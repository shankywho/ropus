"use client";

import React, { useState } from "react";
import { DataProvenanceBadge } from "@/components/ropus/DataProvenanceBadge";
import {
  User,
  Smartphone,
  Globe,
  CreditCard,
  Building,
  ShieldAlert,
  Layers,
  Info,
} from "lucide-react";

interface GraphEntity {
  id: string;
  type: "customer" | "device" | "ip" | "transaction" | "payout_account" | "connected_account";
  label: string;
  sublabel: string;
  riskStatus: "HIGH_RISK" | "BENIGN" | "SUSPICIOUS";
  firstSeen: string;
  lastSeen: string;
  connectionsCount: number;
  hopLevel: number;
  signals: string[];
  x: number;
  y: number;
}

interface GraphLink {
  source: string;
  target: string;
  relation: string;
  hop: number;
}

const INITIAL_NODES: GraphEntity[] = [
  {
    id: "usr_sarah_connor",
    type: "customer",
    label: "Sarah Connor",
    sublabel: "usr_sarah_connor",
    riskStatus: "HIGH_RISK",
    firstSeen: "2024-03-12",
    lastSeen: "2026-08-22 17:42:00",
    connectionsCount: 4,
    hopLevel: 0,
    signals: ["Target of ATO liquidity drain", "Monetary Spike: $14,500"],
    x: 350,
    y: 180,
  },
  {
    id: "dev_mule_cluster_99",
    type: "device",
    label: "Hardware Canvas Mule #99",
    sublabel: "dev_mule_cluster_99",
    riskStatus: "HIGH_RISK",
    firstSeen: "2026-08-01",
    lastSeen: "2026-08-22 17:42:00",
    connectionsCount: 14,
    hopLevel: 1,
    signals: ["Shared canvas hash across 14 synthetic accounts", "WebGL entropy 0.94"],
    x: 180,
    y: 90,
  },
  {
    id: "ip_198_51_100_44",
    type: "ip",
    label: "198.51.100.44",
    sublabel: "Datacenter Proxy (CY)",
    riskStatus: "HIGH_RISK",
    firstSeen: "2026-08-22 17:20:00",
    lastSeen: "2026-08-22 17:42:00",
    connectionsCount: 8,
    hopLevel: 1,
    signals: ["Bulletproof VPN Subnet", "Impossible travel velocity (8,420 km/h)"],
    x: 520,
    y: 90,
  },
  {
    id: "tx_order_88419",
    type: "transaction",
    label: "Wire Transfer #88419",
    sublabel: "$14,500.00 USD",
    riskStatus: "HIGH_RISK",
    firstSeen: "2026-08-22 17:42:00",
    lastSeen: "2026-08-22 17:42:00",
    connectionsCount: 2,
    hopLevel: 1,
    signals: ["Velocity anomaly (>400% baseline)", "Blocked in 1.42ms"],
    x: 350,
    y: 300,
  },
  {
    id: "payout_offshore_882",
    type: "payout_account",
    label: "Offshore IBAN Mule",
    sublabel: "CY88...9124",
    riskStatus: "HIGH_RISK",
    firstSeen: "2026-08-20",
    lastSeen: "2026-08-22 17:42:00",
    connectionsCount: 3,
    hopLevel: 2,
    signals: ["Recipient IBAN quarantined across 3 institutional banks"],
    x: 520,
    y: 300,
  },
  {
    id: "usr_mule_alex_m",
    type: "connected_account",
    label: "Alex Murphy (Compromised)",
    sublabel: "usr_alex_murphy",
    riskStatus: "SUSPICIOUS",
    firstSeen: "2025-11-04",
    lastSeen: "2026-08-22 16:30:00",
    connectionsCount: 3,
    hopLevel: 2,
    signals: ["Co-occurred on device dev_mule_cluster_99 2 hours prior"],
    x: 80,
    y: 220,
  },
];

const INITIAL_LINKS: GraphLink[] = [
  { source: "usr_sarah_connor", target: "dev_mule_cluster_99", relation: "LOGIN_DEVICE", hop: 1 },
  { source: "usr_sarah_connor", target: "ip_198_51_100_44", relation: "SESSION_IP", hop: 1 },
  { source: "usr_sarah_connor", target: "tx_order_88419", relation: "INITIATED_TXN", hop: 1 },
  { source: "tx_order_88419", target: "payout_offshore_882", relation: "BENEFICIARY_IBAN", hop: 2 },
  { source: "dev_mule_cluster_99", target: "usr_mule_alex_m", relation: "SHARED_FINGERPRINT", hop: 2 },
];

