import { apiGet, apiPost, apiDelete } from "../client";
import {
  ConnectedAccount,
  SyncStatus,
  Platform,
  OAuthInitResponse,
  ApiResponse,
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

export interface ListConnectionsResponse {
  data: ConnectedAccount[];
  total: number;
}

export interface PlatformInfo {
  id: string;
  name: string;
  description: string;
  icon: string;
}

export const accountsApi = {
  /**
   * List all connected accounts
   * Backend: GET /connections
   */
  async listConnections(
    params?: ListConnectionsParams
  ): Promise<ListConnectionsResponse> {
    const response = await apiGet<ListConnectionsResponse>("/connections", {
      params,
    });
    return response;
  },

  /**
   * Get a single connected account
   * Backend: GET /connections/:id
   */
  async getConnection(accountId: string): Promise<ConnectedAccount> {
    const response = await apiGet<ApiResponse<ConnectedAccount>>(
      `/connections/${accountId}`
    );
    return response.data;
  },

  /**
   * Initiate OAuth connection for a platform
   * Backend: POST /connections/:platform/connect
   * Returns auth URL to redirect user to platform's OAuth page
   */
  async initiateConnection(
    platform: Platform,
    redirectUrl?: string
  ): Promise<OAuthInitResponse> {
    const params = redirectUrl
      ? `?redirect_url=${encodeURIComponent(redirectUrl)}`
      : "";
    const response = await apiPost<{ auth_url: string; platform: string }>(
      `/connections/${platform}/connect${params}`
    );
    return {
      authUrl: response.auth_url,
      state: "", // State is handled by backend
    };
  },

  /**
   * Disconnect an account
   * Backend: DELETE /connections/:id
   */
  async disconnect(accountId: string): Promise<void> {
    await apiDelete(`/connections/${accountId}`);
  },

  /**
   * Trigger sync for an account
   * Backend: POST /connections/:id/sync
   */
  async syncAccount(accountId: string): Promise<SyncStatus> {
    const response = await apiPost<ApiResponse<SyncStatus>>(
      `/connections/${accountId}/sync`
    );
    return response.data;
  },

  /**
   * Get sync status for an account
   * Backend: GET /connections/:id/sync-status
   */
  async getSyncStatus(accountId: string): Promise<SyncStatus> {
    const response = await apiGet<ApiResponse<SyncStatus>>(
      `/connections/${accountId}/sync-status`
    );
    return response.data;
  },

  /**
   * Reconnect an expired/errored account by initiating new OAuth flow
   */
  async reconnect(
    accountId: string,
    platform: Platform
  ): Promise<OAuthInitResponse> {
    // For reconnect, we initiate a new connection with the same platform
    return this.initiateConnection(
      platform,
      `/dashboard/connections?reconnect=${accountId}`
    );
  },

  /**
   * Get available platforms for connection
   * Backend: GET /connections/platforms
   */
  async getAvailablePlatforms(): Promise<PlatformInfo[]> {
    const response = await apiGet<{ data: PlatformInfo[] }>(
      "/connections/platforms"
    );
    return response.data;
  },
};

export default accountsApi;
