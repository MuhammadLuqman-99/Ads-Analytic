"use client"

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query"
import {
    mockConnectedAccounts,
    type ConnectedAccount,
    type Platform
} from "@/lib/mock-data"

// Simulate API delay
const delay = (ms: number) => new Promise(resolve => setTimeout(resolve, ms))

// Connected Accounts Hook
export function useConnectedAccounts() {
    return useQuery<ConnectedAccount[]>({
        queryKey: ['connected-accounts'],
        queryFn: async () => {
            await delay(400)
            return mockConnectedAccounts
        },
    })
}

// Single Account Hook
export function useConnectedAccount(id: string) {
    return useQuery<ConnectedAccount | undefined>({
        queryKey: ['connected-account', id],
        queryFn: async () => {
            await delay(200)
            return mockConnectedAccounts.find(a => a.id === id)
        },
        enabled: !!id,
    })
}

// Disconnect Account Mutation
export function useDisconnectAccount() {
    const queryClient = useQueryClient()

    return useMutation({
        mutationFn: async (accountId: string) => {
            await delay(800)
            // In real app, this would call the API
            console.log('Disconnecting account:', accountId)
            return { success: true }
        },
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['connected-accounts'] })
        },
    })
}

// Connect Account Mutation (OAuth flow simulation)
export function useConnectAccount() {
    const queryClient = useQueryClient()

    return useMutation({
        mutationFn: async (platform: Platform) => {
            await delay(1500)
            // In real app, this would redirect to OAuth flow
            console.log('Initiating OAuth for:', platform)
            return {
                success: true,
                redirectUrl: `/api/auth/${platform}/connect`
            }
        },
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['connected-accounts'] })
        },
    })
}

// Sync Account Data Mutation
export function useSyncAccount() {
    const queryClient = useQueryClient()

    return useMutation({
        mutationFn: async (accountId: string) => {
            await delay(2000)
            console.log('Syncing account:', accountId)
            return { success: true, syncedAt: new Date().toISOString() }
        },
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['connected-accounts'] })
            queryClient.invalidateQueries({ queryKey: ['campaigns'] })
            queryClient.invalidateQueries({ queryKey: ['dashboard-metrics'] })
        },
    })
}

// Sync All Accounts Mutation
export function useSyncAllAccounts() {
    const queryClient = useQueryClient()

    return useMutation({
        mutationFn: async () => {
            await delay(3000)
            console.log('Syncing all accounts')
            return { success: true, syncedAt: new Date().toISOString() }
        },
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['connected-accounts'] })
            queryClient.invalidateQueries({ queryKey: ['campaigns'] })
            queryClient.invalidateQueries({ queryKey: ['dashboard-metrics'] })
            queryClient.invalidateQueries({ queryKey: ['daily-metrics'] })
            queryClient.invalidateQueries({ queryKey: ['platform-metrics'] })
        },
    })
}
