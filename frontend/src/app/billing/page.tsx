"use client";

import React from "react";
import { CreditCard, Zap, Download, Check } from "lucide-react";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";

export default function BillingPage() {
  const currentPlan = {
    name: "Enterprise Dedicated",
    price: "$24,999",
    period: "month",
    included: "50,000,000 requests/mo",
    overage: "$0.0008 / req",
    status: "ACTIVE",
  };

  const usageStats = {
    evaluated: 8430000,
    quota: 50000000,
    percentUsed: 16.86,
    investigations: 42000,
    storageGB: 240,
  };

  const invoices = [
    { id: "inv_2026_07", date: "2026-07-31", amount: "$24,999.00", status: "PAID" },
    { id: "inv_2026_06", date: "2026-06-30", amount: "$24,999.00", status: "PAID" },
    { id: "inv_2026_05", date: "2026-05-31", amount: "$24,999.00", status: "PAID" },
  ];

  return (
    <div className="flex-1 p-6 md:p-8 space-y-8 max-w-6xl mx-auto">
      {/* Header */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 border-b border-slate-800 pb-5">
        <div>
          <h1 className="text-2xl font-bold text-white tracking-tight flex items-center gap-2.5">
            <CreditCard className="size-6 text-indigo-400" />
            <span>Subscription & Billing</span>
          </h1>
          <p className="text-sm text-slate-400 mt-1">
            Manage your enterprise tier, review real-time usage quotas, and download tax invoices.
          </p>
        </div>
      </div>

      {/* Plan & Usage Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Active Plan Card */}
        <Card className="bg-slate-900/80 border-slate-800 lg:col-span-1">
          <CardHeader>
            <div className="flex items-center justify-between">
              <Badge className="bg-indigo-500/20 text-indigo-300">CURRENT PLAN</Badge>
              <Badge className="bg-emerald-500/20 text-emerald-300">{currentPlan.status}</Badge>
            </div>
            <CardTitle className="text-xl font-bold text-white mt-2">{currentPlan.name}</CardTitle>
            <div className="flex items-baseline gap-1 mt-2">
              <span className="text-3xl font-black text-white">{currentPlan.price}</span>
              <span className="text-xs text-slate-400">/{currentPlan.period}</span>
            </div>
          </CardHeader>
          <CardContent className="space-y-3 text-xs text-slate-300">
            <div className="flex items-center gap-2">
              <Check className="size-4 text-emerald-400 shrink-0" />
              <span>{currentPlan.included}</span>
            </div>
            <div className="flex items-center gap-2">
              <Check className="size-4 text-emerald-400 shrink-0" />
              <span>Dedicated Multi-AZ Kubernetes Cluster</span>
            </div>
            <div className="flex items-center gap-2">
              <Check className="size-4 text-emerald-400 shrink-0" />
              <span>Custom XGBoost & Graph Neural Net Retraining</span>
            </div>
            <div className="flex items-center gap-2">
              <Check className="size-4 text-emerald-400 shrink-0" />
              <span>24/7 Priority SLA & Dedicated Solutions Architect</span>
            </div>
            <div className="pt-4 border-t border-slate-800">
              <Button variant="outline" className="w-full text-xs border-slate-700 hover:bg-slate-800">
                Change Subscription Plan
              </Button>
            </div>
          </CardContent>
        </Card>

        {/* Real-time Usage Meter */}
        <Card className="bg-slate-900/80 border-slate-800 lg:col-span-2">
          <CardHeader>
            <CardTitle className="text-base font-semibold text-white flex items-center gap-2">
              <Zap className="size-4 text-amber-400" />
              <span>Billing Cycle Metering (August 2026)</span>
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-6">
            <div>
              <div className="flex items-center justify-between text-xs mb-2 font-mono">
                <span className="text-slate-400">Monthly Risk Checks</span>
                <span className="text-white font-bold">
                  {usageStats.evaluated.toLocaleString()} / {usageStats.quota.toLocaleString()} ({usageStats.percentUsed}%)
                </span>
              </div>
              <div className="h-2.5 w-full bg-slate-800 rounded-full overflow-hidden">
                <div
                  className="h-full bg-indigo-500 rounded-full"
                  style={{ width: `${usageStats.percentUsed}%` }}
                />
              </div>
            </div>

            <div className="grid grid-cols-2 md:grid-cols-3 gap-4 pt-2">
              <div className="p-3 bg-slate-800/40 rounded-lg border border-slate-800">
                <p className="text-[11px] text-slate-400 font-mono">AI Agent Investigations</p>
                <p className="text-lg font-bold text-white mt-1">{usageStats.investigations.toLocaleString()}</p>
              </div>
              <div className="p-3 bg-slate-800/40 rounded-lg border border-slate-800">
                <p className="text-[11px] text-slate-400 font-mono">Knowledge Graph Nodes</p>
                <p className="text-lg font-bold text-white mt-1">1.84M</p>
              </div>
              <div className="p-3 bg-slate-800/40 rounded-lg border border-slate-800">
                <p className="text-[11px] text-slate-400 font-mono">Encrypted Storage</p>
                <p className="text-lg font-bold text-white mt-1">{usageStats.storageGB} GB</p>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Invoices History */}
      <Card className="bg-slate-900/80 border-slate-800">
        <CardHeader>
          <CardTitle className="text-base font-semibold text-white">Invoice History</CardTitle>
        </CardHeader>
        <CardContent>
          <table className="w-full text-left text-xs">
            <thead className="border-b border-slate-800 text-slate-400 uppercase font-mono">
              <tr>
                <th className="pb-3">Invoice ID</th>
                <th className="pb-3">Billing Date</th>
                <th className="pb-3">Amount</th>
                <th className="pb-3">Status</th>
                <th className="pb-3 text-right">Download</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-800/60 font-mono">
              {invoices.map((inv) => (
                <tr key={inv.id} className="hover:bg-slate-800/30">
                  <td className="py-3 text-indigo-300 font-medium">{inv.id}</td>
                  <td className="py-3 text-slate-400">{inv.date}</td>
                  <td className="py-3 text-white font-bold">{inv.amount}</td>
                  <td className="py-3">
                    <Badge className="bg-emerald-500/20 text-emerald-300">{inv.status}</Badge>
                  </td>
                  <td className="py-3 text-right">
                    <Button variant="ghost" size="sm" className="h-7 text-xs text-slate-400 hover:text-white">
                      <Download className="size-3 mr-1" /> PDF
                    </Button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </CardContent>
      </Card>
    </div>
  );
}
