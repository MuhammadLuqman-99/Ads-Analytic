"use client";

import Link from "next/link";
import { useState } from "react";

export default function PrivacyPolicyPage() {
  const [lang, setLang] = useState<"ms" | "en">("ms");

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <header className="bg-white border-b border-gray-200 sticky top-0 z-10">
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

            <div className="flex items-center gap-2 bg-gray-100 rounded-lg p-1">
              <button
                onClick={() => setLang("ms")}
                className={`px-3 py-1.5 rounded-md text-sm font-medium transition-colors ${
                  lang === "ms" ? "bg-white text-gray-900 shadow-sm" : "text-gray-600"
                }`}
              >
                BM
              </button>
              <button
                onClick={() => setLang("en")}
                className={`px-3 py-1.5 rounded-md text-sm font-medium transition-colors ${
                  lang === "en" ? "bg-white text-gray-900 shadow-sm" : "text-gray-600"
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
          <h1 className="text-3xl font-bold text-gray-900 mb-2">
            {lang === "ms" ? "Polisi Privasi" : "Privacy Policy"}
          </h1>
          <p className="text-gray-500 mb-8">
            {lang === "ms" ? "Kemas kini terakhir" : "Last updated"}: 1 Februari 2026
          </p>

          <div className="prose prose-gray max-w-none">
            {lang === "ms" ? <PrivacyContentMS /> : <PrivacyContentEN />}
          </div>
        </div>

        {/* Legal Nav */}
        <div className="mt-8 flex flex-wrap justify-center gap-4 text-sm text-gray-500">
          <Link href="/privacy" className="text-blue-600 font-medium">
            {lang === "ms" ? "Polisi Privasi" : "Privacy Policy"}
          </Link>
          <span>•</span>
          <Link href="/terms" className="hover:text-gray-900">
            {lang === "ms" ? "Terma Perkhidmatan" : "Terms of Service"}
          </Link>
          <span>•</span>
          <Link href="/cookies" className="hover:text-gray-900">
            {lang === "ms" ? "Polisi Cookie" : "Cookie Policy"}
          </Link>
        </div>
      </main>
    </div>
  );
}

function PrivacyContentMS() {
  return (
    <>
      <section className="mb-8">
        <h2 className="text-xl font-semibold text-gray-900 mb-4">1. Pengenalan</h2>
        <p className="text-gray-600 mb-4">
          AdsAnalytic (&quot;kami&quot;, &quot;kita&quot;, atau &quot;Syarikat&quot;) komited untuk melindungi privasi anda.
          Polisi Privasi ini menerangkan bagaimana kami mengumpul, menggunakan, mendedahkan, dan melindungi
          maklumat anda apabila anda menggunakan platform AdsAnalytic.
        </p>
        <p className="text-gray-600">
          Kami mematuhi Akta Perlindungan Data Peribadi 2010 (PDPA) Malaysia dan komited untuk memastikan
          data peribadi anda dilindungi mengikut undang-undang yang terpakai.
        </p>
      </section>

      <section className="mb-8">
        <h2 className="text-xl font-semibold text-gray-900 mb-4">2. Data Yang Dikumpul</h2>
        <p className="text-gray-600 mb-4">Kami mengumpul jenis data berikut:</p>

        <h3 className="text-lg font-medium text-gray-800 mb-2">2.1 Maklumat Akaun</h3>
        <ul className="list-disc list-inside text-gray-600 mb-4 space-y-1">
          <li>Nama dan alamat emel</li>
          <li>Nama syarikat/perniagaan</li>
          <li>Nombor telefon (pilihan)</li>
          <li>Kata laluan (disimpan dalam bentuk hash yang selamat)</li>
        </ul>

        <h3 className="text-lg font-medium text-gray-800 mb-2">2.2 Data Platform Iklan</h3>
        <ul className="list-disc list-inside text-gray-600 mb-4 space-y-1">
          <li>Data prestasi kempen (impressions, clicks, spend, conversions)</li>
          <li>ID akaun iklan dan nama kempen</li>
          <li>Metrik ROAS, CTR, CPA, dan lain-lain</li>
          <li>Data demografi agregat (tanpa maklumat peribadi pengguna akhir)</li>
        </ul>

        <h3 className="text-lg font-medium text-gray-800 mb-2">2.3 OAuth Tokens</h3>
        <ul className="list-disc list-inside text-gray-600 space-y-1">
          <li>Access tokens dan refresh tokens untuk Meta, TikTok, dan Shopee</li>
          <li>Token ini disimpan menggunakan enkripsi AES-256-GCM</li>
          <li>Kami tidak menyimpan kata laluan platform anda</li>
        </ul>
      </section>

      <section className="mb-8">
        <h2 className="text-xl font-semibold text-gray-900 mb-4">3. Bagaimana OAuth Tokens Disimpan</h2>
        <p className="text-gray-600 mb-4">
          Keselamatan token adalah keutamaan kami. Berikut adalah langkah-langkah yang kami ambil:
        </p>
        <ul className="list-disc list-inside text-gray-600 space-y-2">
          <li><strong>Enkripsi AES-256-GCM:</strong> Semua token dienkripsi sebelum disimpan dalam pangkalan data</li>
          <li><strong>Kunci enkripsi berasingan:</strong> Kunci enkripsi disimpan secara berasingan daripada data</li>
          <li><strong>Akses terhad:</strong> Hanya sistem automatik yang boleh mendekripsi token untuk sync data</li>
          <li><strong>Tiada akses manual:</strong> Pasukan kami tidak boleh melihat token anda</li>
          <li><strong>Pembatalan segera:</strong> Anda boleh batalkan akses bila-bila masa melalui platform asal</li>
        </ul>
      </section>

      <section className="mb-8">
        <h2 className="text-xl font-semibold text-gray-900 mb-4">4. API Pihak Ketiga</h2>
        <p className="text-gray-600 mb-4">Kami mengintegrasikan dengan platform berikut melalui API rasmi mereka:</p>

        <div className="bg-gray-50 rounded-lg p-4 mb-4">
          <h4 className="font-medium text-gray-900 mb-2">Meta (Facebook/Instagram)</h4>
          <p className="text-gray-600 text-sm">
            Kami menggunakan Meta Marketing API untuk mengambil data iklan.
            <a href="https://www.facebook.com/privacy/policy" className="text-blue-600 hover:underline ml-1">
              Polisi Privasi Meta
            </a>
          </p>
        </div>

        <div className="bg-gray-50 rounded-lg p-4 mb-4">
          <h4 className="font-medium text-gray-900 mb-2">TikTok</h4>
          <p className="text-gray-600 text-sm">
            Kami menggunakan TikTok Marketing API untuk mengambil data iklan.
            <a href="https://www.tiktok.com/legal/privacy-policy" className="text-blue-600 hover:underline ml-1">
              Polisi Privasi TikTok
            </a>
          </p>
        </div>

        <div className="bg-gray-50 rounded-lg p-4">
          <h4 className="font-medium text-gray-900 mb-2">Shopee</h4>
          <p className="text-gray-600 text-sm">
            Kami menggunakan Shopee Open Platform API untuk mengambil data iklan.
            <a href="https://shopee.com.my/docs/privacy" className="text-blue-600 hover:underline ml-1">
              Polisi Privasi Shopee
            </a>
          </p>
        </div>
      </section>

      <section className="mb-8">
        <h2 className="text-xl font-semibold text-gray-900 mb-4">5. Pematuhan PDPA Malaysia</h2>
        <p className="text-gray-600 mb-4">
          Sebagai syarikat yang beroperasi di Malaysia, kami mematuhi Akta Perlindungan Data Peribadi 2010 (PDPA):
        </p>
        <ul className="list-disc list-inside text-gray-600 space-y-2">
          <li><strong>Prinsip Am:</strong> Data dikumpul untuk tujuan yang sah dan jelas</li>
          <li><strong>Prinsip Notis dan Pilihan:</strong> Anda diberitahu tentang pengumpulan data dan boleh memilih</li>
          <li><strong>Prinsip Pendedahan:</strong> Data tidak didedahkan tanpa persetujuan</li>
          <li><strong>Prinsip Keselamatan:</strong> Langkah keselamatan yang munasabah diambil</li>
          <li><strong>Prinsip Penyimpanan:</strong> Data tidak disimpan lebih lama daripada yang diperlukan</li>
          <li><strong>Prinsip Integriti Data:</strong> Data dijaga tepat dan terkini</li>
          <li><strong>Prinsip Akses:</strong> Anda boleh mengakses dan membetulkan data anda</li>
        </ul>
      </section>

      <section className="mb-8">
        <h2 className="text-xl font-semibold text-gray-900 mb-4">6. Polisi Penyimpanan Data</h2>
        <p className="text-gray-600 mb-4">Kami menyimpan data anda mengikut tempoh berikut:</p>

        <div className="overflow-x-auto">
          <table className="min-w-full border border-gray-200 rounded-lg">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-4 py-3 text-left text-sm font-medium text-gray-900">Jenis Data</th>
                <th className="px-4 py-3 text-left text-sm font-medium text-gray-900">Tempoh Penyimpanan</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200">
              <tr>
                <td className="px-4 py-3 text-sm text-gray-600">Maklumat akaun</td>
                <td className="px-4 py-3 text-sm text-gray-600">Sehingga akaun dipadam</td>
              </tr>
              <tr>
                <td className="px-4 py-3 text-sm text-gray-600">Data kempen iklan</td>
                <td className="px-4 py-3 text-sm text-gray-600">2 tahun atau mengikut pelan langganan</td>
              </tr>
              <tr>
                <td className="px-4 py-3 text-sm text-gray-600">OAuth tokens</td>
                <td className="px-4 py-3 text-sm text-gray-600">Sehingga dibatalkan atau tamat tempoh</td>
              </tr>
              <tr>
                <td className="px-4 py-3 text-sm text-gray-600">Log aktiviti</td>
                <td className="px-4 py-3 text-sm text-gray-600">90 hari</td>
              </tr>
              <tr>
                <td className="px-4 py-3 text-sm text-gray-600">Data pembayaran</td>
                <td className="px-4 py-3 text-sm text-gray-600">7 tahun (keperluan undang-undang)</td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>

      <section className="mb-8">
        <h2 className="text-xl font-semibold text-gray-900 mb-4">7. Hak Anda</h2>
        <p className="text-gray-600 mb-4">Anda mempunyai hak berikut berkaitan data peribadi anda:</p>

        <div className="space-y-4">
          <div className="bg-blue-50 border border-blue-100 rounded-lg p-4">
            <h4 className="font-medium text-blue-900 mb-1">Hak untuk Mengakses</h4>
            <p className="text-blue-800 text-sm">
              Anda boleh meminta salinan semua data peribadi yang kami simpan tentang anda.
            </p>
          </div>

          <div className="bg-green-50 border border-green-100 rounded-lg p-4">
            <h4 className="font-medium text-green-900 mb-1">Hak untuk Eksport Data</h4>
            <p className="text-green-800 text-sm">
              Anda boleh mengeksport semua data anda dalam format JSON atau CSV melalui tetapan akaun.
            </p>
          </div>

          <div className="bg-red-50 border border-red-100 rounded-lg p-4">
            <h4 className="font-medium text-red-900 mb-1">Hak untuk Memadam</h4>
            <p className="text-red-800 text-sm">
              Anda boleh meminta pemadaman akaun dan semua data berkaitan. Proses ini tidak boleh diterbalikkan.
            </p>
          </div>

          <div className="bg-yellow-50 border border-yellow-100 rounded-lg p-4">
            <h4 className="font-medium text-yellow-900 mb-1">Hak untuk Membetulkan</h4>
            <p className="text-yellow-800 text-sm">
              Anda boleh mengemas kini atau membetulkan maklumat peribadi anda pada bila-bila masa.
            </p>
          </div>
        </div>

        <p className="text-gray-600 mt-4">
          Untuk melaksanakan hak-hak ini, sila hubungi kami di{" "}
          <a href="mailto:privacy@adsanalytic.com" className="text-blue-600 hover:underline">
            privacy@adsanalytic.com
          </a>
        </p>
      </section>

      <section className="mb-8">
        <h2 className="text-xl font-semibold text-gray-900 mb-4">8. Hubungi Kami</h2>
        <p className="text-gray-600 mb-4">
          Jika anda mempunyai sebarang soalan tentang Polisi Privasi ini, sila hubungi:
        </p>
        <div className="bg-gray-50 rounded-lg p-4">
          <p className="text-gray-700"><strong>Pegawai Perlindungan Data</strong></p>
          <p className="text-gray-600">AdsAnalytic Sdn Bhd</p>
          <p className="text-gray-600">Emel: privacy@adsanalytic.com</p>
        </div>
      </section>
    </>
  );
}

function PrivacyContentEN() {
  return (
    <>
      <section className="mb-8">
        <h2 className="text-xl font-semibold text-gray-900 mb-4">1. Introduction</h2>
        <p className="text-gray-600 mb-4">
          AdsAnalytic (&quot;we&quot;, &quot;us&quot;, or &quot;Company&quot;) is committed to protecting your privacy.
          This Privacy Policy explains how we collect, use, disclose, and protect your information
          when you use the AdsAnalytic platform.
        </p>
        <p className="text-gray-600">
          We comply with the Malaysian Personal Data Protection Act 2010 (PDPA) and are committed to
          ensuring your personal data is protected in accordance with applicable laws.
        </p>
      </section>

      <section className="mb-8">
        <h2 className="text-xl font-semibold text-gray-900 mb-4">2. Data We Collect</h2>
        <p className="text-gray-600 mb-4">We collect the following types of data:</p>

        <h3 className="text-lg font-medium text-gray-800 mb-2">2.1 Account Information</h3>
        <ul className="list-disc list-inside text-gray-600 mb-4 space-y-1">
          <li>Name and email address</li>
          <li>Company/business name</li>
          <li>Phone number (optional)</li>
          <li>Password (stored in secure hashed form)</li>
        </ul>

        <h3 className="text-lg font-medium text-gray-800 mb-2">2.2 Advertising Platform Data</h3>
        <ul className="list-disc list-inside text-gray-600 mb-4 space-y-1">
          <li>Campaign performance data (impressions, clicks, spend, conversions)</li>
          <li>Ad account IDs and campaign names</li>
          <li>ROAS, CTR, CPA, and other metrics</li>
          <li>Aggregate demographic data (without end-user personal information)</li>
        </ul>

        <h3 className="text-lg font-medium text-gray-800 mb-2">2.3 OAuth Tokens</h3>
        <ul className="list-disc list-inside text-gray-600 space-y-1">
          <li>Access tokens and refresh tokens for Meta, TikTok, and Shopee</li>
          <li>These tokens are stored using AES-256-GCM encryption</li>
          <li>We do not store your platform passwords</li>
        </ul>
      </section>

      <section className="mb-8">
        <h2 className="text-xl font-semibold text-gray-900 mb-4">3. How OAuth Tokens Are Stored</h2>
        <p className="text-gray-600 mb-4">
          Token security is our priority. Here are the measures we take:
        </p>
        <ul className="list-disc list-inside text-gray-600 space-y-2">
          <li><strong>AES-256-GCM Encryption:</strong> All tokens are encrypted before storage in the database</li>
          <li><strong>Separate encryption keys:</strong> Encryption keys are stored separately from the data</li>
          <li><strong>Limited access:</strong> Only automated systems can decrypt tokens for data sync</li>
          <li><strong>No manual access:</strong> Our team cannot view your tokens</li>
          <li><strong>Immediate revocation:</strong> You can revoke access anytime through the original platform</li>
        </ul>
      </section>

      <section className="mb-8">
        <h2 className="text-xl font-semibold text-gray-900 mb-4">4. Third-Party APIs</h2>
        <p className="text-gray-600 mb-4">We integrate with the following platforms through their official APIs:</p>

        <div className="bg-gray-50 rounded-lg p-4 mb-4">
          <h4 className="font-medium text-gray-900 mb-2">Meta (Facebook/Instagram)</h4>
          <p className="text-gray-600 text-sm">
            We use the Meta Marketing API to retrieve advertising data.
            <a href="https://www.facebook.com/privacy/policy" className="text-blue-600 hover:underline ml-1">
              Meta Privacy Policy
            </a>
          </p>
        </div>

        <div className="bg-gray-50 rounded-lg p-4 mb-4">
          <h4 className="font-medium text-gray-900 mb-2">TikTok</h4>
          <p className="text-gray-600 text-sm">
            We use the TikTok Marketing API to retrieve advertising data.
            <a href="https://www.tiktok.com/legal/privacy-policy" className="text-blue-600 hover:underline ml-1">
              TikTok Privacy Policy
            </a>
          </p>
        </div>

        <div className="bg-gray-50 rounded-lg p-4">
          <h4 className="font-medium text-gray-900 mb-2">Shopee</h4>
          <p className="text-gray-600 text-sm">
            We use the Shopee Open Platform API to retrieve advertising data.
            <a href="https://shopee.com.my/docs/privacy" className="text-blue-600 hover:underline ml-1">
              Shopee Privacy Policy
            </a>
          </p>
        </div>
      </section>

      <section className="mb-8">
        <h2 className="text-xl font-semibold text-gray-900 mb-4">5. Malaysian PDPA Compliance</h2>
        <p className="text-gray-600 mb-4">
          As a company operating in Malaysia, we comply with the Personal Data Protection Act 2010 (PDPA):
        </p>
        <ul className="list-disc list-inside text-gray-600 space-y-2">
          <li><strong>General Principle:</strong> Data is collected for lawful and clear purposes</li>
          <li><strong>Notice and Choice Principle:</strong> You are informed about data collection and can choose</li>
          <li><strong>Disclosure Principle:</strong> Data is not disclosed without consent</li>
          <li><strong>Security Principle:</strong> Reasonable security measures are taken</li>
          <li><strong>Retention Principle:</strong> Data is not stored longer than necessary</li>
          <li><strong>Data Integrity Principle:</strong> Data is kept accurate and up-to-date</li>
          <li><strong>Access Principle:</strong> You can access and correct your data</li>
        </ul>
      </section>

      <section className="mb-8">
        <h2 className="text-xl font-semibold text-gray-900 mb-4">6. Data Retention Policy</h2>
        <p className="text-gray-600 mb-4">We retain your data according to the following periods:</p>

        <div className="overflow-x-auto">
          <table className="min-w-full border border-gray-200 rounded-lg">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-4 py-3 text-left text-sm font-medium text-gray-900">Data Type</th>
                <th className="px-4 py-3 text-left text-sm font-medium text-gray-900">Retention Period</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200">
              <tr>
                <td className="px-4 py-3 text-sm text-gray-600">Account information</td>
                <td className="px-4 py-3 text-sm text-gray-600">Until account deletion</td>
              </tr>
              <tr>
                <td className="px-4 py-3 text-sm text-gray-600">Campaign data</td>
                <td className="px-4 py-3 text-sm text-gray-600">2 years or per subscription plan</td>
              </tr>
              <tr>
                <td className="px-4 py-3 text-sm text-gray-600">OAuth tokens</td>
                <td className="px-4 py-3 text-sm text-gray-600">Until revoked or expired</td>
              </tr>
              <tr>
                <td className="px-4 py-3 text-sm text-gray-600">Activity logs</td>
                <td className="px-4 py-3 text-sm text-gray-600">90 days</td>
              </tr>
              <tr>
                <td className="px-4 py-3 text-sm text-gray-600">Payment data</td>
                <td className="px-4 py-3 text-sm text-gray-600">7 years (legal requirement)</td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>

      <section className="mb-8">
        <h2 className="text-xl font-semibold text-gray-900 mb-4">7. Your Rights</h2>
        <p className="text-gray-600 mb-4">You have the following rights regarding your personal data:</p>

        <div className="space-y-4">
          <div className="bg-blue-50 border border-blue-100 rounded-lg p-4">
            <h4 className="font-medium text-blue-900 mb-1">Right to Access</h4>
            <p className="text-blue-800 text-sm">
              You can request a copy of all personal data we hold about you.
            </p>
          </div>

          <div className="bg-green-50 border border-green-100 rounded-lg p-4">
            <h4 className="font-medium text-green-900 mb-1">Right to Export Data</h4>
            <p className="text-green-800 text-sm">
              You can export all your data in JSON or CSV format through account settings.
            </p>
          </div>

          <div className="bg-red-50 border border-red-100 rounded-lg p-4">
            <h4 className="font-medium text-red-900 mb-1">Right to Delete</h4>
            <p className="text-red-800 text-sm">
              You can request deletion of your account and all associated data. This process is irreversible.
            </p>
          </div>

          <div className="bg-yellow-50 border border-yellow-100 rounded-lg p-4">
            <h4 className="font-medium text-yellow-900 mb-1">Right to Rectify</h4>
            <p className="text-yellow-800 text-sm">
              You can update or correct your personal information at any time.
            </p>
          </div>
        </div>

        <p className="text-gray-600 mt-4">
          To exercise these rights, please contact us at{" "}
          <a href="mailto:privacy@adsanalytic.com" className="text-blue-600 hover:underline">
            privacy@adsanalytic.com
          </a>
        </p>
      </section>

      <section className="mb-8">
        <h2 className="text-xl font-semibold text-gray-900 mb-4">8. Contact Us</h2>
        <p className="text-gray-600 mb-4">
          If you have any questions about this Privacy Policy, please contact:
        </p>
        <div className="bg-gray-50 rounded-lg p-4">
          <p className="text-gray-700"><strong>Data Protection Officer</strong></p>
          <p className="text-gray-600">AdsAnalytic Sdn Bhd</p>
          <p className="text-gray-600">Email: privacy@adsanalytic.com</p>
        </div>
      </section>
    </>
  );
}
