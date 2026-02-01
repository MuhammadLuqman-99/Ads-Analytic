// Types
export type Platform = 'meta' | 'tiktok' | 'shopee'
export type CampaignStatus = 'active' | 'paused' | 'completed' | 'draft'

export interface Campaign {
    id: string
    name: string
    platform: Platform
    status: CampaignStatus
    spend: number
    impressions: number
    clicks: number
    conversions: number
    roas: number
    startDate: string
    endDate: string
    accountId: string
}

export interface DailyMetric {
    date: string
    spend: number
    impressions: number
    clicks: number
    conversions: number
}

export interface PlatformMetrics {
    platform: Platform
    spend: number
    impressions: number
    clicks: number
    conversions: number
    roas: number
}

export interface DashboardMetrics {
    totalSpend: number
    totalImpressions: number
    totalClicks: number
    totalConversions: number
    averageRoas: number
    spendChange: number
    impressionsChange: number
    clicksChange: number
    conversionsChange: number
}

export interface ConnectedAccount {
    id: string
    platform: Platform
    accountId: string
    accountName: string
    status: 'connected' | 'disconnected' | 'error'
    lastSync: string
    currency: string
}

// Mock Data
export const mockCampaigns: Campaign[] = [
    {
        id: "1",
        name: "Summer Sale 2024",
        platform: "meta",
        status: "active",
        spend: 4520,
        impressions: 245000,
        clicks: 8432,
        conversions: 342,
        roas: 3.5,
        startDate: "2024-06-01",
        endDate: "2024-08-31",
        accountId: "meta-1"
    },
    {
        id: "2",
        name: "Product Launch - Series X",
        platform: "tiktok",
        status: "active",
        spend: 3210,
        impressions: 180000,
        clicks: 5892,
        conversions: 245,
        roas: 2.9,
        startDate: "2024-07-15",
        endDate: "2024-09-15",
        accountId: "tiktok-1"
    },
    {
        id: "3",
        name: "Flash Sale Weekend",
        platform: "shopee",
        status: "paused",
        spend: 1890,
        impressions: 85000,
        clicks: 3245,
        conversions: 156,
        roas: 4.2,
        startDate: "2024-07-20",
        endDate: "2024-07-22",
        accountId: "shopee-1"
    },
    {
        id: "4",
        name: "Raya Collection Promo",
        platform: "meta",
        status: "active",
        spend: 5200,
        impressions: 320000,
        clicks: 12450,
        conversions: 520,
        roas: 4.1,
        startDate: "2024-03-01",
        endDate: "2024-04-15",
        accountId: "meta-1"
    },
    {
        id: "5",
        name: "Back to School Campaign",
        platform: "tiktok",
        status: "completed",
        spend: 2850,
        impressions: 156000,
        clicks: 4320,
        conversions: 189,
        roas: 2.6,
        startDate: "2024-01-02",
        endDate: "2024-01-31",
        accountId: "tiktok-1"
    },
    {
        id: "6",
        name: "11.11 Mega Sale",
        platform: "shopee",
        status: "completed",
        spend: 8450,
        impressions: 520000,
        clicks: 24500,
        conversions: 1250,
        roas: 5.2,
        startDate: "2024-11-01",
        endDate: "2024-11-11",
        accountId: "shopee-1"
    },
    {
        id: "7",
        name: "Brand Awareness Q1",
        platform: "meta",
        status: "active",
        spend: 2730,
        impressions: 185000,
        clicks: 3890,
        conversions: 125,
        roas: 2.3,
        startDate: "2024-01-01",
        endDate: "2024-03-31",
        accountId: "meta-2"
    },
    {
        id: "8",
        name: "Influencer Collab - Fashion",
        platform: "tiktok",
        status: "active",
        spend: 4100,
        impressions: 420000,
        clicks: 15200,
        conversions: 680,
        roas: 3.8,
        startDate: "2024-06-15",
        endDate: "2024-08-15",
        accountId: "tiktok-1"
    },
    {
        id: "9",
        name: "Weekend Flash Deals",
        platform: "shopee",
        status: "draft",
        spend: 0,
        impressions: 0,
        clicks: 0,
        conversions: 0,
        roas: 0,
        startDate: "2024-08-01",
        endDate: "2024-08-03",
        accountId: "shopee-1"
    },
    {
        id: "10",
        name: "Holiday Special 2024",
        platform: "meta",
        status: "paused",
        spend: 3650,
        impressions: 198000,
        clicks: 6780,
        conversions: 298,
        roas: 3.1,
        startDate: "2024-12-15",
        endDate: "2024-12-31",
        accountId: "meta-1"
    }
]

