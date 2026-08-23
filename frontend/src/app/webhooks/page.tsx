"use client";

import React, { useState } from "react";
import {
  Webhook,
  Send,
} from "lucide-react";

interface WebhookDelivery {
  id: string;
  endpoint: string;
  eventType: string;
  statusCode: number;
  latencyMs: number;
  signature: string;
  sentAt: string;
}

const SEED_DELIVERIES: WebhookDelivery[] = [
  {
    id: "wh_evt_88419",
    endpoint: "https://api.merchant.com/v1/risk-callbacks",
    eventType: "risk.decision.block",
    statusCode: 200,
    latencyMs: 42.1,
    signature: "t=1755884521,v1=9f8a1e9c7a2b5d8e3f1c6d39a04b12ef...",
    sentAt: "2026-08-22 17:42:02",
  },
  {
    id: "wh_evt_88418",
    endpoint: "https://api.merchant.com/v1/risk-callbacks",
    eventType: "risk.decision.approve",
    statusCode: 200,
    latencyMs: 38.5,
    signature: "t=1755884400,v1=7a1b2c3d4e5f60718293a4b5c6d7e8f9...",
    sentAt: "2026-08-22 17:40:00",
  },
];

export default function WebhooksPage() {
  const [deliveries] = useState<WebhookDelivery[]>(SEED_DELIVERIES);
  const [testPayload, setTestPayload] = useState(
    JSON.stringify(
      {
        event_type: "risk.decision.block",
        decision_id: "dec_9f8a1e9c",
        transaction_id: "tx_order_88419",
        risk_score: 0.96,
        verdict: "BLOCK",
      },
      null,
      2
    )
  );
  const [testResult, setTestResult] = useState<string | null>(null);

  const handleTestDispatch = () => {
    setTestResult("HTTP 200 OK — Signature verified with HMAC-SHA256 secret whsec_*** (Latency: 41ms)");
    setTimeout(() => setTestResult(null), 5000);
  };

  return (
    <div className="max-w-6xl mx-auto space-y-6">
      {/* Top Header */}
      <div className="flex flex-wrap items-center justify-between gap-4 pb-4 border-b border-[#1c2536]">
        <div>
          <div className="flex items-center gap-2">
            <h1 className="text-xl font-bold text-[#f4f5f7] tracking-tight">Webhooks & Outbox Egress</h1>
            <span className="text-[10px] font-mono bg-[#0d94fb15] text-[#0d94fb] px-1.5 py-0.5 rounded-[2px] border border-[#0d94fb33]">
              SUBSYSTEM 13
            </span>
          </div>
          <p className="text-xs text-[#97a0af] font-mono mt-0.5">
            Idempotent transactional outbox, HMAC-SHA256 signature verification & delivery retry queues
          </p>
        </div>

        <div className="flex items-center gap-2">
          <span className="px-2.5 py-1 text-xs font-mono bg-[#0f172a] border border-[#1c2536] text-[#04db7c] rounded-[4px] font-bold shadow-xs">
            Egress: 99.98% Success • HMAC: SHA-256
          </span>
        </div>
      </div>

      {testResult && (
        <div className="p-3 bg-[#04db7c15] border border-[#04db7c44] text-[#04db7c] font-mono text-xs rounded-[4px]">
          ✓ {testResult}
        </div>
      )}

      {/* Main Grid: Left Deliveries / Right Test Console */}
      <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
        {/* Left 7 Cols: Recent Webhook Deliveries Table */}
        <div className="lg:col-span-7 bg-[#0f172a] border border-[#1c2536] rounded-[4px] p-5 shadow-xs">
          <div className="flex items-center justify-between pb-3 mb-4 border-b border-[#1c2536]">
            <div className="flex items-center gap-2">
              <Webhook className="w-4 h-4 text-[#0d94fb]" />
              <h2 className="text-xs font-mono font-bold text-[#f4f5f7] uppercase tracking-wider">
                Recent Deliveries
              </h2>
            </div>
            <span className="text-[10px] font-mono text-[#5e6c84]">Transactional Outbox (Postgres)</span>
          </div>

          <div className="space-y-3">
            {deliveries.map((del) => (
              <div key={del.id} className="p-3 bg-[#0a1324] border border-[#1c2536] rounded-[4px] font-mono text-xs space-y-1.5">
                <div className="flex items-center justify-between">
                  <span className="text-[#0d94fb] font-bold">{del.eventType}</span>
                  <span className="text-[#04db7c] font-bold">{del.statusCode} OK</span>
                </div>
                <div className="text-[11px] text-[#97a0af] truncate">{del.endpoint}</div>
                <div className="flex items-center justify-between text-[10px] text-[#5e6c84] pt-1 border-t border-[#1c2536]">
                  <span>Latency: {del.latencyMs}ms</span>
                  <span>{del.sentAt}</span>
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Right 5 Cols: Interactive Test Dispatch Console */}
        <div className="lg:col-span-5 bg-[#0f172a] border border-[#1c2536] rounded-[4px] p-5 shadow-xs flex flex-col justify-between">
          <div>
            <div className="flex items-center justify-between pb-3 mb-3 border-b border-[#1c2536]">
              <span className="text-xs font-mono font-bold text-[#f4f5f7] uppercase tracking-wider">
                Test Webhook Dispatch
              </span>
              <span className="text-[10px] font-mono text-[#5e6c84]">HMAC Testing</span>
            </div>

            <div className="space-y-3">
              <div>
                <label className="text-[11px] font-mono text-[#5e6c84] block mb-1">TARGET URL</label>
                <input
                  type="text"
                  defaultValue="https://api.merchant.com/v1/risk-callbacks"
                  className="w-full px-3 py-1.5 bg-[#070e1c] border border-[#1c2536] rounded-[4px] text-xs font-mono text-[#f4f5f7] focus:outline-none focus:border-[#0d94fb]"
                />
              </div>

              <div>
                <label className="text-[11px] font-mono text-[#5e6c84] block mb-1">JSON PAYLOAD</label>
                <textarea
                  rows={7}
                  value={testPayload}
                  onChange={(e) => setTestPayload(e.target.value)}
                  className="w-full p-2.5 bg-[#070e1c] border border-[#1c2536] rounded-[4px] text-xs font-mono text-[#0d94fb] focus:outline-none focus:border-[#0d94fb]"
                />
              </div>
            </div>
          </div>

          <button
            onClick={handleTestDispatch}
            className="w-full mt-4 py-2 bg-[#0d94fb] hover:bg-[#0b82dc] text-white font-mono text-xs font-bold rounded-[4px] active:scale-95 shadow-xs cursor-pointer flex items-center justify-center gap-1.5"
          >
            <Send className="w-3.5 h-3.5" /> Send Test Webhook & Verify HMAC
          </button>
        </div>
      </div>
    </div>
  );
}
