import { createFileRoute, Link } from "@tanstack/react-router";
import { useSuspenseQuery } from "@tanstack/react-query";
import { useState, useMemo } from "react";
import { Mono, RiskScore, VerdictBadge, StatusPill } from "@/components/ropus/core";
import { DataGrid, Page, PageHead, SectionHead } from "@/components/ropus/page";
import { decisionsQuery, overviewQuery, casesQuery, isLiveBackend } from "@/lib/ropus/api";
import {
  ShieldAlert,
  PlayCircle,
  Network,
  Cpu,
  FileCode2,
  Lock,
  ArrowRight,
  Activity,
  Zap,
  Globe2,
  Sparkles,
  Layers,
  Flame,
} from "lucide-react";

export const Route = createFileRoute("/")({
  head: () => ({
    meta: [
      { title: "AI Risk Manager — ROPUS Risk Control Plane" },
      {
        name: "description",
        content:
          "Enterprise real-time payments fraud, abuse, and chargeback defense platform with multi-signal decisioning, ONNX ML, and GraphSAGE relationship intelligence.",
      },
      { property: "og:title", content: "AI Risk Manager — ROPUS" },
      {
        property: "og:description",
        content:
          "Multi-signal financial crime defense, deterministic AST rules, calibrated BMR ML, and GraphSAGE relationship intelligence.",
      },
      { property: "og:type", content: "website" },
      { name: "twitter:card", content: "summary_large_image" },
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

  const [filterVerdict, setFilterVerdict] = useState<string>("ALL");
  const [searchTerm, setSearchTerm] = useState<string>("");

  const filteredDecisions = useMemo(() => {
    return decisions.filter((d) => {
      const matchVerdict = filterVerdict === "ALL" || d.verdict === filterVerdict;
      const matchSearch =
        searchTerm === "" ||
        d.transactionId.toLowerCase().includes(searchTerm.toLowerCase()) ||
        d.customerId.toLowerCase().includes(searchTerm.toLowerCase()) ||
        d.primarySignal.toLowerCase().includes(searchTerm.toLowerCase());
      return matchVerdict && matchSearch;
    });
  }, [decisions, filterVerdict, searchTerm]);

  const openCasesCount = cases.filter((c) => c.status !== "CLOSED").length;

  return (
    <Page>
      {/* ------------------------------------------------ Page Header */}
      <div className="flex flex-wrap items-start justify-between gap-4 border-b border-border pb-4">
        <div>
          <div className="flex items-center gap-2">
            <span className="border border-navy bg-navy/10 px-2 py-0.5 font-mono text-[9.5px] font-bold text-navy uppercase tracking-[0.06em]">
              COMMAND DECK
            </span>
            <span className="font-mono text-[11px] text-muted-foreground">
              Real-Time Payments Fraud &amp; Financial Crime Defense
            </span>
          </div>
          <h1 className="mt-1 font-sans text-[26px] lg:text-[28px] font-extrabold leading-[1.12] tracking-[-0.04em] text-foreground">
            Risk Control Plane
          </h1>
          <p className="mt-0.5 font-sans text-[12.5px] text-muted-foreground">
            Multi-stage synchronous decisioning (&lt;100ms SLA), deterministic AST rules, Bayes
            Minimum Risk (BMR) cost optimization, and GraphSAGE relationship intelligence.
          </p>
        </div>

        <div className="flex items-center gap-2.5">
          <Link
            to="/demo"
            className="border border-navy bg-navy px-4 py-2 font-mono text-[11.5px] font-bold text-white shadow-xs hover:bg-navy/90 transition-all flex items-center gap-2 cursor-pointer"
          >
            <PlayCircle className="size-4 text-amber-intel" />
            <span>Launch 7-Stage Walkthrough →</span>
          </Link>
        </div>
      </div>

      {/* ------------------------------------------------ Live Real-Time Telemetry Ribbon */}
      <div className="mt-4 flex flex-wrap items-center justify-between gap-3 border border-border bg-card px-4 py-2.5 shadow-2xs font-mono text-[11px]">
        <div className="flex items-center gap-4">
          <div className="flex items-center gap-1.5">
            <span className="size-2 rounded-full bg-emerald-500 animate-pulse" />
            <span className="font-bold text-foreground">SYNCHRONOUS ENGINE:</span>
            <span className="text-emerald-700 font-semibold">&lt;18.4ms p95 SLA</span>
          </div>
          <span className="text-border-strong hidden sm:inline">|</span>
          <div className="hidden sm:flex items-center gap-1.5 text-muted-foreground">
            <span>Throughput:</span>
            <strong className="text-foreground font-semibold">1,240 req/sec</strong>
          </div>
          <span className="text-border-strong hidden md:inline">|</span>
          <div className="hidden md:flex items-center gap-1.5 text-muted-foreground">
            <span>Model:</span>
            <strong className="text-navy font-bold">XGBoost 25F (Beta Calibrated)</strong>
          </div>
        </div>

        <div className="flex items-center gap-3">
          <div className="flex items-center gap-1 text-[10.5px]">
            <span className="text-muted-foreground">Shadow GNN:</span>
            <span className="border border-amber-intel/50 bg-shadow-intel-surface px-1.5 py-0.2 font-bold text-shadow-intel">
              NON-ENFORCING (0% AUTH)
            </span>
          </div>
          <Link
            to="/operations"
            className="text-navy hover:underline font-bold text-[10.5px] flex items-center gap-1"
          >
            <span>SLO Monitor</span>
            <ArrowRight className="size-3" />
          </Link>
        </div>
      </div>

      {/* ------------------------------------------------ YC Founder / Architectural Thesis Banner */}
      <section
        aria-label="Core Engineering Thesis"
        className="mt-4 border border-border border-l-4 border-l-navy bg-card p-5 shadow-xs"
      >
        <div className="flex flex-wrap items-center justify-between gap-4">
          <div className="space-y-1.5">
            <div className="flex items-center gap-2 font-mono text-[9.5px] font-bold tracking-[0.12em] text-navy uppercase">
              <Sparkles className="size-3.5 text-navy" />
              <span>MULTI-SIGNAL DEFENSE ARCHITECTURE</span>
            </div>
            <h2 className="font-sans text-[17px] font-extrabold text-foreground tracking-tight">
              &ldquo;One signal is weak. The combination is decisive.&rdquo;
            </h2>
            <p className="max-w-3xl font-sans text-[12.5px] leading-relaxed text-muted-foreground">
              An international IP or an elevated transaction amount looks benign in isolation. Real
              financial crime is only revealed when deterministic rules, calibrated machine learning
              probabilities (BMR Champion), and inductive graph topologies (GraphSAGE Shadow)
              converge into an auditable defensive verdict.
            </p>
          </div>

          {/* Quick Action Matrix */}
          <div className="grid grid-cols-2 gap-2 font-mono text-[10.5px]">
            <Link
              to="/graph"
              className="border border-border bg-surface hover:bg-secondary p-2.5 flex items-center gap-2 transition-colors shadow-2xs"
            >
              <Network className="size-3.5 text-navy shrink-0" />
              <div>
                <div className="font-bold text-foreground">Fraud Graph</div>
                <div className="text-[9.5px] text-muted-foreground">3-Hop BFS Mule Ring</div>
              </div>
            </Link>
            <Link
              to="/models"
              className="border border-border bg-surface hover:bg-secondary p-2.5 flex items-center gap-2 transition-colors shadow-2xs"
            >
              <Cpu className="size-3.5 text-navy shrink-0" />
              <div>
                <div className="font-bold text-foreground">BMR Optimizer</div>
                <div className="text-[9.5px] text-muted-foreground">Loss Matrix Simulator</div>
              </div>
            </Link>
            <Link
              to="/rules"
              className="border border-border bg-surface hover:bg-secondary p-2.5 flex items-center gap-2 transition-colors shadow-2xs"
            >
              <FileCode2 className="size-3.5 text-navy shrink-0" />
              <div>
                <div className="font-bold text-foreground">Rules Engine</div>
                <div className="text-[9.5px] text-muted-foreground">JSON-AST Dual Control</div>
              </div>
            </Link>
            <Link
              to="/cases"
              className="border border-border bg-surface hover:bg-secondary p-2.5 flex items-center gap-2 transition-colors shadow-2xs"
            >
              <ShieldAlert className="size-3.5 text-blocked shrink-0" />
              <div>
                <div className="font-bold text-foreground">Case Queue</div>
                <div className="text-[9.5px] text-blocked font-bold">
                  {openCasesCount} Active (24h SLA)
                </div>
              </div>
            </Link>
          </div>
        </div>
      </section>

      {/* ------------------------------------------------ Executive KPI Grid */}
      <dl className="mt-5 grid grid-cols-2 divide-y divide-border border-b border-t border-border sm:grid-cols-3 lg:grid-cols-5 sm:divide-y-0 sm:divide-x bg-card shadow-xs">
        <div className="p-4 sm:pr-4">
          <div className="flex items-center justify-between">
            <dt className="font-mono text-[9.5px] font-semibold tracking-[0.12em] text-muted-foreground uppercase">
              Evaluations (24h)
            </dt>
            <StatusPill tone="authoritative">LIVE</StatusPill>
          </div>
          <dd className="mt-2 font-mono text-[28px] leading-none font-extrabold tracking-[-0.05em] text-foreground tabular">
            {metrics.evaluations.toLocaleString()}
          </dd>
          <div className="mt-1 font-mono text-[9.5px] text-emerald-700 font-semibold">
            +18.4% vs 7-day average
          </div>
        </div>

        <div className="p-4 sm:px-4">
          <div className="flex items-center justify-between">
            <dt className="font-mono text-[9.5px] font-semibold tracking-[0.12em] text-muted-foreground uppercase">
              Block Rate
            </dt>
            <StatusPill tone="authoritative">LIVE</StatusPill>
          </div>
          <dd className="mt-2 font-mono text-[28px] leading-none font-extrabold tracking-[-0.05em] text-foreground tabular">
            {(metrics.blockRate * 100).toFixed(2)}%
          </dd>
          <div className="mt-1 font-mono text-[9.5px] text-muted-foreground">
            Strict statutory + AST hard blocks
          </div>
        </div>

        <div className="p-4 sm:px-4">
          <div className="flex items-center justify-between">
            <dt className="font-mono text-[9.5px] font-semibold tracking-[0.12em] text-muted-foreground uppercase">
              Fraud Prevented
            </dt>
            <StatusPill tone="authoritative">24H</StatusPill>
          </div>
          <dd className="mt-2 font-mono text-[26px] leading-none font-extrabold tracking-[-0.05em] text-navy tabular">
            ₹3.82 Cr
          </dd>
          <div className="mt-1 font-mono text-[9.5px] text-muted-foreground">
            $460,000 USD equivalent
          </div>
        </div>

        <div className="p-4 sm:px-4">
          <div className="flex items-center justify-between">
            <dt className="font-mono text-[9.5px] font-semibold tracking-[0.12em] text-muted-foreground uppercase">
              Serving p95 SLA
            </dt>
            <StatusPill tone="authoritative">LIVE</StatusPill>
          </div>
          <dd className="mt-2 font-mono text-[28px] leading-none font-extrabold tracking-[-0.05em] text-foreground tabular">
            18.4 ms
          </dd>
          <div className="mt-1 font-mono text-[9.5px] text-emerald-700 font-semibold">
            Objective &lt;100ms (p99: 62.4ms)
          </div>
        </div>

        <div className="p-4 sm:pl-4">
          <div className="flex items-center justify-between">
            <dt className="font-mono text-[9.5px] font-semibold tracking-[0.12em] text-muted-foreground uppercase">
              FP Friction Cost
            </dt>
            <StatusPill tone="local">BMR OPT</StatusPill>
          </div>
          <dd className="mt-2 font-mono text-[24px] leading-none font-extrabold tracking-[-0.05em] text-foreground tabular">
            ₹40,000 <span className="text-[12px] font-normal text-muted-foreground">/ FP</span>
          </dd>
          <div className="mt-1 font-mono text-[9.5px] text-muted-foreground">
            Minimized by Bayes loss matrix
          </div>
        </div>
      </dl>

      {/* ------------------------------------------------ High-Risk Incident & Converging Evidence Section */}
      <section
        aria-label="Current high-risk incident"
        className="mt-6 border border-blocked/40 border-l-4 border-l-destructive bg-blocked-surface/30 p-5 shadow-xs"
      >
        <div className="flex flex-wrap items-start justify-between gap-4 border-b border-blocked/20 pb-4">
          <div>
            <div className="flex items-center gap-2">
              <span className="flex size-2 shrink-0 rounded-full bg-destructive animate-ping" />
              <span className="font-mono text-[9.5px] font-bold tracking-[0.10em] text-blocked uppercase">
                ACTIVE INCIDENT INVESTIGATION · OUTBOUND IMPS PAYOUT
              </span>
              <StatusPill tone="blocked">VERDICT: BLOCK (0.96)</StatusPill>
            </div>
            <h2 className="mt-1.5 font-sans text-[17px] font-extrabold text-foreground tracking-tight">
              Account Takeover &amp; Money Mule Drainage Attempt (
              <Mono className="text-[14px] font-bold">txn_order_88419</Mono>)
            </h2>
            <p className="mt-0.5 font-sans text-[12.5px] text-muted-foreground">
              Customer <Mono className="font-bold text-foreground">cus_4471029</Mono> initiated an
              IMPS payout of <strong className="text-foreground">₹14,50,000.00 INR</strong> from
              Limassol proxy 12 minutes after an active Bengaluru residential session.
            </p>
          </div>
          <div className="flex items-center gap-2">
            <Link
              to="/decisions/$decisionId"
              params={{ decisionId: "dec_88419_attack" }}
              className="border border-border bg-surface px-3 py-1.5 font-mono text-[11px] font-semibold text-foreground hover:bg-secondary transition-colors"
            >
              TreeSHAP Breakdown →
            </Link>
            <Link
              to="/demo"
              className="border border-blocked/60 bg-destructive px-3.5 py-1.5 font-mono text-[11px] font-bold text-white transition-opacity hover:opacity-90 shadow-xs"
            >
              Step-by-Step Replay →
            </Link>
          </div>
        </div>

        {/* Converging Risk Signals Grid */}
        <div className="mt-4">
          <div className="font-mono text-[9.5px] font-bold tracking-[0.12em] text-muted-foreground uppercase">
            Converging Risk Signals &amp; Layer Attribution (+0.94 Combined Score)
          </div>
          <div className="mt-2.5 grid grid-cols-1 gap-2.5 sm:grid-cols-2 lg:grid-cols-3">
            <div className="border border-border bg-card p-3 shadow-2xs">
              <div className="flex items-center justify-between">
                <span className="font-mono text-[10.5px] font-bold text-blocked">
                  1. Impossible Travel
                </span>
                <span className="font-mono text-[11px] font-bold text-blocked tabular">+0.21</span>
              </div>
              <p className="mt-1 font-sans text-[11.5px] text-muted-foreground leading-relaxed">
                Session from Limassol proxy 12 min after Bengaluru (7,250 km at 36,250 km/h vs 900
                km/h flight ceiling).
              </p>
            </div>

            <div className="border border-border bg-card p-3 shadow-2xs">
              <div className="flex items-center justify-between">
                <span className="font-mono text-[10.5px] font-bold text-blocked">
                  2. Datacenter Proxy ASN
                </span>
                <span className="font-mono text-[11px] font-bold text-blocked tabular">+0.18</span>
              </div>
              <p className="mt-1 font-sans text-[11.5px] text-muted-foreground leading-relaxed">
                IP <Mono className="text-[11px]">198.51.100.44</Mono> matches commercial bulletproof
                hosting ASN 13335.
              </p>
            </div>

            <div className="border border-border bg-card p-3 shadow-2xs">
              <div className="flex items-center justify-between">
                <span className="font-mono text-[10.5px] font-bold text-blocked">
                  3. Device Novelty
                </span>
                <span className="font-mono text-[11px] font-bold text-blocked tabular">+0.18</span>
              </div>
              <p className="mt-1 font-sans text-[11.5px] text-muted-foreground leading-relaxed">
                Hardware canvas <Mono className="text-[11px]">9f8a84b12c</Mono> matches headless
                virtualized emulator.
              </p>
            </div>

            <div className="border border-border bg-card p-3 shadow-2xs">
              <div className="flex items-center justify-between">
                <span className="font-mono text-[10.5px] font-bold text-blocked">
                  4. Velocity Surge
                </span>
                <span className="font-mono text-[11px] font-bold text-blocked tabular">+0.22</span>
              </div>
              <p className="mt-1 font-sans text-[11.5px] text-muted-foreground leading-relaxed">
                ₹14,50,000 amount represents +412% 1h velocity surge over 24h average expenditure
                baseline profile.
              </p>
            </div>

            <div className="border border-border bg-card p-3 shadow-2xs">
              <div className="flex items-center justify-between">
                <span className="font-mono text-[10.5px] font-bold text-shadow-intel">
                  5. GraphSAGE Shadow Ring
                </span>
                <span className="font-mono text-[11px] font-bold text-shadow-intel tabular">
                  +0.17
                </span>
              </div>
              <p className="mt-1 font-sans text-[11.5px] text-muted-foreground leading-relaxed">
                2-hop inductive GNN expands to 14 synthetic accounts linked to payout node{" "}
                <Mono className="text-[10.5px]">PA-77120</Mono>.
              </p>
            </div>

            <div className="border border-border bg-card p-3 shadow-2xs">
              <div className="flex items-center justify-between">
                <span className="font-mono text-[10.5px] font-bold text-authoritative">
                  6. Authoritative BMR ML
                </span>
                <span className="font-mono text-[11px] font-bold text-authoritative tabular">
                  +0.20
                </span>
              </div>
              <p className="mt-1 font-sans text-[11.5px] text-muted-foreground leading-relaxed">
                25-feature ONNX XGBoost tree inference scores P(fraud|x) = 0.9418 (Beta-calibrated
                Champion).
              </p>
            </div>
          </div>
        </div>
      </section>

      {/* ------------------------------------------------ Technique Alignment Matrix */}
      <section className="mt-8">
        <SectionHead
          title="ARCHITECTURE &amp; TECHNIQUE DECOMPOSITION"
          meta="Right defense tool for each layer"
        />
        <div className="mt-2.5 grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-4">
          <div className="border border-border p-3.5 bg-card shadow-2xs">
            <div className="font-mono text-[9.5px] font-bold tracking-[0.10em] text-navy uppercase flex items-center gap-1.5">
              <FileCode2 className="size-3.5" />
              <span>1. DETERMINISTIC RULES</span>
            </div>
            <div className="mt-1.5 font-sans text-[11.5px] leading-relaxed text-muted-foreground">
              Best for statutory sanctions (OFAC), explicit velocity limits, and hard instant blocks
              (0.4ms latency).
            </div>
          </div>
          <div className="border border-border p-3.5 bg-card shadow-2xs">
            <div className="font-mono text-[9.5px] font-bold tracking-[0.10em] text-authoritative uppercase flex items-center gap-1.5">
              <Cpu className="size-3.5" />
              <span>2. AUTHORITATIVE BMR ML</span>
            </div>
            <div className="mt-1.5 font-sans text-[11.5px] leading-relaxed text-muted-foreground">
              Best for continuous high-dimensional non-linear feature interactions and empirical
              probability calibration (2.1ms).
            </div>
          </div>
          <div className="border border-border p-3.5 bg-card shadow-2xs">
            <div className="font-mono text-[9.5px] font-bold tracking-[0.10em] text-shadow-intel uppercase flex items-center gap-1.5">
              <Network className="size-3.5" />
              <span>3. GRAPHSAGE SHADOW</span>
            </div>
            <div className="mt-1.5 font-sans text-[11.5px] leading-relaxed text-muted-foreground">
              Best for discovering coordinated money mule rings, employee collusion, and synthetic
              identity clusters (Shadow mode).
            </div>
          </div>
          <div className="border border-border p-3.5 bg-card shadow-2xs">
            <div className="font-mono text-[9.5px] font-bold tracking-[0.10em] text-navy uppercase flex items-center gap-1.5">
              <Lock className="size-3.5" />
              <span>4. HUMAN GOVERNANCE</span>
            </div>
            <div className="mt-1.5 font-sans text-[11.5px] leading-relaxed text-muted-foreground">
              Owns consequential actions with cryptographic hash-chain audit proofs and closed-loop
              retraining signals.
            </div>
          </div>
        </div>
      </section>

      {/* ------------------------------------------------ Real-Time Decisions Ledger Feed */}
      <section className="mt-8">
        <div className="flex flex-wrap items-center justify-between gap-3 border-b border-border pb-3">
          <SectionHead
            title="REAL-TIME RISK EVALUATIONS"
            meta={`${filteredDecisions.length} of ${decisions.length} decisions`}
          />
          <div className="flex items-center gap-2 font-mono text-[10.5px]">
            {["ALL", "BLOCK", "REVIEW", "CHALLENGE", "APPROVE"].map((v) => (
              <button
                key={v}
                type="button"
                onClick={() => setFilterVerdict(v)}
                className={`px-2 py-0.5 border cursor-pointer ${
                  filterVerdict === v
                    ? "border-navy bg-navy text-white font-bold"
                    : "border-border bg-surface text-muted-foreground hover:text-foreground"
                }`}
              >
                {v}
              </button>
            ))}
            <input
              type="text"
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              placeholder="Search txn, cus, signal..."
              className="border border-border bg-surface px-2 py-0.5 text-[11px] font-mono text-foreground outline-none"
            />
          </div>
        </div>

        <div className="mt-2 border border-border bg-card shadow-xs">
          <DataGrid
            columns={[
              { key: "time", label: "Time" },
              { key: "txn", label: "Transaction" },
              { key: "cus", label: "Customer" },
              { key: "amount", label: "Amount", align: "right" },
              { key: "risk", label: "Risk", align: "right" },
              { key: "verdict", label: "Verdict" },
              { key: "signal", label: "Primary Signal" },
              { key: "action", label: "Inspect" },
            ]}
            rows={filteredDecisions.slice(0, 7).map((d) => ({
              id: d.decisionId,
              cells: [
                <Mono key="time" className="text-muted-foreground">
                  {time(d.evaluatedAt)}
                </Mono>,
                <Link
                  key="txn"
                  to="/decisions/$decisionId"
                  params={{ decisionId: d.decisionId }}
                  className="font-mono text-[11.5px] text-navy font-bold hover:underline"
                >
                  {d.transactionId}
                </Link>,
                <Mono key="cus" className="text-muted-foreground">
                  {d.customerId}
                </Mono>,
                <Mono
                  key="amt"
                  className="font-bold"
                >{`${d.amount.toLocaleString("en-US", { minimumFractionDigits: 2 })} ${d.currency}`}</Mono>,
                <RiskScore key="score" value={d.riskScore} size="sm" showBand />,
                <VerdictBadge key="v" verdict={d.verdict} size="sm" />,
                <span
                  key="sig"
                  className="text-muted-foreground font-sans text-[11.5px] truncate max-w-[200px] block"
                >
                  {d.primarySignal}
                </span>,
                <Link
                  key="act"
                  to="/decisions/$decisionId"
                  params={{ decisionId: d.decisionId }}
                  className="border border-border bg-surface px-2 py-0.5 text-[10.5px] font-mono font-semibold text-foreground hover:bg-secondary"
                >
                  Dossier →
                </Link>,
              ],
            }))}
          />
        </div>
      </section>
    </Page>
  );
}