export const mockDailyMetrics: DailyMetric[] = [
    { date: "2024-01-25", spend: 2450, impressions: 125000, clicks: 4200, conversions: 165 },
    { date: "2024-01-26", spend: 2680, impressions: 132000, clicks: 4580, conversions: 182 },
    { date: "2024-01-27", spend: 3120, impressions: 156000, clicks: 5340, conversions: 215 },
    { date: "2024-01-28", spend: 2890, impressions: 142000, clicks: 4890, conversions: 195 },
    { date: "2024-01-29", spend: 3450, impressions: 178000, clicks: 6120, conversions: 248 },
    { date: "2024-01-30", spend: 3680, impressions: 185000, clicks: 6450, conversions: 262 },
    { date: "2024-01-31", spend: 4120, impressions: 205000, clicks: 7230, conversions: 295 },
    { date: "2024-02-01", spend: 3950, impressions: 198000, clicks: 6890, conversions: 278 },
    { date: "2024-02-02", spend: 4280, impressions: 215000, clicks: 7520, conversions: 305 },
    { date: "2024-02-03", spend: 4560, impressions: 228000, clicks: 7980, conversions: 325 },
    { date: "2024-02-04", spend: 4120, impressions: 206000, clicks: 7150, conversions: 290 },
    { date: "2024-02-05", spend: 3890, impressions: 194000, clicks: 6780, conversions: 275 },
    { date: "2024-02-06", spend: 4350, impressions: 218000, clicks: 7620, conversions: 310 },
    { date: "2024-02-07", spend: 4680, impressions: 234000, clicks: 8190, conversions: 335 },
]

export const mockPlatformMetrics: PlatformMetrics[] = [
    {
        platform: "meta",
        spend: 12450,
        impressions: 650000,
        clicks: 23105,
        conversions: 892,
        roas: 3.2
    },
    {
        platform: "tiktok",
        spend: 8234,
        impressions: 420000,
        clicks: 15432,
        conversions: 645,
        roas: 2.8
    },
    {
        platform: "shopee",
        spend: 3852,
        impressions: 130000,
        clicks: 6694,
        conversions: 310,
        roas: 4.1
    }
]

export const mockDashboardMetrics: DashboardMetrics = {
    totalSpend: 24536,
    totalImpressions: 1200000,
    totalClicks: 45231,
    totalConversions: 1847,
    averageRoas: 3.4,
    spendChange: 12.5,
    impressionsChange: 8.2,
    clicksChange: 15.3,
    conversionsChange: 22.1
}

export const mockConnectedAccounts: ConnectedAccount[] = [
    {
        id: "acc-1",
        platform: "meta",
        accountId: "act_123456789",
        accountName: "My Business - Meta Ads",
        status: "connected",
        lastSync: "2024-02-01T10:30:00Z",
        currency: "MYR"
    },
    {
        id: "acc-2",
        platform: "tiktok",
        accountId: "tiktok_987654321",
        accountName: "TikTok Shop Ads",
        status: "connected",
        lastSync: "2024-02-01T09:45:00Z",
        currency: "MYR"
    },
    {
        id: "acc-3",
        platform: "shopee",
        accountId: "shopee_456789123",
        accountName: "Shopee Store Ads",
        status: "connected",
        lastSync: "2024-02-01T08:15:00Z",
        currency: "MYR"
    }
]

// Utility functions
export function formatCurrency(amount: number, currency: string = "MYR"): string {
    return new Intl.NumberFormat('ms-MY', {
        style: 'currency',
        currency: currency,
        minimumFractionDigits: 0,
        maximumFractionDigits: 0
    }).format(amount)
}

export function formatNumber(num: number): string {
    if (num >= 1000000) {
        return `${(num / 1000000).toFixed(1)}M`
    }
    if (num >= 1000) {
        return `${(num / 1000).toFixed(1)}K`
    }
    return num.toLocaleString()
}

export function getPlatformColor(platform: Platform): string {
    const colors = {
        meta: '#3b82f6',    // blue-500
        tiktok: '#64748b',  // slate-500
        shopee: '#f97316'   // orange-500
    }
    return colors[platform]
}

export function getPlatformName(platform: Platform): string {
    const names = {
        meta: 'Meta',
        tiktok: 'TikTok',
        shopee: 'Shopee'
    }
    return names[platform]
}
