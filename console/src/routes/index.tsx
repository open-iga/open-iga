import { createFileRoute } from '@tanstack/react-router';

const Index = () => {
    return <h3>Home!</h3>;
};

export const Route = createFileRoute('/')({
    component: Index,
});
