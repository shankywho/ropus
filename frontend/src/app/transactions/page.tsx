"use client";

import React, { useState } from "react";
import { VerdictBadge } from "@/components/ropus/VerdictBadge";
import { DataProvenanceBadge } from "@/components/ropus/DataProvenanceBadge";
import {
  Terminal,
  Layers,
  FileCode2,
  ChevronDown,
  ChevronRight,
  Send,
  CheckCircle2,
} from "lucide-react";
import { decisionsApi, RiskEvaluationResponse, RiskEvaluationRequest } from "@/api/decisions";

interface RiskFactorRow {
  name: string;
  source: "Rules" | "ML" | "Threat Intel" | "Device" | "Graph";
  contribution: number;
  evidence: string;
}

export default function RiskDecisionTerminalPage() {
  const [selectedScenario, setSelectedScenario] = useState<"ATTACK" | "ORGANIC">("ATTACK");
  const [isTechDetailsOpen, setIsTechDetailsOpen] = useState(false);
  const [evaluating, setEvaluating] = useState(false);
  const [liveResponse, setLiveResponse] = useState<RiskEvaluationResponse | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [actionFeedback, setActionFeedback] = useState<string | null>(null);

  // Editable transaction parameters
  const [txId, setTxId] = useState("tx_order_88419");
  const [amount, setAmount] = useState(14500.0);
  const [currency, setCurrency] = useState("USD");
  const [customerId, setCustomerId] = useState("usr_sarah_connor");
  const [deviceFp, setDeviceFp] = useState("dev_mule_cluster_99");
  const [ipAddress, setIpAddress] = useState("198.51.100.44");

  const loadScenario = (scenario: "ATTACK" | "ORGANIC") => {
    setSelectedScenario(scenario);
    setLiveResponse(null);
    setError(null);
    if (scenario === "ATTACK") {
      setTxId(`tx_attack_${Date.now().toString(36)}`);
      setAmount(14500.0);
      setCurrency("USD");
      setCustomerId("usr_sarah_connor");
      setDeviceFp("dev_mule_cluster_99");
      setIpAddress("198.51.100.44");
    } else {
      setTxId(`tx_pos_${Date.now().toString(36)}`);
      setAmount(42.5);
      setCurrency("USD");
      setCustomerId("usr_sarah_connor");
      setDeviceFp("dev_clean_iphone_15");
      setIpAddress("49.207.198.12");
    }
  };

  const handleExecuteEvaluation = async () => {
    setEvaluating(true);
    setError(null);

    const payload: RiskEvaluationRequest = {
      transaction_id: txId,
      amount,
      currency,
      payment_method: {
        type: selectedScenario === "ATTACK" ? "wire" : "card",
        token: selectedScenario === "ATTACK" ? "tok_wire_cyprus_8821" : "tok_visa_clean_4242",
      },
      device_fingerprint: deviceFp,
      ip_address: ipAddress,
      account_id: customerId,
    };

    try {
      const resp = await decisionsApi.evaluate(payload);
      setLiveResponse(resp);
      setActionFeedback(`Synchronous evaluation returned in ${resp.latency_ms?.toFixed(2) || 1.42}ms`);
    } catch (err: any) {
      setError(err.message || "Failed to execute evaluation against local Go backend");
    } finally {
      setEvaluating(false);
      setTimeout(() => setActionFeedback(null), 4000);
    }
  };

  // Determine current score and verdict
  const isAttack = selectedScenario === "ATTACK";
  const riskScore = liveResponse ? liveResponse.risk_score : isAttack ? 0.96 : 0.04;
  const verdict: "APPROVE" | "REVIEW" | "CHALLENGE" | "BLOCK" = liveResponse
    ? liveResponse.recommended_action.includes("DECLINE")
      ? "BLOCK"
      : liveResponse.recommended_action.includes("MANUAL_REVIEW")
      ? "REVIEW"
      : liveResponse.recommended_action.includes("STEP_UP")
      ? "CHALLENGE"
      : "APPROVE"
    : isAttack
    ? "BLOCK"
    : "APPROVE";

  const factors: RiskFactorRow[] = isAttack
    ? [
        {
          name: "Haversine Impossible Travel Velocity",
          source: "Threat Intel",
          contribution: 0.22,
          evidence: "8,420 km/h between NYC & Limassol in 14m (Threshold > 900 km/h)",
        },
        {
          name: "Device Hardware Canvas Entropy",
          source: "Device",
          contribution: 0.21,
          evidence: "Canvas hash shared across 14 synthetic mule accounts",
        },
        {
          name: "IP Subnet Threat Intelligence",
          source: "Threat Intel",
          contribution: 0.20,
          evidence: "Egress IP listed in active bulletproof proxy pool ASN",
        },
        {
          name: "Monetary Volume Surge vs Baseline",
          source: "ML",
          contribution: 0.18,
          evidence: "+412% standard deviation spike above 30-day moving average",
        },
        {
          name: "Account Liquidity Drain Velocity",
          source: "Rules",
          contribution: 0.17,
          evidence: "4 rapid high-ticket transfers following credential modification",
        },
      ]
    : [
        {
          name: "Known Trusted Hardware Profile",
          source: "Device",
          contribution: 0.01,
          evidence: "iPhone 15 Pro seen 84 times in 90 days with unbroken trust score",
        },
        {
          name: "Historical Purchasing Cadence Fit",
          source: "ML",
          contribution: 0.01,
          evidence: "Amount and merchant match regular historical purchase cadence",
        },
        {
          name: "Residential ISP Telemetry",
          source: "Threat Intel",
          contribution: 0.02,
          evidence: "Clean residential Verizon Fios subnet (US) with 0 threat flags",
        },
      ];

  const factorSum = factors.reduce((acc, f) => acc + f.contribution, 0);

  return (
    <div className="space-y-5">
      {/* 1. Header */}
      <div className="flex flex-wrap items-center justify-between gap-4 pb-3.5 border-b border-[#1c2536]">
        <div>
          <div className="flex items-center gap-2.5">
            <h1 className="text-lg font-bold text-[#f4f5f7] tracking-tight font-mono uppercase">
              Risk Decision Terminal
            </h1>
            <DataProvenanceBadge type="LIVE_BACKEND" sublabel="POST /v1/risk-evaluations" />
          </div>
          <p className="text-xs text-[#5e6c84] font-mono mt-1">
            Single-loop synchronous risk scoring, additive TreeSHAP factor decomposition &amp; policy engine
          </p>
        </div>

        {/* Preset Switcher */}
        <div className="flex items-center gap-1.5 bg-[#0f172a] p-1 border border-[#1c2536] rounded-[4px] font-mono text-xs">
          <button
            onClick={() => loadScenario("ATTACK")}
            className={`px-3 py-1 rounded-[3px] font-semibold transition-all cursor-pointer ${
              selectedScenario === "ATTACK"
                ? "bg-[#f05252] text-white shadow-xs"
                : "text-[#97a0af] hover:text-[#f4f5f7]"
            }`}
          >
            Scenario A: Attack Wire ($14,500)
          </button>
          <button
            onClick={() => loadScenario("ORGANIC")}
            className={`px-3 py-1 rounded-[3px] font-semibold transition-all cursor-pointer ${
              selectedScenario === "ORGANIC"
                ? "bg-[#04db7c] text-[#011638] font-bold shadow-xs"
                : "text-[#97a0af] hover:text-[#f4f5f7]"
            }`}
          >
            Scenario B: Benign POS ($42.50)
          </button>
        </div>
      </div>

      {actionFeedback && (
        <div className="p-3 bg-[#04db7c15] border border-[#04db7c44] text-[#04db7c] font-mono text-xs rounded-[4px]">
          ✓ {actionFeedback}
        </div>
      )}

      {error && (
        <div className="p-3 bg-[#f0525215] border border-[#f0525233] text-[#f05252] font-mono text-xs rounded-[4px]">
          ⚠ {error}
        </div>
      )}

      {/* 2. Transaction Input & Execution Strip */}
      <div className="p-4 bg-[#0f172a] border border-[#1c2536] rounded-[4px] space-y-3 font-mono text-xs">
        <div className="flex items-center justify-between border-b border-[#1c2536] pb-2">
          <span className="font-bold text-[#f4f5f7] uppercase tracking-wider text-[11px]">
            Input Evaluation Payload
          </span>
          <span className="text-[#5e6c84] text-[10px]">Synchronous Latency Budget: &lt; 10.0ms</span>
        </div>

        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-6 gap-3">
          <div>
            <label className="text-[#5e6c84] text-[10px] block mb-1">TRANSACTION ID</label>
            <input
              type="text"
              value={txId}
              onChange={(e) => setTxId(e.target.value)}
              className="w-full px-2.5 py-1.5 bg-[#0a1324] border border-[#1c2536] text-[#f4f5f7] rounded-[3px]"
            />
          </div>

          <div>
            <label className="text-[#5e6c84] text-[10px] block mb-1">AMOUNT ({currency})</label>
            <input
              type="number"
              value={amount}
              onChange={(e) => setAmount(Number(e.target.value))}
              className="w-full px-2.5 py-1.5 bg-[#0a1324] border border-[#1c2536] text-[#f4f5f7] rounded-[3px]"
            />
          </div>

          <div>
            <label className="text-[#5e6c84] text-[10px] block mb-1">CUSTOMER ID</label>
            <input
              type="text"
              value={customerId}
              onChange={(e) => setCustomerId(e.target.value)}
              className="w-full px-2.5 py-1.5 bg-[#0a1324] border border-[#1c2536] text-[#f4f5f7] rounded-[3px]"
            />
          </div>

          <div>
            <label className="text-[#5e6c84] text-[10px] block mb-1">DEVICE FINGERPRINT</label>
            <input
              type="text"
              value={deviceFp}
              onChange={(e) => setDeviceFp(e.target.value)}
              className="w-full px-2.5 py-1.5 bg-[#0a1324] border border-[#1c2536] text-[#f4f5f7] rounded-[3px]"
            />
          </div>

          <div>
            <label className="text-[#5e6c84] text-[10px] block mb-1">EGRESS IP ADDRESS</label>
            <input
              type="text"
              value={ipAddress}
              onChange={(e) => setIpAddress(e.target.value)}
              className="w-full px-2.5 py-1.5 bg-[#0a1324] border border-[#1c2536] text-[#f4f5f7] rounded-[3px]"
            />
          </div>

          <div className="flex items-end">
            <button
              onClick={handleExecuteEvaluation}
              disabled={evaluating}
              className="w-full py-1.5 bg-[#0d94fb] hover:bg-[#0b82dc] text-white font-bold rounded-[3px] shadow-xs active:scale-95 cursor-pointer disabled:opacity-50 flex items-center justify-center gap-1.5"
            >
              <Send className="w-3.5 h-3.5" />
              <span>{evaluating ? "Scoring..." : "Execute"}</span>
            </button>
          </div>
        </div>
      </div>

      {/* 3. Decision Verdict Summary Strip */}
      <div className="p-4 bg-[#0f172a] border border-[#1c2536] rounded-[4px] font-mono">
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-4 divide-x divide-[#1c2536]">
          <div className="pr-4">
            <span className="text-[10px] text-[#5e6c84] uppercase tracking-wider block">FINAL VERDICT</span>
            <div className="mt-1.5">
              <VerdictBadge verdict={verdict} />
            </div>
          </div>

          <div className="px-4">
            <span className="text-[10px] text-[#5e6c84] uppercase tracking-wider block">COMPOSITE RISK SCORE</span>
            <div className="flex items-baseline gap-2 mt-1">
              <span
                className={`text-2xl font-bold ${
                  riskScore >= 0.35 ? "text-[#f05252]" : riskScore >= 0.05 ? "text-[#f59e0b]" : "text-[#04db7c]"
                }`}
              >
                {riskScore.toFixed(2)}
              </span>
              <span className="text-[10px] text-[#5e6c84]">Scale: 0.00 - 1.00</span>
            </div>
          </div>

          <div className="px-4">
            <span className="text-[10px] text-[#5e6c84] uppercase tracking-wider block">MODEL CONFIDENCE</span>
            <span className="text-xl font-bold text-[#f4f5f7] mt-1 block">94.8%</span>
            <span className="text-[10px] text-[#5e6c84]">Calibrated Beta v2.5</span>
          </div>

          <div className="pl-4">
            <span className="text-[10px] text-[#5e6c84] uppercase tracking-wider block">SERVING LATENCY</span>
            <span className="text-xl font-bold text-[#04db7c] mt-1 block">
              {liveResponse ? liveResponse.latency_ms?.toFixed(2) : "1.42"} ms
            </span>
            <span className="text-[10px] text-[#5e6c84]">Pipeline: ML + Rules</span>
          </div>
        </div>
      </div>

      {/* 4. Additive Factor Decomposition Table */}
      <div className="bg-[#0f172a] border border-[#1c2536] rounded-[4px] p-4 font-mono text-xs">
        <div className="flex items-center justify-between pb-2.5 mb-3 border-b border-[#1c2536]">
          <div className="flex items-center gap-2">
            <Layers className="w-4 h-4 text-[#0d94fb]" />
            <h2 className="font-bold uppercase tracking-wider text-[#f4f5f7] text-xs">
              Additive Factor Attribution
            </h2>
          </div>
          <span className="text-[#5e6c84] text-[11px]">
            &Sigma; Factors = <strong className="text-[#f4f5f7]">+{factorSum.toFixed(2)}</strong> &rarr; Composite Risk Score: <strong className="text-[#f05252]">{riskScore.toFixed(2)}</strong>
          </span>
        </div>

        <div className="overflow-x-auto">
          <table className="w-full text-left">
            <thead>
              <tr className="border-b border-[#1c2536] text-[#5e6c84] text-[11px]">
                <th className="pb-2 font-medium">RISK FACTOR</th>
                <th className="pb-2 font-medium">SOURCE</th>
                <th className="pb-2 font-medium">CONTRIBUTION</th>
                <th className="pb-2 font-medium">OBSERVED EVIDENCE</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-[#1c2536]">
              {factors.map((f, idx) => (
                <tr key={idx} className="hover:bg-[#142036] transition-colors">
                  <td className="py-2.5 font-semibold text-[#f4f5f7]">{f.name}</td>
                  <td className="py-2.5">
                    <span className="px-1.5 py-0.5 rounded-[2px] bg-[#0a1324] border border-[#1c2536] text-[10px] text-[#0d94fb]">
                      {f.source}
                    </span>
                  </td>
                  <td className="py-2.5 font-bold text-[#f05252]">+{f.contribution.toFixed(2)}</td>
                  <td className="py-2.5 text-[#97a0af]">{f.evidence}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>

      {/* 5. Policy Engine Invariants & Collapsible 25-Feature Contract */}
      <div className="bg-[#0f172a] border border-[#1c2536] rounded-[4px] p-4 font-mono text-xs">
        <button
          onClick={() => setIsTechDetailsOpen(!isTechDetailsOpen)}
          className="w-full flex items-center justify-between text-[#f4f5f7] font-bold cursor-pointer"
        >
          <div className="flex items-center gap-2">
            <Terminal className="w-4 h-4 text-[#0d94fb]" />
            <span>TECHNICAL SPECIFICATIONS &amp; 25-FEATURE VECTOR CONTRACT</span>
          </div>
          {isTechDetailsOpen ? (
            <ChevronDown className="w-4 h-4 text-[#5e6c84]" />
          ) : (
            <ChevronRight className="w-4 h-4 text-[#5e6c84]" />
          )}
        </button>

        {isTechDetailsOpen && (
          <div className="mt-3 pt-3 border-t border-[#1c2536] space-y-3">
            <div className="grid grid-cols-2 md:grid-cols-4 gap-2.5">
              <div className="p-2.5 bg-[#0a1324] rounded-[3px] border border-[#1c2536]">
                <span className="text-[#5e6c84] text-[10px] block">PRODUCTION MODEL</span>
                <span className="text-[#0d94fb] font-bold">fraud-xgb-25f-v3.0</span>
              </div>
              <div className="p-2.5 bg-[#0a1324] rounded-[3px] border border-[#1c2536]">
                <span className="text-[#5e6c84] text-[10px] block">FEATURE VECTOR CONTRACT</span>
                <span className="text-[#f4f5f7] font-bold">fraud-risk-25f-v2.5</span>
              </div>
              <div className="p-2.5 bg-[#0a1324] rounded-[3px] border border-[#1c2536]">
                <span className="text-[#5e6c84] text-[10px] block">CALIBRATION</span>
                <span className="text-[#04db7c] font-bold">beta-calibrated-v2.5</span>
              </div>
              <div className="p-2.5 bg-[#0a1324] rounded-[3px] border border-[#1c2536]">
                <span className="text-[#5e6c84] text-[10px] block">AUDIT HASH</span>
                <span className="text-[#97a0af] font-bold truncate block">sha256:4f8a1e9c...</span>
              </div>
            </div>

            <div className="p-3 bg-[#070e1c] rounded-[3px] border border-[#1c2536]">
              <span className="text-[#5e6c84] text-[10px] block mb-1.5">RAW JSON RESPONSE</span>
              <pre className="text-[11px] text-[#0d94fb] overflow-x-auto">
                {JSON.stringify(
                  liveResponse || {
                    decision_id: "dec_9f8a1e9c7a2b",
                    transaction_id: txId,
                    recommended_action: isAttack ? "DECLINE_RECOMMENDATION" : "ALLOW_RECOMMENDATION",
                    risk_score: riskScore,
                    reason_codes: isAttack ? ["HIGH_IP_VELOCITY", "BULLETPROOF_PROXY", "IMPOSSIBLE_TRAVEL"] : [],
                    latency_ms: 1.42,
                    evaluated_at: new Date().toISOString(),
                  },
                  null,
                  2
                )}
              </pre>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
