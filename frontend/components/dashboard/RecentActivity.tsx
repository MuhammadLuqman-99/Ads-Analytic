"use client";

import { formatDistanceToNow, parseISO } from "date-fns";
import {
  RefreshCw,
  CheckCircle,
  AlertTriangle,
  XCircle,
  Clock,
} from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { mockConnectedAccounts, getPlatformName, type Platform } from "@/lib/mock-data";
import { cn } from "@/lib/utils";

interface SyncActivity {
  id: string;
  platform: Platform;
  accountName: string;
  status: "success" | "warning" | "error" | "syncing";
  timestamp: string;
  message?: string;
}

// Generate mock sync activities based on connected accounts
const generateSyncActivities = (): SyncActivity[] => {
  const activities: SyncActivity[] = mockConnectedAccounts.map((account) => ({
    id: account.id,
    platform: account.platform,
    accountName: account.accountName,
    status: account.status === "connected" ? "success" : "error",
    timestamp: account.lastSync,
    message:
      account.status === "connected"
        ? "Data synced successfully"
        : "Connection error",
  }));

  // Add some warning/error activities for demo
  activities.push({
    id: "warn-1",
    platform: "meta",
    accountName: "My Business - Meta Ads",
    status: "warning",
    timestamp: new Date(Date.now() - 3600000 * 2).toISOString(),
    message: "Rate limit reached, sync delayed",
  });

  return activities.sort(
    (a, b) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime()
  );
};

const statusConfig = {
  success: {
    icon: CheckCircle,
    color: "text-emerald-600",
    bg: "bg-emerald-50",
    label: "Synced",
  },
  warning: {
    icon: AlertTriangle,
    color: "text-amber-600",
    bg: "bg-amber-50",
    label: "Warning",
  },
  error: {
    icon: XCircle,
    color: "text-rose-600",
    bg: "bg-rose-50",
    label: "Error",
  },
  syncing: {
    icon: RefreshCw,
    color: "text-blue-600",
    bg: "bg-blue-50",
    label: "Syncing",
  },
};

const platformColors: Record<Platform, string> = {
  meta: "bg-blue-100 text-blue-700",
  tiktok: "bg-slate-100 text-slate-700",
  shopee: "bg-orange-100 text-orange-700",
};

export function RecentActivity() {
  const activities = generateSyncActivities();

  return (
    <Card className="bg-white border-slate-200">
      <CardHeader className="pb-2">
        <div className="flex items-center justify-between">
          <CardTitle className="text-slate-900">Recent Activity</CardTitle>
          <Badge variant="outline" className="text-slate-500">
            <Clock className="h-3 w-3 mr-1" />
            Last hour
          </Badge>
        </div>
      </CardHeader>
      <CardContent>
        <div className="space-y-4">
          {activities.slice(0, 5).map((activity) => {
            const config = statusConfig[activity.status];
            const StatusIcon = config.icon;

            return (
              <div
                key={activity.id + activity.timestamp}
                className="flex items-start gap-3"
              >
                <div className={cn("p-2 rounded-lg", config.bg)}>
                  <StatusIcon
                    className={cn(
                      "h-4 w-4",
                      config.color,
                      activity.status === "syncing" && "animate-spin"
                    )}
                  />
                </div>
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2">
                    <Badge
                      className={cn(
                        "text-xs border-0",
                        platformColors[activity.platform]
                      )}
                    >
                      {getPlatformName(activity.platform)}
                    </Badge>
                    <span className="text-xs text-slate-400">
                      {formatDistanceToNow(parseISO(activity.timestamp), {
                        addSuffix: true,
                      })}
                    </span>
                  </div>
                  <p className="text-sm text-slate-900 mt-1 truncate">
                    {activity.accountName}
                  </p>
                  {activity.message && (
                    <p className="text-xs text-slate-500 mt-0.5">
                      {activity.message}
                    </p>
                  )}
                </div>
              </div>
            );
          })}
        </div>

        {/* Sync Summary */}
        <div className="mt-4 pt-4 border-t border-slate-100">
          <div className="flex items-center justify-between text-sm">
            <span className="text-slate-500">Sync Status</span>
            <div className="flex items-center gap-4">
              <div className="flex items-center gap-1">
                <CheckCircle className="h-4 w-4 text-emerald-500" />
                <span className="text-slate-700">3 connected</span>
              </div>
              <div className="flex items-center gap-1">
                <AlertTriangle className="h-4 w-4 text-amber-500" />
                <span className="text-slate-700">1 warning</span>
              </div>
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}

export function RecentActivitySkeleton() {
  return (
    <Card className="bg-white border-slate-200">
      <CardHeader className="pb-2">
        <div className="h-6 w-32 bg-slate-200 rounded animate-pulse" />
      </CardHeader>
      <CardContent>
        <div className="space-y-4">
          {[1, 2, 3, 4].map((i) => (
            <div key={i} className="flex items-start gap-3 animate-pulse">
              <div className="h-8 w-8 bg-slate-200 rounded-lg" />
              <div className="flex-1 space-y-2">
                <div className="h-3 w-24 bg-slate-200 rounded" />
                <div className="h-4 w-40 bg-slate-200 rounded" />
              </div>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}
