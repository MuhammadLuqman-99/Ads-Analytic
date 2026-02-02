"use client";

import {
  LineChart as RechartsLineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
} from "recharts";
import { cn } from "@/lib/utils";

interface Series {
  key: string;
  name: string;
  color: string;
  strokeWidth?: number;
  dotSize?: number;
}

interface LineChartProps {
  data: Record<string, unknown>[];
  xAxisKey: string;
  series: Series[];
  height?: number;
  showGrid?: boolean;
  showLegend?: boolean;
  showTooltip?: boolean;
  showDots?: boolean;
  curved?: boolean;
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
];

export function LineChart({
  data,
  xAxisKey,
  series,
  height = 300,
  showGrid = true,
  showLegend = true,
  showTooltip = true,
  showDots = true,
  curved = true,
  yAxisFormatter,
  tooltipFormatter,
  className,
}: LineChartProps) {
  return (
    <div className={cn("w-full", className)}>
      <ResponsiveContainer width="100%" height={height}>
        <RechartsLineChart data={data} margin={{ top: 20, right: 30, left: 20, bottom: 5 }}>
          {showGrid && <CartesianGrid strokeDasharray="3 3" stroke="#E2E8F0" />}
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
          {showLegend && (
            <Legend
              wrapperStyle={{ paddingTop: "20px" }}
              iconType="circle"
              iconSize={8}
            />
          )}
          {series.map((s, index) => (
            <Line
              key={s.key}
              type={curved ? "monotone" : "linear"}
              dataKey={s.key}
              name={s.name}
              stroke={s.color || defaultColors[index % defaultColors.length]}
              strokeWidth={s.strokeWidth || 2}
              dot={showDots ? { fill: s.color || defaultColors[index % defaultColors.length], strokeWidth: 2, r: s.dotSize || 4 } : false}
              activeDot={{ r: 6, strokeWidth: 2 }}
            />
          ))}
        </RechartsLineChart>
      </ResponsiveContainer>
    </div>
  );
}
