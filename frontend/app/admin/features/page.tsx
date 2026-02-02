"use client";

import { BarChart3, TrendingUp, Zap } from "lucide-react";
import {
  AdminMetricCard,
  AdminMetricCardSkeleton,
  FeatureHeatmap,
  FeatureHeatmapSkeleton,
} from "@/components/admin";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { useFeatureUsage } from "@/lib/api/hooks/use-admin";

export default function AdminFeaturesPage() {
  const { data: features, isLoading } = useFeatureUsage();

  const featureData = features?.features || {};
  const totalUsage = Object.values(featureData).reduce((sum, count) => sum + count, 0);
  const mostUsed = Object.entries(featureData).sort(([, a], [, b]) => b - a)[0];
  const leastUsed = Object.entries(featureData).sort(([, a], [, b]) => a - b)[0];

  const formatFeatureName = (name: string) =>
    name
      .replace(/_/g, " ")
      .replace(/\b\w/g, (c) => c.toUpperCase());

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-2xl font-bold text-slate-900">Feature Usage</h1>
        <p className="mt-1 text-sm text-slate-600">
          Track which features are most used by your users
        </p>
      </div>

      {/* Summary Metrics */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        {isLoading ? (
          <>
            <AdminMetricCardSkeleton />
            <AdminMetricCardSkeleton />
            <AdminMetricCardSkeleton />
          </>
        ) : (
          <>
            <AdminMetricCard
              title="Total Feature Uses"
              value={totalUsage}
              icon={<BarChart3 className="h-5 w-5 text-indigo-600" />}
              iconBgColor="bg-indigo-100"
              subtitle="All time usage"
            />
            <AdminMetricCard
              title="Most Used Feature"
              value={mostUsed ? formatFeatureName(mostUsed[0]) : "N/A"}
              icon={<TrendingUp className="h-5 w-5 text-emerald-600" />}
              iconBgColor="bg-emerald-100"
              subtitle={mostUsed ? `${mostUsed[1].toLocaleString()} uses` : "No data"}
            />
            <AdminMetricCard
              title="Least Used Feature"
              value={leastUsed ? formatFeatureName(leastUsed[0]) : "N/A"}
              icon={<Zap className="h-5 w-5 text-amber-600" />}
              iconBgColor="bg-amber-100"
              subtitle={leastUsed ? `${leastUsed[1].toLocaleString()} uses` : "No data"}
            />
          </>
        )}
      </div>

      {/* Feature Heatmap */}
      {isLoading ? (
        <FeatureHeatmapSkeleton />
      ) : (
        <FeatureHeatmap features={featureData} />
      )}

      {/* Feature Details */}
      <Card className="bg-white">
        <CardHeader>
          <CardTitle className="text-lg">Feature Breakdown</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-3">
            {Object.entries(featureData)
              .sort(([, a], [, b]) => b - a)
              .map(([feature, count], index) => {
                const percent = totalUsage > 0 ? (count / totalUsage) * 100 : 0;
                return (
                  <div
                    key={feature}
                    className="flex items-center gap-4 p-3 bg-slate-50 rounded-lg"
                  >
                    <div className="flex items-center justify-center h-8 w-8 rounded-full bg-indigo-100 text-indigo-600 font-medium text-sm">
                      {index + 1}
                    </div>
                    <div className="flex-1">
                      <div className="font-medium text-slate-900">
                        {formatFeatureName(feature)}
                      </div>
                      <div className="mt-1 h-2 bg-slate-200 rounded-full overflow-hidden">
                        <div
                          className="h-full bg-indigo-500 rounded-full transition-all duration-500"
                          style={{ width: `${percent}%` }}
                        />
                      </div>
                    </div>
                    <div className="text-right">
                      <div className="font-medium text-slate-900">
                        {count.toLocaleString()}
                      </div>
                      <div className="text-xs text-slate-500">{percent.toFixed(1)}%</div>
                    </div>
                  </div>
                );
              })}
            {Object.keys(featureData).length === 0 && (
              <div className="text-center py-8 text-slate-500">
                No feature usage data yet
              </div>
            )}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
