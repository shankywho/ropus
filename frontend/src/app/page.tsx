"use client";

import React, { useState } from "react";
import Link from "next/link";
import { VerdictBadge } from "@/components/ropus/VerdictBadge";
import { SubsystemHealthModal } from "@/components/ropus/SubsystemHealthModal";
import {
  ArrowUpRight,
  Radio,
} from "lucide-react";

interface RecentDecisionItem {
  id: string;
  transactionId: string;
  customerId: string;
  amount: number;
  currency: string;
  verdict: "APPROVE" | "REVIEW" | "CHALLENGE" | "BLOCK";
  riskScore: number;
  latencyMs: number;
  timeAgo: string;
}

const SEED_DECISIONS: RecentDecisionItem[] = [
  {
    id: "dec_9f8a1e9c",
    transactionId: "tx_order_88419",
    customerId: "usr_sarah_connor",
    amount: 14500.0,
    currency: "USD",
    verdict: "BLOCK",
    riskScore: 0.96,
    latencyMs: 1.42,
    timeAgo: "1m ago",
  },
  {
    id: "dec_7a1b2c3d",
    transactionId: "tx_order_88418",
    customerId: "usr_sarah_connor",
    amount: 42.5,
    currency: "USD",
    verdict: "APPROVE",
    riskScore: 0.04,
    latencyMs: 1.15,
    timeAgo: "3m ago",
  },
  {
    id: "dec_6e5d4c3b",
    transactionId: "tx_order_88417",
    customerId: "usr_marcus_wright",
    amount: 250.0,
    currency: "USD",
    verdict: "APPROVE",
    riskScore: 0.08,
    latencyMs: 0.98,
    timeAgo: "5m ago",
  },
  {
    id: "dec_5f4e3d2c",
    transactionId: "tx_order_88416",
    customerId: "usr_john_connor",
    amount: 3200.0,
    currency: "USD",
    verdict: "CHALLENGE",
    riskScore: 0.58,
    latencyMs: 1.65,
    timeAgo: "8m ago",
  },
  {
    id: "dec_4e3d2c1b",
    transactionId: "tx_order_88415",
    customerId: "usr_kyle_reese",
    amount: 8200.0,
    currency: "USD",
    verdict: "REVIEW",
    riskScore: 0.44,
    latencyMs: 1.28,
    timeAgo: "12m ago",
  },
  {
    id: "dec_3c2b1a0f",
    transactionId: "tx_order_88414",
    customerId: "usr_grace_harper",
    amount: 18.99,
    currency: "USD",
    verdict: "APPROVE",
    riskScore: 0.02,
    latencyMs: 0.85,
    timeAgo: "15m ago",
  },
];

