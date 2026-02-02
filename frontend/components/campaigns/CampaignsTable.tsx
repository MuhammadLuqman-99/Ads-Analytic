"use client";

import { useState, useMemo } from "react";
import {
  useReactTable,
  getCoreRowModel,
  getSortedRowModel,
  getPaginationRowModel,
  getExpandedRowModel,
  flexRender,
  type ColumnDef,
  type SortingState,
  type ExpandedState,
  type RowSelectionState,
} from "@tanstack/react-table";
import {
  ChevronDown,
  ChevronRight,
  ChevronUp,
  ChevronsUpDown,
  Download,
  MoreHorizontal,
  Eye,
  Pause,
  Play,
  Trash2,
} from "lucide-react";
import { format, parseISO } from "date-fns";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { Badge } from "@/components/ui/badge";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
  DropdownMenuSeparator,
} from "@/components/ui/dropdown-menu";
import {
  type Campaign,
  type Platform,
  type CampaignStatus,
  formatCurrency,
  formatNumber,
  getPlatformName,
} from "@/lib/mock-data";
import { cn } from "@/lib/utils";

// Column visibility state type
export type ColumnVisibility = {
  platform: boolean;
  name: boolean;
  status: boolean;
  spend: boolean;
  impressions: boolean;
  clicks: boolean;
  ctr: boolean;
  conversions: boolean;
  roas: boolean;
};

const defaultColumnVisibility: ColumnVisibility = {
  platform: true,
  name: true,
  status: true,
  spend: true,
  impressions: true,
  clicks: true,
  ctr: true,
  conversions: true,
  roas: true,
};

interface CampaignsTableProps {
  data: Campaign[];
  isLoading?: boolean;
  columnVisibility: ColumnVisibility;
  onColumnVisibilityChange: (visibility: ColumnVisibility) => void;
  onExportCSV: (campaigns: Campaign[]) => void;
}

const platformStyles: Record<Platform, { bg: string; icon: string }> = {
  meta: { bg: "bg-blue-600", icon: "M" },
  tiktok: { bg: "bg-black", icon: "T" },
  shopee: { bg: "bg-orange-500", icon: "S" },
};

const statusStyles: Record<CampaignStatus, { bg: string; text: string }> = {
  active: { bg: "bg-emerald-100", text: "text-emerald-700" },
  paused: { bg: "bg-amber-100", text: "text-amber-700" },
  completed: { bg: "bg-slate-100", text: "text-slate-600" },
  draft: { bg: "bg-slate-100", text: "text-slate-500" },
};

