import type { ReactNode } from "react";
import { cn } from "@/lib/utils";
import { DemoTag } from "@/components/ropus/core";
import { isLiveBackend } from "@/lib/ropus/api";

/**
 * Every ROPUS surface uses the same container, the same header rhythm and the
 * same table. One spacing system, no per-page improvisation.
 */
export function Page({ children, className }: { children: ReactNode; className?: string }) {
  return <div className={cn("mx-auto w-full max-w-[1560px] px-5 py-5 lg:px-7", className)}>{children}</div>;
}

export function PageHead({
  title,
  subtitle,
  breadcrumb,
  actions,
}: {
  title: ReactNode;
  subtitle?: ReactNode;
  breadcrumb?: ReactNode;
  actions?: ReactNode;
}) {
  return (
    <header className="flex flex-wrap items-start justify-between gap-x-8 gap-y-3 border-b border-border pb-4">
      <div className="min-w-0">
        {breadcrumb && <div className="mb-1.5 flex items-center gap-2 text-[11.5px] text-muted-foreground">{breadcrumb}</div>}
        <h1 className="text-[20px] leading-tight font-bold tracking-[-0.01em]">{title}</h1>
        {subtitle && <p className="mt-1 max-w-[78ch] text-[12.5px] text-muted-foreground">{subtitle}</p>}
      </div>
      <div className="flex shrink-0 items-center gap-3">
        {actions}
        {!isLiveBackend && <DemoTag />}
      </div>
    </header>
  );
}

export function SectionHead({
  title,
  meta,
  right,
  className,
}: {
  title: string;
  meta?: ReactNode;
  right?: ReactNode;
  className?: string;
}) {
  return (
    <div className={cn("flex flex-wrap items-baseline justify-between gap-3 border-b border-border pb-1.5", className)}>
      <div className="flex items-baseline gap-3">
        <h2 className="text-[11px] font-bold tracking-[0.08em] uppercase">{title}</h2>
        {meta && <span className="text-[11.5px] text-muted-foreground">{meta}</span>}
      </div>
      {right}
    </div>
  );
}

export type Column = { key: string; label: string; align?: "left" | "right"; width?: string };

/** Dense operational table. Rows are arrays of cells, positionally matched to columns. */
export function DataGrid({
  columns,
  rows,
  empty = "No records.",
  onSelect,
  selectedId,
  rowClassName = "py-[7px]",
}: {
  columns: Column[];
  rows: Array<{ id: string; cells: ReactNode[] }>;
  empty?: string;
  onSelect?: (id: string) => void;
  selectedId?: string | null;
  /** Vertical padding utility for body cells; lets dense surfaces tune row height. */
  rowClassName?: string;
}) {
  return (
    <table className="w-full border-collapse text-[12.5px]">
      <thead>
        <tr className="border-b border-border">
          {columns.map((c) => (
            <th
              key={c.key}
              scope="col"
              style={c.width ? { width: c.width } : undefined}
              className={cn(
                "py-1.5 pr-4 text-[10.5px] font-semibold tracking-[0.08em] whitespace-nowrap text-muted-foreground uppercase last:pr-0",
                c.align === "right" ? "text-right" : "text-left",
              )}
            >
              {c.label}
            </th>
          ))}
        </tr>
      </thead>
      <tbody>
        {rows.map((r) => (
          <tr
            key={r.id}
            onClick={onSelect ? () => onSelect(r.id) : undefined}
            tabIndex={onSelect ? 0 : undefined}
            onKeyDown={
              onSelect
                ? (e) => {
                    if (e.key === "Enter" || e.key === " ") {
                      e.preventDefault();
                      onSelect(r.id);
                    }
                  }
                : undefined
            }
            aria-selected={onSelect ? selectedId === r.id : undefined}
            className={cn(
              "border-b border-border last:border-b-0 hover:bg-accent",
              onSelect && "cursor-pointer outline-none focus-visible:bg-accent",
              selectedId === r.id && "bg-accent",
            )}
          >
            {r.cells.map((cell, i) => (
              <td
                key={columns[i]?.key ?? i}
                className={cn(
                  "pr-4 align-top last:pr-0",
                  rowClassName,
                  columns[i]?.align === "right" ? "text-right" : "text-left",
                  selectedId === r.id && i === 0 && "relative",
                )}
              >
                {selectedId === r.id && i === 0 && (
                  <span aria-hidden className="absolute top-0 bottom-0 -left-2 w-[2px] bg-primary" />
                )}
                {cell}
              </td>
            ))}
          </tr>
        ))}
        {rows.length === 0 && (
          <tr>
            <td colSpan={columns.length} className="py-4 text-[12.5px] text-muted-foreground">
              {empty}
            </td>
          </tr>
        )}
      </tbody>
    </table>
  );
}

/** Horizontal metric strip — separators, never cards. */
export function MetricStrip({
  items,
}: {
  items: Array<{ label: string; value: string; sub?: string; tone?: string }>;
}) {
  return (
    <dl className="grid grid-cols-2 divide-x divide-border border-b border-border sm:grid-cols-3 lg:grid-cols-5">
      {items.map((m, i) => (
        <div key={m.label} className={cn("px-5 py-3.5", i === 0 && "pl-0")}>
          <dd className={cn("font-mono text-[21px] leading-none font-bold tabular", m.tone)}>{m.value}</dd>
          <dt className="mt-2 text-[11.5px] font-medium">{m.label}</dt>
          {m.sub && <dd className="mt-0.5 text-[11px] text-muted-foreground">{m.sub}</dd>}
        </div>
      ))}
    </dl>
  );
}

/**
 * Compact telemetry strip. Denser than MetricStrip: value and label share a
 * baseline block, separators are vertical rules, nothing is a card.
 */
export function TelemetryStrip({
  items,
  className,
}: {
  items: Array<{ label: string; value: string; sub?: string; tone?: string }>;
  className?: string;
}) {
  return (
    <dl className={cn("flex flex-wrap divide-x divide-border border-b border-border", className)}>
      {items.map((m, i) => (
        <div key={m.label} className={cn("px-4 py-2.5", i === 0 && "pl-0")}>
          <dd className={cn("font-mono text-[17px] leading-none font-bold tabular", m.tone)}>{m.value}</dd>
          <dt className="mt-1.5 text-[10.5px] font-semibold tracking-[0.07em] text-muted-foreground uppercase">
            {m.label}
          </dt>
          {m.sub && <dd className="mt-0.5 text-[10.5px] text-muted-foreground">{m.sub}</dd>}
        </div>
      ))}
    </dl>
  );
}

/** Label/value row used inside inspectors. Aligned, quiet label, mono-friendly value. */
export function InspectorRow({ label, children }: { label: string; children: ReactNode }) {
  return (
    <div className="flex items-baseline justify-between gap-4 border-b border-border py-[5px] last:border-b-0">
      <span className="text-[10.5px] font-semibold tracking-[0.07em] whitespace-nowrap text-muted-foreground uppercase">
        {label}
      </span>
      <span className="min-w-0 truncate text-right text-[12.5px]">{children}</span>
    </div>
  );
}
