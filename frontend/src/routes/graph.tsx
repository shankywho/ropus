import { createFileRoute, Link } from "@tanstack/react-router";
import { useMemo, useState } from "react";
import { cn } from "@/lib/utils";
import {
  DemoTag,
  EntityNode,
  GraphInspector,
  Mono,
  RelationshipInspector,
  entityRiskMeta,
} from "@/components/ropus/core";
import { useSuspenseQuery } from "@tanstack/react-query";
import { graphQuery } from "@/lib/ropus/api";
import { blockedDecision } from "@/lib/ropus/fixtures";
import type { EntityRisk } from "@/lib/ropus/contracts";

export const Route = createFileRoute("/graph")({
  head: () => ({
    meta: [
      { title: "Fraud Graph — ROPUS" },
      {
        name: "description",
        content:
          "Entity neighbourhood for a blocked wire: customer, device, network origin, payout account and linked accounts.",
      },
      { property: "og:title", content: "Fraud Graph — ROPUS" },
      {
        property: "og:description",
        content: "Traverse the connected entities behind a single risk decision.",
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

function FraudGraphPage() {
  const { data: fraudGraph } = useSuspenseQuery(graphQuery(blockedDecision.decisionId));
  const [hops, setHops] = useState(3);
  const [selectedId, setSelectedId] = useState(fraudGraph.rootId);
  const [pathOnly, setPathOnly] = useState(false);
  const [hoverId, setHoverId] = useState<string | null>(null);
  const [edgeKey, setEdgeKey] = useState<string | null>(null);

  const visible = useMemo(() => {
    const entities = fraudGraph.entities.filter((e) => e.hop <= hops);
    const ids = new Set(entities.map((e) => e.id));
    const relationships = fraudGraph.relationships.filter(
      (r) => ids.has(r.source) && ids.has(r.target) && (!pathOnly || r.onDecisionPath),
    );
    return { entities, relationships, ids };
  }, [hops, pathOnly]);

  const selected =
    fraudGraph.entities.find((e) => e.id === selectedId && visible.ids.has(e.id)) ??
    fraudGraph.entities[0]!;

  const selectedEdge =
    visible.relationships.find((r) => `${r.source}-${r.target}` === edgeKey) ?? null;

  const focusId = hoverId ?? selected.id;

  const relationships = fraudGraph.relationships
    .filter((r) => r.source === selected.id || r.target === selected.id)
    .map((r) => ({
      id: r.source === selected.id ? r.target : r.source,
      label: r.label,
      direction: (r.source === selected.id ? "out" : "in") as "in" | "out",
    }));

  const neighbours = new Set(
    fraudGraph.relationships
      .filter((r) => r.source === focusId || r.target === focusId)
      .map((r) => (r.source === focusId ? r.target : r.source)),
  );

  const pos = (id: string) => {
    const e = fraudGraph.entities.find((n) => n.id === id)!;
    return { x: e.x * 10, y: e.y * 5.6 };
  };

  return (
    <div className="mx-auto w-full max-w-[1560px] px-5 py-5 lg:px-8">
      <div className="flex flex-wrap items-baseline justify-between gap-3 border-b border-border pb-4">
        <div className="min-w-0">
          <div className="flex items-center gap-2">
            <span className="font-mono text-[10px] font-medium tracking-[0.12em] text-muted-foreground uppercase">
              GRAPH TOPOLOGY
            </span>
            <DemoTag />
          </div>
          <h1 className="mt-1 font-sans text-[25px] font-extrabold leading-[1.15] tracking-[-0.045em] text-foreground">
            Neighbourhood of <Mono className="text-[20px] font-bold text-foreground">{fraudGraph.rootId}</Mono>
          </h1>
          <p className="mt-1 font-sans text-[12px] text-muted-foreground">
            Entities connected to decision{" "}
            <Link
              to="/decisions/$decisionId"
              params={{ decisionId: fraudGraph.decisionId }}
              className="font-mono text-[11.5px] text-navy font-medium hover:underline"
            >
              {fraudGraph.decisionId}
            </Link>
            . {visible.entities.length} entities · {visible.relationships.length} relationships.
          </p>
        </div>

        <div className="flex items-center gap-4">
          <div className="flex" role="group" aria-label="Traversal depth">
            {[1, 2, 3].map((h) => (
              <button
                key={h}
                onClick={() => setHops(h)}
                aria-pressed={hops === h}
                className={cn(
                  "-ml-px border border-border px-2.5 py-1 font-mono text-[11px] first:ml-0 transition-colors",
                  hops === h
                    ? "border-navy bg-navy text-white font-medium"
                    : "bg-surface text-muted-foreground hover:bg-secondary hover:text-foreground",
                )}
              >
                {h}-hop
              </button>
            ))}
          </div>
          <label className="flex items-center gap-2 font-sans text-[12px] text-muted-foreground cursor-pointer">
            <input
              type="checkbox"
              checked={pathOnly}
              onChange={(e) => setPathOnly(e.target.checked)}
              className="size-3.5 accent-navy"
            />
            Decision path only
          </label>
        </div>
      </div>

      <div className="mt-6 grid gap-8 xl:grid-cols-[1fr_320px] xl:gap-8">
        <section className="min-w-0">
          <svg
            viewBox="0 0 1000 560"
            preserveAspectRatio="xMidYMid meet"
            className="h-[calc(100vh-240px)] max-h-[850px] min-h-[500px] w-full border border-border bg-surface shadow-2xs"
            role="img"
            aria-label="Fraud entity graph"
          >
            <defs>
              <pattern id="ropus-grid" width="28" height="28" patternUnits="userSpaceOnUse">
                <path d="M28 0H0V28" fill="none" stroke="rgba(28, 43, 69, 0.05)" strokeWidth="1" />
              </pattern>
            </defs>
            <rect width="1000" height="560" fill="url(#ropus-grid)" />
            {visible.relationships.map((r) => {
              const key = `${r.source}-${r.target}`;
              const a = pos(r.source);
              const b = pos(r.target);
              const active = r.source === focusId || r.target === focusId || key === edgeKey;
              const mx = (a.x + b.x) / 2;
              const my = (a.y + b.y) / 2;
              const vertical = Math.abs(a.x - b.x) < 40;
              return (
                <g
                  key={key}
                  className="cursor-pointer"
                  role="button"
                  tabIndex={0}
                  aria-label={`${r.source} ${r.label} ${r.target}`}
                  onClick={() => setEdgeKey(key)}
                  onKeyDown={(e) => {
                    if (e.key === "Enter" || e.key === " ") {
                      e.preventDefault();
                      setEdgeKey(key);
                    }
                  }}
                >
                  <line x1={a.x} y1={a.y} x2={b.x} y2={b.y} stroke="transparent" strokeWidth={14} />
                  <line
                    x1={a.x}
                    y1={a.y}
                    x2={b.x}
                    y2={b.y}
                    stroke={
                      key === edgeKey
                        ? "#F5A524"
                        : active
                          ? "#1D2839"
                          : "#DCD5CB"
                    }
                    strokeWidth={active ? 1.5 : 0.8}
                    opacity={r.onDecisionPath ? (active ? 1 : 0.8) : active ? 0.85 : 0.4}
                    strokeDasharray={r.onDecisionPath ? undefined : "4 4"}
                    className="transition-opacity duration-150"
                  />
                  {active && (
                    <text
                      x={vertical ? mx + 8 : mx}
                      y={vertical ? my + 60 : my - 7}
                      textAnchor={vertical ? "start" : "middle"}
                      fontSize="11"
                      letterSpacing="0.06em"
                      fontFamily="var(--font-mono)"
                      fill="#647182"
                    >
                      {r.label}
                    </text>
                  )}
                </g>
              );
            })}
            {visible.entities.map((e) => (
              <EntityNode
                key={e.id}
                entity={e}
                selected={e.id === selected.id}
                highlighted={e.id === hoverId || neighbours.has(e.id)}
                dimmed={e.id !== focusId && e.id !== selected.id && !neighbours.has(e.id)}
                onSelect={(id) => {
                  setSelectedId(id);
                  setEdgeKey(null);
                }}
                onHover={setHoverId}
              />
            ))}
          </svg>

          <div className="flex flex-wrap items-center gap-x-6 gap-y-2 pt-3 border-t border-border mt-3">
            {risks.map((r) => {
              const meta = entityRiskMeta(r);
              return (
                <span
                  key={r}
                  className="flex items-center gap-1.5 font-mono text-[10px] tracking-[0.06em] text-muted-foreground uppercase"
                >
                  <span aria-hidden className="h-3 w-1" style={{ background: meta.fill }} />
                  {meta.label}
                </span>
              );
            })}
            <span className="flex items-center gap-1.5 font-mono text-[10px] tracking-[0.06em] text-muted-foreground uppercase">
              <span aria-hidden className="h-px w-5 bg-border-strong" /> Decision Path
            </span>
            <span className="flex items-center gap-1.5 font-mono text-[10px] tracking-[0.06em] text-muted-foreground uppercase">
              <span
                aria-hidden
                className="h-px w-5"
                style={{
                  backgroundImage:
                    "repeating-linear-gradient(90deg,var(--border-strong) 0 4px,transparent 4px 8px)",
                }}
              />
              Contextual
            </span>
          </div>
        </section>

        <aside aria-label="Entity inspector" className="min-w-0 border border-border bg-card p-4 shadow-2xs">
          {selectedEdge ? (
            <RelationshipInspector
              relationship={selectedEdge}
              onSelect={(id) => {
                setSelectedId(id);
                setEdgeKey(null);
              }}
              onClear={() => setEdgeKey(null)}
            />
          ) : (
            <GraphInspector
              entity={selected}
              relationships={relationships}
              onSelect={setSelectedId}
            />
          )}
        </aside>
      </div>
    </div>
  );
}
