import type { ReactNode } from "react";
import { cn } from "@/lib/utils";
import { DemoTag } from "@/components/ropus/core";
import { isLiveBackend } from "@/lib/ropus/api";

/**
 * Every ROPUS surface uses the same container, the same header rhythm and the
 * same table. One spacing system, no per-page improvisation.
 */
export function Page({ children, className }: { children: ReactNode; className?: string }) {
  return (
    <div className={cn("mx-auto w-full max-w-[1560px] px-5 py-5 lg:px-8", className)}>
      {children}
    </div>
  );
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
        {breadcrumb && (
          <div className="mb-1.5 flex items-center gap-2 font-mono text-[9.5px] font-semibold tracking-[0.13em] text-muted-foreground uppercase">
            {breadcrumb}
          </div>
        )}
        <h1 className="font-sans text-[24px] lg:text-[26px] font-extrabold leading-[1.12] tracking-[-0.04em] text-foreground">
          {title}
        </h1>
        {subtitle && (
          <p className="mt-1 max-w-[80ch] font-sans text-[12.5px] font-normal leading-relaxed text-muted-foreground">
            {subtitle}
          </p>
        )}
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
    <div
      className={cn(
        "flex flex-wrap items-baseline justify-between gap-3 border-b border-border pb-1.5",
        className,
      )}
    >
      <div className="flex items-baseline gap-3">
        <h2 className="font-mono text-[9.5px] font-bold tracking-[0.12em] text-muted-foreground uppercase">
          {title}
        </h2>
        {meta && <span className="font-mono text-[9.5px] text-muted-foreground">{meta}</span>}
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
  rowClassName = "py-2",
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
    <div className="w-full overflow-x-auto">
      <table className="w-full border-collapse font-sans text-[12px]">
        <thead>
          <tr className="border-b border-border bg-secondary/35">
            {columns.map((c) => (
              <th
                key={c.key}
                scope="col"
                style={c.width ? { width: c.width } : undefined}
                className={cn(
                  "py-2 pr-4 font-mono text-[9px] font-bold tracking-[0.10em] whitespace-nowrap text-muted-foreground uppercase last:pr-0",
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
                "border-b border-border last:border-b-0 hover:bg-secondary/60 transition-colors",
                onSelect && "cursor-pointer outline-none focus-visible:bg-secondary/80",
                selectedId === r.id && "bg-secondary/60",
              )}
            >
              {r.cells.map((cell, i) => (
                <td
                  key={columns[i]?.key ?? i}
                  className={cn(
                    "pr-4 align-middle last:pr-0 font-sans text-[12px]",
                    rowClassName,
                    columns[i]?.align === "right" ? "text-right" : "text-left",
                    selectedId === r.id && i === 0 && "relative",
                  )}
                >
                  {selectedId === r.id && i === 0 && (
                    <span
                      aria-hidden
                      className="absolute top-0 bottom-0 -left-2 w-[2px] bg-amber-intel"
                    />
                  )}
                  {cell}
                </td>
              ))}
            </tr>
          ))}
          {rows.length === 0 && (
            <tr>
              <td colSpan={columns.length} className="py-4 text-[12px] text-muted-foreground">
                {empty}
              </td>
            </tr>
          )}
        </tbody>
      </table>
    </div>
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
          <dd
            className={cn(
              "font-mono text-[28px] leading-none font-medium tracking-[-0.05em] tabular",
              m.tone,
            )}
          >
            {m.value}
          </dd>
          <dt className="mt-2 font-mono text-[9.5px] font-semibold tracking-[0.12em] text-muted-foreground uppercase">
            {m.label}
          </dt>
          {m.sub && <dd className="mt-0.5 font-sans text-[11px] text-muted-foreground">{m.sub}</dd>}
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
          <dd
            className={cn(
              "font-mono text-[20px] leading-none font-semibold tracking-[-0.05em] tabular",
              m.tone,
            )}
          >
            {m.value}
          </dd>
          <dt className="mt-1.5 font-mono text-[9.5px] font-semibold tracking-[0.12em] text-muted-foreground uppercase">
            {m.label}
          </dt>
          {m.sub && (
            <dd className="mt-0.5 font-sans text-[10.5px] text-muted-foreground">{m.sub}</dd>
          )}
        </div>
      ))}
    </dl>
  );
}

/** Label/value row used inside inspectors. Aligned, quiet label, mono-friendly value. */
export function InspectorRow({ label, children }: { label: string; children: ReactNode }) {
  return (
    <div className="flex items-baseline justify-between gap-4 border-b border-border py-[5.5px] last:border-b-0">
      <span className="font-mono text-[9.5px] font-semibold tracking-[0.10em] whitespace-nowrap text-muted-foreground uppercase">
        {label}
      </span>
      <span className="min-w-0 truncate text-right font-sans text-[12px] text-foreground">
        {children}
      </span>
    </div>
  );
}
