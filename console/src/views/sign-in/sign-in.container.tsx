import { useSignIn } from './hooks/use-sign-in.ts';
import { SignIn } from '@/components/sign-in.tsx';

export const SignInContainer = () => {
    const { mutate, isPending } = useSignIn();

    if (isPending) {
        return <SignIn disableProviderSelection={true} />;
    }

    return <SignIn disableProviderSelection={false} onProviderSelection={mutate} />;
};
