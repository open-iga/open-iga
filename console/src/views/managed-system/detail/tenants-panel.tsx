import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Plus } from 'lucide-react';
import { toast } from 'sonner';
import { useQueryClient, type UseQueryResult } from '@tanstack/react-query';
import { match, P } from 'ts-pattern';
import { fetchClient } from '@/utils/openapi/client.ts';
import type { operations } from '@/utils/openapi/schema';
import { DataTable, type DataTableColumn, type DataTableProps } from '@/design-system/components/ui/data-table.tsx';
import { Button } from '@/design-system/components/ui/button.tsx';
import { Spinner } from '@/design-system/components/ui/spinner.tsx';
import {
    Sheet,
    SheetClose,
    SheetContent,
    SheetDescription,
    SheetFooter,
    SheetHeader,
    SheetTitle,
    SheetTrigger,
} from '@/design-system/components/ui/sheet.tsx';
import { Separator } from '@/design-system/components/ui/separator.tsx';
import { usePermission } from '@/hooks/use-permission.ts';
import { TenantsStep } from '@/views/managed-system/onboard/tenants-step.tsx';
import type { TenantInput } from '@/views/managed-system/onboard/use-managed-system-onboarding.ts';

type Tenant = operations['listManagedSystemTenants']['responses'][200]['content']['application/json'][number];
type TenantsQuery = UseQueryResult<Tenant[]>;

const AddTenantSheet = ({ systemId }: { systemId: string }) => {
    const { t } = useTranslation();
    const queryClient = useQueryClient();
    const [open, setOpen] = useState(false);
    const [tenants, setTenants] = useState<TenantInput[]>([]);
    const [busy, setBusy] = useState(false);

    const close = () => {
        setTenants([]);
        setOpen(false);
    };

    const submit = async () => {
        if (tenants.length === 0) return;
        setBusy(true);
        try {
            for (const tenant of tenants) {
                const { error, response } = await fetchClient.POST('/api/v1/managed-systems/{id}/tenant', {
                    params: { path: { id: systemId } },
                    body: { externalId: tenant.externalId, name: tenant.name },
                });
                if (error) {
                    toast.error(
                        response.status === 409
                            ? t('managedSystems.detail.tenantsPanel.nameExists', { name: tenant.name })
                            : t('managedSystems.detail.tenantsPanel.addError'),
                    );
                    return;
                }
            }
            toast.success(t('managedSystems.detail.tenantsPanel.added', { count: tenants.length }));
            close();
        } catch {
            toast.error(t('managedSystems.detail.tenantsPanel.addError'));
        } finally {
            // Refresh even on a partial failure so successfully onboarded tenants show up
            await queryClient.invalidateQueries({ queryKey: ['managed-system-tenants', systemId] });
            setBusy(false);
        }
    };

    return (
        <Sheet
            open={open}
            onOpenChange={(next) =>
                match({ busy, next })
                    .with({ busy: true }, () => {}) // ignore open/close changes while a batch is applying
                    .with({ next: true }, () => setOpen(true))
                    .otherwise(close)
            }
        >
            <SheetTrigger
                render={
                    <Button>
                        <Plus /> {t('managedSystems.detail.tenantsPanel.add')}
                    </Button>
                }
            />
            <SheetContent side="right" className="min-w-1/3 overflow-y-auto">
                <SheetHeader className="p-4">
                    <SheetTitle>{t('managedSystems.detail.tenantsPanel.addTitle')}</SheetTitle>
                    <SheetDescription>{t('managedSystems.detail.tenantsPanel.addDescription')}</SheetDescription>
                </SheetHeader>
                <Separator />
                <TenantsStep
                    showTitleWithDescription={false}
                    addLabel={t('managedSystems.detail.tenantsPanel.batchChanges')}
                    tenants={tenants}
                    onChange={setTenants}
                    disabled={busy}
                />
                <SheetFooter className="sticky bottom-0 flex-row justify-end border-t bg-popover">
                    <SheetClose
                        render={
                            <Button variant="outline" disabled={busy}>
                                {t('managedSystems.onboard.cancel')}
                            </Button>
                        }
                    />
                    <Button disabled={busy || tenants.length === 0} onClick={submit}>
                        {busy && <Spinner />}
                        {t('managedSystems.detail.tenantsPanel.add')}
                    </Button>
                </SheetFooter>
            </SheetContent>
        </Sheet>
    );
};

export const TenantsPanel = ({ systemId, query }: { systemId: string; query: TenantsQuery }) => {
    const { t } = useTranslation();
    const { isAdmin } = usePermission();

    const columns: DataTableColumn<Tenant>[] = [
        { accessorKey: 'name', header: t('managedSystems.detail.tenantsPanel.columns.name') },
        {
            accessorKey: 'externalId',
            header: t('managedSystems.detail.tenantsPanel.columns.externalId'),
            cell: ({ row }) => (
                <code className="font-mono text-xs text-muted-foreground">{row.original.externalId}</code>
            ),
        },
        {
            accessorKey: 'createdAt',
            header: t('managedSystems.detail.tenantsPanel.columns.onboarded'),
            cell: ({ row }) => (
                <span className="font-mono text-xs text-muted-foreground">
                    {new Date(row.original.createdAt).toLocaleString()}
                </span>
            ),
        },
    ];

    const props = match<typeof query, DataTableProps<Tenant>>(query)
        .with({ isLoading: true }, () => ({ status: 'loading', columns }))
        .with({ data: P.when((d) => d === undefined || d.length === 0) }, () => ({
            status: 'noData',
            columns,
            emptyMessage: {
                name: t('managedSystems.detail.tenantsPanel.empty.name'),
                description: t('managedSystems.detail.tenantsPanel.empty.description'),
            },
        }))
        .with({ data: P.not(undefined) }, ({ data }) => ({ status: 'data', data, columns }))
        .otherwise(() => ({ status: 'error', columns, errorMessage: t('error.generic') }));

    return (
        <div className="flex flex-col gap-4">
            {isAdmin && (
                <div className="flex flex-row-reverse">
                    <AddTenantSheet systemId={systemId} />
                </div>
            )}
            <DataTable {...props} />
        </div>
    );
};
