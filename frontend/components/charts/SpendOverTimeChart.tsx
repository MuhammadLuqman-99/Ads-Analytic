"use client"

import {
    LineChart,
    Line,
    XAxis,
    YAxis,
    CartesianGrid,
    Tooltip,
    ResponsiveContainer,
    Area,
    AreaChart
} from "recharts"
import { format, parseISO } from "date-fns"
import { useDailyMetrics } from "@/hooks/use-metrics"
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card"
import { formatCurrency } from "@/lib/mock-data"

interface SpendOverTimeChartProps {
    dateRange?: { from: Date; to: Date }
}

export function SpendOverTimeChart({ dateRange }: SpendOverTimeChartProps) {
    const { data: metrics, isLoading } = useDailyMetrics(dateRange)

    if (isLoading) {
        return (
            <Card className="h-[400px]">
                <CardHeader>
                    <CardTitle>Spend Over Time</CardTitle>
                    <CardDescription>Daily ad spend across all platforms</CardDescription>
                </CardHeader>
                <CardContent className="flex items-center justify-center h-[300px]">
                    <div className="flex items-center gap-2 text-slate-400">
                        <svg className="animate-spin h-5 w-5" viewBox="0 0 24 24">
                            <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none" />
                            <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
                        </svg>
                        <span>Loading chart data...</span>
                    </div>
                </CardContent>
            </Card>
        )
    }

    const chartData = metrics?.map(m => ({
        ...m,
        formattedDate: format(parseISO(m.date), 'MMM dd'),
    })) || []

    return (
        <Card className="h-[400px]">
            <CardHeader>
                <CardTitle>Spend Over Time</CardTitle>
                <CardDescription>Daily ad spend across all platforms</CardDescription>
            </CardHeader>
            <CardContent>
                <ResponsiveContainer width="100%" height={280}>
                    <AreaChart data={chartData} margin={{ top: 10, right: 30, left: 0, bottom: 0 }}>
                        <defs>
                            <linearGradient id="spendGradient" x1="0" y1="0" x2="0" y2="1">
                                <stop offset="5%" stopColor="#6366f1" stopOpacity={0.4} />
                                <stop offset="95%" stopColor="#6366f1" stopOpacity={0} />
                            </linearGradient>
                        </defs>
                        <CartesianGrid strokeDasharray="3 3" stroke="#334155" vertical={false} />
                        <XAxis
                            dataKey="formattedDate"
                            stroke="#64748b"
                            fontSize={12}
                            tickLine={false}
                            axisLine={false}
                        />
                        <YAxis
                            stroke="#64748b"
                            fontSize={12}
                            tickLine={false}
                            axisLine={false}
                            tickFormatter={(value) => `${(value / 1000).toFixed(1)}K`}
                        />
                        <Tooltip
                            contentStyle={{
                                backgroundColor: '#1e293b',
                                border: '1px solid #334155',
                                borderRadius: '8px',
                                color: '#f1f5f9'
                            }}
                            formatter={(value) => [formatCurrency(Number(value)), 'Spend']}
                            labelStyle={{ color: '#94a3b8' }}
                        />
                        <Area
                            type="monotone"
                            dataKey="spend"
                            stroke="#6366f1"
                            strokeWidth={2}
                            fillOpacity={1}
                            fill="url(#spendGradient)"
                        />
                    </AreaChart>
                </ResponsiveContainer>
            </CardContent>
        </Card>
    )
}
