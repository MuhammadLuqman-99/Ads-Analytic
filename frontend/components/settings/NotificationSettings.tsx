"use client";

import { useState } from "react";
import { Bell, Mail, FileText, AlertTriangle, TrendingDown, Zap } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { cn } from "@/lib/utils";

interface NotificationPreference {
  id: string;
  title: string;
  description: string;
  email: boolean;
  push: boolean;
}

interface AlertThreshold {
  id: string;
  name: string;
  metric: string;
  operator: "less_than" | "greater_than";
  value: number;
  enabled: boolean;
}

const defaultPreferences: NotificationPreference[] = [
  {
    id: "campaign_updates",
    title: "Campaign Updates",
    description: "Get notified when campaigns are paused, ended, or have issues",
    email: true,
    push: true,
  },
  {
    id: "sync_alerts",
    title: "Sync Alerts",
    description: "Notifications about data sync status and errors",
    email: true,
    push: false,
  },
  {
    id: "billing",
    title: "Billing & Invoices",
    description: "Payment confirmations and invoice notifications",
    email: true,
    push: false,
  },
  {
    id: "product_updates",
    title: "Product Updates",
    description: "New features, improvements, and tips",
    email: false,
    push: false,
  },
];

const defaultThresholds: AlertThreshold[] = [
  {
    id: "roas_low",
    name: "Low ROAS Alert",
    metric: "ROAS",
    operator: "less_than",
    value: 2.0,
    enabled: true,
  },
  {
    id: "spend_high",
    name: "High Spend Alert",
    metric: "Daily Spend",
    operator: "greater_than",
    value: 500,
    enabled: true,
  },
  {
    id: "ctr_low",
    name: "Low CTR Alert",
    metric: "CTR",
    operator: "less_than",
    value: 1.0,
    enabled: false,
  },
];

// Toggle Switch Component
function Switch({
  checked,
  onChange,
  disabled,
}: {
  checked: boolean;
  onChange: (checked: boolean) => void;
  disabled?: boolean;
}) {
  return (
    <button
      type="button"
      role="switch"
      aria-checked={checked}
      disabled={disabled}
      onClick={() => onChange(!checked)}
      className={cn(
        "relative inline-flex h-6 w-11 items-center rounded-full transition-colors",
        checked ? "bg-blue-600" : "bg-slate-200",
        disabled && "opacity-50 cursor-not-allowed"
      )}
    >
      <span
        className={cn(
          "inline-block h-4 w-4 transform rounded-full bg-white transition-transform",
          checked ? "translate-x-6" : "translate-x-1"
        )}
      />
    </button>
  );
}

