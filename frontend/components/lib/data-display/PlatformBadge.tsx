"use client";

import { cn } from "@/lib/utils";
import { type Platform, getPlatformName } from "@/lib/mock-data";

interface PlatformBadgeProps {
  platform: Platform;
  showName?: boolean;
  size?: "xs" | "sm" | "md" | "lg";
  className?: string;
}

const platformConfig: Record<Platform, { bg: string; text: string; icon: string }> = {
  meta: {
    bg: "bg-blue-100",
    text: "text-blue-700",
    icon: "M",
  },
  tiktok: {
    bg: "bg-slate-900",
    text: "text-white",
    icon: "T",
  },
  shopee: {
    bg: "bg-orange-100",
    text: "text-orange-700",
    icon: "S",
  },
};

const sizeConfig = {
  xs: {
    badge: "h-5 px-1.5 text-xs gap-1",
    icon: "h-4 w-4 text-[10px]",
    iconOnly: "h-5 w-5 text-[10px]",
  },
  sm: {
    badge: "h-6 px-2 text-xs gap-1.5",
    icon: "h-5 w-5 text-xs",
    iconOnly: "h-6 w-6 text-xs",
  },
  md: {
    badge: "h-7 px-2.5 text-sm gap-2",
    icon: "h-6 w-6 text-sm",
    iconOnly: "h-8 w-8 text-sm",
  },
  lg: {
    badge: "h-9 px-3 text-base gap-2",
    icon: "h-8 w-8 text-base",
    iconOnly: "h-10 w-10 text-base",
  },
};

export function PlatformBadge({
  platform,
  showName = true,
  size = "sm",
  className,
}: PlatformBadgeProps) {
  const config = platformConfig[platform];
  const sizeStyles = sizeConfig[size];

  if (!showName) {
    return (
      <div
        className={cn(
          "rounded-lg flex items-center justify-center font-bold",
          sizeStyles.iconOnly,
          config.bg,
          config.text,
          className
        )}
      >
        {config.icon}
      </div>
    );
  }

  return (
    <div
      className={cn(
        "inline-flex items-center rounded-full font-medium",
        sizeStyles.badge,
        config.bg,
        config.text,
        className
      )}
    >
      <span
        className={cn(
          "rounded flex items-center justify-center font-bold",
          sizeStyles.icon,
          config.bg,
          config.text
        )}
      >
        {config.icon}
      </span>
      <span>{getPlatformName(platform)}</span>
    </div>
  );
}

// Simple icon-only variant for tables
export function PlatformIcon({
  platform,
  size = "sm",
  className,
}: {
  platform: Platform;
  size?: "xs" | "sm" | "md" | "lg";
  className?: string;
}) {
  return <PlatformBadge platform={platform} showName={false} size={size} className={className} />;
}
