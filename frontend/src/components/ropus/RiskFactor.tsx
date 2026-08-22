import React from "react";
import { RiskFactor as RiskFactorType } from "@/lib/fixtures";

interface RiskFactorProps {
  factor: RiskFactorType;
  maxContribution?: number;
}

export function RiskFactor({ factor, maxContribution = 0.30 }: RiskFactorProps) {
  const sourceColors: Record<string, { bg: string; text: string; border: string }> = {
    Rules: { bg: "bg-[#38bdf818]", text: "text-[#38bdf8]", border: "border-[#38bdf844]" },
    ML: { bg: "bg-[#a855f718]", text: "text-[#c084fc]", border: "border-[#a855f744]" },
    "Threat Intel": { bg: "bg-[#f59e0b18]", text: "text-[#f59e0b]", border: "border-[#f59e0b44]" },
    Graph: { bg: "bg-[#ec489918]", text: "text-[#f472b6]", border: "border-[#ec489944]" },
    Device: { bg: "bg-[#10b98118]", text: "text-[#34d399]", border: "border-[#10b98144]" },
  };

  const currentSource = sourceColors[factor.source] || sourceColors.Rules;
  const percentage = Math.min((factor.contribution / maxContribution) * 100, 100);

  return (
    <div className="py-2.5 border-b border-[#1c2b48] last:border-none">
      <div className="flex items-center justify-between gap-2 mb-1">
        <div className="flex items-center gap-2">
          <span
            className={`text-[10px] font-mono font-medium px-1.5 py-0.5 rounded border ${currentSource.bg} ${currentSource.text} ${currentSource.border}`}
          >
            {factor.source}
          </span>
          <span className="text-sm font-semibold text-[#f4f5f7]">{factor.name}</span>
        </div>
        <span className="text-xs font-mono font-bold text-[#f05252]">
          +{factor.contribution.toFixed(2)}
        </span>
      </div>

      <p className="text-xs text-[#97a0af] leading-relaxed mb-1.5">{factor.description}</p>

      <div className="h-1 w-full bg-[#14223d] rounded-full overflow-hidden">
        <div
          className="h-full bg-[#f05252] rounded-full"
          style={{ width: `${Math.max(percentage, 4)}%` }}
        />
      </div>
    </div>
  );
}
