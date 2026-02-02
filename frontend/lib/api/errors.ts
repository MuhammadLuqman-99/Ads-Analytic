// API Error codes matching backend
export const ErrorCode = {
  // General errors
  INTERNAL_ERROR: "INTERNAL_ERROR",
  VALIDATION_ERROR: "VALIDATION_ERROR",
  NOT_FOUND: "NOT_FOUND",
  UNAUTHORIZED: "UNAUTHORIZED",
  FORBIDDEN: "FORBIDDEN",
  CONFLICT: "CONFLICT",
  BAD_REQUEST: "BAD_REQUEST",

  // Platform errors
  PLATFORM_ERROR: "PLATFORM_ERROR",
  RATE_LIMITED: "RATE_LIMITED",
  TOKEN_EXPIRED: "TOKEN_EXPIRED",
  TOKEN_INVALID: "TOKEN_INVALID",
  OAUTH_FAILED: "OAUTH_FAILED",
  PLATFORM_TIMEOUT: "PLATFORM_TIMEOUT",
  PLATFORM_UNAVAILABLE: "PLATFORM_UNAVAILABLE",

  // Sync errors
  SYNC_FAILED: "SYNC_FAILED",
  PARTIAL_SYNC: "PARTIAL_SYNC",
  SYNC_CONFLICT: "SYNC_CONFLICT",

  // Subscription errors
  SUBSCRIPTION_LIMIT: "SUBSCRIPTION_LIMIT",
  SUBSCRIPTION_EXPIRED: "SUBSCRIPTION_EXPIRED",
} as const;

export type ErrorCodeType = (typeof ErrorCode)[keyof typeof ErrorCode];

// Field-level validation error
export interface FieldError {
  field: string;
  message: string;
}

// API error detail structure
export interface ApiErrorDetail {
  code: ErrorCodeType;
  message: string;
  details?: FieldError[];
  retry_after?: number; // seconds
  platform?: string;
}

// API error response structure
export interface ApiErrorResponse {
  success: false;
  error: ApiErrorDetail;
}

// Custom API Error class
export class ApiError extends Error {
  public readonly code: ErrorCodeType;
  public readonly status: number;
  public readonly details?: FieldError[];
  public readonly retryAfter?: number;
  public readonly platform?: string;

  constructor(
    message: string,
    code: ErrorCodeType,
    status: number,
    options?: {
      details?: FieldError[];
      retryAfter?: number;
      platform?: string;
    }
  ) {
    super(message);
    this.name = "ApiError";
    this.code = code;
    this.status = status;
    this.details = options?.details;
    this.retryAfter = options?.retryAfter;
    this.platform = options?.platform;
  }

  // Check if error is due to authentication
  isAuthError(): boolean {
    return this.code === ErrorCode.UNAUTHORIZED || this.code === ErrorCode.TOKEN_EXPIRED;
  }

  // Check if error is due to permissions
  isForbidden(): boolean {
    return this.code === ErrorCode.FORBIDDEN;
  }

  // Check if error is rate limited
  isRateLimited(): boolean {
    return this.code === ErrorCode.RATE_LIMITED;
  }

  // Check if error is a validation error
  isValidationError(): boolean {
    return this.code === ErrorCode.VALIDATION_ERROR;
  }

  // Check if error is a platform error
  isPlatformError(): boolean {
    return (
      this.code === ErrorCode.PLATFORM_ERROR ||
      this.code === ErrorCode.PLATFORM_TIMEOUT ||
      this.code === ErrorCode.PLATFORM_UNAVAILABLE
    );
  }

  // Check if error is due to subscription limits
  isSubscriptionError(): boolean {
    return (
      this.code === ErrorCode.SUBSCRIPTION_LIMIT ||
      this.code === ErrorCode.SUBSCRIPTION_EXPIRED
    );
  }

  // Check if error is retryable
  isRetryable(): boolean {
    return (
      this.status >= 500 ||
      this.code === ErrorCode.RATE_LIMITED ||
      this.code === ErrorCode.PLATFORM_TIMEOUT
    );
  }

  // Get user-friendly message
  getUserMessage(): string {
    switch (this.code) {
      case ErrorCode.UNAUTHORIZED:
        return "Your session has expired. Please log in again.";
      case ErrorCode.FORBIDDEN:
        return "You don't have permission to perform this action.";
      case ErrorCode.RATE_LIMITED:
        return `Too many requests. Please try again ${
          this.retryAfter ? `in ${this.retryAfter} seconds` : "later"
        }.`;
      case ErrorCode.PLATFORM_ERROR:
        return `${this.platform || "Platform"} API error: ${this.message}`;
      case ErrorCode.PLATFORM_TIMEOUT:
        return `${this.platform || "Platform"} is taking too long to respond. Please try again.`;
      case ErrorCode.PLATFORM_UNAVAILABLE:
        return `${this.platform || "Platform"} is currently unavailable. Please try again later.`;
      case ErrorCode.SUBSCRIPTION_LIMIT:
        return "You've reached your subscription limit. Please upgrade your plan.";
      case ErrorCode.SUBSCRIPTION_EXPIRED:
        return "Your subscription has expired. Please renew to continue.";
      case ErrorCode.TOKEN_EXPIRED:
        return `Your ${this.platform || ""} connection has expired. Please reconnect.`;
      case ErrorCode.VALIDATION_ERROR:
        if (this.details && this.details.length > 0) {
          return this.details.map((d) => `${d.field}: ${d.message}`).join(", ");
        }
        return this.message;
      case ErrorCode.NOT_FOUND:
        return "The requested resource was not found.";
      default:
        return this.message || "An unexpected error occurred.";
    }
  }

  // Get field errors for forms
  getFieldErrors(): Record<string, string> {
    const errors: Record<string, string> = {};
    if (this.details) {
      this.details.forEach((d) => {
        errors[d.field] = d.message;
      });
    }
    return errors;
  }

  // Create from API response
  static fromResponse(response: ApiErrorResponse, status: number): ApiError {
    const { error } = response;
    return new ApiError(error.message, error.code, status, {
      details: error.details,
      retryAfter: error.retry_after,
      platform: error.platform,
    });
  }
}

// Helper to check if an error is an ApiError
export function isApiError(error: unknown): error is ApiError {
  return error instanceof ApiError;
}

// Helper to extract error message from any error
export function getErrorMessage(error: unknown): string {
  if (isApiError(error)) {
    return error.getUserMessage();
  }
  if (error instanceof Error) {
    return error.message;
  }
  return "An unexpected error occurred";
}
