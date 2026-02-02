"use client";

import { cn } from "@/lib/utils";

interface SkeletonProps {
  className?: string;
  style?: React.CSSProperties;
}

// Base skeleton element
export function Skeleton({ className, style }: SkeletonProps) {
  return (
    <div className={cn("animate-pulse bg-slate-200 rounded", className)} style={style} />
  );
}

// Card skeleton
export function CardSkeleton({ className }: SkeletonProps) {
  return (
    <div className={cn("bg-white border border-slate-200 rounded-lg p-6", className)}>
      <Skeleton className="h-4 w-24 mb-2" />
      <Skeleton className="h-8 w-32 mb-4" />
      <Skeleton className="h-3 w-20" />
    </div>
  );
}

// Metric card skeleton
export function MetricCardSkeleton({ className }: SkeletonProps) {
  return (
    <div className={cn("bg-white border border-slate-200 rounded-lg p-6", className)}>
      <div className="flex items-start justify-between">
        <div className="flex-1">
          <Skeleton className="h-4 w-20 mb-2" />
          <Skeleton className="h-8 w-28 mb-3" />
          <Skeleton className="h-3 w-16" />
        </div>
        <Skeleton className="h-10 w-10 rounded-lg" />
      </div>
    </div>
  );
}

// Table row skeleton
export function TableRowSkeleton({ columns = 5, className }: { columns?: number } & SkeletonProps) {
  return (
    <div className={cn("flex items-center gap-4 p-4 border-b border-slate-100", className)}>
      {Array.from({ length: columns }).map((_, i) => (
        <Skeleton key={i} className="h-4 flex-1" />
      ))}
    </div>
  );
}

// Table skeleton
export function TableSkeleton({ rows = 5, columns = 5, className }: { rows?: number; columns?: number } & SkeletonProps) {
  return (
    <div className={cn("bg-white border border-slate-200 rounded-lg overflow-hidden", className)}>
      {/* Header */}
      <div className="flex items-center gap-4 p-4 bg-slate-50 border-b border-slate-200">
        {Array.from({ length: columns }).map((_, i) => (
          <Skeleton key={i} className="h-4 flex-1" />
        ))}
      </div>
      {/* Rows */}
      {Array.from({ length: rows }).map((_, i) => (
        <TableRowSkeleton key={i} columns={columns} />
      ))}
    </div>
  );
}

// Chart skeleton
export function ChartSkeleton({ height = 300, className }: { height?: number } & SkeletonProps) {
  return (
    <div className={cn("bg-white border border-slate-200 rounded-lg p-6", className)}>
      <Skeleton className="h-5 w-32 mb-6" />
      <div className="flex items-end gap-2" style={{ height }}>
        {Array.from({ length: 12 }).map((_, i) => (
          <Skeleton
            key={i}
            className="flex-1"
            style={{ height: `${Math.random() * 60 + 40}%` }}
          />
        ))}
      </div>
    </div>
  );
}

// List item skeleton
export function ListItemSkeleton({ className }: SkeletonProps) {
  return (
    <div className={cn("flex items-center gap-4 p-4", className)}>
      <Skeleton className="h-10 w-10 rounded-full" />
      <div className="flex-1">
        <Skeleton className="h-4 w-32 mb-2" />
        <Skeleton className="h-3 w-24" />
      </div>
      <Skeleton className="h-6 w-16 rounded-full" />
    </div>
  );
}

// Avatar skeleton
export function AvatarSkeleton({ size = "md", className }: { size?: "sm" | "md" | "lg" } & SkeletonProps) {
  const sizes = {
    sm: "h-8 w-8",
    md: "h-10 w-10",
    lg: "h-12 w-12",
  };

  return <Skeleton className={cn("rounded-full", sizes[size], className)} />;
}

// Text skeleton
export function TextSkeleton({ lines = 3, className }: { lines?: number } & SkeletonProps) {
  return (
    <div className={cn("space-y-2", className)}>
      {Array.from({ length: lines }).map((_, i) => (
        <Skeleton
          key={i}
          className="h-4"
          style={{ width: i === lines - 1 ? "60%" : "100%" }}
        />
      ))}
    </div>
  );
}

// Form field skeleton
export function FormFieldSkeleton({ className }: SkeletonProps) {
  return (
    <div className={cn("space-y-2", className)}>
      <Skeleton className="h-4 w-20" />
      <Skeleton className="h-10 w-full rounded-md" />
    </div>
  );
}
