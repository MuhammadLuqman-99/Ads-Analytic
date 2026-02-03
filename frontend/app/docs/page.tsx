import Link from "next/link";
import { Facebook, Music2, ShoppingBag, ArrowRight, CheckCircle2 } from "lucide-react";

export const metadata = {
  title: "Documentation - AdsAnalytic",
  description: "Learn how to connect your Meta Ads, TikTok Ads, and Shopee Ads accounts to AdsAnalytic.",
};

const platforms = [
  {
    title: "Meta (Facebook) Ads",
    description: "Connect your Facebook and Instagram ad accounts to track performance across Meta platforms.",
    href: "/docs/meta-ads",
    icon: Facebook,
    color: "bg-blue-500",
  },
  {
    title: "TikTok Ads",
    description: "Link your TikTok Ads Manager to monitor your TikTok advertising campaigns.",
    href: "/docs/tiktok-ads",
    icon: Music2,
    color: "bg-slate-900",
  },
  {
    title: "Shopee Ads",
    description: "Connect your Shopee Seller Center to track your Shopee advertising performance.",
    href: "/docs/shopee-ads",
    icon: ShoppingBag,
    color: "bg-orange-500",
  },
];

const features = [
  "Real-time data synchronization",
  "Unified dashboard for all platforms",
  "Automatic ROAS and CPA calculations",
  "Cross-platform performance comparison",
  "Secure OAuth 2.0 authentication",
  "No manual data entry required",
];

export default function DocsPage() {
  return (
    <div className="space-y-12">
      {/* Hero */}
      <div>
        <h1 className="text-3xl font-bold text-slate-900 mb-4">
          Welcome to AdsAnalytic Documentation
        </h1>
        <p className="text-lg text-slate-600 max-w-3xl">
          Learn how to connect your advertising accounts and get the most out of AdsAnalytic.
          Follow our step-by-step guides to link your Meta, TikTok, and Shopee ad accounts.
        </p>
      </div>

      {/* Quick Start */}
      <div>
        <h2 className="text-xl font-semibold text-slate-900 mb-4">Quick Start</h2>
        <div className="bg-blue-50 border border-blue-100 rounded-lg p-6">
          <h3 className="font-medium text-blue-900 mb-2">Getting Started in 3 Steps</h3>
          <ol className="space-y-2 text-blue-800">
            <li className="flex items-start gap-2">
              <span className="flex-shrink-0 w-6 h-6 bg-blue-600 text-white rounded-full flex items-center justify-center text-sm font-medium">1</span>
              <span><strong>Create an account</strong> - Sign up for free at AdsAnalytic</span>
            </li>
            <li className="flex items-start gap-2">
              <span className="flex-shrink-0 w-6 h-6 bg-blue-600 text-white rounded-full flex items-center justify-center text-sm font-medium">2</span>
              <span><strong>Connect your ad platforms</strong> - Follow the guides below to link your accounts</span>
            </li>
            <li className="flex items-start gap-2">
              <span className="flex-shrink-0 w-6 h-6 bg-blue-600 text-white rounded-full flex items-center justify-center text-sm font-medium">3</span>
              <span><strong>View your unified dashboard</strong> - See all your ad data in one place</span>
            </li>
          </ol>
        </div>
      </div>

      {/* Platform Guides */}
      <div>
        <h2 className="text-xl font-semibold text-slate-900 mb-4">Platform Connection Guides</h2>
        <div className="grid gap-4 md:grid-cols-3">
          {platforms.map((platform) => (
            <Link
              key={platform.href}
              href={platform.href}
              className="group block p-6 bg-slate-50 rounded-lg border border-slate-200 hover:border-blue-300 hover:bg-blue-50 transition-colors"
            >
              <div className={`w-12 h-12 ${platform.color} rounded-lg flex items-center justify-center mb-4`}>
                <platform.icon className="w-6 h-6 text-white" />
              </div>
              <h3 className="font-semibold text-slate-900 mb-2 group-hover:text-blue-700">
                {platform.title}
              </h3>
              <p className="text-sm text-slate-600 mb-4">
                {platform.description}
              </p>
              <span className="inline-flex items-center text-sm text-blue-600 font-medium">
                View Guide <ArrowRight className="w-4 h-4 ml-1 group-hover:translate-x-1 transition-transform" />
              </span>
            </Link>
          ))}
        </div>
      </div>

      {/* Features */}
      <div>
        <h2 className="text-xl font-semibold text-slate-900 mb-4">What You Get</h2>
        <div className="grid gap-3 sm:grid-cols-2">
          {features.map((feature) => (
            <div key={feature} className="flex items-center gap-2 text-slate-700">
              <CheckCircle2 className="w-5 h-5 text-green-500 flex-shrink-0" />
              <span>{feature}</span>
            </div>
          ))}
        </div>
      </div>

      {/* CTA */}
      <div className="bg-gradient-to-r from-blue-600 to-blue-700 rounded-lg p-8 text-white">
        <h2 className="text-2xl font-bold mb-2">Ready to get started?</h2>
        <p className="text-blue-100 mb-6">
          Create your free account and start tracking all your ad performance in one place.
        </p>
        <Link
          href="/register"
          className="inline-flex items-center px-6 py-3 bg-white text-blue-600 font-semibold rounded-lg hover:bg-blue-50 transition-colors"
        >
          Create Free Account <ArrowRight className="w-4 h-4 ml-2" />
        </Link>
      </div>
    </div>
  );
}
