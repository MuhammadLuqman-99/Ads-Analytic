"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { cn } from "@/lib/utils";

interface FeatureHeatmapProps {
  features: Record<string, number>;
  className?: string;
}

const featureLabels: Record<string, string> = {
  dashboard_viewed: "Dashboard View",
  campaign_exported: "Export Campaign",
  report_generated: "Generate Report",
  comparison_viewed: "Compare Campaigns",
  settings_updated: "Update Settings",
  platform_connected: "Connect Platform",
  platform_disconnected: "Disconnect Platform",
  filter_applied: "Apply Filter",
  date_range_changed: "Change Date Range",
  sync_triggered: "Trigger Sync",
};

export function FeatureHeatmap({ features, className }: FeatureHeatmapProps) {
  const maxUsage = Math.max(...Object.values(features));
  const sortedFeatures = Object.entries(features).sort(([, a], [, b]) => b - a);

  const getHeatColor = (value: number): string => {
    if (maxUsage === 0) return "bg-slate-100";
    const intensity = value / maxUsage;

    if (intensity >= 0.8) return "bg-indigo-600";
    if (intensity >= 0.6) return "bg-indigo-500";
    if (intensity >= 0.4) return "bg-indigo-400";
    if (intensity >= 0.2) return "bg-indigo-300";
    return "bg-indigo-100";
  };

  const getTextColor = (value: number): string => {
    if (maxUsage === 0) return "text-slate-600";
    const intensity = value / maxUsage;
    return intensity >= 0.4 ? "text-white" : "text-slate-700";
  };

  return (
    <Card className={cn("bg-white", className)}>
      <CardHeader>
        <CardTitle className="text-lg font-semibold text-slate-900">Feature Usage</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-3">
          {sortedFeatures.map(([feature, count]) => {
            const label = featureLabels[feature] || feature.replace(/_/g, " ");
            const bgColor = getHeatColor(count);
            const textColor = getTextColor(count);

            return (
              <div
                key={feature}
                className={cn(
                  "p-4 rounded-lg transition-all hover:scale-105",
                  bgColor
                )}
              >
                <div className={cn("text-xs font-medium truncate", textColor)}>
                  {label}
                </div>
                <div className={cn("mt-1 text-lg font-bold", textColor)}>
                  {count.toLocaleString()}
                </div>
              </div>
            );
          })}
        </div>

        {/* Legend */}
        <div className="mt-6 pt-4 border-t border-slate-200">
          <div className="flex items-center justify-between text-xs text-slate-500">
            <span>Low usage</span>
            <div className="flex items-center gap-1">
              <div className="h-3 w-6 bg-indigo-100 rounded" />
              <div className="h-3 w-6 bg-indigo-300 rounded" />
              <div className="h-3 w-6 bg-indigo-400 rounded" />
              <div className="h-3 w-6 bg-indigo-500 rounded" />
              <div className="h-3 w-6 bg-indigo-600 rounded" />
            </div>
            <span>High usage</span>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}

export function FeatureHeatmapSkeleton() {
  return (
    <Card className="bg-white">
      <CardHeader>
        <div className="h-6 w-32 bg-slate-200 rounded animate-pulse" />
      </CardHeader>
      <CardContent>
        <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-3">
          {[1, 2, 3, 4, 5, 6, 7, 8].map((i) => (
            <div key={i} className="p-4 bg-slate-100 rounded-lg animate-pulse">
              <div className="h-3 w-16 bg-slate-200 rounded" />
              <div className="mt-2 h-5 w-12 bg-slate-200 rounded" />
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}
