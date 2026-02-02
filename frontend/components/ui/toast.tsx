"use client";

import * as React from "react";
import * as ToastPrimitive from "@radix-ui/react-toast";
import { X, CheckCircle2, AlertCircle, AlertTriangle, Info } from "lucide-react";
import { cn } from "@/lib/utils";

// Toast types
export type ToastType = "success" | "error" | "warning" | "info";

export interface Toast {
  id: string;
  type: ToastType;
  title: string;
  message?: string;
  duration?: number;
  action?: {
    label: string;
    onClick: () => void;
  };
}

// Toast context
interface ToastContextValue {
  toasts: Toast[];
  addToast: (toast: Omit<Toast, "id">) => string;
  removeToast: (id: string) => void;
  success: (title: string, message?: string) => string;
  error: (title: string, message?: string) => string;
  warning: (title: string, message?: string) => string;
  info: (title: string, message?: string) => string;
}

const ToastContext = React.createContext<ToastContextValue | undefined>(undefined);

// Generate unique ID
let toastId = 0;
const generateId = () => `toast-${++toastId}`;

// Toast Provider
export function ToastProvider({ children }: { children: React.ReactNode }) {
  const [toasts, setToasts] = React.useState<Toast[]>([]);

  const addToast = React.useCallback((toast: Omit<Toast, "id">) => {
    const id = generateId();
    setToasts((prev) => [...prev, { ...toast, id }]);
    return id;
  }, []);

  const removeToast = React.useCallback((id: string) => {
    setToasts((prev) => prev.filter((t) => t.id !== id));
  }, []);

  const success = React.useCallback(
    (title: string, message?: string) => {
      return addToast({ type: "success", title, message, duration: 5000 });
    },
    [addToast]
  );

  const error = React.useCallback(
    (title: string, message?: string) => {
      return addToast({ type: "error", title, message, duration: 8000 });
    },
    [addToast]
  );

  const warning = React.useCallback(
    (title: string, message?: string) => {
      return addToast({ type: "warning", title, message, duration: 6000 });
    },
    [addToast]
  );

  const info = React.useCallback(
    (title: string, message?: string) => {
      return addToast({ type: "info", title, message, duration: 5000 });
    },
    [addToast]
  );

  return (
    <ToastContext.Provider
      value={{ toasts, addToast, removeToast, success, error, warning, info }}
    >
      <ToastPrimitive.Provider swipeDirection="right">
        {children}
        <ToastViewport />
      </ToastPrimitive.Provider>
    </ToastContext.Provider>
  );
}

// Toast Viewport
function ToastViewport() {
  const context = React.useContext(ToastContext);
  if (!context) return null;

  return (
    <>
      {context.toasts.map((toast) => (
        <ToastItem key={toast.id} toast={toast} onClose={() => context.removeToast(toast.id)} />
      ))}
      <ToastPrimitive.Viewport className="fixed bottom-0 right-0 z-[100] flex max-h-screen w-full flex-col-reverse p-4 sm:bottom-0 sm:right-0 sm:top-auto sm:flex-col md:max-w-[420px]" />
    </>
  );
}

// Toast Item
function ToastItem({ toast, onClose }: { toast: Toast; onClose: () => void }) {
  const typeConfig = {
    success: {
      icon: CheckCircle2,
      bgClass: "bg-emerald-50 border-emerald-200",
      iconClass: "text-emerald-600",
      titleClass: "text-emerald-900",
    },
    error: {
      icon: AlertCircle,
      bgClass: "bg-red-50 border-red-200",
      iconClass: "text-red-600",
      titleClass: "text-red-900",
    },
    warning: {
      icon: AlertTriangle,
      bgClass: "bg-amber-50 border-amber-200",
      iconClass: "text-amber-600",
      titleClass: "text-amber-900",
    },
    info: {
      icon: Info,
      bgClass: "bg-blue-50 border-blue-200",
      iconClass: "text-blue-600",
      titleClass: "text-blue-900",
    },
  };

  const config = typeConfig[toast.type];
  const Icon = config.icon;

  return (
    <ToastPrimitive.Root
      duration={toast.duration || 5000}
      onOpenChange={(open) => {
        if (!open) onClose();
      }}
      className={cn(
        "group pointer-events-auto relative flex w-full items-start gap-3 overflow-hidden rounded-lg border p-4 shadow-lg transition-all",
        "data-[swipe=cancel]:translate-x-0 data-[swipe=end]:translate-x-[var(--radix-toast-swipe-end-x)] data-[swipe=move]:translate-x-[var(--radix-toast-swipe-move-x)] data-[swipe=move]:transition-none",
        "data-[state=open]:animate-in data-[state=closed]:animate-out data-[swipe=end]:animate-out data-[state=closed]:fade-out-80 data-[state=closed]:slide-out-to-right-full data-[state=open]:slide-in-from-bottom-full",
        config.bgClass
      )}
    >
      <Icon className={cn("h-5 w-5 flex-shrink-0 mt-0.5", config.iconClass)} />

      <div className="flex-1 space-y-1">
        <ToastPrimitive.Title className={cn("text-sm font-semibold", config.titleClass)}>
          {toast.title}
        </ToastPrimitive.Title>
        {toast.message && (
          <ToastPrimitive.Description className="text-sm text-slate-600">
            {toast.message}
          </ToastPrimitive.Description>
        )}
        {toast.action && (
          <ToastPrimitive.Action
            altText={toast.action.label}
            onClick={toast.action.onClick}
            className="mt-2 inline-flex h-8 items-center justify-center rounded-md border bg-transparent px-3 text-sm font-medium transition-colors hover:bg-slate-100 focus:outline-none focus:ring-2 focus:ring-slate-400 focus:ring-offset-2"
          >
            {toast.action.label}
          </ToastPrimitive.Action>
        )}
      </div>

      <ToastPrimitive.Close className="absolute right-2 top-2 rounded-md p-1 opacity-0 transition-opacity hover:bg-slate-100 focus:opacity-100 focus:outline-none focus:ring-2 group-hover:opacity-100">
        <X className="h-4 w-4 text-slate-500" />
      </ToastPrimitive.Close>
    </ToastPrimitive.Root>
  );
}

// Hook to use toast
export function useToast() {
  const context = React.useContext(ToastContext);
  if (!context) {
    throw new Error("useToast must be used within a ToastProvider");
  }
  return context;
}

// Standalone toast function (for use outside of React components)
let toastRef: ToastContextValue | null = null;

export function setToastRef(ref: ToastContextValue) {
  toastRef = ref;
}

export const toast = {
  success: (title: string, message?: string) => toastRef?.success(title, message),
  error: (title: string, message?: string) => toastRef?.error(title, message),
  warning: (title: string, message?: string) => toastRef?.warning(title, message),
  info: (title: string, message?: string) => toastRef?.info(title, message),
};

export default ToastProvider;
