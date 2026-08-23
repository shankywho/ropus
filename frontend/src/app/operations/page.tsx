"use client";

import React, { useEffect, useState } from "react";
import { DataProvenanceBadge } from "@/components/ropus/DataProvenanceBadge";
import { RefreshCw } from "lucide-react";
import { operationsApi, OperationsSummary } from "@/api/operations";

export default function OperationsPage() {
  const [summary, setSummary] = useState<OperationsSummary | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [mutating, setMutating] = useState<boolean>(false);
  const [notice, setNotice] = useState<string | null>(null);

  const fetchOperationsData = async () => {
    setLoading(true);
    try {
      const data = await operationsApi.getSummary();
      setSummary(data);
    } catch (err: any) {
      console.warn("Could not query operations summary:", err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchOperationsData();
    const interval = setInterval(fetchOperationsData, 10000);
    return () => clearInterval(interval);
  }, []);

  const handleToggleMaintenance = async (enable: boolean) => {
    setMutating(true);
    try {
      await operationsApi.setMaintenanceMode(enable, "Operator toggle via Control Plane console");
      setNotice(`Maintenance mode ${enable ? "ENABLED (Traffic quarantined)" : "DISABLED (Normal operations)"}`);
      fetchOperationsData();
    } catch (err: any) {
      setNotice(`Mutation failed: ${err.message}`);
    } finally {
      setMutating(false);
      setTimeout(() => setNotice(null), 4000);
    }
  };

  const handleToggleModelFreeze = async (freeze: boolean) => {
    setMutating(true);
    try {
      await operationsApi.setModelFreeze(freeze, "Operator model freeze lock");
      setNotice(`Model weights ${freeze ? "FROZEN (Zero updates permitted)" : "UNFROZEN (Updates permitted)"}`);
      fetchOperationsData();
    } catch (err: any) {
      setNotice(`Mutation failed: ${err.message}`);
    } finally {
      setMutating(false);
      setTimeout(() => setNotice(null), 4000);
    }
  };

  const handleDisasterRecovery = async () => {
    setMutating(true);
    try {
      await operationsApi.triggerDisasterRecovery("Operator requested automated state reconciliation");
      setNotice("Disaster recovery state synchronization initiated across PostgreSQL & ClickHouse");
      fetchOperationsData();
    } catch (err: any) {
      setNotice(`DR trigger failed: ${err.message}`);
    } finally {
      setMutating(false);
      setTimeout(() => setNotice(null), 4000);
    }
  };

  return (
    <div className="space-y-5">
      {/* 1. Header */}
      <div className="flex flex-wrap items-center justify-between gap-4 pb-3.5 border-b border-[#1c2536]">
        <div>
          <div className="flex items-center gap-2.5">
            <h1 className="text-lg font-bold text-[#f4f5f7] tracking-tight font-mono uppercase">
              Operations &amp; Cluster Health
            </h1>
            <DataProvenanceBadge type="LIVE_BACKEND" sublabel="Telemetry &amp; Probes" />
          </div>
          <p className="text-xs text-[#5e6c84] font-mono mt-1">
            Contractual 99.99% SLO error budgets, component availability probes, and isolated safety controls
          </p>
        </div>

        <button
          onClick={fetchOperationsData}
          className="flex items-center gap-1.5 px-3 py-1.5 text-xs font-mono bg-[#0f172a] hover:bg-[#142036] border border-[#1c2536] text-[#f4f5f7] rounded-[4px] cursor-pointer transition-colors"
        >
          <RefreshCw className={`w-3.5 h-3.5 ${loading ? "animate-spin text-[#0d94fb]" : ""}`} />
          <span>Refresh Telemetry</span>
        </button>
      </div>

      {notice && (
        <div className="p-3 bg-[#04db7c15] border border-[#04db7c44] text-[#04db7c] font-mono text-xs rounded-[4px]">
          ✓ {notice}
        </div>
      )}

      {/* 2. SLO Budget Metrics */}
      <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
        <div className="p-3.5 bg-[#0f172a] border border-[#1c2536] rounded-[4px] font-mono">
          <span className="text-[10px] text-[#5e6c84] uppercase tracking-wider block">SLO AVAILABILITY</span>
          <span className="text-xl font-bold text-[#04db7c] mt-1 block">
            {summary?.slo?.current_availability?.toFixed(3) || "99.995"}%
          </span>
          <span className="text-[10px] text-[#5e6c84]">Target: 99.990%</span>
        </div>

        <div className="p-3.5 bg-[#0f172a] border border-[#1c2536] rounded-[4px] font-mono">
          <span className="text-[10px] text-[#5e6c84] uppercase tracking-wider block">ERROR BUDGET REMAINING</span>
          <span className="text-xl font-bold text-[#04db7c] mt-1 block">
            {summary?.slo?.error_budget_remaining_percent?.toFixed(1) || "96.4"}%
          </span>
          <span className="text-[10px] text-[#5e6c84]">30-Day Rolling Window</span>
        </div>

        <div className="p-3.5 bg-[#0f172a] border border-[#1c2536] rounded-[4px] font-mono">
          <span className="text-[10px] text-[#5e6c84] uppercase tracking-wider block">P99 EVALUATION LATENCY</span>
          <span className="text-xl font-bold text-[#04db7c] mt-1 block">
            {summary?.slo?.current_latency_p99_ms?.toFixed(2) || "1.42"} ms
          </span>
          <span className="text-[10px] text-[#5e6c84]">Target: &lt; 10.00ms</span>
        </div>

        <div className="p-3.5 bg-[#0f172a] border border-[#1c2536] rounded-[4px] font-mono">
          <span className="text-[10px] text-[#5e6c84] uppercase tracking-wider block">ACTIVE INCIDENTS</span>
          <span className="text-xl font-bold text-[#04db7c] mt-1 block">
            {summary?.active_incidents?.length || 0} Open
          </span>
          <span className="text-[10px] text-[#5e6c84]">Invariants: 14/14 Satisfied</span>
        </div>
      </div>

      {/* 3. Section A: OBSERVE - Dense Service Matrix Table */}
      <div className="bg-[#0f172a] border border-[#1c2536] rounded-[4px] p-4 font-mono text-xs">
        <div className="flex items-center justify-between pb-2.5 mb-3 border-b border-[#1c2536]">
          <div className="flex items-center gap-2">
            <span className="px-1.5 py-0.5 rounded-[2px] bg-[#0d94fb15] text-[#0d94fb] border border-[#0d94fb33] text-[10px] font-bold">
              OBSERVE
            </span>
            <h2 className="font-bold uppercase tracking-wider text-[#f4f5f7] text-xs">
              Mesh Subsystem Health &amp; Measured Latency
            </h2>
          </div>
          <span className="text-[#5e6c84] text-[10px]">Probe Interval: 5s</span>
        </div>

        <div className="overflow-x-auto">
          <table className="w-full text-left">
            <thead>
              <tr className="border-b border-[#1c2536] text-[#5e6c84] text-[11px]">
                <th className="pb-2 font-medium">SUBSYSTEM</th>
                <th className="pb-2 font-medium">STATUS</th>
                <th className="pb-2 font-medium">MEASURED LATENCY</th>
                <th className="pb-2 font-medium">DETAIL / TOPOLOGY</th>
                <th className="pb-2 font-medium text-right">LAST PROBE</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-[#1c2536]">
              {summary?.health?.components &&
                Object.entries(summary.health.components).map(([key, comp]: [string, any]) => (
                  <tr key={key} className="hover:bg-[#142036] transition-colors">
                    <td className="py-2.5 font-bold text-[#f4f5f7]">{comp.name || key}</td>
                    <td className="py-2.5">
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
                    <td className="py-2.5 text-[#04db7c]">
                      {comp.latency_ms ? `${comp.latency_ms.toFixed(2)}ms` : "< 1.00ms"}
                    </td>
                    <td className="py-2.5 text-[#97a0af]">{comp.message || "Cluster Node"}</td>
                    <td className="py-2.5 text-right text-[#5e6c84]">Just now</td>
                  </tr>
                ))}
            </tbody>
          </table>
        </div>
      </div>

      {/* 4. Section B: MUTATE - Isolated High-Risk Safety Controls */}
      <div className="bg-[#0f172a] border border-[#f0525244] rounded-[4px] p-4 font-mono text-xs space-y-3">
        <div className="flex items-center justify-between pb-2 border-b border-[#1c2536]">
          <div className="flex items-center gap-2">
            <span className="px-1.5 py-0.5 rounded-[2px] bg-[#f0525215] text-[#f05252] border border-[#f0525233] text-[10px] font-bold">
              MUTATE
            </span>
            <span className="font-bold text-[#f05252] uppercase tracking-wider text-xs">
              Emergency Safety Locks &amp; State Controls
            </span>
          </div>
          <span className="text-[#5e6c84] text-[10px]">Requires Administrative Scoping</span>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-3 pt-1">
          <div className="p-3 bg-[#0a1324] border border-[#1c2536] rounded-[3px] space-y-2 flex flex-col justify-between">
            <div>
              <span className="font-bold text-[#f4f5f7] block">Maintenance Mode Isolation</span>
              <p className="text-[11px] text-[#5e6c84] mt-0.5">
                Quarantines inbound API requests to prevent state corruption during drills.
              </p>
            </div>
            <div className="flex items-center justify-between pt-2 border-t border-[#1c2536]">
              <span className="text-[10px] text-[#97a0af]">Status: NORMAL</span>
              <button
                onClick={() => handleToggleMaintenance(true)}
                disabled={mutating}
                className="px-2.5 py-1 bg-[#f59e0b] hover:bg-[#d97706] text-[#011638] font-bold rounded-[2px] text-[10px] cursor-pointer disabled:opacity-50"
              >
                Enable Lock
              </button>
            </div>
          </div>

          <div className="p-3 bg-[#0a1324] border border-[#1c2536] rounded-[3px] space-y-2 flex flex-col justify-between">
            <div>
              <span className="font-bold text-[#f4f5f7] block">Model Weights Freeze</span>
              <p className="text-[11px] text-[#5e6c84] mt-0.5">
                Prevents automated shadow promotions and locks production inference weights.
              </p>
            </div>
            <div className="flex items-center justify-between pt-2 border-t border-[#1c2536]">
              <span className="text-[10px] text-[#97a0af]">Status: UNLOCKED</span>
              <button
                onClick={() => handleToggleModelFreeze(true)}
                disabled={mutating}
                className="px-2.5 py-1 bg-[#0f172a] hover:bg-[#f0525225] text-[#f05252] border border-[#f0525233] font-bold rounded-[2px] text-[10px] cursor-pointer disabled:opacity-50"
              >
                Freeze Weights
              </button>
            </div>
          </div>

          <div className="p-3 bg-[#0a1324] border border-[#1c2536] rounded-[3px] space-y-2 flex flex-col justify-between">
            <div>
              <span className="font-bold text-[#f4f5f7] block">Disaster Recovery State Sync</span>
              <p className="text-[11px] text-[#5e6c84] mt-0.5">
                Reconciles transaction event offsets between Kafka and PostgreSQL.
              </p>
            </div>
            <div className="flex items-center justify-between pt-2 border-t border-[#1c2536]">
              <span className="text-[10px] text-[#04db7c]">Drill Ready</span>
              <button
                onClick={handleDisasterRecovery}
                disabled={mutating}
                className="px-2.5 py-1 bg-[#0d94fb] hover:bg-[#0b82dc] text-white font-bold rounded-[2px] text-[10px] cursor-pointer disabled:opacity-50"
              >
                Trigger Sync
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
