"use client";

import { useState } from "react";
import {
  RefreshCw,
  CheckCircle2,
  AlertCircle,
  Wifi,
  WifiOff,
  ChevronDown,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import { cn } from "@/lib/utils";
import { useEventStream, type SyncStatus } from "@/lib/api/hooks/use-event-stream";
import { formatDistanceToNow } from "date-fns";

interface PlatformConfig {
  name: string;
  icon: string;
  color: string;
}

const platformConfigs: Record<string, PlatformConfig> = {
  meta: { name: "Meta Ads", icon: "M", color: "bg-blue-600" },
  tiktok: { name: "TikTok Ads", icon: "T", color: "bg-black" },
  shopee: { name: "Shopee Ads", icon: "S", color: "bg-orange-500" },
};

interface SyncStatusIndicatorProps {
  onManualSync?: () => void;
  isSyncingManual?: boolean;
  lastSyncedAt?: Date | string | null;
}

export function SyncStatusIndicator({
  onManualSync,
  isSyncingManual = false,
  lastSyncedAt,
}: SyncStatusIndicatorProps) {
  const [open, setOpen] = useState(false);

  const {
    isConnected,
    isSyncing,
    activeSyncs,
    syncStatuses,
  } = useEventStream({ enabled: true, showToasts: true });

  // Format last synced time
  const formatLastSynced = (date: Date | string | null | undefined) => {
    if (!date) return "Never";
    const d = typeof date === "string" ? new Date(date) : date;
    return formatDistanceToNow(d, { addSuffix: true });
  };

  // Get status icon
  const getStatusIcon = (status: SyncStatus["status"]) => {
    switch (status) {
      case "syncing":
        return <RefreshCw className="h-3 w-3 animate-spin" />;
      case "completed":
        return <CheckCircle2 className="h-3 w-3 text-emerald-500" />;
      case "error":
        return <AlertCircle className="h-3 w-3 text-red-500" />;
      default:
        return null;
    }
  };

  // Determine overall status
  const hasError = syncStatuses.some((s) => s.status === "error");
  const isAnySyncing = isSyncing || isSyncingManual;

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          variant="outline"
          size="sm"
          className={cn(
            "gap-2",
            isAnySyncing && "border-blue-300 bg-blue-50",
            hasError && !isAnySyncing && "border-red-300 bg-red-50"
          )}
        >
          {/* Connection status indicator */}
          {isConnected ? (
            <Wifi className="h-3 w-3 text-emerald-500" />
          ) : (
            <WifiOff className="h-3 w-3 text-slate-400" />
          )}

          {/* Sync status */}
          {isAnySyncing ? (
            <>
              <RefreshCw className="h-4 w-4 animate-spin text-blue-600" />
              <span className="text-blue-600">Syncing...</span>
            </>
          ) : hasError ? (
            <>
              <AlertCircle className="h-4 w-4 text-red-500" />
              <span className="text-red-600">Sync Error</span>
            </>
          ) : (
            <>
              <span className="text-slate-600">
                {formatLastSynced(lastSyncedAt)}
              </span>
            </>
          )}

          <ChevronDown className="h-3 w-3 text-slate-400" />
        </Button>
      </PopoverTrigger>

      <PopoverContent className="w-80 p-0" align="end">
        <div className="p-4 border-b border-slate-200">
          <div className="flex items-center justify-between">
            <h4 className="font-semibold text-slate-900">Sync Status</h4>
            <div className="flex items-center gap-2 text-xs">
              {isConnected ? (
                <span className="flex items-center gap-1 text-emerald-600">
                  <span className="h-2 w-2 rounded-full bg-emerald-500 animate-pulse" />
                  Connected
                </span>
              ) : (
                <span className="flex items-center gap-1 text-slate-400">
                  <span className="h-2 w-2 rounded-full bg-slate-300" />
                  Disconnected
                </span>
              )}
            </div>
          </div>
          {lastSyncedAt && (
            <p className="text-xs text-slate-500 mt-1">
              Last synced {formatLastSynced(lastSyncedAt)}
            </p>
          )}
        </div>

        {/* Platform sync statuses */}
        <div className="p-2 max-h-60 overflow-auto">
          {syncStatuses.length > 0 ? (
            <div className="space-y-1">
              {syncStatuses.map((status) => {
                const config = platformConfigs[status.platform] || {
                  name: status.platform,
                  icon: status.platform[0].toUpperCase(),
                  color: "bg-slate-500",
                };

                return (
                  <div
                    key={`${status.platform}-${status.accountId}`}
                    className="flex items-center gap-3 p-2 rounded-lg hover:bg-slate-50"
                  >
                    <div
                      className={cn(
                        "h-8 w-8 rounded-md flex items-center justify-center text-white text-sm font-bold",
                        config.color
                      )}
                    >
                      {config.icon}
                    </div>
                    <div className="flex-1 min-w-0">
                      <p className="text-sm font-medium text-slate-900 truncate">
                        {config.name}
                      </p>
                      <p className="text-xs text-slate-500 truncate">
                        {status.status === "syncing"
                          ? status.message || `${status.progress}% complete`
                          : status.status === "error"
                          ? status.error
                          : status.lastSyncedAt
                          ? formatLastSynced(status.lastSyncedAt)
                          : "Ready"}
                      </p>
                    </div>
                    <div className="flex items-center gap-2">
                      {status.status === "syncing" && (
                        <div className="w-16 h-1.5 bg-slate-200 rounded-full overflow-hidden">
                          <div
                            className="h-full bg-blue-500 transition-all duration-300"
                            style={{ width: `${status.progress}%` }}
                          />
                        </div>
                      )}
                      {getStatusIcon(status.status)}
                    </div>
                  </div>
                );
              })}
            </div>
          ) : (
            <div className="py-8 text-center text-slate-500">
              <RefreshCw className="h-8 w-8 mx-auto mb-2 text-slate-300" />
              <p className="text-sm">No recent sync activity</p>
            </div>
          )}
        </div>

        {/* Manual sync button */}
        <div className="p-3 border-t border-slate-200 bg-slate-50">
          <Button
            onClick={() => {
              onManualSync?.();
              setOpen(false);
            }}
            disabled={isAnySyncing}
            className="w-full"
            size="sm"
          >
            {isAnySyncing ? (
              <>
                <RefreshCw className="h-4 w-4 mr-2 animate-spin" />
                Syncing...
              </>
            ) : (
              <>
                <RefreshCw className="h-4 w-4 mr-2" />
                Sync All Accounts
              </>
            )}
          </Button>
        </div>
      </PopoverContent>
    </Popover>
  );
}

export default SyncStatusIndicator;
