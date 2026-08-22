import React from "react";

interface RiskScoreProps {
  score: number;
  confidence?: number;
  showBar?: boolean;
}

export function RiskScore({ score, confidence, showBar = true }: RiskScoreProps) {
  // Calibrated color mapping
  const getColor = (s: number) => {
    if (s >= 0.80) return "text-[#f05252]";
    if (s >= 0.50) return "text-[#f97316]";
    if (s >= 0.30) return "text-[#f59e0b]";
    return "text-[#04db7c]";
  };

  const getBarBg = (s: number) => {
    if (s >= 0.80) return "bg-[#f05252]";
    if (s >= 0.50) return "bg-[#f97316]";
    if (s >= 0.30) return "bg-[#f59e0b]";
    return "bg-[#04db7c]";
  };

  const colorClass = getColor(score);
  const barBgClass = getBarBg(score);

  return (
    <div className="flex flex-col gap-1.5">
      <div className="flex items-baseline justify-between gap-3">
        <div className="flex items-baseline gap-1">
          <span className={`text-2xl font-bold font-mono tracking-tight ${colorClass}`}>
            {score.toFixed(2)}
          </span>
          <span className="text-xs text-[#6b778c] font-mono">/ 1.00</span>
        </div>
        {confidence !== undefined && (
          <span className="text-[11px] text-[#97a0af] font-mono">
            conf: {(confidence * 100).toFixed(1)}%
          </span>
        )}
      </div>

      {showBar && (
        <div className="h-1.5 w-full bg-[#1c2b48] rounded-full overflow-hidden">
          <div
            className={`h-full ${barBgClass} rounded-full transition-all duration-300`}
            style={{ width: `${Math.min(Math.max(score * 100, 3), 100)}%` }}
          />
        </div>
      )}
    </div>
  );
}
