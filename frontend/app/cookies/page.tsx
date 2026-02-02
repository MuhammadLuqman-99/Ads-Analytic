"use client";

import Link from "next/link";
import { useState } from "react";

export default function CookiePolicyPage() {
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
            {lang === "ms" ? "Polisi Cookie" : "Cookie Policy"}
          </h1>
          <p className="text-gray-500 mb-8">
            {lang === "ms" ? "Kemas kini terakhir" : "Last updated"}: 1 Februari 2026
          </p>

          <div className="prose prose-gray max-w-none">
            {lang === "ms" ? <CookieContentMS /> : <CookieContentEN />}
          </div>
        </div>

        {/* Legal Nav */}
        <div className="mt-8 flex flex-wrap justify-center gap-4 text-sm text-gray-500">
          <Link href="/privacy" className="hover:text-gray-900">
            {lang === "ms" ? "Polisi Privasi" : "Privacy Policy"}
          </Link>
          <span>•</span>
          <Link href="/terms" className="hover:text-gray-900">
            {lang === "ms" ? "Terma Perkhidmatan" : "Terms of Service"}
          </Link>
          <span>•</span>
          <Link href="/cookies" className="text-blue-600 font-medium">
            {lang === "ms" ? "Polisi Cookie" : "Cookie Policy"}
          </Link>
        </div>
      </main>
    </div>
  );
}

