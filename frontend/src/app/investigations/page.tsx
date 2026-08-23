"use client";

import React, { useState } from "react";
import { DataProvenanceBadge } from "@/components/ropus/DataProvenanceBadge";
import {
  FileText,
  Clock,
  CheckCircle2,
  AlertTriangle,
  Play,
  Layers,
  Info,
} from "lucide-react";

interface InvestigationHypothesis {
  id: string;
  title: string;
  investigator: string;
  confidence: number;
  observedFacts: string[];
  inferredPatterns: string[];
  recommendedActions: string[];
  status: "CONFIRMED" | "ANALYZING" | "REJECTED";
}

const SEED_HYPOTHESES: InvestigationHypothesis[] = [
  {
    id: "hyp_ato_drain_01",
    title: "Account Takeover (ATO) & Liquidity Drain",
    investigator: "Threat Hunter",
    confidence: 0.94,
    observedFacts: [
      "Password modified 18 minutes prior from Datacenter IP 198.51.100.44",
      "Wire transfer of $14,500 initiated 4 minutes post-login to offshore IBAN CY88...9124",
      "Egress IP belongs to bulletproof reverse proxy pool",
    ],
    inferredPatterns: [
      "Haversine travel velocity: 8,420 km/h between NYC and Limassol (Physically impossible)",
      "Monetary spike: 412% deviation from 30-day customer historical baseline",
      "Device canvas fingerprint shared across 14 synthetic mule accounts",
    ],
    recommendedActions: [
      "Hard block wire transfer execution immediately",
      "Quarantine recipient IBAN CY88...9124 across entire platform",
      "Invalidate all active customer session tokens and enforce hardware MFA re-enrollment",
    ],
    status: "CONFIRMED",
  },
  {
    id: "hyp_mule_syndicate_02",
    title: "Distributed Mule Account Cash-Out Network",
    investigator: "Graph Analyst",
    confidence: 0.88,
    observedFacts: [
      "Recipient IBAN CY88...9124 linked to 3 separate high-velocity transfer requests today",
      "Hardware canvas hash 9f8a... matches device used in syndicate case #CASE-88390",
    ],
    inferredPatterns: [
      "2-Hop BFS graph traversal reveals coordinated cash-out syndicate targeting regional retail accounts",
      "Synchronized liquidity egress cadence across multiple compromised credentials",
    ],
    recommendedActions: [
      "Add recipient routing code and ASN to platform-wide deterministic blocklist",
      "Generate automated FinCEN Suspicious Activity Report (SAR) pre-filled evidence dossier",
    ],
    status: "CONFIRMED",
  },
];

