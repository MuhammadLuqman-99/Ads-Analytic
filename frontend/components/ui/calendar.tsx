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
    endOfWeek
} from "date-fns"

import { cn } from "@/lib/utils"
import { Button } from "@/components/ui/button"

export interface CalendarProps {
    selected?: Date | null
    onSelect?: (date: Date) => void
    className?: string
}

export function Calendar({ selected, onSelect, className }: CalendarProps) {
    const [currentMonth, setCurrentMonth] = React.useState(selected || new Date())

    const days = React.useMemo(() => {
        const start = startOfWeek(startOfMonth(currentMonth))
        const end = endOfWeek(endOfMonth(currentMonth))
        return eachDayOfInterval({ start, end })
    }, [currentMonth])

    const weekDays = ['Su', 'Mo', 'Tu', 'We', 'Th', 'Fr', 'Sa']

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
                <span className="text-sm font-medium text-white">
                    {format(currentMonth, 'MMMM yyyy')}
                </span>
                <Button
                    variant="ghost"
                    size="icon"
                    onClick={() => setCurrentMonth(addMonths(currentMonth, 1))}
                    className="h-7 w-7"
                >
                    <ChevronRight className="h-4 w-4" />
                </Button>
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
                    const isSelected = selected && isSameDay(day, selected)
                    const isCurrentMonth = isSameMonth(day, currentMonth)
                    const isToday = isSameDay(day, new Date())

                    return (
                        <button
                            key={idx}
                            onClick={() => onSelect?.(day)}
                            className={cn(
                                "h-8 w-8 rounded-md text-sm transition-colors",
                                "hover:bg-slate-700 focus:outline-none focus:ring-2 focus:ring-indigo-500",
                                !isCurrentMonth && "text-slate-600",
                                isCurrentMonth && "text-slate-200",
                                isToday && !isSelected && "border border-indigo-500/50",
                                isSelected && "bg-indigo-600 text-white hover:bg-indigo-500"
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
