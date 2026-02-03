"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";

import { OnboardingProgress } from "@/components/onboarding/OnboardingProgress";
import {
  PlatformConnectCard,
  MetaIcon,
  TikTokIcon,
  ShopeeIcon,
} from "@/components/onboarding/PlatformConnectCard";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";

const steps = [
  { title: "Company Details", description: "Tell us about your business" },
  { title: "Connect Account", description: "Link your ad platforms" },
  { title: "Quick Tour", description: "Learn the basics" },
];

const companyDetailsSchema = z.object({
  companySize: z.string().min(1, "Please select company size"),
  industry: z.string().min(1, "Please select your industry"),
  monthlyAdSpend: z.string().min(1, "Please select monthly ad spend"),
  primaryGoal: z.string().min(1, "Please select your primary goal"),
});

type CompanyDetailsFormData = z.infer<typeof companyDetailsSchema>;

export default function OnboardingPage() {
  const router = useRouter();
  const [currentStep, setCurrentStep] = useState(0);
  const [isLoading, setIsLoading] = useState(false);
  const [connectedPlatforms, setConnectedPlatforms] = useState<string[]>([]);
  const [showTourModal, setShowTourModal] = useState(false);

  const {
    register,
    handleSubmit,
    formState: { errors },
    setValue,
    watch,
  } = useForm<CompanyDetailsFormData>({
    resolver: zodResolver(companyDetailsSchema),
    defaultValues: {
      companySize: "",
      industry: "",
      monthlyAdSpend: "",
      primaryGoal: "",
    },
  });

  const handleCompanyDetailsSubmit = async (data: CompanyDetailsFormData) => {
    setIsLoading(true);
    try {
      // Store in localStorage for now (backend endpoint not implemented yet)
      localStorage.setItem("onboarding_company_details", JSON.stringify({
        company_size: data.companySize,
        industry: data.industry,
        monthly_ad_spend: data.monthlyAdSpend,
        primary_goal: data.primaryGoal,
      }));
      setCurrentStep(1);
    } catch (error) {
      console.error("Failed to save company details:", error);
      // Still proceed to next step
      setCurrentStep(1);
    } finally {
      setIsLoading(false);
    }
  };

  const handleConnectPlatform = async (platform: string) => {
    // Redirect to OAuth flow
    window.location.href = `${process.env.NEXT_PUBLIC_API_URL}/api/v1/oauth/${platform}/authorize`;
  };

  const handleSkipConnect = () => {
    setCurrentStep(2);
    setShowTourModal(true);
  };

  const handleContinueFromConnect = () => {
    if (connectedPlatforms.length > 0) {
      setCurrentStep(2);
      setShowTourModal(true);
    }
  };

  const handleFinishOnboarding = async () => {
    setIsLoading(true);
    try {
      // Mark onboarding as complete in localStorage
      localStorage.setItem("onboarding_completed", "true");
      router.push("/dashboard");
    } catch (error) {
      console.error("Failed to complete onboarding:", error);
      router.push("/dashboard");
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-slate-50">
      {/* Header */}
      <header className="bg-white border-b border-slate-200">
        <div className="max-w-4xl mx-auto px-4 py-4">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 bg-gradient-to-br from-blue-600 to-blue-700 rounded-xl flex items-center justify-center">
              <svg
                className="w-6 h-6 text-white"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"
                />
              </svg>
            </div>
            <span className="text-xl font-bold text-slate-900">Ads Analytics</span>
          </div>
        </div>
      </header>

      {/* Progress */}
      <div className="max-w-4xl mx-auto px-4 py-8">
        <OnboardingProgress steps={steps} currentStep={currentStep} />
      </div>

      {/* Content */}
      <div className="max-w-2xl mx-auto px-4 pb-12">
        {/* Step 1: Company Details */}
        {currentStep === 0 && (
          <div className="bg-white rounded-xl shadow-sm border border-slate-200 p-8">
            <h2 className="text-2xl font-bold text-slate-900 mb-2">
              Tell us about your business
            </h2>
            <p className="text-slate-600 mb-8">
              This helps us personalize your experience and provide better insights.
            </p>

            <form onSubmit={handleSubmit(handleCompanyDetailsSubmit)} className="space-y-6">
              <div className="space-y-2">
                <Label htmlFor="companySize">Company Size</Label>
                <Select onValueChange={(value) => setValue("companySize", value)}>
                  <SelectTrigger>
                    <SelectValue placeholder="Select company size" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="1-10">1-10 employees</SelectItem>
                    <SelectItem value="11-50">11-50 employees</SelectItem>
                    <SelectItem value="51-200">51-200 employees</SelectItem>
                    <SelectItem value="201-500">201-500 employees</SelectItem>
                    <SelectItem value="500+">500+ employees</SelectItem>
                  </SelectContent>
                </Select>
                {errors.companySize && (
                  <p className="text-sm text-red-600">{errors.companySize.message}</p>
                )}
              </div>

              <div className="space-y-2">
                <Label htmlFor="industry">Industry</Label>
                <Select onValueChange={(value) => setValue("industry", value)}>
                  <SelectTrigger>
                    <SelectValue placeholder="Select your industry" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="ecommerce">E-commerce</SelectItem>
                    <SelectItem value="saas">SaaS / Software</SelectItem>
                    <SelectItem value="agency">Marketing Agency</SelectItem>
                    <SelectItem value="retail">Retail</SelectItem>
                    <SelectItem value="finance">Finance</SelectItem>
                    <SelectItem value="healthcare">Healthcare</SelectItem>
                    <SelectItem value="education">Education</SelectItem>
                    <SelectItem value="other">Other</SelectItem>
                  </SelectContent>
                </Select>
                {errors.industry && (
                  <p className="text-sm text-red-600">{errors.industry.message}</p>
                )}
              </div>

              <div className="space-y-2">
                <Label htmlFor="monthlyAdSpend">Monthly Ad Spend</Label>
                <Select onValueChange={(value) => setValue("monthlyAdSpend", value)}>
                  <SelectTrigger>
                    <SelectValue placeholder="Select monthly ad spend" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="<1000">Less than RM 1,000</SelectItem>
                    <SelectItem value="1000-5000">RM 1,000 - RM 5,000</SelectItem>
                    <SelectItem value="5000-20000">RM 5,000 - RM 20,000</SelectItem>
                    <SelectItem value="20000-100000">RM 20,000 - RM 100,000</SelectItem>
                    <SelectItem value="100000+">More than RM 100,000</SelectItem>
                  </SelectContent>
                </Select>
                {errors.monthlyAdSpend && (
                  <p className="text-sm text-red-600">{errors.monthlyAdSpend.message}</p>
                )}
              </div>

              <div className="space-y-2">
                <Label htmlFor="primaryGoal">Primary Goal</Label>
                <Select onValueChange={(value) => setValue("primaryGoal", value)}>
                  <SelectTrigger>
                    <SelectValue placeholder="What's your main goal?" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="consolidate">Consolidate ad data in one place</SelectItem>
                    <SelectItem value="optimize">Optimize ad performance</SelectItem>
                    <SelectItem value="reporting">Better reporting & analytics</SelectItem>
                    <SelectItem value="budget">Manage budget across platforms</SelectItem>
                    <SelectItem value="scale">Scale ad campaigns</SelectItem>
                  </SelectContent>
                </Select>
                {errors.primaryGoal && (
                  <p className="text-sm text-red-600">{errors.primaryGoal.message}</p>
                )}
              </div>

              <Button type="submit" className="w-full h-11" disabled={isLoading}>
                {isLoading ? "Saving..." : "Continue"}
              </Button>
            </form>
          </div>
        )}

        {/* Step 2: Connect Ad Account */}
        {currentStep === 1 && (
          <div className="bg-white rounded-xl shadow-sm border border-slate-200 p-8">
            <h2 className="text-2xl font-bold text-slate-900 mb-2">
              Connect your ad accounts
            </h2>
            <p className="text-slate-600 mb-8">
              Link at least one platform to start tracking your campaigns.
            </p>

            <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-8">
              <PlatformConnectCard
                platform="meta"
                name="Meta Ads"
                description="Facebook & Instagram"
                icon={<MetaIcon />}
                isConnected={connectedPlatforms.includes("meta")}
                onConnect={() => handleConnectPlatform("meta")}
              />
              <PlatformConnectCard
                platform="tiktok"
                name="TikTok Ads"
                description="TikTok for Business"
                icon={<TikTokIcon />}
                isConnected={connectedPlatforms.includes("tiktok")}
                onConnect={() => handleConnectPlatform("tiktok")}
              />
              <PlatformConnectCard
                platform="shopee"
                name="Shopee Ads"
                description="Shopee Seller Center"
                icon={<ShopeeIcon />}
                isConnected={connectedPlatforms.includes("shopee")}
                onConnect={() => handleConnectPlatform("shopee")}
              />
            </div>

            <div className="flex gap-4">
              <Button
                variant="outline"
                className="flex-1"
                onClick={handleSkipConnect}
              >
                Skip for now
              </Button>
              <Button
                className="flex-1"
                onClick={handleContinueFromConnect}
                disabled={connectedPlatforms.length === 0}
              >
                Continue
              </Button>
            </div>
          </div>
        )}

        {/* Step 3: Quick Tour */}
        {currentStep === 2 && (
          <div className="bg-white rounded-xl shadow-sm border border-slate-200 p-8">
            <div className="text-center">
              <div className="w-20 h-20 mx-auto bg-blue-100 rounded-full flex items-center justify-center mb-6">
                <svg
                  className="w-10 h-10 text-blue-600"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M13 10V3L4 14h7v7l9-11h-7z"
                  />
                </svg>
              </div>

              <h2 className="text-2xl font-bold text-slate-900 mb-2">
                You&apos;re all set!
              </h2>
              <p className="text-slate-600 mb-8">
                Here&apos;s a quick overview of what you can do with Ads Analytics.
              </p>

              <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8 text-left">
                <div className="p-4 rounded-lg bg-slate-50">
                  <div className="w-10 h-10 bg-blue-100 rounded-lg flex items-center justify-center mb-3">
                    <svg
                      className="w-5 h-5 text-blue-600"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={2}
                        d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"
                      />
                    </svg>
                  </div>
                  <h3 className="font-semibold text-slate-900 mb-1">Dashboard</h3>
                  <p className="text-sm text-slate-500">
                    View all your ad performance metrics in one unified dashboard.
                  </p>
                </div>

                <div className="p-4 rounded-lg bg-slate-50">
                  <div className="w-10 h-10 bg-green-100 rounded-lg flex items-center justify-center mb-3">
                    <svg
                      className="w-5 h-5 text-green-600"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={2}
                        d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6"
                      />
                    </svg>
                  </div>
                  <h3 className="font-semibold text-slate-900 mb-1">Analytics</h3>
                  <p className="text-sm text-slate-500">
                    Deep dive into trends and optimize your campaign performance.
                  </p>
                </div>

                <div className="p-4 rounded-lg bg-slate-50">
                  <div className="w-10 h-10 bg-purple-100 rounded-lg flex items-center justify-center mb-3">
                    <svg
                      className="w-5 h-5 text-purple-600"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={2}
                        d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
                      />
                    </svg>
                  </div>
                  <h3 className="font-semibold text-slate-900 mb-1">Reports</h3>
                  <p className="text-sm text-slate-500">
                    Generate and schedule automated reports for stakeholders.
                  </p>
                </div>
              </div>

              <Button
                className="w-full h-11"
                onClick={handleFinishOnboarding}
                disabled={isLoading}
              >
                {isLoading ? "Completing setup..." : "Go to Dashboard"}
              </Button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
