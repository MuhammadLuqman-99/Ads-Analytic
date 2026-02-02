"use client";

import { useState } from "react";
import { format } from "date-fns";
import {
  CreditCard,
  Check,
  Zap,
  Building2,
  Download,
  Plus,
  Trash2,
  AlertCircle,
  ArrowRight,
} from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { cn } from "@/lib/utils";

type Plan = "free" | "pro" | "business";

interface PlanDetails {
  name: string;
  price: number;
  period: string;
  accounts: number;
  apiCalls: string;
  features: string[];
  popular?: boolean;
}

interface UsageStats {
  accounts: { used: number; limit: number };
  apiCalls: { used: number; limit: number };
  storage: { used: number; limit: number };
}

interface PaymentMethod {
  id: string;
  type: "visa" | "mastercard" | "amex";
  last4: string;
  expiry: string;
  isDefault: boolean;
}

interface Invoice {
  id: string;
  date: Date;
  amount: number;
  status: "paid" | "pending" | "failed";
  description: string;
}

const plans: Record<Plan, PlanDetails> = {
  free: {
    name: "Free",
    price: 0,
    period: "forever",
    accounts: 2,
    apiCalls: "1,000/month",
    features: [
      "2 ad accounts",
      "Basic analytics",
      "7-day data retention",
      "Email support",
    ],
  },
  pro: {
    name: "Pro",
    price: 29,
    period: "month",
    accounts: 5,
    apiCalls: "10,000/month",
    features: [
      "5 ad accounts",
      "Advanced analytics",
      "30-day data retention",
      "Priority support",
      "Custom reports",
      "Team collaboration",
    ],
    popular: true,
  },
  business: {
    name: "Business",
    price: 99,
    period: "month",
    accounts: 20,
    apiCalls: "Unlimited",
    features: [
      "20 ad accounts",
      "Enterprise analytics",
      "Unlimited data retention",
      "Dedicated support",
      "API access",
      "White-label options",
      "SSO integration",
    ],
  },
};

const mockUsageStats: UsageStats = {
  accounts: { used: 3, limit: 5 },
  apiCalls: { used: 7500, limit: 10000 },
  storage: { used: 2.5, limit: 10 },
};

const mockPaymentMethods: PaymentMethod[] = [
  { id: "1", type: "visa", last4: "4242", expiry: "12/25", isDefault: true },
  { id: "2", type: "mastercard", last4: "8888", expiry: "06/24", isDefault: false },
];

const mockInvoices: Invoice[] = [
  { id: "INV-001", date: new Date(2024, 0, 1), amount: 29, status: "paid", description: "Pro Plan - January 2024" },
  { id: "INV-002", date: new Date(2023, 11, 1), amount: 29, status: "paid", description: "Pro Plan - December 2023" },
  { id: "INV-003", date: new Date(2023, 10, 1), amount: 29, status: "paid", description: "Pro Plan - November 2023" },
  { id: "INV-004", date: new Date(2023, 9, 1), amount: 29, status: "paid", description: "Pro Plan - October 2023" },
];

const cardIcons: Record<string, string> = {
  visa: "V",
  mastercard: "M",
  amex: "A",
};

