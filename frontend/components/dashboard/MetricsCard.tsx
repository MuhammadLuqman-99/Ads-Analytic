"use client";

import { TrendingUp, TrendingDown } from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import { cn } from "@/lib/utils";

interface MetricsCardProps {
  title: string;
  value: string;
  change: number;
  icon: React.ReactNode;
  iconBgColor: string;
  subtitle?: string;
}

export function MetricsCard({
  title,
  value,
  change,
  icon,
  iconBgColor,
  subtitle,
}: MetricsCardProps) {
  const isPositive = change >= 0;

  return (
    <Card className="bg-white border-slate-200 hover:shadow-md transition-shadow">
      <CardContent className="p-6">
        <div className="flex items-start justify-between">
          <div className={cn("p-3 rounded-xl", iconBgColor)}>{icon}</div>
          <div
            className={cn(
              "flex items-center gap-1 text-sm font-medium px-2 py-1 rounded-full",
              isPositive
                ? "text-emerald-700 bg-emerald-50"
                : "text-rose-700 bg-rose-50"
            )}
          >
            {isPositive ? (
              <TrendingUp className="h-3.5 w-3.5" />
            ) : (
              <TrendingDown className="h-3.5 w-3.5" />
            )}
            {Math.abs(change).toFixed(1)}%
          </div>
        </div>
        <div className="mt-4">
          <p className="text-sm font-medium text-slate-500">{title}</p>
          <p className="mt-1 text-2xl font-bold text-slate-900">{value}</p>
          {subtitle && (
            <p className="mt-1 text-xs text-slate-400">{subtitle}</p>
          )}
        </div>
      </CardContent>
    </Card>
  );
}

export function MetricsCardSkeleton() {
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
