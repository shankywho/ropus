"use client";

import React, { useState, useEffect } from "react";
import Link from "next/link";
import {
  ShieldAlert,
  ShieldCheck,
  AlertTriangle,
  PlayCircle,
  FolderKanban,
  Sliders,
  Clock,
  Cpu,
  Layers,
  Lock,
  ArrowRight,
  TrendingUp,
  Activity,
  Zap,
  CheckCircle2,
  XCircle,
} from "lucide-react";
import {
  PieChart,
  Pie,
  Cell,
  AreaChart,
  Area,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
} from "recharts";
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";

// Synthetic Timeline Data for 24h Hourly Evaluation Volume & Latency
const latencyTimelineData = [
  { time: "00:00", volume: 24000, p95_latency: 12.4, p50_latency: 4.1 },
  { time: "02:00", volume: 18000, p95_latency: 11.8, p50_latency: 3.9 },
  { time: "04:00", volume: 15000, p95_latency: 11.2, p50_latency: 3.8 },
  { time: "06:00", volume: 28000, p95_latency: 13.1, p50_latency: 4.3 },
  { time: "08:00", volume: 62000, p95_latency: 15.6, p50_latency: 4.8 },
  { time: "10:00", volume: 94000, p95_latency: 16.2, p50_latency: 5.1 },
  { time: "12:00", volume: 112000, p95_latency: 17.8, p50_latency: 5.4 },
  { time: "14:00", volume: 108000, p95_latency: 16.9, p50_latency: 5.2 },
  { time: "16:00", volume: 98000, p95_latency: 15.4, p50_latency: 4.9 },
  { time: "18:00", volume: 124000, p95_latency: 18.5, p50_latency: 5.8 },
  { time: "20:00", volume: 135000, p95_latency: 19.2, p50_latency: 6.1 },
  { time: "22:00", volume: 86000, p95_latency: 14.7, p50_latency: 4.6 },
];

// Outcome Distribution Data
const outcomeDistributionData = [
  { name: "ALLOW (84.5%)", value: 84.5, count: 1085420, color: "#10b981" },
  { name: "STEP-UP 3DS (8.2%)", value: 8.2, count: 105330, color: "#6366f1" },
  { name: "MANUAL REVIEW (4.8%)", value: 4.8, count: 61650, color: "#f59e0b" },
  { name: "DECLINE (2.5%)", value: 2.5, count: 32120, color: "#f43f5e" },
];

// Recent Flagged Decisions Stream
const recentFlaggedTransactions = [
  {
    id: "txn_88a91c2b",
    amount: "₹95,000",
    action: "DECLINE_RECOMMENDATION",
    score: 95,
    reason: "HIGH_IP_VELOCITY_BLOCK",
    time: "2 mins ago",
  },
  {
    id: "txn_44e73f1a",
    amount: "₹48,000",
    action: "MANUAL_REVIEW",
    score: 72,
    reason: "NEW_DEVICE_HIGH_TICKET",
    time: "6 mins ago",
  },
  {
    id: "txn_19b88c44",
    amount: "₹18,500",
    action: "STEP_UP_RECOMMENDATION",
    score: 52,
    reason: "OFF_HOURS_ACTIVITY",
    time: "11 mins ago",
  },
  {
    id: "txn_91a02f5e",
    amount: "₹1,250",
    action: "ALLOW_RECOMMENDATION",
    score: 12,
    reason: "LOW_ANOMALY_BASELINE",
    time: "14 mins ago",
  },
];

