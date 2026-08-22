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
      <div className="flex flex-wrap items-center justify-between gap-4 pb-4 border-b border-[#1c2b48]">
        <div>
          <h1 className="text-xl font-bold text-[#f4f5f7] tracking-tight">Risk Decision Intelligence</h1>
          <p className="text-xs text-[#97a0af] font-mono mt-0.5">
            Single-loop synchronous evaluation pipeline & additive factor attribution
          </p>
        </div>

        {/* Toggle between canonical attack scenario and organic benign purchase */}
        <div className="flex items-center gap-2 bg-[#0b1528] p-1 border border-[#1c2b48] rounded">
          <button
            onClick={() => setSelectedScenario("BLOCKED")}
            className={`px-3 py-1 text-xs font-mono rounded font-semibold transition-all ${
              selectedScenario === "BLOCKED"
                ? "bg-[#f0525222] text-[#f05252] border border-[#f0525244]"
                : "text-[#97a0af] hover:text-[#f4f5f7]"
            }`}
          >
            Scenario A: Attack Wire ($14,500)
          </button>
          <button
            onClick={() => setSelectedScenario("APPROVED")}
            className={`px-3 py-1 text-xs font-mono rounded font-semibold transition-all ${
              selectedScenario === "APPROVED"
                ? "bg-[#04db7c22] text-[#04db7c] border border-[#04db7c44]"
                : "text-[#97a0af] hover:text-[#f4f5f7]"
            }`}
          >
            Scenario B: Organic POS ($42.50)
          </button>
        </div>
      </div>

      {actionFeedback && (
        <div className="p-3 bg-[#04db7c15] border border-[#04db7c44] text-[#04db7c] font-mono text-xs rounded flex items-center justify-between">
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
        <div className="lg:col-span-7 bg-[#0b1528] border border-[#1c2b48] rounded p-5">
          <div className="flex items-center justify-between pb-3 mb-2 border-b border-[#1c2b48]">
            <div className="flex items-center gap-2">
              <Layers className="w-4 h-4 text-[#38bdf8]" />
              <h2 className="text-xs font-mono font-bold tracking-wider text-[#f4f5f7] uppercase">
                Additive Factor Attribution (Sum = {decision.riskScore.toFixed(2)})
              </h2>
            </div>
            <span className="text-[10px] font-mono text-[#6b778c]">Exact Parity</span>
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
          <div className="bg-[#0b1528] border border-[#1c2b48] rounded p-5 flex-1">
            <div className="flex items-center justify-between pb-3 mb-3 border-b border-[#1c2b48]">
              <div className="flex items-center gap-2">
                <FileCode2 className="w-4 h-4 text-[#f59e0b]" />
                <h2 className="text-xs font-mono font-bold tracking-wider text-[#f4f5f7] uppercase">
                  Triggered Policy Rules
                </h2>
              </div>
              <span className="text-[10px] font-mono bg-[#14223d] text-[#97a0af] px-1.5 py-0.5 rounded">
                Rules Engine
              </span>
            </div>

            {decision.triggeredRules.length === 0 ? (
              <div className="py-6 text-center text-xs text-[#6b778c] font-mono">
                No policy rule violations triggered for this transaction.
              </div>
            ) : (
              <div className="space-y-2.5">
                {decision.triggeredRules.map((rule, idx) => (
                  <div
                    key={idx}
                    className="p-2.5 bg-[#070e1c] border border-[#1c2b48] rounded font-mono text-xs"
                  >
                    <div className="flex items-center justify-between gap-2 mb-1">
                      <span className="font-bold text-[#f59e0b]">{rule.id}</span>
                      <span className="text-[10px] bg-[#f0525215] text-[#f05252] px-1.5 py-0.2 rounded border border-[#f0525233]">
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
          <div className="bg-[#0b1528] border border-[#1c2b48] rounded p-4">
            <span className="text-[11px] font-mono font-bold text-[#6b778c] uppercase block mb-2.5">
              Analyst Authority Actions
            </span>
            <div className="grid grid-cols-2 gap-2">
              <button
                onClick={() => handleAction("CONFIRM_FRAUD_BLOCK")}
                className="px-3 py-2 text-xs font-mono font-bold bg-[#f0525218] hover:bg-[#f0525228] text-[#f05252] border border-[#f0525244] rounded active:scale-95 text-center"
              >
                Confirm Block
              </button>
              <button
                onClick={() => handleAction("OVERRIDE_FALSE_POSITIVE")}
                className="px-3 py-2 text-xs font-mono font-bold bg-[#04db7c18] hover:bg-[#04db7c28] text-[#04db7c] border border-[#04db7c44] rounded active:scale-95 text-center"
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

      {/* 4. Collapsed by Default Technical Details Panel */}
      <div className="bg-[#0b1528] border border-[#1c2b48] rounded overflow-hidden">
        <button
          onClick={() => setIsTechDetailsOpen(!isTechDetailsOpen)}
          className="w-full px-5 py-3.5 flex items-center justify-between text-left hover:bg-[#0f1c34] transition-colors"
        >
          <div className="flex items-center gap-2">
            <Terminal className="w-4 h-4 text-[#6b778c]" />
            <span className="text-xs font-mono font-bold text-[#f4f5f7] uppercase tracking-wider">
              Technical Details & Subsystem Inspection
            </span>
            <span className="text-[10px] font-mono text-[#6b778c]">
              (ML Features, Threat Intel Dump, Graph Snapshot, Webhooks, Audit Chain)
            </span>
          </div>
          {isTechDetailsOpen ? (
            <ChevronDown className="w-4 h-4 text-[#6b778c]" />
          ) : (
            <ChevronRight className="w-4 h-4 text-[#6b778c]" />
          )}
        </button>

        {isTechDetailsOpen && (
          <div className="p-5 border-t border-[#1c2b48] bg-[#070e1c] space-y-6">
            {/* Grid 1: ML Features & Threat Intel Dump */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              {/* ML Features */}
              <div className="p-4 bg-[#0b1528] border border-[#1c2b48] rounded">
                <div className="flex items-center justify-between pb-2 mb-3 border-b border-[#1c2b48]">
                  <span className="text-xs font-mono font-bold text-[#c084fc]">
                    XGBoost 25-Feature Extractor Vector
                  </span>
                  <span className="text-[10px] font-mono text-[#6b778c]">Subsystem 04</span>
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
              <div className="p-4 bg-[#0b1528] border border-[#1c2b48] rounded">
                <div className="flex items-center justify-between pb-2 mb-3 border-b border-[#1c2b48]">
                  <span className="text-xs font-mono font-bold text-[#f59e0b]">
                    Threat Intelligence Telemetry
                  </span>
                  <span className="text-[10px] font-mono text-[#6b778c]">Subsystem 06</span>
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
              <div className="p-4 bg-[#0b1528] border border-[#1c2b48] rounded">
                <div className="flex items-center justify-between pb-2 mb-3 border-b border-[#1c2b48]">
                  <span className="text-xs font-mono font-bold text-[#f472b6]">
                    Fraud Graph 3-Hop Traversal Snapshot
                  </span>
                  <span className="text-[10px] font-mono text-[#6b778c]">Subsystem 05</span>
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
              <div className="p-4 bg-[#0b1528] border border-[#1c2b48] rounded">
                <div className="flex items-center justify-between pb-2 mb-3 border-b border-[#1c2b48]">
                  <span className="text-xs font-mono font-bold text-[#38bdf8]">
                    Webhook Signed Delivery Record
                  </span>
                  <span className="text-[10px] font-mono text-[#6b778c]">Subsystem 13</span>
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
            <div className="p-4 bg-[#0b1528] border border-[#1c2b48] rounded">
              <div className="flex items-center justify-between pb-2 mb-2 border-b border-[#1c2b48]">
                <span className="text-xs font-mono font-bold text-[#6b778c]">
                  CANONICAL INGRESS PAYLOAD (POST /v1/risk/evaluate)
                </span>
                <span className="text-[10px] font-mono text-[#6b778c]">SHA-256 Chain: {decision.technicalDetails.auditHashChain.slice(0, 16)}...</span>
              </div>
              <pre className="text-[11px] font-mono text-[#38bdf8] overflow-x-auto p-2 bg-[#070e1c] rounded">
                {JSON.stringify(decision.technicalDetails.rawPayload, null, 2)}
              </pre>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
