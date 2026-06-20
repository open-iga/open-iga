import { createFileRoute, useNavigate } from '@tanstack/react-router';
import { SignInCallbackContainer } from '@/views/sign-in-callback';
import { z } from 'zod';
import { toast } from 'sonner';
import { t } from 'i18next';
import { useEffect } from 'react';

const searchSchema = z.object({
    code: z.string(),
    state: z.string(),
});

const ErrorComponent = () => {
    const navigate = useNavigate();

    useEffect(() => {
        toast.error(t('auth.login.authCallbackError'));
        navigate({ to: '/auth/sign-in' }).catch(() => {});
    }, []);

    return null;
};

export const Route = createFileRoute('/auth/$provider/callback')({
    component: SignInCallbackContainer,
    validateSearch: searchSchema,
    errorComponent: ErrorComponent,
});
