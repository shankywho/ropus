import { createFileRoute, Link } from "@tanstack/react-router";
import { useState, useEffect, useRef } from "react";
import { Mono, RiskScore, VerdictBadge, PriorityTag, CaseStatusTag } from "@/components/ropus/core";
import { Page } from "@/components/ropus/page";
import { blockedDecision, baselineDecision } from "@/lib/ropus/fixtures";
import { fraudGraph } from "@/lib/ropus/graph-fixture";
import { demoControls, demoStages, useDemoState, type DemoStageInfo } from "@/lib/ropus/demo-store";
import { evaluateRisk, type LiveEvaluationResponse } from "@/lib/ropus/api";
import { cn } from "@/lib/utils";

export const Route = createFileRoute("/demo")({
  head: () => ({
    meta: [
      { title: "Demo Walkthrough — ROPUS Risk Control Plane" },
      {
        name: "description",
        content:
          "Deterministic 7-stage incident replay: baseline transaction, attack ingress, rules engine triggers, ML calibration, fraud graph traversal, AI investigation, and analyst block.",
      },
      { property: "og:title", content: "Demo Walkthrough — ROPUS Risk Control Plane" },
      {
        property: "og:description",
        content: "Deterministic 7-stage walkthrough of the ROPUS risk decisioning pipeline.",
      },
      { property: "og:type", content: "website" },
      { name: "twitter:card", content: "summary" },
    ],
  }),
  component: DemoPage,
});

