import type { Metadata } from "next";
import { Inter } from "next/font/google";
import "./globals.css";
import { QueryProvider } from "@/lib/query-provider";

const inter = Inter({
  subsets: ["latin"],
  display: 'swap',
  variable: '--font-inter',
});

export const metadata: Metadata = {
  title: "AdsAnalytics - Multi-Platform Ad Analytics Dashboard",
  description: "Unified analytics dashboard for Meta, TikTok, and Shopee advertising campaigns. Track spend, ROAS, conversions, and performance metrics in one place.",
  keywords: ["ads analytics", "meta ads", "tiktok ads", "shopee ads", "roas", "marketing dashboard"],
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" className="dark">
      <body className={`${inter.variable} font-sans bg-slate-900 antialiased`}>
        <QueryProvider>
          {children}
        </QueryProvider>
      </body>
    </html>
  );
}
