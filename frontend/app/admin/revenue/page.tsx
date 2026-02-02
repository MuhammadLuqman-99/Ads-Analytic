"use client";

import { DollarSign, TrendingUp, TrendingDown, Percent } from "lucide-react";
import { AdminMetricCard, AdminMetricCardSkeleton } from "@/components/admin";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { useRevenue, useAdminDashboard } from "@/lib/api/hooks/use-admin";

export default function AdminRevenuePage() {
  const { data: dashboard, isLoading: isDashboardLoading } = useAdminDashboard();
  const { data: revenue, isLoading: isRevenueLoading } = useRevenue();

  const isLoading = isDashboardLoading || isRevenueLoading;

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-2xl font-bold text-slate-900">Revenue Metrics</h1>
        <p className="mt-1 text-sm text-slate-600">
          Financial performance and subscription metrics
        </p>
      </div>

      {/* Key Revenue Metrics */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        {isLoading ? (
          <>
            <AdminMetricCardSkeleton />
            <AdminMetricCardSkeleton />
            <AdminMetricCardSkeleton />
            <AdminMetricCardSkeleton />
          </>
        ) : (
          <>
            <AdminMetricCard
              title="MRR"
              value={revenue?.mrr || dashboard?.metrics.mrr || 0}
              format="currency"
              icon={<DollarSign className="h-5 w-5 text-emerald-600" />}
              iconBgColor="bg-emerald-100"
              subtitle="Monthly recurring revenue"
            />
            <AdminMetricCard
              title="ARR"
              value={(revenue?.mrr || dashboard?.metrics.mrr || 0) * 12}
              format="currency"
              icon={<TrendingUp className="h-5 w-5 text-blue-600" />}
              iconBgColor="bg-blue-100"
              subtitle="Annual recurring revenue"
            />
            <AdminMetricCard
              title="Churn Rate"
              value={revenue?.churnRate || dashboard?.metrics.churnRate || 0}
              format="percent"
              icon={<TrendingDown className="h-5 w-5 text-rose-600" />}
              iconBgColor="bg-rose-100"
              subtitle="Monthly churn"
            />
            <AdminMetricCard
              title="Net Revenue Retention"
              value={100 - (revenue?.churnRate || dashboard?.metrics.churnRate || 0)}
              format="percent"
              icon={<Percent className="h-5 w-5 text-indigo-600" />}
              iconBgColor="bg-indigo-100"
              subtitle="After churn"
            />
          </>
        )}
      </div>

      {/* Revenue Breakdown */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <Card className="bg-white">
          <CardHeader>
            <CardTitle className="text-lg">Revenue by Plan</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {[
                { name: "Free", users: 0, revenue: 0, color: "bg-slate-500" },
                { name: "Pro (RM99/mo)", users: 0, revenue: 0, color: "bg-blue-500" },
                { name: "Business (RM299/mo)", users: 0, revenue: 0, color: "bg-indigo-500" },
              ].map((plan) => (
                <div key={plan.name} className="flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    <div className={`h-3 w-3 rounded-full ${plan.color}`} />
                    <span className="text-sm font-medium text-slate-700">{plan.name}</span>
                  </div>
                  <div className="text-right">
                    <div className="text-sm font-medium text-slate-900">
                      RM {plan.revenue.toLocaleString()}
                    </div>
                    <div className="text-xs text-slate-500">{plan.users} users</div>
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>

        <Card className="bg-white">
          <CardHeader>
            <CardTitle className="text-lg">Key Metrics</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              <div className="flex items-center justify-between p-3 bg-slate-50 rounded-lg">
                <span className="text-sm text-slate-600">Average Revenue Per User</span>
                <span className="text-lg font-bold text-slate-900">
                  RM {dashboard?.metrics.totalUsers ?
                    ((revenue?.mrr || dashboard?.metrics.mrr || 0) / dashboard.metrics.totalUsers).toFixed(2)
                    : "0.00"}
                </span>
              </div>
              <div className="flex items-center justify-between p-3 bg-slate-50 rounded-lg">
                <span className="text-sm text-slate-600">Paying Customers</span>
                <span className="text-lg font-bold text-slate-900">
                  {dashboard?.metrics.activeUsers || 0}
                </span>
              </div>
              <div className="flex items-center justify-between p-3 bg-slate-50 rounded-lg">
                <span className="text-sm text-slate-600">Conversion Rate (Free to Paid)</span>
                <span className="text-lg font-bold text-slate-900">0%</span>
              </div>
              <div className="flex items-center justify-between p-3 bg-slate-50 rounded-lg">
                <span className="text-sm text-slate-600">Customer Lifetime Value</span>
                <span className="text-lg font-bold text-slate-900">RM 0</span>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
