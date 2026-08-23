import { createFileRoute, Link } from "@tanstack/react-router";
import { Disclosure, CodeBlock, KeyValue } from "@/components/ropus/primitives";
import { Page } from "@/components/ropus/page";
import {
  DecisionSummary,
  DemoTag,
  EvidenceList,
  Metric,
  Mono,
  RiskFactorList,
  RiskScore,
  VerdictBadge,
} from "@/components/ropus/core";
import { useSuspenseQuery } from "@tanstack/react-query";
import { decisionQuery } from "@/lib/ropus/api";
import type { RiskDecision } from "@/lib/ropus/contracts";

export const Route = createFileRoute("/decisions/$decisionId")({
  head: () => ({
    meta: [
      { title: "Risk Decision — ROPUS" },
      {
        name: "description",
        content:
          "Score attribution, contributing risk factors and observed / inferred / recommended explainability for a single ROPUS risk decision.",
      },
      { property: "og:title", content: "Risk Decision — ROPUS" },
      {
        property: "og:description",
        content: "Score attribution and explainability for a single risk decision.",
      },
      { property: "og:type", content: "website" },
      { name: "twitter:card", content: "summary" },
    ],
  }),
  loader: ({ context, params }) => context.queryClient.ensureQueryData(decisionQuery(params.decisionId)),
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

  return (
    <article>
      {/* ---------------------------------------------------------- heading */}
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
              ["Evaluated", new Date(d.evaluatedAt).toISOString().replace("T", " ").slice(0, 19) + "Z"],
            ].map(([k, v]) => (
              <div key={k}>
                <dt className="text-[10.5px] font-semibold tracking-[0.08em] text-muted-foreground uppercase">{k}</dt>
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
                <dd className="font-mono text-foreground tabular">{(d.confidence * 100).toFixed(0)}%</dd>
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

      {/* ------------------------------------------------------ two columns */}
      <div className="grid gap-x-12 gap-y-8 pt-6 lg:grid-cols-[minmax(0,1fr)_380px]">

        {/* left: the story */}
        <div className="space-y-10">
          <section>
            <h2 className="text-[11px] font-bold tracking-[0.08em] uppercase">
              Why this decision
            </h2>
            <p className="mt-2 max-w-[62ch] text-[13.5px] leading-relaxed text-foreground/90">
              {d.evidence.find((e) => e.kind === "INFERRED")?.text ??
                "No elevated pattern detected; the transaction matches the customer baseline."}
            </p>
          </section>

          <section>
            <h2 className="text-[11px] font-bold tracking-[0.08em] uppercase">Decision record</h2>
            <div className="mt-3 max-w-[520px]">
              <DecisionSummary decision={d} amount={money(d.amount, d.currency)} />
            </div>
          </section>

          <div className="space-y-8">
            <EvidenceList items={d.evidence} kind="OBSERVED" />
            <EvidenceList items={d.evidence} kind="INFERRED" />
            <EvidenceList items={d.evidence} kind="RECOMMENDED" />
          </div>

          <div className="space-y-3">
            <Disclosure summary={`Model inference — ${d.inference.model} ${d.inference.version}`}>
              <div className="pt-1">
                <p className="text-[12.5px] text-muted-foreground">
                  Predicted fraud probability{" "}
                  <Mono className="text-foreground">{d.inference.probability.toFixed(4)}</Mono>. Feature
                  contributions are model attributions, not observed facts.
                </p>
                <KeyValue
                  rows={d.inference.features.map((f) => [
                    <span key={f.name} className="font-mono text-[12px]">{f.name}</span>,
                    <Mono key={`${f.name}-v`}>+{f.contribution.toFixed(2)}</Mono>,
                  ])}
                />
              </div>
            </Disclosure>

            <Disclosure summary={`Triggered rules (${d.rules.length})`}>
              {d.rules.length === 0 ? (
                <p className="py-2 text-[12.5px] text-muted-foreground">No rules triggered.</p>
              ) : (
                <ul>
                  {d.rules.map((r) => (
                    <li key={r.id} className="border-t border-border py-2.5 first:border-t-0">
                      <div className="flex items-baseline gap-3">
                        <Mono className="text-muted-foreground">{r.id}</Mono>
                        <span className="text-[13px] font-medium">{r.name}</span>
                      </div>
                      <p className="mt-0.5 text-[12px] text-muted-foreground">{r.outcome}</p>
                    </li>
                  ))}
                </ul>
              )}
            </Disclosure>

            <Disclosure summary="Raw evaluation request">
              <CodeBlock code={JSON.stringify(d.rawRequest, null, 2)} language="json" />
            </Disclosure>
          </div>
        </div>

        {/* right: attribution + context */}
        <aside className="space-y-9">
          <section>
            <div className="flex items-baseline justify-between">
              <h2 className="text-[11px] font-bold tracking-[0.08em] uppercase">Score attribution</h2>
              <Mono className="text-muted-foreground">
                {d.baseScore.toFixed(2)} → {d.riskScore.toFixed(2)}
              </Mono>
            </div>
            <p className="mt-1.5 text-[11.5px] text-muted-foreground">
              Baseline plus {delta > 0 ? "+" : ""}
              {delta.toFixed(2)} from {d.factors.length} contributing factor
              {d.factors.length === 1 ? "" : "s"}.
            </p>
            <div className="mt-3">
              <RiskFactorList factors={d.factors} />
            </div>
            <dl className="mt-4 border-t border-border pt-3 font-mono text-[11.5px] text-muted-foreground tabular">
              {(() => {
                const sum = d.factors.reduce((t, f) => t + f.weight, 0);
                const raw = d.baseScore + sum;
                return (
                  <>
                    <div className="flex justify-between py-0.5">
                      <dt>base score</dt>
                      <dd>{d.baseScore.toFixed(2)}</dd>
                    </div>
                    <div className="flex justify-between py-0.5">
                      <dt>+ contributions</dt>
                      <dd>+{sum.toFixed(2)}</dd>
                    </div>
                    <div className="flex justify-between py-0.5">
                      <dt>= raw score</dt>
                      <dd>{raw.toFixed(2)}</dd>
                    </div>
                    {raw > 1 && (
                      <div className="flex justify-between py-0.5">
                        <dt>capped at</dt>
                        <dd>1.00</dd>
                      </div>
                    )}
                    <div className="flex justify-between border-t border-border py-1 text-foreground">
                      <dt>final decision score</dt>
                      <dd>{d.riskScore.toFixed(2)}</dd>
                    </div>
                  </>
                );
              })()}
            </dl>
          </section>

          <section className="border-t border-border pt-6">
            <h2 className="text-[11px] font-bold tracking-[0.08em] uppercase">Session context</h2>
            <KeyValue
              rows={[
                ["IP", <Mono key="ip">{d.threatIntel.ip}</Mono>],
                ["Network", <span key="asn" className="text-[12.5px]">{d.threatIntel.asn}</span>],
                [
                  "IP reputation",
                  <span
                    key="rep"
                    className={
                      d.threatIntel.ipReputation === "clean"
                        ? "text-[12.5px] text-approve"
                        : "text-[12.5px] text-block"
                    }
                  >
                    {d.threatIntel.ipReputation}
                  </span>,
                ],
                [
                  "Proxy / VPN / Tor",
                  <Mono key="pvt">
                    {[d.threatIntel.proxy, d.threatIntel.vpn, d.threatIntel.tor]
                      .map((v) => (v ? "yes" : "no"))
                      .join(" · ")}
                  </Mono>,
                ],
                ["Device", <span key="dev" className="text-[12.5px]">{d.threatIntel.deviceNovelty}</span>],
              ]}
            />
          </section>

          <section className="border-t border-border pt-6">
            <h2 className="text-[11px] font-bold tracking-[0.08em] uppercase">Delivery</h2>
            <div className="mt-3 grid grid-cols-2 gap-6">
              <Metric label="Webhook status" value={String(d.webhook.status)} sub={`attempt ${d.webhook.attempt}`} />
              <Metric label="Evaluation latency" value={`${d.latencyMs.toFixed(1)}`} sub="milliseconds" />
            </div>
            <p className="mt-3 truncate font-mono text-[11px] text-muted-foreground">{d.webhook.endpoint}</p>
          </section>

          {d.caseId && (
            <section className="border-t border-border pt-6">
              <h2 className="text-[11px] font-bold tracking-[0.08em] uppercase">Linked case</h2>
              <Link
                to="/cases/$caseId"
                params={{ caseId: d.caseId }}
                className="mt-2 inline-block font-mono text-[12.5px] text-primary hover:underline"
              >
                {d.caseId}
              </Link>
            </section>
          )}
        </aside>
      </div>
    </article>
  );
}
