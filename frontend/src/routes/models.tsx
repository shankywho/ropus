import { createFileRoute } from "@tanstack/react-router";
import { useState } from "react";
import { DataGrid, InspectorRow, Page, PageHead, SectionHead, TelemetryStrip } from "@/components/ropus/page";
import { Mono } from "@/components/ropus/core";
import { models, modelDetails, type ModelRecord } from "@/lib/ropus/platform-fixtures";
import { blockedDecision } from "@/lib/ropus/fixtures";
import { cn } from "@/lib/utils";

export const Route = createFileRoute("/models")({
  head: () => ({
    meta: [
      { title: "Models — ROPUS" },
      { name: "description", content: "Deployed risk models, shadow candidates, drift, latency and feature attribution." },
      { property: "og:title", content: "Models — ROPUS" },
      { property: "og:description", content: "Model registry with stage, AUC, drift and serving latency." },
      { property: "og:type", content: "website" },
      { name: "twitter:card", content: "summary" },
    ],
  }),
  component: ModelsPage,
});

const stageTone: Record<ModelRecord["stage"], string> = {
  PRODUCTION: "bg-approve",
  SHADOW: "bg-warning",
  RETIRED: "bg-border-strong",
};

const key = (m: ModelRecord) => `${m.id}@${m.version}`;