export function BillingSettings() {
  const [currentPlan] = useState<Plan>("pro");
  const [usageStats] = useState(mockUsageStats);
  const [paymentMethods, setPaymentMethods] = useState(mockPaymentMethods);
  const [invoices] = useState(mockInvoices);
  const [isChangingPlan, setIsChangingPlan] = useState(false);

  const handleUpgrade = async (plan: Plan) => {
    setIsChangingPlan(true);
    await new Promise((resolve) => setTimeout(resolve, 1500));
    console.log("Upgrading to:", plan);
    setIsChangingPlan(false);
  };

  const handleRemovePaymentMethod = (id: string) => {
    if (confirm("Are you sure you want to remove this payment method?")) {
      setPaymentMethods((prev) => prev.filter((p) => p.id !== id));
    }
  };

  const handleSetDefault = (id: string) => {
    setPaymentMethods((prev) =>
      prev.map((p) => ({ ...p, isDefault: p.id === id }))
    );
  };

  return (
    <div className="space-y-6">
      {/* Current Plan */}
      <Card className="bg-white border-slate-200">
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle className="flex items-center gap-2 text-slate-900">
                <Zap className="h-5 w-5" />
                Current Plan
              </CardTitle>
              <CardDescription>
                You are currently on the {plans[currentPlan].name} plan
              </CardDescription>
            </div>
            <Badge className="bg-blue-100 text-blue-700 text-lg px-4 py-1">
              {plans[currentPlan].name}
            </Badge>
          </div>
        </CardHeader>
        <CardContent>
          <div className="flex items-baseline gap-2 mb-6">
            <span className="text-4xl font-bold text-slate-900">
              ${plans[currentPlan].price}
            </span>
            <span className="text-slate-500">/{plans[currentPlan].period}</span>
          </div>

          {/* Usage Stats */}
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4 p-4 bg-slate-50 rounded-lg">
            <div>
              <div className="flex items-center justify-between text-sm mb-1">
                <span className="text-slate-600">Ad Accounts</span>
                <span className="font-medium text-slate-900">
                  {usageStats.accounts.used} / {usageStats.accounts.limit}
                </span>
              </div>
              <div className="h-2 bg-slate-200 rounded-full overflow-hidden">
                <div
                  className="h-full bg-blue-500 rounded-full"
                  style={{
                    width: `${(usageStats.accounts.used / usageStats.accounts.limit) * 100}%`,
                  }}
                />
              </div>
            </div>
            <div>
              <div className="flex items-center justify-between text-sm mb-1">
                <span className="text-slate-600">API Calls</span>
                <span className="font-medium text-slate-900">
                  {usageStats.apiCalls.used.toLocaleString()} / {usageStats.apiCalls.limit.toLocaleString()}
                </span>
              </div>
              <div className="h-2 bg-slate-200 rounded-full overflow-hidden">
                <div
                  className={cn(
                    "h-full rounded-full",
                    usageStats.apiCalls.used / usageStats.apiCalls.limit > 0.8
                      ? "bg-amber-500"
                      : "bg-blue-500"
                  )}
                  style={{
                    width: `${(usageStats.apiCalls.used / usageStats.apiCalls.limit) * 100}%`,
                  }}
                />
              </div>
            </div>
            <div>
              <div className="flex items-center justify-between text-sm mb-1">
                <span className="text-slate-600">Storage</span>
                <span className="font-medium text-slate-900">
                  {usageStats.storage.used} GB / {usageStats.storage.limit} GB
                </span>
              </div>
              <div className="h-2 bg-slate-200 rounded-full overflow-hidden">
                <div
                  className="h-full bg-blue-500 rounded-full"
                  style={{
                    width: `${(usageStats.storage.used / usageStats.storage.limit) * 100}%`,
                  }}
                />
              </div>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Plan Comparison */}
      <Card className="bg-white border-slate-200">
        <CardHeader>
          <CardTitle className="text-slate-900">Compare Plans</CardTitle>
          <CardDescription>
            Choose the plan that best fits your needs
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
            {(Object.keys(plans) as Plan[]).map((planKey) => {
              const plan = plans[planKey];
              const isCurrent = planKey === currentPlan;

              return (
                <div
                  key={planKey}
                  className={cn(
                    "relative p-6 rounded-xl border-2 transition-all",
                    isCurrent
                      ? "border-blue-500 bg-blue-50"
                      : "border-slate-200 hover:border-slate-300",
                    plan.popular && !isCurrent && "border-purple-300"
                  )}
                >
                  {plan.popular && (
                    <Badge className="absolute -top-3 left-1/2 -translate-x-1/2 bg-purple-600">
                      Most Popular
                    </Badge>
                  )}

                  <div className="text-center mb-6">
                    <h3 className="text-xl font-bold text-slate-900">{plan.name}</h3>
                    <div className="mt-2">
                      <span className="text-3xl font-bold text-slate-900">
                        ${plan.price}
                      </span>
                      <span className="text-slate-500">/{plan.period}</span>
                    </div>
                  </div>

                  <ul className="space-y-3 mb-6">
                    {plan.features.map((feature, index) => (
                      <li key={index} className="flex items-center gap-2">
                        <Check className="h-4 w-4 text-emerald-500 flex-shrink-0" />
                        <span className="text-sm text-slate-600">{feature}</span>
                      </li>
                    ))}
                  </ul>

                  {isCurrent ? (
                    <Button className="w-full" disabled>
                      Current Plan
                    </Button>
                  ) : planKey === "free" && currentPlan !== "free" ? (
                    <Button variant="outline" className="w-full">
                      Downgrade
                    </Button>
                  ) : (
                    <Button
                      className="w-full"
                      onClick={() => handleUpgrade(planKey)}
                      disabled={isChangingPlan}
                    >
                      {isChangingPlan ? "Processing..." : "Upgrade"}
                      <ArrowRight className="h-4 w-4 ml-2" />
                    </Button>
                  )}
                </div>
              );
            })}
          </div>
        </CardContent>
      </Card>

      {/* Payment Methods */}
      <Card className="bg-white border-slate-200">
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle className="flex items-center gap-2 text-slate-900">
                <CreditCard className="h-5 w-5" />
                Payment Methods
              </CardTitle>
              <CardDescription>
                Manage your payment methods for billing
              </CardDescription>
            </div>
            <Button variant="outline">
              <Plus className="h-4 w-4 mr-2" />
              Add Card
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          <div className="space-y-3">
            {paymentMethods.map((method) => (
              <div
                key={method.id}
                className={cn(
                  "flex items-center justify-between p-4 rounded-lg border",
                  method.isDefault ? "border-blue-200 bg-blue-50" : "border-slate-200"
                )}
              >
                <div className="flex items-center gap-4">
                  <div className="w-12 h-8 rounded bg-slate-800 flex items-center justify-center text-white font-bold">
                    {cardIcons[method.type]}
                  </div>
                  <div>
                    <p className="font-medium text-slate-900">
                      {method.type.charAt(0).toUpperCase() + method.type.slice(1)} ending in {method.last4}
                    </p>
                    <p className="text-sm text-slate-500">Expires {method.expiry}</p>
                  </div>
                </div>
                <div className="flex items-center gap-2">
                  {method.isDefault ? (
                    <Badge className="bg-blue-100 text-blue-700">Default</Badge>
                  ) : (
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => handleSetDefault(method.id)}
                    >
                      Set as Default
                    </Button>
                  )}
                  {!method.isDefault && (
                    <Button
                      variant="ghost"
                      size="icon"
                      className="h-8 w-8 text-red-500 hover:text-red-600"
                      onClick={() => handleRemovePaymentMethod(method.id)}
                    >
                      <Trash2 className="h-4 w-4" />
                    </Button>
                  )}
                </div>
              </div>
            ))}

            {paymentMethods.length === 0 && (
              <div className="text-center py-8">
                <CreditCard className="h-12 w-12 text-slate-300 mx-auto mb-3" />
                <p className="text-slate-500">No payment methods added</p>
                <Button className="mt-4">
                  <Plus className="h-4 w-4 mr-2" />
                  Add Payment Method
                </Button>
              </div>
            )}
          </div>

          <div className="mt-4 p-3 bg-slate-50 rounded-lg flex items-center gap-3">
            <AlertCircle className="h-5 w-5 text-slate-400" />
            <p className="text-sm text-slate-600">
              Your card information is securely stored and processed by Stripe.
            </p>
          </div>
        </CardContent>
      </Card>

      {/* Invoice History */}
      <Card className="bg-white border-slate-200">
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-slate-900">
            <Building2 className="h-5 w-5" />
            Invoice History
          </CardTitle>
          <CardDescription>
            View and download your past invoices
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow className="bg-slate-50">
                <TableHead>Invoice</TableHead>
                <TableHead>Date</TableHead>
                <TableHead>Description</TableHead>
                <TableHead>Amount</TableHead>
                <TableHead>Status</TableHead>
                <TableHead className="text-right">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {invoices.map((invoice) => (
                <TableRow key={invoice.id}>
                  <TableCell className="font-mono text-sm">{invoice.id}</TableCell>
                  <TableCell>{format(invoice.date, "MMM d, yyyy")}</TableCell>
                  <TableCell>{invoice.description}</TableCell>
                  <TableCell className="font-medium">${invoice.amount}</TableCell>
                  <TableCell>
                    <Badge
                      className={cn(
                        invoice.status === "paid" && "bg-emerald-100 text-emerald-700",
                        invoice.status === "pending" && "bg-amber-100 text-amber-700",
                        invoice.status === "failed" && "bg-red-100 text-red-700"
                      )}
                    >
                      {invoice.status.charAt(0).toUpperCase() + invoice.status.slice(1)}
                    </Badge>
                  </TableCell>
                  <TableCell className="text-right">
                    <Button variant="ghost" size="sm">
                      <Download className="h-4 w-4 mr-1" />
                      PDF
                    </Button>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </div>
  );
}
