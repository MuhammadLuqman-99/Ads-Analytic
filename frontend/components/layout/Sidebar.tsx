"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import {
  LayoutDashboard,
  Megaphone,
  BarChart3,
  Link2,
  Settings,
  CreditCard,
  ChevronLeft,
  ChevronRight,
  BookOpen,
} from "lucide-react";

import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import { useSidebarStore, useOrganizationStore } from "@/stores/app-store";

interface NavItem {
  title: string;
  href: string;
  icon: React.ElementType;
  badge?: string;
  requiresPaidPlan?: boolean;
}

const navItems: NavItem[] = [
  {
    title: "Dashboard",
    href: "/dashboard",
    icon: LayoutDashboard,
  },
  {
    title: "Campaigns",
    href: "/dashboard/campaigns",
    icon: Megaphone,
  },
  {
    title: "Analytics",
    href: "/dashboard/analytics",
    icon: BarChart3,
  },
  {
    title: "Connections",
    href: "/dashboard/connections",
    icon: Link2,
  },
  {
    title: "Settings",
    href: "/dashboard/settings",
    icon: Settings,
  },
  {
    title: "Billing",
    href: "/dashboard/billing",
    icon: CreditCard,
    requiresPaidPlan: true,
  },
];

export function Sidebar() {
  const pathname = usePathname();
  const { isCollapsed, toggleCollapsed } = useSidebarStore();
  const { selectedOrganization } = useOrganizationStore();

  const isPaidPlan = selectedOrganization?.plan !== "free";

  return (
    <aside
      className={cn(
        "fixed left-0 top-0 z-40 h-screen bg-white border-r border-slate-200 transition-all duration-300 hidden md:flex flex-col",
        isCollapsed ? "w-[70px]" : "w-[240px]"
      )}
    >
      {/* Logo */}
      <div className="h-16 flex items-center justify-between px-4 border-b border-slate-200">
        <Link href="/dashboard" className="flex items-center gap-3">
          <div className="w-9 h-9 bg-gradient-to-br from-blue-600 to-blue-700 rounded-lg flex items-center justify-center flex-shrink-0">
            <BarChart3 className="w-5 h-5 text-white" />
          </div>
          {!isCollapsed && (
            <span className="text-lg font-bold text-slate-900">Ads Analytics</span>
          )}
        </Link>
      </div>

      {/* Navigation */}
      <nav className="flex-1 p-3 space-y-1 overflow-y-auto">
        {navItems.map((item) => {
          // Hide billing for free plan
          if (item.requiresPaidPlan && !isPaidPlan) {
            return null;
          }

          const isActive =
            item.href === "/dashboard"
              ? pathname === "/dashboard"
              : pathname.startsWith(item.href);

          return (
            <Link key={item.href} href={item.href}>
              <div
                className={cn(
                  "flex items-center gap-3 px-3 py-2.5 rounded-lg transition-colors group",
                  isActive
                    ? "bg-blue-50 text-blue-600"
                    : "text-slate-600 hover:bg-slate-50 hover:text-slate-900"
                )}
              >
                <item.icon
                  className={cn(
                    "w-5 h-5 flex-shrink-0",
                    isActive ? "text-blue-600" : "text-slate-400 group-hover:text-slate-600"
                  )}
                />
                {!isCollapsed && (
                  <span className="font-medium text-sm">{item.title}</span>
                )}
                {!isCollapsed && item.badge && (
                  <span className="ml-auto text-xs bg-blue-100 text-blue-600 px-2 py-0.5 rounded-full">
                    {item.badge}
                  </span>
                )}
              </div>
            </Link>
          );
        })}
      </nav>

      {/* Help & Docs */}
      <div className="px-3 pb-2">
        <Link href="/docs" target="_blank">
          <div
            className={cn(
              "flex items-center gap-3 px-3 py-2.5 rounded-lg transition-colors group",
              "text-slate-600 hover:bg-slate-50 hover:text-slate-900"
            )}
          >
            <BookOpen className="w-5 h-5 flex-shrink-0 text-slate-400 group-hover:text-slate-600" />
            {!isCollapsed && (
              <span className="font-medium text-sm">Help & Docs</span>
            )}
          </div>
        </Link>
      </div>

      {/* Collapse Button */}
      <div className="p-3 border-t border-slate-200">
        <Button
          variant="ghost"
          size="sm"
          className={cn("w-full justify-center", !isCollapsed && "justify-start")}
          onClick={toggleCollapsed}
        >
          {isCollapsed ? (
            <ChevronRight className="w-4 h-4" />
          ) : (
            <>
              <ChevronLeft className="w-4 h-4 mr-2" />
              <span className="text-sm">Collapse</span>
            </>
          )}
        </Button>
      </div>
    </aside>
  );
}
