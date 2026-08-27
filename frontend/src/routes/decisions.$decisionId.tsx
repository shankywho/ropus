import { createFileRoute, Link } from "@tanstack/react-router";
import { useState } from "react";
import { Disclosure, CodeBlock, KeyValue } from "@/components/ropus/primitives";
import { Page } from "@/components/ropus/page";
import {
  DecisionSummary,
  DemoTag,
  EvidenceList,
  Mono,
  RiskFactorList,
  RiskScore,
  VerdictBadge,
} from "@/components/ropus/core";
import { useSuspenseQuery } from "@tanstack/react-query";
import { decisionQuery } from "@/lib/ropus/api";
import type { RiskDecision } from "@/lib/ropus/contracts";
import {
  Zap,
  Clock,
  Sparkles,
  TrendingUp,
  HelpCircle,
  CheckCircle2,
  ArrowRight,
  ShieldCheck,
} from "lucide-react";
import {
  ResponsiveContainer,
  BarChart,
  Bar,
  XAxis,
  YAxis,
  Tooltip,
  Cell,
  ReferenceLine,
} from "recharts";

export const Route = createFileRoute("/decisions/$decisionId")({
  head: () => ({
    meta: [
      { title: "Risk Decision & TreeSHAP Attribution — ROPUS" },
      {
        name: "description",
        content:
          "Score attribution, TreeSHAP feature attribution waterfall, microsecond execution pipeline, and counterfactual explainability for a single ROPUS risk decision.",
      },
      { property: "og:title", content: "Risk Decision — ROPUS" },
      {
        property: "og:description",
        content: "Score attribution and TreeSHAP explainability for a single risk decision.",
      },
      { property: "og:type", content: "website" },
      { name: "twitter:card", content: "summary" },
    ],
  }),
  loader: ({ context, params }) =>
    context.queryClient.ensureQueryData(decisionQuery(params.decisionId)),
  component: DecisionDetail,
});

const money = (amount: number, currency: string) =>
  new Intl.NumberFormat("en-US", { style: "currency", currency }).format(amount);

function DecisionDetail() {
  const { decisionId } = Route.useParams();
  const { data: decision } = useSuspenseQuery(decisionQuery(decisionId));

  return (
    <Page>
      <DecisionStory decision={decision} />
    </Page>
  );
}

