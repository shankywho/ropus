import { createFileRoute } from "@tanstack/react-router";
import { useState } from "react";
import {
  DataGrid,
  InspectorRow,
  Page,
  PageHead,
  SectionHead,
  TelemetryStrip,
} from "@/components/ropus/page";
import { Mono, StatusPill } from "@/components/ropus/core";
import { models, modelDetails, type ModelRecord } from "@/lib/ropus/platform-fixtures";
import { blockedDecision } from "@/lib/ropus/fixtures";
import { cn } from "@/lib/utils";

export const Route = createFileRoute("/models")({
  head: () => ({
    meta: [
      { title: "Models — ROPUS" },
      {
        name: "description",
        content: "Deployed risk models, shadow candidates, drift, latency and feature attribution.",
      },
      { property: "og:title", content: "Models — ROPUS" },
      {
        property: "og:description",
        content: "Model registry with stage, AUC, drift and serving latency.",
      },
      { property: "og:type", content: "website" },
      { name: "twitter:card", content: "summary" },
    ],
  }),
  component: ModelsPage,
});

const key = (m: ModelRecord) => `${m.id}@${m.version}`;

function StageText({ stage }: { stage: ModelRecord["stage"] }) {
  if (stage === "PRODUCTION") {
    return <StatusPill tone="authoritative">AUTHORITATIVE</StatusPill>;
  }
  if (stage === "SHADOW") {
    return <StatusPill tone="shadow">SHADOW</StatusPill>;
  }
  return <StatusPill tone="neutral">RETIRED</StatusPill>;
}

