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
        title="AI Risk Manager"
        subtitle="Detect coordinated financial fraud without blocking legitimate customers."
      />

      {/* ------------------------------------------------ Problem Taste & Core Thesis Banner */}
      <section
        aria-label="Core Engineering Thesis"
        className="mt-4 rounded border border-primary/30 bg-primary/5 p-4"
      >
        <div className="flex flex-wrap items-center justify-between gap-3">
          <div>
            <div className="flex items-center gap-2 font-mono text-[11px] font-bold text-primary uppercase">
              <span>● MULTI-SIGNAL DEFENSE ARCHITECTURE</span>
            </div>
            <h2 className="mt-1 text-[15px] font-bold text-foreground">
              "One signal is weak. The combination is not."
            </h2>
            <p className="mt-0.5 max-w-3xl text-[12.5px] leading-relaxed text-muted-foreground">
              An international IP or a higher transaction amount looks benign in isolation. Real
              financial crime is only revealed when deterministic rules, calibrated machine learning
              probabilities, and entity graph topologies converge into an auditable defensive
              verdict.
            </p>
          </div>
          <Link
            to="/demo"
            className="rounded bg-primary px-4 py-2 font-mono text-[12px] font-bold text-white shadow-xs transition-opacity hover:opacity-90 whitespace-nowrap"
          >
            Launch 7-Stage Walkthrough →
          </Link>
        </div>
      </section>

      {/* ------------------------------------------------ System Summary & Measured Evaluation Metrics */}
      <dl className="mt-6 grid grid-cols-2 divide-y divide-border border-b border-t border-border sm:grid-cols-3 lg:grid-cols-5 sm:divide-y-0 sm:divide-x">
        <div className="py-4 sm:pr-4">
          <div className="flex items-center justify-between">
            <dt className="text-[11px] font-semibold tracking-wider text-muted-foreground uppercase">
              Evaluations (24h)
            </dt>
            <span className="font-mono text-[9px] font-bold text-approve uppercase">
              Live/Local
            </span>
          </div>
          <dd className="mt-1.5 font-mono text-[26px] leading-none font-bold text-foreground tabular">
            {metrics.evaluations.toLocaleString()}
          </dd>
        </div>
        <div className="py-4 sm:px-4">
          <div className="flex items-center justify-between">
            <dt className="text-[11px] font-semibold tracking-wider text-muted-foreground uppercase">
              Block rate
            </dt>
            <span className="font-mono text-[9px] font-bold text-approve uppercase">
              Live/Local
            </span>
          </div>
          <dd className="mt-1.5 font-mono text-[26px] leading-none font-bold text-foreground tabular">
            {(metrics.blockRate * 100).toFixed(2)}%
          </dd>
        </div>
        <div className="py-4 sm:px-4">
          <div className="flex items-center justify-between">
            <dt className="text-[11px] font-semibold tracking-wider text-muted-foreground uppercase">
              Precision
            </dt>
            <span className="font-mono text-[9px] font-bold text-primary uppercase">
              Offline eval
            </span>
          </div>
          <dd className="mt-1.5 font-mono text-[26px] leading-none font-bold text-foreground tabular">
            10.08%
          </dd>
          <div className="mt-1 font-mono text-[10px] text-muted-foreground">
            TP/(TP+FP) · 13/129
          </div>
        </div>
        <div className="py-4 sm:px-4">
          <div className="flex items-center justify-between">
            <dt className="text-[11px] font-semibold tracking-wider text-muted-foreground uppercase">
              Recall
            </dt>
            <span className="font-mono text-[9px] font-bold text-primary uppercase">
              Offline eval
            </span>
          </div>
          <dd className="mt-1.5 font-mono text-[26px] leading-none font-bold text-foreground tabular">
            25.00%
          </dd>
          <div className="mt-1 font-mono text-[10px] text-muted-foreground">TP/(TP+FN) · 13/52</div>
        </div>
        <div className="py-4 sm:pl-4">
          <div className="flex items-center justify-between">
            <dt className="text-[11px] font-semibold tracking-wider text-muted-foreground uppercase">
              FP Cost
            </dt>
            <span className="font-mono text-[9px] font-bold text-primary uppercase">
              Offline eval
            </span>
          </div>
          <dd className="mt-1.5 font-mono text-[26px] leading-none font-bold text-foreground tabular">
            ₹40,000 <span className="text-[14px] font-normal text-muted-foreground">/ FP</span>
          </dd>
          <div className="mt-1 font-mono text-[10px] text-muted-foreground">
            ₹46.4L total (116 FP on test set)
          </div>
        </div>
      </dl>

      {/* ------------------------------------------------ High-Risk Incident & Evidence Section */}
      <section
        aria-label="Current high-risk incident"
        className="mt-6 rounded border border-block/40 bg-block/5 p-5"
      >
        <div className="flex flex-wrap items-start justify-between gap-4 border-b border-block/20 pb-4">
          <div>
            <div className="flex items-center gap-2">
              <span className="flex size-2.5 shrink-0 rounded-full bg-block" />
              <span className="font-mono text-[11px] font-bold tracking-wider text-block uppercase">
                REPRESENTATIVE INCIDENT: ₹14,50,000.00 INR OUTBOUND IMPS PAYOUT
              </span>
              <span className="rounded bg-block/10 px-2 py-0.5 font-mono text-[10px] font-bold text-block">
                VERDICT: BLOCK (0.96)
              </span>
            </div>
            <h2 className="mt-1.5 text-[16px] font-bold text-foreground">
              Account Takeover &amp; Money Mule Drainage Attempt (
              <Mono className="text-[14px]">txn_order_88419</Mono>)
            </h2>
            <p className="mt-0.5 text-[12.5px] text-muted-foreground">
              Customer <Mono className="font-bold text-foreground">cus_4471029</Mono> initiated an
              IMPS payout from Limassol proxy 12 minutes after an active Bengaluru session.
            </p>
          </div>
          <div className="flex items-center gap-2">
            <Link
              to="/demo"
              className="rounded border border-block/40 bg-surface px-3 py-1.5 font-mono text-[12px] font-bold text-block transition-colors hover:bg-block hover:text-white"
            >
              Inspect Walkthrough →
            </Link>
          </div>
        </div>

        {/* Concise Evidence Signals Grid */}
        <div className="mt-4">
          <div className="text-[11px] font-semibold tracking-wider text-muted-foreground uppercase">
            Converging Risk Signals &amp; Layer Attribution
          </div>
          <div className="mt-2.5 grid grid-cols-1 gap-2.5 sm:grid-cols-2 lg:grid-cols-3">
            <div className="rounded border border-border bg-surface p-3">
              <div className="flex items-center justify-between">
                <span className="font-mono text-[11px] font-bold text-block">
                  1. Impossible Travel
                </span>
                <span className="font-mono text-[11px] font-bold text-block">+0.21</span>
              </div>
              <p className="mt-1 text-[12px] text-muted-foreground">
                Session from Limassol proxy 12 min after Bengaluru (7,250 km at 36,250 km/h vs 900
                km/h flight ceiling).
              </p>
            </div>

            <div className="rounded border border-border bg-surface p-3">
              <div className="flex items-center justify-between">
                <span className="font-mono text-[11px] font-bold text-block">
                  2. Datacenter Proxy ASN
                </span>
                <span className="font-mono text-[11px] font-bold text-block">+0.18</span>
              </div>
              <p className="mt-1 text-[12px] text-muted-foreground">
                IP <Mono className="text-[11px]">198.51.100.44</Mono> matches commercial bulletproof
                hosting ASN 13335.
              </p>
            </div>

            <div className="rounded border border-border bg-surface p-3">
              <div className="flex items-center justify-between">
                <span className="font-mono text-[11px] font-bold text-block">
                  3. Device Novelty
                </span>
                <span className="font-mono text-[11px] font-bold text-block">+0.18</span>
              </div>
              <p className="mt-1 text-[12px] text-muted-foreground">
                Hardware canvas <Mono className="text-[11px]">9f8a84b12c</Mono> matches headless
                virtualized emulator.
              </p>
            </div>

            <div className="rounded border border-border bg-surface p-3">
              <div className="flex items-center justify-between">
                <span className="font-mono text-[11px] font-bold text-block">
                  4. Velocity Surge
                </span>
                <span className="font-mono text-[11px] font-bold text-block">+0.22</span>
              </div>
              <p className="mt-1 text-[12px] text-muted-foreground">
                ₹14,50,000 amount represents +412% 1h velocity surge over 24h average expenditure
                baseline.
              </p>
            </div>

            <div className="rounded border border-border bg-surface p-3">
              <div className="flex items-center justify-between">
                <span className="font-mono text-[11px] font-bold text-block">
                  5. Fraud Graph Syndicate
                </span>
                <span className="font-mono text-[11px] font-bold text-block">+0.17</span>
              </div>
              <p className="mt-1 text-[12px] text-muted-foreground">
                3-hop BFS expands to 14 synthetic accounts linked to payout node{" "}
                <Mono className="text-[11px]">PA-77120</Mono>.
              </p>
            </div>

            <div className="rounded border border-border bg-surface p-3">
              <div className="flex items-center justify-between">
                <span className="font-mono text-[11px] font-bold text-primary">
                  6. Calibrated ML Posterior
                </span>
                <span className="font-mono text-[11px] font-bold text-primary">+0.20</span>
              </div>
              <p className="mt-1 text-[12px] text-muted-foreground">
                25-feature ONNX XGBoost tree inference scores P(fraud|x) = 0.9418 (Beta-calibrated).
              </p>
            </div>
          </div>
        </div>
      </section>

      {/* ------------------------------------------------ Technique Alignment Matrix */}
      <section className="mt-8">
        <SectionHead
          title="Architecture &amp; Technique Decomposition"
          meta="Right tool for each layer"
        />
        <div className="mt-2.5 grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-4">
          <div className="border border-border p-3.5 rounded bg-surface">
            <div className="font-mono text-[11px] font-bold text-primary">DETERMINISTIC RULES</div>
            <div className="mt-1 text-[12px] text-muted-foreground">
              Best for statutory sanctions (OFAC), explicit velocity limits, and hard instant blocks
              (0.4ms latency).
            </div>
          </div>
          <div className="border border-border p-3.5 rounded bg-surface">
            <div className="font-mono text-[11px] font-bold text-primary">
              CALIBRATED XGBOOST ML
            </div>
            <div className="mt-1 text-[12px] text-muted-foreground">
              Best for continuous high-dimensional non-linear feature interactions and empirical
              probability calibration (2.1ms).
            </div>
          </div>
          <div className="border border-border p-3.5 rounded bg-surface">
            <div className="font-mono text-[11px] font-bold text-primary">
              FRAUD KNOWLEDGE GRAPH
            </div>
            <div className="mt-1 text-[12px] text-muted-foreground">
              Best for discovering coordinated money mule rings, device sharing, and synthetic
              identity clusters (1.2ms).
            </div>
          </div>
          <div className="border border-border p-3.5 rounded bg-surface">
            <div className="font-mono text-[11px] font-bold text-primary">HUMAN GOVERNANCE</div>
            <div className="mt-1 text-[12px] text-muted-foreground">
              Owns consequential actions with cryptographic hash-chain audit proofs and closed-loop
              retraining signals.
            </div>
          </div>
        </div>
      </section>

      {/* ------------------------------------------------ Recent Decisions */}
      <section className="mt-8">
        <SectionHead
          title="Recent risk decisions"
          meta={`${Math.min(decisions.length, 6)} of ${metrics.evaluations.toLocaleString()} total`}
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
            rows={decisions.slice(0, 6).map((d) => ({
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
