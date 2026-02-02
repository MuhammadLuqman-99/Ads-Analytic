"use client";

import { formatDistanceToNow } from "date-fns";
import {
  RefreshCw,
  Unplug,
  AlertTriangle,
  CheckCircle,
  Clock,
  MoreVertical,
  ExternalLink,
  Settings,
} from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { cn } from "@/lib/utils";
import type { Platform, ConnectedAccount, ConnectionStatus } from "@/lib/api/types";
import { getPlatformName, getPlatformBgClass, getPlatformTextClass, getPlatformInitial } from "@/lib/platform-utils";

interface ConnectionCardProps {
  account: ConnectedAccount;
  onRefresh: (accountId: string) => void;
  onDisconnect: (accountId: string) => void;
  onReconnect: (accountId: string, platform: Platform) => void;
  onViewDetails: (account: ConnectedAccount) => void;
  isRefreshing?: boolean;
}

const statusConfig: Record<
  ConnectionStatus,
  { label: string; color: string; icon: React.ElementType }
> = {
  active: { label: "Active", color: "bg-emerald-100 text-emerald-700", icon: CheckCircle },
  error: { label: "Error", color: "bg-red-100 text-red-700", icon: AlertTriangle },
  expired: { label: "Expired", color: "bg-amber-100 text-amber-700", icon: AlertTriangle },
  syncing: { label: "Syncing", color: "bg-blue-100 text-blue-700", icon: RefreshCw },
  disconnected: { label: "Disconnected", color: "bg-slate-100 text-slate-700", icon: Unplug },
};

export function ConnectionCard({
  account,
  onRefresh,
  onDisconnect,
  onReconnect,
  onViewDetails,
  isRefreshing,
}: ConnectionCardProps) {
  const status = statusConfig[account.status] || statusConfig.disconnected;
  const StatusIcon = status.icon;
  const hasError = account.status === "error" || account.status === "expired";

  return (
    <Card
      className={cn(
        "bg-white border-slate-200 hover:border-slate-300 transition-colors",
        hasError && "border-amber-200"
      )}
    >
      <CardContent className="p-6">
        {/* Header */}
        <div className="flex items-start justify-between mb-4">
          <div className="flex items-center gap-3">
            {/* Platform Logo */}
            <div
              className={cn(
                "w-12 h-12 rounded-lg flex items-center justify-center font-bold text-lg",
                getPlatformBgClass(account.platform),
                getPlatformTextClass(account.platform)
              )}
            >
              {getPlatformInitial(account.platform)}
            </div>
            <div>
              <h3 className="font-semibold text-slate-900">{account.platformAccountName}</h3>
              <p className="text-sm text-slate-500">{account.platformAccountId}</p>
            </div>
          </div>

          {/* Actions Menu */}
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" size="icon" className="h-8 w-8">
                <MoreVertical className="h-4 w-4" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuItem onClick={() => onViewDetails(account)}>
                <Settings className="h-4 w-4 mr-2" />
                View Details
              </DropdownMenuItem>
              <DropdownMenuItem onClick={() => onRefresh(account.id)}>
                <RefreshCw className="h-4 w-4 mr-2" />
                Sync Now
              </DropdownMenuItem>
              <DropdownMenuSeparator />
              <DropdownMenuItem
                onClick={() => onDisconnect(account.id)}
                className="text-red-600 focus:text-red-600"
              >
                <Unplug className="h-4 w-4 mr-2" />
                Disconnect
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>

        {/* Status and Platform */}
        <div className="flex items-center gap-2 mb-4">
          <Badge variant={account.platform}>{getPlatformName(account.platform)}</Badge>
          <Badge className={cn("gap-1", status.color)}>
            <StatusIcon
              className={cn("h-3 w-3", account.status === "syncing" && "animate-spin")}
            />
            {status.label}
          </Badge>
        </div>

        {/* Error Message */}
        {account.error && (
          <div className="mb-4 p-3 rounded-lg bg-amber-50 border border-amber-200">
            <div className="flex items-start gap-2">
              <AlertTriangle className="h-4 w-4 text-amber-600 mt-0.5 flex-shrink-0" />
              <div className="flex-1 min-w-0">
                <p className="text-sm text-amber-800">{account.error.message}</p>
                {account.error.type === "token_expired" && (
                  <Button
                    size="sm"
                    variant="outline"
                    className="mt-2 border-amber-300 text-amber-700 hover:bg-amber-100"
                    onClick={() => onReconnect(account.id, account.platform)}
                  >
                    <ExternalLink className="h-3 w-3 mr-1" />
                    Reconnect Account
                  </Button>
                )}
                {account.error.type === "permission_denied" && (
                  <p className="text-xs text-amber-600 mt-1">
                    Please check your account permissions in the platform settings.
                  </p>
                )}
                {(account.error.type === "api_error" || account.error.type === "rate_limit") && (
                  <Button
                    size="sm"
                    variant="outline"
                    className="mt-2 border-amber-300 text-amber-700 hover:bg-amber-100"
                    onClick={() => onRefresh(account.id)}
                  >
                    <RefreshCw className="h-3 w-3 mr-1" />
                    Retry
                  </Button>
                )}
              </div>
            </div>
          </div>
        )}

        {/* Last Sync */}
        <div className="flex items-center justify-between text-sm">
          <div className="flex items-center gap-1 text-slate-500">
            <Clock className="h-4 w-4" />
            {account.lastSyncAt ? (
              <span>Synced {formatDistanceToNow(new Date(account.lastSyncAt), { addSuffix: true })}</span>
            ) : (
              <span>Never synced</span>
            )}
          </div>

          {!hasError && (
            <Button
              variant="ghost"
              size="sm"
              onClick={() => onRefresh(account.id)}
              disabled={isRefreshing || account.status === "syncing"}
              className="h-8"
            >
              <RefreshCw
                className={cn("h-4 w-4 mr-1", isRefreshing && "animate-spin")}
              />
              Refresh
            </Button>
          )}
        </div>
      </CardContent>
    </Card>
  );
}
