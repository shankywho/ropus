import { createFileRoute, Link } from "@tanstack/react-router";
import { useState } from "react";
import {
  DataGrid,
  InspectorRow,
  Page,
  PageHead,
  SectionHead,
  TelemetryStrip,
} from "@/components/ropus/page";
import { Mono } from "@/components/ropus/core";
import { rules, ruleDetails, type RuleRecord } from "@/lib/ropus/platform-fixtures";
import { cn } from "@/lib/utils";

export const Route = createFileRoute("/rules")({
  head: () => ({
    meta: [
      { title: "Rules — ROPUS" },
      {
        name: "description",
        content:
          "Deterministic decisioning rules, their policies, weights, hit volume and precision.",
      },
      { property: "og:title", content: "Rules — ROPUS" },
      {
        property: "og:description",
        content: "Policy rules with weight, state, 24h hit volume and precision.",
      },
      { property: "og:type", content: "website" },
      { name: "twitter:card", content: "summary" },
    ],
  }),
  component: RulesPage,
});

const stateTone: Record<RuleRecord["state"], string> = {
  ENABLED: "bg-approve",
  SHADOW: "bg-warning",
  DISABLED: "bg-border-strong",
};

const actionTone: Record<RuleRecord["action"], string> = {
  BLOCK: "text-block",
  REVIEW: "text-warning",
  CHALLENGE: "text-primary",
  SCORE: "text-muted-foreground",
};

const filters = ["ALL", "ENABLED", "SHADOW", "DISABLED"] as const;

/** State reads as text plus a 4px semantic square — never a pill. */
function StateText({ state }: { state: RuleRecord["state"] }) {
  return (
    <span className="flex items-center gap-1.5 whitespace-nowrap">
      <span aria-hidden className={cn("size-1.5", stateTone[state])} />
      <span
        className={cn(
          "text-[11px] font-semibold tracking-[0.06em]",
          state === "DISABLED" ? "text-muted-foreground" : undefined,
        )}
      >
        {state}
      </span>
    </span>
  );
}

