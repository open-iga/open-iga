import { createFileRoute } from '@tanstack/react-router';
import { SignInContainer } from '../../views/sign-in';

export const Route = createFileRoute('/auth/sign-in')({
    component: SignInContainer,
});
