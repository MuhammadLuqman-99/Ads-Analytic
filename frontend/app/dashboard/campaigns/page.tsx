"use client";

import { useState, useMemo } from "react";
import { Download, RefreshCw } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import {
  CampaignFilters,
  CampaignsTable,
  ColumnCustomization,
  EmptyState,
  loadColumnVisibility,
  type CampaignFiltersState,
  type ColumnVisibility,
} from "@/components/campaigns";
import { useCampaigns } from "@/hooks/use-metrics";
import { useDateRangeStore } from "@/stores/app-store";
import { type Campaign, type CampaignStatus } from "@/lib/mock-data";

const defaultFilters: CampaignFiltersState = {
  search: "",
  platforms: [],
  statuses: [],
  dateRange: null,
  useGlobalDateRange: true,
};

export default function CampaignsPage() {
  const { dateRange: globalDateRange } = useDateRangeStore();
  const [filters, setFilters] = useState<CampaignFiltersState>(defaultFilters);
  const [columnVisibility, setColumnVisibility] = useState<ColumnVisibility>(() =>
    loadColumnVisibility()
  );

  // Build query filters
  const queryFilters = useMemo(() => {
    const activeDateRange = filters.useGlobalDateRange
      ? globalDateRange
      : filters.dateRange;

    return {
      platforms: filters.platforms.length > 0 ? filters.platforms : undefined,
      status: filters.statuses.length === 1 ? filters.statuses[0] : undefined,
      search: filters.search || undefined,
      dateRange: activeDateRange || undefined,
    };
  }, [filters, globalDateRange]);

  const { data: campaigns, isLoading, refetch, isRefetching } = useCampaigns(queryFilters);

  // Filter campaigns locally for multiple statuses
  const filteredCampaigns = useMemo(() => {
    if (!campaigns) return [];

    let result = [...campaigns];

    // Apply multi-status filter (useCampaigns only supports single status)
    if (filters.statuses.length > 1) {
      result = result.filter((c) =>
        filters.statuses.includes(c.status as CampaignStatus)
      );
    }

    return result;
  }, [campaigns, filters.statuses]);

  // Check if we have any connected accounts (mock check)
  const hasConnectedAccounts = true; // In real app, check from a hook

  // Export to CSV function
  const handleExportCSV = (campaignsToExport: Campaign[]) => {
    const headers = [
      "ID",
      "Name",
      "Platform",
      "Status",
      "Spend",
      "Impressions",
      "Clicks",
      "CTR",
      "Conversions",
      "ROAS",
      "Start Date",
      "End Date",
    ];

    const rows = campaignsToExport.map((c) => [
      c.id,
      c.name,
      c.platform,
      c.status,
      c.spend.toString(),
      c.impressions.toString(),
      c.clicks.toString(),
      ((c.clicks / c.impressions) * 100).toFixed(2) + "%",
      c.conversions.toString(),
      c.roas.toFixed(2) + "x",
      c.startDate,
      c.endDate,
    ]);

    const csvContent = [
      headers.join(","),
      ...rows.map((row) =>
        row.map((cell) => `"${cell.replace(/"/g, '""')}"`).join(",")
      ),
    ].join("\n");

    const blob = new Blob([csvContent], { type: "text/csv;charset=utf-8;" });
    const link = document.createElement("a");
    const url = URL.createObjectURL(blob);
    link.setAttribute("href", url);
    link.setAttribute(
      "download",
      `campaigns-export-${new Date().toISOString().split("T")[0]}.csv`
    );
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
  };

  const clearFilters = () => {
    setFilters(defaultFilters);
  };

  // Determine which empty state to show
  const getEmptyStateType = () => {
    if (!hasConnectedAccounts) return "no-accounts";
    if (
      filteredCampaigns.length === 0 &&
      (filters.search ||
        filters.platforms.length > 0 ||
        filters.statuses.length > 0)
    ) {
      return "no-results";
    }
    return "no-campaigns";
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
        <div>
          <h1 className="text-3xl font-bold text-slate-900">Campaigns</h1>
          <p className="mt-1 text-slate-500">
            Manage and monitor all your advertising campaigns
          </p>
        </div>
        <div className="flex items-center gap-3">
          <Button
            variant="outline"
            onClick={() => refetch()}
            disabled={isRefetching}
          >
            <RefreshCw
              className={`h-4 w-4 mr-2 ${isRefetching ? "animate-spin" : ""}`}
            />
            Refresh
          </Button>
          <Button
            variant="outline"
            onClick={() => handleExportCSV(filteredCampaigns)}
            disabled={filteredCampaigns.length === 0}
          >
            <Download className="h-4 w-4 mr-2" />
            Export All
          </Button>
        </div>
      </div>

      {/* Filters and Column Customization */}
      <Card className="bg-white border-slate-200">
        <CardContent className="pt-6">
          <div className="flex flex-col lg:flex-row lg:items-start lg:justify-between gap-4">
            <div className="flex-1">
              <CampaignFilters
                filters={filters}
                onFiltersChange={setFilters}
                globalDateRange={globalDateRange}
              />
            </div>
            <ColumnCustomization
              visibility={columnVisibility}
              onVisibilityChange={setColumnVisibility}
            />
          </div>
        </CardContent>
      </Card>

      {/* Campaigns Table or Empty State */}
      {!isLoading && filteredCampaigns.length === 0 ? (
        <EmptyState
          type={getEmptyStateType()}
          onClearFilters={
            getEmptyStateType() === "no-results" ? clearFilters : undefined
          }
        />
      ) : (
        <Card className="bg-white border-slate-200">
          <CardHeader className="pb-0">
            <div className="flex items-center justify-between">
              <CardTitle className="text-slate-900">
                All Campaigns
                {!isLoading && (
                  <span className="ml-2 text-sm font-normal text-slate-500">
                    ({filteredCampaigns.length} campaigns)
                  </span>
                )}
              </CardTitle>
            </div>
          </CardHeader>
          <CardContent className="pt-4">
            <CampaignsTable
              data={filteredCampaigns}
              isLoading={isLoading}
              columnVisibility={columnVisibility}
              onColumnVisibilityChange={setColumnVisibility}
              onExportCSV={handleExportCSV}
            />
          </CardContent>
        </Card>
      )}
    </div>
  );
}
