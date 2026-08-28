import { useState, type ReactNode } from "react";
import { cn } from "@/lib/utils";
import { DATA_SOURCE } from "@/lib/ropus/fixtures";
import type { EvidenceItem, FactorSource, RiskFactorRecord, Verdict } from "@/lib/ropus/contracts";

/* ------------------------------------------------------------------ mono id */

export function Mono({
  children,
  className,
}: {
  children: ReactNode;
  className?: string | undefined;
}) {
  return <span className={cn("font-mono text-[11.5px] tabular", className)}>{children}</span>;
}

/* ---------------------------------------------------------------- demo flag */

export function DemoTag({ className }: { className?: string }) {
  return (
    <span
      title={DATA_SOURCE.note}
      className={cn(
        "inline-flex items-center gap-1.5 border border-amber-intel/40 bg-shadow-intel-surface px-1.5 py-0.5 font-mono text-[9.5px] font-semibold tracking-[0.06em] text-shadow-intel uppercase",
        className,
      )}
    >
      <span aria-hidden className="size-1 rounded-full bg-amber-intel" />
      DEMO DATA
    </span>
  );
}

/* -------------------------------------------------------------- risk score */

const scoreBand = (v: number) =>
  v >= 0.8
    ? { tone: "text-blocked", label: "Critical" }
    : v >= 0.55
      ? { tone: "text-amber-intel", label: "Elevated" }
      : v >= 0.35
        ? { tone: "text-local", label: "Moderate" }
        : { tone: "text-authoritative", label: "Low" };

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
            "font-mono text-[38px] lg:text-[42px] leading-none font-medium tracking-[-0.05em] tabular transition-colors duration-200",
            band.tone,
          )}
        >
          {value.toFixed(2)}
        </div>
        {showBand && (
          <div className="mt-1.5 font-mono text-[9.5px] font-semibold tracking-[0.12em] text-muted-foreground uppercase">
            {band.label} risk
          </div>
        )}
      </div>
    );
  }
  return (
    <span className={cn("font-mono text-[12px] font-semibold tabular", band.tone)}>
      {value.toFixed(2)}
    </span>
  );
}

/* ------------------------------------------------------------ verdict badge */

const verdictStyles: Record<Verdict, string> = {
  APPROVE: "border-authoritative/35 bg-authoritative-surface text-authoritative",
  REVIEW: "border-amber-intel/40 bg-shadow-intel-surface text-shadow-intel",
  CHALLENGE: "border-local/35 bg-local-surface text-local",
  BLOCK: "border-blocked/40 bg-blocked-surface text-blocked",
};

export function VerdictBadge({ verdict, size = "sm" }: { verdict: Verdict; size?: "sm" | "lg" }) {
  return (
    <span
      className={cn(
        "inline-flex items-center gap-1.5 border font-mono font-semibold tracking-[0.06em] uppercase",
        verdictStyles[verdict],
        size === "sm" ? "px-2 py-0.5 text-[9.5px]" : "px-2.5 py-1 text-[11px]",
      )}
    >
      <span
        aria-hidden
        className={cn(
          "size-1 rounded-full",
          verdict === "APPROVE"
            ? "bg-authoritative"
            : verdict === "REVIEW"
              ? "bg-amber-intel"
              : verdict === "CHALLENGE"
                ? "bg-local"
                : "bg-blocked",
        )}
      />
      {verdict}
    </span>
  );
}

/* --------------------------------------------------------- semantic status pills */

export type SemanticTone = "authoritative" | "shadow" | "blocked" | "local" | "synthetic" | "neutral";