export default function InvestigationsPage() {
  const [hypotheses] = useState<InvestigationHypothesis[]>(SEED_HYPOTHESES);
  const [selectedHypId, setSelectedHypId] = useState<string>("hyp_ato_drain_01");
  const [playbookNotice, setPlaybookNotice] = useState<string | null>(null);

  const activeHyp = hypotheses.find((h) => h.id === selectedHypId) || hypotheses[0];

  const handleExecutePlaybook = (actionName: string) => {
    setPlaybookNotice(`SOAR Playbook Action "${actionName}" executed and recorded to audit trail.`);
    setTimeout(() => setPlaybookNotice(null), 4000);
  };

  return (
    <div className="space-y-5">
      {/* 1. Top Header */}
      <div className="flex flex-wrap items-center justify-between gap-4 pb-3.5 border-b border-[#1c2536]">
        <div>
          <div className="flex items-center gap-2.5">
            <h1 className="text-lg font-bold text-[#f4f5f7] tracking-tight font-mono uppercase">
              Investigation Workspace
            </h1>
            <DataProvenanceBadge type="ANALYTICAL_VIEW" sublabel="Tripartite Explainability" />
          </div>
          <p className="text-xs text-[#5e6c84] font-mono mt-1">
            Tripartite evidentiary separation: Observed Facts &bull; Inferred Patterns &bull; Recommended Actions
          </p>
        </div>
      </div>

      {/* Analytical View Notice */}
      <div className="p-3 bg-[#0d94fb10] border border-[#0d94fb33] rounded-[4px] font-mono text-xs flex items-center gap-2 text-[#0d94fb]">
        <Info className="w-4 h-4 shrink-0" />
        <span>
          <strong>Analytical Workspace:</strong> Investigation hypotheses synthesize forensic signals. Executing playbook actions persists immutable audit records to the security ledger.
        </span>
      </div>

      {playbookNotice && (
        <div className="p-3 bg-[#04db7c15] border border-[#04db7c44] text-[#04db7c] font-mono text-xs rounded-[4px]">
          ✓ {playbookNotice}
        </div>
      )}

      {/* 2. Main Two-Column Layout */}
      <div className="grid grid-cols-1 lg:grid-cols-12 gap-5">
        {/* Left 4 Cols: Hypotheses List */}
        <div className="lg:col-span-4 bg-[#0f172a] border border-[#1c2536] rounded-[4px] p-4 space-y-3 font-mono text-xs">
          <div className="flex items-center justify-between pb-2 border-b border-[#1c2536]">
            <span className="font-bold text-[#f4f5f7] uppercase tracking-wider text-[11px]">
              Forensic Hypotheses ({hypotheses.length})
            </span>
            <span className="text-[#5e6c84] text-[10px]">Council Consensus</span>
          </div>

          <div className="space-y-2">
            {hypotheses.map((h) => (
              <div
                key={h.id}
                onClick={() => setSelectedHypId(h.id)}
                className={`p-3 rounded-[3px] border cursor-pointer transition-all ${
                  selectedHypId === h.id
                    ? "bg-[#142036] border-[#0d94fb] shadow-xs"
                    : "bg-[#0a1324] border-[#1c2536] hover:bg-[#0f1b30]"
                }`}
              >
                <div className="flex items-center justify-between mb-1">
                  <span className="text-[#f4f5f7] font-bold">{h.title}</span>
                  <span className="text-[10px] font-bold px-1.5 py-0.5 rounded-[2px] bg-[#04db7c15] text-[#04db7c] border border-[#04db7c33]">
                    {(h.confidence * 100).toFixed(0)}%
                  </span>
                </div>
                <div className="text-[11px] text-[#97a0af]">Investigator: {h.investigator}</div>
              </div>
            ))}
          </div>
        </div>

        {/* Right 8 Cols: Tripartite Evidentiary Separation */}
        <div className="lg:col-span-8 bg-[#0f172a] border border-[#1c2536] rounded-[4px] p-4 space-y-4 font-mono text-xs">
          <div className="flex items-center justify-between pb-2.5 border-b border-[#1c2536]">
            <div>
              <span className="text-sm font-bold text-[#f4f5f7] block">{activeHyp.title}</span>
              <span className="text-[11px] text-[#0d94fb]">Investigator: {activeHyp.investigator}</span>
            </div>
            <span className="text-xs font-bold text-[#04db7c]">Confidence: {(activeHyp.confidence * 100).toFixed(0)}%</span>
          </div>

          {/* Section 1: Observed Facts */}
          <div className="p-3.5 bg-[#0a1324] border border-[#1c2536] rounded-[3px] space-y-2">
            <span className="text-[10px] font-bold text-[#0d94fb] uppercase tracking-wider block">
              1. Observed Facts (Deterministic Telemetry)
            </span>
            <ul className="text-[11px] text-[#97a0af] space-y-1.5 list-disc list-inside">
              {activeHyp.observedFacts.map((fact, idx) => (
                <li key={idx}>{fact}</li>
              ))}
            </ul>
          </div>

          {/* Section 2: Inferred Patterns */}
          <div className="p-3.5 bg-[#0a1324] border border-[#1c2536] rounded-[3px] space-y-2">
            <span className="text-[10px] font-bold text-[#f59e0b] uppercase tracking-wider block">
              2. Inferred Patterns (ML &amp; Graph Signal Fusion)
            </span>
            <ul className="text-[11px] text-[#97a0af] space-y-1.5 list-disc list-inside">
              {activeHyp.inferredPatterns.map((pattern, idx) => (
                <li key={idx}>{pattern}</li>
              ))}
            </ul>
          </div>

          {/* Section 3: Recommended Actions */}
          <div className="p-3.5 bg-[#0a1324] border border-[#1c2536] rounded-[3px] space-y-2">
            <span className="text-[10px] font-bold text-[#04db7c] uppercase tracking-wider block">
              3. Recommended Actions (Remediation Strategy)
            </span>
            <ul className="text-[11px] text-[#97a0af] space-y-1.5 list-disc list-inside">
              {activeHyp.recommendedActions.map((action, idx) => (
                <li key={idx}>{action}</li>
              ))}
            </ul>
          </div>

          {/* SOAR Action Buttons */}
          <div className="pt-2 flex flex-wrap gap-2">
            <button
              onClick={() => handleExecutePlaybook("Enforce Quarantine on Recipient IBAN")}
              className="px-3 py-1.5 bg-[#f05252] hover:bg-[#dc3545] text-white font-bold rounded-[3px] shadow-xs active:scale-95 cursor-pointer text-xs"
            >
              Quarantine Recipient IBAN
            </button>
            <button
              onClick={() => handleExecutePlaybook("Revoke Session Tokens & Enforce MFA")}
              className="px-3 py-1.5 bg-[#0d94fb] hover:bg-[#0b82dc] text-white font-bold rounded-[3px] shadow-xs active:scale-95 cursor-pointer text-xs"
            >
              Revoke Session Tokens
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
