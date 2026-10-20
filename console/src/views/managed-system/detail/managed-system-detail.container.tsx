import { PageWrapper } from '@/components/page-wrapper.tsx';
import { useQuery } from '@tanstack/react-query';
import { fetchClient } from '@/utils/openapi/client.ts';
import { useTranslation } from 'react-i18next';
import type { ReactNode } from 'react';
import { getRouteApi, Link } from '@tanstack/react-router';
import { Button } from '@/design-system/components/ui/button.tsx';
import { Spinner } from '@/design-system/components/ui/spinner.tsx';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/design-system/components/ui/tabs.tsx';
import { CapabilityReview } from '@/views/managed-system/onboard/capability-review.tsx';
import { TenantsPanel } from './tenants-panel.tsx';

const routeApi = getRouteApi('/managed-systems/$id');

const TAB = { tenants: 'tenants', spec: 'spec' } as const;

const shortHash = (hash: string) => `${hash.slice(0, 16)}…${hash.slice(-4)}`;

const Meta = ({ label, value }: { label: string; value: ReactNode }) => (
    <div className="flex min-w-0 flex-col gap-1 bg-card px-4 py-3">
        <dt className="text-xs text-muted-foreground">{label}</dt>
        <dd className="min-w-0 text-sm">{value}</dd>
    </div>
);

export const ManagedSystemDetailContainer = () => {
    const { t } = useTranslation();
    const { id } = routeApi.useParams();

    const system = useQuery({
        queryKey: ['managed-system', id],
        queryFn: async () => {
            const { data, error } = await fetchClient.GET('/api/v1/managed-systems/{id}', {
                params: { path: { id } },
            });
            if (error) throw error;
            return data;
        },
    });

    const tenants = useQuery({
        queryKey: ['managed-system-tenants', id],
        queryFn: async () => {
            const { data, error } = await fetchClient.GET('/api/v1/managed-systems/{id}/tenant', {
                params: { path: { id } },
            });
            if (error) throw error;
            return data;
        },
    });

    if (system.isLoading) {
        return (
            <PageWrapper title={t('managedSystems.detail.title')} description="">
                <Spinner />
            </PageWrapper>
        );
    }

    if (system.error || !system.data) {
        return (
            <PageWrapper title={t('managedSystems.detail.notFound')} description="">
                <Button variant="outline" render={<Link to="/managed-systems" />}>
                    {t('managedSystems.detail.backToList')}
                </Button>
            </PageWrapper>
        );
    }

    const { name, connectorHash, connectorSpec, connectorUrl, createdAt } = system.data;

    return (
        <PageWrapper title={name} description={connectorSpec.description}>
            <dl className="grid grid-cols-1 gap-px overflow-hidden rounded-xl border bg-border sm:grid-cols-2 lg:grid-cols-4">
                <Meta
                    label={t('managedSystems.onboard.connectorUrl.name')}
                    value={
                        <span className="block truncate font-mono text-xs" title={connectorUrl}>
                            {connectorUrl}
                        </span>
                    }
                />
                <Meta
                    label={t('managedSystems.onboard.connectorHash.name')}
                    value={
                        <code className="font-mono text-xs" title={connectorHash}>
                            {shortHash(connectorHash)}
                        </code>
                    }
                />
                <Meta
                    label={t('managedSystems.detail.tenants')}
                    value={<span className="tabular-nums">{tenants.data?.length ?? 0}</span>}
                />
                <Meta label={t('managedSystems.detail.onboarded')} value={new Date(createdAt).toLocaleString()} />
            </dl>

            <Tabs defaultValue={TAB.tenants}>
                <TabsList variant="line" className="w-full justify-start rounded-none border-b">
                    <TabsTrigger value={TAB.tenants} className="flex-none">
                        {t('managedSystems.detail.tabs.tenants')}
                    </TabsTrigger>
                    <TabsTrigger value={TAB.spec} className="flex-none">
                        {t('managedSystems.detail.tabs.spec')}
                    </TabsTrigger>
                </TabsList>
                <TabsContent value={TAB.tenants} className="pt-4">
                    <TenantsPanel systemId={id} query={tenants} />
                </TabsContent>
                <TabsContent value={TAB.spec} className="pt-4">
                    <CapabilityReview spec={connectorSpec} missingEnvVars={[]} />
                </TabsContent>
            </Tabs>
        </PageWrapper>
    );
};
