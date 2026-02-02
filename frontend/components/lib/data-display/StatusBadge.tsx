"use client";

import { CheckCircle, PauseCircle, XCircle, AlertCircle, Clock, Loader2 } from "lucide-react";
import { cn } from "@/lib/utils";

type Status = "active" | "paused" | "error" | "pending" | "completed" | "draft" | "syncing";

interface StatusBadgeProps {
  status: Status;
  size?: "xs" | "sm" | "md" | "lg";
  showIcon?: boolean;
  className?: string;
}

const statusConfig: Record<Status, {
  label: string;
  bg: string;
  text: string;
  icon: typeof CheckCircle;
  animate?: boolean;
}> = {
  active: {
    label: "Active",
    bg: "bg-emerald-100",
    text: "text-emerald-700",
    icon: CheckCircle,
  },
  paused: {
    label: "Paused",
    bg: "bg-amber-100",
    text: "text-amber-700",
    icon: PauseCircle,
  },
  error: {
    label: "Error",
    bg: "bg-red-100",
    text: "text-red-700",
    icon: XCircle,
  },
  pending: {
    label: "Pending",
    bg: "bg-slate-100",
    text: "text-slate-700",
    icon: Clock,
  },
  completed: {
    label: "Completed",
    bg: "bg-blue-100",
    text: "text-blue-700",
    icon: CheckCircle,
  },
  draft: {
    label: "Draft",
    bg: "bg-slate-100",
    text: "text-slate-500",
    icon: AlertCircle,
  },
  syncing: {
    label: "Syncing",
    bg: "bg-blue-100",
    text: "text-blue-700",
    icon: Loader2,
    animate: true,
  },
};

const sizeConfig = {
  xs: { badge: "px-1.5 py-0.5 text-[10px] gap-1", icon: "h-3 w-3" },
  sm: { badge: "px-2 py-0.5 text-xs gap-1", icon: "h-3.5 w-3.5" },
  md: { badge: "px-2.5 py-1 text-sm gap-1.5", icon: "h-4 w-4" },
  lg: { badge: "px-3 py-1.5 text-base gap-2", icon: "h-5 w-5" },
};

export function StatusBadge({
  status,
  size = "sm",
  showIcon = true,
  className,
}: StatusBadgeProps) {
  const config = statusConfig[status];
  const sizeStyles = sizeConfig[size];
  const Icon = config.icon;

  return (
    <span
      className={cn(
        "inline-flex items-center rounded-full font-medium",
        sizeStyles.badge,
        config.bg,
        config.text,
        className
      )}
    >
      {showIcon && (
        <Icon className={cn(sizeStyles.icon, config.animate && "animate-spin")} />
      )}
      {config.label}
    </span>
  );
}

// Dot-only status indicator for compact views
export function StatusDot({
  status,
  size = "sm",
  className,
}: {
  status: Status;
  size?: "xs" | "sm" | "md";
  className?: string;
}) {
  const config = statusConfig[status];
  const dotSizes = {
    xs: "h-1.5 w-1.5",
    sm: "h-2 w-2",
    md: "h-2.5 w-2.5",
  };

  return (
    <span
      className={cn(
        "inline-block rounded-full",
        dotSizes[size],
        config.text.replace("text-", "bg-").replace("-700", "-500"),
        config.animate && "animate-pulse",
        className
      )}
    />
  );
}
