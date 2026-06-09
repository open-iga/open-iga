import { Route } from '../../routes/auth/$provider/callback.tsx';
import type { SupportedOauthProvider } from '../types';
import { isSupportedOauthProvider } from '@/views/sign-in/sign-in.utils.ts';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import { useNavigate } from '@tanstack/react-router';
import { useSignInCallback } from '@/views/sign-in-callback/hooks/use-sign-in-callback.ts';
import { SignIn } from '@/components/sign-in.tsx';
import { useEffect } from 'react';

type SignInCallbackProps = {
    state: string;
    code: string;
    provider: SupportedOauthProvider;
};

const SignInCallback = ({ state, code, provider }: SignInCallbackProps) => {
    const { mutate } = useSignInCallback({ state, code, provider });

    useEffect(() => {
        mutate();
    }, []);

    return <SignIn disableProviderSelection={true} providerToHighlight={provider} />;
};

export const SignInCallbackContainer = () => {
    const { provider } = Route.useParams();
    const { code, state } = Route.useSearch();
    const { t } = useTranslation();
    const navigate = useNavigate();

    if (!isSupportedOauthProvider(provider)) {
        toast.error(t('auth.provider.invalidProvider'));
        return navigate({ to: '/auth/sign-in' });
    }

    return <SignInCallback code={code} provider={provider} state={state} />;
};
