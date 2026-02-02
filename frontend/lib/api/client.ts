import axios, {
  AxiosError,
  AxiosInstance,
  AxiosRequestConfig,
  InternalAxiosRequestConfig,
} from "axios";
import { ApiError, ApiErrorResponse, ErrorCode, isApiError } from "./errors";

// ============================================
// Configuration
// ============================================

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api/v1";
const TOKEN_EXPIRY_KEY = "auth_token_expiry";
const REFRESH_THRESHOLD_MS = 5 * 60 * 1000; // 5 minutes before expiry

// ============================================
// Token Management (Cookie-based)
// ============================================

/**
 * TokenManager for cookie-based authentication
 * Tokens are stored in httpOnly cookies, but we track expiry time locally
 * to know when to trigger refresh
 */
class TokenManager {
  private static instance: TokenManager;
  private tokenExpiry: number | null = null;
  private refreshPromise: Promise<unknown> | null = null;

  static getInstance(): TokenManager {
    if (!TokenManager.instance) {
      TokenManager.instance = new TokenManager();
    }
    return TokenManager.instance;
  }

  initialize(): void {
    if (typeof window === "undefined") return;

    try {
      const stored = localStorage.getItem(TOKEN_EXPIRY_KEY);
      if (stored) {
        this.tokenExpiry = parseInt(stored, 10);
      }
    } catch (error) {
      console.error("Failed to load token expiry from storage:", error);
      this.clearTokens();
    }
  }

  /**
   * Check if token is expired
   * Note: With httpOnly cookies, we can only estimate based on stored expiry
   */
  isTokenExpired(): boolean {
    if (!this.tokenExpiry) return true;
    return Date.now() >= this.tokenExpiry;
  }

  /**
   * Check if token is expiring soon (within 5 minutes)
   */
  isTokenExpiringSoon(): boolean {
    if (!this.tokenExpiry) return false; // Don't refresh if we don't know expiry
    return Date.now() >= this.tokenExpiry - REFRESH_THRESHOLD_MS;
  }

  /**
   * Set token expiry time (tokens themselves are in httpOnly cookies)
   */
  setTokenExpiry(expiresAt: number): void {
    this.tokenExpiry = expiresAt;

    if (typeof window !== "undefined") {
      localStorage.setItem(TOKEN_EXPIRY_KEY, expiresAt.toString());
    }
  }

  /**
   * Clear local token state
   */
  clearTokens(): void {
    this.tokenExpiry = null;
    this.refreshPromise = null;

    if (typeof window !== "undefined") {
      localStorage.removeItem(TOKEN_EXPIRY_KEY);
    }
  }

  /**
   * Refresh tokens with deduplication
   */
  async refreshTokens<T>(refreshFn: () => Promise<T>): Promise<T> {
    // If already refreshing, return the existing promise
    if (this.refreshPromise) {
      return this.refreshPromise as Promise<T>;
    }

    // Start new refresh
    this.refreshPromise = refreshFn()
      .then((result) => {
        this.refreshPromise = null;
        return result;
      })
      .catch((error) => {
        this.refreshPromise = null;
        this.clearTokens();
        throw error;
      });

    return this.refreshPromise as Promise<T>;
  }

  /**
   * Get token expiry timestamp
   */
  getTokenExpiry(): number | null {
    return this.tokenExpiry;
  }
}

// Export singleton instance
export const tokenManager = TokenManager.getInstance();

// ============================================
// API Client Factory
// ============================================

