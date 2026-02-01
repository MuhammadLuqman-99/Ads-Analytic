"use client";

import { useState } from "react";
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  Legend,
} from "recharts";
import { format, parseISO } from "date-fns";
import { useDailyMetrics } from "@/hooks/use-metrics";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { formatCurrency, type Platform } from "@/lib/mock-data";
import { cn } from "@/lib/utils";

interface PlatformToggle {
  key: "meta" | "tiktok" | "shopee";
  label: string;
  color: string;
  dataKey: string;
}

const platforms: PlatformToggle[] = [
  { key: "meta", label: "Meta", color: "#3b82f6", dataKey: "metaSpend" },
  { key: "tiktok", label: "TikTok", color: "#000000", dataKey: "tiktokSpend" },
  { key: "shopee", label: "Shopee", color: "#f97316", dataKey: "shopeeSpend" },
];

interface SpendChartProps {
  dateRange?: { from: Date; to: Date };
}

export function SpendChart({ dateRange }: SpendChartProps) {
  const { data: metrics, isLoading } = useDailyMetrics(dateRange);
  const [activePlatforms, setActivePlatforms] = useState<Set<string>>(
    new Set(["meta", "tiktok", "shopee"])
  );

  const togglePlatform = (platform: string) => {
    const newSet = new Set(activePlatforms);
    if (newSet.has(platform)) {
      if (newSet.size > 1) {
        newSet.delete(platform);
      }
    } else {
      newSet.add(platform);
    }
    setActivePlatforms(newSet);
  };

  if (isLoading) {
    return (
      <Card className="bg-white border-slate-200">
        <CardHeader className="pb-2">
          <CardTitle className="text-slate-900">Spend Over Time</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex gap-2 mb-4">
            {[1, 2, 3].map((i) => (
              <div key={i} className="h-8 w-20 bg-slate-200 rounded animate-pulse" />
            ))}
          </div>
          <div className="h-[300px] flex items-center justify-center">
            <div className="flex items-center gap-2 text-slate-400">
              <svg className="animate-spin h-5 w-5" viewBox="0 0 24 24">
                <circle
                  className="opacity-25"
                  cx="12"
                  cy="12"
                  r="10"
                  stroke="currentColor"
                  strokeWidth="4"
                  fill="none"
                />
                <path
                  className="opacity-75"
                  fill="currentColor"
                  d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"
                />
              </svg>
              <span>Loading chart...</span>
            </div>
          </div>
        </CardContent>
      </Card>
    );
  }

  const chartData =
    metrics?.map((m) => ({
      ...m,
      formattedDate: format(parseISO(m.date), "MMM dd"),
    })) || [];

  const CustomTooltip = ({ active, payload, label }: any) => {
    if (!active || !payload) return null;

    return (
      <div className="bg-white border border-slate-200 rounded-lg shadow-lg p-3">
        <p className="text-sm font-medium text-slate-900 mb-2">{label}</p>
        <div className="space-y-1">
          {payload.map((entry: any) => (
            <div key={entry.dataKey} className="flex items-center justify-between gap-4">
              <div className="flex items-center gap-2">
                <div
                  className="w-3 h-3 rounded-full"
                  style={{ backgroundColor: entry.color }}
                />
                <span className="text-sm text-slate-600">
                  {entry.dataKey === "metaSpend"
                    ? "Meta"
                    : entry.dataKey === "tiktokSpend"
                    ? "TikTok"
                    : "Shopee"}
                </span>
              </div>
              <span className="text-sm font-medium text-slate-900">
                {formatCurrency(entry.value)}
              </span>
            </div>
          ))}
          <div className="pt-2 mt-2 border-t border-slate-100 flex items-center justify-between">
            <span className="text-sm font-medium text-slate-600">Total</span>
            <span className="text-sm font-bold text-slate-900">
              {formatCurrency(
                payload.reduce((sum: number, entry: any) => sum + entry.value, 0)
              )}
            </span>
          </div>
        </div>
      </div>
    );
  };

  return (
    <Card className="bg-white border-slate-200">
      <CardHeader className="pb-2">
        <div className="flex items-center justify-between">
          <CardTitle className="text-slate-900">Spend Over Time</CardTitle>
        </div>
      </CardHeader>
      <CardContent>
        {/* Platform Toggles */}
        <div className="flex flex-wrap gap-2 mb-4">
          {platforms.map((platform) => (
            <Button
              key={platform.key}
              variant="outline"
              size="sm"
              onClick={() => togglePlatform(platform.key)}
              className={cn(
                "gap-2 transition-all",
                activePlatforms.has(platform.key)
                  ? "border-2"
                  : "opacity-50 border-slate-200"
              )}
              style={{
                borderColor: activePlatforms.has(platform.key)
                  ? platform.color
                  : undefined,
              }}
            >
              <div
                className="w-3 h-3 rounded-full"
                style={{ backgroundColor: platform.color }}
              />
              {platform.label}
            </Button>
          ))}
        </div>

        {/* Chart */}
        <ResponsiveContainer width="100%" height={300}>
          <LineChart data={chartData} margin={{ top: 10, right: 10, left: 0, bottom: 0 }}>
            <CartesianGrid strokeDasharray="3 3" stroke="#e2e8f0" vertical={false} />
            <XAxis
              dataKey="formattedDate"
              stroke="#94a3b8"
              fontSize={12}
              tickLine={false}
              axisLine={false}
            />
            <YAxis
              stroke="#94a3b8"
              fontSize={12}
              tickLine={false}
              axisLine={false}
              tickFormatter={(value) => `${(value / 1000).toFixed(0)}K`}
            />
            <Tooltip content={<CustomTooltip />} />
            {platforms.map(
              (platform) =>
                activePlatforms.has(platform.key) && (
                  <Line
                    key={platform.key}
                    type="monotone"
                    dataKey={platform.dataKey}
                    stroke={platform.color}
                    strokeWidth={2}
                    dot={false}
                    activeDot={{ r: 6, strokeWidth: 2, fill: "#fff" }}
                  />
                )
            )}
          </LineChart>
        </ResponsiveContainer>
      </CardContent>
    </Card>
  );
}