export function NotificationSettings() {
  const [preferences, setPreferences] = useState(defaultPreferences);
  const [thresholds, setThresholds] = useState(defaultThresholds);
  const [weeklyReport, setWeeklyReport] = useState(true);
  const [isSaving, setIsSaving] = useState(false);

  const updatePreference = (
    id: string,
    field: "email" | "push",
    value: boolean
  ) => {
    setPreferences((prev) =>
      prev.map((p) => (p.id === id ? { ...p, [field]: value } : p))
    );
  };

  const updateThreshold = (
    id: string,
    field: "value" | "enabled",
    value: number | boolean
  ) => {
    setThresholds((prev) =>
      prev.map((t) => (t.id === id ? { ...t, [field]: value } : t))
    );
  };

  const handleSave = async () => {
    setIsSaving(true);
    await new Promise((resolve) => setTimeout(resolve, 1000));
    console.log("Saved:", { preferences, thresholds, weeklyReport });
    setIsSaving(false);
  };

  return (
    <div className="space-y-6 max-w-2xl">
      {/* Email Preferences */}
      <Card className="bg-white border-slate-200">
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-slate-900">
            <Mail className="h-5 w-5" />
            Email Preferences
          </CardTitle>
          <CardDescription>
            Choose what emails you want to receive
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          {preferences.map((pref) => (
            <div
              key={pref.id}
              className="flex items-start justify-between gap-4 pb-4 border-b border-slate-100 last:border-0 last:pb-0"
            >
              <div className="flex-1">
                <p className="font-medium text-slate-900">{pref.title}</p>
                <p className="text-sm text-slate-500 mt-0.5">{pref.description}</p>
              </div>
              <div className="flex items-center gap-4">
                <div className="flex items-center gap-2">
                  <span className="text-xs text-slate-500">Email</span>
                  <Switch
                    checked={pref.email}
                    onChange={(checked) => updatePreference(pref.id, "email", checked)}
                  />
                </div>
              </div>
            </div>
          ))}
        </CardContent>
      </Card>

      {/* Weekly Report */}
      <Card className="bg-white border-slate-200">
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-slate-900">
            <FileText className="h-5 w-5" />
            Weekly Report
          </CardTitle>
          <CardDescription>
            Receive a summary of your advertising performance every week
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex items-center justify-between p-4 bg-slate-50 rounded-lg">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 rounded-lg bg-blue-100 flex items-center justify-center">
                <Zap className="h-5 w-5 text-blue-600" />
              </div>
              <div>
                <p className="font-medium text-slate-900">Weekly Performance Report</p>
                <p className="text-sm text-slate-500">
                  Sent every Monday at 9:00 AM
                </p>
              </div>
            </div>
            <Switch checked={weeklyReport} onChange={setWeeklyReport} />
          </div>

          {weeklyReport && (
            <div className="mt-4 p-4 bg-blue-50 border border-blue-200 rounded-lg">
              <p className="text-sm text-blue-800">
                Your weekly report includes:
              </p>
              <ul className="mt-2 space-y-1 text-sm text-blue-700">
                <li>• Total spend and revenue across all platforms</li>
                <li>• Top performing campaigns</li>
                <li>• Week-over-week comparison</li>
                <li>• Recommendations for improvement</li>
              </ul>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Alert Thresholds */}
      <Card className="bg-white border-slate-200">
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-slate-900">
            <AlertTriangle className="h-5 w-5" />
            Alert Thresholds
          </CardTitle>
          <CardDescription>
            Get notified when metrics cross your defined thresholds
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          {thresholds.map((threshold) => (
            <div
              key={threshold.id}
              className={cn(
                "p-4 rounded-lg border transition-colors",
                threshold.enabled
                  ? "bg-white border-slate-200"
                  : "bg-slate-50 border-slate-100"
              )}
            >
              <div className="flex items-center justify-between mb-3">
                <div className="flex items-center gap-2">
                  <TrendingDown
                    className={cn(
                      "h-4 w-4",
                      threshold.enabled ? "text-amber-500" : "text-slate-400"
                    )}
                  />
                  <span className="font-medium text-slate-900">{threshold.name}</span>
                </div>
                <Switch
                  checked={threshold.enabled}
                  onChange={(checked) => updateThreshold(threshold.id, "enabled", checked)}
                />
              </div>

              <div className="flex items-center gap-3">
                <span className="text-sm text-slate-600">
                  Alert when {threshold.metric} is{" "}
                  {threshold.operator === "less_than" ? "less than" : "greater than"}
                </span>
                <Input
                  type="number"
                  value={threshold.value}
                  onChange={(e) =>
                    updateThreshold(threshold.id, "value", parseFloat(e.target.value))
                  }
                  disabled={!threshold.enabled}
                  className="w-24 h-8"
                  step={threshold.metric === "CTR" || threshold.metric === "ROAS" ? 0.1 : 1}
                />
                <span className="text-sm text-slate-600">
                  {threshold.metric === "Daily Spend" && "$"}
                  {threshold.metric === "CTR" && "%"}
                  {threshold.metric === "ROAS" && "x"}
                </span>
              </div>
            </div>
          ))}

          <Button variant="outline" className="w-full">
            <AlertTriangle className="h-4 w-4 mr-2" />
            Add Custom Alert
          </Button>
        </CardContent>
      </Card>

      {/* Save Button */}
      <div className="flex justify-end">
        <Button onClick={handleSave} disabled={isSaving}>
          {isSaving ? "Saving..." : "Save Preferences"}
        </Button>
      </div>
    </div>
  );
}
