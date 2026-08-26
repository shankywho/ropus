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
import { Mono, StatusPill } from "@/components/ropus/core";
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

const actionTone: Record<RuleRecord["action"], string> = {
  BLOCK: "text-blocked font-semibold",
  REVIEW: "text-amber-intel font-semibold",
  CHALLENGE: "text-local font-semibold",
  SCORE: "text-muted-foreground",
};

const filters = ["ALL", "ENABLED", "SHADOW", "DISABLED"] as const;

function StateText({ state }: { state: RuleRecord["state"] }) {
  if (state === "ENABLED") {
    return <StatusPill tone="authoritative">ENABLED</StatusPill>;
  }
  if (state === "SHADOW") {
    return <StatusPill tone="shadow">SHADOW</StatusPill>;
  }
  return <StatusPill tone="neutral">DISABLED</StatusPill>;
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
        title="Deterministic Rules Engine"
        subtitle="Deterministic JSON-AST conditions evaluated before the ML model on every request. Evaluated in <0.5ms with strict Maker-Checker dual control."
      />

      <TelemetryStrip
        items={[
          { label: "Rules", value: String(rules.length), sub: "in this tenant" },
          {
            label: "Enabled",
            value: String(rules.filter((r) => r.state === "ENABLED").length),
            sub: "scoring live traffic",
            tone: "text-authoritative",
          },
          {
            label: "Shadow",
            value: String(rules.filter((r) => r.state === "SHADOW").length),
            sub: "evaluated, non-blocking",
            tone: "text-amber-intel",
          },
          {
            label: "Hits (24h)",
            value: rules.reduce((s, r) => s + r.hits24h, 0).toLocaleString(),
            sub: "active triggers",
          },
          {
            label: "Hard Blocks",
            value: String(rules.filter((r) => r.action === "BLOCK").length),
            sub: "instant decline",
            tone: "text-blocked",
          },
        ]}
      />

      <div
        className={cn(
          "mt-4 grid gap-6",
          selected ? "xl:grid-cols-[minmax(0,1fr)_340px] xl:gap-8" : "grid-cols-1",
        )}
      >
        <section className="min-w-0">
          <SectionHead
            title="RULE SET"
            meta={`${rows.length} rules matching filter`}
            right={
              <div className="flex items-center gap-3">
                {filters.map((f) => (
                  <button
                    key={f}
                    type="button"
                    onClick={() => setFilter(f)}
                    aria-pressed={filter === f}
                    className={cn(
                      "border-b-2 pb-1 font-mono text-[10px] font-medium tracking-[0.08em] uppercase transition-colors",
                      filter === f
                        ? "border-b-navy text-foreground font-bold"
                        : "border-b-transparent text-muted-foreground hover:text-foreground",
                    )}
                  >
                    {f}
                  </button>
                ))}
              </div>
            }
          />
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
                  { key: "hits", label: "Hits (24h)", align: "right" as const },
                  { key: "precision", label: "Precision", align: "right" as const },
                  { key: "state", label: "State" },
                  ...(selected ? [] : [{ key: "updated", label: "Updated" }]),
                ]}
                rows={rows.map((r) => ({
                  id: r.id,
                  cells: [
                    <Mono className="text-[12px] font-medium whitespace-nowrap text-foreground">{r.id}</Mono>,
                    <span className="text-muted-foreground">{r.name}</span>,
                    ...(selected
                      ? []
                      : [<Mono className="text-[11px] text-muted-foreground">{r.policy}</Mono>]),
                    <span className="font-mono text-[10px] tracking-[0.06em] text-muted-foreground uppercase">
                      {r.scope}
                    </span>,
                    <span
                      className={cn(
                        "font-mono text-[10.5px] tracking-[0.06em] uppercase",
                        actionTone[r.action],
                      )}
                    >
                      {r.action}
                    </span>,
                    <Mono className="font-medium text-foreground">+{r.weight.toFixed(2)}</Mono>,
                    <Mono className="text-muted-foreground">{r.hits24h.toLocaleString()}</Mono>,
                    <Mono className={r.precision < 0.4 ? "text-amber-intel" : undefined}>
                      {(r.precision * 100).toFixed(0)}%
                    </Mono>,
                    <StateText state={r.state} />,
                    ...(selected
                      ? []
                      : [
                          <span className="whitespace-nowrap text-muted-foreground">
                            <Mono className="text-[11px] text-muted-foreground">
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
          <p className="mt-2.5 border-l-2 border-l-border-strong pl-2.5 font-sans text-[11px] text-muted-foreground">
            Low-precision heuristic rules are retained deliberately: they contribute score increments rather than a hard verdict, and become decisive when combined with ML calibration.
          </p>
        </section>

        {selected && (
          <aside
            aria-label="Rule inspector"
            className="min-w-0 xl:border-l xl:border-border xl:pl-6"
          >
            <div className="border border-border bg-card p-4 shadow-2xs">
              <SectionHead
                title="RULE INSPECTOR"
                right={
                  <button
                    type="button"
                    onClick={() => setSelectedId(null)}
                    className="font-mono text-[10px] font-medium tracking-[0.06em] text-muted-foreground uppercase hover:text-foreground"
                  >
                    Close
                  </button>
                }
              />
              <div className="pt-2">
                <Mono className="text-[13px] font-bold text-foreground">{selected.id}</Mono>
                <p className="mt-0.5 font-sans text-[11.5px] text-muted-foreground">{selected.name}</p>
              </div>

              <div className="mt-3">
                <InspectorRow label="Status">
                  <StateText state={selected.state} />
                </InspectorRow>
                <InspectorRow label="Policy">
                  <Mono className="text-[11.5px] text-muted-foreground">{selected.policy}</Mono>
                </InspectorRow>
                <InspectorRow label="Scope">
                  <span className="font-mono text-[10px] tracking-[0.06em] uppercase">{selected.scope}</span>
                </InspectorRow>
                <InspectorRow label="Action">
                  <span
                    className={cn(
                      "font-mono text-[10.5px] tracking-[0.06em] uppercase",
                      actionTone[selected.action],
                    )}
                  >
                    {selected.action}
                  </span>
                </InspectorRow>
                <InspectorRow label="Weight">
                  <Mono className="font-medium text-foreground">+{selected.weight.toFixed(2)}</Mono>
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
                  <div className="mt-5 border-t border-border pt-3">
                    <div className="font-mono text-[9.5px] font-medium tracking-[0.1em] text-muted-foreground uppercase">
                      JSON-AST EVALUATION LOGIC
                    </div>
                    <dl className="mt-2 font-mono text-[11px] leading-[1.55] bg-secondary/50 p-2.5 border border-border">
                      <div className="flex gap-2">
                        <dt className="w-10 shrink-0 text-muted-foreground font-semibold">IF</dt>
                        <dd className="text-foreground">{detail.logic.if}</dd>
                      </div>
                      {detail.logic.and && (
                        <div className="mt-1 flex gap-2">
                          <dt className="w-10 shrink-0 text-muted-foreground font-semibold">AND</dt>
                          <dd className="text-foreground">{detail.logic.and}</dd>
                        </div>
                      )}
                      <div className="mt-1 flex gap-2">
                        <dt className="w-10 shrink-0 text-muted-foreground font-semibold">THEN</dt>
                        <dd className="text-foreground">{detail.logic.then}</dd>
                      </div>
                    </dl>
                  </div>

                  <div className="mt-5 border-t border-border pt-3">
                    <div className="font-mono text-[9.5px] font-medium tracking-[0.1em] text-muted-foreground uppercase">
                      RECENT EVALUATION ACTIVITY
                    </div>
                    <div className="mt-2">
                      <InspectorRow label="Hits">
                        <Mono>{selected.hits24h.toLocaleString()}</Mono>
                      </InspectorRow>
                      <InspectorRow label="Avg Contribution">
                        <Mono>+{detail.avgContribution.toFixed(2)}</Mono>
                      </InspectorRow>
                      <InspectorRow label="Last Triggered">
                        <Mono className="text-[11px]">{detail.lastTriggeredAt ?? "never"}</Mono>
                      </InspectorRow>
                      <InspectorRow label="Triggered By">
                        <Mono className="text-[11px]">{detail.lastTriggeredBy ?? "—"}</Mono>
                      </InspectorRow>
                    </div>
                    {detail.lastDecisionId && (
                      <Link
                        to="/decisions/$decisionId"
                        params={{ decisionId: detail.lastDecisionId }}
                        className="mt-2.5 inline-block font-sans text-[11px] text-navy font-semibold hover:underline"
                      >
                        View Decision Audit →
                      </Link>
                    )}
                  </div>
                </>
              )}
            </div>
          </aside>
        )}
      </div>
    </Page>
  );
}
