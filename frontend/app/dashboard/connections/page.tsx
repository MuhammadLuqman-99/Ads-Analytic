"use client";

import { useEffect, useState } from "react";
import { useSearchParams } from "next/navigation";
import { Plus, Zap, Loader2, AlertCircle, CheckCircle2 } from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
  ConnectionCard,
  AddConnectionModal,
  AccountDetailsModal,
  type PlanLimits,
} from "@/components/connections";
import { useConnections, useSyncStatus } from "@/lib/api/hooks";
import type { Platform, ConnectedAccount } from "@/lib/api/types";

// Toast notification component
function Toast({
  message,
  type,
  onClose,
}: {
  message: string;
  type: "success" | "error";
  onClose: () => void;
}) {
  useEffect(() => {
    const timer = setTimeout(onClose, 5000);
    return () => clearTimeout(timer);
  }, [onClose]);

  return (
    <div
      className={`fixed bottom-4 right-4 z-50 flex items-center gap-3 px-4 py-3 rounded-lg shadow-lg ${
        type === "success"
          ? "bg-emerald-50 text-emerald-800 border border-emerald-200"
          : "bg-red-50 text-red-800 border border-red-200"
      }`}
    >
      {type === "success" ? (
        <CheckCircle2 className="h-5 w-5 text-emerald-600" />
      ) : (
        <AlertCircle className="h-5 w-5 text-red-600" />
      )}
      <p className="text-sm font-medium">{message}</p>
      <button
        onClick={onClose}
        className="ml-2 text-current opacity-50 hover:opacity-100"
      >
        ×
      </button>
    </div>
  );
}

// Error messages mapping
const errorMessages: Record<string, string> = {
  permission_denied: "You denied permission to access your ad account. Please try again and grant the required permissions.",
  access_denied: "Access was denied. Please try again.",
  invalid_state: "Your session expired. Please try connecting again.",
  state_expired: "Your session expired. Please try connecting again.",
  token_exchange_failed: "Failed to connect your account. Please try again later.",
  callback_failed: "Something went wrong. Please try again.",
  no_code: "Authorization was not completed. Please try again.",
  platform_mismatch: "Platform mismatch detected. Please try again.",
};

