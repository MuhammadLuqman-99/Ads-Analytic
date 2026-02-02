import { apiGet, apiPost, getApiClient } from "../client";
import {
  Campaign,
  CampaignFilters,
  CampaignMetrics,
  PaginatedResponse,
  ApiResponse,
  DateRange,
  TimeSeriesData,
} from "../types";

// ============================================
// Campaigns API Service
// ============================================

export interface CampaignListResponse extends PaginatedResponse<Campaign> {
  summary?: {
    totalSpend: number;
    totalConversions: number;
    averageRoas: number;
  };
}

export const campaignsApi = {
  /**
   * List campaigns with filters and pagination
   */
  async list(filters?: CampaignFilters): Promise<CampaignListResponse> {
    const params = {
      ...filters,
      dateRange: filters?.dateRange
        ? JSON.stringify(filters.dateRange)
        : undefined,
      platforms: filters?.platforms?.join(","),
      statuses: filters?.statuses?.join(","),
      accountIds: filters?.accountIds?.join(","),
    };

    const response = await apiGet<CampaignListResponse>("/campaigns", {
      params,
    });
    return response;
  },

  /**
   * Get a single campaign by ID
   */
  async getById(
    campaignId: string,
    options?: { includeMetrics?: boolean; dateRange?: DateRange }
  ): Promise<Campaign> {
    const params = {
      includeMetrics: options?.includeMetrics,
      dateRange: options?.dateRange
        ? JSON.stringify(options.dateRange)
        : undefined,
    };

    const response = await apiGet<ApiResponse<Campaign>>(
      `/campaigns/${campaignId}`,
      { params }
    );
    return response.data;
  },

  /**
   * Get campaign metrics for a specific date range
   */
  async getMetrics(
    campaignId: string,
    dateRange: DateRange
  ): Promise<CampaignMetrics> {
    const response = await apiGet<ApiResponse<CampaignMetrics>>(
      `/campaigns/${campaignId}/metrics`,
      { params: { dateRange: JSON.stringify(dateRange) } }
    );
    return response.data;
  },

  /**
   * Get campaign performance time series
   */
  async getTimeSeries(
    campaignId: string,
    params: {
      dateRange: DateRange;
      granularity?: "hour" | "day" | "week" | "month";
    }
  ): Promise<TimeSeriesData> {
    const response = await apiGet<ApiResponse<TimeSeriesData>>(
      `/campaigns/${campaignId}/timeseries`,
      {
        params: {
          dateRange: JSON.stringify(params.dateRange),
          granularity: params.granularity || "day",
        },
      }
    );
    return response.data;
  },

  /**
   * Export campaigns to CSV
   */
  async export(
    filters?: CampaignFilters,
    format: "csv" | "xlsx" = "csv"
  ): Promise<Blob> {
    const params = {
      ...filters,
      format,
      dateRange: filters?.dateRange
        ? JSON.stringify(filters.dateRange)
        : undefined,
      platforms: filters?.platforms?.join(","),
      statuses: filters?.statuses?.join(","),
    };

    const response = await getApiClient().get("/campaigns/export", {
      params,
      responseType: "blob",
    });

    return response.data;
  },

  /**
   * Download export as file
   */
  async downloadExport(
    filters?: CampaignFilters,
    format: "csv" | "xlsx" = "csv"
  ): Promise<void> {
    const blob = await this.export(filters, format);
    const filename = `campaigns-export-${new Date().toISOString().split("T")[0]}.${format}`;

    // Create download link
    const url = window.URL.createObjectURL(blob);
    const link = document.createElement("a");
    link.href = url;
    link.setAttribute("download", filename);
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    window.URL.revokeObjectURL(url);
  },

  /**
   * Pause a campaign
   */
  async pause(campaignId: string): Promise<Campaign> {
    const response = await apiPost<ApiResponse<Campaign>>(
      `/campaigns/${campaignId}/pause`
    );
    return response.data;
  },

  /**
   * Resume a paused campaign
   */
  async resume(campaignId: string): Promise<Campaign> {
    const response = await apiPost<ApiResponse<Campaign>>(
      `/campaigns/${campaignId}/resume`
    );
    return response.data;
  },

  /**
   * Get campaign insights/recommendations
   */
  async getInsights(
    campaignId: string
  ): Promise<{
    score: number;
    recommendations: string[];
    opportunities: { type: string; impact: string; description: string }[];
  }> {
    const response = await apiGet<
      ApiResponse<{
        score: number;
        recommendations: string[];
        opportunities: { type: string; impact: string; description: string }[];
      }>
    >(`/campaigns/${campaignId}/insights`);
    return response.data;
  },

  /**
   * Compare multiple campaigns
   */
  async compare(
    campaignIds: string[],
    dateRange: DateRange
  ): Promise<{
    campaigns: (Campaign & { metrics: CampaignMetrics })[];
    comparison: Record<string, { best: string; worst: string }>;
  }> {
    const response = await apiPost<
      ApiResponse<{
        campaigns: (Campaign & { metrics: CampaignMetrics })[];
        comparison: Record<string, { best: string; worst: string }>;
      }>
    >("/campaigns/compare", {
      campaignIds,
      dateRange,
    });
    return response.data;
  },
};

export default campaignsApi;
