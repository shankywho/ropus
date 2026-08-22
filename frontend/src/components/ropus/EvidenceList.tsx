import React, { useState } from "react";
import { CheckCircle2, ShieldAlert, ArrowRight } from "lucide-react";

interface EvidenceListProps {
  observed: string[];
  inferred: string[];
  recommended: string[];
  onActionConfirm?: (action: string) => void;
}

export function EvidenceList({
  observed,
  inferred,
  recommended,
  onActionConfirm,
}: EvidenceListProps) {
  const [confirmedActions, setConfirmedActions] = useState<Record<number, boolean>>({});

  const handleAction = (idx: number, actionText: string) => {
    setConfirmedActions((prev) => ({ ...prev, [idx]: true }));
    if (onActionConfirm) {
      onActionConfirm(actionText);
    }
  };

  return (
    <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
      {/* 1. OBSERVED FACTS */}
      <div className="bg-[#0f172a] border border-[#1c2536] rounded-[4px] p-4 flex flex-col justify-between shadow-xs">
        <div>
          <div className="flex items-center justify-between pb-2 mb-3 border-b border-[#1c2536]">
            <span className="text-xs font-mono font-bold tracking-wider text-[#0d94fb] uppercase">
              Observed Facts
            </span>
            <span className="text-[10px] font-mono bg-[#0d94fb15] text-[#0d94fb] px-1.5 py-0.5 rounded-[2px] border border-[#0d94fb33]">
              TELEMETRY
            </span>
          </div>

          <ul className="space-y-2.5">
            {observed.map((fact, i) => (
              <li key={i} className="text-xs text-[#f4f5f7] leading-relaxed flex items-start gap-2">
                <span className="text-[#0d94fb] mt-0.5">•</span>
                <span>{fact}</span>
              </li>
            ))}
          </ul>
        </div>
        <p className="text-[11px] text-[#5e6c84] font-mono mt-4 pt-2 border-t border-[#1c2536]">
          Verifiable immutable logs
        </p>
      </div>

      {/* 2. INFERRED PATTERNS */}
      <div className="bg-[#0f172a] border border-[#1c2536] rounded-[4px] p-4 flex flex-col justify-between shadow-xs">
        <div>
          <div className="flex items-center justify-between pb-2 mb-3 border-b border-[#1c2b48]">
            <span className="text-xs font-mono font-bold tracking-wider text-[#c084fc] uppercase">
              Inferred Patterns
            </span>
            <span className="text-[10px] font-mono bg-[#a855f718] text-[#c084fc] px-1.5 py-0.5 rounded-[2px] border border-[#a855f744]">
              AI INFERRED
            </span>
          </div>

          <ul className="space-y-2.5">
            {inferred.map((pattern, i) => (
              <li key={i} className="text-xs text-[#f4f5f7] leading-relaxed flex items-start gap-2">
                <ShieldAlert className="w-3.5 h-3.5 text-[#c084fc] mt-0.5 shrink-0" />
                <span>{pattern.replace(/^AI INFERRED:\s*/i, "")}</span>
              </li>
            ))}
          </ul>
        </div>
        <p className="text-[11px] text-[#5e6c84] font-mono mt-4 pt-2 border-t border-[#1c2536]">
          Deductions from ML & Graph
        </p>
      </div>

      {/* 3. RECOMMENDED ACTIONS */}
      <div className="bg-[#0f172a] border border-[#1c2536] rounded-[4px] p-4 flex flex-col justify-between shadow-xs">
        <div>
          <div className="flex items-center justify-between pb-2 mb-3 border-b border-[#1c2536]">
            <span className="text-xs font-mono font-bold tracking-wider text-[#f59e0b] uppercase">
              Recommended Actions
            </span>
            <span className="text-[10px] font-mono bg-[#f59e0b18] text-[#f59e0b] px-1.5 py-0.5 rounded-[2px] border border-[#f59e0b44]">
              HUMAN CONFIRMATION
            </span>
          </div>

          <div className="space-y-2">
            {recommended.map((action, i) => {
              const isDone = confirmedActions[i];
              return (
                <div
                  key={i}
                  className={`p-2.5 rounded-[4px] border text-xs flex items-center justify-between gap-2 transition-colors ${
                    isDone
                      ? "bg-[#04db7c12] border-[#04db7c44] text-[#04db7c]"
                      : "bg-[#0a1324] border-[#1c2536] text-[#f4f5f7] hover:border-[#2c3b52]"
                  }`}
                >
                  <span className="leading-tight font-medium">{action}</span>
                  <button
                    onClick={() => handleAction(i, action)}
                    disabled={isDone}
                    className={`px-2.5 py-1 text-[10px] font-mono font-bold rounded-[4px] shrink-0 transition-all ${
                      isDone
                        ? "bg-transparent text-[#04db7c] cursor-default"
                        : "bg-[#0d94fb] hover:bg-[#0b82dc] text-white shadow-xs active:scale-95 cursor-pointer"
                    }`}
                  >
                    {isDone ? (
                      <span className="flex items-center gap-1">
                        <CheckCircle2 className="w-3 h-3 text-[#04db7c]" /> Done
                      </span>
                    ) : (
                      <span className="flex items-center gap-1">
                        Execute <ArrowRight className="w-2.5 h-2.5" />
                      </span>
                    )}
                  </button>
                </div>
              );
            })}
          </div>
        </div>
        <p className="text-[11px] text-[#5e6c84] font-mono mt-4 pt-2 border-t border-[#1c2536]">
          Requires explicit analyst authorization
        </p>
      </div>
    </div>
  );
}