export function CampaignsTable({
  data,
  isLoading,
  columnVisibility,
  onColumnVisibilityChange,
  onExportCSV,
}: CampaignsTableProps) {
  const [sorting, setSorting] = useState<SortingState>([]);
  const [expanded, setExpanded] = useState<ExpandedState>({});
  const [rowSelection, setRowSelection] = useState<RowSelectionState>({});

  const columns = useMemo<ColumnDef<Campaign>[]>(
    () => [
      // Selection column
      {
        id: "select",
        header: ({ table }) => (
          <Checkbox
            checked={
              table.getIsAllPageRowsSelected() ||
              (table.getIsSomePageRowsSelected() && "indeterminate")
            }
            onCheckedChange={(value) => table.toggleAllPageRowsSelected(!!value)}
            aria-label="Select all"
          />
        ),
        cell: ({ row }) => (
          <Checkbox
            checked={row.getIsSelected()}
            onCheckedChange={(value) => row.toggleSelected(!!value)}
            aria-label="Select row"
          />
        ),
        enableSorting: false,
        size: 40,
      },
      // Expand column
      {
        id: "expand",
        header: () => null,
        cell: ({ row }) => (
          <button
            onClick={() => row.toggleExpanded()}
            className="p-1 hover:bg-slate-100 rounded"
          >
            {row.getIsExpanded() ? (
              <ChevronDown className="h-4 w-4 text-slate-500" />
            ) : (
              <ChevronRight className="h-4 w-4 text-slate-500" />
            )}
          </button>
        ),
        size: 40,
      },
      // Platform
      {
        accessorKey: "platform",
        header: "Platform",
        cell: ({ row }) => {
          const platform = row.getValue("platform") as Platform;
          const style = platformStyles[platform];
          return (
            <div
              className={cn(
                "h-8 w-8 rounded-lg flex items-center justify-center text-white font-bold text-sm",
                style.bg
              )}
            >
              {style.icon}
            </div>
          );
        },
        size: 80,
      },
      // Name
      {
        accessorKey: "name",
        header: ({ column }) => (
          <button
            onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
            className="flex items-center gap-1 hover:text-slate-900"
          >
            Campaign Name
            {column.getIsSorted() === "asc" ? (
              <ChevronUp className="h-4 w-4" />
            ) : column.getIsSorted() === "desc" ? (
              <ChevronDown className="h-4 w-4" />
            ) : (
              <ChevronsUpDown className="h-4 w-4 opacity-50" />
            )}
          </button>
        ),
        cell: ({ row }) => (
          <div className="font-medium text-slate-900">{row.getValue("name")}</div>
        ),
        size: 250,
      },
      // Status
      {
        accessorKey: "status",
        header: "Status",
        cell: ({ row }) => {
          const status = row.getValue("status") as CampaignStatus;
          const style = statusStyles[status];
          return (
            <Badge className={cn("border-0 capitalize", style.bg, style.text)}>
              {status}
            </Badge>
          );
        },
        size: 100,
      },
      // Spend
      {
        accessorKey: "spend",
        header: ({ column }) => (
          <button
            onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
            className="flex items-center gap-1 hover:text-slate-900"
          >
            Spend
            {column.getIsSorted() === "asc" ? (
              <ChevronUp className="h-4 w-4" />
            ) : column.getIsSorted() === "desc" ? (
              <ChevronDown className="h-4 w-4" />
            ) : (
              <ChevronsUpDown className="h-4 w-4 opacity-50" />
            )}
          </button>
        ),
        cell: ({ row }) => (
          <div className="text-right font-medium">
            {formatCurrency(row.getValue("spend"))}
          </div>
        ),
        size: 120,
      },
      // Impressions
      {
        accessorKey: "impressions",
        header: ({ column }) => (
          <button
            onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
            className="flex items-center gap-1 hover:text-slate-900"
          >
            Impressions
            {column.getIsSorted() === "asc" ? (
              <ChevronUp className="h-4 w-4" />
            ) : column.getIsSorted() === "desc" ? (
              <ChevronDown className="h-4 w-4" />
            ) : (
              <ChevronsUpDown className="h-4 w-4 opacity-50" />
            )}
          </button>
        ),
        cell: ({ row }) => (
          <div className="text-right">{formatNumber(row.getValue("impressions"))}</div>
        ),
        size: 120,
      },
      // Clicks
      {
        accessorKey: "clicks",
        header: ({ column }) => (
          <button
            onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
            className="flex items-center gap-1 hover:text-slate-900"
          >
            Clicks
            {column.getIsSorted() === "asc" ? (
              <ChevronUp className="h-4 w-4" />
            ) : column.getIsSorted() === "desc" ? (
              <ChevronDown className="h-4 w-4" />
            ) : (
              <ChevronsUpDown className="h-4 w-4 opacity-50" />
            )}
          </button>
        ),
        cell: ({ row }) => (
          <div className="text-right">{formatNumber(row.getValue("clicks"))}</div>
        ),
        size: 100,
      },
      // CTR (calculated)
      {
        id: "ctr",
        header: "CTR",
        cell: ({ row }) => {
          const clicks = row.getValue("clicks") as number;
          const impressions = row.getValue("impressions") as number;
          const ctr = impressions > 0 ? (clicks / impressions) * 100 : 0;
          return <div className="text-right">{ctr.toFixed(2)}%</div>;
        },
        size: 80,
      },
      // Conversions
      {
        accessorKey: "conversions",
        header: ({ column }) => (
          <button
            onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
            className="flex items-center gap-1 hover:text-slate-900"
          >
            Conv.
            {column.getIsSorted() === "asc" ? (
              <ChevronUp className="h-4 w-4" />
            ) : column.getIsSorted() === "desc" ? (
              <ChevronDown className="h-4 w-4" />
            ) : (
              <ChevronsUpDown className="h-4 w-4 opacity-50" />
            )}
          </button>
        ),
        cell: ({ row }) => (
          <div className="text-right">{formatNumber(row.getValue("conversions"))}</div>
        ),
        size: 100,
      },
      // ROAS
      {
        accessorKey: "roas",
        header: ({ column }) => (
          <button
            onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
            className="flex items-center gap-1 hover:text-slate-900"
          >
            ROAS
            {column.getIsSorted() === "asc" ? (
              <ChevronUp className="h-4 w-4" />
            ) : column.getIsSorted() === "desc" ? (
              <ChevronDown className="h-4 w-4" />
            ) : (
              <ChevronsUpDown className="h-4 w-4 opacity-50" />
            )}
          </button>
        ),
        cell: ({ row }) => {
          const roas = row.getValue("roas") as number;
          return (
            <div
              className={cn(
                "text-right font-semibold",
                roas >= 3
                  ? "text-emerald-600"
                  : roas >= 2
                  ? "text-amber-600"
                  : "text-rose-600"
              )}
            >
              {roas.toFixed(1)}x
            </div>
          );
        },
        size: 80,
      },
      // Actions
      {
        id: "actions",
        cell: ({ row }) => {
          const campaign = row.original;
          return (
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="ghost" size="sm">
                  <MoreHorizontal className="h-4 w-4" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                <DropdownMenuItem>
                  <Eye className="h-4 w-4 mr-2" />
                  View Details
                </DropdownMenuItem>
                {campaign.status === "active" ? (
                  <DropdownMenuItem>
                    <Pause className="h-4 w-4 mr-2" />
                    Pause Campaign
                  </DropdownMenuItem>
                ) : campaign.status === "paused" ? (
                  <DropdownMenuItem>
                    <Play className="h-4 w-4 mr-2" />
                    Resume Campaign
                  </DropdownMenuItem>
                ) : null}
                <DropdownMenuSeparator />
                <DropdownMenuItem className="text-rose-600">
                  <Trash2 className="h-4 w-4 mr-2" />
                  Delete
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          );
        },
        size: 50,
      },
    ],
    []
  );

  // Filter columns based on visibility
  const visibleColumns = useMemo(() => {
    return columns.filter((col) => {
      const id = col.id || (col as any).accessorKey;
      if (id === "select" || id === "expand" || id === "actions") return true;
      return columnVisibility[id as keyof ColumnVisibility] ?? true;
    });
  }, [columns, columnVisibility]);

  const table = useReactTable({
    data,
    columns: visibleColumns,
    state: {
      sorting,
      expanded,
      rowSelection,
    },
    onSortingChange: setSorting,
    onExpandedChange: setExpanded,
    onRowSelectionChange: setRowSelection,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    getExpandedRowModel: getExpandedRowModel(),
    initialState: {
      pagination: {
        pageSize: 10,
      },
    },
  });

  const selectedCampaigns = table
    .getFilteredSelectedRowModel()
    .rows.map((row) => row.original);

  if (isLoading) {
    return <CampaignsTableSkeleton />;
  }

  return (
    <div className="space-y-4">
      {/* Bulk Actions */}
      {selectedCampaigns.length > 0 && (
        <div className="flex items-center gap-4 p-3 bg-blue-50 rounded-lg">
          <span className="text-sm font-medium text-blue-900">
            {selectedCampaigns.length} campaign(s) selected
          </span>
          <Button
            size="sm"
            variant="outline"
            onClick={() => onExportCSV(selectedCampaigns)}
          >
            <Download className="h-4 w-4 mr-2" />
            Export CSV
          </Button>
          <Button
            size="sm"
            variant="ghost"
            onClick={() => setRowSelection({})}
          >
            Clear selection
          </Button>
        </div>
      )}

      {/* Table */}
      <div className="border border-slate-200 rounded-lg overflow-hidden">
        <Table>
          <TableHeader>
            {table.getHeaderGroups().map((headerGroup) => (
              <TableRow key={headerGroup.id} className="bg-slate-50 hover:bg-slate-50">
                {headerGroup.headers.map((header) => (
                  <TableHead
                    key={header.id}
                    className="text-slate-500 font-medium"
                    style={{ width: header.getSize() }}
                  >
                    {header.isPlaceholder
                      ? null
                      : flexRender(
                          header.column.columnDef.header,
                          header.getContext()
                        )}
                  </TableHead>
                ))}
              </TableRow>
            ))}
          </TableHeader>
          <TableBody>
            {table.getRowModel().rows.length === 0 ? (
              <TableRow>
                <TableCell
                  colSpan={visibleColumns.length}
                  className="h-32 text-center text-slate-500"
                >
                  No campaigns found
                </TableCell>
              </TableRow>
            ) : (
              table.getRowModel().rows.map((row) => (
                <>
                  <TableRow
                    key={row.id}
                    data-state={row.getIsSelected() && "selected"}
                    className={cn(
                      "hover:bg-slate-50",
                      row.getIsSelected() && "bg-blue-50"
                    )}
                  >
                    {row.getVisibleCells().map((cell) => (
                      <TableCell key={cell.id}>
                        {flexRender(
                          cell.column.columnDef.cell,
                          cell.getContext()
                        )}
                      </TableCell>
                    ))}
                  </TableRow>
                  {/* Expanded Row */}
                  {row.getIsExpanded() && (
                    <TableRow className="bg-slate-50">
                      <TableCell colSpan={visibleColumns.length} className="p-4">
                        <CampaignExpandedDetails campaign={row.original} />
                      </TableCell>
                    </TableRow>
                  )}
                </>
              ))
            )}
          </TableBody>
        </Table>
      </div>

      {/* Pagination */}
      <div className="flex items-center justify-between">
        <div className="text-sm text-slate-500">
          Showing {table.getState().pagination.pageIndex * table.getState().pagination.pageSize + 1} to{" "}
          {Math.min(
            (table.getState().pagination.pageIndex + 1) * table.getState().pagination.pageSize,
            data.length
          )}{" "}
          of {data.length} campaigns
        </div>
        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => table.previousPage()}
            disabled={!table.getCanPreviousPage()}
          >
            Previous
          </Button>
          <div className="flex items-center gap-1">
            {Array.from({ length: table.getPageCount() }, (_, i) => (
              <Button
                key={i}
                variant={table.getState().pagination.pageIndex === i ? "default" : "outline"}
                size="sm"
                onClick={() => table.setPageIndex(i)}
                className="w-8"
              >
                {i + 1}
              </Button>
            ))}
          </div>
          <Button
            variant="outline"
            size="sm"
            onClick={() => table.nextPage()}
            disabled={!table.getCanNextPage()}
          >
            Next
          </Button>
        </div>
      </div>
    </div>
  );
}