function ModelsPage() {
  const [selectedKey, setSelectedKey] = useState<string | null>(null);
  const [compareKey, setCompareKey] = useState<string | null>(null);

  const production = models.filter((m) => m.stage === "PRODUCTION");
  const features = blockedDecision.inference.features;
  const maxFeature = Math.max(...features.map((f) => f.contribution));

  const selected = models.find((m) => key(m) === selectedKey) ?? null;
  const detail = selected ? modelDetails[key(selected)] : undefined;
  const compare = models.find((m) => key(m) === compareKey) ?? null;

  return (
    <Page>
      <PageHead
        title="Model Registry & Governance"
        subtitle="Authoritative production champions score live traffic with 100% decision authority. GraphSAGE shadow candidates are strictly non-enforcing (0% customer authority)."
      />

      <TelemetryStrip
        items={[
          { label: "Production Champions", value: String(production.length), sub: "100% BMR Authority" },
          {
            label: "Shadow Candidates",
            value: String(models.filter((m) => m.stage === "SHADOW").length),
            sub: "GraphSAGE Non-Enforcing",
          },
          { label: "Serving p99", value: "19.6 ms", sub: "production_model_v8_bmr" },
          {
            label: "Max Drift (PSI)",
            value: Math.max(...production.map((m) => m.driftPsi)).toFixed(2),
            sub: "Stable baseline (<0.10)",
            tone: "text-authoritative",
          },
          { label: "Confirmed Cases", value: "0 / 50", sub: "Promotion Gated", tone: "text-shadow-intel" },
        ]}
      />

      <div className="mt-4 grid gap-6 xl:grid-cols-[minmax(0,1fr)_340px] xl:gap-8">
        <section className="min-w-0">
          <SectionHead title="MODEL REGISTRY" meta={`${models.length} registered models`} />
          <div className="overflow-x-auto">
            <div className={selected ? "min-w-[700px]" : "min-w-[880px]"}>
              <DataGrid
                onSelect={(id) => {
                  setSelectedKey((cur) => (cur === id ? null : id));
                  setCompareKey(null);
                }}
                selectedId={selectedKey}
                rowClassName="py-2.5"
                columns={[
                  { key: "id", label: "Model" },
                  { key: "name", label: "Purpose" },
                  { key: "version", label: "Version" },
                  { key: "stage", label: "Stage" },
                  { key: "auc", label: "ROC-AUC (eval)", align: "right" },
                  { key: "prAuc", label: "PR-AUC (eval)", align: "right" },
                  { key: "psi", label: "Drift PSI", align: "right" },
                  { key: "p99", label: "p99 SLA", align: "right" },
                  { key: "features", label: "Features", align: "right" },
                  { key: "share", label: "Authority", align: "right" as const },
                  ...(selected ? [] : [{ key: "trained", label: "Trained" }]),
                ]}
                rows={models.map((m) => ({
                  id: key(m),
                  cells: [
                    <Mono className="text-[12px] font-medium whitespace-nowrap text-foreground">{m.id}</Mono>,
                    <span className="text-muted-foreground">{m.name}</span>,
                    <Mono>{m.version}</Mono>,
                    <StageText stage={m.stage} />,
                    <Mono className="font-medium">{m.auc.toFixed(3)}</Mono>,
                    <Mono className="font-medium text-authoritative">
                      {m.prAuc !== undefined ? m.prAuc.toFixed(4) : "—"}
                    </Mono>,
                    <Mono className={m.driftPsi > 0.1 ? "text-amber-intel" : "text-muted-foreground"}>
                      {m.driftPsi.toFixed(2)}
                    </Mono>,
                    <Mono className="text-muted-foreground">{m.p99Ms.toFixed(1)}ms</Mono>,
                    <Mono className="text-muted-foreground">{m.features}</Mono>,
                    <span className="flex items-center justify-end gap-2">
                      <span aria-hidden className="hidden h-[2px] w-10 bg-secondary sm:block">
                        <span
                          className={cn("block h-full", m.callShare > 0 ? "bg-authoritative" : "bg-muted-foreground/30")}
                          style={{ width: `${m.callShare * 100}%` }}
                        />
                      </span>
                      <Mono className={m.callShare === 0 ? "text-muted-foreground" : "text-authoritative font-medium"}>
                        {(m.callShare * 100).toFixed(0)}%
                      </Mono>
                    </span>,
                    ...(selected
                      ? []
                      : [<Mono className="text-muted-foreground">{m.trainedOn}</Mono>]),
                  ],
                }))}
              />
            </div>
          </div>
          <p className="mt-2.5 border-l-2 border-l-amber-intel pl-2.5 font-sans text-[11px] text-muted-foreground">
            ROC-AUC and PR-AUC reflect evaluation on chronological hold-out splits (IEEE-CIS benchmark).
            GraphSAGE relationship models operate in strictly non-enforcing shadow mode with 0% customer decision authority until 50 confirmed real collusion cases accumulate.
          </p>

          {selected && compare && (
            <div className="mt-5 border border-border bg-card p-4 shadow-2xs">
              <SectionHead
                title="VERSION COMPARISON"
                meta={`${selected.version} vs ${compare.version}`}
                right={
                  <button
                    type="button"
                    onClick={() => setCompareKey(null)}
                    className="font-mono text-[10px] font-medium tracking-[0.06em] text-muted-foreground uppercase hover:text-foreground"
                  >
                    Close
                  </button>
                }
              />
              <table className="w-full border-collapse font-sans text-[11.5px] mt-2">
                <thead>
                  <tr className="border-b border-border bg-muted/40">
                    <th className="py-1.5 pr-4 text-left font-mono text-[9px] font-medium tracking-[0.1em] text-muted-foreground uppercase">
                      Metric
                    </th>
                    <th className="py-1.5 pr-4 text-right font-mono text-[11px] font-medium">
                      {selected.version}
                    </th>
                    <th className="py-1.5 text-right font-mono text-[11px] font-medium">
                      {compare.version}
                    </th>
                  </tr>
                </thead>
                <tbody>
                  {(
                    [
                      ["AUC", selected.auc.toFixed(3), compare.auc.toFixed(3)],
                      ["Drift PSI", selected.driftPsi.toFixed(2), compare.driftPsi.toFixed(2)],
                      ["p99 SLA", `${selected.p99Ms.toFixed(1)}ms`, `${compare.p99Ms.toFixed(1)}ms`],
                      [
                        "Customer Decision Authority",
                        `${(selected.callShare * 100).toFixed(0)}%`,
                        `${(compare.callShare * 100).toFixed(0)}%`,
                      ],
                      ["Features", String(selected.features), String(compare.features)],
                    ] as const
                  ).map(([label, a, b]) => (
                    <tr key={label} className="border-b border-border last:border-b-0">
                      <td className="py-[6px] pr-4 text-muted-foreground">{label}</td>
                      <td className="py-[6px] pr-4 text-right">
                        <Mono>{a}</Mono>
                      </td>
                      <td className="py-[6px] text-right">
                        <Mono>{b}</Mono>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </section>

        <aside className="min-w-0 xl:border-l xl:border-border xl:pl-6">
          {selected && (
            <div className="mb-6 border border-border bg-card p-4 shadow-2xs">
              <SectionHead
                title="MODEL INSPECTOR"
                right={
                  <button
                    type="button"
                    onClick={() => setSelectedKey(null)}
                    className="font-mono text-[10px] font-medium tracking-[0.06em] text-muted-foreground uppercase hover:text-foreground"
                  >
                    Close
                  </button>
                }
              />
              <div className="pt-2">
                <Mono className="text-[13px] font-bold text-foreground">{selected.id}</Mono>
                <p className="mt-0.5 font-sans text-[11.5px] text-muted-foreground">
                  {selected.name}
                  {detail && (
                    <>
                      {" · serves as "}
                      <Mono className="text-[11px] text-muted-foreground">
                        {detail.servingAlias}
                      </Mono>
                    </>
                  )}
                </p>
              </div>
              <div className="mt-3">
                <InspectorRow label="Version">
                  <Mono className="font-medium">{selected.version}</Mono>
                </InspectorRow>
                <InspectorRow label="Stage">
                  <StageText stage={selected.stage} />
                </InspectorRow>
                <InspectorRow label="ROC-AUC">
                  <Mono>{selected.auc.toFixed(3)}</Mono>
                </InspectorRow>
                {selected.prAuc !== undefined && (
                  <InspectorRow label="PR-AUC">
                    <Mono className="font-medium text-authoritative">{selected.prAuc.toFixed(4)}</Mono>
                  </InspectorRow>
                )}
                <InspectorRow label="p99 SLA">
                  <Mono>{selected.p99Ms.toFixed(1)} ms</Mono>
                </InspectorRow>
                <InspectorRow label="Drift">
                  <Mono className={selected.driftPsi > 0.1 ? "text-amber-intel" : undefined}>
                    {selected.driftPsi.toFixed(2)} PSI
                  </Mono>
                </InspectorRow>
                <InspectorRow label="Features">
                  <Mono>{selected.features}</Mono>
                </InspectorRow>
                <InspectorRow label="Decision Authority">
                  <Mono>{(selected.callShare * 100).toFixed(0)}%</Mono>
                </InspectorRow>
              </div>

              <div className="mt-5 border-t border-border pt-3">
                <div className="font-mono text-[9.5px] font-medium tracking-[0.1em] text-muted-foreground uppercase">
                  OFFLINE BENCHMARK EVALUATION
                </div>
                <div className="mt-2">
                  <InspectorRow label="Precision (eval)">
                    <Mono>10.08%</Mono>
                  </InspectorRow>
                  <InspectorRow label="Recall (eval)">
                    <Mono>25.00%</Mono>
                  </InspectorRow>
                  <InspectorRow label="F1 Score">
                    <Mono>0.1436</Mono>
                  </InspectorRow>
                  <InspectorRow label="PR-AUC (eval)">
                    <Mono className="font-medium text-authoritative">
                      {selected.prAuc !== undefined ? selected.prAuc.toFixed(4) : "0.0688"}
                    </Mono>
                  </InspectorRow>
                  <InspectorRow label="ROC-AUC (eval)">
                    <Mono>{selected.auc.toFixed(3)}</Mono>
                  </InspectorRow>
                  <InspectorRow label="False Positive Cost">
                    <Mono>₹40,000 / $500</Mono>
                  </InspectorRow>
                </div>
              </div>

              {detail && (
                <>
                  <div className="mt-5 border-t border-border pt-3">
                    <div className="font-mono text-[9.5px] font-medium tracking-[0.1em] text-muted-foreground uppercase">
                      LIFECYCLE & GOVERNANCE
                    </div>
                    <div className="mt-2">
                      <InspectorRow label="Training">
                        <Mono className="text-[11px]">{selected.trainedOn}</Mono>
                      </InspectorRow>
                      <InspectorRow label="Promotion">
                        <Mono className="text-[11px]">{detail.promotedOn ?? "not promoted"}</Mono>
                      </InspectorRow>
                      <InspectorRow label="Previous Version">
                        <Mono className="text-[11px]">{detail.previousVersion ?? "—"}</Mono>
                      </InspectorRow>
                      <InspectorRow label="Shadow Candidate">
                        <Mono className="text-[11px]">{detail.shadowCandidate ?? "—"}</Mono>
                      </InspectorRow>
                    </div>
                    {detail.shadowCandidate && (
                      <button
                        type="button"
                        onClick={() => setCompareKey(`${selected.id}@${detail.shadowCandidate}`)}
                        className="mt-2.5 font-sans text-[11px] text-navy font-semibold hover:underline"
                      >
                        Compare with {detail.shadowCandidate} →
                      </button>
                    )}
                  </div>
                </>
              )}
            </div>
          )}

          <div className="border border-border bg-card p-4 shadow-2xs">
            <SectionHead
              title="FEATURE ATTRIBUTION"
              meta={`${blockedDecision.inference.model} ${blockedDecision.inference.version}`}
            />
            <p className="mt-1.5 font-sans text-[11.5px] text-muted-foreground">
              Local feature attributions for decision{" "}
              <Mono className="text-foreground">{blockedDecision.decisionId}</Mono>.
            </p>
            <div className="mt-2.5">
              {features.map((f) => (
                <div key={f.name} className="border-t border-border py-[6px] first:border-t-0">
                  <div className="flex items-baseline justify-between gap-4">
                    <Mono className="text-[11px]">{f.name}</Mono>
                    <Mono className="font-medium text-foreground">+{f.contribution.toFixed(2)}</Mono>
                  </div>
                  <div className="mt-1.5 h-[2px] w-full bg-secondary">
                    <div
                      className="h-full bg-navy"
                      style={{ width: `${(f.contribution / maxFeature) * 100}%` }}
                    />
                  </div>
                </div>
              ))}
            </div>
            <div className="mt-3 border-t border-border pt-2.5 text-[11.5px]">
              <div className="flex items-baseline justify-between gap-4">
                <span className="font-sans text-muted-foreground">Predicted Fraud Probability</span>
                <Mono className="font-medium text-destructive">
                  {blockedDecision.inference.probability.toFixed(4)}
                </Mono>
              </div>
              <p className="mt-1.5 font-sans text-[10.5px] text-muted-foreground">
                Calibrated ML probability is combined with pre-rules, post-rules, and GraphSAGE shadow indicators.
              </p>
            </div>
          </div>
        </aside>
      </div>
    </Page>
  );
}
