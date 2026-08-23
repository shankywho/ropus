import { createFileRoute, Link } from "@tanstack/react-router";
import { useSuspenseQuery } from "@tanstack/react-query";
import { Mono, RiskScore, VerdictBadge } from "@/components/ropus/core";
import { DataGrid, MetricStrip, Page, PageHead, SectionHead } from "@/components/ropus/page";
import { decisionsQuery, overviewQuery } from "@/lib/ropus/api";
import { casesQuery } from "@/lib/ropus/api";
import type { ServiceState, Verdict } from "@/lib/ropus/contracts";
import { cn } from "@/lib/utils";

export const Route = createFileRoute("/")({
  head: () => ({
    meta: [
      { title: "Risk Intelligence Overview — ROPUS" },
      {
        name: "description",
        content:
          "Live view of transaction decisions, verdict distribution, decision latency, open cases and platform health.",
      },
      { property: "og:title", content: "Risk Intelligence Overview — ROPUS" },
      {
        property: "og:description",
        content: "Transaction decisions, verdict distribution, open investigations and platform health.",
      },
      { property: "og:type", content: "website" },
      { name: "twitter:card", content: "summary" },
    ],
  }),
  loader: ({ context }) =>
    Promise.all([
      context.queryClient.ensureQueryData(overviewQuery()),
      context.queryClient.ensureQueryData(decisionsQuery()),
      context.queryClient.ensureQueryData(casesQuery()),
    ]),
  component: Overview,
});

const verdictBar: Record<Verdict, string> = {
  APPROVE: "bg-approve",
  REVIEW: "bg-review",
  CHALLENGE: "bg-challenge",
  BLOCK: "bg-block",
};

const stateTone: Record<ServiceState, string> = {
  HEALTHY: "text-muted-foreground",
  DEGRADED: "text-warning",
  UNAVAILABLE: "text-block",
};

const pct = (n: number) => `${(n * 100).toFixed(2)}%`;
const time = (iso: string) => iso.slice(11, 19) + "Z";

