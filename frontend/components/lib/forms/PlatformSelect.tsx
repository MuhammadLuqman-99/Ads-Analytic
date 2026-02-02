"use client";

import { useState, useRef, useEffect } from "react";
import { Check, ChevronDown, X } from "lucide-react";
import { cn } from "@/lib/utils";

type Platform = "meta" | "google" | "tiktok" | "linkedin" | "twitter" | "snapchat";

interface PlatformOption {
  value: Platform;
  label: string;
  logo: string;
  color: string;
}

const platforms: PlatformOption[] = [
  {
    value: "meta",
    label: "Meta Ads",
    logo: "/logos/meta.svg",
    color: "bg-blue-500",
  },
  {
    value: "google",
    label: "Google Ads",
    logo: "/logos/google.svg",
    color: "bg-red-500",
  },
  {
    value: "tiktok",
    label: "TikTok Ads",
    logo: "/logos/tiktok.svg",
    color: "bg-slate-900",
  },
  {
    value: "linkedin",
    label: "LinkedIn Ads",
    logo: "/logos/linkedin.svg",
    color: "bg-blue-700",
  },
  {
    value: "twitter",
    label: "Twitter/X Ads",
    logo: "/logos/twitter.svg",
    color: "bg-slate-800",
  },
  {
    value: "snapchat",
    label: "Snapchat Ads",
    logo: "/logos/snapchat.svg",
    color: "bg-yellow-400",
  },
];

// Fallback logo component when image fails to load
function PlatformLogo({
  platform,
  size = "md",
}: {
  platform: PlatformOption;
  size?: "sm" | "md" | "lg";
}) {
  const [imageError, setImageError] = useState(false);

  const sizeClasses = {
    sm: "h-4 w-4 text-[10px]",
    md: "h-5 w-5 text-xs",
    lg: "h-6 w-6 text-sm",
  };

  if (imageError) {
    return (
      <div
        className={cn(
          "rounded flex items-center justify-center text-white font-bold",
          platform.color,
          sizeClasses[size]
        )}
      >
        {platform.label[0]}
      </div>
    );
  }

  return (
    <img
      src={platform.logo}
      alt={platform.label}
      className={sizeClasses[size]}
      onError={() => setImageError(true)}
    />
  );
}

interface PlatformSelectProps {
  value?: Platform[];
  onChange?: (value: Platform[]) => void;
  placeholder?: string;
  multiple?: boolean;
  disabled?: boolean;
  className?: string;
  availablePlatforms?: Platform[];
}

