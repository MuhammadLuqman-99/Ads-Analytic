"use client";

import { useState } from "react";
import { Plus, Zap } from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
  ConnectionCard,
  AddConnectionModal,
  AccountDetailsModal,
  mockAdAccounts,
  mockSyncHistory,
  mockPlanLimits,
  type AdAccount,
  type PlanLimits,
} from "@/components/connections";
import { type Platform } from "@/lib/mock-data";

export default function ConnectionsPage() {
  const [accounts, setAccounts] = useState<AdAccount[]>(mockAdAccounts);
  const [planLimits, setPlanLimits] = useState<PlanLimits>(mockPlanLimits);
  const [isAddModalOpen, setIsAddModalOpen] = useState(false);
  const [selectedAccount, setSelectedAccount] = useState<AdAccount | null>(null);
  const [isDetailsModalOpen, setIsDetailsModalOpen] = useState(false);
  const [refreshingId, setRefreshingId] = useState<string | null>(null);

  const handleRefresh = async (accountId: string) => {
    setRefreshingId(accountId);
    // Update status to syncing
    setAccounts((prev) =>
      prev.map((acc) =>
        acc.id === accountId ? { ...acc, status: "syncing" as const } : acc
      )
    );

    // Simulate sync
    await new Promise((resolve) => setTimeout(resolve, 2000));

    // Update with success
    setAccounts((prev) =>
      prev.map((acc) =>
        acc.id === accountId
          ? { ...acc, status: "active" as const, lastSyncAt: new Date(), error: undefined }
          : acc
      )
    );
    setRefreshingId(null);
  };

  const handleDisconnect = (accountId: string) => {
    if (confirm("Are you sure you want to disconnect this account?")) {
      setAccounts((prev) => prev.filter((acc) => acc.id !== accountId));
      setPlanLimits((prev) => ({
        ...prev,
        accountsUsed: Math.max(0, prev.accountsUsed - 1),
      }));
    }
  };

  const handleReconnect = async (accountId: string) => {
    // Simulate OAuth flow
    await new Promise((resolve) => setTimeout(resolve, 1500));

    // Update account status
    setAccounts((prev) =>
      prev.map((acc) =>
        acc.id === accountId
          ? { ...acc, status: "active" as const, error: undefined, lastSyncAt: new Date() }
          : acc
      )
    );
  };

  const handleConnect = async (platform: Platform) => {
    // Simulate OAuth flow
    await new Promise((resolve) => setTimeout(resolve, 2000));

    // Add new account
    const newAccount: AdAccount = {
      id: String(Date.now()),
      platform,
      accountId: `${platform}_${Math.random().toString(36).substr(2, 9)}`,
      accountName: `New ${platform.charAt(0).toUpperCase() + platform.slice(1)} Account`,
      status: "active",
      lastSyncAt: new Date(),
      connectedAt: new Date(),
    };

    setAccounts((prev) => [...prev, newAccount]);
    setPlanLimits((prev) => ({
      ...prev,
      accountsUsed: prev.accountsUsed + 1,
    }));
  };

  const handleViewDetails = (account: AdAccount) => {
    setSelectedAccount(account);
    setIsDetailsModalOpen(true);
  };

  const activeAccounts = accounts.filter((a) => a.status === "active").length;
  const errorAccounts = accounts.filter(
    (a) => a.status === "error" || a.status === "expired"
  ).length;

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
        <div>
          <h1 className="text-3xl font-bold text-slate-900">Connections</h1>
          <p className="mt-1 text-slate-500">
            Manage your connected advertising accounts
          </p>
        </div>
        <Button onClick={() => setIsAddModalOpen(true)}>
          <Plus className="h-4 w-4 mr-2" />
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
                  className="h-full bg-blue-500 rounded-full"
                  style={{
                    width: `${(planLimits.accountsUsed / planLimits.accountsLimit) * 100}%`,
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
                <p className="text-2xl font-bold text-emerald-600">{activeAccounts}</p>
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
                <p className="text-2xl font-bold text-amber-600">{errorAccounts}</p>
              </div>
              {errorAccounts > 0 && (
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
        {accounts.map((account) => (
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
              <Button variant="outline" onClick={() => setIsAddModalOpen(true)}>
                Connect Account
              </Button>
            </CardContent>
          </Card>
        )}
      </div>

      {/* Empty State */}
      {accounts.length === 0 && (
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
            <Button onClick={() => setIsAddModalOpen(true)}>
              <Plus className="h-4 w-4 mr-2" />
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
        syncHistory={mockSyncHistory}
        onSync={handleRefresh}
        onReconnect={handleReconnect}
        isSyncing={refreshingId === selectedAccount?.id}
      />
    </div>
  );
}
