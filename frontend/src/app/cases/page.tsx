"use client";

import React, { useState } from "react";
import { EvidenceList } from "@/components/ropus/EvidenceList";
import { CaseTimeline, TimelineEntry } from "@/components/ropus/CaseTimeline";
import { CANONICAL_BLOCKED_DECISION } from "@/lib/fixtures";
import {
  FileText,
  Clock,
} from "lucide-react";

interface CaseItem {
  id: string;
  transactionId: string;
  customerId: string;
  amount: number;
  currency: string;
  riskScore: number;
  priority: "P0_CRITICAL" | "P1_HIGH" | "P2_MEDIUM";
  status: "OPEN" | "IN_REVIEW" | "CONFIRMED_FRAUD" | "OVERRIDDEN";
  createdAt: string;
  assignedTo: string;
}

const INITIAL_CASES: CaseItem[] = [
  {
    id: "CASE-88419",
    transactionId: "tx_order_88419",
    customerId: "usr_sarah_connor",
    amount: 14500.0,
    currency: "USD",
    riskScore: 0.96,
    priority: "P0_CRITICAL",
    status: "OPEN",
    createdAt: "2026-08-22 17:42:02",
    assignedTo: "Lead Analyst (You)",
  },
  {
    id: "CASE-88415",
    transactionId: "tx_order_88415",
    customerId: "usr_alex_murphy",
    amount: 8200.0,
    currency: "USD",
    riskScore: 0.88,
    priority: "P1_HIGH",
    status: "IN_REVIEW",
    createdAt: "2026-08-22 16:30:10",
    assignedTo: "Compliance Team",
  },
  {
    id: "CASE-88390",
    transactionId: "tx_order_88390",
    customerId: "usr_ellen_ripley",
    amount: 320.0,
    currency: "USD",
    riskScore: 0.52,
    priority: "P2_MEDIUM",
    status: "CONFIRMED_FRAUD",
    createdAt: "2026-08-22 11:15:40",
    assignedTo: "Auto-Rules Action",
  },
];

const INITIAL_TIMELINE: TimelineEntry[] = [
  {
    id: "tl_001",
    timestamp: "2026-08-22 17:42:01.420",
    actor: "UnifiedRiskPipeline",
    action: "EVALUATION_COMPLETED",
    details: "Risk Score 0.96 (BLOCK) evaluated in 1.42ms. Rule RULE-VELOCITY-04 matched.",
    sha256Hash: "4f8a1e9c7a2b5d8e3f1c6d39a04b12ef7a2b5d8e3f1c6d39a04b12ef8a19bc7f",
  },
  {
    id: "tl_002",
    timestamp: "2026-08-22 17:42:02.100",
    actor: "AIAgentCouncil",
    action: "DOSSIER_SYNTHESIZED",
    details: "Threat Hunter, Graph Analyst, and AML Officer synthesized tripartite evidentiary dossier.",
    sha256Hash: "8e3f1c6d39a04b12ef7a2b5d8e3f1c6d39a04b12ef8a19bc7f4f8a1e9c7a2b5d",
  },
  {
    id: "tl_003",
    timestamp: "2026-08-22 17:42:02.500",
    actor: "CaseManager",
    action: "CASE_OPENED_P0",
    details: "Case #CASE-88419 opened with P0 Critical priority; assigned to Lead Analyst.",
    sha256Hash: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
  },
];

