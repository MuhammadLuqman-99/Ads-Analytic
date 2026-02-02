import {
  Header,
  Hero,
  Problem,
  Features,
  Platforms,
  Pricing,
  FAQ,
  Footer,
} from "@/components/landing";
import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "AdsAnalytic - Semua Ads Data Dalam Satu Dashboard",
  description:
    "Platform analitik iklan terbaik untuk peniaga e-commerce Malaysia. Gabungkan data Meta Ads, TikTok Ads, dan Shopee Ads dalam satu dashboard. Jimat masa, tingkatkan ROAS.",
  keywords: [
    "ads analytics malaysia",
    "meta ads dashboard",
    "tiktok ads analytics",
    "shopee ads tracking",
    "roas calculator",
    "e-commerce analytics",
    "iklan facebook",
    "iklan tiktok",
    "analitik shopee",
    "dashboard iklan",
    "multi-platform ads",
  ],
  authors: [{ name: "AdsAnalytic" }],
  creator: "AdsAnalytic",
  publisher: "AdsAnalytic",
  openGraph: {
    type: "website",
    locale: "ms_MY",
    url: "https://adsanalytic.com",
    siteName: "AdsAnalytic",
    title: "AdsAnalytic - Semua Ads Data Dalam Satu Dashboard",
    description:
      "Platform analitik iklan terbaik untuk peniaga e-commerce Malaysia. Gabungkan data Meta Ads, TikTok Ads, dan Shopee Ads dalam satu dashboard.",
    images: [
      {
        url: "/og-image.png",
        width: 1200,
        height: 630,
        alt: "AdsAnalytic Dashboard Preview",
      },
    ],
  },
  twitter: {
    card: "summary_large_image",
    title: "AdsAnalytic - Semua Ads Data Dalam Satu Dashboard",
    description:
      "Platform analitik iklan terbaik untuk peniaga e-commerce Malaysia.",
    images: ["/og-image.png"],
  },
  robots: {
    index: true,
    follow: true,
    googleBot: {
      index: true,
      follow: true,
      "max-video-preview": -1,
      "max-image-preview": "large",
      "max-snippet": -1,
    },
  },
  alternates: {
    canonical: "https://adsanalytic.com",
    languages: {
      "ms-MY": "https://adsanalytic.com",
      "en-US": "https://adsanalytic.com/en",
    },
  },
};

export default function LandingPage() {
  return (
    <main className="min-h-screen">
      <Header />
      <Hero />
      <Problem />
      <Features />
      <Platforms />
      <Pricing />
      <FAQ />
      <Footer />
    </main>
  );
}
