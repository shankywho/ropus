import { createFileRoute, Link } from "@tanstack/react-router";
import { useSuspenseQuery } from "@tanstack/react-query";
import { Mono, RiskScore, VerdictBadge, StatusPill } from "@/components/ropus/core";
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

  return (
    <Page>
      <PageHead
        title="AI Risk Manager"
        subtitle="Institutional multi-signal financial crime defense and GraphSAGE relationship intelligence."
      />

      {/* ------------------------------------------------ Problem Taste & Core Thesis Banner */}
      <section
        aria-label="Core Engineering Thesis"
        className="mt-4 border border-border border-l-4 border-l-navy bg-card p-4 shadow-xs"
      >
        <div className="flex flex-wrap items-center justify-between gap-3">
          <div>
            <div className="flex items-center gap-2 font-mono text-[10px] font-medium tracking-[0.12em] text-navy uppercase">
              <span>● MULTI-SIGNAL DEFENSE ARCHITECTURE</span>
            </div>
            <h2 className="mt-1 font-sans text-[15px] font-bold text-foreground">
              "One signal is weak. The combination is decisive."
            </h2>
            <p className="mt-0.5 max-w-3xl font-sans text-[12px] leading-relaxed text-muted-foreground">
              An international IP or an elevated transaction amount looks benign in isolation. Real
              financial crime is only revealed when deterministic rules, calibrated machine learning
              probabilities (BMR Champion), and inductive graph topologies (GraphSAGE Shadow) converge into an auditable defensive
              verdict.
            </p>
          </div>
          <Link
            to="/demo"
            className="border border-navy bg-navy px-3.5 py-1.5 font-mono text-[11px] font-medium text-white shadow-xs transition-opacity hover:opacity-90 whitespace-nowrap"
          >
            Launch 7-Stage Walkthrough →
          </Link>
        </div>
      </section>

      {/* ------------------------------------------------ System Summary & Measured Evaluation Metrics */}
      <dl className="mt-6 grid grid-cols-2 divide-y divide-border border-b border-t border-border sm:grid-cols-3 lg:grid-cols-5 sm:divide-y-0 sm:divide-x bg-card">
        <div className="p-4 sm:pr-4">
          <div className="flex items-center justify-between">
            <dt className="font-mono text-[9.5px] font-medium tracking-[0.12em] text-muted-foreground uppercase">
              Evaluations (24h)
            </dt>
            <StatusPill tone="authoritative">LIVE</StatusPill>
          </div>
          <dd className="mt-2 font-mono text-[28px] leading-none font-medium tracking-[-0.06em] text-foreground tabular">
            {metrics.evaluations.toLocaleString()}
          </dd>
        </div>
        <div className="p-4 sm:px-4">
          <div className="flex items-center justify-between">
            <dt className="font-mono text-[9.5px] font-medium tracking-[0.12em] text-muted-foreground uppercase">
              Block Rate
            </dt>
            <StatusPill tone="authoritative">LIVE</StatusPill>
          </div>
          <dd className="mt-2 font-mono text-[28px] leading-none font-medium tracking-[-0.06em] text-foreground tabular">
            {(metrics.blockRate * 100).toFixed(2)}%
          </dd>
        </div>
        <div className="p-4 sm:px-4">
          <div className="flex items-center justify-between">
            <dt className="font-mono text-[9.5px] font-medium tracking-[0.12em] text-muted-foreground uppercase">
              Precision
            </dt>
            <StatusPill tone="local">OFFLINE</StatusPill>
          </div>
          <dd className="mt-2 font-mono text-[28px] leading-none font-medium tracking-[-0.06em] text-foreground tabular">
            10.08%
          </dd>
          <div className="mt-1 font-mono text-[9.5px] text-muted-foreground">
            TP/(TP+FP) · 13/129
          </div>
        </div>
        <div className="p-4 sm:px-4">
          <div className="flex items-center justify-between">
            <dt className="font-mono text-[9.5px] font-medium tracking-[0.12em] text-muted-foreground uppercase">
              Recall
            </dt>
            <StatusPill tone="local">OFFLINE</StatusPill>
          </div>
          <dd className="mt-2 font-mono text-[28px] leading-none font-medium tracking-[-0.06em] text-foreground tabular">
            25.00%
          </dd>
          <div className="mt-1 font-mono text-[9.5px] text-muted-foreground">TP/(TP+FN) · 13/52</div>
        </div>
        <div className="p-4 sm:pl-4">
          <div className="flex items-center justify-between">
            <dt className="font-mono text-[9.5px] font-medium tracking-[0.12em] text-muted-foreground uppercase">
              FP Cost
            </dt>
            <StatusPill tone="local">OFFLINE</StatusPill>
          </div>
          <dd className="mt-2 font-mono text-[24px] leading-none font-medium tracking-[-0.06em] text-foreground tabular">
            ₹40,000 <span className="text-[12px] font-normal text-muted-foreground">/ FP</span>
          </dd>
          <div className="mt-1 font-mono text-[9.5px] text-muted-foreground">
            ₹46.4L total (116 FP holdout)
          </div>
        </div>
      </dl>

      {/* ------------------------------------------------ High-Risk Incident & Evidence Section */}
      <section
        aria-label="Current high-risk incident"
        className="mt-6 border border-blocked/40 border-l-4 border-l-destructive bg-blocked-surface/30 p-5"
      >
        <div className="flex flex-wrap items-start justify-between gap-4 border-b border-blocked/20 pb-4">
          <div>
            <div className="flex items-center gap-2">
              <span className="flex size-2 shrink-0 rounded-full bg-destructive" />
              <span className="font-mono text-[10px] font-medium tracking-[0.1em] text-blocked uppercase">
                REPRESENTATIVE INCIDENT: ₹14,50,000.00 INR OUTBOUND IMPS PAYOUT
              </span>
              <StatusPill tone="blocked">VERDICT: BLOCK (0.96)</StatusPill>
            </div>
            <h2 className="mt-1.5 font-sans text-[16px] font-bold text-foreground">
              Account Takeover &amp; Money Mule Drainage Attempt (
              <Mono className="text-[13px]">txn_order_88419</Mono>)
            </h2>
            <p className="mt-0.5 font-sans text-[12px] text-muted-foreground">
              Customer <Mono className="font-semibold text-foreground">cus_4471029</Mono> initiated an
              IMPS payout from Limassol proxy 12 minutes after an active Bengaluru session.
            </p>
          </div>
          <div className="flex items-center gap-2">
            <Link
              to="/demo"
              className="border border-blocked/40 bg-surface px-3 py-1.5 font-mono text-[11px] font-medium text-blocked transition-colors hover:bg-destructive hover:text-white"
            >
              Inspect Walkthrough →
            </Link>
          </div>
        </div>

        {/* Concise Evidence Signals Grid */}
        <div className="mt-4">
          <div className="font-mono text-[10px] font-medium tracking-[0.12em] text-muted-foreground uppercase">
            Converging Risk Signals &amp; Layer Attribution
          </div>
          <div className="mt-2.5 grid grid-cols-1 gap-2.5 sm:grid-cols-2 lg:grid-cols-3">
            <div className="border border-border bg-card p-3 shadow-2xs">
              <div className="flex items-center justify-between">
                <span className="font-mono text-[10.5px] font-medium text-blocked">
                  1. Impossible Travel
                </span>
                <span className="font-mono text-[11px] font-medium text-blocked">+0.21</span>
              </div>
              <p className="mt-1 font-sans text-[11.5px] text-muted-foreground">
                Session from Limassol proxy 12 min after Bengaluru (7,250 km at 36,250 km/h vs 900
                km/h flight ceiling).
              </p>
            </div>

            <div className="border border-border bg-card p-3 shadow-2xs">
              <div className="flex items-center justify-between">
                <span className="font-mono text-[10.5px] font-medium text-blocked">
                  2. Datacenter Proxy ASN
                </span>
                <span className="font-mono text-[11px] font-medium text-blocked">+0.18</span>
              </div>
              <p className="mt-1 font-sans text-[11.5px] text-muted-foreground">
                IP <Mono className="text-[11px]">198.51.100.44</Mono> matches commercial bulletproof
                hosting ASN 13335.
              </p>
            </div>

            <div className="border border-border bg-card p-3 shadow-2xs">
              <div className="flex items-center justify-between">
                <span className="font-mono text-[10.5px] font-medium text-blocked">
                  3. Device Novelty
                </span>
                <span className="font-mono text-[11px] font-medium text-blocked">+0.18</span>
              </div>
              <p className="mt-1 font-sans text-[11.5px] text-muted-foreground">
                Hardware canvas <Mono className="text-[11px]">9f8a84b12c</Mono> matches headless
                virtualized emulator.
              </p>
            </div>

            <div className="border border-border bg-card p-3 shadow-2xs">
              <div className="flex items-center justify-between">
                <span className="font-mono text-[10.5px] font-medium text-blocked">
                  4. Velocity Surge
                </span>
                <span className="font-mono text-[11px] font-medium text-blocked">+0.22</span>
              </div>
              <p className="mt-1 font-sans text-[11.5px] text-muted-foreground">
                ₹14,50,000 amount represents +412% 1h velocity surge over 24h average expenditure
                baseline.
              </p>
            </div>

            <div className="border border-border bg-card p-3 shadow-2xs">
              <div className="flex items-center justify-between">
                <span className="font-mono text-[10.5px] font-medium text-shadow-intel">
                  5. GraphSAGE Shadow Ring
                </span>
                <span className="font-mono text-[11px] font-medium text-shadow-intel">+0.17</span>
              </div>
              <p className="mt-1 font-sans text-[11.5px] text-muted-foreground">
                2-hop inductive GNN expands to 14 synthetic accounts linked to payout node{" "}
                <Mono className="text-[10.5px]">PA-77120</Mono>.
              </p>
            </div>

            <div className="border border-border bg-card p-3 shadow-2xs">
              <div className="flex items-center justify-between">
                <span className="font-mono text-[10.5px] font-medium text-authoritative">
                  6. Authoritative BMR ML
                </span>
                <span className="font-mono text-[11px] font-medium text-authoritative">+0.20</span>
              </div>
              <p className="mt-1 font-sans text-[11.5px] text-muted-foreground">
                25-feature ONNX XGBoost tree inference scores P(fraud|x) = 0.9418 (Beta-calibrated Champion).
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
          <div className="border border-border p-3.5 bg-card shadow-2xs">
            <div className="font-mono text-[10px] font-medium tracking-[0.1em] text-navy uppercase">1. DETERMINISTIC RULES</div>
            <div className="mt-1.5 font-sans text-[11.5px] leading-relaxed text-muted-foreground">
              Best for statutory sanctions (OFAC), explicit velocity limits, and hard instant blocks
              (0.4ms latency).
            </div>
          </div>
          <div className="border border-border p-3.5 bg-card shadow-2xs">
            <div className="font-mono text-[10px] font-medium tracking-[0.1em] text-authoritative uppercase">
              2. AUTHORITATIVE BMR ML
            </div>
            <div className="mt-1.5 font-sans text-[11.5px] leading-relaxed text-muted-foreground">
              Best for continuous high-dimensional non-linear feature interactions and empirical
              probability calibration (2.1ms).
            </div>
          </div>
          <div className="border border-border p-3.5 bg-card shadow-2xs">
            <div className="font-mono text-[10px] font-medium tracking-[0.1em] text-shadow-intel uppercase">
              3. GRAPHSAGE SHADOW
            </div>
            <div className="mt-1.5 font-sans text-[11.5px] leading-relaxed text-muted-foreground">
              Best for discovering coordinated money mule rings, employee collusion, and synthetic
              identity clusters (Shadow mode).
            </div>
          </div>
          <div className="border border-border p-3.5 bg-card shadow-2xs">
            <div className="font-mono text-[10px] font-medium tracking-[0.1em] text-navy uppercase">4. HUMAN GOVERNANCE</div>
            <div className="mt-1.5 font-sans text-[11.5px] leading-relaxed text-muted-foreground">
              Owns consequential actions with cryptographic hash-chain audit proofs and closed-loop
              retraining signals.
            </div>
          </div>
        </div>
      </section>

      {/* ------------------------------------------------ Recent Decisions */}
      <section className="mt-8">
        <SectionHead
          title="Recent Risk Decisions"
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
              { key: "signal", label: "Primary Signal" },
            ]}
            rows={decisions.slice(0, 6).map((d) => ({
              id: d.decisionId,
              cells: [
                <Mono className="text-muted-foreground">{time(d.evaluatedAt)}</Mono>,
                <Link
                  to="/decisions/$decisionId"
                  params={{ decisionId: d.decisionId }}
                  className="font-mono text-[11.5px] text-navy font-medium hover:underline"
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
