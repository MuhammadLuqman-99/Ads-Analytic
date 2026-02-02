"use client";

import { useState, useMemo } from "react";
import {
  RefreshCw,
  DollarSign,
  TrendingUp,
  ShoppingCart,
  Wallet,
  CalendarIcon,
  AlertCircle,
} from "lucide-react";
import { format, subDays, differenceInHours } from "date-fns";
import { Button } from "@/components/ui/button";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import { Calendar } from "@/components/ui/calendar";
import {
  MetricsCard,
  MetricsCardSkeleton,
  SpendChart,
  PlatformCards,
  TopCampaigns,
  RecentActivity,
} from "@/components/dashboard";
import { useDashboard } from "@/lib/api/hooks/use-dashboard";
import type { DateRange } from "@/lib/api/types";
import type { Platform } from "@/lib/mock-data";
import { cn } from "@/lib/utils";

// Helper to format currency
function formatCurrency(value: number): string {
  return new Intl.NumberFormat("en-MY", {
    style: "currency",
    currency: "MYR",
    minimumFractionDigits: 0,
    maximumFractionDigits: 0,
  }).format(value);
}

// Helper to format numbers
function formatNumber(value: number): string {
  return new Intl.NumberFormat("en-MY").format(value);
}

// Date range presets
const DATE_PRESETS = [
  { label: "Last 7 days", days: 7 },
  { label: "Last 14 days", days: 14 },
  { label: "Last 30 days", days: 30 },
  { label: "Last 90 days", days: 90 },
];

