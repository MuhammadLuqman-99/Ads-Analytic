// Main API client export
export {
  tokenManager,
  getApiClient,
  configureApiClient,
  setRefreshTokenFunction,
  apiGet,
  apiPost,
  apiPut,
  apiPatch,
  apiDelete,
} from "./client";

// Types
export * from "./types";

// Services
export {
  authApi,
  accountsApi,
  campaignsApi,
  analyticsApi,
  settingsApi,
} from "./services";

// Hooks
export {
  // Auth
  useAuth,
  useRequireAuth,
  authKeys,
  auth,
  // Dashboard
  useDashboard,
  useRealTimeMetrics,
  dashboardKeys,
  dashboard,
  // Campaigns
  useCampaigns,
  useCampaignsInfinite,
  useCampaign,
  useCampaignComparison,
  campaignKeys,
  campaigns,
  // Connections
  useConnections,
  useConnection,
  useSyncStatus,
  useAvailablePlatforms,
  connectionKeys,
  connections,
  // Analytics
  useAnalytics,
  useSpendBreakdown,
  useConversionFunnel,
  useInsights,
  useTopBottomCampaigns,
  useReportGenerator,
  analyticsKeys,
  analytics,
} from "./hooks";

// Re-export hook types
export type {
  UseDashboardOptions,
  UseCampaignsOptions,
  UseCampaignOptions,
  UseConnectionsOptions,
  UseAnalyticsOptions,
} from "./hooks";