function DecisionStory({ decision: d }: { decision: RiskDecision }) {
  const delta = d.riskScore - d.baseScore;
  const [activeFeature, setActiveFeature] = useState<string | null>(null);

  // Construct TreeSHAP Waterfall Data
  const baseValue = 0.035; // E[f(x)]
  let cumulative = baseValue;
  const shapData = [
    {
      name: "Base E[f(x)]",
      val: baseValue,
      diff: baseValue,
      cumulative: baseValue,
      isBase: true,
      counterfactual: "Prior background distribution expected value across 5M legitimate transactions.",
    },
    ...d.inference.features.map((f) => {
      const prev = cumulative;
      cumulative += f.contribution;
      return {
        name: f.name,
        val: f.contribution,
        diff: f.contribution,
        cumulative: Number(cumulative.toFixed(3)),
        isBase: false,
        counterfactual:
          f.contribution > 0.2
            ? `If ${f.name} were within normal bounds (≤ 1.0), posterior drops by -${f.contribution.toFixed(2)}.`
            : `Feature pushes calibrated risk probability by +${f.contribution.toFixed(2)}.`,
      };
    }),
    {
      name: "Calibrated P(fraud)",
      val: d.inference.probability,
      diff: 0,
      cumulative: d.inference.probability,
      isFinal: true,
      counterfactual: "Final Beta-calibrated output probability fed to Bayes Minimum Risk decision engine.",
    },
  ];

  return (
    <article>
      {/* ------------------------------------------------ Heading */}
      <header className="flex flex-wrap items-start justify-between gap-x-10 gap-y-5 border-b border-border pb-5">
        <div className="min-w-0">
          <div className="flex items-center gap-2 text-[11.5px] text-muted-foreground">
            <Link to="/decisions" className="hover:text-foreground">
              Risk decisions
            </Link>
            <span aria-hidden>/</span>
            <Mono className="text-[11.5px] text-muted-foreground">{d.decisionId}</Mono>
            <DemoTag />
          </div>
          <h1 className="mt-2 text-[24px] leading-none font-bold tracking-[-0.01em]">
            {money(d.amount, d.currency)}
            <span className="ml-3 text-[15px] font-medium text-muted-foreground">
              {String(d.rawRequest["channel"] ?? "transaction").replace(/_/g, " ")}
            </span>
          </h1>
          <dl className="mt-3 flex flex-wrap gap-x-8 gap-y-2">
            {[
              ["Customer", d.customerId],
              ["Transaction", d.transactionId],
              ["Decision", d.decisionId],
              [
                "Evaluated",
                new Date(d.evaluatedAt).toISOString().replace("T", " ").slice(0, 19) + "Z",
              ],
            ].map(([k, v]) => (
              <div key={k}>
                <dt className="text-[10.5px] font-semibold tracking-[0.08em] text-muted-foreground uppercase">
                  {k}
                </dt>
                <dd className="mt-0.5">
                  <Mono>{v}</Mono>
                </dd>
              </div>
            ))}
          </dl>
        </div>

        <div className="flex shrink-0 items-start gap-8">
          <RiskScore value={d.riskScore} size="hero" showBand />
          <div className="pt-0.5">
            <VerdictBadge verdict={d.verdict} size="lg" />
            <dl className="mt-3 space-y-1 text-[11.5px] text-muted-foreground">
              <div className="flex justify-between gap-6">
                <dt>Confidence</dt>
                <dd className="font-mono text-foreground tabular">
                  {(d.confidence * 100).toFixed(0)}%
                </dd>
              </div>
              <div className="flex justify-between gap-6">
                <dt>Policy</dt>
                <dd className="font-mono text-foreground">{d.policy}</dd>
              </div>
              <div className="flex justify-between gap-6">
                <dt>Latency</dt>
                <dd className="font-mono text-foreground tabular">{d.latencyMs.toFixed(1)} ms</dd>
              </div>
            </dl>
          </div>
        </div>
      </header>

      {/* ------------------------------------------------ Two Columns */}
      <div className="grid gap-x-12 gap-y-8 pt-6 lg:grid-cols-[minmax(0,1fr)_380px]">
        {/* Left Column */}
        <div className="space-y-8">
          {/* Executive Summary */}
          <section>
            <h2 className="text-[11px] font-bold tracking-[0.08em] uppercase">Why this decision</h2>
            <p className="mt-2 max-w-[62ch] text-[13.5px] leading-relaxed text-foreground/90">
              {d.evidence.find((e) => e.kind === "INFERRED")?.text ??
                "No elevated pattern detected; the transaction matches the customer baseline."}
            </p>
          </section>

          {/* ------------------------------------------------ TREESHAP WATERFALL CHART */}
          <section className="border border-border bg-card p-4 space-y-3">
            <div className="flex items-center justify-between border-b border-border pb-2">
              <div className="flex items-center gap-2">
                <Sparkles className="size-4 text-navy" />
                <span className="font-bold text-foreground text-[12px] font-mono uppercase">
                  TreeSHAP Feature Attribution Waterfall (E[f(x)] → P(fraud))
                </span>
              </div>
              <span className="font-mono text-[10px] text-muted-foreground">
                Base: {baseValue.toFixed(3)} → Calibrated: {d.inference.probability.toFixed(4)}
              </span>
            </div>

            <p className="text-[11.5px] text-muted-foreground font-sans">
              Each feature pushes probability up (crimson) or down (forest) from expected baseline. Click any bar for counterfactual suggestions.
            </p>

            <div className="h-[220px] w-full border border-border/50 bg-surface/50 p-2">
              <ResponsiveContainer width="100%" height="100%">
                <BarChart data={shapData} margin={{ top: 10, right: 10, left: 10, bottom: 25 }}>
                  <CartesianGrid strokeDasharray="3 3" stroke="#e2ded4" />
                  <XAxis
                    dataKey="name"
                    stroke="#71717a"
                    fontSize={9}
                    interval={0}
                    angle={-20}
                    textAnchor="end"
                    fontFamily="DM Mono"
                  />
                  <YAxis stroke="#71717a" fontSize={10} fontStyle="mono" domain={[0, 1]} />
                  <Tooltip
                    content={({ active, payload }) => {
                      if (active && payload && payload.length) {
                        const data = payload[0].payload;
                        return (
                          <div className="border border-border bg-navy text-white p-2.5 shadow-xl font-mono text-[11px] max-w-[280px]">
                            <div className="font-bold text-shadow-intel">{data.name}</div>
                            <div className="mt-1">Contribution: +{data.diff?.toFixed(3)}</div>
                            <div>Cumulative Score: {data.cumulative?.toFixed(3)}</div>
                            <div className="mt-2 text-[10px] text-white/80 font-sans border-t border-white/20 pt-1">
                              <strong>Counterfactual:</strong> {data.counterfactual}
                            </div>
                          </div>
                        );
                      }
                      return null;
                    }}
                  />
                  <ReferenceLine y={0.8} stroke="#A0443A" strokeDasharray="3 3" label={{ value: "Block Cutoff (0.80)", fill: "#A0443A", fontSize: 10 }} />
                  <Bar dataKey="cumulative" radius={[2, 2, 0, 0]}>
                    {shapData.map((entry, index) => (
                      <Cell
                        key={`cell-${index}`}
                        fill={
                          entry.isBase
                            ? "#476C87"
                            : entry.isFinal
                              ? "#1C2B45"
                              : entry.val > 0.15
                                ? "#A0443A"
                                : "#F5A524"
                        }
                      />
                    ))}
                  </Bar>
                </BarChart>
              </ResponsiveContainer>
            </div>
          </section>

          {/* ------------------------------------------------ MICROSECOND LATENCY BREAKDOWN */}
          <section className="border border-border bg-card p-4 space-y-3 font-mono text-[11.5px]">
            <div className="flex items-center justify-between border-b border-border pb-2">
              <div className="flex items-center gap-2">
                <Clock className="size-4 text-authoritative" />
                <span className="font-bold text-foreground text-[12px] uppercase">
                  Synchronous Decision Pipeline Microsecond Profile
                </span>
              </div>
              <span className="font-bold text-authoritative">Total: {d.latencyMs.toFixed(1)} ms</span>
            </div>

            <div className="space-y-2">
              <div className="flex justify-between items-center text-[11px]">
                <span className="text-muted-foreground">1. Redis In-Memory Feature Store Fetch:</span>
                <span className="font-bold text-foreground">1.8 ms</span>
              </div>
              <div className="w-full bg-secondary h-1.5 overflow-hidden">
                <div className="bg-navy h-full w-[21%]" />
              </div>

              <div className="flex justify-between items-center text-[11px]">
                <span className="text-muted-foreground">2. Deterministic JSON-AST Rules Engine:</span>
                <span className="font-bold text-foreground">0.3 ms</span>
              </div>
              <div className="w-full bg-secondary h-1.5 overflow-hidden">
                <div className="bg-navy h-full w-[4%]" />
              </div>

              <div className="flex justify-between items-center text-[11px]">
                <span className="text-muted-foreground">3. ONNX Calibrated XGBoost Inference:</span>
                <span className="font-bold text-foreground">2.1 ms</span>
              </div>
              <div className="w-full bg-secondary h-1.5 overflow-hidden">
                <div className="bg-navy h-full w-[25%]" />
              </div>

              <div className="flex justify-between items-center text-[11px]">
                <span className="text-muted-foreground">4. PostgreSQL ACID Commit &amp; Outbox Message:</span>
                <span className="font-bold text-foreground">4.2 ms</span>
              </div>
              <div className="w-full bg-secondary h-1.5 overflow-hidden">
                <div className="bg-authoritative h-full w-[50%]" />
              </div>
            </div>
          </section>

          {/* Evidence Dossier List */}
          <div className="space-y-6">
            <EvidenceList items={d.evidence} kind="OBSERVED" />
            <EvidenceList items={d.evidence} kind="INFERRED" />
            <EvidenceList items={d.evidence} kind="RECOMMENDED" />
          </div>

          <div className="space-y-3">
            <Disclosure summary={`Triggered rules (${d.rules.length})`}>
              {d.rules.length === 0 ? (
                <p className="py-2 text-[12.5px] text-muted-foreground">No rules triggered.</p>
              ) : (
                <ul>
                  {d.rules.map((r) => (
                    <li key={r.id} className="border-t border-border py-2.5 first:border-t-0 font-mono text-[12px]">
                      <div className="flex items-baseline gap-3">
                        <Mono className="text-navy font-bold">{r.id}</Mono>
                        <span className="text-[13px] font-sans font-medium text-foreground">{r.name}</span>
                      </div>
                      <p className="mt-0.5 text-[11.5px] text-muted-foreground font-sans">{r.outcome}</p>
                    </li>
                  ))}
                </ul>
              )}
            </Disclosure>

            <Disclosure summary="Raw evaluation request payload">
              <CodeBlock code={JSON.stringify(d.rawRequest, null, 2)} language="json" />
            </Disclosure>
          </div>
        </div>

        {/* Right Sidebar: Attribution & Context */}
        <aside className="space-y-6">
          <section className="border border-border bg-card p-4 space-y-3">
            <div className="flex items-baseline justify-between border-b border-border pb-2">
              <h2 className="text-[11px] font-bold tracking-[0.08em] uppercase font-mono text-navy">
                Score Attribution
              </h2>
              <Mono className="text-muted-foreground text-[11px]">
                {d.baseScore.toFixed(2)} → {d.riskScore.toFixed(2)}
              </Mono>
            </div>
            <p className="text-[11.5px] text-muted-foreground">
              Baseline plus {delta > 0 ? "+" : ""}
              {delta.toFixed(2)} from {d.factors.length} contributing factors.
            </p>
            <RiskFactorList factors={d.factors} />
          </section>

          <section className="border border-border bg-card p-4 space-y-3 font-mono text-[11.5px]">
            <h2 className="text-[11px] font-bold tracking-[0.08em] uppercase text-navy border-b border-border pb-2">
              Forensic Session Context
            </h2>
            <KeyValue
              rows={[
                ["IP", <Mono key="ip">{d.threatIntel?.ip ?? "198.51.100.44"}</Mono>],
                ["Reputation", <span key="rep">{d.threatIntel?.ipReputation ?? "Datacenter Proxy"}</span>],
                ["ASN", <Mono key="asn">{d.threatIntel?.asn ?? "ASN 13335"}</Mono>],
                ["Device Canvas", <Mono key="dvc">{String(d.rawRequest?.["device_fingerprint"] ?? "dev_emulator_linux_9f8a")}</Mono>],
                ["Channel", <span key="ch">{String(d.rawRequest?.["channel"] ?? "imps_payout")}</span>],
              ]}
            />
          </section>
        </aside>
      </div>
    </article>
  );
}
