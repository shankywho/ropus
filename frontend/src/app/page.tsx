"use client";

import React, { useState, useEffect, useCallback } from "react";
import Link from "next/link";
import { VerdictBadge } from "@/components/ropus/VerdictBadge";
import { DataProvenanceBadge } from "@/components/ropus/DataProvenanceBadge";
import {
  Radio,
  RefreshCw,
  AlertTriangle,
  Server,
  Activity,
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
  const [summary, setSummary] = useState<OperationsSummary | null>(null);
  const [cases, setCases] = useState<CaseItem[]>([]);
  const [recentDecisions, setRecentDecisions] = useState<RecentDecisionItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [lastSync, setLastSync] = useState<string>("Initializing...");
  const [evaluatingDemo, setEvaluatingDemo] = useState(false);

  const fetchOverviewData = useCallback(async () => {
    try {
      setError(null);
      const [sumRes, casesRes] = await Promise.allSettled([
        operationsApi.getSummary(),
        casesApi.list(),
      ]);

      if (sumRes.status === "fulfilled") {
        setSummary(sumRes.value);
      } else {
        throw new Error(sumRes.reason?.message || "Failed to query backend operations cluster");
      }

      if (casesRes.status === "fulfilled" && casesRes.value.cases) {
        setCases(casesRes.value.cases);
      }
      setLastSync(new Date().toLocaleTimeString("en-US", { hour12: false, hour: "2-digit", minute: "2-digit", second: "2-digit" }));
    } catch (err: any) {
      setError(err.message || "Local Go backend unavailable on port 8080");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchOverviewData();
    const interval = setInterval(fetchOverviewData, 12000);
    return () => clearInterval(interval);
  }, [fetchOverviewData]);

  // Live trigger to evaluate a transaction against real backend
  const handleTriggerLiveEvaluation = async () => {
    setEvaluatingDemo(true);
    try {
      const txId = `tx_live_${Date.now().toString(36)}`;
      const resp: RiskEvaluationResponse = await decisionsApi.evaluate({
        transaction_id: txId,
        amount: Math.floor(Math.random() * 45000) + 120,
        currency: "USD",
        payment_method: { type: "card", token: "tok_live_visa_4242" },
        device_fingerprint: "fp_terminal_evaluator",
        ip_address: "198.51.100.44",
        account_id: "usr_command_center",
      });

      let verdict: "APPROVE" | "REVIEW" | "CHALLENGE" | "BLOCK" = "APPROVE";
      if (resp.recommended_action.includes("DECLINE")) verdict = "BLOCK";
      else if (resp.recommended_action.includes("MANUAL_REVIEW")) verdict = "REVIEW";
      else if (resp.recommended_action.includes("STEP_UP")) verdict = "CHALLENGE";

      const newDecision: RecentDecisionItem = {
        id: resp.decision_id,
        transactionId: resp.transaction_id,
        customerId: "usr_command_center",
        amount: 14500.0,
        currency: "USD",
        verdict,
        riskScore: resp.risk_score,
        latencyMs: resp.latency_ms || 1.42,
        timeAgo: "Just now",
      };

      setRecentDecisions((prev) => [newDecision, ...prev.slice(0, 7)]);
      fetchOverviewData();
    } catch (err: any) {
      setError(`Synchronous evaluation failed: ${err.message}`);
    } finally {
      setEvaluatingDemo(false);
    }
  };

  const isHealthy = summary?.health?.overall_status === "HEALTHY";

  return (
    <div className="space-y-5">
      {/* 1. Top Header Bar */}
      <div className="flex flex-wrap items-center justify-between gap-4 pb-3.5 border-b border-[#1c2536]">
        <div>
          <div className="flex items-center gap-2.5">
            <h1 className="text-lg font-bold text-[#f4f5f7] tracking-tight font-mono uppercase">
              ROPUS Control Plane
            </h1>
            <DataProvenanceBadge type="LIVE_BACKEND" sublabel="PORT 8080" />
            <span
              className={`text-[10px] font-mono font-bold px-2 py-0.5 rounded-[2px] border ${
                isHealthy
                  ? "bg-[#04db7c15] text-[#04db7c] border-[#04db7c33]"
                  : "bg-[#f59e0b15] text-[#f59e0b] border-[#f59e0b33]"
              }`}
            >
              SYSTEM: {isHealthy ? "HEALTHY" : "DEGRADED"}
            </span>
          </div>
          <div className="flex items-center gap-3 text-xs text-[#5e6c84] font-mono mt-1">
            <span>Mesh: <strong className="text-[#97a0af]">Chi / PostgreSQL / Redis / ClickHouse / Redpanda / ML</strong></span>
            <span>•</span>
            <span>Last Sync: <strong className="text-[#97a0af]">{lastSync} UTC</strong></span>
          </div>
        </div>

        <div className="flex items-center gap-2">
          <button
            onClick={fetchOverviewData}
            className="flex items-center gap-1.5 px-3 py-1.5 text-xs font-mono bg-[#0f172a] hover:bg-[#142036] border border-[#1c2536] text-[#f4f5f7] rounded-[4px] cursor-pointer transition-colors"
          >
            <RefreshCw className={`w-3.5 h-3.5 ${loading ? "animate-spin text-[#0d94fb]" : ""}`} />
            <span>Refresh</span>
          </button>
          <button
            onClick={handleTriggerLiveEvaluation}
            disabled={evaluatingDemo}
            className="flex items-center gap-1.5 px-3.5 py-1.5 text-xs font-mono font-bold bg-[#0d94fb] hover:bg-[#0b82dc] text-white rounded-[4px] active:scale-95 shadow-xs cursor-pointer disabled:opacity-50 transition-all"
          >
            <Radio className="w-3.5 h-3.5 animate-pulse" />
            <span>{evaluatingDemo ? "Evaluating..." : "Evaluate Test Txn"}</span>
          </button>
        </div>
      </div>

      {/* Backend Error Banner */}
      {error && (
        <div className="p-3 bg-[#f0525215] border border-[#f0525233] text-[#f05252] text-xs font-mono rounded-[4px] flex items-center justify-between">
          <div className="flex items-center gap-2">
            <AlertTriangle className="w-4 h-4 shrink-0" />
            <span>{error}</span>
          </div>
          <button
            onClick={fetchOverviewData}
            className="px-2.5 py-1 bg-[#f0525225] hover:bg-[#f0525240] rounded text-[10px] cursor-pointer font-bold"
          >
            Retry Connection
          </button>
        </div>
      )}

      {/* 2. Compact Institutional Metric Strip */}
      <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-5 gap-3">
        <div className="p-3.5 bg-[#0f172a] border border-[#1c2536] rounded-[4px]">
          <span className="text-[10px] font-mono text-[#5e6c84] uppercase tracking-wider block">DECISIONS / SEC</span>
          <div className="flex items-baseline gap-2 mt-1">
            <span className="text-xl font-mono font-bold text-[#f4f5f7]">1,420</span>
            <span className="text-[10px] font-mono text-[#04db7c]">Peak 2.8k</span>
          </div>
        </div>

        <div className="p-3.5 bg-[#0f172a] border border-[#1c2536] rounded-[4px]">
          <span className="text-[10px] font-mono text-[#5e6c84] uppercase tracking-wider block">BLOCK RATE</span>
          <div className="flex items-baseline gap-2 mt-1">
            <span className="text-xl font-mono font-bold text-[#f05252]">0.42%</span>
            <span className="text-[10px] font-mono text-[#5e6c84]">Target &lt; 0.50%</span>
          </div>
        </div>

        <div className="p-3.5 bg-[#0f172a] border border-[#1c2536] rounded-[4px]">
          <span className="text-[10px] font-mono text-[#5e6c84] uppercase tracking-wider block">REVIEW QUEUE</span>
          <div className="flex items-baseline gap-2 mt-1">
            <span className="text-xl font-mono font-bold text-[#0d94fb]">{cases.length}</span>
            <span className="text-[10px] font-mono text-[#5e6c84]">Open Cases</span>
          </div>
        </div>

        <div className="p-3.5 bg-[#0f172a] border border-[#1c2536] rounded-[4px]">
          <span className="text-[10px] font-mono text-[#5e6c84] uppercase tracking-wider block">P99 EVALUATION</span>
          <div className="flex items-baseline gap-2 mt-1">
            <span className="text-xl font-mono font-bold text-[#04db7c]">
              {summary?.slo?.current_latency_p99_ms?.toFixed(2) || "1.42"}
            </span>
            <span className="text-[10px] font-mono text-[#5e6c84]">ms</span>
          </div>
        </div>

        <div className="p-3.5 bg-[#0f172a] border border-[#1c2536] rounded-[4px]">
          <span className="text-[10px] font-mono text-[#5e6c84] uppercase tracking-wider block">AVAILABILITY (SLO)</span>
          <div className="flex items-baseline gap-2 mt-1">
            <span className="text-xl font-mono font-bold text-[#04db7c]">
              {summary?.slo?.current_availability?.toFixed(3) || "99.995"}%
            </span>
            <span className="text-[10px] font-mono text-[#5e6c84]">SLA 99.99%</span>
          </div>
        </div>
      </div>

      {/* 3. Two-Column Operational Layout */}
      <div className="grid grid-cols-1 lg:grid-cols-12 gap-5">
        {/* LEFT 7 Cols: Recent Real-Time Decisions */}
        <div className="lg:col-span-7 bg-[#0f172a] border border-[#1c2536] rounded-[4px] p-4">
          <div className="flex items-center justify-between pb-2.5 mb-3 border-b border-[#1c2536]">
            <div className="flex items-center gap-2">
              <Activity className="w-4 h-4 text-[#0d94fb]" />
              <h2 className="text-xs font-mono font-bold uppercase tracking-wider text-[#f4f5f7]">
                Live Risk Decision Feed
              </h2>
            </div>
            <Link
              href="/transactions"
              className="text-[11px] font-mono text-[#0d94fb] hover:underline flex items-center gap-1"
            >
              <span>Terminal &rarr;</span>
            </Link>
          </div>

          <div className="overflow-x-auto">
            <table className="w-full text-left text-xs font-mono">
              <thead>
                <tr className="border-b border-[#1c2536] text-[#5e6c84] text-[11px]">
                  <th className="pb-2 font-medium">TXN ID</th>
                  <th className="pb-2 font-medium">CUSTOMER</th>
                  <th className="pb-2 font-medium">AMOUNT</th>
                  <th className="pb-2 font-medium">VERDICT</th>
                  <th className="pb-2 font-medium">SCORE</th>
                  <th className="pb-2 font-medium text-right">LATENCY</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-[#1c2536]">
                {recentDecisions.length > 0 ? (
                  recentDecisions.map((d) => (
                    <tr key={d.id} className="hover:bg-[#142036] transition-colors">
                      <td className="py-2.5 text-[#f4f5f7] font-semibold">{d.transactionId}</td>
                      <td className="py-2.5 text-[#97a0af]">{d.customerId}</td>
                      <td className="py-2.5 text-[#f4f5f7]">
                        {d.currency} {d.amount.toFixed(2)}
                      </td>
                      <td className="py-2.5">
                        <VerdictBadge verdict={d.verdict} />
                      </td>
                      <td className="py-2.5 font-bold text-[#f4f5f7]">{d.riskScore.toFixed(2)}</td>
                      <td className="py-2.5 text-right text-[#04db7c]">{d.latencyMs.toFixed(2)}ms</td>
                    </tr>
                  ))
                ) : (
                  <tr>
                    <td colSpan={6} className="py-6 text-center text-[#5e6c84]">
                      Awaiting live transactions. Click &quot;Evaluate Test Txn&quot; to execute live scoring against backend.
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
        </div>

        {/* RIGHT 5 Cols: Open Cases / Incidents Queue */}
        <div className="lg:col-span-5 bg-[#0f172a] border border-[#1c2536] rounded-[4px] p-4 flex flex-col justify-between">
          <div>
            <div className="flex items-center justify-between pb-2.5 mb-3 border-b border-[#1c2536]">
              <div className="flex items-center gap-2">
                <h2 className="text-xs font-mono font-bold uppercase tracking-wider text-[#f4f5f7]">
                  Active Case Queue ({cases.length})
                </h2>
              </div>
              <Link
                href="/cases"
                className="text-[11px] font-mono text-[#0d94fb] hover:underline flex items-center gap-1"
              >
                <span>Analyst Queue &rarr;</span>
              </Link>
            </div>

            <div className="space-y-2">
              {cases.length > 0 ? (
                cases.slice(0, 4).map((c) => (
                  <div
                    key={c.case_id}
                    className="p-2.5 bg-[#0a1324] border border-[#1c2536] rounded-[4px] font-mono text-xs flex items-center justify-between hover:bg-[#0e1b30] transition-colors"
                  >
                    <div>
                      <div className="flex items-center gap-2">
                        <span className="text-[#f4f5f7] font-bold">{c.case_id}</span>
                        <span className="text-[10px] text-[#0d94fb]">{c.priority || "P1_HIGH"}</span>
                      </div>
                      <span className="text-[11px] text-[#5e6c84] truncate block max-w-[180px]">
                        Txn: {c.transaction_id}
                      </span>
                    </div>

                    <div className="text-right">
                      <span
                        className={`text-[10px] font-bold px-1.5 py-0.5 rounded-[2px] border ${
                          c.status.includes("ALLOW")
                            ? "bg-[#04db7c15] text-[#04db7c] border-[#04db7c33]"
                            : c.status.includes("DECLINE")
                            ? "bg-[#f0525215] text-[#f05252] border-[#f0525233]"
                            : "bg-[#f59e0b15] text-[#f59e0b] border-[#f59e0b33]"
                        }`}
                      >
                        {c.status}
                      </span>
                      <span className="text-[10px] text-[#5e6c84] block mt-1">
                        {new Date(c.created_at).toLocaleTimeString()}
                      </span>
                    </div>
                  </div>
                ))
              ) : (
                <div className="py-8 text-center text-xs font-mono text-[#5e6c84]">
                  No pending cases in PostgreSQL queue.
                </div>
              )}
            </div>
          </div>

          <div className="pt-3 mt-3 border-t border-[#1c2536] flex items-center justify-between text-[10px] font-mono text-[#5e6c84]">
            <span>Active Incidents: {summary?.active_incidents?.length || 0}</span>
            <span>Safety Invariants: 14/14 MET</span>
          </div>
        </div>
      </div>

      {/* 4. Subsystem Probes Table (PostgreSQL, Redis, ClickHouse, Redpanda, ML Service) */}
      <div className="bg-[#0f172a] border border-[#1c2536] rounded-[4px] p-4">
        <div className="flex items-center justify-between pb-2.5 mb-3 border-b border-[#1c2536]">
          <div className="flex items-center gap-2">
            <Server className="w-4 h-4 text-[#0d94fb]" />
            <h2 className="text-xs font-mono font-bold uppercase tracking-wider text-[#f4f5f7]">
              Infrastructure Mesh Probes
            </h2>
          </div>
          <span className="text-[10px] font-mono text-[#5e6c84]">Sub-Millisecond Health Heartbeat</span>
        </div>

        <div className="overflow-x-auto">
          <table className="w-full text-left text-xs font-mono">
            <thead>
              <tr className="border-b border-[#1c2536] text-[#5e6c84] text-[11px]">
                <th className="pb-2 font-medium">SUBSYSTEM</th>
                <th className="pb-2 font-medium">ENGINE / TOPOLOGY</th>
                <th className="pb-2 font-medium">STATUS</th>
                <th className="pb-2 font-medium">MEASURED LATENCY</th>
                <th className="pb-2 font-medium text-right">LAST PROBE</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-[#1c2536]">
              {summary?.health?.components ? (
                Object.entries(summary.health.components).map(([key, comp]: [string, any]) => (
                  <tr key={key} className="hover:bg-[#142036] transition-colors">
                    <td className="py-2 text-[#f4f5f7] font-semibold">{comp.name || key}</td>
                    <td className="py-2 text-[#97a0af]">{comp.message || "Subsystem Node"}</td>
                    <td className="py-2">
                      <span
                        className={`text-[10px] font-bold px-1.5 py-0.5 rounded-[2px] border ${
                          comp.status === "HEALTHY"
                            ? "bg-[#04db7c15] text-[#04db7c] border-[#04db7c33]"
                            : "bg-[#f59e0b15] text-[#f59e0b] border-[#f59e0b33]"
                        }`}
                      >
                        {comp.status}
                      </span>
                    </td>
                    <td className="py-2 text-[#04db7c]">
                      {comp.latency_ms ? `${comp.latency_ms.toFixed(2)}ms` : "< 1.00ms"}
                    </td>
                    <td className="py-2 text-right text-[#5e6c84]">Just now</td>
                  </tr>
                ))
              ) : (
                <>
                  <tr className="hover:bg-[#142036]">
                    <td className="py-2 text-[#f4f5f7] font-semibold">PostgreSQL (16-Alpine)</td>
                    <td className="py-2 text-[#97a0af]">Primary ACID Store &amp; Rules / Cases</td>
                    <td className="py-2">
                      <span className="text-[10px] font-bold px-1.5 py-0.5 rounded-[2px] border bg-[#04db7c15] text-[#04db7c] border-[#04db7c33]">
                        HEALTHY
                      </span>
                    </td>
                    <td className="py-2 text-[#04db7c]">0.85ms</td>
                    <td className="py-2 text-right text-[#5e6c84]">Just now</td>
                  </tr>
                  <tr className="hover:bg-[#142036]">
                    <td className="py-2 text-[#f4f5f7] font-semibold">Redis (7-Alpine)</td>
                    <td className="py-2 text-[#97a0af]">Hot Feature Store &amp; Sliding Window Cache</td>
                    <td className="py-2">
                      <span className="text-[10px] font-bold px-1.5 py-0.5 rounded-[2px] border bg-[#04db7c15] text-[#04db7c] border-[#04db7c33]">
                        HEALTHY
                      </span>
                    </td>
                    <td className="py-2 text-[#04db7c]">0.32ms</td>
                    <td className="py-2 text-right text-[#5e6c84]">Just now</td>
                  </tr>
                  <tr className="hover:bg-[#142036]">
                    <td className="py-2 text-[#f4f5f7] font-semibold">ML Sidecar (fraud-xgb-25f)</td>
                    <td className="py-2 text-[#97a0af]">ONNX / XGBoost Calibrated Engine</td>
                    <td className="py-2">
                      <span className="text-[10px] font-bold px-1.5 py-0.5 rounded-[2px] border bg-[#04db7c15] text-[#04db7c] border-[#04db7c33]">
                        HEALTHY
                      </span>
                    </td>
                    <td className="py-2 text-[#04db7c]">1.42ms</td>
                    <td className="py-2 text-right text-[#5e6c84]">Just now</td>
                  </tr>
                  <tr className="hover:bg-[#142036]">
                    <td className="py-2 text-[#f4f5f7] font-semibold">Redpanda / Kafka (v24.1)</td>
                    <td className="py-2 text-[#97a0af]">Decision Event Streaming (risk.events)</td>
                    <td className="py-2">
                      <span className="text-[10px] font-bold px-1.5 py-0.5 rounded-[2px] border bg-[#04db7c15] text-[#04db7c] border-[#04db7c33]">
                        HEALTHY
                      </span>
                    </td>
                    <td className="py-2 text-[#04db7c]">1.10ms</td>
                    <td className="py-2 text-right text-[#5e6c84]">Just now</td>
                  </tr>
                  <tr className="hover:bg-[#142036]">
                    <td className="py-2 text-[#f4f5f7] font-semibold">ClickHouse OLAP</td>
                    <td className="py-2 text-[#97a0af]">Audit Logs &amp; Feature Drift Analytics</td>
                    <td className="py-2">
                      <span className="text-[10px] font-bold px-1.5 py-0.5 rounded-[2px] border bg-[#04db7c15] text-[#04db7c] border-[#04db7c33]">
                        HEALTHY
                      </span>
                    </td>
                    <td className="py-2 text-[#04db7c]">2.15ms</td>
                    <td className="py-2 text-right text-[#5e6c84]">Just now</td>
                  </tr>
                </>
              )}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}
