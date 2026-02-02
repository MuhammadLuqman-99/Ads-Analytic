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
   */
  async getSummary(dateRange: DateRange): Promise<DashboardSummary> {
    const response = await apiGet<ApiResponse<DashboardSummary>>(
      "/analytics/summary",
      {
        params: {
          from: dateRange.from,
          to: dateRange.to,
        },
      }
    );
    return response.data;
  },

  /**
   * Get time series data
   */
  async getTimeSeries(params: AnalyticsParams): Promise<TimeSeriesData> {
    const response = await apiGet<ApiResponse<TimeSeriesData>>(
      "/analytics/timeseries",
      {
        params: {
          from: params.dateRange.from,
          to: params.dateRange.to,
          granularity: params.granularity || "day",
          platforms: params.platforms?.join(","),
          accountIds: params.accountIds?.join(","),
          campaignIds: params.campaignIds?.join(","),
          metrics: params.metrics?.join(","),
          groupBy: params.groupBy,
        },
      }
    );
    return response.data;
  },

  /**
   * Get platform comparison data
   */
  async getPlatformComparison(
    dateRange: DateRange,
    platforms?: Platform[]
  ): Promise<PlatformMetrics[]> {
    const response = await apiGet<ApiResponse<PlatformMetrics[]>>(
      "/analytics/platforms",
      {
        params: {
          from: dateRange.from,
          to: dateRange.to,
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
      "/analytics/top-campaigns",
      {
        params: {
          from: dateRange.from,
          to: dateRange.to,
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
      "/analytics/bottom-campaigns",
      {
        params: {
          from: dateRange.from,
          to: dateRange.to,
          metric: options?.metric || "roas",
          limit: options?.limit || 5,
          platforms: options?.platforms?.join(","),
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
        from: dateRange.from,
        to: dateRange.to,
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
        from: dateRange.from,
        to: dateRange.to,
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
        from: dateRange.from,
        to: dateRange.to,
      },
    });
    return response.data;
  },
};

export default analyticsApi;