export default function ConnectionsPage() {
  const searchParams = useSearchParams();
  const [isAddModalOpen, setIsAddModalOpen] = useState(false);
  const [selectedAccount, setSelectedAccount] = useState<ConnectedAccount | null>(null);
  const [isDetailsModalOpen, setIsDetailsModalOpen] = useState(false);
  const [refreshingId, setRefreshingId] = useState<string | null>(null);
  const [toast, setToast] = useState<{
    message: string;
    type: "success" | "error";
  } | null>(null);

  // Use real API hook
  const {
    connections,
    total,
    activeCount,
    errorCount,
    isLoading,
    connect,
    disconnect,
    sync,
    reconnect,
    isConnecting,
    refetch,
  } = useConnections();

  // Get sync status for selected account
  const { status: syncStatus } = useSyncStatus(selectedAccount?.id || "");

  // Handle OAuth callback query params
  useEffect(() => {
    const success = searchParams.get("success");
    const error = searchParams.get("error");
    const platform = searchParams.get("platform");
    const message = searchParams.get("message");
    const accountId = searchParams.get("account_id");

    if (success === "true" && platform) {
      setToast({
        message: `Successfully connected your ${platform.charAt(0).toUpperCase() + platform.slice(1)} account!`,
        type: "success",
      });
      // Refresh connections list
      refetch();
      // Trigger initial sync if we have the account ID
      if (accountId) {
        sync(accountId).catch(console.error);
      }
      // Clear query params
      window.history.replaceState({}, "", "/dashboard/connections");
    } else if (error) {
      const errorMessage = errorMessages[error] || message || "Failed to connect account. Please try again.";
      setToast({
        message: errorMessage,
        type: "error",
      });
      // Clear query params
      window.history.replaceState({}, "", "/dashboard/connections");
    }
  }, [searchParams, refetch, sync]);

  // Plan limits (TODO: fetch from API)
  const planLimits: PlanLimits = {
    accountsLimit: 5,
    accountsUsed: total,
    currentPlan: "free",
  };

  const handleRefresh = async (accountId: string) => {
    setRefreshingId(accountId);
    try {
      await sync(accountId);
      setToast({ message: "Sync started successfully", type: "success" });
    } catch (error) {
      setToast({ message: "Failed to start sync", type: "error" });
    } finally {
      setRefreshingId(null);
    }
  };

  const handleDisconnect = async (accountId: string) => {
    if (confirm("Are you sure you want to disconnect this account?")) {
      try {
        await disconnect(accountId);
        setToast({ message: "Account disconnected successfully", type: "success" });
      } catch (error) {
        setToast({ message: "Failed to disconnect account", type: "error" });
      }
    }
  };

  const handleReconnect = async (accountId: string, platform: Platform) => {
    try {
      await reconnect({ accountId, platform });
    } catch (error) {
      setToast({ message: "Failed to initiate reconnection", type: "error" });
    }
  };

  const handleConnect = async (platform: Platform) => {
    try {
      await connect(platform);
      // The hook will redirect to OAuth URL
    } catch (error) {
      setToast({ message: "Failed to initiate connection", type: "error" });
    }
  };

  const handleViewDetails = (account: ConnectedAccount) => {
    setSelectedAccount(account);
    setIsDetailsModalOpen(true);
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <Loader2 className="h-8 w-8 animate-spin text-blue-600" />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Toast Notification */}
      {toast && (
        <Toast
          message={toast.message}
          type={toast.type}
          onClose={() => setToast(null)}
        />
      )}

      {/* Header */}
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
        <div>
          <h1 className="text-3xl font-bold text-slate-900">Connections</h1>
          <p className="mt-1 text-slate-500">
            Manage your connected advertising accounts
          </p>
        </div>
        <Button
          onClick={() => setIsAddModalOpen(true)}
          disabled={isConnecting}
        >
          {isConnecting ? (
            <Loader2 className="h-4 w-4 mr-2 animate-spin" />
          ) : (
            <Plus className="h-4 w-4 mr-2" />
          )}
          Connect Account
        </Button>
      </div>

      {/* Stats Bar */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <Card className="bg-white border-slate-200">
          <CardContent className="p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-slate-500">Connected Accounts</p>
                <p className="text-2xl font-bold text-slate-900">
                  {planLimits.accountsUsed} / {planLimits.accountsLimit}
                </p>
              </div>
              <div className="h-12 w-12 rounded-full bg-blue-100 flex items-center justify-center">
                <Zap className="h-6 w-6 text-blue-600" />
              </div>
            </div>
            <div className="mt-3">
              <div className="h-2 bg-slate-100 rounded-full overflow-hidden">
                <div
                  className="h-full bg-blue-500 rounded-full transition-all"
                  style={{
                    width: `${Math.min((planLimits.accountsUsed / planLimits.accountsLimit) * 100, 100)}%`,
                  }}
                />
              </div>
              <p className="text-xs text-slate-500 mt-1">
                {planLimits.currentPlan.charAt(0).toUpperCase() +
                  planLimits.currentPlan.slice(1)}{" "}
                Plan
              </p>
            </div>
          </CardContent>
        </Card>

        <Card className="bg-white border-slate-200">
          <CardContent className="p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-slate-500">Active</p>
                <p className="text-2xl font-bold text-emerald-600">{activeCount}</p>
              </div>
              <Badge className="bg-emerald-100 text-emerald-700">Syncing</Badge>
            </div>
          </CardContent>
        </Card>

        <Card className="bg-white border-slate-200">
          <CardContent className="p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-slate-500">Needs Attention</p>
                <p className="text-2xl font-bold text-amber-600">{errorCount}</p>
              </div>
              {errorCount > 0 && (
                <Badge className="bg-amber-100 text-amber-700">Action Required</Badge>
              )}
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Upgrade CTA if at limit */}
      {planLimits.accountsUsed >= planLimits.accountsLimit && (
        <Card className="bg-gradient-to-r from-blue-500 to-purple-600 border-0">
          <CardContent className="p-6">
            <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
              <div className="text-white">
                <h3 className="text-lg font-semibold">
                  You&apos;ve reached your account limit
                </h3>
                <p className="text-blue-100 mt-1">
                  Upgrade to connect more ad accounts and unlock advanced features.
                </p>
              </div>
              <Button
                variant="secondary"
                className="bg-white text-blue-600 hover:bg-blue-50"
                onClick={() => (window.location.href = "/dashboard/billing")}
              >
                Upgrade Plan
              </Button>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Connections Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {connections.map((account) => (
          <ConnectionCard
            key={account.id}
            account={account}
            onRefresh={handleRefresh}
            onDisconnect={handleDisconnect}
            onReconnect={handleReconnect}
            onViewDetails={handleViewDetails}
            isRefreshing={refreshingId === account.id}
          />
        ))}

        {/* Add New Connection Card */}
        {planLimits.accountsUsed < planLimits.accountsLimit && (
          <Card className="bg-white border-slate-200 border-dashed hover:border-slate-300 transition-colors">
            <CardContent className="p-6 flex flex-col items-center justify-center h-full min-h-[280px]">
              <div className="w-12 h-12 rounded-full bg-slate-100 flex items-center justify-center mb-4">
                <Plus className="h-6 w-6 text-slate-400" />
              </div>
              <p className="font-medium text-slate-900 mb-1">Add Connection</p>
              <p className="text-sm text-slate-500 text-center mb-4">
                Connect a new advertising platform
              </p>
              <Button
                variant="outline"
                onClick={() => setIsAddModalOpen(true)}
                disabled={isConnecting}
              >
                Connect Account
              </Button>
            </CardContent>
          </Card>
        )}
      </div>

      {/* Empty State */}
      {connections.length === 0 && (
        <Card className="bg-white border-slate-200 border-dashed">
          <CardContent className="py-16 flex flex-col items-center justify-center">
            <div className="w-16 h-16 rounded-full bg-slate-100 flex items-center justify-center mb-4">
              <Zap className="h-8 w-8 text-slate-400" />
            </div>
            <h3 className="text-lg font-semibold text-slate-900 mb-2">
              No Accounts Connected
            </h3>
            <p className="text-slate-500 text-center max-w-md mb-6">
              Connect your first ad account to start tracking campaigns and
              performance metrics across platforms.
            </p>
            <Button
              onClick={() => setIsAddModalOpen(true)}
              disabled={isConnecting}
            >
              {isConnecting ? (
                <Loader2 className="h-4 w-4 mr-2 animate-spin" />
              ) : (
                <Plus className="h-4 w-4 mr-2" />
              )}
              Connect Your First Account
            </Button>
          </CardContent>
        </Card>
      )}

      {/* Modals */}
      <AddConnectionModal
        isOpen={isAddModalOpen}
        onClose={() => setIsAddModalOpen(false)}
        onConnect={handleConnect}
        planLimits={planLimits}
      />

      <AccountDetailsModal
        isOpen={isDetailsModalOpen}
        onClose={() => {
          setIsDetailsModalOpen(false);
          setSelectedAccount(null);
        }}
        account={selectedAccount}
        syncStatus={syncStatus}
        onSync={handleRefresh}
        onReconnect={handleReconnect}
        isSyncing={refreshingId === selectedAccount?.id}
      />
    </div>
  );
}
