// Auth hooks
export {
  useAuth,
  useRequireAuth,
  authKeys,
  default as auth,
} from "./use-auth";

// Dashboard hooks
export {
  useDashboard,
  useRealTimeMetrics,
  dashboardKeys,
  default as dashboard,
} from "./use-dashboard";

// Campaign hooks
export {
  useCampaigns,
  useCampaignsInfinite,
  useCampaign,
  useCampaignComparison,
  campaignKeys,
  default as campaigns,
} from "./use-campaigns";

// Connection hooks
export {
  useConnections,
  useConnection,
  useSyncStatus,
  useAvailablePlatforms,
  connectionKeys,
  default as connections,
} from "./use-connections";

// Analytics hooks
export {
  useAnalytics,
  useSpendBreakdown,
  useConversionFunnel,
  useInsights,
  useTopBottomCampaigns,
  useReportGenerator,
  analyticsKeys,
  default as analytics,
} from "./use-analytics";

// Event stream hooks (SSE)
export {
  useEventStream,
  default as eventStream,
} from "./use-event-stream";

// Admin hooks
export {
  useAdminDashboard,
  useAdminMetrics,
  useActiveUsers,
  useChurnedUsers,
  useFunnel,
  usePlatformBreakdown,
  useFeatureUsage,
  useRevenue,
  useEventBreakdown,
  useCohortAnalysis,
  useTopUsers,
  adminKeys,
} from "./use-admin";

// Re-export types
export type { UseDashboardOptions } from "./use-dashboard";
export type { UseCampaignsOptions, UseCampaignOptions } from "./use-campaigns";
export type { UseConnectionsOptions } from "./use-connections";
export type { UseAnalyticsOptions } from "./use-analytics";
export type {
  UseEventStreamOptions,
  SyncStatus,
  ToastNotification,
  EventType,
  SyncStartedData,
  SyncProgressData,
  SyncCompletedData,
  SyncErrorData,
  DataUpdatedData,
} from "./use-event-stream";
