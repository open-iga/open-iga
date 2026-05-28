import { fetchClient } from '../../openapi/client.ts';
import { useMutation } from '@tanstack/react-query';
import type { SupportedOauthProvider } from '../types';

export const supportedProviders: SupportedOauthProvider[] = ['google'];

type SignInContainerProps = {
    disableAll?: boolean;
};

export const SignInContainer = ({ disableAll = false }: SignInContainerProps) => {
    const { mutate, isPending } = useMutation({
        mutationFn: (provider: SupportedOauthProvider) =>
            fetchClient.GET('/api/v1/auth/{provider}', {
                params: { path: { provider } },
            }),
        onSuccess: ({ data }) => {
            if (data?.authCodeUrl) {
                globalThis.location.href = data.authCodeUrl;
            }
        },
        onError: (error) => {
            console.error(error);
        },
    });

    const disabled = isPending || disableAll;
    return supportedProviders.map((provider) => (
        <button disabled={disabled} key={provider} onClick={() => mutate(provider)}>
            {provider}
        </button>
    ));
};
