import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "AdsAnalytics - Multi-Platform Ad Analytics Dashboard",
  description: "Unified analytics dashboard for Meta, TikTok, and Shopee advertising campaigns",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" className="dark">
      <body className="bg-slate-900 antialiased">
        {children}
      </body>
    </html>
  );
}
