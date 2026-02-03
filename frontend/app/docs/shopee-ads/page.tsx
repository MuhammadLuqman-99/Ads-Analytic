import Link from "next/link";
import { AlertCircle, CheckCircle2, ExternalLink } from "lucide-react";

export const metadata = {
  title: "Connect Shopee Ads - AdsAnalytic Documentation",
  description: "Step-by-step guide to connect your Shopee Seller Center and Ads account to AdsAnalytic.",
};

export default function ShopeeAdsPage() {
  return (
    <div className="space-y-8">
      {/* Header */}
      <div>
        <h1 className="text-3xl font-bold text-slate-900 mb-4">
          Connect Shopee Ads
        </h1>
        <p className="text-lg text-slate-600">
          Follow this guide to connect your Shopee Seller Center and start tracking your Shopee advertising campaigns.
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
            <span>You need an active <strong>Shopee Seller Center</strong> account</span>
          </li>
          <li className="flex items-start gap-2">
            <CheckCircle2 className="w-5 h-5 text-amber-600 flex-shrink-0 mt-0.5" />
            <span>Your shop must be <strong>verified</strong> and in good standing</span>
          </li>
          <li className="flex items-start gap-2">
            <CheckCircle2 className="w-5 h-5 text-amber-600 flex-shrink-0 mt-0.5" />
            <span>You must have <strong>Shopee Ads</strong> enabled on your account</span>
          </li>
        </ul>
      </div>

      {/* Region Notice */}
      <div className="bg-blue-50 border border-blue-200 rounded-lg p-6">
        <h2 className="flex items-center gap-2 text-lg font-semibold text-blue-900 mb-3">
          Supported Regions
        </h2>
        <p className="text-blue-800 mb-3">
          AdsAnalytic supports Shopee accounts from the following regions:
        </p>
        <div className="grid grid-cols-2 md:grid-cols-4 gap-2">
          {["Malaysia (MY)", "Singapore (SG)", "Indonesia (ID)", "Thailand (TH)", "Vietnam (VN)", "Philippines (PH)", "Taiwan (TW)", "Brazil (BR)"].map((region) => (
            <div key={region} className="bg-white px-3 py-2 rounded border border-blue-100 text-sm text-blue-800">
              {region}
            </div>
          ))}
        </div>
      </div>

      {/* Steps */}
      <div className="space-y-6">
        <h2 className="text-xl font-semibold text-slate-900">Step-by-Step Guide</h2>

        {/* Step 1 */}
        <div className="border border-slate-200 rounded-lg overflow-hidden">
          <div className="bg-slate-50 px-6 py-4 border-b border-slate-200">
            <h3 className="font-semibold text-slate-900">Step 1: Access Shopee Open Platform</h3>
          </div>
          <div className="p-6 space-y-4">
            <ol className="list-decimal list-inside space-y-2 text-slate-700">
              <li>Go to <a href="https://open.shopee.com" target="_blank" rel="noopener noreferrer" className="text-blue-600 hover:underline inline-flex items-center gap-1">Shopee Open Platform <ExternalLink className="w-3 h-3" /></a></li>
              <li>Click <strong>&quot;Log in&quot;</strong> with your Shopee Seller account</li>
              <li>Navigate to <strong>&quot;My Apps&quot;</strong> in the dashboard</li>
            </ol>
            <div className="bg-slate-100 rounded-lg p-4 text-sm text-slate-600">
              <strong>Note:</strong> You must use your Shopee Seller account, not a regular buyer account. The email should match your Seller Center login.
            </div>
          </div>
        </div>

        {/* Step 2 */}
        <div className="border border-slate-200 rounded-lg overflow-hidden">
          <div className="bg-slate-50 px-6 py-4 border-b border-slate-200">
            <h3 className="font-semibold text-slate-900">Step 2: Create an App</h3>
          </div>
          <div className="p-6 space-y-4">
            <ol className="list-decimal list-inside space-y-2 text-slate-700">
              <li>Click <strong>&quot;Create App&quot;</strong></li>
              <li>Fill in the app details:
                <ul className="list-disc list-inside ml-4 mt-2 space-y-1 text-slate-600">
                  <li>App Name: &quot;AdsAnalytic Integration&quot;</li>
                  <li>App Type: Select <strong>&quot;Private App&quot;</strong></li>
                  <li>Description: Brief description of your use case</li>
                </ul>
              </li>
              <li>Click <strong>&quot;Submit&quot;</strong></li>
            </ol>
            <div className="bg-blue-50 border border-blue-200 rounded-lg p-4 text-sm text-blue-800">
              <strong>Why Private App?</strong> Private apps are for your own shop integration and don&apos;t require the approval process needed for public apps.
            </div>
          </div>
        </div>

        {/* Step 3 */}
        <div className="border border-slate-200 rounded-lg overflow-hidden">
          <div className="bg-slate-50 px-6 py-4 border-b border-slate-200">
            <h3 className="font-semibold text-slate-900">Step 3: Get Your API Credentials</h3>
          </div>
          <div className="p-6 space-y-4">
            <ol className="list-decimal list-inside space-y-2 text-slate-700">
              <li>Once your app is created, click on it to view details</li>
              <li>Go to the <strong>&quot;App Credentials&quot;</strong> section</li>
              <li>Copy your <strong>Partner ID</strong> and <strong>Partner Key</strong></li>
              <li>Save these credentials securely</li>
            </ol>
            <div className="bg-red-50 border border-red-200 rounded-lg p-4 text-sm text-red-800">
              <strong>Important:</strong> Keep your Partner Key secure. Never share it publicly or commit it to version control.
            </div>
          </div>
        </div>

        {/* Step 4 */}
        <div className="border border-slate-200 rounded-lg overflow-hidden">
          <div className="bg-slate-50 px-6 py-4 border-b border-slate-200">
            <h3 className="font-semibold text-slate-900">Step 4: Configure API Permissions</h3>
          </div>
          <div className="p-6 space-y-4">
            <ol className="list-decimal list-inside space-y-2 text-slate-700">
              <li>In your app settings, go to <strong>&quot;API List&quot;</strong></li>
              <li>Enable the following API categories:
                <ul className="list-disc list-inside ml-4 mt-2 space-y-1 text-slate-600">
                  <li><strong>Shop API</strong> - For shop information</li>
                  <li><strong>Ads API</strong> - For advertising data</li>
                  <li><strong>Analytics API</strong> - For performance metrics</li>
                </ul>
              </li>
              <li>Click <strong>&quot;Save&quot;</strong> to apply permissions</li>
            </ol>
          </div>
        </div>

        {/* Step 5 */}
        <div className="border border-slate-200 rounded-lg overflow-hidden">
          <div className="bg-slate-50 px-6 py-4 border-b border-slate-200">
            <h3 className="font-semibold text-slate-900">Step 5: Authorize Your Shop</h3>
          </div>
          <div className="p-6 space-y-4">
            <ol className="list-decimal list-inside space-y-2 text-slate-700">
              <li>Go to <strong>&quot;Shop Authorization&quot;</strong> tab</li>
              <li>Click <strong>&quot;Add Shop&quot;</strong></li>
              <li>You&apos;ll be redirected to Shopee to authorize access</li>
              <li>Log in with your Seller account and click <strong>&quot;Confirm Authorization&quot;</strong></li>
              <li>Copy the <strong>Shop ID</strong> and <strong>Access Token</strong> provided</li>
            </ol>
            <div className="bg-amber-50 border border-amber-200 rounded-lg p-4 text-sm text-amber-800">
              <strong>Token Refresh:</strong> Shopee access tokens expire periodically. AdsAnalytic will automatically refresh your tokens to maintain the connection.
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
              <li>Select <strong>&quot;Shopee Ads&quot;</strong></li>
              <li>Select your <strong>Region</strong> (e.g., MY, SG, ID)</li>
              <li>Enter your Partner ID, Partner Key, Shop ID, and Access Token</li>
              <li>Click <strong>&quot;Connect&quot;</strong></li>
            </ol>
            <div className="bg-green-50 border border-green-200 rounded-lg p-4 text-sm text-green-800">
              <strong>Success!</strong> Once connected, AdsAnalytic will automatically sync your Shopee Ads data including campaign performance, product ads, and discovery ads.
            </div>
          </div>
        </div>
      </div>

      {/* Data Available */}
      <div className="space-y-4">
        <h2 className="text-xl font-semibold text-slate-900">Data We Sync</h2>
        <div className="grid gap-4 md:grid-cols-2">
          <div className="border border-slate-200 rounded-lg p-4">
            <h3 className="font-semibold text-slate-900 mb-2">Campaign Data</h3>
            <ul className="space-y-1 text-sm text-slate-600">
              <li>• Search Ads performance</li>
              <li>• Discovery Ads metrics</li>
              <li>• Product Boost campaigns</li>
              <li>• Shop Ads statistics</li>
            </ul>
          </div>
          <div className="border border-slate-200 rounded-lg p-4">
            <h3 className="font-semibold text-slate-900 mb-2">Metrics</h3>
            <ul className="space-y-1 text-sm text-slate-600">
              <li>• Impressions and clicks</li>
              <li>• Click-through rate (CTR)</li>
              <li>• Cost per click (CPC)</li>
              <li>• ROAS and conversions</li>
            </ul>
          </div>
        </div>
      </div>

      {/* Troubleshooting */}
      <div className="space-y-4">
        <h2 className="text-xl font-semibold text-slate-900">Troubleshooting</h2>
        <div className="space-y-4">
          <details className="group border border-slate-200 rounded-lg">
            <summary className="px-6 py-4 cursor-pointer font-medium text-slate-900 hover:bg-slate-50">
              I can&apos;t find the Open Platform option
            </summary>
            <div className="px-6 pb-4 text-slate-600">
              The Shopee Open Platform is only available for verified sellers. Make sure your shop is verified and in good standing. Some new shops may need to wait 30 days before accessing the API.
            </div>
          </details>
          <details className="group border border-slate-200 rounded-lg">
            <summary className="px-6 py-4 cursor-pointer font-medium text-slate-900 hover:bg-slate-50">
              Shop authorization failed
            </summary>
            <div className="px-6 pb-4 text-slate-600">
              Ensure you&apos;re logged into the correct Shopee Seller account. The account must be the shop owner or have admin permissions. Try logging out of all Shopee sessions and try again.
            </div>
          </details>
          <details className="group border border-slate-200 rounded-lg">
            <summary className="px-6 py-4 cursor-pointer font-medium text-slate-900 hover:bg-slate-50">
              Missing Ads API permission
            </summary>
            <div className="px-6 pb-4 text-slate-600">
              The Ads API may not be available for all regions or shop types. Check if your shop has Shopee Ads enabled in Seller Center. Contact Shopee support if you believe you should have access.
            </div>
          </details>
          <details className="group border border-slate-200 rounded-lg">
            <summary className="px-6 py-4 cursor-pointer font-medium text-slate-900 hover:bg-slate-50">
              Data not syncing
            </summary>
            <div className="px-6 pb-4 text-slate-600">
              Check that all required APIs are enabled (Shop, Ads, Analytics). Verify your access token hasn&apos;t expired. If issues persist, try disconnecting and reconnecting your Shopee account.
            </div>
          </details>
        </div>
      </div>

      {/* Next Steps */}
      <div className="flex items-center justify-between pt-6 border-t border-slate-200">
        <Link
          href="/docs/tiktok-ads"
          className="text-slate-600 hover:text-slate-900"
        >
          &larr; TikTok Ads Guide
        </Link>
        <Link
          href="/docs"
          className="text-blue-600 hover:text-blue-700 font-medium"
        >
          Back to Documentation &rarr;
        </Link>
      </div>
    </div>
  );
}
