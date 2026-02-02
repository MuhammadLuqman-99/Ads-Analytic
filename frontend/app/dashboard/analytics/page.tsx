"use client";

import { useState, useMemo } from "react";
import { format, subDays } from "date-fns";
import { FileText, BarChart3, Layout } from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  ReportBuilder,
  ChartVisualization,
  PlatformComparisonReport,
  DailyPerformanceReport,
  TopBottomPerformersReport,
  SpendDistributionReport,
  ExportOptions,
  ComparisonDashboard,
  exportToCSV,
  exportToPDF,
  type ReportConfig,
} from "@/components/analytics";
import { useDateRangeStore } from "@/stores/app-store";
import { type Platform } from "@/lib/mock-data";

// Mock data for analytics
const generateMockPlatformData = () => [
  { platform: "meta" as Platform, spend: 15420, impressions: 892000, clicks: 28540, conversions: 1245, roas: 4.2, ctr: 3.2 },
  { platform: "tiktok" as Platform, spend: 8750, impressions: 1250000, clicks: 45600, conversions: 890, roas: 3.8, ctr: 3.65 },
  { platform: "shopee" as Platform, spend: 5230, impressions: 425000, clicks: 12400, conversions: 567, roas: 5.1, ctr: 2.92 },
];

const generateMockDailyData = () => {
  const data = [];
  for (let i = 29; i >= 0; i--) {
    const date = subDays(new Date(), i);
    data.push({
      date: format(date, "MMM d"),
      spend: Math.floor(Math.random() * 1500) + 500,
      impressions: Math.floor(Math.random() * 50000) + 20000,
      clicks: Math.floor(Math.random() * 2000) + 500,
      conversions: Math.floor(Math.random() * 100) + 20,
    });
  }
  return data;
};

const generateMockCampaigns = () => ({
  top: [
    { id: "1", name: "Summer Sale 2024", platform: "meta" as Platform, spend: 5420, roas: 6.2, conversions: 456, trend: "up" as const },
    { id: "2", name: "Brand Awareness TikTok", platform: "tiktok" as Platform, spend: 3200, roas: 5.1, conversions: 312, trend: "up" as const },
    { id: "3", name: "Retargeting Meta", platform: "meta" as Platform, spend: 2100, roas: 4.8, conversions: 189, trend: "stable" as const },
  ],
  bottom: [
    { id: "4", name: "New Product Launch", platform: "shopee" as Platform, spend: 1800, roas: 1.2, conversions: 34, trend: "down" as const },
    { id: "5", name: "Holiday Promo", platform: "tiktok" as Platform, spend: 2400, roas: 1.5, conversions: 56, trend: "down" as const },
    { id: "6", name: "Flash Sale Weekend", platform: "meta" as Platform, spend: 1200, roas: 1.8, conversions: 42, trend: "stable" as const },
  ],
});

const defaultReportConfig: ReportConfig = {
  metrics: ["spend", "impressions", "clicks", "conversions", "roas"],
  dimensions: ["platform"],
  dateRange: null,
  comparisonDateRange: null,
  compareEnabled: false,
};

