"use client"

import { useQuery } from "@tanstack/react-query"
import {
    mockCampaigns,
    mockDailyMetrics,
    mockPlatformMetrics,
    mockDashboardMetrics,
    type Campaign,
    type DailyMetric,
    type PlatformMetrics,
    type DashboardMetrics,
    type Platform
} from "@/lib/mock-data"

// Simulate API delay
const delay = (ms: number) => new Promise(resolve => setTimeout(resolve, ms))

// Dashboard Metrics Hook
export function useDashboardMetrics() {
    return useQuery<DashboardMetrics>({
        queryKey: ['dashboard-metrics'],
        queryFn: async () => {
            await delay(500)
            return mockDashboardMetrics
        },
    })
}

// Daily Metrics Hook (for charts)
export function useDailyMetrics(dateRange?: { from: Date; to: Date }) {
    return useQuery<DailyMetric[]>({
        queryKey: ['daily-metrics', dateRange],
        queryFn: async () => {
            await delay(400)
            if (dateRange) {
                return mockDailyMetrics.filter(m => {
                    const date = new Date(m.date)
                    return date >= dateRange.from && date <= dateRange.to
                })
            }
            return mockDailyMetrics
        },
    })
}

// Platform Metrics Hook (for bar chart)
export function usePlatformMetrics() {
    return useQuery<PlatformMetrics[]>({
        queryKey: ['platform-metrics'],
        queryFn: async () => {
            await delay(300)
            return mockPlatformMetrics
        },
    })
}

// Campaigns Hook with filtering
interface CampaignsFilter {
    platforms?: Platform[]
    status?: string
    search?: string
    dateRange?: { from: Date; to: Date }
}

export function useCampaigns(filters?: CampaignsFilter) {
    return useQuery<Campaign[]>({
        queryKey: ['campaigns', filters],
        queryFn: async () => {
            await delay(600)

            let result = [...mockCampaigns]

            if (filters?.platforms && filters.platforms.length > 0) {
                result = result.filter(c => filters.platforms!.includes(c.platform))
            }

            if (filters?.status && filters.status !== 'all') {
                result = result.filter(c => c.status === filters.status)
            }

            if (filters?.search) {
                const searchLower = filters.search.toLowerCase()
                result = result.filter(c =>
                    c.name.toLowerCase().includes(searchLower) ||
                    c.id.includes(searchLower)
                )
            }

            if (filters?.dateRange) {
                result = result.filter(c => {
                    const startDate = new Date(c.startDate)
                    return startDate >= filters.dateRange!.from && startDate <= filters.dateRange!.to
                })
            }

            return result
        },
    })
}

// Single Campaign Hook
export function useCampaign(id: string) {
    return useQuery<Campaign | undefined>({
        queryKey: ['campaign', id],
        queryFn: async () => {
            await delay(300)
            return mockCampaigns.find(c => c.id === id)
        },
        enabled: !!id,
    })
}
