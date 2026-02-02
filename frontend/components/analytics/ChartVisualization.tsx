"use client";

import { useState } from "react";
import {
  LineChart,
  Line,
  BarChart,
  Bar,
  PieChart,
  Pie,
  Cell,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
} from "recharts";
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
          <ResponsiveContainer width="100%" height={350}>
            <LineChart data={data} margin={{ top: 20, right: 30, left: 20, bottom: 5 }}>
              <CartesianGrid strokeDasharray="3 3" stroke="#E2E8F0" />
              <XAxis
                dataKey="name"
                stroke="#64748B"
                fontSize={12}
                tickLine={false}
              />
              <YAxis stroke="#64748B" fontSize={12} tickLine={false} />
              <Tooltip
                contentStyle={{
                  backgroundColor: "white",
                  border: "1px solid #E2E8F0",
                  borderRadius: "8px",
                  boxShadow: "0 4px 6px -1px rgb(0 0 0 / 0.1)",
                }}
              />
              <Legend />
              {metrics.map((metric, index) => (
                <Line
                  key={metric}
                  type="monotone"
                  dataKey={metric}
                  stroke={colors[index % colors.length]}
                  strokeWidth={2}
                  dot={{ fill: colors[index % colors.length], strokeWidth: 2 }}
                  activeDot={{ r: 6 }}
                />
              ))}
            </LineChart>
          </ResponsiveContainer>
        );

      case "bar":
        return (
          <ResponsiveContainer width="100%" height={350}>
            <BarChart data={data} margin={{ top: 20, right: 30, left: 20, bottom: 5 }}>
              <CartesianGrid strokeDasharray="3 3" stroke="#E2E8F0" />
              <XAxis
                dataKey="name"
                stroke="#64748B"
                fontSize={12}
                tickLine={false}
              />
              <YAxis stroke="#64748B" fontSize={12} tickLine={false} />
              <Tooltip
                contentStyle={{
                  backgroundColor: "white",
                  border: "1px solid #E2E8F0",
                  borderRadius: "8px",
                  boxShadow: "0 4px 6px -1px rgb(0 0 0 / 0.1)",
                }}
              />
              <Legend />
              {metrics.map((metric, index) => (
                <Bar
                  key={metric}
                  dataKey={metric}
                  fill={colors[index % colors.length]}
                  radius={[4, 4, 0, 0]}
                />
              ))}
            </BarChart>
          </ResponsiveContainer>
        );

      case "pie":
        // For pie chart, we need to restructure data - always include fill
        const pieData = metrics.flatMap((metric, mIndex) =>
          data.map((d, dIndex) => ({
            name: `${d.name} - ${metric}`,
            value: d[metric] as number,
            fill: colors[(mIndex * data.length + dIndex) % colors.length],
          }))
        );

        // If single metric, show by category
        const simplePieData =
          metrics.length === 1
            ? data.map((d, i) => ({
                name: d.name,
                value: d[metrics[0]] as number,
                fill: colors[i % colors.length],
              }))
            : pieData;

        return (
          <ResponsiveContainer width="100%" height={350}>
            <PieChart>
              <Pie
                data={simplePieData}
                cx="50%"
                cy="50%"
                labelLine={false}
                label={({ name, percent }) =>
                  `${name} (${((percent ?? 0) * 100).toFixed(0)}%)`
                }
                outerRadius={120}
                fill="#8884d8"
                dataKey="value"
              >
                {simplePieData.map((entry, index) => (
                  <Cell
                    key={`cell-${index}`}
                    fill={entry.fill}
                  />
                ))}
              </Pie>
              <Tooltip
                contentStyle={{
                  backgroundColor: "white",
                  border: "1px solid #E2E8F0",
                  borderRadius: "8px",
                  boxShadow: "0 4px 6px -1px rgb(0 0 0 / 0.1)",
                }}
              />
              <Legend />
            </PieChart>
          </ResponsiveContainer>
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
