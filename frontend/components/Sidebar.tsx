'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'
import {
  LayoutDashboard,
  Megaphone,
  Link2,
  Settings,
  BarChart3
} from 'lucide-react'
import { cn } from '@/lib/utils'

const navigation = [
  { name: 'Dashboard', href: '/dashboard', icon: LayoutDashboard },
  { name: 'Campaigns', href: '/campaigns', icon: Megaphone },
  { name: 'Analytics', href: '/analytics', icon: BarChart3 },
  { name: 'Connect', href: '/connect', icon: Link2 },
  { name: 'Settings', href: '/settings', icon: Settings },
]

export default function Sidebar() {
  const pathname = usePathname()

  return (
    <div className="hidden lg:fixed lg:inset-y-0 lg:z-50 lg:flex lg:w-72 lg:flex-col">
      <div className="flex grow flex-col gap-y-5 overflow-y-auto bg-gradient-to-b from-slate-900 to-slate-800 px-6 pb-4 border-r border-slate-700/50">
        {/* Logo */}
        <div className="flex h-16 shrink-0 items-center">
          <div className="flex items-center gap-3">
            <div className="h-10 w-10 rounded-xl bg-gradient-to-br from-indigo-500 to-purple-600 flex items-center justify-center shadow-lg shadow-indigo-500/25">
              <BarChart3 className="h-5 w-5 text-white" />
            </div>
            <div>
              <span className="text-xl font-bold bg-gradient-to-r from-indigo-400 to-purple-400 bg-clip-text text-transparent">
                AdsAnalytics
              </span>
              <p className="text-[10px] text-slate-500 tracking-wider uppercase">Multi-Platform</p>
            </div>
          </div>
        </div>

        {/* Navigation */}
        <nav className="flex flex-1 flex-col">
          <ul role="list" className="flex flex-1 flex-col gap-y-7">
            <li>
              <ul role="list" className="-mx-2 space-y-1">
                {navigation.map((item) => {
                  const isActive = pathname === item.href ||
                    (item.href === '/dashboard' && pathname === '/')
                  return (
                    <li key={item.name}>
                      <Link
                        href={item.href}
                        className={cn(
                          'group flex gap-x-3 rounded-lg p-3 text-sm leading-6 font-medium transition-all duration-200',
                          isActive
                            ? 'bg-gradient-to-r from-indigo-600/20 to-purple-600/20 text-white border border-indigo-500/30'
                            : 'text-slate-400 hover:text-white hover:bg-slate-800/50'
                        )}
                      >
                        <item.icon
                          className={cn(
                            'h-5 w-5 shrink-0 transition-colors duration-200',
                            isActive
                              ? 'text-indigo-400'
                              : 'text-slate-500 group-hover:text-indigo-400'
                          )}
                        />
                        {item.name}
                        {isActive && (
                          <span className="ml-auto w-1.5 h-1.5 rounded-full bg-indigo-400" />
                        )}
                      </Link>
                    </li>
                  )
                })}
              </ul>
            </li>

            {/* Connected Platforms Card */}
            <li className="mt-auto">
              <div className="rounded-xl bg-gradient-to-br from-slate-800/80 to-slate-800/40 p-4 border border-slate-700/50">
                <p className="text-xs font-medium text-slate-400 uppercase tracking-wider">Connected Platforms</p>
                <div className="mt-3 flex gap-2">
                  <div className="h-9 w-9 rounded-lg bg-blue-600 flex items-center justify-center text-xs text-white font-bold shadow-lg shadow-blue-600/25">
                    M
                  </div>
                  <div className="h-9 w-9 rounded-lg bg-black flex items-center justify-center text-xs text-white font-bold border border-slate-700">
                    T
                  </div>
                  <div className="h-9 w-9 rounded-lg bg-orange-500 flex items-center justify-center text-xs text-white font-bold shadow-lg shadow-orange-500/25">
                    S
                  </div>
                </div>
                <Link
                  href="/connect"
                  className="mt-3 block text-xs text-indigo-400 hover:text-indigo-300 transition-colors"
                >
                  Manage connections →
                </Link>
              </div>
            </li>
          </ul>
        </nav>
      </div>
    </div>
  )
}