export default function AnalyticsPage() {
  const { dateRange: globalDateRange } = useDateRangeStore();
  const [activeTab, setActiveTab] = useState("prebuilt");
  const [reportConfig, setReportConfig] = useState<ReportConfig>(defaultReportConfig);
  const [isGenerating, setIsGenerating] = useState(false);
  const [generatedReportData, setGeneratedReportData] = useState<Record<string, unknown>[] | null>(null);

  // Mock data
  const platformData = useMemo(() => generateMockPlatformData(), []);
  const dailyData = useMemo(() => generateMockDailyData(), []);
  const campaignData = useMemo(() => generateMockCampaigns(), []);

  // Comparison data
  const comparisonMetrics = useMemo(() => [
    { label: "Total Spend", currentValue: 29400, previousValue: 25800, format: "currency" as const },
    { label: "Conversions", currentValue: 2702, previousValue: 2150, format: "number" as const },
    { label: "Avg. ROAS", currentValue: 4.37, previousValue: 3.85, format: "multiplier" as const },
    { label: "Avg. CTR", currentValue: 3.26, previousValue: 2.98, format: "percentage" as const },
  ], []);

  const handleGenerateReport = () => {
    setIsGenerating(true);
    // Simulate report generation
    setTimeout(() => {
      const mockReportData = platformData.map((p) => ({
        Platform: p.platform,
        Spend: p.spend,
        Impressions: p.impressions,
        Clicks: p.clicks,
        Conversions: p.conversions,
        ROAS: p.roas,
        CTR: p.ctr,
      }));
      setGeneratedReportData(mockReportData);
      setIsGenerating(false);
    }, 1000);
  };

  const handleExportCSV = () => {
    const dataToExport = generatedReportData || platformData.map((p) => ({
      Platform: p.platform,
      Spend: p.spend,
      Impressions: p.impressions,
      Clicks: p.clicks,
      Conversions: p.conversions,
      ROAS: p.roas,
      CTR: p.ctr,
    }));
    exportToCSV(dataToExport, "analytics-report");
  };

  const handleExportPDF = () => {
    exportToPDF("analytics-content", "analytics-report");
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
        <div>
          <h1 className="text-3xl font-bold text-slate-900">Analytics</h1>
          <p className="mt-1 text-slate-500">
            Deep dive into your advertising performance and trends
          </p>
        </div>
        <ExportOptions
          onExportCSV={handleExportCSV}
          onExportPDF={handleExportPDF}
        />
      </div>

      {/* Main Content */}
      <div id="analytics-content">
        <Tabs value={activeTab} onValueChange={setActiveTab}>
          <TabsList className="mb-6">
            <TabsTrigger value="prebuilt" className="gap-2">
              <Layout className="h-4 w-4" />
              Pre-built Reports
            </TabsTrigger>
            <TabsTrigger value="custom" className="gap-2">
              <FileText className="h-4 w-4" />
              Report Builder
            </TabsTrigger>
            <TabsTrigger value="comparison" className="gap-2">
              <BarChart3 className="h-4 w-4" />
              Comparison Mode
            </TabsTrigger>
          </TabsList>

          {/* Pre-built Reports Tab */}
          <TabsContent value="prebuilt" className="space-y-6">
            <div className="grid grid-cols-1 xl:grid-cols-2 gap-6">
              <PlatformComparisonReport data={platformData} />
              <SpendDistributionReport data={platformData} />
            </div>
            <DailyPerformanceReport data={dailyData} />
            <TopBottomPerformersReport
              topCampaigns={campaignData.top}
              bottomCampaigns={campaignData.bottom}
            />
          </TabsContent>

          {/* Report Builder Tab */}
          <TabsContent value="custom" className="space-y-6">
            <ReportBuilder
              config={reportConfig}
              onConfigChange={setReportConfig}
              onGenerateReport={handleGenerateReport}
              isGenerating={isGenerating}
            />

            {generatedReportData && (
              <ChartVisualization
                title="Generated Report"
                data={generatedReportData.map((row) => ({
                  name: String(row.Platform || row.name || ""),
                  ...Object.fromEntries(
                    Object.entries(row)
                      .filter(([key]) => key !== "Platform" && key !== "name")
                      .map(([key, value]) => [key.toLowerCase(), value])
                  ),
                }))}
                metrics={reportConfig.metrics}
                allowedChartTypes={["line", "bar", "pie", "table"]}
              />
            )}
          </TabsContent>

          {/* Comparison Mode Tab */}
          <TabsContent value="comparison" className="space-y-6">
            <ComparisonDashboard
              currentPeriod={
                globalDateRange
                  ? `${format(globalDateRange.from, "MMM d")} - ${format(globalDateRange.to, "MMM d, yyyy")}`
                  : "Last 30 days"
              }
              previousPeriod="Previous 30 days"
              summaryMetrics={comparisonMetrics}
              detailedMetrics={[
                { label: "Meta Spend", currentValue: 15420, previousValue: 13200, format: "currency" },
                { label: "TikTok Spend", currentValue: 8750, previousValue: 7800, format: "currency" },
                { label: "Shopee Spend", currentValue: 5230, previousValue: 4800, format: "currency" },
                { label: "Total Impressions", currentValue: 2567000, previousValue: 2150000, format: "number" },
                { label: "Total Clicks", currentValue: 86540, previousValue: 72300, format: "number" },
                { label: "Total Conversions", currentValue: 2702, previousValue: 2150, format: "number" },
              ]}
            />
          </TabsContent>
        </Tabs>
      </div>
    </div>
  );
}