/** Smoothly tweens the score progression for clear presentation feedback */
function useTweenedScore(target: number, durationMs = 400) {
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

const speeds: Array<[number, string]> = [
  [3500, "0.5×"],
  [2200, "1×"],
  [1200, "2×"],
];

function DemoPage() {
  const { stage, playing, intervalMs } = useDemoState();
  const current: DemoStageInfo = demoStages[stage] || demoStages[0];
  const tweenedScore = useTweenedScore(current.simulatedScore);
  const [confirmedBlock, setConfirmedBlock] = useState(false);
  const [liveEvalResult, setLiveEvalResult] = useState<LiveEvaluationResponse | null>(null);
  const [liveLoading, setLiveLoading] = useState(false);

  const handleLiveEvaluate = async () => {
    setLiveLoading(true);
    try {
      const res = await evaluateRisk({
        transaction_id: "txn_order_88419",
        customer_id: "cus_4471029",
        amount: 1450000,
        currency: "USD",
        ip_address: "198.51.100.44",
        device_id: "9f8a84b12c",
      });
      setLiveEvalResult(res);
    } catch (e) {
      console.error("Live evaluate failed", e);
    } finally {
      setLiveLoading(false);
    }
  };

  // Reset confirmation state when leaving stage 7 or resetting
  useEffect(() => {
    if (stage !== 6) {
      setConfirmedBlock(false);
    }
  }, [stage]);

  const nextStage = () => {
    if (stage < demoStages.length - 1) {
      demoControls.next();
    } else {
      demoControls.reset();
    }
  };

  return (
    <Page>
      {/* ------------------------------------------------ Header & Controls */}
      <div className="flex flex-wrap items-start justify-between gap-4 border-b border-border pb-4">
        <div>
          <div className="flex items-center gap-2">
            <span className="rounded bg-primary/10 px-2 py-0.5 font-mono text-[11px] font-bold text-primary">
              DEMO REPLAY
            </span>
            <span className="font-mono text-[11px] text-muted-foreground">
              Deterministic 7-Stage Incident Trace
            </span>
          </div>
          <h1 className="mt-1 text-[20px] font-bold tracking-tight text-foreground">
            Attack Evaluation Lifecycle
          </h1>
          <p className="mt-0.5 text-[13px] text-muted-foreground">
            Normal baseline → high-risk wire → rules triggers → ML calibration → graph traversal →
            AI evidence → analyst freeze.
          </p>
        </div>

        {/* Playback Controls */}
        <div className="flex flex-wrap items-center gap-2">
          <button
            type="button"
            onClick={() => (playing ? demoControls.pause() : demoControls.play())}
            className={cn(
              "rounded px-3 py-1.5 font-mono text-[12px] font-bold transition-colors",
              playing
                ? "bg-warning text-foreground"
                : "bg-primary text-primary-foreground hover:bg-primary/90",
            )}
          >
            {playing ? "⏸ Pause Replay" : stage === 0 ? "▶ Start 7-Stage Replay" : "▶ Resume"}
          </button>
          <button
            type="button"
            onClick={() => demoControls.back()}
            disabled={stage === 0}
            className="rounded border border-border bg-surface px-2.5 py-1.5 font-mono text-[12px] font-medium text-foreground hover:bg-accent disabled:opacity-40"
          >
            ← Back
          </button>
          <button
            type="button"
            onClick={() => demoControls.next()}
            disabled={stage === demoStages.length - 1}
            className="rounded border border-border bg-surface px-2.5 py-1.5 font-mono text-[12px] font-medium text-foreground hover:bg-accent disabled:opacity-40"
          >
            Step →
          </button>
          <button
            type="button"
            onClick={() => demoControls.reset()}
            className="rounded border border-border bg-surface px-2.5 py-1.5 font-mono text-[12px] text-muted-foreground hover:text-foreground"
          >
            Reset
          </button>

          <div className="ml-2 flex items-center gap-1 border-l border-border pl-2">
            {speeds.map(([ms, label]) => (
              <button
                key={label}
                type="button"
                onClick={() => demoControls.setSpeed(ms)}
                aria-pressed={intervalMs === ms}
                className={cn(
                  "px-1.5 py-1 font-mono text-[11px] rounded",
                  intervalMs === ms
                    ? "bg-foreground text-background font-bold"
                    : "text-muted-foreground hover:text-foreground",
                )}
              >
                {label}
              </button>
            ))}
          </div>
        </div>
      </div>

      {/* ------------------------------------------------ 7-Stage Timeline Navigation Pills */}
      <nav
        aria-label="Demo Stages"
        className="mt-4 grid grid-cols-2 gap-2 sm:grid-cols-4 lg:grid-cols-7"
      >
        {demoStages.map((s, idx) => {
          const isActive = stage === idx;
          const isPassed = stage > idx;
          return (
            <button
              key={s.id}
              type="button"
              onClick={() => demoControls.goTo(idx)}
              className={cn(
                "flex flex-col items-start border p-2 text-left transition-all rounded",
                isActive
                  ? "border-primary bg-primary/5 shadow-xs ring-1 ring-primary"
                  : isPassed
                    ? "border-border bg-accent/40 text-foreground"
                    : "border-border/60 bg-surface text-muted-foreground hover:border-border",
              )}
            >
              <div className="flex w-full items-center justify-between text-[10px] font-mono">
                <span
                  className={cn("font-bold", isActive ? "text-primary" : "text-muted-foreground")}
                >
                  STAGE {s.stageNumber}
                </span>
                {isPassed && <span className="text-approve font-bold">✓</span>}
              </div>
              <div className="mt-1 truncate text-[11.5px] font-semibold">
                {s.label.split(" ")[0]} {s.label.split(" ")[1] ?? ""}
              </div>
            </button>
          );
        })}
      </nav>

      {/* ------------------------------------------------ Current Stage Core Presentation Area */}
      <div className="mt-6 grid gap-6 lg:grid-cols-[380px_minmax(0,1fr)]">
        {/* Left Column: Stage Summary, Score Gauge & Factor Attribution */}
        <section className="flex flex-col justify-between rounded border border-border bg-surface p-5">
          <div className="space-y-4">
            <div>
              <div className="flex items-center justify-between">
                <span className="font-mono text-[11px] font-bold tracking-wider text-primary uppercase">
                  Stage {current.stageNumber} of 7
                </span>
                <VerdictBadge verdict={current.verdict} size="sm" />
              </div>
              <h2 className="mt-1 text-[17px] font-bold text-foreground">{current.label}</h2>
              <p className="mt-1 text-[13px] leading-relaxed text-muted-foreground">
                {current.shortExplanation}
              </p>
            </div>

            {/* Score Display */}
            <div className="border-t border-border pt-4">
              <div className="text-[10.5px] font-semibold tracking-wider text-muted-foreground uppercase">
                Risk Score Invariant
              </div>
              <div className="mt-2 flex items-baseline justify-between">
                <RiskScore value={tweenedScore} size="hero" showBand />
                <div className="text-right font-mono text-[11px] text-muted-foreground">
                  <div>Base: 0.02</div>
                  <div>Attribution: +{(tweenedScore - 0.02).toFixed(2)}</div>
                </div>
              </div>
            </div>

            {/* Additive Factors Summary */}
            <div className="border-t border-border pt-4">
              <div className="text-[10.5px] font-semibold tracking-wider text-muted-foreground uppercase">
                Active Factor Contributions
              </div>
              <ul className="mt-2 space-y-1.5 text-[12px]">
                {stage === 0 && (
                  <li className="flex items-center justify-between font-mono text-muted-foreground">
                    <span>Base Account Trust</span>
                    <span className="text-approve font-bold">+0.02</span>
                  </li>
                )}
                {stage >= 1 && (
                  <li className="flex items-center justify-between font-mono">
                    <span className="text-foreground">Unrecognised Device & Geo</span>
                    <span className="text-block font-bold">+0.26</span>
                  </li>
                )}
                {stage >= 2 && (
                  <li className="flex items-center justify-between font-mono">
                    <span className="text-foreground">Velocity Surge + Impossible Speed</span>
                    <span className="text-block font-bold">+0.37</span>
                  </li>
                )}
                {stage >= 3 && (
                  <li className="flex items-center justify-between font-mono">
                    <span className="text-foreground">XGBoost Posterior Calibration</span>
                    <span className="text-block font-bold">+0.23</span>
                  </li>
                )}
                {stage >= 4 && (
                  <li className="flex items-center justify-between font-mono">
                    <span className="text-foreground">Syndicate Cluster Centrality</span>
                    <span className="text-block font-bold">+0.06</span>
                  </li>
                )}
                {stage >= 5 && (
                  <li className="flex items-center justify-between font-mono">
                    <span className="text-foreground">Agent Tripartite Evidence</span>
                    <span className="text-block font-bold">+0.02</span>
                  </li>
                )}
              </ul>
            </div>
          </div>

          {/* Next Button */}
          <div className="mt-6 border-t border-border pt-4">
            <button
              type="button"
              onClick={nextStage}
              className="w-full rounded bg-foreground py-2.5 text-center font-mono text-[12.5px] font-bold text-background transition-opacity hover:opacity-90"
            >
              {stage < 6 ? `Next: Stage ${stage + 2} →` : "Restart Demo Replay ↺"}
            </button>
          </div>
        </section>

        {/* Right Column: Deep Technical Telemetry & Surface Details */}
        <section className="rounded border border-border bg-surface p-5">
          <div className="flex items-center justify-between border-b border-border pb-3">
            <h3 className="font-mono text-[12px] font-bold tracking-wider text-muted-foreground uppercase">
              Engine Telemetry & State Artifacts
            </h3>
            <span className="font-mono text-[11px] text-muted-foreground">
              Latency: {stage < 2 ? "1.1ms" : stage < 4 ? "3.2ms" : "38.4ms"}
            </span>
          </div>

          <div className="mt-4">
            {/* STAGE 1: Baseline */}
            {stage === 0 && (
              <div className="space-y-4">
                <div className="rounded border border-border bg-accent/30 p-3">
                  <div className="font-mono text-[11px] font-bold text-primary">
                    CUSTOMER PROFILE: cus_4471029
                  </div>
                  <div className="mt-1 text-[13px] text-foreground">
                    Tenure: 14 months · Clean history (0 disputes) · Average ticket: $48.20
                  </div>
                </div>

                <div className="grid grid-cols-2 gap-3 text-[12.5px]">
                  <div className="border border-border p-3 rounded">
                    <div className="label-xs">Trusted Device</div>
                    <Mono className="mt-1 font-bold">dev_safari_ios_01</Mono>
                    <div className="mt-1 text-muted-foreground">
                      Mobile Safari · iOS 17.4 · Canvas 4a12
                    </div>
                  </div>
                  <div className="border border-border p-3 rounded">
                    <div className="label-xs">Home Subnet</div>
                    <Mono className="mt-1 font-bold">Singtel Mobile (Singapore)</Mono>
                    <div className="mt-1 text-muted-foreground">ASN 4657 · 103.82°E, 1.35°N</div>
                  </div>
                </div>

                <div className="border border-border p-3 rounded">
                  <div className="label-xs">Recent History (Last 24h)</div>
                  <div className="mt-2 space-y-1 font-mono text-[11.5px] text-muted-foreground">
                    <div className="flex justify-between">
                      <span>txn_99120 · FairPrice Supermarket</span>
                      <span className="text-foreground">$38.50 USD · APPROVE (0.01)</span>
                    </div>
                    <div className="flex justify-between">
                      <span>txn_99118 · Grab Taxi Singapore</span>
                      <span className="text-foreground">$14.20 USD · APPROVE (0.02)</span>
                    </div>
                  </div>
                </div>
              </div>
            )}

            {/* STAGE 2: Attack Ingress */}
            {stage === 1 && (
              <div className="space-y-4">
                <div className="rounded border border-block/40 bg-block/5 p-3">
                  <div className="font-mono text-[11px] font-bold text-block">
                    INBOUND API INGRESS: POST /v1/risk-evaluations
                  </div>
                  <div className="mt-1 text-[13px] text-foreground font-semibold">
                    Amount: $14,500.00 USD (Wire Transfer to Beneficiary PA-77120)
                  </div>
                </div>

                <div className="border border-border p-3 font-mono text-[11.5px] bg-accent/20 rounded">
                  <div className="flex items-center justify-between text-muted-foreground font-semibold">
                    <span>// Inbound JSON Payload Contract</span>
                    <button
                      type="button"
                      onClick={handleLiveEvaluate}
                      disabled={liveLoading}
                      className="rounded bg-primary px-2.5 py-1 text-[11px] font-bold text-white transition-opacity hover:opacity-90 disabled:opacity-50"
                    >
                      {liveLoading
                        ? "Evaluating on Go Backend..."
                        : "⚡ Execute Live Backend Evaluation"}
                    </button>
                  </div>
                  <pre className="mt-2 text-foreground overflow-x-auto">
                    {`{
  "transaction_id": "txn_order_88419",
  "customer_id": "cus_4471029",
  "amount": 1450000,
  "currency": "USD",
  "ip_address": "198.51.100.44",
  "device_fingerprint": "9f8a84b12c"
}`}
                  </pre>
                </div>

                {liveEvalResult && (
                  <div className="rounded border border-approve/50 bg-approve/10 p-3 font-mono text-[12px] text-foreground">
                    <div className="flex items-center justify-between text-approve font-bold">
                      <span className="flex items-center gap-1.5">
                        <span className="size-2 rounded-full bg-approve" />
                        LIVE GO ORCHESTRATOR RESPONSE
                      </span>
                      <span className="text-muted-foreground font-normal">
                        Latency: {liveEvalResult.latency_ms}ms
                      </span>
                    </div>
                    <div className="mt-2 grid grid-cols-2 gap-2 text-[11.5px]">
                      <div>
                        Decision:{" "}
                        <span className="font-bold text-primary">{liveEvalResult.decision_id}</span>
                      </div>
                      <div>
                        Action:{" "}
                        <span className="font-bold text-block">
                          {liveEvalResult.recommended_action}
                        </span>
                      </div>
                    </div>
                    <div className="mt-1 text-[11px] text-muted-foreground">
                      Signals:{" "}
                      {liveEvalResult.reason_codes?.join(" · ") ||
                        "RULE_SIGNAL:HIGH_TRANSACTION_AMOUNT"}
                    </div>
                  </div>
                )}
              </div>
            )}

            {/* STAGE 3: Rules Engine */}
            {stage === 2 && (
              <div className="space-y-4">
                <div className="rounded border border-warning/40 bg-warning/5 p-3">
                  <div className="font-mono text-[11px] font-bold text-warning">
                    DETERMINISTIC AST RULES EVALUATED (0.4ms)
                  </div>
                  <div className="mt-1 text-[13px] text-foreground">
                    3 hard policy condition nodes triggered short-circuit actions.
                  </div>
                </div>

                <div className="space-y-2">
                  <div className="border border-border p-2.5 rounded flex items-start justify-between">
                    <div>
                      <div className="font-mono text-[12px] font-bold text-block">
                        RULE_IMPOSSIBLE_TRAVEL_SPEED
                      </div>
                      <div className="text-[12px] text-muted-foreground">
                        Singapore to Cyprus in 12 min (51,350 km/h &gt; 900 km/h threshold)
                      </div>
                    </div>
                    <span className="font-mono text-[11px] font-bold text-block">+0.21</span>
                  </div>

                  <div className="border border-border p-2.5 rounded flex items-start justify-between">
                    <div>
                      <div className="font-mono text-[12px] font-bold text-block">
                        RULE_VELOCITY_SURGE_1H
                      </div>
                      <div className="text-[12px] text-muted-foreground">
                        Amount velocity +412% over 24h average window
                      </div>
                    </div>
                    <span className="font-mono text-[11px] font-bold text-block">+0.22</span>
                  </div>

                  <div className="border border-border p-2.5 rounded flex items-start justify-between">
                    <div>
                      <div className="font-mono text-[12px] font-bold text-block">
                        RULE_DATACENTER_PROXY_ASN
                      </div>
                      <div className="text-[12px] text-muted-foreground">
                        IP 198.51.100.44 belongs to hosting datacenter ASN 13335
                      </div>
                    </div>
                    <span className="font-mono text-[11px] font-bold text-block">+0.18</span>
                  </div>
                </div>
              </div>
            )}

            {/* STAGE 4: ML Inference */}
            {stage === 3 && (
              <div className="space-y-4">
                <div className="rounded border border-primary/40 bg-primary/5 p-3">
                  <div className="font-mono text-[11px] font-bold text-primary">
                    25-FEATURE ONNX XGBOOST INFERENCE (2.1ms)
                  </div>
                  <div className="mt-1 text-[13px] text-foreground font-semibold">
                    Calibrated Posterior: P(fraud | x) = 0.9418 · Expected Loss = $13,656.10
                  </div>
                </div>

                <div className="grid grid-cols-3 gap-2 font-mono text-[11.5px]">
                  <div className="border border-border p-2 rounded">
                    <div className="text-[10px] text-muted-foreground uppercase">
                      Raw Model Output
                    </div>
                    <div className="mt-1 font-bold text-[14px]">0.8712</div>
                  </div>
                  <div className="border border-border p-2 rounded">
                    <div className="text-[10px] text-muted-foreground uppercase">
                      Beta Calibrated
                    </div>
                    <div className="mt-1 font-bold text-[14px] text-block">0.9418</div>
                  </div>
                  <div className="border border-border p-2 rounded">
                    <div className="text-[10px] text-muted-foreground uppercase">Expected Loss</div>
                    <div className="mt-1 font-bold text-[14px] text-block">$13,656.10</div>
                  </div>
                </div>

                <div className="border border-border p-3 font-mono text-[11px] bg-accent/20 rounded">
                  <div className="text-muted-foreground font-semibold">
                    // Continuous Feature Tensor Snippet
                  </div>
                  <div className="mt-1 grid grid-cols-2 gap-x-4 gap-y-1 text-foreground">
                    <div>geo_velocity_kmh: 51350.0</div>
                    <div>device_entropy: 0.92</div>
                    <div>is_vpn_proxy: 1.0</div>
                    <div>card_testing_ratio: 0.00</div>
                    <div>canvas_drift: 0.88</div>
                    <div>user_tenure_months: 14.0</div>
                  </div>
                </div>
              </div>
            )}

            {/* STAGE 5: Fraud Graph */}
            {stage === 4 && (
              <div className="space-y-4">
                <div className="rounded border border-block/40 bg-block/5 p-3">
                  <div className="font-mono text-[11px] font-bold text-block">
                    3-HOP BFS GRAPH DISCOVERY (1.2ms)
                  </div>
                  <div className="mt-1 text-[13px] text-foreground font-semibold">
                    Syndicate Ring Detected: Device canvas linked to 14 synthetic accounts and
                    confirmed chargebacks.
                  </div>
                </div>

                <div className="border border-border p-3 rounded font-mono text-[11.5px] space-y-2">
                  <div className="flex items-center gap-2">
                    <span className="size-2 rounded-full bg-block" />
                    <span className="font-bold">cus_4471029</span>
                    <span className="text-muted-foreground">→ USED_DEVICE →</span>
                    <span className="font-bold text-block">dev_emulator_linux_9f8a</span>
                  </div>
                  <div className="flex items-center gap-2 pl-4">
                    <span className="size-2 rounded-full bg-block" />
                    <span className="text-muted-foreground">→ SHARED_CANVAS →</span>
                    <span className="font-bold text-block">14 Synthetic Mule Accounts</span>
                  </div>
                  <div className="flex items-center gap-2 pl-8">
                    <span className="size-2 rounded-full bg-block" />
                    <span className="text-muted-foreground">→ PAYOUT_LINK →</span>
                    <span className="font-bold text-block">PA-77120 (Confirmed Mule Depot)</span>
                  </div>
                </div>

                <div className="text-right">
                  <Link to="/graph" className="font-mono text-[12px] text-primary hover:underline">
                    Open Full Interactive Fraud Graph →
                  </Link>
                </div>
              </div>
            )}

            {/* STAGE 6: AI Investigator */}
            {stage === 5 && (
              <div className="space-y-4">
                <div className="rounded border border-primary/40 bg-primary/5 p-3">
                  <div className="font-mono text-[11px] font-bold text-primary">
                    AGENT COUNCIL TRIPARTITE DOSSIER GENERATED
                  </div>
                  <div className="mt-1 text-[13px] text-foreground font-semibold">
                    Case CASE-88419 opened at Priority P0 (Critical) with 15m SLA.
                  </div>
                </div>

                <div className="space-y-2 text-[12.5px]">
                  <div className="border-l-2 border-primary bg-accent/20 p-2.5">
                    <div className="font-bold text-primary font-mono text-[11px]">
                      1. OBSERVED FACTS
                    </div>
                    <p className="mt-0.5 text-muted-foreground">
                      $14,500 outbound wire to Cyprus. Egress from Limassol datacenter IP
                      198.51.100.44. Linux emulator canvas 9f8a.
                    </p>
                  </div>
                  <div className="border-l-2 border-warning bg-warning/5 p-2.5">
                    <div className="font-bold text-warning font-mono text-[11px]">
                      2. INFERRED PATTERNS
                    </div>
                    <p className="mt-0.5 text-muted-foreground">
                      Account Takeover (ATO) probability 96%. Device canvas matches active Limassol
                      mule ring (degree centrality = 16).
                    </p>
                  </div>
                  <div className="border-l-2 border-block bg-block/5 p-2.5">
                    <div className="font-bold text-block font-mono text-[11px]">
                      3. RECOMMENDED ACTION
                    </div>
                    <p className="mt-0.5 text-foreground font-semibold">
                      Block transaction, freeze customer session, and add beneficiary PA-77120 to
                      global syndicate blacklist.
                    </p>
                  </div>
                </div>
              </div>
            )}

            {/* STAGE 7: Human Analyst Action */}
            {stage === 6 && (
              <div className="space-y-4">
                <div
                  className={cn(
                    "rounded border p-3 transition-colors",
                    confirmedBlock ? "border-approve bg-approve/10" : "border-block bg-block/5",
                  )}
                >
                  <div className="font-mono text-[11px] font-bold text-foreground">
                    {confirmedBlock
                      ? "ACTION CONFIRMED & AUDITED"
                      : "HUMAN GOVERNANCE REVIEW REQUIRED"}
                  </div>
                  <div className="mt-1 text-[13px] text-foreground font-semibold">
                    {confirmedBlock
                      ? "Transaction BLOCKED. Hash chain entry appended to audit ledger. Ground truth sent to ML retraining pipeline."
                      : "Reviewing CASE-88419 (Analyst: m.okafor). Human authorization required to confirm block."}
                  </div>
                </div>

                {!confirmedBlock ? (
                  <div className="space-y-3">
                    <button
                      type="button"
                      onClick={() => setConfirmedBlock(true)}
                      className="w-full rounded bg-block py-3 font-mono text-[13px] font-bold text-block-foreground shadow-sm hover:bg-block/90 active:scale-[0.99]"
                    >
                      CONFIRM BLOCK &amp; FREEZE ACCOUNT
                    </button>
                    <div className="text-center text-[11px] text-muted-foreground">
                      Authorizing analyst:{" "}
                      <Mono className="font-bold text-foreground">
                        m.okafor (Senior Fraud Lead)
                      </Mono>
                    </div>
                  </div>
                ) : (
                  <div className="space-y-2 border border-border p-3 bg-accent/20 rounded font-mono text-[11.5px]">
                    <div className="text-approve font-bold flex items-center gap-1.5">
                      <span>✓</span> Case Outcome: CONFIRMED_FRAUD
                    </div>
                    <div className="text-muted-foreground">
                      Audit Hash:{" "}
                      <Mono className="text-foreground">sha256:4f8a1e9c7a2b53018e6...</Mono>
                    </div>
                    <div className="text-muted-foreground">
                      ML Retraining Event:{" "}
                      <Mono className="text-foreground">evt_retrain_ground_truth_88419</Mono>
                    </div>
                  </div>
                )}
              </div>
            )}
          </div>
        </section>
      </div>
    </Page>
  );
}
