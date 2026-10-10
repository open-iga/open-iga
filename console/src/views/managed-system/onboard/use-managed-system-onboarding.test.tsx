import { useManagedSystemOnboarding } from './use-managed-system-onboarding.ts';
import { toast } from 'sonner';
import { mockHttpHandlers, mockServer } from '@/test-utils/msw.ts';
import { HttpResponse } from 'msw';
import { act, renderHook, waitFor } from '@testing-library/react';
import { Wrapper } from '@/test-utils/common-wrappers.tsx';

vi.mock('@/utils/openapi/client.ts');

const mockedToastError = vi.mocked(toast.error);

const spec = {
    name: 'slack',
    description: 'Slack connector',
    config: [],
    allowedDomains: ['slack.com'],
    actions: {},
    entitlements: {},
};

const body = { connectorUrl: 'https://example.com/x.wasm', connectorSha: 'sha' };

describe('useManagedSystemOnboarding', () => {
    beforeEach(() => {
        mockedToastError.mockReset();
    });

    it('should return input phase before submitting', () => {
        const { result } = renderHook(useManagedSystemOnboarding, { wrapper: Wrapper });

        expect(result.current.phase).toBe('input');
    });

    it('should move to review and expose the spec when validation succeeds', async () => {
        mockServer.use(
            mockHttpHandlers.post('/api/v1/connectors/onboarding-requests', () =>
                HttpResponse.json({ onboardingId: 'onboard-id' }, { status: 201 }),
            ),
            mockHttpHandlers.get('/api/v1/connectors/onboarding-requests/{onboarding-id}', () =>
                HttpResponse.json({ status: 'success', connectorSpec: spec }, { status: 200 }),
            ),
        );

        const { result } = renderHook(useManagedSystemOnboarding, { wrapper: Wrapper });

        result.current.validate(body).catch(() => {});

        await waitFor(() => expect(result.current.phase).toBe('review'));
        expect(result.current.spec?.name).toBe('slack');
        expect(mockedToastError).not.toHaveBeenCalled();
    });

    it('should stay on input and surface the error when validation fails', async () => {
        mockServer.use(
            mockHttpHandlers.post('/api/v1/connectors/onboarding-requests', () =>
                HttpResponse.json({ onboardingId: 'onboard-id' }, { status: 201 }),
            ),
            mockHttpHandlers.get('/api/v1/connectors/onboarding-requests/{onboarding-id}', () =>
                HttpResponse.json({ status: 'failed', error: 'validation failed' }, { status: 200 }),
            ),
        );

        const { result } = renderHook(useManagedSystemOnboarding, { wrapper: Wrapper });

        result.current.validate(body).catch(() => {});

        await waitFor(() => expect(result.current.error).toBe('validation failed'));
        expect(result.current.phase).toBe('input');
        expect(mockedToastError).toHaveBeenCalledWith('managedSystems.onboard.validationError', {
            description: 'validation failed',
        });
    });

    it('should strip a sha256: prefix from the hash before sending the request', async () => {
        let sentBody: { connectorUrl: string; connectorSha: string } | undefined;
        mockServer.use(
            mockHttpHandlers.post('/api/v1/connectors/onboarding-requests', async ({ request }) => {
                sentBody = (await request.json()) as typeof sentBody;
                return HttpResponse.json({ onboardingId: 'onboard-id' }, { status: 201 });
            }),
            mockHttpHandlers.get('/api/v1/connectors/onboarding-requests/{onboarding-id}', () =>
                HttpResponse.json({ status: 'success', connectorSpec: spec }, { status: 200 }),
            ),
        );

        const { result } = renderHook(useManagedSystemOnboarding, { wrapper: Wrapper });

        result.current
            .validate({ connectorUrl: 'https://example.com/x.wasm', connectorSha: 'sha256:abc123' })
            .catch(() => {});

        await waitFor(() => expect(result.current.phase).toBe('review'));
        expect(sentBody?.connectorSha).toBe('abc123');
    });

    it('should error when creating the onboarding request fails', async () => {
        mockServer.use(
            mockHttpHandlers.post('/api/v1/connectors/onboarding-requests', () =>
                HttpResponse.json(null, { status: 500 }),
            ),
        );

        const { result } = renderHook(useManagedSystemOnboarding, { wrapper: Wrapper });

        result.current.validate(body).catch(() => {});

        await waitFor(() =>
            expect(mockedToastError).toHaveBeenCalledWith('managedSystems.onboard.validationError', {
                description: 'managedSystems.onboard.validationError',
            }),
        );
        expect(result.current.phase).toBe('input');
    });

    it('should toast onboarding error when onboarding the managed system fails', async () => {
        mockServer.use(
            mockHttpHandlers.post('/api/v1/connectors/onboarding-requests', () =>
                HttpResponse.json({ onboardingId: 'onboard-id' }, { status: 201 }),
            ),
            mockHttpHandlers.get('/api/v1/connectors/onboarding-requests/{onboarding-id}', () =>
                HttpResponse.json({ status: 'success', connectorSpec: spec }, { status: 200 }),
            ),
            mockHttpHandlers.post('/api/v1/managed-systems', () => HttpResponse.error()),
        );

        const { result } = renderHook(useManagedSystemOnboarding, { wrapper: Wrapper });

        result.current.validate(body).catch(() => {});
        await waitFor(() => expect(result.current.phase).toBe('review'));

        await act(async () => {
            await result.current.onboard().catch(() => {});
        });

        expect(mockedToastError).toHaveBeenCalledWith('managedSystems.onboard.onboardingError', expect.anything());
    });

    it('should error when validation times out', async () => {
        vi.useFakeTimers();
        try {
            mockServer.use(
                mockHttpHandlers.post('/api/v1/connectors/onboarding-requests', () =>
                    HttpResponse.json({ onboardingId: 'onboard-id' }, { status: 201 }),
                ),
                mockHttpHandlers.get('/api/v1/connectors/onboarding-requests/{onboarding-id}', () =>
                    HttpResponse.json({ status: 'pending' }, { status: 200 }),
                ),
            );

            const { result } = renderHook(useManagedSystemOnboarding, { wrapper: Wrapper });

            act(() => {
                result.current.validate(body).catch(() => {});
            });

            await act(async () => {
                await vi.advanceTimersByTimeAsync(90 * 2000 + 1000);
            });

            expect(result.current.phase).toBe('input');
            expect(mockedToastError).toHaveBeenCalledWith('managedSystems.onboard.validationError', {
                description: 'managedSystems.onboard.validationError',
            });
        } finally {
            vi.useRealTimers();
        }
    });
});
