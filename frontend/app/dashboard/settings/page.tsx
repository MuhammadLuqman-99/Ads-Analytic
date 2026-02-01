"use client";

import { useState } from "react";
import { User, Bell, Shield, Globe, Palette } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { cn } from "@/lib/utils";

const settingsTabs = [
  { id: "profile", label: "Profile", icon: User },
  { id: "notifications", label: "Notifications", icon: Bell },
  { id: "security", label: "Security", icon: Shield },
  { id: "preferences", label: "Preferences", icon: Palette },
];

export default function SettingsPage() {
  const [activeTab, setActiveTab] = useState("profile");

  return (
    <div>
      {/* Header */}
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-slate-900">Settings</h1>
        <p className="mt-1 text-slate-500">
          Manage your account settings and preferences
        </p>
      </div>

      <div className="flex flex-col lg:flex-row gap-8">
        {/* Tabs */}
        <nav className="lg:w-64 space-y-1">
          {settingsTabs.map((tab) => (
            <button
              key={tab.id}
              onClick={() => setActiveTab(tab.id)}
              className={cn(
                "w-full flex items-center gap-3 px-4 py-3 rounded-lg text-left transition-colors",
                activeTab === tab.id
                  ? "bg-blue-50 text-blue-600"
                  : "text-slate-600 hover:bg-slate-50"
              )}
            >
              <tab.icon className="h-5 w-5" />
              <span className="font-medium">{tab.label}</span>
            </button>
          ))}
        </nav>

        {/* Content */}
        <div className="flex-1">
          {activeTab === "profile" && (
            <Card className="bg-white border-slate-200">
              <CardHeader>
                <CardTitle className="text-slate-900">Profile Settings</CardTitle>
                <CardDescription>
                  Update your personal information and profile details
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-6">
                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                  <div className="space-y-2">
                    <Label htmlFor="firstName">First Name</Label>
                    <Input id="firstName" placeholder="John" />
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="lastName">Last Name</Label>
                    <Input id="lastName" placeholder="Doe" />
                  </div>
                </div>
                <div className="space-y-2">
                  <Label htmlFor="email">Email</Label>
                  <Input id="email" type="email" placeholder="john@example.com" />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="company">Company</Label>
                  <Input id="company" placeholder="Acme Inc." />
                </div>
                <Button>Save Changes</Button>
              </CardContent>
            </Card>
          )}

          {activeTab === "notifications" && (
            <Card className="bg-white border-slate-200">
              <CardHeader>
                <CardTitle className="text-slate-900">Notification Preferences</CardTitle>
                <CardDescription>
                  Choose how you want to receive notifications
                </CardDescription>
              </CardHeader>
              <CardContent>
                <div className="space-y-4">
                  {[
                    { id: "email", label: "Email notifications", description: "Receive updates via email" },
                    { id: "push", label: "Push notifications", description: "Browser push notifications" },
                    { id: "weekly", label: "Weekly digest", description: "Weekly performance summary" },
                    { id: "alerts", label: "Alert notifications", description: "Critical campaign alerts" },
                  ].map((item) => (
                    <div
                      key={item.id}
                      className="flex items-center justify-between py-3 border-b border-slate-100 last:border-0"
                    >
                      <div>
                        <p className="font-medium text-slate-900">{item.label}</p>
                        <p className="text-sm text-slate-500">{item.description}</p>
                      </div>
                      <input type="checkbox" className="h-4 w-4" defaultChecked />
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>
          )}

          {activeTab === "security" && (
            <Card className="bg-white border-slate-200">
              <CardHeader>
                <CardTitle className="text-slate-900">Security Settings</CardTitle>
                <CardDescription>
                  Manage your password and security preferences
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-6">
                <div className="space-y-2">
                  <Label htmlFor="currentPassword">Current Password</Label>
                  <Input id="currentPassword" type="password" />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="newPassword">New Password</Label>
                  <Input id="newPassword" type="password" />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="confirmPassword">Confirm New Password</Label>
                  <Input id="confirmPassword" type="password" />
                </div>
                <Button>Update Password</Button>
              </CardContent>
            </Card>
          )}

          {activeTab === "preferences" && (
            <Card className="bg-white border-slate-200">
              <CardHeader>
                <CardTitle className="text-slate-900">Preferences</CardTitle>
                <CardDescription>
                  Customize your dashboard experience
                </CardDescription>
              </CardHeader>
              <CardContent>
                <div className="space-y-6">
                  <div className="space-y-2">
                    <Label>Default Currency</Label>
                    <select className="w-full border border-slate-200 rounded-lg px-3 py-2">
                      <option>MYR (Malaysian Ringgit)</option>
                      <option>USD (US Dollar)</option>
                      <option>SGD (Singapore Dollar)</option>
                    </select>
                  </div>
                  <div className="space-y-2">
                    <Label>Timezone</Label>
                    <select className="w-full border border-slate-200 rounded-lg px-3 py-2">
                      <option>Asia/Kuala_Lumpur (GMT+8)</option>
                      <option>Asia/Singapore (GMT+8)</option>
                      <option>UTC</option>
                    </select>
                  </div>
                  <div className="space-y-2">
                    <Label>Date Format</Label>
                    <select className="w-full border border-slate-200 rounded-lg px-3 py-2">
                      <option>DD/MM/YYYY</option>
                      <option>MM/DD/YYYY</option>
                      <option>YYYY-MM-DD</option>
                    </select>
                  </div>
                  <Button>Save Preferences</Button>
                </div>
              </CardContent>
            </Card>
          )}
        </div>
      </div>
    </div>
  );
}
