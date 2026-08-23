import { useState, type ReactNode } from "react";
import { cn } from "@/lib/utils";
import { DATA_SOURCE } from "@/lib/ropus/fixtures";
import type {
  EvidenceItem,
  FactorSource,
  RiskFactorRecord,
  Verdict,
} from "@/lib/ropus/contracts";

/* ------------------------------------------------------------------ mono id */

export function Mono({ children, className }: { children: ReactNode; className?: string | undefined }) {
  return <span className={cn("font-mono text-[12.5px] tabular", className)}>{children}</span>;
}

/* ---------------------------------------------------------------- demo flag */

export function DemoTag({ className }: { className?: string }) {
  return (
    <span
      title={DATA_SOURCE.note}
      className={cn(
        "inline-flex items-center gap-1.5 border border-border-strong px-1.5 py-0.5 font-mono text-[10px] tracking-[0.08em] text-muted-foreground uppercase",
        className,
      )}
    >
      <span aria-hidden className="size-1 rounded-full bg-warning" />
      Demo data
    </span>
  );
}

/* -------------------------------------------------------------- risk score */

const scoreBand = (v: number) =>
  v >= 0.8
    ? { tone: "text-block", label: "Critical" }
    : v >= 0.55
      ? { tone: "text-warning", label: "Elevated" }
      : v >= 0.35
        ? { tone: "text-primary", label: "Moderate" }
        : { tone: "text-approve", label: "Low" };

/**
 * The single score representation. `hero` is used once per screen — nothing
 * else may compete with it visually.
 */
export function RiskScore({
  value,
  size = "inline",
  showBand = false,
}: {
  value: number;
  size?: "inline" | "hero";
  showBand?: boolean;
}) {
  const band = scoreBand(value);
  if (size === "hero") {
    return (
      <div>
        <div
          className={cn(
            "font-mono text-[44px] leading-none font-bold tabular transition-colors duration-200",
            band.tone,
          )}
        >
          {value.toFixed(2)}
        </div>
        {showBand && (
          <div className="mt-1.5 text-[11px] font-semibold tracking-[0.08em] text-muted-foreground uppercase">
            {band.label} risk
          </div>
        )}
      </div>
    );
  }
  return <span className={cn("font-mono text-[12.5px] font-semibold tabular", band.tone)}>{value.toFixed(2)}</span>;
}

/* ------------------------------------------------------------ verdict badge */

const verdictStyles: Record<Verdict, string> = {
  APPROVE: "border-approve/40 text-approve",
  REVIEW: "border-warning/40 text-warning",
  CHALLENGE: "border-primary/40 text-primary",
  BLOCK: "border-block/45 text-block",
};

export function VerdictBadge({ verdict, size = "sm" }: { verdict: Verdict; size?: "sm" | "lg" }) {
  return (
    <span
      className={cn(
        "inline-flex items-center border font-semibold tracking-[0.08em] uppercase",
        verdictStyles[verdict],
        size === "sm" ? "px-1.5 py-0.5 text-[10.5px]" : "px-2.5 py-1 text-[12px]",
      )}
    >
      {verdict}
    </span>
  );
}

/* --------------------------------------------------------- status indicator */

export function StatusIndicator({
  state,
  label,
}: {
  state: "OK" | "DEGRADED" | "DOWN";
  label?: string;
}) {
  const map = {
    OK: "bg-approve text-approve",
    DEGRADED: "bg-warning text-warning",
    DOWN: "bg-block text-block",
  } as const;
  return (
    <span className={cn("inline-flex items-center gap-1.5 text-[11px] font-semibold tracking-[0.06em]", map[state].split(" ")[1])}>
      <span aria-hidden className={cn("size-1.5", map[state].split(" ")[0])} />
      {label ?? state}
    </span>
  );
}

/* ------------------------------------------------------------- source tag */

const sourceLabels: Record<FactorSource, string> = {
  RULES: "Rules",
  ML: "ML",
  THREAT_INTEL: "Threat intel",
  GRAPH: "Graph",
  DEVICE: "Device",
};

export function SourceTag({ source }: { source: FactorSource }) {
  return (
    <span className="shrink-0 font-mono text-[10.5px] tracking-[0.06em] text-muted-foreground uppercase">
      {sourceLabels[source]}
    </span>
  );
}

/* ------------------------------------------------------------- risk factor */