export function StatusPill({
  tone = "neutral",
  children,
  className,
}: {
  tone?: SemanticTone;
  children: ReactNode;
  className?: string;
}) {
  const tones: Record<SemanticTone, { box: string; dot: string }> = {
    authoritative: {
      box: "border-authoritative/35 bg-authoritative-surface text-authoritative",
      dot: "bg-authoritative",
    },
    shadow: {
      box: "border-amber-intel/40 bg-shadow-intel-surface text-shadow-intel",
      dot: "bg-amber-intel",
    },
    blocked: {
      box: "border-blocked/40 bg-blocked-surface text-blocked",
      dot: "bg-blocked",
    },
    local: {
      box: "border-local/35 bg-local-surface text-local",
      dot: "bg-local",
    },
    synthetic: {
      box: "border-synthetic/35 bg-synthetic-surface text-synthetic",
      dot: "bg-synthetic",
    },
    neutral: {
      box: "border-border-strong bg-muted/60 text-muted-foreground",
      dot: "bg-muted-foreground",
    },
  };

  return (
    <span
      className={cn(
        "inline-flex items-center gap-1.5 border px-2 py-0.5 font-mono text-[9.5px] font-semibold tracking-[0.06em] uppercase",
        tones[tone].box,
        className,
      )}
    >
      <span aria-hidden className={cn("size-1 rounded-full", tones[tone].dot)} />
      {children}
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
    OK: "bg-authoritative text-authoritative",
    DEGRADED: "bg-amber-intel text-shadow-intel",
    DOWN: "bg-blocked text-blocked",
  } as const;
  return (
    <span
      className={cn(
        "inline-flex items-center gap-1.5 font-mono text-[9.5px] font-semibold tracking-[0.06em]",
        map[state].split(" ")[1],
      )}
    >
      <span aria-hidden className={cn("size-1.5 rounded-full", map[state].split(" ")[0])} />
      {label ?? state}
    </span>
  );
}

/* ------------------------------------------------------------- source tag */

const sourceLabels: Record<FactorSource, string> = {
  RULES: "Rules",
  ML: "ML (BMR)",
  THREAT_INTEL: "Threat Intel",
  GRAPH: "GraphSAGE Shadow",
  DEVICE: "Device",
};

export function SourceTag({ source }: { source: FactorSource }) {
  return (
    <span className="shrink-0 font-mono text-[9.5px] font-medium tracking-[0.08em] text-muted-foreground uppercase">
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
        <span className="font-sans text-[12.5px] font-semibold">{factor.label}</span>
        <div className="flex shrink-0 items-baseline gap-3">
          <SourceTag source={factor.source} />
          <Mono className="w-12 text-right font-bold text-[11.5px]">+{factor.weight.toFixed(2)}</Mono>
        </div>
      </div>
      {factor.detail && (
        <p className="mt-0.5 font-mono text-[10.5px] text-muted-foreground tabular">
          {factor.detail}
        </p>
      )}
      <div className="mt-2 h-[2px] w-full bg-secondary">
        <div
          className="h-full bg-primary/70 transition-[width] duration-200"
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
      <span className="font-mono text-[9.5px] font-semibold tracking-[0.06em] text-authoritative uppercase">
        Authorized by Analyst
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
            "px-2 py-1 font-mono text-[11px] font-medium text-white transition-colors duration-150 cursor-pointer",
            item.action.destructive ? "bg-destructive hover:bg-destructive/90" : "bg-primary hover:bg-primary/90",
          )}
        >
          Confirm
        </button>
        <button
          type="button"
          onClick={() => setState("idle")}
          className="px-2 py-1 font-sans text-[11.5px] font-medium text-muted-foreground hover:text-foreground cursor-pointer"
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
      className="border border-border-strong bg-surface px-2 py-1 font-sans text-[11.5px] font-medium transition-colors duration-150 hover:bg-secondary cursor-pointer"
    >
      {item.action.label}
    </button>
  );
}

const kindMeta = {
  OBSERVED: {
    title: "Observed Facts",
    note: "Recorded deterministically by the platform runtime.",
    rule: "border-t-authoritative",
  },
  INFERRED: {
    title: "Inferred Patterns",
    note: "Model, GraphSAGE shadow, and ML behavioral intelligence.",
    rule: "border-t-local",
  },
  RECOMMENDED: {
    title: "Recommended Actions",
    note: "Proposed to human analyst. Requires explicit authorization.",
    rule: "border-t-amber-intel",
  },
} as const;

/**
 * Three-way explainability split. Deliberately not three identical cards:
 * observed is a plain fact list, inferred is labelled model output, recommended
 * is a list of actions that require explicit confirmation.
 */