interface ApiClientOptions {
  baseURL?: string;
  timeout?: number;
  onUnauthorized?: () => void;
  onForbidden?: () => void;
  onRateLimited?: (retryAfter: number) => void;
  onServerError?: (error: ApiError) => void;
  onPlatformError?: (platform: string, error: ApiError) => void;
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
let refreshTokenFunction: (() => Promise<any>) | null = null;

// eslint-disable-next-line @typescript-eslint/no-explicit-any
export function setRefreshTokenFunction(fn: () => Promise<any>): void {
  refreshTokenFunction = fn;
}

function createApiClient(options: ApiClientOptions = {}): AxiosInstance {
  const client = axios.create({
    baseURL: options.baseURL || API_BASE_URL,
    timeout: options.timeout || 30000,
    headers: {
      "Content-Type": "application/json",
    },
    withCredentials: true, // Send/receive httpOnly cookies
  });

  // Request Interceptor
  client.interceptors.request.use(
    async (config: InternalAxiosRequestConfig) => {
      // Skip refresh for public endpoints and refresh endpoint itself
      const skipRefreshEndpoints = [
        "/auth/login",
        "/auth/register",
        "/auth/forgot-password",
        "/auth/reset-password",
        "/auth/refresh",
        "/auth/session",
      ];
      const isSkipRefreshEndpoint = skipRefreshEndpoints.some((endpoint) =>
        config.url?.includes(endpoint)
      );

      // Proactively refresh token if expiring soon (cookies sent automatically)
      if (!isSkipRefreshEndpoint && tokenManager.isTokenExpiringSoon() && refreshTokenFunction) {
        try {
          await tokenManager.refreshTokens(refreshTokenFunction);
        } catch (error) {
          // Token refresh failed, request will proceed and backend will decide
          console.warn("Token refresh failed:", error);
        }
      }

      // Cookies are sent automatically via withCredentials: true
      return config;
    },
    (error) => Promise.reject(error)
  );

  // Response Interceptor
  client.interceptors.response.use(
    (response) => response,
    async (error: AxiosError<ApiErrorResponse>) => {
      const originalRequest = error.config as InternalAxiosRequestConfig & {
        _retry?: boolean;
      };
      const status = error.response?.status || 0;

      // Parse API error from response
      const parseApiError = (): ApiError => {
        if (error.response?.data?.error) {
          return ApiError.fromResponse(error.response.data, status);
        }
        // Network or other errors
        return new ApiError(
          error.message || "A network error occurred",
          ErrorCode.INTERNAL_ERROR,
          status
        );
      };

      // Handle 401 Unauthorized
      if (status === 401 && !originalRequest._retry) {
        originalRequest._retry = true;

        // Try to refresh token (cookies sent automatically)
        if (refreshTokenFunction) {
          try {
            await tokenManager.refreshTokens(refreshTokenFunction);
            // Retry original request (new cookies will be sent automatically)
            return client(originalRequest);
          } catch (refreshError) {
            // Refresh failed, redirect to login
            tokenManager.clearTokens();
            options.onUnauthorized?.();
            return Promise.reject(parseApiError());
          }
        }

        // No refresh function, redirect to login
        tokenManager.clearTokens();
        options.onUnauthorized?.();
        return Promise.reject(parseApiError());
      }

      const apiError = parseApiError();

      // Handle 403 Forbidden
      if (status === 403) {
        options.onForbidden?.();
      }

      // Handle 429 Rate Limited
      if (status === 429 && apiError.retryAfter) {
        options.onRateLimited?.(apiError.retryAfter);
      }

      // Handle 500+ server errors
      if (status >= 500) {
        options.onServerError?.(apiError);
      }

      // Handle platform errors
      if (apiError.isPlatformError() && apiError.platform) {
        options.onPlatformError?.(apiError.platform, apiError);
      }

      return Promise.reject(apiError);
    }
  );

  return client;
}

// ============================================
// Default API Client
// ============================================

let defaultClient: AxiosInstance | null = null;
let clientOptions: ApiClientOptions = {};

export function configureApiClient(options: ApiClientOptions): void {
  clientOptions = options;
  defaultClient = null; // Reset client to pick up new options
}

export function getApiClient(): AxiosInstance {
  if (!defaultClient) {
    defaultClient = createApiClient(clientOptions);
  }
  return defaultClient;
}

// ============================================
// Helper Functions
// ============================================

export async function apiGet<T>(
  url: string,
  config?: AxiosRequestConfig
): Promise<T> {
  const response = await getApiClient().get<T>(url, config);
  return response.data;
}

export async function apiPost<T, D = unknown>(
  url: string,
  data?: D,
  config?: AxiosRequestConfig
): Promise<T> {
  const response = await getApiClient().post<T>(url, data, config);
  return response.data;
}

export async function apiPut<T, D = unknown>(
  url: string,
  data?: D,
  config?: AxiosRequestConfig
): Promise<T> {
  const response = await getApiClient().put<T>(url, data, config);
  return response.data;
}

export async function apiPatch<T, D = unknown>(
  url: string,
  data?: D,
  config?: AxiosRequestConfig
): Promise<T> {
  const response = await getApiClient().patch<T>(url, data, config);
  return response.data;
}

export async function apiDelete<T>(
  url: string,
  config?: AxiosRequestConfig
): Promise<T> {
  const response = await getApiClient().delete<T>(url, config);
  return response.data;
}

// ============================================
// Initialize
// ============================================

// Initialize token manager on client side
if (typeof window !== "undefined") {
  tokenManager.initialize();
}
