import { fetchClient } from '@/utils/openapi/client.ts';
import { useMutation } from '@tanstack/react-query';
import type { SupportedOauthProvider } from '@/views/types';
import { useNavigate } from '@tanstack/react-router';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';

type UseSignInCallbackArgs = {
    state: string;
    code: string;
    provider: SupportedOauthProvider;
};

export const useSignInCallback = ({ state, provider, code }: UseSignInCallbackArgs) => {
    const navigate = useNavigate();
    const { t } = useTranslation();

    return useMutation({
        mutationFn: () =>
            fetchClient.POST('/api/v1/auth/{provider}/callback', {
                params: { path: { provider }, query: { code, state } },
            }),
        onSuccess: async ({ data, response, error }) => {
            if (data?.redirect) {
                globalThis.location.href = data.redirect;
            } else {
                toast.error(t('auth.login.error'), {
                    description: `${response.status}: ${error?.message ?? t('auth.login.noErrorDetails')}`,
                });

                await navigate({ to: '/auth/sign-in' });
            }
        },
        onError: async (error) => {
            toast.error(t('auth.login.error'), {
                description: error.message,
            });

            await navigate({ to: '/auth/sign-in' });
        },
    });
};
