"use client";

import { useState } from "react";
import { UserMinus, AlertTriangle, Clock } from "lucide-react";
import {
  AdminMetricCard,
  AdminMetricCardSkeleton,
  UsersTable,
  UsersTableSkeleton,
} from "@/components/admin";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { useChurnedUsers } from "@/lib/api/hooks/use-admin";

export default function AdminChurnedPage() {
  const [days, setDays] = useState(30);
  const { data, isLoading } = useChurnedUsers(days);

  const churnedUsers = data?.users || [];

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-slate-900">Churned Users</h1>
          <p className="mt-1 text-sm text-slate-600">
            Users who haven&apos;t been active recently
          </p>
        </div>
        <Select value={days.toString()} onValueChange={(v) => setDays(parseInt(v))}>
          <SelectTrigger className="w-40">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="7">Last 7 days</SelectItem>
            <SelectItem value="14">Last 14 days</SelectItem>
            <SelectItem value="30">Last 30 days</SelectItem>
            <SelectItem value="60">Last 60 days</SelectItem>
            <SelectItem value="90">Last 90 days</SelectItem>
          </SelectContent>
        </Select>
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
              title="Churned Users"
              value={churnedUsers.length}
              icon={<UserMinus className="h-5 w-5 text-rose-600" />}
              iconBgColor="bg-rose-100"
              subtitle={`Inactive for ${days}+ days`}
            />
            <AdminMetricCard
              title="At Risk"
              value={churnedUsers.filter((u) => u.totalSessions > 5).length}
              icon={<AlertTriangle className="h-5 w-5 text-amber-600" />}
              iconBgColor="bg-amber-100"
              subtitle="Had high engagement"
            />
            <AdminMetricCard
              title="Avg Inactive Days"
              value={days}
              icon={<Clock className="h-5 w-5 text-slate-600" />}
              iconBgColor="bg-slate-100"
              subtitle="Threshold for churn"
            />
          </>
        )}
      </div>

      {/* Churned Users Table */}
      {isLoading ? (
        <UsersTableSkeleton />
      ) : (
        <UsersTable
          title={`Churned Users (${days}+ days inactive)`}
          users={churnedUsers}
          showChurnBadge
        />
      )}
    </div>
  );
}
