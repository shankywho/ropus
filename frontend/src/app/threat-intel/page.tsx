"use client";

import React, { useState } from "react";
import { DataProvenanceBadge } from "@/components/ropus/DataProvenanceBadge";
import {
  Globe,
  Radio,
  ShieldAlert,
  Terminal,
  Activity,
  Layers,
  Info,
} from "lucide-react";

interface ThreatSignal {
  id: string;
  type: "IP_PROXY" | "VELOCITY_ANOMALY" | "FINGERPRINT_CLUSTER" | "TOR_EXIT_NODE";
  indicator: string;
  source: string;
  confidence: number;
  observed: string;
  lastUpdated: string;
  status: "ACTIVE_BLOCK" | "MONITORING";
}

const SEED_SIGNALS: ThreatSignal[] = [
  {
    id: "sig_01_proxy_cy",
    type: "IP_PROXY",
    indicator: "198.51.100.44 / ASN 49505 (Datacenter VPN, Limassol)",
    source: "Bulletproof Reverse Proxy Pool Feed",
    confidence: 0.98,
    observed: "14 matching sessions in last 24h",
    lastUpdated: "2026-08-22 17:40:00 UTC",
    status: "ACTIVE_BLOCK",
  },
  {
    id: "sig_02_haversine",
    type: "VELOCITY_ANOMALY",
    indicator: "NYC (US) &rarr; Limassol (CY) (8,420 km/h in 14m)",
    source: "Haversine Impossible Travel Engine",
    confidence: 0.99,
    observed: "Velocity exceeds 900 km/h threshold",
    lastUpdated: "2026-08-22 17:42:01 UTC",
    status: "ACTIVE_BLOCK",
  },
  {
    id: "sig_03_canvas_entropy",
    type: "FINGERPRINT_CLUSTER",
    indicator: "Canvas Hash 9f8a... (Linux Emulator v4.2)",
    source: "Hardware Entropy Cluster Analyzer",
    confidence: 0.94,
    observed: "Shared across 14 synthetic accounts",
    lastUpdated: "2026-08-22 16:30:00 UTC",
    status: "ACTIVE_BLOCK",
  },
  {
    id: "sig_04_tor_exit",
    type: "TOR_EXIT_NODE",
    indicator: "185.220.101.5 / Tor Directory Node",
    source: "Tor Project Directory Consensual Feed",
    confidence: 0.99,
    observed: "Direct onion egress match",
    lastUpdated: "2026-08-22 15:10:00 UTC",
    status: "MONITORING",
  },
];

export default function ThreatIntelPage() {
  const [signals] = useState<ThreatSignal[]>(SEED_SIGNALS);

  return (
    <div className="space-y-5">
      {/* 1. Top Header */}
      <div className="flex flex-wrap items-center justify-between gap-4 pb-3.5 border-b border-[#1c2536]">
        <div>
          <div className="flex items-center gap-2.5">
            <h1 className="text-lg font-bold text-[#f4f5f7] tracking-tight font-mono uppercase">
              Threat Intelligence &amp; Signals
            </h1>
            <DataProvenanceBadge type="SEEDED_INTELLIGENCE" sublabel="Forensic Feeds" />
          </div>
          <p className="text-xs text-[#5e6c84] font-mono mt-1">
            Curated threat indicators: bulletproof proxy pools, Haversine velocity anomalies &amp; adversary fingerprints
          </p>
        </div>
      </div>

      {/* Provenance Notice */}
      <div className="p-3 bg-[#f59e0b10] border border-[#f59e0b33] rounded-[4px] font-mono text-xs flex items-center gap-2 text-[#f59e0b]">
        <Info className="w-4 h-4 shrink-0" />
        <span>
          <strong>Seeded Intelligence Dataset:</strong> These threat indicators represent a curated baseline for forensic signal enrichment.
        </span>
      </div>

      {/* 2. Structured Threat Indicators Table */}
      <div className="bg-[#0f172a] border border-[#1c2536] rounded-[4px] p-4 font-mono text-xs">
        <div className="flex items-center justify-between pb-2.5 mb-3 border-b border-[#1c2536]">
          <div className="flex items-center gap-2">
            <Globe className="w-4 h-4 text-[#0d94fb]" />
            <h2 className="font-bold uppercase tracking-wider text-[#f4f5f7] text-xs">
              Threat Indicator Catalog ({signals.length})
            </h2>
          </div>
          <span className="text-[#5e6c84] text-[10px]">Real-Time Feature Enrichment</span>
        </div>

        <div className="overflow-x-auto">
          <table className="w-full text-left">
            <thead>
              <tr className="border-b border-[#1c2536] text-[#5e6c84] text-[11px]">
                <th className="pb-2 font-medium">INDICATOR</th>
                <th className="pb-2 font-medium">TYPE</th>
                <th className="pb-2 font-medium">INTELLIGENCE SOURCE</th>
                <th className="pb-2 font-medium">CONFIDENCE</th>
                <th className="pb-2 font-medium">OBSERVED EVIDENCE</th>
                <th className="pb-2 font-medium text-right">STATUS</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-[#1c2536]">
              {signals.map((sig) => (
                <tr key={sig.id} className="hover:bg-[#142036] transition-colors">
                  <td
                    className="py-3 font-bold text-[#f4f5f7]"
                    dangerouslySetInnerHTML={{ __html: sig.indicator }}
                  />
                  <td className="py-3">
                    <span className="px-1.5 py-0.5 rounded-[2px] bg-[#0a1324] border border-[#1c2536] text-[10px] text-[#0d94fb]">
                      {sig.type}
                    </span>
                  </td>
                  <td className="py-3 text-[#97a0af]">{sig.source}</td>
                  <td className="py-3 font-bold text-[#04db7c]">{(sig.confidence * 100).toFixed(0)}%</td>
                  <td className="py-3 text-[#97a0af]">{sig.observed}</td>
                  <td className="py-3 text-right">
                    <span
                      className={`text-[10px] font-bold px-2 py-0.5 rounded-[2px] border ${
                        sig.status === "ACTIVE_BLOCK"
                          ? "bg-[#f0525215] text-[#f05252] border-[#f0525233]"
                          : "bg-[#f59e0b15] text-[#f59e0b] border-[#f59e0b33]"
                      }`}
                    >
                      {sig.status}
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
