"use client"

import Sidebar from "@/components/Sidebar"
import { ConnectedAccountCard } from "@/components/ConnectedAccountCard"
import { ConnectPlatformButton } from "@/components/ConnectPlatformButton"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { useConnectedAccounts } from "@/hooks/use-accounts"
import { type Platform } from "@/lib/mock-data"

const availablePlatforms: Platform[] = ['meta', 'tiktok', 'shopee']

export default function ConnectPage() {
    const { data: accounts, isLoading } = useConnectedAccounts()

    const connectedPlatforms = accounts?.map(a => a.platform) || []
    const unconnectedPlatforms = availablePlatforms.filter(p => !connectedPlatforms.includes(p))

    return (
        <div className="min-h-screen bg-slate-900">
            <Sidebar />

            <main className="lg:pl-72">
                <div className="px-4 sm:px-6 lg:px-8 py-8">
                    {/* Header */}
                    <div className="mb-8">
                        <h1 className="text-3xl font-bold text-white">Connected Accounts</h1>
                        <p className="mt-1 text-slate-400">Manage your ad platform connections and sync settings</p>
                    </div>

                    {/* Connected Accounts */}
                    <section className="mb-10">
                        <h2 className="text-xl font-semibold text-white mb-4 flex items-center gap-2">
                            <span className="w-2 h-2 rounded-full bg-emerald-400" />
                            Active Connections
                        </h2>

                        {isLoading ? (
                            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                                {[...Array(3)].map((_, i) => (
                                    <Card key={i} className="animate-pulse">
                                        <div className="h-2 bg-slate-700" />
                                        <CardContent className="p-6">
                                            <div className="flex items-center gap-4 mb-6">
                                                <div className="h-14 w-14 bg-slate-700 rounded-xl" />
                                                <div>
                                                    <div className="h-5 w-32 bg-slate-700 rounded mb-2" />
                                                    <div className="h-4 w-24 bg-slate-700 rounded" />
                                                </div>
                                            </div>
                                            <div className="h-10 bg-slate-700 rounded" />
                                        </CardContent>
                                    </Card>
                                ))}
                            </div>
                        ) : accounts && accounts.length > 0 ? (
                            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                                {accounts.map((account) => (
                                    <ConnectedAccountCard key={account.id} account={account} />
                                ))}
                            </div>
                        ) : (
                            <Card>
                                <CardContent className="p-8 text-center">
                                    <p className="text-slate-400">No connected accounts yet</p>
                                    <p className="text-sm text-slate-500 mt-1">Connect a platform below to get started</p>
                                </CardContent>
                            </Card>
                        )}
                    </section>

                    {/* Add New Connection */}
                    <section>
                        <h2 className="text-xl font-semibold text-white mb-4 flex items-center gap-2">
                            <span className="w-2 h-2 rounded-full bg-indigo-400" />
                            Add New Connection
                        </h2>

                        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                            {availablePlatforms.map((platform) => (
                                <ConnectPlatformButton
                                    key={platform}
                                    platform={platform}
                                    isConnected={connectedPlatforms.includes(platform)}
                                />
                            ))}
                        </div>
                    </section>

                    {/* Help Section */}
                    <section className="mt-10">
                        <Card className="bg-gradient-to-r from-slate-800/50 to-indigo-900/20 border-indigo-500/20">
                            <CardHeader>
                                <CardTitle className="text-lg">Need Help Connecting?</CardTitle>
                                <CardDescription>
                                    Learn how to set up your ad platform connections
                                </CardDescription>
                            </CardHeader>
                            <CardContent>
                                <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                                    <div className="p-4 rounded-lg bg-slate-800/50">
                                        <h4 className="font-medium text-white mb-2">Meta Ads Setup</h4>
                                        <p className="text-sm text-slate-400 mb-3">
                                            Connect your Meta Business account to sync Facebook and Instagram ads.
                                        </p>
                                        <a href="#" className="text-sm text-indigo-400 hover:text-indigo-300">
                                            View guide →
                                        </a>
                                    </div>
                                    <div className="p-4 rounded-lg bg-slate-800/50">
                                        <h4 className="font-medium text-white mb-2">TikTok Ads Setup</h4>
                                        <p className="text-sm text-slate-400 mb-3">
                                            Link your TikTok for Business account to import campaign data.
                                        </p>
                                        <a href="#" className="text-sm text-indigo-400 hover:text-indigo-300">
                                            View guide →
                                        </a>
                                    </div>
                                    <div className="p-4 rounded-lg bg-slate-800/50">
                                        <h4 className="font-medium text-white mb-2">Shopee Ads Setup</h4>
                                        <p className="text-sm text-slate-400 mb-3">
                                            Connect your Shopee Seller account to track store ad performance.
                                        </p>
                                        <a href="#" className="text-sm text-indigo-400 hover:text-indigo-300">
                                            View guide →
                                        </a>
                                    </div>
                                </div>
                            </CardContent>
                        </Card>
                    </section>
                </div>
            </main>
        </div>
    )
}
