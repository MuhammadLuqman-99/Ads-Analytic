"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { cn } from "@/lib/utils";

interface PlatformBreakdownProps {
  platforms: Record<string, number>;
  className?: string;
}

const platformColors: Record<string, { bg: string; bar: string }> = {
  meta: { bg: "bg-blue-100", bar: "bg-blue-500" },
  facebook: { bg: "bg-blue-100", bar: "bg-blue-500" },
  tiktok: { bg: "bg-slate-100", bar: "bg-slate-800" },
  shopee: { bg: "bg-orange-100", bar: "bg-orange-500" },
  google: { bg: "bg-green-100", bar: "bg-green-500" },
  lazada: { bg: "bg-purple-100", bar: "bg-purple-500" },
};

const platformLabels: Record<string, string> = {
  meta: "Meta Ads",
  facebook: "Meta Ads",
  tiktok: "TikTok Ads",
  shopee: "Shopee Ads",
  google: "Google Ads",
  lazada: "Lazada Ads",
};

export function PlatformBreakdown({ platforms, className }: PlatformBreakdownProps) {
  const total = Object.values(platforms).reduce((sum, count) => sum + count, 0);
  const sortedPlatforms = Object.entries(platforms).sort(([, a], [, b]) => b - a);

  return (
    <Card className={cn("bg-white", className)}>
      <CardHeader>
        <CardTitle className="text-lg font-semibold text-slate-900">Platform Breakdown</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="space-y-4">
          {sortedPlatforms.map(([platform, count]) => {
            const percent = total > 0 ? (count / total) * 100 : 0;
            const colors = platformColors[platform.toLowerCase()] || {
              bg: "bg-slate-100",
              bar: "bg-slate-500",
            };
            const label = platformLabels[platform.toLowerCase()] || platform;

            return (
              <div key={platform} className="space-y-2">
                <div className="flex items-center justify-between text-sm">
                  <div className="flex items-center gap-2">
                    <div className={cn("h-3 w-3 rounded-full", colors.bar)} />
                    <span className="font-medium text-slate-700">{label}</span>
                  </div>
                  <span className="text-slate-500">
                    {count.toLocaleString()} ({percent.toFixed(1)}%)
                  </span>
                </div>
                <div className="h-2 bg-slate-100 rounded-full overflow-hidden">
                  <div
                    className={cn("h-full rounded-full transition-all duration-500", colors.bar)}
                    style={{ width: `${percent}%` }}
                  />
                </div>
              </div>
            );
          })}
        </div>

        {/* Total */}
        <div className="mt-6 pt-4 border-t border-slate-200">
          <div className="flex items-center justify-between">
            <span className="text-sm text-slate-600">Total Connections</span>
            <span className="text-lg font-bold text-slate-900">{total.toLocaleString()}</span>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}

export function PlatformBreakdownSkeleton() {
  return (
    <Card className="bg-white">
      <CardHeader>
        <div className="h-6 w-40 bg-slate-200 rounded animate-pulse" />
      </CardHeader>
      <CardContent className="space-y-4">
        {[1, 2, 3].map((i) => (
          <div key={i} className="space-y-2 animate-pulse">
            <div className="flex justify-between">
              <div className="h-4 w-24 bg-slate-200 rounded" />
              <div className="h-4 w-16 bg-slate-200 rounded" />
            </div>
            <div className="h-2 bg-slate-200 rounded-full" />
          </div>
        ))}
      </CardContent>
    </Card>
  );
}
