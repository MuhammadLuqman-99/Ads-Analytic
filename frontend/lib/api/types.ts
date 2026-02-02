// ============================================
// Base API Types
// ============================================

export interface ApiError {
  code: string;
  message: string;
  details?: Record<string, string[]>;
  status: number;
}

export interface ApiResponse<T> {
  data: T;
  message?: string;
  success: boolean;
}

export interface PaginatedResponse<T> {
  data: T[];
  pagination: {
    page: number;
    limit: number;
    total: number;
    totalPages: number;
    hasNext: boolean;
    hasPrev: boolean;
  };
}

export interface DateRange {
  from: Date | string;
  to: Date | string;
}

// ============================================
// User & Organization Types
// ============================================

export type UserRole = "owner" | "admin" | "member" | "viewer";

export interface User {
  id: string;
  email: string;
  name: string;
  avatarUrl?: string;
  role: UserRole;
  organizationId: string;
  emailVerified: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface Organization {
  id: string;
  name: string;
  slug: string;
  logoUrl?: string;
  plan: "free" | "starter" | "pro" | "enterprise";
  planExpiresAt?: string;
  settings: {
    timezone: string;
    currency: string;
    dateFormat: string;
  };
  limits: {
    maxAccounts: number;
    maxUsers: number;
    dataRetentionDays: number;
  };
  createdAt: string;
  updatedAt: string;
}

export interface TeamMember {
  id: string;
  userId: string;
  user: Pick<User, "id" | "email" | "name" | "avatarUrl">;
  role: UserRole;
  invitedAt: string;
  joinedAt?: string;
  status: "pending" | "active" | "inactive";
}

// ============================================
// Connected Account Types
// ============================================

export type Platform = "meta" | "google" | "tiktok" | "shopee" | "linkedin";

export type ConnectionStatus = "active" | "error" | "expired" | "syncing" | "disconnected";

export type ConnectionErrorType =
  | "token_expired"
  | "permission_denied"
  | "api_error"
  | "rate_limit"
  | "account_suspended";

export interface ConnectionError {
  type: ConnectionErrorType;
  message: string;
  occurredAt: string;
  retryable: boolean;
}

export interface ConnectedAccount {
  id: string;
  platform: Platform;
  platformAccountId: string;
  platformAccountName: string;
  status: ConnectionStatus;
  error?: ConnectionError;
  tokenExpiresAt?: string;
  lastSyncAt?: string;
  lastSuccessfulSyncAt?: string;
  syncFrequency: number; // minutes
  dataFreshness: "fresh" | "stale" | "outdated";
  connectedAt: string;
  connectedBy: string;
}

export interface SyncStatus {
  accountId: string;
  status: "idle" | "syncing" | "completed" | "failed";
  progress: number; // 0-100
  startedAt?: string;
  completedAt?: string;
  error?: string;
  syncedRecords?: number;
}

// ============================================
// Campaign Types (Normalized)
// ============================================

export type CampaignStatus = "active" | "paused" | "completed" | "draft" | "archived" | "pending_review";

export type CampaignObjective =
  | "awareness"
  | "traffic"
  | "engagement"
  | "leads"
  | "app_installs"
  | "conversions"
  | "sales"
  | "store_visits";

export interface Campaign {
  id: string;
  externalId: string;
  accountId: string;
  platform: Platform;
  name: string;
  status: CampaignStatus;
  objective?: CampaignObjective;
  budget: {
    amount: number;
    currency: string;
    type: "daily" | "lifetime";
  };
  schedule: {
    startDate: string;
    endDate?: string;
    timezone: string;
  };
  targeting?: {
    locations?: string[];
    ageRange?: { min: number; max: number };
    genders?: ("male" | "female" | "all")[];
    interests?: string[];
    customAudiences?: string[];
  };
  createdAt: string;
  updatedAt: string;
  // Metrics (optional, populated when fetched with metrics)
  metrics?: CampaignMetrics;
}

export interface CampaignMetrics {
  spend: number;
  impressions: number;
  clicks: number;
  conversions: number;
  revenue: number;
  roas: number;
  ctr: number;
  cpc: number;
  cpm: number;
  conversionRate: number;
  costPerConversion: number;
}

// ============================================
// Analytics & Metrics Types
// ============================================

export interface AdMetrics {
  spend: number;
  impressions: number;
  clicks: number;
  conversions: number;
  revenue: number;
  roas: number;
  ctr: number;
  cpc: number;
  cpm: number;
  reach?: number;
  frequency?: number;
  videoViews?: number;
  videoCompletions?: number;
}

export interface DashboardSummary {
  dateRange: DateRange;
  totals: AdMetrics;
  changes: {
    spend: number;
    impressions: number;
    clicks: number;
    conversions: number;
    revenue: number;
    roas: number;
  };
  platformBreakdown: PlatformMetrics[];
  topCampaigns: CampaignPerformance[];
  bottomCampaigns: CampaignPerformance[];
}

export interface PlatformMetrics {
  platform: Platform;
  metrics: AdMetrics;
  accountCount: number;
  campaignCount: number;
  change: number; // percentage change from previous period
}

export interface CampaignPerformance {
  id: string;
  name: string;
  platform: Platform;
  status: CampaignStatus;
  metrics: Pick<AdMetrics, "spend" | "roas" | "conversions" | "ctr">;
  change: number;
  trend: "up" | "down" | "stable";
}

export interface TimeSeriesDataPoint {
  date: string;
  metrics: AdMetrics;
  byPlatform?: Record<Platform, AdMetrics>;
}

export interface TimeSeriesData {
  dateRange: DateRange;
  granularity: "hour" | "day" | "week" | "month";
  data: TimeSeriesDataPoint[];
  totals: AdMetrics;
}

// ============================================
// Filter & Query Types
// ============================================

export interface CampaignFilters {
  platforms?: Platform[];
  statuses?: CampaignStatus[];
  accountIds?: string[];
  search?: string;
  dateRange?: DateRange;
  minSpend?: number;
  maxSpend?: number;
  minRoas?: number;
  sortBy?: keyof CampaignMetrics | "name" | "createdAt";
  sortOrder?: "asc" | "desc";
  page?: number;
  limit?: number;
}

export interface AnalyticsParams {
  dateRange: DateRange;
  compareRange?: DateRange;
  platforms?: Platform[];
  accountIds?: string[];
  campaignIds?: string[];
  granularity?: "hour" | "day" | "week" | "month";
  metrics?: (keyof AdMetrics)[];
  groupBy?: "platform" | "account" | "campaign" | "date";
}

// ============================================
// Auth Types
// ============================================

export interface LoginCredentials {
  email: string;
  password: string;
  rememberMe?: boolean;
}

export interface RegisterData {
  email: string;
  password: string;
  name: string;
  organizationName?: string;
}

export interface AuthTokens {
  accessToken: string;
  refreshToken: string;
  expiresIn: number;
  tokenType: "Bearer";
}

export interface AuthResponse {
  user: User;
  organization: Organization;
  tokens?: AuthTokens; // Optional since we use httpOnly cookies
  expiresAt?: string; // Token expiry time (sent for refresh scheduling)
}

export interface PasswordResetRequest {
  email: string;
}

export interface PasswordResetConfirm {
  token: string;
  password: string;
}

// ============================================
// Settings Types
// ============================================

export interface ProfileUpdateData {
  name?: string;
  avatarUrl?: string;
  timezone?: string;
}

export interface PasswordChangeData {
  currentPassword: string;
  newPassword: string;
}

export interface OrganizationUpdateData {
  name?: string;
  logoUrl?: string;
  settings?: Partial<Organization["settings"]>;
}

export interface NotificationSettings {
  email: {
    weeklyReport: boolean;
    campaignAlerts: boolean;
    syncErrors: boolean;
    budgetAlerts: boolean;
  };
  alertThresholds: {
    roasBelow: number;
    spendAbove: number;
    ctrBelow: number;
  };
}

// ============================================
// Billing Types
// ============================================

export interface BillingInfo {
  plan: Organization["plan"];
  status: "active" | "past_due" | "canceled" | "trialing";
  currentPeriodStart: string;
  currentPeriodEnd: string;
  cancelAtPeriodEnd: boolean;
  paymentMethod?: {
    type: "card" | "bank";
    last4: string;
    expiryMonth?: number;
    expiryYear?: number;
  };
}

export interface Invoice {
  id: string;
  number: string;
  amount: number;
  currency: string;
  status: "paid" | "open" | "void" | "uncollectible";
  dueDate: string;
  paidAt?: string;
  invoiceUrl: string;
}

export interface UsageStats {
  accounts: { used: number; limit: number };
  users: { used: number; limit: number };
  dataRetention: { current: number; limit: number };
  apiCalls: { used: number; limit: number };
}

// ============================================
// OAuth Types
// ============================================

export interface OAuthInitResponse {
  authUrl: string;
  state: string;
}

export interface OAuthCallbackData {
  platform: Platform;
  code: string;
  state: string;
}
