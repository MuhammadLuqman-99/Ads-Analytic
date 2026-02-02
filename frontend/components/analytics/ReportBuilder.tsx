"use client";

import { useState } from "react";
import { format } from "date-fns";
import { Calendar, Play, RotateCcw } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { Label } from "@/components/ui/label";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import { Calendar as CalendarComponent } from "@/components/ui/calendar";
import { cn } from "@/lib/utils";

export interface ReportConfig {
  metrics: string[];
  dimensions: string[];
  dateRange: { from: Date; to: Date } | null;
  comparisonDateRange: { from: Date; to: Date } | null;
  compareEnabled: boolean;
}

interface ReportBuilderProps {
  config: ReportConfig;
  onConfigChange: (config: ReportConfig) => void;
  onGenerateReport: () => void;
  isGenerating?: boolean;
}

const metricOptions = [
  { id: "spend", label: "Spend", description: "Total ad spend" },
  { id: "impressions", label: "Impressions", description: "Times ads were shown" },
  { id: "clicks", label: "Clicks", description: "Number of clicks" },
  { id: "ctr", label: "CTR", description: "Click-through rate" },
  { id: "conversions", label: "Conversions", description: "Completed goals" },
  { id: "cpc", label: "CPC", description: "Cost per click" },
  { id: "cpm", label: "CPM", description: "Cost per 1000 impressions" },
  { id: "roas", label: "ROAS", description: "Return on ad spend" },
  { id: "revenue", label: "Revenue", description: "Total revenue" },
];

const dimensionOptions = [
  { id: "platform", label: "Platform", description: "Meta, TikTok, Shopee" },
  { id: "campaign", label: "Campaign", description: "By campaign" },
  { id: "date", label: "Date", description: "Daily breakdown" },
  { id: "status", label: "Status", description: "Active, Paused, etc" },
];

