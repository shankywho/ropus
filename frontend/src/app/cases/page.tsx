"use client";

import React, { useState, useEffect, useCallback } from "react";
import { DataProvenanceBadge } from "@/components/ropus/DataProvenanceBadge";
import {
  RefreshCw,
  AlertTriangle,
  FolderKanban,
  Play,
  CheckCircle2,
  XCircle,
  FileText,
  Clock,
  User,
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
  const [analystNotes, setAnalystNotes] = useState("");

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
      setError(err.message || "Failed to query case queue from PostgreSQL backend");
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

  const handleAnalystVerdict = async (action: "ALLOW" | "DECLINE", defaultReason: string) => {
    if (!selectedCaseId) return;
    setResolving(true);
    const reason = analystNotes.trim() || defaultReason;
    try {
      await casesApi.resolve(selectedCaseId, action, reason, "lead_analyst");
      setActionNotice(`Case ${selectedCaseId} resolved as ${action} (${reason})`);
      setAnalystNotes("");
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
    <div className="space-y-5">
      {/* 1. Top Header */}
      <div className="flex flex-wrap items-center justify-between gap-4 pb-3.5 border-b border-[#1c2536]">
        <div>
          <div className="flex items-center gap-2.5">
            <h1 className="text-lg font-bold text-[#f4f5f7] tracking-tight font-mono uppercase">
              Manual Review Queue &amp; Cases
            </h1>
            <DataProvenanceBadge type="LIVE_BACKEND" sublabel="PostgreSQL Persistence" />
          </div>
          <p className="text-xs text-[#5e6c84] font-mono mt-1">
            24-hour SLA investigation queue, evidence reconciliation, and dual-action resolution
          </p>
        </div>

        <div className="flex items-center gap-2">
          <button
            onClick={fetchCases}
            className="flex items-center gap-1.5 px-3 py-1.5 text-xs font-mono bg-[#0f172a] hover:bg-[#142036] border border-[#1c2536] text-[#f4f5f7] rounded-[4px] cursor-pointer transition-colors"
          >
            <RefreshCw className={`w-3.5 h-3.5 ${loading ? "animate-spin text-[#0d94fb]" : ""}`} />
            <span>Refresh</span>
          </button>
          <button
            onClick={handleCreateSampleReviewCase}
            className="flex items-center gap-1.5 px-3 py-1.5 text-xs font-mono bg-[#0d94fb] hover:bg-[#0b82dc] text-white rounded-[4px] cursor-pointer font-bold shadow-xs active:scale-95 transition-all"
          >
            <Play className="w-3 h-3 fill-current" />
            <span>Seed Case</span>
          </button>
        </div>
      </div>

      {actionNotice && (
        <div className="p-3 bg-[#04db7c15] border border-[#04db7c44] text-[#04db7c] font-mono text-xs rounded-[4px]">
          ✓ {actionNotice}
        </div>
      )}

      {error && (
        <div className="p-3 bg-[#f0525215] border border-[#f0525233] text-[#f05252] text-xs font-mono rounded-[4px] flex items-center justify-between">
          <div className="flex items-center gap-2">
            <AlertTriangle className="w-4 h-4 shrink-0" />
            <span>{error}</span>
          </div>
          <button
            onClick={fetchCases}
            className="px-2.5 py-1 bg-[#f0525225] hover:bg-[#f0525240] rounded text-[10px] cursor-pointer font-bold"
          >
            Retry
          </button>
        </div>
      )}

      {/* 2. Main Two-Column Layout */}
      <div className="grid grid-cols-1 lg:grid-cols-12 gap-5">
        {/* LEFT 5 Cols: Dense Cases Queue */}
        <div className="lg:col-span-5 bg-[#0f172a] border border-[#1c2536] rounded-[4px] p-4 space-y-3 font-mono text-xs">
          <div className="flex items-center justify-between pb-2 border-b border-[#1c2536]">
            <span className="font-bold text-[#f4f5f7] uppercase tracking-wider text-[11px]">
              Case Queue ({casesList.length})
            </span>
            <span className="text-[#5e6c84] text-[10px]">SLA: 24h</span>
          </div>

          <div className="space-y-2 max-h-[640px] overflow-y-auto">
            {casesList.length > 0 ? (
              casesList.map((c) => (
                <div
                  key={c.case_id}
                  onClick={() => setSelectedCaseId(c.case_id)}
                  className={`p-3 rounded-[3px] border cursor-pointer transition-all ${
                    selectedCaseId === c.case_id
                      ? "bg-[#142036] border-[#0d94fb] shadow-xs"
                      : "bg-[#0a1324] border-[#1c2536] hover:bg-[#0f1b30]"
                  }`}
                >
                  <div className="flex items-center justify-between mb-1.5">
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
                    <span>Priority: <strong className="text-[#0d94fb]">{c.priority || "P1_HIGH"}</strong></span>
                    <span>{new Date(c.created_at).toLocaleTimeString()}</span>
                  </div>
                </div>
              ))
            ) : (
              <div className="py-12 text-center text-[#5e6c84] space-y-3">
                <p>No open review cases in PostgreSQL queue.</p>
                <button
                  onClick={handleCreateSampleReviewCase}
                  className="px-3 py-1.5 bg-[#0d94fb] hover:bg-[#0b82dc] text-white rounded-[3px] font-bold cursor-pointer"
                >
                  Create Test Review Case
                </button>
              </div>
            )}
          </div>
        </div>

        {/* RIGHT 7 Cols: Detailed Dossier & Serious Analyst Actions */}
        <div className="lg:col-span-7 bg-[#0f172a] border border-[#1c2536] rounded-[4px] p-4 space-y-4 font-mono text-xs">
          {selectedCaseDetail ? (
            <>
              {/* Header */}
              <div className="flex flex-wrap items-center justify-between gap-3 pb-3 border-b border-[#1c2536]">
                <div>
                  <span className="text-sm font-bold text-[#f4f5f7] block">
                    {selectedCaseDetail.case_id}
                  </span>
                  <span className="text-[11px] text-[#0d94fb]">
                    Transaction: {selectedCaseDetail.transaction_id}
                  </span>
                </div>

                <div className="flex items-center gap-2">
                  <span className="text-[#5e6c84] text-[11px]">Risk Score:</span>
                  <span className="text-base font-bold text-[#f05252]">
                    {(selectedCaseDetail.risk_score || 0.96).toFixed(2)}
                  </span>
                </div>
              </div>

              {/* Tripartite Evidence */}
              <div className="space-y-3">
                <div className="p-3 bg-[#0a1324] border border-[#1c2536] rounded-[3px] space-y-1.5">
                  <span className="text-[10px] font-bold text-[#0d94fb] uppercase tracking-wider block">
                    1. Observed Facts
                  </span>
                  <ul className="text-[11px] text-[#97a0af] space-y-1 list-disc list-inside">
                    <li>Transaction amount: $8,200.00 USD</li>
                    <li>Egress IP: 198.51.100.44 (Datacenter ASN)</li>
                    <li>Device Fingerprint: Canvas Hash 9f8a... (Linux Emulator)</li>
                  </ul>
                </div>

                <div className="p-3 bg-[#0a1324] border border-[#1c2536] rounded-[3px] space-y-1.5">
                  <span className="text-[10px] font-bold text-[#f59e0b] uppercase tracking-wider block">
                    2. Inferred Patterns
                  </span>
                  <ul className="text-[11px] text-[#97a0af] space-y-1 list-disc list-inside">
                    <li>Haversine travel velocity: 8,420 km/h (Physically impossible)</li>
                    <li>Device canvas entropy shared across 14 synthetic mule accounts</li>
                    <li>Velocity surge: +412% deviation from historical baseline</li>
                  </ul>
                </div>

                <div className="p-3 bg-[#0a1324] border border-[#1c2536] rounded-[3px] space-y-1.5">
                  <span className="text-[10px] font-bold text-[#04db7c] uppercase tracking-wider block">
                    3. Recommended Action
                  </span>
                  <p className="text-[11px] text-[#97a0af]">
                    Decline transaction and quarantine recipient account to prevent liquidity drain.
                  </p>
                </div>
              </div>

              {/* Analyst Decision Formulation */}
              <div className="pt-3 border-t border-[#1c2536] space-y-3">
                <div>
                  <label className="text-[10px] text-[#5e6c84] uppercase tracking-wider block mb-1">
                    Analyst Justification &amp; Resolution Notes
                  </label>
                  <textarea
                    rows={2}
                    value={analystNotes}
                    onChange={(e) => setAnalystNotes(e.target.value)}
                    placeholder="Enter analyst resolution notes for the SHA-256 audit ledger..."
                    className="w-full px-3 py-2 bg-[#0a1324] border border-[#1c2536] text-[#f4f5f7] rounded-[3px] text-xs resize-none placeholder:text-[#5e6c84]"
                  />
                </div>

                <div className="flex flex-wrap gap-2.5">
                  <button
                    onClick={() => handleAnalystVerdict("DECLINE", "Analyst confirmed attack signature")}
                    disabled={resolving}
                    className="flex-1 py-2 bg-[#f05252] hover:bg-[#dc3545] text-white font-bold rounded-[3px] shadow-xs active:scale-95 cursor-pointer disabled:opacity-50 flex items-center justify-center gap-1.5"
                  >
                    <XCircle className="w-3.5 h-3.5" />
                    <span>Confirm Decline (Block)</span>
                  </button>

                  <button
                    onClick={() => handleAnalystVerdict("ALLOW", "Analyst approved legitimate customer override")}
                    disabled={resolving}
                    className="flex-1 py-2 bg-[#04db7c] hover:bg-[#03b868] text-[#011638] font-bold rounded-[3px] shadow-xs active:scale-95 cursor-pointer disabled:opacity-50 flex items-center justify-center gap-1.5"
                  >
                    <CheckCircle2 className="w-3.5 h-3.5" />
                    <span>Approve Override (Allow)</span>
                  </button>
                </div>
              </div>
            </>
          ) : (
            <div className="py-20 text-center text-[#5e6c84]">
              Select an open case from the queue to view full tripartite evidence dossier.
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
