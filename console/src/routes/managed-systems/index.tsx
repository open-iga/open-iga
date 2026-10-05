import { createFileRoute } from '@tanstack/react-router';
import { ManagedSystemContainer } from '@/views/managed-system';

export const Route = createFileRoute('/managed-systems/')({
    component: ManagedSystemContainer,
});
