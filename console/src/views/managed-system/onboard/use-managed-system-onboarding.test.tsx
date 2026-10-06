import { useManagedSystemOnboarding } from './use-managed-system-onboarding.ts';
import { toast } from 'sonner';
import { mockHttpHandlers, mockServer } from '@/test-utils/msw.ts';
import { HttpResponse } from 'msw';
import { act, renderHook, waitFor } from '@testing-library/react';
import { Wrapper } from '@/test-utils/common-wrappers.tsx';

vi.mock('@/utils/openapi/client.ts');

const mockedToastError = vi.mocked(toast.error);

describe('useManagedSystemOnboarding', () => {
    beforeEach(() => {
        mockedToastError.mockReset();
    });

    it('should return idle state before submitting', () => {
        const { result } = renderHook(useManagedSystemOnboarding, { wrapper: Wrapper });

        expect(result.current.state).toBe('idle');
    });

    it('should onboard the managed system and resolve to success when validation succeeds', async () => {
        mockServer.use(
            mockHttpHandlers.post('/api/v1/connectors/onboarding-requests', () =>
                HttpResponse.json({ onboardingId: 'onboard-id' }, { status: 201 }),
            ),
            mockHttpHandlers.get('/api/v1/connectors/onboarding-requests/{onboarding-id}', () =>
                HttpResponse.json({ status: 'success' }, { status: 200 }),
            ),
            mockHttpHandlers.post('/api/v1/managed-systems', () =>
                HttpResponse.json({ id: 'system-id' }, { status: 200 }),
            ),
        );

        const { result } = renderHook(useManagedSystemOnboarding, { wrapper: Wrapper });

        result.current.submit({ connectorUrl: 'https://example.com/x.wasm', connectorSha: 'sha' }).catch(() => {});

        await waitFor(() => expect(result.current.state).toBe('success'));

        expect(result.current.error).toBeUndefined();
        expect(mockedToastError).not.toHaveBeenCalled();
    });

    it('should show an error toast and resolve to error when connector validation fails', async () => {
        mockServer.use(
            mockHttpHandlers.post('/api/v1/connectors/onboarding-requests', () =>
                HttpResponse.json({ onboardingId: 'onboard-id' }, { status: 201 }),
            ),
            mockHttpHandlers.get('/api/v1/connectors/onboarding-requests/{onboarding-id}', () =>
                HttpResponse.json({ status: 'failed', error: 'validation failed' }, { status: 200 }),
            ),
        );

        const { result } = renderHook(useManagedSystemOnboarding, { wrapper: Wrapper });

        result.current.submit({ connectorUrl: 'https://example.com/x.wasm', connectorSha: 'sha' }).catch(() => {});

        await waitFor(() => expect(result.current.state).toBe('error'));

        expect(mockedToastError).toHaveBeenCalledWith('managedSystems.onboard.onboardingError', {
            description: 'validation failed',
        });
    });

    it('should show an error toast and resolve to error when managed system onboarding fails', async () => {
        mockServer.use(
            mockHttpHandlers.post('/api/v1/connectors/onboarding-requests', () =>
                HttpResponse.json({ onboardingId: 'onboard-id' }, { status: 201 }),
            ),
            mockHttpHandlers.get('/api/v1/connectors/onboarding-requests/{onboarding-id}', () =>
                HttpResponse.json({ status: 'success' }, { status: 200 }),
            ),
            mockHttpHandlers.post('/api/v1/managed-systems', () => HttpResponse.error()),
        );

        const { result } = renderHook(useManagedSystemOnboarding, { wrapper: Wrapper });

        result.current.submit({ connectorUrl: 'https://example.com/x.wasm', connectorSha: 'sha' }).catch(() => {});

        await waitFor(() => expect(result.current.state).toBe('error'));

        expect(mockedToastError).toHaveBeenCalled();
    });

    it('should show an error toast and resolve to error when creating the onboarding request fails', async () => {
        mockServer.use(
            mockHttpHandlers.post('/api/v1/connectors/onboarding-requests', () =>
                HttpResponse.json(null, { status: 500 }),
            ),
        );

        const { result } = renderHook(useManagedSystemOnboarding, { wrapper: Wrapper });

        result.current.submit({ connectorUrl: 'https://example.com/x.wasm', connectorSha: 'sha' }).catch(() => {});

        await waitFor(() => expect(result.current.state).toBe('error'));

        expect(mockedToastError).toHaveBeenCalledWith('managedSystems.onboard.onboardingError', {
            description: 'managedSystems.onboard.validationError',
        });
    });

    it('should show an error toast and resolve to error when validation times out', async () => {
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
                result.current.submit({ connectorUrl: 'https://example.com/x.wasm', connectorSha: 'sha' }).catch(() => {});
            });

            await act(async () => {
                await vi.advanceTimersByTimeAsync(90 * 2000 + 1000);
            });

            expect(result.current.state).toBe('error');
            expect(mockedToastError).toHaveBeenCalledWith('managedSystems.onboard.onboardingError', {
                description: 'managedSystems.onboard.onboardingError',
            });
        } finally {
            vi.useRealTimers();
        }
    });
});
