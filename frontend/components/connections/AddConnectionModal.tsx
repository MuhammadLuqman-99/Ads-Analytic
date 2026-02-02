"use client";

import { useState } from "react";
import { X, ExternalLink, Check, AlertCircle, Loader2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { cn } from "@/lib/utils";
import { type Platform } from "@/lib/api/types";
import { type PlanLimits } from "./types";

interface AddConnectionModalProps {
  isOpen: boolean;
  onClose: () => void;
  onConnect: (platform: Platform) => Promise<void>;
  planLimits: PlanLimits;
}

type ConnectionStep = "select" | "connecting" | "success" | "error";

// Only show supported platforms in the connection modal
const supportedPlatforms: Platform[] = ["meta", "tiktok", "shopee"];

const platformConfigs: Record<
  Platform,
  { name: string; description: string; bg: string; text: string; icon: string }
> = {
  meta: {
    name: "Meta Ads",
    description: "Connect Facebook & Instagram ads accounts",
    bg: "bg-blue-100 hover:bg-blue-200",
    text: "text-blue-600",
    icon: "M",
  },
  google: {
    name: "Google Ads",
    description: "Connect Google Ads accounts",
    bg: "bg-red-100 hover:bg-red-200",
    text: "text-red-600",
    icon: "G",
  },
  tiktok: {
    name: "TikTok Ads",
    description: "Connect TikTok for Business accounts",
    bg: "bg-slate-900 hover:bg-slate-800",
    text: "text-white",
    icon: "T",
  },
  shopee: {
    name: "Shopee Ads",
    description: "Connect Shopee Seller Center accounts",
    bg: "bg-orange-100 hover:bg-orange-200",
    text: "text-orange-600",
    icon: "S",
  },
  linkedin: {
    name: "LinkedIn Ads",
    description: "Connect LinkedIn Marketing accounts",
    bg: "bg-blue-100 hover:bg-blue-200",
    text: "text-blue-700",
    icon: "L",
  },
};

export function AddConnectionModal({
  isOpen,
  onClose,
  onConnect,
  planLimits,
}: AddConnectionModalProps) {
  const [step, setStep] = useState<ConnectionStep>("select");
  const [selectedPlatform, setSelectedPlatform] = useState<Platform | null>(null);
  const [error, setError] = useState<string | null>(null);

  const isAtLimit = planLimits.accountsUsed >= planLimits.accountsLimit;

  const handleSelectPlatform = async (platform: Platform) => {
    if (isAtLimit) return;

    setSelectedPlatform(platform);
    setStep("connecting");
    setError(null);

    try {
      await onConnect(platform);
      setStep("success");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Connection failed");
      setStep("error");
    }
  };

  const handleClose = () => {
    setStep("select");
    setSelectedPlatform(null);
    setError(null);
    onClose();
  };

  const handleRetry = () => {
    setStep("select");
    setError(null);
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      {/* Backdrop */}
      <div
        className="absolute inset-0 bg-black/50 backdrop-blur-sm"
        onClick={handleClose}
      />

      {/* Modal */}
      <div className="relative bg-white rounded-xl shadow-xl max-w-lg w-full mx-4 max-h-[90vh] overflow-y-auto">
        {/* Header */}
        <div className="flex items-center justify-between p-6 border-b border-slate-200">
          <h2 className="text-xl font-semibold text-slate-900">
            {step === "select" && "Connect Ad Account"}
            {step === "connecting" && "Connecting..."}
            {step === "success" && "Connected!"}
            {step === "error" && "Connection Failed"}
          </h2>
          <Button variant="ghost" size="icon" onClick={handleClose}>
            <X className="h-5 w-5" />
          </Button>
        </div>

        {/* Content */}
        <div className="p-6">
          {/* Platform Selection */}
          {step === "select" && (
            <>
              {/* Limit Warning */}
              {isAtLimit && (
                <div className="mb-6 p-4 rounded-lg bg-amber-50 border border-amber-200">
                  <div className="flex items-start gap-3">
                    <AlertCircle className="h-5 w-5 text-amber-600 mt-0.5" />
                    <div>
                      <p className="font-medium text-amber-800">
                        Account Limit Reached
                      </p>
                      <p className="text-sm text-amber-700 mt-1">
                        You&apos;ve connected {planLimits.accountsUsed} of{" "}
                        {planLimits.accountsLimit} accounts on your{" "}
                        {planLimits.currentPlan} plan.
                      </p>
                      <Button
                        size="sm"
                        className="mt-3"
                        onClick={() => {
                          handleClose();
                          window.location.href = "/dashboard/billing";
                        }}
                      >
                        Upgrade Plan
                      </Button>
                    </div>
                  </div>
                </div>
              )}

              {/* Usage Display */}
              {!isAtLimit && (
                <div className="mb-6">
                  <div className="flex items-center justify-between text-sm mb-2">
                    <span className="text-slate-600">Connected Accounts</span>
                    <span className="font-medium text-slate-900">
                      {planLimits.accountsUsed} / {planLimits.accountsLimit}
                    </span>
                  </div>
                  <div className="h-2 bg-slate-100 rounded-full overflow-hidden">
                    <div
                      className="h-full bg-blue-500 rounded-full transition-all"
                      style={{
                        width: `${(planLimits.accountsUsed / planLimits.accountsLimit) * 100}%`,
                      }}
                    />
                  </div>
                </div>
              )}

              {/* Platform Options */}
              <div className="space-y-3">
                <p className="text-sm text-slate-500 mb-4">
                  Select a platform to connect your ad account
                </p>
                {supportedPlatforms.map((platform) => {
                  const config = platformConfigs[platform];
                  return (
                    <button
                      key={platform}
                      onClick={() => handleSelectPlatform(platform)}
                      disabled={isAtLimit}
                      className={cn(
                        "w-full flex items-center gap-4 p-4 rounded-lg border-2 border-transparent transition-all",
                        isAtLimit
                          ? "opacity-50 cursor-not-allowed bg-slate-100"
                          : "hover:border-blue-500 bg-slate-50 hover:bg-white"
                      )}
                    >
                      <div
                        className={cn(
                          "w-12 h-12 rounded-lg flex items-center justify-center font-bold text-xl transition-colors",
                          config.bg,
                          config.text
                        )}
                      >
                        {config.icon}
                      </div>
                      <div className="flex-1 text-left">
                        <p className="font-medium text-slate-900">{config.name}</p>
                        <p className="text-sm text-slate-500">{config.description}</p>
                      </div>
                      <ExternalLink className="h-5 w-5 text-slate-400" />
                    </button>
                  );
                })}
              </div>
            </>
          )}

          {/* Connecting State */}
          {step === "connecting" && selectedPlatform && (
            <div className="text-center py-8">
              <div
                className={cn(
                  "w-20 h-20 rounded-2xl flex items-center justify-center font-bold text-3xl mx-auto mb-6",
                  platformConfigs[selectedPlatform].bg,
                  platformConfigs[selectedPlatform].text
                )}
              >
                {platformConfigs[selectedPlatform].icon}
              </div>
              <div className="flex items-center justify-center gap-2 mb-4">
                <Loader2 className="h-5 w-5 animate-spin text-blue-600" />
                <span className="text-slate-600">
                  Connecting to {platformConfigs[selectedPlatform].name}...
                </span>
              </div>
              <p className="text-sm text-slate-500">
                You&apos;ll be redirected to authorize access to your ad account.
              </p>
            </div>
          )}

          {/* Success State */}
          {step === "success" && selectedPlatform && (
            <div className="text-center py-8">
              <div className="w-20 h-20 rounded-full bg-emerald-100 flex items-center justify-center mx-auto mb-6">
                <Check className="h-10 w-10 text-emerald-600" />
              </div>
              <h3 className="text-lg font-semibold text-slate-900 mb-2">
                Successfully Connected!
              </h3>
              <p className="text-slate-600 mb-6">
                Your {platformConfigs[selectedPlatform].name} account has been
                connected. Data will start syncing shortly.
              </p>
              <Button onClick={handleClose}>Done</Button>
            </div>
          )}

          {/* Error State */}
          {step === "error" && (
            <div className="text-center py-8">
              <div className="w-20 h-20 rounded-full bg-red-100 flex items-center justify-center mx-auto mb-6">
                <AlertCircle className="h-10 w-10 text-red-600" />
              </div>
              <h3 className="text-lg font-semibold text-slate-900 mb-2">
                Connection Failed
              </h3>
              <p className="text-slate-600 mb-2">
                {error || "Unable to connect your account. Please try again."}
              </p>
              <p className="text-sm text-slate-500 mb-6">
                If the problem persists, please contact support.
              </p>
              <div className="flex items-center justify-center gap-3">
                <Button variant="outline" onClick={handleClose}>
                  Cancel
                </Button>
                <Button onClick={handleRetry}>Try Again</Button>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
