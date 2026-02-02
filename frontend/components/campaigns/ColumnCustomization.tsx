"use client";

import { Settings2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import { Checkbox } from "@/components/ui/checkbox";
import { Label } from "@/components/ui/label";
import { type ColumnVisibility } from "./CampaignsTable";

interface ColumnCustomizationProps {
  visibility: ColumnVisibility;
  onVisibilityChange: (visibility: ColumnVisibility) => void;
}

const columnOptions: { key: keyof ColumnVisibility; label: string }[] = [
  { key: "platform", label: "Platform" },
  { key: "name", label: "Campaign Name" },
  { key: "status", label: "Status" },
  { key: "spend", label: "Spend" },
  { key: "impressions", label: "Impressions" },
  { key: "clicks", label: "Clicks" },
  { key: "ctr", label: "CTR" },
  { key: "conversions", label: "Conversions" },
  { key: "roas", label: "ROAS" },
];

const STORAGE_KEY = "campaigns-column-visibility";

export function ColumnCustomization({
  visibility,
  onVisibilityChange,
}: ColumnCustomizationProps) {
  const toggleColumn = (key: keyof ColumnVisibility) => {
    const newVisibility = {
      ...visibility,
      [key]: !visibility[key],
    };
    onVisibilityChange(newVisibility);
    // Save to localStorage
    if (typeof window !== "undefined") {
      localStorage.setItem(STORAGE_KEY, JSON.stringify(newVisibility));
    }
  };

  const resetToDefault = () => {
    const defaultVisibility: ColumnVisibility = {
      platform: true,
      name: true,
      status: true,
      spend: true,
      impressions: true,
      clicks: true,
      ctr: true,
      conversions: true,
      roas: true,
    };
    onVisibilityChange(defaultVisibility);
    if (typeof window !== "undefined") {
      localStorage.removeItem(STORAGE_KEY);
    }
  };

  const visibleCount = Object.values(visibility).filter(Boolean).length;

  return (
    <Popover>
      <PopoverTrigger asChild>
        <Button variant="outline" size="sm" className="gap-2">
          <Settings2 className="h-4 w-4" />
          Columns
          <span className="text-slate-400">({visibleCount})</span>
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-[220px] p-3" align="end">
        <div className="space-y-3">
          <div className="flex items-center justify-between">
            <span className="text-sm font-medium">Show columns</span>
            <Button
              variant="ghost"
              size="sm"
              className="h-auto p-0 text-xs text-blue-600 hover:text-blue-700"
              onClick={resetToDefault}
            >
              Reset
            </Button>
          </div>
          <div className="space-y-2">
            {columnOptions.map((column) => (
              <div key={column.key} className="flex items-center space-x-2">
                <Checkbox
                  id={`column-${column.key}`}
                  checked={visibility[column.key]}
                  onCheckedChange={() => toggleColumn(column.key)}
                />
                <Label
                  htmlFor={`column-${column.key}`}
                  className="flex-1 cursor-pointer text-sm"
                >
                  {column.label}
                </Label>
              </div>
            ))}
          </div>
        </div>
      </PopoverContent>
    </Popover>
  );
}

// Helper function to load visibility from localStorage
export function loadColumnVisibility(): ColumnVisibility {
  if (typeof window === "undefined") {
    return {
      platform: true,
      name: true,
      status: true,
      spend: true,
      impressions: true,
      clicks: true,
      ctr: true,
      conversions: true,
      roas: true,
    };
  }

  try {
    const stored = localStorage.getItem(STORAGE_KEY);
    if (stored) {
      return JSON.parse(stored);
    }
  } catch {
    // Ignore parsing errors
  }

  return {
    platform: true,
    name: true,
    status: true,
    spend: true,
    impressions: true,
    clicks: true,
    ctr: true,
    conversions: true,
    roas: true,
  };
}
