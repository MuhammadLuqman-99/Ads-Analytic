"use client";

import { useState } from "react";
import { Search, X, Calendar, Filter } from "lucide-react";
import { format } from "date-fns";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import { Calendar as CalendarComponent } from "@/components/ui/calendar";
import { Checkbox } from "@/components/ui/checkbox";
import { Label } from "@/components/ui/label";
import { type Platform, type CampaignStatus, getPlatformName } from "@/lib/mock-data";
import { cn } from "@/lib/utils";

export interface CampaignFiltersState {
  search: string;
  platforms: Platform[];
  statuses: CampaignStatus[];
  dateRange: { from: Date; to: Date } | null;
  useGlobalDateRange: boolean;
}

interface CampaignFiltersProps {
  filters: CampaignFiltersState;
  onFiltersChange: (filters: CampaignFiltersState) => void;
  globalDateRange?: { from: Date; to: Date };
}

const platformOptions: { value: Platform; label: string; color: string }[] = [
  { value: "meta", label: "Meta", color: "bg-blue-100 text-blue-700" },
  { value: "tiktok", label: "TikTok", color: "bg-slate-100 text-slate-700" },
  { value: "shopee", label: "Shopee", color: "bg-orange-100 text-orange-700" },
];

const statusOptions: { value: CampaignStatus; label: string; color: string }[] = [
  { value: "active", label: "Active", color: "bg-emerald-100 text-emerald-700" },
  { value: "paused", label: "Paused", color: "bg-amber-100 text-amber-700" },
  { value: "completed", label: "Completed", color: "bg-slate-100 text-slate-600" },
  { value: "draft", label: "Draft", color: "bg-slate-100 text-slate-500" },
];

