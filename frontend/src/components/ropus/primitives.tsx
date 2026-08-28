import { cn } from "@/lib/utils";
import type { ReactNode } from "react";
import type { Verdict, ServiceState as Health } from "@/lib/ropus/contracts";
import { isLiveBackend } from "@/lib/ropus/api";

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
    <section className={cn("border border-border bg-card shadow-xs", className)}>
      {(title || actions) && (
        <header className="flex items-start justify-between gap-4 border-b border-border bg-card px-4 py-2.5">
          <div>
            <h2 className="font-sans text-[13px] font-bold tracking-tight text-foreground">{title}</h2>
            {description && <p className="mt-0.5 font-sans text-[11.5px] text-muted-foreground">{description}</p>}
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
    <header className="border-b border-border bg-surface px-6 py-6 lg:px-8">
      <div className="flex flex-wrap items-start justify-between gap-4">
        <div className="max-w-3xl">
          <h1 className="font-sans text-[24px] lg:text-[26px] font-extrabold leading-[1.12] tracking-[-0.04em] text-foreground">{title}</h1>
          {description && (
            <p className="mt-1.5 font-sans text-[12.5px] leading-relaxed text-muted-foreground">
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
  return <span className={cn("font-mono text-[11.5px] tabular", className)}>{children}</span>;
}

export function DemoBadge({ className }: { className?: string }) {
  return isLiveBackend ? (
    <span
      title="Connected to authoritative Go API at localhost:8080"
      className={cn(
        "inline-flex items-center gap-1.5 border border-authoritative/35 bg-authoritative-surface px-1.5 py-0.5 font-mono text-[9.5px] font-semibold tracking-[0.06em] text-authoritative uppercase",
        className,
      )}
    >
      <span aria-hidden className="size-1 rounded-full bg-authoritative" />
      LIVE BACKEND
    </span>
  ) : (
    <span
      title="Offline fixtures rendered locally"
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

/* ------------------------------------------------------------------ status */

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

export function StatusBadge({
  tone = "neutral",
  children,
}: {
  tone?: "neutral" | "positive" | "warning" | "danger" | "info" | "synthetic";
  children: ReactNode;
}) {
  const tones = {
    neutral: "border-border-strong bg-muted/60 text-muted-foreground",
    positive: "border-authoritative/35 bg-authoritative-surface text-authoritative",
    warning: "border-amber-intel/40 bg-shadow-intel-surface text-shadow-intel",
    danger: "border-blocked/40 bg-blocked-surface text-blocked",
    info: "border-local/35 bg-local-surface text-local",
    synthetic: "border-synthetic/35 bg-synthetic-surface text-synthetic",
  } as const;
  return (
    <span
      className={cn(
        "inline-flex items-center gap-1.5 border px-2 py-0.5 font-mono text-[9.5px] font-semibold tracking-[0.06em] uppercase",
        tones[tone],
      )}
    >
      <span
        aria-hidden
        className={cn(
          "size-1 rounded-full",
          tone === "positive"
            ? "bg-authoritative"
            : tone === "warning"
              ? "bg-amber-intel"
              : tone === "danger"
                ? "bg-blocked"
                : tone === "info"
                  ? "bg-local"
                  : tone === "synthetic"
                    ? "bg-synthetic"
                    : "bg-muted-foreground",
        )}
      />
      {children}
    </span>
  );
}

export function HealthIndicator({ state }: { state: Health }) {
  const map = {
    HEALTHY: { tone: "bg-authoritative", text: "text-authoritative", glyph: "●" },
    DEGRADED: { tone: "bg-amber-intel", text: "text-shadow-intel", glyph: "◐" },
    UNAVAILABLE: { tone: "bg-blocked", text: "text-blocked", glyph: "○" },
  } as const;
  const s = map[state];
  return (
    <span
      className={cn(
        "inline-flex items-center gap-1.5 font-mono text-[9.5px] font-semibold tracking-[0.06em]",
        s.text,
      )}
    >
      <span aria-hidden className={cn("size-1.5 rounded-full", s.tone)} />
      {state}
    </span>
  );
}

export function RiskScore({ value, size = "sm" }: { value: number; size?: "sm" | "xl" }) {
  const tone =
    value >= 0.8
      ? "text-blocked"
      : value >= 0.55
        ? "text-amber-intel"
        : value >= 0.35
          ? "text-local"
          : "text-authoritative";
  const band =
    value >= 0.8 ? "Critical" : value >= 0.55 ? "Elevated" : value >= 0.35 ? "Moderate" : "Low";
  if (size === "xl") {
    return (
      <div>
        <div className={cn("font-mono text-[36px] lg:text-[40px] leading-none font-medium tracking-[-0.05em] tabular", tone)}>
          {value.toFixed(2)}
        </div>
        <div className="mt-1 font-mono text-[9.5px] font-semibold tracking-[0.12em] text-muted-foreground uppercase">
          {band} risk
        </div>
      </div>
    );
  }
  return (
    <span className={cn("font-mono text-[12px] font-semibold tabular", tone)}>
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
    <div className="border-r border-b border-border bg-card px-4 py-3.5 last:border-r-0">
      <div className="font-mono text-[9.5px] font-semibold tracking-[0.12em] text-muted-foreground uppercase">{label}</div>
      <div className="mt-1.5 font-mono text-[28px] leading-none font-medium tracking-[-0.05em] tabular">{value}</div>
      <div className="mt-1.5 flex items-center gap-1.5 font-mono text-[10.5px] text-muted-foreground">
        {delta && (
          <span className={cn("tabular font-semibold", dir === "up" ? "text-blocked" : "text-authoritative")}>
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
      <table className="w-full border-collapse font-sans text-[12px]">
        {caption && <caption className="sr-only">{caption}</caption>}
        <thead>
          <tr className="border-b border-border bg-secondary/35">
            {columns.map((c) => (
              <th
                key={c.key}
                scope="col"
                className={cn(
                  "px-3 py-2 font-mono text-[9px] font-bold tracking-[0.10em] whitespace-nowrap text-muted-foreground uppercase first:pl-3 last:pr-3",
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
        "border-b border-border transition-colors last:border-b-0 hover:bg-secondary/60",
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
        "px-3 py-2 align-middle font-sans text-[12px] first:pl-3 last:pr-3",
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
          <dt className="font-mono text-[9.5px] font-semibold tracking-[0.10em] text-muted-foreground uppercase">{k}</dt>
          <dd className="text-right font-sans text-[12px] font-medium text-foreground">{v}</dd>
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
      title: "Observed Facts",
      note: "Recorded deterministically by the platform. Not interpreted.",
      border: "border-l-authoritative",
      badge: null as ReactNode,
    },
    inferred: {
      title: "Inferred Patterns",
      note: "Model, GraphSAGE shadow, and heuristic interpretation.",
      border: "border-l-local",
      badge: <StatusBadge tone="info">AI GENERATED</StatusBadge>,
    },
    recommended: {
      title: "Recommended Actions",
      note: "Proposals requiring explicit analyst authorization.",
      border: "border-l-warning",
      badge: <StatusBadge tone="warning">PROPOSAL</StatusBadge>,
    },
  }[kind];

  return (
    <div className={cn("border border-border border-l-2 bg-card shadow-xs", meta.border)}>
      <div className="flex items-center justify-between gap-3 border-b border-border px-4 py-2.5">
        <div>
          <h3 className="font-sans text-[13px] font-bold text-foreground">{meta.title}</h3>
          <p className="font-sans text-[11.5px] text-muted-foreground">{meta.note}</p>
        </div>
        {meta.badge}
      </div>
      <ul className="divide-y divide-border">
        {items.map((it, i) => (
          <li key={i} className="flex gap-3 px-4 py-2 font-sans text-[12px]">
            <span className="font-mono text-[10.5px] text-muted-foreground tabular font-medium">
              {String(i + 1).padStart(2, "0")}
            </span>
            <span className="text-foreground/90">{it}</span>
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
      <h3 className="font-sans text-[13px] font-bold text-foreground">{title}</h3>
      <p className="max-w-md font-sans text-[12px] text-muted-foreground">{description}</p>
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
      className="border border-blocked/40 border-l-2 border-l-blocked bg-blocked-surface px-5 py-4"
    >
      <div className="flex items-center gap-2">
        <StatusBadge tone="danger">Error</StatusBadge>
        {code && <Mono className="text-blocked font-bold">{code}</Mono>}
      </div>
      <h3 className="mt-2 font-sans text-[13px] font-bold text-foreground">{title}</h3>
      <p className="mt-0.5 font-sans text-[12px] text-muted-foreground">{detail}</p>
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
            className="h-4 animate-pulse bg-secondary"
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
  tone?: "info" | "warning" | "danger" | "positive";
  title: string;
  children?: ReactNode;
}) {
  const tones = {
    info: "border-l-local bg-local-surface text-local",
    warning: "border-l-amber-intel bg-shadow-intel-surface text-shadow-intel",
    danger: "border-l-blocked bg-blocked-surface text-blocked",
    positive: "border-l-authoritative bg-authoritative-surface text-authoritative",
  } as const;
  return (
    <div className={cn("border border-border border-l-2 px-4 py-2.5 font-sans text-[12px]", tones[tone])}>
      <span className="font-semibold">{title}</span>
      {children && <span className="text-muted-foreground"> — {children}</span>}
    </div>
  );
}

/* ------------------------------------------------------------------- code  */

export function CodeBlock({ code, language }: { code: string; language?: string }) {
  return (
    <div className="border border-sidebar-border bg-sidebar text-sidebar-foreground">
      <div className="flex items-center justify-between border-b border-sidebar-border px-3 py-1.5">
        <span className="font-mono text-[9.5px] font-semibold tracking-[0.08em] text-sidebar-muted uppercase">
          {language ?? "text"}
        </span>
        <CopyButton value={code} />
      </div>
      <pre className="overflow-x-auto px-3 py-3 font-mono text-[11.5px] leading-[1.6] text-sidebar-foreground">
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
      className="border border-sidebar-border bg-sidebar-accent px-1.5 py-0.5 font-mono text-[9.5px] font-medium text-sidebar-muted transition-colors hover:text-white cursor-pointer"
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
          <h2 className="font-sans text-[13.5px] font-bold tracking-tight text-foreground">{title}</h2>
          {description && (
            <p className="mt-0.5 font-sans text-[11.5px] text-muted-foreground">{description}</p>
          )}
        </div>
        {actions && <div className="flex shrink-0 items-center gap-2">{actions}</div>}
      </div>
      <div className="pt-4">{children}</div>
    </section>
  );
}

export function SectionLabel({ children }: { children: ReactNode }) {
  return <div className="font-mono text-[9.5px] font-semibold tracking-[0.12em] text-muted-foreground uppercase">{children}</div>;
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
          <span className="font-sans text-[12.5px] font-semibold text-foreground">{summary}</span>
          {detail && <span className="ml-2 font-sans text-[11.5px] text-muted-foreground">{detail}</span>}
        </span>
        <span className="shrink-0 font-sans text-[11.5px] font-medium text-primary">
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
    <div className="border-t border-border py-2.5 first:border-t-0">
      <div className="flex items-baseline justify-between gap-4">
        <span className="font-sans text-[12.5px] font-semibold text-foreground">{name}</span>
        <span
          className={cn(
            "font-mono text-[9.5px] font-semibold tracking-[0.08em] uppercase",
            kind === "observed" ? "text-authoritative" : "text-local",
          )}
        >
          {kind}
        </span>
      </div>
      <p className="mt-0.5 font-mono text-[10.5px] text-muted-foreground tabular">{detail}</p>
      <div className="mt-2 flex items-center gap-3">
        <div className="h-[2px] flex-1 bg-secondary">
          <div
            className={cn("h-full", kind === "observed" ? "bg-authoritative" : "bg-local")}
            style={{ width: `${(weight / max) * 100}%` }}
          />
        </div>
        <Mono className="w-12 shrink-0 text-right font-bold">+{weight.toFixed(2)}</Mono>
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
    observed: "border-t-authoritative",
    inferred: "border-t-local",
    recommended: "border-t-amber-intel",
  }[kind];
  return (
    <div className={cn("border-t-2 pt-3.5", accent)}>
      <div className="flex items-baseline gap-2">
        <Mono className="text-[10px] text-muted-foreground font-medium">{step}</Mono>
        <h3 className="font-mono text-[9.5px] font-bold tracking-[0.12em] uppercase text-foreground">{title}</h3>
      </div>
      <ul className="mt-2.5 space-y-1.5">
        {items.map((i) => (
          <li key={i} className="font-sans text-[12px] leading-relaxed text-foreground/90">
            {i}
          </li>
        ))}
      </ul>
      {footer && <div className="mt-2.5 font-sans text-[11.5px]">{footer}</div>}
    </div>
  );
}
