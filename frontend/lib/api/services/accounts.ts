import { apiGet, apiPost, apiDelete } from "../client";
import {
  ConnectedAccount,
  SyncStatus,
  Platform,
  OAuthInitResponse,
  OAuthCallbackData,
  ApiResponse,
  PaginatedResponse,
} from "../types";

// ============================================
// Accounts API Service
// ============================================

export interface ListConnectionsParams {
  platform?: Platform;
  status?: ConnectedAccount["status"];
  page?: number;
  limit?: number;
}

export const accountsApi = {
  /**
   * List all connected accounts
   */
  async listConnections(
    params?: ListConnectionsParams
  ): Promise<PaginatedResponse<ConnectedAccount>> {
    const response = await apiGet<PaginatedResponse<ConnectedAccount>>(
      "/accounts",
      { params }
    );
    return response;
  },

  /**
   * Get a single connected account
   */
  async getConnection(accountId: string): Promise<ConnectedAccount> {
    const response = await apiGet<ApiResponse<ConnectedAccount>>(
      `/accounts/${accountId}`
    );
    return response.data;
  },

  /**
   * Initiate OAuth connection for a platform
   */
  async initiateConnection(platform: Platform): Promise<OAuthInitResponse> {
    const response = await apiPost<ApiResponse<OAuthInitResponse>>(
      `/accounts/connect/${platform}`
    );
    return response.data;
  },

  /**
   * Complete OAuth callback
   */
  async completeConnection(data: OAuthCallbackData): Promise<ConnectedAccount> {
    const response = await apiPost<ApiResponse<ConnectedAccount>>(
      "/accounts/callback",
      data
    );
    return response.data;
  },

  /**
   * Disconnect an account
   */
  async disconnect(accountId: string): Promise<void> {
    await apiDelete(`/accounts/${accountId}`);
  },

  /**
   * Trigger sync for an account
   */
  async syncAccount(accountId: string): Promise<SyncStatus> {
    const response = await apiPost<ApiResponse<SyncStatus>>(
      `/accounts/${accountId}/sync`
    );
    return response.data;
  },

  /**
   * Trigger sync for all accounts
   */
  async syncAllAccounts(): Promise<SyncStatus[]> {
    const response = await apiPost<ApiResponse<SyncStatus[]>>("/accounts/sync-all");
    return response.data;
  },

  /**
   * Get sync status for an account
   */
  async getSyncStatus(accountId: string): Promise<SyncStatus> {
    const response = await apiGet<ApiResponse<SyncStatus>>(
      `/accounts/${accountId}/sync-status`
    );
    return response.data;
  },

  /**
   * Get sync status for all accounts
   */
  async getAllSyncStatus(): Promise<SyncStatus[]> {
    const response = await apiGet<ApiResponse<SyncStatus[]>>("/accounts/sync-status");
    return response.data;
  },

  /**
   * Reconnect an expired/errored account
   */
  async reconnect(accountId: string): Promise<OAuthInitResponse> {
    const response = await apiPost<ApiResponse<OAuthInitResponse>>(
      `/accounts/${accountId}/reconnect`
    );
    return response.data;
  },

  /**
   * Update account settings (sync frequency, etc.)
   */
  async updateSettings(
    accountId: string,
    settings: { syncFrequency?: number }
  ): Promise<ConnectedAccount> {
    const response = await apiPost<ApiResponse<ConnectedAccount>>(
      `/accounts/${accountId}/settings`,
      settings
    );
    return response.data;
  },

  /**
   * Get available platforms for connection
   */
  async getAvailablePlatforms(): Promise<Platform[]> {
    const response = await apiGet<ApiResponse<Platform[]>>("/accounts/platforms");
    return response.data;
  },
};

export default accountsApi;
