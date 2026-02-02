"use client";

import { useQuery } from "@tanstack/react-query";
import { analyticsApi } from "../services";
import type {
  DateRange,
  DashboardSummary,
  TimeSeriesData,
  PlatformMetrics,
  CampaignPerformance,
  Platform,
} from "../types";

// ============================================
// Query Keys
// ============================================

export const dashboardKeys = {
  all: ["dashboard"] as const,
  summary: (dateRange: DateRange) =>
    [...dashboardKeys.all, "summary", dateRange] as const,
  timeSeries: (dateRange: DateRange, granularity?: string) =>
    [...dashboardKeys.all, "timeSeries", dateRange, granularity] as const,
  platforms: (dateRange: DateRange) =>
    [...dashboardKeys.all, "platforms", dateRange] as const,
  topCampaigns: (dateRange: DateRange) =>
    [...dashboardKeys.all, "topCampaigns", dateRange] as const,
  bottomCampaigns: (dateRange: DateRange) =>
    [...dashboardKeys.all, "bottomCampaigns", dateRange] as const,
  realtime: () => [...dashboardKeys.all, "realtime"] as const,
};

// ============================================
// useDashboard Hook
// ============================================

export interface UseDashboardOptions {
  dateRange: DateRange;
  platforms?: Platform[];
  enabled?: boolean;
}

export function useDashboard(options: UseDashboardOptions) {
  const { dateRange, platforms, enabled = true } = options;

  // Dashboard summary
  const summaryQuery = useQuery({
    queryKey: dashboardKeys.summary(dateRange),
    queryFn: () => analyticsApi.getSummary(dateRange),
    enabled,
    staleTime: 5 * 60 * 1000, // 5 minutes
  });

  // Time series data
  const timeSeriesQuery = useQuery({
    queryKey: dashboardKeys.timeSeries(dateRange, "day"),
    queryFn: () =>
      analyticsApi.getTimeSeries({
        dateRange,
        granularity: "day",
        platforms,
      }),
    enabled,
    staleTime: 5 * 60 * 1000,
  });

  // Platform comparison
  const platformsQuery = useQuery({
    queryKey: dashboardKeys.platforms(dateRange),
    queryFn: () => analyticsApi.getPlatformComparison(dateRange, platforms),
    enabled,
    staleTime: 5 * 60 * 1000,
  });

  // Top campaigns
  const topCampaignsQuery = useQuery({
    queryKey: dashboardKeys.topCampaigns(dateRange),
    queryFn: () =>
      analyticsApi.getTopCampaigns(dateRange, {
        limit: 5,
        metric: "roas",
        platforms,
      }),
    enabled,
    staleTime: 5 * 60 * 1000,
  });

  // Bottom campaigns
  const bottomCampaignsQuery = useQuery({
    queryKey: dashboardKeys.bottomCampaigns(dateRange),
    queryFn: () =>
      analyticsApi.getBottomCampaigns(dateRange, {
        limit: 5,
        metric: "roas",
        platforms,
      }),
    enabled,
    staleTime: 5 * 60 * 1000,
  });

  // Combined loading state
  const isLoading =
    summaryQuery.isLoading ||
    timeSeriesQuery.isLoading ||
    platformsQuery.isLoading ||
    topCampaignsQuery.isLoading ||
    bottomCampaignsQuery.isLoading;

  // Combined error
  const error =
    summaryQuery.error ||
    timeSeriesQuery.error ||
    platformsQuery.error ||
    topCampaignsQuery.error ||
    bottomCampaignsQuery.error;

  // Refetch all
  const refetch = async () => {
    await Promise.all([
      summaryQuery.refetch(),
      timeSeriesQuery.refetch(),
      platformsQuery.refetch(),
      topCampaignsQuery.refetch(),
      bottomCampaignsQuery.refetch(),
    ]);
  };

  return {
    // Data
    summary: summaryQuery.data,
    timeSeries: timeSeriesQuery.data,
    platforms: platformsQuery.data,
    topCampaigns: topCampaignsQuery.data,
    bottomCampaigns: bottomCampaignsQuery.data,

    // States
    isLoading,
    error,
    isFetching:
      summaryQuery.isFetching ||
      timeSeriesQuery.isFetching ||
      platformsQuery.isFetching,

    // Actions
    refetch,

    // Individual queries (for granular control)
    queries: {
      summary: summaryQuery,
      timeSeries: timeSeriesQuery,
      platforms: platformsQuery,
      topCampaigns: topCampaignsQuery,
      bottomCampaigns: bottomCampaignsQuery,
    },
  };
}

// ============================================
// useRealTimeMetrics Hook
// ============================================

export function useRealTimeMetrics(platforms?: Platform[]) {
  return useQuery({
    queryKey: dashboardKeys.realtime(),
    queryFn: () => analyticsApi.getRealTimeMetrics(platforms),
    refetchInterval: 60 * 1000, // Refresh every minute
    staleTime: 30 * 1000, // 30 seconds
  });
}

export default useDashboard;
