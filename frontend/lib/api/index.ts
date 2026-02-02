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

// Errors
export {
  ApiError,
  ErrorCode,
  isApiError,
  getErrorMessage,
} from "./errors";
export type {
  ErrorCodeType,
  FieldError,
  ApiErrorDetail,
  ApiErrorResponse,
} from "./errors";

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
  // Event Stream
  useEventStream,
  eventStream,
} from "./hooks";

// Re-export hook types
export type {
  UseDashboardOptions,
  UseCampaignsOptions,
  UseCampaignOptions,
  UseConnectionsOptions,
  UseAnalyticsOptions,
  UseEventStreamOptions,
  SyncStatus,
  ToastNotification,
} from "./hooks";
