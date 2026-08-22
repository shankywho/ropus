"use client";

import React, { useState } from "react";
import { Network, ShieldAlert, ZoomIn, ZoomOut } from "lucide-react";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";

interface GraphNode {
  id: string;
  label: string;
  type: "USER" | "DEVICE" | "ACCOUNT" | "MERCHANT" | "CAMPAIGN";
  riskScore: number;
  x: number;
  y: number;
}

interface GraphEdge {
  from: string;
  to: string;
  relationship: string;
}

const mockNodes: GraphNode[] = [
  { id: "usr_synth_01", label: "usr_synthetic_01", type: "USER", riskScore: 94, x: 180, y: 120 },
  { id: "usr_synth_02", label: "usr_synthetic_02", type: "USER", riskScore: 91, x: 280, y: 80 },
  { id: "usr_synth_03", label: "usr_synthetic_03", type: "USER", riskScore: 89, x: 380, y: 130 },
  { id: "dev_emul_root", label: "dev_emulator_root_pool", type: "DEVICE", riskScore: 98, x: 280, y: 220 },
  { id: "acc_mule_layer", label: "acc_mule_layer_99", type: "ACCOUNT", riskScore: 95, x: 160, y: 320 },
  { id: "camp_phantom", label: "CAMP-PHANTOM-CARDING", type: "CAMPAIGN", riskScore: 99, x: 420, y: 300 },
  { id: "merch_crypto", label: "CryptoLiquidityExpress", type: "MERCHANT", riskScore: 65, x: 280, y: 380 },
];

const mockEdges: GraphEdge[] = [
  { from: "usr_synth_01", to: "dev_emul_root", relationship: "SHARED_DEVICE" },
  { from: "usr_synth_02", to: "dev_emul_root", relationship: "SHARED_DEVICE" },
  { from: "usr_synth_03", to: "dev_emul_root", relationship: "SHARED_DEVICE" },
  { from: "usr_synth_01", to: "acc_mule_layer", relationship: "TRANSFERS_FUNDS" },
  { from: "dev_emul_root", to: "camp_phantom", relationship: "INFRASTRUCTURE_FOR" },
  { from: "acc_mule_layer", to: "merch_crypto", relationship: "CASHOUT_RAIL" },
];

