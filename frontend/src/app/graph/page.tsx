"use client";

import React, { useState } from "react";
import {
  User,
  Smartphone,
  Globe,
  CreditCard,
  Building,
  ShieldAlert,
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
    signals: ["Newly linked offshore payout rail", "Recipient of 3 prior fraudulent transfers"],
    x: 350,
    y: 410,
  },
  // 2-Hop / 3-Hop Nodes
  {
    id: "usr_synthetic_mule_12",
    type: "connected_account",
    label: "Synthetic Mule Account #12",
    sublabel: "usr_mule_12",
    riskStatus: "HIGH_RISK",
    firstSeen: "2026-08-15",
    lastSeen: "2026-08-21 11:15:00",
    connectionsCount: 3,
    hopLevel: 2,
    signals: ["Created using shared canvas hash", "Confirmed chargeback loss $8,200"],
    x: 60,
    y: 60,
  },
  {
    id: "usr_synthetic_mule_14",
    type: "connected_account",
    label: "Synthetic Mule Account #14",
    sublabel: "usr_mule_14",
    riskStatus: "HIGH_RISK",
    firstSeen: "2026-08-18",
    lastSeen: "2026-08-22 09:30:00",
    connectionsCount: 2,
    hopLevel: 2,
    signals: ["Created using shared canvas hash", "Dispute pending"],
    x: 60,
    y: 160,
  },
  {
    id: "ip_cluster_proxy_pool",
    type: "ip",
    label: "198.51.100.0/24 Subnet",
    sublabel: "Proxy Farm",
    riskStatus: "HIGH_RISK",
    firstSeen: "2026-07-10",
    lastSeen: "2026-08-22 17:42:00",
    connectionsCount: 42,
    hopLevel: 3,
    signals: ["Coordinated proxy rotation pool", "Host to 14 confirmed fraud syndicates"],
    x: 640,
    y: 60,
  },
];

const INITIAL_LINKS: GraphLink[] = [
  { source: "usr_sarah_connor", target: "dev_mule_cluster_99", relation: "USED_DEVICE", hop: 1 },
  { source: "usr_sarah_connor", target: "ip_198_51_100_44", relation: "USED_IP", hop: 1 },
  { source: "usr_sarah_connor", target: "tx_order_88419", relation: "INITIATED_TX", hop: 1 },
  { source: "tx_order_88419", target: "payout_offshore_882", relation: "DESTINATION_PAYOUT", hop: 2 },
  { source: "dev_mule_cluster_99", target: "usr_synthetic_mule_12", relation: "MATCHES_CANVAS", hop: 2 },
  { source: "dev_mule_cluster_99", target: "usr_synthetic_mule_14", relation: "MATCHES_CANVAS", hop: 2 },
  { source: "ip_198_51_100_44", target: "ip_cluster_proxy_pool", relation: "BELONGS_TO_CIDR", hop: 3 },
];

