"use client";

import { format, formatDistanceToNow } from "date-fns";
import {
  X,
  RefreshCw,
  CheckCircle,
  XCircle,
  AlertTriangle,
  Clock,
  Calendar,
  Hash,
  ExternalLink,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";
import type { Platform, ConnectedAccount, ConnectionStatus, SyncStatus } from "@/lib/api/types";
import { getPlatformName, getPlatformBgClass, getPlatformTextClass, getPlatformInitial } from "@/lib/platform-utils";

interface AccountDetailsModalProps {
  isOpen: boolean;
  onClose: () => void;
  account: ConnectedAccount | null;
  syncStatus?: SyncStatus | null;
  onSync: (accountId: string) => void;
  onReconnect: (accountId: string, platform: Platform) => void;
  isSyncing?: boolean;
}

const statusConfig: Record<ConnectionStatus, { label: string; color: string }> = {
  active: { label: "Active", color: "bg-emerald-100 text-emerald-700" },
  error: { label: "Error", color: "bg-red-100 text-red-700" },
  expired: { label: "Expired", color: "bg-amber-100 text-amber-700" },
  syncing: { label: "Syncing", color: "bg-blue-100 text-blue-700" },
  disconnected: { label: "Disconnected", color: "bg-slate-100 text-slate-700" },
};

export function AccountDetailsModal({
  isOpen,
  onClose,
  account,
  syncStatus,
  onSync,
  onReconnect,
  isSyncing,
}: AccountDetailsModalProps) {
  if (!isOpen || !account) return null;

  const status = statusConfig[account.status] || statusConfig.disconnected;
  const hasError = account.status === "error" || account.status === "expired";

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      {/* Backdrop */}
      <div
        className="absolute inset-0 bg-black/50 backdrop-blur-sm"
        onClick={onClose}
      />

      {/* Modal */}
      <div className="relative bg-white rounded-xl shadow-xl max-w-2xl w-full mx-4 max-h-[90vh] overflow-y-auto">
        {/* Header */}
        <div className="flex items-center justify-between p-6 border-b border-slate-200">
          <div className="flex items-center gap-4">
            <div
              className={cn(
                "w-12 h-12 rounded-lg flex items-center justify-center font-bold text-xl",
                getPlatformBgClass(account.platform),
                getPlatformTextClass(account.platform)
              )}
            >
              {getPlatformInitial(account.platform)}
            </div>
            <div>
              <h2 className="text-xl font-semibold text-slate-900">
                {account.platformAccountName}
              </h2>
              <div className="flex items-center gap-2 mt-1">
                <Badge variant={account.platform}>
                  {getPlatformName(account.platform)}
                </Badge>
                <Badge className={status.color}>{status.label}</Badge>
              </div>
            </div>
          </div>
          <Button variant="ghost" size="icon" onClick={onClose}>
            <X className="h-5 w-5" />
          </Button>
        </div>

        {/* Content */}
        <div className="p-6 space-y-6">
          {/* Error Alert */}
          {account.error && (
            <div className="p-4 rounded-lg bg-amber-50 border border-amber-200">
              <div className="flex items-start gap-3">
                <AlertTriangle className="h-5 w-5 text-amber-600 mt-0.5" />
                <div className="flex-1">
                  <p className="font-medium text-amber-800">
                    {account.error.type === "token_expired" && "Token Expired"}
                    {account.error.type === "permission_denied" && "Permission Denied"}
                    {account.error.type === "api_error" && "API Error"}
                    {account.error.type === "rate_limit" && "Rate Limited"}
                    {account.error.type === "account_suspended" && "Account Suspended"}
                  </p>
                  <p className="text-sm text-amber-700 mt-1">{account.error.message}</p>
                  <p className="text-xs text-amber-600 mt-2">
                    Occurred {formatDistanceToNow(new Date(account.error.occurredAt), { addSuffix: true })}
                  </p>
                </div>
              </div>
            </div>
          )}

          {/* Account Info */}
          <div>
            <h3 className="text-sm font-medium text-slate-500 mb-3">Account Information</h3>
            <div className="grid grid-cols-2 gap-4">
              <div className="p-3 bg-slate-50 rounded-lg">
                <div className="flex items-center gap-2 text-slate-500 mb-1">
                  <Hash className="h-4 w-4" />
                  <span className="text-xs">Account ID</span>
                </div>
                <p className="font-mono text-sm text-slate-900">{account.platformAccountId}</p>
              </div>
              <div className="p-3 bg-slate-50 rounded-lg">
                <div className="flex items-center gap-2 text-slate-500 mb-1">
                  <Calendar className="h-4 w-4" />
                  <span className="text-xs">Connected</span>
                </div>
                <p className="text-sm text-slate-900">
                  {format(new Date(account.connectedAt), "MMM d, yyyy")}
                </p>
              </div>
              <div className="p-3 bg-slate-50 rounded-lg">
                <div className="flex items-center gap-2 text-slate-500 mb-1">
                  <Clock className="h-4 w-4" />
                  <span className="text-xs">Last Sync</span>
                </div>
                <p className="text-sm text-slate-900">
                  {account.lastSyncAt
                    ? formatDistanceToNow(new Date(account.lastSyncAt), { addSuffix: true })
                    : "Never"}
                </p>
              </div>
              <div className="p-3 bg-slate-50 rounded-lg">
                <div className="flex items-center gap-2 text-slate-500 mb-1">
                  <RefreshCw className="h-4 w-4" />
                  <span className="text-xs">Sync Status</span>
                </div>
                <p className="text-sm text-slate-900">
                  {syncStatus?.status === "syncing"
                    ? `In Progress (${syncStatus.progress}%)`
                    : account.status === "syncing"
                      ? "In Progress"
                      : "Idle"}
                </p>
              </div>
            </div>
          </div>

          {/* Sync Status */}
          <div>
            <h3 className="text-sm font-medium text-slate-500 mb-3">Last Sync</h3>
            {!syncStatus && !account.lastSuccessfulSyncAt ? (
              <p className="text-sm text-slate-400 text-center py-6">
                No sync history available
              </p>
            ) : (
              <div className="space-y-2">
                {syncStatus && (
                  <div className="flex items-center justify-between p-3 bg-slate-50 rounded-lg">
                    <div className="flex items-center gap-3">
                      {syncStatus.status === "completed" && (
                        <CheckCircle className="h-5 w-5 text-emerald-500" />
                      )}
                      {syncStatus.status === "failed" && (
                        <XCircle className="h-5 w-5 text-red-500" />
                      )}
                      {syncStatus.status === "syncing" && (
                        <RefreshCw className="h-5 w-5 text-blue-500 animate-spin" />
                      )}
                      {syncStatus.status === "idle" && (
                        <Clock className="h-5 w-5 text-slate-400" />
                      )}
                      <div>
                        <p className="text-sm font-medium text-slate-900">
                          {syncStatus.status === "completed" && "Sync Completed"}
                          {syncStatus.status === "failed" && "Sync Failed"}
                          {syncStatus.status === "syncing" && "Syncing..."}
                          {syncStatus.status === "idle" && "Ready"}
                        </p>
                        {syncStatus.completedAt && (
                          <p className="text-xs text-slate-500">
                            {format(new Date(syncStatus.completedAt), "MMM d, yyyy 'at' h:mm a")}
                          </p>
                        )}
                      </div>
                    </div>
                    <div className="text-right">
                      {syncStatus.syncedRecords !== undefined && (
                        <p className="text-sm text-slate-900">
                          {syncStatus.syncedRecords.toLocaleString()} records
                        </p>
                      )}
                      {syncStatus.error && (
                        <p className="text-xs text-red-500 max-w-[200px] truncate">
                          {syncStatus.error}
                        </p>
                      )}
                    </div>
                  </div>
                )}
              </div>
            )}
          </div>

          {/* Actions */}
          <div className="flex items-center justify-between pt-4 border-t border-slate-200">
            <Button variant="outline" onClick={onClose}>
              Close
            </Button>
            <div className="flex items-center gap-3">
              {hasError && account.error?.type === "token_expired" && (
                <Button onClick={() => onReconnect(account.id, account.platform)}>
                  <ExternalLink className="h-4 w-4 mr-2" />
                  Reconnect Account
                </Button>
              )}
              {!hasError && (
                <Button
                  onClick={() => onSync(account.id)}
                  disabled={isSyncing || account.status === "syncing"}
                >
                  <RefreshCw
                    className={cn("h-4 w-4 mr-2", isSyncing && "animate-spin")}
                  />
                  Sync Now
                </Button>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