function CookieContentMS() {
  return (
    <>
      <section className="mb-8">
        <h2 className="text-xl font-semibold text-gray-900 mb-4">1. Apa Itu Cookie?</h2>
        <p className="text-gray-600 mb-4">
          Cookie adalah fail teks kecil yang disimpan pada peranti anda apabila anda melawat laman
          web. Cookie membantu laman web mengingati maklumat tentang lawatan anda, seperti pilihan
          bahasa dan tetapan lain.
        </p>
        <p className="text-gray-600">
          Kami menggunakan cookie untuk memastikan laman web kami berfungsi dengan baik dan untuk
          meningkatkan pengalaman pengguna anda.
        </p>
      </section>

      <section className="mb-8">
        <h2 className="text-xl font-semibold text-gray-900 mb-4">2. Jenis Cookie Yang Kami Gunakan</h2>

        <div className="space-y-6">
          {/* Essential Cookies */}
          <div className="bg-green-50 border border-green-200 rounded-lg p-6">
            <div className="flex items-start gap-4">
              <div className="w-10 h-10 bg-green-500 rounded-lg flex items-center justify-center flex-shrink-0">
                <svg className="w-5 h-5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
                </svg>
              </div>
              <div>
                <h3 className="text-lg font-semibold text-green-900 mb-2">Cookie Penting (Essential)</h3>
                <p className="text-green-800 text-sm mb-3">
                  Cookie ini diperlukan untuk laman web berfungsi dan tidak boleh dimatikan.
                </p>
                <div className="overflow-x-auto">
                  <table className="min-w-full text-sm">
                    <thead>
                      <tr className="border-b border-green-200">
                        <th className="text-left py-2 pr-4 font-medium text-green-900">Nama</th>
                        <th className="text-left py-2 pr-4 font-medium text-green-900">Tujuan</th>
                        <th className="text-left py-2 font-medium text-green-900">Tempoh</th>
                      </tr>
                    </thead>
                    <tbody className="text-green-700">
                      <tr className="border-b border-green-100">
                        <td className="py-2 pr-4 font-mono text-xs">session_token</td>
                        <td className="py-2 pr-4">Mengekalkan sesi log masuk anda</td>
                        <td className="py-2">Sesi</td>
                      </tr>
                      <tr className="border-b border-green-100">
                        <td className="py-2 pr-4 font-mono text-xs">csrf_token</td>
                        <td className="py-2 pr-4">Perlindungan keselamatan CSRF</td>
                        <td className="py-2">Sesi</td>
                      </tr>
                      <tr className="border-b border-green-100">
                        <td className="py-2 pr-4 font-mono text-xs">auth_refresh</td>
                        <td className="py-2 pr-4">Memperbaharui token akses</td>
                        <td className="py-2">7 hari</td>
                      </tr>
                      <tr>
                        <td className="py-2 pr-4 font-mono text-xs">cookie_consent</td>
                        <td className="py-2 pr-4">Menyimpan pilihan cookie anda</td>
                        <td className="py-2">1 tahun</td>
                      </tr>
                    </tbody>
                  </table>
                </div>
              </div>
            </div>
          </div>

          {/* Functional Cookies */}
          <div className="bg-blue-50 border border-blue-200 rounded-lg p-6">
            <div className="flex items-start gap-4">
              <div className="w-10 h-10 bg-blue-500 rounded-lg flex items-center justify-center flex-shrink-0">
                <svg className="w-5 h-5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                </svg>
              </div>
              <div>
                <h3 className="text-lg font-semibold text-blue-900 mb-2">Cookie Fungsi (Functional)</h3>
                <p className="text-blue-800 text-sm mb-3">
                  Cookie ini membolehkan ciri-ciri lanjutan dan personalisasi.
                </p>
                <div className="overflow-x-auto">
                  <table className="min-w-full text-sm">
                    <thead>
                      <tr className="border-b border-blue-200">
                        <th className="text-left py-2 pr-4 font-medium text-blue-900">Nama</th>
                        <th className="text-left py-2 pr-4 font-medium text-blue-900">Tujuan</th>
                        <th className="text-left py-2 font-medium text-blue-900">Tempoh</th>
                      </tr>
                    </thead>
                    <tbody className="text-blue-700">
                      <tr className="border-b border-blue-100">
                        <td className="py-2 pr-4 font-mono text-xs">language</td>
                        <td className="py-2 pr-4">Menyimpan pilihan bahasa</td>
                        <td className="py-2">1 tahun</td>
                      </tr>
                      <tr className="border-b border-blue-100">
                        <td className="py-2 pr-4 font-mono text-xs">theme</td>
                        <td className="py-2 pr-4">Tema gelap/cerah</td>
                        <td className="py-2">1 tahun</td>
                      </tr>
                      <tr>
                        <td className="py-2 pr-4 font-mono text-xs">dashboard_layout</td>
                        <td className="py-2 pr-4">Susun atur dashboard pilihan</td>
                        <td className="py-2">1 tahun</td>
                      </tr>
                    </tbody>
                  </table>
                </div>
              </div>
            </div>
          </div>

          {/* Analytics Cookies */}
          <div className="bg-purple-50 border border-purple-200 rounded-lg p-6">
            <div className="flex items-start gap-4">
              <div className="w-10 h-10 bg-purple-500 rounded-lg flex items-center justify-center flex-shrink-0">
                <svg className="w-5 h-5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
                </svg>
              </div>
              <div>
                <h3 className="text-lg font-semibold text-purple-900 mb-2">Cookie Analitik (Analytics)</h3>
                <p className="text-purple-800 text-sm mb-3">
                  Cookie ini membantu kami memahami bagaimana pelawat menggunakan laman web kami.
                  Data dikumpul secara agregat dan tanpa nama.
                </p>
                <div className="overflow-x-auto">
                  <table className="min-w-full text-sm">
                    <thead>
                      <tr className="border-b border-purple-200">
                        <th className="text-left py-2 pr-4 font-medium text-purple-900">Nama</th>
                        <th className="text-left py-2 pr-4 font-medium text-purple-900">Tujuan</th>
                        <th className="text-left py-2 font-medium text-purple-900">Tempoh</th>
                      </tr>
                    </thead>
                    <tbody className="text-purple-700">
                      <tr className="border-b border-purple-100">
                        <td className="py-2 pr-4 font-mono text-xs">_ga</td>
                        <td className="py-2 pr-4">Google Analytics - ID pelawat unik</td>
                        <td className="py-2">2 tahun</td>
                      </tr>
                      <tr>
                        <td className="py-2 pr-4 font-mono text-xs">_gid</td>
                        <td className="py-2 pr-4">Google Analytics - ID sesi</td>
                        <td className="py-2">24 jam</td>
                      </tr>
                    </tbody>
                  </table>
                </div>
                <p className="text-purple-700 text-sm mt-3">
                  <strong>Nota:</strong> Anda boleh menolak cookie analitik tanpa menjejaskan
                  penggunaan laman web.
                </p>
              </div>
            </div>
          </div>
        </div>
      </section>

      <section className="mb-8">
        <h2 className="text-xl font-semibold text-gray-900 mb-4">3. Mengurus Cookie</h2>
        <p className="text-gray-600 mb-4">
          Anda boleh mengawal dan mengurus cookie melalui beberapa cara:
        </p>

        <div className="space-y-4">
          <div className="bg-gray-50 rounded-lg p-4">
            <h4 className="font-medium text-gray-900 mb-2">Tetapan Pelayar</h4>
            <p className="text-gray-600 text-sm">
              Kebanyakan pelayar membenarkan anda melihat, mengurus, dan memadam cookie melalui
              tetapan. Perhatian: Mematikan cookie penting mungkin menjejaskan fungsi laman web.
            </p>
          </div>

          <div className="bg-gray-50 rounded-lg p-4">
            <h4 className="font-medium text-gray-900 mb-2">Banner Persetujuan Cookie</h4>
            <p className="text-gray-600 text-sm">
              Apabila anda mula-mula melawat laman web kami, anda akan melihat banner yang
              membolehkan anda memilih kategori cookie yang ingin anda terima.
            </p>
          </div>

          <div className="bg-gray-50 rounded-lg p-4">
            <h4 className="font-medium text-gray-900 mb-2">Opt-Out Google Analytics</h4>
            <p className="text-gray-600 text-sm">
              Anda boleh memasang{" "}
              <a
                href="https://tools.google.com/dlpage/gaoptout"
                target="_blank"
                rel="noopener noreferrer"
                className="text-blue-600 hover:underline"
              >
                Google Analytics Opt-out Browser Add-on
              </a>{" "}
              untuk menghalang data anda digunakan oleh Google Analytics.
            </p>
          </div>
        </div>
      </section>

      <section className="mb-8">
        <h2 className="text-xl font-semibold text-gray-900 mb-4">4. Cookie Pihak Ketiga</h2>
        <p className="text-gray-600 mb-4">
          Kami menggunakan perkhidmatan pihak ketiga yang mungkin menetapkan cookie mereka sendiri:
        </p>
        <ul className="list-disc list-inside text-gray-600 space-y-2">
          <li>
            <strong>Google Analytics:</strong> Untuk analitik laman web
            <a href="https://policies.google.com/privacy" className="text-blue-600 hover:underline ml-1">
              (Polisi Privasi Google)
            </a>
          </li>
          <li>
            <strong>Stripe:</strong> Untuk pemprosesan pembayaran
            <a href="https://stripe.com/privacy" className="text-blue-600 hover:underline ml-1">
              (Polisi Privasi Stripe)
            </a>
          </li>
        </ul>
      </section>

      <section className="mb-8">
        <h2 className="text-xl font-semibold text-gray-900 mb-4">5. Perubahan Polisi</h2>
        <p className="text-gray-600">
          Kami mungkin mengemas kini Polisi Cookie ini dari semasa ke semasa. Perubahan akan
          dipaparkan di halaman ini dengan tarikh kemas kini yang baharu. Kami menggalakkan anda
          menyemak polisi ini secara berkala.
        </p>
      </section>

      <section className="mb-8">
        <h2 className="text-xl font-semibold text-gray-900 mb-4">6. Hubungi Kami</h2>
        <p className="text-gray-600 mb-4">
          Jika anda mempunyai sebarang soalan tentang penggunaan cookie kami, sila hubungi:
        </p>
        <div className="bg-gray-50 rounded-lg p-4">
          <p className="text-gray-700"><strong>AdsAnalytic Sdn Bhd</strong></p>
          <p className="text-gray-600">Emel: privacy@adsanalytic.com</p>
        </div>
      </section>
    </>
  );
}

