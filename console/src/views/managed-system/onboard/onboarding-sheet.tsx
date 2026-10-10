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
import { Check, CirclePlus, ShieldAlert } from 'lucide-react';
import { Field, FieldDescription, FieldGroup, FieldLabel, FieldSet } from '@/design-system/components/ui/field.tsx';
import { Input } from '@/design-system/components/ui/input.tsx';
import { Spinner } from '@/design-system/components/ui/spinner.tsx';
import { Separator } from '@/design-system/components/ui/separator.tsx';
import { useForm } from '@tanstack/react-form';
import { useQueryClient } from '@tanstack/react-query';
import { useState } from 'react';
import { match } from 'ts-pattern';
import { toast } from 'sonner';
import { cn } from '@/design-system/lib/utils.ts';
import { usePermission } from '@/hooks/use-permission.ts';
import { useManagedSystemOnboarding, type OnboardingPhase } from './use-managed-system-onboarding.ts';
import { CapabilityReview } from './capability-review.tsx';
import type { ConnectorSpec } from './types.ts';

const STEPS = ['validate', 'review', 'onboard'] as const;
const stepIndex: Record<OnboardingPhase, number> = { input: 0, validating: 0, review: 1, onboarding: 2 };

const Stepper = ({ phase }: { phase: OnboardingPhase }) => {
    const { t } = useTranslation();
    const current = stepIndex[phase];
    return (
        <ol className="mt-3 flex items-center gap-2 text-xs">
            {STEPS.map((step, i) => (
                <li key={step} className="flex items-center gap-2">
                    <span
                        className={cn(
                            'flex items-center gap-1.5',
                            i <= current ? 'text-foreground' : 'text-muted-foreground',
                        )}
                    >
                        <span
                            className={cn(
                                'flex size-4 items-center justify-center rounded-full border text-[10px] tabular-nums',
                                i < current && 'border-primary bg-primary text-primary-foreground',
                                i === current && 'border-ring',
                            )}
                        >
                            {i < current ? <Check className="size-2.5" /> : i + 1}
                        </span>
                        <span className="hidden sm:inline">{t(`managedSystems.onboard.steps.${step}`)}</span>
                    </span>
                    {i < STEPS.length - 1 && <span className="h-px w-4 bg-border" aria-hidden />}
                </li>
            ))}
        </ol>
    );
};

const ReviewPanel = ({ spec }: { spec: ConnectorSpec }) => {
    const { t } = useTranslation();
    return (
        <div className="flex flex-col gap-5 px-4 mt-4">
            <div className="flex items-start gap-3 rounded-lg border border-amber-500/30 bg-amber-500/5 p-3">
                <ShieldAlert className="mt-0.5 size-4 shrink-0 text-amber-500" aria-hidden />
                <div className="flex flex-col gap-1 text-sm">
                    <p className="font-medium">{t('managedSystems.onboard.review.heading', { name: spec.name })}</p>
                    <p className="text-xs leading-relaxed text-muted-foreground">
                        {t('managedSystems.onboard.review.description')}
                    </p>
                </div>
            </div>
            <div className="flex flex-col gap-1">
                <h3 className="text-lg font-semibold">{spec.name}</h3>
                <p className="text-sm text-muted-foreground">{spec.description}</p>
            </div>
            <CapabilityReview spec={spec} />
        </div>
    );
};

export const OnboardingSheet = () => {
    const { t } = useTranslation();
    const [open, setOpen] = useState(false);
    const { isAdmin } = usePermission();
    const queryClient = useQueryClient();
    const { phase, spec, error, validate, onboard, reset } = useManagedSystemOnboarding();

    const busy = phase === 'validating' || phase === 'onboarding';

    const form = useForm({
        defaultValues: { connectorUrl: '', connectorSha: '' },
        onSubmit: ({ value }) => validate(value).catch(() => {}), // error surfaced by the hook's toast
    });

    const close = () => {
        form.reset();
        reset();
        setOpen(false);
    };

    const handleOpenChange = (nextOpen: boolean) => {
        if (busy) return;
        if (!nextOpen) close();
        else setOpen(nextOpen);
    };

    const handleOnboard = async () => {
        try {
            await onboard();
            await queryClient.invalidateQueries({ queryKey: ['managed-systems'] });
            toast.success(t('managedSystems.onboard.onboardedTitle'), {
                description: t('managedSystems.onboard.onboardedDescription', { name: spec?.name }),
            });
            close();
        } catch {
            // error surfaced by the hook's toast; keep the sheet open to retry
        }
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
            <SheetContent side="right" className="min-w-1/2 overflow-y-auto">
                <SheetHeader className="p-4">
                    <SheetTitle>{t('managedSystems.onboard.name')}</SheetTitle>
                    <SheetDescription>{t('managedSystems.onboard.description')}</SheetDescription>
                    <Stepper phase={phase} />
                </SheetHeader>
                <Separator />

                {match(phase)
                    .with('input', 'validating', () => (
                        <form
                            id="onboard-form"
                            onSubmit={(e) => {
                                e.preventDefault();
                                form.handleSubmit();
                            }}
                            className="grid flex-1 auto-rows-min gap-6 px-4 mt-4"
                        >
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
                                                    disabled={busy}
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
                                                    disabled={busy}
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
                            {error && <p className="text-sm text-red-600">{error}</p>}
                        </form>
                    ))
                    .with('review', 'onboarding', () => spec && <ReviewPanel spec={spec} />)
                    .exhaustive()}

                <SheetFooter className="sticky bottom-0 flex-row justify-end border-t bg-popover">
                    {match(phase)
                        .with('review', 'onboarding', () => (
                            <>
                                <Button variant="outline" disabled={busy} onClick={reset}>
                                    {t('managedSystems.onboard.reject')}
                                </Button>
                                <Button disabled={busy} onClick={handleOnboard}>
                                    {phase === 'onboarding' && <Spinner />}
                                    {t('managedSystems.onboard.approveAndOnboard')}
                                </Button>
                            </>
                        ))
                        .with('input', 'validating', () => (
                            <>
                                <SheetClose
                                    render={
                                        <Button variant="outline" disabled={busy}>
                                            {t('managedSystems.onboard.cancel')}
                                        </Button>
                                    }
                                />
                                <form.Subscribe
                                    selector={(s) =>
                                        s.values.connectorUrl.trim() !== '' && s.values.connectorSha.trim() !== ''
                                    }
                                >
                                    {(canSubmit) => (
                                        <Button type="submit" form="onboard-form" disabled={busy || !canSubmit}>
                                            {phase === 'validating' && <Spinner />}
                                            {t('managedSystems.onboard.validate')}
                                        </Button>
                                    )}
                                </form.Subscribe>
                            </>
                        ))
                        .exhaustive()}
                </SheetFooter>
            </SheetContent>
        </Sheet>
    );
};
