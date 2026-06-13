import type { router } from '../main.tsx';

declare module '@tanstack/react-router' {
    interface Register {
        router: typeof router;
    }
}
