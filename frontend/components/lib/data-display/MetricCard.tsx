"use client";

import { type LucideIcon } from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import { cn } from "@/lib/utils";
import { TrendIndicator } from "./TrendIndicator";

interface MetricCardProps {
  label: string;
  value: string | number;
  change?: number;
  changeLabel?: string;
  icon?: LucideIcon;
  iconColor?: string;
  iconBgColor?: string;
  trend?: "up" | "down" | "neutral";
  size?: "sm" | "md" | "lg";
  className?: string;
}

export function MetricCard({
  label,
  value,
  change,
  changeLabel,
  icon: Icon,
  iconColor = "text-blue-600",
  iconBgColor = "bg-blue-100",
  trend,
  size = "md",
  className,
}: MetricCardProps) {
  const sizeStyles = {
    sm: {
      card: "p-4",
      value: "text-xl",
      label: "text-xs",
      icon: "h-8 w-8",
      iconSize: "h-4 w-4",
    },
    md: {
      card: "p-6",
      value: "text-2xl",
      label: "text-sm",
      icon: "h-10 w-10",
      iconSize: "h-5 w-5",
    },
    lg: {
      card: "p-8",
      value: "text-3xl",
      label: "text-base",
      icon: "h-12 w-12",
      iconSize: "h-6 w-6",
    },
  };

  const styles = sizeStyles[size];

  // Auto-detect trend from change if not provided
  const effectiveTrend = trend ?? (change !== undefined ? (change > 0 ? "up" : change < 0 ? "down" : "neutral") : undefined);

  return (
    <Card className={cn("bg-white border-slate-200", className)}>
      <CardContent className={styles.card}>
        <div className="flex items-start justify-between">
          <div className="flex-1">
            <p className={cn("font-medium text-slate-500", styles.label)}>
              {label}
            </p>
            <p className={cn("font-bold text-slate-900 mt-1", styles.value)}>
              {value}
            </p>
            {(change !== undefined || changeLabel) && (
              <div className="flex items-center gap-1 mt-2">
                {effectiveTrend && <TrendIndicator trend={effectiveTrend} size="sm" />}
                <span
                  className={cn(
                    "text-xs font-medium",
                    effectiveTrend === "up" && "text-emerald-600",
                    effectiveTrend === "down" && "text-red-600",
                    effectiveTrend === "neutral" && "text-slate-500"
                  )}
                >
                  {change !== undefined && (
                    <>
                      {change > 0 ? "+" : ""}
                      {typeof change === "number" ? change.toFixed(1) : change}%
                    </>
                  )}
                  {changeLabel && ` ${changeLabel}`}
                </span>
              </div>
            )}
          </div>
          {Icon && (
            <div
              className={cn(
                "rounded-lg flex items-center justify-center",
                styles.icon,
                iconBgColor
              )}
            >
              <Icon className={cn(styles.iconSize, iconColor)} />
            </div>
          )}
        </div>
      </CardContent>
    </Card>
  );
}
