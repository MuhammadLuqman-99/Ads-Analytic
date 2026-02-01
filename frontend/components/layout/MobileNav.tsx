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
  X,
} from "lucide-react";

import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet";
import { useSidebarStore, useOrganizationStore } from "@/stores/app-store";

interface NavItem {
  title: string;
  href: string;
  icon: React.ElementType;
  requiresPaidPlan?: boolean;
}

const navItems: NavItem[] = [
  { title: "Dashboard", href: "/dashboard", icon: LayoutDashboard },
  { title: "Campaigns", href: "/dashboard/campaigns", icon: Megaphone },
  { title: "Analytics", href: "/dashboard/analytics", icon: BarChart3 },
  { title: "Connections", href: "/dashboard/connections", icon: Link2 },
  { title: "Settings", href: "/dashboard/settings", icon: Settings },
  { title: "Billing", href: "/dashboard/billing", icon: CreditCard, requiresPaidPlan: true },
];

// Bottom navigation items for quick access
const bottomNavItems: NavItem[] = [
  { title: "Dashboard", href: "/dashboard", icon: LayoutDashboard },
  { title: "Campaigns", href: "/dashboard/campaigns", icon: Megaphone },
  { title: "Analytics", href: "/dashboard/analytics", icon: BarChart3 },
  { title: "Settings", href: "/dashboard/settings", icon: Settings },
];

export function MobileSidebar() {
  const pathname = usePathname();
  const { isMobileOpen, setMobileOpen } = useSidebarStore();
  const { selectedOrganization } = useOrganizationStore();

  const isPaidPlan = selectedOrganization?.plan !== "free";

  return (
    <Sheet open={isMobileOpen} onOpenChange={setMobileOpen}>
      <SheetContent side="left" className="w-[280px] p-0">
        <SheetHeader className="h-16 flex flex-row items-center justify-between px-4 border-b border-slate-200">
          <div className="flex items-center gap-3">
            <div className="w-9 h-9 bg-gradient-to-br from-blue-600 to-blue-700 rounded-lg flex items-center justify-center">
              <BarChart3 className="w-5 h-5 text-white" />
            </div>
            <SheetTitle className="text-lg font-bold">Ads Analytics</SheetTitle>
          </div>
        </SheetHeader>

        <nav className="p-3 space-y-1">
          {navItems.map((item) => {
            if (item.requiresPaidPlan && !isPaidPlan) return null;

            const isActive =
              item.href === "/dashboard"
                ? pathname === "/dashboard"
                : pathname.startsWith(item.href);

            return (
              <Link
                key={item.href}
                href={item.href}
                onClick={() => setMobileOpen(false)}
              >
                <div
                  className={cn(
                    "flex items-center gap-3 px-3 py-3 rounded-lg transition-colors",
                    isActive
                      ? "bg-blue-50 text-blue-600"
                      : "text-slate-600 hover:bg-slate-50 hover:text-slate-900"
                  )}
                >
                  <item.icon
                    className={cn(
                      "w-5 h-5",
                      isActive ? "text-blue-600" : "text-slate-400"
                    )}
                  />
                  <span className="font-medium">{item.title}</span>
                </div>
              </Link>
            );
          })}
        </nav>
      </SheetContent>
    </Sheet>
  );
}

export function BottomNavigation() {
  const pathname = usePathname();

  return (
    <nav className="fixed bottom-0 left-0 right-0 z-50 bg-white border-t border-slate-200 md:hidden">
      <div className="flex items-center justify-around h-16">
        {bottomNavItems.map((item) => {
          const isActive =
            item.href === "/dashboard"
              ? pathname === "/dashboard"
              : pathname.startsWith(item.href);

          return (
            <Link
              key={item.href}
              href={item.href}
              className={cn(
                "flex flex-col items-center justify-center flex-1 h-full gap-1 transition-colors",
                isActive ? "text-blue-600" : "text-slate-500"
              )}
            >
              <item.icon className="w-5 h-5" />
              <span className="text-xs font-medium">{item.title}</span>
            </Link>
          );
        })}
      </div>
    </nav>
  );
}
