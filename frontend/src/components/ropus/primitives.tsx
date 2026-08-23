import { cn } from "@/lib/utils";
import type { ReactNode } from "react";
import { DATA_SOURCE, type Health, type Verdict } from "@/lib/ropus-data";

/* ---------------------------------------------------------------- surfaces */

export function Panel({
  title,
  description,
  actions,
  children,
  className,
  bodyClassName,
}: {
  title?: ReactNode;
  description?: ReactNode;
  actions?: ReactNode;
  children: ReactNode;
  className?: string;
  bodyClassName?: string;
}) {
  return (
    <section className={cn("border border-border bg-card", className)}>
      {(title || actions) && (
        <header className="flex items-start justify-between gap-4 border-b border-border px-4 py-2.5">
          <div>
            <h2 className="text-[13px] font-bold tracking-tight">{title}</h2>
            {description && <p className="mt-0.5 text-xs text-muted-foreground">{description}</p>}
          </div>
          {actions && <div className="flex shrink-0 items-center gap-2">{actions}</div>}
        </header>
      )}
      <div className={cn(bodyClassName)}>{children}</div>
    </section>
  );
}

export function PageHeader({
  title,
  description,
  controls,
  meta,
}: {
  title: string;
  description?: string;
  controls?: ReactNode;
  meta?: ReactNode;
}) {
  return (
    <header className="border-b border-border bg-surface px-6 py-7 lg:px-10">
      <div className="flex flex-wrap items-start justify-between gap-4">
        <div className="max-w-2xl">
          <h1 className="text-[22px] leading-tight font-bold tracking-tight">{title}</h1>
          {description && (
            <p className="mt-2 text-[13.5px] leading-relaxed text-muted-foreground">
              {description}
            </p>
          )}
          {meta}
        </div>
        {controls && <div className="flex flex-wrap items-center gap-2">{controls}</div>}
      </div>
    </header>
  );
}

export function Mono({ children, className }: { children: ReactNode; className?: string }) {
  return <span className={cn("font-mono text-[12.5px] tabular", className)}>{children}</span>;
}

export function DemoBadge({ className }: { className?: string }) {
  return (
    <span
      title={DATA_SOURCE.note}
      className={cn(
        "inline-flex items-center gap-1.5 border border-warning/40 bg-warning-surface px-1.5 py-0.5 text-[10px] font-semibold tracking-[0.08em] text-warning uppercase",
        className,
      )}
    >
      <span aria-hidden className="size-1.5 rounded-full bg-warning" />
      {DATA_SOURCE.label}
    </span>
  );
}

/* ------------------------------------------------------------------ status */

const verdictStyles: Record<Verdict, string> = {
  APPROVE: "border-approve/35 bg-approve-surface text-approve",
  REVIEW: "border-review/35 bg-review-surface text-review",
  CHALLENGE: "border-challenge/35 bg-challenge-surface text-challenge",
  BLOCK: "border-block/35 bg-block-surface text-block",
};

export function VerdictBadge({ verdict, size = "sm" }: { verdict: Verdict; size?: "sm" | "lg" }) {
  return (
    <span
      className={cn(
        "inline-flex items-center border font-semibold tracking-[0.06em] uppercase",
        verdictStyles[verdict],
        size === "sm" ? "px-1.5 py-0.5 text-[10.5px]" : "px-2.5 py-1 text-xs",
      )}
    >
      {verdict}
    </span>
  );
}

export function StatusBadge({
  tone = "neutral",
  children,
}: {
  tone?: "neutral" | "positive" | "warning" | "danger" | "info";
  children: ReactNode;
}) {
  const tones = {
    neutral: "border-border-strong bg-neutral-surface text-muted-foreground",
    positive: "border-approve/35 bg-approve-surface text-approve",
    warning: "border-review/35 bg-review-surface text-review",
    danger: "border-block/35 bg-block-surface text-block",
    info: "border-challenge/35 bg-challenge-surface text-challenge",
  } as const;
  return (
    <span
      className={cn(
        "inline-flex items-center border px-1.5 py-0.5 text-[10.5px] font-semibold tracking-[0.05em] uppercase",
        tones[tone],
      )}
    >
      {children}
    </span>
  );
}

