"use client";

import React from "react";
import { DemoController } from "@/components/ropus/DemoController";
import { DecisionSummary } from "@/components/ropus/DecisionSummary";
import { RiskFactor } from "@/components/ropus/RiskFactor";
import { EvidenceList } from "@/components/ropus/EvidenceList";
import { useDemoStore } from "@/lib/demo-store";
import {
  Radio,
} from "lucide-react";

export default function DemoPage() {
  const { stage, stepIndex, activeDecision, events, takeAnalystAction, analystActionTaken } =
    useDemoStore();

  return (
    <div className="max-w-6xl mx-auto space-y-6">
      {/* Top Header */}
      <div className="flex flex-wrap items-center justify-between gap-4 pb-4 border-b border-[#1c2536]">
        <div>
          <div className="flex items-center gap-2">
            <h1 className="text-xl font-bold text-[#f4f5f7] tracking-tight">Interactive 5-Minute Investor Demo</h1>
            <span className="text-[10px] font-mono bg-[#f59e0b15] text-[#f59e0b] px-1.5 py-0.5 rounded-[2px] border border-[#f59e0b33]">
              SCRIPTED STATE MACHINE
            </span>
          </div>
          <p className="text-xs text-[#97a0af] font-mono mt-0.5">
            Scripted end-to-end attack simulation: Organic baseline → ATO → Graph syndicate → 0.96 Block → Dossier → Governance
          </p>
        </div>

        <div className="flex items-center gap-2">
          <span className="w-2 h-2 rounded-full bg-[#04db7c] animate-pulse" />
          <span className="text-xs font-mono text-[#04db7c] font-bold">
            STAGE: {stage} ({stepIndex}/9)
          </span>
        </div>
      </div>

      {/* 1. Playback Controller Bar */}
      <DemoController />

      {/* 2. Main Demo Layout: Left Stream Ticker / Right Live Decision State */}
      <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
        {/* Left 4 Cols: Live Event Stream Ticker (Kafka Event Proof) */}
        <div className="lg:col-span-4 bg-[#0f172a] border border-[#1c2536] rounded-[4px] p-4 flex flex-col justify-between max-h-[640px] shadow-xs">
          <div>
            <div className="flex items-center justify-between pb-3 mb-3 border-b border-[#1c2536]">
              <div className="flex items-center gap-2">
                <Radio className="w-3.5 h-3.5 text-[#0d94fb] animate-pulse" />
                <span className="text-xs font-mono font-bold text-[#f4f5f7] uppercase tracking-wider">
                  Live Event Stream
                </span>
              </div>
              <span className="text-[10px] font-mono text-[#5e6c84]">Kafka Egress</span>
            </div>

            <div className="space-y-2 overflow-y-auto max-h-[500px] pr-1">
              {events.map((evt) => (
                <div
                  key={evt.id}
                  className="p-2.5 bg-[#070e1c] border border-[#1c2536] rounded-[4px] text-xs font-mono"
                >
                  <div className="flex items-center justify-between gap-2 mb-1">
                    <span
                      className={`text-[10px] font-bold px-1.5 py-0.2 rounded-[2px] border ${
                        evt.status === "CRITICAL"
                          ? "bg-[#f0525215] text-[#f05252] border-[#f0525233]"
                          : evt.status === "WARN"
                          ? "bg-[#f59e0b15] text-[#f59e0b] border-[#f59e0b33]"
                          : "bg-[#04db7c15] text-[#04db7c] border-[#04db7c33]"
                      }`}
                    >
                      {evt.type}
                    </span>
                    <span className="text-[10px] text-[#5e6c84]">{evt.time}</span>
                  </div>
                  <p className="text-xs text-[#f4f5f7] leading-relaxed">{evt.summary}</p>
                </div>
              ))}
            </div>
          </div>

          <div className="pt-3 mt-3 border-t border-[#1c2536] text-[10px] font-mono text-[#5e6c84] text-center">
            Sub-10ms Stream Ingestion • Idempotent Dispatch
          </div>
        </div>

        {/* Right 8 Cols: Real-Time Evaluated Decision State */}
        <div className="lg:col-span-8 space-y-6">
          {/* Active Decision Top Card */}
          <DecisionSummary
            transactionId={activeDecision.transactionId}
            customerId={activeDecision.customerId}
            amount={activeDecision.amount}
            currency={activeDecision.currency}
            verdict={activeDecision.verdict}
            riskScore={activeDecision.riskScore}
            confidence={activeDecision.confidence}
            latencyMs={activeDecision.latencyMs}
            modelVersion={activeDecision.modelVersion}
            isDemo={true}
          />

          {/* Additive Factors and Rules */}
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {/* Additive Factors */}
            <div className="bg-[#0f172a] border border-[#1c2536] rounded-[4px] p-4 shadow-xs">
              <div className="flex items-center justify-between pb-2 mb-2 border-b border-[#1c2536]">
                <span className="text-xs font-mono font-bold text-[#0d94fb] uppercase">
                  Additive Factor Sum: {activeDecision.riskScore.toFixed(2)}
                </span>
                <span className="text-[10px] font-mono text-[#5e6c84]">Parity</span>
              </div>
              <div className="space-y-1">
                {activeDecision.riskFactors.map((f, i) => (
                  <RiskFactor key={i} factor={f} />
                ))}
              </div>
            </div>

            {/* Rules & Action State */}
            <div className="bg-[#0f172a] border border-[#1c2536] rounded-[4px] p-4 flex flex-col justify-between shadow-xs">
              <div>
                <div className="flex items-center justify-between pb-2 mb-2 border-b border-[#1c2536]">
                  <span className="text-xs font-mono font-bold text-[#f59e0b] uppercase">
                    Rule Invariants
                  </span>
                  <span className="text-[10px] font-mono text-[#5e6c84]">Precedence</span>
                </div>

                {activeDecision.triggeredRules.length === 0 ? (
                  <p className="text-xs text-[#5e6c84] font-mono py-4">
                    Zero rule violations. Evaluated as benign organic transaction.
                  </p>
                ) : (
                  <div className="space-y-2">
                    {activeDecision.triggeredRules.map((r, i) => (
                      <div key={i} className="p-2 bg-[#0a1324] border border-[#1c2536] rounded-[4px] text-xs font-mono">
                        <div className="flex items-center justify-between text-[#f59e0b] font-bold text-[11px]">
                          <span>{r.id}</span>
                          <span>{r.action}</span>
                        </div>
                        <p className="text-[11px] text-[#97a0af] mt-0.5">{r.name}</p>
                      </div>
                    ))}
                  </div>
                )}
              </div>

              {/* Interactive Action Trigger */}
              <div className="pt-3 border-t border-[#1c2536] mt-3">
                <span className="text-[10px] font-mono text-[#5e6c84] block mb-2">ANALYST DECISION:</span>
                {analystActionTaken ? (
                  <div className="p-2 bg-[#04db7c15] border border-[#04db7c33] text-[#04db7c] font-mono text-xs rounded-[4px] text-center font-bold">
                    ✓ {analystActionTaken} Logged
                  </div>
                ) : (
                  <div className="grid grid-cols-2 gap-2">
                    <button
                      onClick={() => takeAnalystAction("CONFIRM_FRAUD_BLOCK")}
                      className="px-2.5 py-1.5 text-xs font-mono font-bold bg-[#f05252] hover:bg-[#d93b3b] text-white rounded-[4px] active:scale-95 text-center shadow-xs cursor-pointer"
                    >
                      Confirm Block
                    </button>
                    <button
                      onClick={() => takeAnalystAction("OVERRIDE_ALLOW")}
                      className="px-2.5 py-1.5 text-xs font-mono font-bold bg-[#142036] hover:bg-[#1c2b48] text-[#04db7c] border border-[#04db7c44] rounded-[4px] active:scale-95 text-center cursor-pointer"
                    >
                      Override
                    </button>
                  </div>
                )}
              </div>
            </div>
          </div>

          {/* Tripartite Evidence List */}
          <div>
            <span className="text-xs font-mono font-bold text-[#5e6c84] uppercase block mb-2">
              Tripartite Dossier (Observed / Inferred / Recommended)
            </span>
            <EvidenceList
              observed={activeDecision.evidence.observed}
              inferred={activeDecision.evidence.inferred}
              recommended={activeDecision.evidence.recommended}
              onActionConfirm={(act) => takeAnalystAction(act)}
            />
          </div>
        </div>
      </div>
    </div>
  );
}
