import type { SupportedOauthProvider } from '@/views/types';
import { fetchClient } from '@/utils/openapi/client.ts';
import { useMutation } from '@tanstack/react-query';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';

export const useSignIn = () => {
    const { t } = useTranslation();

    return useMutation({
        mutationFn: (provider: SupportedOauthProvider) =>
            fetchClient.GET('/api/v1/auth/{provider}', {
                params: { path: { provider } },
            }),
        onSuccess: ({ data, error, response }) => {
            if (data && 'authCodeUrl' in data) {
                globalThis.location.href = data.authCodeUrl;
                return;
            }

            if (data && 'redirect' in data) {
                globalThis.location.href = data.redirect;
                return;
            }

            toast.error(t('auth.login.error'), {
                description: `${response.status}: ${error?.message ?? t('auth.login.noErrorDetails')}`,
            });
        },
        onError: (error) => {
            toast.error(t('auth.login.error'), { description: error.message });
        },
    });
};
