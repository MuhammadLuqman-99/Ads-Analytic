"use client";

import Link from "next/link";
import { useState } from "react";

export default function TermsOfServicePage() {
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
            {lang === "ms" ? "Terma Perkhidmatan" : "Terms of Service"}
          </h1>
          <p className="text-gray-500 mb-8">
            {lang === "ms" ? "Kemas kini terakhir" : "Last updated"}: 1 Februari 2026
          </p>

          <div className="prose prose-gray max-w-none">
            {lang === "ms" ? <TermsContentMS /> : <TermsContentEN />}
          </div>
        </div>

        {/* Legal Nav */}
        <div className="mt-8 flex flex-wrap justify-center gap-4 text-sm text-gray-500">
          <Link href="/privacy" className="hover:text-gray-900">
            {lang === "ms" ? "Polisi Privasi" : "Privacy Policy"}
          </Link>
          <span>•</span>
          <Link href="/terms" className="text-blue-600 font-medium">
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

function TermsContentMS() {
  return (
    <>
      <section className="mb-8">
        <h2 className="text-xl font-semibold text-gray-900 mb-4">1. Pengenalan</h2>
        <p className="text-gray-600 mb-4">
          Selamat datang ke AdsAnalytic. Terma Perkhidmatan ini (&quot;Terma&quot;) mengawal penggunaan anda
          terhadap perkhidmatan kami. Dengan mengakses atau menggunakan AdsAnalytic, anda bersetuju
          untuk terikat dengan Terma ini.
        </p>
        <p className="text-gray-600">
          Sila baca Terma ini dengan teliti sebelum menggunakan perkhidmatan kami. Jika anda tidak
          bersetuju dengan mana-mana bahagian Terma ini, anda tidak boleh menggunakan perkhidmatan kami.
        </p>
      </section>

      <section className="mb-8">
        <h2 className="text-xl font-semibold text-gray-900 mb-4">2. Penerangan Perkhidmatan</h2>
        <p className="text-gray-600 mb-4">AdsAnalytic menyediakan:</p>
        <ul className="list-disc list-inside text-gray-600 space-y-2">
          <li>Dashboard analitik untuk menggabungkan data iklan dari pelbagai platform (Meta, TikTok, Shopee)</li>
          <li>Pengiraan metrik seperti ROAS, CTR, CPA, dan CPM</li>
          <li>Penyegerakan data automatik dari platform pengiklanan</li>
          <li>Eksport laporan dalam format PDF dan Excel</li>
          <li>Alat kolaborasi pasukan untuk pengurusan kempen</li>
        </ul>
      </section>

      <section className="mb-8">
        <h2 className="text-xl font-semibold text-gray-900 mb-4">3. Kelayakan</h2>
        <p className="text-gray-600 mb-4">Untuk menggunakan AdsAnalytic, anda mesti:</p>
        <ul className="list-disc list-inside text-gray-600 space-y-2">
          <li>Berumur sekurang-kurangnya 18 tahun</li>
          <li>Mempunyai kapasiti undang-undang untuk memasuki perjanjian yang mengikat</li>
          <li>Mempunyai akaun pengiklanan yang sah di platform yang disokong</li>
          <li>Mematuhi terma perkhidmatan platform pengiklanan yang berkaitan</li>
        </ul>
      </section>

      <section className="mb-8">
        <h2 className="text-xl font-semibold text-gray-900 mb-4">4. Tanggungjawab Pengguna</h2>
        <p className="text-gray-600 mb-4">Sebagai pengguna AdsAnalytic, anda bersetuju untuk:</p>

        <h3 className="text-lg font-medium text-gray-800 mb-2">4.1 Keselamatan Akaun</h3>
        <ul className="list-disc list-inside text-gray-600 mb-4 space-y-1">
          <li>Menjaga kerahsiaan kata laluan anda</li>
          <li>Memberitahu kami dengan segera jika terdapat akses tanpa kebenaran</li>
          <li>Tidak berkongsi akaun dengan pihak lain tanpa kebenaran</li>
        </ul>

        <h3 className="text-lg font-medium text-gray-800 mb-2">4.2 Penggunaan Yang Dibenarkan</h3>
        <ul className="list-disc list-inside text-gray-600 mb-4 space-y-1">
          <li>Menggunakan perkhidmatan hanya untuk tujuan perniagaan yang sah</li>
          <li>Tidak cuba menggodam, menyalahgunakan, atau mengganggu perkhidmatan</li>
          <li>Tidak menggunakan perkhidmatan untuk aktiviti haram</li>
          <li>Mematuhi semua undang-undang dan peraturan yang terpakai</li>
        </ul>

        <h3 className="text-lg font-medium text-gray-800 mb-2">4.3 Data dan Kandungan</h3>
        <ul className="list-disc list-inside text-gray-600 space-y-1">
          <li>Memastikan anda mempunyai hak untuk menyambungkan akaun pengiklanan</li>
          <li>Bertanggungjawab atas ketepatan data yang dikongsi</li>
          <li>Tidak memuat naik kandungan yang melanggar hak cipta atau hak pihak ketiga</li>
        </ul>
      </section>

      <section className="mb-8">
        <h2 className="text-xl font-semibold text-gray-900 mb-4">5. Pelan dan Pembayaran</h2>

        <h3 className="text-lg font-medium text-gray-800 mb-2">5.1 Pelan Langganan</h3>
        <div className="bg-gray-50 rounded-lg p-4 mb-4">
          <table className="min-w-full">
            <thead>
              <tr>
                <th className="text-left text-sm font-medium text-gray-900 pb-2">Pelan</th>
                <th className="text-left text-sm font-medium text-gray-900 pb-2">Harga</th>
                <th className="text-left text-sm font-medium text-gray-900 pb-2">Billing</th>
              </tr>
            </thead>
            <tbody className="text-sm text-gray-600">
              <tr>
                <td className="py-1">Percuma</td>
                <td className="py-1">RM 0</td>
                <td className="py-1">-</td>
              </tr>
              <tr>
                <td className="py-1">Pro</td>
                <td className="py-1">RM 99/bulan</td>
                <td className="py-1">Bulanan</td>
              </tr>
              <tr>
                <td className="py-1">Business</td>
                <td className="py-1">RM 299/bulan</td>
                <td className="py-1">Bulanan</td>
              </tr>
            </tbody>
          </table>
        </div>

        <h3 className="text-lg font-medium text-gray-800 mb-2">5.2 Terma Pembayaran</h3>
        <ul className="list-disc list-inside text-gray-600 mb-4 space-y-1">
          <li>Pembayaran diproses pada awal setiap tempoh billing</li>
          <li>Harga tidak termasuk cukai yang terpakai (SST)</li>
          <li>Pembayaran boleh dibuat melalui kad kredit/debit atau FPX</li>
          <li>Langganan akan diperbaharui secara automatik melainkan dibatalkan</li>
        </ul>

        <h3 className="text-lg font-medium text-gray-800 mb-2">5.3 Polisi Bayaran Balik</h3>
        <div className="bg-green-50 border border-green-200 rounded-lg p-4">
          <p className="text-green-800 mb-2">
            <strong>Jaminan 30 Hari Wang Dikembalikan</strong>
          </p>
          <ul className="list-disc list-inside text-green-700 text-sm space-y-1">
            <li>Bayaran balik penuh jika diminta dalam 30 hari pertama langganan berbayar</li>
            <li>Tiada bayaran balik selepas 30 hari</li>
            <li>Bayaran balik akan dikreditkan dalam 5-10 hari bekerja</li>
            <li>Pelan Percuma tidak layak untuk bayaran balik</li>
          </ul>
        </div>
      </section>

      <section className="mb-8">
        <h2 className="text-xl font-semibold text-gray-900 mb-4">6. Had Liabiliti</h2>
        <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-4 mb-4">
          <p className="text-yellow-800 text-sm">
            <strong>PENTING:</strong> Sila baca bahagian ini dengan teliti kerana ia mengehadkan liabiliti kami kepada anda.
          </p>
        </div>

        <p className="text-gray-600 mb-4">Setakat yang dibenarkan oleh undang-undang:</p>
        <ul className="list-disc list-inside text-gray-600 space-y-2">
          <li>
            <strong>Tiada Jaminan:</strong> Perkhidmatan disediakan &quot;sebagaimana adanya&quot; tanpa sebarang
            jaminan tersurat atau tersirat
          </li>
          <li>
            <strong>Had Ganti Rugi:</strong> Liabiliti maksimum kami terhad kepada jumlah yang anda
            bayar kepada kami dalam 12 bulan sebelumnya
          </li>
          <li>
            <strong>Pengecualian:</strong> Kami tidak bertanggungjawab atas kerosakan tidak langsung,
            sampingan, khas, atau punitif
          </li>
          <li>
            <strong>Data Pihak Ketiga:</strong> Kami tidak menjamin ketepatan data dari platform
            pengiklanan pihak ketiga
          </li>
        </ul>
      </section>

      <section className="mb-8">
        <h2 className="text-xl font-semibold text-gray-900 mb-4">7. Penamatan Akaun</h2>

        <h3 className="text-lg font-medium text-gray-800 mb-2">7.1 Pembatalan Oleh Anda</h3>
        <p className="text-gray-600 mb-4">
          Anda boleh membatalkan akaun anda pada bila-bila masa melalui tetapan akaun. Selepas
          pembatalan:
        </p>
        <ul className="list-disc list-inside text-gray-600 mb-4 space-y-1">
          <li>Akses kepada ciri berbayar akan tamat pada akhir tempoh billing semasa</li>
          <li>Data anda akan disimpan selama 30 hari sebelum pemadaman</li>
          <li>Anda boleh mengeksport data sebelum pemadaman</li>
        </ul>

        <h3 className="text-lg font-medium text-gray-800 mb-2">7.2 Penamatan Oleh Kami</h3>
        <p className="text-gray-600 mb-4">Kami berhak untuk menamatkan atau menggantung akaun anda jika:</p>
        <ul className="list-disc list-inside text-gray-600 space-y-1">
          <li>Anda melanggar Terma ini</li>
          <li>Anda terlibat dalam aktiviti penipuan atau haram</li>
          <li>Anda tidak membayar yuran yang tertunggak</li>
          <li>Atas permintaan pihak berkuasa undang-undang</li>
        </ul>
      </section>

      <section className="mb-8">
        <h2 className="text-xl font-semibold text-gray-900 mb-4">8. Perubahan Terma</h2>
        <p className="text-gray-600 mb-4">
          Kami berhak untuk mengubah Terma ini pada bila-bila masa. Perubahan material akan
          dimaklumkan melalui:
        </p>
        <ul className="list-disc list-inside text-gray-600 space-y-1">
          <li>Emel kepada alamat yang didaftarkan</li>
          <li>Notifikasi dalam aplikasi</li>
          <li>Pengumuman di laman web</li>
        </ul>
        <p className="text-gray-600 mt-4">
          Penggunaan berterusan selepas perubahan berkuatkuasa akan dianggap sebagai penerimaan
          Terma yang dikemas kini.
        </p>
      </section>

      <section className="mb-8">
        <h2 className="text-xl font-semibold text-gray-900 mb-4">9. Undang-undang dan Bidang Kuasa</h2>
        <p className="text-gray-600">
          Terma ini ditadbir oleh undang-undang Malaysia. Sebarang pertikaian akan diselesaikan
          secara eksklusif di mahkamah Malaysia.
        </p>
      </section>

      <section className="mb-8">
        <h2 className="text-xl font-semibold text-gray-900 mb-4">10. Hubungi Kami</h2>
        <p className="text-gray-600 mb-4">
          Untuk sebarang pertanyaan mengenai Terma ini, sila hubungi:
        </p>
        <div className="bg-gray-50 rounded-lg p-4">
          <p className="text-gray-700"><strong>AdsAnalytic Sdn Bhd</strong></p>
          <p className="text-gray-600">Emel: legal@adsanalytic.com</p>
        </div>
      </section>
    </>
  );
}

function TermsContentEN() {
  return (
    <>
      <section className="mb-8">
        <h2 className="text-xl font-semibold text-gray-900 mb-4">1. Introduction</h2>
        <p className="text-gray-600 mb-4">
          Welcome to AdsAnalytic. These Terms of Service (&quot;Terms&quot;) govern your use of our services.
          By accessing or using AdsAnalytic, you agree to be bound by these Terms.
        </p>
        <p className="text-gray-600">
          Please read these Terms carefully before using our services. If you do not agree with
          any part of these Terms, you may not use our services.
        </p>
      </section>

      <section className="mb-8">
        <h2 className="text-xl font-semibold text-gray-900 mb-4">2. Service Description</h2>
        <p className="text-gray-600 mb-4">AdsAnalytic provides:</p>
        <ul className="list-disc list-inside text-gray-600 space-y-2">
          <li>Analytics dashboard for aggregating advertising data from multiple platforms (Meta, TikTok, Shopee)</li>
          <li>Calculation of metrics such as ROAS, CTR, CPA, and CPM</li>
          <li>Automatic data synchronization from advertising platforms</li>
          <li>Report export in PDF and Excel formats</li>
          <li>Team collaboration tools for campaign management</li>
        </ul>
      </section>

      <section className="mb-8">
        <h2 className="text-xl font-semibold text-gray-900 mb-4">3. Eligibility</h2>
        <p className="text-gray-600 mb-4">To use AdsAnalytic, you must:</p>
        <ul className="list-disc list-inside text-gray-600 space-y-2">
          <li>Be at least 18 years old</li>
          <li>Have legal capacity to enter into binding agreements</li>
          <li>Have valid advertising accounts on supported platforms</li>
          <li>Comply with the terms of service of relevant advertising platforms</li>
        </ul>
      </section>

      <section className="mb-8">
        <h2 className="text-xl font-semibold text-gray-900 mb-4">4. User Responsibilities</h2>
        <p className="text-gray-600 mb-4">As an AdsAnalytic user, you agree to:</p>

        <h3 className="text-lg font-medium text-gray-800 mb-2">4.1 Account Security</h3>
        <ul className="list-disc list-inside text-gray-600 mb-4 space-y-1">
          <li>Maintain the confidentiality of your password</li>
          <li>Notify us immediately of any unauthorized access</li>
          <li>Not share your account with others without authorization</li>
        </ul>

        <h3 className="text-lg font-medium text-gray-800 mb-2">4.2 Acceptable Use</h3>
        <ul className="list-disc list-inside text-gray-600 mb-4 space-y-1">
          <li>Use the service only for legitimate business purposes</li>
          <li>Not attempt to hack, abuse, or interfere with the service</li>
          <li>Not use the service for illegal activities</li>
          <li>Comply with all applicable laws and regulations</li>
        </ul>

        <h3 className="text-lg font-medium text-gray-800 mb-2">4.3 Data and Content</h3>
        <ul className="list-disc list-inside text-gray-600 space-y-1">
          <li>Ensure you have the right to connect advertising accounts</li>
          <li>Be responsible for the accuracy of shared data</li>
          <li>Not upload content that violates copyright or third-party rights</li>
        </ul>
      </section>

      <section className="mb-8">
        <h2 className="text-xl font-semibold text-gray-900 mb-4">5. Plans and Payment</h2>

        <h3 className="text-lg font-medium text-gray-800 mb-2">5.1 Subscription Plans</h3>
        <div className="bg-gray-50 rounded-lg p-4 mb-4">
          <table className="min-w-full">
            <thead>
              <tr>
                <th className="text-left text-sm font-medium text-gray-900 pb-2">Plan</th>
                <th className="text-left text-sm font-medium text-gray-900 pb-2">Price</th>
                <th className="text-left text-sm font-medium text-gray-900 pb-2">Billing</th>
              </tr>
            </thead>
            <tbody className="text-sm text-gray-600">
              <tr>
                <td className="py-1">Free</td>
                <td className="py-1">RM 0</td>
                <td className="py-1">-</td>
              </tr>
              <tr>
                <td className="py-1">Pro</td>
                <td className="py-1">RM 99/month</td>
                <td className="py-1">Monthly</td>
              </tr>
              <tr>
                <td className="py-1">Business</td>
                <td className="py-1">RM 299/month</td>
                <td className="py-1">Monthly</td>
              </tr>
            </tbody>
          </table>
        </div>

        <h3 className="text-lg font-medium text-gray-800 mb-2">5.2 Payment Terms</h3>
        <ul className="list-disc list-inside text-gray-600 mb-4 space-y-1">
          <li>Payment is processed at the beginning of each billing period</li>
          <li>Prices exclude applicable taxes (SST)</li>
          <li>Payment can be made via credit/debit card or FPX</li>
          <li>Subscriptions renew automatically unless cancelled</li>
        </ul>

        <h3 className="text-lg font-medium text-gray-800 mb-2">5.3 Refund Policy</h3>
        <div className="bg-green-50 border border-green-200 rounded-lg p-4">
          <p className="text-green-800 mb-2">
            <strong>30-Day Money-Back Guarantee</strong>
          </p>
          <ul className="list-disc list-inside text-green-700 text-sm space-y-1">
            <li>Full refund if requested within 30 days of first paid subscription</li>
            <li>No refunds after 30 days</li>
            <li>Refunds will be credited within 5-10 business days</li>
            <li>Free plan is not eligible for refunds</li>
          </ul>
        </div>
      </section>

      <section className="mb-8">
        <h2 className="text-xl font-semibold text-gray-900 mb-4">6. Limitation of Liability</h2>
        <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-4 mb-4">
          <p className="text-yellow-800 text-sm">
            <strong>IMPORTANT:</strong> Please read this section carefully as it limits our liability to you.
          </p>
        </div>

        <p className="text-gray-600 mb-4">To the extent permitted by law:</p>
        <ul className="list-disc list-inside text-gray-600 space-y-2">
          <li>
            <strong>No Warranty:</strong> The service is provided &quot;as is&quot; without any express
            or implied warranties
          </li>
          <li>
            <strong>Damages Cap:</strong> Our maximum liability is limited to the amount you paid
            us in the preceding 12 months
          </li>
          <li>
            <strong>Exclusions:</strong> We are not liable for indirect, incidental, special,
            or punitive damages
          </li>
          <li>
            <strong>Third-Party Data:</strong> We do not guarantee the accuracy of data from
            third-party advertising platforms
          </li>
        </ul>
      </section>

      <section className="mb-8">
        <h2 className="text-xl font-semibold text-gray-900 mb-4">7. Account Termination</h2>

        <h3 className="text-lg font-medium text-gray-800 mb-2">7.1 Cancellation By You</h3>
        <p className="text-gray-600 mb-4">
          You can cancel your account at any time through account settings. After cancellation:
        </p>
        <ul className="list-disc list-inside text-gray-600 mb-4 space-y-1">
          <li>Access to paid features ends at the end of the current billing period</li>
          <li>Your data will be retained for 30 days before deletion</li>
          <li>You can export your data before deletion</li>
        </ul>

        <h3 className="text-lg font-medium text-gray-800 mb-2">7.2 Termination By Us</h3>
        <p className="text-gray-600 mb-4">We reserve the right to terminate or suspend your account if:</p>
        <ul className="list-disc list-inside text-gray-600 space-y-1">
          <li>You violate these Terms</li>
          <li>You engage in fraudulent or illegal activities</li>
          <li>You fail to pay outstanding fees</li>
          <li>At the request of legal authorities</li>
        </ul>
      </section>

      <section className="mb-8">
        <h2 className="text-xl font-semibold text-gray-900 mb-4">8. Changes to Terms</h2>
        <p className="text-gray-600 mb-4">
          We reserve the right to modify these Terms at any time. Material changes will be
          notified through:
        </p>
        <ul className="list-disc list-inside text-gray-600 space-y-1">
          <li>Email to your registered address</li>
          <li>In-app notification</li>
          <li>Website announcement</li>
        </ul>
        <p className="text-gray-600 mt-4">
          Continued use after changes take effect will be deemed acceptance of the updated Terms.
        </p>
      </section>

      <section className="mb-8">
        <h2 className="text-xl font-semibold text-gray-900 mb-4">9. Governing Law and Jurisdiction</h2>
        <p className="text-gray-600">
          These Terms are governed by the laws of Malaysia. Any disputes will be resolved
          exclusively in Malaysian courts.
        </p>
      </section>

      <section className="mb-8">
        <h2 className="text-xl font-semibold text-gray-900 mb-4">10. Contact Us</h2>
        <p className="text-gray-600 mb-4">
          For any questions about these Terms, please contact:
        </p>
        <div className="bg-gray-50 rounded-lg p-4">
          <p className="text-gray-700"><strong>AdsAnalytic Sdn Bhd</strong></p>
          <p className="text-gray-600">Email: legal@adsanalytic.com</p>
        </div>
      </section>
    </>
  );
}
