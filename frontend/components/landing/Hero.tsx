import Link from "next/link";

export function Hero() {
  return (
    <section className="pt-32 pb-20 px-4 sm:px-6 lg:px-8 bg-gradient-to-b from-blue-50 via-white to-white">
      <div className="max-w-7xl mx-auto">
        <div className="text-center max-w-4xl mx-auto">
          {/* Badge */}
          <div className="inline-flex items-center px-4 py-2 bg-blue-100 text-blue-700 rounded-full text-sm font-medium mb-8">
            <span className="w-2 h-2 bg-green-500 rounded-full mr-2 animate-pulse"></span>
            Dipercayai 500+ peniaga e-commerce Malaysia
          </div>

          {/* Headline */}
          <h1 className="text-4xl sm:text-5xl lg:text-6xl font-bold text-gray-900 leading-tight mb-6">
            Semua Ads Data{" "}
            <span className="text-transparent bg-clip-text bg-gradient-to-r from-blue-600 to-purple-600">
              Dalam Satu Dashboard
            </span>
          </h1>

          {/* Subheadline */}
          <p className="text-xl text-gray-600 mb-10 max-w-2xl mx-auto leading-relaxed">
            Gabungkan data iklan dari Meta, TikTok, dan Shopee dalam satu tempat.
            Jimat masa, tingkatkan ROAS, dan buat keputusan lebih bijak untuk bisnes e-commerce anda.
          </p>

          {/* CTA Buttons */}
          <div className="flex flex-col sm:flex-row items-center justify-center gap-4 mb-12">
            <Link
              href="/register"
              className="w-full sm:w-auto bg-gradient-to-r from-blue-600 to-purple-600 text-white px-8 py-4 rounded-xl font-semibold text-lg hover:opacity-90 transition-all shadow-xl shadow-blue-500/25 hover:shadow-2xl hover:shadow-blue-500/30 hover:-translate-y-0.5"
            >
              Cuba Percuma 14 Hari
            </Link>
            <a
              href="#features"
              className="w-full sm:w-auto flex items-center justify-center gap-2 text-gray-700 px-8 py-4 rounded-xl font-semibold text-lg hover:bg-gray-100 transition-colors"
            >
              <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z" />
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
              Lihat Demo
            </a>
          </div>

          {/* Trust indicators */}
          <div className="flex flex-wrap items-center justify-center gap-6 text-sm text-gray-500">
            <div className="flex items-center gap-2">
              <svg className="w-5 h-5 text-green-500" fill="currentColor" viewBox="0 0 20 20">
                <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
              </svg>
              Tiada kad kredit diperlukan
            </div>
            <div className="flex items-center gap-2">
              <svg className="w-5 h-5 text-green-500" fill="currentColor" viewBox="0 0 20 20">
                <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
              </svg>
              Setup dalam 5 minit
            </div>
            <div className="flex items-center gap-2">
              <svg className="w-5 h-5 text-green-500" fill="currentColor" viewBox="0 0 20 20">
                <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
              </svg>
              Cancel bila-bila masa
            </div>
          </div>
        </div>

        {/* Dashboard Preview */}
        <div className="mt-16 relative">
          <div className="absolute inset-0 bg-gradient-to-t from-white via-transparent to-transparent z-10 pointer-events-none"></div>
          <div className="bg-gray-900 rounded-2xl shadow-2xl overflow-hidden border border-gray-800">
            {/* Browser chrome */}
            <div className="bg-gray-800 px-4 py-3 flex items-center gap-2">
              <div className="flex gap-1.5">
                <div className="w-3 h-3 rounded-full bg-red-500"></div>
                <div className="w-3 h-3 rounded-full bg-yellow-500"></div>
                <div className="w-3 h-3 rounded-full bg-green-500"></div>
              </div>
              <div className="flex-1 flex justify-center">
                <div className="bg-gray-700 rounded-lg px-4 py-1 text-gray-400 text-sm">
                  app.adsanalytic.com/dashboard
                </div>
              </div>
            </div>
            {/* Dashboard mockup */}
            <div className="bg-gray-100 p-6">
              <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
                {/* Metric cards */}
                {[
                  { label: "Jumlah Spend", value: "RM 12,450", change: "+12%", color: "blue" },
                  { label: "Revenue", value: "RM 45,230", change: "+28%", color: "green" },
                  { label: "ROAS", value: "3.63x", change: "+0.5x", color: "purple" },
                  { label: "Conversions", value: "1,247", change: "+18%", color: "orange" },
                ].map((metric, i) => (
                  <div key={i} className="bg-white rounded-xl p-4 shadow-sm">
                    <p className="text-gray-500 text-sm">{metric.label}</p>
                    <p className="text-2xl font-bold text-gray-900 mt-1">{metric.value}</p>
                    <p className="text-green-600 text-sm mt-1">{metric.change} vs minggu lepas</p>
                  </div>
                ))}
              </div>
              {/* Platform breakdown */}
              <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                {[
                  { platform: "Meta Ads", spend: "RM 5,200", roas: "4.2x", color: "from-blue-500 to-blue-600" },
                  { platform: "TikTok Ads", spend: "RM 4,100", roas: "3.8x", color: "from-pink-500 to-rose-500" },
                  { platform: "Shopee Ads", spend: "RM 3,150", roas: "2.9x", color: "from-orange-500 to-red-500" },
                ].map((platform, i) => (
                  <div key={i} className="bg-white rounded-xl p-4 shadow-sm">
                    <div className={`w-10 h-10 rounded-lg bg-gradient-to-br ${platform.color} mb-3`}></div>
                    <p className="font-semibold text-gray-900">{platform.platform}</p>
                    <div className="flex justify-between mt-2 text-sm">
                      <span className="text-gray-500">Spend: {platform.spend}</span>
                      <span className="text-green-600 font-medium">ROAS: {platform.roas}</span>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