export function ReportBuilder({
  config,
  onConfigChange,
  onGenerateReport,
  isGenerating,
}: ReportBuilderProps) {
  const [isDateOpen, setIsDateOpen] = useState(false);
  const [isComparisonDateOpen, setIsComparisonDateOpen] = useState(false);

  const toggleMetric = (metricId: string) => {
    const newMetrics = config.metrics.includes(metricId)
      ? config.metrics.filter((m) => m !== metricId)
      : [...config.metrics, metricId];
    onConfigChange({ ...config, metrics: newMetrics });
  };

  const toggleDimension = (dimensionId: string) => {
    const newDimensions = config.dimensions.includes(dimensionId)
      ? config.dimensions.filter((d) => d !== dimensionId)
      : [...config.dimensions, dimensionId];
    onConfigChange({ ...config, dimensions: newDimensions });
  };

  const resetConfig = () => {
    onConfigChange({
      metrics: ["spend", "impressions", "clicks", "conversions", "roas"],
      dimensions: ["platform"],
      dateRange: null,
      comparisonDateRange: null,
      compareEnabled: false,
    });
  };

  const canGenerate = config.metrics.length > 0 && config.dimensions.length > 0;

  return (
    <Card className="bg-white border-slate-200">
      <CardHeader className="pb-4">
        <div className="flex items-center justify-between">
          <CardTitle className="text-slate-900">Report Builder</CardTitle>
          <Button variant="ghost" size="sm" onClick={resetConfig}>
            <RotateCcw className="h-4 w-4 mr-2" />
            Reset
          </Button>
        </div>
      </CardHeader>
      <CardContent className="space-y-6">
        {/* Metrics Selection */}
        <div>
          <h3 className="text-sm font-medium text-slate-900 mb-3">
            Select Metrics
          </h3>
          <div className="grid grid-cols-2 md:grid-cols-3 gap-3">
            {metricOptions.map((metric) => (
              <div
                key={metric.id}
                className={cn(
                  "flex items-start space-x-3 p-3 rounded-lg border transition-colors cursor-pointer",
                  config.metrics.includes(metric.id)
                    ? "border-blue-500 bg-blue-50"
                    : "border-slate-200 hover:border-slate-300"
                )}
                onClick={() => toggleMetric(metric.id)}
              >
                <Checkbox
                  id={`metric-${metric.id}`}
                  checked={config.metrics.includes(metric.id)}
                  onCheckedChange={() => toggleMetric(metric.id)}
                />
                <div className="flex-1">
                  <Label
                    htmlFor={`metric-${metric.id}`}
                    className="text-sm font-medium cursor-pointer"
                  >
                    {metric.label}
                  </Label>
                  <p className="text-xs text-slate-500 mt-0.5">
                    {metric.description}
                  </p>
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Dimensions Selection */}
        <div>
          <h3 className="text-sm font-medium text-slate-900 mb-3">
            Group By (Dimensions)
          </h3>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
            {dimensionOptions.map((dimension) => (
              <div
                key={dimension.id}
                className={cn(
                  "flex items-start space-x-3 p-3 rounded-lg border transition-colors cursor-pointer",
                  config.dimensions.includes(dimension.id)
                    ? "border-blue-500 bg-blue-50"
                    : "border-slate-200 hover:border-slate-300"
                )}
                onClick={() => toggleDimension(dimension.id)}
              >
                <Checkbox
                  id={`dimension-${dimension.id}`}
                  checked={config.dimensions.includes(dimension.id)}
                  onCheckedChange={() => toggleDimension(dimension.id)}
                />
                <div className="flex-1">
                  <Label
                    htmlFor={`dimension-${dimension.id}`}
                    className="text-sm font-medium cursor-pointer"
                  >
                    {dimension.label}
                  </Label>
                  <p className="text-xs text-slate-500 mt-0.5">
                    {dimension.description}
                  </p>
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Date Range */}
        <div className="flex flex-wrap items-end gap-4">
          <div>
            <h3 className="text-sm font-medium text-slate-900 mb-2">
              Date Range
            </h3>
            <Popover open={isDateOpen} onOpenChange={setIsDateOpen}>
              <PopoverTrigger asChild>
                <Button variant="outline" className="w-[280px] justify-start">
                  <Calendar className="h-4 w-4 mr-2" />
                  {config.dateRange
                    ? `${format(config.dateRange.from, "MMM d, yyyy")} - ${format(
                        config.dateRange.to,
                        "MMM d, yyyy"
                      )}`
                    : "Select date range"}
                </Button>
              </PopoverTrigger>
              <PopoverContent className="w-auto p-0" align="start">
                <CalendarComponent
                  mode="range"
                  selected={config.dateRange || undefined}
                  onSelect={(range) => {
                    if (range?.from && range?.to) {
                      onConfigChange({
                        ...config,
                        dateRange: { from: range.from, to: range.to },
                      });
                      setIsDateOpen(false);
                    }
                  }}
                  numberOfMonths={2}
                />
              </PopoverContent>
            </Popover>
          </div>

          {/* Comparison Toggle */}
          <div className="flex items-center space-x-2">
            <Checkbox
              id="compare-enabled"
              checked={config.compareEnabled}
              onCheckedChange={(checked) =>
                onConfigChange({ ...config, compareEnabled: checked === true })
              }
            />
            <Label htmlFor="compare-enabled" className="cursor-pointer">
              Compare with another period
            </Label>
          </div>

          {/* Comparison Date Range */}
          {config.compareEnabled && (
            <Popover
              open={isComparisonDateOpen}
              onOpenChange={setIsComparisonDateOpen}
            >
              <PopoverTrigger asChild>
                <Button variant="outline" className="w-[280px] justify-start">
                  <Calendar className="h-4 w-4 mr-2" />
                  {config.comparisonDateRange
                    ? `${format(
                        config.comparisonDateRange.from,
                        "MMM d, yyyy"
                      )} - ${format(config.comparisonDateRange.to, "MMM d, yyyy")}`
                    : "Select comparison period"}
                </Button>
              </PopoverTrigger>
              <PopoverContent className="w-auto p-0" align="start">
                <CalendarComponent
                  mode="range"
                  selected={config.comparisonDateRange || undefined}
                  onSelect={(range) => {
                    if (range?.from && range?.to) {
                      onConfigChange({
                        ...config,
                        comparisonDateRange: { from: range.from, to: range.to },
                      });
                      setIsComparisonDateOpen(false);
                    }
                  }}
                  numberOfMonths={2}
                />
              </PopoverContent>
            </Popover>
          )}
        </div>

        {/* Generate Button */}
        <div className="flex items-center justify-between pt-4 border-t border-slate-100">
          <p className="text-sm text-slate-500">
            {config.metrics.length} metrics, {config.dimensions.length} dimensions
            selected
          </p>
          <Button
            onClick={onGenerateReport}
            disabled={!canGenerate || isGenerating}
          >
            <Play className="h-4 w-4 mr-2" />
            {isGenerating ? "Generating..." : "Generate Report"}
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}
