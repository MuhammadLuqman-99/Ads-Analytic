"use client";

import { Plus, RefreshCw, CheckCircle, AlertCircle, XCircle } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";

interface Connection {
  id: string;
  platform: "meta" | "tiktok" | "shopee";
  name: string;
  accountId: string;
  status: "connected" | "expired" | "error";
  lastSync?: string;
}

const mockConnections: Connection[] = [
  {
    id: "1",
    platform: "meta",
    name: "My Business Page",
    accountId: "act_123456789",
    status: "connected",
    lastSync: "2 hours ago",
  },
  {
    id: "2",
    platform: "tiktok",
    name: "TikTok Business",
    accountId: "tt_987654321",
    status: "connected",
    lastSync: "1 hour ago",
  },
  {
    id: "3",
    platform: "shopee",
    name: "Shopee Store",
    accountId: "shop_456789123",
    status: "expired",
  },
];

const platformConfig = {
  meta: {
    name: "Meta Ads",
    color: "bg-blue-600",
    letter: "M",
  },
  tiktok: {
    name: "TikTok Ads",
    color: "bg-black",
    letter: "T",
  },
  shopee: {
    name: "Shopee Ads",
    color: "bg-orange-500",
    letter: "S",
  },
};

const statusConfig = {
  connected: {
    label: "Connected",
    icon: CheckCircle,
    color: "text-emerald-600 bg-emerald-50",
  },
  expired: {
    label: "Token Expired",
    icon: AlertCircle,
    color: "text-amber-600 bg-amber-50",
  },
  error: {
    label: "Error",
    icon: XCircle,
    color: "text-red-600 bg-red-50",
  },
};

export default function ConnectionsPage() {
  return (
    <div>
      {/* Header */}
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4 mb-8">
        <div>
          <h1 className="text-3xl font-bold text-slate-900">Connections</h1>
          <p className="mt-1 text-slate-500">
            Manage your connected advertising accounts
          </p>
        </div>
        <Button>
          <Plus className="h-4 w-4 mr-2" />
          Connect Account
        </Button>
      </div>

      {/* Connections Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {mockConnections.map((connection) => {
          const platform = platformConfig[connection.platform];
          const status = statusConfig[connection.status];
          const StatusIcon = status.icon;

          return (
            <Card key={connection.id} className="bg-white border-slate-200">
              <CardContent className="p-6">
                <div className="flex items-start justify-between mb-4">
                  <div className="flex items-center gap-3">
                    <div
                      className={`h-12 w-12 rounded-lg ${platform.color} flex items-center justify-center text-white font-bold text-lg`}
                    >
                      {platform.letter}
                    </div>
                    <div>
                      <p className="font-semibold text-slate-900">
                        {platform.name}
                      </p>
                      <p className="text-sm text-slate-500">{connection.name}</p>
                    </div>
                  </div>
                </div>

                <div className="space-y-3">
                  <div className="flex items-center justify-between text-sm">
                    <span className="text-slate-500">Account ID</span>
                    <span className="font-mono text-slate-900">
                      {connection.accountId}
                    </span>
                  </div>

                  <div className="flex items-center justify-between text-sm">
                    <span className="text-slate-500">Status</span>
                    <Badge className={status.color}>
                      <StatusIcon className="h-3 w-3 mr-1" />
                      {status.label}
                    </Badge>
                  </div>

                  {connection.lastSync && (
                    <div className="flex items-center justify-between text-sm">
                      <span className="text-slate-500">Last Sync</span>
                      <span className="text-slate-900">{connection.lastSync}</span>
                    </div>
                  )}
                </div>

                <div className="flex gap-2 mt-4 pt-4 border-t border-slate-200">
                  {connection.status === "expired" ? (
                    <Button variant="outline" size="sm" className="flex-1">
                      Reconnect
                    </Button>
                  ) : (
                    <Button variant="outline" size="sm" className="flex-1">
                      <RefreshCw className="h-3 w-3 mr-1" />
                      Sync
                    </Button>
                  )}
                  <Button variant="ghost" size="sm" className="text-red-600">
                    Disconnect
                  </Button>
                </div>
              </CardContent>
            </Card>
          );
        })}

        {/* Add New Connection Card */}
        <Card className="bg-white border-slate-200 border-dashed">
          <CardContent className="p-6 flex flex-col items-center justify-center h-full min-h-[280px]">
            <div className="w-12 h-12 rounded-full bg-slate-100 flex items-center justify-center mb-4">
              <Plus className="h-6 w-6 text-slate-400" />
            </div>
            <p className="font-medium text-slate-900 mb-1">Add Connection</p>
            <p className="text-sm text-slate-500 text-center mb-4">
              Connect a new advertising platform
            </p>
            <Button variant="outline">Connect Account</Button>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
