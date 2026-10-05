import { useCurrentUser } from '@/hooks/use-current-user.ts';

export const usePermission = () => {
    const { roles } = useCurrentUser();

    const isAdmin = roles.includes('admin');

    return {
        isAdmin,
    };
};