export default function DashboardPage() {
  // Date range state - default to last 30 days
  const [dateRange, setDateRange] = useState<DateRange>(() => ({
    from: subDays(new Date(), 30),
    to: new Date(),
  }));
  const [calendarOpen, setCalendarOpen] = useState(false);
  const [selectedPlatform, setSelectedPlatform] = useState<Platform | null>(null);

  // Use the real dashboard hook
  const {
    summary,
    isLoading,
    isFetching,
    error,
    refetch,
  } = useDashboard({
    dateRange,
    enabled: true,
  });

  // Check if data is stale (last sync > 1 hour ago)
  const isDataStale = useMemo(() => {
    if (!summary) return false;
    // Check if summary has a lastSyncedAt field
    const lastSynced = (summary as { lastSyncedAt?: string }).lastSyncedAt;
    if (!lastSynced) return false;
    return differenceInHours(new Date(), new Date(lastSynced)) > 1;
  }, [summary]);

  const handlePlatformClick = (platform: Platform) => {
    setSelectedPlatform(selectedPlatform === platform ? null : platform);
  };

  const handleDatePreset = (days: number) => {
    setDateRange({
      from: subDays(new Date(), days),
      to: new Date(),
    });
    setCalendarOpen(false);
  };

  const handleCalendarSelect = (range: { from?: Date; to?: Date } | undefined) => {
    if (range?.from) {
      setDateRange({
        from: range.from,
        to: range.to || range.from,
      });
      if (range.to) {
        setCalendarOpen(false);
      }
    }
  };

  // Format date range for display
  const dateRangeLabel = useMemo(() => {
    const from = dateRange.from instanceof Date ? dateRange.from : new Date(dateRange.from);
    const to = dateRange.to instanceof Date ? dateRange.to : new Date(dateRange.to);
    return `${format(from, "MMM d, yyyy")} - ${format(to, "MMM d, yyyy")}`;
  }, [dateRange]);

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
        <div>
          <h1 className="text-3xl font-bold text-slate-900">Dashboard</h1>
          <p className="mt-1 text-slate-500">
            Overview of your ad performance across all platforms
          </p>
        </div>
        <div className="flex items-center gap-3">
          {/* Date Range Picker */}
          <Popover open={calendarOpen} onOpenChange={setCalendarOpen}>
            <PopoverTrigger asChild>
              <Button variant="outline" className="justify-start text-left font-normal">
                <CalendarIcon className="mr-2 h-4 w-4" />
                {dateRangeLabel}
              </Button>
            </PopoverTrigger>
            <PopoverContent className="w-auto p-0" align="end">
              <div className="flex">
                <div className="border-r border-slate-700 p-3 space-y-1">
                  {DATE_PRESETS.map((preset) => (
                    <Button
                      key={preset.days}
                      variant="ghost"
                      size="sm"
                      className="w-full justify-start"
                      onClick={() => handleDatePreset(preset.days)}
                    >
                      {preset.label}
                    </Button>
                  ))}
                </div>
                <Calendar
                  mode="range"
                  selected={{
                    from: dateRange.from instanceof Date ? dateRange.from : new Date(dateRange.from),
                    to: dateRange.to instanceof Date ? dateRange.to : new Date(dateRange.to),
                  }}
                  onSelect={handleCalendarSelect}
                  numberOfMonths={2}
                  disabled={{ after: new Date() }}
                />
              </div>
            </PopoverContent>
          </Popover>

          {/* Sync Button */}
          <Button
            onClick={() => refetch()}
            disabled={isFetching}
          >
            <RefreshCw
              className={cn("h-4 w-4 mr-2", isFetching && "animate-spin")}
            />
            {isFetching ? "Syncing..." : "Refresh"}
          </Button>
        </div>
      </div>

      {/* Stale Data Warning */}
      {isDataStale && (
        <div className="flex items-center gap-2 p-3 bg-amber-50 border border-amber-200 rounded-lg text-amber-800">
          <AlertCircle className="h-4 w-4" />
          <span className="text-sm">
            Data may be outdated. Last synced more than 1 hour ago.
          </span>
          <Button
            variant="ghost"
            size="sm"
            className="ml-auto text-amber-800 hover:text-amber-900 hover:bg-amber-100"
            onClick={() => refetch()}
          >
            Refresh Now
          </Button>
        </div>
      )}

      {/* Error State */}
      {error && (
        <div className="flex items-center gap-2 p-3 bg-red-50 border border-red-200 rounded-lg text-red-800">
          <AlertCircle className="h-4 w-4" />
          <span className="text-sm">
            Failed to load dashboard data. Please try again.
          </span>
          <Button
            variant="ghost"
            size="sm"
            className="ml-auto text-red-800 hover:text-red-900 hover:bg-red-100"
            onClick={() => refetch()}
          >
            Retry
          </Button>
        </div>
      )}

      {/* Key Metrics Row */}
      {isLoading ? (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
          {[1, 2, 3, 4].map((i) => (
            <MetricsCardSkeleton key={i} />
          ))}
        </div>
      ) : (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
          <MetricsCard
            title="Total Spend"
            value={formatCurrency(summary?.totals?.spend || 0)}
            change={summary?.changes?.spend || 0}
            icon={<Wallet className="h-5 w-5 text-white" />}
            iconBgColor="bg-gradient-to-br from-indigo-500 to-indigo-600"
            subtitle="vs last period"
          />
          <MetricsCard
            title="Total Revenue"
            value={formatCurrency(summary?.totals?.revenue || 0)}
            change={summary?.changes?.revenue || 0}
            icon={<DollarSign className="h-5 w-5 text-white" />}
            iconBgColor="bg-gradient-to-br from-emerald-500 to-emerald-600"
            subtitle="vs last period"
          />
          <MetricsCard
            title="Overall ROAS"
            value={`${summary?.totals?.roas?.toFixed(1) || 0}x`}
            change={summary?.changes?.roas || 0}
            icon={<TrendingUp className="h-5 w-5 text-white" />}
            iconBgColor="bg-gradient-to-br from-amber-500 to-amber-600"
            subtitle="Return on ad spend"
          />
          <MetricsCard
            title="Total Conversions"
            value={formatNumber(summary?.totals?.conversions || 0)}
            change={summary?.changes?.conversions || 0}
            icon={<ShoppingCart className="h-5 w-5 text-white" />}
            iconBgColor="bg-gradient-to-br from-purple-500 to-purple-600"
            subtitle="vs last period"
          />
        </div>
      )}

      {/* Spend Chart */}
      <SpendChart />

      {/* Platform Performance */}
      <div>
        <h2 className="text-lg font-semibold text-slate-900 mb-4">
          Platform Performance
          {selectedPlatform && (
            <span className="text-sm font-normal text-slate-500 ml-2">
              (Click again to clear filter)
            </span>
          )}
        </h2>
        <PlatformCards
          selectedPlatform={selectedPlatform}
          onPlatformClick={handlePlatformClick}
        />
      </div>

      {/* Bottom Row: Top Campaigns + Recent Activity */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <div className="lg:col-span-2">
          <TopCampaigns platformFilter={selectedPlatform} />
        </div>
        <div>
          <RecentActivity />
        </div>
      </div>
    </div>
  );
}
