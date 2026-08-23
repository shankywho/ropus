"use client";

import React, { useState } from "react";
import {
  Radio,
  Search,
} from "lucide-react";

interface ThreatSignal {
  id: string;
  ip: string;
  asn: string;
  country: string;
  category: "BULLETPROOF_PROXY" | "IMPOSSIBLE_TRAVEL" | "DEVICE_ENTROPY_HIGH" | "TOR_EXIT_NODE";
  severity: "CRITICAL" | "HIGH" | "MEDIUM";
  details: string;
  haversineSpeed?: string;
  observedAt: string;
}

const SEED_SIGNALS: ThreatSignal[] = [
  {
    id: "sig_001",
    ip: "198.51.100.44",
    asn: "AS4134 (Chinanet Backbone)",
    country: "CY",
    category: "IMPOSSIBLE_TRAVEL",
    severity: "CRITICAL",
    details: "Haversine travel velocity: 8,420 km/h between NYC (17:28) and Limassol (17:42).",
    haversineSpeed: "8,420 km/h",
    observedAt: "2026-08-22 17:42:01",
  },
  {
    id: "sig_002",
    ip: "198.51.100.44",
    asn: "AS4134",
    country: "CY",
    category: "BULLETPROOF_PROXY",
    severity: "CRITICAL",
    details: "Host flagged in global threat feed as active bulletproof reverse proxy pool.",
    observedAt: "2026-08-22 17:40:15",
  },
  {
    id: "sig_003",
    ip: "203.0.113.88",
    asn: "AS13335 (Cloudflare WARP)",
    country: "US",
    category: "DEVICE_ENTROPY_HIGH",
    severity: "HIGH",
    details: "Hardware canvas fingerprint entropy 0.94 matching known mule emulator cluster #99.",
    observedAt: "2026-08-22 16:30:10",
  },
  {
    id: "sig_004",
    ip: "185.220.101.5",
    asn: "AS200052 (Tor Relay Network)",
    country: "DE",
    category: "TOR_EXIT_NODE",
    severity: "HIGH",
    details: "Direct connection from authenticated Tor Exit Node during credential change.",
    observedAt: "2026-08-22 15:10:02",
  },
];

