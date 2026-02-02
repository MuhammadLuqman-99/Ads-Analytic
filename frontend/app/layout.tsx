import type { Metadata } from "next";
import { Inter } from "next/font/google";
import "./globals.css";
import { Providers } from "@/components/providers/Providers";

const inter = Inter({
  subsets: ["latin"],
  display: 'swap',
  variable: '--font-inter',
});

export const metadata: Metadata = {
  title: {
    default: "AdsAnalytic - Semua Ads Data Dalam Satu Dashboard",
    template: "%s | AdsAnalytic",
  },
  description: "Platform analitik iklan terbaik untuk peniaga e-commerce Malaysia. Gabungkan data Meta Ads, TikTok Ads, dan Shopee Ads dalam satu dashboard.",
  keywords: ["ads analytics", "meta ads", "tiktok ads", "shopee ads", "roas", "marketing dashboard", "e-commerce malaysia"],
  metadataBase: new URL("https://adsanalytic.com"),
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="ms">
      <body className={`${inter.variable} font-sans antialiased`}>
        <Providers>
          {children}
        </Providers>
      </body>
    </html>
  );
}
