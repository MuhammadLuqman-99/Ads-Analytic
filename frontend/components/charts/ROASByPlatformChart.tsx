"use client"

import {
    BarChart,
    Bar,
    XAxis,
    YAxis,
    CartesianGrid,
    Tooltip,
    ResponsiveContainer,
    Cell,
    Legend
} from "recharts"
import { usePlatformMetrics } from "@/hooks/use-metrics"
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card"
import { getPlatformColor, getPlatformName } from "@/lib/mock-data"

export function ROASByPlatformChart() {
    const { data: metrics, isLoading } = usePlatformMetrics()

    if (isLoading) {
        return (
            <Card className="h-[400px]">
                <CardHeader>
                    <CardTitle>ROAS by Platform</CardTitle>
                    <CardDescription>Return on ad spend comparison</CardDescription>
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
        name: getPlatformName(m.platform),
        platform: m.platform,
        roas: m.roas,
        spend: m.spend,
        conversions: m.conversions,
    })) || []

    return (
        <Card className="h-[400px]">
            <CardHeader>
                <CardTitle>ROAS by Platform</CardTitle>
                <CardDescription>Return on ad spend comparison</CardDescription>
            </CardHeader>
            <CardContent>
                <ResponsiveContainer width="100%" height={280}>
                    <BarChart data={chartData} margin={{ top: 20, right: 30, left: 0, bottom: 5 }}>
                        <CartesianGrid strokeDasharray="3 3" stroke="#334155" vertical={false} />
                        <XAxis
                            dataKey="name"
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
                            tickFormatter={(value) => `${value}x`}
                            domain={[0, 'dataMax + 1']}
                        />
                        <Tooltip
                            contentStyle={{
                                backgroundColor: '#1e293b',
                                border: '1px solid #334155',
                                borderRadius: '8px',
                                color: '#f1f5f9'
                            }}
                            formatter={(value, name) => {
                                if (name === 'roas') return [`${Number(value).toFixed(1)}x`, 'ROAS']
                                return [value, name]
                            }}
                            labelStyle={{ color: '#94a3b8' }}
                        />
                        <Bar
                            dataKey="roas"
                            radius={[6, 6, 0, 0]}
                            maxBarSize={80}
                        >
                            {chartData.map((entry, index) => (
                                <Cell
                                    key={`cell-${index}`}
                                    fill={getPlatformColor(entry.platform)}
                                    className="transition-opacity hover:opacity-80"
                                />
                            ))}
                        </Bar>
                    </BarChart>
                </ResponsiveContainer>

                {/* Custom Legend */}
                <div className="flex justify-center gap-6 mt-2">
                    {chartData.map((entry) => (
                        <div key={entry.platform} className="flex items-center gap-2">
                            <div
                                className="w-3 h-3 rounded-sm"
                                style={{ backgroundColor: getPlatformColor(entry.platform) }}
                            />
                            <span className="text-sm text-slate-400">{entry.name}</span>
                        </div>
                    ))}
                </div>
            </CardContent>
        </Card>
    )
}
