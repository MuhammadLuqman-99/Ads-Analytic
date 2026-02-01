"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

interface PlatformConnectCardProps {
  platform: "meta" | "tiktok" | "shopee";
  name: string;
  description: string;
  icon: React.ReactNode;
  isConnected?: boolean;
  onConnect: () => Promise<void>;
}

export function PlatformConnectCard({
  platform,
  name,
  description,
  icon,
  isConnected = false,
  onConnect,
}: PlatformConnectCardProps) {
  const [isLoading, setIsLoading] = useState(false);

  const handleConnect = async () => {
    setIsLoading(true);
    try {
      await onConnect();
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div
      className={cn(
        "relative p-6 rounded-xl border-2 transition-all",
        isConnected
          ? "border-green-500 bg-green-50"
          : "border-slate-200 hover:border-blue-300 hover:shadow-md"
      )}
    >
      {isConnected && (
        <div className="absolute top-3 right-3">
          <div className="w-6 h-6 bg-green-500 rounded-full flex items-center justify-center">
            <svg
              className="w-4 h-4 text-white"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M5 13l4 4L19 7"
              />
            </svg>
          </div>
        </div>
      )}

      <div className="flex flex-col items-center text-center space-y-4">
        <div className="w-16 h-16 rounded-2xl bg-slate-100 flex items-center justify-center">
          {icon}
        </div>

        <div>
          <h3 className="font-semibold text-lg text-slate-900">{name}</h3>
          <p className="text-sm text-slate-500 mt-1">{description}</p>
        </div>

        <Button
          variant={isConnected ? "outline" : "default"}
          className="w-full"
          onClick={handleConnect}
          disabled={isLoading || isConnected}
        >
          {isLoading ? (
            <svg
              className="animate-spin h-4 w-4 mr-2"
              xmlns="http://www.w3.org/2000/svg"
              fill="none"
              viewBox="0 0 24 24"
            >
              <circle
                className="opacity-25"
                cx="12"
                cy="12"
                r="10"
                stroke="currentColor"
                strokeWidth="4"
              />
              <path
                className="opacity-75"
                fill="currentColor"
                d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
              />
            </svg>
          ) : isConnected ? (
            "Connected"
          ) : (
            "Connect"
          )}
        </Button>
      </div>
    </div>
  );
}

// Platform icons
export const MetaIcon = () => (
  <svg className="w-8 h-8" viewBox="0 0 24 24" fill="none">
    <path
      d="M12 2C6.477 2 2 6.477 2 12c0 4.991 3.657 9.128 8.438 9.879V14.89h-2.54V12h2.54V9.797c0-2.506 1.492-3.89 3.777-3.89 1.094 0 2.238.195 2.238.195v2.46h-1.26c-1.243 0-1.63.771-1.63 1.562V12h2.773l-.443 2.89h-2.33v6.989C18.343 21.129 22 16.99 22 12c0-5.523-4.477-10-10-10z"
      fill="#1877F2"
    />
  </svg>
);

export const TikTokIcon = () => (
  <svg className="w-8 h-8" viewBox="0 0 24 24" fill="none">
    <path
      d="M19.59 6.69a4.83 4.83 0 01-3.77-4.25V2h-3.45v13.67a2.89 2.89 0 01-5.2 1.74 2.89 2.89 0 012.31-4.64 2.93 2.93 0 01.88.13V9.4a6.84 6.84 0 00-1-.05A6.33 6.33 0 005 20.1a6.34 6.34 0 0010.86-4.43v-7a8.16 8.16 0 004.77 1.52v-3.4a4.85 4.85 0 01-1-.1z"
      fill="#000000"
    />
  </svg>
);

export const ShopeeIcon = () => (
  <svg className="w-8 h-8" viewBox="0 0 24 24" fill="none">
    <path
      d="M12 2C9.243 2 7 4.243 7 7h2c0-1.654 1.346-3 3-3s3 1.346 3 3h2c0-2.757-2.243-5-5-5z"
      fill="#EE4D2D"
    />
    <path
      d="M20 7H4c-1.103 0-2 .897-2 2v11c0 1.103.897 2 2 2h16c1.103 0 2-.897 2-2V9c0-1.103-.897-2-2-2zm-8 11c-2.206 0-4-1.794-4-4s1.794-4 4-4 4 1.794 4 4-1.794 4-4 4z"
      fill="#EE4D2D"
    />
  </svg>
);