export function EvidenceList({
  items,
  kind,
}: {
  items: EvidenceItem[];
  kind: keyof typeof kindMeta;
}) {
  const meta = kindMeta[kind];
  const rows = items.filter((i) => i.kind === kind);

  return (
    <section className={cn("border-t-2 pt-3", meta.rule)}>
      <div className="flex items-baseline justify-between gap-3">
        <h3 className="font-mono text-[9.5px] font-bold tracking-[0.12em] uppercase text-foreground">{meta.title}</h3>
        {kind !== "OBSERVED" && (
          <span className="font-mono text-[9px] font-semibold tracking-[0.10em] text-muted-foreground uppercase">
            AI GENERATED
          </span>
        )}
      </div>
      <p className="mt-1 font-sans text-[11.5px] text-muted-foreground">{meta.note}</p>

      <ul className="mt-3">
        {rows.map((item) => (
          <li key={item.id} className="border-t border-border py-2.5 first:border-t-0 first:pt-0">
            {kind === "OBSERVED" ? (
              <div className="flex items-baseline justify-between gap-3">
                <span className="font-sans text-[12.5px] leading-relaxed text-foreground/90">{item.text}</span>
                {item.source && <SourceTag source={item.source} />}
              </div>
            ) : kind === "INFERRED" ? (
              <p className="font-sans text-[12.5px] leading-relaxed text-foreground/90">{item.text}</p>
            ) : (
              <div className="flex flex-wrap items-center justify-between gap-3">
                <span className="font-sans text-[12.5px] leading-relaxed text-foreground/90">{item.text}</span>
                <RecommendedAction item={item} />
              </div>
            )}
          </li>
        ))}
        {rows.length === 0 && (
          <li className="py-2.5 font-sans text-[11.5px] text-muted-foreground">
            None recorded for this decision.
          </li>
        )}
      </ul>
    </section>
  );
}

/* ------------------------------------------------------------------ metric */

export function Metric({ label, value, sub }: { label: string; value: string; sub?: string }) {
  return (
    <div>
      <div className="font-mono text-[28px] leading-none font-medium tracking-[-0.05em] tabular">{value}</div>
      <div className="mt-2 font-mono text-[9.5px] font-semibold tracking-[0.12em] text-muted-foreground uppercase">{label}</div>
      {sub && (
        <div className="mt-0.5 font-mono text-[10.5px] text-muted-foreground tabular">{sub}</div>
      )}
    </div>
  );
}

/* -------------------------------------------------------- decision summary */

/**
 * Compact technical record of the decision. Deliberately a definition table,
 * not a grid of KPI cards.
 */
export function DecisionSummary({
  decision,
  amount,
}: {
  decision: import("@/lib/ropus/contracts").RiskDecision;
  amount: string;
}) {
  const rows: Array<[string, ReactNode]> = [
    ["Verdict", <VerdictBadge key="v" verdict={decision.verdict} />],
    ["Risk Score", <Mono key="s" className="font-semibold text-foreground">{decision.riskScore.toFixed(2)}</Mono>],
    ["Confidence", <Mono key="c">{(decision.confidence * 100).toFixed(1)}%</Mono>],
    ["Model", <Mono key="m">{`${decision.model}-${decision.modelVersion}`}</Mono>],
    ["Policy", <Mono key="p">{decision.policy}</Mono>],
    ["Latency", <Mono key="l">{decision.latencyMs.toFixed(1)} ms</Mono>],
    ["Amount", <Mono key="a" className="font-semibold">{`${amount} ${decision.currency}`}</Mono>],
    [
      "Rules",
      decision.rules.length ? (
        <div key="r" className="flex flex-col items-end gap-0.5">
          {decision.rules.map((r) => (
            <Mono key={r.id}>{r.id}</Mono>
          ))}
        </div>
      ) : (
        <span key="r0" className="font-sans text-[11.5px] text-muted-foreground">
          none triggered
        </span>
      ),
    ],
  ];

  return (
    <dl className="divide-y divide-border border-y border-border">
      {rows.map(([k, v]) => (
        <div key={k} className="flex items-baseline justify-between gap-6 py-1.5">
          <dt className="font-mono text-[9.5px] font-semibold tracking-[0.10em] text-muted-foreground uppercase">{k}</dt>
          <dd className="text-right font-sans text-[12px]">{v}</dd>
        </div>
      ))}
    </dl>
  );
}

/* -------------------------------------------------------------- graph bits */

const entityRisk: Record<
  import("@/lib/ropus/contracts").EntityRisk,
  { fill: string; text: string; bg: string; border: string; label: string }
> = {
  CLEAN: { fill: "#647182", text: "text-muted-foreground", bg: "bg-muted-foreground", border: "border-muted-foreground", label: "Clean" },
  WATCH: { fill: "#476C87", text: "text-local", bg: "bg-local", border: "border-local", label: "Watch" },
  SUSPECT: { fill: "#F5A524", text: "text-shadow-intel", bg: "bg-amber-intel", border: "border-amber-intel", label: "Suspect" },
  CONFIRMED_FRAUD: { fill: "#CE3227", text: "text-blocked", bg: "bg-destructive", border: "border-destructive", label: "Confirmed Fraud" },
};

