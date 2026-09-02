import { createFileRoute } from "@tanstack/react-router";
import { useSuspenseQuery } from "@tanstack/react-query";
import { Page, PageHead, SectionHead, TelemetryStrip } from "@/components/ropus/page";
import { Mono } from "@/components/ropus/core";
import { overviewQuery } from "@/lib/ropus/api";
import { opsEvents, slos, type OpsEvent } from "@/lib/ropus/platform-fixtures";
import type { ServiceState } from "@/lib/ropus/contracts";
import { cn } from "@/lib/utils";

export const Route = createFileRoute("/operations")({
  head: () => ({
    meta: [
      { title: "Operations & SLO Monitor — ROPUS" },
      {
        name: "description",
        content:
          "Service health, error budgets and the operational event log for the synchronous risk decision path.",
      },
      { property: "og:title", content: "Operations — ROPUS" },
      {
        property: "og:description",
        content: "Component health, SLO budgets and recent operational events.",
      },
      { property: "og:type", content: "website" },
      { name: "twitter:card", content: "summary" },
    ],
  }),
  loader: ({ context }) => context.queryClient.ensureQueryData(overviewQuery()),
  component: OperationsPage,
});

const stateBadge: Record<ServiceState, string> = {
  HEALTHY: "border-authoritative/40 bg-approve-surface text-authoritative",
  DEGRADED: "border-amber-intel/45 bg-shadow-intel-surface text-shadow-intel",
  UNAVAILABLE: "border-blocked/45 bg-blocked-surface text-blocked",
};

const sevTone: Record<OpsEvent["severity"], { text: string; rail: string }> = {
  INFO: { text: "text-muted-foreground", rail: "bg-border-strong" },
  WARN: { text: "text-shadow-intel", rail: "bg-amber-intel" },
  CRITICAL: { text: "text-blocked", rail: "bg-blocked" },
};

function StateBadge({ state }: { state: ServiceState }) {
  return (
    <span
      className={cn(
        "inline-flex items-center gap-1.5 rounded-[2px] border px-1.5 py-[1px] font-mono text-[9.5px] font-semibold tracking-[0.06em] uppercase",
        stateBadge[state],
      )}
    >
      <span aria-hidden className="size-[5px] bg-current rounded-full" />
      {state}
    </span>
  );
}

