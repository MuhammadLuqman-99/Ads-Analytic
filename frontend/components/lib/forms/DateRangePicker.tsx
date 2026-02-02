"use client";

import { useState, useCallback } from "react";
import { Calendar, ChevronDown, X } from "lucide-react";
import { format, subDays, startOfMonth, endOfMonth, startOfYear } from "date-fns";
import { Button } from "@/components/ui/button";
import { Calendar as CalendarComponent } from "@/components/ui/calendar";
import { cn } from "@/lib/utils";

export interface DateRange {
  from: Date;
  to: Date;
}

interface DateRangePreset {
  label: string;
  value: string;
  getRange: () => DateRange;
}

const defaultPresets: DateRangePreset[] = [
  {
    label: "Last 7 days",
    value: "7d",
    getRange: () => ({
      from: subDays(new Date(), 7),
      to: new Date(),
    }),
  },
  {
    label: "Last 14 days",
    value: "14d",
    getRange: () => ({
      from: subDays(new Date(), 14),
      to: new Date(),
    }),
  },
  {
    label: "Last 30 days",
    value: "30d",
    getRange: () => ({
      from: subDays(new Date(), 30),
      to: new Date(),
    }),
  },
  {
    label: "Last 90 days",
    value: "90d",
    getRange: () => ({
      from: subDays(new Date(), 90),
      to: new Date(),
    }),
  },
  {
    label: "This month",
    value: "this-month",
    getRange: () => ({
      from: startOfMonth(new Date()),
      to: new Date(),
    }),
  },
  {
    label: "Last month",
    value: "last-month",
    getRange: () => {
      const lastMonth = subDays(startOfMonth(new Date()), 1);
      return {
        from: startOfMonth(lastMonth),
        to: endOfMonth(lastMonth),
      };
    },
  },
  {
    label: "Year to date",
    value: "ytd",
    getRange: () => ({
      from: startOfYear(new Date()),
      to: new Date(),
    }),
  },
];

interface DateRangePickerProps {
  value?: DateRange;
  onChange?: (range: DateRange) => void;
  presets?: DateRangePreset[];
  showPresets?: boolean;
  placeholder?: string;
  align?: "start" | "center" | "end";
  className?: string;
  disabled?: boolean;
}

export function DateRangePicker({
  value,
  onChange,
  presets = defaultPresets,
  showPresets = true,
  placeholder = "Select date range",
  align = "start",
  className,
  disabled = false,
}: DateRangePickerProps) {
  const [isOpen, setIsOpen] = useState(false);
  const [selectedPreset, setSelectedPreset] = useState<string | null>(null);
  const [tempRange, setTempRange] = useState<{
    from: Date | undefined;
    to: Date | undefined;
  }>({
    from: value?.from,
    to: value?.to,
  });

  const handlePresetClick = useCallback(
    (preset: DateRangePreset) => {
      const range = preset.getRange();
      setSelectedPreset(preset.value);
      setTempRange(range);
      onChange?.(range);
      setIsOpen(false);
    },
    [onChange]
  );

  const handleDateSelect = useCallback(
    (range: { from: Date | undefined; to: Date | undefined } | undefined) => {
      if (!range) return;

      setTempRange(range);
      setSelectedPreset(null);

      if (range.from && range.to) {
        onChange?.({ from: range.from, to: range.to });
      }
    },
    [onChange]
  );

  const handleClear = useCallback(
    (e: React.MouseEvent) => {
      e.stopPropagation();
      setTempRange({ from: undefined, to: undefined });
      setSelectedPreset(null);
    },
    []
  );

  const formatDateRange = () => {
    if (!value?.from) return placeholder;

    if (selectedPreset) {
      const preset = presets.find((p) => p.value === selectedPreset);
      if (preset) return preset.label;
    }

    if (value.to) {
      return `${format(value.from, "MMM d, yyyy")} - ${format(
        value.to,
        "MMM d, yyyy"
      )}`;
    }

    return format(value.from, "MMM d, yyyy");
  };

  return (
    <div className={cn("relative", className)}>
      <Button
        variant="outline"
        onClick={() => !disabled && setIsOpen(!isOpen)}
        disabled={disabled}
        className={cn(
          "w-full justify-between text-left font-normal",
          !value && "text-slate-500"
        )}
      >
        <div className="flex items-center gap-2">
          <Calendar className="h-4 w-4 text-slate-400" />
          <span className="truncate">{formatDateRange()}</span>
        </div>
        <div className="flex items-center gap-1">
          {value?.from && (
            <span
              onClick={handleClear}
              className="p-0.5 hover:bg-slate-200 rounded cursor-pointer"
            >
              <X className="h-3 w-3 text-slate-400" />
            </span>
          )}
          <ChevronDown className="h-4 w-4 text-slate-400" />
        </div>
      </Button>

      {isOpen && (
        <>
          <div
            className="fixed inset-0 z-40"
            onClick={() => setIsOpen(false)}
          />
          <div
            className={cn(
              "absolute z-50 mt-2 bg-white rounded-lg shadow-lg border border-slate-200",
              align === "start" && "left-0",
              align === "center" && "left-1/2 -translate-x-1/2",
              align === "end" && "right-0"
            )}
          >
            <div className="flex">
              {showPresets && (
                <div className="border-r border-slate-200 p-3 w-40">
                  <p className="text-xs font-medium text-slate-500 mb-2">
                    Quick select
                  </p>
                  <div className="space-y-1">
                    {presets.map((preset) => (
                      <button
                        key={preset.value}
                        onClick={() => handlePresetClick(preset)}
                        className={cn(
                          "w-full text-left px-2 py-1.5 text-sm rounded hover:bg-slate-100 transition-colors",
                          selectedPreset === preset.value &&
                            "bg-blue-50 text-blue-600 font-medium"
                        )}
                      >
                        {preset.label}
                      </button>
                    ))}
                  </div>
                </div>
              )}
              <div className="p-3">
                <CalendarComponent
                  mode="range"
                  selected={tempRange}
                  onSelect={handleDateSelect as (range: { from?: Date; to?: Date } | undefined) => void}
                  numberOfMonths={2}
                  disabled={{ after: new Date() }}
                />
                <div className="flex justify-end gap-2 mt-3 pt-3 border-t border-slate-200">
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => setIsOpen(false)}
                  >
                    Cancel
                  </Button>
                  <Button
                    size="sm"
                    onClick={() => {
                      if (tempRange.from && tempRange.to) {
                        onChange?.({ from: tempRange.from, to: tempRange.to });
                      }
                      setIsOpen(false);
                    }}
                    disabled={!tempRange.from || !tempRange.to}
                  >
                    Apply
                  </Button>
                </div>
              </div>
            </div>
          </div>
        </>
      )}
    </div>
  );
}

// Compact version for toolbars
export function CompactDateRangePicker({
  value,
  onChange,
  className,
}: Pick<DateRangePickerProps, "value" | "onChange" | "className">) {
  return (
    <DateRangePicker
      value={value}
      onChange={onChange}
      showPresets={false}
      className={cn("w-auto", className)}
    />
  );
}
