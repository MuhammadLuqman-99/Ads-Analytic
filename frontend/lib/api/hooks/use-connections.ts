"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { accountsApi, ListConnectionsResponse } from "../services/accounts";
import type { Platform, ConnectedAccount, SyncStatus } from "../types";

// ============================================
// Query Keys
// ============================================

export const connectionKeys = {
  all: ["connections"] as const,
  lists: () => [...connectionKeys.all, "list"] as const,
  list: (params?: Parameters<typeof accountsApi.listConnections>[0]) =>
    [...connectionKeys.lists(), params] as const,
  details: () => [...connectionKeys.all, "detail"] as const,
  detail: (id: string) => [...connectionKeys.details(), id] as const,
  syncStatus: () => [...connectionKeys.all, "syncStatus"] as const,
  syncStatusById: (id: string) => [...connectionKeys.syncStatus(), id] as const,
  platforms: () => [...connectionKeys.all, "platforms"] as const,
};

// ============================================
// useConnections Hook
// ============================================

export interface UseConnectionsOptions {
  platform?: Platform;
  status?: ConnectedAccount["status"];
  enabled?: boolean;
}

export function useConnections(options: UseConnectionsOptions = {}) {
  const { platform, status, enabled = true } = options;
  const queryClient = useQueryClient();

  // List connections
  const query = useQuery({
    queryKey: connectionKeys.list({ platform, status }),
    queryFn: () => accountsApi.listConnections({ platform, status }),
    enabled,
    staleTime: 2 * 60 * 1000,
  });

  // Connect mutation (initiates OAuth)
  const connectMutation = useMutation({
    mutationFn: (platform: Platform) =>
      accountsApi.initiateConnection(platform, "/dashboard/connections"),
    onSuccess: (data) => {
      // Redirect to OAuth URL
      if (typeof window !== "undefined" && data.authUrl) {
        window.location.href = data.authUrl;
      }
    },
  });

  // Disconnect mutation
  const disconnectMutation = useMutation({
    mutationFn: (accountId: string) => accountsApi.disconnect(accountId),
    onSuccess: (_, accountId) => {
      // Remove from cache
      queryClient.setQueryData(
        connectionKeys.list({ platform, status }),
        (old: ListConnectionsResponse | undefined) => {
          if (!old) return old;
          return {
            ...old,
            data: old.data.filter((a) => a.id !== accountId),
            total: old.total - 1,
          };
        }
      );
      queryClient.invalidateQueries({ queryKey: connectionKeys.lists() });
    },
  });

  // Sync mutation
  const syncMutation = useMutation({
    mutationFn: (accountId: string) => accountsApi.syncAccount(accountId),
    onSuccess: (syncStatus, accountId) => {
      // Update sync status in cache
      queryClient.setQueryData(connectionKeys.syncStatusById(accountId), syncStatus);

      // Update account status to syncing
      queryClient.setQueryData(
        connectionKeys.list({ platform, status }),
        (old: ListConnectionsResponse | undefined) => {
          if (!old) return old;
          return {
            ...old,
            data: old.data.map((a) =>
              a.id === accountId ? { ...a, status: "syncing" as const } : a
            ),
          };
        }
      );
    },
  });

  // Reconnect mutation
  const reconnectMutation = useMutation({
    mutationFn: ({
      accountId,
      platform,
    }: {
      accountId: string;
      platform: Platform;
    }) => accountsApi.reconnect(accountId, platform),
    onSuccess: (data) => {
      if (typeof window !== "undefined" && data.authUrl) {
        window.location.href = data.authUrl;
      }
    },
  });

  return {
    // Data
    connections: query.data?.data || [],
    total: query.data?.total || 0,

    // Computed
    activeCount: query.data?.data.filter((a) => a.status === "active").length || 0,
    errorCount:
      query.data?.data.filter((a) => a.status === "error" || a.status === "expired")
        .length || 0,

    // States
    isLoading: query.isLoading,
    isFetching: query.isFetching,
    error: query.error,

    // Actions
    refetch: query.refetch,
    connect: connectMutation.mutateAsync,
    disconnect: disconnectMutation.mutateAsync,
    sync: syncMutation.mutateAsync,
    reconnect: reconnectMutation.mutateAsync,

    // Mutation states
    isConnecting: connectMutation.isPending,
    isDisconnecting: disconnectMutation.isPending,
    isSyncing: syncMutation.isPending,
    isReconnecting: reconnectMutation.isPending,
  };
}

// ============================================
// useConnection Hook (single)
// ============================================

export function useConnection(accountId: string, enabled = true) {
  const query = useQuery({
    queryKey: connectionKeys.detail(accountId),
    queryFn: () => accountsApi.getConnection(accountId),
    enabled: enabled && !!accountId,
    staleTime: 2 * 60 * 1000,
  });

  return {
    connection: query.data,
    isLoading: query.isLoading,
    error: query.error,
    refetch: query.refetch,
  };
}

// ============================================
// useSyncStatus Hook
// ============================================

export function useSyncStatus(accountId: string) {
  const query = useQuery({
    queryKey: connectionKeys.syncStatusById(accountId),
    queryFn: () => accountsApi.getSyncStatus(accountId),
    enabled: !!accountId,
    refetchInterval: (query) => {
      // Poll while syncing
      const data = query.state.data as SyncStatus | undefined;
      if (data?.status === "syncing") {
        return 2000; // 2 seconds
      }
      return false;
    },
  });

  return {
    status: query.data,
    isLoading: query.isLoading,
    refetch: query.refetch,
  };
}

// ============================================
// useAvailablePlatforms Hook
// ============================================

export function useAvailablePlatforms() {
  return useQuery({
    queryKey: connectionKeys.platforms(),
    queryFn: accountsApi.getAvailablePlatforms,
    staleTime: 60 * 60 * 1000, // 1 hour (rarely changes)
  });
}

export default useConnections;
