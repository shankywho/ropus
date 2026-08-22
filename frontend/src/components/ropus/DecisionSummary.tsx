import React from "react";
import { VerdictBadge, VerdictType } from "./VerdictBadge";
import { RiskScore } from "./RiskScore";

interface DecisionSummaryProps {
  transactionId: string;
  customerId: string;
  amount: number;
  currency: string;
  verdict: VerdictType;
  riskScore: number;
  confidence: number;
  latencyMs: number;
  modelVersion: string;
  policyVersion?: string;
  isDemo?: boolean;
}

export function DecisionSummary({
  transactionId,
  customerId,
  amount,
  currency,
  verdict,
  riskScore,
  confidence,
  latencyMs,
  modelVersion,
  isDemo = false,
}: DecisionSummaryProps) {
  return (
    <div className="bg-[#0f172a] border border-[#1c2536] rounded-[4px] p-4 mb-6 shadow-xs">
      <div className="flex flex-wrap items-center justify-between gap-4 pb-3 border-b border-[#1c2536]">
        <div className="flex items-center gap-3">
          <VerdictBadge verdict={verdict} size="lg" />
          <div>
            <div className="flex items-center gap-2">
              <span className="text-xs text-[#5e6c84] font-mono">TXN:</span>
              <span className="text-sm font-mono font-bold text-[#f4f5f7] tracking-tight">
                {transactionId}
              </span>
              {isDemo && (
                <span className="text-[10px] font-mono bg-[#f59e0b15] text-[#f59e0b] px-1.5 py-0.5 rounded-[2px] border border-[#f59e0b33]">
                  DEMO DATA
                </span>
              )}
            </div>
            <p className="text-xs text-[#97a0af] font-mono mt-0.5">
              Customer: <span className="text-[#f4f5f7] font-semibold">{customerId}</span>
            </p>
          </div>
        </div>

        <div className="flex items-center gap-6">
          {/* Amount */}
          <div className="text-right">
            <span className="text-[11px] text-[#5e6c84] font-mono block uppercase">TRANSACTION AMOUNT</span>
            <span className="text-base font-mono font-bold text-[#f4f5f7]">
              ${amount.toLocaleString("en-US", { minimumFractionDigits: 2 })} {currency}
            </span>
          </div>

          {/* Latency */}
          <div className="text-right border-l border-[#1c2536] pl-6">
            <span className="text-[11px] text-[#5e6c84] font-mono block uppercase">EVAL LATENCY</span>
            <span className="text-sm font-mono font-semibold text-[#04db7c]">
              {latencyMs.toFixed(2)} ms
            </span>
          </div>

          {/* Model Runtime */}
          <div className="text-right border-l border-[#1c2536] pl-6 hidden md:block">
            <span className="text-[11px] text-[#5e6c84] font-mono block uppercase">MODEL RUNTIME</span>
            <span className="text-xs font-mono text-[#0d94fb]">{modelVersion}</span>
          </div>
        </div>
      </div>

      <div className="pt-3 flex items-center justify-between gap-4">
        <div className="w-52">
          <RiskScore score={riskScore} confidence={confidence} />
        </div>
        <div className="text-xs text-[#97a0af] font-mono">
          <span className="text-[#5e6c84]">Arbitration:</span>{" "}
          <span className="text-[#f4f5f7] font-semibold">
            {riskScore >= 0.80 ? "Level 1 Regulatory Blacklist / Immediate Block" : "Level 5 Baseline Organic Settlement"}
          </span>
        </div>
      </div>
    </div>
  );
}
