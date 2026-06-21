import { useLogout } from './use-logout.ts';
import { toast } from 'sonner';
import { mockHttpHandlers, mockServer } from '@/test-utils/msw.ts';
import { HttpResponse } from 'msw';
import { renderHook, waitFor } from '@testing-library/react';
import { Wrapper } from '@/test-utils/common-wrappers.tsx';

vi.mock('@/utils/openapi/client.ts');

const navigateMock = vi.hoisted(() => vi.fn(() => Promise.resolve()));
vi.mock('@tanstack/react-router', () => ({
    useNavigate: () => navigateMock,
}));

const mockedToastError = vi.mocked(toast.error);

describe('useLogout', () => {
    beforeEach(() => {
        mockedToastError.mockReset();
        navigateMock.mockClear();
    });

    it('should navigate to sign-in when logout is successful', async () => {
        mockServer.use(
            mockHttpHandlers.post('/api/v1/auth/logout', () => {
                return HttpResponse.json({ message: 'Session deactivated' }, { status: 200 });
            }),
        );

        const { result } = renderHook(useLogout, { wrapper: Wrapper });

        result.current.logout();
        await waitFor(() => expect(navigateMock).toHaveBeenCalledWith({ to: '/auth/sign-in' }));

        expect(mockedToastError).not.toHaveBeenCalled();
    });

    it('should display toast when logout is not successful', async () => {
        mockServer.use(
            mockHttpHandlers.post('/api/v1/auth/logout', () => {
                return HttpResponse.json(null, { status: 500 });
            }),
        );

        const { result } = renderHook(useLogout, { wrapper: Wrapper });

        result.current.logout();
        await waitFor(() => expect(mockedToastError).toHaveBeenCalled());

        expect(mockedToastError).toHaveBeenCalledWith('auth.logout.error', expect.any(Object));
        expect(navigateMock).not.toHaveBeenCalled();
    });
});
