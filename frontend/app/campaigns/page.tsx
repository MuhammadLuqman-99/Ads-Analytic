"use client"

import * as React from "react"
import { Search } from "lucide-react"
import Sidebar from "@/components/Sidebar"
import { CampaignTable } from "@/components/CampaignTable"
import { PlatformFilter } from "@/components/PlatformFilter"
import { StatusFilter } from "@/components/StatusFilter"
import { DateRangePicker, type DateRange } from "@/components/DateRangePicker"
import { Input } from "@/components/ui/input"
import { useCampaigns } from "@/hooks/use-metrics"
import { type Platform } from "@/lib/mock-data"
import { subDays } from "date-fns"

export default function CampaignsPage() {
    const [selectedPlatforms, setSelectedPlatforms] = React.useState<Platform[]>([])
    const [selectedStatus, setSelectedStatus] = React.useState<string>("all")
    const [searchQuery, setSearchQuery] = React.useState("")
    const [dateRange, setDateRange] = React.useState<DateRange>({
        from: subDays(new Date(), 30),
        to: new Date(),
    })

    const { data: campaigns, isLoading } = useCampaigns({
        platforms: selectedPlatforms.length > 0 ? selectedPlatforms : undefined,
        status: selectedStatus !== 'all' ? selectedStatus : undefined,
        search: searchQuery || undefined,
    })

    const handleSearchChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        setSearchQuery(e.target.value)
    }

    return (
        <div className="min-h-screen bg-slate-900">
            <Sidebar />

            <main className="lg:pl-72">
                <div className="px-4 sm:px-6 lg:px-8 py-8">
                    {/* Header */}
                    <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4 mb-8">
                        <div>
                            <h1 className="text-3xl font-bold text-white">Campaigns</h1>
                            <p className="mt-1 text-slate-400">Manage all your ad campaigns across platforms</p>
                        </div>
                        <div className="text-sm text-slate-400">
                            {campaigns?.length || 0} campaigns found
                        </div>
                    </div>

                    {/* Filters */}
                    <div className="flex flex-col lg:flex-row gap-4 mb-6">
                        {/* Search */}
                        <div className="relative flex-1 max-w-md">
                            <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-slate-500" />
                            <Input
                                placeholder="Search campaigns..."
                                value={searchQuery}
                                onChange={handleSearchChange}
                                className="pl-10"
                            />
                        </div>

                        {/* Filter Controls */}
                        <div className="flex flex-wrap gap-3">
                            <PlatformFilter
                                selected={selectedPlatforms}
                                onSelectionChange={setSelectedPlatforms}
                            />
                            <StatusFilter
                                value={selectedStatus}
                                onChange={setSelectedStatus}
                            />
                            <DateRangePicker
                                dateRange={dateRange}
                                onDateRangeChange={setDateRange}
                            />
                        </div>
                    </div>

                    {/* Active Filters */}
                    {(selectedPlatforms.length > 0 || selectedStatus !== 'all' || searchQuery) && (
                        <div className="flex flex-wrap items-center gap-2 mb-4">
                            <span className="text-sm text-slate-400">Active filters:</span>
                            {selectedPlatforms.map((platform) => (
                                <button
                                    key={platform}
                                    onClick={() => setSelectedPlatforms(selectedPlatforms.filter(p => p !== platform))}
                                    className="inline-flex items-center gap-1 px-2 py-1 rounded-md bg-slate-800 text-sm text-slate-300 hover:bg-slate-700 transition-colors"
                                >
                                    {platform}
                                    <span className="text-slate-500">×</span>
                                </button>
                            ))}
                            {selectedStatus !== 'all' && (
                                <button
                                    onClick={() => setSelectedStatus('all')}
                                    className="inline-flex items-center gap-1 px-2 py-1 rounded-md bg-slate-800 text-sm text-slate-300 hover:bg-slate-700 transition-colors"
                                >
                                    {selectedStatus}
                                    <span className="text-slate-500">×</span>
                                </button>
                            )}
                            {searchQuery && (
                                <button
                                    onClick={() => setSearchQuery('')}
                                    className="inline-flex items-center gap-1 px-2 py-1 rounded-md bg-slate-800 text-sm text-slate-300 hover:bg-slate-700 transition-colors"
                                >
                                    &quot;{searchQuery}&quot;
                                    <span className="text-slate-500">×</span>
                                </button>
                            )}
                            <button
                                onClick={() => {
                                    setSelectedPlatforms([])
                                    setSelectedStatus('all')
                                    setSearchQuery('')
                                }}
                                className="text-sm text-indigo-400 hover:text-indigo-300 transition-colors"
                            >
                                Clear all
                            </button>
                        </div>
                    )}

                    {/* Table */}
                    <CampaignTable campaigns={campaigns || []} isLoading={isLoading} />
                </div>
            </main>
        </div>
    )
}
