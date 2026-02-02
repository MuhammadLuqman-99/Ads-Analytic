"use client";

import { useState } from "react";
import { AlertTriangle, Trash2, X, AlertCircle, HelpCircle } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { cn } from "@/lib/utils";

type DialogVariant = "danger" | "warning" | "info";

interface ConfirmDialogProps {
  isOpen: boolean;
  onClose: () => void;
  onConfirm: () => void | Promise<void>;
  title: string;
  description?: string;
  confirmLabel?: string;
  cancelLabel?: string;
  variant?: DialogVariant;
  confirmText?: string; // Text user must type to confirm (for dangerous actions)
  isLoading?: boolean;
  className?: string;
}

const variantConfig: Record<DialogVariant, {
  icon: typeof AlertTriangle;
  iconBg: string;
  iconColor: string;
  confirmButton: string;
}> = {
  danger: {
    icon: Trash2,
    iconBg: "bg-red-100",
    iconColor: "text-red-600",
    confirmButton: "bg-red-600 hover:bg-red-700 text-white",
  },
  warning: {
    icon: AlertTriangle,
    iconBg: "bg-amber-100",
    iconColor: "text-amber-600",
    confirmButton: "bg-amber-600 hover:bg-amber-700 text-white",
  },
  info: {
    icon: HelpCircle,
    iconBg: "bg-blue-100",
    iconColor: "text-blue-600",
    confirmButton: "bg-blue-600 hover:bg-blue-700 text-white",
  },
};

export function ConfirmDialog({
  isOpen,
  onClose,
  onConfirm,
  title,
  description,
  confirmLabel = "Confirm",
  cancelLabel = "Cancel",
  variant = "danger",
  confirmText,
  isLoading = false,
  className,
}: ConfirmDialogProps) {
  const [inputValue, setInputValue] = useState("");
  const [isConfirming, setIsConfirming] = useState(false);

  const config = variantConfig[variant];
  const Icon = config.icon;

  const canConfirm = confirmText ? inputValue === confirmText : true;

  const handleConfirm = async () => {
    setIsConfirming(true);
    try {
      await onConfirm();
      setInputValue("");
      onClose();
    } finally {
      setIsConfirming(false);
    }
  };

  const handleClose = () => {
    setInputValue("");
    onClose();
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      {/* Backdrop */}
      <div
        className="absolute inset-0 bg-black/50 backdrop-blur-sm"
        onClick={handleClose}
      />

      {/* Dialog */}
      <div
        className={cn(
          "relative bg-white rounded-xl shadow-xl max-w-md w-full mx-4",
          className
        )}
      >
        {/* Close button */}
        <button
          onClick={handleClose}
          className="absolute top-4 right-4 text-slate-400 hover:text-slate-600"
        >
          <X className="h-5 w-5" />
        </button>

        <div className="p-6">
          {/* Icon */}
          <div
            className={cn(
              "h-12 w-12 rounded-full flex items-center justify-center mx-auto mb-4",
              config.iconBg
            )}
          >
            <Icon className={cn("h-6 w-6", config.iconColor)} />
          </div>

          {/* Title */}
          <h3 className="text-lg font-semibold text-slate-900 text-center mb-2">
            {title}
          </h3>

          {/* Description */}
          {description && (
            <p className="text-slate-500 text-center mb-4">{description}</p>
          )}

          {/* Confirm text input */}
          {confirmText && (
            <div className="mb-6">
              <p className="text-sm text-slate-600 mb-2">
                Type{" "}
                <span className="font-mono font-bold text-slate-900">
                  {confirmText}
                </span>{" "}
                to confirm:
              </p>
              <Input
                value={inputValue}
                onChange={(e) => setInputValue(e.target.value)}
                placeholder={confirmText}
                className={cn(
                  variant === "danger" && "border-red-200 focus:border-red-500"
                )}
              />
            </div>
          )}

          {/* Actions */}
          <div className="flex items-center gap-3 justify-end">
            <Button variant="outline" onClick={handleClose} disabled={isConfirming}>
              {cancelLabel}
            </Button>
            <Button
              className={config.confirmButton}
              onClick={handleConfirm}
              disabled={!canConfirm || isConfirming || isLoading}
            >
              {isConfirming || isLoading ? "Processing..." : confirmLabel}
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
}

// Delete confirmation preset
export function DeleteConfirmDialog({
  isOpen,
  onClose,
  onConfirm,
  itemName,
  requireConfirmText = false,
  isLoading,
}: {
  isOpen: boolean;
  onClose: () => void;
  onConfirm: () => void | Promise<void>;
  itemName: string;
  requireConfirmText?: boolean;
  isLoading?: boolean;
}) {
  return (
    <ConfirmDialog
      isOpen={isOpen}
      onClose={onClose}
      onConfirm={onConfirm}
      title={`Delete ${itemName}?`}
      description={`This action cannot be undone. ${itemName} will be permanently deleted.`}
      confirmLabel="Delete"
      variant="danger"
      confirmText={requireConfirmText ? "DELETE" : undefined}
      isLoading={isLoading}
    />
  );
}

// Hook for easy dialog management
export function useConfirmDialog() {
  const [isOpen, setIsOpen] = useState(false);
  const [config, setConfig] = useState<Omit<ConfirmDialogProps, "isOpen" | "onClose">>({
    onConfirm: () => {},
    title: "",
  });

  const confirm = (options: Omit<ConfirmDialogProps, "isOpen" | "onClose">) => {
    setConfig(options);
    setIsOpen(true);
  };

  const close = () => {
    setIsOpen(false);
  };

  const Dialog = () => (
    <ConfirmDialog isOpen={isOpen} onClose={close} {...config} />
  );

  return { confirm, close, Dialog, isOpen };
}
