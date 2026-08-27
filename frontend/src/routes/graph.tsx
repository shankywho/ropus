import { createFileRoute, Link } from "@tanstack/react-router";
import { useMemo, useState, useRef, useEffect, useCallback } from "react";
import { cn } from "@/lib/utils";
import {
  DemoTag,
  Mono,
  StatusPill,
  entityRiskMeta,
} from "@/components/ropus/core";
import { useSuspenseQuery } from "@tanstack/react-query";
import { graphQuery } from "@/lib/ropus/api";
import { blockedDecision } from "@/lib/ropus/fixtures";
import type { EntityRisk } from "@/lib/ropus/contracts";
import {
  ZoomIn,
  ZoomOut,
  Maximize2,
  Minimize2,
  Compass,
  Layers,
  Sparkles,
  ShieldAlert,
  User,
  Smartphone,
  Globe,
  CreditCard,
  Building2,
  ArrowRight,
  PlusCircle,
  Share2,
  Info,
} from "lucide-react";
import { toast } from "sonner";

export const Route = createFileRoute("/graph")({
  head: () => ({
    meta: [
      { title: "Fraud Knowledge Graph — ROPUS Risk Control Plane" },
      {
        name: "description",
        content:
          "Interactive force-directed entity graph for syndicate discovery, GNN shadow traversal, multi-hop BFS relationships, and shared velocity forensics.",
      },
      { property: "og:title", content: "Fraud Knowledge Graph — ROPUS" },
      {
        property: "og:description",
        content: "Traverse connected entities behind risk decisions with GraphSAGE shadow inference.",
      },
      { property: "og:type", content: "website" },
      { name: "twitter:card", content: "summary_large_image" },
    ],
  }),
  loader: ({ context }) =>
    context.queryClient.ensureQueryData(graphQuery(blockedDecision.decisionId)),
  component: FraudGraphPage,
});

const risks: EntityRisk[] = ["CLEAN", "WATCH", "SUSPECT", "CONFIRMED_FRAUD"];

