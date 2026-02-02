"use client";

import { useQuery, useMutation } from "@tanstack/react-query";
import { analyticsApi } from "../services";
import type { AnalyticsParams, DateRange, Platform, AdMetrics } from "../types";

// ============================================
// Query Keys
// ============================================

export const analyticsKeys = {
  all: ["analytics"] as const,
  summary: (dateRange: DateRange) =>
    [...analyticsKeys.all, "summary", dateRange] as const,
  timeSeries: (params: AnalyticsParams) =>
    [...analyticsKeys.all, "timeSeries", params] as const,
  platforms: (dateRange: DateRange, platforms?: Platform[]) =>
    [...analyticsKeys.all, "platforms", dateRange, platforms] as const,
  comparison: (current: DateRange, previous: DateRange) =>
    [...analyticsKeys.all, "comparison", current, previous] as const,
  spendBreakdown: (dateRange: DateRange, dimension: string) =>
    [...analyticsKeys.all, "spendBreakdown", dateRange, dimension] as const,
  funnel: (dateRange: DateRange) =>
    [...analyticsKeys.all, "funnel", dateRange] as const,
  insights: (dateRange: DateRange) =>
    [...analyticsKeys.all, "insights", dateRange] as const,
  topCampaigns: (dateRange: DateRange, metric?: string) =>
    [...analyticsKeys.all, "topCampaigns", dateRange, metric] as const,
  bottomCampaigns: (dateRange: DateRange, metric?: string) =>
    [...analyticsKeys.all, "bottomCampaigns", dateRange, metric] as const,
};

// ============================================
// useAnalytics Hook
// ============================================

export interface UseAnalyticsOptions extends AnalyticsParams {
  enabled?: boolean;
}

export function useAnalytics(options: UseAnalyticsOptions) {
  const {
    dateRange,
    compareRange,
    platforms,
    accountIds,
    campaignIds,
    granularity = "day",
    metrics,
    groupBy,
    enabled = true,
  } = options;

  // Time series data
  const timeSeriesQuery = useQuery({
    queryKey: analyticsKeys.timeSeries({
      dateRange,
      platforms,
      accountIds,
      campaignIds,
      granularity,
      metrics,
      groupBy,
    }),
    queryFn: () =>
      analyticsApi.getTimeSeries({
        dateRange,
        platforms,
        accountIds,
        campaignIds,
        granularity,
        metrics,
        groupBy,
      }),
    enabled,
    staleTime: 5 * 60 * 1000,
  });

  // Platform comparison
  const platformsQuery = useQuery({
    queryKey: analyticsKeys.platforms(dateRange, platforms),
    queryFn: () => analyticsApi.getPlatformComparison(dateRange, platforms),
    enabled,
    staleTime: 5 * 60 * 1000,
  });

  // Period comparison (if compare range provided)
  const comparisonQuery = useQuery({
    queryKey: compareRange
      ? analyticsKeys.comparison(dateRange, compareRange)
      : [],
    queryFn: () =>
      analyticsApi.compareRanges(dateRange, compareRange!, { platforms, accountIds }),
    enabled: enabled && !!compareRange,
    staleTime: 5 * 60 * 1000,
  });

  return {
    // Data
    timeSeries: timeSeriesQuery.data,
    platforms: platformsQuery.data,
    comparison: comparisonQuery.data,

    // States
    isLoading:
      timeSeriesQuery.isLoading ||
      platformsQuery.isLoading ||
      (!!compareRange && comparisonQuery.isLoading),
    error: timeSeriesQuery.error || platformsQuery.error || comparisonQuery.error,
    isFetching:
      timeSeriesQuery.isFetching ||
      platformsQuery.isFetching ||
      comparisonQuery.isFetching,

    // Actions
    refetch: async () => {
      await Promise.all([
        timeSeriesQuery.refetch(),
        platformsQuery.refetch(),
        compareRange && comparisonQuery.refetch(),
      ]);
    },

    // Individual queries
    queries: {
      timeSeries: timeSeriesQuery,
      platforms: platformsQuery,
      comparison: comparisonQuery,
    },
  };
}

// ============================================
// useSpendBreakdown Hook
// ============================================

export function useSpendBreakdown(
  dateRange: DateRange,
  dimension: "platform" | "account" | "campaign" | "day" | "week"
) {
  return useQuery({
    queryKey: analyticsKeys.spendBreakdown(dateRange, dimension),
    queryFn: () => analyticsApi.getSpendBreakdown(dateRange, dimension),
    staleTime: 5 * 60 * 1000,
  });
}

// ============================================
// useConversionFunnel Hook
// ============================================

export function useConversionFunnel(
  dateRange: DateRange,
  options?: { platforms?: Platform[]; campaignIds?: string[] }
) {
  return useQuery({
    queryKey: analyticsKeys.funnel(dateRange),
    queryFn: () => analyticsApi.getConversionFunnel(dateRange, options),
    staleTime: 5 * 60 * 1000,
  });
}

// ============================================
// useInsights Hook
// ============================================

export function useInsights(dateRange: DateRange) {
  return useQuery({
    queryKey: analyticsKeys.insights(dateRange),
    queryFn: () => analyticsApi.getInsights(dateRange),
    staleTime: 10 * 60 * 1000, // 10 minutes
  });
}

// ============================================
// useTopBottomCampaigns Hook
// ============================================

export function useTopBottomCampaigns(
  dateRange: DateRange,
  options?: {
    metric?: keyof AdMetrics;
    limit?: number;
    platforms?: Platform[];
  }
) {
  const topQuery = useQuery({
    queryKey: analyticsKeys.topCampaigns(dateRange, options?.metric),
    queryFn: () => analyticsApi.getTopCampaigns(dateRange, options),
    staleTime: 5 * 60 * 1000,
  });

  const bottomQuery = useQuery({
    queryKey: analyticsKeys.bottomCampaigns(dateRange, options?.metric),
    queryFn: () => analyticsApi.getBottomCampaigns(dateRange, options),
    staleTime: 5 * 60 * 1000,
  });

  return {
    top: topQuery.data,
    bottom: bottomQuery.data,
    isLoading: topQuery.isLoading || bottomQuery.isLoading,
    error: topQuery.error || bottomQuery.error,
  };
}

// ============================================
// useReportGenerator Hook
// ============================================

export function useReportGenerator() {
  const generateMutation = useMutation({
    mutationFn: analyticsApi.generateReport,
  });

  const downloadReport = async (params: Parameters<typeof analyticsApi.generateReport>[0]) => {
    const result = await generateMutation.mutateAsync({
      ...params,
      format: params.format || "pdf",
    });

    if (result instanceof Blob) {
      const url = window.URL.createObjectURL(result);
      const link = document.createElement("a");
      link.href = url;
      link.setAttribute(
        "download",
        `report-${new Date().toISOString().split("T")[0]}.${params.format || "pdf"}`
      );
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      window.URL.revokeObjectURL(url);
    }

    return result;
  };

  return {
    generateReport: generateMutation.mutateAsync,
    downloadReport,
    isGenerating: generateMutation.isPending,
    error: generateMutation.error,
  };
}

export default useAnalytics;
