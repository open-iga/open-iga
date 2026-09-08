import { useQuery } from '@tanstack/react-query';
import { fetchClient } from '@/utils/openapi/client.ts';

export const useCurrentUser = () => {
    const { isPending, data, isError, error } = useQuery({
        queryKey: ['user-details'],
        queryFn: () => fetchClient.GET('/api/v1/users'),
    });

    return {
        isPending,
        isError,
        error,
        firstName: data?.data?.firstName || 'FirstName',
        lastName: data?.data?.lastName || 'LastName',
    };
};
