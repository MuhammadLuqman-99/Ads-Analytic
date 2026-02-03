"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { BarChart3, BookOpen, ChevronRight, Home } from "lucide-react";
import { MetaIcon, TikTokIcon, ShopeeIcon2 } from "@/components/icons";
import { cn } from "@/lib/utils";
import { ComponentType, SVGProps } from "react";

interface DocsLayoutProps {
  children: React.ReactNode;
}

interface NavItem {
  title: string;
  href: string;
  icon: ComponentType<{ className?: string }>;
}

interface NavSection {
  title: string;
  items: NavItem[];
}

const navigation: NavSection[] = [
  {
    title: "Getting Started",
    items: [
      { title: "Overview", href: "/docs", icon: BookOpen },
    ],
  },
  {
    title: "Platform Guides",
    items: [
      { title: "Connect Meta Ads", href: "/docs/meta-ads", icon: MetaIcon },
      { title: "Connect TikTok Ads", href: "/docs/tiktok-ads", icon: TikTokIcon },
      { title: "Connect Shopee Ads", href: "/docs/shopee-ads", icon: ShopeeIcon2 },
    ],
  },
];

export function DocsLayout({ children }: DocsLayoutProps) {
  const pathname = usePathname();

  return (
    <div className="min-h-screen bg-slate-50">
      {/* Header */}
      <header className="sticky top-0 z-50 bg-white border-b border-slate-200">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex items-center justify-between h-16">
            <Link href="/" className="flex items-center gap-3">
              <div className="w-9 h-9 bg-gradient-to-br from-blue-600 to-blue-700 rounded-lg flex items-center justify-center">
                <BarChart3 className="w-5 h-5 text-white" />
              </div>
              <span className="text-lg font-bold text-slate-900">AdsAnalytic</span>
            </Link>
            <nav className="flex items-center gap-4">
              <Link
                href="/"
                className="text-sm text-slate-600 hover:text-slate-900 flex items-center gap-1"
              >
                <Home className="w-4 h-4" />
                Home
              </Link>
              <Link
                href="/login"
                className="text-sm bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 transition-colors"
              >
                Sign In
              </Link>
            </nav>
          </div>
        </div>
      </header>

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="flex gap-8">
          {/* Sidebar */}
          <aside className="hidden lg:block w-64 flex-shrink-0">
            <nav className="sticky top-24 space-y-6">
              {navigation.map((section) => (
                <div key={section.title}>
                  <h3 className="text-xs font-semibold text-slate-500 uppercase tracking-wider mb-3">
                    {section.title}
                  </h3>
                  <ul className="space-y-1">
                    {section.items.map((item) => {
                      const isActive = pathname === item.href;
                      return (
                        <li key={item.href}>
                          <Link
                            href={item.href}
                            className={cn(
                              "flex items-center gap-2 px-3 py-2 text-sm rounded-lg transition-colors",
                              isActive
                                ? "bg-blue-50 text-blue-700 font-medium"
                                : "text-slate-600 hover:bg-slate-100 hover:text-slate-900"
                            )}
                          >
                            <item.icon className="w-4 h-4" />
                            {item.title}
                          </Link>
                        </li>
                      );
                    })}
                  </ul>
                </div>
              ))}
            </nav>
          </aside>

          {/* Main Content */}
          <main className="flex-1 min-w-0">
            {/* Breadcrumb */}
            <nav className="flex items-center gap-2 text-sm text-slate-500 mb-6">
              <Link href="/docs" className="hover:text-slate-700">
                Docs
              </Link>
              {pathname !== "/docs" && (
                <>
                  <ChevronRight className="w-4 h-4" />
                  <span className="text-slate-900">
                    {navigation
                      .flatMap((s) => s.items)
                      .find((i) => i.href === pathname)?.title || "Guide"}
                  </span>
                </>
              )}
            </nav>

            {/* Content */}
            <div className="bg-white rounded-xl border border-slate-200 p-8">
              {children}
            </div>
          </main>
        </div>
      </div>

      {/* Footer */}
      <footer className="border-t border-slate-200 mt-16">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          <p className="text-center text-sm text-slate-500">
            &copy; {new Date().getFullYear()} AdsAnalytic. All rights reserved.
          </p>
        </div>
      </footer>
    </div>
  );
}
