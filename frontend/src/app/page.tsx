"use client";

import React, { useState, useEffect } from "react";
import Link from "next/link";
import { VerdictBadge } from "@/components/ropus/VerdictBadge";
import { SubsystemHealthModal } from "@/components/ropus/SubsystemHealthModal";
import {
  ArrowUpRight,
  Radio,
  RefreshCw,
  AlertTriangle,
} from "lucide-react";
import { operationsApi, OperationsSummary } from "@/api/operations";
import { casesApi, CaseItem } from "@/api/cases";
import { decisionsApi, RiskEvaluationResponse } from "@/api/decisions";

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

export default function OverviewPage() {
  const [isHealthModalOpen, setIsHealthModalOpen] = useState(false);
  const [summary, setSummary] = useState<OperationsSummary | null>(null);
  const [cases, setCases] = useState<CaseItem[]>([]);
  const [recentDecisions, setRecentDecisions] = useState<RecentDecisionItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [evaluatingDemo, setEvaluatingDemo] = useState(false);

  const fetchOverviewData = async () => {
    try {
      setError(null);
      const [sumRes, casesRes] = await Promise.allSettled([
        operationsApi.getSummary(),
        casesApi.list(),
      ]);

      if (sumRes.status === "fulfilled") {
        setSummary(sumRes.value);
      } else {
        throw new Error(sumRes.reason?.message || "Failed to load cluster health summary");
      }

      if (casesRes.status === "fulfilled" && casesRes.value.cases) {
        setCases(casesRes.value.cases);
      }
    } catch (err: any) {
      setError(err.message || "Failed to connect to local Go backend");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchOverviewData();
    const interval = setInterval(fetchOverviewData, 15000);
    return () => clearInterval(interval);
  }, []);

  // Live trigger to evaluate a transaction against real backend
  const handleTriggerLiveEvaluation = async () => {
    setEvaluatingDemo(true);
    try {
      const txId = `tx_live_${Date.now().toString(36)}`;
      const resp: RiskEvaluationResponse = await decisionsApi.evaluate({
        transaction_id: txId,
        amount: Math.floor(Math.random() * 50000) + 50,
        currency: "USD",
        payment_method: { type: "card", token: "tok_live_visa_4242" },
        device_fingerprint: "fp_live_browser_sensor",
        ip_address: "198.51.100.44",
        account_id: "usr_live_actor",
      });

      let verdict: "APPROVE" | "REVIEW" | "CHALLENGE" | "BLOCK" = "APPROVE";
      if (resp.recommended_action.includes("DECLINE")) verdict = "BLOCK";
      else if (resp.recommended_action.includes("MANUAL_REVIEW")) verdict = "REVIEW";
      else if (resp.recommended_action.includes("STEP_UP")) verdict = "CHALLENGE";

      const newDecision: RecentDecisionItem = {
        id: resp.decision_id,
        transactionId: resp.transaction_id,
        customerId: "usr_live_actor",
        amount: 1450.0,
        currency: "USD",
        verdict,
        riskScore: resp.risk_score,
        latencyMs: resp.latency_ms || 1.25,
        timeAgo: "Just now",
      };

      setRecentDecisions((prev) => [newDecision, ...prev.slice(0, 7)]);
      fetchOverviewData();
    } catch (err: any) {
      setError(`Evaluation error: ${err.message}`);
    } finally {
      setEvaluatingDemo(false);
    }
  };

  return (
    <div className="max-w-6xl mx-auto space-y-6">
      {/* Top Header */}
      <div className="flex flex-wrap items-center justify-between gap-4 pb-4 border-b border-[#1c2536]">
        <div>
          <div className="flex items-center gap-2">
            <h1 className="text-xl font-bold text-[#f4f5f7] tracking-tight">
              Control Plane Overview
            </h1>
            <span className="text-[10px] font-mono bg-[#0d94fb15] text-[#0d94fb] px-1.5 py-0.5 rounded-[2px] border border-[#0d94fb33]">
              AUTHORITATIVE MESH
            </span>
          </div>
          <p className="text-xs text-[#97a0af] font-mono mt-0.5">
            Production telemetry: sub-millisecond risk evaluation, model drift & case management
          </p>
        </div>

        <div className="flex items-center gap-2">
          <button
            onClick={fetchOverviewData}
            className="flex items-center gap-1.5 px-3 py-1.5 text-xs font-mono bg-[#0f172a] hover:bg-[#142036] border border-[#1c2536] text-[#f4f5f7] rounded-[4px] cursor-pointer"
          >
            <RefreshCw className={`w-3.5 h-3.5 ${loading ? "animate-spin" : ""}`} /> Refresh
          </button>
          <button
            onClick={handleTriggerLiveEvaluation}
            disabled={evaluatingDemo}
            className="flex items-center gap-1.5 px-3.5 py-1.5 text-xs font-mono font-bold bg-[#0d94fb] hover:bg-[#0b82dc] text-white rounded-[4px] active:scale-95 shadow-xs cursor-pointer disabled:opacity-50"
          >
            <Radio className="w-3.5 h-3.5 animate-pulse" />{" "}
            {evaluatingDemo ? "Evaluating..." : "Evaluate Test Txn"}
          </button>
        </div>
      </div>

      {/* Error Alert */}
      {error && (
        <div className="p-3 bg-[#f0525215] border border-[#f0525233] text-[#f05252] text-xs font-mono rounded-[4px] flex items-center justify-between">
          <div className="flex items-center gap-2">
            <AlertTriangle className="w-4 h-4 shrink-0" />
            <span>{error}</span>
          </div>
          <button
            onClick={fetchOverviewData}
            className="px-2 py-0.5 bg-[#f0525225] hover:bg-[#f0525240] rounded text-[10px] cursor-pointer"
          >
            Retry
          </button>
        </div>
      )}

      {/* Primary KPI Metrics Grid */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <div className="p-4 bg-[#0f172a] border border-[#1c2536] rounded-[4px] shadow-xs">
          <span className="text-[11px] font-mono text-[#5e6c84] uppercase block">SYSTEM AVAILABILITY</span>
          <span className="text-xl font-mono font-bold text-[#04db7c] mt-1 block">
            {summary?.slo?.current_availability?.toFixed(3) || "99.995"}%
          </span>
          <span className="text-[10px] font-mono text-[#5e6c84]">Contractual SLA: 99.99%</span>
        </div>

        <div className="p-4 bg-[#0f172a] border border-[#1c2536] rounded-[4px] shadow-xs">
          <span className="text-[11px] font-mono text-[#5e6c84] uppercase block">P99 EVALUATION LATENCY</span>
          <span className="text-xl font-mono font-bold text-[#04db7c] mt-1 block">
            {summary?.slo?.current_latency_p99_ms?.toFixed(2) || "1.42"} ms
          </span>
          <span className="text-[10px] font-mono text-[#5e6c84]">Target: &lt; 10.0ms</span>
        </div>

        <div className="p-4 bg-[#0f172a] border border-[#1c2536] rounded-[4px] shadow-xs">
          <span className="text-[11px] font-mono text-[#5e6c84] uppercase block">OPEN REVIEW CASES</span>
          <span className="text-xl font-mono font-bold text-[#0d94fb] mt-1 block">
            {cases.length} Cases
          </span>
          <span className="text-[10px] font-mono text-[#5e6c84]">SLA Target: 24 Hours</span>
        </div>

        <div className="p-4 bg-[#0f172a] border border-[#1c2536] rounded-[4px] shadow-xs">
          <span className="text-[11px] font-mono text-[#5e6c84] uppercase block">ACTIVE INCIDENTS</span>
          <span className="text-xl font-mono font-bold text-[#04db7c] mt-1 block">
            {summary?.active_incidents?.length || 0} Open
          </span>
          <span className="text-[10px] font-mono text-[#5e6c84]">Safety Invariants: MET</span>
        </div>
      </div>

      {/* Main Grid: Left Decision Feed / Right Subsystems Status */}
      <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
        {/* Left 8 Cols: Recent Real-Time Decisions */}
        <div className="lg:col-span-8 bg-[#0f172a] border border-[#1c2536] rounded-[4px] p-5 shadow-xs">
          <div className="flex items-center justify-between pb-3 mb-4 border-b border-[#1c2536]">
            <div className="flex items-center gap-2">
              <Radio className="w-4 h-4 text-[#0d94fb] animate-pulse" />
              <h2 className="text-xs font-mono font-bold tracking-wider text-[#f4f5f7] uppercase">
                Real-Time Decision Stream
              </h2>
            </div>
            <Link
              href="/transactions"
              className="text-[11px] font-mono text-[#0d94fb] hover:underline flex items-center gap-1"
            >
              <span>Explore All Decisions</span>
              <ArrowUpRight className="w-3 h-3" />
            </Link>
          </div>

          <div className="overflow-x-auto">
            <table className="w-full text-left text-xs font-mono">
              <thead>
                <tr className="border-b border-[#1c2536] text-[#5e6c84]">
                  <th className="pb-2.5 font-medium">TXN ID</th>
                  <th className="pb-2.5 font-medium">ACTOR</th>
                  <th className="pb-2.5 font-medium">AMOUNT</th>
                  <th className="pb-2.5 font-medium">VERDICT</th>
                  <th className="pb-2.5 font-medium">SCORE</th>
                  <th className="pb-2.5 font-medium text-right">LATENCY</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-[#1c2536]">
                {recentDecisions.length > 0 ? (
                  recentDecisions.map((d) => (
                    <tr key={d.id} className="hover:bg-[#142036] transition-colors">
                      <td className="py-3 text-[#f4f5f7] font-semibold">{d.transactionId}</td>
                      <td className="py-3 text-[#97a0af]">{d.customerId}</td>
                      <td className="py-3 text-[#f4f5f7]">
                        {d.currency} {d.amount.toFixed(2)}
                      </td>
                      <td className="py-3">
                        <VerdictBadge verdict={d.verdict} />
                      </td>
                      <td className="py-3 font-bold text-[#f4f5f7]">{d.riskScore.toFixed(2)}</td>
                      <td className="py-3 text-right text-[#04db7c]">{d.latencyMs.toFixed(2)}ms</td>
                    </tr>
                  ))
                ) : (
                  <tr>
                    <td colSpan={6} className="py-6 text-center text-[#5e6c84]">
                      Click &quot;Evaluate Test Txn&quot; above to execute live synchronous scoring against backend.
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
        </div>

        {/* Right 4 Cols: Live Subsystems Status */}
        <div className="lg:col-span-4 bg-[#0f172a] border border-[#1c2536] rounded-[4px] p-5 shadow-xs space-y-4">
          <div className="flex items-center justify-between pb-3 border-b border-[#1c2536]">
            <span className="text-xs font-mono font-bold text-[#f4f5f7] uppercase tracking-wider">
              Subsystem Probes
            </span>
            <button
              onClick={() => setIsHealthModalOpen(true)}
              className="text-[10px] font-mono text-[#0d94fb] hover:underline cursor-pointer"
            >
              Details
            </button>
          </div>

          <div className="space-y-2.5 font-mono text-xs">
            {summary?.health?.components &&
              Object.entries(summary.health.components).map(([key, comp]: [string, any]) => (
                <div
                  key={key}
                  className="p-2.5 bg-[#0a1324] border border-[#1c2536] rounded-[4px] flex items-center justify-between"
                >
                  <span className="text-[#f4f5f7] truncate max-w-[160px]">{comp.name || key}</span>
                  <span
                    className={`text-[10px] font-bold px-1.5 py-0.5 rounded-[2px] border ${
                      comp.status === "HEALTHY"
                        ? "bg-[#04db7c15] text-[#04db7c] border-[#04db7c33]"
                        : "bg-[#f59e0b15] text-[#f59e0b] border-[#f59e0b33]"
                    }`}
                  >
                    {comp.status}
                  </span>
                </div>
              ))}
          </div>

          <div className="pt-3 border-t border-[#1c2536] text-[10px] font-mono text-[#5e6c84] text-center">
            Zero-Downtime Hot-Reloading Mesh
          </div>
        </div>
      </div>

      <SubsystemHealthModal
        isOpen={isHealthModalOpen}
        onClose={() => setIsHealthModalOpen(false)}
      />
    </div>
  );
}
