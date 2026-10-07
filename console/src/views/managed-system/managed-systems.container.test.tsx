import type { ReactNode } from 'react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { render, waitFor } from '@testing-library/react';
import { delay, HttpResponse } from 'msw';
import { ManagedSystemContainer } from './managed-system.container.tsx';
import { mockHttpHandlers, mockServer } from '@/test-utils/msw.ts';

vi.mock('@/utils/openapi/client.ts');

const { dataTableProps } = vi.hoisted(() => ({ dataTableProps: vi.fn() }));

vi.mock('@/design-system/components/ui/data-table.tsx', () => ({
    DataTable: (props: unknown) => {
        dataTableProps(props);
        return null;
    },
}));

vi.mock('@/views/managed-system/onboard/onboarding-sheet.tsx', () => ({
    OnboardingSheet: () => null,
}));

const createWrapper = () => {
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    return ({ children }: { children: ReactNode }) => (
        <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
    );
};

describe('<ManagedSystemContainer />', () => {
    beforeEach(() => {
        dataTableProps.mockClear();
    });

    it('should return noData state when data is an empty list in query', async () => {
        mockServer.use(
            mockHttpHandlers.get('/api/v1/managed-systems', async () => {
                await delay(50);
                return HttpResponse.json([], { status: 200 });
            }),
        );

        render(<ManagedSystemContainer />, { wrapper: createWrapper() });

        expect(dataTableProps).toHaveBeenLastCalledWith(expect.objectContaining({ status: 'loading' }));

        await waitFor(() =>
            expect(dataTableProps).toHaveBeenLastCalledWith(
                expect.objectContaining({
                    status: 'noData',
                    emptyMessage: {
                        name: 'managedSystems.emptyManagedSystem.name',
                        description: 'managedSystems.emptyManagedSystem.description',
                    },
                }),
            ),
        );
    });

    it('should return data state when data has some managed connectors in query', async () => {
        const managedSystem = {
            id: 'id-1',
            name: 'AWS',
            connectorUrl: 'https://example.com/aws.wasm',
            connectorHash: 'sha256',
            createdAt: '2026-10-05T12:00:00Z',
        };

        mockServer.use(
            mockHttpHandlers.get('/api/v1/managed-systems', async () => {
                await delay(50);
                return HttpResponse.json([managedSystem], { status: 200 });
            }),
        );

        render(<ManagedSystemContainer />, { wrapper: createWrapper() });

        expect(dataTableProps).toHaveBeenLastCalledWith(expect.objectContaining({ status: 'loading' }));

        await waitFor(() =>
            expect(dataTableProps).toHaveBeenLastCalledWith(
                expect.objectContaining({ status: 'data', data: [managedSystem] }),
            ),
        );
    });

    it('should return error state when the response is a non-2xx status', async () => {
        mockServer.use(
            mockHttpHandlers.get('/api/v1/managed-systems', async () => {
                await delay(50);
                return HttpResponse.json({ message: 'boom' }, { status: 500 });
            }),
        );

        render(<ManagedSystemContainer />, { wrapper: createWrapper() });

        expect(dataTableProps).toHaveBeenLastCalledWith(expect.objectContaining({ status: 'loading' }));

        await waitFor(() =>
            expect(dataTableProps).toHaveBeenLastCalledWith(
                expect.objectContaining({ status: 'error', errorMessage: 'error.generic' }),
            ),
        );
    });

    it('should return error when the query fails', async () => {
        mockServer.use(
            mockHttpHandlers.get('/api/v1/managed-systems', async () => {
                await delay(50);
                return HttpResponse.error();
            }),
        );

        render(<ManagedSystemContainer />, { wrapper: createWrapper() });

        expect(dataTableProps).toHaveBeenLastCalledWith(expect.objectContaining({ status: 'loading' }));

        await waitFor(() =>
            expect(dataTableProps).toHaveBeenLastCalledWith(
                expect.objectContaining({ status: 'error', errorMessage: 'error.generic' }),
            ),
        );
    });
});
