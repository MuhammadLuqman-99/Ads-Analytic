import Sidebar from "@/components/Sidebar";
import MetricCard from "@/components/MetricCard";
import PlatformCard from "@/components/PlatformCard";

function DollarIcon({ className }: { className?: string }) {
  return (
    <svg className={className} fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor">
      <path strokeLinecap="round" strokeLinejoin="round" d="M12 6v12m-3-2.818l.879.659c1.171.879 3.07.879 4.242 0 1.172-.879 1.172-2.303 0-3.182C13.536 12.219 12.768 12 12 12c-.725 0-1.45-.22-2.003-.659-1.106-.879-1.106-2.303 0-3.182s2.9-.879 4.006 0l.415.33M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
    </svg>
  )
}

function EyeIcon({ className }: { className?: string }) {
  return (
    <svg className={className} fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor">
      <path strokeLinecap="round" strokeLinejoin="round" d="M2.036 12.322a1.012 1.012 0 010-.639C3.423 7.51 7.36 4.5 12 4.5c4.638 0 8.573 3.007 9.963 7.178.07.207.07.431 0 .639C20.577 16.49 16.64 19.5 12 19.5c-4.638 0-8.573-3.007-9.963-7.178z" />
      <path strokeLinecap="round" strokeLinejoin="round" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
    </svg>
  )
}

function CursorIcon({ className }: { className?: string }) {
  return (
    <svg className={className} fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor">
      <path strokeLinecap="round" strokeLinejoin="round" d="M15.042 21.672L13.684 16.6m0 0l-2.51 2.225.569-9.47 5.227 7.917-3.286-.672zM12 2.25V4.5m5.834.166l-1.591 1.591M20.25 10.5H18M7.757 14.743l-1.59 1.59M6 10.5H3.75m4.007-4.243l-1.59-1.59" />
    </svg>
  )
}

function ShoppingCartIcon({ className }: { className?: string }) {
  return (
    <svg className={className} fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor">
      <path strokeLinecap="round" strokeLinejoin="round" d="M2.25 3h1.386c.51 0 .955.343 1.087.835l.383 1.437M7.5 14.25a3 3 0 00-3 3h15.75m-12.75-3h11.218c1.121-2.3 2.1-4.684 2.924-7.138a60.114 60.114 0 00-16.536-1.84M7.5 14.25L5.106 5.272M6 20.25a.75.75 0 11-1.5 0 .75.75 0 011.5 0zm12.75 0a.75.75 0 11-1.5 0 .75.75 0 011.5 0z" />
    </svg>
  )
}

