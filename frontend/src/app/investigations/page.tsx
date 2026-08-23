"use client";

import React, { useState } from "react";
import {
  Bot,
} from "lucide-react";

interface InvestigationHypothesis {
  id: string;
  clusterId: string;
  confidence: number;
  threatType: "ACCOUNT_TAKEOVER" | "SYNTHETIC_IDENTITY_RING" | "MONEY_MULE_FARM";
  agentDossier: {
    threatHunter: string;
    graphAnalyst: string;
    amlOfficer: string;
  };
  recommendedActions: string[];
  status: "OPEN" | "CONFIRMED" | "DISMISSED";
  generatedAt: string;
}

const SEED_HYPOTHESES: InvestigationHypothesis[] = [
  {
    id: "hyp_001",
    clusterId: "#SYNTHETIC-MULE-CLUSTER-99",
    confidence: 0.96,
    threatType: "ACCOUNT_TAKEOVER",
    agentDossier: {
      threatHunter: "Impossible travel velocity (8,420 km/h) & bulletproof proxy pool verified on egress IP.",
      graphAnalyst: "Device hardware canvas entropy matches 14 synthetic accounts linked to Cyprus payout IBAN.",
      amlOfficer: "Transaction velocity spike >400% baseline following credential reset. SAR filing mandatory.",
    },
    recommendedActions: [
      "Revoke all active session bearer tokens for usr_sarah_connor",
      "Block outbound wire transfer #88419 ($14,500.00)",
      "Quarantine recipient IBAN CY88...9124 across entire platform",
    ],
    status: "OPEN",
    generatedAt: "2026-08-22 17:42:02",
  },
  {
    id: "hyp_002",
    clusterId: "#TOR-CARD-TESTING-POOL-04",
    confidence: 0.88,
    threatType: "SYNTHETIC_IDENTITY_RING",
    agentDossier: {
      threatHunter: "Burst of 42 micro-authorizations ($1.00 - $3.50) from rotating Tor exit nodes.",
      graphAnalyst: "Cards share identical BIN issuer prefix with randomized billing postal codes.",
      amlOfficer: "Automated velocity rules halted remaining 38 authorization attempts.",
    },
    recommendedActions: [
      "Enable aggressive CAPTCHA challenge on payment gateway",
      "Add BIN range 4111-22xx to heightened step-up policy",
    ],
    status: "CONFIRMED",
    generatedAt: "2026-08-22 14:15:00",
  },
];

