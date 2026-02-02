import { apiGet, apiPost } from "../client";
import {
  DashboardSummary,
  TimeSeriesData,
  PlatformMetrics,
  AnalyticsParams,
  DateRange,
  Platform,
  AdMetrics,
  ApiResponse,
  CampaignPerformance,
} from "../types";

// ============================================
// Analytics API Service
// ============================================

// Helper to format date for backend API (expects YYYY-MM-DD)
function formatDate(date: Date | string): string {
  if (typeof date === "string") {
    // If already a string, try to parse and format
    const parsed = new Date(date);
    if (!isNaN(parsed.getTime())) {
      return parsed.toISOString().split("T")[0];
    }
    return date;
  }
  return date.toISOString().split("T")[0];
}

export interface ComparisonData {
  current: AdMetrics;
  previous: AdMetrics;
  changes: Record<keyof AdMetrics, number>;
  byPlatform: {
    platform: Platform;
    current: AdMetrics;
    previous: AdMetrics;
    change: number;
  }[];
}

export interface ReportData {
  title: string;
  dateRange: DateRange;
  summary: AdMetrics;
  timeSeries: TimeSeriesData;
  platformBreakdown: PlatformMetrics[];
  topCampaigns: CampaignPerformance[];
  insights: string[];
}

export const analyticsApi = {
  /**
   * Get dashboard summary
   * Backend: GET /dashboard/summary
   */
  async getSummary(dateRange: DateRange): Promise<DashboardSummary> {
    const response = await apiGet<ApiResponse<DashboardSummary>>(
      "/dashboard/summary",
      {
        params: {
          start_date: formatDate(dateRange.from),
          end_date: formatDate(dateRange.to),
        },
      }
    );
    return response.data;
  },

  /**
   * Get time series data
   * Backend: GET /dashboard/timeseries
   */
  async getTimeSeries(params: AnalyticsParams): Promise<TimeSeriesData> {
    const response = await apiGet<ApiResponse<TimeSeriesData>>(
      "/dashboard/timeseries",
      {
        params: {
          start_date: formatDate(params.dateRange.from),
          end_date: formatDate(params.dateRange.to),
          granularity: params.granularity || "day",
          platforms: params.platforms?.join(","),
        },
      }
    );
    return response.data;
  },

  /**
   * Get platform comparison data
   * Backend: GET /dashboard/platforms
   */
  async getPlatformComparison(
    dateRange: DateRange,
    platforms?: Platform[]
  ): Promise<PlatformMetrics[]> {
    const response = await apiGet<ApiResponse<PlatformMetrics[]>>(
      "/dashboard/platforms",
      {
        params: {
          start_date: formatDate(dateRange.from),
          end_date: formatDate(dateRange.to),
          platforms: platforms?.join(","),
        },
      }
    );
    return response.data;
  },

  /**
   * Compare two date ranges
   */
  async compareRanges(
    currentRange: DateRange,
    previousRange: DateRange,
    options?: {
      platforms?: Platform[];
      accountIds?: string[];
    }
  ): Promise<ComparisonData> {
    const response = await apiPost<ApiResponse<ComparisonData>>(
      "/analytics/compare",
      {
        currentRange,
        previousRange,
        ...options,
      }
    );
    return response.data;
  },

  /**
   * Get top performing campaigns
   * Backend: GET /dashboard/top-campaigns
   */
  async getTopCampaigns(
    dateRange: DateRange,
    options?: {
      metric?: keyof AdMetrics;
      limit?: number;
      platforms?: Platform[];
    }
  ): Promise<CampaignPerformance[]> {
    const response = await apiGet<ApiResponse<CampaignPerformance[]>>(
      "/dashboard/top-campaigns",
      {
        params: {
          start_date: formatDate(dateRange.from),
          end_date: formatDate(dateRange.to),
          metric: options?.metric || "roas",
          limit: options?.limit || 5,
          platforms: options?.platforms?.join(","),
        },
      }
    );
    return response.data;
  },

  /**
   * Get worst performing campaigns
   * Backend: GET /dashboard/top-campaigns with sort=asc
   */
  async getBottomCampaigns(
    dateRange: DateRange,
    options?: {
      metric?: keyof AdMetrics;
      limit?: number;
      platforms?: Platform[];
    }
  ): Promise<CampaignPerformance[]> {
    const response = await apiGet<ApiResponse<CampaignPerformance[]>>(
      "/dashboard/top-campaigns",
      {
        params: {
          start_date: formatDate(dateRange.from),
          end_date: formatDate(dateRange.to),
          metric: options?.metric || "roas",
          limit: options?.limit || 5,
          platforms: options?.platforms?.join(","),
          sort: "asc", // Get bottom performers
        },
      }
    );
    return response.data;
  },

  /**
   * Generate a report
   */
  async generateReport(params: {
    dateRange: DateRange;
    platforms?: Platform[];
    reportType: "overview" | "platform" | "campaign" | "custom";
    metrics?: (keyof AdMetrics)[];
    format?: "json" | "pdf" | "csv";
  }): Promise<ReportData | Blob> {
    if (params.format === "pdf" || params.format === "csv") {
      const response = await apiPost<Blob>(
        "/analytics/reports/generate",
        params,
        { responseType: "blob" }
      );
      return response;
    }

    const response = await apiPost<ApiResponse<ReportData>>(
      "/analytics/reports/generate",
      params
    );
    return response.data;
  },

  /**
   * Get real-time metrics (last 24 hours, refreshed frequently)
   */
  async getRealTimeMetrics(platforms?: Platform[]): Promise<{
    current: AdMetrics;
    hourlyTrend: { hour: string; metrics: AdMetrics }[];
    activeAlerts: { type: string; message: string; severity: "info" | "warning" | "error" }[];
  }> {
    const response = await apiGet<
      ApiResponse<{
        current: AdMetrics;
        hourlyTrend: { hour: string; metrics: AdMetrics }[];
        activeAlerts: { type: string; message: string; severity: "info" | "warning" | "error" }[];
      }>
    >("/analytics/realtime", {
      params: { platforms: platforms?.join(",") },
    });
    return response.data;
  },

  /**
   * Get spend breakdown by different dimensions
   */
  async getSpendBreakdown(
    dateRange: DateRange,
    dimension: "platform" | "account" | "campaign" | "day" | "week"
  ): Promise<{ name: string; value: number; percentage: number }[]> {
    const response = await apiGet<
      ApiResponse<{ name: string; value: number; percentage: number }[]>
    >("/analytics/spend-breakdown", {
      params: {
        start_date: formatDate(dateRange.from),
        end_date: formatDate(dateRange.to),
        dimension,
      },
    });
    return response.data;
  },

  /**
   * Get conversion funnel data
   */
  async getConversionFunnel(
    dateRange: DateRange,
    options?: { platforms?: Platform[]; campaignIds?: string[] }
  ): Promise<{
    steps: { name: string; value: number; conversionRate: number }[];
    overallRate: number;
  }> {
    const response = await apiGet<
      ApiResponse<{
        steps: { name: string; value: number; conversionRate: number }[];
        overallRate: number;
      }>
    >("/analytics/funnel", {
      params: {
        start_date: formatDate(dateRange.from),
        end_date: formatDate(dateRange.to),
        platforms: options?.platforms?.join(","),
        campaignIds: options?.campaignIds?.join(","),
      },
    });
    return response.data;
  },

  /**
   * Get AI-generated insights
   */
  async getInsights(
    dateRange: DateRange
  ): Promise<{
    insights: { type: string; title: string; description: string; impact: "high" | "medium" | "low" }[];
    recommendations: string[];
    anomalies: { metric: string; expected: number; actual: number; deviation: number }[];
  }> {
    const response = await apiGet<
      ApiResponse<{
        insights: { type: string; title: string; description: string; impact: "high" | "medium" | "low" }[];
        recommendations: string[];
        anomalies: { metric: string; expected: number; actual: number; deviation: number }[];
      }>
    >("/analytics/insights", {
      params: {
        start_date: formatDate(dateRange.from),
        end_date: formatDate(dateRange.to),
      },
    });
    return response.data;
  },
};

export default analyticsApi;
