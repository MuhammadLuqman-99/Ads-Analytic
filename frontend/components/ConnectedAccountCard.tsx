"use client"

import {
    RefreshCw,
    Unlink,
    CheckCircle2,
    AlertCircle,
    Clock
} from "lucide-react"
import { format, formatDistanceToNow, parseISO } from "date-fns"
import { Card, CardContent } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import {
    type ConnectedAccount,
    type Platform,
    getPlatformName
} from "@/lib/mock-data"
import { useSyncAccount, useDisconnectAccount } from "@/hooks/use-accounts"

interface ConnectedAccountCardProps {
    account: ConnectedAccount
}

const platformIcons: Record<Platform, React.ReactNode> = {
    meta: (
        <svg className="w-8 h-8" viewBox="0 0 24 24" fill="currentColor">
            <path d="M12 2.04c-5.5 0-10 4.49-10 10.02 0 5 3.66 9.15 8.44 9.9v-7H7.9v-2.9h2.54V9.85c0-2.52 1.49-3.91 3.78-3.91 1.1 0 2.24.2 2.24.2v2.47h-1.26c-1.24 0-1.63.78-1.63 1.57v1.88h2.78l-.45 2.9h-2.33v7a10 10 0 008.44-9.9c0-5.53-4.5-10.02-10-10.02z" />
        </svg>
    ),
    tiktok: (
        <svg className="w-8 h-8" viewBox="0 0 24 24" fill="currentColor">
            <path d="M19.59 6.69a4.83 4.83 0 01-3.77-4.25V2h-3.45v13.67a2.89 2.89 0 01-5.2 1.74 2.89 2.89 0 012.31-4.64 2.93 2.93 0 01.88.13V9.4a6.84 6.84 0 00-1-.05A6.33 6.33 0 005 20.1a6.34 6.34 0 0010.86-4.43v-7a8.16 8.16 0 004.77 1.52v-3.4a4.85 4.85 0 01-1-.1z" />
        </svg>
    ),
    shopee: (
        <svg className="w-8 h-8" viewBox="0 0 24 24" fill="currentColor">
            <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-2 15l-5-5 1.41-1.41L10 14.17l7.59-7.59L19 8l-9 9z" />
        </svg>
    ),
}

const platformColors: Record<Platform, string> = {
    meta: 'bg-blue-600',
    tiktok: 'bg-black',
    shopee: 'bg-orange-500',
}

const statusConfig = {
    connected: {
        icon: CheckCircle2,
        color: 'text-emerald-400',
        badge: 'success' as const,
        label: 'Connected',
    },
    disconnected: {
        icon: AlertCircle,
        color: 'text-slate-400',
        badge: 'secondary' as const,
        label: 'Disconnected',
    },
    error: {
        icon: AlertCircle,
        color: 'text-red-400',
        badge: 'destructive' as const,
        label: 'Error',
    },
}

export function ConnectedAccountCard({ account }: ConnectedAccountCardProps) {
    const syncMutation = useSyncAccount()
    const disconnectMutation = useDisconnectAccount()

    const status = statusConfig[account.status]
    const StatusIcon = status.icon

    const handleSync = () => {
        syncMutation.mutate(account.id)
    }

    const handleDisconnect = () => {
        if (window.confirm(`Are you sure you want to disconnect ${account.accountName}?`)) {
            disconnectMutation.mutate(account.id)
        }
    }

    const lastSyncDate = parseISO(account.lastSync)
    const lastSyncFormatted = formatDistanceToNow(lastSyncDate, { addSuffix: true })

    return (
        <Card className="overflow-hidden">
            <div className={`h-2 ${platformColors[account.platform]}`} />
            <CardContent className="p-6">
                <div className="flex items-start justify-between">
                    <div className="flex items-center gap-4">
                        <div className={`p-3 rounded-xl ${platformColors[account.platform]} text-white`}>
                            {platformIcons[account.platform]}
                        </div>
                        <div>
                            <h3 className="font-semibold text-white text-lg">{account.accountName}</h3>
                            <p className="text-sm text-slate-400">{getPlatformName(account.platform)} Ads</p>
                            <p className="text-xs text-slate-500 mt-1">Account ID: {account.accountId}</p>
                        </div>
                    </div>
                    <Badge variant={status.badge}>
                        <StatusIcon className="w-3 h-3 mr-1" />
                        {status.label}
                    </Badge>
                </div>

                <div className="mt-6 pt-4 border-t border-slate-700/50">
                    <div className="flex items-center justify-between mb-4">
                        <div className="flex items-center gap-2 text-sm text-slate-400">
                            <Clock className="w-4 h-4" />
                            <span>Last synced: {lastSyncFormatted}</span>
                        </div>
                        <span className="text-xs text-slate-500">
                            {format(lastSyncDate, 'MMM dd, yyyy HH:mm')}
                        </span>
                    </div>

                    <div className="flex gap-2">
                        <Button
                            variant="secondary"
                            size="sm"
                            onClick={handleSync}
                            disabled={syncMutation.isPending}
                            className="flex-1"
                        >
                            <RefreshCw className={`w-4 h-4 mr-2 ${syncMutation.isPending ? 'animate-spin' : ''}`} />
                            {syncMutation.isPending ? 'Syncing...' : 'Sync Now'}
                        </Button>
                        <Button
                            variant="outline"
                            size="sm"
                            onClick={handleDisconnect}
                            disabled={disconnectMutation.isPending}
                            className="text-red-400 hover:text-red-300 hover:border-red-400/50"
                        >
                            <Unlink className="w-4 h-4" />
                        </Button>
                    </div>
                </div>
            </CardContent>
        </Card>
    )
}
