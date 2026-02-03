"use client";

import {
  createContext,
  useContext,
  useEffect,
  useState,
  useCallback,
  useMemo,
  ReactNode,
} from "react";
import { useRouter, usePathname } from "next/navigation";
import { authApi, SessionResponse } from "@/lib/api/services/auth";
import { tokenManager } from "@/lib/api/client";

// ============================================
// Types
// ============================================

interface User {
  id: string;
  email: string;
  role: string;
}

interface AuthState {
  user: User | null;
  organizationId: string | null;
  permissions: string[];
  isAuthenticated: boolean;
  isLoading: boolean;
}

interface AuthContextType extends AuthState {
  login: (email: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
  refreshSession: () => Promise<void>;
}

// ============================================
// Context
// ============================================

const AuthContext = createContext<AuthContextType | undefined>(undefined);

// ============================================
// Constants
// ============================================

const PUBLIC_PATHS = ["/login", "/register", "/forgot-password", "/reset-password"];
const REFRESH_INTERVAL_MS = 4 * 60 * 1000; // Refresh every 4 minutes (before 5 min threshold)

// ============================================
// Provider Component
// ============================================

interface AuthProviderProps {
  children: ReactNode;
}

export function AuthProvider({ children }: AuthProviderProps) {
  const router = useRouter();
  const pathname = usePathname();

  const [state, setState] = useState<AuthState>({
    user: null,
    organizationId: null,
    permissions: [],
    isAuthenticated: false,
    isLoading: true,
  });

  // Check session on mount
  const checkSession = useCallback(async () => {
    try {
      const session: SessionResponse = await authApi.getSession();

      if (session.authenticated && session.user) {
        setState({
          user: session.user,
          organizationId: session.organizationId || null,
          permissions: session.permissions || [],
          isAuthenticated: true,
          isLoading: false,
        });
      } else {
        setState({
          user: null,
          organizationId: null,
          permissions: [],
          isAuthenticated: false,
          isLoading: false,
        });
      }
    } catch (error) {
      console.error("Session check failed:", error);
      setState({
        user: null,
        organizationId: null,
        permissions: [],
        isAuthenticated: false,
        isLoading: false,
      });
    }
  }, []);

  // Refresh session (for auto-refresh)
  const refreshSession = useCallback(async () => {
    try {
      await authApi.refreshToken();
      // After refresh, verify the session is still valid
      await checkSession();
    } catch (error) {
      console.error("Session refresh failed:", error);
      // If refresh fails, clear state and redirect
      setState({
        user: null,
        organizationId: null,
        permissions: [],
        isAuthenticated: false,
        isLoading: false,
      });
      tokenManager.clearTokens();
    }
  }, [checkSession]);

  // Login
  const login = useCallback(async (email: string, password: string) => {
    setState((prev) => ({ ...prev, isLoading: true }));

    try {
      const result = await authApi.login({ email, password });

      setState({
        user: result.user ? {
          id: result.user.id,
          email: result.user.email,
          role: result.user.role || "viewer",
        } : null,
        organizationId: result.organization?.id || null,
        permissions: [],
        isAuthenticated: true,
        isLoading: false,
      });

      router.push("/dashboard");
    } catch (error) {
      setState((prev) => ({ ...prev, isLoading: false }));
      throw error;
    }
  }, [router]);

  // Logout
  const logout = useCallback(async () => {
    try {
      await authApi.logout();
    } catch (error) {
      // Continue with logout even if API fails
      console.error("Logout API call failed:", error);
    } finally {
      setState({
        user: null,
        organizationId: null,
        permissions: [],
        isAuthenticated: false,
        isLoading: false,
      });
      tokenManager.clearTokens();
      router.push("/login");
    }
  }, [router]);

  // Initialize: check session on mount
  useEffect(() => {
    checkSession();
  }, [checkSession]);

  // Auto-refresh token before expiry
  useEffect(() => {
    if (!state.isAuthenticated) return;

    const interval = setInterval(() => {
      if (tokenManager.isTokenExpiringSoon()) {
        refreshSession();
      }
    }, REFRESH_INTERVAL_MS);

    return () => clearInterval(interval);
  }, [state.isAuthenticated, refreshSession]);

  // Note: Redirect logic is handled by NextAuth middleware (middleware.ts)
  // The AuthProvider only manages backend API auth state via cookies
  // Do NOT add redirect logic here as it conflicts with NextAuth middleware

  // Memoize context value
  const contextValue = useMemo<AuthContextType>(
    () => ({
      ...state,
      login,
      logout,
      refreshSession,
    }),
    [state, login, logout, refreshSession]
  );

  return (
    <AuthContext.Provider value={contextValue}>
      {children}
    </AuthContext.Provider>
  );
}

// ============================================
// Hook
// ============================================

export function useAuthContext(): AuthContextType {
  const context = useContext(AuthContext);

  if (context === undefined) {
    throw new Error("useAuthContext must be used within an AuthProvider");
  }

  return context;
}

// ============================================
// Utility Hooks
// ============================================

/**
 * Hook that returns true when auth is ready (not loading)
 */
export function useAuthReady(): boolean {
  const { isLoading } = useAuthContext();
  return !isLoading;
}

/**
 * Hook that returns true when user is authenticated
 */
export function useIsAuthenticated(): boolean {
  const { isAuthenticated, isLoading } = useAuthContext();
  return !isLoading && isAuthenticated;
}

/**
 * Hook that checks if user has required permission
 */
export function useHasPermission(permission: string): boolean {
  const { permissions } = useAuthContext();

  if (permissions.includes("*")) return true;
  if (permissions.includes(permission)) return true;

  // Check wildcard patterns (e.g., "campaigns:*" matches "campaigns:read")
  const [resource] = permission.split(":");
  if (permissions.includes(`${resource}:*`)) return true;

  return false;
}

/**
 * Hook that checks if user has required role
 */
export function useHasRole(roles: string | string[]): boolean {
  const { user } = useAuthContext();
  const roleArray = Array.isArray(roles) ? roles : [roles];
  return user ? roleArray.includes(user.role) : false;
}

export default AuthProvider;
