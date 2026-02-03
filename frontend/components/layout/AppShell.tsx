"use client";

import { useEffect } from "react";

import { cn } from "@/lib/utils";
import { Sidebar } from "./Sidebar";
import { Header } from "./Header";
import { MobileSidebar, BottomNavigation } from "./MobileNav";
import { useSidebarStore, useOrganizationStore } from "@/stores/app-store";
import { useAuthContext } from "@/components/providers/AuthProvider";

interface AppShellProps {
  children: React.ReactNode;
}

export function AppShell({ children }: AppShellProps) {
  const { user, isAuthenticated, isLoading } = useAuthContext();
  const { isCollapsed } = useSidebarStore();
  const { setOrganizations, setSelectedOrganization, selectedOrganization } =
    useOrganizationStore();

  // Initialize organization from session or fetch from API
  useEffect(() => {
    const initializeOrganization = async () => {
      // Wait for auth to be ready
      if (isLoading) return;

      // If we already have an organization selected, skip
      if (selectedOrganization) return;

      try {
        // Fetch organizations from API
        const response = await fetch(
          `${process.env.NEXT_PUBLIC_API_URL}/organizations`,
          {
            credentials: "include",
          }
        );

        if (response.ok) {
          const data = await response.json();
          if (data.data && data.data.length > 0) {
            setOrganizations(data.data);
            setSelectedOrganization(data.data[0]);
          }
        }
      } catch (error) {
        console.error("Failed to fetch organizations:", error);
        // Set a default organization for demo purposes
        const defaultOrg = {
          id: "demo-org",
          name: user?.email ? `${user.email.split("@")[0]}'s Org` : "My Organization",
          plan: "free" as const,
        };
        setOrganizations([defaultOrg]);
        setSelectedOrganization(defaultOrg);
      }
    };

    initializeOrganization();
  }, [isLoading, isAuthenticated, user, selectedOrganization, setOrganizations, setSelectedOrganization]);

  // Show loading state while checking auth
  if (isLoading) {
    return (
      <div className="min-h-screen bg-slate-50 flex items-center justify-center">
        <div className="animate-spin h-8 w-8 border-4 border-blue-600 border-t-transparent rounded-full" />
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-slate-50">
      {/* Desktop Sidebar */}
      <Sidebar />

      {/* Mobile Sidebar (Sheet) */}
      <MobileSidebar />

      {/* Main Content */}
      <div
        className={cn(
          "min-h-screen transition-all duration-300",
          isCollapsed ? "md:pl-[70px]" : "md:pl-[240px]"
        )}
      >
        {/* Header */}
        <Header />

        {/* Page Content */}
        <main className="p-4 md:p-6 pb-20 md:pb-6">{children}</main>
      </div>

      {/* Mobile Bottom Navigation */}
      <BottomNavigation />
    </div>
  );
}
