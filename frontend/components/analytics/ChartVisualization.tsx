"use client";

import { useState } from "react";
import { BarChart3, LineChartIcon, PieChartIcon, Table2 } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { cn } from "@/lib/utils";
import { formatCurrency, formatNumber } from "@/lib/mock-data";
import {
  LineChart as LibLineChart,
  BarChart as LibBarChart,
  PieChart as LibPieChart,
} from "@/components/lib";

export type ChartType = "line" | "bar" | "pie" | "table";

interface DataPoint {
  name: string;
  [key: string]: string | number;
}

interface ChartVisualizationProps {
  title: string;
  data: DataPoint[];
  metrics: string[];
  defaultChartType?: ChartType;
  allowedChartTypes?: ChartType[];
  colors?: string[];
  className?: string;
}

const CHART_COLORS = [
  "#3B82F6", // blue
  "#10B981", // emerald
  "#F59E0B", // amber
  "#EF4444", // red
  "#8B5CF6", // violet
  "#EC4899", // pink
  "#06B6D4", // cyan
  "#84CC16", // lime
];

const chartTypeIcons = {
  line: LineChartIcon,
  bar: BarChart3,
  pie: PieChartIcon,
  table: Table2,
};

export function ChartVisualization({
  title,
  data,
  metrics,
  defaultChartType = "line",
  allowedChartTypes = ["line", "bar", "pie", "table"],
  colors = CHART_COLORS,
  className,
}: ChartVisualizationProps) {
  const [chartType, setChartType] = useState<ChartType>(defaultChartType);

  const formatValue = (value: number, metric: string) => {
    if (metric === "spend" || metric === "revenue" || metric === "cpc" || metric === "cpm") {
      return formatCurrency(value);
    }
    if (metric === "ctr" || metric === "roas") {
      return `${value.toFixed(2)}${metric === "ctr" ? "%" : "x"}`;
    }
    return formatNumber(value);
  };

  const renderChart = () => {
    switch (chartType) {
      case "line":
        return (
          <LibLineChart
            data={data}
            xAxisKey="name"
            series={metrics.map((metric, index) => ({
              key: metric,
              name: metric.charAt(0).toUpperCase() + metric.slice(1),
              color: colors[index % colors.length],
            }))}
            height={350}
          />
        );

      case "bar":
        return (
          <LibBarChart
            data={data}
            xAxisKey="name"
            series={metrics.map((metric, index) => ({
              key: metric,
              name: metric.charAt(0).toUpperCase() + metric.slice(1),
              color: colors[index % colors.length],
            }))}
            height={350}
          />
        );

      case "pie":
        // For pie chart, we need to restructure data
        const simplePieData =
          metrics.length === 1
            ? data.map((d, i) => ({
                name: d.name,
                value: d[metrics[0]] as number,
                color: colors[i % colors.length],
              }))
            : metrics.flatMap((metric, mIndex) =>
                data.map((d, dIndex) => ({
                  name: `${d.name} - ${metric}`,
                  value: d[metric] as number,
                  color: colors[(mIndex * data.length + dIndex) % colors.length],
                }))
              );

        return (
          <LibPieChart
            data={simplePieData}
            height={350}
            showLabels
            colors={colors}
          />
        );

      case "table":
        return (
          <div className="max-h-[350px] overflow-auto">
            <Table>
              <TableHeader>
                <TableRow className="bg-slate-50">
                  <TableHead className="text-slate-500">Name</TableHead>
                  {metrics.map((metric) => (
                    <TableHead key={metric} className="text-slate-500 text-right">
                      {metric.charAt(0).toUpperCase() + metric.slice(1)}
                    </TableHead>
                  ))}
                </TableRow>
              </TableHeader>
              <TableBody>
                {data.map((row, index) => (
                  <TableRow key={index} className="hover:bg-slate-50">
                    <TableCell className="font-medium text-slate-900">
                      {row.name}
                    </TableCell>
                    {metrics.map((metric) => (
                      <TableCell key={metric} className="text-right">
                        {formatValue(row[metric] as number, metric)}
                      </TableCell>
                    ))}
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        );
    }
  };

  return (
    <Card className={cn("bg-white border-slate-200", className)}>
      <CardHeader className="flex flex-row items-center justify-between pb-2">
        <CardTitle className="text-slate-900">{title}</CardTitle>
        <div className="flex items-center gap-1 bg-slate-100 rounded-lg p-1">
          {allowedChartTypes.map((type) => {
            const Icon = chartTypeIcons[type];
            return (
              <Button
                key={type}
                variant="ghost"
                size="sm"
                className={cn(
                  "h-8 w-8 p-0",
                  chartType === type && "bg-white shadow-sm"
                )}
                onClick={() => setChartType(type)}
              >
                <Icon className="h-4 w-4" />
              </Button>
            );
          })}
        </div>
      </CardHeader>
      <CardContent>{renderChart()}</CardContent>
    </Card>
  );
}
