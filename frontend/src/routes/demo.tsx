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
  const [failureScenario, setFailureScenario] = useState<
    "NONE" | "ML_TIMEOUT" | "REDIS_DOWN" | "KAFKA_DOWN"
  >("NONE");

  const handleLiveEvaluate = async () => {
    setLiveLoading(true);
    try {
      const res = await evaluateRisk({
        transaction_id: "txn_order_88419",
        customer_id: "cus_4471029",
        amount: 145000000,
        currency: "INR",
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
                <div className="rounded border border-border bg-accent/30 p-3.5">
                  <div className="flex items-center justify-between font-mono text-[11px]">
                    <span className="font-bold text-primary">CUSTOMER PROFILE: cus_4471029</span>
                    <span className="text-approve font-bold">STATUS: TRUSTED ACTIVE</span>
                  </div>
                  <div className="mt-1 text-[13px] text-foreground">
                    Tenure: 14 months · Clean history (0 disputes) · 90-day moving average ticket:
                    ₹480.00 INR
                  </div>
                </div>

                <div className="grid grid-cols-2 gap-3 text-[12.5px]">
                  <div className="border border-border p-3 rounded bg-surface">
                    <div className="text-[10.5px] font-semibold text-muted-foreground uppercase">
                      Trusted Hardware Canvas
                    </div>
                    <Mono className="mt-1 font-bold text-foreground">dev_safari_ios_01</Mono>
                    <div className="mt-1 text-[12px] text-muted-foreground">
                      Mobile Safari · iOS 17.4 · Canvas 4a12 · Entropy 0.24
                    </div>
                  </div>
                  <div className="border border-border p-3 rounded bg-surface">
                    <div className="text-[10.5px] font-semibold text-muted-foreground uppercase">
                      Primary Network Subnet
                    </div>
                    <Mono className="mt-1 font-bold text-foreground">
                      Jio Fiber (Bengaluru, KA)
                    </Mono>
                    <div className="mt-1 text-[12px] text-muted-foreground">
                      ASN 55836 · 12.97°N, 77.59°E · Residential
                    </div>
                  </div>
                </div>

                <div className="border border-border p-3 rounded bg-surface">
                  <div className="text-[10.5px] font-semibold text-muted-foreground uppercase">
                    Recent Normal Transactions (Last 24h)
                  </div>
                  <div className="mt-2 space-y-1.5 font-mono text-[11.5px]">
                    <div className="flex justify-between text-muted-foreground">
                      <span>txn_99120 · Blinkit Quick Commerce (Bengaluru)</span>
                      <span className="text-foreground font-semibold">
                        ₹385.00 INR · APPROVE (0.01)
                      </span>
                    </div>
                    <div className="flex justify-between text-muted-foreground">
                      <span>txn_99118 · Namma Yatri Rides (Bengaluru)</span>
                      <span className="text-foreground font-semibold">
                        ₹142.00 INR · APPROVE (0.02)
                      </span>
                    </div>
                  </div>
                </div>

                <div className="rounded border border-primary/30 bg-primary/5 p-3 text-[12.5px] text-muted-foreground">
                  <span className="font-bold text-foreground">Baseline Invariant:</span> Normal
                  organic behavior is established by low amount deviation, residential IP ASN,
                  consistent hardware entropy, and zero graph connectivity to known chargeback
                  syndicates.
                </div>
              </div>
            )}

            {/* STAGE 2: Suspicious Ingress & Risk Signals */}
            {stage === 1 && (
              <div className="space-y-4">
                <div className="rounded border border-block/40 bg-block/5 p-3.5">
                  <div className="flex items-center justify-between">
                    <span className="font-mono text-[11px] font-bold text-block uppercase">
                      INBOUND API INGRESS: POST /v1/risk-evaluations
                    </span>
                    <span className="font-mono text-[10px] font-bold text-block">
                      HIGH-RISK IMPS PAYOUT
                    </span>
                  </div>
                  <div className="mt-1 text-[14px] text-foreground font-semibold">
                    Amount: ₹14,50,000.00 INR (Outbound IMPS to Beneficiary PA-77120)
                  </div>
                  <div className="mt-0.5 text-[12px] text-muted-foreground">
                    Origin: Limassol proxy · IP: 198.51.100.44 · Device: dev_emulator_linux_9f8a
                  </div>
                </div>

                {/* Signals & Features Extracted */}
                <div className="space-y-2">
                  <div className="text-[11px] font-semibold tracking-wider text-muted-foreground uppercase">
                    Extracted Risk Signals &amp; Features (Context: 1.42ms)
                  </div>
                  <div className="grid grid-cols-2 gap-2 text-[11.5px] font-mono">
                    <div className="rounded border border-border bg-surface p-2">
                      <span className="text-muted-foreground">Geographic Jump:</span>{" "}
                      <span className="font-bold text-block">7,250 km (12 min)</span>
                    </div>
                    <div className="rounded border border-border bg-surface p-2">
                      <span className="text-muted-foreground">Implied Travel Speed:</span>{" "}
                      <span className="font-bold text-block">36,250 km/h (&gt; 900)</span>
                    </div>
                    <div className="rounded border border-border bg-surface p-2">
                      <span className="text-muted-foreground">Device Novelty / Entropy:</span>{" "}
                      <span className="font-bold text-block">0.92 (Emulator VM)</span>
                    </div>
                    <div className="rounded border border-border bg-surface p-2">
                      <span className="text-muted-foreground">Amount Deviation:</span>{" "}
                      <span className="font-bold text-block">+302,000% over avg</span>
                    </div>
                    <div className="rounded border border-border bg-surface p-2">
                      <span className="text-muted-foreground">ASN Reputation:</span>{" "}
                      <span className="font-bold text-block">ASN 13335 (Datacenter)</span>
                    </div>
                    <div className="rounded border border-border bg-surface p-2">
                      <span className="text-muted-foreground">Beneficiary Novelty:</span>{" "}
                      <span className="font-bold text-block">PA-77120 (Unseen)</span>
                    </div>
                  </div>
                </div>

                <div className="border border-border p-3 font-mono text-[11.5px] bg-accent/20 rounded">
                  <div className="flex items-center justify-between text-muted-foreground font-semibold">
                    <span>// Inbound Request JSON</span>
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
                  <pre className="mt-2 text-foreground overflow-x-auto text-[11px]">
                    {`{
  "transaction_id": "txn_order_88419",
  "customer_id": "cus_4471029",
  "amount": 145000000,
  "currency": "INR",
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
                        Decision ID:{" "}
                        <span className="font-bold text-primary">{liveEvalResult.decision_id}</span>
                      </div>
                      <div>
                        Verdict:{" "}
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
                <div className="rounded border border-warning/40 bg-warning/5 p-3.5">
                  <div className="flex items-center justify-between">
                    <span className="font-mono text-[11px] font-bold text-warning uppercase">
                      DETERMINISTIC AST RULES EVALUATED (0.4ms)
                    </span>
                    <span className="font-mono text-[10px] font-bold text-warning">
                      4 RULES TRIGGERED
                    </span>
                  </div>
                  <div className="mt-1 text-[13px] text-foreground">
                    Policy rules evaluate first to enforce deterministic compliance controls before
                    statistical models.
                  </div>
                </div>

                <div className="space-y-2">
                  <div className="border border-border p-3 rounded bg-surface flex items-start justify-between">
                    <div>
                      <div className="font-mono text-[12px] font-bold text-block">
                        RULE_IMPOSSIBLE_TRAVEL_SPEED
                      </div>
                      <div className="text-[12px] text-muted-foreground mt-0.5">
                        Origin: Bengaluru, KA → Limassol proxy | Distance: 7,250 km in 12m (36,250
                        km/h &gt; 900 km/h ceiling)
                      </div>
                    </div>
                    <span className="font-mono text-[11px] font-bold text-block">+0.21</span>
                  </div>

                  <div className="border border-border p-3 rounded bg-surface flex items-start justify-between">
                    <div>
                      <div className="font-mono text-[12px] font-bold text-block">
                        RULE_VELOCITY_SURGE_1H
                      </div>
                      <div className="text-[12px] text-muted-foreground mt-0.5">
                        1-hour expenditure surge (+412% over 24-hour moving average spend profile)
                      </div>
                    </div>
                    <span className="font-mono text-[11px] font-bold text-block">+0.22</span>
                  </div>

                  <div className="border border-border p-3 rounded bg-surface flex items-start justify-between">
                    <div>
                      <div className="font-mono text-[12px] font-bold text-block">
                        RULE_DATACENTER_PROXY_ASN
                      </div>
                      <div className="text-[12px] text-muted-foreground mt-0.5">
                        Source IP 198.51.100.44 matches bulletproof hosting ASN 13335 (Anonymization
                        network)
                      </div>
                    </div>
                    <span className="font-mono text-[11px] font-bold text-block">+0.18</span>
                  </div>

                  <div className="border border-border p-3 rounded bg-surface flex items-start justify-between">
                    <div>
                      <div className="font-mono text-[12px] font-bold text-block">
                        RULE_NEW_BENEFICIARY_HIGH_AMOUNT
                      </div>
                      <div className="text-[12px] text-muted-foreground mt-0.5">
                        Unverified beneficiary PA-77120 receiving high-value transfer (&gt;
                        ₹2,00,000)
                      </div>
                    </div>
                    <span className="font-mono text-[11px] font-bold text-block">+0.12</span>
                  </div>
                </div>
              </div>
            )}

            {/* STAGE 4: ML Inference & Evaluation Metrics */}
            {stage === 3 && (
              <div className="space-y-4">
                <div className="rounded border border-primary/40 bg-primary/5 p-3.5">
                  <div className="flex items-center justify-between">
                    <span className="font-mono text-[11px] font-bold text-primary uppercase">
                      LIVE INFERENCE: 25-FEATURE ONNX XGBOOST (2.1ms)
                    </span>
                    <span className="font-mono text-[10px] font-bold text-primary">
                      BETA CALIBRATED
                    </span>
                  </div>
                  <div className="mt-1 text-[13px] text-foreground font-semibold">
                    Calibrated Posterior: P(fraud | x) = 0.9418 · Expected Monetary Loss =
                    ₹13,65,610.00
                  </div>
                </div>

                <div className="grid grid-cols-3 gap-2 font-mono text-[11.5px]">
                  <div className="border border-border p-2.5 rounded bg-surface">
                    <div className="text-[10px] text-muted-foreground uppercase">
                      Raw Model Output
                    </div>
                    <div className="mt-1 font-bold text-[14px]">0.8712</div>
                  </div>
                  <div className="border border-border p-2.5 rounded bg-surface">
                    <div className="text-[10px] text-muted-foreground uppercase">
                      Beta Calibrated
                    </div>
                    <div className="mt-1 font-bold text-[14px] text-block">0.9418</div>
                  </div>
                  <div className="border border-border p-2.5 rounded bg-surface">
                    <div className="text-[10px] text-muted-foreground uppercase">
                      Decision Weight
                    </div>
                    <div className="mt-1 font-bold text-[14px] text-block">+0.20</div>
                  </div>
                </div>

                {/* OFFLINE MODEL EVALUATION SECTION */}
                <div className="rounded border border-border bg-surface p-3.5">
                  <div className="flex items-center justify-between border-b border-border pb-2">
                    <span className="font-mono text-[11px] font-bold text-foreground uppercase">
                      OFFLINE MODEL EVALUATION (HELD-OUT TEST SET)
                    </span>
                    <span className="font-mono text-[10px] text-muted-foreground">
                      N = 1,200 test samples · Threshold: 0.50
                    </span>
                  </div>
                  <div className="mt-3 grid grid-cols-2 gap-3 sm:grid-cols-4 font-mono text-[11.5px]">
                    <div className="border border-border p-2 rounded bg-accent/20">
                      <div className="text-[10px] text-muted-foreground">Precision</div>
                      <div className="mt-0.5 font-bold text-[13px] text-foreground">10.08%</div>
                      <div className="text-[9.5px] text-muted-foreground">TP/(TP+FP) = 13/129</div>
                    </div>
                    <div className="border border-border p-2 rounded bg-accent/20">
                      <div className="text-[10px] text-muted-foreground">Recall</div>
                      <div className="mt-0.5 font-bold text-[13px] text-foreground">25.00%</div>
                      <div className="text-[9.5px] text-muted-foreground">TP/(TP+FN) = 13/52</div>
                    </div>
                    <div className="border border-border p-2 rounded bg-accent/20">
                      <div className="text-[10px] text-muted-foreground">F1 Score</div>
                      <div className="mt-0.5 font-bold text-[13px] text-foreground">14.36%</div>
                      <div className="text-[9.5px] text-muted-foreground">Harmonic Mean</div>
                    </div>
                    <div className="border border-border p-2 rounded bg-accent/20">
                      <div className="text-[10px] text-muted-foreground">False-Positive Cost</div>
                      <div className="mt-0.5 font-bold text-[13px] text-block">₹40,000 / FP</div>
                      <div className="text-[9.5px] text-muted-foreground">
                        ₹46.4L total (116 FP)
                      </div>
                    </div>
                  </div>
                  <div className="mt-2 text-[11px] text-muted-foreground">
                    <span className="font-bold text-foreground">Cost Formula:</span> False-Positive
                    Cost = FP × ₹40,000. Expected Cost(Decline) = (1 - P(fraud)) × ₹40,000.
                  </div>
                </div>
              </div>
            )}

            {/* STAGE 5: Fraud Graph */}
            {stage === 4 && (
              <div className="space-y-4">
                <div className="rounded border border-block/40 bg-block/5 p-3.5">
                  <div className="flex items-center justify-between">
                    <span className="font-mono text-[11px] font-bold text-block uppercase">
                      3-HOP BFS FRAUD GRAPH TRAVERSAL (1.2ms)
                    </span>
                    <span className="font-mono text-[10px] font-bold text-block">
                      MULE RING DETECTED
                    </span>
                  </div>
                  <div className="mt-1 text-[13px] text-foreground font-semibold">
                    Syndicate Discovery: Hardware canvas hash links customer to 14 synthetic
                    accounts and confirmed cashout node PA-77120.
                  </div>
                </div>

                <div className="border border-border p-3.5 rounded font-mono text-[11.5px] space-y-2.5 bg-surface">
                  <div className="flex items-center gap-2">
                    <span className="size-2.5 rounded-full bg-primary" />
                    <span className="font-bold text-foreground">1-Hop: Target Account</span>
                    <span className="text-muted-foreground">cus_4471029</span>
                    <span className="text-muted-foreground">→ USED_DEVICE →</span>
                    <span className="font-bold text-block">dev_emulator_linux_9f8a</span>
                  </div>
                  <div className="flex items-center gap-2 pl-4">
                    <span className="size-2.5 rounded-full bg-warning" />
                    <span className="font-bold text-foreground">2-Hop: Shared Canvas</span>
                    <span className="text-muted-foreground">Canvas 9f8a84b12c</span>
                    <span className="text-muted-foreground">→ LINKED →</span>
                    <span className="font-bold text-block">14 Synthetic Mule Accounts</span>
                  </div>
                  <div className="flex items-center gap-2 pl-8">
                    <span className="size-2.5 rounded-full bg-block" />
                    <span className="font-bold text-foreground">3-Hop: Payout Depot</span>
                    <span className="text-muted-foreground">All 14 accounts</span>
                    <span className="text-muted-foreground">→ PAYOUT_DESTINATION →</span>
                    <span className="font-bold text-block">PA-77120 (Confirmed Mule Cashout)</span>
                  </div>
                </div>

                <div className="flex items-center justify-between pt-1">
                  <div className="font-mono text-[11.5px] text-muted-foreground">
                    Graph Centrality: <span className="font-bold text-block">Degree 16</span>{" "}
                    (Threshold &ge; 4 triggers syndicate block)
                  </div>
                  <Link
                    to="/graph"
                    className="font-mono text-[12px] text-primary hover:underline font-semibold"
                  >
                    Open Interactive Fraud Graph →
                  </Link>
                </div>
              </div>
            )}

            {/* STAGE 6: Decision & Precedence */}
            {stage === 5 && (
              <div className="space-y-4">
                <div className="rounded border border-block/40 bg-block/5 p-3.5">
                  <div className="flex items-center justify-between">
                    <span className="font-mono text-[11px] font-bold text-block uppercase">
                      5-LEVEL DECISION PRECEDENCE ARBITRATION
                    </span>
                    <span className="font-mono text-[10px] font-bold text-block">
                      FINAL VERDICT: BLOCK
                    </span>
                  </div>
                  <div className="mt-1 text-[13px] text-foreground font-semibold">
                    Composite Risk Score: 0.96 / 1.00 · Policy Action: BLOCK_AND_REVIEW ·
                    Confidence: 94%
                  </div>
                </div>

                <div className="space-y-2 text-[12px] font-mono">
                  <div className="border border-border p-2.5 rounded bg-surface flex items-center justify-between">
                    <span>Level 1: Statutory Compliance / Sanctions</span>
                    <span className="text-approve font-bold">PASS (No OFAC/RBI match)</span>
                  </div>
                  <div className="border border-border p-2.5 rounded bg-surface flex items-center justify-between">
                    <span>Level 2: Declarative Policy Rules</span>
                    <span className="text-block font-bold">ESCALATE (Speed + Proxy fired)</span>
                  </div>
                  <div className="border border-border p-2.5 rounded bg-surface flex items-center justify-between">
                    <span>Level 3: Fraud Knowledge Graph</span>
                    <span className="text-block font-bold">ESCALATE (Degree 16 Mule Ring)</span>
                  </div>
                  <div className="border border-border p-2.5 rounded bg-surface flex items-center justify-between">
                    <span>Level 4: Calibrated XGBoost ML</span>
                    <span className="text-block font-bold">
                      P(fraud) = 0.9418 (Block Band &ge; 0.80)
                    </span>
                  </div>
                </div>

                <div className="border border-border p-3 rounded bg-accent/20 text-[12px]">
                  <div className="font-semibold text-foreground">Definitive Reasons:</div>
                  <ul className="mt-1 space-y-1 text-muted-foreground list-disc list-inside">
                    <li>
                      Cross-border impossible travel from overseas proxy (Limassol) at 36,250 km/h
                      after active Bengaluru session
                    </li>
                    <li>
                      IP address originates from commercial bulletproof proxy / VPN (ASN 13335)
                    </li>
                    <li>
                      Entity hardware identifier is linked across 14 synthetic mule accounts in
                      graph
                    </li>
                    <li>XGBoost calibrated posterior probability 94.18%</li>
                  </ul>
                </div>
              </div>
            )}

            {/* STAGE 7: Human Action & Audit Hash Chain */}
            {stage === 6 && (
              <div className="space-y-4">
                <div
                  className={cn(
                    "rounded border p-3.5 transition-colors",
                    confirmedBlock ? "border-approve bg-approve/10" : "border-block bg-block/5",
                  )}
                >
                  <div className="font-mono text-[11px] font-bold text-foreground uppercase">
                    {confirmedBlock
                      ? "ACTION CONFIRMED & AUDITED TO HASH CHAIN"
                      : "HUMAN GOVERNANCE REVIEW REQUIRED (SLA: 15m)"}
                  </div>
                  <div className="mt-1 text-[13px] text-foreground font-semibold">
                    {confirmedBlock
                      ? "Transaction BLOCKED. Cryptographic SHA-256 hash chain entry appended. Retraining label emitted."
                      : "Reviewing CASE-88419 (Analyst: a.sharma). Human authorization required to confirm payment block."}
                  </div>
                </div>

                {!confirmedBlock ? (
                  <div className="space-y-3">
                    <button
                      type="button"
                      onClick={() => setConfirmedBlock(true)}
                      className="w-full rounded bg-block py-3 font-mono text-[13px] font-bold text-white shadow-xs hover:bg-block/90 active:scale-[0.99] transition-all"
                    >
                      CONFIRM BLOCK &amp; FREEZE ACCOUNT
                    </button>
                    <div className="text-center text-[11.5px] text-muted-foreground font-mono">
                      Authorizing analyst:{" "}
                      <span className="font-bold text-foreground">
                        a.sharma (Senior Fraud Lead)
                      </span>
                    </div>
                  </div>
                ) : (
                  <div className="space-y-2.5 border border-border p-3.5 bg-accent/20 rounded font-mono text-[11.5px]">
                    <div className="text-approve font-bold flex items-center gap-1.5 text-[12.5px]">
                      <span>✓</span> Case Outcome: CONFIRMED_FRAUD (Analyst: a.sharma)
                    </div>
                    <div className="text-muted-foreground">
                      Cryptographic Audit Hash:{" "}
                      <Mono className="text-foreground font-bold">
                        sha256:4f8a1e9c7a2b53018e69d72...
                      </Mono>
                    </div>
                    <div className="text-muted-foreground">
                      Hash Chain Link:{" "}
                      <Mono className="text-foreground">
                        H_i = SHA-256(H_i-1 || EntryID || Timestamp)
                      </Mono>
                    </div>
                    <div className="text-muted-foreground">
                      Closed-Loop Retraining Event:{" "}
                      <Mono className="text-foreground">evt_retrain_ground_truth_88419</Mono>
                    </div>
                  </div>
                )}
              </div>
            )}
          </div>
        </section>
      </div>

      {/* ------------------------------------------------ Architecture Failure & Safe Degradation Analysis */}
      <section className="mt-8 rounded border border-border bg-surface p-5">
        <div className="flex flex-wrap items-center justify-between gap-3 border-b border-border pb-3">
          <div>
            <div className="flex items-center gap-2 font-mono text-[11px] font-bold text-primary uppercase">
              <span>● BUILD QUALITY &amp; RESILIENCE</span>
            </div>
            <h3 className="mt-1 text-[15px] font-bold text-foreground">
              Failure Recovery &amp; Safe Degradation Invariants
            </h3>
            <p className="mt-0.5 text-[12.5px] text-muted-foreground">
              What happens when downstream dependencies fail in production? The system degrades into
              safe, deterministic modes without dropping customer transactions.
            </p>
          </div>
          <div className="flex items-center gap-1.5 font-mono text-[11.5px]">
            <button
              type="button"
              onClick={() => setFailureScenario("NONE")}
              className={cn(
                "rounded px-2.5 py-1 font-semibold transition-colors cursor-pointer",
                failureScenario === "NONE"
                  ? "bg-foreground text-background font-bold"
                  : "bg-accent/40 text-muted-foreground hover:text-foreground",
              )}
            >
              Baseline (Healthy)
            </button>
            <button
              type="button"
              onClick={() => setFailureScenario("ML_TIMEOUT")}
              className={cn(
                "rounded px-2.5 py-1 font-semibold transition-colors cursor-pointer",
                failureScenario === "ML_TIMEOUT"
                  ? "bg-warning text-foreground font-bold"
                  : "bg-accent/40 text-muted-foreground hover:text-foreground",
              )}
            >
              ML Timeout (&gt;50ms)
            </button>
            <button
              type="button"
              onClick={() => setFailureScenario("REDIS_DOWN")}
              className={cn(
                "rounded px-2.5 py-1 font-semibold transition-colors cursor-pointer",
                failureScenario === "REDIS_DOWN"
                  ? "bg-warning text-foreground font-bold"
                  : "bg-accent/40 text-muted-foreground hover:text-foreground",
              )}
            >
              Redis Unavailable
            </button>
            <button
              type="button"
              onClick={() => setFailureScenario("KAFKA_DOWN")}
              className={cn(
                "rounded px-2.5 py-1 font-semibold transition-colors cursor-pointer",
                failureScenario === "KAFKA_DOWN"
                  ? "bg-warning text-foreground font-bold"
                  : "bg-accent/40 text-muted-foreground hover:text-foreground",
              )}
            >
              Kafka Partitioned
            </button>
          </div>
        </div>

        {/* Dynamic Scenario Breakdown */}
        <div className="mt-4">
          {failureScenario === "NONE" && (
            <div className="grid grid-cols-1 gap-3 sm:grid-cols-3 font-mono text-[12px]">
              <div className="rounded border border-border bg-accent/10 p-3">
                <div className="text-approve font-bold flex items-center gap-1.5">
                  <span className="size-2 rounded-full bg-approve" />
                  ML Sidecar (gRPC / ONNX)
                </div>
                <div className="mt-1 text-[11.5px] text-muted-foreground">
                  Healthy · p99 Latency: 2.1ms · Timeout Budget: 50ms
                </div>
              </div>
              <div className="rounded border border-border bg-accent/10 p-3">
                <div className="text-approve font-bold flex items-center gap-1.5">
                  <span className="size-2 rounded-full bg-approve" />
                  Redis Feature Store
                </div>
                <div className="mt-1 text-[11.5px] text-muted-foreground">
                  Healthy · In-memory sliding velocity &amp; canvas cache
                </div>
              </div>
              <div className="rounded border border-border bg-accent/10 p-3">
                <div className="text-approve font-bold flex items-center gap-1.5">
                  <span className="size-2 rounded-full bg-approve" />
                  Event Stream &amp; Audit
                </div>
                <div className="mt-1 text-[11.5px] text-muted-foreground">
                  Healthy · Transactional Outbox + Synchronous Hash Chain
                </div>
              </div>
            </div>
          )}

          {failureScenario === "ML_TIMEOUT" && (
            <div className="rounded border border-warning/40 bg-warning/5 p-4 font-mono text-[12px]">
              <div className="flex items-center justify-between">
                <span className="font-bold text-warning uppercase">
                  SIMULATED FAILURE: ML SIDECAR DEADLINE EXCEEDED (&gt;50ms)
                </span>
                <span className="rounded bg-warning/20 px-2 py-0.5 text-[10px] font-bold text-warning">
                  STATUS: DEGRADED (is_degraded: true)
                </span>
              </div>
              <div className="mt-2 space-y-1.5 text-foreground text-[12px]">
                <div>
                  1. <span className="text-muted-foreground">Orchestrator Action:</span> Context
                  cancellation fires at 50ms deadline; cancels downstream gRPC call to ONNX sidecar.
                </div>
                <div>
                  2. <span className="text-muted-foreground">Fallback Policy:</span> Engages Tier 4
                  deterministic heuristic rule fallback (
                  <Mono className="text-foreground">RULE_HIGH_TRANSACTION_AMOUNT</Mono>,{" "}
                  <Mono className="text-foreground">RULE_UNUSUAL_HOUR</Mono>).
                </div>
                <div>
                  3. <span className="text-muted-foreground">Decision Safety:</span> Assigned
                  neutral ML baseline contribution (+0.10); decision completes in 1.8ms total;
                  payload marked <Mono className="text-warning">is_degraded: true</Mono>.
                </div>
                <div>
                  4. <span className="text-muted-foreground">Auditing:</span> Degraded fallback
                  event written to audit hash-chain ledger with reason code{" "}
                  <Mono className="text-warning">ML_INFERENCE_TIMEOUT_FALLBACK</Mono>.
                </div>
              </div>
            </div>
          )}

          {failureScenario === "REDIS_DOWN" && (
            <div className="rounded border border-warning/40 bg-warning/5 p-4 font-mono text-[12px]">
              <div className="flex items-center justify-between">
                <span className="font-bold text-warning uppercase">
                  SIMULATED FAILURE: REDIS DEVICE FEATURE STORE UNREACHABLE
                </span>
                <span className="rounded bg-warning/20 px-2 py-0.5 text-[10px] font-bold text-warning">
                  STATUS: DEGRADED (is_degraded: true)
                </span>
              </div>
              <div className="mt-2 space-y-1.5 text-foreground text-[12px]">
                <div>
                  1. <span className="text-muted-foreground">Backend Handler:</span>{" "}
                  <Mono className="text-warning">device_redis.go</Mono> catches connection timeout;
                  sets <Mono className="text-warning">IsDegraded: true</Mono> and logs{" "}
                  <Mono className="text-warning">VELOCITY_FEATURE_STORE_UNAVAILABLE</Mono>.
                </div>
                <div>
                  2. <span className="text-muted-foreground">Safe Defaults:</span> Emits neutral
                  reputation baseline (
                  <Mono className="text-foreground">DeviceReputationScore = 0.50</Mono>, 1h count =
                  0) preventing false-positive customer rejections.
                </div>
                <div>
                  3. <span className="text-muted-foreground">Local Cache:</span> In-memory LRU ring
                  buffer maintains active customer sessions locally until Redis reconnects.
                </div>
                <div>
                  4. <span className="text-muted-foreground">Zero Downtime:</span> Synchronous
                  transaction flow continues without customer disruption or latency spike.
                </div>
              </div>
            </div>
          )}

          {failureScenario === "KAFKA_DOWN" && (
            <div className="rounded border border-warning/40 bg-warning/5 p-4 font-mono text-[12px]">
              <div className="flex items-center justify-between">
                <span className="font-bold text-warning uppercase">
                  SIMULATED FAILURE: KAFKA / STREAMING BROKER PARTITIONED
                </span>
                <span className="rounded bg-warning/20 px-2 py-0.5 text-[10px] font-bold text-warning">
                  STATUS: BUFFERED IN OUTBOX
                </span>
              </div>
              <div className="mt-2 space-y-1.5 text-foreground text-[12px]">
                <div>
                  1. <span className="text-muted-foreground">Resilience Mechanism:</span>{" "}
                  <Mono className="text-warning">health_manager.go</Mono> catches broker disconnect
                  and invokes <Mono className="text-foreground">BufferFallbackEvent</Mono>.
                </div>
                <div>
                  2. <span className="text-muted-foreground">Transactional Outbox:</span> Audit
                  events and ML retraining labels are committed atomically into PostgreSQL{" "}
                  <Mono className="text-foreground">outbox_events</Mono> table.
                </div>
                <div>
                  3. <span className="text-muted-foreground">Replay Worker:</span> Background
                  consumer flushes buffered events with exponential backoff once Kafka recovers.
                </div>
                <div>
                  4. <span className="text-muted-foreground">Zero Loss:</span> Exactly-once delivery
                  semantics preserved; synchronous risk decision path remains 100% unblocked.
                </div>
              </div>
            </div>
          )}
        </div>
      </section>
    </Page>
  );
}
