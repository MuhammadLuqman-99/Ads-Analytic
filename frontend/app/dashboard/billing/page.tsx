"use client";

import { Check, CreditCard, Download } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";

const plans = [
  {
    id: "free",
    name: "Free",
    price: 0,
    description: "For individuals getting started",
    features: [
      "1 ad account connection",
      "7 days data retention",
      "Basic analytics",
      "Email support",
    ],
  },
  {
    id: "pro",
    name: "Pro",
    price: 99,
    description: "For growing businesses",
    popular: true,
    features: [
      "5 ad account connections",
      "30 days data retention",
      "Advanced analytics",
      "Custom reports",
      "Priority support",
      "API access",
    ],
  },
  {
    id: "business",
    name: "Business",
    price: 299,
    description: "For large organizations",
    features: [
      "Unlimited ad accounts",
      "Unlimited data retention",
      "White-label reports",
      "Dedicated account manager",
      "Custom integrations",
      "SSO & advanced security",
    ],
  },
];

const invoices = [
  { id: "INV-001", date: "Jan 1, 2024", amount: "RM 99.00", status: "Paid" },
  { id: "INV-002", date: "Dec 1, 2023", amount: "RM 99.00", status: "Paid" },
  { id: "INV-003", date: "Nov 1, 2023", amount: "RM 99.00", status: "Paid" },
];

export default function BillingPage() {
  const currentPlan = "pro";

  return (
    <div>
      {/* Header */}
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-slate-900">Billing</h1>
        <p className="mt-1 text-slate-500">
          Manage your subscription and billing information
        </p>
      </div>

      {/* Current Plan */}
      <Card className="bg-white border-slate-200 mb-8">
        <CardHeader>
          <CardTitle className="text-slate-900">Current Plan</CardTitle>
          <CardDescription>
            You are currently on the <span className="font-semibold">Pro</span> plan
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex items-center justify-between">
            <div>
              <p className="text-3xl font-bold text-slate-900">
                RM 99<span className="text-lg font-normal text-slate-500">/month</span>
              </p>
              <p className="text-sm text-slate-500 mt-1">Next billing date: Feb 1, 2024</p>
            </div>
            <Button variant="outline">Manage Subscription</Button>
          </div>
        </CardContent>
      </Card>

      {/* Plans */}
      <div className="mb-8">
        <h2 className="text-xl font-semibold text-slate-900 mb-4">Available Plans</h2>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
          {plans.map((plan) => (
            <Card
              key={plan.id}
              className={cn(
                "bg-white relative",
                plan.popular
                  ? "border-blue-500 border-2"
                  : "border-slate-200"
              )}
            >
              {plan.popular && (
                <Badge className="absolute -top-3 left-1/2 -translate-x-1/2 bg-blue-600">
                  Most Popular
                </Badge>
              )}
              <CardHeader>
                <CardTitle className="text-slate-900">{plan.name}</CardTitle>
                <CardDescription>{plan.description}</CardDescription>
              </CardHeader>
              <CardContent>
                <div className="mb-6">
                  <span className="text-3xl font-bold text-slate-900">
                    RM {plan.price}
                  </span>
                  <span className="text-slate-500">/month</span>
                </div>
                <ul className="space-y-3 mb-6">
                  {plan.features.map((feature, index) => (
                    <li key={index} className="flex items-center gap-2 text-sm">
                      <Check className="h-4 w-4 text-emerald-500" />
                      <span className="text-slate-600">{feature}</span>
                    </li>
                  ))}
                </ul>
                <Button
                  className="w-full"
                  variant={currentPlan === plan.id ? "outline" : "default"}
                  disabled={currentPlan === plan.id}
                >
                  {currentPlan === plan.id ? "Current Plan" : "Upgrade"}
                </Button>
              </CardContent>
            </Card>
          ))}
        </div>
      </div>

      {/* Payment Method */}
      <Card className="bg-white border-slate-200 mb-8">
        <CardHeader>
          <CardTitle className="text-slate-900">Payment Method</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-4">
              <div className="w-12 h-8 bg-slate-100 rounded flex items-center justify-center">
                <CreditCard className="h-5 w-5 text-slate-600" />
              </div>
              <div>
                <p className="font-medium text-slate-900">Visa ending in 4242</p>
                <p className="text-sm text-slate-500">Expires 12/2025</p>
              </div>
            </div>
            <Button variant="outline" size="sm">Update</Button>
          </div>
        </CardContent>
      </Card>

      {/* Invoices */}
      <Card className="bg-white border-slate-200">
        <CardHeader>
          <CardTitle className="text-slate-900">Invoice History</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="divide-y divide-slate-100">
            {invoices.map((invoice) => (
              <div
                key={invoice.id}
                className="flex items-center justify-between py-4 first:pt-0 last:pb-0"
              >
                <div>
                  <p className="font-medium text-slate-900">{invoice.id}</p>
                  <p className="text-sm text-slate-500">{invoice.date}</p>
                </div>
                <div className="flex items-center gap-4">
                  <Badge variant="outline" className="text-emerald-600 border-emerald-200">
                    {invoice.status}
                  </Badge>
                  <span className="font-medium text-slate-900">{invoice.amount}</span>
                  <Button variant="ghost" size="sm">
                    <Download className="h-4 w-4" />
                  </Button>
                </div>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
