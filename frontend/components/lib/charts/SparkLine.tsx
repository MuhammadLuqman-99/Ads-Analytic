"use client";

import { LineChart, Line, ResponsiveContainer, YAxis } from "recharts";
import { cn } from "@/lib/utils";

interface SparkLineProps {
  data: number[];
  color?: string;
  height?: number;
  width?: number;
  showFill?: boolean;
  strokeWidth?: number;
  trend?: "up" | "down" | "neutral";
  className?: string;
}

export function SparkLine({
  data,
  color,
  height = 30,
  width = 80,
  showFill = false,
  strokeWidth = 1.5,
  trend,
  className,
}: SparkLineProps) {
  // Auto-detect trend if not provided
  const effectiveTrend = trend ?? (
    data.length >= 2
      ? data[data.length - 1] > data[0]
        ? "up"
        : data[data.length - 1] < data[0]
        ? "down"
        : "neutral"
      : "neutral"
  );

  // Default colors based on trend
  const defaultColor = {
    up: "#10B981",    // emerald
    down: "#EF4444",  // red
    neutral: "#64748B", // slate
  };

  const lineColor = color || defaultColor[effectiveTrend];

  // Format data for recharts
  const chartData = data.map((value, index) => ({ value, index }));

  return (
    <div className={cn("inline-block", className)} style={{ width, height }}>
      <ResponsiveContainer width="100%" height="100%">
        <LineChart data={chartData} margin={{ top: 2, right: 2, left: 2, bottom: 2 }}>
          <YAxis hide domain={["dataMin", "dataMax"]} />
          <Line
            type="monotone"
            dataKey="value"
            stroke={lineColor}
            strokeWidth={strokeWidth}
            dot={false}
            fill={showFill ? lineColor : "none"}
            fillOpacity={showFill ? 0.1 : 0}
          />
        </LineChart>
      </ResponsiveContainer>
    </div>
  );
}

// SparkBar variant for bar representation
interface SparkBarProps {
  data: number[];
  color?: string;
  height?: number;
  width?: number;
  gap?: number;
  className?: string;
}

export function SparkBar({
  data,
  color = "#3B82F6",
  height = 30,
  width = 80,
  gap = 2,
  className,
}: SparkBarProps) {
  const max = Math.max(...data);
  const barWidth = (width - (data.length - 1) * gap) / data.length;

  return (
    <div
      className={cn("inline-flex items-end", className)}
      style={{ width, height, gap }}
    >
      {data.map((value, index) => (
        <div
          key={index}
          className="rounded-t"
          style={{
            width: barWidth,
            height: `${(value / max) * 100}%`,
            backgroundColor: color,
            minHeight: 2,
          }}
        />
      ))}
    </div>
  );
}

// SparkArea variant
export function SparkArea({
  data,
  color = "#3B82F6",
  height = 30,
  width = 80,
  className,
}: SparkLineProps) {
  return (
    <SparkLine
      data={data}
      color={color}
      height={height}
      width={width}
      showFill
      className={className}
    />
  );
}