/** Additive contribution row. One bar style, everywhere. */
export function RiskFactor({ factor, max }: { factor: RiskFactorRecord; max: number }) {
  return (
    <div className="border-t border-border py-2.5 first:border-t-0">
      <div className="flex items-baseline justify-between gap-4">
        <span className="text-[13px] font-medium">{factor.label}</span>
        <div className="flex shrink-0 items-baseline gap-3">
          <SourceTag source={factor.source} />
          <Mono className="w-11 text-right font-semibold">+{factor.weight.toFixed(2)}</Mono>
        </div>
      </div>
      {factor.detail && (
        <p className="mt-0.5 font-mono text-[11.5px] text-muted-foreground tabular">{factor.detail}</p>
      )}
      <div className="mt-2 h-[2px] w-full bg-neutral-surface">
        <div
          className="h-full bg-foreground/60 transition-[width] duration-200"
          style={{ width: `${Math.max(2, (factor.weight / max) * 100)}%` }}
        />
      </div>
    </div>
  );
}

export function RiskFactorList({ factors }: { factors: RiskFactorRecord[] }) {
  const sorted = [...factors].sort((a, b) => b.weight - a.weight);
  const max = Math.max(...sorted.map((f) => f.weight), 0.01);
  return (
    <div>
      {sorted.map((f) => (
        <RiskFactor key={f.id} factor={f} max={max} />
      ))}
    </div>
  );
}

/* ---------------------------------------------------------- evidence list */

function RecommendedAction({ item }: { item: EvidenceItem }) {
  const [state, setState] = useState<"idle" | "confirming" | "authorized">("idle");
  if (!item.action) return null;

  if (state === "authorized") {
    return (
      <span className="font-mono text-[11px] tracking-[0.06em] text-approve uppercase">
        Authorized by analyst
      </span>
    );
  }
  if (state === "confirming") {
    return (
      <span className="flex items-center gap-2">
        <button
          type="button"
          onClick={() => setState("authorized")}
          className={cn(
            "px-2 py-1 text-[12px] font-semibold text-white transition-colors duration-150",
            item.action.destructive ? "bg-block" : "bg-primary",
          )}
        >
          Confirm
        </button>
        <button
          type="button"
          onClick={() => setState("idle")}
          className="px-2 py-1 text-[12px] font-medium text-muted-foreground hover:text-foreground"
        >
          Cancel
        </button>
      </span>
    );
  }
  return (
    <button
      type="button"
      onClick={() => setState("confirming")}
      className="border border-border-strong px-2 py-1 text-[12px] font-medium transition-colors duration-150 hover:bg-accent"
    >
      {item.action.label}
    </button>
  );
}

const kindMeta = {
  OBSERVED: {
    title: "Observed",
    note: "Recorded by the platform.",
    rule: "border-t-approve",
  },
  INFERRED: {
    title: "Inferred",
    note: "Model, graph and agent interpretation.",
    rule: "border-t-primary",
  },
  RECOMMENDED: {
    title: "Recommended",
    note: "Proposed to a human. Nothing executes without confirmation.",
    rule: "border-t-warning",
  },
} as const;

/**
 * Three-way explainability split. Deliberately not three identical cards:
 * observed is a plain fact list, inferred is labelled model output, recommended
 * is a list of actions that require explicit confirmation.
 */
export function EvidenceList({ items, kind }: { items: EvidenceItem[]; kind: keyof typeof kindMeta }) {
  const meta = kindMeta[kind];
  const rows = items.filter((i) => i.kind === kind);

  return (
    <section className={cn("border-t-2 pt-3", meta.rule)}>
      <div className="flex items-baseline justify-between gap-3">
        <h3 className="text-[11px] font-bold tracking-[0.08em] uppercase">{meta.title}</h3>
        {kind !== "OBSERVED" && (
          <span className="font-mono text-[10px] tracking-[0.08em] text-muted-foreground uppercase">
            AI generated
          </span>
        )}
      </div>
      <p className="mt-1 text-[11.5px] text-muted-foreground">{meta.note}</p>

      <ul className="mt-3">
        {rows.map((item) => (
          <li key={item.id} className="border-t border-border py-2.5 first:border-t-0 first:pt-0">
            {kind === "OBSERVED" ? (
              <div className="flex items-baseline justify-between gap-3">
                <span className="text-[13px] leading-relaxed">{item.text}</span>
                {item.source && <SourceTag source={item.source} />}
              </div>
            ) : kind === "INFERRED" ? (
              <p className="text-[13px] leading-relaxed text-foreground/90">{item.text}</p>
            ) : (
              <div className="flex flex-wrap items-center justify-between gap-3">
                <span className="text-[13px] leading-relaxed">{item.text}</span>
                <RecommendedAction item={item} />
              </div>
            )}
          </li>
        ))}
        {rows.length === 0 && (
          <li className="py-2.5 text-[12.5px] text-muted-foreground">None recorded for this decision.</li>
        )}
      </ul>
    </section>
  );
}

