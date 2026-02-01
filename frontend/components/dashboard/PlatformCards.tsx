"use client";

import { TrendingUp, TrendingDown } from "lucide-react";
import { usePlatformMetrics } from "@/hooks/use-metrics";
import { Card, CardContent } from "@/components/ui/card";
import { formatCurrency, formatNumber, type Platform } from "@/lib/mock-data";
import { cn } from "@/lib/utils";

interface PlatformConfig {
  name: string;
  icon: string;
  bgColor: string;
  textColor: string;
}

const platformConfigs: Record<Platform, PlatformConfig> = {
  meta: {
    name: "Meta Ads",
    icon: "M",
    bgColor: "bg-blue-600",
    textColor: "text-blue-600",
  },
  tiktok: {
    name: "TikTok Ads",
    icon: "T",
    bgColor: "bg-black",
    textColor: "text-black",
  },
  shopee: {
    name: "Shopee Ads",
    icon: "S",
    bgColor: "bg-orange-500",
    textColor: "text-orange-500",
  },
};

interface PlatformCardsProps {
  onPlatformClick?: (platform: Platform) => void;
  selectedPlatform?: Platform | null;
}

export function PlatformCards({ onPlatformClick, selectedPlatform }: PlatformCardsProps) {
  const { data: platforms, isLoading } = usePlatformMetrics();

  if (isLoading) {
    return (
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        {[1, 2, 3].map((i) => (
          <PlatformCardSkeleton key={i} />
        ))}
      </div>
    );
  }

  return (
    <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
      {platforms?.map((platform) => {
        const config = platformConfigs[platform.platform];
        const isSelected = selectedPlatform === platform.platform;

        return (
          <Card
            key={platform.platform}
            className={cn(
              "bg-white border-slate-200 cursor-pointer transition-all hover:shadow-md",
              isSelected && "ring-2 ring-offset-2",
              isSelected && config.textColor.replace("text-", "ring-")
            )}
            onClick={() => onPlatformClick?.(platform.platform)}
          >
            <CardContent className="p-5">
              {/* Header */}
              <div className="flex items-center gap-3 mb-4">
                <div
                  className={cn(
                    "h-10 w-10 rounded-lg flex items-center justify-center text-white font-bold text-lg shadow-lg",
                    config.bgColor
                  )}
                >
                  {config.icon}
                </div>
                <div className="flex-1">
                  <p className="font-semibold text-slate-900">{config.name}</p>
                  <p className="text-xs text-slate-500">
                    {formatNumber(platform.impressions)} impressions
                  </p>
                </div>
              </div>

              {/* Metrics */}
              <div className="grid grid-cols-3 gap-3">
                <div>
                  <p className="text-xs text-slate-500 mb-1">Spend</p>
                  <p className="text-sm font-semibold text-slate-900">
                    {formatCurrency(platform.spend)}
                  </p>
                  <TrendIndicator value={platform.spendChange} />
                </div>
                <div>
                  <p className="text-xs text-slate-500 mb-1">ROAS</p>
                  <p className="text-sm font-semibold text-slate-900">
                    {platform.roas.toFixed(1)}x
                  </p>
                  <TrendIndicator value={platform.roasChange} />
                </div>
                <div>
                  <p className="text-xs text-slate-500 mb-1">Conv.</p>
                  <p className="text-sm font-semibold text-slate-900">
                    {formatNumber(platform.conversions)}
                  </p>
                </div>
              </div>
            </CardContent>
          </Card>
        );
      })}
    </div>
  );
}

function TrendIndicator({ value }: { value: number }) {
  const isPositive = value >= 0;

  return (
    <div
      className={cn(
        "flex items-center gap-0.5 text-xs font-medium mt-1",
        isPositive ? "text-emerald-600" : "text-rose-600"
      )}
    >
      {isPositive ? (
        <TrendingUp className="h-3 w-3" />
      ) : (
        <TrendingDown className="h-3 w-3" />
      )}
      {Math.abs(value).toFixed(1)}%
    </div>
  );
}

function PlatformCardSkeleton() {
  return (
    <Card className="bg-white border-slate-200">
      <CardContent className="p-5 animate-pulse">
        <div className="flex items-center gap-3 mb-4">
          <div className="h-10 w-10 bg-slate-200 rounded-lg" />
          <div className="flex-1 space-y-2">
            <div className="h-4 w-24 bg-slate-200 rounded" />
            <div className="h-3 w-32 bg-slate-200 rounded" />
          </div>
        </div>
        <div className="grid grid-cols-3 gap-3">
          {[1, 2, 3].map((i) => (
            <div key={i} className="space-y-2">
              <div className="h-3 w-10 bg-slate-200 rounded" />
              <div className="h-4 w-16 bg-slate-200 rounded" />
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}
