"use client";

import { Layers, TrendingUp, Users } from "lucide-react";
import {
  AdminMetricCard,
  AdminMetricCardSkeleton,
  PlatformBreakdown,
  PlatformBreakdownSkeleton,
} from "@/components/admin";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { usePlatformBreakdown, useAdminDashboard } from "@/lib/api/hooks/use-admin";

export default function AdminPlatformsPage() {
  const { data: platforms, isLoading: isPlatformsLoading } = usePlatformBreakdown();
  const { data: dashboard, isLoading: isDashboardLoading } = useAdminDashboard();

  const platformData = platforms?.platforms || {};
  const totalConnections = Object.values(platformData).reduce((sum, count) => sum + count, 0);
  const uniquePlatforms = Object.keys(platformData).length;

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-2xl font-bold text-slate-900">Platform Analytics</h1>
        <p className="mt-1 text-sm text-slate-600">
          Platform connection and usage statistics
        </p>
      </div>

      {/* Summary Metrics */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        {isPlatformsLoading ? (
          <>
            <AdminMetricCardSkeleton />
            <AdminMetricCardSkeleton />
            <AdminMetricCardSkeleton />
          </>
        ) : (
          <>
            <AdminMetricCard
              title="Total Connections"
              value={totalConnections}
              icon={<Layers className="h-5 w-5 text-indigo-600" />}
              iconBgColor="bg-indigo-100"
              subtitle="Across all platforms"
            />
            <AdminMetricCard
              title="Platforms Supported"
              value={uniquePlatforms}
              icon={<TrendingUp className="h-5 w-5 text-emerald-600" />}
              iconBgColor="bg-emerald-100"
              subtitle="Active integrations"
            />
            <AdminMetricCard
              title="Avg Connections/User"
              value={
                dashboard?.metrics.activeUsers
                  ? (totalConnections / dashboard.metrics.activeUsers).toFixed(1)
                  : "0"
              }
              icon={<Users className="h-5 w-5 text-blue-600" />}
              iconBgColor="bg-blue-100"
              subtitle="Per active user"
            />
          </>
        )}
      </div>

      {/* Platform Breakdown */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {isPlatformsLoading ? (
          <PlatformBreakdownSkeleton />
        ) : (
          <PlatformBreakdown platforms={platformData} />
        )}

        {/* Platform Health */}
        <Card className="bg-white">
          <CardHeader>
            <CardTitle className="text-lg">Platform Health</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {Object.entries(platformData).map(([platform, count]) => (
                <div
                  key={platform}
                  className="flex items-center justify-between p-3 bg-slate-50 rounded-lg"
                >
                  <div>
                    <div className="font-medium text-slate-900 capitalize">{platform} Ads</div>
                    <div className="text-sm text-slate-500">{count} connections</div>
                  </div>
                  <div className="flex items-center gap-2">
                    <div className="h-2 w-2 rounded-full bg-emerald-500" />
                    <span className="text-sm text-emerald-600">Healthy</span>
                  </div>
                </div>
              ))}
              {Object.keys(platformData).length === 0 && (
                <div className="text-center py-8 text-slate-500">
                  No platform connections yet
                </div>
              )}
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Platform Adoption Trends */}
      <Card className="bg-white">
        <CardHeader>
          <CardTitle className="text-lg">Platform Popularity</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            {[
              { name: "Meta Ads", color: "bg-blue-500", popular: true },
              { name: "TikTok Ads", color: "bg-slate-800", popular: true },
              { name: "Shopee Ads", color: "bg-orange-500", popular: false },
              { name: "Google Ads", color: "bg-green-500", popular: false },
            ].map((platform) => (
              <div
                key={platform.name}
                className="p-4 border border-slate-200 rounded-lg text-center"
              >
                <div className={`h-12 w-12 mx-auto rounded-full ${platform.color} mb-3`} />
                <div className="font-medium text-slate-900">{platform.name}</div>
                {platform.popular && (
                  <span className="inline-block mt-2 px-2 py-1 text-xs bg-amber-100 text-amber-700 rounded-full">
                    Popular
                  </span>
                )}
              </div>
            ))}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
