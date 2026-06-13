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
            if (data?.authCodeUrl) {
                globalThis.location.href = data.authCodeUrl;
            } else {
                toast.error(t('auth.signIn.error'), {
                    description: `${response.status}: ${error?.message ?? t('auth.signIn.noErrorDetails')}`,
                });
            }
        },
        onError: (error) => {
            toast.error(t('auth.signIn.error'), { description: error.message });
        },
    });
};
