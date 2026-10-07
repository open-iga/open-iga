import type { ReactNode } from 'react';
import { mockServer } from './src/test-utils/msw';

vi.mock('sonner', () => ({
    toast: {
        error: vi.fn(),
    },
}));

vi.mock('react-i18next', () => ({
    useTranslation: () => ({
        t: (key: string) => key,
    }),
}));

vi.mock('@/components/page-wrapper.tsx', () => ({
    PageWrapper: ({ children }: { children: ReactNode }) => children,
}));

beforeAll(() => mockServer.listen());
afterEach(() => mockServer.resetHandlers());
afterAll(() => mockServer.close());
