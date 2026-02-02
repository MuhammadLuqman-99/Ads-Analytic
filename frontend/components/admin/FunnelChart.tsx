"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { cn } from "@/lib/utils";

interface FunnelStep {
  name: string;
  count: number;
  conversionRate: number;
}

interface FunnelChartProps {
  title: string;
  steps: FunnelStep[];
  className?: string;
}

export function FunnelChart({ title, steps, className }: FunnelChartProps) {
  const maxCount = Math.max(...steps.map((s) => s.count));

  return (
    <Card className={cn("bg-white", className)}>
      <CardHeader>
        <CardTitle className="text-lg font-semibold text-slate-900">{title}</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="space-y-4">
          {steps.map((step, index) => {
            const widthPercent = maxCount > 0 ? (step.count / maxCount) * 100 : 0;
            const isLast = index === steps.length - 1;

            return (
              <div key={step.name} className="space-y-2">
                <div className="flex items-center justify-between text-sm">
                  <span className="font-medium text-slate-700">{step.name}</span>
                  <span className="text-slate-500">
                    {step.count.toLocaleString()}
                    {!isLast && (
                      <span className="ml-2 text-xs text-slate-400">
                        ({step.conversionRate.toFixed(1)}%)
                      </span>
                    )}
                  </span>
                </div>
                <div className="relative h-8 bg-slate-100 rounded-lg overflow-hidden">
                  <div
                    className={cn(
                      "absolute inset-y-0 left-0 rounded-lg transition-all duration-500",
                      index === 0 && "bg-indigo-500",
                      index === 1 && "bg-indigo-400",
                      index === 2 && "bg-indigo-300",
                      index >= 3 && "bg-indigo-200"
                    )}
                    style={{ width: `${widthPercent}%` }}
                  />
                  <div className="absolute inset-0 flex items-center px-3">
                    <span className="text-xs font-medium text-white drop-shadow">
                      {widthPercent.toFixed(0)}%
                    </span>
                  </div>
                </div>
                {!isLast && (
                  <div className="flex items-center justify-center">
                    <div className="w-px h-4 bg-slate-300" />
                  </div>
                )}
              </div>
            );
          })}
        </div>

        {/* Overall Conversion */}
        {steps.length >= 2 && (
          <div className="mt-6 pt-4 border-t border-slate-200">
            <div className="flex items-center justify-between">
              <span className="text-sm text-slate-600">Overall Conversion</span>
              <span className="text-lg font-bold text-indigo-600">
                {((steps[steps.length - 1].count / steps[0].count) * 100).toFixed(1)}%
              </span>
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  );
}

export function FunnelChartSkeleton() {
  return (
    <Card className="bg-white">
      <CardHeader>
        <div className="h-6 w-40 bg-slate-200 rounded animate-pulse" />
      </CardHeader>
      <CardContent className="space-y-4">
        {[1, 2, 3, 4].map((i) => (
          <div key={i} className="space-y-2 animate-pulse">
            <div className="flex justify-between">
              <div className="h-4 w-24 bg-slate-200 rounded" />
              <div className="h-4 w-16 bg-slate-200 rounded" />
            </div>
            <div className="h-8 bg-slate-200 rounded-lg" />
          </div>
        ))}
      </CardContent>
    </Card>
  );
}
