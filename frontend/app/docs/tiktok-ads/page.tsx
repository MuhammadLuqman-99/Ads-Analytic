import Link from "next/link";
import { AlertCircle, CheckCircle2, ExternalLink } from "lucide-react";

export const metadata = {
  title: "Connect TikTok Ads - AdsAnalytic Documentation",
  description: "Step-by-step guide to connect your TikTok Ads Manager account to AdsAnalytic.",
};

export default function TikTokAdsPage() {
  return (
    <div className="space-y-8">
      {/* Header */}
      <div>
        <h1 className="text-3xl font-bold text-slate-900 mb-4">
          Connect TikTok Ads
        </h1>
        <p className="text-lg text-slate-600">
          Follow this guide to connect your TikTok Ads Manager and start tracking your TikTok advertising campaigns.
        </p>
      </div>

      {/* Prerequisites */}
      <div className="bg-amber-50 border border-amber-200 rounded-lg p-6">
        <h2 className="flex items-center gap-2 text-lg font-semibold text-amber-900 mb-3">
          <AlertCircle className="w-5 h-5" />
          Before You Begin
        </h2>
        <ul className="space-y-2 text-amber-800">
          <li className="flex items-start gap-2">
            <CheckCircle2 className="w-5 h-5 text-amber-600 flex-shrink-0 mt-0.5" />
            <span>You need a <strong>TikTok for Business</strong> account</span>
          </li>
          <li className="flex items-start gap-2">
            <CheckCircle2 className="w-5 h-5 text-amber-600 flex-shrink-0 mt-0.5" />
            <span>You must have <strong>Admin access</strong> to your TikTok Ads Manager</span>
          </li>
          <li className="flex items-start gap-2">
            <CheckCircle2 className="w-5 h-5 text-amber-600 flex-shrink-0 mt-0.5" />
            <span>Your TikTok Business Center must be <strong>verified</strong></span>
          </li>
        </ul>
      </div>

      {/* Steps */}
      <div className="space-y-6">
        <h2 className="text-xl font-semibold text-slate-900">Step-by-Step Guide</h2>

        {/* Step 1 */}
        <div className="border border-slate-200 rounded-lg overflow-hidden">
          <div className="bg-slate-50 px-6 py-4 border-b border-slate-200">
            <h3 className="font-semibold text-slate-900">Step 1: Access TikTok for Business</h3>
          </div>
          <div className="p-6 space-y-4">
            <ol className="list-decimal list-inside space-y-2 text-slate-700">
              <li>Go to <a href="https://business.tiktok.com" target="_blank" rel="noopener noreferrer" className="text-blue-600 hover:underline inline-flex items-center gap-1">TikTok Business Center <ExternalLink className="w-3 h-3" /></a></li>
              <li>Sign in with your TikTok account that has admin access</li>
              <li>Select your Business Center from the dropdown menu</li>
            </ol>
            <div className="bg-slate-100 rounded-lg p-4 text-sm text-slate-600">
              <strong>Tip:</strong> If you don&apos;t have a TikTok for Business account, you can create one at{" "}
              <a href="https://getstarted.tiktok.com/business" target="_blank" rel="noopener noreferrer" className="text-blue-600 hover:underline">
                getstarted.tiktok.com/business
              </a>
            </div>
          </div>
        </div>

        {/* Step 2 */}
        <div className="border border-slate-200 rounded-lg overflow-hidden">
          <div className="bg-slate-50 px-6 py-4 border-b border-slate-200">
            <h3 className="font-semibold text-slate-900">Step 2: Navigate to Developer Portal</h3>
          </div>
          <div className="p-6 space-y-4">
            <ol className="list-decimal list-inside space-y-2 text-slate-700">
              <li>Go to <a href="https://developers.tiktok.com" target="_blank" rel="noopener noreferrer" className="text-blue-600 hover:underline inline-flex items-center gap-1">TikTok Developer Portal <ExternalLink className="w-3 h-3" /></a></li>
              <li>Click <strong>&quot;Log in&quot;</strong> and use your TikTok Business account</li>
              <li>Navigate to <strong>&quot;My Apps&quot;</strong> section</li>
            </ol>
          </div>
        </div>

        {/* Step 3 */}
        <div className="border border-slate-200 rounded-lg overflow-hidden">
          <div className="bg-slate-50 px-6 py-4 border-b border-slate-200">
            <h3 className="font-semibold text-slate-900">Step 3: Create a Marketing API App</h3>
          </div>
          <div className="p-6 space-y-4">
            <ol className="list-decimal list-inside space-y-2 text-slate-700">
              <li>Click <strong>&quot;Create App&quot;</strong></li>
              <li>Select <strong>&quot;Marketing API&quot;</strong> as the app type</li>
              <li>Enter app details:
                <ul className="list-disc list-inside ml-4 mt-2 space-y-1 text-slate-600">
                  <li>App Name: &quot;AdsAnalytic Integration&quot;</li>
                  <li>App Description: Brief description of your use case</li>
                </ul>
              </li>
              <li>Click <strong>&quot;Submit for Review&quot;</strong></li>
            </ol>
            <div className="bg-blue-50 border border-blue-200 rounded-lg p-4 text-sm text-blue-800">
              <strong>Note:</strong> Marketing API apps require approval from TikTok. This process typically takes 1-3 business days.
            </div>
          </div>
        </div>

        {/* Step 4 */}
        <div className="border border-slate-200 rounded-lg overflow-hidden">
          <div className="bg-slate-50 px-6 py-4 border-b border-slate-200">
            <h3 className="font-semibold text-slate-900">Step 4: Get Your App Credentials</h3>
          </div>
          <div className="p-6 space-y-4">
            <ol className="list-decimal list-inside space-y-2 text-slate-700">
              <li>Once approved, go to your app in <strong>&quot;My Apps&quot;</strong></li>
              <li>Navigate to the <strong>&quot;App Info&quot;</strong> tab</li>
              <li>Copy your <strong>App ID</strong> and <strong>App Secret</strong></li>
              <li>Save these credentials securely</li>
            </ol>
            <div className="bg-red-50 border border-red-200 rounded-lg p-4 text-sm text-red-800">
              <strong>Important:</strong> Keep your App Secret secure. Never share it publicly or commit it to version control.
            </div>
          </div>
        </div>

        {/* Step 5 */}
        <div className="border border-slate-200 rounded-lg overflow-hidden">
          <div className="bg-slate-50 px-6 py-4 border-b border-slate-200">
            <h3 className="font-semibold text-slate-900">Step 5: Generate Access Token</h3>
          </div>
          <div className="p-6 space-y-4">
            <ol className="list-decimal list-inside space-y-2 text-slate-700">
              <li>In your app settings, go to <strong>&quot;Authorization&quot;</strong></li>
              <li>Click <strong>&quot;Generate Access Token&quot;</strong></li>
              <li>Select the Ad Accounts you want to connect</li>
              <li>Choose the following permissions:
                <ul className="list-disc list-inside ml-4 mt-2 space-y-1 text-slate-600">
                  <li>Ad Account Management</li>
                  <li>Ads Management</li>
                  <li>Ads Reporting</li>
                </ul>
              </li>
              <li>Click <strong>&quot;Confirm&quot;</strong> to generate your access token</li>
              <li><strong>Copy and save your access token</strong></li>
            </ol>
            <div className="bg-amber-50 border border-amber-200 rounded-lg p-4 text-sm text-amber-800">
              <strong>Token Expiry:</strong> TikTok access tokens expire after 24 hours. For long-term access, you&apos;ll need to use the refresh token to get new access tokens automatically.
            </div>
          </div>
        </div>

        {/* Step 6 */}
        <div className="border border-slate-200 rounded-lg overflow-hidden">
          <div className="bg-slate-50 px-6 py-4 border-b border-slate-200">
            <h3 className="font-semibold text-slate-900">Step 6: Connect in AdsAnalytic</h3>
          </div>
          <div className="p-6 space-y-4">
            <ol className="list-decimal list-inside space-y-2 text-slate-700">
              <li>Log in to your AdsAnalytic dashboard</li>
              <li>Go to <strong>Connections</strong> in the sidebar</li>
              <li>Click <strong>&quot;Connect Platform&quot;</strong></li>
              <li>Select <strong>&quot;TikTok Ads&quot;</strong></li>
              <li>Enter your App ID, App Secret, and Access Token</li>
              <li>Click <strong>&quot;Connect&quot;</strong></li>
            </ol>
            <div className="bg-green-50 border border-green-200 rounded-lg p-4 text-sm text-green-800">
              <strong>Success!</strong> Once connected, AdsAnalytic will automatically sync your TikTok ad data and handle token refresh automatically.
            </div>
          </div>
        </div>
      </div>

      {/* Troubleshooting */}
      <div className="space-y-4">
        <h2 className="text-xl font-semibold text-slate-900">Troubleshooting</h2>
        <div className="space-y-4">
          <details className="group border border-slate-200 rounded-lg">
            <summary className="px-6 py-4 cursor-pointer font-medium text-slate-900 hover:bg-slate-50">
              My app is stuck in review
            </summary>
            <div className="px-6 pb-4 text-slate-600">
              TikTok Marketing API approval typically takes 1-3 business days. Ensure your app description clearly explains your use case. If it&apos;s been longer than 5 business days, contact TikTok Developer Support.
            </div>
          </details>
          <details className="group border border-slate-200 rounded-lg">
            <summary className="px-6 py-4 cursor-pointer font-medium text-slate-900 hover:bg-slate-50">
              Access token expired
            </summary>
            <div className="px-6 pb-4 text-slate-600">
              TikTok access tokens expire after 24 hours. AdsAnalytic automatically handles token refresh using your refresh token. If you&apos;re seeing token errors, try disconnecting and reconnecting your account.
            </div>
          </details>
          <details className="group border border-slate-200 rounded-lg">
            <summary className="px-6 py-4 cursor-pointer font-medium text-slate-900 hover:bg-slate-50">
              I don&apos;t see all my Ad Accounts
            </summary>
            <div className="px-6 pb-4 text-slate-600">
              When generating your access token, make sure you selected all the Ad Accounts you want to connect. You can regenerate the token and select additional accounts.
            </div>
          </details>
          <details className="group border border-slate-200 rounded-lg">
            <summary className="px-6 py-4 cursor-pointer font-medium text-slate-900 hover:bg-slate-50">
              Permission denied errors
            </summary>
            <div className="px-6 pb-4 text-slate-600">
              Ensure your app has all required permissions (Ad Account Management, Ads Management, Ads Reporting). You may need to request additional permissions through the TikTok Developer Portal.
            </div>
          </details>
        </div>
      </div>

      {/* Next Steps */}
      <div className="flex items-center justify-between pt-6 border-t border-slate-200">
        <Link
          href="/docs/meta-ads"
          className="text-slate-600 hover:text-slate-900"
        >
          &larr; Meta Ads Guide
        </Link>
        <Link
          href="/docs/shopee-ads"
          className="text-blue-600 hover:text-blue-700 font-medium"
        >
          Next: Connect Shopee Ads &rarr;
        </Link>
      </div>
    </div>
  );
}