function CookieContentEN() {
  return (
    <>
      <section className="mb-8">
        <h2 className="text-xl font-semibold text-gray-900 mb-4">1. What Are Cookies?</h2>
        <p className="text-gray-600 mb-4">
          Cookies are small text files stored on your device when you visit a website. Cookies
          help websites remember information about your visit, such as language preferences and
          other settings.
        </p>
        <p className="text-gray-600">
          We use cookies to ensure our website functions properly and to improve your user experience.
        </p>
      </section>

      <section className="mb-8">
        <h2 className="text-xl font-semibold text-gray-900 mb-4">2. Types of Cookies We Use</h2>

        <div className="space-y-6">
          {/* Essential Cookies */}
          <div className="bg-green-50 border border-green-200 rounded-lg p-6">
            <div className="flex items-start gap-4">
              <div className="w-10 h-10 bg-green-500 rounded-lg flex items-center justify-center flex-shrink-0">
                <svg className="w-5 h-5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
                </svg>
              </div>
              <div>
                <h3 className="text-lg font-semibold text-green-900 mb-2">Essential Cookies</h3>
                <p className="text-green-800 text-sm mb-3">
                  These cookies are required for the website to function and cannot be disabled.
                </p>
                <div className="overflow-x-auto">
                  <table className="min-w-full text-sm">
                    <thead>
                      <tr className="border-b border-green-200">
                        <th className="text-left py-2 pr-4 font-medium text-green-900">Name</th>
                        <th className="text-left py-2 pr-4 font-medium text-green-900">Purpose</th>
                        <th className="text-left py-2 font-medium text-green-900">Duration</th>
                      </tr>
                    </thead>
                    <tbody className="text-green-700">
                      <tr className="border-b border-green-100">
                        <td className="py-2 pr-4 font-mono text-xs">session_token</td>
                        <td className="py-2 pr-4">Maintains your login session</td>
                        <td className="py-2">Session</td>
                      </tr>
                      <tr className="border-b border-green-100">
                        <td className="py-2 pr-4 font-mono text-xs">csrf_token</td>
                        <td className="py-2 pr-4">CSRF security protection</td>
                        <td className="py-2">Session</td>
                      </tr>
                      <tr className="border-b border-green-100">
                        <td className="py-2 pr-4 font-mono text-xs">auth_refresh</td>
                        <td className="py-2 pr-4">Renews access token</td>
                        <td className="py-2">7 days</td>
                      </tr>
                      <tr>
                        <td className="py-2 pr-4 font-mono text-xs">cookie_consent</td>
                        <td className="py-2 pr-4">Stores your cookie preferences</td>
                        <td className="py-2">1 year</td>
                      </tr>
                    </tbody>
                  </table>
                </div>
              </div>
            </div>
          </div>

          {/* Functional Cookies */}
          <div className="bg-blue-50 border border-blue-200 rounded-lg p-6">
            <div className="flex items-start gap-4">
              <div className="w-10 h-10 bg-blue-500 rounded-lg flex items-center justify-center flex-shrink-0">
                <svg className="w-5 h-5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                </svg>
              </div>
              <div>
                <h3 className="text-lg font-semibold text-blue-900 mb-2">Functional Cookies</h3>
                <p className="text-blue-800 text-sm mb-3">
                  These cookies enable advanced features and personalization.
                </p>
                <div className="overflow-x-auto">
                  <table className="min-w-full text-sm">
                    <thead>
                      <tr className="border-b border-blue-200">
                        <th className="text-left py-2 pr-4 font-medium text-blue-900">Name</th>
                        <th className="text-left py-2 pr-4 font-medium text-blue-900">Purpose</th>
                        <th className="text-left py-2 font-medium text-blue-900">Duration</th>
                      </tr>
                    </thead>
                    <tbody className="text-blue-700">
                      <tr className="border-b border-blue-100">
                        <td className="py-2 pr-4 font-mono text-xs">language</td>
                        <td className="py-2 pr-4">Stores language preference</td>
                        <td className="py-2">1 year</td>
                      </tr>
                      <tr className="border-b border-blue-100">
                        <td className="py-2 pr-4 font-mono text-xs">theme</td>
                        <td className="py-2 pr-4">Dark/light theme</td>
                        <td className="py-2">1 year</td>
                      </tr>
                      <tr>
                        <td className="py-2 pr-4 font-mono text-xs">dashboard_layout</td>
                        <td className="py-2 pr-4">Preferred dashboard layout</td>
                        <td className="py-2">1 year</td>
                      </tr>
                    </tbody>
                  </table>
                </div>
              </div>
            </div>
          </div>

          {/* Analytics Cookies */}
          <div className="bg-purple-50 border border-purple-200 rounded-lg p-6">
            <div className="flex items-start gap-4">
              <div className="w-10 h-10 bg-purple-500 rounded-lg flex items-center justify-center flex-shrink-0">
                <svg className="w-5 h-5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
                </svg>
              </div>
              <div>
                <h3 className="text-lg font-semibold text-purple-900 mb-2">Analytics Cookies</h3>
                <p className="text-purple-800 text-sm mb-3">
                  These cookies help us understand how visitors use our website.
                  Data is collected in aggregate and anonymously.
                </p>
                <div className="overflow-x-auto">
                  <table className="min-w-full text-sm">
                    <thead>
                      <tr className="border-b border-purple-200">
                        <th className="text-left py-2 pr-4 font-medium text-purple-900">Name</th>
                        <th className="text-left py-2 pr-4 font-medium text-purple-900">Purpose</th>
                        <th className="text-left py-2 font-medium text-purple-900">Duration</th>
                      </tr>
                    </thead>
                    <tbody className="text-purple-700">
                      <tr className="border-b border-purple-100">
                        <td className="py-2 pr-4 font-mono text-xs">_ga</td>
                        <td className="py-2 pr-4">Google Analytics - Unique visitor ID</td>
                        <td className="py-2">2 years</td>
                      </tr>
                      <tr>
                        <td className="py-2 pr-4 font-mono text-xs">_gid</td>
                        <td className="py-2 pr-4">Google Analytics - Session ID</td>
                        <td className="py-2">24 hours</td>
                      </tr>
                    </tbody>
                  </table>
                </div>
                <p className="text-purple-700 text-sm mt-3">
                  <strong>Note:</strong> You can decline analytics cookies without affecting
                  your use of the website.
                </p>
              </div>
            </div>
          </div>
        </div>
      </section>

      <section className="mb-8">
        <h2 className="text-xl font-semibold text-gray-900 mb-4">3. Managing Cookies</h2>
        <p className="text-gray-600 mb-4">
          You can control and manage cookies in several ways:
        </p>

        <div className="space-y-4">
          <div className="bg-gray-50 rounded-lg p-4">
            <h4 className="font-medium text-gray-900 mb-2">Browser Settings</h4>
            <p className="text-gray-600 text-sm">
              Most browsers allow you to view, manage, and delete cookies through settings.
              Note: Disabling essential cookies may affect website functionality.
            </p>
          </div>

          <div className="bg-gray-50 rounded-lg p-4">
            <h4 className="font-medium text-gray-900 mb-2">Cookie Consent Banner</h4>
            <p className="text-gray-600 text-sm">
              When you first visit our website, you will see a banner that allows you to choose
              which cookie categories you want to accept.
            </p>
          </div>

          <div className="bg-gray-50 rounded-lg p-4">
            <h4 className="font-medium text-gray-900 mb-2">Google Analytics Opt-Out</h4>
            <p className="text-gray-600 text-sm">
              You can install the{" "}
              <a
                href="https://tools.google.com/dlpage/gaoptout"
                target="_blank"
                rel="noopener noreferrer"
                className="text-blue-600 hover:underline"
              >
                Google Analytics Opt-out Browser Add-on
              </a>{" "}
              to prevent your data from being used by Google Analytics.
            </p>
          </div>
        </div>
      </section>

      <section className="mb-8">
        <h2 className="text-xl font-semibold text-gray-900 mb-4">4. Third-Party Cookies</h2>
        <p className="text-gray-600 mb-4">
          We use third-party services that may set their own cookies:
        </p>
        <ul className="list-disc list-inside text-gray-600 space-y-2">
          <li>
            <strong>Google Analytics:</strong> For website analytics
            <a href="https://policies.google.com/privacy" className="text-blue-600 hover:underline ml-1">
              (Google Privacy Policy)
            </a>
          </li>
          <li>
            <strong>Stripe:</strong> For payment processing
            <a href="https://stripe.com/privacy" className="text-blue-600 hover:underline ml-1">
              (Stripe Privacy Policy)
            </a>
          </li>
        </ul>
      </section>

      <section className="mb-8">
        <h2 className="text-xl font-semibold text-gray-900 mb-4">5. Policy Changes</h2>
        <p className="text-gray-600">
          We may update this Cookie Policy from time to time. Changes will be posted on this page
          with an updated date. We encourage you to review this policy periodically.
        </p>
      </section>

      <section className="mb-8">
        <h2 className="text-xl font-semibold text-gray-900 mb-4">6. Contact Us</h2>
        <p className="text-gray-600 mb-4">
          If you have any questions about our use of cookies, please contact:
        </p>
        <div className="bg-gray-50 rounded-lg p-4">
          <p className="text-gray-700"><strong>AdsAnalytic Sdn Bhd</strong></p>
          <p className="text-gray-600">Email: privacy@adsanalytic.com</p>
        </div>
      </section>
    </>
  );
}
