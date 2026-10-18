import { createFileRoute } from '@tanstack/react-router';
import { ManagedSystemDetailContainer } from '@/views/managed-system';

export const Route = createFileRoute('/managed-systems/$id')({
    component: ManagedSystemDetailContainer,
});
