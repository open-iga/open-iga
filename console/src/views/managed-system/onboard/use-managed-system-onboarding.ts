import { useMutation } from '@tanstack/react-query';
import { fetchClient } from '@/utils/openapi/client.ts';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';

const POLL_INTERVAL_MS = 2000;
const MAX_POLLS = 90; // ~3 min at the poll interval; backend validation times out before this

export type ValidationStatus = 'pending' | 'success' | 'failed';
export type OnboardingState = 'idle' | 'processing' | 'success' | 'error';

const sleep = (ms: number) => new Promise((resolve) => setTimeout(resolve, ms));

export const useManagedSystemOnboarding = () => {
    const { t } = useTranslation();

    const onboarding = useMutation({
        mutationFn: async (body: { connectorUrl: string; connectorSha: string }) => {
            // 1. submit the connector and get an onboarding id
            const created = await fetchClient.POST('/api/v1/connectors/onboarding-requests', { body });
            const onboardingId = created.data?.onboardingId;
            if (!onboardingId) {
                throw new Error(t('managedSystems.onboard.validationError'));
            }

            // 2. poll until validation resolves
            for (let attempt = 0; attempt < MAX_POLLS; attempt++) {
                const result = await fetchClient.GET('/api/v1/connectors/onboarding-requests/{onboarding-id}', {
                    params: { path: { 'onboarding-id': onboardingId } },
                });
                const status = result.data?.status;

                if (status === 'failed') {
                    throw new Error(result.data?.error ?? t('managedSystems.onboard.validationError'));
                }

                if (status === 'success') {
                    // 3. create the managed system (throws on failure -> onError)
                    return fetchClient.POST('/api/v1/managed-systems', {
                        body: { connectorOnboardingId: onboardingId },
                    });
                }

                await sleep(POLL_INTERVAL_MS);
            }

            throw new Error(t('managedSystems.onboard.onboardingError'));
        },
        onError: (error) => toast.error(t('managedSystems.onboard.onboardingError'), { description: error.message }),
    });

    const resolveState = (): OnboardingState => {
        if (onboarding.isPending) return 'processing';
        if (onboarding.isSuccess) return 'success';
        if (onboarding.isError) return 'error';
        return 'idle';
    };

    // mutateAsync so callers can await success (it rejects on error; onError still toasts).
    return { error: onboarding.error?.message, state: resolveState(), submit: onboarding.mutateAsync };
};
