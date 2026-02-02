"use client";

import { type LucideIcon, Inbox, Search, FileX, FolderOpen, Link2, AlertCircle } from "lucide-react";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

type EmptyStateVariant = "default" | "search" | "error" | "no-data" | "no-connection" | "custom";

interface EmptyStateProps {
  variant?: EmptyStateVariant;
  icon?: LucideIcon;
  title?: string;
  description?: string;
  action?: {
    label: string;
    onClick: () => void;
    variant?: "default" | "outline" | "secondary";
  };
  secondaryAction?: {
    label: string;
    onClick: () => void;
  };
  size?: "sm" | "md" | "lg";
  className?: string;
}

const variantDefaults: Record<EmptyStateVariant, { icon: LucideIcon; title: string; description: string }> = {
  default: {
    icon: Inbox,
    title: "No data available",
    description: "There's nothing here yet. Get started by creating your first item.",
  },
  search: {
    icon: Search,
    title: "No results found",
    description: "We couldn't find anything matching your search. Try different keywords.",
  },
  error: {
    icon: AlertCircle,
    title: "Something went wrong",
    description: "We encountered an error while loading this content. Please try again.",
  },
  "no-data": {
    icon: FolderOpen,
    title: "No data yet",
    description: "Start adding data to see it appear here.",
  },
  "no-connection": {
    icon: Link2,
    title: "No connections",
    description: "Connect your ad accounts to start tracking performance.",
  },
  custom: {
    icon: FileX,
    title: "Empty",
    description: "",
  },
};

const sizeConfig = {
  sm: {
    container: "py-8",
    iconContainer: "h-12 w-12",
    icon: "h-6 w-6",
    title: "text-base",
    description: "text-sm",
  },
  md: {
    container: "py-12",
    iconContainer: "h-16 w-16",
    icon: "h-8 w-8",
    title: "text-lg",
    description: "text-sm",
  },
  lg: {
    container: "py-16",
    iconContainer: "h-20 w-20",
    icon: "h-10 w-10",
    title: "text-xl",
    description: "text-base",
  },
};

export function EmptyState({
  variant = "default",
  icon,
  title,
  description,
  action,
  secondaryAction,
  size = "md",
  className,
}: EmptyStateProps) {
  const defaults = variantDefaults[variant];
  const Icon = icon || defaults.icon;
  const displayTitle = title || defaults.title;
  const displayDescription = description || defaults.description;
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
          "rounded-full bg-slate-100 flex items-center justify-center mb-4",
          sizeStyles.iconContainer
        )}
      >
        <Icon className={cn("text-slate-400", sizeStyles.icon)} />
      </div>

      <h3 className={cn("font-semibold text-slate-900 mb-2", sizeStyles.title)}>
        {displayTitle}
      </h3>

      {displayDescription && (
        <p
          className={cn(
            "text-slate-500 max-w-md mb-6",
            sizeStyles.description
          )}
        >
          {displayDescription}
        </p>
      )}

      {(action || secondaryAction) && (
        <div className="flex items-center gap-3">
          {secondaryAction && (
            <Button variant="outline" onClick={secondaryAction.onClick}>
              {secondaryAction.label}
            </Button>
          )}
          {action && (
            <Button
              variant={action.variant || "default"}
              onClick={action.onClick}
            >
              {action.label}
            </Button>
          )}
        </div>
      )}
    </div>
  );
}

// Preset empty states for common use cases
export function SearchEmptyState({
  onClear,
  className,
}: {
  onClear?: () => void;
  className?: string;
}) {
  return (
    <EmptyState
      variant="search"
      action={onClear ? { label: "Clear filters", onClick: onClear, variant: "outline" } : undefined}
      className={className}
    />
  );
}

export function NoDataEmptyState({
  onAction,
  actionLabel = "Get started",
  className,
}: {
  onAction?: () => void;
  actionLabel?: string;
  className?: string;
}) {
  return (
    <EmptyState
      variant="no-data"
      action={onAction ? { label: actionLabel, onClick: onAction } : undefined}
      className={className}
    />
  );
}

export function NoConnectionEmptyState({
  onConnect,
  className,
}: {
  onConnect?: () => void;
  className?: string;
}) {
  return (
    <EmptyState
      variant="no-connection"
      action={onConnect ? { label: "Connect Account", onClick: onConnect } : undefined}
      className={className}
    />
  );
}
