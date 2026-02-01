"use client"

import { ExternalLink, Loader2 } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import { type Platform, getPlatformName } from "@/lib/mock-data"
import { useConnectAccount } from "@/hooks/use-accounts"

interface ConnectPlatformButtonProps {
    platform: Platform
    isConnected: boolean
}

const platformConfig: Record<Platform, {
    icon: React.ReactNode
    color: string
    hoverColor: string
    description: string
}> = {
    meta: {
        icon: (
            <svg className="w-6 h-6" viewBox="0 0 24 24" fill="currentColor">
                <path d="M12 2.04c-5.5 0-10 4.49-10 10.02 0 5 3.66 9.15 8.44 9.9v-7H7.9v-2.9h2.54V9.85c0-2.52 1.49-3.91 3.78-3.91 1.1 0 2.24.2 2.24.2v2.47h-1.26c-1.24 0-1.63.78-1.63 1.57v1.88h2.78l-.45 2.9h-2.33v7a10 10 0 008.44-9.9c0-5.53-4.5-10.02-10-10.02z" />
            </svg>
        ),
        color: 'bg-blue-600 border-blue-600',
        hoverColor: 'hover:bg-blue-500 hover:border-blue-500',
        description: 'Connect your Meta Business account to sync Facebook and Instagram ads data.',
    },
    tiktok: {
        icon: (
            <svg className="w-6 h-6" viewBox="0 0 24 24" fill="currentColor">
                <path d="M19.59 6.69a4.83 4.83 0 01-3.77-4.25V2h-3.45v13.67a2.89 2.89 0 01-5.2 1.74 2.89 2.89 0 012.31-4.64 2.93 2.93 0 01.88.13V9.4a6.84 6.84 0 00-1-.05A6.33 6.33 0 005 20.1a6.34 6.34 0 0010.86-4.43v-7a8.16 8.16 0 004.77 1.52v-3.4a4.85 4.85 0 01-1-.1z" />
            </svg>
        ),
        color: 'bg-slate-900 border-slate-700',
        hoverColor: 'hover:bg-slate-800 hover:border-slate-600',
        description: 'Connect your TikTok for Business account to sync TikTok ads data.',
    },
    shopee: {
        icon: (
            <svg className="w-6 h-6" viewBox="0 0 24 24" fill="currentColor">
                <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-2 15l-5-5 1.41-1.41L10 14.17l7.59-7.59L19 8l-9 9z" />
            </svg>
        ),
        color: 'bg-orange-600 border-orange-600',
        hoverColor: 'hover:bg-orange-500 hover:border-orange-500',
        description: 'Connect your Shopee Seller account to sync Shopee ads data.',
    },
}

export function ConnectPlatformButton({ platform, isConnected }: ConnectPlatformButtonProps) {
    const connectMutation = useConnectAccount()
    const config = platformConfig[platform]

    const handleConnect = () => {
        connectMutation.mutate(platform)
    }

    if (isConnected) {
        return (
            <Card className="opacity-60">
                <CardContent className="p-6">
                    <div className="flex items-center gap-4">
                        <div className={`p-3 rounded-xl ${config.color} text-white`}>
                            {config.icon}
                        </div>
                        <div className="flex-1">
                            <h3 className="font-medium text-white">{getPlatformName(platform)}</h3>
                            <p className="text-sm text-slate-400">Already connected</p>
                        </div>
                        <Button variant="ghost" disabled size="sm">
                            Connected
                        </Button>
                    </div>
                </CardContent>
            </Card>
        )
    }

    return (
        <Card className="group transition-all duration-200 hover:border-slate-600">
            <CardContent className="p-6">
                <div className="flex items-start gap-4">
                    <div className={`p-3 rounded-xl ${config.color} ${config.hoverColor} text-white transition-colors`}>
                        {config.icon}
                    </div>
                    <div className="flex-1">
                        <h3 className="font-medium text-white">{getPlatformName(platform)}</h3>
                        <p className="text-sm text-slate-400 mt-1">{config.description}</p>
                    </div>
                </div>
                <div className="mt-4">
                    <Button
                        onClick={handleConnect}
                        disabled={connectMutation.isPending}
                        className={`w-full ${config.color} ${config.hoverColor} text-white border transition-colors`}
                    >
                        {connectMutation.isPending ? (
                            <>
                                <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                                Connecting...
                            </>
                        ) : (
                            <>
                                <ExternalLink className="w-4 h-4 mr-2" />
                                Connect {getPlatformName(platform)}
                            </>
                        )}
                    </Button>
                </div>
            </CardContent>
        </Card>
    )
}
