'use client'

interface PlatformCardProps {
    name: string
    platform: 'meta' | 'tiktok' | 'shopee'
    spend: string
    impressions: string
    clicks: string
    conversions: string
    roas: string
    connected: boolean
}

const platformConfig = {
    meta: {
        color: 'from-blue-500 to-blue-600',
        bg: 'bg-blue-500',
        label: 'Meta Ads',
        icon: 'M'
    },
    tiktok: {
        color: 'from-slate-800 to-slate-900',
        bg: 'bg-slate-800',
        label: 'TikTok Ads',
        icon: 'T'
    },
    shopee: {
        color: 'from-orange-500 to-orange-600',
        bg: 'bg-orange-500',
        label: 'Shopee Ads',
        icon: 'S'
    }
}

export default function PlatformCard({ name, platform, spend, impressions, clicks, conversions, roas, connected }: PlatformCardProps) {
    const config = platformConfig[platform]

    return (
        <div className="rounded-xl bg-slate-800/50 border border-slate-700/50 overflow-hidden hover:border-slate-600/50 transition-all duration-200">
            <div className={`bg-gradient-to-r ${config.color} p-4`}>
                <div className="flex items-center justify-between">
                    <div className="flex items-center gap-3">
                        <div className="h-10 w-10 rounded-full bg-white/20 flex items-center justify-center text-white font-bold">
                            {config.icon}
                        </div>
                        <div>
                            <h3 className="font-semibold text-white">{config.label}</h3>
                            <p className="text-sm text-white/70">{name}</p>
                        </div>
                    </div>
                    <span className={`px-2 py-1 rounded-full text-xs font-medium ${connected ? 'bg-emerald-500/20 text-emerald-300' : 'bg-slate-500/20 text-slate-300'}`}>
                        {connected ? 'Connected' : 'Disconnected'}
                    </span>
                </div>
            </div>
            <div className="p-4 grid grid-cols-2 gap-4">
                <div>
                    <p className="text-xs text-slate-400">Spend</p>
                    <p className="text-lg font-semibold text-white">{spend}</p>
                </div>
                <div>
                    <p className="text-xs text-slate-400">ROAS</p>
                    <p className="text-lg font-semibold text-emerald-400">{roas}x</p>
                </div>
                <div>
                    <p className="text-xs text-slate-400">Impressions</p>
                    <p className="text-lg font-semibold text-white">{impressions}</p>
                </div>
                <div>
                    <p className="text-xs text-slate-400">Clicks</p>
                    <p className="text-lg font-semibold text-white">{clicks}</p>
                </div>
            </div>
        </div>
    )
}
