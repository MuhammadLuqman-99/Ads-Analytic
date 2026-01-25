interface MetricCardProps {
    title: string
    value: string
    change?: string
    changeType?: 'positive' | 'negative' | 'neutral'
    icon: React.ReactNode
}

export default function MetricCard({ title, value, change, changeType = 'neutral', icon }: MetricCardProps) {
    const changeColors = {
        positive: 'text-emerald-400',
        negative: 'text-rose-400',
        neutral: 'text-slate-400',
    }

    return (
        <div className="rounded-xl bg-slate-800/50 border border-slate-700/50 p-6 hover:border-slate-600/50 transition-all duration-200">
            <div className="flex items-center justify-between">
                <div className="flex-shrink-0 p-3 rounded-lg bg-slate-700/50">
                    {icon}
                </div>
                {change && (
                    <span className={`text-sm font-medium ${changeColors[changeType]}`}>
                        {changeType === 'positive' && '↑'}
                        {changeType === 'negative' && '↓'}
                        {change}
                    </span>
                )}
            </div>
            <div className="mt-4">
                <h3 className="text-sm font-medium text-slate-400">{title}</h3>
                <p className="mt-1 text-2xl font-bold text-white">{value}</p>
            </div>
        </div>
    )
}
