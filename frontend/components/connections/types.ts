// UI-specific types for the connections components

export interface PlanLimits {
  currentPlan: "free" | "pro" | "business";
  accountsUsed: number;
  accountsLimit: number;
}
