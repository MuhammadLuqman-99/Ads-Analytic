"use client"

import * as React from "react"
import { Check, Filter } from "lucide-react"
import { cn } from "@/lib/utils"
import { Button } from "@/components/ui/button"
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover"
import { Badge } from "@/components/ui/badge"
import { type Platform, getPlatformName } from "@/lib/mock-data"

interface PlatformFilterProps {
    selected: Platform[]
    onSelectionChange: (platforms: Platform[]) => void
}

const platforms: { value: Platform; label: string; color: string }[] = [
    { value: "meta", label: "Meta", color: "bg-blue-500" },
    { value: "tiktok", label: "TikTok", color: "bg-slate-500" },
    { value: "shopee", label: "Shopee", color: "bg-orange-500" },
]

export function PlatformFilter({ selected, onSelectionChange }: PlatformFilterProps) {
    const [open, setOpen] = React.useState(false)

    const togglePlatform = (platform: Platform) => {
        if (selected.includes(platform)) {
            onSelectionChange(selected.filter(p => p !== platform))
        } else {
            onSelectionChange([...selected, platform])
        }
    }

    const clearAll = () => {
        onSelectionChange([])
    }

    const selectAll = () => {
        onSelectionChange(platforms.map(p => p.value))
    }

    return (
        <Popover open={open} onOpenChange={setOpen}>
            <PopoverTrigger asChild>
                <Button variant="outline" className="h-10 gap-2 min-w-[140px]">
                    <Filter className="h-4 w-4" />
                    <span>Platforms</span>
                    {selected.length > 0 && (
                        <Badge variant="secondary" className="ml-1 rounded-full px-1.5 py-0.5 text-xs">
                            {selected.length}
                        </Badge>
                    )}
                </Button>
            </PopoverTrigger>
            <PopoverContent className="w-56 p-0" align="start">
                <div className="p-3 border-b border-slate-700">
                    <div className="flex items-center justify-between">
                        <span className="text-sm font-medium text-white">Filter by Platform</span>
                        <div className="flex gap-1">
                            <button
                                onClick={selectAll}
                                className="text-xs text-indigo-400 hover:text-indigo-300 transition-colors"
                            >
                                All
                            </button>
                            <span className="text-slate-600">|</span>
                            <button
                                onClick={clearAll}
                                className="text-xs text-slate-400 hover:text-slate-300 transition-colors"
                            >
                                Clear
                            </button>
                        </div>
                    </div>
                </div>
                <div className="p-2">
                    {platforms.map((platform) => {
                        const isSelected = selected.includes(platform.value)
                        return (
                            <button
                                key={platform.value}
                                onClick={() => togglePlatform(platform.value)}
                                className={cn(
                                    "flex items-center gap-3 w-full px-3 py-2 rounded-md text-sm transition-colors",
                                    isSelected
                                        ? "bg-slate-700 text-white"
                                        : "text-slate-300 hover:bg-slate-700/50"
                                )}
                            >
                                <div className={cn("w-2.5 h-2.5 rounded-full", platform.color)} />
                                <span className="flex-1 text-left">{platform.label}</span>
                                {isSelected && <Check className="h-4 w-4 text-indigo-400" />}
                            </button>
                        )
                    })}
                </div>
            </PopoverContent>
        </Popover>
    )
}
