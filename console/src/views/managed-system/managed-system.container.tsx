import { PageWrapper } from '@/components/page-wrapper.tsx';
import { OnboardingSheet } from '@/views/managed-system/onboard/onboarding-sheet.tsx';
import { useQuery } from '@tanstack/react-query';
import { fetchClient } from '@/utils/openapi/client.ts';
import { DataTable, type DataTableColumn, type DataTableProps } from '@/design-system/components/ui/data-table.tsx';
import { useTranslation } from 'react-i18next';
import { match, P } from 'ts-pattern';

export const ManagedSystemContainer = () => {
    const { t } = useTranslation();
    const query = useQuery({
        queryKey: ['managed-systems'],
        queryFn: () => fetchClient.GET('/api/v1/managed-systems'),
    });

    const managedSystems = query.data?.data ?? [];
    const columns: DataTableColumn<(typeof managedSystems)[0]>[] = [
        { accessorKey: 'id', header: t('managedSystems.name'), meta: { hidden: true } },
        { accessorKey: 'name', header: t('managedSystems.name') },
        { accessorKey: 'connectorUrl', header: t('managedSystems.onboard.connectorUrl.name') },
        { accessorKey: 'connectorHash', header: t('managedSystems.onboard.connectorHash.name') },
        {
            accessorKey: 'createdAt',
            header: t('managedSystems.onboard.createAt'),
            cell: ({ row }) => new Date(row.original.createdAt).toLocaleString(),
        },
    ];

    const props = match<typeof query, DataTableProps<(typeof managedSystems)[0]>>(query)
        .with({ isLoading: true }, () => ({ status: 'loading', columns }))
        .with({ data: { data: P.when((t) => t === undefined || t?.length === 0) } }, () => ({
            status: 'noData',
            columns,
            emptyMessage: {
                name: t('managedSystems.emptyManagedSystem.name'),
                description: t('managedSystems.emptyManagedSystem.description'),
            },
        }))
        .with({ data: { data: P.not(undefined) } }, ({ data }) => ({
            status: 'data',
            data: data.data,
            columns,
        }))
        .otherwise(() => ({ status: 'error', columns, errorMessage: t('error.generic') }));

    return (
        <PageWrapper title={t('managedSystems.name')} description={t('managedSystems.description')}>
            <div className="flex flex-row-reverse">
                <OnboardingSheet />
            </div>
            <DataTable {...props} />
        </PageWrapper>
    );
};