function OperationsPage() {
  const { data: metrics } = useSuspenseQuery(overviewQuery());
  const degraded = metrics.services.filter((s) => s.state !== "HEALTHY");

  return (
    <Page>
      <PageHead
        title="SRE Operations & SLO Monitor"
        subtitle="Health of the synchronous decision path and background streaming infrastructure. Degradation in shadow graph or event streaming does not block live decisioning."
      />

      <TelemetryStrip
        className="mt-3"
        items={[
          { label: "Components", value: String(metrics.services.length), sub: "instrumented" },
          {
            label: "Degraded",
            value: String(degraded.length),
            sub: degraded.map((d) => d.name).join(", ") || "none",
            tone: degraded.length ? "text-amber-intel font-bold" : "text-authoritative",
          },
          {
            label: "Decision p99",
            value: `${metrics.p99LatencyMs.toFixed(1)} ms`,
            sub: "objective < 80 ms",
          },
          {
            label: "Availability",
            value: "99.995%",
            sub: "trailing 30 days",
            tone: "text-authoritative",
          },
          {
            label: "Open incidents",
            value: "1",
            sub: "graph traversal latency",
            tone: "text-shadow-intel",
          },
        ]}
      />

      {degraded.length > 0 && (
        <div className="mt-4 border border-border border-l-4 border-l-amber-intel bg-shadow-intel-surface px-4 py-3 shadow-xs font-mono text-[11.5px]">
          <div className="flex items-baseline gap-2">
            <span className="text-[10px] font-bold tracking-[0.10em] text-shadow-intel uppercase">
              Current impact
            </span>
            <Mono className="text-[11px] text-muted-foreground">
              {degraded.length} components degraded
            </Mono>
          </div>
          <p className="mt-1 max-w-[110ch] text-[12px] font-sans text-muted-foreground leading-relaxed">
            Synchronous decisioning remains healthy. Graph context is delayed and 3-hop traversals
            are queued. Decision stream consumer lag is isolated to partition 6.
          </p>
        </div>
      )}

      <div className="mt-6 grid gap-8 xl:grid-cols-[minmax(0,1fr)_360px] xl:gap-10">
        <section className="min-w-0 space-y-4">
          <SectionHead title="SYSTEM COMPONENTS" meta={`${metrics.services.length} instrumented`} />
          <div className="border border-border bg-card overflow-x-auto shadow-xs">
            <table className="w-full min-w-[720px] border-collapse font-sans text-[12px]">
              <thead>
                <tr className="border-b border-border bg-secondary/40 font-mono text-[9px] uppercase font-bold text-muted-foreground tracking-wider">
                  {["Component", "State", "p99", "Detail"].map((h, i) => (
                    <th
                      key={h}
                      scope="col"
                      className={cn(
                        "py-2 px-3 whitespace-nowrap text-muted-foreground last:pr-3",
                        i === 2 ? "text-right" : "text-left",
                      )}
                    >
                      {h}
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody className="divide-y divide-border">
                {metrics.services.map((s) => {
                  const bad = s.state !== "HEALTHY";
                  return (
                    <tr
                      key={s.name}
                      className={cn(
                        "hover:bg-secondary/40 transition-colors",
                        bad && "bg-shadow-intel-surface/40",
                      )}
                    >
                      <td className="relative py-2.5 px-3 align-middle font-semibold text-foreground">
                        {bad && (
                          <span
                            aria-hidden
                            className="absolute top-0 bottom-0 left-0 w-[2px] bg-amber-intel"
                          />
                        )}
                        <span>{s.name}</span>
                      </td>
                      <td className="py-2.5 px-3 align-middle">
                        <StateBadge state={s.state} />
                      </td>
                      <td className="py-2.5 px-3 text-right align-middle">
                        <Mono
                          className={cn(
                            "font-semibold text-[11.5px] tabular",
                            bad ? "text-shadow-intel" : "text-muted-foreground",
                          )}
                        >
                          {s.p99Ms ? `${s.p99Ms.toFixed(1)} ms` : "—"}
                        </Mono>
                      </td>
                      <td className="py-2.5 px-3 align-middle text-[11.5px] text-muted-foreground font-sans">
                        {s.detail}
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>

          <div className="mt-8 space-y-3">
            <SectionHead title="OPERATIONAL EVENT LOG" meta="last 24 hours" />
            <ol className="border border-border bg-card divide-y divide-border shadow-xs font-mono text-[11px]">
              {opsEvents.map((e) => {
                const tone = sevTone[e.severity];
                return (
                  <li
                    key={e.id}
                    className="flex items-baseline gap-3 px-3 py-2.5 hover:bg-secondary/40 transition-colors"
                  >
                    <span
                      aria-hidden
                      className={cn("h-3 w-[2px] shrink-0 self-center", tone.rail)}
                    />
                    <Mono className="w-[140px] shrink-0 text-[10.5px] text-muted-foreground tabular">
                      {e.at}
                    </Mono>
                    <span
                      className={cn(
                        "w-[65px] shrink-0 text-[9.5px] font-bold tracking-[0.08em] uppercase",
                        tone.text,
                      )}
                    >
                      {e.severity}
                    </span>
                    <span className="w-[140px] shrink-0 font-sans text-[12px] font-semibold text-foreground">
                      {e.component}
                    </span>
                    <span className="min-w-0 font-sans text-[12px] text-muted-foreground">
                      {e.text}
                    </span>
                  </li>
                );
              })}
            </ol>
          </div>
        </section>

        <aside className="min-w-0 space-y-6">
          <div>
            <SectionHead title="ERROR BUDGETS" meta="30-day window" />
            <div className="border border-border bg-card p-4 shadow-xs space-y-3 font-mono text-[11.5px]">
              {slos.map((s) => {
                const used = Math.min(1, s.budget);
                const exhausted = used >= 1;
                return (
                  <div
                    key={s.name}
                    className="border-b border-border/50 pb-3 last:border-b-0 last:pb-0"
                  >
                    <div className="flex items-baseline justify-between gap-4">
                      <span className="font-sans text-[12px] font-medium text-foreground">
                        {s.name}
                      </span>
                      <Mono
                        className={cn(
                          "font-bold tabular text-[11.5px]",
                          s.healthy ? "text-authoritative" : "text-amber-intel",
                        )}
                      >
                        {s.actual}
                      </Mono>
                    </div>
                    <div className="mt-1.5 h-[5px] w-full bg-secondary overflow-hidden rounded-xs">
                      <div
                        className={cn(
                          "h-full",
                          exhausted
                            ? "bg-blocked"
                            : s.healthy
                              ? "bg-authoritative"
                              : "bg-amber-intel",
                        )}
                        style={{ width: `${used * 100}%` }}
                      />
                    </div>
                    <div className="mt-1 flex items-baseline justify-between gap-3 text-[10px] text-muted-foreground font-mono">
                      <span>
                        target <Mono className="text-[10px]">{s.target}</Mono>
                      </span>
                      <span>
                        <Mono className="text-[10px]">{(used * 100).toFixed(0)}%</Mono> consumed ·{" "}
                        <Mono className="text-[10px]">{((1 - used) * 100).toFixed(0)}%</Mono>{" "}
                        remaining
                      </span>
                    </div>
                    {exhausted && (
                      <div className="mt-1.5 inline-flex items-center gap-1.5 border border-blocked/40 bg-blocked-surface px-1.5 py-[1px] font-mono text-[9px] font-bold tracking-[0.06em] text-blocked uppercase">
                        <span aria-hidden className="size-[4px] bg-current rounded-full" />
                        Action needed
                      </div>
                    )}
                  </div>
                );
              })}
            </div>
          </div>

          <div>
            <SectionHead title="OPEN INCIDENT" meta="1 active" />
            <div className="border border-border bg-card p-4 shadow-xs font-mono text-[11px] space-y-2">
              <div className="flex items-baseline justify-between gap-3 border-b border-border pb-2">
                <span className="font-sans text-[12.5px] font-bold text-foreground">
                  Graph traversal latency
                </span>
                <Mono className="text-[10.5px] text-shadow-intel font-bold">INC-2261</Mono>
              </div>
              <dl className="space-y-1">
                {[
                  ["Severity", "SEV-3"],
                  ["Opened", "2026-08-22 17:38Z"],
                  ["Component", "Graph service"],
                  ["Owner", "r.duarte"],
                ].map(([k, v]) => (
                  <div key={k} className="flex items-baseline justify-between gap-4 py-[2px]">
                    <dt className="text-[9.5px] font-semibold tracking-[0.07em] text-muted-foreground uppercase">
                      {k}
                    </dt>
                    <dd>
                      <Mono className="text-[11px] font-medium text-foreground">{v}</Mono>
                    </dd>
                  </div>
                ))}
              </dl>
              <p className="mt-2 text-[11.5px] text-muted-foreground font-sans leading-relaxed border-t border-border pt-2">
                Graph traversal has consumed its full budget for this window. 3-hop traversals are
                queued behind 1- and 2-hop requests until the backlog clears.
              </p>
            </div>
          </div>
        </aside>
      </div>
    </Page>
  );
}
