import {
    columnResizingFeature,
    columnSizingFeature,
    columnVisibilityFeature,
    tableFeatures,
    useTable,
    type ColumnDef,
    type RowData,
} from '@tanstack/react-table';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/design-system/components/ui/table.tsx';
import { match } from 'ts-pattern';
import { Skeleton } from '@/design-system/components/ui/skeleton.tsx';

// columnMeta lets a column declare `meta: { hidden: true }` inline to start hidden.
type ColumnMeta = { hidden?: boolean };

const features = tableFeatures({
    columnVisibilityFeature,
    columnSizingFeature,
    columnResizingFeature,
    columnMeta: {} as ColumnMeta,
});

export type DataTableColumn<TData extends RowData> = ColumnDef<typeof features, TData>;

export type DataTableProps<TData extends RowData> =
    | { status: 'loading'; columns: DataTableColumn<TData>[] }
    | { status: 'error'; columns: DataTableColumn<TData>[]; errorMessage: string }
    | {
          status: 'data';
          columns: DataTableColumn<TData>[];
          data: TData[];
      }
    | {
          status: 'noData';
          columns: DataTableColumn<TData>[];
          emptyMessage: {
              name: string;
              description: string;
          };
      };

// Build the initial visibility map from columns flagged `meta.hidden`.
const hiddenColumns = <TData extends RowData>(columns: DataTableColumn<TData>[]): Record<string, boolean> => {
    const visibility: Record<string, boolean> = {};
    for (const column of columns) {
        if (!column.meta?.hidden) {
            continue;
        }
        const id =
            column.id ??
            ('accessorKey' in column && typeof column.accessorKey === 'string' ? column.accessorKey : undefined);
        if (id) {
            visibility[id] = false;
        }
    }
    return visibility;
};

export const DataTable = <TData extends RowData>(props: DataTableProps<TData>) => {
    const { columns } = props;
    const data = 'data' in props ? props.data : [];

    const table = useTable({
        features,
        columns,
        data,
        initialState: { columnVisibility: hiddenColumns(columns) },
        columnResizeMode: 'onChange',
    });

    return (
        <div className="overflow-hidden rounded-md border">
            <Table className="table-auto min-w-full">
                <TableHeader>
                    {table.getHeaderGroups().map((headerGroup) => (
                        <TableRow key={headerGroup.id}>
                            {headerGroup.headers.map((header) => (
                                <TableHead
                                    key={header.id}
                                    className="relative overflow-hidden"
                                    style={{ width: header.getSize() }}
                                >
                                    {header.isPlaceholder ? null : <table.FlexRender header={header} />}
                                    {header.column.getCanResize() && (
                                        <button
                                            type="button"
                                            aria-label="Resize column"
                                            onMouseDown={header.getResizeHandler()}
                                            onTouchStart={header.getResizeHandler()}
                                            className="absolute top-0 right-0 h-full w-1 cursor-col-resize touch-none select-none border-0 bg-transparent p-0 hover:bg-primary/30"
                                        />
                                    )}
                                </TableHead>
                            ))}
                        </TableRow>
                    ))}
                </TableHeader>
                <TableBody>
                    {match(props)
                        .with({ status: 'loading' }, () => (
                            <TableRow>
                                {table
                                    .getHeaderGroups()
                                    .at(0)
                                    ?.headers.map((header) => (
                                        <TableCell key={`${header.id}-skeleton`} className="overflow-hidden">
                                            <Skeleton className="h-4 w-20" />
                                        </TableCell>
                                    ))}
                            </TableRow>
                        ))
                        .with({ status: 'data' }, () =>
                            table.getRowModel().rows.map((row) => (
                                <TableRow key={row.id}>
                                    {row.getVisibleCells().map((cell) => (
                                        <TableCell key={cell.id} className="overflow-hidden">
                                            <table.FlexRender cell={cell} />
                                        </TableCell>
                                    ))}
                                </TableRow>
                            )),
                        )
                        .with({ status: 'noData' }, ({ emptyMessage }) => (
                            <TableRow>
                                <TableCell colSpan={columns.length}>
                                    <div className="flex h-32 flex-col items-center justify-center gap-2 text-center text-wrap">
                                        <p className="font-medium">{emptyMessage.name}</p>
                                        <p className="max-w-md text-sm text-muted-foreground">
                                            {emptyMessage.description}
                                        </p>
                                    </div>
                                </TableCell>
                            </TableRow>
                        ))
                        .with({ status: 'error' }, ({ errorMessage }) => (
                            <TableRow>
                                <TableCell colSpan={columns.length}>
                                    <div className="flex h-32 flex-col items-center justify-center gap-2 text-center text-wrap">
                                        <p className="font-medium">{errorMessage}</p>
                                    </div>
                                </TableCell>
                            </TableRow>
                        ))
                        .exhaustive()}
                </TableBody>
            </Table>
        </div>
    );
};
