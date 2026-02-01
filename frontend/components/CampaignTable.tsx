"use client"

import * as React from "react"
import {
    ArrowUpDown,
    ArrowUp,
    ArrowDown,
    MoreHorizontal,
    Eye,
    Pause,
    Play,
    Trash2
} from "lucide-react"
import {
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
} from "@/components/ui/table"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import {
    type Campaign,
    type Platform,
    formatCurrency,
    formatNumber,
    getPlatformName
} from "@/lib/mock-data"
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuLabel,
    DropdownMenuSeparator,
    DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"

type SortColumn = 'name' | 'platform' | 'status' | 'spend' | 'impressions' | 'clicks' | 'roas'
type SortDirection = 'asc' | 'desc'

interface CampaignTableProps {
    campaigns: Campaign[]
    isLoading?: boolean
}

const statusVariants: Record<Campaign['status'], 'success' | 'warning' | 'secondary' | 'default'> = {
    active: 'success',
    paused: 'warning',
    completed: 'secondary',
    draft: 'default',
}

const platformVariants: Record<Platform, 'meta' | 'tiktok' | 'shopee'> = {
    meta: 'meta',
    tiktok: 'tiktok',
    shopee: 'shopee',
}

export function CampaignTable({ campaigns, isLoading }: CampaignTableProps) {
    const [sortColumn, setSortColumn] = React.useState<SortColumn>('spend')
    const [sortDirection, setSortDirection] = React.useState<SortDirection>('desc')

    const handleSort = (column: SortColumn) => {
        if (sortColumn === column) {
            setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc')
        } else {
            setSortColumn(column)
            setSortDirection('desc')
        }
    }

    const sortedCampaigns = React.useMemo(() => {
        return [...campaigns].sort((a, b) => {
            let aValue: string | number
            let bValue: string | number

            switch (sortColumn) {
                case 'name':
                    aValue = a.name.toLowerCase()
                    bValue = b.name.toLowerCase()
                    break
                case 'platform':
                    aValue = a.platform
                    bValue = b.platform
                    break
                case 'status':
                    aValue = a.status
                    bValue = b.status
                    break
                case 'spend':
                    aValue = a.spend
                    bValue = b.spend
                    break
                case 'impressions':
                    aValue = a.impressions
                    bValue = b.impressions
                    break
                case 'clicks':
                    aValue = a.clicks
                    bValue = b.clicks
                    break
                case 'roas':
                    aValue = a.roas
                    bValue = b.roas
                    break
            }

            if (aValue < bValue) return sortDirection === 'asc' ? -1 : 1
            if (aValue > bValue) return sortDirection === 'asc' ? 1 : -1
            return 0
        })
    }, [campaigns, sortColumn, sortDirection])

    const SortableHeader = ({ column, children }: { column: SortColumn; children: React.ReactNode }) => {
        const isActive = sortColumn === column
        return (
            <Button
                variant="ghost"
                onClick={() => handleSort(column)}
                className="h-8 px-2 -ml-2 font-medium text-slate-400 hover:text-white"
            >
                {children}
                {isActive ? (
                    sortDirection === 'asc' ? (
                        <ArrowUp className="ml-2 h-4 w-4 text-indigo-400" />
                    ) : (
                        <ArrowDown className="ml-2 h-4 w-4 text-indigo-400" />
                    )
                ) : (
                    <ArrowUpDown className="ml-2 h-4 w-4 opacity-50" />
                )}
            </Button>
        )
    }

    if (isLoading) {
        return (
            <div className="rounded-xl border border-slate-700/50 bg-slate-800/50 overflow-hidden">
                <div className="flex items-center justify-center h-[400px]">
                    <div className="flex items-center gap-2 text-slate-400">
                        <svg className="animate-spin h-5 w-5" viewBox="0 0 24 24">
                            <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none" />
                            <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
                        </svg>
                        <span>Loading campaigns...</span>
                    </div>
                </div>
            </div>
        )
    }

    if (campaigns.length === 0) {
        return (
            <div className="rounded-xl border border-slate-700/50 bg-slate-800/50 overflow-hidden">
                <div className="flex flex-col items-center justify-center h-[400px] gap-2">
                    <div className="text-lg text-slate-400">No campaigns found</div>
                    <p className="text-sm text-slate-500">Try adjusting your filters</p>
                </div>
            </div>
        )
    }

    return (
        <div className="rounded-xl border border-slate-700/50 bg-slate-800/50 overflow-hidden">
            <Table>
                <TableHeader>
                    <TableRow className="hover:bg-transparent border-slate-700/50">
                        <TableHead className="w-[300px]">
                            <SortableHeader column="name">Campaign</SortableHeader>
                        </TableHead>
                        <TableHead>
                            <SortableHeader column="platform">Platform</SortableHeader>
                        </TableHead>
                        <TableHead>
                            <SortableHeader column="status">Status</SortableHeader>
                        </TableHead>
                        <TableHead className="text-right">
                            <SortableHeader column="spend">Spend</SortableHeader>
                        </TableHead>
                        <TableHead className="text-right">
                            <SortableHeader column="impressions">Impressions</SortableHeader>
                        </TableHead>
                        <TableHead className="text-right">
                            <SortableHeader column="clicks">Clicks</SortableHeader>
                        </TableHead>
                        <TableHead className="text-right">
                            <SortableHeader column="roas">ROAS</SortableHeader>
                        </TableHead>
                        <TableHead className="w-[50px]"></TableHead>
                    </TableRow>
                </TableHeader>
                <TableBody>
                    {sortedCampaigns.map((campaign) => (
                        <TableRow key={campaign.id} className="border-slate-700/50">
                            <TableCell>
                                <div>
                                    <div className="font-medium text-white">{campaign.name}</div>
                                    <div className="text-xs text-slate-500">ID: {campaign.id}</div>
                                </div>
                            </TableCell>
                            <TableCell>
                                <Badge variant={platformVariants[campaign.platform]}>
                                    {getPlatformName(campaign.platform)}
                                </Badge>
                            </TableCell>
                            <TableCell>
                                <Badge variant={statusVariants[campaign.status]} className="capitalize">
                                    {campaign.status}
                                </Badge>
                            </TableCell>
                            <TableCell className="text-right font-medium">
                                {formatCurrency(campaign.spend)}
                            </TableCell>
                            <TableCell className="text-right text-slate-300">
                                {formatNumber(campaign.impressions)}
                            </TableCell>
                            <TableCell className="text-right text-slate-300">
                                {formatNumber(campaign.clicks)}
                            </TableCell>
                            <TableCell className="text-right">
                                <span className={campaign.roas >= 3 ? 'text-emerald-400 font-medium' : campaign.roas >= 2 ? 'text-amber-400' : 'text-slate-400'}>
                                    {campaign.roas > 0 ? `${campaign.roas.toFixed(1)}x` : '-'}
                                </span>
                            </TableCell>
                            <TableCell>
                                <DropdownMenu>
                                    <DropdownMenuTrigger asChild>
                                        <Button variant="ghost" size="icon" className="h-8 w-8">
                                            <MoreHorizontal className="h-4 w-4" />
                                            <span className="sr-only">Open menu</span>
                                        </Button>
                                    </DropdownMenuTrigger>
                                    <DropdownMenuContent align="end" className="w-48">
                                        <DropdownMenuLabel>Actions</DropdownMenuLabel>
                                        <DropdownMenuSeparator />
                                        <DropdownMenuItem className="gap-2 cursor-pointer">
                                            <Eye className="h-4 w-4" />
                                            View Details
                                        </DropdownMenuItem>
                                        {campaign.status === 'active' ? (
                                            <DropdownMenuItem className="gap-2 cursor-pointer">
                                                <Pause className="h-4 w-4" />
                                                Pause Campaign
                                            </DropdownMenuItem>
                                        ) : campaign.status === 'paused' ? (
                                            <DropdownMenuItem className="gap-2 cursor-pointer">
                                                <Play className="h-4 w-4" />
                                                Resume Campaign
                                            </DropdownMenuItem>
                                        ) : null}
                                        <DropdownMenuSeparator />
                                        <DropdownMenuItem className="gap-2 cursor-pointer text-red-400 focus:text-red-400">
                                            <Trash2 className="h-4 w-4" />
                                            Delete
                                        </DropdownMenuItem>
                                    </DropdownMenuContent>
                                </DropdownMenu>
                            </TableCell>
                        </TableRow>
                    ))}
                </TableBody>
            </Table>
        </div>
    )
}
