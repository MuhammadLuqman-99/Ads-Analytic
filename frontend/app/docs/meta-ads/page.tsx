import Link from "next/link";
import { AlertCircle, CheckCircle2, ExternalLink } from "lucide-react";

export const metadata = {
  title: "Connect Meta (Facebook) Ads - AdsAnalytic Documentation",
  description: "Step-by-step guide to connect your Meta (Facebook/Instagram) Ads account to AdsAnalytic.",
};

export default function MetaAdsPage() {
  return (
    <div className="space-y-8">
      {/* Header */}
      <div>
        <h1 className="text-3xl font-bold text-slate-900 mb-4">
          Connect Meta (Facebook) Ads
        </h1>
        <p className="text-lg text-slate-600">
          Follow this guide to connect your Meta Business Suite and start tracking your Facebook and Instagram ad campaigns.
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
            <span>You need a <strong>Meta Business Suite</strong> account</span>
          </li>
          <li className="flex items-start gap-2">
            <CheckCircle2 className="w-5 h-5 text-amber-600 flex-shrink-0 mt-0.5" />
            <span>You must have <strong>Admin access</strong> to your Ad Account</span>
          </li>
          <li className="flex items-start gap-2">
            <CheckCircle2 className="w-5 h-5 text-amber-600 flex-shrink-0 mt-0.5" />
            <span>Your business must be <strong>verified</strong> on Meta Business Suite</span>
          </li>
        </ul>
      </div>

      {/* Steps */}
      <div className="space-y-6">
        <h2 className="text-xl font-semibold text-slate-900">Step-by-Step Guide</h2>

        {/* Step 1 */}
        <div className="border border-slate-200 rounded-lg overflow-hidden">
          <div className="bg-slate-50 px-6 py-4 border-b border-slate-200">
            <h3 className="font-semibold text-slate-900">Step 1: Access Meta Business Suite</h3>
          </div>
          <div className="p-6 space-y-4">
            <ol className="list-decimal list-inside space-y-2 text-slate-700">
              <li>Go to <a href="https://business.facebook.com" target="_blank" rel="noopener noreferrer" className="text-blue-600 hover:underline inline-flex items-center gap-1">Meta Business Suite <ExternalLink className="w-3 h-3" /></a></li>
              <li>Sign in with your Facebook account that has admin access</li>
              <li>Select your Business Account from the dropdown menu</li>
            </ol>
            <div className="bg-slate-100 rounded-lg p-4 text-sm text-slate-600">
              <strong>Tip:</strong> If you don&apos;t have a Business Suite account, you can create one at{" "}
              <a href="https://business.facebook.com/overview" target="_blank" rel="noopener noreferrer" className="text-blue-600 hover:underline">
                business.facebook.com/overview
              </a>
            </div>
          </div>
        </div>

        {/* Step 2 */}
        <div className="border border-slate-200 rounded-lg overflow-hidden">
          <div className="bg-slate-50 px-6 py-4 border-b border-slate-200">
            <h3 className="font-semibold text-slate-900">Step 2: Navigate to Business Settings</h3>
          </div>
          <div className="p-6 space-y-4">
            <ol className="list-decimal list-inside space-y-2 text-slate-700">
              <li>Click on the <strong>gear icon</strong> (Settings) in the bottom left corner</li>
              <li>Select <strong>&quot;Business Settings&quot;</strong></li>
              <li>In the left sidebar, click on <strong>&quot;Users&quot;</strong> &rarr; <strong>&quot;System Users&quot;</strong></li>
            </ol>
          </div>
        </div>

        {/* Step 3 */}
        <div className="border border-slate-200 rounded-lg overflow-hidden">
          <div className="bg-slate-50 px-6 py-4 border-b border-slate-200">
            <h3 className="font-semibold text-slate-900">Step 3: Create a System User</h3>
          </div>
          <div className="p-6 space-y-4">
            <ol className="list-decimal list-inside space-y-2 text-slate-700">
              <li>Click the <strong>&quot;Add&quot;</strong> button</li>
              <li>Enter a name for your system user (e.g., &quot;AdsAnalytic Integration&quot;)</li>
              <li>Set the role to <strong>&quot;Admin&quot;</strong></li>
              <li>Click <strong>&quot;Create System User&quot;</strong></li>
            </ol>
            <div className="bg-blue-50 border border-blue-200 rounded-lg p-4 text-sm text-blue-800">
              <strong>Why System User?</strong> System users provide a secure way to access your ad data without using personal account credentials.
            </div>
          </div>
        </div>

        {/* Step 4 */}
        <div className="border border-slate-200 rounded-lg overflow-hidden">
          <div className="bg-slate-50 px-6 py-4 border-b border-slate-200">
            <h3 className="font-semibold text-slate-900">Step 4: Assign Ad Account Access</h3>
          </div>
          <div className="p-6 space-y-4">
            <ol className="list-decimal list-inside space-y-2 text-slate-700">
              <li>Click on your newly created system user</li>
              <li>Click <strong>&quot;Add Assets&quot;</strong></li>
              <li>Select <strong>&quot;Ad Accounts&quot;</strong></li>
              <li>Choose the ad accounts you want to connect</li>
              <li>Set permissions to <strong>&quot;Manage campaigns&quot;</strong> (or higher)</li>
              <li>Click <strong>&quot;Save Changes&quot;</strong></li>
            </ol>
          </div>
        </div>

        {/* Step 5 */}
        <div className="border border-slate-200 rounded-lg overflow-hidden">
          <div className="bg-slate-50 px-6 py-4 border-b border-slate-200">
            <h3 className="font-semibold text-slate-900">Step 5: Generate Access Token</h3>
          </div>
          <div className="p-6 space-y-4">
            <ol className="list-decimal list-inside space-y-2 text-slate-700">
              <li>Click on your system user again</li>
              <li>Click <strong>&quot;Generate New Token&quot;</strong></li>
              <li>Select the app (or create one if needed)</li>
              <li>Choose token expiration: <strong>&quot;Never&quot;</strong> is recommended</li>
              <li>Select the following permissions:
                <ul className="list-disc list-inside ml-4 mt-2 space-y-1 text-slate-600">
                  <li>ads_read</li>
                  <li>ads_management</li>
                  <li>business_management</li>
                  <li>read_insights</li>
                </ul>
              </li>
              <li>Click <strong>&quot;Generate Token&quot;</strong></li>
              <li><strong>Copy and save your token securely</strong> - you won&apos;t be able to see it again!</li>
            </ol>
            <div className="bg-red-50 border border-red-200 rounded-lg p-4 text-sm text-red-800">
              <strong>Important:</strong> Keep your access token secure. Never share it publicly or commit it to version control.
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
              <li>Select <strong>&quot;Meta (Facebook)&quot;</strong></li>
              <li>Paste your access token</li>
              <li>Click <strong>&quot;Connect&quot;</strong></li>
            </ol>
            <div className="bg-green-50 border border-green-200 rounded-lg p-4 text-sm text-green-800">
              <strong>Success!</strong> Once connected, AdsAnalytic will automatically sync your ad data every hour.
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
              I don&apos;t see my Ad Account in the list
            </summary>
            <div className="px-6 pb-4 text-slate-600">
              Make sure you have admin access to the ad account. You may need to request access from your business administrator.
            </div>
          </details>
          <details className="group border border-slate-200 rounded-lg">
            <summary className="px-6 py-4 cursor-pointer font-medium text-slate-900 hover:bg-slate-50">
              My token expired
            </summary>
            <div className="px-6 pb-4 text-slate-600">
              Generate a new token following Step 5 above. When creating the token, select &quot;Never&quot; for expiration to avoid this issue in the future.
            </div>
          </details>
          <details className="group border border-slate-200 rounded-lg">
            <summary className="px-6 py-4 cursor-pointer font-medium text-slate-900 hover:bg-slate-50">
              Data is not syncing
            </summary>
            <div className="px-6 pb-4 text-slate-600">
              Check that your token has all the required permissions (ads_read, ads_management, business_management, read_insights). You may need to regenerate the token with the correct permissions.
            </div>
          </details>
        </div>
      </div>

      {/* Next Steps */}
      <div className="flex items-center justify-between pt-6 border-t border-slate-200">
        <Link
          href="/docs"
          className="text-slate-600 hover:text-slate-900"
        >
          &larr; Back to Docs
        </Link>
        <Link
          href="/docs/tiktok-ads"
          className="text-blue-600 hover:text-blue-700 font-medium"
        >
          Next: Connect TikTok Ads &rarr;
        </Link>
      </div>
    </div>
  );
}