export default function FraudGraphExplorerPage() {
  const [selectedNode, setSelectedNode] = useState<GraphNode | null>(mockNodes[3]);

  return (
    <div className="flex-1 p-6 md:p-8 space-y-6 max-w-7xl mx-auto">
      {/* Header */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 border-b border-slate-800 pb-5">
        <div>
          <h1 className="text-2xl font-bold text-white tracking-tight flex items-center gap-2.5">
            <Network className="size-6 text-indigo-400" />
            <span>Fraud Graph Explorer</span>
          </h1>
          <p className="text-sm text-slate-400 mt-1">
            Real-time entity resolution, syndicate cluster visualization, and shared identifier lineages.
          </p>
        </div>
        <div className="flex items-center gap-3">
          <Badge className="bg-rose-500/20 text-rose-300 border-rose-500/30 text-xs px-3 py-1 font-mono">
            Active Syndicate Ring (Degree: 14)
          </Badge>
        </div>
      </div>

      {/* Main Grid: Interactive Graph on Left, Node Details on Right */}
      <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
        {/* Graph Canvas */}
        <div className="lg:col-span-8 bg-slate-900 border border-slate-800 rounded-2xl p-6 relative overflow-hidden flex flex-col justify-between min-h-[500px] shadow-2xl">
          {/* Top Controls */}
          <div className="flex items-center justify-between z-10">
            <div className="flex items-center gap-2 text-xs text-slate-400">
              <span className="flex items-center gap-1.5"><span className="size-2 rounded-full bg-rose-500" /> Critical Risk</span>
              <span className="flex items-center gap-1.5"><span className="size-2 rounded-full bg-indigo-500" /> Infrastructure</span>
              <span className="flex items-center gap-1.5"><span className="size-2 rounded-full bg-amber-500" /> Accounts</span>
            </div>
            <div className="flex items-center gap-2">
              <Button size="icon" variant="outline" className="size-7 bg-slate-800 border-slate-700 text-slate-200">
                <ZoomIn className="size-3.5" />
              </Button>
              <Button size="icon" variant="outline" className="size-7 bg-slate-800 border-slate-700 text-slate-200">
                <ZoomOut className="size-3.5" />
              </Button>
            </div>
          </div>

          {/* SVG Graph Visualization */}
          <svg className="w-full h-full absolute inset-0 pointer-events-auto">
            {/* Draw Edges */}
            {mockEdges.map((e, idx) => {
              const src = mockNodes.find((n) => n.id === e.from);
              const dst = mockNodes.find((n) => n.id === e.to);
              if (!src || !dst) return null;
              return (
                <g key={idx}>
                  <line
                    x1={src.x}
                    y1={src.y}
                    x2={dst.x}
                    y2={dst.y}
                    stroke="#475569"
                    strokeWidth="2"
                    strokeDasharray="4 2"
                  />
                  <text
                    x={(src.x + dst.x) / 2}
                    y={(src.y + dst.y) / 2 - 6}
                    fill="#94a3b8"
                    fontSize="9"
                    textAnchor="middle"
                    className="font-mono select-none"
                  >
                    {e.relationship}
                  </text>
                </g>
              );
            })}

            {/* Draw Nodes */}
            {mockNodes.map((node) => {
              const isSelected = selectedNode?.id === node.id;
              const fillColor =
                node.type === "DEVICE"
                  ? "#6366f1"
                  : node.type === "CAMPAIGN"
                  ? "#f43f5e"
                  : node.type === "ACCOUNT"
                  ? "#f59e0b"
                  : "#10b981";

              return (
                <g
                  key={node.id}
                  onClick={() => setSelectedNode(node)}
                  className="cursor-pointer"
                >
                  <circle
                    cx={node.x}
                    cy={node.y}
                    r={isSelected ? 22 : 18}
                    fill={fillColor}
                    fillOpacity={0.2}
                    stroke={fillColor}
                    strokeWidth={isSelected ? 3 : 2}
                    className="transition-all"
                  />
                  <circle cx={node.x} cy={node.y} r={6} fill={fillColor} />
                  <text
                    x={node.x}
                    y={node.y + 30}
                    fill="#f1f5f9"
                    fontSize="11"
                    fontWeight="bold"
                    textAnchor="middle"
                    className="font-mono select-none"
                  >
                    {node.label}
                  </text>
                </g>
              );
            })}
          </svg>

          {/* Bottom Banner */}
          <div className="z-10 bg-slate-950/80 border border-slate-800 p-3 rounded-xl flex items-center justify-between text-xs">
            <span className="text-slate-400">Transnational Syndicate Cluster: <span className="text-indigo-300 font-bold">Phantom-Carding-09</span></span>
            <span className="font-mono text-emerald-400">Graph Ingestion SLA: 2.75M ops/sec</span>
          </div>
        </div>

        {/* Node Detail Inspector */}
        <div className="lg:col-span-4">
          <Card className="bg-slate-900 border-slate-800 text-slate-100 shadow-xl h-full flex flex-col justify-between">
            <CardHeader className="border-b border-slate-800 p-5">
              <CardTitle className="text-base font-semibold flex items-center gap-2">
                <ShieldAlert className="size-5 text-indigo-400" />
                <span>Entity Inspector</span>
              </CardTitle>
            </CardHeader>
            <CardContent className="p-5 space-y-5 flex-1">
              {selectedNode ? (
                <>
                  <div>
                    <div className="text-xs text-slate-400 font-mono">Entity Identifier</div>
                    <div className="text-sm font-bold font-mono text-white mt-0.5">{selectedNode.label}</div>
                  </div>

                  <div className="grid grid-cols-2 gap-3">
                    <div className="p-3 rounded-lg bg-slate-950/60 border border-slate-800">
                      <div className="text-[11px] text-slate-400">Node Type</div>
                      <div className="text-xs font-bold text-indigo-300 mt-1">{selectedNode.type}</div>
                    </div>
                    <div className="p-3 rounded-lg bg-slate-950/60 border border-slate-800">
                      <div className="text-[11px] text-slate-400">Cluster Risk</div>
                      <div className="text-xs font-bold text-rose-400 mt-1">{selectedNode.riskScore}%</div>
                    </div>
                  </div>

                  <div className="p-4 rounded-xl bg-slate-950/70 border border-slate-800 text-xs text-slate-300 space-y-2">
                    <div className="font-semibold text-slate-200">Syndicate Linkage Summary:</div>
                    <p>
                      This entity serves as a central hub connecting 3 synthetic user identities, routing funds across 1 mule account to merchant cashout rails.
                    </p>
                  </div>

                  <Button className="w-full bg-rose-600 hover:bg-rose-500 text-white text-xs font-semibold">
                    Freeze Entire Cluster (7 Entities)
                  </Button>
                </>
              ) : (
                <div className="text-center py-12 text-slate-500 text-xs">
                  Click on any node in the graph to inspect entity attributes and lineage.
                </div>
              )}
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
}
