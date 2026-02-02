"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import {
  LayoutDashboard,
  Users,
  TrendingUp,
  Activity,
  Settings,
  LogOut,
  ChevronLeft,
  BarChart3,
  Layers,
  DollarSign,
  UserMinus,
} from "lucide-react";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import { useAuth } from "@/lib/api/hooks/use-auth";

const navigation = [
  { name: "Dashboard", href: "/admin", icon: LayoutDashboard },
  { name: "Active Users", href: "/admin/users", icon: Users },
  { name: "Churned Users", href: "/admin/churned", icon: UserMinus },
  { name: "Revenue", href: "/admin/revenue", icon: DollarSign },
  { name: "Funnels", href: "/admin/funnels", icon: TrendingUp },
  { name: "Platforms", href: "/admin/platforms", icon: Layers },
  { name: "Features", href: "/admin/features", icon: BarChart3 },
  { name: "Events", href: "/admin/events", icon: Activity },
];

export function AdminSidebar() {
  const pathname = usePathname();
  const { logout } = useAuth();

  return (
    <aside className="fixed left-0 top-0 z-40 h-screen w-64 border-r border-slate-200 bg-white">
      {/* Header */}
      <div className="flex h-16 items-center justify-between border-b border-slate-200 px-4">
        <div className="flex items-center gap-2">
          <div className="h-8 w-8 rounded-lg bg-indigo-600 flex items-center justify-center">
            <Settings className="h-5 w-5 text-white" />
          </div>
          <span className="font-semibold text-slate-900">Admin Panel</span>
        </div>
        <Link href="/dashboard">
          <Button variant="ghost" size="icon" className="h-8 w-8">
            <ChevronLeft className="h-4 w-4" />
          </Button>
        </Link>
      </div>

      {/* Navigation */}
      <nav className="flex-1 space-y-1 px-3 py-4">
        {navigation.map((item) => {
          const isActive = pathname === item.href;
          return (
            <Link
              key={item.name}
              href={item.href}
              className={cn(
                "flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-colors",
                isActive
                  ? "bg-indigo-50 text-indigo-700"
                  : "text-slate-600 hover:bg-slate-100 hover:text-slate-900"
              )}
            >
              <item.icon className="h-5 w-5" />
              {item.name}
            </Link>
          );
        })}
      </nav>

      {/* Footer */}
      <div className="border-t border-slate-200 p-4">
        <Button
          variant="ghost"
          className="w-full justify-start gap-3 text-slate-600 hover:text-slate-900"
          onClick={() => logout()}
        >
          <LogOut className="h-5 w-5" />
          Log Out
        </Button>
      </div>
    </aside>
  );
}
