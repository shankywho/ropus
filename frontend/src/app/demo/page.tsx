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
      <div className="flex flex-wrap items-center justify-between gap-4 pb-4 border-b border-[#1c2b48]">
        <div>
          <h1 className="text-xl font-bold text-[#f4f5f7] tracking-tight">Interactive 5-Minute Investor Demo</h1>
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
        <div className="lg:col-span-4 bg-[#0b1528] border border-[#1c2b48] rounded p-4 flex flex-col justify-between max-h-[640px]">
          <div>
            <div className="flex items-center justify-between pb-3 mb-3 border-b border-[#1c2b48]">
              <div className="flex items-center gap-2">
                <Radio className="w-3.5 h-3.5 text-[#38bdf8] animate-pulse" />
                <span className="text-xs font-mono font-bold text-[#f4f5f7] uppercase tracking-wider">
                  Live Event Stream
                </span>
              </div>
              <span className="text-[10px] font-mono text-[#6b778c]">Kafka Egress</span>
            </div>

            <div className="space-y-2 overflow-y-auto max-h-[500px] pr-1">
              {events.map((evt) => (
                <div
                  key={evt.id}
                  className="p-2.5 bg-[#070e1c] border border-[#1c2b48] rounded text-xs font-mono"
                >
                  <div className="flex items-center justify-between gap-2 mb-1">
                    <span
                      className={`text-[10px] font-bold px-1.5 py-0.2 rounded border ${
                        evt.status === "CRITICAL"
                          ? "bg-[#f0525215] text-[#f05252] border-[#f0525233]"
                          : evt.status === "WARN"
                          ? "bg-[#f59e0b15] text-[#f59e0b] border-[#f59e0b33]"
                          : "bg-[#04db7c15] text-[#04db7c] border-[#04db7c33]"
                      }`}
                    >
                      {evt.type}
                    </span>
                    <span className="text-[10px] text-[#6b778c]">{evt.time}</span>
                  </div>
                  <p className="text-xs text-[#f4f5f7] leading-relaxed">{evt.summary}</p>
                </div>
              ))}
            </div>
          </div>

          <div className="pt-3 mt-3 border-t border-[#1c2b48] text-[10px] font-mono text-[#6b778c] text-center">
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
            <div className="bg-[#0b1528] border border-[#1c2b48] rounded p-4">
              <div className="flex items-center justify-between pb-2 mb-2 border-b border-[#1c2b48]">
                <span className="text-xs font-mono font-bold text-[#38bdf8] uppercase">
                  Additive Factor Sum: {activeDecision.riskScore.toFixed(2)}
                </span>
                <span className="text-[10px] font-mono text-[#6b778c]">Parity</span>
              </div>
              <div className="space-y-1">
                {activeDecision.riskFactors.map((f, i) => (
                  <RiskFactor key={i} factor={f} />
                ))}
              </div>
            </div>

            {/* Rules & Action State */}
            <div className="bg-[#0b1528] border border-[#1c2b48] rounded p-4 flex flex-col justify-between">
              <div>
                <div className="flex items-center justify-between pb-2 mb-2 border-b border-[#1c2b48]">
                  <span className="text-xs font-mono font-bold text-[#f59e0b] uppercase">
                    Rule Invariants
                  </span>
                  <span className="text-[10px] font-mono text-[#6b778c]">Precedence</span>
                </div>

                {activeDecision.triggeredRules.length === 0 ? (
                  <p className="text-xs text-[#6b778c] font-mono py-4">
                    Zero rule violations. Evaluated as benign organic transaction.
                  </p>
                ) : (
                  <div className="space-y-2">
                    {activeDecision.triggeredRules.map((r, i) => (
                      <div key={i} className="p-2 bg-[#070e1c] border border-[#1c2b48] rounded text-xs font-mono">
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
              <div className="pt-3 border-t border-[#1c2b48] mt-3">
                <span className="text-[10px] font-mono text-[#6b778c] block mb-2">ANALYST DECISION:</span>
                {analystActionTaken ? (
                  <div className="p-2 bg-[#04db7c15] border border-[#04db7c33] text-[#04db7c] font-mono text-xs rounded text-center font-bold">
                    ✓ {analystActionTaken} Logged
                  </div>
                ) : (
                  <div className="grid grid-cols-2 gap-2">
                    <button
                      onClick={() => takeAnalystAction("CONFIRM_FRAUD_BLOCK")}
                      className="px-2.5 py-1.5 text-xs font-mono font-bold bg-[#f0525218] hover:bg-[#f0525228] text-[#f05252] border border-[#f0525244] rounded active:scale-95 text-center"
                    >
                      Confirm Block
                    </button>
                    <button
                      onClick={() => takeAnalystAction("OVERRIDE_ALLOW")}
                      className="px-2.5 py-1.5 text-xs font-mono font-bold bg-[#04db7c18] hover:bg-[#04db7c28] text-[#04db7c] border border-[#04db7c44] rounded active:scale-95 text-center"
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
            <span className="text-xs font-mono font-bold text-[#6b778c] uppercase block mb-2">
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