export default function OverviewPage() {
  const [isHealthModalOpen, setIsHealthModalOpen] = useState(false);
  const [decisions] = useState<RecentDecisionItem[]>(SEED_DECISIONS);

  return (
    <div className="max-w-6xl mx-auto space-y-6">
      {/* Top Header */}
      <div className="flex flex-wrap items-center justify-between gap-4 pb-4 border-b border-[#1c2b48]">
        <div>
          <h1 className="text-xl font-bold text-[#f4f5f7] tracking-tight">System Control Plane</h1>
          <p className="text-xs text-[#97a0af] font-mono mt-0.5">
            Synchronous evaluation telemetry, live stream ingestion & cluster health
          </p>
        </div>

        {/* Clickable System Health Badge -> Opens Breaker Inspector */}
        <button
          onClick={() => setIsHealthModalOpen(true)}
          className="flex items-center gap-2 px-3 py-1.5 bg-[#0b1528] hover:bg-[#0f1c34] border border-[#04db7c44] rounded text-xs font-mono text-[#04db7c] transition-all cursor-pointer shadow-xs active:scale-95"
        >
          <span className="w-2 h-2 rounded-full bg-[#04db7c] animate-pulse" />
          <span className="font-bold">ALL SYSTEMS NORMAL (CIRCUIT BREAKERS CLOSED)</span>
          <span className="text-[10px] text-[#97a0af] underline ml-1">Inspect</span>
        </button>
      </div>

      {/* 6 Core Metrics Grid (Clean, Calm, Anti-Clutter) */}
      <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-3">
        <div className="p-3.5 bg-[#0b1528] border border-[#1c2b48] rounded">
          <span className="text-[11px] font-mono text-[#6b778c] block">EVALUATIONS (24H)</span>
          <span className="text-lg font-mono font-bold text-[#f4f5f7] mt-1 block">1,842,910</span>
          <span className="text-[10px] font-mono text-[#04db7c]">100% SLA Met</span>
        </div>

        <div className="p-3.5 bg-[#0b1528] border border-[#1c2b48] rounded">
          <span className="text-[11px] font-mono text-[#6b778c] block">APPROVE RATE</span>
          <span className="text-lg font-mono font-bold text-[#04db7c] mt-1 block">98.42%</span>
          <span className="text-[10px] font-mono text-[#6b778c]">1,813,792 txns</span>
        </div>

        <div className="p-3.5 bg-[#0b1528] border border-[#1c2b48] rounded">
          <span className="text-[11px] font-mono text-[#6b778c] block">BLOCK / CHALLENGE</span>
          <span className="text-lg font-mono font-bold text-[#f05252] mt-1 block">1.58%</span>
          <span className="text-[10px] font-mono text-[#6b778c]">29,118 intercepted</span>
        </div>

        <div className="p-3.5 bg-[#0b1528] border border-[#1c2b48] rounded">
          <span className="text-[11px] font-mono text-[#6b778c] block">AVG RISK SCORE</span>
          <span className="text-lg font-mono font-bold text-[#38bdf8] mt-1 block">0.078</span>
          <span className="text-[10px] font-mono text-[#6b778c]">Calibrated Posterior</span>
        </div>

        <div className="p-3.5 bg-[#0b1528] border border-[#1c2b48] rounded">
          <span className="text-[11px] font-mono text-[#6b778c] block">P99 LATENCY</span>
          <span className="text-lg font-mono font-bold text-[#04db7c] mt-1 block">1.42 ms</span>
          <span className="text-[10px] font-mono text-[#6b778c]">&lt; 10.0ms Target</span>
        </div>

        <div
          onClick={() => setIsHealthModalOpen(true)}
          className="p-3.5 bg-[#0b1528] border border-[#1c2b48] hover:border-[#04db7c44] rounded cursor-pointer transition-colors"
        >
          <span className="text-[11px] font-mono text-[#6b778c] block">HEALTH STATUS</span>
          <span className="text-lg font-mono font-bold text-[#04db7c] mt-1 block">100.0%</span>
          <span className="text-[10px] font-mono text-[#0d94fb] underline">8/8 Breakers</span>
        </div>
      </div>

      {/* Live Recent Risk Decisions Feed (Kafka Egress Proof) */}
      <div className="bg-[#0b1528] border border-[#1c2b48] rounded p-5">
        <div className="flex items-center justify-between pb-3 mb-4 border-b border-[#1c2b48]">
          <div className="flex items-center gap-2">
            <Radio className="w-4 h-4 text-[#38bdf8] animate-pulse" />
            <h2 className="text-xs font-mono font-bold text-[#f4f5f7] uppercase tracking-wider">
              Live Decision Stream (Kafka Real-Time Feed)
            </h2>
          </div>
          <span className="text-[10px] font-mono text-[#6b778c]">Append-Only Telemetry</span>
        </div>

        <div className="overflow-x-auto">
          <table className="w-full text-left text-xs font-mono">
            <thead>
              <tr className="border-b border-[#1c2b48] text-[#6b778c]">
                <th className="pb-2.5 font-medium">TRANSACTION ID</th>
                <th className="pb-2.5 font-medium">CUSTOMER</th>
                <th className="pb-2.5 font-medium">AMOUNT</th>
                <th className="pb-2.5 font-medium">RISK SCORE</th>
                <th className="pb-2.5 font-medium">VERDICT</th>
                <th className="pb-2.5 font-medium">LATENCY</th>
                <th className="pb-2.5 font-medium text-right">ACTION</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-[#1c2b48]">
              {decisions.map((dec) => (
                <tr key={dec.id} className="hover:bg-[#0f1c34] transition-colors">
                  <td className="py-3 text-[#f4f5f7] font-semibold">{dec.transactionId}</td>
                  <td className="py-3 text-[#97a0af]">{dec.customerId}</td>
                  <td className="py-3 text-[#f4f5f7] font-bold">
                    ${dec.amount.toLocaleString("en-US", { minimumFractionDigits: 2 })} {dec.currency}
                  </td>
                  <td className="py-3">
                    <span
                      className={`font-bold ${
                        dec.riskScore >= 0.8
                          ? "text-[#f05252]"
                          : dec.riskScore >= 0.3
                          ? "text-[#f59e0b]"
                          : "text-[#04db7c]"
                      }`}
                    >
                      {dec.riskScore.toFixed(2)}
                    </span>
                  </td>
                  <td className="py-3">
                    <VerdictBadge verdict={dec.verdict} size="sm" />
                  </td>
                  <td className="py-3 text-[#97a0af]">{dec.latencyMs.toFixed(2)} ms</td>
                  <td className="py-3 text-right">
                    <Link
                      href="/transactions"
                      className="inline-flex items-center gap-1 text-[11px] text-[#0d94fb] hover:underline"
                    >
                      Inspect <ArrowUpRight className="w-3 h-3" />
                    </Link>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>

      {/* Subsystem Health Expander Modal */}
      <SubsystemHealthModal
        isOpen={isHealthModalOpen}
        onClose={() => setIsHealthModalOpen(false)}
      />
    </div>
  );
}
