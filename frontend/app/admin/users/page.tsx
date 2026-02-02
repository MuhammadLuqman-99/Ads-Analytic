"use client";

import { useState } from "react";
import { Users, Activity, CalendarDays, TrendingUp } from "lucide-react";
import {
  AdminMetricCard,
  AdminMetricCardSkeleton,
  UsersTable,
  UsersTableSkeleton,
} from "@/components/admin";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { useActiveUsers, useTopUsers } from "@/lib/api/hooks/use-admin";

export default function AdminUsersPage() {
  const [metric, setMetric] = useState("events");
  const { data: activeUsers, isLoading: isActiveLoading } = useActiveUsers();
  const { data: topUsers, isLoading: isTopUsersLoading } = useTopUsers(metric, 20);

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-2xl font-bold text-slate-900">Active Users</h1>
        <p className="mt-1 text-sm text-slate-600">
          User activity and engagement metrics
        </p>
      </div>

      {/* Active User Metrics */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        {isActiveLoading ? (
          <>
            <AdminMetricCardSkeleton />
            <AdminMetricCardSkeleton />
            <AdminMetricCardSkeleton />
          </>
        ) : (
          <>
            <AdminMetricCard
              title="DAU"
              value={activeUsers?.dau || 0}
              icon={<Activity className="h-5 w-5 text-blue-600" />}
              iconBgColor="bg-blue-100"
              subtitle="Daily active users"
            />
            <AdminMetricCard
              title="WAU"
              value={activeUsers?.wau || 0}
              icon={<CalendarDays className="h-5 w-5 text-purple-600" />}
              iconBgColor="bg-purple-100"
              subtitle="Weekly active users"
            />
            <AdminMetricCard
              title="MAU"
              value={activeUsers?.mau || 0}
              icon={<TrendingUp className="h-5 w-5 text-amber-600" />}
              iconBgColor="bg-amber-100"
              subtitle="Monthly active users"
            />
          </>
        )}
      </div>

      {/* DAU/WAU and DAU/MAU Ratios */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <Card className="bg-white">
          <CardHeader>
            <CardTitle className="text-lg">DAU/WAU Ratio (Stickiness)</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold text-indigo-600">
              {activeUsers?.wau
                ? ((activeUsers.dau / activeUsers.wau) * 100).toFixed(1)
                : 0}
              %
            </div>
            <p className="mt-1 text-sm text-slate-500">
              Higher ratio indicates better daily engagement
            </p>
          </CardContent>
        </Card>

        <Card className="bg-white">
          <CardHeader>
            <CardTitle className="text-lg">DAU/MAU Ratio (Monthly Stickiness)</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold text-purple-600">
              {activeUsers?.mau
                ? ((activeUsers.dau / activeUsers.mau) * 100).toFixed(1)
                : 0}
              %
            </div>
            <p className="mt-1 text-sm text-slate-500">
              Industry benchmark: 10-20% is healthy
            </p>
          </CardContent>
        </Card>
      </div>

      {/* Top Users */}
      <Card className="bg-white">
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle className="text-lg">Top Users</CardTitle>
            <Select value={metric} onValueChange={setMetric}>
              <SelectTrigger className="w-40">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="events">By Events</SelectItem>
                <SelectItem value="sessions">By Sessions</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </CardHeader>
        <CardContent>
          {isTopUsersLoading ? (
            <div className="animate-pulse space-y-3">
              {[1, 2, 3, 4, 5].map((i) => (
                <div key={i} className="h-12 bg-slate-100 rounded" />
              ))}
            </div>
          ) : (
            <UsersTable title="" users={topUsers?.users || []} />
          )}
        </CardContent>
      </Card>
    </div>
  );
}