export function HealthIndicator({ state }: { state: Health }) {
  const map = {
    HEALTHY: { tone: "bg-approve", text: "text-approve", glyph: "●" },
    DEGRADED: { tone: "bg-review", text: "text-review", glyph: "◐" },
    UNAVAILABLE: { tone: "bg-block", text: "text-block", glyph: "○" },
  } as const;
  const s = map[state];
  return (
    <span
      className={cn(
        "inline-flex items-center gap-1.5 text-[11px] font-semibold tracking-[0.05em]",
        s.text,
      )}
    >
      <span aria-hidden className={cn("size-2", s.tone)} />
      {state}
    </span>
  );
}

export function RiskScore({ value, size = "sm" }: { value: number; size?: "sm" | "xl" }) {
  const tone =
    value >= 0.8
      ? "text-block"
      : value >= 0.55
        ? "text-review"
        : value >= 0.35
          ? "text-challenge"
          : "text-approve";
  const band =
    value >= 0.8 ? "Critical" : value >= 0.55 ? "Elevated" : value >= 0.35 ? "Moderate" : "Low";
  if (size === "xl") {
    return (
      <div>
        <div className={cn("font-mono text-4xl leading-none font-bold tabular", tone)}>
          {value.toFixed(2)}
        </div>
        <div className="mt-1 text-[11px] font-semibold tracking-[0.06em] text-muted-foreground uppercase">
          {band} risk
        </div>
      </div>
    );
  }
  return (
    <span className={cn("font-mono text-[12.5px] font-semibold tabular", tone)}>
      {value.toFixed(2)}
    </span>
  );
}

/* ----------------------------------------------------------------- metrics */

export function MetricBlock({
  label,
  value,
  delta,
  dir,
  window: win,
}: {
  label: string;
  value: string;
  delta?: string;
  dir?: "up" | "down";
  window?: string;
}) {
  return (
    <div className="border-r border-b border-border px-4 py-3 last:border-r-0">
      <div className="label-xs">{label}</div>
      <div className="mt-1.5 font-mono text-xl leading-none font-bold tabular">{value}</div>
      <div className="mt-1.5 flex items-center gap-1.5 text-[11px] text-muted-foreground">
        {delta && (
          <span className="font-mono tabular">
            {dir === "up" ? "▲" : "▼"} {delta}
          </span>
        )}
        {win && <span>{win}</span>}
      </div>
    </div>
  );
}

/* ------------------------------------------------------------------ tables */

