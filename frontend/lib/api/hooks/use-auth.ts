"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import { authApi, SessionResponse } from "../services/auth";
import type {
  User,
  LoginCredentials,
  RegisterData,
  AuthResponse,
  PasswordResetRequest,
  PasswordResetConfirm,
} from "../types";

// ============================================
// Query Keys
// ============================================

export const authKeys = {
  all: ["auth"] as const,
  session: () => [...authKeys.all, "session"] as const,
  user: () => [...authKeys.all, "user"] as const,
};

// ============================================
// useSession Hook (Cookie-based)
// ============================================

/**
 * Hook to check authentication status using the session endpoint
 * This is the primary way to check if the user is authenticated
 * when using httpOnly cookies for token storage
 */
export function useSession() {
  const query = useQuery({
    queryKey: authKeys.session(),
    queryFn: authApi.getSession,
    staleTime: 5 * 60 * 1000, // 5 minutes
    retry: false,
    refetchOnWindowFocus: true,
  });

  return {
    session: query.data,
    isAuthenticated: query.data?.authenticated ?? false,
    user: query.data?.user,
    organizationId: query.data?.organizationId,
    permissions: query.data?.permissions ?? [],
    isLoading: query.isLoading,
    error: query.error,
    refetch: query.refetch,
  };
}

// ============================================
// useAuth Hook
// ============================================

export function useAuth() {
  const queryClient = useQueryClient();
  const router = useRouter();

  // Get session status (preferred method with httpOnly cookies)
  const {
    session,
    isAuthenticated,
    user: sessionUser,
    isLoading,
    error,
    refetch: refetchSession,
  } = useSession();

  // Get current user profile (detailed info)
  const userQuery = useQuery({
    queryKey: authKeys.user(),
    queryFn: authApi.getCurrentUser,
    enabled: isAuthenticated,
    staleTime: 5 * 60 * 1000, // 5 minutes
    retry: false,
  });

  // Login mutation
  const loginMutation = useMutation({
    mutationFn: (credentials: LoginCredentials) => authApi.login(credentials),
    onSuccess: (data: AuthResponse) => {
      // Invalidate session to refetch
      queryClient.invalidateQueries({ queryKey: authKeys.session() });
      queryClient.setQueryData(authKeys.user(), data.user);
      router.push("/dashboard");
    },
  });

  // Register mutation
  const registerMutation = useMutation({
    mutationFn: (data: RegisterData) => authApi.register(data),
    onSuccess: (data: AuthResponse) => {
      queryClient.invalidateQueries({ queryKey: authKeys.session() });
      queryClient.setQueryData(authKeys.user(), data.user);
      router.push("/onboarding");
    },
  });

  // Logout mutation
  const logoutMutation = useMutation({
    mutationFn: authApi.logout,
    onSuccess: () => {
      queryClient.clear();
      router.push("/login");
    },
    onError: () => {
      // Always redirect on logout, even if API fails
      queryClient.clear();
      router.push("/login");
    },
  });

  // Forgot password mutation
  const forgotPasswordMutation = useMutation({
    mutationFn: (data: PasswordResetRequest) => authApi.forgotPassword(data),
  });

  // Reset password mutation
  const resetPasswordMutation = useMutation({
    mutationFn: (data: PasswordResetConfirm) => authApi.resetPassword(data),
    onSuccess: () => {
      router.push("/login");
    },
  });

  // Verify email mutation
  const verifyEmailMutation = useMutation({
    mutationFn: (token: string) => authApi.verifyEmail(token),
    onSuccess: () => {
      refetchSession();
    },
  });

  // Resend verification email mutation
  const resendVerificationMutation = useMutation({
    mutationFn: authApi.resendVerificationEmail,
  });

  return {
    // User state
    user: userQuery.data ?? (sessionUser as User | undefined),
    isLoading: isLoading || userQuery.isLoading,
    isAuthenticated,
    error: error || userQuery.error,

    // Session data
    session,
    organizationId: session?.organizationId,
    permissions: session?.permissions ?? [],

    // Actions
    login: loginMutation.mutateAsync,
    register: registerMutation.mutateAsync,
    logout: logoutMutation.mutate,
    forgotPassword: forgotPasswordMutation.mutateAsync,
    resetPassword: resetPasswordMutation.mutateAsync,
    verifyEmail: verifyEmailMutation.mutateAsync,
    resendVerification: resendVerificationMutation.mutate,
    refetchUser: userQuery.refetch,
    refetchSession,

    // Mutation states
    isLoggingIn: loginMutation.isPending,
    isRegistering: registerMutation.isPending,
    isLoggingOut: logoutMutation.isPending,
    loginError: loginMutation.error,
    registerError: registerMutation.error,
  };
}

// ============================================
// Auth Guard Hook
// ============================================

export function useRequireAuth(redirectTo = "/login") {
  const { isAuthenticated, isLoading } = useAuth();
  const router = useRouter();

  if (!isLoading && !isAuthenticated) {
    router.push(redirectTo);
  }

  return { isAuthenticated, isLoading };
}

// ============================================
// Permission Check Hook
// ============================================

/**
 * Hook to check if user has a specific permission
 */
export function usePermission(permission: string): boolean {
  const { permissions, isAuthenticated } = useAuth();

  if (!isAuthenticated) return false;
  if (permissions.includes("*")) return true;
  if (permissions.includes(permission)) return true;

  // Check wildcard patterns (e.g., "campaigns:*" matches "campaigns:read")
  const [resource] = permission.split(":");
  if (permissions.includes(`${resource}:*`)) return true;

  return false;
}

/**
 * Hook to check if user has any of the specified roles
 */
export function useRole(roles: string | string[]): boolean {
  const { user, isAuthenticated } = useAuth();
  const roleArray = Array.isArray(roles) ? roles : [roles];

  if (!isAuthenticated || !user?.role) return false;
  return roleArray.includes(user.role);
}

export default useAuth;
