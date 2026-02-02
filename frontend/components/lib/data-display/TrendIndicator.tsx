"use client";

import { TrendingUp, TrendingDown, Minus } from "lucide-react";
import { cn } from "@/lib/utils";

type Trend = "up" | "down" | "neutral";

interface TrendIndicatorProps {
  trend: Trend;
  value?: number | string;
  showLabel?: boolean;
  size?: "xs" | "sm" | "md" | "lg";
  className?: string;
}

const trendConfig: Record<Trend, { icon: typeof TrendingUp; color: string; bgColor: string; label: string }> = {
  up: {
    icon: TrendingUp,
    color: "text-emerald-600",
    bgColor: "bg-emerald-100",
    label: "Increase",
  },
  down: {
    icon: TrendingDown,
    color: "text-red-600",
    bgColor: "bg-red-100",
    label: "Decrease",
  },
  neutral: {
    icon: Minus,
    color: "text-slate-500",
    bgColor: "bg-slate-100",
    label: "No change",
  },
};

const sizeConfig = {
  xs: { icon: "h-3 w-3", text: "text-xs", padding: "px-1 py-0.5" },
  sm: { icon: "h-4 w-4", text: "text-xs", padding: "px-1.5 py-0.5" },
  md: { icon: "h-4 w-4", text: "text-sm", padding: "px-2 py-1" },
  lg: { icon: "h-5 w-5", text: "text-base", padding: "px-2.5 py-1" },
};

export function TrendIndicator({
  trend,
  value,
  showLabel = false,
  size = "sm",
  className,
}: TrendIndicatorProps) {
  const config = trendConfig[trend];
  const sizeStyles = sizeConfig[size];
  const Icon = config.icon;

  // If just showing the icon
  if (value === undefined && !showLabel) {
    return (
      <Icon className={cn(sizeStyles.icon, config.color, className)} />
    );
  }

  // With value or label
  return (
    <span
      className={cn(
        "inline-flex items-center gap-1 rounded-full font-medium",
        sizeStyles.padding,
        sizeStyles.text,
        config.bgColor,
        config.color,
        className
      )}
    >
      <Icon className={sizeStyles.icon} />
      {value !== undefined && (
        <span>
          {trend === "up" && "+"}
          {value}
          {typeof value === "number" && "%"}
        </span>
      )}
      {showLabel && !value && <span>{config.label}</span>}
    </span>
  );
}

// Utility function to determine trend from a change value
export function getTrendFromChange(change: number): Trend {
  if (change > 0) return "up";
  if (change < 0) return "down";
  return "neutral";
}
