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
      { title: "Operations — ROPUS" },
      {
        name: "description",
        content:
          "Service health, error budgets and the operational event log for the decision path.",
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
  HEALTHY: "border-approve/40 bg-approve/10 text-approve",
  DEGRADED: "border-warning/45 bg-warning/12 text-warning",
  UNAVAILABLE: "border-block/45 bg-block/12 text-block",
};

const sevTone: Record<OpsEvent["severity"], { text: string; rail: string }> = {
  INFO: { text: "text-muted-foreground", rail: "bg-border-strong" },
  WARN: { text: "text-warning", rail: "bg-warning" },
  CRITICAL: { text: "text-block", rail: "bg-block" },
};

function StateBadge({ state }: { state: ServiceState }) {
  return (
    <span
      className={cn(
        "inline-flex items-center gap-1.5 rounded-[3px] border px-1.5 py-[1px] text-[10px] font-bold tracking-[0.07em] uppercase",
        stateBadge[state],
      )}
    >
      <span aria-hidden className="size-[5px] bg-current" />
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
        title="Operations"
        subtitle="Health of the synchronous decision path and everything behind it. Degradation in graph or streaming does not stop decisioning; it reduces context."
      />

      <TelemetryStrip
        className="mt-3"
        items={[
          { label: "Components", value: String(metrics.services.length), sub: "instrumented" },
          {
            label: "Degraded",
            value: String(degraded.length),
            sub: degraded.map((d) => d.name).join(", ") || "none",
            tone: degraded.length ? "text-warning" : "",
          },
          {
            label: "Decision p99",
            value: `${metrics.p99LatencyMs.toFixed(1)} ms`,
            sub: "objective < 80 ms",
          },
          { label: "Availability", value: "99.995%", sub: "trailing 30 days" },
          { label: "Open incidents", value: "1", sub: "graph traversal latency" },
        ]}
      />

      {degraded.length > 0 && (
        <div className="mt-4 border border-border border-l-2 border-l-warning bg-warning/[0.06] px-4 py-3">
          <div className="flex items-baseline gap-2">
            <span className="text-[10.5px] font-bold tracking-[0.08em] text-warning uppercase">
              Current impact
            </span>
            <Mono className="text-[11px] text-muted-foreground">
              {degraded.length} components degraded
            </Mono>
          </div>
          <p className="mt-1 max-w-[110ch] text-[12.5px] text-muted-foreground">
            Synchronous decisioning remains healthy. Graph context is delayed and 3-hop traversals
            are queued. Decision stream consumer lag is isolated to partition 6.
          </p>
        </div>
      )}

      <div className="mt-6 grid gap-8 xl:grid-cols-[minmax(0,1fr)_360px] xl:gap-10">
        <section className="min-w-0">
          <SectionHead title="Components" meta={`${metrics.services.length} instrumented`} />
          <div className="mt-1 overflow-x-auto">
            <table className="w-full min-w-[720px] border-collapse text-[12.5px]">
              <thead>
                <tr className="border-b border-border">
                  {["Component", "State", "p99", "Detail"].map((h, i) => (
                    <th
                      key={h}
                      scope="col"
                      className={cn(
                        "py-1.5 pr-4 text-[10.5px] font-semibold tracking-[0.08em] whitespace-nowrap text-muted-foreground uppercase last:pr-0",
                        i === 2 ? "text-right" : "text-left",
                      )}
                    >
                      {h}
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {metrics.services.map((s) => {
                  const bad = s.state !== "HEALTHY";
                  return (
                    <tr
                      key={s.name}
                      className={cn(
                        "border-b border-border last:border-b-0 hover:bg-accent",
                        bad && "bg-warning/[0.05]",
                      )}
                    >
                      <td className="relative py-2.5 pr-4 align-middle">
                        {bad && (
                          <span
                            aria-hidden
                            className="absolute top-0 bottom-0 -left-2 w-[2px] bg-warning"
                          />
                        )}
                        <span className="text-[13px] font-semibold">{s.name}</span>
                      </td>
                      <td className="py-2.5 pr-4 align-middle">
                        <StateBadge state={s.state} />
                      </td>
                      <td className="py-2.5 pr-4 text-right align-middle">
                        <Mono
                          className={cn(
                            "font-semibold",
                            bad ? "text-warning" : "text-muted-foreground",
                          )}
                        >
                          {s.p99Ms ? `${s.p99Ms.toFixed(1)} ms` : "—"}
                        </Mono>
                      </td>
                      <td className="py-2.5 align-middle text-[12px] text-muted-foreground">
                        {s.detail}
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>

          <div className="mt-8">
            <SectionHead title="Event log" meta="last 24 hours" />
            <ol className="mt-1">
              {opsEvents.map((e) => {
                const tone = sevTone[e.severity];
                return (
                  <li
                    key={e.id}
                    className="flex items-baseline gap-4 border-b border-border py-2.5 last:border-b-0 hover:bg-accent"
                  >
                    <span
                      aria-hidden
                      className={cn("h-3 w-[2px] shrink-0 self-center", tone.rail)}
                    />
                    <Mono className="w-[150px] shrink-0 text-[11.5px] text-muted-foreground">
                      {e.at}
                    </Mono>
                    <span
                      className={cn(
                        "w-[70px] shrink-0 text-[10px] font-bold tracking-[0.08em]",
                        tone.text,
                      )}
                    >
                      {e.severity}
                    </span>
                    <span className="w-[140px] shrink-0 text-[12.5px] font-semibold">
                      {e.component}
                    </span>
                    <span className="min-w-0 text-[12.5px] text-muted-foreground">{e.text}</span>
                  </li>
                );
              })}
            </ol>
          </div>
        </section>

        <aside className="min-w-0">
          <SectionHead title="Error budgets" meta="30-day window" />
          <div className="mt-1">
            {slos.map((s) => {
              const used = Math.min(1, s.budget);
              const exhausted = used >= 1;
              return (
                <div key={s.name} className="border-b border-border py-2.5 last:border-b-0">
                  <div className="flex items-baseline justify-between gap-4">
                    <span className="text-[12.5px] font-medium">{s.name}</span>
                    <Mono className={cn("font-semibold", s.healthy ? undefined : "text-warning")}>
                      {s.actual}
                    </Mono>
                  </div>
                  <div className="mt-1.5 h-[6px] w-full bg-neutral-surface">
                    <div
                      className={cn(
                        "h-full",
                        exhausted ? "bg-block" : s.healthy ? "bg-approve" : "bg-warning",
                      )}
                      style={{ width: `${used * 100}%` }}
                    />
                  </div>
                  <div className="mt-1 flex items-baseline justify-between gap-3 text-[11px] text-muted-foreground">
                    <span>
                      target <Mono className="text-[11px]">{s.target}</Mono>
                    </span>
                    <span>
                      <Mono className="text-[11px]">{(used * 100).toFixed(0)}%</Mono> consumed ·{" "}
                      <Mono className="text-[11px]">{((1 - used) * 100).toFixed(0)}%</Mono>{" "}
                      remaining
                    </span>
                  </div>
                  {exhausted && (
                    <div className="mt-1.5 inline-flex items-center gap-1.5 border border-block/40 bg-block/10 px-1.5 py-[1px] text-[10px] font-bold tracking-[0.07em] text-block uppercase">
                      <span aria-hidden className="size-[5px] bg-current" />
                      Action needed
                    </div>
                  )}
                </div>
              );
            })}
          </div>

          <div className="mt-8">
            <SectionHead title="Open incident" meta="1 active" />
            <div className="mt-1 border-b border-border py-2.5">
              <div className="flex items-baseline justify-between gap-3">
                <span className="text-[12.5px] font-semibold">Graph traversal latency</span>
                <Mono className="text-[11px] text-muted-foreground">INC-2261</Mono>
              </div>
              <dl className="mt-2">
                {[
                  ["Severity", "SEV-3"],
                  ["Opened", "2026-08-22 17:38Z"],
                  ["Component", "Graph service"],
                  ["Owner", "r.duarte"],
                ].map(([k, v]) => (
                  <div key={k} className="flex items-baseline justify-between gap-4 py-[3px]">
                    <dt className="text-[10.5px] font-semibold tracking-[0.07em] text-muted-foreground uppercase">
                      {k}
                    </dt>
                    <dd>
                      <Mono className="text-[11.5px]">{v}</Mono>
                    </dd>
                  </div>
                ))}
              </dl>
              <p className="mt-2 text-[11.5px] text-muted-foreground">
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
