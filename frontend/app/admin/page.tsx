"use client";

import {
  Users,
  UserCheck,
  CalendarDays,
  TrendingUp,
  DollarSign,
  Percent,
  Activity,
  Layers,
} from "lucide-react";
import {
  AdminMetricCard,
  AdminMetricCardSkeleton,
  FunnelChart,
  FunnelChartSkeleton,
  PlatformBreakdown,
  PlatformBreakdownSkeleton,
  FeatureHeatmap,
  FeatureHeatmapSkeleton,
  UsersTable,
  UsersTableSkeleton,
} from "@/components/admin";
import {
  useAdminDashboard,
  useFunnel,
  usePlatformBreakdown,
  useFeatureUsage,
  useTopUsers,
} from "@/lib/api/hooks/use-admin";

export default function AdminDashboardPage() {
  const { data: dashboard, isLoading: isDashboardLoading } = useAdminDashboard();
  const { data: funnel, isLoading: isFunnelLoading } = useFunnel("activation");
  const { data: platforms, isLoading: isPlatformsLoading } = usePlatformBreakdown();
  const { data: features, isLoading: isFeaturesLoading } = useFeatureUsage();
  const { data: topUsers, isLoading: isTopUsersLoading } = useTopUsers("events", 5);

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-2xl font-bold text-slate-900">Admin Dashboard</h1>
        <p className="mt-1 text-sm text-slate-600">
          Internal analytics and product metrics overview
        </p>
      </div>

      {/* Key Metrics */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        {isDashboardLoading ? (
          <>
            <AdminMetricCardSkeleton />
            <AdminMetricCardSkeleton />
            <AdminMetricCardSkeleton />
            <AdminMetricCardSkeleton />
          </>
        ) : (
          <>
            <AdminMetricCard
              title="Total Users"
              value={dashboard?.metrics.totalUsers || 0}
              icon={<Users className="h-5 w-5 text-indigo-600" />}
              iconBgColor="bg-indigo-100"
              subtitle="All registered users"
            />
            <AdminMetricCard
              title="Active Users"
              value={dashboard?.metrics.activeUsers || 0}
              icon={<UserCheck className="h-5 w-5 text-emerald-600" />}
              iconBgColor="bg-emerald-100"
              subtitle="Active in last 30 days"
            />
            <AdminMetricCard
              title="MRR"
              value={dashboard?.metrics.mrr || 0}
              format="currency"
              icon={<DollarSign className="h-5 w-5 text-green-600" />}
              iconBgColor="bg-green-100"
              subtitle="Monthly recurring revenue"
            />
            <AdminMetricCard
              title="Churn Rate"
              value={dashboard?.metrics.churnRate || 0}
              format="percent"
              icon={<Percent className="h-5 w-5 text-rose-600" />}
              iconBgColor="bg-rose-100"
              subtitle="Last 30 days"
            />
          </>
        )}
      </div>

      {/* DAU, WAU, MAU */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        {isDashboardLoading ? (
          <>
            <AdminMetricCardSkeleton />
            <AdminMetricCardSkeleton />
            <AdminMetricCardSkeleton />
          </>
        ) : (
          <>
            <AdminMetricCard
              title="DAU"
              value={dashboard?.dau || 0}
              icon={<Activity className="h-5 w-5 text-blue-600" />}
              iconBgColor="bg-blue-100"
              subtitle="Daily active users"
            />
            <AdminMetricCard
              title="WAU"
              value={dashboard?.wau || 0}
              icon={<CalendarDays className="h-5 w-5 text-purple-600" />}
              iconBgColor="bg-purple-100"
              subtitle="Weekly active users"
            />
            <AdminMetricCard
              title="MAU"
              value={dashboard?.mau || 0}
              icon={<TrendingUp className="h-5 w-5 text-amber-600" />}
              iconBgColor="bg-amber-100"
              subtitle="Monthly active users"
            />
          </>
        )}
      </div>

      {/* Charts Row */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Conversion Funnel */}
        {isFunnelLoading ? (
          <FunnelChartSkeleton />
        ) : (
          <FunnelChart
            title="User Activation Funnel"
            steps={
              funnel?.steps || [
                { name: "Registered", count: 0, conversionRate: 100 },
                { name: "Platform Connected", count: 0, conversionRate: 0 },
                { name: "First Sync", count: 0, conversionRate: 0 },
                { name: "Active User", count: 0, conversionRate: 0 },
              ]
            }
          />
        )}

        {/* Platform Breakdown */}
        {isPlatformsLoading ? (
          <PlatformBreakdownSkeleton />
        ) : (
          <PlatformBreakdown
            platforms={
              platforms?.platforms || {
                meta: 0,
                tiktok: 0,
                shopee: 0,
              }
            }
          />
        )}
      </div>

      {/* Feature Usage Heatmap */}
      {isFeaturesLoading ? (
        <FeatureHeatmapSkeleton />
      ) : (
        <FeatureHeatmap
          features={
            features?.features || {
              dashboard_viewed: 0,
              campaign_exported: 0,
              report_generated: 0,
              comparison_viewed: 0,
            }
          }
        />
      )}

      {/* Top Users */}
      {isTopUsersLoading ? (
        <UsersTableSkeleton />
      ) : (
        <UsersTable title="Top Active Users" users={topUsers?.users || []} />
      )}
    </div>
  );
}