export function PlatformSelect({
  value = [],
  onChange,
  placeholder = "Select platforms",
  multiple = true,
  disabled = false,
  className,
  availablePlatforms,
}: PlatformSelectProps) {
  const [isOpen, setIsOpen] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);

  const filteredPlatforms = availablePlatforms
    ? platforms.filter((p) => availablePlatforms.includes(p.value))
    : platforms;

  const selectedPlatforms = filteredPlatforms.filter((p) =>
    value.includes(p.value)
  );

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (
        containerRef.current &&
        !containerRef.current.contains(event.target as Node)
      ) {
        setIsOpen(false);
      }
    };

    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  const handleSelect = (platform: Platform) => {
    if (!multiple) {
      onChange?.([platform]);
      setIsOpen(false);
      return;
    }

    const isSelected = value.includes(platform);
    if (isSelected) {
      onChange?.(value.filter((v) => v !== platform));
    } else {
      onChange?.([...value, platform]);
    }
  };

  const handleRemove = (platform: Platform, e: React.MouseEvent) => {
    e.stopPropagation();
    onChange?.(value.filter((v) => v !== platform));
  };

  const handleClearAll = (e: React.MouseEvent) => {
    e.stopPropagation();
    onChange?.([]);
  };

  return (
    <div ref={containerRef} className={cn("relative", className)}>
      <button
        type="button"
        onClick={() => !disabled && setIsOpen(!isOpen)}
        disabled={disabled}
        className={cn(
          "w-full min-h-[40px] px-3 py-2 border border-slate-200 rounded-lg",
          "flex items-center gap-2 text-left bg-white",
          "hover:border-slate-300 focus:outline-none focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500",
          "disabled:opacity-50 disabled:cursor-not-allowed",
          isOpen && "ring-2 ring-blue-500/20 border-blue-500"
        )}
      >
        <div className="flex-1 flex flex-wrap gap-1.5">
          {selectedPlatforms.length === 0 ? (
            <span className="text-slate-400">{placeholder}</span>
          ) : (
            selectedPlatforms.map((platform) => (
              <span
                key={platform.value}
                className="inline-flex items-center gap-1.5 px-2 py-0.5 bg-slate-100 rounded-md text-sm"
              >
                <PlatformLogo platform={platform} size="sm" />
                <span className="text-slate-700">{platform.label}</span>
                {multiple && (
                  <button
                    onClick={(e) => handleRemove(platform.value, e)}
                    className="text-slate-400 hover:text-slate-600"
                  >
                    <X className="h-3 w-3" />
                  </button>
                )}
              </span>
            ))
          )}
        </div>
        <div className="flex items-center gap-1 ml-2">
          {selectedPlatforms.length > 0 && multiple && (
            <button
              onClick={handleClearAll}
              className="p-0.5 hover:bg-slate-200 rounded"
            >
              <X className="h-4 w-4 text-slate-400" />
            </button>
          )}
          <ChevronDown
            className={cn(
              "h-4 w-4 text-slate-400 transition-transform",
              isOpen && "rotate-180"
            )}
          />
        </div>
      </button>

      {isOpen && (
        <div className="absolute z-50 w-full mt-1 bg-white border border-slate-200 rounded-lg shadow-lg overflow-hidden">
          <div className="max-h-64 overflow-y-auto">
            {filteredPlatforms.map((platform) => {
              const isSelected = value.includes(platform.value);
              return (
                <button
                  key={platform.value}
                  onClick={() => handleSelect(platform.value)}
                  className={cn(
                    "w-full flex items-center gap-3 px-3 py-2.5 text-left",
                    "hover:bg-slate-50 transition-colors",
                    isSelected && "bg-blue-50"
                  )}
                >
                  <PlatformLogo platform={platform} />
                  <span
                    className={cn(
                      "flex-1 text-sm",
                      isSelected ? "text-blue-700 font-medium" : "text-slate-700"
                    )}
                  >
                    {platform.label}
                  </span>
                  {isSelected && (
                    <Check className="h-4 w-4 text-blue-600" />
                  )}
                </button>
              );
            })}
          </div>
          {multiple && filteredPlatforms.length > 1 && (
            <div className="border-t border-slate-200 p-2 flex justify-between">
              <button
                onClick={() => onChange?.(filteredPlatforms.map((p) => p.value))}
                className="text-sm text-blue-600 hover:text-blue-700"
              >
                Select all
              </button>
              <button
                onClick={() => onChange?.([])}
                className="text-sm text-slate-500 hover:text-slate-700"
              >
                Clear all
              </button>
            </div>
          )}
        </div>
      )}
    </div>
  );
}

// Single platform select variant
export function SinglePlatformSelect(
  props: Omit<PlatformSelectProps, "multiple">
) {
  return <PlatformSelect {...props} multiple={false} />;
}

// Platform filter chips for dashboards
export function PlatformFilterChips({
  value = [],
  onChange,
  className,
}: {
  value?: Platform[];
  onChange?: (value: Platform[]) => void;
  className?: string;
}) {
  const handleToggle = (platform: Platform) => {
    const isSelected = value.includes(platform);
    if (isSelected) {
      onChange?.(value.filter((v) => v !== platform));
    } else {
      onChange?.([...value, platform]);
    }
  };

  return (
    <div className={cn("flex flex-wrap gap-2", className)}>
      {platforms.map((platform) => {
        const isSelected = value.includes(platform.value);
        return (
          <button
            key={platform.value}
            onClick={() => handleToggle(platform.value)}
            className={cn(
              "inline-flex items-center gap-2 px-3 py-1.5 rounded-full text-sm transition-colors",
              isSelected
                ? "bg-blue-100 text-blue-700 border-2 border-blue-300"
                : "bg-slate-100 text-slate-600 border-2 border-transparent hover:bg-slate-200"
            )}
          >
            <PlatformLogo platform={platform} size="sm" />
            {platform.label}
          </button>
        );
      })}
    </div>
  );
}
