import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Plus, Trash2 } from 'lucide-react';
import { toast } from 'sonner';
import { Button } from '@/design-system/components/ui/button.tsx';
import { Field, FieldDescription, FieldGroup, FieldLabel, FieldSet } from '@/design-system/components/ui/field.tsx';
import { Input } from '@/design-system/components/ui/input.tsx';
import type { TenantInput } from './use-managed-system-onboarding.ts';

export const TenantsStep = ({
    tenants,
    onChange,
    disabled,
    showTitleWithDescription = true,
    addLabel,
}: {
    tenants: TenantInput[];
    onChange: (next: TenantInput[]) => void;
    disabled: boolean;
    showTitleWithDescription?: boolean;
    addLabel?: string;
}) => {
    const { t } = useTranslation();
    const [externalId, setExternalId] = useState('');
    const [name, setName] = useState('');
    const canAdd = externalId.trim() !== '' && name.trim() !== '';

    const add = () => {
        if (!canAdd) return;
        const id = externalId.trim();
        if (tenants.some(({ externalId }) => externalId === id)) {
            toast.error(t('managedSystems.onboard.tenants.duplicate'));
            return;
        }
        onChange([...tenants, { externalId: id, name: name.trim() }]);
        setExternalId('');
        setName('');
    };

    return (
        <div className="flex flex-col gap-5 px-4 mt-4">
            {showTitleWithDescription && (
                <div className="flex flex-col gap-1">
                    <h3 className="text-sm font-medium">{t('managedSystems.onboard.tenants.title')}</h3>
                    <p className="text-xs text-muted-foreground">{t('managedSystems.onboard.tenants.description')}</p>
                </div>
            )}
            <form action={add} className="flex flex-col gap-5">
                <FieldSet>
                    <FieldGroup>
                        <Field>
                            <FieldLabel htmlFor="tenant-external-id">
                                {t('managedSystems.onboard.tenants.externalId')}
                            </FieldLabel>
                            <Input
                                id="tenant-external-id"
                                type="text"
                                disabled={disabled}
                                value={externalId}
                                onChange={(e) => setExternalId(e.target.value)}
                            />
                            <FieldDescription>
                                {t('managedSystems.onboard.tenants.externalIdDescription')}
                            </FieldDescription>
                        </Field>
                        <Field>
                            <FieldLabel htmlFor="tenant-name">{t('managedSystems.onboard.tenants.name')}</FieldLabel>
                            <Input
                                id="tenant-name"
                                type="text"
                                disabled={disabled}
                                value={name}
                                onChange={(e) => setName(e.target.value)}
                            />
                            <FieldDescription>{t('managedSystems.onboard.tenants.nameDescription')}</FieldDescription>
                        </Field>
                    </FieldGroup>
                </FieldSet>
                <Button type="submit" className="self-end" disabled={disabled || !canAdd}>
                    <Plus /> {addLabel ?? t('managedSystems.onboard.tenants.add')}
                </Button>
            </form>
            {tenants.length === 0 ? (
                <p className="rounded-lg border border-dashed p-3 text-xs text-muted-foreground">
                    {t('managedSystems.onboard.tenants.empty')}
                </p>
            ) : (
                <ul className="flex flex-col divide-y rounded-lg border">
                    {tenants.map((tenant) => (
                        <li
                            key={tenant.externalId}
                            className="flex items-center justify-between gap-3 py-1.5 pr-1.5 pl-3 text-sm"
                        >
                            <span className="font-medium">{tenant.name}</span>
                            <div className="flex items-center gap-2">
                                <code className="font-mono text-xs text-muted-foreground">{tenant.externalId}</code>
                                <Button
                                    variant="destructive"
                                    size="icon"
                                    disabled={disabled}
                                    aria-label={t('managedSystems.onboard.tenants.remove', { name: tenant.name })}
                                    onClick={() => onChange(tenants.filter((x) => x.externalId !== tenant.externalId))}
                                >
                                    <Trash2 />
                                </Button>
                            </div>
                        </li>
                    ))}
                </ul>
            )}
        </div>
    );
};
