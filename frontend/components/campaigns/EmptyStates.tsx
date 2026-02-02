"use client";

import Link from "next/link";
import { Search, Link2, FolderOpen, Plus } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";

interface EmptyStateProps {
  type: "no-campaigns" | "no-results" | "no-accounts";
  onClearFilters?: () => void;
}

export function EmptyState({ type, onClearFilters }: EmptyStateProps) {
  if (type === "no-accounts") {
    return (
      <Card className="border-2 border-dashed border-slate-200">
        <CardContent className="flex flex-col items-center justify-center py-16">
          <div className="h-16 w-16 rounded-full bg-slate-100 flex items-center justify-center mb-4">
            <Link2 className="h-8 w-8 text-slate-400" />
          </div>
          <h3 className="text-lg font-semibold text-slate-900 mb-2">
            No Ad Accounts Connected
          </h3>
          <p className="text-slate-500 text-center max-w-md mb-6">
            Connect your ad accounts from Meta, TikTok, or Shopee to start
            tracking your campaigns and performance metrics.
          </p>
          <Link href="/dashboard/connections">
            <Button>
              <Plus className="h-4 w-4 mr-2" />
              Connect Ad Account
            </Button>
          </Link>
        </CardContent>
      </Card>
    );
  }

  if (type === "no-results") {
    return (
      <Card className="border-2 border-dashed border-slate-200">
        <CardContent className="flex flex-col items-center justify-center py-16">
          <div className="h-16 w-16 rounded-full bg-slate-100 flex items-center justify-center mb-4">
            <Search className="h-8 w-8 text-slate-400" />
          </div>
          <h3 className="text-lg font-semibold text-slate-900 mb-2">
            No Campaigns Found
          </h3>
          <p className="text-slate-500 text-center max-w-md mb-6">
            No campaigns match your current filters. Try adjusting your search
            criteria or clearing the filters.
          </p>
          {onClearFilters && (
            <Button variant="outline" onClick={onClearFilters}>
              Clear All Filters
            </Button>
          )}
        </CardContent>
      </Card>
    );
  }

  // no-campaigns
  return (
    <Card className="border-2 border-dashed border-slate-200">
      <CardContent className="flex flex-col items-center justify-center py-16">
        <div className="h-16 w-16 rounded-full bg-slate-100 flex items-center justify-center mb-4">
          <FolderOpen className="h-8 w-8 text-slate-400" />
        </div>
        <h3 className="text-lg font-semibold text-slate-900 mb-2">
          No Campaigns Yet
        </h3>
        <p className="text-slate-500 text-center max-w-md mb-6">
          Your connected ad accounts don&apos;t have any campaigns yet. Create
          campaigns in your ad platforms and they will appear here automatically.
        </p>
        <div className="flex gap-3">
          <Button variant="outline" asChild>
            <Link href="/dashboard/connections">Check Connections</Link>
          </Button>
          <Button onClick={() => window.location.reload()}>
            Refresh Data
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}