/* ------------------------------------------------------------------ metric */

export function Metric({ label, value, sub }: { label: string; value: string; sub?: string }) {
  return (
    <div>
      <div className="font-mono text-[22px] leading-none font-semibold tabular">{value}</div>
      <div className="mt-2 text-[11.5px] text-muted-foreground">{label}</div>
      {sub && <div className="mt-0.5 font-mono text-[11px] text-muted-foreground tabular">{sub}</div>}
    </div>
  );
}

/* -------------------------------------------------------- decision summary */

/**
 * Compact technical record of the decision. Deliberately a definition table,
 * not a grid of KPI cards.
 */
export function DecisionSummary({ decision, amount }: { decision: import("@/lib/ropus/contracts").RiskDecision; amount: string }) {
  const rows: Array<[string, ReactNode]> = [
    ["Verdict", <VerdictBadge key="v" verdict={decision.verdict} />],
    ["Risk score", <Mono key="s">{decision.riskScore.toFixed(2)}</Mono>],
    ["Confidence", <Mono key="c">{(decision.confidence * 100).toFixed(1)}%</Mono>],
    ["Model", <Mono key="m">{`${decision.model}-${decision.modelVersion}`}</Mono>],
    ["Policy", <Mono key="p">{decision.policy}</Mono>],
    ["Latency", <Mono key="l">{decision.latencyMs.toFixed(1)} ms</Mono>],
    ["Amount", <Mono key="a">{`${amount} ${decision.currency}`}</Mono>],
    [
      "Rules",
      decision.rules.length ? (
        <div key="r" className="flex flex-col items-end gap-0.5">
          {decision.rules.map((r) => (
            <Mono key={r.id}>{r.id}</Mono>
          ))}
        </div>
      ) : (
        <span key="r0" className="text-[12.5px] text-muted-foreground">
          none triggered
        </span>
      ),
    ],
  ];

  return (
    <dl className="divide-y divide-border border-y border-border">
      {rows.map(([k, v]) => (
        <div key={k} className="flex items-baseline justify-between gap-6 py-1.5">
          <dt className="label-xs">{k}</dt>
          <dd className="text-right">{v}</dd>
        </div>
      ))}
    </dl>
  );
}

/* -------------------------------------------------------------- graph bits */

const entityRisk: Record<
  import("@/lib/ropus/contracts").EntityRisk,
  { fill: string; text: string; label: string }
> = {
  CLEAN: { fill: "var(--muted-foreground)", text: "text-muted-foreground", label: "Clean" },
  WATCH: { fill: "var(--challenge)", text: "text-challenge", label: "Watch" },
  SUSPECT: { fill: "var(--review)", text: "text-review", label: "Suspect" },
  CONFIRMED_FRAUD: { fill: "var(--block)", text: "text-block", label: "Confirmed fraud" },
};

export function entityRiskMeta(risk: import("@/lib/ropus/contracts").EntityRisk) {
  return entityRisk[risk];
}

