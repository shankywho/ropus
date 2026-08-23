import { createFileRoute, Link } from "@tanstack/react-router";
import { useSuspenseQuery } from "@tanstack/react-query";
import { Mono, RiskScore, VerdictBadge } from "@/components/ropus/core";
import { DataGrid, Page, PageHead, SectionHead } from "@/components/ropus/page";
import { decisionsQuery, overviewQuery, casesQuery } from "@/lib/ropus/api";

export const Route = createFileRoute("/")({
  head: () => ({
    meta: [
      { title: "Risk Control Plane Overview — ROPUS" },
      {
        name: "description",
        content: "Monitor risk decisions, identify threats, and manage investigations.",
      },
      { property: "og:title", content: "Risk Control Plane Overview — ROPUS" },
      {
        property: "og:description",
        content: "Monitor risk decisions, identify threats, and manage investigations.",
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

const time = (iso: string) => iso.slice(11, 19) + "Z";

function Overview() {
  const { data: metrics } = useSuspenseQuery(overviewQuery());
  const { data: decisions } = useSuspenseQuery(decisionsQuery());
  const { data: cases } = useSuspenseQuery(casesQuery());

  const queue = cases.filter((c) => c.status !== "CLOSED");
  const highRiskDec = decisions.find((d) => d.riskScore >= 0.8) ?? decisions[0];

  return (
    <Page>
      <PageHead
        title="Overview"
        subtitle="Monitor risk decisions, identify threats, and manage investigations."
      />

      {/* ------------------------------------------------ Top 3 Metrics */}
      <dl className="mt-4 grid grid-cols-1 divide-y divide-border border-b border-border sm:grid-cols-3 sm:divide-y-0 sm:divide-x">
        <div className="py-4 sm:pr-6">
          <dt className="text-[11.5px] font-semibold tracking-wider text-muted-foreground uppercase">
            Evaluations (24h)
          </dt>
          <dd className="mt-1.5 font-mono text-[28px] leading-none font-bold text-foreground tabular">
            {metrics.evaluations.toLocaleString()}
          </dd>
        </div>
        <div className="py-4 sm:px-6">
          <dt className="text-[11.5px] font-semibold tracking-wider text-muted-foreground uppercase">
            Block rate
          </dt>
          <dd className="mt-1.5 font-mono text-[28px] leading-none font-bold text-foreground tabular">
            {(metrics.blockRate * 100).toFixed(2)}%
          </dd>
        </div>
        <div className="py-4 sm:pl-6">
          <dt className="text-[11.5px] font-semibold tracking-wider text-muted-foreground uppercase">
            Open cases
          </dt>
          <dd className="mt-1.5 font-mono text-[28px] leading-none font-bold text-foreground tabular">
            {metrics.openCases ?? queue.length}
          </dd>
        </div>
      </dl>

      {/* ------------------------------------------------ High-Risk Incident Banner */}
      {highRiskDec && (
        <section
          aria-label="Latest high-risk incident"
          className="mt-6 flex flex-wrap items-center justify-between gap-4 rounded border border-block/40 bg-block/5 px-4 py-3.5"
        >
          <div className="flex items-center gap-3">
            <span className="flex size-2.5 shrink-0 rounded-full bg-block" />
            <div>
              <div className="flex items-center gap-2">
                <span className="font-mono text-[11px] font-bold tracking-wider text-block uppercase">
                  LATEST HIGH-RISK INCIDENT
                </span>
                <span className="font-mono text-[11px] text-muted-foreground">
                  · {highRiskDec.amount.toLocaleString("en-US", { minimumFractionDigits: 2 })}{" "}
                  {highRiskDec.currency}
                </span>
              </div>
              <p className="mt-0.5 text-[13px] text-foreground">
                <Mono className="font-bold">{highRiskDec.transactionId}</Mono> blocked due to{" "}
                {highRiskDec.primarySignal.toLowerCase()}. Score:{" "}
                <span className="font-mono font-bold text-block">
                  {highRiskDec.riskScore.toFixed(2)}
                </span>
              </p>
            </div>
          </div>
          <div className="flex items-center gap-2.5">
            <Link
              to="/decisions/$decisionId"
              params={{ decisionId: highRiskDec.decisionId }}
              className="rounded border border-block/40 bg-surface px-3 py-1.5 font-mono text-[12px] font-bold text-block transition-colors hover:bg-block hover:text-white"
            >
              Inspect Decision
            </Link>
            <Link
              to="/cases/$caseId"
              params={{ caseId: highRiskDec.caseId ?? "CASE-88419" }}
              className="rounded bg-block px-3 py-1.5 font-mono text-[12px] font-bold text-white transition-colors hover:bg-block/90"
            >
              Review Case
            </Link>
          </div>
        </section>
      )}

      {/* ------------------------------------------------ Recent Decisions */}
      <section className="mt-8">
        <SectionHead
          title="Recent decisions"
          meta={`${Math.min(decisions.length, 8)} of ${metrics.evaluations.toLocaleString()}`}
        />
        <div className="mt-2">
          <DataGrid
            columns={[
              { key: "time", label: "Time" },
              { key: "txn", label: "Transaction" },
              { key: "cus", label: "Customer" },
              { key: "amount", label: "Amount", align: "right" },
              { key: "risk", label: "Risk", align: "right" },
              { key: "verdict", label: "Verdict" },
              { key: "signal", label: "Primary signal" },
            ]}
            rows={decisions.slice(0, 8).map((d) => ({
              id: d.decisionId,
              cells: [
                <Mono className="text-muted-foreground">{time(d.evaluatedAt)}</Mono>,
                <Link
                  to="/decisions/$decisionId"
                  params={{ decisionId: d.decisionId }}
                  className="font-mono text-[12.5px] text-primary hover:underline"
                >
                  {d.transactionId}
                </Link>,
                <Mono className="text-muted-foreground">{d.customerId}</Mono>,
                <Mono>{`${d.amount.toLocaleString("en-US", { minimumFractionDigits: 2 })} ${d.currency}`}</Mono>,
                <RiskScore value={d.riskScore} />,
                <VerdictBadge verdict={d.verdict} />,
                <span className="text-muted-foreground">{d.primarySignal}</span>,
              ],
            }))}
          />
        </div>
      </section>
    </Page>
  );
}
