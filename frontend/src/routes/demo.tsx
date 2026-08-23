import { createFileRoute, Link } from "@tanstack/react-router";
import {
  EvidenceList,
  Mono,
  RiskFactorList,
  RiskScore,
  VerdictBadge,
} from "@/components/ropus/core";
import { Page, PageHead, SectionHead } from "@/components/ropus/page";
import { blockedDecision } from "@/lib/ropus/fixtures";
import { fraudGraph } from "@/lib/ropus/graph-fixture";
import { demoControls, demoStages, useDemoState } from "@/lib/ropus/demo-store";
import { cn } from "@/lib/utils";
import { useEffect, useRef, useState } from "react";

/** Eases a score change so escalations read as movement, not a jump. */
function useTweenedScore(target: number, durationMs = 520) {
  const [value, setValue] = useState(target);
  const from = useRef(target);
  useEffect(() => {
    const start = performance.now();
    const origin = from.current;
    if (origin === target) return;
    let raf = 0;
    const step = (now: number) => {
      const t = Math.min(1, (now - start) / durationMs);
      const eased = 1 - Math.pow(1 - t, 3);
      setValue(origin + (target - origin) * eased);
      if (t < 1) raf = requestAnimationFrame(step);
      else from.current = target;
    };
    raf = requestAnimationFrame(step);
    return () => cancelAnimationFrame(raf);
  }, [target, durationMs]);
  useEffect(() => {
    from.current = value;
  }, [value]);
  return value;
}

export const Route = createFileRoute("/demo")({
  head: () => ({
    meta: [
      { title: "Demo Scenario — ROPUS Risk Control Plane" },
      {
        name: "description",
        content:
          "Deterministic ROPUS demo: account takeover to blocked wire, fraud graph expansion and analyst-confirmed case.",
      },
      { property: "og:title", content: "Demo Scenario — ROPUS Risk Control Plane" },
      {
        property: "og:description",
        content: "Deterministic walkthrough of one attack sequence across decisioning, graph and case surfaces.",
      },
      { property: "og:type", content: "website" },
      { name: "twitter:card", content: "summary" },
    ],
  }),
  component: DemoPage,
});

/** Score is revealed progressively — the same arithmetic as the decision page. */
function scoreAtStage(stage: number) {
  if (stage < 4) return blockedDecision.baseScore;
  const revealed = factorsAtStage(stage);
  const raw = revealed.reduce((sum, f) => sum + f.weight, blockedDecision.baseScore);
  return Math.min(raw, blockedDecision.riskScore);
}

function factorsAtStage(stage: number) {
  if (stage < 5) return [];
  if (stage === 5) return blockedDecision.factors.slice(0, 3);
  if (stage === 6) return blockedDecision.factors.slice(0, 5);
  return blockedDecision.factors;
}

const speeds: Array<[number, string]> = [
  [2200, "0.5×"],
  [1400, "1×"],
  [700, "2×"],
];