export default function FraudGraphPage() {
  const [nodes] = useState<GraphEntity[]>(INITIAL_NODES);
  const [links] = useState<GraphLink[]>(INITIAL_LINKS);
  const [selectedEntityId, setSelectedEntityId] = useState<string>("usr_sarah_connor");
  const [hopFilter, setHopFilter] = useState<number>(3);

  const selectedNode = nodes.find((n) => n.id === selectedEntityId) || nodes[0];
  const filteredNodes = nodes.filter((n) => n.hopLevel <= hopFilter);
  const filteredLinks = links.filter((l) => l.hop <= hopFilter);

  const getEntityIcon = (type: GraphEntity["type"]) => {
    switch (type) {
      case "customer":
        return <User className="w-3.5 h-3.5" />;
      case "device":
        return <Smartphone className="w-3.5 h-3.5" />;
      case "ip":
        return <Globe className="w-3.5 h-3.5" />;
      case "transaction":
        return <CreditCard className="w-3.5 h-3.5" />;
      case "payout_account":
        return <Building className="w-3.5 h-3.5" />;
      case "connected_account":
        return <ShieldAlert className="w-3.5 h-3.5" />;
    }
  };

  return (
    <div className="space-y-5">
      {/* 1. Top Header */}
      <div className="flex flex-wrap items-center justify-between gap-4 pb-3.5 border-b border-[#1c2536]">
        <div>
          <div className="flex items-center gap-2.5">
            <h1 className="text-lg font-bold text-[#f4f5f7] tracking-tight font-mono uppercase">
              Forensic Graph Visualization
            </h1>
            <DataProvenanceBadge type="ANALYTICAL_VIEW" sublabel="BFS Entity Topology" />
          </div>
          <p className="text-xs text-[#5e6c84] font-mono mt-1">
            Multi-hop relational syndication across Account &bull; Device &bull; IP &bull; Beneficiary nodes
          </p>
        </div>

        {/* Hop Filter */}
        <div className="flex items-center gap-1.5 bg-[#0f172a] p-1 border border-[#1c2536] rounded-[4px] font-mono text-xs">
          <span className="text-[#5e6c84] text-[10px] pl-1.5 uppercase">Depth:</span>
          {[1, 2, 3].map((hop) => (
            <button
              key={hop}
              onClick={() => setHopFilter(hop)}
              className={`px-2.5 py-1 rounded-[3px] font-semibold transition-all cursor-pointer ${
                hopFilter === hop
                  ? "bg-[#0d94fb] text-white shadow-xs"
                  : "text-[#97a0af] hover:text-[#f4f5f7]"
              }`}
            >
              {hop}-Hop
            </button>
          ))}
        </div>
      </div>

      {/* Analytical View Notice */}
      <div className="p-3 bg-[#0d94fb10] border border-[#0d94fb33] rounded-[4px] font-mono text-xs flex items-center gap-2 text-[#0d94fb]">
        <Info className="w-4 h-4 shrink-0" />
        <span>
          <strong>Forensic Visualization View:</strong> Graph relationships shown here represent an analytical topology. Live risk evaluations evaluate graph features directly in Redis feature stores.
        </span>
      </div>

      {/* 2. Main Two-Column Canvas */}
      <div className="grid grid-cols-1 lg:grid-cols-12 gap-5">
        {/* Left 8 Cols: Interactive SVG Entity Topology */}
        <div className="lg:col-span-8 bg-[#0f172a] border border-[#1c2536] rounded-[4px] p-4 relative min-h-[460px] flex flex-col justify-between font-mono">
          <div className="flex items-center justify-between border-b border-[#1c2536] pb-2 text-xs">
            <span className="font-bold text-[#f4f5f7] uppercase tracking-wider text-[11px]">
              Multi-Hop Topology ({filteredNodes.length} Nodes &bull; {filteredLinks.length} Edges)
            </span>
            <span className="text-[#5e6c84] text-[10px]">BFS Subgraph</span>
          </div>

          <div className="relative flex-1 flex items-center justify-center my-4 overflow-hidden">
            <svg className="w-full h-80" viewBox="0 0 600 380">
              {/* Edges */}
              {filteredLinks.map((link, idx) => {
                const s = nodes.find((n) => n.id === link.source);
                const t = nodes.find((n) => n.id === link.target);
                if (!s || !t) return null;
                return (
                  <g key={idx}>
                    <line
                      x1={s.x}
                      y1={s.y}
                      x2={t.x}
                      y2={t.y}
                      stroke="#1c2536"
                      strokeWidth={1.5}
                      strokeDasharray={link.hop === 2 ? "3,3" : "none"}
                    />
                    <text
                      x={(s.x + t.x) / 2}
                      y={(s.y + t.y) / 2 - 4}
                      fill="#5e6c84"
                      fontSize="9"
                      textAnchor="middle"
                      className="font-mono"
                    >
                      {link.relation}
                    </text>
                  </g>
                );
              })}

              {/* Nodes */}
              {filteredNodes.map((node) => {
                const isSelected = node.id === selectedEntityId;
                return (
                  <g
                    key={node.id}
                    transform={`translate(${node.x}, ${node.y})`}
                    onClick={() => setSelectedEntityId(node.id)}
                    className="cursor-pointer"
                  >
                    <circle
                      r={isSelected ? 22 : 18}
                      fill={node.riskStatus === "HIGH_RISK" ? "#f0525220" : "#0d94fb20"}
                      stroke={
                        isSelected
                          ? "#0d94fb"
                          : node.riskStatus === "HIGH_RISK"
                          ? "#f05252"
                          : "#0d94fb"
                      }
                      strokeWidth={isSelected ? 2.5 : 1.5}
                    />
                    <text
                      y={4}
                      textAnchor="middle"
                      fill={node.riskStatus === "HIGH_RISK" ? "#f05252" : "#0d94fb"}
                      fontSize="10"
                      fontWeight="bold"
                    >
                      {node.type.substring(0, 2).toUpperCase()}
                    </text>
                    <text
                      y={32}
                      textAnchor="middle"
                      fill="#f4f5f7"
                      fontSize="10"
                      fontWeight="bold"
                    >
                      {node.label}
                    </text>
                  </g>
                );
              })}
            </svg>
          </div>

          <div className="pt-2 border-t border-[#1c2536] flex items-center justify-between text-[10px] text-[#5e6c84]">
            <span>Click any node to inspect forensic attributes &bull; Scroll to zoom</span>
            <span>Syndicate Cluster Risk: CRITICAL</span>
          </div>
        </div>

        {/* Right 4 Cols: Selected Entity Inspector */}
        <div className="lg:col-span-4 bg-[#0f172a] border border-[#1c2536] rounded-[4px] p-4 space-y-4 font-mono text-xs">
          <div className="flex items-center justify-between pb-2 border-b border-[#1c2536]">
            <span className="font-bold text-[#f4f5f7] uppercase tracking-wider text-[11px]">
              Entity Inspector
            </span>
            <span
              className={`text-[10px] font-bold px-1.5 py-0.5 rounded-[2px] border ${
                selectedNode.riskStatus === "HIGH_RISK"
                  ? "bg-[#f0525215] text-[#f05252] border-[#f0525233]"
                  : "bg-[#0d94fb15] text-[#0d94fb] border-[#0d94fb33]"
              }`}
            >
              {selectedNode.riskStatus}
            </span>
          </div>

          <div className="space-y-2.5">
            <div>
              <span className="text-[#5e6c84] text-[10px] block">ENTITY LABEL</span>
              <span className="text-[#f4f5f7] font-bold text-sm">{selectedNode.label}</span>
            </div>

            <div>
              <span className="text-[#5e6c84] text-[10px] block">IDENTIFIER</span>
              <span className="text-[#0d94fb]">{selectedNode.sublabel}</span>
            </div>

            <div className="grid grid-cols-2 gap-2 pt-1">
              <div className="p-2 bg-[#0a1324] border border-[#1c2536] rounded-[3px]">
                <span className="text-[#5e6c84] text-[10px] block">TYPE</span>
                <span className="text-[#f4f5f7] font-bold uppercase text-[11px]">{selectedNode.type}</span>
              </div>
              <div className="p-2 bg-[#0a1324] border border-[#1c2536] rounded-[3px]">
                <span className="text-[#5e6c84] text-[10px] block">CONNECTIONS</span>
                <span className="text-[#f05252] font-bold text-[11px]">{selectedNode.connectionsCount} Linked</span>
              </div>
            </div>

            <div className="pt-2 border-t border-[#1c2536]">
              <span className="text-[#5e6c84] text-[10px] block mb-1.5 uppercase font-bold">
                Observed Risk Signals
              </span>
              <ul className="space-y-1 text-[11px] text-[#97a0af] list-disc list-inside">
                {selectedNode.signals.map((s, idx) => (
                  <li key={idx}>{s}</li>
                ))}
              </ul>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
