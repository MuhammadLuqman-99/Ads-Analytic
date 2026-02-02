import { Metadata } from "next";
import { AdminLayout } from "@/components/admin";

export const metadata: Metadata = {
  title: "Admin Panel | AdsAnalytic",
  description: "Internal admin dashboard for analytics and metrics",
  robots: "noindex, nofollow",
};

export default function AdminRootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return <AdminLayout>{children}</AdminLayout>;
}
