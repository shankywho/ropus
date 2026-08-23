"use client";

import React from "react";
import { useDemoStore } from "@/lib/demo-store";
import { DemoController } from "@/components/ropus/DemoController";
import { VerdictBadge } from "@/components/ropus/VerdictBadge";
import { DataProvenanceBadge } from "@/components/ropus/DataProvenanceBadge";
import {
  Radio,
} from "lucide-react";

export default function DemoPage() {
  const {
    stage,
    stepIndex,
    activeDecision,
    events,
  } = useDemoStore();

  const timelineEvents = [
    { time: "17:42:01.120", event: "DEVICE_FINGERPRINT_CHANGE", source: "Device Sensor", status: "CRITICAL", desc: "Canvas hash 9f8a... (Linux Emulator v4.2)" },
    { time: "17:42:03.450", event: "GEO_VELOCITY_ANOMALY", source: "Threat Intel", status: "CRITICAL", desc: "Haversine velocity 8,420 km/h (NYC to Limassol)" },
    { time: "17:42:04.100", event: "MONETARY_VOLUME_SPIKE", source: "Feature Store", status: "HIGH", desc: "$14,500.00 wire transfer (+412% over 30d mean)" },
    { time: "17:42:05.280", event: "FRAUD_GRAPH_MATCH", source: "BFS Engine", status: "HIGH", desc: "2-hop link to mule account CY88...9124" },
    { time: "17:42:06.650", event: "ML_SCORE_SYNCHRONIZED", source: "fraud-xgb-25f", status: "CRITICAL", desc: "Model score 0.96 (Beta calibrated v2.5)" },
    { time: "17:42:07.142", event: "DETERMINISTIC_POLICY_BLOCK", source: "Rules AST", status: "BLOCKED", desc: "Rule RULE-IMPOSSIBLE-TRAVEL executed in 1.42ms" },
    { time: "17:42:08.500", event: "CASE_DISPATCHED_P0", source: "Case Manager", status: "PENDING", desc: "Case #CASE-88419 assigned to Lead Analyst" },
    { time: "17:42:10.020", event: "WEBHOOK_DELIVERED", source: "Transactional Outbox", status: "ACK", desc: "HMAC-SHA256 payload delivered to merchant endpoint" },
  ];

  return (
    <div className="space-y-5 font-mono">
      {/* 1. Top Header */}
      <div className="flex flex-wrap items-center justify-between gap-4 pb-3.5 border-b border-[#1c2536]">
        <div>
          <div className="flex items-center gap-2.5">
            <h1 className="text-lg font-bold text-[#f4f5f7] tracking-tight uppercase">
              7-Stage Adversarial Scenario Walkthrough
            </h1>
            <DataProvenanceBadge type="DEMO_MODE" sublabel="Deterministic Playback" />
          </div>
          <p className="text-xs text-[#5e6c84] mt-1 font-sans">
            Automated simulation of an Account Takeover (ATO) and liquidity drain attack against the ROPUS risk engine
          </p>
        </div>
      </div>

      {/* 2. Interactive Playback Controller */}
      <DemoController />

      {/* 3. Active Stage Scenario Hero Strip */}
      <div className="p-4 bg-[#0f172a] border border-[#0d94fb44] rounded-[4px] space-y-2.5 shadow-xs">
        <div className="flex items-center justify-between border-b border-[#1c2536] pb-2 text-xs">
          <div className="flex items-center gap-2">
            <span className="px-2 py-0.5 rounded-[2px] bg-[#0d94fb15] text-[#0d94fb] border border-[#0d94fb33] font-bold text-[10px]">
              STEP {stepIndex}: {stage}
            </span>
            <span className="text-[#f4f5f7] font-bold text-xs">{activeDecision.transactionId}</span>
          </div>
          <div className="flex items-center gap-2">
            <span className="text-[#5e6c84] text-[10px]">VERDICT:</span>
            <VerdictBadge verdict={activeDecision.verdict} />
          </div>
        </div>

        <div className="grid grid-cols-2 sm:grid-cols-4 gap-3 pt-1 text-xs">
          <div className="p-2.5 bg-[#0a1324] rounded-[3px] border border-[#1c2536]">
            <span className="text-[#5e6c84] text-[10px] block">AMOUNT</span>
            <span className="text-[#f4f5f7] font-bold">
              {activeDecision.currency} {activeDecision.amount.toLocaleString()}
            </span>
          </div>
          <div className="p-2.5 bg-[#0a1324] rounded-[3px] border border-[#1c2536]">
            <span className="text-[#5e6c84] text-[10px] block">RISK SCORE</span>
            <span
              className={`font-bold ${
                activeDecision.riskScore >= 0.35 ? "text-[#f05252]" : "text-[#04db7c]"
              }`}
            >
              {activeDecision.riskScore.toFixed(2)}
            </span>
          </div>
          <div className="p-2.5 bg-[#0a1324] rounded-[3px] border border-[#1c2536]">
            <span className="text-[#5e6c84] text-[10px] block">LATENCY</span>
            <span className="text-[#04db7c] font-bold">{activeDecision.latencyMs.toFixed(2)} ms</span>
          </div>
          <div className="p-2.5 bg-[#0a1324] rounded-[3px] border border-[#1c2536]">
            <span className="text-[#5e6c84] text-[10px] block">MODEL</span>
            <span className="text-[#0d94fb] font-bold">{activeDecision.modelVersion}</span>
          </div>
        </div>
      </div>

      {/* 4. Cinematic Operational Event Timeline */}
      <div className="bg-[#0f172a] border border-[#1c2536] rounded-[4px] p-4 text-xs space-y-3">
        <div className="flex items-center justify-between pb-2.5 border-b border-[#1c2536]">
          <div className="flex items-center gap-2">
            <Radio className="w-4 h-4 text-[#0d94fb] animate-pulse" />
            <h2 className="font-bold uppercase tracking-wider text-[#f4f5f7] text-xs">
              Synchronous Execution Stream
            </h2>
          </div>
          <span className="text-[#5e6c84] text-[10px]">Deterministic Event Log ({events.length} Events)</span>
        </div>

        <div className="overflow-x-auto">
          <table className="w-full text-left">
            <thead>
              <tr className="border-b border-[#1c2536] text-[#5e6c84] text-[11px]">
                <th className="pb-2 font-medium">TIMESTAMP</th>
                <th className="pb-2 font-medium">EVENT TYPE</th>
                <th className="pb-2 font-medium">SUBSYSTEM</th>
                <th className="pb-2 font-medium">DESCRIPTION</th>
                <th className="pb-2 font-medium text-right">STATUS</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-[#1c2536]">
              {timelineEvents.map((evt, idx) => (
                <tr key={idx} className="hover:bg-[#142036] transition-colors">
                  <td className="py-2.5 text-[#5e6c84]">{evt.time}</td>
                  <td className="py-2.5 font-bold text-[#f4f5f7]">{evt.event}</td>
                  <td className="py-2.5">
                    <span className="px-1.5 py-0.5 rounded-[2px] bg-[#0a1324] border border-[#1c2536] text-[10px] text-[#0d94fb]">
                      {evt.source}
                    </span>
                  </td>
                  <td className="py-2.5 text-[#97a0af] font-sans">{evt.desc}</td>
                  <td className="py-2.5 text-right">
                    <span
                      className={`text-[10px] font-bold px-1.5 py-0.5 rounded-[2px] border ${
                        evt.status === "CRITICAL" || evt.status === "BLOCKED"
                          ? "bg-[#f0525215] text-[#f05252] border-[#f0525233]"
                          : evt.status === "HIGH" || evt.status === "PENDING"
                          ? "bg-[#f59e0b15] text-[#f59e0b] border-[#f59e0b33]"
                          : "bg-[#04db7c15] text-[#04db7c] border-[#04db7c33]"
                      }`}
                    >
                      {evt.status}
                    </span>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}
