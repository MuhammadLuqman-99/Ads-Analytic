export { authApi, default as auth } from "./auth";
export { accountsApi, default as accounts } from "./accounts";
export { campaignsApi, default as campaigns } from "./campaigns";
export { analyticsApi, default as analytics } from "./analytics";
export { settingsApi, default as settings } from "./settings";
export { adminApi } from "./admin";

// Re-export types for convenience
export type { ListConnectionsParams } from "./accounts";
export type { CampaignListResponse } from "./campaigns";
export type { ComparisonData, ReportData } from "./analytics";
export type {
  AdminMetrics,
  DashboardData,
  TimeSeriesData,
  UserProfile,
  FunnelData,
  CohortData,
  FeatureUsageData,
} from "./admin";
