import { createFileRoute } from '@tanstack/react-router';
import { Logout } from '@/views/logout';

export const Route = createFileRoute('/auth/logout')({
    component: Logout,
});
