import { SignInContainer, supportedProviders } from '../sign-in';
import { Route } from '../../routes/auth/$provider/callback.tsx';
import { useMutation } from '@tanstack/react-query';
import { fetchClient } from '../../openapi/client.ts';
import { useEffect, useRef } from 'react';
import type { SupportedOauthProvider } from '../types';

type SignInCallbackProps = {
    state: string;
    code: string;
    provider: SupportedOauthProvider;
};

const SignInCallback = ({ state, code, provider }: SignInCallbackProps) => {
    const { mutate } = useMutation({
        mutationFn: () =>
            fetchClient.POST('/api/v1/auth/{provider}/callback', {
                params: { path: { provider }, query: { code, state } },
            }),
        onSuccess: ({ data }) => {
            if (data?.redirect) {
                globalThis.location.href = data.redirect;
            }
        },
    });

    const called = useRef(false);
    useEffect(() => {
        if (called.current) {
            return;
        }

        called.current = true;
        mutate();
    }, []);

    return <SignInContainer disableAll />;
};

export const SignInCallbackContainer = () => {
    const { provider } = Route.useParams();
    const { code, state } = Route.useSearch();

    if (!supportedProviders.includes(provider as (typeof supportedProviders)[number])) {
        return <div>Invalid provider</div>;
    }

    return <SignInCallback code={code} provider={provider as SupportedOauthProvider} state={state} />;
};
