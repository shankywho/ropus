"use client";

import React, { useState } from "react";
import { Activity, Search, RefreshCw } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { RiskExplanationCard } from "@/components/explanation/RiskExplanationCard";

interface TransactionItem {
  id: string;
  userId: string;
  amount: string;
  merchant: string;
  location: string;
  score: number;
  decision: "APPROVE" | "REVIEW" | "CHALLENGE" | "BLOCK";
  reasons: string[];
  explanation: string;
  time: string;
  breakdown: {
    graph_intelligence_weight: number;
    behavior_analysis_weight: number;
    threat_intelligence_weight: number;
    machine_learning_weight: number;
  };
  graphSignals?: string[];
}

const mockTransactions: TransactionItem[] = [
  {
    id: "tx_9981_mule_burst",
    userId: "usr_synthetic_bot_01",
    amount: "$14,500.00",
    merchant: "CryptoLiquidityExpress",
    location: "Limassol, CY",
    score: 94,
    decision: "BLOCK",
    reasons: [
      "Device telemetry indicates VPN and emulator environment",
      "Fraud knowledge graph detected connection to active syndicate ring",
      "High risk geolocation observed: CY",
    ],
    explanation: "Transaction blocked due to critical risk score (94.0%). High correlation with malicious fraud cluster and emulator spoofing.",
    time: "Just now",
    breakdown: {
      threat_intelligence_weight: 0.95,
      graph_intelligence_weight: 0.92,
      behavior_analysis_weight: 0.85,
      machine_learning_weight: 0.88,
    },
    graphSignals: ["Entity linked to known transnational carding cluster (degree: 14)"],
  },
  {
    id: "tx_4421_ato_wire",
    userId: "usr_vip_enterprise_88",
    amount: "$4,200.00",
    merchant: "GlobalWireTransfer",
    location: "Lagos, NG",
    score: 68,
    decision: "CHALLENGE",
    reasons: [
      "Impossible travel velocity anomaly triggered (> 5,000 miles/hr velocity)",
      "High risk geolocation observed: NG",
    ],
    explanation: "Transaction challenged with risk score (68.0%) due to elevated amount ($4,200.00) and geographic anomaly.",
    time: "2 mins ago",
    breakdown: {
      threat_intelligence_weight: 0.70,
      graph_intelligence_weight: 0.35,
      behavior_analysis_weight: 0.80,
      machine_learning_weight: 0.65,
    },
  },
  {
    id: "tx_1001_clean_retail",
    userId: "usr_john_doe",
    amount: "$125.50",
    merchant: "AmazonRetail",
    location: "Seattle, US",
    score: 12,
    decision: "APPROVE",
    reasons: [],
    explanation: "Transaction approved with low risk score (12.0%). No anomalous behavior detected.",
    time: "5 mins ago",
    breakdown: {
      threat_intelligence_weight: 0.10,
      graph_intelligence_weight: 0.05,
      behavior_analysis_weight: 0.15,
      machine_learning_weight: 0.12,
    },
  },
];

export default function TransactionsExplorerPage() {
  const [selectedTx, setSelectedTx] = useState<TransactionItem>(mockTransactions[0]);
  const [searchTerm, setSearchTerm] = useState("");

  const filtered = mockTransactions.filter(
    (t) =>
      t.id.toLowerCase().includes(searchTerm.toLowerCase()) ||
      t.userId.toLowerCase().includes(searchTerm.toLowerCase()) ||
      t.merchant.toLowerCase().includes(searchTerm.toLowerCase())
  );

  return (
    <div className="flex-1 p-6 md:p-8 space-y-6 max-w-7xl mx-auto">
      {/* Header */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 border-b border-slate-800 pb-5">
        <div>
          <h1 className="text-2xl font-bold text-white tracking-tight flex items-center gap-2.5">
            <Activity className="size-6 text-indigo-400" />
            <span>Transaction Risk Explorer</span>
          </h1>
          <p className="text-sm text-slate-400 mt-1">
            Real-time multi-engine transaction stream, factor attributions, and AI explainability.
          </p>
        </div>
        <div className="flex items-center gap-3">
          <Button variant="outline" className="border-slate-800 bg-slate-900 text-slate-200 text-xs">
            <RefreshCw className="size-3.5 mr-1.5" /> Live Streaming (Sub-10ms)
          </Button>
        </div>
      </div>

      {/* Main Grid: Stream on Left, Detail / Explanation on Right */}
      <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
        {/* Left Column: Transaction Feed */}
        <div className="lg:col-span-5 space-y-4">
          <div className="relative">
            <Search className="absolute left-3 top-2.5 size-4 text-slate-500" />
            <input
              type="text"
              placeholder="Filter by transaction ID, user, merchant..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="w-full pl-9 pr-4 py-2 bg-slate-900 border border-slate-800 rounded-lg text-xs text-slate-200 focus:outline-none focus:border-indigo-500"
            />
          </div>

          <div className="space-y-2">
            {filtered.map((tx) => (
              <div
                key={tx.id}
                onClick={() => setSelectedTx(tx)}
                className={`p-4 rounded-xl border transition-all cursor-pointer ${
                  selectedTx.id === tx.id
                    ? "bg-slate-800/90 border-indigo-500/50 shadow-lg shadow-indigo-500/10"
                    : "bg-slate-900/60 border-slate-800 hover:bg-slate-800/40"
                }`}
              >
                <div className="flex items-center justify-between">
                  <span className="text-xs font-mono font-bold text-slate-200">{tx.id}</span>
                  <Badge
                    className={`text-[10px] px-2 py-0.5 font-bold ${
                      tx.decision === "APPROVE"
                        ? "bg-emerald-500/20 text-emerald-300"
                        : tx.decision === "BLOCK"
                        ? "bg-rose-500/20 text-rose-300"
                        : "bg-blue-500/20 text-blue-300"
                    }`}
                  >
                    {tx.decision} ({tx.score}%)
                  </Badge>
                </div>
                <div className="flex items-center justify-between text-xs text-slate-400 mt-2">
                  <span>{tx.merchant}</span>
                  <span className="font-semibold text-slate-200">{tx.amount}</span>
                </div>
                <div className="flex items-center justify-between text-[11px] text-slate-500 font-mono mt-1">
                  <span>{tx.location}</span>
                  <span>{tx.time}</span>
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Right Column: AI Explanation Experience */}
        <div className="lg:col-span-7">
          <RiskExplanationCard
            transactionId={selectedTx.id}
            amount={selectedTx.amount}
            riskScore={selectedTx.score}
            decision={selectedTx.decision}
            reasons={selectedTx.reasons}
            humanExplanation={selectedTx.explanation}
            breakdown={selectedTx.breakdown}
            graphSignals={selectedTx.graphSignals}
          />
        </div>
      </div>
    </div>
  );
}