function RulesPage() {
  const [filter, setFilter] = useState<(typeof filters)[number]>("ALL");
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const rows = rules.filter((r) => filter === "ALL" || r.state === filter);
  const selected = rows.find((r) => r.id === selectedId) ?? null;
  const detail = selected ? ruleDetails[selected.id] : undefined;

  return (
    <Page>
      <PageHead
        title="Rules"
        subtitle="Deterministic conditions evaluated before the model on every request. Weights are additive contributions to the risk score."
      />

      <TelemetryStrip
        items={[
          { label: "Rules", value: String(rules.length), sub: "in this tenant" },
          {
            label: "Enabled",
            value: String(rules.filter((r) => r.state === "ENABLED").length),
            sub: "scoring live traffic",
          },
          {
            label: "Shadow",
            value: String(rules.filter((r) => r.state === "SHADOW").length),
            sub: "evaluated, not applied",
          },
          {
            label: "Hits",
            value: rules.reduce((s, r) => s + r.hits24h, 0).toLocaleString(),
            sub: "last 24 hours",
          },
          {
            label: "Hard stops",
            value: String(rules.filter((r) => r.action === "BLOCK").length),
            sub: "action = BLOCK",
          },
        ]}
      />

      <div
        className={cn(
          "mt-4 grid gap-6",
          selected ? "xl:grid-cols-[minmax(0,1fr)_330px] xl:gap-8" : "grid-cols-1",
        )}
      >
        <section className="min-w-0">
          <SectionHead
            title="Rule set"
            meta={`${rows.length} shown`}
            right={
              <div className="flex items-center gap-3.5">
                {filters.map((f) => (
                  <button
                    key={f}
                    type="button"
                    onClick={() => setFilter(f)}
                    aria-pressed={filter === f}
                    className={cn(
                      "border-b-[1.5px] pb-[3px] text-[11px] font-semibold tracking-[0.06em] uppercase",
                      filter === f
                        ? "border-b-foreground text-foreground"
                        : "border-b-transparent text-muted-foreground hover:text-foreground",
                    )}
                  >
                    {f}
                  </button>
                ))}
              </div>
            }
          />
          {/* Policy and Updated live in the inspector, so they yield column width when it is open. */}
          <div className="overflow-x-auto">
            <div className={selected ? "min-w-[720px]" : "min-w-[920px]"}>
              <DataGrid
                onSelect={(id) => setSelectedId((cur) => (cur === id ? null : id))}
                selectedId={selectedId}
                rowClassName="py-2.5"
                columns={[
                  { key: "id", label: "Rule" },
                  { key: "name", label: "Condition" },
                  ...(selected ? [] : [{ key: "policy", label: "Policy" }]),
                  { key: "scope", label: "Scope" },
                  { key: "action", label: "Action" },
                  { key: "weight", label: "Weight", align: "right" as const },
                  { key: "hits", label: "Hits 24h", align: "right" as const },
                  { key: "precision", label: "Precision", align: "right" as const },
                  { key: "state", label: "State" },
                  ...(selected ? [] : [{ key: "updated", label: "Updated" }]),
                ]}
                rows={rows.map((r) => ({
                  id: r.id,
                  cells: [
                    <Mono className="text-[12px] font-semibold whitespace-nowrap">{r.id}</Mono>,
                    <span className="text-muted-foreground">{r.name}</span>,
                    ...(selected
                      ? []
                      : [<Mono className="text-[11.5px] text-muted-foreground">{r.policy}</Mono>]),
                    <span className="text-[11px] tracking-[0.06em] text-muted-foreground uppercase">
                      {r.scope}
                    </span>,
                    <span
                      className={cn(
                        "text-[11px] font-semibold tracking-[0.06em]",
                        actionTone[r.action],
                      )}
                    >
                      {r.action}
                    </span>,
                    <Mono className="font-semibold">+{r.weight.toFixed(2)}</Mono>,
                    <Mono className="text-muted-foreground">{r.hits24h.toLocaleString()}</Mono>,
                    <Mono className={r.precision < 0.4 ? "text-warning" : undefined}>
                      {(r.precision * 100).toFixed(0)}%
                    </Mono>,
                    <StateText state={r.state} />,
                    ...(selected
                      ? []
                      : [
                          <span className="whitespace-nowrap text-muted-foreground">
                            <Mono className="text-[11.5px] text-muted-foreground">
                              {r.updatedAt}
                            </Mono>{" "}
                            · {r.updatedBy}
                          </span>,
                        ]),
                  ],
                }))}
              />
            </div>
          </div>
          <p className="mt-2.5 border-l-2 border-l-border-strong pl-2.5 text-[11px] text-muted-foreground">
            Low-precision rules are retained deliberately: they contribute score rather than a
            verdict, and are only decisive in combination.
          </p>
        </section>

        {selected && (
          <aside
            aria-label="Rule inspector"
            className="min-w-0 xl:border-l xl:border-border xl:pl-6"
          >
            <SectionHead
              title="Rule"
              right={
                <button
                  type="button"
                  onClick={() => setSelectedId(null)}
                  className="text-[11px] font-semibold tracking-[0.06em] text-muted-foreground uppercase hover:text-foreground"
                >
                  Close
                </button>
              }
            />
            <div className="pt-2">
              <Mono className="text-[13.5px] font-bold">{selected.id}</Mono>
              <p className="mt-0.5 text-[12.5px] text-muted-foreground">{selected.name}</p>
            </div>

            <div className="mt-3">
              <InspectorRow label="Status">
                <StateText state={selected.state} />
              </InspectorRow>
              <InspectorRow label="Policy">
                <Mono className="text-[12px] text-muted-foreground">{selected.policy}</Mono>
              </InspectorRow>
              <InspectorRow label="Scope">
                <span className="text-[11px] tracking-[0.06em] uppercase">{selected.scope}</span>
              </InspectorRow>
              <InspectorRow label="Action">
                <span
                  className={cn(
                    "text-[11px] font-semibold tracking-[0.06em]",
                    actionTone[selected.action],
                  )}
                >
                  {selected.action}
                </span>
              </InspectorRow>
              <InspectorRow label="Weight">
                <Mono className="font-semibold">+{selected.weight.toFixed(2)}</Mono>
              </InspectorRow>
              <InspectorRow label="Hits">
                <Mono>{selected.hits24h.toLocaleString()} / 24h</Mono>
              </InspectorRow>
              <InspectorRow label="Precision">
                <Mono>{(selected.precision * 100).toFixed(0)}%</Mono>
              </InspectorRow>
            </div>

            {detail && (
              <>
                <h3 className="mt-5 border-b border-border pb-1.5 text-[10.5px] font-bold tracking-[0.08em] uppercase">
                  Example logic
                </h3>
                <dl className="mt-2 font-mono text-[11.5px] leading-[1.55]">
                  <div className="flex gap-2">
                    <dt className="w-10 shrink-0 text-muted-foreground">IF</dt>
                    <dd>{detail.logic.if}</dd>
                  </div>
                  {detail.logic.and && (
                    <div className="mt-1 flex gap-2">
                      <dt className="w-10 shrink-0 text-muted-foreground">AND</dt>
                      <dd>{detail.logic.and}</dd>
                    </div>
                  )}
                  <div className="mt-1 flex gap-2">
                    <dt className="w-10 shrink-0 text-muted-foreground">THEN</dt>
                    <dd>{detail.logic.then}</dd>
                  </div>
                </dl>

                <h3 className="mt-5 border-b border-border pb-1.5 text-[10.5px] font-bold tracking-[0.08em] uppercase">
                  Recent evaluation activity
                </h3>
                <div className="mt-1">
                  <InspectorRow label="Hits">
                    <Mono>{selected.hits24h.toLocaleString()}</Mono>
                  </InspectorRow>
                  <InspectorRow label="Avg contribution">
                    <Mono>+{detail.avgContribution.toFixed(2)}</Mono>
                  </InspectorRow>
                  <InspectorRow label="Last triggered">
                    <Mono className="text-[11.5px]">{detail.lastTriggeredAt ?? "never"}</Mono>
                  </InspectorRow>
                  <InspectorRow label="Triggered by">
                    <Mono className="text-[11.5px]">{detail.lastTriggeredBy ?? "—"}</Mono>
                  </InspectorRow>
                </div>
                {detail.lastDecisionId && (
                  <Link
                    to="/decisions/$decisionId"
                    params={{ decisionId: detail.lastDecisionId }}
                    className="mt-2.5 inline-block text-[11.5px] text-primary hover:underline"
                  >
                    View decision →
                  </Link>
                )}
              </>
            )}
          </aside>
        )}
      </div>
    </Page>
  );
}
