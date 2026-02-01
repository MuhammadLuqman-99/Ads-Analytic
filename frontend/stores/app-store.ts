import { create } from "zustand";
import { persist } from "zustand/middleware";

// ============================================================================
// Organization Store
// ============================================================================
export interface Organization {
  id: string;
  name: string;
  logo?: string;
  plan: "free" | "pro" | "business";
}

interface OrganizationState {
  organizations: Organization[];
  selectedOrganization: Organization | null;
  setOrganizations: (orgs: Organization[]) => void;
  setSelectedOrganization: (org: Organization) => void;
}

export const useOrganizationStore = create<OrganizationState>()(
  persist(
    (set) => ({
      organizations: [],
      selectedOrganization: null,
      setOrganizations: (organizations) => set({ organizations }),
      setSelectedOrganization: (selectedOrganization) => set({ selectedOrganization }),
    }),
    {
      name: "organization-storage",
    }
  )
);

// ============================================================================
// Date Range Store
// ============================================================================
export type DateRangePreset =
  | "today"
  | "yesterday"
  | "last7days"
  | "last30days"
  | "thisMonth"
  | "lastMonth"
  | "last90days"
  | "custom";

export interface DateRange {
  from: Date;
  to: Date;
}

interface DateRangeState {
  preset: DateRangePreset;
  dateRange: DateRange;
  compareEnabled: boolean;
  compareDateRange: DateRange | null;
  setPreset: (preset: DateRangePreset) => void;
  setDateRange: (range: DateRange) => void;
  setCompareEnabled: (enabled: boolean) => void;
  setCompareDateRange: (range: DateRange | null) => void;
}

const getPresetDateRange = (preset: DateRangePreset): DateRange => {
  const now = new Date();
  const today = new Date(now.getFullYear(), now.getMonth(), now.getDate());

  switch (preset) {
    case "today":
      return { from: today, to: today };
    case "yesterday":
      const yesterday = new Date(today);
      yesterday.setDate(yesterday.getDate() - 1);
      return { from: yesterday, to: yesterday };
    case "last7days":
      const last7 = new Date(today);
      last7.setDate(last7.getDate() - 6);
      return { from: last7, to: today };
    case "last30days":
      const last30 = new Date(today);
      last30.setDate(last30.getDate() - 29);
      return { from: last30, to: today };
    case "thisMonth":
      return {
        from: new Date(now.getFullYear(), now.getMonth(), 1),
        to: today,
      };
    case "lastMonth":
      const lastMonthStart = new Date(now.getFullYear(), now.getMonth() - 1, 1);
      const lastMonthEnd = new Date(now.getFullYear(), now.getMonth(), 0);
      return { from: lastMonthStart, to: lastMonthEnd };
    case "last90days":
      const last90 = new Date(today);
      last90.setDate(last90.getDate() - 89);
      return { from: last90, to: today };
    default:
      const defaultLast30 = new Date(today);
      defaultLast30.setDate(defaultLast30.getDate() - 29);
      return { from: defaultLast30, to: today };
  }
};

export const useDateRangeStore = create<DateRangeState>()(
  persist(
    (set) => ({
      preset: "last30days",
      dateRange: getPresetDateRange("last30days"),
      compareEnabled: false,
      compareDateRange: null,
      setPreset: (preset) =>
        set({
          preset,
          dateRange: preset !== "custom" ? getPresetDateRange(preset) : undefined,
        }),
      setDateRange: (dateRange) => set({ dateRange, preset: "custom" }),
      setCompareEnabled: (compareEnabled) => set({ compareEnabled }),
      setCompareDateRange: (compareDateRange) => set({ compareDateRange }),
    }),
    {
      name: "date-range-storage",
      partialize: (state) => ({
        preset: state.preset,
        compareEnabled: state.compareEnabled,
      }),
    }
  )
);

// ============================================================================
// Sidebar Store
// ============================================================================
interface SidebarState {
  isCollapsed: boolean;
  isMobileOpen: boolean;
  setCollapsed: (collapsed: boolean) => void;
  toggleCollapsed: () => void;
  setMobileOpen: (open: boolean) => void;
  toggleMobileOpen: () => void;
}

export const useSidebarStore = create<SidebarState>()(
  persist(
    (set) => ({
      isCollapsed: false,
      isMobileOpen: false,
      setCollapsed: (isCollapsed) => set({ isCollapsed }),
      toggleCollapsed: () => set((state) => ({ isCollapsed: !state.isCollapsed })),
      setMobileOpen: (isMobileOpen) => set({ isMobileOpen }),
      toggleMobileOpen: () => set((state) => ({ isMobileOpen: !state.isMobileOpen })),
    }),
    {
      name: "sidebar-storage",
      partialize: (state) => ({ isCollapsed: state.isCollapsed }),
    }
  )
);

// ============================================================================
// Notification Store
// ============================================================================
export interface Notification {
  id: string;
  title: string;
  message: string;
  type: "info" | "success" | "warning" | "error";
  read: boolean;
  createdAt: Date;
  link?: string;
}

interface NotificationState {
  notifications: Notification[];
  unreadCount: number;
  addNotification: (notification: Omit<Notification, "id" | "read" | "createdAt">) => void;
  markAsRead: (id: string) => void;
  markAllAsRead: () => void;
  removeNotification: (id: string) => void;
  clearAll: () => void;
}

export const useNotificationStore = create<NotificationState>((set) => ({
  notifications: [],
  unreadCount: 0,
  addNotification: (notification) =>
    set((state) => {
      const newNotification: Notification = {
        ...notification,
        id: crypto.randomUUID(),
        read: false,
        createdAt: new Date(),
      };
      return {
        notifications: [newNotification, ...state.notifications],
        unreadCount: state.unreadCount + 1,
      };
    }),
  markAsRead: (id) =>
    set((state) => ({
      notifications: state.notifications.map((n) =>
        n.id === id ? { ...n, read: true } : n
      ),
      unreadCount: Math.max(0, state.unreadCount - 1),
    })),
  markAllAsRead: () =>
    set((state) => ({
      notifications: state.notifications.map((n) => ({ ...n, read: true })),
      unreadCount: 0,
    })),
  removeNotification: (id) =>
    set((state) => {
      const notification = state.notifications.find((n) => n.id === id);
      return {
        notifications: state.notifications.filter((n) => n.id !== id),
        unreadCount: notification?.read ? state.unreadCount : state.unreadCount - 1,
      };
    }),
  clearAll: () => set({ notifications: [], unreadCount: 0 }),
}));