export function DataTable({
  columns,
  children,
  caption,
}: {
  columns: Array<{ key: string; label: string; align?: "left" | "right" }>;
  children: ReactNode;
  caption?: string;
}) {
  return (
    <div className="overflow-x-auto">
      <table className="w-full border-collapse text-[13px]">
        {caption && <caption className="sr-only">{caption}</caption>}
        <thead>
          <tr className="border-b border-border">
            {columns.map((c) => (
              <th
                key={c.key}
                scope="col"
                className={cn(
                  "label-xs px-3 py-2 font-semibold whitespace-nowrap first:pl-0 last:pr-0",
                  c.align === "right" ? "text-right" : "text-left",
                )}
              >
                {c.label}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>{children}</tbody>
      </table>
    </div>
  );
}

export function Row({ children, className }: { children: ReactNode; className?: string }) {
  return (
    <tr
      className={cn(
        "border-b border-border transition-colors last:border-b-0 hover:bg-accent",
        className,
      )}
    >
      {children}
    </tr>
  );
}

export function Cell({
  children,
  align,
  className,
}: {
  children: ReactNode;
  align?: "right";
  className?: string;
}) {
  return (
    <td
      className={cn(
        "px-3 py-2 align-middle first:pl-0 last:pr-0",
        align === "right" && "text-right",
        className,
      )}
    >
      {children}
    </td>
  );
}

/* -------------------------------------------------------------- key/values */

export function KeyValue({
  rows,
  columns = 1,
}: {
  rows: Array<[ReactNode, ReactNode]>;
  columns?: 1 | 2;
}) {
  return (
    <dl className={cn("grid", columns === 2 ? "sm:grid-cols-2" : "grid-cols-1")}>
      {rows.map(([k, v], i) => (
        <div
          key={i}
          className="flex items-baseline justify-between gap-4 border-b border-border px-4 py-2 last:border-b-0"
        >
          <dt className="text-[12px] text-muted-foreground">{k}</dt>
          <dd className="text-right text-[13px] font-medium">{v}</dd>
        </div>
      ))}
    </dl>
  );
}

/* --------------------------------------------------------- explainability  */

export function EvidenceGroup({
  kind,
  items,
}: {
  kind: "observed" | "inferred" | "recommended";
  items: string[];
}) {
  const meta = {
    observed: {
      title: "Observed facts",
      note: "Recorded by the platform. Not interpreted.",
      border: "border-l-approve",
      badge: null as ReactNode,
    },
    inferred: {
      title: "Inferred patterns",
      note: "Model and heuristic interpretation of the observed facts.",
      border: "border-l-challenge",
      badge: <StatusBadge tone="info">AI generated</StatusBadge>,
    },
    recommended: {
      title: "Recommended actions",
      note: "Proposals requiring analyst authorization.",
      border: "border-l-review",
      badge: <StatusBadge tone="info">AI generated</StatusBadge>,
    },
  }[kind];

  return (
    <div className={cn("border border-border border-l-2 bg-card", meta.border)}>
      <div className="flex items-center justify-between gap-3 border-b border-border px-4 py-2">
        <div>
          <h3 className="text-[13px] font-bold">{meta.title}</h3>
          <p className="text-[11px] text-muted-foreground">{meta.note}</p>
        </div>
        {meta.badge}
      </div>
      <ul className="divide-y divide-border">
        {items.map((it, i) => (
          <li key={i} className="flex gap-3 px-4 py-2 text-[13px]">
            <span className="font-mono text-[11px] text-muted-foreground tabular">
              {String(i + 1).padStart(2, "0")}
            </span>
            <span>{it}</span>
          </li>
        ))}
      </ul>
    </div>
  );
}

/* ------------------------------------------------------------------ states */

export function EmptyState({
  title,
  description,
  action,
}: {
  title: string;
  description: string;
  action?: ReactNode;
}) {
  return (
    <div className="flex flex-col items-start gap-2 border border-dashed border-border-strong bg-card px-5 py-8">
      <h3 className="text-[13px] font-bold">{title}</h3>
      <p className="max-w-md text-[13px] text-muted-foreground">{description}</p>
      {action}
    </div>
  );
}

export function ErrorState({
  title,
  detail,
  code,
}: {
  title: string;
  detail: string;
  code?: string;
}) {
  return (
    <div
      role="alert"
      className="border border-block/40 border-l-2 border-l-block bg-block-surface px-5 py-4"
    >
      <div className="flex items-center gap-2">
        <StatusBadge tone="danger">Error</StatusBadge>
        {code && <Mono className="text-block">{code}</Mono>}
      </div>
      <h3 className="mt-2 text-[13px] font-bold">{title}</h3>
      <p className="mt-0.5 text-[13px] text-muted-foreground">{detail}</p>
    </div>
  );
}

export function LoadingState({ rows = 4, label = "Loading" }: { rows?: number; label?: string }) {
  return (
    <div className="px-4 py-3" aria-busy="true" aria-live="polite">
      <span className="sr-only">{label}</span>
      <div className="space-y-2">
        {Array.from({ length: rows }).map((_, i) => (
          <div
            key={i}
            className="h-4 animate-pulse bg-neutral-surface"
            style={{ width: `${100 - i * 9}%` }}
          />
        ))}
      </div>
    </div>
  );
}

export function NoticeBar({
  tone = "info",
  title,
  children,
}: {
  tone?: "info" | "warning" | "danger";
  title: string;
  children?: ReactNode;
}) {
  const tones = {
    info: "border-l-challenge bg-challenge-surface/50",
    warning: "border-l-review bg-review-surface/50",
    danger: "border-l-block bg-block-surface/60",
  } as const;
  return (
    <div className={cn("border border-border border-l-2 px-4 py-2.5 text-[13px]", tones[tone])}>
      <span className="font-semibold">{title}</span>
      {children && <span className="text-muted-foreground"> — {children}</span>}
    </div>
  );
}

/* ------------------------------------------------------------------- code  */

export function CodeBlock({ code, language }: { code: string; language?: string }) {
  return (
    <div className="border border-border bg-navy">
      <div className="flex items-center justify-between border-b border-white/10 px-3 py-1.5">
        <span className="font-mono text-[11px] tracking-[0.06em] text-navy-foreground/60 uppercase">
          {language ?? "text"}
        </span>
        <CopyButton value={code} />
      </div>
      <pre className="overflow-x-auto px-3 py-3 font-mono text-[12.5px] leading-[1.6] text-navy-foreground">
        <code>{code}</code>
      </pre>
    </div>
  );
}

export function CopyButton({ value, label = "Copy" }: { value: string; label?: string }) {
  return (
    <button
      type="button"
      onClick={() => void navigator.clipboard?.writeText(value)}
      className="rounded-[2px] border border-white/15 px-1.5 py-0.5 font-mono text-[11px] text-navy-foreground/80 transition-colors hover:bg-white/10"
    >
      {label}
    </button>
  );
}

/* ------------------------------------------------- hierarchy without cards */

export function Section({
  title,
  description,
  actions,
  children,
  className,
}: {
  title: ReactNode;
  description?: ReactNode;
  actions?: ReactNode;
  children: ReactNode;
  className?: string;
}) {
  return (
    <section className={cn("min-w-0", className)}>
      <div className="flex flex-wrap items-baseline justify-between gap-3 border-b border-border pb-2">
        <div>
          <h2 className="text-[13px] font-bold tracking-tight">{title}</h2>
          {description && (
            <p className="mt-0.5 text-[12.5px] text-muted-foreground">{description}</p>
          )}
        </div>
        {actions && <div className="flex shrink-0 items-center gap-2">{actions}</div>}
      </div>
      <div className="pt-4">{children}</div>
    </section>
  );
}

export function SectionLabel({ children }: { children: ReactNode }) {
  return <div className="label-xs">{children}</div>;
}

/** Progressive disclosure: summary line first, forensic detail on demand. */
export function Disclosure({
  summary,
  detail,
  cta = "View detail",
  children,
  defaultOpen = false,
}: {
  summary: ReactNode;
  detail?: ReactNode;
  cta?: string;
  children: ReactNode;
  defaultOpen?: boolean;
}) {
  return (
    <details
      open={defaultOpen}
      className="group border-t border-border py-3 first:border-t-0 [&_summary::-webkit-details-marker]:hidden"
    >
      <summary className="flex cursor-pointer list-none items-baseline justify-between gap-4">
        <span>
          <span className="text-[13px] font-semibold">{summary}</span>
          {detail && <span className="ml-2 text-[12.5px] text-muted-foreground">{detail}</span>}
        </span>
        <span className="shrink-0 text-[12.5px] font-medium text-primary">
          {cta}{" "}
          <span aria-hidden className="inline-block transition-transform group-open:rotate-90">
            →
          </span>
        </span>
      </summary>
      <div className="pt-3">{children}</div>
    </details>
  );
}

/** Single consistent contribution visualization for risk factors. */
export function FactorRow({
  name,
  detail,
  weight,
  max,
  kind,
}: {
  name: string;
  detail: string;
  weight: number;
  max: number;
  kind: "observed" | "inferred";
}) {
  return (
    <div className="border-t border-border py-3 first:border-t-0">
      <div className="flex items-baseline justify-between gap-4">
        <span className="text-[13px] font-semibold">{name}</span>
        <span
          className={cn(
            "text-[10.5px] font-semibold tracking-[0.08em] uppercase",
            kind === "observed" ? "text-approve" : "text-challenge",
          )}
        >
          {kind}
        </span>
      </div>
      <p className="mt-0.5 font-mono text-[11.5px] text-muted-foreground tabular">{detail}</p>
      <div className="mt-2 flex items-center gap-3">
        <div className="h-[3px] flex-1 bg-neutral-surface">
          <div
            className={cn("h-full", kind === "observed" ? "bg-approve/70" : "bg-challenge/70")}
            style={{ width: `${(weight / max) * 100}%` }}
          />
        </div>
        <Mono className="w-12 shrink-0 text-right font-semibold">+{weight.toFixed(2)}</Mono>
      </div>
    </div>
  );
}

/** One stage of the observed → inferred → recommended explanation flow. */
export function ExplanationStage({
  step,
  title,
  kind,
  items,
  footer,
}: {
  step: string;
  title: string;
  kind: "observed" | "inferred" | "recommended";
  items: string[];
  footer?: ReactNode;
}) {
  const accent = {
    observed: "border-t-approve",
    inferred: "border-t-challenge",
    recommended: "border-t-review",
  }[kind];
  return (
    <div className={cn("border-t-2 pt-4", accent)}>
      <div className="flex items-baseline gap-2">
        <Mono className="text-[11px] text-muted-foreground">{step}</Mono>
        <h3 className="label-xs !text-foreground">{title}</h3>
      </div>
      <ul className="mt-3 space-y-2">
        {items.map((i) => (
          <li key={i} className="text-[13px] leading-relaxed">
            {i}
          </li>
        ))}
      </ul>
      {footer && <div className="mt-3 text-[12.5px]">{footer}</div>}
    </div>
  );
}