export default function Dashboard() {
  return (
    <div className="min-h-screen bg-slate-900">
      <Sidebar />

      <main className="lg:pl-72">
        <div className="px-4 sm:px-6 lg:px-8 py-8">
          {/* Header */}
          <div className="flex items-center justify-between mb-8">
            <div>
              <h1 className="text-3xl font-bold text-white">Dashboard</h1>
              <p className="mt-1 text-slate-400">Overview of your ad performance across all platforms</p>
            </div>
            <div className="flex items-center gap-3">
              <select className="bg-slate-800 border border-slate-700 text-white rounded-lg px-4 py-2 text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent">
                <option>Last 7 days</option>
                <option>Last 30 days</option>
                <option>Last 90 days</option>
                <option>This month</option>
                <option>Last month</option>
              </select>
              <button className="bg-indigo-600 hover:bg-indigo-500 text-white rounded-lg px-4 py-2 text-sm font-medium transition-colors">
                Sync Now
              </button>
            </div>
          </div>

          {/* Metrics Grid */}
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
            <MetricCard
              title="Total Spend"
              value="RM 24,536.00"
              change="+12.5%"
              changeType="negative"
              icon={<DollarIcon className="h-6 w-6 text-indigo-400" />}
            />
            <MetricCard
              title="Total Impressions"
              value="1.2M"
              change="+8.2%"
              changeType="positive"
              icon={<EyeIcon className="h-6 w-6 text-purple-400" />}
            />
            <MetricCard
              title="Total Clicks"
              value="45,231"
              change="+15.3%"
              changeType="positive"
              icon={<CursorIcon className="h-6 w-6 text-emerald-400" />}
            />
            <MetricCard
              title="Conversions"
              value="1,847"
              change="+22.1%"
              changeType="positive"
              icon={<ShoppingCartIcon className="h-6 w-6 text-amber-400" />}
            />
          </div>

          {/* Platform Performance */}
          <div className="mb-8">
            <h2 className="text-xl font-semibold text-white mb-4">Platform Performance</h2>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
              <PlatformCard
                name="My Business"
                platform="meta"
                spend="RM 12,450"
                impressions="650K"
                clicks="23,105"
                conversions="892"
                roas="3.2"
                connected={true}
              />
              <PlatformCard
                name="TikTok Shop"
                platform="tiktok"
                spend="RM 8,234"
                impressions="420K"
                clicks="15,432"
                conversions="645"
                roas="2.8"
                connected={true}
              />
              <PlatformCard
                name="Shopee Store"
                platform="shopee"
                spend="RM 3,852"
                impressions="130K"
                clicks="6,694"
                conversions="310"
                roas="4.1"
                connected={true}
              />
            </div>
          </div>

          {/* Recent Campaigns */}
          <div className="rounded-xl bg-slate-800/50 border border-slate-700/50 overflow-hidden">
            <div className="p-6 border-b border-slate-700/50">
              <h2 className="text-xl font-semibold text-white">Recent Campaigns</h2>
            </div>
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead className="bg-slate-800/50">
                  <tr>
                    <th className="text-left text-xs font-medium text-slate-400 uppercase tracking-wider px-6 py-3">Campaign</th>
                    <th className="text-left text-xs font-medium text-slate-400 uppercase tracking-wider px-6 py-3">Platform</th>
                    <th className="text-left text-xs font-medium text-slate-400 uppercase tracking-wider px-6 py-3">Status</th>
                    <th className="text-right text-xs font-medium text-slate-400 uppercase tracking-wider px-6 py-3">Spend</th>
                    <th className="text-right text-xs font-medium text-slate-400 uppercase tracking-wider px-6 py-3">Impressions</th>
                    <th className="text-right text-xs font-medium text-slate-400 uppercase tracking-wider px-6 py-3">Clicks</th>
                    <th className="text-right text-xs font-medium text-slate-400 uppercase tracking-wider px-6 py-3">ROAS</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-700/50">
                  <tr className="hover:bg-slate-700/20 transition-colors">
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="text-sm font-medium text-white">Summer Sale 2024</div>
                      <div className="text-xs text-slate-400">ID: 123456789</div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <span className="inline-flex items-center gap-1 px-2 py-1 rounded-full text-xs font-medium bg-blue-500/20 text-blue-400">
                        Meta
                      </span>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <span className="inline-flex items-center gap-1 px-2 py-1 rounded-full text-xs font-medium bg-emerald-500/20 text-emerald-400">
                        Active
                      </span>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-right text-sm text-white">RM 4,520</td>
                    <td className="px-6 py-4 whitespace-nowrap text-right text-sm text-slate-300">245K</td>
                    <td className="px-6 py-4 whitespace-nowrap text-right text-sm text-slate-300">8,432</td>
                    <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium text-emerald-400">3.5x</td>
                  </tr>
                  <tr className="hover:bg-slate-700/20 transition-colors">
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="text-sm font-medium text-white">Product Launch - Series X</div>
                      <div className="text-xs text-slate-400">ID: 987654321</div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <span className="inline-flex items-center gap-1 px-2 py-1 rounded-full text-xs font-medium bg-slate-500/20 text-slate-300">
                        TikTok
                      </span>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <span className="inline-flex items-center gap-1 px-2 py-1 rounded-full text-xs font-medium bg-emerald-500/20 text-emerald-400">
                        Active
                      </span>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-right text-sm text-white">RM 3,210</td>
                    <td className="px-6 py-4 whitespace-nowrap text-right text-sm text-slate-300">180K</td>
                    <td className="px-6 py-4 whitespace-nowrap text-right text-sm text-slate-300">5,892</td>
                    <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium text-emerald-400">2.9x</td>
                  </tr>
                  <tr className="hover:bg-slate-700/20 transition-colors">
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="text-sm font-medium text-white">Flash Sale Weekend</div>
                      <div className="text-xs text-slate-400">ID: 456789123</div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <span className="inline-flex items-center gap-1 px-2 py-1 rounded-full text-xs font-medium bg-orange-500/20 text-orange-400">
                        Shopee
                      </span>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <span className="inline-flex items-center gap-1 px-2 py-1 rounded-full text-xs font-medium bg-amber-500/20 text-amber-400">
                        Paused
                      </span>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-right text-sm text-white">RM 1,890</td>
                    <td className="px-6 py-4 whitespace-nowrap text-right text-sm text-slate-300">85K</td>
                    <td className="px-6 py-4 whitespace-nowrap text-right text-sm text-slate-300">3,245</td>
                    <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium text-emerald-400">4.2x</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>
      </main>
    </div>
  )
}
