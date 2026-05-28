import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { createRootRoute, Outlet } from '@tanstack/react-router';

const queryClient = new QueryClient();

const RootLayout = () => (
    <QueryClientProvider client={queryClient}>
        <Outlet />
    </QueryClientProvider>
);

export const Route = createRootRoute({ component: RootLayout });
