import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { useSignInCallback } from './use-sign-in-callback.ts';
import { toast } from 'sonner';
import { mockHttpHandlers, mockServer } from '@/test-utils/msw.ts';
import { HttpResponse } from 'msw';
import { renderHook, waitFor } from '@testing-library/react';
import { Wrapper } from '@/test-utils/common-wrappers.tsx';

const navigateMock = vi.hoisted(() => vi.fn(() => Promise.resolve()));
vi.mock('@tanstack/react-router', () => ({
    useNavigate: () => navigateMock,
}));

vi.mock('@/utils/openapi/client.ts');

const mockedToastError = vi.mocked(toast.error);

describe('useSignInCallback', () => {
    beforeEach(() => {
        mockedToastError.mockReset();
        navigateMock.mockClear();
        vi.stubGlobal('location', { href: '' } as unknown as Location);
    });

    afterEach(() => {
        vi.unstubAllGlobals();
    });

    it('redirects to the home page on success', async () => {
        mockServer.use(
            mockHttpHandlers.post('/api/v1/auth/{provider}/callback', () => {
                return HttpResponse.json({ redirect: '/home' }, { status: 200 });
            }),
        );

        const { result } = renderHook(
            () => useSignInCallback({ provider: 'google', code: 'code-1', state: 'state-1' }),
            {
                wrapper: Wrapper,
            },
        );

        result.current.mutate();
        await waitFor(() => expect(result.current.isSuccess).toBe(true));

        expect(globalThis.location.href).toBe('/home');
        expect(mockedToastError).not.toHaveBeenCalled();
        expect(navigateMock).not.toHaveBeenCalled();
    });

    it('shows an error toast and navigates to sign in when redirect is missing', async () => {
        mockServer.use(
            mockHttpHandlers.post('/api/v1/auth/{provider}/callback', () => {
                return HttpResponse.json(null, { status: 500 });
            }),
        );

        const { result } = renderHook(
            () => useSignInCallback({ provider: 'google', code: 'code-1', state: 'state-1' }),
            {
                wrapper: Wrapper,
            },
        );

        result.current.mutate();
        await waitFor(() => expect(mockedToastError).toHaveBeenCalled());

        expect(mockedToastError).toHaveBeenCalledWith('auth.login.error', {
            description: '500: auth.login.noErrorDetails',
        });
        expect(navigateMock).toHaveBeenCalledWith({ to: '/auth/sign-in' });
        expect(globalThis.location.href).toBe('');
    });
});
