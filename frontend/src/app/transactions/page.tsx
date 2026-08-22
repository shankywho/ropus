"use client";

import React, { useState } from "react";
import { DecisionSummary } from "@/components/ropus/DecisionSummary";
import { RiskFactor } from "@/components/ropus/RiskFactor";
import { EvidenceList } from "@/components/ropus/EvidenceList";
import { CANONICAL_BLOCKED_DECISION, CANONICAL_APPROVED_DECISION, RiskDecisionPayload } from "@/lib/fixtures";
import {
  ChevronDown,
  ChevronRight,
  Terminal,
  Activity,
  Layers,
  FileCode2,
} from "lucide-react";

export default function RiskDecisionPage() {
  const [selectedScenario, setSelectedScenario] = useState<"BLOCKED" | "APPROVED">("BLOCKED");
  const [isTechDetailsOpen, setIsTechDetailsOpen] = useState(false);
  const [actionFeedback, setActionFeedback] = useState<string | null>(null);

  const decision: RiskDecisionPayload =
    selectedScenario === "BLOCKED" ? CANONICAL_BLOCKED_DECISION : CANONICAL_APPROVED_DECISION;

  const handleAction = (actionName: string) => {
    setActionFeedback(`Action "${actionName}" executed & logged to SHA-256 ledger`);
    setTimeout(() => setActionFeedback(null), 4000);
  };

  return (
    <div className="max-w-6xl mx-auto space-y-6">
      {/* Top Header / Switcher */}
      <div className="flex flex-wrap items-center justify-between gap-4 pb-4 border-b border-[#1c2536]">
        <div>
          <div className="flex items-center gap-2">
            <h1 className="text-xl font-bold text-[#f4f5f7] tracking-tight">Risk Decision Intelligence</h1>
            <span className="text-[10px] font-mono bg-[#0d94fb15] text-[#0d94fb] px-1.5 py-0.5 rounded-[2px] border border-[#0d94fb33]">
              LIVE INFERENCE
            </span>
          </div>
          <p className="text-xs text-[#97a0af] font-mono mt-0.5">
            Single-loop synchronous evaluation pipeline & additive factor attribution
          </p>
        </div>

        {/* Toggle between canonical attack scenario and organic benign purchase */}
        <div className="flex items-center gap-1.5 bg-[#0f172a] p-1 border border-[#1c2536] rounded-[4px]">
          <button
            onClick={() => setSelectedScenario("BLOCKED")}
            className={`px-3 py-1.5 text-xs font-mono rounded-[3px] font-semibold transition-all cursor-pointer ${
              selectedScenario === "BLOCKED"
                ? "bg-[#f05252] text-white shadow-xs"
                : "text-[#97a0af] hover:text-[#f4f5f7]"
            }`}
          >
            Scenario A: Attack Wire ($14,500)
          </button>
          <button
            onClick={() => setSelectedScenario("APPROVED")}
            className={`px-3 py-1.5 text-xs font-mono rounded-[3px] font-semibold transition-all cursor-pointer ${
              selectedScenario === "APPROVED"
                ? "bg-[#04db7c] text-[#011638] font-bold shadow-xs"
                : "text-[#97a0af] hover:text-[#f4f5f7]"
            }`}
          >
            Scenario B: Organic POS ($42.50)
          </button>
        </div>
      </div>

      {actionFeedback && (
        <div className="p-3 bg-[#04db7c15] border border-[#04db7c44] text-[#04db7c] font-mono text-xs rounded-[4px] flex items-center justify-between">
          <span>✓ {actionFeedback}</span>
        </div>
      )}

      {/* 1. Primary Decision Summary */}
      <DecisionSummary
        transactionId={decision.transactionId}
        customerId={decision.customerId}
        amount={decision.amount}
        currency={decision.currency}
        verdict={decision.verdict}
        riskScore={decision.riskScore}
        confidence={decision.confidence}
        latencyMs={decision.latencyMs}
        modelVersion={decision.modelVersion}
        isDemo={true}
      />

      {/* 2. Middle Grid: Additive Risk Factors & Triggered Rules */}
      <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
        {/* Left 7 cols: Additive Risk Factors */}
        <div className="lg:col-span-7 bg-[#0f172a] border border-[#1c2536] rounded-[4px] p-5 shadow-xs">
          <div className="flex items-center justify-between pb-3 mb-2 border-b border-[#1c2536]">
            <div className="flex items-center gap-2">
              <Layers className="w-4 h-4 text-[#0d94fb]" />
              <h2 className="text-xs font-mono font-bold tracking-wider text-[#f4f5f7] uppercase">
                Additive Factor Attribution (Sum = {decision.riskScore.toFixed(2)})
              </h2>
            </div>
            <span className="text-[10px] font-mono text-[#5e6c84]">Exact Parity</span>
          </div>

          <div className="space-y-1">
            {decision.riskFactors.map((factor, idx) => (
              <RiskFactor key={idx} factor={factor} />
            ))}
          </div>
        </div>

        {/* Right 5 cols: Triggered Rules & Direct Action Controls */}
        <div className="lg:col-span-5 flex flex-col gap-6">
          {/* Triggered Rules Box */}
          <div className="bg-[#0f172a] border border-[#1c2536] rounded-[4px] p-5 flex-1 shadow-xs">
            <div className="flex items-center justify-between pb-3 mb-3 border-b border-[#1c2536]">
              <div className="flex items-center gap-2">
                <FileCode2 className="w-4 h-4 text-[#f59e0b]" />
                <h2 className="text-xs font-mono font-bold tracking-wider text-[#f4f5f7] uppercase">
                  Triggered Policy Rules
                </h2>
              </div>
              <span className="text-[10px] font-mono bg-[#142036] text-[#97a0af] px-1.5 py-0.5 rounded-[2px]">
                Rules Engine
              </span>
            </div>

            {decision.triggeredRules.length === 0 ? (
              <div className="py-6 text-center text-xs text-[#5e6c84] font-mono">
                No policy rule violations triggered for this transaction.
              </div>
            ) : (
              <div className="space-y-2.5">
                {decision.triggeredRules.map((rule, idx) => (
                  <div
                    key={idx}
                    className="p-2.5 bg-[#0a1324] border border-[#1c2536] rounded-[4px] font-mono text-xs"
                  >
                    <div className="flex items-center justify-between gap-2 mb-1">
                      <span className="font-bold text-[#f59e0b]">{rule.id}</span>
                      <span className="text-[10px] bg-[#f0525215] text-[#f05252] px-1.5 py-0.2 rounded-[2px] border border-[#f0525233]">
                        {rule.action}
                      </span>
                    </div>
                    <p className="text-[#f4f5f7] font-semibold text-xs mb-0.5">{rule.name}</p>
                    <p className="text-[11px] text-[#97a0af]">{rule.reason}</p>
                  </div>
                ))}
              </div>
            )}
          </div>

          {/* Quick Analyst Actions */}
          <div className="bg-[#0f172a] border border-[#1c2536] rounded-[4px] p-4 shadow-xs">
            <span className="text-[11px] font-mono font-bold text-[#5e6c84] uppercase block mb-2.5">
              Analyst Authority Actions
            </span>
            <div className="grid grid-cols-2 gap-2">
              <button
                onClick={() => handleAction("CONFIRM_FRAUD_BLOCK")}
                className="px-3 py-2 text-xs font-mono font-bold bg-[#f05252] hover:bg-[#d93b3b] text-white rounded-[4px] active:scale-95 text-center shadow-xs cursor-pointer"
              >
                Confirm Block
              </button>
              <button
                onClick={() => handleAction("OVERRIDE_FALSE_POSITIVE")}
                className="px-3 py-2 text-xs font-mono font-bold bg-[#142036] hover:bg-[#1c2b48] text-[#04db7c] border border-[#04db7c44] rounded-[4px] active:scale-95 text-center cursor-pointer"
              >
                Override
              </button>
            </div>
          </div>
        </div>
      </div>

      {/* 3. Three-Way Explainability Dossier: Observed / Inferred / Recommended */}
      <div>
        <div className="flex items-center gap-2 mb-3">
          <Activity className="w-4 h-4 text-[#c084fc]" />
          <h2 className="text-xs font-mono font-bold tracking-wider text-[#f4f5f7] uppercase">
            Tripartite Explainability Evidence
          </h2>
        </div>
        <EvidenceList
          observed={decision.evidence.observed}
          inferred={decision.evidence.inferred}
          recommended={decision.evidence.recommended}
          onActionConfirm={(act) => handleAction(act)}
        />
      </div>

      {/* 4. Collapsed Technical Details Panel */}
      <div className="bg-[#0f172a] border border-[#1c2536] rounded-[4px] overflow-hidden shadow-xs">
        <button
          onClick={() => setIsTechDetailsOpen(!isTechDetailsOpen)}
          className="w-full px-5 py-3.5 flex items-center justify-between text-left hover:bg-[#142036] transition-colors cursor-pointer"
        >
          <div className="flex items-center gap-2">
            <Terminal className="w-4 h-4 text-[#0d94fb]" />
            <span className="text-xs font-mono font-bold text-[#f4f5f7] uppercase tracking-wider">
              Technical Details & Subsystem Inspection
            </span>
            <span className="text-[10px] font-mono text-[#5e6c84]">
              (ML 25-Feature Vector, Threat Intel Dump, Graph Snapshot, Signed Webhooks, Audit Chain)
            </span>
          </div>
          {isTechDetailsOpen ? (
            <ChevronDown className="w-4 h-4 text-[#5e6c84]" />
          ) : (
            <ChevronRight className="w-4 h-4 text-[#5e6c84]" />
          )}
        </button>

        {isTechDetailsOpen && (
          <div className="p-5 border-t border-[#1c2536] bg-[#070e1c] space-y-6">
            {/* Grid 1: ML Features & Threat Intel Dump */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              {/* ML Features */}
              <div className="p-4 bg-[#0f172a] border border-[#1c2536] rounded-[4px]">
                <div className="flex items-center justify-between pb-2 mb-3 border-b border-[#1c2536]">
                  <span className="text-xs font-mono font-bold text-[#c084fc]">
                    XGBoost 25-Feature Extractor Vector
                  </span>
                  <span className="text-[10px] font-mono text-[#5e6c84]">Subsystem 04</span>
                </div>
                <div className="space-y-1.5 font-mono text-xs">
                  {Object.entries(decision.technicalDetails.mlFeatures).map(([k, v]) => (
                    <div key={k} className="flex items-center justify-between text-[11px]">
                      <span className="text-[#97a0af]">{k}:</span>
                      <span className="text-[#f4f5f7] font-semibold">{String(v)}</span>
                    </div>
                  ))}
                </div>
              </div>

              {/* Threat Intel Signal Dump */}
              <div className="p-4 bg-[#0f172a] border border-[#1c2536] rounded-[4px]">
                <div className="flex items-center justify-between pb-2 mb-3 border-b border-[#1c2536]">
                  <span className="text-xs font-mono font-bold text-[#f59e0b]">
                    Threat Intelligence Telemetry
                  </span>
                  <span className="text-[10px] font-mono text-[#5e6c84]">Subsystem 06</span>
                </div>
                <div className="space-y-1.5 font-mono text-xs">
                  <div className="flex items-center justify-between text-[11px]">
                    <span className="text-[#97a0af]">Egress IP:</span>
                    <span className="text-[#f4f5f7] font-semibold">{decision.technicalDetails.threatIntel.ip}</span>
                  </div>
                  <div className="flex items-center justify-between text-[11px]">
                    <span className="text-[#97a0af]">ASN:</span>
                    <span className="text-[#f4f5f7]">{decision.technicalDetails.threatIntel.asn}</span>
                  </div>
                  <div className="flex items-center justify-between text-[11px]">
                    <span className="text-[#97a0af]">Bulletproof Proxy Flag:</span>
                    <span className={decision.technicalDetails.threatIntel.isProxy ? "text-[#f05252] font-bold" : "text-[#04db7c]"}>
                      {decision.technicalDetails.threatIntel.isProxy ? "TRUE" : "FALSE"}
                    </span>
                  </div>
                  <div className="flex items-center justify-between text-[11px]">
                    <span className="text-[#97a0af]">Haversine Geo Velocity:</span>
                    <span className="text-[#f05252] font-bold">
                      {decision.technicalDetails.threatIntel.geoVelocityKmh} km/h
                    </span>
                  </div>
                  <div className="flex items-center justify-between text-[11px]">
                    <span className="text-[#97a0af]">Distance from Prior Session:</span>
                    <span className="text-[#f4f5f7]">
                      {decision.technicalDetails.threatIntel.distanceKm} km
                    </span>
                  </div>
                </div>
              </div>
            </div>

            {/* Grid 2: Graph Context & Webhook Delivery Record */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              {/* Graph Snapshot */}
              <div className="p-4 bg-[#0f172a] border border-[#1c2536] rounded-[4px]">
                <div className="flex items-center justify-between pb-2 mb-3 border-b border-[#1c2536]">
                  <span className="text-xs font-mono font-bold text-[#f472b6]">
                    Fraud Graph 3-Hop Traversal Snapshot
                  </span>
                  <span className="text-[10px] font-mono text-[#5e6c84]">Subsystem 05</span>
                </div>
                <div className="space-y-1.5 font-mono text-xs">
                  <div className="flex items-center justify-between text-[11px]">
                    <span className="text-[#97a0af]">Root Entity:</span>
                    <span className="text-[#f4f5f7]">{decision.technicalDetails.graphSnapshot.rootId}</span>
                  </div>
                  <div className="flex items-center justify-between text-[11px]">
                    <span className="text-[#97a0af]">Connected 3-Hop Neighbors:</span>
                    <span className="text-[#f4f5f7]">{decision.technicalDetails.graphSnapshot.connectedNodes} nodes</span>
                  </div>
                  <div className="flex items-center justify-between text-[11px]">
                    <span className="text-[#97a0af]">Degree Centrality:</span>
                    <span className="text-[#f05252] font-bold">{decision.technicalDetails.graphSnapshot.degreeCentrality}</span>
                  </div>
                  <div className="flex items-center justify-between text-[11px]">
                    <span className="text-[#97a0af]">Mule Syndicate Cluster:</span>
                    <span className={decision.technicalDetails.graphSnapshot.syndicateClusterDetected ? "text-[#f05252] font-bold" : "text-[#04db7c]"}>
                      {decision.technicalDetails.graphSnapshot.syndicateClusterDetected ? "CONFIRMED (Degree >= 4)" : "NEGATIVE"}
                    </span>
                  </div>
                </div>
              </div>

              {/* Webhook Delivery Record */}
              <div className="p-4 bg-[#0f172a] border border-[#1c2536] rounded-[4px]">
                <div className="flex items-center justify-between pb-2 mb-3 border-b border-[#1c2536]">
                  <span className="text-xs font-mono font-bold text-[#0d94fb]">
                    Webhook Signed Delivery Record
                  </span>
                  <span className="text-[10px] font-mono text-[#5e6c84]">Subsystem 13</span>
                </div>
                <div className="space-y-1.5 font-mono text-xs">
                  <div className="flex items-center justify-between text-[11px]">
                    <span className="text-[#97a0af]">Event ID:</span>
                    <span className="text-[#f4f5f7]">{decision.technicalDetails.webhookDelivery.id}</span>
                  </div>
                  <div className="flex items-center justify-between text-[11px]">
                    <span className="text-[#97a0af]">Destination:</span>
                    <span className="text-[#f4f5f7] truncate max-w-xs">{decision.technicalDetails.webhookDelivery.endpoint}</span>
                  </div>
                  <div className="flex items-center justify-between text-[11px]">
                    <span className="text-[#97a0af]">Status:</span>
                    <span className="text-[#04db7c] font-bold">
                      {decision.technicalDetails.webhookDelivery.status} ({decision.technicalDetails.webhookDelivery.statusCode})
                    </span>
                  </div>
                  <div className="flex items-center justify-between text-[11px]">
                    <span className="text-[#97a0af]">HMAC-SHA256 Sig:</span>
                    <span className="text-[#97a0af] truncate max-w-xs">{decision.technicalDetails.webhookDelivery.signature}</span>
                  </div>
                </div>
              </div>
            </div>

            {/* Raw JSON Request */}
            <div className="p-4 bg-[#0f172a] border border-[#1c2536] rounded-[4px]">
              <div className="flex items-center justify-between pb-2 mb-2 border-b border-[#1c2536]">
                <span className="text-xs font-mono font-bold text-[#5e6c84]">
                  CANONICAL INGRESS PAYLOAD (POST /v1/risk/evaluate)
                </span>
                <span className="text-[10px] font-mono text-[#5e6c84]">SHA-256 Chain: {decision.technicalDetails.auditHashChain.slice(0, 16)}...</span>
              </div>
              <pre className="text-[11px] font-mono text-[#0d94fb] overflow-x-auto p-2 bg-[#070e1c] rounded-[4px]">
                {JSON.stringify(decision.technicalDetails.rawPayload, null, 2)}
              </pre>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