/** One entity in the fraud graph. Rendered inside an SVG canvas. */
export function EntityNode({
  entity,
  selected,
  dimmed,
  highlighted,
  onSelect,
  onHover,
}: {
  entity: import("@/lib/ropus/contracts").GraphEntity;
  selected: boolean;
  dimmed: boolean;
  highlighted?: boolean;
  onSelect: (id: string) => void;
  onHover?: (id: string | null) => void;
}) {
  const meta = entityRisk[entity.risk];
  const x = entity.x * 10;
  const y = entity.y * 5.6;
  // Hierarchy: the root entity reads larger, first-hop next, context smallest.
  const half = entity.hop === 0 ? 21 : entity.hop === 1 ? 17 : 14;
  return (
    <g
      role="button"
      tabIndex={0}
      aria-label={`${entity.type} ${entity.id}`}
      aria-pressed={selected}
      onClick={() => onSelect(entity.id)}
      onMouseEnter={() => onHover?.(entity.id)}
      onMouseLeave={() => onHover?.(null)}
      onFocus={() => onHover?.(entity.id)}
      onBlur={() => onHover?.(null)}
      onKeyDown={(e) => {
        if (e.key === "Enter" || e.key === " ") {
          e.preventDefault();
          onSelect(entity.id);
        }
      }}
      className="cursor-pointer transition-opacity duration-150"
      opacity={dimmed ? 0.22 : 1}
    >
      {selected && (
        <rect
          x={x - (half + 5)}
          y={y - (half + 5)}
          width={(half + 5) * 2}
          height={(half + 5) * 2}
          fill="none"
          stroke="var(--primary)"
          strokeWidth={1}
          opacity={0.7}
        />
      )}
      <rect
        x={x - half}
        y={y - half}
        width={half * 2}
        height={half * 2}
        fill={selected ? "var(--accent)" : "var(--surface)"}
        stroke={selected ? "var(--foreground)" : highlighted ? "var(--primary)" : "var(--border-strong)"}
        strokeWidth={selected || highlighted ? 1.5 : 1}
      />
      <rect x={x - half} y={y - half} width={5} height={half * 2} fill={meta.fill} />
      <text
        x={x}
        y={y + half + 17}
        textAnchor="middle"
        fontSize="13"
        fontWeight={selected ? 700 : 400}
        fontFamily="var(--font-mono)"
        fill="var(--foreground)"
      >
        {entity.id}
      </text>
      <text
        x={x}
        y={y + half + 33}
        textAnchor="middle"
        fontSize="11"
        letterSpacing="0.08em"
        fontFamily="var(--font-mono)"
        fill="var(--muted-foreground)"
      >
        {entity.type}
      </text>
    </g>
  );
}

/** Metadata for one relationship, shown when an analyst clicks an edge. */
export function RelationshipInspector({
  relationship,
  onSelect,
  onClear,
}: {
  relationship: import("@/lib/ropus/contracts").GraphRelationship;
  onSelect: (id: string) => void;
  onClear: () => void;
}) {
  return (
    <div>
      <div className="flex items-baseline justify-between gap-3">
        <div className="label-xs">Relationship</div>
        <button onClick={onClear} className="text-[11.5px] text-primary hover:underline">
          Back to entity
        </button>
      </div>
      <div className="mt-1 text-[14px] font-semibold">{relationship.label}</div>
      <dl className="mt-4 border-t border-border">
        {([
          ["Source", relationship.source],
          ["Target", relationship.target],
          ["Path", relationship.onDecisionPath ? "decision path" : "contextual"],
        ] as Array<[string, string]>).map(([k, v]) => (
          <div key={k} className="flex items-baseline justify-between gap-4 border-b border-border py-1.5">
            <dt className="text-[12.5px] text-muted-foreground">{k}</dt>
            <dd>
              <button onClick={() => onSelect(v)} className="font-mono text-[12px] text-primary hover:underline">
                {v}
              </button>
            </dd>
          </div>
        ))}
      </dl>
    </div>
  );
}


