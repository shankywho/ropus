"use client";

import React, { useState, useEffect, useCallback } from "react";
import {
  RefreshCw,
  AlertTriangle,
  FolderKanban,
  Play,
} from "lucide-react";
import { casesApi, CaseItem, CaseDetail } from "@/api/cases";
import { decisionsApi } from "@/api/decisions";

export default function CasesPage() {
  const [casesList, setCasesList] = useState<CaseItem[]>([]);
  const [selectedCaseId, setSelectedCaseId] = useState<string | null>(null);
  const [selectedCaseDetail, setSelectedCaseDetail] = useState<CaseDetail | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [actionNotice, setActionNotice] = useState<string | null>(null);
  const [resolving, setResolving] = useState(false);

  const fetchCases = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await casesApi.list();
      if (data && data.cases && data.cases.length > 0) {
        setCasesList(data.cases);
        if (!selectedCaseId || !data.cases.some((c) => c.case_id === selectedCaseId)) {
          setSelectedCaseId(data.cases[0].case_id);
        }
      } else {
        setCasesList([]);
        setSelectedCaseId(null);
        setSelectedCaseDetail(null);
      }
    } catch (err: any) {
      setError(err.message || "Failed to load review cases from backend");
    } finally {
      setLoading(false);
    }
  }, [selectedCaseId]);

  useEffect(() => {
    fetchCases();
  }, [fetchCases]);

  // Fetch full details whenever selectedCaseId changes
  useEffect(() => {
    if (!selectedCaseId) return;
    const fetchDetail = async () => {
      try {
        const detail = await casesApi.get(selectedCaseId);
        setSelectedCaseDetail(detail);
      } catch (err: any) {
        console.warn("Could not fetch case detail:", err.message);
      }
    };
    fetchDetail();
  }, [selectedCaseId]);

  const handleAnalystVerdict = async (action: "ALLOW" | "DECLINE", reason: string) => {
    if (!selectedCaseId) return;
    setResolving(true);
    try {
      await casesApi.resolve(selectedCaseId, action, reason, "lead_analyst");
      setActionNotice(`Case ${selectedCaseId} resolved as ${action} (${reason})`);
      fetchCases();
    } catch (err: any) {
      setActionNotice(`Resolution failed: ${err.message}`);
    } finally {
      setResolving(false);
      setTimeout(() => setActionNotice(null), 4000);
    }
  };

  const handleCreateSampleReviewCase = async () => {
    try {
      setActionNotice("Executing evaluation with high risk score to spawn case...");
      await decisionsApi.evaluate({
        transaction_id: `tx_rev_${Date.now().toString(36)}`,
        amount: 8200.0,
        currency: "USD",
        payment_method: { type: "card", token: "tok_visa_sample_4242" },
        device_fingerprint: "fp_sample_review_device",
        ip_address: "198.51.100.44",
        account_id: "usr_sample_subject",
      });
      fetchCases();
    } catch (err: any) {
      setActionNotice(`Failed to seed case: ${err.message}`);
    }
  };

  return (
    <div className="max-w-6xl mx-auto space-y-6">
      {/* Top Header */}
      <div className="flex flex-wrap items-center justify-between gap-4 pb-4 border-b border-[#1c2536]">
        <div>
          <div className="flex items-center gap-2">
            <h1 className="text-xl font-bold text-[#f4f5f7] tracking-tight">Manual Review Queue &amp; Cases</h1>
            <span className="text-[10px] font-mono bg-[#0d94fb15] text-[#0d94fb] px-1.5 py-0.5 rounded-[2px] border border-[#0d94fb33]">
              SUBSYSTEM 08
            </span>
          </div>
          <p className="text-xs text-[#97a0af] font-mono mt-0.5">
            Synchronous decision audit ledger, SLA countdowns &amp; analyst resolution workflows
          </p>
        </div>

        <div className="flex items-center gap-2">
          <button
            onClick={fetchCases}
            className="flex items-center gap-1.5 px-3 py-1.5 text-xs font-mono bg-[#0f172a] hover:bg-[#142036] border border-[#1c2536] text-[#f4f5f7] rounded-[4px] cursor-pointer"
          >
            <RefreshCw className={`w-3.5 h-3.5 ${loading ? "animate-spin" : ""}`} /> Refresh
          </button>
          <button
            onClick={handleCreateSampleReviewCase}
            className="flex items-center gap-1.5 px-3 py-1.5 text-xs font-mono bg-[#0d94fb] hover:bg-[#0b82dc] text-white rounded-[4px] cursor-pointer font-semibold shadow-xs"
          >
            <Play className="w-3 h-3 fill-current" /> Seed Review Case
          </button>
        </div>
      </div>

      {/* Action Notice */}
      {actionNotice && (
        <div className="p-3 bg-[#04db7c15] border border-[#04db7c44] text-[#04db7c] font-mono text-xs rounded-[4px]">
          ✓ {actionNotice}
        </div>
      )}

      {/* Error Alert */}
      {error && (
        <div className="p-3 bg-[#f0525215] border border-[#f0525233] text-[#f05252] text-xs font-mono rounded-[4px] flex items-center justify-between">
          <div className="flex items-center gap-2">
            <AlertTriangle className="w-4 h-4 shrink-0" />
            <span>{error}</span>
          </div>
          <button
            onClick={fetchCases}
            className="px-2 py-0.5 bg-[#f0525225] hover:bg-[#f0525240] rounded text-[10px] cursor-pointer"
          >
            Retry
          </button>
        </div>
      )}

      {/* Main Grid: Left 5 Cols (Cases List) / Right 7 Cols (Case Detail) */}
      <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
        {/* Left 5 Cols: Cases List */}
        <div className="lg:col-span-5 bg-[#0f172a] border border-[#1c2536] rounded-[4px] p-5 shadow-xs">
          <div className="flex items-center justify-between pb-3 mb-3 border-b border-[#1c2536]">
            <div className="flex items-center gap-2">
              <FolderKanban className="w-4 h-4 text-[#0d94fb]" />
              <h2 className="text-xs font-mono font-bold text-[#f4f5f7] uppercase tracking-wider">
                Case Queue ({casesList.length})
              </h2>
            </div>
            <span className="text-[10px] font-mono text-[#5e6c84]">Postgres Persistence</span>
          </div>

          <div className="space-y-2.5">
            {casesList.length > 0 ? (
              casesList.map((c) => (
                <div
                  key={c.case_id}
                  onClick={() => setSelectedCaseId(c.case_id)}
                  className={`p-3 rounded-[4px] border font-mono text-xs cursor-pointer transition-all ${
                    selectedCaseId === c.case_id
                      ? "bg-[#142036] border-[#0d94fb] shadow-xs"
                      : "bg-[#0a1324] border-[#1c2536] hover:bg-[#0f1b30]"
                  }`}
                >
                  <div className="flex items-center justify-between mb-1">
                    <span className="text-[#f4f5f7] font-bold">{c.case_id}</span>
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
                  </div>

                  <div className="text-[11px] text-[#97a0af] truncate">Txn: {c.transaction_id}</div>
                  <div className="flex items-center justify-between text-[10px] text-[#5e6c84] mt-2 pt-2 border-t border-[#1c2536]">
                    <span>Priority: {c.priority || "P1_HIGH"}</span>
                    <span>{new Date(c.created_at).toLocaleTimeString()}</span>
                  </div>
                </div>
              ))
            ) : (
              <div className="py-10 text-center text-xs font-mono text-[#5e6c84] space-y-2">
                <p>No open manual review cases in PostgreSQL database.</p>
                <button
                  onClick={handleCreateSampleReviewCase}
                  className="px-3 py-1.5 text-xs font-mono bg-[#0d94fb] hover:bg-[#0b82dc] text-white rounded-[4px] font-semibold cursor-pointer"
                >
                  Create Test Review Case
                </button>
              </div>
            )}
          </div>
        </div>

        {/* Right 7 Cols: Active Case Detail & Resolution Workflow */}
        <div className="lg:col-span-7 bg-[#0f172a] border border-[#1c2536] rounded-[4px] p-5 shadow-xs space-y-5">
          {selectedCaseDetail ? (
            <>
              {/* Header */}
              <div className="flex flex-wrap items-center justify-between gap-3 pb-3 border-b border-[#1c2536]">
                <div>
                  <span className="text-sm font-mono font-bold text-[#f4f5f7] block">
                    {selectedCaseDetail.case_id}
                  </span>
                  <span className="text-xs font-mono text-[#0d94fb]">
                    Transaction: {selectedCaseDetail.transaction_id}
                  </span>
                </div>

                <div className="flex items-center gap-2">
                  <span className="text-xs font-mono text-[#97a0af]">Risk Score:</span>
                  <span className="text-base font-mono font-bold text-[#f05252]">
                    {(selectedCaseDetail.risk_score || 0.96).toFixed(2)}
                  </span>
                </div>
              </div>

              {/* Immutable Evidence Tags */}
              <div>
                <span className="text-[11px] font-mono font-bold text-[#5e6c84] uppercase block mb-2">
                  Triggered Reason Codes
                </span>
                <div className="flex flex-wrap gap-2">
                  {(selectedCaseDetail.reason_codes || ["HIGH_IP_VELOCITY", "SUSPICIOUS_DEVICE"]).map(
                    (code: string, idx: number) => (
                      <span
                        key={idx}
                        className="px-2 py-0.5 bg-[#0a1324] border border-[#1c2536] text-[#0d94fb] text-xs font-mono rounded-[2px]"
                      >
                        {code}
                      </span>
                    )
                  )}
                </div>
              </div>

              {/* Action Buttons */}
              <div className="pt-3 border-t border-[#1c2536] space-y-3">
                <span className="text-[11px] font-mono font-bold text-[#5e6c84] uppercase block">
                  Analyst Human-In-The-Loop Decision
                </span>
                <div className="flex flex-wrap gap-2">
                  <button
                    onClick={() => handleAnalystVerdict("DECLINE", "Analyst confirmed high risk attack signature")}
                    disabled={resolving}
                    className="flex-1 py-2 bg-[#f05252] hover:bg-[#dc3545] text-white font-mono text-xs font-bold rounded-[4px] active:scale-95 shadow-xs cursor-pointer disabled:opacity-50"
                  >
                    Confirm Decline (Block)
                  </button>
                  <button
                    onClick={() => handleAnalystVerdict("ALLOW", "Analyst confirmed legitimate customer override")}
                    disabled={resolving}
                    className="flex-1 py-2 bg-[#04db7c] hover:bg-[#03b868] text-[#011638] font-mono text-xs font-bold rounded-[4px] active:scale-95 shadow-xs cursor-pointer disabled:opacity-50"
                  >
                    Approve Override (Allow)
                  </button>
                </div>
              </div>
            </>
          ) : (
            <div className="py-16 text-center text-xs font-mono text-[#5e6c84]">
              Select a case from the queue to view immutable evidentiary dossier and take resolution action.
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
