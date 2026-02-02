"use client";

import { useState } from "react";

export function FAQ() {
  const [openIndex, setOpenIndex] = useState<number | null>(0);

  const faqs = [
    {
      question: "Adakah data saya selamat?",
      answer:
        "Ya, keselamatan data adalah keutamaan kami. Kami menggunakan enkripsi AES-256 untuk semua data sensitif, SSL/TLS untuk semua connections, dan mematuhi standard keselamatan industri. Server kami dihoskan di AWS dengan SOC 2 compliance. Kami tidak pernah share atau jual data anda kepada pihak ketiga.",
    },
    {
      question: "Berapa lama masa untuk setup?",
      answer:
        "Setup mengambil masa kurang dari 5 minit! Anda hanya perlu register, connect platform ads anda (Meta, TikTok, Shopee), dan data akan mula sync secara automatik. Tiada technical knowledge diperlukan.",
    },
    {
      question: "Boleh cancel bila-bila masa?",
      answer:
        "Ya, anda boleh cancel subscription bila-bila masa tanpa sebarang penalti. Jika cancel dalam 30 hari pertama, kami akan refund sepenuhnya. Tiada lock-in contract atau hidden fees.",
    },
    {
      question: "Platform mana yang disokong?",
      answer:
        "Buat masa ini, kami menyokong Meta Ads (Facebook & Instagram), TikTok Ads, dan Shopee Ads. Google Ads, Lazada Ads, dan platform lain akan datang tidak lama lagi. Subscribe newsletter kami untuk updates!",
    },
    {
      question: "Adakah trial percuma?",
      answer:
        "Ya! Plan Pro dan Business datang dengan 14 hari trial percuma. Tiada kad kredit diperlukan untuk mula. Anda boleh test semua features sebelum decide.",
    },
    {
      question: "Bagaimana data di-sync?",
      answer:
        "Data di-sync secara automatik menggunakan official API dari setiap platform. Plan Percuma sync sekali sehari, Plan Pro sync setiap jam, dan Plan Business sync real-time. Anda juga boleh trigger manual sync bila-bila masa.",
    },
    {
      question: "Boleh tambah team members?",
      answer:
        "Ya! Plan Pro membenarkan sehingga 3 team members, dan Plan Business membenarkan unlimited team members. Setiap member boleh ada role yang berbeza (Admin, Editor, Viewer).",
    },
    {
      question: "Ada mobile app?",
      answer:
        "Dashboard kami adalah fully responsive dan berfungsi dengan baik pada mobile browser. Dedicated mobile app sedang dalam development dan akan dilancarkan Q2 2026.",
    },
  ];

  return (
    <section id="faq" className="py-20 px-4 sm:px-6 lg:px-8 bg-gray-50">
      <div className="max-w-3xl mx-auto">
        <div className="text-center mb-16">
          <div className="inline-flex items-center px-4 py-2 bg-blue-100 text-blue-700 rounded-full text-sm font-medium mb-4">
            FAQ
          </div>
          <h2 className="text-3xl sm:text-4xl font-bold text-gray-900 mb-4">
            Soalan Lazim
          </h2>
          <p className="text-xl text-gray-600">
            Ada soalan? Kami ada jawapan.
          </p>
        </div>

        <div className="space-y-4">
          {faqs.map((faq, index) => (
            <div
              key={index}
              className="bg-white rounded-xl border border-gray-200 overflow-hidden"
            >
              <button
                onClick={() => setOpenIndex(openIndex === index ? null : index)}
                className="w-full px-6 py-5 flex items-center justify-between text-left hover:bg-gray-50 transition-colors"
              >
                <span className="font-semibold text-gray-900 pr-4">{faq.question}</span>
                <svg
                  className={`w-5 h-5 text-gray-500 flex-shrink-0 transition-transform duration-200 ${
                    openIndex === index ? "rotate-180" : ""
                  }`}
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                >
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
                </svg>
              </button>
              {openIndex === index && (
                <div className="px-6 pb-5">
                  <p className="text-gray-600 leading-relaxed">{faq.answer}</p>
                </div>
              )}
            </div>
          ))}
        </div>

        {/* Contact CTA */}
        <div className="text-center mt-12">
          <p className="text-gray-600 mb-4">Masih ada soalan?</p>
          <a
            href="mailto:support@adsanalytic.com"
            className="inline-flex items-center gap-2 text-blue-600 font-medium hover:text-blue-700"
          >
            <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
            </svg>
            Hubungi kami di support@adsanalytic.com
          </a>
        </div>
      </div>
    </section>
  );
}
