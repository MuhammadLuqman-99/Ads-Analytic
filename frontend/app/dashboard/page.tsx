"use client"

import { RefreshCw, DollarSign, Eye, MousePointer, ShoppingCart, TrendingUp } from "lucide-react"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { SpendOverTimeChart } from "@/components/charts/SpendOverTimeChart"
import { ROASByPlatformChart } from "@/components/charts/ROASByPlatformChart"
import { useDashboardMetrics } from "@/hooks/use-metrics"
import { useSyncAllAccounts } from "@/hooks/use-accounts"
import { formatCurrency, formatNumber } from "@/lib/mock-data"

function MetricsCard({
    title,
    value,
    change,
    changeType,
    icon: Icon,
    iconColor
}: {
    title: string
    value: string
    change: number
    changeType: 'positive' | 'negative'
    icon: React.ComponentType<{ className?: string }>
    iconColor: string
}) {
    return (
        <Card className="bg-white border-slate-200">
            <CardContent className="p-6">
                <div className="flex items-center justify-between">
                    <div className={`p-3 rounded-xl bg-gradient-to-br ${iconColor} shadow-lg`}>
                        <Icon className="h-5 w-5 text-white" />
                    </div>
                    <div className={`flex items-center gap-1 text-sm font-medium ${changeType === 'positive' ? 'text-emerald-600' : 'text-rose-600'
                        }`}>
                        {changeType === 'positive' ? '↑' : '↓'}
                        {Math.abs(change)}%
                    </div>
                </div>
                <div className="mt-4">
                    <p className="text-sm font-medium text-slate-500">{title}</p>
                    <p className="mt-1 text-2xl font-bold text-slate-900">{value}</p>
                </div>
            </CardContent>
        </Card>
    )
}