function DemoPage() {
  const { stage, playing, intervalMs } = useDemoState();
  const current = demoStages[stage]!;
  const score = useTweenedScore(scoreAtStage(stage));
  const factors = factorsAtStage(stage);
  const decided = stage >= 8;
  const relationships = fraudGraph.relationships.slice(0, stage < 6 ? 0 : stage === 6 ? 3 : undefined);

  return (
    <Page>
      <PageHead
        title="Demo scenario"
        subtitle="Deterministic replay of one attack sequence across every ROPUS surface. Fixed data, fixed ordering, no simulated processing delay."
        actions={
          <div className="flex items-center gap-2">
            <button
              type="button"
              onClick={() => (playing ? demoControls.pause() : demoControls.play())}
              className="border border-primary bg-primary px-3 py-1 text-[12px] font-semibold text-primary-foreground"
            >
              {playing ? "Pause" : stage === 0 ? "Start replay" : "Resume"}
            </button>
            <button
              type="button"
              onClick={() => demoControls.back()}
              disabled={stage === 0}
              className="border border-border-strong px-2.5 py-1 text-[12px] font-medium disabled:opacity-40"
            >
              Back
            </button>
            <button
              type="button"
              onClick={() => demoControls.next()}
              disabled={stage === demoStages.length - 1}
              className="border border-border-strong px-2.5 py-1 text-[12px] font-medium disabled:opacity-40"
            >
              Step
            </button>
            <button
              type="button"
              onClick={() => demoControls.reset()}
              className="border border-border-strong px-2.5 py-1 text-[12px] font-medium"
            >
              Reset
            </button>
            <span className="ml-2 flex items-center gap-2.5">
              {speeds.map(([ms, label]) => (
                <button
                  key={label}
                  type="button"
                  onClick={() => demoControls.setSpeed(ms)}
                  aria-pressed={intervalMs === ms}
                  className={cn(
                    "font-mono text-[11.5px]",
                    intervalMs === ms ? "font-bold text-foreground" : "text-muted-foreground hover:text-foreground",
                  )}
                >
                  {label}
                </button>
              ))}
            </span>
            <Mono className="ml-1 text-muted-foreground">
              {String(stage).padStart(2, "0")}/{demoStages.length - 1}
            </Mono>
          </div>
        }
      />

      <div className="mt-5 grid gap-8 xl:grid-cols-[210px_minmax(0,1fr)_330px] xl:gap-10">
        {/* sequence */}
        <nav aria-label="Scenario stages" className="min-w-0">
          <SectionHead title="Sequence" />
          <ol className="mt-2 border-l border-border">
            {demoStages.map((s) => {
              const done = s.id < stage;
              const active = s.id === stage;
              return (
                <li key={s.id}>
                  <button
                    type="button"
                    onClick={() => demoControls.goTo(s.id)}
                    className="relative block w-full py-[5px] pl-4 text-left"
                  >
                    <span
                      aria-hidden
                      className={cn(
                        "absolute top-[11px] -left-[3.5px] size-[6px]",
                        active ? "bg-primary" : done ? "bg-border-strong" : "bg-border",
                      )}
                    />
                    <span
                      className={cn(
                        "text-[12.5px]",
                        active ? "font-semibold text-foreground" : done ? "text-foreground/70" : "text-muted-foreground",
                      )}
                    >
                      {s.label}
                    </span>
                  </button>
                </li>
              );
            })}
          </ol>
        </nav>

        {/* stage output */}
        <div className="min-w-0">
          <section className="flex flex-wrap items-start justify-between gap-6 border-b border-border pb-5">
            <div className="min-w-0 max-w-[58ch]">
              <h2 className="text-[11px] font-bold tracking-[0.08em] uppercase">{current.label}</h2>
              <p className="mt-1.5 text-[14px] leading-snug">{current.note}</p>
              <dl className="mt-4 flex flex-wrap gap-x-8 gap-y-2">
                <div>
                  <dt className="text-[10.5px] font-semibold tracking-[0.08em] text-muted-foreground uppercase">Customer</dt>
                  <dd className="mt-0.5">
                    <Mono>{blockedDecision.customerId}</Mono>
                  </dd>
                </div>
                <div>
                  <dt className="text-[10.5px] font-semibold tracking-[0.08em] text-muted-foreground uppercase">Transaction</dt>
                  <dd className="mt-0.5">
                    <Mono className={stage >= 4 ? undefined : "text-muted-foreground"}>
                      {stage >= 4 ? blockedDecision.transactionId : "—"}
                    </Mono>
                  </dd>
                </div>
                <div>
                  <dt className="text-[10.5px] font-semibold tracking-[0.08em] text-muted-foreground uppercase">Decision</dt>
                  <dd className="mt-0.5">
                    {decided ? (
                      <Link
                        to="/decisions/$decisionId"
                        params={{ decisionId: blockedDecision.decisionId }}
                        className="font-mono text-[12.5px] text-primary hover:underline"
                      >
                        {blockedDecision.decisionId}
                      </Link>
                    ) : (
                      <Mono className="text-muted-foreground">pending</Mono>
                    )}
                  </dd>
                </div>
                <div>
                  <dt className="text-[10.5px] font-semibold tracking-[0.08em] text-muted-foreground uppercase">Case</dt>
                  <dd className="mt-0.5">
                    {stage >= 9 ? (
                      <Link
                        to="/cases/$caseId"
                        params={{ caseId: blockedDecision.caseId! }}
                        className="font-mono text-[12.5px] text-primary hover:underline"
                      >
                        {blockedDecision.caseId}
                      </Link>
                    ) : (
                      <Mono className="text-muted-foreground">—</Mono>
                    )}
                  </dd>
                </div>
              </dl>
            </div>
            <div className="text-right">
              <RiskScore value={score} size="hero" showBand />
              <div className="mt-2">
                {decided ? (
                  <VerdictBadge verdict={blockedDecision.verdict} size="lg" />
                ) : (
                  <span className="text-[11.5px] text-muted-foreground">verdict pending</span>
                )}
              </div>
            </div>
          </section>

          <section className="mt-6">
            <SectionHead
              title="Signals evaluated"
              meta={factors.length ? `${factors.length} of ${blockedDecision.factors.length} returned` : "none yet"}
            />
            {factors.length === 0 ? (
              <p className="mt-2 text-[12.5px] text-muted-foreground">
                No factors returned yet — the transaction has not reached the decision path.
              </p>
            ) : (
              <div className="mt-1">
                <RiskFactorList factors={factors} />
              </div>
            )}
          </section>

          {decided && (
            <div className="mt-7 space-y-6">
              <EvidenceList items={blockedDecision.evidence} kind="OBSERVED" />
              <EvidenceList items={blockedDecision.evidence} kind="INFERRED" />
              {stage >= 10 && <EvidenceList items={blockedDecision.evidence} kind="RECOMMENDED" />}
            </div>
          )}
        </div>

        {/* graph reveal */}
        <aside className="min-w-0">
          <SectionHead
            title="Connected entities"
            meta={relationships.length ? `${relationships.length} edges` : "not traversed"}
          />
          {relationships.length === 0 ? (
            <p className="mt-2 text-[12.5px] text-muted-foreground">
              The graph is expanded once a payout account enters the decision path.
            </p>
          ) : (
            <>
              <ul className="mt-1">
                {relationships.map((r) => (
                  <li key={`${r.source}-${r.target}`} className="border-b border-border py-2 last:border-b-0">
                    <div className="flex items-baseline justify-between gap-3">
                      <Mono className="text-[12px]">{r.source}</Mono>
                      <span className="font-mono text-[10px] tracking-[0.08em] text-muted-foreground uppercase">
                        {r.label}
                      </span>
                    </div>
                    <Mono className="text-[12px] text-muted-foreground">→ {r.target}</Mono>
                  </li>
                ))}
              </ul>
              <Link to="/graph" className="mt-3 inline-block text-[12px] text-primary hover:underline">
                Open the fraud graph →
              </Link>
            </>
          )}

          <div className="mt-7">
            <SectionHead title="Decision record" />
            <dl className="mt-1 text-[12.5px]">
              {[
                ["Policy", blockedDecision.policy],
                ["Model", `${blockedDecision.model}-${blockedDecision.modelVersion}`],
                ["Latency", decided ? `${blockedDecision.latencyMs.toFixed(1)} ms` : "—"],
                ["Verdict", decided ? blockedDecision.verdict : "pending"],
                ["Case", stage >= 9 ? blockedDecision.caseId! : "—"],
              ].map(([k, v]) => (
                <div key={k} className="flex items-baseline justify-between gap-4 border-b border-border py-1.5 last:border-b-0">
                  <dt className="text-muted-foreground">{k}</dt>
                  <dd>
                    <Mono>{v}</Mono>
                  </dd>
                </div>
              ))}
            </dl>
          </div>
        </aside>
      </div>
    </Page>
  );
}