export default function InvestigationsPage() {
  const [hypotheses] = useState<InvestigationHypothesis[]>(SEED_HYPOTHESES);
  const [actionNotice, setActionNotice] = useState<string | null>(null);

  const handleExecuteAction = (hypId: string, actionText: string) => {
    setActionNotice(`Executed and logged to cryptographic ledger: "${actionText}"`);
    setTimeout(() => setActionNotice(null), 4000);
  };

  return (
    <div className="max-w-6xl mx-auto space-y-6">
      {/* Top Header */}
      <div className="flex flex-wrap items-center justify-between gap-4 pb-4 border-b border-[#1c2536]">
        <div>
          <div className="flex items-center gap-2">
            <h1 className="text-xl font-bold text-[#f4f5f7] tracking-tight">Autonomous AI Investigations</h1>
            <span className="text-[10px] font-mono bg-[#c084fc15] text-[#c084fc] px-1.5 py-0.5 rounded-[2px] border border-[#c084fc33]">
              SUBSYSTEM 07
            </span>
          </div>
          <p className="text-xs text-[#97a0af] font-mono mt-0.5">
            Multi-agent consensus council: Threat Hunter, Graph Analyst & AML Compliance Officer
          </p>
        </div>

        <div className="flex items-center gap-2">
          <span className="px-2.5 py-1 text-xs font-mono bg-[#0f172a] border border-[#1c2536] text-[#04db7c] rounded-[4px] font-bold shadow-xs">
            Council Agents: 3 Active • Consensus: 0.96
          </span>
        </div>
      </div>

      {actionNotice && (
        <div className="p-3 bg-[#04db7c15] border border-[#04db7c44] text-[#04db7c] font-mono text-xs rounded-[4px] flex items-center justify-between">
          <span>✓ {actionNotice}</span>
        </div>
      )}

      {/* Hypotheses Cards */}
      <div className="space-y-6">
        {hypotheses.map((hyp: InvestigationHypothesis) => (
          <div key={hyp.id} className="bg-[#0f172a] border border-[#1c2536] rounded-[4px] p-5 shadow-xs space-y-4">
            <div className="flex flex-wrap items-center justify-between gap-3 pb-3 border-b border-[#1c2536]">
              <div className="flex items-center gap-3">
                <span className="text-sm font-mono font-bold text-[#f4f5f7]">{hyp.id}</span>
                <span className="text-xs font-mono text-[#0d94fb]">{hyp.clusterId}</span>
                <span
                  className={`text-[10px] font-bold px-1.5 py-0.5 rounded-[2px] border ${
                    hyp.status === "OPEN"
                      ? "bg-[#f0525215] text-[#f05252] border-[#f0525233]"
                      : "bg-[#04db7c15] text-[#04db7c] border-[#04db7c33]"
                  }`}
                >
                  {hyp.status}
                </span>
              </div>

              <div className="text-xs font-mono text-[#97a0af]">
                Confidence: <span className="text-[#04db7c] font-bold">{(hyp.confidence * 100).toFixed(0)}%</span> •{" "}
                {hyp.generatedAt}
              </div>
            </div>

            {/* Tripartite Agent Council Dossier */}
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              <div className="p-3.5 bg-[#0a1324] border border-[#1c2536] rounded-[4px]">
                <div className="flex items-center gap-1.5 mb-2 text-xs font-mono font-bold text-[#f59e0b]">
                  <Bot className="w-3.5 h-3.5" /> Threat Hunter Agent
                </div>
                <p className="text-xs text-[#f4f5f7] leading-relaxed font-sans">{hyp.agentDossier.threatHunter}</p>
              </div>

              <div className="p-3.5 bg-[#0a1324] border border-[#1c2536] rounded-[4px]">
                <div className="flex items-center gap-1.5 mb-2 text-xs font-mono font-bold text-[#0d94fb]">
                  <Bot className="w-3.5 h-3.5" /> Graph Analyst Agent
                </div>
                <p className="text-xs text-[#f4f5f7] leading-relaxed font-sans">{hyp.agentDossier.graphAnalyst}</p>
              </div>

              <div className="p-3.5 bg-[#0a1324] border border-[#1c2536] rounded-[4px]">
                <div className="flex items-center gap-1.5 mb-2 text-xs font-mono font-bold text-[#c084fc]">
                  <Bot className="w-3.5 h-3.5" /> AML Compliance Officer
                </div>
                <p className="text-xs text-[#f4f5f7] leading-relaxed font-sans">{hyp.agentDossier.amlOfficer}</p>
              </div>
            </div>

            {/* Recommended SOAR Remediations */}
            <div className="pt-2 border-t border-[#1c2536]">
              <span className="text-[11px] font-mono font-bold text-[#5e6c84] uppercase block mb-2">
                Autonomous SOAR Playbook Remediations (Human Authorization Required)
              </span>
              <div className="space-y-2">
                {hyp.recommendedActions.map((action: string, idx: number) => (
                  <div
                    key={idx}
                    className="p-2.5 bg-[#0a1324] border border-[#1c2536] rounded-[4px] flex items-center justify-between text-xs font-mono"
                  >
                    <span className="text-[#f4f5f7]">{action}</span>
                    <button
                      onClick={() => handleExecuteAction(hyp.id, action)}
                      className="px-2.5 py-1 text-[10px] font-mono font-bold bg-[#0d94fb] hover:bg-[#0b82dc] text-white rounded-[4px] active:scale-95 shadow-xs cursor-pointer"
                    >
                      Execute
                    </button>
                  </div>
                ))}
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
