import { createFileRoute } from '@tanstack/react-router';
import { HomeContainer } from '@/views/home/home.container.tsx';

export const Route = createFileRoute('/')({
    component: HomeContainer,
});
