"use client"

import * as React from "react"
import { ChevronLeft, ChevronRight } from "lucide-react"
import {
    format,
    addMonths,
    subMonths,
    startOfMonth,
    endOfMonth,
    eachDayOfInterval,
    isSameMonth,
    isSameDay,
    startOfWeek,
    endOfWeek,
    isWithinInterval,
    isBefore,
    isAfter
} from "date-fns"

import { cn } from "@/lib/utils"
import { Button } from "@/components/ui/button"

// Single date selection props
export interface CalendarSingleProps {
    mode?: "single"
    selected?: Date | null
    onSelect?: (date: Date) => void
    className?: string
    numberOfMonths?: number
}

// Range date selection props
export interface CalendarRangeProps {
    mode: "range"
    selected?: { from: Date; to: Date } | undefined
    onSelect?: (range: { from?: Date; to?: Date } | undefined) => void
    className?: string
    numberOfMonths?: number
}

export type CalendarProps = CalendarSingleProps | CalendarRangeProps

function isRangeMode(props: CalendarProps): props is CalendarRangeProps {
    return props.mode === "range"
}

export function Calendar(props: CalendarProps) {
    const { className, numberOfMonths = 1 } = props

    const getInitialMonth = () => {
        if (isRangeMode(props)) {
            return props.selected?.from || new Date()
        }
        return props.selected || new Date()
    }

    const [currentMonth, setCurrentMonth] = React.useState(getInitialMonth())
    const [rangeStart, setRangeStart] = React.useState<Date | null>(null)

    const weekDays = ['Su', 'Mo', 'Tu', 'We', 'Th', 'Fr', 'Sa']

    const handleDayClick = (day: Date) => {
        if (isRangeMode(props)) {
            if (!rangeStart) {
                setRangeStart(day)
                props.onSelect?.({ from: day, to: undefined })
            } else {
                if (isBefore(day, rangeStart)) {
                    props.onSelect?.({ from: day, to: rangeStart })
                } else {
                    props.onSelect?.({ from: rangeStart, to: day })
                }
                setRangeStart(null)
            }
        } else {
            props.onSelect?.(day)
        }
    }

    const isDateSelected = (day: Date): boolean => {
        if (isRangeMode(props)) {
            if (!props.selected) return false
            if (props.selected.from && props.selected.to) {
                return isSameDay(day, props.selected.from) || isSameDay(day, props.selected.to)
            }
            return props.selected.from ? isSameDay(day, props.selected.from) : false
        }
        return props.selected ? isSameDay(day, props.selected) : false
    }

    const isDateInRange = (day: Date): boolean => {
        if (!isRangeMode(props) || !props.selected?.from || !props.selected?.to) return false
        return isWithinInterval(day, { start: props.selected.from, end: props.selected.to })
    }

    const isRangeStart = (day: Date): boolean => {
        if (!isRangeMode(props) || !props.selected?.from) return false
        return isSameDay(day, props.selected.from)
    }

    const isRangeEnd = (day: Date): boolean => {
        if (!isRangeMode(props) || !props.selected?.to) return false
        return isSameDay(day, props.selected.to)
    }

    const renderMonth = (monthOffset: number) => {
        const month = addMonths(currentMonth, monthOffset)
        const days = eachDayOfInterval({
            start: startOfWeek(startOfMonth(month)),
            end: endOfWeek(endOfMonth(month))
        })

        return (
            <div key={monthOffset}>
                <div className="text-center mb-4">
                    <span className="text-sm font-medium text-slate-900">
                        {format(month, 'MMMM yyyy')}
                    </span>
                </div>
                <div className="grid grid-cols-7 gap-1 mb-2">
                    {weekDays.map((day) => (
                        <div
                            key={day}
                            className="text-center text-xs font-medium text-slate-500 py-1"
                        >
                            {day}
                        </div>
                    ))}
                </div>
                <div className="grid grid-cols-7 gap-1">
                    {days.map((day, idx) => {
                        const selected = isDateSelected(day)
                        const inRange = isDateInRange(day)
                        const isStart = isRangeStart(day)
                        const isEnd = isRangeEnd(day)
                        const isCurrentMonth = isSameMonth(day, month)
                        const isToday = isSameDay(day, new Date())

                        return (
                            <button
                                key={idx}
                                onClick={() => handleDayClick(day)}
                                className={cn(
                                    "h-8 w-8 rounded-md text-sm transition-colors",
                                    "hover:bg-slate-100 focus:outline-none focus:ring-2 focus:ring-blue-500",
                                    !isCurrentMonth && "text-slate-400",
                                    isCurrentMonth && "text-slate-900",
                                    isToday && !selected && "border border-blue-500",
                                    selected && "bg-blue-600 text-white hover:bg-blue-500",
                                    inRange && !selected && "bg-blue-100",
                                    isStart && "rounded-r-none",
                                    isEnd && "rounded-l-none",
                                    inRange && !isStart && !isEnd && "rounded-none"
                                )}
                            >
                                {format(day, 'd')}
                            </button>
                        )
                    })}
                </div>
            </div>
        )
    }

    return (
        <div className={cn("p-3", className)}>
            <div className="flex items-center justify-between mb-4">
                <Button
                    variant="ghost"
                    size="icon"
                    onClick={() => setCurrentMonth(subMonths(currentMonth, 1))}
                    className="h-7 w-7"
                >
                    <ChevronLeft className="h-4 w-4" />
                </Button>
                <div className="flex-1" />
                <Button
                    variant="ghost"
                    size="icon"
                    onClick={() => setCurrentMonth(addMonths(currentMonth, 1))}
                    className="h-7 w-7"
                >
                    <ChevronRight className="h-4 w-4" />
                </Button>
            </div>
            <div className={cn(
                "grid gap-4",
                numberOfMonths === 2 && "grid-cols-2"
            )}>
                {Array.from({ length: numberOfMonths }, (_, i) => renderMonth(i))}
            </div>
        </div>
    )
}
