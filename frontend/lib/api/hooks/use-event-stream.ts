"use client";

import { useEffect, useRef, useCallback, useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { dashboardKeys } from "./use-dashboard";

// Simple toast notification type for external handling
export interface ToastNotification {
  id: string;
  type: "info" | "success" | "error" | "warning";
  title: string;
  message?: string;
}

// Event types from backend
export type EventType =
  | "connected"
  | "sync:started"
  | "sync:progress"
  | "sync:completed"
  | "sync:error"
  | "data:updated";

// Event data types
export interface SyncStartedData {
  platform: string;
  account_id: string;
}

export interface SyncProgressData {
  platform: string;
  account_id: string;
  progress_percent: number;
  message?: string;
}

export interface SyncCompletedData {
  platform: string;
  account_id: string;
  records_synced: number;
  duration: number;
  duration_str: string;
}

export interface SyncErrorData {
  platform: string;
  account_id: string;
  error_message: string;
  retryable: boolean;
}

export interface DataUpdatedData {
  affected: string[];
}

export interface SSEEvent<T = unknown> {
  id: string;
  type: EventType;
  timestamp: string;
  data: T;
}

export interface SyncStatus {
  platform: string;
  accountId: string;
  status: "idle" | "syncing" | "completed" | "error";
  progress: number;
  message?: string;
  lastSyncedAt?: Date;
  error?: string;
}

export interface UseEventStreamOptions {
  enabled?: boolean;
  onSyncStarted?: (data: SyncStartedData) => void;
  onSyncProgress?: (data: SyncProgressData) => void;
  onSyncCompleted?: (data: SyncCompletedData) => void;
  onSyncError?: (data: SyncErrorData) => void;
  onDataUpdated?: (data: DataUpdatedData) => void;
  onToast?: (notification: ToastNotification) => void;
  showToasts?: boolean;
}

const RECONNECT_DELAY = 5000; // 5 seconds
const POLLING_INTERVAL = 30000; // 30 seconds fallback

export function useEventStream(options: UseEventStreamOptions = {}) {
  const {
    enabled = true,
    onSyncStarted,
    onSyncProgress,
    onSyncCompleted,
    onSyncError,
    onDataUpdated,
    onToast,
    showToasts = true,
  } = options;

  // Helper to show toast notifications
  const showToast = useCallback(
    (notification: ToastNotification) => {
      if (showToasts) {
        if (onToast) {
          onToast(notification);
        } else {
          // Fallback to console
          const prefix = notification.type === "error" ? "ERROR" : notification.type.toUpperCase();
          console.log(`[${prefix}] ${notification.title}${notification.message ? `: ${notification.message}` : ""}`);
        }
      }
    },
    [showToasts, onToast]
  );

  const queryClient = useQueryClient();
  const eventSourceRef = useRef<EventSource | null>(null);
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null);
  const pollingIntervalRef = useRef<NodeJS.Timeout | null>(null);

  const [isConnected, setIsConnected] = useState(false);
  const [syncStatuses, setSyncStatuses] = useState<Map<string, SyncStatus>>(
    new Map()
  );

  // Get sync status key
  const getSyncKey = (platform: string, accountId: string) =>
    `${platform}:${accountId}`;

  // Update sync status
  const updateSyncStatus = useCallback(
    (platform: string, accountId: string, update: Partial<SyncStatus>) => {
      setSyncStatuses((prev) => {
        const key = getSyncKey(platform, accountId);
        const existing = prev.get(key) || {
          platform,
          accountId,
          status: "idle" as const,
          progress: 0,
        };
        const newMap = new Map(prev);
        newMap.set(key, { ...existing, ...update });
        return newMap;
      });
    },
    []
  );

  // Invalidate affected caches
  const invalidateAffectedCaches = useCallback(
    (affected: string[]) => {
      affected.forEach((key) => {
        switch (key) {
          case "dashboard":
            queryClient.invalidateQueries({ queryKey: dashboardKeys.all });
            break;
          case "campaigns":
            queryClient.invalidateQueries({ queryKey: ["campaigns"] });
            break;
          case "platforms":
            queryClient.invalidateQueries({ queryKey: ["platforms"] });
            break;
          case "accounts":
            queryClient.invalidateQueries({ queryKey: ["accounts"] });
            break;
          default:
            // Invalidate by key name
            queryClient.invalidateQueries({ queryKey: [key] });
        }
      });
    },
    [queryClient]
  );

  // Handle incoming events
  const handleEvent = useCallback(
    (eventType: EventType, data: unknown) => {
      switch (eventType) {
        case "connected":
          setIsConnected(true);
          break;

        case "sync:started": {
          const syncData = data as SyncStartedData;
          updateSyncStatus(syncData.platform, syncData.account_id, {
            status: "syncing",
            progress: 0,
            message: "Starting sync...",
          });
          onSyncStarted?.(syncData);
          showToast({
            id: `sync-${syncData.platform}-${syncData.account_id}`,
            type: "info",
            title: `Syncing ${syncData.platform}...`,
          });
          break;
        }

        case "sync:progress": {
          const progressData = data as SyncProgressData;
          updateSyncStatus(progressData.platform, progressData.account_id, {
            status: "syncing",
            progress: progressData.progress_percent,
            message: progressData.message,
          });
          onSyncProgress?.(progressData);
          break;
        }

        case "sync:completed": {
          const completedData = data as SyncCompletedData;
          updateSyncStatus(completedData.platform, completedData.account_id, {
            status: "completed",
            progress: 100,
            message: `Synced ${completedData.records_synced} records`,
            lastSyncedAt: new Date(),
          });
          onSyncCompleted?.(completedData);
          showToast({
            id: `sync-${completedData.platform}-${completedData.account_id}`,
            type: "success",
            title: `${completedData.platform} sync completed`,
            message: `${completedData.records_synced} records synced`,
          });
          // Auto-invalidate dashboard after sync
          invalidateAffectedCaches(["dashboard", "campaigns"]);
          break;
        }

        case "sync:error": {
          const errorData = data as SyncErrorData;
          updateSyncStatus(errorData.platform, errorData.account_id, {
            status: "error",
            progress: 0,
            error: errorData.error_message,
          });
          onSyncError?.(errorData);
          showToast({
            id: `sync-${errorData.platform}-${errorData.account_id}`,
            type: "error",
            title: `${errorData.platform} sync failed`,
            message: errorData.error_message,
          });
          break;
        }

        case "data:updated": {
          const updatedData = data as DataUpdatedData;
          onDataUpdated?.(updatedData);
          invalidateAffectedCaches(updatedData.affected);
          break;
        }
      }
    },
    [
      updateSyncStatus,
      onSyncStarted,
      onSyncProgress,
      onSyncCompleted,
      onSyncError,
      onDataUpdated,
      showToast,
      invalidateAffectedCaches,
    ]
  );

  // Connect to SSE
  const connect = useCallback(() => {
    if (eventSourceRef.current) {
      eventSourceRef.current.close();
    }

    const baseUrl = process.env.NEXT_PUBLIC_API_URL || "/api/v1";
    const url = `${baseUrl}/events/stream`;

    try {
      const eventSource = new EventSource(url, { withCredentials: true });
      eventSourceRef.current = eventSource;

      eventSource.onopen = () => {
        setIsConnected(true);
        // Clear polling when SSE connects
        if (pollingIntervalRef.current) {
          clearInterval(pollingIntervalRef.current);
          pollingIntervalRef.current = null;
        }
      };

      eventSource.onerror = () => {
        setIsConnected(false);
        eventSource.close();
        eventSourceRef.current = null;

        // Schedule reconnect
        if (reconnectTimeoutRef.current) {
          clearTimeout(reconnectTimeoutRef.current);
        }
        reconnectTimeoutRef.current = setTimeout(connect, RECONNECT_DELAY);

        // Start fallback polling
        if (!pollingIntervalRef.current) {
          pollingIntervalRef.current = setInterval(() => {
            // Polling fallback - just invalidate dashboard
            queryClient.invalidateQueries({ queryKey: dashboardKeys.all });
          }, POLLING_INTERVAL);
        }
      };

      // Listen for specific event types
      const eventTypes: EventType[] = [
        "connected",
        "sync:started",
        "sync:progress",
        "sync:completed",
        "sync:error",
        "data:updated",
      ];

      eventTypes.forEach((eventType) => {
        eventSource.addEventListener(eventType, (event: MessageEvent) => {
          try {
            const data = JSON.parse(event.data);
            handleEvent(eventType, data.data || data);
          } catch (e) {
            console.error("Failed to parse SSE event:", e);
          }
        });
      });
    } catch (error) {
      console.error("Failed to create EventSource:", error);
      setIsConnected(false);
    }
  }, [handleEvent, queryClient]);

  // Disconnect
  const disconnect = useCallback(() => {
    if (eventSourceRef.current) {
      eventSourceRef.current.close();
      eventSourceRef.current = null;
    }
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current);
      reconnectTimeoutRef.current = null;
    }
    if (pollingIntervalRef.current) {
      clearInterval(pollingIntervalRef.current);
      pollingIntervalRef.current = null;
    }
    setIsConnected(false);
  }, []);

  // Effect to manage connection
  useEffect(() => {
    if (enabled) {
      connect();
    } else {
      disconnect();
    }

    return () => {
      disconnect();
    };
  }, [enabled, connect, disconnect]);

  // Get sync status for a specific platform/account
  const getSyncStatus = useCallback(
    (platform: string, accountId: string): SyncStatus | undefined => {
      return syncStatuses.get(getSyncKey(platform, accountId));
    },
    [syncStatuses]
  );

  // Check if any sync is in progress
  const isSyncing = Array.from(syncStatuses.values()).some(
    (s) => s.status === "syncing"
  );

  // Get all active syncs
  const activeSyncs = Array.from(syncStatuses.values()).filter(
    (s) => s.status === "syncing"
  );

  return {
    isConnected,
    isSyncing,
    activeSyncs,
    syncStatuses: Array.from(syncStatuses.values()),
    getSyncStatus,
    connect,
    disconnect,
  };
}

export default useEventStream;
