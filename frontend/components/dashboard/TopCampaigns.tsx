"use client";

import { useState } from "react";
import Link from "next/link";
import { ArrowRight, ArrowUpDown } from "lucide-react";
import { useCampaigns } from "@/hooks/use-metrics";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  formatCurrency,
  getPlatformName,
  type Platform,
  type CampaignStatus,
} from "@/lib/mock-data";
import { cn } from "@/lib/utils";

type SortBy = "roas" | "spend";

const platformStyles: Record<Platform, { bg: string; text: string }> = {
  meta: { bg: "bg-blue-100", text: "text-blue-700" },
  tiktok: { bg: "bg-slate-100", text: "text-slate-700" },
  shopee: { bg: "bg-orange-100", text: "text-orange-700" },
};

const statusStyles: Record<CampaignStatus, { bg: string; text: string }> = {
  active: { bg: "bg-emerald-100", text: "text-emerald-700" },
  paused: { bg: "bg-amber-100", text: "text-amber-700" },
  completed: { bg: "bg-slate-100", text: "text-slate-600" },
  draft: { bg: "bg-slate-100", text: "text-slate-500" },
};

interface TopCampaignsProps {
  platformFilter?: Platform | null;
}

export function TopCampaigns({ platformFilter }: TopCampaignsProps) {
  const [sortBy, setSortBy] = useState<SortBy>("roas");
  const { data: allCampaigns, isLoading } = useCampaigns(
    platformFilter ? { platforms: [platformFilter] } : undefined
  );

  if (isLoading) {
    return <TopCampaignsSkeleton />;
  }

  // Sort and take top 5
  const sortedCampaigns = [...(allCampaigns || [])]
    .filter((c) => c.status === "active" || c.status === "completed")
    .sort((a, b) => (sortBy === "roas" ? b.roas - a.roas : b.spend - a.spend))
    .slice(0, 5);

  return (
    <Card className="bg-white border-slate-200">
      <CardHeader className="flex flex-row items-center justify-between pb-2">
        <CardTitle className="text-slate-900">Top Campaigns</CardTitle>
        <div className="flex items-center gap-2">
          <Button
            variant="ghost"
            size="sm"
            onClick={() => setSortBy(sortBy === "roas" ? "spend" : "roas")}
            className="text-slate-600"
          >
            <ArrowUpDown className="h-4 w-4 mr-1" />
            {sortBy === "roas" ? "By ROAS" : "By Spend"}
          </Button>
        </div>
      </CardHeader>
      <CardContent>
        <Table>
          <TableHeader>
            <TableRow className="hover:bg-transparent">
              <TableHead className="text-slate-500">Campaign</TableHead>
              <TableHead className="text-slate-500">Platform</TableHead>
              <TableHead className="text-slate-500 text-right">Spend</TableHead>
              <TableHead className="text-slate-500 text-right">ROAS</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {sortedCampaigns.length === 0 ? (
              <TableRow>
                <TableCell
                  colSpan={4}
                  className="text-center py-8 text-slate-500"
                >
                  No campaigns found
                </TableCell>
              </TableRow>
            ) : (
              sortedCampaigns.map((campaign) => {
                const platformStyle = platformStyles[campaign.platform];
                const statusStyle = statusStyles[campaign.status];

                return (
                  <TableRow key={campaign.id} className="hover:bg-slate-50">
                    <TableCell>
                      <div>
                        <p className="font-medium text-slate-900">
                          {campaign.name}
                        </p>
                        <Badge
                          variant="outline"
                          className={cn(
                            "mt-1 text-xs",
                            statusStyle.bg,
                            statusStyle.text,
                            "border-0"
                          )}
                        >
                          {campaign.status}
                        </Badge>
                      </div>
                    </TableCell>
                    <TableCell>
                      <Badge
                        className={cn(
                          platformStyle.bg,
                          platformStyle.text,
                          "border-0"
                        )}
                      >
                        {getPlatformName(campaign.platform)}
                      </Badge>
                    </TableCell>
                    <TableCell className="text-right font-medium text-slate-900">
                      {formatCurrency(campaign.spend)}
                    </TableCell>
                    <TableCell className="text-right">
                      <span
                        className={cn(
                          "font-semibold",
                          campaign.roas >= 3
                            ? "text-emerald-600"
                            : campaign.roas >= 2
                            ? "text-amber-600"
                            : "text-rose-600"
                        )}
                      >
                        {campaign.roas.toFixed(1)}x
                      </span>
                    </TableCell>
                  </TableRow>
                );
              })
            )}
          </TableBody>
        </Table>

        <div className="mt-4 pt-4 border-t border-slate-100">
          <Link href="/dashboard/campaigns">
            <Button variant="ghost" className="w-full text-blue-600 hover:text-blue-700">
              View all campaigns
              <ArrowRight className="h-4 w-4 ml-2" />
            </Button>
          </Link>
        </div>
      </CardContent>
    </Card>
  );
}

function TopCampaignsSkeleton() {
  return (
    <Card className="bg-white border-slate-200">
      <CardHeader className="pb-2">
        <div className="h-6 w-32 bg-slate-200 rounded animate-pulse" />
      </CardHeader>
      <CardContent>
        <div className="space-y-4">
          {[1, 2, 3, 4, 5].map((i) => (
            <div key={i} className="flex items-center justify-between animate-pulse">
              <div className="flex-1 space-y-2">
                <div className="h-4 w-48 bg-slate-200 rounded" />
                <div className="h-3 w-16 bg-slate-200 rounded" />
              </div>
              <div className="h-4 w-20 bg-slate-200 rounded" />
              <div className="h-4 w-16 bg-slate-200 rounded ml-4" />
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}
