import { apiGet, apiPost, apiPut, apiDelete } from "../client";
import {
  User,
  Organization,
  TeamMember,
  ProfileUpdateData,
  PasswordChangeData,
  OrganizationUpdateData,
  NotificationSettings,
  BillingInfo,
  Invoice,
  UsageStats,
  ApiResponse,
  PaginatedResponse,
  UserRole,
} from "../types";

// ============================================
// Settings API Service
// ============================================

export const settingsApi = {
  // ==========================================
  // Profile Settings
  // ==========================================

  /**
   * Get current user profile
   */
  async getProfile(): Promise<User> {
    const response = await apiGet<ApiResponse<User>>("/settings/profile");
    return response.data;
  },

  /**
   * Update user profile
   */
  async updateProfile(data: ProfileUpdateData): Promise<User> {
    const response = await apiPut<ApiResponse<User>>("/settings/profile", data);
    return response.data;
  },

  /**
   * Change password
   */
  async changePassword(data: PasswordChangeData): Promise<void> {
    await apiPost("/settings/profile/password", data);
  },

  /**
   * Upload avatar
   */
  async uploadAvatar(file: File): Promise<{ url: string }> {
    const formData = new FormData();
    formData.append("avatar", file);

    const response = await apiPost<ApiResponse<{ url: string }>>(
      "/settings/profile/avatar",
      formData,
      {
        headers: {
          "Content-Type": "multipart/form-data",
        },
      }
    );
    return response.data;
  },

  /**
   * Delete user account
   */
  async deleteAccount(password: string): Promise<void> {
    await apiDelete("/settings/profile", {
      data: { password },
    });
  },

  /**
   * Enable/disable two-factor authentication
   */
  async setupTwoFactor(): Promise<{ qrCode: string; secret: string }> {
    const response = await apiPost<ApiResponse<{ qrCode: string; secret: string }>>(
      "/settings/profile/2fa/setup"
    );
    return response.data;
  },

  async enableTwoFactor(code: string): Promise<{ backupCodes: string[] }> {
    const response = await apiPost<ApiResponse<{ backupCodes: string[] }>>(
      "/settings/profile/2fa/enable",
      { code }
    );
    return response.data;
  },

  async disableTwoFactor(code: string): Promise<void> {
    await apiPost("/settings/profile/2fa/disable", { code });
  },

  // ==========================================
  // Organization Settings
  // ==========================================

  /**
   * Get organization details
   */
  async getOrganization(): Promise<Organization> {
    const response = await apiGet<ApiResponse<Organization>>("/settings/organization");
    return response.data;
  },

  /**
   * Update organization
   */
  async updateOrganization(data: OrganizationUpdateData): Promise<Organization> {
    const response = await apiPut<ApiResponse<Organization>>(
      "/settings/organization",
      data
    );
    return response.data;
  },

  /**
   * Upload organization logo
   */
  async uploadLogo(file: File): Promise<{ url: string }> {
    const formData = new FormData();
    formData.append("logo", file);

    const response = await apiPost<ApiResponse<{ url: string }>>(
      "/settings/organization/logo",
      formData,
      {
        headers: {
          "Content-Type": "multipart/form-data",
        },
      }
    );
    return response.data;
  },

  // ==========================================
  // Team Members
  // ==========================================

  /**
   * List team members
   */
  async listTeamMembers(params?: {
    page?: number;
    limit?: number;
    status?: TeamMember["status"];
  }): Promise<PaginatedResponse<TeamMember>> {
    const response = await apiGet<PaginatedResponse<TeamMember>>(
      "/settings/team",
      { params }
    );
    return response;
  },

  /**
   * Invite team member
   */
  async inviteTeamMember(data: {
    email: string;
    role: UserRole;
    message?: string;
  }): Promise<TeamMember> {
    const response = await apiPost<ApiResponse<TeamMember>>(
      "/settings/team/invite",
      data
    );
    return response.data;
  },

  /**
   * Resend invitation
   */
  async resendInvitation(memberId: string): Promise<void> {
    await apiPost(`/settings/team/${memberId}/resend-invite`);
  },

  /**
   * Update team member role
   */
  async updateMemberRole(memberId: string, role: UserRole): Promise<TeamMember> {
    const response = await apiPut<ApiResponse<TeamMember>>(
      `/settings/team/${memberId}`,
      { role }
    );
    return response.data;
  },

  /**
   * Remove team member
   */
  async removeMember(memberId: string): Promise<void> {
    await apiDelete(`/settings/team/${memberId}`);
  },

  /**
   * Cancel pending invitation
   */
  async cancelInvitation(memberId: string): Promise<void> {
    await apiDelete(`/settings/team/${memberId}/invitation`);
  },

  // ==========================================
  // Notification Settings
  // ==========================================

  /**
   * Get notification settings
   */
  async getNotificationSettings(): Promise<NotificationSettings> {
    const response = await apiGet<ApiResponse<NotificationSettings>>(
      "/settings/notifications"
    );
    return response.data;
  },

  /**
   * Update notification settings
   */
  async updateNotificationSettings(
    data: Partial<NotificationSettings>
  ): Promise<NotificationSettings> {
    const response = await apiPut<ApiResponse<NotificationSettings>>(
      "/settings/notifications",
      data
    );
    return response.data;
  },

  /**
   * Test notification (email)
   */
  async testNotification(type: "email"): Promise<void> {
    await apiPost("/settings/notifications/test", { type });
  },

  // ==========================================
  // Billing Settings
  // ==========================================

  /**
   * Get billing information
   */
  async getBilling(): Promise<BillingInfo> {
    const response = await apiGet<ApiResponse<BillingInfo>>("/settings/billing");
    return response.data;
  },

  /**
   * Get usage statistics
   */
  async getUsage(): Promise<UsageStats> {
    const response = await apiGet<ApiResponse<UsageStats>>("/settings/billing/usage");
    return response.data;
  },

  /**
   * List invoices
   */
  async listInvoices(params?: {
    page?: number;
    limit?: number;
  }): Promise<PaginatedResponse<Invoice>> {
    const response = await apiGet<PaginatedResponse<Invoice>>(
      "/settings/billing/invoices",
      { params }
    );
    return response;
  },

  /**
   * Download invoice
   */
  async downloadInvoice(invoiceId: string): Promise<Blob> {
    const response = await apiGet<Blob>(
      `/settings/billing/invoices/${invoiceId}/download`,
      { responseType: "blob" }
    );
    return response;
  },

  /**
   * Get checkout session for plan upgrade
   */
  async createCheckoutSession(plan: Organization["plan"]): Promise<{ url: string }> {
    const response = await apiPost<ApiResponse<{ url: string }>>(
      "/settings/billing/checkout",
      { plan }
    );
    return response.data;
  },

  /**
   * Create portal session for billing management
   */
  async createPortalSession(): Promise<{ url: string }> {
    const response = await apiPost<ApiResponse<{ url: string }>>(
      "/settings/billing/portal"
    );
    return response.data;
  },

  /**
   * Cancel subscription
   */
  async cancelSubscription(feedback?: string): Promise<void> {
    await apiPost("/settings/billing/cancel", { feedback });
  },

  /**
   * Resume canceled subscription
   */
  async resumeSubscription(): Promise<BillingInfo> {
    const response = await apiPost<ApiResponse<BillingInfo>>(
      "/settings/billing/resume"
    );
    return response.data;
  },

  // ==========================================
  // API Keys (for integrations)
  // ==========================================

  /**
   * List API keys
   */
  async listApiKeys(): Promise<
    { id: string; name: string; lastUsed?: string; createdAt: string }[]
  > {
    const response = await apiGet<
      ApiResponse<{ id: string; name: string; lastUsed?: string; createdAt: string }[]>
    >("/settings/api-keys");
    return response.data;
  },

  /**
   * Create API key
   */
  async createApiKey(name: string): Promise<{
    id: string;
    name: string;
    key: string; // Only returned on creation
    createdAt: string;
  }> {
    const response = await apiPost<
      ApiResponse<{
        id: string;
        name: string;
        key: string;
        createdAt: string;
      }>
    >("/settings/api-keys", { name });
    return response.data;
  },

  /**
   * Delete API key
   */
  async deleteApiKey(keyId: string): Promise<void> {
    await apiDelete(`/settings/api-keys/${keyId}`);
  },
};

export default settingsApi;
