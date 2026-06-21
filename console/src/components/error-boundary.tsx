import { Component, type ReactNode } from 'react';

interface ErrorBoundaryProps {
    children: ReactNode;
}

interface ErrorBoundaryState {
    hasError: boolean;
}

interface ErrorDetailsProps {
    summary: string;
    details: string;
}

const globalError = {
    summary: 'Runtime Error 🐞',
    details: 'You have reached the Global Error Boundary. Open console to get more details',
};

export const ErrorDetails = ({ summary, details }: ErrorDetailsProps) => (
    <div className="w-screen h-screen flex flex-col items-center justify-center p-6">
        <h1
            style={{ textDecorationSkipInk: 'none' }}
            className="scroll-m-20 text-center text-xl font-extrabold tracking-tight text-balance underline decoration-ink decoration-2 decoration-destructive decoration-wavy"
        >
            {summary}
        </h1>
        <p className="mt-2 text-muted-foreground text-center">{details}</p>
    </div>
);

export class GlobalErrorBoundary extends Component<ErrorBoundaryProps, ErrorBoundaryState> {
    constructor(props: { fallback: ReactNode; children: ReactNode }) {
        super(props);
        this.state = { hasError: false };
    }

    static getDerivedStateFromError() {
        return { hasError: true };
    }

    componentDidCatch(error: unknown) {
        console.error({ runtimeError: error });
    }

    render() {
        if (this.state.hasError) {
            return <ErrorDetails summary={globalError.summary} details={globalError.details} />;
        }

        return this.props.children;
    }
}
