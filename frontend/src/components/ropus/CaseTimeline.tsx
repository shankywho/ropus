import React from "react";
import { Terminal } from "lucide-react";

export interface TimelineEntry {
  id: string;
  timestamp: string;
  actor: string;
  action: string;
  details: string;
  sha256Hash: string;
}

interface CaseTimelineProps {
  entries: TimelineEntry[];
}

export function CaseTimeline({ entries }: CaseTimelineProps) {
  return (
    <div className="space-y-3">
      {entries.map((entry, idx) => (
        <div
          key={entry.id || idx}
          className="p-3 bg-[#070e1c] border border-[#1c2b48] rounded text-xs"
        >
          <div className="flex items-center justify-between gap-2 mb-1.5">
            <div className="flex items-center gap-2">
              <span className="font-mono text-[#38bdf8] font-bold">{entry.actor}</span>
              <span className="text-[#6b778c]">•</span>
              <span className="text-[#f4f5f7] font-semibold">{entry.action}</span>
            </div>
            <span className="text-[11px] font-mono text-[#6b778c]">{entry.timestamp}</span>
          </div>

          <p className="text-xs text-[#97a0af] leading-relaxed mb-2">{entry.details}</p>

          <div className="flex items-center gap-1.5 pt-1.5 border-t border-[#14223d] text-[10px] font-mono text-[#6b778c]">
            <Terminal className="w-3 h-3 text-[#6b778c]" />
            <span>SHA-256:</span>
            <span className="text-[#97a0af] truncate max-w-xs">{entry.sha256Hash}</span>
          </div>
        </div>
      ))}
    </div>
  );
}
