import React from "react";
import { Play, Pause, RotateCcw, SkipForward } from "lucide-react";
import { useDemoStore, DemoStage } from "@/lib/demo-store";

export function DemoController() {
  const {
    stage,
    stepIndex,
    isPlaying,
    startDemo,
    pauseDemo,
    resumeDemo,
    resetDemo,
    nextStep,
  } = useDemoStore();

  const STAGES: { key: DemoStage; label: string }[] = [
    { key: "IDLE", label: "0. Idle" },
    { key: "NORMAL_ACTIVITY", label: "1. Baseline" },
    { key: "ATTACK_TRIGGERED", label: "2. Attack Inbound" },
    { key: "SIGNALS_ARRIVING", label: "3. Threat Signals" },
    { key: "GRAPH_REVEALED", label: "4. Mule Graph" },
    { key: "SCORE_ESCALATED", label: "5. Risk Escalation" },
    { key: "DOSSIER_GENERATED", label: "6. AI Dossier" },
    { key: "CASE_CREATED", label: "7. P0 Case" },
    { key: "ANALYST_CONFIRMED", label: "8. Human Action" },
    { key: "COMPLETE", label: "9. Feedback Loop" },
  ];

  return (
    <div className="bg-[#0f172a] border border-[#1c2536] rounded-[4px] p-4 mb-6 shadow-xs">
      <div className="flex flex-wrap items-center justify-between gap-4">
        <div className="flex items-center gap-3">
          <div className="flex items-center gap-1.5">
            <span className="w-2 h-2 rounded-full bg-[#f59e0b] animate-pulse" />
            <span className="text-xs font-mono font-bold text-[#f4f5f7] tracking-wider uppercase">
              Deterministic 7-Stage Demo Engine
            </span>
          </div>
          <span className="text-[10px] font-mono bg-[#f59e0b18] text-[#f59e0b] px-2 py-0.5 rounded-[2px] border border-[#f59e0b33]">
            DEMO ENVIRONMENT
          </span>
        </div>

        {/* Playback Controls (Razorpay Blade) */}
        <div className="flex items-center gap-2">
          {stage === "IDLE" ? (
            <button
              onClick={startDemo}
              className="px-3.5 py-1.5 text-xs font-mono font-bold bg-[#0d94fb] hover:bg-[#0b82dc] text-white rounded-[4px] flex items-center gap-1.5 active:scale-95 shadow-xs cursor-pointer"
            >
              <Play className="w-3.5 h-3.5 fill-current" /> START DEMO
            </button>
          ) : isPlaying ? (
            <button
              onClick={pauseDemo}
              className="px-3.5 py-1.5 text-xs font-mono font-bold bg-[#142036] hover:bg-[#1c2b48] text-[#f4f5f7] rounded-[4px] border border-[#1c2536] flex items-center gap-1.5 active:scale-95 cursor-pointer shadow-xs"
            >
              <Pause className="w-3.5 h-3.5 fill-current" /> PAUSE
            </button>
          ) : (
            <button
              onClick={resumeDemo}
              className="px-3.5 py-1.5 text-xs font-mono font-bold bg-[#0d94fb] hover:bg-[#0b82dc] text-white rounded-[4px] flex items-center gap-1.5 active:scale-95 shadow-xs cursor-pointer"
            >
              <Play className="w-3.5 h-3.5 fill-current" /> RESUME
            </button>
          )}

          <button
            onClick={nextStep}
            disabled={stage === "COMPLETE"}
            className="px-3 py-1.5 text-xs font-mono bg-[#0a1324] hover:bg-[#142036] text-[#f4f5f7] rounded-[4px] border border-[#1c2536] flex items-center gap-1.5 disabled:opacity-50 active:scale-95 cursor-pointer"
          >
            <SkipForward className="w-3.5 h-3.5" /> NEXT STEP
          </button>

          <button
            onClick={resetDemo}
            className="px-3 py-1.5 text-xs font-mono bg-[#0a1324] hover:bg-[#142036] text-[#97a0af] hover:text-[#f4f5f7] rounded-[4px] border border-[#1c2536] flex items-center gap-1.5 active:scale-95 cursor-pointer"
          >
            <RotateCcw className="w-3.5 h-3.5" /> RESET
          </button>
        </div>
      </div>

      {/* Stage Tracker */}
      <div className="mt-3 pt-3 border-t border-[#1c2536] flex items-center justify-between gap-1 overflow-x-auto">
        {STAGES.map((s, idx) => {
          const isActive = idx === stepIndex;
          const isPassed = idx < stepIndex;

          return (
            <div
              key={s.key}
              className={`px-2 py-1 text-[10px] font-mono rounded-[3px] whitespace-nowrap border ${
                isActive
                  ? "bg-[#0d94fb22] border-[#0d94fb] text-[#0d94fb] font-bold"
                  : isPassed
                  ? "bg-[#04db7c11] border-[#04db7c33] text-[#04db7c]"
                  : "bg-[#070e1c] border-[#1c2536] text-[#5e6c84]"
              }`}
            >
              {s.label}
            </div>
          );
        })}
      </div>
    </div>
  );
}
