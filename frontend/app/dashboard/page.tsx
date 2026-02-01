"use client";

import { useState } from "react";
import {
  RefreshCw,
  DollarSign,
  TrendingUp,
  ShoppingCart,
  Wallet,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  MetricsCard,
  MetricsCardSkeleton,
  SpendChart,
  PlatformCards,
  TopCampaigns,
  RecentActivity,
} from "@/components/dashboard";
import { useDashboardMetrics } from "@/hooks/use-metrics";
import { useSyncAllAccounts } from "@/hooks/use-accounts";
import { formatCurrency, formatNumber, type Platform } from "@/lib/mock-data";

export default function DashboardPage() {
  const { data: metrics, isLoading } = useDashboardMetrics();
  const syncMutation = useSyncAllAccounts();
  const [selectedPlatform, setSelectedPlatform] = useState<Platform | null>(null);

  const handlePlatformClick = (platform: Platform) => {
    setSelectedPlatform(selectedPlatform === platform ? null : platform);
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
        <div>
          <h1 className="text-3xl font-bold text-slate-900">Dashboard</h1>
          <p className="mt-1 text-slate-500">
            Overview of your ad performance across all platforms
          </p>
        </div>
        <Button
          onClick={() => syncMutation.mutate()}
          disabled={syncMutation.isPending}
        >
          <RefreshCw
            className={`h-4 w-4 mr-2 ${syncMutation.isPending ? "animate-spin" : ""}`}
          />
          {syncMutation.isPending ? "Syncing..." : "Sync Now"}
        </Button>
      </div>

      {/* Key Metrics Row */}
      {isLoading ? (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
          {[1, 2, 3, 4].map((i) => (
            <MetricsCardSkeleton key={i} />
          ))}
        </div>
      ) : (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
          <MetricsCard
            title="Total Spend"
            value={formatCurrency(metrics?.totalSpend || 0)}
            change={metrics?.spendChange || 0}
            icon={<Wallet className="h-5 w-5 text-white" />}
            iconBgColor="bg-gradient-to-br from-indigo-500 to-indigo-600"
            subtitle="vs last period"
          />
          <MetricsCard
            title="Total Revenue"
            value={formatCurrency(metrics?.totalRevenue || 0)}
            change={metrics?.revenueChange || 0}
            icon={<DollarSign className="h-5 w-5 text-white" />}
            iconBgColor="bg-gradient-to-br from-emerald-500 to-emerald-600"
            subtitle="vs last period"
          />
          <MetricsCard
            title="Overall ROAS"
            value={`${metrics?.averageRoas?.toFixed(1) || 0}x`}
            change={metrics?.roasChange || 0}
            icon={<TrendingUp className="h-5 w-5 text-white" />}
            iconBgColor="bg-gradient-to-br from-amber-500 to-amber-600"
            subtitle="Return on ad spend"
          />
          <MetricsCard
            title="Total Conversions"
            value={formatNumber(metrics?.totalConversions || 0)}
            change={metrics?.conversionsChange || 0}
            icon={<ShoppingCart className="h-5 w-5 text-white" />}
            iconBgColor="bg-gradient-to-br from-purple-500 to-purple-600"
            subtitle="vs last period"
          />
        </div>
      )}

      {/* Spend Chart */}
      <SpendChart />

      {/* Platform Performance */}
      <div>
        <h2 className="text-lg font-semibold text-slate-900 mb-4">
          Platform Performance
          {selectedPlatform && (
            <span className="text-sm font-normal text-slate-500 ml-2">
              (Click again to clear filter)
            </span>
          )}
        </h2>
        <PlatformCards
          selectedPlatform={selectedPlatform}
          onPlatformClick={handlePlatformClick}
        />
      </div>

      {/* Bottom Row: Top Campaigns + Recent Activity */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <div className="lg:col-span-2">
          <TopCampaigns platformFilter={selectedPlatform} />
        </div>
        <div>
          <RecentActivity />
        </div>
      </div>
    </div>
  );
}
