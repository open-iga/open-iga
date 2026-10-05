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
import { Button } from '@/design-system/components/ui/button.tsx';
import { useTranslation } from 'react-i18next';
import { CirclePlus } from 'lucide-react';
import { Field, FieldDescription, FieldGroup, FieldLabel, FieldSet } from '@/design-system/components/ui/field.tsx';
import { Input } from '@/design-system/components/ui/input.tsx';
import { Spinner } from '@/design-system/components/ui/spinner.tsx';
import { useForm } from '@tanstack/react-form';
import { useState } from 'react';
import { usePermission } from '@/hooks/use-permission.ts';
import { useManagedSystemOnboarding } from './use-managed-system-onboarding.ts';

export const OnboardingSheet = () => {
    const { t } = useTranslation();
    const [open, setOpen] = useState(false);
    const { isAdmin } = usePermission();
    const { state, submit } = useManagedSystemOnboarding();

    const processing = state === 'processing';

    const form = useForm({
        defaultValues: {
            connectorUrl: '',
            connectorSha: '',
        },
        onSubmit: async ({ value, formApi }) => {
            try {
                await submit(value);
                formApi.reset();
                setOpen(false);
            } catch {
                // noop: failure is surfaced by the hook's onError toast; keep the sheet open to retry
            }
        },
    });

    const handleOpenChange = (nextOpen: boolean) => {
        if (!nextOpen && processing) {
            return;
        }
        setOpen(nextOpen);
    };

    return (
        <Sheet open={open} onOpenChange={handleOpenChange}>
            <SheetTrigger
                render={
                    <Button disabled={!isAdmin}>
                        <CirclePlus /> {t('managedSystems.onboard.name')}
                    </Button>
                }
            />
            <SheetContent side="right">
                <form action={form.handleSubmit}>
                    <SheetHeader className="p-4">
                        <SheetTitle>{t('managedSystems.onboard.name')}</SheetTitle>
                        <SheetDescription>{t('managedSystems.onboard.description')}</SheetDescription>
                    </SheetHeader>
                    <div className="grid flex-1 auto-rows-min gap-6 px-4 mt-4">
                        <FieldSet>
                            <FieldGroup>
                                <Field>
                                    <FieldLabel htmlFor="connector-url">
                                        {t('managedSystems.onboard.connectorUrl.name')}
                                    </FieldLabel>
                                    <form.Field name="connectorUrl">
                                        {(field) => (
                                            <Input
                                                id="connector-url"
                                                type="text"
                                                value={field.state.value}
                                                onChange={(e) => field.handleChange(e.target.value)}
                                            />
                                        )}
                                    </form.Field>
                                    <FieldDescription>
                                        {t('managedSystems.onboard.connectorUrl.description')}
                                    </FieldDescription>
                                </Field>
                                <Field>
                                    <FieldLabel htmlFor="connector-hash">
                                        {t('managedSystems.onboard.connectorHash.name')}
                                    </FieldLabel>
                                    <form.Field name="connectorSha">
                                        {(field) => (
                                            <Input
                                                id="connector-hash"
                                                type="text"
                                                value={field.state.value}
                                                onChange={(e) => field.handleChange(e.target.value)}
                                            />
                                        )}
                                    </form.Field>
                                    <FieldDescription>
                                        {t('managedSystems.onboard.connectorHash.description')}
                                    </FieldDescription>
                                </Field>
                            </FieldGroup>
                        </FieldSet>
                    </div>
                    <SheetFooter>
                        <form.Subscribe
                            selector={(state) =>
                                state.values.connectorUrl.trim() !== '' && state.values.connectorSha.trim() !== ''
                            }
                        >
                            {(canSubmit) => (
                                <Button type="submit" disabled={processing || !canSubmit}>
                                    {processing && <Spinner />}
                                    {t('managedSystems.onboard.validateAndOnboard')}
                                </Button>
                            )}
                        </form.Subscribe>
                        <SheetClose render={<Button variant="outline">{t('managedSystems.onboard.cancel')}</Button>} />
                    </SheetFooter>
                </form>
            </SheetContent>
        </Sheet>
    );
};
