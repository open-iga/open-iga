import { createFileRoute } from '@tanstack/react-router';
import { SignInCallbackContainer } from '@/views/sign-in-callback';

export const Route = createFileRoute('/auth/$provider/callback')({
    component: SignInCallbackContainer,
    validateSearch: (search) => ({
        code: search.code as string,
        state: search.state as string,
    }),
});
