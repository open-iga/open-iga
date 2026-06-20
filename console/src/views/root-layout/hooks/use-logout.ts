import { useMutation } from '@tanstack/react-query';
import { fetchClient } from '@/utils/openapi/client.ts';
import { toast } from 'sonner';
import { useNavigate } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';

export const useLogout = () => {
    const { t } = useTranslation();
    const navigate = useNavigate();

    const { mutate: logout, isPending } = useMutation({
        mutationFn: () => fetchClient.POST('/api/v1/auth/logout'),
        onSuccess: ({ data, response }) => {
            if (data?.message) {
                navigate({ to: '/auth/sign-in' }).catch(() => {});
            } else {
                toast.error(t('auth.logout.error'), { description: response.statusText });
            }
        },
    });

    return { logout, isPending };
};