export default function OverviewAnalyticsDashboard() {
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    setMounted(true);
  }, []);

  const renderActionBadge = (action: string) => {
    switch (action) {
      case "ALLOW_RECOMMENDATION":
        return <Badge variant="success" className="text-[10px]">ALLOW</Badge>;
      case "STEP_UP_RECOMMENDATION":
        return <Badge variant="info" className="text-[10px]">STEP-UP (3DS)</Badge>;
      case "MANUAL_REVIEW":
        return <Badge variant="warning" className="text-[10px]">REVIEW (24h)</Badge>;
      case "DECLINE_RECOMMENDATION":
        return <Badge variant="danger" className="text-[10px]">DECLINE</Badge>;
      default:
        return <Badge variant="outline" className="text-[10px]">{action}</Badge>;
    }
  };

  return (
    <div className="p-6 md:p-10 max-w-7xl mx-auto w-full space-y-8">
      {/* Top Banner */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 border-b border-slate-800 pb-6">
        <div>
          <div className="flex items-center gap-2 mb-1">
            <h1 className="text-2xl font-bold tracking-tight text-white flex items-center gap-2">
              <Activity className="size-6 text-indigo-400" />
              AI Risk Manager — Analytics Dashboard
            </h1>
            <Badge variant="success" className="gap-1">
              <span className="size-1.5 rounded-full bg-emerald-400 animate-pulse" />
              Live Telemetry
            </Badge>
          </div>
          <p className="text-sm text-slate-400">
            Real-time fraud decisioning overview, p95 latency tracking, and portfolio risk telemetry.
          </p>
        </div>

        <div className="flex items-center gap-3">
          <Link href="/playground">
            <Button className="bg-gradient-to-r from-indigo-600 to-purple-600 hover:from-indigo-500 hover:to-purple-500 text-white gap-2 shadow-lg shadow-indigo-500/25">
              <PlayCircle className="size-4" />
              Open Risk Playground
            </Button>
          </Link>
        </div>
      </div>

      {/* Top Row: KPI Metric Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        {/* Total Decisions (24h) */}
        <Card className="bg-slate-900/70 border-slate-800 shadow-lg backdrop-blur">
          <CardHeader className="pb-2">
            <div className="flex items-center justify-between">
              <CardDescription className="text-xs font-medium text-slate-400">
                Total Decisions (24h)
              </CardDescription>
              <Zap className="size-4 text-indigo-400" />
            </div>
            <CardTitle className="text-2xl font-bold font-mono text-white mt-1">
              1,284,520
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex items-center gap-1.5 text-xs text-emerald-400 font-medium">
              <TrendingUp className="size-3.5" />
              <span>+14.2%</span>
              <span className="text-slate-500 font-normal">vs previous 24h</span>
            </div>
          </CardContent>
        </Card>

        {/* System Latency (p95) */}
        <Card className="bg-slate-900/70 border-slate-800 shadow-lg backdrop-blur">
          <CardHeader className="pb-2">
            <div className="flex items-center justify-between">
              <CardDescription className="text-xs font-medium text-slate-400">
                System Latency (p95)
              </CardDescription>
              <Clock className="size-4 text-emerald-400" />
            </div>
            <CardTitle className="text-2xl font-bold font-mono text-emerald-400 mt-1 flex items-baseline gap-1.5">
              14.2 ms
              <span className="text-xs text-slate-500 font-normal font-sans">/ 100ms SLA</span>
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-xs text-slate-400 flex items-center gap-1">
              <Cpu className="size-3 text-indigo-400" />
              <span>ONNX Runtime sub-millisecond graph</span>
            </p>
          </CardContent>
        </Card>

        {/* Active Rules */}
        <Card className="bg-slate-900/70 border-slate-800 shadow-lg backdrop-blur">
          <CardHeader className="pb-2">
            <div className="flex items-center justify-between">
              <CardDescription className="text-xs font-medium text-slate-400">
                Rules Governance
              </CardDescription>
              <Sliders className="size-4 text-purple-400" />
            </div>
            <CardTitle className="text-2xl font-bold font-mono text-purple-400 mt-1">
              18 Active
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-xs text-slate-400 flex items-center justify-between">
              <span>Maker-Checker: Dual Control</span>
              <Badge variant="warning" className="text-[10px] py-0 px-1.5">2 Pending</Badge>
            </p>
          </CardContent>
        </Card>

        {/* Open Review Cases */}
        <Card className="bg-slate-900/70 border-slate-800 shadow-lg backdrop-blur">
          <CardHeader className="pb-2">
            <div className="flex items-center justify-between">
              <CardDescription className="text-xs font-medium text-slate-400">
                Manual Review Queue
              </CardDescription>
              <FolderKanban className="size-4 text-amber-400" />
            </div>
            <CardTitle className="text-2xl font-bold font-mono text-amber-400 mt-1">
              42 Open
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-xs text-slate-400 flex items-center gap-1">
              <span className="size-2 rounded-full bg-amber-400 animate-pulse" />
              <span>24h SLA Countdown Active</span>
            </p>
          </CardContent>
        </Card>
      </div>

      {/* Middle Section: Two Recharts Charts */}
      <div className="grid grid-cols-1 lg:grid-cols-12 gap-8 items-start">
        {/* Chart 1: Latency & Volume Timeline (7 cols) */}
        <div className="lg:col-span-7 space-y-6">
          <Card className="bg-slate-900/70 border-slate-800 shadow-xl backdrop-blur">
            <CardHeader className="pb-2">
              <div className="flex items-center justify-between">
                <div>
                  <CardTitle className="text-base text-white flex items-center gap-2">
                    <Clock className="size-4 text-indigo-400" />
                    Decision Latency &amp; Volume Timeline (24h)
                  </CardTitle>
                  <CardDescription className="text-xs text-slate-400">
                    p95 Latency guarantee (&lt;100ms budget) vs hourly transaction volume.
                  </CardDescription>
                </div>
                <Badge variant="outline" className="font-mono text-xs text-emerald-400 border-emerald-500/30 bg-emerald-500/10">
                  p95: 14.2ms
                </Badge>
              </div>
            </CardHeader>

            <CardContent className="pt-4">
              <div className="h-72 w-full">
                {mounted ? (
                  <ResponsiveContainer width="100%" height="100%">
                    <AreaChart data={latencyTimelineData} margin={{ top: 10, right: 10, left: -20, bottom: 0 }}>
                      <defs>
                        <linearGradient id="latencyGradient" x1="0" y1="0" x2="0" y2="1">
                          <stop offset="5%" stopColor="#6366f1" stopOpacity={0.4} />
                          <stop offset="95%" stopColor="#6366f1" stopOpacity={0.0} />
                        </linearGradient>
                      </defs>
                      <CartesianGrid strokeDasharray="3 3" stroke="#1e293b" vertical={false} />
                      <XAxis dataKey="time" stroke="#64748b" fontSize={11} tickLine={false} />
                      <YAxis stroke="#64748b" fontSize={11} tickLine={false} unit="ms" />
                      <Tooltip
                        contentStyle={{
                          backgroundColor: "#0f172a",
                          borderColor: "#334155",
                          borderRadius: "8px",
                          fontSize: "12px",
                          color: "#f8fafc",
                        }}
                      />
                      <Area
                        type="monotone"
                        dataKey="p95_latency"
                        name="p95 Latency (ms)"
                        stroke="#818cf8"
                        strokeWidth={2}
                        fillOpacity={1}
                        fill="url(#latencyGradient)"
                      />
                    </AreaChart>
                  </ResponsiveContainer>
                ) : (
                  <div className="h-full flex items-center justify-center text-slate-500 text-xs">
                    Loading timeline...
                  </div>
                )}
              </div>
            </CardContent>
          </Card>
        </div>

        {/* Chart 2: Outcome Distribution Pie Chart (5 cols) */}
        <div className="lg:col-span-5 space-y-6">
          <Card className="bg-slate-900/70 border-slate-800 shadow-xl backdrop-blur">
            <CardHeader className="pb-2">
              <CardTitle className="text-base text-white flex items-center gap-2">
                <ShieldCheck className="size-4 text-emerald-400" />
                Outcome Recommendation Breakdown
              </CardTitle>
              <CardDescription className="text-xs text-slate-400">
                Distribution across 1.28M evaluated transactions in the last 24h.
              </CardDescription>
            </CardHeader>

            <CardContent className="pt-2">
              <div className="h-56 w-full flex items-center justify-center">
                {mounted ? (
                  <ResponsiveContainer width="100%" height="100%">
                    <PieChart>
                      <Pie
                        data={outcomeDistributionData}
                        cx="50%"
                        cy="50%"
                        innerRadius={55}
                        outerRadius={80}
                        paddingAngle={4}
                        dataKey="value"
                      >
                        {outcomeDistributionData.map((entry, index) => (
                          <Cell key={`cell-${index}`} fill={entry.color} />
                        ))}
                      </Pie>
                      <Tooltip
                        contentStyle={{
                          backgroundColor: "#0f172a",
                          borderColor: "#334155",
                          borderRadius: "8px",
                          fontSize: "12px",
                          color: "#f8fafc",
                        }}
                      />
                    </PieChart>
                  </ResponsiveContainer>
                ) : (
                  <div className="h-full flex items-center justify-center text-slate-500 text-xs">
                    Loading distribution...
                  </div>
                )}
              </div>

              {/* Custom Legend Chips */}
              <div className="grid grid-cols-2 gap-2 mt-2 pt-3 border-t border-slate-800 text-xs">
                {outcomeDistributionData.map((item, idx) => (
                  <div key={idx} className="flex items-center gap-2 p-1.5 rounded bg-slate-950/60 border border-slate-800/80">
                    <span className="size-2 rounded-full shrink-0" style={{ backgroundColor: item.color }} />
                    <div className="truncate">
                      <p className="text-slate-300 font-medium truncate">{item.name}</p>
                      <p className="text-[10px] text-slate-500 font-mono">{item.count.toLocaleString()} txns</p>
                    </div>
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>
        </div>
      </div>

      {/* Bottom Section: Live Activity Feed & Quick Actions */}
      <div className="grid grid-cols-1 lg:grid-cols-12 gap-8 items-start">
        {/* Recent High-Impact Flagged Evaluations (7 cols) */}
        <div className="lg:col-span-7 space-y-6">
          <Card className="bg-slate-900/70 border-slate-800 shadow-xl backdrop-blur">
            <CardHeader className="pb-3">
              <div className="flex items-center justify-between">
                <CardTitle className="text-base text-white flex items-center gap-2">
                  <ShieldAlert className="size-4 text-rose-400" />
                  Live Decision Stream
                </CardTitle>
                <Link href="/playground" className="text-xs text-indigo-400 hover:underline flex items-center gap-1">
                  Test in Playground <ArrowRight className="size-3" />
                </Link>
              </div>
              <CardDescription className="text-xs text-slate-400">
                Recent decisions streamed from the Go orchestrator via Redpanda outbox.
              </CardDescription>
            </CardHeader>

            <CardContent className="space-y-3">
              {recentFlaggedTransactions.map((item) => (
                <div
                  key={item.id}
                  className="flex items-center justify-between p-3 rounded-lg bg-slate-950/70 border border-slate-800 hover:border-slate-700 transition-colors"
                >
                  <div className="space-y-1">
                    <div className="flex items-center gap-2">
                      <span className="text-xs font-mono font-semibold text-white">{item.id}</span>
                      {renderActionBadge(item.action)}
                    </div>
                    <p className="text-[11px] text-slate-400 font-mono flex items-center gap-1.5">
                      <span className="size-1.5 rounded-full bg-indigo-400" />
                      {item.reason}
                    </p>
                  </div>

                  <div className="text-right space-y-0.5">
                    <p className="text-xs font-mono font-bold text-white">{item.amount}</p>
                    <p className="text-[10px] text-slate-500 font-mono">{item.time}</p>
                  </div>
                </div>
              ))}
            </CardContent>
          </Card>
        </div>

        {/* Quick Launch & Architecture Guarantee Cards (5 cols) */}
        <div className="lg:col-span-5 space-y-4">
          <Card className="bg-slate-900/70 border-slate-800 shadow-xl backdrop-blur">
            <CardHeader className="pb-3">
              <CardTitle className="text-base text-white">Interactive Modules</CardTitle>
              <CardDescription className="text-xs text-slate-400">
                Access core features of the AI Risk Manager platform.
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-2.5">
              <Link href="/playground" className="block">
                <div className="p-3 rounded-lg bg-slate-950/70 border border-slate-800 hover:border-indigo-500/40 transition-all flex items-center justify-between group">
                  <div className="flex items-center gap-3">
                    <div className="size-8 rounded bg-emerald-500/10 text-emerald-400 flex items-center justify-center">
                      <PlayCircle className="size-4" />
                    </div>
                    <div>
                      <p className="text-xs font-semibold text-white group-hover:text-indigo-300">
                        Risk Evaluation Playground
                      </p>
                      <p className="text-[11px] text-slate-400">Simulate live transaction scoring</p>
                    </div>
                  </div>
                  <ArrowRight className="size-4 text-slate-600 group-hover:text-white" />
                </div>
              </Link>

              <Link href="/cases" className="block">
                <div className="p-3 rounded-lg bg-slate-950/70 border border-slate-800 hover:border-amber-500/40 transition-all flex items-center justify-between group">
                  <div className="flex items-center gap-3">
                    <div className="size-8 rounded bg-amber-500/10 text-amber-400 flex items-center justify-center">
                      <FolderKanban className="size-4" />
                    </div>
                    <div>
                      <p className="text-xs font-semibold text-white group-hover:text-amber-300">
                        Manual Review Queue
                      </p>
                      <p className="text-[11px] text-slate-400">Investigate cases with 24h SLA</p>
                    </div>
                  </div>
                  <ArrowRight className="size-4 text-slate-600 group-hover:text-white" />
                </div>
              </Link>

              <Link href="/rules" className="block">
                <div className="p-3 rounded-lg bg-slate-950/70 border border-slate-800 hover:border-purple-500/40 transition-all flex items-center justify-between group">
                  <div className="flex items-center gap-3">
                    <div className="size-8 rounded bg-purple-500/10 text-purple-400 flex items-center justify-center">
                      <Sliders className="size-4" />
                    </div>
                    <div>
                      <p className="text-xs font-semibold text-white group-hover:text-purple-300">
                        Rules Governance
                      </p>
                      <p className="text-[11px] text-slate-400">Maker-Checker dual-control builder</p>
                    </div>
                  </div>
                  <ArrowRight className="size-4 text-slate-600 group-hover:text-white" />
                </div>
              </Link>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
}
