"use client";

import { useState } from "react";
import { FunnelChart, FunnelChartSkeleton } from "@/components/admin";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { useFunnel } from "@/lib/api/hooks/use-admin";

const funnels = [
  { id: "activation", name: "User Activation", description: "From signup to first value" },
  { id: "conversion", name: "Free to Paid", description: "Conversion to paid plans" },
  { id: "engagement", name: "Feature Engagement", description: "Feature adoption funnel" },
];

export default function AdminFunnelsPage() {
  const [selectedFunnel, setSelectedFunnel] = useState("activation");
  const { data: funnel, isLoading } = useFunnel(selectedFunnel);

  // Default funnel steps for each type
  const defaultSteps: Record<string, { name: string; count: number; conversionRate: number }[]> = {
    activation: [
      { name: "Registered", count: 0, conversionRate: 100 },
      { name: "Email Verified", count: 0, conversionRate: 0 },
      { name: "Platform Connected", count: 0, conversionRate: 0 },
      { name: "First Sync Complete", count: 0, conversionRate: 0 },
      { name: "Dashboard Viewed", count: 0, conversionRate: 0 },
    ],
    conversion: [
      { name: "Free Users", count: 0, conversionRate: 100 },
      { name: "Viewed Pricing", count: 0, conversionRate: 0 },
      { name: "Started Checkout", count: 0, conversionRate: 0 },
      { name: "Completed Payment", count: 0, conversionRate: 0 },
    ],
    engagement: [
      { name: "Active Users", count: 0, conversionRate: 100 },
      { name: "Used Dashboard", count: 0, conversionRate: 0 },
      { name: "Generated Report", count: 0, conversionRate: 0 },
      { name: "Exported Data", count: 0, conversionRate: 0 },
    ],
  };

  const currentFunnelInfo = funnels.find((f) => f.id === selectedFunnel);

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-slate-900">Conversion Funnels</h1>
          <p className="mt-1 text-sm text-slate-600">
            Track user journey through key conversion paths
          </p>
        </div>
        <Select value={selectedFunnel} onValueChange={setSelectedFunnel}>
          <SelectTrigger className="w-48">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {funnels.map((f) => (
              <SelectItem key={f.id} value={f.id}>
                {f.name}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      {/* Funnel Info */}
      {currentFunnelInfo && (
        <div className="p-4 bg-indigo-50 rounded-lg border border-indigo-100">
          <h3 className="font-medium text-indigo-900">{currentFunnelInfo.name}</h3>
          <p className="text-sm text-indigo-700">{currentFunnelInfo.description}</p>
        </div>
      )}

      {/* Funnel Chart */}
      {isLoading ? (
        <FunnelChartSkeleton />
      ) : (
        <FunnelChart
          title={currentFunnelInfo?.name || "Funnel"}
          steps={funnel?.steps || defaultSteps[selectedFunnel]}
        />
      )}

      {/* All Funnels Overview */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {funnels.map((f) => (
          <div
            key={f.id}
            className={`p-4 rounded-lg border cursor-pointer transition-all ${
              selectedFunnel === f.id
                ? "border-indigo-500 bg-indigo-50"
                : "border-slate-200 bg-white hover:border-indigo-300"
            }`}
            onClick={() => setSelectedFunnel(f.id)}
          >
            <h4 className="font-medium text-slate-900">{f.name}</h4>
            <p className="text-sm text-slate-500 mt-1">{f.description}</p>
            <div className="mt-3 text-2xl font-bold text-indigo-600">
              {selectedFunnel === f.id && funnel?.overallConversionRate
                ? `${funnel.overallConversionRate.toFixed(1)}%`
                : "—"}
            </div>
            <p className="text-xs text-slate-400">Overall conversion</p>
          </div>
        ))}
      </div>
    </div>
  );
}
