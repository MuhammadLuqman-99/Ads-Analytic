"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { signOut, useSession } from "next-auth/react";
import { format } from "date-fns";
import {
  Menu,
  Bell,
  ChevronDown,
  Building2,
  Plus,
  Check,
  Calendar,
  User,
  Settings,
  CreditCard,
  LogOut,
} from "lucide-react";

import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { Calendar as CalendarComponent } from "@/components/ui/calendar";
import {
  useSidebarStore,
  useOrganizationStore,
  useDateRangeStore,
  useNotificationStore,
  type DateRangePreset,
} from "@/stores/app-store";

const datePresets: { label: string; value: DateRangePreset }[] = [
  { label: "Today", value: "today" },
  { label: "Yesterday", value: "yesterday" },
  { label: "Last 7 days", value: "last7days" },
  { label: "Last 30 days", value: "last30days" },
  { label: "This month", value: "thisMonth" },
  { label: "Last month", value: "lastMonth" },
  { label: "Last 90 days", value: "last90days" },
];

export function Header() {
  const router = useRouter();
  const { data: session } = useSession();
  const { toggleMobileOpen } = useSidebarStore();
  const { organizations, selectedOrganization, setSelectedOrganization } =
    useOrganizationStore();
  const { preset, dateRange, setPreset, setDateRange } = useDateRangeStore();
  const { notifications, unreadCount, markAsRead, markAllAsRead } =
    useNotificationStore();

  const [isCalendarOpen, setIsCalendarOpen] = useState(false);

  const handleSignOut = async () => {
    await signOut({ callbackUrl: "/login" });
  };

  const getInitials = (name?: string | null) => {
    if (!name) return "U";
    return name
      .split(" ")
      .map((n) => n[0])
      .join("")
      .toUpperCase()
      .slice(0, 2);
  };

  const formatDateRange = () => {
    if (preset !== "custom") {
      return datePresets.find((p) => p.value === preset)?.label || "Select date";
    }
    return `${format(dateRange.from, "MMM d")} - ${format(dateRange.to, "MMM d, yyyy")}`;
  };

  return (
    <header className="h-16 bg-white border-b border-slate-200 flex items-center justify-between px-4 md:px-6">
      {/* Left Section */}
      <div className="flex items-center gap-4">
        {/* Mobile Menu Button */}
        <Button
          variant="ghost"
          size="icon"
          className="md:hidden"
          onClick={toggleMobileOpen}
        >
          <Menu className="w-5 h-5" />
        </Button>

        {/* Organization Switcher */}
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="outline" className="gap-2 max-w-[200px]">
              <Building2 className="w-4 h-4 text-slate-500" />
              <span className="truncate">
                {selectedOrganization?.name || "Select Organization"}
              </span>
              <ChevronDown className="w-4 h-4 text-slate-400" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="start" className="w-[240px]">
            <DropdownMenuLabel>Organizations</DropdownMenuLabel>
            <DropdownMenuSeparator />
            {organizations.map((org) => (
              <DropdownMenuItem
                key={org.id}
                onClick={() => setSelectedOrganization(org)}
                className="gap-2"
              >
                <div className="w-8 h-8 rounded-lg bg-slate-100 flex items-center justify-center text-xs font-medium">
                  {org.name.slice(0, 2).toUpperCase()}
                </div>
                <div className="flex-1 min-w-0">
                  <p className="text-sm font-medium truncate">{org.name}</p>
                  <p className="text-xs text-slate-500 capitalize">{org.plan} plan</p>
                </div>
                {selectedOrganization?.id === org.id && (
                  <Check className="w-4 h-4 text-blue-600" />
                )}
              </DropdownMenuItem>
            ))}
            <DropdownMenuSeparator />
            <DropdownMenuItem className="gap-2">
              <Plus className="w-4 h-4" />
              <span>Create Organization</span>
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>

        {/* Date Range Picker */}
        <Popover open={isCalendarOpen} onOpenChange={setIsCalendarOpen}>
          <PopoverTrigger asChild>
            <Button variant="outline" className="gap-2 hidden sm:flex">
              <Calendar className="w-4 h-4 text-slate-500" />
              <span>{formatDateRange()}</span>
              <ChevronDown className="w-4 h-4 text-slate-400" />
            </Button>
          </PopoverTrigger>
          <PopoverContent className="w-auto p-0" align="start">
            <div className="flex">
              {/* Presets */}
              <div className="border-r border-slate-200 p-2 space-y-1">
                {datePresets.map((presetItem) => (
                  <button
                    key={presetItem.value}
                    onClick={() => {
                      setPreset(presetItem.value);
                      setIsCalendarOpen(false);
                    }}
                    className={cn(
                      "w-full text-left px-3 py-2 text-sm rounded-md hover:bg-slate-100 transition-colors",
                      preset === presetItem.value && "bg-slate-100 font-medium"
                    )}
                  >
                    {presetItem.label}
                  </button>
                ))}
              </div>
              {/* Calendar */}
              <div className="p-2">
                <CalendarComponent
                  mode="range"
                  selected={{
                    from: dateRange.from,
                    to: dateRange.to,
                  }}
                  onSelect={(range) => {
                    if (range?.from && range?.to) {
                      setDateRange({ from: range.from, to: range.to });
                    }
                  }}
                  numberOfMonths={2}
                />
              </div>
            </div>
          </PopoverContent>
        </Popover>
      </div>

      {/* Right Section */}
      <div className="flex items-center gap-2">
        {/* Notifications */}
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" size="icon" className="relative">
              <Bell className="w-5 h-5" />
              {unreadCount > 0 && (
                <span className="absolute -top-1 -right-1 w-5 h-5 bg-red-500 text-white text-xs rounded-full flex items-center justify-center">
                  {unreadCount > 9 ? "9+" : unreadCount}
                </span>
              )}
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end" className="w-[360px]">
            <div className="flex items-center justify-between px-4 py-2">
              <DropdownMenuLabel className="p-0">Notifications</DropdownMenuLabel>
              {unreadCount > 0 && (
                <button
                  onClick={markAllAsRead}
                  className="text-xs text-blue-600 hover:text-blue-700"
                >
                  Mark all as read
                </button>
              )}
            </div>
            <DropdownMenuSeparator />
            {notifications.length === 0 ? (
              <div className="p-8 text-center text-slate-500">
                <Bell className="w-8 h-8 mx-auto mb-2 opacity-50" />
                <p className="text-sm">No notifications yet</p>
              </div>
            ) : (
              <div className="max-h-[300px] overflow-y-auto">
                {notifications.slice(0, 5).map((notification) => (
                  <DropdownMenuItem
                    key={notification.id}
                    onClick={() => markAsRead(notification.id)}
                    className={cn(
                      "flex flex-col items-start gap-1 p-3 cursor-pointer",
                      !notification.read && "bg-blue-50"
                    )}
                  >
                    <p className="font-medium text-sm">{notification.title}</p>
                    <p className="text-xs text-slate-500 line-clamp-2">
                      {notification.message}
                    </p>
                    <p className="text-xs text-slate-400">
                      {format(new Date(notification.createdAt), "MMM d, h:mm a")}
                    </p>
                  </DropdownMenuItem>
                ))}
              </div>
            )}
            {notifications.length > 0 && (
              <>
                <DropdownMenuSeparator />
                <DropdownMenuItem
                  className="justify-center text-blue-600"
                  onClick={() => router.push("/dashboard/notifications")}
                >
                  View all notifications
                </DropdownMenuItem>
              </>
            )}
          </DropdownMenuContent>
        </DropdownMenu>

        {/* User Menu */}
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" className="gap-2 pl-2 pr-3">
              <Avatar className="w-8 h-8">
                <AvatarImage src={session?.user?.image || undefined} />
                <AvatarFallback className="text-xs">
                  {getInitials(session?.user?.name)}
                </AvatarFallback>
              </Avatar>
              <span className="hidden sm:inline-block text-sm font-medium max-w-[100px] truncate">
                {session?.user?.name || "User"}
              </span>
              <ChevronDown className="w-4 h-4 text-slate-400" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end" className="w-[200px]">
            <div className="px-2 py-1.5">
              <p className="font-medium text-sm">{session?.user?.name}</p>
              <p className="text-xs text-slate-500 truncate">
                {session?.user?.email}
              </p>
            </div>
            <DropdownMenuSeparator />
            <DropdownMenuItem onClick={() => router.push("/dashboard/profile")}>
              <User className="w-4 h-4 mr-2" />
              Profile
            </DropdownMenuItem>
            <DropdownMenuItem onClick={() => router.push("/dashboard/settings")}>
              <Settings className="w-4 h-4 mr-2" />
              Settings
            </DropdownMenuItem>
            <DropdownMenuItem onClick={() => router.push("/dashboard/billing")}>
              <CreditCard className="w-4 h-4 mr-2" />
              Billing
            </DropdownMenuItem>
            <DropdownMenuSeparator />
            <DropdownMenuItem onClick={handleSignOut} className="text-red-600">
              <LogOut className="w-4 h-4 mr-2" />
              Sign out
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    </header>
  );
}
