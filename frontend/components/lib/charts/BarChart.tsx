"use client";

import {
  BarChart as RechartsBarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
  Cell,
} from "recharts";
import { cn } from "@/lib/utils";

interface BarSeries {
  key: string;
  name: string;
  color?: string;
  stackId?: string;
}

interface BarChartProps {
  data: Record<string, unknown>[];
  xAxisKey: string;
  series: BarSeries[];
  height?: number;
  horizontal?: boolean;
  stacked?: boolean;
  showGrid?: boolean;
  showLegend?: boolean;
  showTooltip?: boolean;
  barSize?: number;
  barRadius?: number;
  colors?: string[];
  colorByData?: boolean;
  colorKey?: string;
  yAxisFormatter?: (value: number) => string;
  tooltipFormatter?: (value: number, name: string) => string;
  className?: string;
}

// Helper to wrap formatter for Recharts compatibility
const wrapFormatter = (formatter?: (value: number, name: string) => string) => {
  if (!formatter) return undefined;
  return (value: number | undefined, name: string | undefined) => {
    return formatter(Number(value) || 0, name || "");
  };
};

const defaultColors = [
  "#3B82F6", // blue
  "#10B981", // emerald
  "#F59E0B", // amber
  "#EF4444", // red
  "#8B5CF6", // violet
  "#EC4899", // pink
  "#06B6D4", // cyan
];

export function BarChart({
  data,
  xAxisKey,
  series,
  height = 300,
  horizontal = false,
  stacked = false,
  showGrid = true,
  showLegend = true,
  showTooltip = true,
  barSize = 20,
  barRadius = 4,
  colors = defaultColors,
  colorByData = false,
  colorKey,
  yAxisFormatter,
  tooltipFormatter,
  className,
}: BarChartProps) {
  const ChartComponent = horizontal ? RechartsBarChart : RechartsBarChart;

  return (
    <div className={cn("w-full", className)}>
      <ResponsiveContainer width="100%" height={height}>
        <ChartComponent
          data={data}
          layout={horizontal ? "vertical" : "horizontal"}
          margin={{ top: 20, right: 30, left: 20, bottom: 5 }}
        >
          {showGrid && <CartesianGrid strokeDasharray="3 3" stroke="#E2E8F0" />}

          {horizontal ? (
            <>
              <XAxis
                type="number"
                stroke="#64748B"
                fontSize={12}
                tickLine={false}
                axisLine={{ stroke: "#E2E8F0" }}
                tickFormatter={yAxisFormatter}
              />
              <YAxis
                type="category"
                dataKey={xAxisKey}
                stroke="#64748B"
                fontSize={12}
                tickLine={false}
                axisLine={{ stroke: "#E2E8F0" }}
                width={100}
              />
            </>
          ) : (
            <>
              <XAxis
                dataKey={xAxisKey}
                stroke="#64748B"
                fontSize={12}
                tickLine={false}
                axisLine={{ stroke: "#E2E8F0" }}
              />
              <YAxis
                stroke="#64748B"
                fontSize={12}
                tickLine={false}
                axisLine={{ stroke: "#E2E8F0" }}
                tickFormatter={yAxisFormatter}
              />
            </>
          )}

          {showTooltip && (
            <Tooltip
              formatter={wrapFormatter(tooltipFormatter)}
              contentStyle={{
                backgroundColor: "white",
                border: "1px solid #E2E8F0",
                borderRadius: "8px",
                boxShadow: "0 4px 6px -1px rgb(0 0 0 / 0.1)",
              }}
              labelStyle={{ color: "#1E293B", fontWeight: 500 }}
            />
          )}

          {showLegend && series.length > 1 && (
            <Legend
              wrapperStyle={{ paddingTop: "20px" }}
              iconType="square"
              iconSize={10}
            />
          )}

          {series.map((s, index) => (
            <Bar
              key={s.key}
              dataKey={s.key}
              name={s.name}
              fill={s.color || colors[index % colors.length]}
              barSize={barSize}
              radius={[barRadius, barRadius, 0, 0]}
              stackId={stacked ? "stack" : s.stackId}
            >
              {colorByData && colorKey && data.map((entry, i) => (
                <Cell key={`cell-${i}`} fill={colors[i % colors.length]} />
              ))}
            </Bar>
          ))}
        </ChartComponent>
      </ResponsiveContainer>
    </div>
  );
}
