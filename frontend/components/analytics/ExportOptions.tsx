"use client";

import { useState } from "react";
import { Download, FileText, Mail, Clock } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";

interface ExportOptionsProps {
  onExportCSV: () => void;
  onExportPDF: () => void;
  onScheduleReport?: () => void;
  isExporting?: boolean;
  reportTitle?: string;
}

export function ExportOptions({
  onExportCSV,
  onExportPDF,
  onScheduleReport,
  isExporting,
  reportTitle = "Analytics Report",
}: ExportOptionsProps) {
  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="outline" disabled={isExporting}>
          <Download className="h-4 w-4 mr-2" />
          {isExporting ? "Exporting..." : "Export"}
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" className="w-56">
        <DropdownMenuItem onClick={onExportCSV}>
          <FileText className="h-4 w-4 mr-2" />
          Download as CSV
        </DropdownMenuItem>
        <DropdownMenuItem onClick={onExportPDF}>
          <FileText className="h-4 w-4 mr-2" />
          Download as PDF
        </DropdownMenuItem>
        <DropdownMenuSeparator />
        <DropdownMenuItem disabled className="opacity-60">
          <Mail className="h-4 w-4 mr-2" />
          Schedule Report (Email)
          <Badge variant="secondary" className="ml-auto text-xs">
            Soon
          </Badge>
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}

// CSV Export utility function
export function exportToCSV(data: Record<string, unknown>[], filename: string) {
  if (data.length === 0) return;

  const headers = Object.keys(data[0]);
  const csvContent = [
    headers.join(","),
    ...data.map((row) =>
      headers
        .map((header) => {
          const value = row[header];
          const stringValue = String(value ?? "");
          return `"${stringValue.replace(/"/g, '""')}"`;
        })
        .join(",")
    ),
  ].join("\n");

  const blob = new Blob([csvContent], { type: "text/csv;charset=utf-8;" });
  const link = document.createElement("a");
  const url = URL.createObjectURL(blob);
  link.setAttribute("href", url);
  link.setAttribute("download", `${filename}-${new Date().toISOString().split("T")[0]}.csv`);
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
}

// PDF Report Component (for react-pdf)
export function ReportPDFContent({
  title,
  dateRange,
  metrics,
  data,
}: {
  title: string;
  dateRange: string;
  metrics: { label: string; value: string }[];
  data: Record<string, unknown>[];
}) {
  // This is a placeholder - actual PDF generation will be done dynamically
  return null;
}

// PDF Export utility using browser print
export async function exportToPDF(elementId: string, filename: string) {
  const element = document.getElementById(elementId);
  if (!element) return;

  // Create a new window for printing
  const printWindow = window.open("", "_blank");
  if (!printWindow) return;

  // Get computed styles
  const styles = Array.from(document.styleSheets)
    .map((styleSheet) => {
      try {
        return Array.from(styleSheet.cssRules)
          .map((rule) => rule.cssText)
          .join("\n");
      } catch {
        return "";
      }
    })
    .join("\n");

  printWindow.document.write(`
    <!DOCTYPE html>
    <html>
      <head>
        <title>${filename}</title>
        <style>
          ${styles}
          @media print {
            body { -webkit-print-color-adjust: exact !important; print-color-adjust: exact !important; }
          }
          body { font-family: system-ui, -apple-system, sans-serif; padding: 20px; }
        </style>
      </head>
      <body>
        ${element.innerHTML}
      </body>
    </html>
  `);

  printWindow.document.close();
  printWindow.focus();

  // Wait for content to load then print
  setTimeout(() => {
    printWindow.print();
    printWindow.close();
  }, 500);
}

// Schedule Report Modal Content
export function ScheduleReportForm() {
  return (
    <Card className="bg-white border-slate-200">
      <CardHeader>
        <div className="flex items-center gap-2">
          <Clock className="h-5 w-5 text-slate-500" />
          <CardTitle className="text-slate-900">Schedule Report</CardTitle>
          <Badge variant="warning">Coming Soon</Badge>
        </div>
      </CardHeader>
      <CardContent>
        <p className="text-slate-500 text-sm">
          Automated report scheduling will allow you to receive reports via email
          on a weekly or monthly basis. This feature is currently in development.
        </p>
      </CardContent>
    </Card>
  );
}