function CampaignExpandedDetails({ campaign }: { campaign: Campaign }) {
  return (
    <div className="grid grid-cols-2 md:grid-cols-4 gap-6">
      <div>
        <p className="text-xs text-slate-500 mb-1">Campaign ID</p>
        <p className="font-mono text-sm">{campaign.id}</p>
      </div>
      <div>
        <p className="text-xs text-slate-500 mb-1">Account ID</p>
        <p className="font-mono text-sm">{campaign.accountId}</p>
      </div>
      <div>
        <p className="text-xs text-slate-500 mb-1">Start Date</p>
        <p className="text-sm">{format(parseISO(campaign.startDate), "MMM d, yyyy")}</p>
      </div>
      <div>
        <p className="text-xs text-slate-500 mb-1">End Date</p>
        <p className="text-sm">{format(parseISO(campaign.endDate), "MMM d, yyyy")}</p>
      </div>
      <div>
        <p className="text-xs text-slate-500 mb-1">Cost per Click</p>
        <p className="text-sm font-medium">
          {campaign.clicks > 0
            ? formatCurrency(campaign.spend / campaign.clicks)
            : "N/A"}
        </p>
      </div>
      <div>
        <p className="text-xs text-slate-500 mb-1">Cost per Conversion</p>
        <p className="text-sm font-medium">
          {campaign.conversions > 0
            ? formatCurrency(campaign.spend / campaign.conversions)
            : "N/A"}
        </p>
      </div>
      <div>
        <p className="text-xs text-slate-500 mb-1">Conversion Rate</p>
        <p className="text-sm font-medium">
          {campaign.clicks > 0
            ? `${((campaign.conversions / campaign.clicks) * 100).toFixed(2)}%`
            : "N/A"}
        </p>
      </div>
      <div>
        <p className="text-xs text-slate-500 mb-1">Platform</p>
        <p className="text-sm">{getPlatformName(campaign.platform)}</p>
      </div>
    </div>
  );
}

function CampaignsTableSkeleton() {
  return (
    <div className="border border-slate-200 rounded-lg overflow-hidden">
      <Table>
        <TableHeader>
          <TableRow className="bg-slate-50">
            {[1, 2, 3, 4, 5, 6, 7, 8].map((i) => (
              <TableHead key={i}>
                <div className="h-4 w-20 bg-slate-200 rounded animate-pulse" />
              </TableHead>
            ))}
          </TableRow>
        </TableHeader>
        <TableBody>
          {[1, 2, 3, 4, 5].map((row) => (
            <TableRow key={row}>
              {[1, 2, 3, 4, 5, 6, 7, 8].map((cell) => (
                <TableCell key={cell}>
                  <div className="h-4 w-16 bg-slate-200 rounded animate-pulse" />
                </TableCell>
              ))}
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  );
}

export { defaultColumnVisibility };