export function FraudGraphPage() {
  const { data: fraudGraph } = useSuspenseQuery(graphQuery(blockedDecision.decisionId));
  const [hops, setHops] = useState<number>(3);
  const [selectedId, setSelectedId] = useState<string>(fraudGraph.rootId);
  const [pathOnly, setPathOnly] = useState<boolean>(false);
  const [hoverId, setHoverId] = useState<string | null>(null);
  const [edgeKey, setEdgeKey] = useState<string | null>(null);

  // Zoom & Pan state
  const [zoom, setZoom] = useState<number>(1);
  const [pan, setPan] = useState<{ x: number; y: number }>({ x: 0, y: 0 });
  const [isDragging, setIsDragging] = useState<boolean>(false);
  const [dragStart, setDragStart] = useState<{ x: number; y: number }>({ x: 0, y: 0 });
  const [draggedNodeId, setDraggedNodeId] = useState<string | null>(null);
  const [nodePositions, setNodePositions] = useState<Record<string, { x: number; y: number }>>({});

  // Initialize node positions based on fixture coordinates
  useEffect(() => {
    const posMap: Record<string, { x: number; y: number }> = {};
    fraudGraph.entities.forEach((e) => {
      posMap[e.id] = { x: e.x * 10, y: e.y * 5.6 };
    });
    setNodePositions(posMap);
  }, [fraudGraph]);

  const visible = useMemo(() => {
    const entities = fraudGraph.entities.filter((e) => e.hop <= hops);
    const ids = new Set(entities.map((e) => e.id));
    const relationships = fraudGraph.relationships.filter(
      (r) => ids.has(r.source) && ids.has(r.target) && (!pathOnly || r.onDecisionPath),
    );
    return { entities, relationships, ids };
  }, [fraudGraph, hops, pathOnly]);

  const selected =
    fraudGraph.entities.find((e) => e.id === selectedId && visible.ids.has(e.id)) ??
    fraudGraph.entities[0]!;

  const selectedEdge =
    visible.relationships.find((r) => `${r.source}-${r.target}` === edgeKey) ?? null;

  const focusId = hoverId ?? selected.id;

  const connectedRelationships = fraudGraph.relationships
    .filter((r) => r.source === selected.id || r.target === selected.id)
    .map((r) => ({
      id: r.source === selected.id ? r.target : r.source,
      label: r.label,
      direction: (r.source === selected.id ? "out" : "in") as "in" | "out",
    }));

  const neighbours = useMemo(() => {
    return new Set(
      fraudGraph.relationships
        .filter((r) => r.source === focusId || r.target === focusId)
        .map((r) => (r.source === focusId ? r.target : r.source)),
    );
  }, [fraudGraph, focusId]);

  const getNodePos = useCallback(
    (id: string) => {
      if (nodePositions[id]) return nodePositions[id];
      const e = fraudGraph.entities.find((n) => n.id === id);
      if (e) return { x: e.x * 10, y: e.y * 5.6 };
      return { x: 500, y: 300 };
    },
    [nodePositions, fraudGraph],
  );

  // Pan Handlers
  const handleMouseDown = (e: React.MouseEvent<SVGSVGElement>) => {
    if (draggedNodeId) return;
    setIsDragging(true);
    setDragStart({ x: e.clientX - pan.x, y: e.clientY - pan.y });
  };

  const handleMouseMove = (e: React.MouseEvent<SVGSVGElement>) => {
    if (draggedNodeId) {
      const rect = e.currentTarget.getBoundingClientRect();
      const newX = (e.clientX - rect.left - pan.x) / zoom;
      const newY = (e.clientY - rect.top - pan.y) / zoom;
      setNodePositions((prev) => ({
        ...prev,
        [draggedNodeId]: { x: Math.max(50, Math.min(950, newX)), y: Math.max(50, Math.min(550, newY)) },
      }));
      return;
    }
    if (isDragging) {
      setPan({ x: e.clientX - dragStart.x, y: e.clientY - dragStart.y });
    }
  };

  const handleMouseUp = () => {
    setIsDragging(false);
    setDraggedNodeId(null);
  };

  const handleZoomIn = () => setZoom((z) => Math.min(2.5, z + 0.2));
  const handleZoomOut = () => setZoom((z) => Math.max(0.4, z - 0.2));
  const handleResetView = () => {
    setZoom(1);
    setPan({ x: 0, y: 0 });
  };

  const handleCenterOnNode = (id: string) => {
    const pos = getNodePos(id);
    setPan({ x: 500 - pos.x * zoom, y: 300 - pos.y * zoom });
  };

  const handleCreateCase = (entityId: string) => {
    toast.success(`Case created for entity ${entityId}`, {
      description: "Priority P1 investigation opened with 3-hop graph snapshot attached.",
    });
  };

  const getNodeIcon = (type: string) => {
    switch (type) {
      case "CUSTOMER":
        return <User className="size-3.5" />;
      case "DEVICE":
        return <Smartphone className="size-3.5" />;
      case "IP":
        return <Globe className="size-3.5" />;
      case "CARD":
        return <CreditCard className="size-3.5" />;
      case "ACCOUNT":
        return <Building2 className="size-3.5" />;
      default:
        return <Share2 className="size-3.5" />;
    }
  };

  return (
    <div className="mx-auto w-full max-w-[1560px] px-5 py-5 lg:px-8 space-y-5">
      {/* ------------------------------------------------ Header */}
      <div className="flex flex-wrap items-baseline justify-between gap-3 border-b border-border pb-4">
        <div className="min-w-0">
          <div className="flex items-center gap-2">
            <span className="border border-navy bg-navy/10 px-2 py-0.5 font-mono text-[10px] font-bold text-navy uppercase tracking-[0.06em]">
              GRAPH TOPOLOGY &amp; GNN SHADOW
            </span>
            <DemoTag />
            <span className="border border-shadow-intel/40 bg-shadow-intel-surface px-1.5 py-0.5 font-mono text-[9.5px] font-bold text-shadow-intel">
              GraphSAGE: NON-ENFORCING (0% AUTHORITY)
            </span>
          </div>
          <h1 className="mt-1 font-sans text-[25px] font-extrabold leading-[1.15] tracking-[-0.045em] text-foreground">
            Syndicate Neighbourhood of <Mono className="text-[20px] font-bold text-foreground">{fraudGraph.rootId}</Mono>
          </h1>
          <p className="mt-0.5 font-sans text-[12px] text-muted-foreground">
            3-hop in-memory BFS &amp; inductive GraphSAGE embeddings linking customer, headless device emulator, proxy ASN 13335, and cashout mule accounts.
          </p>
        </div>

        {/* View Controls & Filters */}
        <div className="flex flex-wrap items-center gap-2 font-mono text-[11px]">
          <div className="flex items-center border border-border bg-surface p-0.5">
            <span className="px-2 text-[10px] text-muted-foreground uppercase">Hops:</span>
            {[1, 2, 3].map((h) => (
              <button
                key={h}
                type="button"
                onClick={() => setHops(h)}
                className={cn(
                  "px-2.5 py-1 transition-colors cursor-pointer",
                  hops === h
                    ? "bg-navy text-white font-bold"
                    : "text-muted-foreground hover:text-foreground",
                )}
              >
                {h} {h === 1 ? "Hop" : "Hops"}
              </button>
            ))}
          </div>

          <button
            type="button"
            onClick={() => setPathOnly((p) => !p)}
            className={cn(
              "flex items-center gap-1.5 border px-3 py-1 font-medium transition-colors cursor-pointer",
              pathOnly
                ? "border-blocked bg-blocked-surface text-blocked font-bold"
                : "border-border bg-surface text-muted-foreground hover:text-foreground",
            )}
          >
            <Sparkles className="size-3.5" />
            <span>{pathOnly ? "Decision Path Only" : "Show All Edges"}</span>
          </button>
        </div>
      </div>

      {/* ------------------------------------------------ Main Workspace Grid */}
      <div className="grid gap-5 lg:grid-cols-[1fr_380px]">
        {/* Left Column: Interactive Graph Viewport */}
        <div className="relative flex flex-col border border-border bg-card shadow-xs overflow-hidden h-[640px]">
          {/* Viewport Floating Toolbar */}
          <div className="absolute top-3 left-3 z-10 flex items-center gap-1 border border-border bg-surface/90 backdrop-blur-xs p-1 shadow-xs font-mono text-[11px]">
            <button
              type="button"
              onClick={handleZoomIn}
              title="Zoom In"
              className="p-1.5 text-muted-foreground hover:text-foreground hover:bg-secondary cursor-pointer"
            >
              <ZoomIn className="size-3.5" />
            </button>
            <button
              type="button"
              onClick={handleZoomOut}
              title="Zoom Out"
              className="p-1.5 text-muted-foreground hover:text-foreground hover:bg-secondary cursor-pointer"
            >
              <ZoomOut className="size-3.5" />
            </button>
            <button
              type="button"
              onClick={handleResetView}
              title="Reset View"
              className="p-1.5 text-muted-foreground hover:text-foreground hover:bg-secondary cursor-pointer"
            >
              <Maximize2 className="size-3.5" />
            </button>
            <button
              type="button"
              onClick={() => handleCenterOnNode(selected.id)}
              title="Center on Selected Node"
              className="p-1.5 text-navy hover:bg-secondary cursor-pointer font-bold flex items-center gap-1 text-[10.5px]"
            >
              <Compass className="size-3.5" />
              <span>Center</span>
            </button>
            <span className="border-l border-border pl-2 pr-1 text-[10px] text-muted-foreground">
              {Math.round(zoom * 100)}%
            </span>
          </div>

          {/* Graph Legend Overlay */}
          <div className="absolute top-3 right-3 z-10 border border-border bg-surface/90 backdrop-blur-xs p-2.5 shadow-xs font-mono text-[10px] space-y-1">
            <div className="text-muted-foreground uppercase font-bold text-[9px] mb-1">Entity Risk Class</div>
            {risks.map((r) => {
              const meta = entityRiskMeta(r);
              return (
                <div key={r} className="flex items-center gap-2">
                  <span className={cn("size-2 rounded-full", meta.bg)} />
                  <span className="text-foreground">{r}</span>
                </div>
              );
            })}
          </div>

          {/* SVG Interactive Canvas */}
          <svg
            className="w-full h-full cursor-grab active:cursor-grabbing select-none bg-background/50"
            viewBox="0 0 1000 600"
            onMouseDown={handleMouseDown}
            onMouseMove={handleMouseMove}
            onMouseUp={handleMouseUp}
          >
            <defs>
              <pattern id="graph-grid" width="40" height="40" patternUnits="userSpaceOnUse">
                <path d="M 40 0 L 0 0 0 40" fill="none" stroke="currentColor" strokeWidth="0.5" className="text-border/40" />
              </pattern>
              <marker
                id="arrow-default"
                viewBox="0 0 10 10"
                refX="28"
                refY="5"
                markerWidth="6"
                markerHeight="6"
                orient="auto-start-reverse"
              >
                <path d="M 0 1.5 L 8 5 L 0 8.5 z" className="fill-border-strong" />
              </marker>
              <marker
                id="arrow-active"
                viewBox="0 0 10 10"
                refX="28"
                refY="5"
                markerWidth="6"
                markerHeight="6"
                orient="auto-start-reverse"
              >
                <path d="M 0 1.5 L 8 5 L 0 8.5 z" className="fill-blocked" />
              </marker>
            </defs>

            {/* Grid Background */}
            <rect width="1000" height="600" fill="url(#graph-grid)" />

            <g transform={`translate(${pan.x}, ${pan.y}) scale(${zoom})`}>
              {/* Edges */}
              {visible.relationships.map((r) => {
                const s = getNodePos(r.source);
                const t = getNodePos(r.target);
                const isSelected = `${r.source}-${r.target}` === edgeKey;
                const isOnDecisionPath = r.onDecisionPath;
                const isIncidentFocused = r.source === focusId || r.target === focusId;

                const midX = (s.x + t.x) / 2;
                const midY = (s.y + t.y) / 2;

                return (
                  <g
                    key={`${r.source}-${r.target}`}
                    onClick={(e) => {
                      e.stopPropagation();
                      setEdgeKey(`${r.source}-${r.target}`);
                    }}
                    className="cursor-pointer group"
                  >
                    <line
                      x1={s.x}
                      y1={s.y}
                      x2={t.x}
                      y2={t.y}
                      strokeWidth={isSelected || isIncidentFocused ? 2.5 : isOnDecisionPath ? 2 : 1.2}
                      markerEnd={isOnDecisionPath || isIncidentFocused ? "url(#arrow-active)" : "url(#arrow-default)"}
                      className={cn(
                        "transition-colors",
                        isSelected
                          ? "stroke-navy stroke-dasharray-none"
                          : isOnDecisionPath
                            ? "stroke-blocked"
                            : isIncidentFocused
                              ? "stroke-amber-intel"
                              : "stroke-border-strong group-hover:stroke-foreground",
                      )}
                      strokeDasharray={isOnDecisionPath ? "4 2" : undefined}
                    />
                    {/* Edge Label Badge */}
                    <rect
                      x={midX - 35}
                      y={midY - 9}
                      width={70}
                      height={18}
                      className="fill-surface stroke-border"
                      rx="2"
                    />
                    <text
                      x={midX}
                      y={midY + 3.5}
                      textAnchor="middle"
                      className="font-mono text-[9px] font-bold fill-muted-foreground group-hover:fill-foreground pointer-events-none"
                    >
                      {r.label}
                    </text>
                  </g>
                );
              })}

              {/* Nodes */}
              {visible.entities.map((e) => {
                const pos = getNodePos(e.id);
                const isSelected = e.id === selected.id;
                const isHovered = e.id === hoverId;
                const isNeighbour = neighbours.has(e.id);
                const isRoot = e.id === fraudGraph.rootId;
                const meta = entityRiskMeta(e.risk);

                return (
                  <g
                    key={e.id}
                    transform={`translate(${pos.x}, ${pos.y})`}
                    onClick={(eEv) => {
                      eEv.stopPropagation();
                      setSelectedId(e.id);
                      setEdgeKey(null);
                    }}
                    onMouseDown={(eEv) => {
                      eEv.stopPropagation();
                      setDraggedNodeId(e.id);
                    }}
                    onMouseEnter={() => setHoverId(e.id)}
                    onMouseLeave={() => setHoverId(null)}
                    className="cursor-pointer group"
                  >
                    {/* Radar Pulse on high risk nodes */}
                    {(e.risk === "CONFIRMED_FRAUD" || e.risk === "SUSPECT") && (
                      <circle
                        r="30"
                        className="fill-none stroke-blocked/30 animate-ping"
                        strokeWidth="1.5"
                      />
                    )}

                    {/* Outer Selection Ring */}
                    <circle
                      r={isRoot ? 24 : 20}
                      className={cn(
                        "transition-all",
                        isSelected
                          ? "fill-navy/15 stroke-navy stroke-2"
                          : isHovered
                            ? "fill-surface stroke-foreground stroke-2"
                            : isNeighbour
                              ? "fill-surface stroke-amber-intel stroke-1.5"
                              : "fill-surface stroke-border stroke-1",
                      )}
                    />

                    {/* Inner Risk Ring */}
                    <circle
                      r={isRoot ? 17 : 14}
                      className={cn("transition-colors", meta.bg, meta.border)}
                      strokeWidth="2"
                    />

                    {/* Node Identifier Text */}
                    <text
                      y={isRoot ? 36 : 30}
                      textAnchor="middle"
                      className={cn(
                        "font-mono text-[10.5px] font-bold tracking-tight pointer-events-none transition-colors",
                        isSelected ? "fill-navy text-[11px]" : "fill-foreground",
                      )}
                    >
                      {e.id}
                    </text>

                    {/* Subtitle / Hop badge */}
                    <text
                      y={isRoot ? 47 : 41}
                      textAnchor="middle"
                      className="font-sans text-[9px] fill-muted-foreground pointer-events-none"
                    >
                      {e.type} · Hop {e.hop}
                    </text>
                  </g>
                );
              })}
            </g>
          </svg>

          {/* Minimap Viewport Indicator in Bottom Left */}
          <div className="absolute bottom-3 left-3 z-10 border border-border bg-surface/90 backdrop-blur-xs p-2 shadow-xs font-mono text-[10px] space-y-1 hidden sm:block">
            <div className="flex items-center gap-1.5 font-bold text-foreground">
              <Layers className="size-3 text-navy" />
              <span>Visible Topology: {visible.entities.length} Nodes · {visible.relationships.length} Edges</span>
            </div>
            <div className="text-muted-foreground text-[9px]">
              Drag nodes to rearrange · Click node to inspect details
            </div>
          </div>
        </div>

        {/* Right Column: Slide-Over Inspector Panel */}
        <aside className="border border-border bg-card p-5 shadow-xs flex flex-col justify-between space-y-4">
          <div className="space-y-4">
            {/* Header Title */}
            <div className="flex items-center justify-between border-b border-border pb-3">
              <div className="flex items-center gap-2">
                <span className="font-mono text-[10px] font-bold text-navy uppercase tracking-wider">
                  ENTITY FORENSIC INSPECTOR
                </span>
              </div>
              <StatusPill tone={selected.risk === "CONFIRMED_FRAUD" ? "blocked" : selected.risk === "SUSPECT" ? "shadow" : "authoritative"}>
                {selected.risk}
              </StatusPill>
            </div>

            {/* Selected Node Details Card */}
            <div className="border border-border bg-surface p-3.5 space-y-3 font-mono text-[11.5px]">
              <div className="flex items-start justify-between gap-2">
                <div>
                  <div className="font-bold text-foreground text-[13px] flex items-center gap-1.5">
                    {getNodeIcon(selected.type)}
                    <span>{selected.id}</span>
                  </div>
                  <div className="mt-0.5 text-muted-foreground text-[11px] font-sans">
                    {selected.label}
                  </div>
                </div>
                <span className="border border-border bg-secondary/50 px-2 py-0.5 text-[10px] font-bold">
                  HOP {selected.hop}
                </span>
              </div>

              <div className="grid grid-cols-2 gap-2 text-[11px] border-t border-border pt-2.5">
                <div>
                  <span className="text-muted-foreground block text-[9.5px] uppercase">Entity Type:</span>
                  <span className="font-bold text-foreground">{selected.type}</span>
                </div>
                <div>
                  <span className="text-muted-foreground block text-[9.5px] uppercase">First Observed:</span>
                  <span className="text-foreground">{selected.firstSeen}</span>
                </div>
              </div>

              {/* Attributes Key-Value Table */}
              <div className="border-t border-border pt-2.5 space-y-1">
                <span className="text-muted-foreground block text-[9.5px] uppercase font-bold">Entity Attributes:</span>
                {selected.attributes.map(([k, v]) => (
                  <div key={k} className="flex justify-between text-[11px]">
                    <span className="text-muted-foreground">{k}:</span>
                    <span className="font-bold text-foreground">{v}</span>
                  </div>
                ))}
              </div>

              {/* Signals and Flags */}
              {selected.signals && selected.signals.length > 0 && (
                <div className="border-t border-border pt-2.5 space-y-1">
                  <span className="text-muted-foreground block text-[9.5px] uppercase font-bold text-blocked">Forensic Signals:</span>
                  <ul className="space-y-1 font-sans text-[11px] text-muted-foreground list-disc list-inside">
                    {selected.signals.map((sig, idx) => (
                      <li key={idx} className="leading-snug">{sig}</li>
                    ))}
                  </ul>
                </div>
              )}
            </div>

            {/* Connected Relationships Edge List */}
            <div className="space-y-2 font-mono text-[11px]">
              <div className="flex items-center justify-between text-[10px] font-bold text-muted-foreground uppercase">
                <span>Adjacent Edges ({connectedRelationships.length})</span>
                <span>Hop Neighborhood</span>
              </div>
              <div className="space-y-1.5 max-h-[160px] overflow-y-auto">
                {connectedRelationships.map((rel, idx) => (
                  <div
                    key={idx}
                    onClick={() => {
                      setSelectedId(rel.id);
                    }}
                    className="border border-border/70 bg-surface p-2 flex items-center justify-between cursor-pointer hover:bg-secondary/60 transition-colors"
                  >
                    <div className="flex items-center gap-1.5 truncate">
                      <span className="text-muted-foreground font-normal">{rel.direction === "out" ? "→" : "←"}</span>
                      <span className="font-bold text-navy">{rel.label}:</span>
                      <span className="truncate text-foreground">{rel.id}</span>
                    </div>
                    <ArrowRight className="size-3 text-muted-foreground shrink-0" />
                  </div>
                ))}
              </div>
            </div>
          </div>

          {/* Action Bar */}
          <div className="border-t border-border pt-3 space-y-2 font-mono text-[11px]">
            <button
              type="button"
              onClick={() => handleCreateCase(selected.id)}
              className="w-full border border-navy bg-navy py-2 text-white font-bold hover:bg-navy/90 transition-opacity flex items-center justify-center gap-1.5 cursor-pointer"
            >
              <PlusCircle className="size-3.5" />
              <span>Open Forensic Case for {selected.id}</span>
            </button>
            <Link
              to="/cases/CASE-88419"
              className="w-full border border-border bg-surface py-2 text-center text-foreground font-medium hover:bg-secondary transition-colors block"
            >
              View Linked Case Dossier (CASE-88419) →
            </Link>
          </div>
        </aside>
      </div>
    </div>
  );
}
