import { useMutation } from '@tanstack/react-query';
import { useState } from 'react';
import { match } from 'ts-pattern';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import { fetchClient } from '@/utils/openapi/client.ts';
import type { ConnectorSpec } from './types.ts';

const POLL_INTERVAL_MS = 2000;
const MAX_POLLS = 90; // ~3 min at the poll interval; backend validation times out before this

export type OnboardingPhase = 'input' | 'validating' | 'review' | 'onboarding';
type ValidationResult = { onboardingId: string; spec: ConnectorSpec };

const sleep = (ms: number) => new Promise((resolve) => setTimeout(resolve, ms));

export const useManagedSystemOnboarding = () => {
    const { t } = useTranslation();
    const [result, setResult] = useState<ValidationResult | null>(null);

    // Step 1: submit the connector (url + sha), then poll until validation resolves to a spec.
    const validation = useMutation({
        mutationFn: async (input: { connectorUrl: string; connectorSha: string }): Promise<ValidationResult> => {
            const body = { ...input, connectorSha: input.connectorSha.replace(/^sha256:/, '') };
            const created = await fetchClient.POST('/api/v1/connectors/onboarding-requests', { body });
            const onboardingId = created.data?.onboardingId;
            if (!onboardingId) {
                throw new Error(t('managedSystems.onboard.validationError'));
            }

            for (let attempt = 0; attempt < MAX_POLLS; attempt++) {
                const { data } = await fetchClient.GET('/api/v1/connectors/onboarding-requests/{onboarding-id}', {
                    params: { path: { 'onboarding-id': onboardingId } },
                });

                const resolved = match(data?.status)
                    .with('success', (): ValidationResult => {
                        if (!data?.connectorSpec) {
                            throw new Error(t('managedSystems.onboard.validationError'));
                        }
                        return { onboardingId, spec: data.connectorSpec };
                    })
                    .with('failed', () => {
                        throw new Error(data?.error ?? t('managedSystems.onboard.validationError'));
                    })
                    .otherwise(() => null); // pending: keep polling

                if (resolved) {
                    return resolved;
                }
                await sleep(POLL_INTERVAL_MS);
            }

            throw new Error(t('managedSystems.onboard.validationError'));
        },
        onSuccess: setResult,
        onError: (error) => toast.error(t('managedSystems.onboard.validationError'), { description: error.message }),
    });

    // Step 2: onboard the managed system using the validated onboarding id.
    const onboarding = useMutation({
        mutationFn: async () => {
            if (!result) {
                throw new Error(t('managedSystems.onboard.onboardingError'));
            }
            const created = await fetchClient.POST('/api/v1/managed-systems', {
                body: { connectorOnboardingId: result.onboardingId },
            });
            return created.data;
        },
        onError: (error) => toast.error(t('managedSystems.onboard.onboardingError'), { description: error.message }),
    });

    const reset = () => {
        setResult(null);
        validation.reset();
        onboarding.reset();
    };

    const phase = match({
        onboarding: onboarding.isPending,
        validating: validation.isPending,
        review: result !== null,
    })
        .returnType<OnboardingPhase>()
        .with({ onboarding: true }, () => 'onboarding')
        .with({ validating: true }, () => 'validating')
        .with({ review: true }, () => 'review')
        .otherwise(() => 'input');

    return {
        phase,
        spec: result?.spec ?? null,
        error: validation.isError ? validation.error.message : undefined,
        validate: validation.mutateAsync,
        onboard: onboarding.mutateAsync,
        reset,
    };
};
