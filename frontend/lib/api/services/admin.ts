import { apiGet } from "../client";

// ============================================
// Types
// ============================================

export interface AdminMetrics {
  totalUsers: number;
  activeUsers: number;
  totalRevenue: number;
  mrr: number;
  churnRate: number;
  platformConnections: Record<string, number>;
}

export interface TimeSeriesData {
  timestamps: string[];
  values: number[];
  labels?: string[];
}

export interface DashboardData {
  metrics: AdminMetrics;
  dau: number;
  wau: number;
  mau: number;
  recentSignups: number;
  platformBreakdown: Record<string, number>;
}

export interface UserProfile {
  userId: string;
  email: string;
  name: string;
  firstSeenAt: string;
  lastSeenAt: string;
  totalEvents: number;
  totalSessions: number;
  properties: Record<string, unknown>;
}

export interface FunnelStep {
  name: string;
  count: number;
  conversionRate: number;
}

export interface FunnelData {
  name: string;
  steps: FunnelStep[];
  overallConversionRate: number;
}

export interface CohortData {
  period: string;
  cohorts: {
    cohort: string;
    size: number;
    retentionByPeriod: Record<string, number>;
  }[];
}

export interface FeatureUsageData {
  features: Record<string, number>;
}

export interface EventBreakdown {
  events: Record<string, number>;
}

// ============================================
// Admin API Service
// ============================================

export const adminApi = {
  /**
   * Get admin dashboard overview
   */
  getDashboard: () => apiGet<DashboardData>("/admin/dashboard"),

  /**
   * Get detailed metrics
   */
  getMetrics: () => apiGet<AdminMetrics>("/admin/metrics"),

  /**
   * Get active users (DAU, WAU, MAU)
   */
  getActiveUsers: (params?: { from?: string; to?: string; granularity?: string }) =>
    apiGet<{
      dau: number;
      wau: number;
      mau: number;
      timeSeries?: TimeSeriesData;
    }>("/admin/users/active", { params }),

  /**
   * Get churned users list
   */
  getChurnedUsers: (days = 30) =>
    apiGet<{ users: UserProfile[] }>("/admin/users/churned", { params: { days } }),

  /**
   * Get conversion funnel data
   */
  getFunnel: (name: string, from?: string, to?: string) =>
    apiGet<FunnelData>(`/admin/funnels/${name}`, { params: { from, to } }),

  /**
   * Get platform connection breakdown
   */
  getPlatformBreakdown: () =>
    apiGet<{ platforms: Record<string, number> }>("/admin/platforms"),

  /**
   * Get feature usage heatmap data
   */
  getFeatureUsage: (from?: string, to?: string) =>
    apiGet<FeatureUsageData>("/admin/features", { params: { from, to } }),

  /**
   * Get revenue metrics
   */
  getRevenue: (from?: string, to?: string) =>
    apiGet<{
      mrr: number;
      churnRate: number;
      timeSeries?: TimeSeriesData;
    }>("/admin/revenue", { params: { from, to } }),

  /**
   * Get event breakdown by type
   */
  getEvents: (from?: string, to?: string) =>
    apiGet<EventBreakdown>("/admin/events", { params: { from, to } }),

  /**
   * Get cohort analysis data
   */
  getCohortAnalysis: (from?: string, to?: string, period = "week") =>
    apiGet<CohortData>("/admin/cohorts", { params: { from, to, period } }),

  /**
   * Get active users time series
   */
  getActiveUsersTimeSeries: (from: string, to: string, granularity = "day") =>
    apiGet<TimeSeriesData>("/admin/users/active/timeseries", {
      params: { from, to, granularity },
    }),

  /**
   * Get revenue time series
   */
  getRevenueTimeSeries: (from: string, to: string) =>
    apiGet<TimeSeriesData>("/admin/revenue/timeseries", { params: { from, to } }),

  /**
   * Get top users by activity
   */
  getTopUsers: (metric = "events", limit = 10) =>
    apiGet<{ users: UserProfile[] }>("/admin/users/top", { params: { metric, limit } }),
};