export default function DashboardPage() {
    const { data: metrics, isLoading } = useDashboardMetrics()
    const syncMutation = useSyncAllAccounts()

    return (
        <div>
                    {/* Header */}
                    <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4 mb-8">
                        <div>
                            <h1 className="text-3xl font-bold text-slate-900">Dashboard</h1>
                            <p className="mt-1 text-slate-500">Overview of your ad performance across all platforms</p>
                        </div>
                        <div className="flex items-center gap-3">
                            <Button
                                onClick={() => syncMutation.mutate()}
                                disabled={syncMutation.isPending}
                            >
                                <RefreshCw className={`h-4 w-4 mr-2 ${syncMutation.isPending ? 'animate-spin' : ''}`} />
                                {syncMutation.isPending ? 'Syncing...' : 'Sync Now'}
                            </Button>
                        </div>
                    </div>

                    {/* Metrics Grid */}
                    {isLoading ? (
                        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-5 gap-4 mb-8">
                            {[...Array(5)].map((_, i) => (
                                <Card key={i} className="animate-pulse bg-white">
                                    <CardContent className="p-6">
                                        <div className="h-12 w-12 bg-slate-200 rounded-xl mb-4" />
                                        <div className="h-4 w-20 bg-slate-200 rounded mb-2" />
                                        <div className="h-6 w-32 bg-slate-200 rounded" />
                                    </CardContent>
                                </Card>
                            ))}
                        </div>
                    ) : (
                        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-5 gap-4 mb-8">
                            <MetricsCard
                                title="Total Spend"
                                value={formatCurrency(metrics?.totalSpend || 0)}
                                change={metrics?.spendChange || 0}
                                changeType="negative"
                                icon={DollarSign}
                                iconColor="from-indigo-500 to-indigo-600"
                            />
                            <MetricsCard
                                title="Impressions"
                                value={formatNumber(metrics?.totalImpressions || 0)}
                                change={metrics?.impressionsChange || 0}
                                changeType="positive"
                                icon={Eye}
                                iconColor="from-purple-500 to-purple-600"
                            />
                            <MetricsCard
                                title="Clicks"
                                value={formatNumber(metrics?.totalClicks || 0)}
                                change={metrics?.clicksChange || 0}
                                changeType="positive"
                                icon={MousePointer}
                                iconColor="from-cyan-500 to-cyan-600"
                            />
                            <MetricsCard
                                title="Conversions"
                                value={formatNumber(metrics?.totalConversions || 0)}
                                change={metrics?.conversionsChange || 0}
                                changeType="positive"
                                icon={ShoppingCart}
                                iconColor="from-emerald-500 to-emerald-600"
                            />
                            <MetricsCard
                                title="Avg. ROAS"
                                value={`${metrics?.averageRoas?.toFixed(1) || 0}x`}
                                change={8.5}
                                changeType="positive"
                                icon={TrendingUp}
                                iconColor="from-amber-500 to-amber-600"
                            />
                        </div>
                    )}

                    {/* Charts */}
                    <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-8">
                        <SpendOverTimeChart />
                        <ROASByPlatformChart />
                    </div>

                    {/* Platform Summary */}
                    <Card className="bg-white border-slate-200">
                        <CardHeader className="border-b border-slate-200">
                            <CardTitle className="text-slate-900">Platform Summary</CardTitle>
                        </CardHeader>
                        <CardContent className="p-0">
                            <div className="grid grid-cols-1 md:grid-cols-3 divide-y md:divide-y-0 md:divide-x divide-slate-200">
                                {/* Meta */}
                                <div className="p-6">
                                    <div className="flex items-center gap-3 mb-4">
                                        <div className="h-10 w-10 rounded-lg bg-blue-600 flex items-center justify-center text-white font-bold shadow-lg shadow-blue-600/25">
                                            M
                                        </div>
                                        <div>
                                            <p className="font-semibold text-slate-900">Meta Ads</p>
                                            <p className="text-xs text-slate-500">Facebook & Instagram</p>
                                        </div>
                                    </div>
                                    <div className="space-y-2">
                                        <div className="flex justify-between text-sm">
                                            <span className="text-slate-500">Spend</span>
                                            <span className="text-slate-900 font-medium">RM 12,450</span>
                                        </div>
                                        <div className="flex justify-between text-sm">
                                            <span className="text-slate-500">Conversions</span>
                                            <span className="text-slate-900 font-medium">892</span>
                                        </div>
                                        <div className="flex justify-between text-sm">
                                            <span className="text-slate-500">ROAS</span>
                                            <span className="text-emerald-600 font-medium">3.2x</span>
                                        </div>
                                    </div>
                                </div>

                                {/* TikTok */}
                                <div className="p-6">
                                    <div className="flex items-center gap-3 mb-4">
                                        <div className="h-10 w-10 rounded-lg bg-black flex items-center justify-center text-white font-bold">
                                            T
                                        </div>
                                        <div>
                                            <p className="font-semibold text-slate-900">TikTok Ads</p>
                                            <p className="text-xs text-slate-500">TikTok for Business</p>
                                        </div>
                                    </div>
                                    <div className="space-y-2">
                                        <div className="flex justify-between text-sm">
                                            <span className="text-slate-500">Spend</span>
                                            <span className="text-slate-900 font-medium">RM 8,234</span>
                                        </div>
                                        <div className="flex justify-between text-sm">
                                            <span className="text-slate-500">Conversions</span>
                                            <span className="text-slate-900 font-medium">645</span>
                                        </div>
                                        <div className="flex justify-between text-sm">
                                            <span className="text-slate-500">ROAS</span>
                                            <span className="text-emerald-600 font-medium">2.8x</span>
                                        </div>
                                    </div>
                                </div>

                                {/* Shopee */}
                                <div className="p-6">
                                    <div className="flex items-center gap-3 mb-4">
                                        <div className="h-10 w-10 rounded-lg bg-orange-500 flex items-center justify-center text-white font-bold shadow-lg shadow-orange-500/25">
                                            S
                                        </div>
                                        <div>
                                            <p className="font-semibold text-slate-900">Shopee Ads</p>
                                            <p className="text-xs text-slate-500">Shopee Marketing</p>
                                        </div>
                                    </div>
                                    <div className="space-y-2">
                                        <div className="flex justify-between text-sm">
                                            <span className="text-slate-500">Spend</span>
                                            <span className="text-slate-900 font-medium">RM 3,852</span>
                                        </div>
                                        <div className="flex justify-between text-sm">
                                            <span className="text-slate-500">Conversions</span>
                                            <span className="text-slate-900 font-medium">310</span>
                                        </div>
                                        <div className="flex justify-between text-sm">
                                            <span className="text-slate-500">ROAS</span>
                                            <span className="text-emerald-600 font-medium">4.1x</span>
                                        </div>
                                    </div>
                                </div>
                            </div>
                        </CardContent>
                    </Card>
        </div>
    )
}