function StageText({ stage }: { stage: ModelRecord["stage"] }) {
  return (
    <span className="flex items-center gap-1.5 whitespace-nowrap">
      <span aria-hidden className={cn("size-1.5", stageTone[stage])} />
      <span
        className={cn(
          "text-[11px] font-semibold tracking-[0.06em]",
          stage === "RETIRED" ? "text-muted-foreground" : undefined,
        )}
      >
        {stage}
      </span>
    </span>
  );
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
        title="Models"
        subtitle="Model registry for this tenant. Production models score live traffic; shadow models are mirrored and never affect a verdict."
      />

      <TelemetryStrip
        items={[
          { label: "Production models", value: String(production.length), sub: "serving verdicts" },
          { label: "Shadow", value: String(models.filter((m) => m.stage === "SHADOW").length), sub: "mirrored traffic" },
          { label: "Serving p99", value: "19.6 ms", sub: "mdl_wire_risk" },
          {
            label: "Max drift (PSI)",
            value: Math.max(...production.map((m) => m.driftPsi)).toFixed(2),
            sub: "mdl_card_cnp — review",
            tone: "text-warning",
          },
          { label: "Last promotion", value: "2026-08-12", sub: "v4.3.0-rc2 to shadow" },
        ]}
      />

      <div className="mt-4 grid gap-6 xl:grid-cols-[minmax(0,1fr)_330px] xl:gap-8">
        <section className="min-w-0">
          <SectionHead title="Model registry" meta={`${models.length} versions`} />
          {/* Trained date moves into the inspector when it is open. */}
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
              { key: "auc", label: "AUC", align: "right" },
              { key: "psi", label: "Drift PSI", align: "right" },
              { key: "p99", label: "p99", align: "right" },
              { key: "features", label: "Features", align: "right" },
              { key: "share", label: "Traffic", align: "right" as const },
              ...(selected ? [] : [{ key: "trained", label: "Trained" }]),
            ]}
            rows={models.map((m) => ({
              id: key(m),
              cells: [
                <Mono className="text-[12px] font-semibold whitespace-nowrap">{m.id}</Mono>,
                <span className="text-muted-foreground">{m.name}</span>,
                <Mono>{m.version}</Mono>,
                <StageText stage={m.stage} />,
                <Mono className="font-semibold">{m.auc.toFixed(3)}</Mono>,
                <Mono className={m.driftPsi > 0.1 ? "text-warning" : undefined}>{m.driftPsi.toFixed(2)}</Mono>,
                <Mono className="text-muted-foreground">{m.p99Ms.toFixed(1)}ms</Mono>,
                <Mono className="text-muted-foreground">{m.features}</Mono>,
                <span className="flex items-center justify-end gap-2">
                  <span aria-hidden className="hidden h-[2px] w-10 bg-neutral-surface sm:block">
                    <span className="block h-full bg-foreground/45" style={{ width: `${m.callShare * 100}%` }} />
                  </span>
                  <Mono className={m.callShare === 0 ? "text-muted-foreground" : undefined}>
                    {(m.callShare * 100).toFixed(0)}%
                  </Mono>
                </span>,
                ...(selected ? [] : [<Mono className="text-muted-foreground">{m.trainedOn}</Mono>]),
              ],
            }))}
          />
          </div>
          </div>
          <p className="mt-2.5 border-l-2 border-l-warning pl-2.5 text-[11px] text-muted-foreground">
            PSI above 0.10 opens an operations review; mdl_card_cnp is scheduled for retraining on 2026-09-01.
          </p>

          {selected && compare && (
            <div className="mt-5">
              <SectionHead
                title="Version comparison"
                meta={`${selected.version} vs ${compare.version}`}
                right={
                  <button
                    type="button"
                    onClick={() => setCompareKey(null)}
                    className="text-[11px] font-semibold tracking-[0.06em] text-muted-foreground uppercase hover:text-foreground"
                  >
                    Close
                  </button>
                }
              />
              <table className="w-full border-collapse text-[12.5px]">
                <thead>
                  <tr className="border-b border-border">
                    <th className="py-1.5 pr-4 text-left text-[10.5px] font-semibold tracking-[0.08em] text-muted-foreground uppercase">
                      Metric
                    </th>
                    <th className="py-1.5 pr-4 text-right font-mono text-[11.5px] font-semibold">{selected.version}</th>
                    <th className="py-1.5 text-right font-mono text-[11.5px] font-semibold">{compare.version}</th>
                  </tr>
                </thead>
                <tbody>
                  {(
                    [
                      ["AUC", selected.auc.toFixed(3), compare.auc.toFixed(3)],
                      ["Drift PSI", selected.driftPsi.toFixed(2), compare.driftPsi.toFixed(2)],
                      ["p99", `${selected.p99Ms.toFixed(1)}ms`, `${compare.p99Ms.toFixed(1)}ms`],
                      ["Traffic", `${(selected.callShare * 100).toFixed(0)}%`, `${(compare.callShare * 100).toFixed(0)}%`],
                      ["Features", String(selected.features), String(compare.features)],
                    ] as const
                  ).map(([label, a, b]) => (
                    <tr key={label} className="border-b border-border last:border-b-0">
                      <td className="py-[7px] pr-4 text-muted-foreground">{label}</td>
                      <td className="py-[7px] pr-4 text-right">
                        <Mono>{a}</Mono>
                      </td>
                      <td className="py-[7px] text-right">
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
            <div className="mb-6">
              <SectionHead
                title="Model"
                right={
                  <button
                    type="button"
                    onClick={() => setSelectedKey(null)}
                    className="text-[11px] font-semibold tracking-[0.06em] text-muted-foreground uppercase hover:text-foreground"
                  >
                    Close
                  </button>
                }
              />
              <div className="pt-2">
                <Mono className="text-[13.5px] font-bold">{selected.id}</Mono>
                <p className="mt-0.5 text-[12.5px] text-muted-foreground">
                  {selected.name}
                  {detail && (
                    <>
                      {" · serves as "}
                      <Mono className="text-[11.5px] text-muted-foreground">{detail.servingAlias}</Mono>
                    </>
                  )}
                </p>
              </div>
              <div className="mt-3">
                <InspectorRow label="Version">
                  <Mono className="font-semibold">{selected.version}</Mono>
                </InspectorRow>
                <InspectorRow label="Stage">
                  <StageText stage={selected.stage} />
                </InspectorRow>
                <InspectorRow label="AUC">
                  <Mono>{selected.auc.toFixed(3)}</Mono>
                </InspectorRow>
                <InspectorRow label="p99">
                  <Mono>{selected.p99Ms.toFixed(1)} ms</Mono>
                </InspectorRow>
                <InspectorRow label="Drift">
                  <Mono className={selected.driftPsi > 0.1 ? "text-warning" : undefined}>
                    {selected.driftPsi.toFixed(2)} PSI
                  </Mono>
                </InspectorRow>
                <InspectorRow label="Features">
                  <Mono>{selected.features}</Mono>
                </InspectorRow>
                <InspectorRow label="Traffic">
                  <Mono>{(selected.callShare * 100).toFixed(0)}%</Mono>
                </InspectorRow>
              </div>

              {detail && (
                <>
                  <h3 className="mt-5 border-b border-border pb-1.5 text-[10.5px] font-bold tracking-[0.08em] uppercase">
                    Model lifecycle
                  </h3>
                  <div className="mt-1">
                    <InspectorRow label="Training">
                      <Mono className="text-[11.5px]">{selected.trainedOn}</Mono>
                    </InspectorRow>
                    <InspectorRow label="Promotion">
                      <Mono className="text-[11.5px]">{detail.promotedOn ?? "not promoted"}</Mono>
                    </InspectorRow>
                    <InspectorRow label="Previous version">
                      <Mono className="text-[11.5px]">{detail.previousVersion ?? "—"}</Mono>
                    </InspectorRow>
                    <InspectorRow label="Shadow candidate">
                      <Mono className="text-[11.5px]">{detail.shadowCandidate ?? "—"}</Mono>
                    </InspectorRow>
                  </div>
                  {detail.shadowCandidate && (
                    <button
                      type="button"
                      onClick={() => setCompareKey(`${selected.id}@${detail.shadowCandidate}`)}
                      className="mt-2.5 text-[11.5px] text-primary hover:underline"
                    >
                      Compare with {detail.shadowCandidate} →
                    </button>
                  )}
                </>
              )}
            </div>
          )}

          <SectionHead
            title="Feature attribution"
            meta={`${blockedDecision.inference.model} ${blockedDecision.inference.version}`}
          />
          <p className="mt-2 text-[11.5px] text-muted-foreground">
            Attributions for decision <Mono className="text-foreground">{blockedDecision.decisionId}</Mono>. Model
            output, not observed fact.
          </p>
          <div className="mt-2.5">
            {features.map((f) => (
              <div key={f.name} className="border-t border-border py-[7px] first:border-t-0">
                <div className="flex items-baseline justify-between gap-4">
                  <Mono className="text-[11.5px]">{f.name}</Mono>
                  <Mono className="font-semibold">+{f.contribution.toFixed(2)}</Mono>
                </div>
                <div className="mt-1.5 h-[2px] w-full bg-neutral-surface">
                  <div className="h-full bg-foreground/60" style={{ width: `${(f.contribution / maxFeature) * 100}%` }} />
                </div>
              </div>
            ))}
          </div>
          <div className="mt-3 border-t border-border pt-2 text-[12px]">
            <div className="flex items-baseline justify-between gap-4">
              <span className="text-muted-foreground">Predicted fraud probability</span>
              <Mono className="font-semibold">{blockedDecision.inference.probability.toFixed(4)}</Mono>
            </div>
            <p className="mt-1.5 text-[11px] text-muted-foreground">
              Model probability is one contributor to the final decision score.
            </p>
          </div>
        </aside>
      </div>
    </Page>
  );
}
