import { useSignIn } from './use-sign-in.ts';
import { toast } from 'sonner';
import { renderHook, waitFor } from '@testing-library/react';
import { Wrapper } from '@/test-utils/common-wrappers.tsx';
import { mockHttpHandlers, mockServer } from '@/test-utils/msw.ts';
import { HttpResponse } from 'msw';

vi.mock('@/utils/openapi/client.ts');

const toastErrorMock = vi.mocked(toast.error);

describe('useSignIn', () => {
    beforeEach(() => {
        vi.stubGlobal('location', { href: '' });
    });

    afterEach(() => {
        vi.unstubAllGlobals();
        toastErrorMock.mockReset();
    });

    it('redirects to auth code url on success', async () => {
        const authCodeUrlMock = 'auth-code-url-mock';
        mockServer.use(
            mockHttpHandlers.get('/api/v1/auth/{provider}', () => {
                return HttpResponse.json({ authCodeUrl: authCodeUrlMock }, { status: 200 });
            }),
        );

        const { result } = renderHook(useSignIn, { wrapper: Wrapper });

        result.current.mutate('google');
        await waitFor(() => expect(globalThis.location.href).toBe(authCodeUrlMock));

        expect(toastErrorMock).not.toHaveBeenCalled();
    });

    it('shows an error toast when auth url is missing', async () => {
        const messageMock = 'auth-code-url-mock';
        mockServer.use(
            mockHttpHandlers.get('/api/v1/auth/{provider}', () => {
                return HttpResponse.json({ message: messageMock }, { status: 500 });
            }),
        );

        const { result } = renderHook(useSignIn, { wrapper: Wrapper });

        result.current.mutate('google');
        await waitFor(() => expect(result.current.isSuccess).toBe(true));

        expect(toastErrorMock).toHaveBeenCalledTimes(1);
    });
});
