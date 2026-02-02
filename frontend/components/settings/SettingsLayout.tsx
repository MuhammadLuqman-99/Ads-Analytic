"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { User, Building2, Bell, CreditCard } from "lucide-react";
import { cn } from "@/lib/utils";

interface SettingsTab {
  name: string;
  href: string;
  icon: React.ElementType;
  description: string;
}

const settingsTabs: SettingsTab[] = [
  {
    name: "Profile",
    href: "/dashboard/settings",
    icon: User,
    description: "Manage your personal information",
  },
  {
    name: "Organization",
    href: "/dashboard/settings/organization",
    icon: Building2,
    description: "Team and organization settings",
  },
  {
    name: "Notifications",
    href: "/dashboard/settings/notifications",
    icon: Bell,
    description: "Email and alert preferences",
  },
  {
    name: "Billing",
    href: "/dashboard/settings/billing",
    icon: CreditCard,
    description: "Plans, payments, and invoices",
  },
];

interface SettingsLayoutProps {
  children: React.ReactNode;
}

export function SettingsLayout({ children }: SettingsLayoutProps) {
  const pathname = usePathname();

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-3xl font-bold text-slate-900">Settings</h1>
        <p className="mt-1 text-slate-500">
          Manage your account settings and preferences
        </p>
      </div>

      {/* Tabs Navigation */}
      <div className="border-b border-slate-200">
        <nav className="-mb-px flex space-x-8 overflow-x-auto">
          {settingsTabs.map((tab) => {
            const isActive =
              pathname === tab.href ||
              (tab.href !== "/dashboard/settings" &&
                pathname.startsWith(tab.href));
            const Icon = tab.icon;

            return (
              <Link
                key={tab.name}
                href={tab.href}
                className={cn(
                  "flex items-center gap-2 py-4 px-1 border-b-2 text-sm font-medium whitespace-nowrap transition-colors",
                  isActive
                    ? "border-blue-500 text-blue-600"
                    : "border-transparent text-slate-500 hover:text-slate-700 hover:border-slate-300"
                )}
              >
                <Icon className="h-4 w-4" />
                {tab.name}
              </Link>
            );
          })}
        </nav>
      </div>

      {/* Content */}
      <div>{children}</div>
    </div>
  );
}