function Overview() {
  const { data: metrics } = useSuspenseQuery(overviewQuery());
  const { data: decisions } = useSuspenseQuery(decisionsQuery());
  const { data: cases } = useSuspenseQuery(casesQuery());

  const total = metrics.distribution.reduce((s, d) => s + d.count, 0);
  const degraded = metrics.services.filter((s) => s.state !== "HEALTHY");
  const queue = cases.filter((c) => c.status !== "CLOSED");

  return (
    <Page>
      <PageHead
        title="Overview"
        subtitle="Decisioning throughput, verdict mix, open investigations and the health of the synchronous decision path."
      />

      <MetricStrip
        items={[
          { label: "Evaluations", value: metrics.evaluations.toLocaleString(), sub: metrics.windowLabel },
          { label: "Block rate", value: pct(metrics.blockRate), sub: "of evaluations" },
          { label: "Review rate", value: pct(metrics.reviewRate), sub: "of evaluations" },
          { label: "p99 latency", value: `${metrics.p99LatencyMs.toFixed(1)} ms`, sub: "decision API" },
          { label: "Open cases", value: String(metrics.openCases), sub: "investigation queue" },
        ]}
      />

      {degraded.length > 0 && (
        <section
          aria-label="System alert"
          className="mt-4 flex flex-wrap items-baseline gap-x-8 gap-y-1.5 border-y border-border border-l-2 border-l-warning bg-surface py-2 pl-3"
        >
          <span className="text-[10.5px] font-semibold tracking-[0.08em] text-warning uppercase">System alert</span>
          <span className="text-[12.5px] font-semibold">{degraded.length} services degraded</span>
          {degraded.map((s) => (
            <span key={s.name} className="text-[12.5px] text-muted-foreground">
              <span className="text-foreground/80">{s.name}</span> — {s.detail.toLowerCase()}
            </span>
          ))}
          <span className="ml-auto pr-3 text-[12px] text-muted-foreground">
            Synchronous decisioning is unaffected.
          </span>
        </section>
      )}


      <div className="mt-6 grid gap-8 xl:grid-cols-[minmax(0,1fr)_300px_300px] xl:gap-10">
        <section className="min-w-0">
          <SectionHead title="Verdict distribution" meta={`${total.toLocaleString()} evaluations`} />
          <div className="mt-3 flex h-1.5 w-full overflow-hidden">
            {metrics.distribution.map((d) => (
              <div
                key={d.verdict}
                className={verdictBar[d.verdict]}
                style={{ width: `${(d.count / total) * 100}%` }}
                title={`${d.verdict} ${((d.count / total) * 100).toFixed(2)}%`}
              />
            ))}
          </div>
          <dl className="mt-4 grid grid-cols-2 gap-x-8 gap-y-4 sm:grid-cols-4">
            {metrics.distribution.map((d) => (
              <div key={d.verdict}>
                <div className="flex items-center gap-1.5">
                  <span aria-hidden className={cn("size-1.5", verdictBar[d.verdict])} />
                  <dt className="text-[10.5px] font-semibold tracking-[0.08em] text-muted-foreground uppercase">
                    {d.verdict}
                  </dt>
                </div>
                <dd className="mt-1.5 font-mono text-[16px] leading-none font-semibold tabular">
                  {((d.count / total) * 100).toFixed(2)}%
                </dd>
                <dd className="mt-1 font-mono text-[11px] text-muted-foreground tabular">
                  {d.count.toLocaleString()}
                </dd>
              </div>
            ))}
          </dl>
        </section>

        <section className="min-w-0">
          <SectionHead title="System status" meta={degraded.length ? `${degraded.length} degraded` : "all healthy"} />
          <table className="mt-1 w-full text-[12.5px]">
            <caption className="sr-only">Component health and p99 latency</caption>
            <tbody>
              {metrics.services.map((s) => (
                <tr key={s.name} className="border-b border-border last:border-b-0">
                  <td className="py-[7px] pr-3">{s.name}</td>
                  <td className={cn("py-[7px] pr-3 text-[10.5px] font-semibold tracking-[0.06em]", stateTone[s.state])}>
                    {s.state}
                  </td>
                  <td className="py-[7px] text-right">
                    <Mono className="text-muted-foreground">{s.p99Ms ? `${s.p99Ms.toFixed(1)}ms` : "—"}</Mono>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </section>

        <section className="min-w-0">
          <SectionHead
            title="Case queue"
            right={
              <Link to="/cases" className="text-[11.5px] text-primary hover:underline">
                All cases →
              </Link>
            }
          />
          <ul className="mt-1 text-[12.5px]">
            {queue.slice(0, 6).map((c) => (
              <li key={c.caseId} className="border-b border-border py-[7px] last:border-b-0">
                <div className="flex items-baseline justify-between gap-3">
                  <Link
                    to="/cases/$caseId"
                    params={{ caseId: c.caseId }}
                    className="font-mono text-[12px] text-primary hover:underline"
                  >
                    {c.caseId}
                  </Link>
                  <span
                    className={cn(
                      "font-mono text-[11.5px] font-semibold",
                      c.priority === "P1" ? "text-block" : "text-muted-foreground",
                    )}
                  >
                    {c.priority}
                  </span>
                  <Mono className="ml-auto truncate text-muted-foreground">{c.assignee ?? "unassigned"}</Mono>
                </div>
                <p className="mt-0.5 truncate text-muted-foreground">{c.summary}</p>
              </li>
            ))}
          </ul>

        </section>
      </div>

      <section className="mt-8">
        <SectionHead
          title="Recent decisions"
          meta={`${decisions.length} of ${metrics.evaluations.toLocaleString()}`}
          right={
            <Link to="/decisions" className="text-[11.5px] text-primary hover:underline">
              All decisions →
            </Link>
          }
        />
        <div className="mt-1">
          <DataGrid
            columns={[
              { key: "time", label: "Time" },
              { key: "txn", label: "Transaction" },
              { key: "cus", label: "Customer" },
              { key: "amount", label: "Amount", align: "right" },
              { key: "risk", label: "Risk", align: "right" },
              { key: "verdict", label: "Verdict" },
              { key: "signal", label: "Primary signal" },
              { key: "latency", label: "Latency", align: "right" },
              { key: "case", label: "Case" },
            ]}
            rows={decisions.map((d) => ({
              id: d.decisionId,
              cells: [
                <Mono className="text-muted-foreground">{time(d.evaluatedAt)}</Mono>,
                <Link
                  to="/decisions/$decisionId"
                  params={{ decisionId: d.decisionId }}
                  className="font-mono text-[12px] text-primary hover:underline"
                >
                  {d.transactionId}
                </Link>,
                <Mono className="text-muted-foreground">{d.customerId}</Mono>,
                <Mono>{d.amount.toLocaleString("en-US", { minimumFractionDigits: 2 })}</Mono>,
                <RiskScore value={d.riskScore} />,
                <VerdictBadge verdict={d.verdict} />,
                <span className="text-muted-foreground">{d.primarySignal}</span>,
                <Mono className="text-muted-foreground">{d.latencyMs.toFixed(1)}ms</Mono>,
                d.caseId ? (
                  <Link
                    to="/cases/$caseId"
                    params={{ caseId: d.caseId }}
                    className="font-mono text-[12px] text-primary hover:underline"
                  >
                    {d.caseId}
                  </Link>
                ) : (
                  <span className="text-muted-foreground">—</span>
                ),
              ],
            }))}
          />
        </div>
      </section>
    </Page>
  );
}
