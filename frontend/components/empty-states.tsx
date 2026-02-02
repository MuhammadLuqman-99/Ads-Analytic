"use client";

import { ReactNode } from "react";
import Link from "next/link";
import {
  BarChart3,
  Calendar,
  FolderOpen,
  Link2,
  Search,
  ServerOff,
  ShieldAlert,
  Wifi,
  AlertTriangle,
  RefreshCw,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";

interface EmptyStateProps {
  icon?: ReactNode;
  title: string;
  description: string;
  action?: {
    label: string;
    href?: string;
    onClick?: () => void;
  };
  secondaryAction?: {
    label: string;
    href?: string;
    onClick?: () => void;
  };
}

// Base Empty State Component
export function EmptyState({
  icon,
  title,
  description,
  action,
  secondaryAction,
}: EmptyStateProps) {
  return (
    <Card className="border-dashed">
      <CardContent className="flex flex-col items-center justify-center py-12 text-center">
        {icon && (
          <div className="mb-4 rounded-full bg-slate-100 p-3">
            {icon}
          </div>
        )}
        <h3 className="text-lg font-semibold text-slate-900 mb-2">{title}</h3>
        <p className="text-sm text-slate-500 max-w-sm mb-6">{description}</p>
        <div className="flex gap-3">
          {secondaryAction && (
            secondaryAction.href ? (
              <Link href={secondaryAction.href}>
                <Button variant="outline">{secondaryAction.label}</Button>
              </Link>
            ) : (
              <Button variant="outline" onClick={secondaryAction.onClick}>
                {secondaryAction.label}
              </Button>
            )
          )}
          {action && (
            action.href ? (
              <Link href={action.href}>
                <Button>{action.label}</Button>
              </Link>
            ) : (
              <Button onClick={action.onClick}>{action.label}</Button>
            )
          )}
        </div>
      </CardContent>
    </Card>
  );
}

// No Data for Date Range
export function NoDataForPeriod({
  onChangeDateRange,
}: {
  onChangeDateRange?: () => void;
}) {
  return (
    <EmptyState
      icon={<Calendar className="h-6 w-6 text-slate-400" />}
      title="No data for this period"
      description="There's no data available for the selected date range. Try selecting a different time period."
      action={
        onChangeDateRange
          ? { label: "Change Date Range", onClick: onChangeDateRange }
          : undefined
      }
    />
  );
}

// No Campaigns Found
export function NoCampaignsFound({
  hasFilters,
  onClearFilters,
}: {
  hasFilters?: boolean;
  onClearFilters?: () => void;
}) {
  return (
    <EmptyState
      icon={<FolderOpen className="h-6 w-6 text-slate-400" />}
      title="No campaigns found"
      description={
        hasFilters
          ? "No campaigns match your current filters. Try adjusting your search criteria."
          : "You don't have any campaigns yet. Connect an ad account to start tracking your campaigns."
      }
      action={
        hasFilters && onClearFilters
          ? { label: "Clear Filters", onClick: onClearFilters }
          : { label: "Connect Account", href: "/dashboard/connections" }
      }
    />
  );
}

// No Accounts Connected
export function NoAccountsConnected() {
  return (
    <EmptyState
      icon={<Link2 className="h-6 w-6 text-slate-400" />}
      title="No accounts connected"
      description="Connect your ad platform accounts to start tracking performance and syncing data."
      action={{ label: "Connect Account", href: "/dashboard/connections" }}
    />
  );
}

// Search No Results
export function SearchNoResults({
  query,
  onClear,
}: {
  query: string;
  onClear?: () => void;
}) {
  return (
    <EmptyState
      icon={<Search className="h-6 w-6 text-slate-400" />}
      title="No results found"
      description={`We couldn't find anything matching "${query}". Try a different search term.`}
      action={onClear ? { label: "Clear Search", onClick: onClear } : undefined}
    />
  );
}

// No Analytics Data
export function NoAnalyticsData() {
  return (
    <EmptyState
      icon={<BarChart3 className="h-6 w-6 text-slate-400" />}
      title="No analytics data yet"
      description="Once your connected accounts sync, you'll see your performance metrics and analytics here."
      action={{ label: "View Connections", href: "/dashboard/connections" }}
    />
  );
}

// Error States

// Network Error
export function NetworkError({ onRetry }: { onRetry?: () => void }) {
  return (
    <EmptyState
      icon={<Wifi className="h-6 w-6 text-red-400" />}
      title="Connection error"
      description="Unable to connect to the server. Please check your internet connection and try again."
      action={onRetry ? { label: "Try Again", onClick: onRetry } : undefined}
    />
  );
}

// Server Error
export function ServerError({ onRetry }: { onRetry?: () => void }) {
  return (
    <EmptyState
      icon={<ServerOff className="h-6 w-6 text-red-400" />}
      title="Server error"
      description="Something went wrong on our end. Please try again later or contact support if the problem persists."
      action={onRetry ? { label: "Retry", onClick: onRetry } : undefined}
      secondaryAction={{ label: "Contact Support", href: "/support" }}
    />
  );
}

// Access Denied
export function AccessDenied() {
  return (
    <EmptyState
      icon={<ShieldAlert className="h-6 w-6 text-amber-500" />}
      title="Access denied"
      description="You don't have permission to view this content. Contact your administrator if you believe this is a mistake."
      action={{ label: "Go to Dashboard", href: "/dashboard" }}
    />
  );
}

// Rate Limited
export function RateLimited({
  retryAfter,
  onRetry,
}: {
  retryAfter?: number;
  onRetry?: () => void;
}) {
  return (
    <EmptyState
      icon={<AlertTriangle className="h-6 w-6 text-amber-500" />}
      title="Too many requests"
      description={
        retryAfter
          ? `Please wait ${retryAfter} seconds before trying again.`
          : "You've made too many requests. Please wait a moment and try again."
      }
      action={onRetry ? { label: "Try Again", onClick: onRetry } : undefined}
    />
  );
}

// Platform Error
export function PlatformError({
  platform,
  message,
  onRetry,
  onReconnect,
}: {
  platform: string;
  message?: string;
  onRetry?: () => void;
  onReconnect?: () => void;
}) {
  return (
    <EmptyState
      icon={<AlertTriangle className="h-6 w-6 text-amber-500" />}
      title={`${platform} error`}
      description={
        message ||
        `We're having trouble connecting to ${platform}. This might be a temporary issue.`
      }
      action={onRetry ? { label: "Retry", onClick: onRetry } : undefined}
      secondaryAction={
        onReconnect ? { label: "Reconnect Account", onClick: onReconnect } : undefined
      }
    />
  );
}

// Generic Loading Failed
export function LoadingFailed({
  title = "Failed to load",
  message,
  onRetry,
}: {
  title?: string;
  message?: string;
  onRetry?: () => void;
}) {
  return (
    <EmptyState
      icon={<RefreshCw className="h-6 w-6 text-slate-400" />}
      title={title}
      description={message || "Something went wrong while loading this content. Please try again."}
      action={onRetry ? { label: "Retry", onClick: onRetry } : undefined}
    />
  );
}

export default EmptyState;
