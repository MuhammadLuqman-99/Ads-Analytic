"use client";

import { useQuery } from "@tanstack/react-query";
import { adminApi } from "../services/admin";

// ============================================
// Query Keys
// ============================================

export const adminKeys = {
  all: ["admin"] as const,
  dashboard: () => [...adminKeys.all, "dashboard"] as const,
  metrics: () => [...adminKeys.all, "metrics"] as const,
  activeUsers: (params?: { from?: string; to?: string }) =>
    [...adminKeys.all, "activeUsers", params] as const,
  churnedUsers: (days: number) => [...adminKeys.all, "churnedUsers", days] as const,
  funnel: (name: string, from?: string, to?: string) =>
    [...adminKeys.all, "funnel", name, from, to] as const,
  platforms: () => [...adminKeys.all, "platforms"] as const,
  features: (from?: string, to?: string) =>
    [...adminKeys.all, "features", from, to] as const,
  revenue: (from?: string, to?: string) =>
    [...adminKeys.all, "revenue", from, to] as const,
  events: (from?: string, to?: string) =>
    [...adminKeys.all, "events", from, to] as const,
  cohorts: (from?: string, to?: string, period?: string) =>
    [...adminKeys.all, "cohorts", from, to, period] as const,
  topUsers: (metric: string, limit: number) =>
    [...adminKeys.all, "topUsers", metric, limit] as const,
};

// ============================================
// Hooks
// ============================================

/**
 * Hook to get admin dashboard overview data
 */
export function useAdminDashboard() {
  return useQuery({
    queryKey: adminKeys.dashboard(),
    queryFn: adminApi.getDashboard,
    staleTime: 60 * 1000, // 1 minute
    refetchInterval: 5 * 60 * 1000, // Refresh every 5 minutes
  });
}

/**
 * Hook to get detailed admin metrics
 */
export function useAdminMetrics() {
  return useQuery({
    queryKey: adminKeys.metrics(),
    queryFn: adminApi.getMetrics,
    staleTime: 60 * 1000,
  });
}

/**
 * Hook to get active users metrics
 */
export function useActiveUsers(params?: { from?: string; to?: string; granularity?: string }) {
  return useQuery({
    queryKey: adminKeys.activeUsers(params),
    queryFn: () => adminApi.getActiveUsers(params),
    staleTime: 60 * 1000,
  });
}

/**
 * Hook to get churned users
 */
export function useChurnedUsers(days = 30) {
  return useQuery({
    queryKey: adminKeys.churnedUsers(days),
    queryFn: () => adminApi.getChurnedUsers(days),
    staleTime: 5 * 60 * 1000,
  });
}

/**
 * Hook to get funnel data
 */
export function useFunnel(name: string, from?: string, to?: string) {
  return useQuery({
    queryKey: adminKeys.funnel(name, from, to),
    queryFn: () => adminApi.getFunnel(name, from, to),
    staleTime: 5 * 60 * 1000,
    enabled: !!name,
  });
}

/**
 * Hook to get platform breakdown
 */
export function usePlatformBreakdown() {
  return useQuery({
    queryKey: adminKeys.platforms(),
    queryFn: adminApi.getPlatformBreakdown,
    staleTime: 5 * 60 * 1000,
  });
}

/**
 * Hook to get feature usage
 */
export function useFeatureUsage(from?: string, to?: string) {
  return useQuery({
    queryKey: adminKeys.features(from, to),
    queryFn: () => adminApi.getFeatureUsage(from, to),
    staleTime: 5 * 60 * 1000,
  });
}

/**
 * Hook to get revenue metrics
 */
export function useRevenue(from?: string, to?: string) {
  return useQuery({
    queryKey: adminKeys.revenue(from, to),
    queryFn: () => adminApi.getRevenue(from, to),
    staleTime: 5 * 60 * 1000,
  });
}

/**
 * Hook to get event breakdown
 */
export function useEventBreakdown(from?: string, to?: string) {
  return useQuery({
    queryKey: adminKeys.events(from, to),
    queryFn: () => adminApi.getEvents(from, to),
    staleTime: 5 * 60 * 1000,
  });
}

/**
 * Hook to get cohort analysis
 */
export function useCohortAnalysis(from?: string, to?: string, period = "week") {
  return useQuery({
    queryKey: adminKeys.cohorts(from, to, period),
    queryFn: () => adminApi.getCohortAnalysis(from, to, period),
    staleTime: 10 * 60 * 1000,
  });
}

/**
 * Hook to get top users
 */
export function useTopUsers(metric = "events", limit = 10) {
  return useQuery({
    queryKey: adminKeys.topUsers(metric, limit),
    queryFn: () => adminApi.getTopUsers(metric, limit),
    staleTime: 5 * 60 * 1000,
  });
}
