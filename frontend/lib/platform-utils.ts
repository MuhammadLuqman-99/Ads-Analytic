import type { Platform } from "./api/types";

/**
 * Platform display configuration
 */
export const platformConfig: Record<
  Platform,
  { name: string; color: string; bgColor: string; textColor: string }
> = {
  meta: {
    name: "Meta",
    color: "#3b82f6", // blue-500
    bgColor: "bg-blue-100",
    textColor: "text-blue-600",
  },
  google: {
    name: "Google",
    color: "#ef4444", // red-500
    bgColor: "bg-red-100",
    textColor: "text-red-600",
  },
  tiktok: {
    name: "TikTok",
    color: "#64748b", // slate-500
    bgColor: "bg-slate-900",
    textColor: "text-white",
  },
  shopee: {
    name: "Shopee",
    color: "#f97316", // orange-500
    bgColor: "bg-orange-100",
    textColor: "text-orange-600",
  },
  linkedin: {
    name: "LinkedIn",
    color: "#0077b5",
    bgColor: "bg-blue-100",
    textColor: "text-blue-700",
  },
};

/**
 * Get human-readable platform name
 */
export function getPlatformName(platform: Platform): string {
  return platformConfig[platform]?.name ?? platform;
}

/**
 * Get platform color (hex)
 */
export function getPlatformColor(platform: Platform): string {
  return platformConfig[platform]?.color ?? "#64748b";
}

/**
 * Get platform background CSS class
 */
export function getPlatformBgClass(platform: Platform): string {
  return platformConfig[platform]?.bgColor ?? "bg-slate-100";
}

/**
 * Get platform text CSS class
 */
export function getPlatformTextClass(platform: Platform): string {
  return platformConfig[platform]?.textColor ?? "text-slate-600";
}

/**
 * Get platform logo initial (for fallback display)
 */
export function getPlatformInitial(platform: Platform): string {
  const initials: Record<Platform, string> = {
    meta: "M",
    google: "G",
    tiktok: "T",
    shopee: "S",
    linkedin: "L",
  };
  return initials[platform] ?? platform.charAt(0).toUpperCase();
}

/**
 * Format currency
 */
export function formatCurrency(
  amount: number,
  currency: string = "MYR"
): string {
  return new Intl.NumberFormat("ms-MY", {
    style: "currency",
    currency: currency,
    minimumFractionDigits: 0,
    maximumFractionDigits: 0,
  }).format(amount);
}

/**
 * Format large numbers with K/M suffix
 */
export function formatNumber(num: number): string {
  if (num >= 1000000) {
    return `${(num / 1000000).toFixed(1)}M`;
  }
  if (num >= 1000) {
    return `${(num / 1000).toFixed(1)}K`;
  }
  return num.toLocaleString();
}

/**
 * Format percentage
 */
export function formatPercentage(value: number, decimals: number = 1): string {
  return `${value >= 0 ? "+" : ""}${value.toFixed(decimals)}%`;
}