export function CampaignFilters({
  filters,
  onFiltersChange,
  globalDateRange,
}: CampaignFiltersProps) {
  const [isDateOpen, setIsDateOpen] = useState(false);
  const [isPlatformOpen, setIsPlatformOpen] = useState(false);
  const [isStatusOpen, setIsStatusOpen] = useState(false);

  const updateFilter = <K extends keyof CampaignFiltersState>(
    key: K,
    value: CampaignFiltersState[K]
  ) => {
    onFiltersChange({ ...filters, [key]: value });
  };

  const togglePlatform = (platform: Platform) => {
    const newPlatforms = filters.platforms.includes(platform)
      ? filters.platforms.filter((p) => p !== platform)
      : [...filters.platforms, platform];
    updateFilter("platforms", newPlatforms);
  };

  const toggleStatus = (status: CampaignStatus) => {
    const newStatuses = filters.statuses.includes(status)
      ? filters.statuses.filter((s) => s !== status)
      : [...filters.statuses, status];
    updateFilter("statuses", newStatuses);
  };

  const clearFilters = () => {
    onFiltersChange({
      search: "",
      platforms: [],
      statuses: [],
      dateRange: null,
      useGlobalDateRange: true,
    });
  };

  const hasActiveFilters =
    filters.search ||
    filters.platforms.length > 0 ||
    filters.statuses.length > 0 ||
    filters.dateRange;

  const activeDateRange = filters.useGlobalDateRange
    ? globalDateRange
    : filters.dateRange;

  return (
    <div className="space-y-4">
      {/* Main Filters Row */}
      <div className="flex flex-wrap items-center gap-3">
        {/* Search */}
        <div className="relative flex-1 min-w-[200px] max-w-md">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-slate-400" />
          <Input
            placeholder="Search campaigns..."
            value={filters.search}
            onChange={(e) => updateFilter("search", e.target.value)}
            className="pl-10 bg-white"
          />
          {filters.search && (
            <button
              onClick={() => updateFilter("search", "")}
              className="absolute right-3 top-1/2 -translate-y-1/2 text-slate-400 hover:text-slate-600"
            >
              <X className="h-4 w-4" />
            </button>
          )}
        </div>

        {/* Platform Filter */}
        <Popover open={isPlatformOpen} onOpenChange={setIsPlatformOpen}>
          <PopoverTrigger asChild>
            <Button variant="outline" className="gap-2">
              <Filter className="h-4 w-4" />
              Platform
              {filters.platforms.length > 0 && (
                <Badge variant="secondary" className="ml-1">
                  {filters.platforms.length}
                </Badge>
              )}
            </Button>
          </PopoverTrigger>
          <PopoverContent className="w-[200px] p-3" align="start">
            <div className="space-y-2">
              {platformOptions.map((platform) => (
                <div
                  key={platform.value}
                  className="flex items-center space-x-2"
                >
                  <Checkbox
                    id={`platform-${platform.value}`}
                    checked={filters.platforms.includes(platform.value)}
                    onCheckedChange={() => togglePlatform(platform.value)}
                  />
                  <Label
                    htmlFor={`platform-${platform.value}`}
                    className="flex-1 cursor-pointer"
                  >
                    <Badge className={cn("border-0", platform.color)}>
                      {platform.label}
                    </Badge>
                  </Label>
                </div>
              ))}
            </div>
          </PopoverContent>
        </Popover>

        {/* Status Filter */}
        <Popover open={isStatusOpen} onOpenChange={setIsStatusOpen}>
          <PopoverTrigger asChild>
            <Button variant="outline" className="gap-2">
              Status
              {filters.statuses.length > 0 && (
                <Badge variant="secondary" className="ml-1">
                  {filters.statuses.length}
                </Badge>
              )}
            </Button>
          </PopoverTrigger>
          <PopoverContent className="w-[200px] p-3" align="start">
            <div className="space-y-2">
              {statusOptions.map((status) => (
                <div key={status.value} className="flex items-center space-x-2">
                  <Checkbox
                    id={`status-${status.value}`}
                    checked={filters.statuses.includes(status.value)}
                    onCheckedChange={() => toggleStatus(status.value)}
                  />
                  <Label
                    htmlFor={`status-${status.value}`}
                    className="flex-1 cursor-pointer"
                  >
                    <Badge className={cn("border-0", status.color)}>
                      {status.label}
                    </Badge>
                  </Label>
                </div>
              ))}
            </div>
          </PopoverContent>
        </Popover>

        {/* Date Range */}
        <Popover open={isDateOpen} onOpenChange={setIsDateOpen}>
          <PopoverTrigger asChild>
            <Button variant="outline" className="gap-2">
              <Calendar className="h-4 w-4" />
              {activeDateRange
                ? `${format(activeDateRange.from, "MMM d")} - ${format(
                    activeDateRange.to,
                    "MMM d"
                  )}`
                : "Date Range"}
            </Button>
          </PopoverTrigger>
          <PopoverContent className="w-auto p-3" align="start">
            <div className="space-y-3">
              <div className="flex items-center space-x-2">
                <Checkbox
                  id="use-global-date"
                  checked={filters.useGlobalDateRange}
                  onCheckedChange={(checked) =>
                    updateFilter("useGlobalDateRange", checked === true)
                  }
                />
                <Label htmlFor="use-global-date" className="cursor-pointer">
                  Use global date range
                </Label>
              </div>
              {!filters.useGlobalDateRange && (
                <CalendarComponent
                  mode="range"
                  selected={
                    filters.dateRange
                      ? { from: filters.dateRange.from, to: filters.dateRange.to }
                      : undefined
                  }
                  onSelect={(range) => {
                    if (range?.from && range?.to) {
                      updateFilter("dateRange", { from: range.from, to: range.to });
                    }
                  }}
                  numberOfMonths={2}
                />
              )}
            </div>
          </PopoverContent>
        </Popover>

        {/* Clear Filters */}
        {hasActiveFilters && (
          <Button variant="ghost" size="sm" onClick={clearFilters}>
            <X className="h-4 w-4 mr-1" />
            Clear
          </Button>
        )}
      </div>

      {/* Active Filters Display */}
      {hasActiveFilters && (
        <div className="flex flex-wrap items-center gap-2">
          <span className="text-sm text-slate-500">Active filters:</span>
          {filters.platforms.map((platform) => {
            const option = platformOptions.find((p) => p.value === platform);
            return (
              <Badge
                key={platform}
                variant="outline"
                className="gap-1 cursor-pointer hover:bg-slate-100"
                onClick={() => togglePlatform(platform)}
              >
                {option?.label}
                <X className="h-3 w-3" />
              </Badge>
            );
          })}
          {filters.statuses.map((status) => {
            const option = statusOptions.find((s) => s.value === status);
            return (
              <Badge
                key={status}
                variant="outline"
                className="gap-1 cursor-pointer hover:bg-slate-100"
                onClick={() => toggleStatus(status)}
              >
                {option?.label}
                <X className="h-3 w-3" />
              </Badge>
            );
          })}
        </div>
      )}
    </div>
  );
}