export default function ThreatIntelligencePage() {
  const [signals] = useState<ThreatSignal[]>(SEED_SIGNALS);
  const [searchQuery, setSearchQuery] = useState("");

  const filtered = signals.filter(
    (s) =>
      s.ip.toLowerCase().includes(searchQuery.toLowerCase()) ||
      s.details.toLowerCase().includes(searchQuery.toLowerCase()) ||
      s.category.toLowerCase().includes(searchQuery.toLowerCase())
  );

  return (
    <div className="max-w-6xl mx-auto space-y-6">
      {/* Top Header */}
      <div className="flex flex-wrap items-center justify-between gap-4 pb-4 border-b border-[#1c2536]">
        <div>
          <div className="flex items-center gap-2">
            <h1 className="text-xl font-bold text-[#f4f5f7] tracking-tight">Threat Intelligence & Signals</h1>
            <span className="text-[10px] font-mono bg-[#f59e0b15] text-[#f59e0b] px-1.5 py-0.5 rounded-[2px] border border-[#f59e0b33]">
              SUBSYSTEM 06
            </span>
          </div>
          <p className="text-xs text-[#97a0af] font-mono mt-0.5">
            Haversine impossible velocity calculation, bulletproof proxy detection & adversary fingerprinting
          </p>
        </div>

        <div className="flex items-center gap-2">
          <div className="relative">
            <Search className="w-3.5 h-3.5 text-[#5e6c84] absolute left-3 top-2.5" />
            <input
              type="text"
              placeholder="Search IP, ASN, Category..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="pl-8 pr-3 py-1.5 bg-[#0f172a] border border-[#1c2536] text-xs font-mono rounded-[4px] text-[#f4f5f7] placeholder-[#5e6c84] focus:outline-none focus:border-[#0d94fb]"
            />
          </div>
        </div>
      </div>

      {/* Top Metrics Cards */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <div className="p-4 bg-[#0f172a] border border-[#1c2536] rounded-[4px] shadow-xs">
          <span className="text-[11px] font-mono text-[#5e6c84] uppercase block">ACTIVE THREAT POOLS</span>
          <span className="text-base font-mono font-bold text-[#f05252] mt-1 block">14 Subnets</span>
          <span className="text-[10px] font-mono text-[#5e6c84]">Automatic IP Blocking</span>
        </div>

        <div className="p-4 bg-[#0f172a] border border-[#1c2536] rounded-[4px] shadow-xs">
          <span className="text-[11px] font-mono text-[#5e6c84] uppercase block">HAVERSINE VELOCITY MAX</span>
          <span className="text-base font-mono font-bold text-[#f59e0b] mt-1 block">8,420 km/h</span>
          <span className="text-[10px] font-mono text-[#5e6c84]">Threshold: &gt;900 km/h</span>
        </div>

        <div className="p-4 bg-[#0f172a] border border-[#1c2536] rounded-[4px] shadow-xs">
          <span className="text-[11px] font-mono text-[#5e6c84] uppercase block">PROXY DISCOVERY RATE</span>
          <span className="text-base font-mono font-bold text-[#0d94fb] mt-1 block">99.8%</span>
          <span className="text-[10px] font-mono text-[#5e6c84]">Sub-1ms Redis Cache</span>
        </div>

        <div className="p-4 bg-[#0f172a] border border-[#1c2536] rounded-[4px] shadow-xs">
          <span className="text-[11px] font-mono text-[#5e6c84] uppercase block">ADVERSARY FINGERPRINTS</span>
          <span className="text-base font-mono font-bold text-[#04db7c] mt-1 block">1,420 Tracked</span>
          <span className="text-[10px] font-mono text-[#5e6c84]">Entropy Resolution: Active</span>
        </div>
      </div>

      {/* Threat Signals Feed Table */}
      <div className="bg-[#0f172a] border border-[#1c2536] rounded-[4px] p-5 shadow-xs">
        <div className="flex items-center justify-between pb-3 mb-4 border-b border-[#1c2536]">
          <div className="flex items-center gap-2">
            <Radio className="w-4 h-4 text-[#f59e0b] animate-pulse" />
            <h2 className="text-xs font-mono font-bold text-[#f4f5f7] uppercase tracking-wider">
              Live Threat Egress ({filtered.length} Active Feeds)
            </h2>
          </div>
          <span className="text-[10px] font-mono text-[#5e6c84]">Synchronous Pipeline Integration</span>
        </div>

        <div className="overflow-x-auto">
          <table className="w-full text-left text-xs font-mono">
            <thead>
              <tr className="border-b border-[#1c2536] text-[#5e6c84]">
                <th className="pb-2.5 font-medium">CATEGORY</th>
                <th className="pb-2.5 font-medium">IP / HOST</th>
                <th className="pb-2.5 font-medium">ASN / GEO</th>
                <th className="pb-2.5 font-medium">THREAT DETAILS</th>
                <th className="pb-2.5 font-medium">SEVERITY</th>
                <th className="pb-2.5 font-medium text-right">OBSERVED</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-[#1c2536]">
              {filtered.map((sig) => (
                <tr key={sig.id} className="hover:bg-[#142036] transition-colors">
                  <td className="py-3">
                    <span className="text-[10px] font-bold px-1.5 py-0.5 rounded-[2px] bg-[#f59e0b15] text-[#f59e0b] border border-[#f59e0b33]">
                      {sig.category}
                    </span>
                  </td>
                  <td className="py-3 text-[#f4f5f7] font-semibold">{sig.ip}</td>
                  <td className="py-3 text-[#97a0af]">
                    {sig.asn} ({sig.country})
                  </td>
                  <td className="py-3 text-[#f4f5f7] max-w-md">{sig.details}</td>
                  <td className="py-3">
                    <span
                      className={`text-[10px] font-bold px-1.5 py-0.5 rounded-[2px] border ${
                        sig.severity === "CRITICAL"
                          ? "bg-[#f0525215] text-[#f05252] border-[#f0525233]"
                          : "bg-[#f59e0b15] text-[#f59e0b] border-[#f59e0b33]"
                      }`}
                    >
                      {sig.severity}
                    </span>
                  </td>
                  <td className="py-3 text-[#97a0af] text-right">{sig.observedAt}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}
