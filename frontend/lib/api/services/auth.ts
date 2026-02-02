import { apiGet, apiPost, tokenManager, setRefreshTokenFunction } from "../client";
import {
  AuthResponse,
  LoginCredentials,
  RegisterData,
  PasswordResetRequest,
  PasswordResetConfirm,
  User,
  ApiResponse,
} from "../types";

// ============================================
// Session Response Type
// ============================================

export interface SessionResponse {
  authenticated: boolean;
  user?: {
    id: string;
    email: string;
    role: string;
  };
  organizationId?: string;
  permissions?: string[];
}

// ============================================
// Auth API Service (Cookie-based)
// ============================================

export const authApi = {
  /**
   * Login with email and password
   * Tokens are stored in httpOnly cookies by the backend
   */
  async login(credentials: LoginCredentials): Promise<AuthResponse> {
    const response = await apiPost<ApiResponse<AuthResponse>>("/auth/login", credentials);

    // Store token expiry info for refresh timing (not the token itself)
    if (response.data.expiresAt) {
      tokenManager.setTokenExpiry(new Date(response.data.expiresAt).getTime());
    }

    return response.data;
  },

  /**
   * Register a new user
   * Tokens are stored in httpOnly cookies by the backend
   */
  async register(data: RegisterData): Promise<AuthResponse> {
    const response = await apiPost<ApiResponse<AuthResponse>>("/auth/register", data);

    // Store token expiry info for refresh timing
    if (response.data.expiresAt) {
      tokenManager.setTokenExpiry(new Date(response.data.expiresAt).getTime());
    }

    return response.data;
  },

  /**
   * Logout current user
   * Backend will clear httpOnly cookies
   */
  async logout(): Promise<void> {
    try {
      await apiPost("/auth/logout");
    } finally {
      // Clear local state
      tokenManager.clearTokens();
    }
  },

  /**
   * Refresh access token
   * Refresh token is sent via cookie automatically
   */
  async refreshToken(): Promise<{ expiresAt: string }> {
    const response = await apiPost<ApiResponse<{ expiresAt: string }>>("/auth/refresh");

    // Update token expiry info
    if (response.data.expiresAt) {
      tokenManager.setTokenExpiry(new Date(response.data.expiresAt).getTime());
    }

    return response.data;
  },

  /**
   * Get current session status
   * Uses optionally authenticated endpoint to check if user is logged in
   */
  async getSession(): Promise<SessionResponse> {
    const response = await apiGet<ApiResponse<SessionResponse>>("/auth/session");
    return response.data;
  },

  /**
   * Request password reset email
   */
  async forgotPassword(data: PasswordResetRequest): Promise<void> {
    await apiPost("/auth/forgot-password", data);
  },

  /**
   * Reset password with token
   */
  async resetPassword(data: PasswordResetConfirm): Promise<void> {
    await apiPost("/auth/reset-password", data);
  },

  /**
   * Get current user profile
   */
  async getCurrentUser(): Promise<User> {
    const response = await apiGet<ApiResponse<User>>("/user/me");
    return response.data;
  },

  /**
   * Verify email with token
   */
  async verifyEmail(token: string): Promise<void> {
    await apiPost("/auth/verify-email", { token });
  },

  /**
   * Resend verification email
   */
  async resendVerificationEmail(): Promise<void> {
    await apiPost("/auth/resend-verification");
  },

  /**
   * Check if user might be authenticated (local check based on expiry)
   * For definitive check, use getSession()
   */
  isAuthenticated(): boolean {
    return !tokenManager.isTokenExpired();
  },
};

// Set up refresh token function for the API client
setRefreshTokenFunction(authApi.refreshToken);

export default authApi;
