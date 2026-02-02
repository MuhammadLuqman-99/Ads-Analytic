"use client";

import { Activity, TrendingUp, Clock } from "lucide-react";
import { AdminMetricCard, AdminMetricCardSkeleton } from "@/components/admin";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { useEventBreakdown } from "@/lib/api/hooks/use-admin";

const eventColors: Record<string, string> = {
  user_registered: "bg-emerald-500",
  user_logged_in: "bg-blue-500",
  platform_connected: "bg-indigo-500",
  platform_disconnected: "bg-rose-500",
  dashboard_viewed: "bg-purple-500",
  campaign_exported: "bg-amber-500",
  report_generated: "bg-cyan-500",
  settings_updated: "bg-slate-500",
  plan_upgraded: "bg-green-500",
  plan_downgraded: "bg-red-500",
};

const eventDescriptions: Record<string, string> = {
  user_registered: "New user signups",
  user_logged_in: "User login events",
  platform_connected: "Platform integrations",
  platform_disconnected: "Disconnected platforms",
  dashboard_viewed: "Dashboard page views",
  campaign_exported: "Campaign data exports",
  report_generated: "Reports created",
  settings_updated: "Settings changes",
  plan_upgraded: "Plan upgrades",
  plan_downgraded: "Plan downgrades",
};

export default function AdminEventsPage() {
  const { data: events, isLoading } = useEventBreakdown();

  const eventData = events?.events || {};
  const totalEvents = Object.values(eventData).reduce((sum, count) => sum + count, 0);
  const eventTypes = Object.keys(eventData).length;

  const formatEventName = (name: string) =>
    name
      .replace(/_/g, " ")
      .replace(/\b\w/g, (c) => c.toUpperCase());

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-2xl font-bold text-slate-900">Event Analytics</h1>
        <p className="mt-1 text-sm text-slate-600">
          Track all user events and actions in the system
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
              title="Total Events"
              value={totalEvents}
              icon={<Activity className="h-5 w-5 text-indigo-600" />}
              iconBgColor="bg-indigo-100"
              subtitle="All tracked events"
            />
            <AdminMetricCard
              title="Event Types"
              value={eventTypes}
              icon={<TrendingUp className="h-5 w-5 text-emerald-600" />}
              iconBgColor="bg-emerald-100"
              subtitle="Unique event types"
            />
            <AdminMetricCard
              title="Avg Events/Day"
              value={Math.round(totalEvents / 30)}
              icon={<Clock className="h-5 w-5 text-blue-600" />}
              iconBgColor="bg-blue-100"
              subtitle="Last 30 days"
            />
          </>
        )}
      </div>

      {/* Events Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {isLoading
          ? Array(9)
              .fill(0)
              .map((_, i) => (
                <Card key={i} className="bg-white animate-pulse">
                  <CardContent className="p-4">
                    <div className="h-4 w-24 bg-slate-200 rounded" />
                    <div className="mt-2 h-6 w-16 bg-slate-200 rounded" />
                  </CardContent>
                </Card>
              ))
          : Object.entries(eventData)
              .sort(([, a], [, b]) => b - a)
              .map(([event, count]) => {
                const percent = totalEvents > 0 ? (count / totalEvents) * 100 : 0;
                const color = eventColors[event] || "bg-slate-500";
                const description = eventDescriptions[event] || "Event activity";

                return (
                  <Card key={event} className="bg-white hover:shadow-md transition-shadow">
                    <CardContent className="p-4">
                      <div className="flex items-start justify-between">
                        <div className={`h-3 w-3 rounded-full ${color}`} />
                        <Badge variant="secondary" className="text-xs">
                          {percent.toFixed(1)}%
                        </Badge>
                      </div>
                      <div className="mt-3">
                        <div className="font-medium text-slate-900">
                          {formatEventName(event)}
                        </div>
                        <div className="text-xs text-slate-500 mt-1">{description}</div>
                      </div>
                      <div className="mt-3 text-2xl font-bold text-slate-900">
                        {count.toLocaleString()}
                      </div>
                    </CardContent>
                  </Card>
                );
              })}
      </div>

      {Object.keys(eventData).length === 0 && !isLoading && (
        <Card className="bg-white">
          <CardContent className="py-12 text-center">
            <Activity className="h-12 w-12 mx-auto text-slate-300" />
            <h3 className="mt-4 font-medium text-slate-900">No events yet</h3>
            <p className="mt-1 text-sm text-slate-500">
              Events will appear here as users interact with the platform
            </p>
          </CardContent>
        </Card>
      )}

      {/* Event Log */}
      <Card className="bg-white">
        <CardHeader>
          <CardTitle className="text-lg">Event Distribution</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-2">
            {Object.entries(eventData)
              .sort(([, a], [, b]) => b - a)
              .map(([event, count]) => {
                const percent = totalEvents > 0 ? (count / totalEvents) * 100 : 0;
                const color = eventColors[event] || "bg-slate-500";

                return (
                  <div key={event} className="flex items-center gap-3">
                    <div className={`h-2 w-2 rounded-full ${color}`} />
                    <span className="flex-1 text-sm text-slate-700">
                      {formatEventName(event)}
                    </span>
                    <div className="w-32 h-2 bg-slate-100 rounded-full overflow-hidden">
                      <div
                        className={`h-full rounded-full ${color}`}
                        style={{ width: `${percent}%` }}
                      />
                    </div>
                    <span className="text-sm text-slate-500 w-20 text-right">
                      {count.toLocaleString()}
                    </span>
                  </div>
                );
              })}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
