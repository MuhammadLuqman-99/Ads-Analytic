import { type Platform } from "@/lib/mock-data";

export type ConnectionStatus = "active" | "error" | "expired" | "syncing";

export interface AdAccount {
  id: string;
  platform: Platform;
  accountId: string;
  accountName: string;
  status: ConnectionStatus;
  lastSyncAt: Date | null;
  connectedAt: Date;
  error?: {
    type: "token_expired" | "permission_denied" | "api_error" | "rate_limit";
    message: string;
    occurredAt: Date;
  };
}

export interface SyncHistory {
  id: string;
  accountId: string;
  status: "success" | "failed" | "partial";
  startedAt: Date;
  completedAt: Date;
  recordsSynced?: number;
  error?: string;
}

export interface PlanLimits {
  currentPlan: "free" | "pro" | "business";
  accountsUsed: number;
  accountsLimit: number;
}

// Mock data
export const mockAdAccounts: AdAccount[] = [
  {
    id: "1",
    platform: "meta",
    accountId: "act_123456789",
    accountName: "MyBusiness Meta Ads",
    status: "active",
    lastSyncAt: new Date(Date.now() - 1000 * 60 * 15), // 15 min ago
    connectedAt: new Date(Date.now() - 1000 * 60 * 60 * 24 * 30), // 30 days ago
  },
  {
    id: "2",
    platform: "tiktok",
    accountId: "7123456789012345678",
    accountName: "TikTok Business Account",
    status: "active",
    lastSyncAt: new Date(Date.now() - 1000 * 60 * 30), // 30 min ago
    connectedAt: new Date(Date.now() - 1000 * 60 * 60 * 24 * 14), // 14 days ago
  },
  {
    id: "3",
    platform: "shopee",
    accountId: "shop_987654321",
    accountName: "Shopee Seller Center",
    status: "expired",
    lastSyncAt: new Date(Date.now() - 1000 * 60 * 60 * 24 * 2), // 2 days ago
    connectedAt: new Date(Date.now() - 1000 * 60 * 60 * 24 * 60), // 60 days ago
    error: {
      type: "token_expired",
      message: "Access token has expired. Please reconnect your account.",
      occurredAt: new Date(Date.now() - 1000 * 60 * 60 * 24 * 2),
    },
  },
];

export const mockSyncHistory: SyncHistory[] = [
  {
    id: "s1",
    accountId: "1",
    status: "success",
    startedAt: new Date(Date.now() - 1000 * 60 * 15),
    completedAt: new Date(Date.now() - 1000 * 60 * 14),
    recordsSynced: 1250,
  },
  {
    id: "s2",
    accountId: "1",
    status: "success",
    startedAt: new Date(Date.now() - 1000 * 60 * 60 * 2),
    completedAt: new Date(Date.now() - 1000 * 60 * 60 * 2 + 60000),
    recordsSynced: 1180,
  },
  {
    id: "s3",
    accountId: "1",
    status: "failed",
    startedAt: new Date(Date.now() - 1000 * 60 * 60 * 5),
    completedAt: new Date(Date.now() - 1000 * 60 * 60 * 5 + 30000),
    error: "Rate limit exceeded. Retry scheduled.",
  },
  {
    id: "s4",
    accountId: "2",
    status: "success",
    startedAt: new Date(Date.now() - 1000 * 60 * 30),
    completedAt: new Date(Date.now() - 1000 * 60 * 29),
    recordsSynced: 890,
  },
  {
    id: "s5",
    accountId: "3",
    status: "failed",
    startedAt: new Date(Date.now() - 1000 * 60 * 60 * 48),
    completedAt: new Date(Date.now() - 1000 * 60 * 60 * 48 + 10000),
    error: "Token expired",
  },
];

export const mockPlanLimits: PlanLimits = {
  currentPlan: "pro",
  accountsUsed: 3,
  accountsLimit: 5,
};
