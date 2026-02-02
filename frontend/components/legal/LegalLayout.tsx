"use client";

import Link from "next/link";
import { useState } from "react";

interface LegalLayoutProps {
  children: React.ReactNode;
  title: string;
  lastUpdated: string;
}

export function LegalLayout({ children, title, lastUpdated }: LegalLayoutProps) {
  const [language, setLanguage] = useState<"ms" | "en">("ms");

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <header className="bg-white border-b border-gray-200">
        <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
          <div className="flex items-center justify-between">
            <Link href="/" className="flex items-center space-x-2">
              <div className="w-8 h-8 bg-gradient-to-br from-blue-600 to-purple-600 rounded-lg flex items-center justify-center">
                <svg className="w-5 h-5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
                </svg>
              </div>
              <span className="text-xl font-bold text-gray-900">AdsAnalytic</span>
            </Link>

            {/* Language Toggle */}
            <div className="flex items-center gap-2 bg-gray-100 rounded-lg p-1">
              <button
                onClick={() => setLanguage("ms")}
                className={`px-3 py-1.5 rounded-md text-sm font-medium transition-colors ${
                  language === "ms"
                    ? "bg-white text-gray-900 shadow-sm"
                    : "text-gray-600 hover:text-gray-900"
                }`}
              >
                BM
              </button>
              <button
                onClick={() => setLanguage("en")}
                className={`px-3 py-1.5 rounded-md text-sm font-medium transition-colors ${
                  language === "en"
                    ? "bg-white text-gray-900 shadow-sm"
                    : "text-gray-600 hover:text-gray-900"
                }`}
              >
                EN
              </button>
            </div>
          </div>
        </div>
      </header>

      {/* Content */}
      <main className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
        <div className="bg-white rounded-2xl shadow-sm border border-gray-200 p-8 md:p-12">
          <h1 className="text-3xl font-bold text-gray-900 mb-2">{title}</h1>
          <p className="text-gray-500 mb-8">
            {language === "ms" ? "Kemas kini terakhir" : "Last updated"}: {lastUpdated}
          </p>

          <div className="prose prose-gray max-w-none" data-language={language}>
            {children}
          </div>
        </div>

        {/* Legal Nav */}
        <div className="mt-8 flex flex-wrap justify-center gap-4 text-sm text-gray-500">
          <Link href="/privacy" className="hover:text-gray-900">
            {language === "ms" ? "Polisi Privasi" : "Privacy Policy"}
          </Link>
          <span>•</span>
          <Link href="/terms" className="hover:text-gray-900">
            {language === "ms" ? "Terma Perkhidmatan" : "Terms of Service"}
          </Link>
          <span>•</span>
          <Link href="/cookies" className="hover:text-gray-900">
            {language === "ms" ? "Polisi Cookie" : "Cookie Policy"}
          </Link>
        </div>
      </main>

      {/* Footer */}
      <footer className="border-t border-gray-200 bg-white mt-12">
        <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-6 text-center text-gray-500 text-sm">
          © {new Date().getFullYear()} AdsAnalytic. {language === "ms" ? "Hak cipta terpelihara." : "All rights reserved."}
        </div>
      </footer>
    </div>
  );
}

interface LegalSectionProps {
  titleMs: string;
  titleEn: string;
  children: React.ReactNode;
}

export function LegalSection({ titleMs, titleEn, children }: LegalSectionProps) {
  return (
    <section className="mb-8">
      <h2 className="text-xl font-semibold text-gray-900 mb-4 [*[data-language='ms']_&]:block [*[data-language='en']_&]:hidden">
        {titleMs}
      </h2>
      <h2 className="text-xl font-semibold text-gray-900 mb-4 [*[data-language='ms']_&]:hidden [*[data-language='en']_&]:block">
        {titleEn}
      </h2>
      {children}
    </section>
  );
}

interface BilingualTextProps {
  ms: React.ReactNode;
  en: React.ReactNode;
}

export function BilingualText({ ms, en }: BilingualTextProps) {
  return (
    <>
      <div className="[*[data-language='ms']_&]:block [*[data-language='en']_&]:hidden">
        {ms}
      </div>
      <div className="[*[data-language='ms']_&]:hidden [*[data-language='en']_&]:block">
        {en}
      </div>
    </>
  );
}
