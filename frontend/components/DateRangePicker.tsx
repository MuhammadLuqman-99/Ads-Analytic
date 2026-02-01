"use client"

import * as React from "react"
import { CalendarIcon } from "lucide-react"
import { format, subDays, startOfMonth, endOfMonth, subMonths } from "date-fns"

import { cn } from "@/lib/utils"
import { Button } from "@/components/ui/button"
import { Calendar } from "@/components/ui/calendar"
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover"

export interface DateRange {
    from: Date
    to: Date
}

interface DateRangePickerProps {
    dateRange: DateRange
    onDateRangeChange: (range: DateRange) => void
}

const presets = [
    {
        label: "Last 7 days",
        getValue: () => ({
            from: subDays(new Date(), 7),
            to: new Date(),
        }),
    },
    {
        label: "Last 14 days",
        getValue: () => ({
            from: subDays(new Date(), 14),
            to: new Date(),
        }),
    },
    {
        label: "Last 30 days",
        getValue: () => ({
            from: subDays(new Date(), 30),
            to: new Date(),
        }),
    },
    {
        label: "This month",
        getValue: () => ({
            from: startOfMonth(new Date()),
            to: endOfMonth(new Date()),
        }),
    },
    {
        label: "Last month",
        getValue: () => ({
            from: startOfMonth(subMonths(new Date(), 1)),
            to: endOfMonth(subMonths(new Date(), 1)),
        }),
    },
]

export function DateRangePicker({ dateRange, onDateRangeChange }: DateRangePickerProps) {
    const [open, setOpen] = React.useState(false)
    const [selectingStart, setSelectingStart] = React.useState(true)

    const handleDateSelect = (date: Date) => {
        if (selectingStart) {
            onDateRangeChange({ from: date, to: dateRange.to })
            setSelectingStart(false)
        } else {
            if (date < dateRange.from) {
                onDateRangeChange({ from: date, to: dateRange.from })
            } else {
                onDateRangeChange({ from: dateRange.from, to: date })
            }
            setSelectingStart(true)
            setOpen(false)
        }
    }

    const handlePresetClick = (preset: typeof presets[0]) => {
        onDateRangeChange(preset.getValue())
        setOpen(false)
    }

    return (
        <Popover open={open} onOpenChange={setOpen}>
            <PopoverTrigger asChild>
                <Button
                    variant="outline"
                    className={cn(
                        "h-10 gap-2 min-w-[280px] justify-start text-left font-normal",
                        !dateRange && "text-muted-foreground"
                    )}
                >
                    <CalendarIcon className="h-4 w-4" />
                    {dateRange?.from ? (
                        dateRange.to ? (
                            <>
                                {format(dateRange.from, "MMM dd, yyyy")} -{" "}
                                {format(dateRange.to, "MMM dd, yyyy")}
                            </>
                        ) : (
                            format(dateRange.from, "MMM dd, yyyy")
                        )
                    ) : (
                        <span>Pick a date range</span>
                    )}
                </Button>
            </PopoverTrigger>
            <PopoverContent className="w-auto p-0" align="start">
                <div className="flex">
                    {/* Presets */}
                    <div className="border-r border-slate-700 p-3 space-y-1">
                        <p className="text-xs text-slate-400 font-medium mb-2 px-2">Quick Select</p>
                        {presets.map((preset) => (
                            <button
                                key={preset.label}
                                onClick={() => handlePresetClick(preset)}
                                className="w-full px-3 py-1.5 text-sm text-left text-slate-300 hover:bg-slate-700 rounded-md transition-colors"
                            >
                                {preset.label}
                            </button>
                        ))}
                    </div>

                    {/* Calendar */}
                    <div className="p-3">
                        <div className="mb-2 text-xs text-slate-400 text-center">
                            {selectingStart ? "Select start date" : "Select end date"}
                        </div>
                        <Calendar
                            selected={selectingStart ? dateRange.from : dateRange.to}
                            onSelect={handleDateSelect}
                        />
                    </div>
                </div>

                {/* Selected Range Display */}
                <div className="border-t border-slate-700 p-3 flex items-center justify-between">
                    <div className="text-sm text-slate-400">
                        <span className="text-white">{format(dateRange.from, "MMM dd")}</span>
                        {" → "}
                        <span className="text-white">{format(dateRange.to, "MMM dd, yyyy")}</span>
                    </div>
                    <Button
                        size="sm"
                        onClick={() => setOpen(false)}
                    >
                        Apply
                    </Button>
                </div>
            </PopoverContent>
        </Popover>
    )
}
