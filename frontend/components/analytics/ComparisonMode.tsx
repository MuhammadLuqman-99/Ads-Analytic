"use client";

import { ArrowUp, ArrowDown, Minus } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { cn } from "@/lib/utils";
import { formatCurrency, formatNumber } from "@/lib/mock-data";

interface MetricComparison {
  label: string;
  currentValue: number;
  previousValue: number;
  format: "currency" | "number" | "percentage" | "multiplier";
}

interface ComparisonModeProps {
  currentPeriod: string;
  previousPeriod: string;
  metrics: MetricComparison[];
}

function formatValue(value: number, format: MetricComparison["format"]): string {
  switch (format) {
    case "currency":
      return formatCurrency(value);
    case "number":
      return formatNumber(value);
    case "percentage":
      return `${value.toFixed(2)}%`;
    case "multiplier":
      return `${value.toFixed(2)}x`;
    default:
      return String(value);
  }
}

function calculateChange(current: number, previous: number): { delta: number; percentage: number } {
  const delta = current - previous;
  const percentage = previous !== 0 ? ((current - previous) / previous) * 100 : 0;
  return { delta, percentage };
}

export function ComparisonMode({ currentPeriod, previousPeriod, metrics }: ComparisonModeProps) {
  return (
    <Card className="bg-white border-slate-200">
      <CardHeader>
        <CardTitle className="text-slate-900">Period Comparison</CardTitle>
        <div className="flex items-center gap-4 text-sm text-slate-500">
          <span className="flex items-center gap-2">
            <div className="w-3 h-3 rounded bg-blue-500" />
            {currentPeriod}
          </span>
          <span>vs</span>
          <span className="flex items-center gap-2">
            <div className="w-3 h-3 rounded bg-slate-300" />
            {previousPeriod}
          </span>
        </div>
      </CardHeader>
      <CardContent>
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          {metrics.map((metric) => {
            const { delta, percentage } = calculateChange(metric.currentValue, metric.previousValue);
            const isPositive = delta > 0;
            const isNegative = delta < 0;
            const isNeutral = delta === 0;

            return (
              <div
                key={metric.label}
                className="p-4 rounded-lg border border-slate-200 bg-slate-50"
              >
                <p className="text-sm text-slate-500 mb-1">{metric.label}</p>

                {/* Current Value */}
                <p className="text-2xl font-bold text-slate-900">
                  {formatValue(metric.currentValue, metric.format)}
                </p>

                {/* Previous Value */}
                <p className="text-sm text-slate-400 mt-1">
                  was {formatValue(metric.previousValue, metric.format)}
                </p>

                {/* Delta and Percentage Change */}
                <div className={cn(
                  "flex items-center gap-2 mt-3 pt-3 border-t border-slate-200",
                  isPositive && "text-emerald-600",
                  isNegative && "text-red-600",
                  isNeutral && "text-slate-500"
                )}>
                  {isPositive && <ArrowUp className="h-4 w-4" />}
                  {isNegative && <ArrowDown className="h-4 w-4" />}
                  {isNeutral && <Minus className="h-4 w-4" />}

                  <span className="text-sm font-medium">
                    {isPositive && "+"}
                    {formatValue(Math.abs(delta), metric.format)}
                  </span>

                  <span className={cn(
                    "text-xs px-1.5 py-0.5 rounded",
                    isPositive && "bg-emerald-100",
                    isNegative && "bg-red-100",
                    isNeutral && "bg-slate-100"
                  )}>
                    {isPositive && "+"}
                    {percentage.toFixed(1)}%
                  </span>
                </div>
              </div>
            );
          })}
        </div>
      </CardContent>
    </Card>
  );
}

// Comparison Bar Component for side-by-side visualization
export function ComparisonBar({
  label,
  currentValue,
  previousValue,
  maxValue,
  format = "number",
}: {
  label: string;
  currentValue: number;
  previousValue: number;
  maxValue: number;
  format?: MetricComparison["format"];
}) {
  const currentWidth = (currentValue / maxValue) * 100;
  const previousWidth = (previousValue / maxValue) * 100;
  const { percentage } = calculateChange(currentValue, previousValue);
  const isPositive = percentage > 0;
  const isNegative = percentage < 0;

  return (
    <div className="space-y-2">
      <div className="flex items-center justify-between">
        <span className="text-sm font-medium text-slate-700">{label}</span>
        <div className="flex items-center gap-2">
          <span className="text-sm text-slate-900">
            {formatValue(currentValue, format)}
          </span>
          <span className={cn(
            "text-xs px-1.5 py-0.5 rounded",
            isPositive && "bg-emerald-100 text-emerald-700",
            isNegative && "bg-red-100 text-red-700",
            !isPositive && !isNegative && "bg-slate-100 text-slate-600"
          )}>
            {isPositive && "+"}
            {percentage.toFixed(1)}%
          </span>
        </div>
      </div>

      <div className="relative h-8">
        {/* Previous Period Bar (Background) */}
        <div
          className="absolute top-0 h-3 bg-slate-200 rounded"
          style={{ width: `${previousWidth}%` }}
        />
        {/* Current Period Bar (Foreground) */}
        <div
          className="absolute top-4 h-3 bg-blue-500 rounded"
          style={{ width: `${currentWidth}%` }}
        />
      </div>

      <div className="flex items-center gap-4 text-xs text-slate-500">
        <span className="flex items-center gap-1">
          <div className="w-2 h-2 rounded bg-blue-500" />
          Current: {formatValue(currentValue, format)}
        </span>
        <span className="flex items-center gap-1">
          <div className="w-2 h-2 rounded bg-slate-200" />
          Previous: {formatValue(previousValue, format)}
        </span>
      </div>
    </div>
  );
}

// Full Comparison Dashboard
export function ComparisonDashboard({
  currentPeriod,
  previousPeriod,
  summaryMetrics,
  detailedMetrics,
}: {
  currentPeriod: string;
  previousPeriod: string;
  summaryMetrics: MetricComparison[];
  detailedMetrics: {
    label: string;
    currentValue: number;
    previousValue: number;
    format: MetricComparison["format"];
  }[];
}) {
  const maxValue = Math.max(
    ...detailedMetrics.flatMap((m) => [m.currentValue, m.previousValue])
  );

  return (
    <div className="space-y-6">
      {/* Summary Cards */}
      <ComparisonMode
        currentPeriod={currentPeriod}
        previousPeriod={previousPeriod}
        metrics={summaryMetrics}
      />

      {/* Detailed Comparison Bars */}
      <Card className="bg-white border-slate-200">
        <CardHeader>
          <CardTitle className="text-slate-900">Detailed Comparison</CardTitle>
        </CardHeader>
        <CardContent className="space-y-6">
          {detailedMetrics.map((metric) => (
            <ComparisonBar
              key={metric.label}
              label={metric.label}
              currentValue={metric.currentValue}
              previousValue={metric.previousValue}
              maxValue={maxValue}
              format={metric.format}
            />
          ))}
        </CardContent>
      </Card>
    </div>
  );
}