/** Right-hand detail surface for the selected entity. */
export function GraphInspector({
  entity,
  relationships,
  onSelect,
}: {
  entity: import("@/lib/ropus/contracts").GraphEntity;
  relationships: Array<{ id: string; label: string; direction: "in" | "out" }>;
  onSelect: (id: string) => void;
}) {
  const meta = entityRisk[entity.risk];
  return (
    <div>
      <div className="label-xs">Entity</div>
      <div className="mt-1 flex items-baseline justify-between gap-3">
        <Mono className="text-[14px] font-semibold">{entity.id}</Mono>
        <span className={cn("text-[10.5px] font-semibold tracking-[0.08em] uppercase", meta.text)}>{meta.label}</span>
      </div>
      <p className="mt-0.5 text-[12.5px] text-muted-foreground">{entity.label}</p>

      <dl className="mt-4 border-t border-border">
        {[
          ["Type", entity.type] as [string, string],
          ["Hop from root", String(entity.hop)],
          ["First seen", entity.firstSeen],
          ["Last seen", entity.lastSeen],
          ...entity.attributes,
        ].map(([k, v]) => (
          <div key={k} className="flex items-baseline justify-between gap-4 border-b border-border py-1.5">
            <dt className="text-[12.5px] text-muted-foreground">{k}</dt>
            <dd className="text-right">
              <Mono className="text-[12px]">{v}</Mono>
            </dd>
          </div>
        ))}
      </dl>

      <div className="mt-5 label-xs">Relationships</div>
      <ul className="mt-1.5">
        {relationships.map((r) => (
          <li key={r.id + r.label} className="flex items-baseline justify-between gap-3 border-b border-border py-1.5">
            <span className="text-[12.5px] text-muted-foreground">
              {r.direction === "out" ? "→" : "←"} {r.label}
            </span>
            <button onClick={() => onSelect(r.id)} className="font-mono text-[12px] text-primary hover:underline">
              {r.id}
            </button>
          </li>
        ))}
      </ul>

      <div className="mt-5 label-xs">Observed signals</div>
      <ul className="mt-1.5 space-y-2">
        {entity.signals.map((s) => (
          <li key={s} className="border-l-2 border-l-border-strong pl-3 text-[12.5px] leading-relaxed">
            {s}
          </li>
        ))}
      </ul>
    </div>
  );
}

/* ------------------------------------------------------------------ cases */

const caseStatusMeta: Record<
  import("@/lib/ropus/contracts").CaseStatus,
  { label: string; tone: string }
> = {
  OPEN: { label: "Open", tone: "text-primary" },
  IN_REVIEW: { label: "In review", tone: "text-warning" },
  ESCALATED: { label: "Escalated", tone: "text-block" },
  CLOSED: { label: "Closed", tone: "text-muted-foreground" },
};

export function CaseStatusTag({ status }: { status: import("@/lib/ropus/contracts").CaseStatus }) {
  const meta = caseStatusMeta[status];
  return (
    <span className={cn("text-[12.5px] font-medium", meta.tone)}>{meta.label}</span>
  );
}

export function PriorityTag({ priority }: { priority: import("@/lib/ropus/contracts").CasePriority }) {
  const tone =
    priority === "P1" ? "text-block" : priority === "P2" ? "text-warning" : "text-muted-foreground";
  return <span className={cn("font-mono text-[12px] font-semibold tabular", tone)}>{priority}</span>;
}

export function SlaTag({ minutes }: { minutes: number }) {
  if (minutes === 0) return <span className="text-[12.5px] text-muted-foreground">—</span>;
  const breached = minutes < 0;
  const abs = Math.abs(minutes);
  const text = `${Math.floor(abs / 60)}h ${abs % 60}m${breached ? " over" : ""}`;
  return (
    <Mono className={breached ? "text-block" : abs <= 60 ? "text-warning" : "text-muted-foreground"}>
      {text}
    </Mono>
  );
}

const actorMeta: Record<import("@/lib/ropus/contracts").CaseActorKind, string> = {
  SYSTEM: "System",
  ANALYST: "Analyst",
  AI: "AI agent",
};

/**
 * Append-only case history. One vertical rule, one row per event — no cards,
 * no icon set. AI-written entries are explicitly labelled.
 */
export function CaseTimeline({ events }: { events: import("@/lib/ropus/contracts").CaseEvent[] }) {
  return (
    <ol className="border-l border-border">
      {events.map((e) => (
        <li key={e.id} className="relative py-3 pl-5">
          <span
            aria-hidden
            className={cn(
              "absolute top-[18px] -left-[3.5px] size-[6px] rounded-full",
              e.actorKind === "AI" ? "bg-warning" : e.actorKind === "ANALYST" ? "bg-primary" : "bg-border-strong",
            )}
          />
          <div className="flex flex-wrap items-baseline gap-x-3 gap-y-1">
            <Mono className="text-[11.5px] text-muted-foreground">{e.at.replace("T", " ").replace("Z", "Z")}</Mono>
            <span className="font-mono text-[10px] tracking-[0.08em] text-muted-foreground uppercase">
              {actorMeta[e.actorKind]}
            </span>
            <Mono className="text-[11.5px] text-muted-foreground">{e.actor}</Mono>
          </div>
          <p className="mt-1 text-[13px] leading-relaxed">{e.text}</p>
        </li>
      ))}
    </ol>
  );
}