export function entityRiskMeta(risk: import("@/lib/ropus/contracts").EntityRisk) {
  return entityRisk[risk];
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
        <div className="font-mono text-[9.5px] font-bold tracking-[0.12em] text-muted-foreground uppercase">
          RELATIONSHIP
        </div>
        <button onClick={onClear} className="font-sans text-[11px] text-primary hover:underline cursor-pointer">
          Back to Entity
        </button>
      </div>
      <div className="mt-1 font-sans text-[13.5px] font-bold text-foreground">{relationship.label}</div>
      <dl className="mt-4 border-t border-border">
        {(
          [
            ["Source", relationship.source],
            ["Target", relationship.target],
            ["Path", relationship.onDecisionPath ? "decision path" : "contextual"],
          ] as Array<[string, string]>
        ).map(([k, v]) => (
          <div
            key={k}
            className="flex items-baseline justify-between gap-4 border-b border-border py-1.5"
          >
            <dt className="font-mono text-[9.5px] font-semibold tracking-[0.08em] text-muted-foreground uppercase">{k}</dt>
            <dd>
              <button
                onClick={() => onSelect(v)}
                className="font-mono text-[11.5px] text-primary hover:underline cursor-pointer"
              >
                {v}
              </button>
            </dd>
          </div>
        ))}
      </dl>
    </div>
  );
}

/* ------------------------------------------------------------------ cases */

const caseStatusMeta: Record<
  import("@/lib/ropus/contracts").CaseStatus,
  { label: string; tone: string; pillTone: SemanticTone }
> = {
  OPEN: { label: "Open", tone: "text-local", pillTone: "local" },
  IN_REVIEW: { label: "In Review", tone: "text-shadow-intel", pillTone: "shadow" },
  ESCALATED: { label: "Escalated", tone: "text-blocked", pillTone: "blocked" },
  CLOSED: { label: "Closed", tone: "text-muted-foreground", pillTone: "neutral" },
};

export function CaseStatusTag({ status }: { status: import("@/lib/ropus/contracts").CaseStatus }) {
  const meta = caseStatusMeta[status];
  return <StatusPill tone={meta.pillTone}>{meta.label}</StatusPill>;
}

export function PriorityTag({
  priority,
}: {
  priority: import("@/lib/ropus/contracts").CasePriority;
}) {
  const tone =
    priority === "P1" ? "text-blocked font-bold" : priority === "P2" ? "text-amber-intel font-bold" : "text-muted-foreground";
  return (
    <span className={cn("font-mono text-[11px] font-bold tabular", tone)}>{priority}</span>
  );
}

export function SlaTag({ minutes }: { minutes: number }) {
  if (minutes === 0) return <span className="font-sans text-[11.5px] text-muted-foreground">—</span>;
  const breached = minutes < 0;
  const abs = Math.abs(minutes);
  const text = `${Math.floor(abs / 60)}h ${abs % 60}m${breached ? " over" : ""}`;
  return (
    <Mono
      className={breached ? "text-blocked font-bold" : abs <= 60 ? "text-amber-intel font-bold" : "text-muted-foreground"}
    >
      {text}
    </Mono>
  );
}

const actorMeta: Record<import("@/lib/ropus/contracts").CaseActorKind, string> = {
  SYSTEM: "System",
  ANALYST: "Analyst",
  AI: "AI Agent",
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
              e.actorKind === "AI"
                ? "bg-amber-intel"
                : e.actorKind === "ANALYST"
                  ? "bg-navy"
                  : "bg-border-strong",
            )}
          />
          <div className="flex flex-wrap items-baseline gap-x-3 gap-y-1">
            <Mono className="text-[10px] text-muted-foreground font-medium">
              {e.at.replace("T", " ").replace("Z", "Z")}
            </Mono>
            <span className="font-mono text-[9px] font-bold tracking-[0.10em] text-muted-foreground uppercase">
              {actorMeta[e.actorKind]}
            </span>
            <Mono className="text-[10px] text-muted-foreground">{e.actor}</Mono>
          </div>
          <p className="mt-1 font-sans text-[12px] leading-relaxed text-foreground/90">{e.text}</p>
        </li>
      ))}
    </ol>
  );
}

