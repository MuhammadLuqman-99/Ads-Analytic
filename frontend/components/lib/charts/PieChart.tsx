"use client";

import {
  PieChart as RechartsPieChart,
  Pie,
  Cell,
  Tooltip,
  Legend,
  ResponsiveContainer,
} from "recharts";
import { cn } from "@/lib/utils";

interface DataPoint {
  name: string;
  value: number;
  color?: string;
}

interface PieChartProps {
  data: DataPoint[];
  height?: number;
  innerRadius?: number;
  outerRadius?: number;
  showLegend?: boolean;
  showTooltip?: boolean;
  showLabels?: boolean;
  centerLabel?: {
    value: string | number;
    label: string;
  };
  colors?: string[];
  valueFormatter?: (value: number) => string;
  className?: string;
}

const defaultColors = [
  "#3B82F6", // blue
  "#10B981", // emerald
  "#F59E0B", // amber
  "#EF4444", // red
  "#8B5CF6", // violet
  "#EC4899", // pink
  "#06B6D4", // cyan
  "#84CC16", // lime
];

export function PieChart({
  data,
  height = 300,
  innerRadius = 60,
  outerRadius = 100,
  showLegend = true,
  showTooltip = true,
  showLabels = false,
  centerLabel,
  colors = defaultColors,
  valueFormatter = (v) => v.toLocaleString(),
  className,
}: PieChartProps) {
  const total = data.reduce((sum, d) => sum + d.value, 0);

  return (
    <div className={cn("w-full relative", className)}>
      <ResponsiveContainer width="100%" height={height}>
        <RechartsPieChart>
          <Pie
            data={data}
            cx="50%"
            cy="50%"
            innerRadius={centerLabel ? innerRadius : 0}
            outerRadius={outerRadius}
            paddingAngle={2}
            dataKey="value"
            label={
              showLabels
                ? ({ name, percent }) =>
                    `${name} (${((percent ?? 0) * 100).toFixed(0)}%)`
                : undefined
            }
            labelLine={showLabels}
          >
            {data.map((entry, index) => (
              <Cell
                key={`cell-${index}`}
                fill={entry.color || colors[index % colors.length]}
              />
            ))}
          </Pie>

          {showTooltip && (
            <Tooltip
              formatter={(value: number | undefined) => [valueFormatter(Number(value) || 0), "Value"]}
              contentStyle={{
                backgroundColor: "white",
                border: "1px solid #E2E8F0",
                borderRadius: "8px",
                boxShadow: "0 4px 6px -1px rgb(0 0 0 / 0.1)",
              }}
            />
          )}

          {showLegend && (
            <Legend
              layout="vertical"
              align="right"
              verticalAlign="middle"
              iconType="circle"
              iconSize={10}
              formatter={(value, entry) => {
                const item = data.find((d) => d.name === value);
                const percent = item ? ((item.value / total) * 100).toFixed(1) : 0;
                return (
                  <span className="text-sm text-slate-600">
                    {value} ({percent}%)
                  </span>
                );
              }}
            />
          )}
        </RechartsPieChart>
      </ResponsiveContainer>

      {/* Center Label */}
      {centerLabel && (
        <div
          className="absolute left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2 text-center pointer-events-none"
          style={{ marginLeft: showLegend ? "-60px" : "0" }}
        >
          <p className="text-2xl font-bold text-slate-900">{centerLabel.value}</p>
          <p className="text-sm text-slate-500">{centerLabel.label}</p>
        </div>
      )}
    </div>
  );
}

// Donut variant with center label
export function DonutChart(props: Omit<PieChartProps, "innerRadius"> & { innerRadius?: number }) {
  return <PieChart {...props} innerRadius={props.innerRadius ?? 60} />;
}
