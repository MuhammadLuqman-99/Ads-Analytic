"use client";

import { useQuery, useMutation, useQueryClient, useInfiniteQuery } from "@tanstack/react-query";
import { campaignsApi } from "../services";
import type {
  Campaign,
  CampaignFilters,
  DateRange,
} from "../types";

// ============================================
// Query Keys
// ============================================

export const campaignKeys = {
  all: ["campaigns"] as const,
  lists: () => [...campaignKeys.all, "list"] as const,
  list: (filters?: CampaignFilters) =>
    [...campaignKeys.lists(), filters] as const,
  details: () => [...campaignKeys.all, "detail"] as const,
  detail: (id: string) => [...campaignKeys.details(), id] as const,
  metrics: (id: string, dateRange?: DateRange) =>
    [...campaignKeys.detail(id), "metrics", dateRange] as const,
  timeSeries: (id: string, dateRange?: DateRange) =>
    [...campaignKeys.detail(id), "timeSeries", dateRange] as const,
  insights: (id: string) => [...campaignKeys.detail(id), "insights"] as const,
};

// ============================================
// useCampaigns Hook
// ============================================

export interface UseCampaignsOptions {
  filters?: CampaignFilters;
  enabled?: boolean;
}

export function useCampaigns(options: UseCampaignsOptions = {}) {
  const { filters, enabled = true } = options;
  const queryClient = useQueryClient();

  // List campaigns
  const query = useQuery({
    queryKey: campaignKeys.list(filters),
    queryFn: () => campaignsApi.list(filters),
    enabled,
    staleTime: 2 * 60 * 1000, // 2 minutes
  });

  // Pause campaign mutation
  const pauseMutation = useMutation({
    mutationFn: (campaignId: string) => campaignsApi.pause(campaignId),
    onSuccess: (updatedCampaign) => {
      // Update the campaign in the list
      queryClient.setQueryData(
        campaignKeys.list(filters),
        (old: Awaited<ReturnType<typeof campaignsApi.list>> | undefined) => {
          if (!old) return old;
          return {
            ...old,
            data: old.data.map((c) =>
              c.id === updatedCampaign.id ? updatedCampaign : c
            ),
          };
        }
      );
      // Update individual campaign cache
      queryClient.setQueryData(
        campaignKeys.detail(updatedCampaign.id),
        updatedCampaign
      );
    },
  });

  // Resume campaign mutation
  const resumeMutation = useMutation({
    mutationFn: (campaignId: string) => campaignsApi.resume(campaignId),
    onSuccess: (updatedCampaign) => {
      queryClient.setQueryData(
        campaignKeys.list(filters),
        (old: Awaited<ReturnType<typeof campaignsApi.list>> | undefined) => {
          if (!old) return old;
          return {
            ...old,
            data: old.data.map((c) =>
              c.id === updatedCampaign.id ? updatedCampaign : c
            ),
          };
        }
      );
      queryClient.setQueryData(
        campaignKeys.detail(updatedCampaign.id),
        updatedCampaign
      );
    },
  });

  // Export mutation
  const exportMutation = useMutation({
    mutationFn: (format: "csv" | "xlsx" = "csv") =>
      campaignsApi.downloadExport(filters, format),
  });

  return {
    // Data
    campaigns: query.data?.data || [],
    pagination: query.data?.pagination,
    summary: query.data?.summary,

    // States
    isLoading: query.isLoading,
    isFetching: query.isFetching,
    error: query.error,

    // Actions
    refetch: query.refetch,
    pause: pauseMutation.mutateAsync,
    resume: resumeMutation.mutateAsync,
    exportCampaigns: exportMutation.mutate,

    // Mutation states
    isPausing: pauseMutation.isPending,
    isResuming: resumeMutation.isPending,
    isExporting: exportMutation.isPending,
  };
}

// ============================================
// useCampaignsInfinite Hook (for infinite scroll)
// ============================================

export function useCampaignsInfinite(filters?: Omit<CampaignFilters, "page">) {
  return useInfiniteQuery({
    queryKey: [...campaignKeys.lists(), "infinite", filters],
    queryFn: async ({ pageParam = 1 }) => {
      return campaignsApi.list({ ...filters, page: pageParam });
    },
    initialPageParam: 1,
    getNextPageParam: (lastPage) => {
      if (lastPage.pagination.hasNext) {
        return lastPage.pagination.page + 1;
      }
      return undefined;
    },
    staleTime: 2 * 60 * 1000,
  });
}

// ============================================
// useCampaign Hook (single campaign)
// ============================================

export interface UseCampaignOptions {
  campaignId: string;
  includeMetrics?: boolean;
  dateRange?: DateRange;
  enabled?: boolean;
}

export function useCampaign(options: UseCampaignOptions) {
  const { campaignId, includeMetrics, dateRange, enabled = true } = options;

  // Campaign details
  const detailQuery = useQuery({
    queryKey: campaignKeys.detail(campaignId),
    queryFn: () =>
      campaignsApi.getById(campaignId, { includeMetrics, dateRange }),
    enabled: enabled && !!campaignId,
    staleTime: 2 * 60 * 1000,
  });

  // Campaign metrics (separate query for flexibility)
  const metricsQuery = useQuery({
    queryKey: campaignKeys.metrics(campaignId, dateRange),
    queryFn: () => campaignsApi.getMetrics(campaignId, dateRange!),
    enabled: enabled && !!campaignId && !!dateRange && !includeMetrics,
    staleTime: 2 * 60 * 1000,
  });

  // Time series
  const timeSeriesQuery = useQuery({
    queryKey: campaignKeys.timeSeries(campaignId, dateRange),
    queryFn: () =>
      campaignsApi.getTimeSeries(campaignId, {
        dateRange: dateRange!,
        granularity: "day",
      }),
    enabled: enabled && !!campaignId && !!dateRange,
    staleTime: 2 * 60 * 1000,
  });

  // Insights
  const insightsQuery = useQuery({
    queryKey: campaignKeys.insights(campaignId),
    queryFn: () => campaignsApi.getInsights(campaignId),
    enabled: enabled && !!campaignId,
    staleTime: 10 * 60 * 1000, // 10 minutes (insights change less frequently)
  });

  return {
    // Data
    campaign: detailQuery.data,
    metrics: detailQuery.data?.metrics || metricsQuery.data,
    timeSeries: timeSeriesQuery.data,
    insights: insightsQuery.data,

    // States
    isLoading: detailQuery.isLoading,
    error: detailQuery.error,

    // Individual queries
    queries: {
      detail: detailQuery,
      metrics: metricsQuery,
      timeSeries: timeSeriesQuery,
      insights: insightsQuery,
    },
  };
}

// ============================================
// useCampaignComparison Hook
// ============================================

export function useCampaignComparison(
  campaignIds: string[],
  dateRange: DateRange
) {
  return useQuery({
    queryKey: ["campaigns", "comparison", campaignIds, dateRange],
    queryFn: () => campaignsApi.compare(campaignIds, dateRange),
    enabled: campaignIds.length >= 2 && !!dateRange,
    staleTime: 5 * 60 * 1000,
  });
}

export default useCampaigns;
