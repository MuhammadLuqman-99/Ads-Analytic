"use client";

import { TrendingUp, TrendingDown, Minus } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { cn } from "@/lib/utils";
import {
  formatCurrency,
  formatNumber,
  getPlatformName,
  type Platform,
} from "@/lib/mock-data";
import {
  BarChart as LibBarChart,
  LineChart as LibLineChart,
  DonutChart,
} from "@/components/lib";

const PLATFORM_COLORS: Record<Platform, string> = {
  meta: "#3B82F6",
  tiktok: "#000000",
  shopee: "#EE4D2D",
};

interface PlatformData {
  platform: Platform;
  spend: number;
  impressions: number;
  clicks: number;
  conversions: number;
  roas: number;
  ctr: number;
}

interface DailyData {
  date: string;
  spend: number;
  impressions: number;
  clicks: number;
  conversions: number;
}

interface CampaignPerformance {
  id: string;
  name: string;
  platform: Platform;
  spend: number;
  roas: number;
  conversions: number;
  trend: "up" | "down" | "stable";
}

export function PlatformComparisonReport({ data }: { data: PlatformData[] }) {
  const chartData = data.map((d) => ({
    name: getPlatformName(d.platform),
    Spend: d.spend,
    Conversions: d.conversions,
    ROAS: d.roas,
    platform: d.platform,
    color: PLATFORM_COLORS[d.platform],
  }));

  return (
    <Card className="bg-white border-slate-200">
      <CardHeader>
        <CardTitle className="text-slate-900">Platform Comparison</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <div>
            <h4 className="text-sm font-medium text-slate-700 mb-4">Spend by Platform</h4>
            <LibBarChart
              data={chartData as unknown as Record<string, unknown>[]}
              xAxisKey="name"
              series={[{ key: "Spend", name: "Spend", color: "#3B82F6" }]}
              height={250}
              showLegend={false}
              colorByData
              colors={data.map((d) => PLATFORM_COLORS[d.platform])}
              tooltipFormatter={(value) => formatCurrency(value)}
            />
          </div>
          <div>
            <h4 className="text-sm font-medium text-slate-700 mb-4">Key Metrics</h4>
            <Table>
              <TableHeader>
                <TableRow className="bg-slate-50">
                  <TableHead>Platform</TableHead>
                  <TableHead className="text-right">Spend</TableHead>
                  <TableHead className="text-right">Conv.</TableHead>
                  <TableHead className="text-right">ROAS</TableHead>
                  <TableHead className="text-right">CTR</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {data.map((row) => (
                  <TableRow key={row.platform}>
                    <TableCell>
                      <Badge variant={row.platform}>{getPlatformName(row.platform)}</Badge>
                    </TableCell>
                    <TableCell className="text-right">{formatCurrency(row.spend)}</TableCell>
                    <TableCell className="text-right">{formatNumber(row.conversions)}</TableCell>
                    <TableCell className="text-right">{row.roas.toFixed(2)}x</TableCell>
                    <TableCell className="text-right">{row.ctr.toFixed(2)}%</TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}

export function DailyPerformanceReport({ data }: { data: DailyData[] }) {
  return (
    <Card className="bg-white border-slate-200">
      <CardHeader>
        <CardTitle className="text-slate-900">Daily Performance Trend</CardTitle>
      </CardHeader>
      <CardContent>
        <LibLineChart
          data={data as unknown as Record<string, unknown>[]}
          xAxisKey="date"
          series={[
            { key: "spend", name: "Spend", color: "#3B82F6" },
            { key: "conversions", name: "Conversions", color: "#10B981" },
          ]}
          height={300}
          showDots={false}
        />
      </CardContent>
    </Card>
  );
}

export function TopBottomPerformersReport({
  topCampaigns,
  bottomCampaigns,
}: {
  topCampaigns: CampaignPerformance[];
  bottomCampaigns: CampaignPerformance[];
}) {
  const TrendIcon = ({ trend }: { trend: "up" | "down" | "stable" }) => {
    if (trend === "up") return <TrendingUp className="h-4 w-4 text-emerald-500" />;
    if (trend === "down") return <TrendingDown className="h-4 w-4 text-red-500" />;
    return <Minus className="h-4 w-4 text-slate-400" />;
  };

  const CampaignList = ({ campaigns, type }: { campaigns: CampaignPerformance[]; type: "top" | "bottom" }) => (
    <div>
      <h4 className={cn("text-sm font-medium mb-3", type === "top" ? "text-emerald-700" : "text-red-700")}>
        {type === "top" ? "Top Performers" : "Needs Attention"}
      </h4>
      <div className="space-y-2">
        {campaigns.map((campaign, index) => (
          <div
            key={campaign.id}
            className={cn(
              "flex items-center justify-between p-3 rounded-lg border",
              type === "top" ? "bg-emerald-50 border-emerald-100" : "bg-red-50 border-red-100"
            )}
          >
            <div className="flex items-center gap-3">
              <span className={cn("text-lg font-bold", type === "top" ? "text-emerald-600" : "text-red-600")}>
                #{index + 1}
              </span>
              <div>
                <p className="text-sm font-medium text-slate-900">{campaign.name}</p>
                <Badge variant={campaign.platform} className="mt-1">{getPlatformName(campaign.platform)}</Badge>
              </div>
            </div>
            <div className="text-right">
              <div className="flex items-center gap-1 justify-end">
                <span className="text-sm font-medium text-slate-900">{campaign.roas.toFixed(2)}x ROAS</span>
                <TrendIcon trend={campaign.trend} />
              </div>
              <p className="text-xs text-slate-500 mt-1">{formatCurrency(campaign.spend)} spent</p>
            </div>
          </div>
        ))}
      </div>
    </div>
  );

  return (
    <Card className="bg-white border-slate-200">
      <CardHeader>
        <CardTitle className="text-slate-900">Top & Bottom Performers</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <CampaignList campaigns={topCampaigns} type="top" />
          <CampaignList campaigns={bottomCampaigns} type="bottom" />
        </div>
      </CardContent>
    </Card>
  );
}

export function SpendDistributionReport({ data }: { data: PlatformData[] }) {
  const pieData = data.map((d) => ({
    name: getPlatformName(d.platform),
    value: d.spend,
    color: PLATFORM_COLORS[d.platform],
  }));
  const total = pieData.reduce((sum, d) => sum + d.value, 0);

  return (
    <Card className="bg-white border-slate-200">
      <CardHeader>
        <CardTitle className="text-slate-900">Spend Distribution</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 items-center">
          <DonutChart
            data={pieData}
            height={250}
            showLegend={false}
            centerLabel={{ value: formatCurrency(total), label: "Total Spend" }}
            valueFormatter={(v) => formatCurrency(v)}
          />
          <div className="space-y-4">
            {pieData.map((item) => {
              const percentage = ((item.value / total) * 100).toFixed(1);
              return (
                <div key={item.name} className="flex items-center gap-4">
                  <div className="w-4 h-4 rounded" style={{ backgroundColor: item.color }} />
                  <div className="flex-1">
                    <div className="flex items-center justify-between">
                      <span className="text-sm font-medium text-slate-900">{item.name}</span>
                      <span className="text-sm text-slate-500">{percentage}%</span>
                    </div>
                    <div className="mt-1 h-2 bg-slate-100 rounded-full overflow-hidden">
                      <div className="h-full rounded-full" style={{ width: `${percentage}%`, backgroundColor: item.color }} />
                    </div>
                    <p className="text-xs text-slate-500 mt-1">{formatCurrency(item.value)}</p>
                  </div>
                </div>
              );
            })}
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