export default function CasesPage() {
  const [casesList, setCasesList] = useState<CaseItem[]>(INITIAL_CASES);
  const [selectedCaseId, setSelectedCaseId] = useState<string>("CASE-88419");
  const [timeline, setTimeline] = useState<TimelineEntry[]>(INITIAL_TIMELINE);
  const [actionNotice, setActionNotice] = useState<string | null>(null);

  const activeCase = casesList.find((c) => c.id === selectedCaseId) || casesList[0];

  const handleAnalystVerdict = (verdict: "CONFIRM_FRAUD" | "ESCALATE_AML" | "OVERRIDE") => {
    const updatedStatus =
      verdict === "CONFIRM_FRAUD"
        ? "CONFIRMED_FRAUD"
        : verdict === "OVERRIDE"
        ? "OVERRIDDEN"
        : "IN_REVIEW";

    setCasesList((prev) =>
      prev.map((c) => (c.id === activeCase.id ? { ...c, status: updatedStatus } : c))
    );

    const newLog: TimelineEntry = {
      id: `tl_${Date.now()}`,
      timestamp: new Date().toISOString().replace("T", " ").slice(0, 19),
      actor: "Lead Analyst (You)",
      action: verdict === "CONFIRM_FRAUD" ? "ANALYST_CONFIRMED_FRAUD" : verdict === "ESCALATE_AML" ? "ESCALATED_TO_AML_SAR" : "OVERRIDDEN_FALSE_POSITIVE",
      details:
        verdict === "CONFIRM_FRAUD"
          ? "Confirmed unauthorized ATO transaction. Wire transfer halted, session revoked, SAR drafted."
          : verdict === "ESCALATE_AML"
          ? "Escalated to AML Compliance Unit with complete evidentiary dossier attached."
          : "Overridden by human analyst as authorized high-value business wire.",
      sha256Hash: "a1b2c3d4e5f60718293a4b5c6d7e8f90123456789abcdef0123456789abcdef0",
    };

    setTimeline((prev) => [newLog, ...prev]);
    setActionNotice(`Case ${activeCase.id} updated: ${verdict} recorded.`);
    setTimeout(() => setActionNotice(null), 4000);
  };

  const getPriorityBadge = (p: CaseItem["priority"]) => {
    switch (p) {
      case "P0_CRITICAL":
        return <span className="px-1.5 py-0.5 text-[10px] font-mono font-bold bg-[#f0525215] text-[#f05252] border border-[#f0525233] rounded-[2px]">P0 CRITICAL</span>;
      case "P1_HIGH":
        return <span className="px-1.5 py-0.5 text-[10px] font-mono font-bold bg-[#f9731615] text-[#f97316] border border-[#f9731633] rounded-[2px]">P1 HIGH</span>;
      case "P2_MEDIUM":
        return <span className="px-1.5 py-0.5 text-[10px] font-mono font-bold bg-[#f59e0b15] text-[#f59e0b] border border-[#f59e0b33] rounded-[2px]">P2 MED</span>;
    }
  };

  const getStatusBadge = (s: CaseItem["status"]) => {
    switch (s) {
      case "OPEN":
        return <span className="px-1.5 py-0.5 text-[10px] font-mono font-bold bg-[#f0525215] text-[#f05252] border border-[#f0525233] rounded-[2px]">OPEN</span>;
      case "IN_REVIEW":
        return <span className="px-1.5 py-0.5 text-[10px] font-mono font-bold bg-[#f59e0b15] text-[#f59e0b] border border-[#f59e0b33] rounded-[2px]">IN REVIEW</span>;
      case "CONFIRMED_FRAUD":
        return <span className="px-1.5 py-0.5 text-[10px] font-mono font-bold bg-[#04db7c15] text-[#04db7c] border border-[#04db7c33] rounded-[2px]">CONFIRMED FRAUD</span>;
      case "OVERRIDDEN":
        return <span className="px-1.5 py-0.5 text-[10px] font-mono font-bold bg-[#5e6c8415] text-[#97a0af] border border-[#5e6c8433] rounded-[2px]">OVERRIDDEN</span>;
    }
  };

  return (
    <div className="max-w-6xl mx-auto space-y-6">
      {/* Top Header */}
      <div className="flex flex-wrap items-center justify-between gap-4 pb-4 border-b border-[#1c2536]">
        <div>
          <div className="flex items-center gap-2">
            <h1 className="text-xl font-bold text-[#f4f5f7] tracking-tight">Case Review & AI Investigation</h1>
            <span className="text-[10px] font-mono bg-[#0d94fb15] text-[#0d94fb] px-1.5 py-0.5 rounded-[2px] border border-[#0d94fb33]">
              HUMAN GOVERNANCE
            </span>
          </div>
          <p className="text-xs text-[#97a0af] font-mono mt-0.5">
            Human-in-the-loop governance, multi-agent dossiers & immutable SHA-256 audit ledger
          </p>
        </div>

        <div className="flex items-center gap-2">
          <span className="text-xs font-mono text-[#5e6c84]">QUEUE METRIC:</span>
          <span className="px-2.5 py-1 text-xs font-mono bg-[#0f172a] border border-[#1c2536] text-[#f4f5f7] rounded-[4px] font-bold shadow-xs">
            1 P0 Active • 0 Breached SLAs
          </span>
        </div>
      </div>

      {actionNotice && (
        <div className="p-3 bg-[#04db7c15] border border-[#04db7c44] text-[#04db7c] font-mono text-xs rounded-[4px] flex items-center justify-between">
          <span>✓ {actionNotice}</span>
        </div>
      )}

      {/* Main Grid: Left Review Queue List / Right Detailed Active Case */}
      <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
        {/* Left 4 Cols: Review Queue Table */}
        <div className="lg:col-span-4 bg-[#0f172a] border border-[#1c2536] rounded-[4px] p-4 flex flex-col justify-between shadow-xs">
          <div>
            <div className="flex items-center justify-between pb-3 mb-3 border-b border-[#1c2536]">
              <span className="text-xs font-mono font-bold text-[#f4f5f7] uppercase tracking-wider">
                Review Cases ({casesList.length})
              </span>
              <span className="text-[10px] font-mono text-[#5e6c84]">Tier 1 Queue</span>
            </div>

            <div className="space-y-2">
              {casesList.map((c) => {
                const isSelected = selectedCaseId === c.id;
                return (
                  <button
                    key={c.id}
                    onClick={() => setSelectedCaseId(c.id)}
                    className={`w-full p-3 rounded-[4px] border text-left transition-all cursor-pointer ${
                      isSelected
                        ? "bg-[#0a1324] border-[#0d94fb] ring-1 ring-[#0d94fb]"
                        : "bg-[#070e1c] border-[#1c2536] hover:border-[#2c3b52]"
                    }`}
                  >
                    <div className="flex items-center justify-between gap-2 mb-1.5">
                      <span className="text-xs font-mono font-bold text-[#f4f5f7]">{c.id}</span>
                      {getPriorityBadge(c.priority)}
                    </div>

                    <div className="flex items-center justify-between text-xs font-mono text-[#97a0af] mb-1">
                      <span>{c.customerId}</span>
                      <span className="text-[#f4f5f7] font-bold">
                        ${c.amount.toLocaleString()} {c.currency}
                      </span>
                    </div>

                    <div className="flex items-center justify-between text-[10px] font-mono text-[#5e6c84]">
                      <span>Score: {c.riskScore.toFixed(2)}</span>
                      {getStatusBadge(c.status)}
                    </div>
                  </button>
                );
              })}
            </div>
          </div>

          <div className="pt-3 mt-3 border-t border-[#1c2536] text-[10px] font-mono text-[#5e6c84] text-center">
            Persisted in PostgreSQL • Redis Cache Active
          </div>
        </div>

        {/* Right 8 Cols: Active Case Investigation & Governance Actions */}
        <div className="lg:col-span-8 space-y-6">
          {/* Active Case Header Box */}
          <div className="bg-[#0f172a] border border-[#1c2536] rounded-[4px] p-5 shadow-xs">
            <div className="flex flex-wrap items-center justify-between gap-3 pb-3 mb-4 border-b border-[#1c2536]">
              <div className="flex items-center gap-3">
                <span className="text-base font-bold font-mono text-[#f4f5f7]">{activeCase.id}</span>
                {getPriorityBadge(activeCase.priority)}
                {getStatusBadge(activeCase.status)}
              </div>

              <div className="text-xs font-mono text-[#97a0af]">
                Assigned: <span className="text-[#f4f5f7] font-semibold">{activeCase.assignedTo}</span>
              </div>
            </div>

            <div className="grid grid-cols-2 md:grid-cols-4 gap-4 font-mono text-xs mb-4">
              <div>
                <span className="text-[10px] text-[#5e6c84] block">TRANSACTION ID</span>
                <span className="text-[#f4f5f7] font-semibold">{activeCase.transactionId}</span>
              </div>
              <div>
                <span className="text-[10px] text-[#5e6c84] block">AMOUNT</span>
                <span className="text-[#f4f5f7] font-bold">
                  ${activeCase.amount.toLocaleString()} {activeCase.currency}
                </span>
              </div>
              <div>
                <span className="text-[10px] text-[#5e6c84] block">EVALUATED SCORE</span>
                <span className="text-[#f05252] font-bold">{activeCase.riskScore.toFixed(2)} (BLOCK)</span>
              </div>
              <div>
                <span className="text-[10px] text-[#5e6c84] block">OPENED AT</span>
                <span className="text-[#97a0af]">{activeCase.createdAt}</span>
              </div>
            </div>

            {/* Analyst Action Buttons (Razorpay Blade Styling) */}
            <div className="pt-3 border-t border-[#1c2536] flex flex-wrap items-center justify-between gap-3">
              <span className="text-xs font-mono text-[#5e6c84]">EXECUTE GOVERNANCE ACTION:</span>
              <div className="flex items-center gap-2">
                <button
                  onClick={() => handleAnalystVerdict("CONFIRM_FRAUD")}
                  className="px-3 py-1.5 text-xs font-mono font-bold bg-[#f05252] hover:bg-[#d93b3b] text-white rounded-[4px] active:scale-95 shadow-xs cursor-pointer"
                >
                  Confirm Fraud Block
                </button>
                <button
                  onClick={() => handleAnalystVerdict("ESCALATE_AML")}
                  className="px-3 py-1.5 text-xs font-mono font-semibold bg-[#f59e0b18] hover:bg-[#f59e0b28] text-[#f59e0b] border border-[#f59e0b44] rounded-[4px] active:scale-95 cursor-pointer"
                >
                  Escalate SAR
                </button>
                <button
                  onClick={() => handleAnalystVerdict("OVERRIDE")}
                  className="px-3 py-1.5 text-xs font-mono bg-[#142036] hover:bg-[#1c2b48] text-[#97a0af] hover:text-[#f4f5f7] border border-[#1c2536] rounded-[4px] active:scale-95 cursor-pointer"
                >
                  Mark False Positive
                </button>
              </div>
            </div>
          </div>

          {/* AI Investigator Dossier (Observed / Inferred / Recommended) */}
          <div className="bg-[#0f172a] border border-[#1c2536] rounded-[4px] p-5 shadow-xs">
            <div className="flex items-center justify-between pb-3 mb-3 border-b border-[#1c2536]">
              <div className="flex items-center gap-2">
                <FileText className="w-4 h-4 text-[#0d94fb]" />
                <h2 className="text-xs font-mono font-bold text-[#f4f5f7] uppercase tracking-wider">
                  AI Investigator Council Dossier
                </h2>
              </div>
              <span className="text-[10px] font-mono text-[#5e6c84]">Multi-Persona Consensus (Subsystem 07)</span>
            </div>

            <EvidenceList
              observed={CANONICAL_BLOCKED_DECISION.evidence.observed}
              inferred={CANONICAL_BLOCKED_DECISION.evidence.inferred}
              recommended={CANONICAL_BLOCKED_DECISION.evidence.recommended}
              onActionConfirm={() => handleAnalystVerdict("CONFIRM_FRAUD")}
            />
          </div>

          {/* Immutable Audit Timeline */}
          <div className="bg-[#0f172a] border border-[#1c2536] rounded-[4px] p-5 shadow-xs">
            <div className="flex items-center justify-between pb-3 mb-3 border-b border-[#1c2536]">
              <div className="flex items-center gap-2">
                <Clock className="w-4 h-4 text-[#04db7c]" />
                <h2 className="text-xs font-mono font-bold text-[#f4f5f7] uppercase tracking-wider">
                  Immutable Cryptographic Audit Trail
                </h2>
              </div>
              <span className="text-[10px] font-mono text-[#5e6c84]">SHA-256 Hash Chain (Subsystem 08)</span>
            </div>

            <CaseTimeline entries={timeline} />
          </div>
        </div>
      </div>
    </div>
  );
}
