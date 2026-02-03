// Shared pricing configuration - Single source of truth
// Update prices here and they will reflect across the entire app

export type PlanId = "free" | "pro" | "business";

export interface PlanFeature {
  text: string;
  included: boolean;
}

export interface Plan {
  id: PlanId;
  name: string;
  nameMs: string; // Malay name
  price: number;
  currency: string;
  period: string;
  periodMs: string; // Malay period
  description: string;
  descriptionMs: string; // Malay description
  accounts: number | "unlimited";
  apiCalls: string;
  dataRetention: string;
  features: PlanFeature[];
  popular?: boolean;
  ctaText: string;
  ctaTextMs: string; // Malay CTA
  ctaLink: string;
}

export const CURRENCY = "RM";
export const CURRENCY_SYMBOL = "RM";

export const plans: Plan[] = [
  {
    id: "free",
    name: "Free",
    nameMs: "Percuma",
    price: 0,
    currency: CURRENCY,
    period: "forever",
    periodMs: "selamanya",
    description: "For beginners just getting started",
    descriptionMs: "Untuk peniaga yang baru bermula",
    accounts: 1,
    apiCalls: "1,000/month",
    dataRetention: "7 days",
    features: [
      { text: "1 platform connection", included: true },
      { text: "Basic dashboard", included: true },
      { text: "7 days data history", included: true },
      { text: "CSV export", included: true },
      { text: "Email support", included: true },
      { text: "Cross-platform analytics", included: false },
      { text: "Auto-sync", included: false },
    ],
    popular: false,
    ctaText: "Start Free",
    ctaTextMs: "Mula Percuma",
    ctaLink: "/register",
  },
  {
    id: "pro",
    name: "Pro",
    nameMs: "Pro",
    price: 99,
    currency: CURRENCY,
    period: "month",
    periodMs: "bulan",
    description: "For serious businesses looking to grow",
    descriptionMs: "Untuk peniaga yang serius grow bisnes",
    accounts: 3,
    apiCalls: "10,000/month",
    dataRetention: "30 days",
    features: [
      { text: "3 platform connections", included: true },
      { text: "Cross-platform ROAS", included: true },
      { text: "30 days data history", included: true },
      { text: "Hourly auto-sync", included: true },
      { text: "PDF & Excel export", included: true },
      { text: "Priority support", included: true },
      { text: "Custom dashboard", included: true },
      { text: "Team collaboration (3 users)", included: true },
    ],
    popular: true,
    ctaText: "Try 14 Days Free",
    ctaTextMs: "Cuba 14 Hari Percuma",
    ctaLink: "/register?plan=pro",
  },
  {
    id: "business",
    name: "Business",
    nameMs: "Business",
    price: 299,
    currency: CURRENCY,
    period: "month",
    periodMs: "bulan",
    description: "For agencies and enterprise businesses",
    descriptionMs: "Untuk agency dan bisnes enterprise",
    accounts: "unlimited",
    apiCalls: "Unlimited",
    dataRetention: "Unlimited",
    features: [
      { text: "Unlimited platform connections", included: true },
      { text: "Unlimited data history", included: true },
      { text: "Real-time sync", included: true },
      { text: "White-label reports", included: true },
      { text: "API access", included: true },
      { text: "Dedicated account manager", included: true },
      { text: "Custom integrations", included: true },
      { text: "Unlimited team members", included: true },
      { text: "SSO authentication", included: true },
      { text: "SLA guarantee", included: true },
    ],
    popular: false,
    ctaText: "Contact Us",
    ctaTextMs: "Hubungi Kami",
    ctaLink: "/contact",
  },
];

// Helper functions
export function getPlanById(id: PlanId): Plan | undefined {
  return plans.find((plan) => plan.id === id);
}

export function formatPrice(price: number, currency: string = CURRENCY): string {
  if (price === 0) return `${currency} 0`;
  return `${currency} ${price}`;
}

export function getPlanFeatures(planId: PlanId): string[] {
  const plan = getPlanById(planId);
  if (!plan) return [];
  return plan.features.filter((f) => f.included).map((f) => f.text);
}

export function getPlanLimitations(planId: PlanId): string[] {
  const plan = getPlanById(planId);
  if (!plan) return [];
  return plan.features.filter((f) => !f.included).map((f) => f.text);
}

// Usage stats limits per plan
export const planLimits: Record<PlanId, { accounts: number; apiCalls: number; storage: number }> = {
  free: { accounts: 1, apiCalls: 1000, storage: 1 },
  pro: { accounts: 3, apiCalls: 10000, storage: 10 },
  business: { accounts: 100, apiCalls: 1000000, storage: 100 },
};