export default function FraudGraphPage() {
  const [hopFilter, setHopFilter] = useState<number>(2);
  const [selectedEntity, setSelectedEntity] = useState<GraphEntity>(INITIAL_NODES[0]);

  const visibleNodes = INITIAL_NODES.filter((n) => n.hopLevel <= hopFilter);
  const visibleLinks = INITIAL_LINKS.filter((l) => l.hop <= hopFilter);

  const getTypeIcon = (type: GraphEntity["type"]) => {
    switch (type) {
      case "customer":
        return <User className="w-4 h-4 text-[#0d94fb]" />;
      case "device":
        return <Smartphone className="w-4 h-4 text-[#c084fc]" />;
      case "ip":
        return <Globe className="w-4 h-4 text-[#f59e0b]" />;
      case "transaction":
        return <CreditCard className="w-4 h-4 text-[#f05252]" />;
      case "payout_account":
        return <Building className="w-4 h-4 text-[#f472b6]" />;
      case "connected_account":
        return <ShieldAlert className="w-4 h-4 text-[#f05252]" />;
    }
  };

  return (
    <div className="max-w-6xl mx-auto space-y-6">
      {/* Top Header & Hop Controls */}
      <div className="flex flex-wrap items-center justify-between gap-4 pb-4 border-b border-[#1c2536]">
        <div>
          <div className="flex items-center gap-2">
            <h1 className="text-xl font-bold text-[#f4f5f7] tracking-tight">Fraud Knowledge Graph</h1>
            <span className="text-[10px] font-mono bg-[#0d94fb15] text-[#0d94fb] px-1.5 py-0.5 rounded-[2px] border border-[#0d94fb33]">
              GRAPH 3.0
            </span>
          </div>
          <p className="text-xs text-[#97a0af] font-mono mt-0.5">
            Real-time in-memory 3-hop BFS entity resolution & syndicate ring discovery
          </p>
        </div>

        {/* Hop Depth Filter Buttons */}
        <div className="flex items-center gap-2">
          <span className="text-xs font-mono text-[#5e6c84]">TRAVERSAL DEPTH:</span>
          <div className="flex bg-[#0f172a] p-1 border border-[#1c2536] rounded-[4px]">
            {[1, 2, 3].map((hop) => (
              <button
                key={hop}
                onClick={() => setHopFilter(hop)}
                className={`px-3 py-1 text-xs font-mono rounded-[3px] font-semibold transition-all cursor-pointer ${
                  hopFilter === hop
                    ? "bg-[#0d94fb] text-white shadow-xs"
                    : "text-[#97a0af] hover:text-[#f4f5f7]"
                }`}
              >
                {hop}-Hop {hop === 1 ? "(Direct)" : hop === 2 ? "(Syndicate)" : "(Cluster)"}
              </button>
            ))}
          </div>
        </div>
      </div>

      {/* Main Canvas & Inspector Layout */}
      <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
        {/* Left 8 Cols: Interactive Visual Graph Canvas */}
        <div className="lg:col-span-8 bg-[#0f172a] border border-[#1c2536] rounded-[4px] p-4 relative min-h-[500px] flex flex-col justify-between overflow-hidden shadow-xs">
          {/* Canvas Badge */}
          <div className="flex items-center justify-between z-10">
            <div className="flex items-center gap-2">
              <span className="text-xs font-mono font-bold text-[#f4f5f7]">
                Active Cluster: #SYNTHETIC-MULE-CLUSTER-99
              </span>
              <span className="text-[10px] font-mono bg-[#f0525218] text-[#f05252] px-1.5 py-0.5 rounded-[2px] border border-[#f0525233]">
                CONFIRMED SYNDICATE
              </span>
            </div>
            <span className="text-[10px] font-mono text-[#5e6c84]">
              {visibleNodes.length} Nodes • {visibleLinks.length} Edges Rendered
            </span>
          </div>

          {/* SVG Visual Graph Rendering */}
          <div className="relative w-full h-[440px] my-2 bg-[#070e1c] border border-[#1c2536] rounded-[4px]">
            <svg className="absolute inset-0 w-full h-full">
              {/* Edges */}
              {visibleLinks.map((link, idx) => {
                const sourceNode = visibleNodes.find((n) => n.id === link.source);
                const targetNode = visibleNodes.find((n) => n.id === link.target);
                if (!sourceNode || !targetNode) return null;

                return (
                  <g key={idx}>
                    <line
                      x1={sourceNode.x}
                      y1={sourceNode.y}
                      x2={targetNode.x}
                      y2={targetNode.y}
                      stroke={link.hop === 1 ? "#0d94fb" : "#f05252"}
                      strokeWidth={link.hop === 1 ? "1.5" : "1"}
                      strokeDasharray={link.hop > 1 ? "4 4" : "none"}
                      opacity="0.65"
                    />
                  </g>
                );
              })}
            </svg>

            {/* Nodes */}
            {visibleNodes.map((node) => {
              const isSelected = selectedEntity.id === node.id;
              const isRoot = node.id === "usr_sarah_connor";

              return (
                <button
                  key={node.id}
                  onClick={() => setSelectedEntity(node)}
                  style={{ left: `${node.x}px`, top: `${node.y}px` }}
                  className={`absolute -translate-x-1/2 -translate-y-1/2 p-2.5 rounded-[4px] border text-left cursor-pointer transition-all shadow-md ${
                    isSelected
                      ? "bg-[#0a1324] border-[#0d94fb] ring-2 ring-[#0d94fb55] z-20 scale-105"
                      : isRoot
                      ? "bg-[#0f172a] border-[#0d94fb] hover:border-[#0d94fb] z-10"
                      : "bg-[#0f172a] border-[#1c2536] hover:border-[#2c3b52]"
                  }`}
                >
                  <div className="flex items-center gap-1.5 mb-1">
                    {getTypeIcon(node.type)}
                    <span className="text-xs font-mono font-bold text-[#f4f5f7] whitespace-nowrap">
                      {node.label}
                    </span>
                  </div>
                  <div className="flex items-center justify-between gap-3 text-[10px] font-mono text-[#97a0af]">
                    <span>{node.sublabel}</span>
                    <span className={node.riskStatus === "HIGH_RISK" ? "text-[#f05252] font-bold" : "text-[#04db7c]"}>
                      {node.connectionsCount} links
                    </span>
                  </div>
                </button>
              );
            })}
          </div>

          {/* Canvas Bottom Legend */}
          <div className="flex items-center justify-between text-[11px] font-mono text-[#5e6c84] pt-2 border-t border-[#1c2536]">
            <div className="flex items-center gap-4">
              <span className="flex items-center gap-1">
                <span className="w-2 h-0.5 bg-[#0d94fb] inline-block" /> 1-Hop Direct
              </span>
              <span className="flex items-center gap-1">
                <span className="w-2 h-0.5 bg-[#f05252] inline-block border-t border-dashed" /> 2/3-Hop Shared Canvas / CIDR
              </span>
            </div>
            <span>Click any node to inspect entity signals</span>
          </div>
        </div>

        {/* Right 4 Cols: Compact Entity Inspector */}
        <div className="lg:col-span-4 bg-[#0f172a] border border-[#1c2536] rounded-[4px] p-5 flex flex-col justify-between shadow-xs">
          <div>
            <div className="flex items-center justify-between pb-3 mb-4 border-b border-[#1c2536]">
              <span className="text-xs font-mono font-bold tracking-wider text-[#0d94fb] uppercase">
                Entity Inspector
              </span>
              <span className="text-[10px] font-mono bg-[#142036] text-[#97a0af] px-1.5 py-0.5 rounded-[2px] uppercase">
                {selectedEntity.type}
              </span>
            </div>

            {/* Entity Header */}
            <div className="mb-4">
              <h3 className="text-base font-bold text-[#f4f5f7] font-mono leading-snug">
                {selectedEntity.label}
              </h3>
              <p className="text-xs text-[#97a0af] font-mono mt-0.5">{selectedEntity.id}</p>
            </div>

            {/* Core Metrics */}
            <div className="space-y-2.5 p-3 bg-[#0a1324] border border-[#1c2536] rounded-[4px] mb-4 font-mono text-xs">
              <div className="flex items-center justify-between">
                <span className="text-[#5e6c84]">RISK STATUS:</span>
                <span
                  className={`font-bold px-1.5 py-0.2 rounded-[2px] text-[10px] ${
                    selectedEntity.riskStatus === "HIGH_RISK"
                      ? "bg-[#f0525215] text-[#f05252] border border-[#f0525233]"
                      : "bg-[#04db7c15] text-[#04db7c] border border-[#04db7c33]"
                  }`}
                >
                  {selectedEntity.riskStatus}
                </span>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-[#5e6c84]">DEGREE CENTRALITY:</span>
                <span className="text-[#f4f5f7] font-semibold">{selectedEntity.connectionsCount} Linked Edges</span>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-[#5e6c84]">HOP DISTANCE:</span>
                <span className="text-[#f4f5f7]">{selectedEntity.hopLevel} Hops from Root</span>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-[#5e6c84]">FIRST OBSERVED:</span>
                <span className="text-[#97a0af]">{selectedEntity.firstSeen}</span>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-[#5e6c84]">LAST ACTIVE:</span>
                <span className="text-[#97a0af]">{selectedEntity.lastSeen}</span>
              </div>
            </div>

            {/* Signals on this Entity */}
            <div>
              <span className="text-[11px] font-mono font-bold text-[#5e6c84] uppercase block mb-2">
                Correlated Graph Signals
              </span>
              <ul className="space-y-2">
                {selectedEntity.signals.map((sig, i) => (
                  <li
                    key={i}
                    className="p-2 bg-[#070e1c] border border-[#1c2536] rounded-[4px] text-xs text-[#f4f5f7] leading-relaxed flex items-start gap-2"
                  >
                    <ShieldAlert className="w-3.5 h-3.5 text-[#f05252] shrink-0 mt-0.5" />
                    <span>{sig}</span>
                  </li>
                ))}
              </ul>
            </div>
          </div>

          <div className="pt-4 mt-4 border-t border-[#1c2536] text-[11px] text-[#5e6c84] font-mono">
            Graph traversal latency: <span className="text-[#04db7c] font-semibold">1.10 ms</span>
          </div>
        </div>
      </div>
    </div>
  );
}
