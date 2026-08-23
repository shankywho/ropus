import React from "react";

export type ProvenanceType = "LIVE_BACKEND" | "ANALYTICAL_VIEW" | "SEEDED_INTELLIGENCE" | "DEMO_MODE";

interface DataProvenanceBadgeProps {
  type: ProvenanceType;
  sublabel?: string;
}

export function DataProvenanceBadge({ type, sublabel }: DataProvenanceBadgeProps) {
  const configs: Record<ProvenanceType, { label: string; bg: string; text: string; border: string; dot: string }> = {
    LIVE_BACKEND: {
      label: "LIVE BACKEND",
      bg: "bg-[#04db7c10]",
      text: "text-[#04db7c]",
      border: "border-[#04db7c33]",
      dot: "bg-[#04db7c]",
    },
    ANALYTICAL_VIEW: {
      label: "ANALYTICAL VIEW",
      bg: "bg-[#0d94fb10]",
      text: "text-[#0d94fb]",
      border: "border-[#0d94fb33]",
      dot: "bg-[#0d94fb]",
    },
    SEEDED_INTELLIGENCE: {
      label: "SEEDED INTELLIGENCE",
      bg: "bg-[#f59e0b10]",
      text: "text-[#f59e0b]",
      border: "border-[#f59e0b33]",
      dot: "bg-[#f59e0b]",
    },
    DEMO_MODE: {
      label: "DEMO MODE",
      bg: "bg-[#a855f710]",
      text: "text-[#c084fc]",
      border: "border-[#a855f733]",
      dot: "bg-[#a855f7]",
    },
  };

  const c = configs[type];

  return (
    <span
      className={`inline-flex items-center gap-1.5 px-2 py-0.5 rounded-[3px] border font-mono text-[10px] font-semibold tracking-wider ${c.bg} ${c.text} ${c.border}`}
    >
      <span className={`w-1.5 h-1.5 rounded-full ${c.dot} ${type === "LIVE_BACKEND" ? "animate-pulse" : ""}`} />
      <span>{c.label}</span>
      {sublabel && <span className="text-[#5e6c84] font-normal border-l border-[#1c2536] pl-1.5">{sublabel}</span>}
    </span>
  );
}
