import { useTranslation } from 'react-i18next';
import { Globe, KeyRound, Layers, UserCog } from 'lucide-react';
import { cn } from '@/design-system/lib/utils.ts';
import type { ConnectorSpec, Operation } from './types.ts';

const methodTone: Record<string, string> = {
    GET: 'text-emerald-600 border-emerald-500/30 bg-emerald-500/10',
    POST: 'text-sky-600 border-sky-500/30 bg-sky-500/10',
    PUT: 'text-amber-600 border-amber-500/30 bg-amber-500/10',
    PATCH: 'text-violet-600 border-violet-500/30 bg-violet-500/10',
    DELETE: 'text-red-600 border-red-500/30 bg-red-500/10',
};

const MethodBadge = ({ method }: { method: string }) => (
    <span
        className={cn(
            'inline-flex h-5 w-16 shrink-0 items-center justify-center rounded border font-mono text-[11px] font-semibold',
            methodTone[method.toUpperCase()] ?? 'text-muted-foreground border-border bg-muted/40',
        )}
    >
        {method.toUpperCase()}
    </span>
);

// Operations live under optional keys (create/enable/.../grant); drop the absent ones.
const definedOps = (group: Record<string, Operation | undefined>): [string, Operation][] =>
    Object.entries(group).filter((entry): entry is [string, Operation] => Boolean(entry[1]));

const countOps = (spec: ConnectorSpec) => {
    const groups = [...Object.values(spec.actions), ...Object.values(spec.entitlements)];
    const ops = groups.flatMap((group) => definedOps(group).map(([, op]) => op));
    return { operations: ops.length, endpoints: ops.reduce((n, op) => n + op.endpoints.length, 0) };
};

const OperationRow = ({ name, op }: { name: string; op: Operation }) => (
    <li className="flex flex-col gap-2 px-3 py-3">
        <div className="flex items-baseline justify-between gap-3">
            <span className="font-mono text-xs font-medium">{name}</span>
            <span className="text-right text-xs text-muted-foreground">{op.description}</span>
        </div>
        <ul className="flex flex-col gap-1.5">
            {op.endpoints.map((ep, i) => (
                <li key={i} className="flex min-w-0 items-start gap-2">
                    <MethodBadge method={ep.method} />
                    <div className="flex min-w-0 flex-col">
                        <code className="break-all font-mono text-xs text-foreground">{ep.url}</code>
                        <span className="text-xs text-muted-foreground">{ep.description}</span>
                    </div>
                </li>
            ))}
        </ul>
    </li>
);

const OperationGroup = ({ title, kind, group }: { title: string; kind: string; group: Record<string, Operation | undefined> }) => {
    const { t } = useTranslation();
    const ops = definedOps(group);
    return (
        <div className="overflow-hidden rounded-lg border">
            <div className="flex items-center justify-between border-b bg-muted/40 px-3 py-2">
                <span className="text-xs font-medium">
                    <span className="text-muted-foreground">{kind} · </span>
                    <span className="font-mono">{title}</span>
                </span>
                <span className="text-xs text-muted-foreground">
                    {t('managedSystems.onboard.capability.operationCount', { count: ops.length })}
                </span>
            </div>
            <ul className="divide-y">
                {ops.map(([name, op]) => (
                    <OperationRow key={name} name={name} op={op} />
                ))}
            </ul>
        </div>
    );
};

const CapabilitySummary = ({ spec }: { spec: ConnectorSpec }) => {
    const { t } = useTranslation();
    const { operations, endpoints } = countOps(spec);
    const stats = [
        { id: 'domains', label: t('managedSystems.onboard.capability.stats.domains'), value: spec.allowedDomains.length },
        { id: 'operations', label: t('managedSystems.onboard.capability.stats.operations'), value: operations },
        { id: 'endpoints', label: t('managedSystems.onboard.capability.stats.endpoints'), value: endpoints },
        { id: 'config', label: t('managedSystems.onboard.capability.stats.config'), value: spec.config.length },
    ];
    return (
        <dl className="grid grid-cols-2 gap-px overflow-hidden rounded-lg border bg-border sm:grid-cols-4">
            {stats.map((s) => (
                <div key={s.id} className="flex flex-col gap-0.5 bg-card px-3 py-2.5">
                    <dt className="text-xs text-muted-foreground">{s.label}</dt>
                    <dd className="text-lg font-semibold tabular-nums">{s.value}</dd>
                </div>
            ))}
        </dl>
    );
};

const Section = ({ icon, title, hint, children }: { icon: React.ReactNode; title: string; hint?: string; children: React.ReactNode }) => (
    <section className="flex flex-col gap-2">
        <h3 className="flex items-center gap-2 text-sm font-medium">
            {icon}
            {title}
        </h3>
        {hint && <p className="text-xs text-muted-foreground">{hint}</p>}
        {children}
    </section>
);

export const CapabilityReview = ({ spec }: { spec: ConnectorSpec }) => {
    const { t } = useTranslation();
    const iconCls = 'size-4 text-muted-foreground';
    return (
        <div className="flex flex-col gap-6">
            <CapabilitySummary spec={spec} />

            <Section
                icon={<Globe className={iconCls} aria-hidden />}
                title={t('managedSystems.onboard.capability.egress.title')}
                hint={t('managedSystems.onboard.capability.egress.hint')}
            >
                <ul className="flex flex-wrap gap-1.5">
                    {spec.allowedDomains.map((domain) => (
                        <li key={domain} className="rounded-md border bg-muted/40 px-2 py-1 font-mono text-xs">
                            {domain}
                        </li>
                    ))}
                </ul>
            </Section>

            <Section icon={<KeyRound className={iconCls} aria-hidden />} title={t('managedSystems.onboard.capability.requiredConfig')}>
                <ul className="divide-y overflow-hidden rounded-lg border">
                    {spec.config.map((c) => (
                        <li key={c.name} className="flex items-center justify-between gap-3 px-3 py-2">
                            <div className="flex min-w-0 flex-col">
                                <code className="font-mono text-xs">{c.name}</code>
                                <span className="text-xs text-muted-foreground">{c.description}</span>
                            </div>
                            <span className="shrink-0 text-xs text-muted-foreground">
                                {c.required
                                    ? t('managedSystems.onboard.capability.required')
                                    : t('managedSystems.onboard.capability.optional')}
                            </span>
                        </li>
                    ))}
                </ul>
            </Section>

            <Section icon={<UserCog className={iconCls} aria-hidden />} title={t('managedSystems.onboard.capability.accountActions')}>
                {Object.entries(spec.actions).map(([resource, group]) => (
                    <OperationGroup
                        key={resource}
                        title={resource}
                        kind={t('managedSystems.onboard.capability.kind.account')}
                        group={group}
                    />
                ))}
            </Section>

            <Section icon={<Layers className={iconCls} aria-hidden />} title={t('managedSystems.onboard.capability.entitlementOps')}>
                {Object.entries(spec.entitlements).map(([resource, group]) => (
                    <OperationGroup
                        key={resource}
                        title={resource}
                        kind={t('managedSystems.onboard.capability.kind.entitlement')}
                        group={group}
                    />
                ))}
            </Section>
        </div>
    );
};
