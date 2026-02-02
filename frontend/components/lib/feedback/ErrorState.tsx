"use client";

import { AlertCircle, RefreshCw, WifiOff, ServerCrash, ShieldX } from "lucide-react";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

type ErrorType = "generic" | "network" | "server" | "permission" | "notFound";

interface ErrorStateProps {
  type?: ErrorType;
  title?: string;
  message?: string;
  error?: Error;
  onRetry?: () => void;
  retryLabel?: string;
  isRetrying?: boolean;
  showDetails?: boolean;
  size?: "sm" | "md" | "lg";
  className?: string;
}

const errorConfig: Record<ErrorType, { icon: typeof AlertCircle; title: string; message: string }> = {
  generic: {
    icon: AlertCircle,
    title: "Something went wrong",
    message: "An unexpected error occurred. Please try again.",
  },
  network: {
    icon: WifiOff,
    title: "Connection error",
    message: "Unable to connect. Please check your internet connection and try again.",
  },
  server: {
    icon: ServerCrash,
    title: "Server error",
    message: "Our servers are having trouble. Please try again in a few moments.",
  },
  permission: {
    icon: ShieldX,
    title: "Access denied",
    message: "You don't have permission to access this resource.",
  },
  notFound: {
    icon: AlertCircle,
    title: "Not found",
    message: "The requested resource could not be found.",
  },
};

const sizeConfig = {
  sm: {
    container: "py-6",
    iconContainer: "h-10 w-10",
    icon: "h-5 w-5",
    title: "text-sm",
    message: "text-xs",
  },
  md: {
    container: "py-10",
    iconContainer: "h-14 w-14",
    icon: "h-7 w-7",
    title: "text-base",
    message: "text-sm",
  },
  lg: {
    container: "py-14",
    iconContainer: "h-16 w-16",
    icon: "h-8 w-8",
    title: "text-lg",
    message: "text-base",
  },
};

export function ErrorState({
  type = "generic",
  title,
  message,
  error,
  onRetry,
  retryLabel = "Try again",
  isRetrying = false,
  showDetails = false,
  size = "md",
  className,
}: ErrorStateProps) {
  const config = errorConfig[type];
  const Icon = config.icon;
  const displayTitle = title || config.title;
  const displayMessage = message || error?.message || config.message;
  const sizeStyles = sizeConfig[size];

  return (
    <div
      className={cn(
        "flex flex-col items-center justify-center text-center",
        sizeStyles.container,
        className
      )}
    >
      <div
        className={cn(
          "rounded-full bg-red-100 flex items-center justify-center mb-4",
          sizeStyles.iconContainer
        )}
      >
        <Icon className={cn("text-red-600", sizeStyles.icon)} />
      </div>

      <h3 className={cn("font-semibold text-slate-900 mb-2", sizeStyles.title)}>
        {displayTitle}
      </h3>

      <p className={cn("text-slate-500 max-w-md mb-4", sizeStyles.message)}>
        {displayMessage}
      </p>

      {showDetails && error && (
        <pre className="mb-4 p-3 bg-slate-100 rounded-lg text-xs text-slate-600 max-w-md overflow-auto text-left">
          {error.stack || error.message}
        </pre>
      )}

      {onRetry && (
        <Button
          onClick={onRetry}
          disabled={isRetrying}
          className="gap-2"
        >
          <RefreshCw className={cn("h-4 w-4", isRetrying && "animate-spin")} />
          {isRetrying ? "Retrying..." : retryLabel}
        </Button>
      )}
    </div>
  );
}

// Network error preset
export function NetworkError({
  onRetry,
  isRetrying,
  className,
}: {
  onRetry?: () => void;
  isRetrying?: boolean;
  className?: string;
}) {
  return (
    <ErrorState
      type="network"
      onRetry={onRetry}
      isRetrying={isRetrying}
      className={className}
    />
  );
}

// Server error preset
export function ServerError({
  onRetry,
  isRetrying,
  className,
}: {
  onRetry?: () => void;
  isRetrying?: boolean;
  className?: string;
}) {
  return (
    <ErrorState
      type="server"
      onRetry={onRetry}
      isRetrying={isRetrying}
      className={className}
    />
  );
}

// Inline error for forms
export function InlineError({
  message,
  className,
}: {
  message: string;
  className?: string;
}) {
  return (
    <div className={cn("flex items-center gap-2 text-red-600 text-sm", className)}>
      <AlertCircle className="h-4 w-4" />
      <span>{message}</span>
    </div>
  );
}
