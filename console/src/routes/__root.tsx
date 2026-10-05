import { createRootRoute, Outlet } from '@tanstack/react-router';
import { RootLayoutContainer } from '@/components/root-layout';
import { GlobalErrorBoundary } from '@/components/error-boundary.tsx';

export const Route = createRootRoute({
    component: () => (
        <GlobalErrorBoundary>
            <RootLayoutContainer>
                <Outlet />
            </RootLayoutContainer>
        </GlobalErrorBoundary>
    ),
});
