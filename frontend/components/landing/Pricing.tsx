import Link from "next/link";

export function Pricing() {
  const plans = [
    {
      name: "Percuma",
      price: "RM 0",
      period: "selamanya",
      description: "Untuk peniaga yang baru bermula",
      features: [
        "1 platform connection",
        "Dashboard asas",
        "7 hari data history",
        "Export CSV",
        "Email support",
      ],
      limitations: [
        "Tiada cross-platform analytics",
        "Tiada auto-sync",
      ],
      cta: "Mula Percuma",
      ctaLink: "/register",
      popular: false,
    },
    {
      name: "Pro",
      price: "RM 99",
      period: "/bulan",
      description: "Untuk peniaga yang serius grow bisnes",
      features: [
        "3 platform connections",
        "Cross-platform ROAS",
        "30 hari data history",
        "Auto-sync setiap jam",
        "Export PDF & Excel",
        "Priority support",
        "Custom dashboard",
        "Team collaboration (3 users)",
      ],
      limitations: [],
      cta: "Cuba 14 Hari Percuma",
      ctaLink: "/register?plan=pro",
      popular: true,
    },
    {
      name: "Business",
      price: "RM 299",
      period: "/bulan",
      description: "Untuk agency dan bisnes enterprise",
      features: [
        "Unlimited platform connections",
        "Unlimited data history",
        "Real-time sync",
        "White-label reports",
        "API access",
        "Dedicated account manager",
        "Custom integrations",
        "Unlimited team members",
        "SSO authentication",
        "SLA guarantee",
      ],
      limitations: [],
      cta: "Hubungi Kami",
      ctaLink: "/contact",
      popular: false,
    },
  ];

  return (
    <section id="pricing" className="py-20 px-4 sm:px-6 lg:px-8 bg-white">
      <div className="max-w-7xl mx-auto">
        <div className="text-center mb-16">
          <div className="inline-flex items-center px-4 py-2 bg-purple-100 text-purple-700 rounded-full text-sm font-medium mb-4">
            Harga Transparent
          </div>
          <h2 className="text-3xl sm:text-4xl font-bold text-gray-900 mb-4">
            Pilih plan yang sesuai untuk bisnes anda
          </h2>
          <p className="text-xl text-gray-600 max-w-2xl mx-auto">
            Tiada hidden fees. Cancel bila-bila masa. Money-back guarantee 30 hari.
          </p>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-8 max-w-6xl mx-auto">
          {plans.map((plan, index) => (
            <div
              key={index}
              className={`relative rounded-2xl p-8 ${
                plan.popular
                  ? "bg-gradient-to-b from-blue-600 to-purple-600 text-white shadow-2xl shadow-blue-500/25 scale-105 z-10"
                  : "bg-white border border-gray-200"
              }`}
            >
              {/* Popular badge */}
              {plan.popular && (
                <div className="absolute -top-4 left-1/2 -translate-x-1/2">
                  <span className="bg-gradient-to-r from-yellow-400 to-orange-400 text-gray-900 text-sm font-bold px-4 py-1 rounded-full shadow-lg">
                    Paling Popular
                  </span>
                </div>
              )}

              {/* Plan header */}
              <div className="text-center mb-8">
                <h3 className={`text-xl font-semibold mb-2 ${plan.popular ? "text-white" : "text-gray-900"}`}>
                  {plan.name}
                </h3>
                <div className="flex items-baseline justify-center gap-1">
                  <span className={`text-4xl font-bold ${plan.popular ? "text-white" : "text-gray-900"}`}>
                    {plan.price}
                  </span>
                  <span className={plan.popular ? "text-blue-200" : "text-gray-500"}>
                    {plan.period}
                  </span>
                </div>
                <p className={`mt-2 text-sm ${plan.popular ? "text-blue-200" : "text-gray-500"}`}>
                  {plan.description}
                </p>
              </div>

              {/* Features */}
              <ul className="space-y-4 mb-8">
                {plan.features.map((feature, i) => (
                  <li key={i} className="flex items-start gap-3">
                    <svg
                      className={`w-5 h-5 flex-shrink-0 ${plan.popular ? "text-green-300" : "text-green-500"}`}
                      fill="currentColor"
                      viewBox="0 0 20 20"
                    >
                      <path
                        fillRule="evenodd"
                        d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z"
                        clipRule="evenodd"
                      />
                    </svg>
                    <span className={plan.popular ? "text-white" : "text-gray-700"}>
                      {feature}
                    </span>
                  </li>
                ))}
                {plan.limitations.map((limitation, i) => (
                  <li key={`limit-${i}`} className="flex items-start gap-3">
                    <svg
                      className={`w-5 h-5 flex-shrink-0 ${plan.popular ? "text-red-300" : "text-gray-400"}`}
                      fill="currentColor"
                      viewBox="0 0 20 20"
                    >
                      <path
                        fillRule="evenodd"
                        d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z"
                        clipRule="evenodd"
                      />
                    </svg>
                    <span className={plan.popular ? "text-blue-200" : "text-gray-400"}>
                      {limitation}
                    </span>
                  </li>
                ))}
              </ul>

              {/* CTA */}
              <Link
                href={plan.ctaLink}
                className={`block w-full py-3 px-6 rounded-xl font-semibold text-center transition-all ${
                  plan.popular
                    ? "bg-white text-blue-600 hover:bg-gray-100"
                    : "bg-gray-900 text-white hover:bg-gray-800"
                }`}
              >
                {plan.cta}
              </Link>
            </div>
          ))}
        </div>

        {/* Guarantee */}
        <div className="text-center mt-12">
          <div className="inline-flex items-center gap-3 bg-green-50 text-green-700 px-6 py-3 rounded-full">
            <svg className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
            </svg>
            <span className="font-medium">30 hari money-back guarantee. Tiada risiko.</span>
          </div>
        </div>
      </div>
    </section>
  );
}
