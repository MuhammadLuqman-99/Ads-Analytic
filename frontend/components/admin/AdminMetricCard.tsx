"use client";

import { TrendingUp, TrendingDown, Minus } from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import { cn } from "@/lib/utils";

interface AdminMetricCardProps {
  title: string;
  value: string | number;
  change?: number;
  subtitle?: string;
  icon: React.ReactNode;
  iconBgColor?: string;
  format?: "number" | "currency" | "percent";
}

export function AdminMetricCard({
  title,
  value,
  change,
  subtitle,
  icon,
  iconBgColor = "bg-indigo-100",
  format = "number",
}: AdminMetricCardProps) {
  const formatValue = (val: string | number): string => {
    if (typeof val === "string") return val;

    switch (format) {
      case "currency":
        return new Intl.NumberFormat("ms-MY", {
          style: "currency",
          currency: "MYR",
          minimumFractionDigits: 0,
          maximumFractionDigits: 0,
        }).format(val);
      case "percent":
        return `${val.toFixed(1)}%`;
      default:
        return new Intl.NumberFormat("en-US").format(val);
    }
  };

  const hasChange = typeof change === "number";
  const isPositive = hasChange && change > 0;
  const isNegative = hasChange && change < 0;
  const isNeutral = hasChange && change === 0;

  return (
    <Card className="bg-white border-slate-200 hover:shadow-md transition-shadow">
      <CardContent className="p-6">
        <div className="flex items-start justify-between">
          <div className={cn("p-3 rounded-xl", iconBgColor)}>{icon}</div>
          {hasChange && (
            <div
              className={cn(
                "flex items-center gap-1 text-sm font-medium px-2 py-1 rounded-full",
                isPositive && "text-emerald-700 bg-emerald-50",
                isNegative && "text-rose-700 bg-rose-50",
                isNeutral && "text-slate-600 bg-slate-100"
              )}
            >
              {isPositive && <TrendingUp className="h-3.5 w-3.5" />}
              {isNegative && <TrendingDown className="h-3.5 w-3.5" />}
              {isNeutral && <Minus className="h-3.5 w-3.5" />}
              {Math.abs(change).toFixed(1)}%
            </div>
          )}
        </div>
        <div className="mt-4">
          <p className="text-sm font-medium text-slate-500">{title}</p>
          <p className="mt-1 text-2xl font-bold text-slate-900">{formatValue(value)}</p>
          {subtitle && <p className="mt-1 text-xs text-slate-400">{subtitle}</p>}
        </div>
      </CardContent>
    </Card>
  );
}

export function AdminMetricCardSkeleton() {
  return (
    <Card className="bg-white border-slate-200">
      <CardContent className="p-6 animate-pulse">
        <div className="flex items-start justify-between">
          <div className="h-12 w-12 bg-slate-200 rounded-xl" />
          <div className="h-6 w-16 bg-slate-200 rounded-full" />
        </div>
        <div className="mt-4 space-y-2">
          <div className="h-4 w-20 bg-slate-200 rounded" />
          <div className="h-7 w-32 bg-slate-200 rounded" />
        </div>
      </CardContent>
    </Card>
  );
}
