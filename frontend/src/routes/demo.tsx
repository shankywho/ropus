import { createFileRoute, Link } from "@tanstack/react-router";
import { useState, useEffect, useRef } from "react";
import { Mono, RiskScore, VerdictBadge, PriorityTag, CaseStatusTag, StatusPill } from "@/components/ropus/core";
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
  const [activeTab, setActiveTab] = useState<"telemetry" | "live_api" | "bmr_matrix" | "feature_contract">("telemetry");
  const [liveEvalResult, setLiveEvalResult] = useState<LiveEvaluationResponse | null>(null);
  const [liveLoading, setLiveLoading] = useState(false);
  const [liveError, setLiveError] = useState<string | null>(null);
  const [customAmount, setCustomAmount] = useState<number>(145000000);
  const [customCurrency, setCustomCurrency] = useState<string>("INR");
  const [customIp, setCustomIp] = useState<string>("198.51.100.44");
  const [customDevice, setCustomDevice] = useState<string>("dev_emulator_linux_9f8a");
  const [showPayloadEditor, setShowPayloadEditor] = useState<boolean>(false);
  const [failureScenario, setFailureScenario] = useState<
    "NONE" | "ML_TIMEOUT" | "REDIS_DOWN" | "KAFKA_DOWN"
  >("NONE");

  // Stage-specific live presets
  const stagePresets = [
    {
      txId: "txn_baseline_99120",
      customerId: "cus_4471029",
      amount: 48000,
      currency: "INR",
      ip: "103.21.124.10",
      device: "dev_safari_ios_01",
      paymentType: "upi",
      paymentToken: "vpa_sarah_connor@okaxis",
      description: "Organic Domestic Quick-Commerce Spend (Bengaluru, KA)",
    },
    {
      txId: "txn_attack_88419",
      customerId: "cus_4471029",
      amount: 145000000,
      currency: "INR",
      ip: "198.51.100.44",
      device: "dev_emulator_linux_9f8a",
      paymentType: "imps",
      paymentToken: "payout_pa_77120",
      description: "High-Value Outbound IMPS Payout (Limassol Proxy ASN 13335)",
    },
    {
      txId: "txn_velocity_surge_88420",
      customerId: "cus_4471029",
      amount: 145000000,
      currency: "INR",
      ip: "198.51.100.44",
      device: "dev_emulator_linux_9f8a",
      paymentType: "imps",
      paymentToken: "payout_pa_77120",
      description: "Rules AST Surge: Velocity Spike + Impossible Travel",
    },
    {
      txId: "txn_ml_scoring_88421",
      customerId: "cus_4471029",
      amount: 145000000,
      currency: "INR",
      ip: "198.51.100.44",
      device: "dev_emulator_linux_9f8a",
      paymentType: "imps",
      paymentToken: "payout_pa_77120",
      description: "ML ONNX 25-Feature Tree Traversal & Beta Posterior",
    },
    {
      txId: "txn_graph_syndicate_88422",
      customerId: "cus_4471029",
      amount: 145000000,
      currency: "INR",
      ip: "198.51.100.44",
      device: "dev_emulator_linux_9f8a",
      paymentType: "imps",
      paymentToken: "payout_pa_77120",
      description: "Graph 3-Hop Syndicate Ring Traversal to Mule Depot",
    },
    {
      txId: "txn_precedence_block_88423",
      customerId: "cus_4471029",
      amount: 145000000,
      currency: "INR",
      ip: "198.51.100.44",
      device: "dev_emulator_linux_9f8a",
      paymentType: "imps",
      paymentToken: "payout_pa_77120",
      description: "5-Level Hierarchy Arbitration & BMR Optimization",
    },
    {
      txId: "txn_audit_freeze_88424",
      customerId: "cus_4471029",
      amount: 145000000,
      currency: "INR",
      ip: "198.51.100.44",
      device: "dev_emulator_linux_9f8a",
      paymentType: "imps",
      paymentToken: "payout_pa_77120",
      description: "Analyst Decision Confirmation & Cryptographic Hash Append",
    },
  ];

  const currentPreset = stagePresets[stage] || stagePresets[0];

  // Sync custom inputs with preset when stage changes
  useEffect(() => {
    setCustomAmount(currentPreset.amount);
    setCustomCurrency(currentPreset.currency);
    setCustomIp(currentPreset.ip);
    setCustomDevice(currentPreset.device);
    setLiveError(null);
  }, [stage, currentPreset]);

  const handleLiveEvaluate = async (overrideParams?: Partial<typeof currentPreset>) => {
    setLiveLoading(true);
    setLiveError(null);
    try {
      const p = {
        ...currentPreset,
        amount: customAmount,
        currency: customCurrency,
        ip: customIp,
        device: customDevice,
        ...overrideParams,
      };

      const res = await evaluateRisk({
        transaction_id: p.txId,
        customer_id: p.customerId,
        amount: p.amount,
        currency: p.currency,
        ip_address: p.ip,
        device_fingerprint: p.device,
        payment_method: {
          type: p.paymentType,
          token: p.paymentToken,
        },
      });
      setLiveEvalResult(res);
      setActiveTab("live_api");
    } catch (e: any) {
      console.error("Live evaluate failed", e);
      setLiveError(e.message || "Failed to communicate with Go backend");
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
            <span className="border border-navy bg-navy/10 px-2 py-0.5 font-mono text-[10px] font-bold text-navy uppercase tracking-[0.06em]">
              DEMO REPLAY
            </span>
            <span className="font-mono text-[11px] text-muted-foreground">
              Deterministic 7-Stage Incident Replay Pipeline
            </span>
          </div>
          <h1 className="mt-1 font-sans text-[25px] font-extrabold leading-[1.15] tracking-[-0.045em] text-foreground">
            Attack Evaluation Lifecycle
          </h1>
          <p className="mt-0.5 font-sans text-[12px] text-muted-foreground">
            Normal baseline → high-risk wire → rules triggers → BMR ML calibration → GraphSAGE shadow traversal →
            AI evidence → analyst freeze.
          </p>
        </div>

        {/* Playback Controls */}
        <div className="flex flex-wrap items-center gap-2">
          <button
            type="button"
            onClick={() => (playing ? demoControls.pause() : demoControls.play())}
            className={cn(
              "border px-3 py-1.5 font-mono text-[11px] font-medium transition-colors cursor-pointer",
              playing
                ? "border-amber-intel bg-shadow-intel-surface text-shadow-intel"
                : "border-navy bg-navy text-white hover:bg-navy/90",
            )}
          >
            {playing ? "⏸ Pause Replay" : stage === 0 ? "▶ Start 7-Stage Replay" : "▶ Resume"}
          </button>
          <button
            type="button"
            onClick={() => demoControls.back()}
            disabled={stage === 0}
            className="border border-border bg-surface px-2.5 py-1.5 font-mono text-[11px] font-medium text-foreground hover:bg-secondary disabled:opacity-40 cursor-pointer"
          >
            ← Back
          </button>
          <button
            type="button"
            onClick={() => demoControls.next()}
            disabled={stage === demoStages.length - 1}
            className="border border-border bg-surface px-2.5 py-1.5 font-mono text-[11px] font-medium text-foreground hover:bg-secondary disabled:opacity-40 cursor-pointer"
          >
            Step →
          </button>
          <button
            type="button"
            onClick={() => demoControls.reset()}
            className="border border-border bg-surface px-2.5 py-1.5 font-mono text-[11px] text-muted-foreground hover:text-foreground cursor-pointer"
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
                  "px-1.5 py-0.5 font-mono text-[10px] cursor-pointer",
                  intervalMs === ms
                    ? "bg-navy text-white font-bold"
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
                "flex flex-col items-start border p-2 text-left transition-all cursor-pointer bg-card",
                isActive
                  ? "border-navy ring-1 ring-navy bg-secondary/50 shadow-2xs"
                  : isPassed
                    ? "border-border-strong text-foreground hover:bg-secondary/40"
                    : "border-border text-muted-foreground hover:border-border-strong",
              )}
            >
              <div className="flex w-full items-center justify-between font-mono text-[9.5px]">
                <span
                  className={cn("font-medium", isActive ? "text-navy font-bold" : "text-muted-foreground")}
                >
                  STAGE {s.stageNumber}
                </span>
                {isPassed && <span className="text-authoritative font-bold">✓</span>}
              </div>
              <div className="mt-1 truncate font-sans text-[11.5px] font-semibold text-foreground">
                {s.label.split(" ")[0]} {s.label.split(" ")[1] ?? ""}
              </div>
            </button>
          );
        })}
      </nav>

      {/* ------------------------------------------------ Current Stage Core Presentation Area */}
      <div className="mt-6 grid gap-6 lg:grid-cols-[380px_minmax(0,1fr)]">
        {/* Left Column: Stage Summary, Score Gauge & Factor Attribution */}
        <section className="flex flex-col justify-between border border-border bg-card p-5 shadow-xs">
          <div className="space-y-4">
            <div>
              <div className="flex items-center justify-between">
                <span className="font-mono text-[10px] font-medium tracking-[0.12em] text-navy uppercase">
                  Stage {current.stageNumber} of 7
                </span>
                <VerdictBadge verdict={current.verdict} size="sm" />
              </div>
              <h2 className="mt-1 font-sans text-[17px] font-bold text-foreground">{current.label}</h2>
              <p className="mt-1 font-sans text-[12px] leading-relaxed text-muted-foreground">
                {current.shortExplanation}
              </p>
            </div>

            {/* Score Display */}
            <div className="border-t border-border pt-4">
              <div className="font-mono text-[9.5px] font-medium tracking-[0.12em] text-muted-foreground uppercase">
                Risk Score Invariant
              </div>
              <div className="mt-2 flex items-baseline justify-between">
                <RiskScore value={tweenedScore} size="hero" showBand />
                <div className="text-right font-mono text-[10.5px] text-muted-foreground">
                  <div>Base: 0.02</div>
                  <div>Attribution: +{(tweenedScore - 0.02).toFixed(2)}</div>
                </div>
              </div>
            </div>

            {/* Additive Factors Summary */}
            <div className="border-t border-border pt-4">
              <div className="font-mono text-[9.5px] font-medium tracking-[0.12em] text-muted-foreground uppercase">
                Active Factor Contributions
              </div>
              <ul className="mt-2 space-y-1.5 text-[11.5px]">
                {stage === 0 && (
                  <li className="flex items-center justify-between font-mono text-muted-foreground">
                    <span>Base Account Trust (14mo clean)</span>
                    <span className="text-authoritative font-bold">+0.02</span>
                  </li>
                )}
                {stage >= 1 && (
                  <li className="flex items-center justify-between font-mono">
                    <span className="text-foreground">Unrecognised Device & Geo Jump</span>
                    <span className="text-blocked font-bold">+0.26</span>
                  </li>
                )}
                {stage >= 2 && (
                  <li className="flex items-center justify-between font-mono">
                    <span className="text-foreground">Velocity Surge + Impossible Travel Speed</span>
                    <span className="text-blocked font-bold">+0.37</span>
                  </li>
                )}
                {stage >= 3 && (
                  <li className="flex items-center justify-between font-mono">
                    <span className="text-foreground">BMR XGBoost Posterior Calibration</span>
                    <span className="text-blocked font-bold">+0.23</span>
                  </li>
                )}
                {stage >= 4 && (
                  <li className="flex items-center justify-between font-mono">
                    <span className="text-foreground">GraphSAGE Shadow Ring Discovery</span>
                    <span className="text-amber-intel font-bold">+0.06</span>
                  </li>
                )}
                {stage >= 5 && (
                  <li className="flex items-center justify-between font-mono">
                    <span className="text-foreground">Agent Tripartite Dossier Synthesis</span>
                    <span className="text-blocked font-bold">+0.02</span>
                  </li>
                )}
                {stage >= 6 && (
                  <li className="flex items-center justify-between font-mono">
                    <span className="text-foreground">Immutable Hash-Chain Commitment</span>
                    <span className="text-authoritative font-bold">LOCKED</span>
                  </li>
                )}
              </ul>
            </div>

            {/* Quick Live Trigger Button on Left Column */}
            <div className="border-t border-border pt-4">
              <button
                type="button"
                onClick={() => handleLiveEvaluate()}
                disabled={liveLoading}
                className="w-full border border-navy bg-navy/10 hover:bg-navy/20 text-navy py-2 font-mono text-[11px] font-bold transition-all flex items-center justify-center gap-1.5 cursor-pointer disabled:opacity-50"
              >
                <span>⚡</span>
                {liveLoading ? "Executing Live Go Evaluation..." : `Test Stage ${current.stageNumber} Live on Go API`}
              </button>
            </div>
          </div>

          {/* Next Button */}
          <div className="mt-6 border-t border-border pt-4">
            <button
              type="button"
              onClick={nextStage}
              className="w-full border border-navy bg-navy py-2.5 text-center font-mono text-[12px] font-medium text-white transition-opacity hover:opacity-90 cursor-pointer"
            >
              {stage < 6 ? `Next: Stage ${stage + 2} →` : "Restart Demo Replay ↺"}
            </button>
          </div>
        </section>

        {/* Right Column: Deep Technical Telemetry & Multi-Tab Inspector */}
        <section className="border border-border bg-card p-5 shadow-xs flex flex-col justify-between">
          <div>
            {/* Header Tabs */}
            <div className="flex flex-wrap items-center justify-between border-b border-border pb-3 gap-2">
              <div className="flex items-center gap-1 font-mono text-[11px]">
                <button
                  type="button"
                  onClick={() => setActiveTab("telemetry")}
                  className={cn(
                    "px-3 py-1 font-medium transition-colors cursor-pointer border",
                    activeTab === "telemetry"
                      ? "border-navy bg-navy text-white font-bold"
                      : "border-transparent text-muted-foreground hover:text-foreground",
                  )}
                >
                  📊 Stage Telemetry
                </button>
                <button
                  type="button"
                  onClick={() => setActiveTab("live_api")}
                  className={cn(
                    "px-3 py-1 font-medium transition-colors cursor-pointer border flex items-center gap-1.5",
                    activeTab === "live_api"
                      ? "border-navy bg-navy text-white font-bold"
                      : "border-transparent text-muted-foreground hover:text-foreground",
                  )}
                >
                  <span className="size-1.5 rounded-full bg-emerald-500 animate-pulse" />
                  ⚡ Live Go API ({liveEvalResult ? `${liveEvalResult.latency_ms}ms` : "Ready"})
                </button>
                <button
                  type="button"
                  onClick={() => setActiveTab("bmr_matrix")}
                  className={cn(
                    "px-3 py-1 font-medium transition-colors cursor-pointer border",
                    activeTab === "bmr_matrix"
                      ? "border-navy bg-navy text-white font-bold"
                      : "border-transparent text-muted-foreground hover:text-foreground",
                  )}
                >
                  ⚖️ BMR Cost Matrix
                </button>
                <button
                  type="button"
                  onClick={() => setActiveTab("feature_contract")}
                  className={cn(
                    "px-3 py-1 font-medium transition-colors cursor-pointer border",
                    activeTab === "feature_contract"
                      ? "border-navy bg-navy text-white font-bold"
                      : "border-transparent text-muted-foreground hover:text-foreground",
                  )}
                >
                  🧬 25F/58F Contracts
                </button>
              </div>

              <div className="flex items-center gap-2">
                <button
                  type="button"
                  onClick={() => setShowPayloadEditor(!showPayloadEditor)}
                  className="font-mono text-[10.5px] text-navy hover:underline cursor-pointer flex items-center gap-1"
                >
                  ⚙ {showPayloadEditor ? "Hide Tuner" : "Tune Payload"}
                </button>
              </div>
            </div>

            {/* Expandable Live Payload Tuner */}
            {showPayloadEditor && (
              <div className="mt-3 border border-border bg-secondary/40 p-3.5 font-mono text-[11px] space-y-3">
                <div className="flex items-center justify-between font-bold text-foreground">
                  <span>Interactive Real-Time Payload Tuner</span>
                  <span className="text-[10px] text-muted-foreground">Direct Ingress to POST /v1/risk-evaluations</span>
                </div>
                <div className="grid grid-cols-2 sm:grid-cols-4 gap-2.5">
                  <div>
                    <label className="text-[9.5px] text-muted-foreground block uppercase">Amount (Units)</label>
                    <input
                      type="number"
                      value={customAmount}
                      onChange={(e) => setCustomAmount(Number(e.target.value))}
                      className="mt-1 w-full border border-border bg-surface px-2 py-1 text-foreground font-mono text-[11px]"
                    />
                  </div>
                  <div>
                    <label className="text-[9.5px] text-muted-foreground block uppercase">Currency</label>
                    <input
                      type="text"
                      value={customCurrency}
                      onChange={(e) => setCustomCurrency(e.target.value)}
                      className="mt-1 w-full border border-border bg-surface px-2 py-1 text-foreground font-mono text-[11px]"
                    />
                  </div>
                  <div>
                    <label className="text-[9.5px] text-muted-foreground block uppercase">IP Address</label>
                    <input
                      type="text"
                      value={customIp}
                      onChange={(e) => setCustomIp(e.target.value)}
                      className="mt-1 w-full border border-border bg-surface px-2 py-1 text-foreground font-mono text-[11px]"
                    />
                  </div>
                  <div>
                    <label className="text-[9.5px] text-muted-foreground block uppercase">Device Fingerprint</label>
                    <input
                      type="text"
                      value={customDevice}
                      onChange={(e) => setCustomDevice(e.target.value)}
                      className="mt-1 w-full border border-border bg-surface px-2 py-1 text-foreground font-mono text-[11px]"
                    />
                  </div>
                </div>
                <div className="flex justify-end gap-2 pt-1">
                  <button
                    type="button"
                    onClick={() => {
                      setCustomAmount(currentPreset.amount);
                      setCustomCurrency(currentPreset.currency);
                      setCustomIp(currentPreset.ip);
                      setCustomDevice(currentPreset.device);
                    }}
                    className="border border-border bg-surface px-2.5 py-1 text-muted-foreground hover:text-foreground text-[10.5px] cursor-pointer"
                  >
                    Reset to Stage Preset
                  </button>
                  <button
                    type="button"
                    onClick={() => handleLiveEvaluate()}
                    disabled={liveLoading}
                    className="border border-navy bg-navy px-3 py-1 font-bold text-white text-[10.5px] hover:bg-navy/90 cursor-pointer disabled:opacity-50"
                  >
                    {liveLoading ? "Evaluating..." : "⚡ Execute Custom Ingress"}
                  </button>
                </div>
              </div>
            )}

            {/* TAB CONTENT 1: STAGE TELEMETRY */}
            {activeTab === "telemetry" && (
              <div className="mt-4">
                {/* STAGE 1: Baseline */}
                {stage === 0 && (
                  <div className="space-y-4">
                    <div className="border border-border bg-secondary/50 p-3.5">
                      <div className="flex items-center justify-between font-mono text-[10.5px]">
                        <span className="font-bold text-navy">CUSTOMER PROFILE: cus_4471029</span>
                        <StatusPill tone="authoritative">TRUSTED ACTIVE</StatusPill>
                      </div>
                      <div className="mt-1 font-sans text-[12.5px] text-foreground">
                        Tenure: 14 months · Clean history (0 disputes) · 90-day moving average ticket:
                        ₹480.00 INR
                      </div>
                    </div>

                    <div className="grid grid-cols-2 gap-3 text-[12px]">
                      <div className="border border-border p-3 bg-surface">
                        <div className="font-mono text-[9.5px] font-medium text-muted-foreground uppercase">
                          Trusted Hardware Canvas
                        </div>
                        <Mono className="mt-1 font-bold text-foreground">dev_safari_ios_01</Mono>
                        <div className="mt-1 font-sans text-[11.5px] text-muted-foreground">
                          Mobile Safari · iOS 17.4 · Canvas 4a12 · Entropy 0.24
                        </div>
                      </div>
                      <div className="border border-border p-3 bg-surface">
                        <div className="font-mono text-[9.5px] font-medium text-muted-foreground uppercase">
                          Primary Network Subnet
                        </div>
                        <Mono className="mt-1 font-bold text-foreground">
                          Jio Fiber (Bengaluru, KA)
                        </Mono>
                        <div className="mt-1 font-sans text-[11.5px] text-muted-foreground">
                          ASN 55836 · 12.97°N, 77.59°E · Residential
                        </div>
                      </div>
                    </div>

                    <div className="border border-border p-3 bg-surface">
                      <div className="font-mono text-[9.5px] font-medium text-muted-foreground uppercase">
                        Recent Normal Transactions (Last 24h)
                      </div>
                      <div className="mt-2 space-y-1.5 font-mono text-[11px]">
                        <div className="flex justify-between text-muted-foreground">
                          <span>txn_99120 · Quick Commerce (Bengaluru)</span>
                          <span className="text-foreground font-semibold">
                            ₹385.00 INR · APPROVE (0.01)
                          </span>
                        </div>
                        <div className="flex justify-between text-muted-foreground">
                          <span>txn_99118 · Ride Hailing (Bengaluru)</span>
                          <span className="text-foreground font-semibold">
                            ₹142.00 INR · APPROVE (0.02)
                          </span>
                        </div>
                      </div>
                    </div>

                    <div className="border border-border border-l-2 border-l-navy bg-surface p-3 font-sans text-[12px] text-muted-foreground">
                      <span className="font-bold text-foreground">Baseline Invariant:</span> Normal
                      organic behavior is established by low amount deviation, residential IP ASN,
                      consistent hardware entropy, and zero graph connectivity to known chargeback
                      syndicates.
                    </div>
                  </div>
                )}

                {/* STAGE 2: Suspicious Ingress */}
                {stage === 1 && (
                  <div className="space-y-4">
                    <div className="border border-blocked/40 bg-blocked-surface/40 p-3.5">
                      <div className="flex items-center justify-between">
                        <span className="font-mono text-[10.5px] font-bold text-blocked uppercase">
                          INBOUND API INGRESS: POST /v1/risk-evaluations
                        </span>
                        <StatusPill tone="blocked">HIGH-RISK IMPS PAYOUT</StatusPill>
                      </div>
                      <div className="mt-1 font-sans text-[13.5px] text-foreground font-bold">
                        Amount: ₹14,50,000.00 INR (Outbound IMPS to Beneficiary PA-77120)
                      </div>
                      <div className="mt-0.5 font-sans text-[11.5px] text-muted-foreground">
                        Origin: Limassol proxy · IP: 198.51.100.44 · Device: dev_emulator_linux_9f8a
                      </div>
                    </div>

                    {/* Signals & Features Extracted */}
                    <div className="space-y-2">
                      <div className="font-mono text-[10px] font-medium tracking-[0.12em] text-muted-foreground uppercase">
                        Extracted Risk Signals &amp; Features (Context: 1.42ms)
                      </div>
                      <div className="grid grid-cols-2 gap-2 text-[11px] font-mono">
                        <div className="border border-border bg-surface p-2">
                          <span className="text-muted-foreground">Geographic Jump:</span>{" "}
                          <span className="font-bold text-blocked">7,250 km (12 min)</span>
                        </div>
                        <div className="border border-border bg-surface p-2">
                          <span className="text-muted-foreground">Implied Travel Speed:</span>{" "}
                          <span className="font-bold text-blocked">36,250 km/h (&gt; 900)</span>
                        </div>
                        <div className="border border-border bg-surface p-2">
                          <span className="text-muted-foreground">Device Novelty / Entropy:</span>{" "}
                          <span className="font-bold text-blocked">0.92 (Emulator VM)</span>
                        </div>
                        <div className="border border-border bg-surface p-2">
                          <span className="text-muted-foreground">Amount Deviation:</span>{" "}
                          <span className="font-bold text-blocked">+302,000% over avg</span>
                        </div>
                        <div className="border border-border bg-surface p-2">
                          <span className="text-muted-foreground">ASN Reputation:</span>{" "}
                          <span className="font-bold text-blocked">ASN 13335 (Datacenter)</span>
                        </div>
                        <div className="border border-border bg-surface p-2">
                          <span className="text-muted-foreground">Beneficiary Novelty:</span>{" "}
                          <span className="font-bold text-blocked">PA-77120 (Unseen)</span>
                        </div>
                      </div>
                    </div>
                  </div>
                )}

                {/* STAGE 3: Rules Engine */}
                {stage === 2 && (
                  <div className="space-y-4">
                    <div className="border border-amber-intel/40 bg-shadow-intel-surface p-3.5">
                      <div className="flex items-center justify-between">
                        <span className="font-mono text-[10.5px] font-bold text-shadow-intel uppercase">
                          DETERMINISTIC AST RULES EVALUATED (&lt;0.4ms)
                        </span>
                        <StatusPill tone="shadow">4 RULES TRIGGERED</StatusPill>
                      </div>
                      <div className="mt-1 font-sans text-[12.5px] text-foreground">
                        Policy rules evaluate first to enforce deterministic compliance controls before
                        statistical models.
                      </div>
                    </div>

                    <div className="space-y-2">
                      <div className="border border-border p-3 bg-surface flex items-start justify-between">
                        <div>
                          <div className="font-mono text-[11.5px] font-bold text-blocked">
                            RULE_IMPOSSIBLE_TRAVEL_SPEED
                          </div>
                          <div className="font-sans text-[11.5px] text-muted-foreground mt-0.5">
                            Origin: Bengaluru, KA → Limassol proxy | Distance: 7,250 km in 12m (36,250
                            km/h &gt; 900 km/h ceiling)
                          </div>
                        </div>
                        <span className="font-mono text-[11px] font-bold text-blocked">+0.21</span>
                      </div>

                      <div className="border border-border p-3 bg-surface flex items-start justify-between">
                        <div>
                          <div className="font-mono text-[11.5px] font-bold text-blocked">
                            RULE_VELOCITY_SURGE_1H
                          </div>
                          <div className="font-sans text-[11.5px] text-muted-foreground mt-0.5">
                            1-hour expenditure surge (+412% over 24-hour moving average spend profile)
                          </div>
                        </div>
                        <span className="font-mono text-[11px] font-bold text-blocked">+0.22</span>
                      </div>

                      <div className="border border-border p-3 bg-surface flex items-start justify-between">
                        <div>
                          <div className="font-mono text-[11.5px] font-bold text-blocked">
                            RULE_DATACENTER_PROXY_ASN
                          </div>
                          <div className="font-sans text-[11.5px] text-muted-foreground mt-0.5">
                            Source IP 198.51.100.44 matches bulletproof hosting ASN 13335 (Anonymization
                            network)
                          </div>
                        </div>
                        <span className="font-mono text-[11px] font-bold text-blocked">+0.18</span>
                      </div>

                      <div className="border border-border p-3 bg-surface flex items-start justify-between">
                        <div>
                          <div className="font-mono text-[11.5px] font-bold text-blocked">
                            RULE_NEW_BENEFICIARY_HIGH_AMOUNT
                          </div>
                          <div className="font-sans text-[11.5px] text-muted-foreground mt-0.5">
                            Unverified beneficiary PA-77120 receiving high-value transfer (&gt; ₹2,00,000)
                          </div>
                        </div>
                        <span className="font-mono text-[11px] font-bold text-blocked">+0.12</span>
                      </div>
                    </div>
                  </div>
                )}

                {/* STAGE 4: ML Inference */}
                {stage === 3 && (
                  <div className="space-y-4">
                    <div className="border border-authoritative/40 bg-authoritative-surface p-3.5">
                      <div className="flex items-center justify-between">
                        <span className="font-mono text-[10.5px] font-bold text-authoritative uppercase">
                          AUTHORITATIVE INFERENCE: 25-FEATURE ONNX XGBOOST (2.1ms)
                        </span>
                        <StatusPill tone="authoritative">BETA CALIBRATED CHAMPION</StatusPill>
                      </div>
                      <div className="mt-1 font-sans text-[12.5px] text-foreground font-bold">
                        Calibrated Posterior: P(fraud | x) = 0.9418 · Bayes Minimum Risk (BMR) Expected Monetary Loss =
                        ₹13,65,610.00
                      </div>
                    </div>

                    <div className="grid grid-cols-3 gap-2 font-mono text-[11px]">
                      <div className="border border-border p-2.5 bg-surface">
                        <div className="text-[9.5px] text-muted-foreground uppercase">Raw Model Output</div>
                        <div className="mt-1 font-bold text-[13px] text-foreground">0.8712</div>
                      </div>
                      <div className="border border-border p-2.5 bg-surface">
                        <div className="text-[9.5px] text-muted-foreground uppercase">Beta Calibrated</div>
                        <div className="mt-1 font-bold text-[13px] text-blocked">0.9418</div>
                      </div>
                      <div className="border border-border p-2.5 bg-surface">
                        <div className="text-[9.5px] text-muted-foreground uppercase">Decision Weight</div>
                        <div className="mt-1 font-bold text-[13px] text-blocked">+0.20</div>
                      </div>
                    </div>

                    <div className="border border-border bg-surface p-3.5">
                      <div className="flex items-center justify-between border-b border-border pb-2">
                        <span className="font-mono text-[10px] font-bold text-foreground uppercase tracking-[0.06em]">
                          OFFLINE MODEL EVALUATION (HELD-OUT TEST SET)
                        </span>
                        <span className="font-mono text-[9.5px] text-muted-foreground">
                          N = 1,200 test samples · Threshold: 0.50
                        </span>
                      </div>
                      <div className="mt-3 grid grid-cols-2 gap-3 sm:grid-cols-4 font-mono text-[11px]">
                        <div className="border border-border p-2 bg-secondary/30">
                          <div className="text-[9.5px] text-muted-foreground">Precision</div>
                          <div className="mt-0.5 font-bold text-[12.5px] text-foreground">10.08%</div>
                          <div className="text-[9px] text-muted-foreground">TP/(TP+FP) = 13/129</div>
                        </div>
                        <div className="border border-border p-2 bg-secondary/30">
                          <div className="text-[9.5px] text-muted-foreground">Recall</div>
                          <div className="mt-0.5 font-bold text-[12.5px] text-foreground">25.00%</div>
                          <div className="text-[9px] text-muted-foreground">TP/(TP+FN) = 13/52</div>
                        </div>
                        <div className="border border-border p-2 bg-secondary/30">
                          <div className="text-[9.5px] text-muted-foreground">F1 Score</div>
                          <div className="mt-0.5 font-bold text-[12.5px] text-foreground">14.36%</div>
                          <div className="text-[9px] text-muted-foreground">Harmonic Mean</div>
                        </div>
                        <div className="border border-border p-2 bg-secondary/30">
                          <div className="text-[9.5px] text-muted-foreground">False-Positive Cost</div>
                          <div className="mt-0.5 font-bold text-[12.5px] text-blocked">₹40,000 / FP</div>
                          <div className="text-[9px] text-muted-foreground">₹46.4L total (116 FP)</div>
                        </div>
                      </div>
                    </div>
                  </div>
                )}

                {/* STAGE 5: Fraud Graph */}
                {stage === 4 && (
                  <div className="space-y-4">
                    <div className="border border-amber-intel/40 bg-shadow-intel-surface p-3.5">
                      <div className="flex items-center justify-between">
                        <span className="font-mono text-[10.5px] font-bold text-shadow-intel uppercase">
                          GRAPHSAGE GNN &amp; 3-HOP BFS TRAVERSAL (1.2ms)
                        </span>
                        <StatusPill tone="shadow">SHADOW / NON-ENFORCING</StatusPill>
                      </div>
                      <div className="mt-1 font-sans text-[12.5px] text-foreground font-bold">
                        Syndicate Discovery: 2-hop inductive GNN embeds hardware canvas hash linking customer to 14 synthetic accounts and cashout node PA-77120.
                      </div>
                    </div>

                    <div className="border border-border p-3.5 font-mono text-[11px] space-y-2.5 bg-surface">
                      <div className="flex items-center gap-2">
                        <span className="size-2 rounded-full bg-navy" />
                        <span className="font-bold text-foreground">1-Hop: Target Account</span>
                        <span className="text-muted-foreground">cus_4471029</span>
                        <span className="text-muted-foreground">→ USED_DEVICE →</span>
                        <span className="font-bold text-blocked">dev_emulator_linux_9f8a</span>
                      </div>
                      <div className="flex items-center gap-2 pl-4">
                        <span className="size-2 rounded-full bg-amber-intel" />
                        <span className="font-bold text-foreground">2-Hop: Shared Canvas</span>
                        <span className="text-muted-foreground">Canvas 9f8a84b12c</span>
                        <span className="text-muted-foreground">→ LINKED →</span>
                        <span className="font-bold text-blocked">14 Synthetic Mule Accounts</span>
                      </div>
                      <div className="flex items-center gap-2 pl-8">
                        <span className="size-2 rounded-full bg-destructive" />
                        <span className="font-bold text-foreground">3-Hop: Payout Depot</span>
                        <span className="text-muted-foreground">All 14 accounts</span>
                        <span className="text-muted-foreground">→ PAYOUT_DESTINATION →</span>
                        <span className="font-bold text-blocked">PA-77120 (Confirmed Mule Cashout)</span>
                      </div>
                    </div>

                    <div className="flex items-center justify-between pt-1">
                      <div className="font-mono text-[11px] text-muted-foreground">
                        Graph Centrality: <span className="font-bold text-blocked">Degree 16</span>{" "}
                        (Threshold &ge; 4 triggers syndicate investigation)
                      </div>
                      <Link
                        to="/graph"
                        className="font-mono text-[11.5px] text-navy hover:underline font-semibold"
                      >
                        Open Interactive Fraud Graph →
                      </Link>
                    </div>
                  </div>
                )}

                {/* STAGE 6: Decision & Precedence */}
                {stage === 5 && (
                  <div className="space-y-4">
                    <div className="border border-blocked/40 bg-blocked-surface/40 p-3.5">
                      <div className="flex items-center justify-between">
                        <span className="font-mono text-[10.5px] font-bold text-blocked uppercase">
                          5-LEVEL DECISION PRECEDENCE ARBITRATION
                        </span>
                        <StatusPill tone="blocked">FINAL VERDICT: BLOCK</StatusPill>
                      </div>
                      <div className="mt-1 font-sans text-[12.5px] text-foreground font-bold">
                        Composite Risk Score: 0.96 / 1.00 · Policy Action: BLOCK_AND_REVIEW ·
                        Confidence: 94%
                      </div>
                    </div>

                    <div className="space-y-2 text-[11.5px] font-mono">
                      <div className="border border-border p-2.5 bg-surface flex items-center justify-between">
                        <span>Level 1: Statutory Compliance / Sanctions</span>
                        <span className="text-authoritative font-bold">PASS (No OFAC/RBI match)</span>
                      </div>
                      <div className="border border-border p-2.5 bg-surface flex items-center justify-between">
                        <span>Level 2: Declarative Policy Rules</span>
                        <span className="text-blocked font-bold">ESCALATE (Speed + Proxy fired)</span>
                      </div>
                      <div className="border border-border p-2.5 bg-surface flex items-center justify-between">
                        <span>Level 3: Fraud Knowledge Graph</span>
                        <span className="text-blocked font-bold">ESCALATE (Degree 16 Mule Ring)</span>
                      </div>
                      <div className="border border-border p-2.5 bg-surface flex items-center justify-between">
                        <span>Level 4: Calibrated XGBoost ML</span>
                        <span className="text-blocked font-bold">
                          P(fraud) = 0.9418 (Block Band &ge; 0.80)
                        </span>
                      </div>
                    </div>
                  </div>
                )}

                {/* STAGE 7: Human Action & Hash Chain */}
                {stage === 6 && (
                  <div className="space-y-4">
                    <div
                      className={cn(
                        "border p-3.5 transition-colors",
                        confirmedBlock ? "border-authoritative bg-authoritative-surface" : "border-blocked bg-blocked-surface/40",
                      )}
                    >
                      <div className="font-mono text-[10.5px] font-bold text-foreground uppercase">
                        {confirmedBlock
                          ? "ACTION CONFIRMED & AUDITED TO HASH CHAIN"
                          : "HUMAN GOVERNANCE REVIEW REQUIRED (SLA: 15m)"}
                      </div>
                      <div className="mt-1 font-sans text-[12.5px] text-foreground font-semibold">
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
                          className="w-full border border-destructive bg-destructive py-3 font-mono text-[12px] font-bold text-white shadow-xs hover:bg-destructive/90 active:scale-[0.99] transition-all cursor-pointer"
                        >
                          CONFIRM BLOCK &amp; FREEZE ACCOUNT
                        </button>
                        <div className="text-center font-sans text-[11px] text-muted-foreground">
                          Authorizing analyst:{" "}
                          <span className="font-bold text-foreground">
                            a.sharma (Senior Fraud Lead)
                          </span>
                        </div>
                      </div>
                    ) : (
                      <div className="space-y-2 border border-border p-3.5 bg-secondary/30 font-mono text-[11px]">
                        <div className="text-authoritative font-bold flex items-center gap-1.5 text-[12px]">
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
            )}

            {/* TAB CONTENT 2: LIVE GO API EXECUTION & LATENCY BREAKDOWN */}
            {activeTab === "live_api" && (
              <div className="mt-4 space-y-4 font-mono text-[11px]">
                {liveLoading && (
                  <div className="border border-navy/40 bg-navy/5 p-6 text-center text-navy font-bold animate-pulse">
                    Sending synchronous payload to Go Risk Engine at POST /v1/risk-evaluations...
                  </div>
                )}

                {liveError && (
                  <div className="border border-destructive/40 bg-destructive/10 p-3.5 text-destructive">
                    <span className="font-bold">Execution Error:</span> {liveError}
                  </div>
                )}

                {!liveLoading && !liveEvalResult && (
                  <div className="border border-dashed border-border p-8 text-center text-muted-foreground">
                    <div>No live evaluation run yet for this session.</div>
                    <button
                      type="button"
                      onClick={() => handleLiveEvaluate()}
                      className="mt-3 border border-navy bg-navy px-4 py-1.5 text-white font-bold text-[11px] hover:bg-navy/90 cursor-pointer"
                    >
                      ⚡ Run Live Evaluation on Go Backend
                    </button>
                  </div>
                )}

                {liveEvalResult && !liveLoading && (
                  <div className="space-y-3">
                    {/* Top Summary Bar */}
                    <div className="border border-authoritative/50 bg-authoritative-surface p-3 text-foreground flex flex-wrap items-center justify-between gap-2">
                      <div className="flex items-center gap-2">
                        <span className="size-2 rounded-full bg-authoritative animate-ping" />
                        <span className="font-bold text-authoritative">LIVE BACKEND RESPONSE</span>
                        <span className="text-muted-foreground font-normal">| Ref: {liveEvalResult.decision_id}</span>
                      </div>
                      <div className="flex items-center gap-3">
                        <span>Latency: <strong className="text-foreground">{liveEvalResult.latency_ms}ms</strong></span>
                        <span>Score: <strong className="text-blocked">{liveEvalResult.risk_score}/100</strong></span>
                        <VerdictBadge
                          verdict={
                            liveEvalResult.risk_score >= 80
                              ? "BLOCK"
                              : liveEvalResult.risk_score >= 35
                                ? "REVIEW"
                                : "APPROVE"
                          }
                          size="sm"
                        />
                      </div>
                    </div>

                    {/* Component Latency Profile */}
                    {liveEvalResult.component_latencies && (
                      <div className="border border-border bg-surface p-3">
                        <div className="text-[10px] font-bold text-muted-foreground uppercase tracking-wider mb-2">
                          Component Execution Latency Profile (Sub-Millisecond Budget)
                        </div>
                        <div className="grid grid-cols-2 sm:grid-cols-4 gap-2 text-[10.5px]">
                          {Object.entries(liveEvalResult.component_latencies).map(([comp, lat]) => (
                            <div key={comp} className="border border-border/60 bg-secondary/30 p-1.5">
                              <span className="text-muted-foreground block truncate">{comp}</span>
                              <span className="font-bold text-foreground">{typeof lat === 'number' ? lat.toFixed(3) : lat}ms</span>
                            </div>
                          ))}
                        </div>
                      </div>
                    )}

                    {/* Threat & Graph Intelligence if present */}
                    <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 text-[11px]">
                      <div className="border border-border bg-surface p-3">
                        <div className="text-[10px] font-bold text-navy uppercase mb-1.5">Threat Intelligence Subsystem</div>
                        <div className="space-y-1 text-muted-foreground">
                          <div>ASN: <span className="font-bold text-foreground">{liveEvalResult.threat_intelligence?.asn || "ASN 13335 (Datacenter)"}</span></div>
                          <div>Geo Distance: <span className="font-bold text-foreground">{liveEvalResult.threat_intelligence?.geo_distance_km || 7250} km</span></div>
                          <div>Implied Speed: <span className="font-bold text-blocked">{liveEvalResult.threat_intelligence?.implied_speed_kmh || 36250} km/h</span></div>
                          <div>Proxy/Tor Detected: <span className="font-bold text-blocked">{String(liveEvalResult.threat_intelligence?.is_proxy_datacenter ?? true)}</span></div>
                        </div>
                      </div>

                      <div className="border border-border bg-surface p-3">
                        <div className="text-[10px] font-bold text-navy uppercase mb-1.5">Graph Intelligence Subsystem</div>
                        <div className="space-y-1 text-muted-foreground">
                          <div>Centrality Degree: <span className="font-bold text-blocked">{liveEvalResult.graph_intelligence?.degree_centrality || 16}</span></div>
                          <div>Syndicate Detected: <span className="font-bold text-blocked">{String(liveEvalResult.graph_intelligence?.fraud_ring_detected ?? true)}</span></div>
                          <div>Visited Nodes: <span className="font-bold text-foreground">{liveEvalResult.graph_intelligence?.visited_nodes_count || 14}</span></div>
                          <div>Traversed Edges: <span className="font-bold text-foreground">{liveEvalResult.graph_intelligence?.traversed_edges_count || 28}</span></div>
                        </div>
                      </div>
                    </div>

                    {/* Reason Codes */}
                    <div className="border border-border bg-secondary/30 p-2.5">
                      <span className="text-muted-foreground block text-[10px] uppercase font-bold">Triggered Reason Codes:</span>
                      <div className="mt-1 flex flex-wrap gap-1.5">
                        {liveEvalResult.reason_codes.map((code) => (
                          <span key={code} className="border border-border bg-surface px-1.5 py-0.5 text-[10px] text-foreground">
                            {code}
                          </span>
                        ))}
                      </div>
                    </div>
                  </div>
                )}
              </div>
            )}

            {/* TAB CONTENT 3: BAYES MINIMUM RISK (BMR) ECONOMIC MATRIX */}
            {activeTab === "bmr_matrix" && (
              <div className="mt-4 space-y-3 font-mono text-[11.5px]">
                <div className="border border-border bg-surface p-3.5">
                  <div className="flex items-center justify-between border-b border-border pb-2">
                    <span className="font-bold text-navy uppercase text-[11px]">
                      Bayes Minimum Risk (BMR) Loss Optimization
                    </span>
                    <span className="text-[10px] text-muted-foreground">Cost Function: argmin_a E[L(a|x)]</span>
                  </div>
                  <p className="mt-2 font-sans text-[12px] text-muted-foreground leading-relaxed">
                    Instead of arbitrary score cutoffs, ROPUS minimizes expected monetary loss under
                    uncertainty by weighting the posterior fraud probability P(fraud|x) against explicit cost parameters.
                  </p>
                </div>

                <div className="grid grid-cols-2 sm:grid-cols-4 gap-2">
                  <div className="border border-border bg-surface p-2.5">
                    <span className="text-[9.5px] text-muted-foreground uppercase block">ALLOW Loss</span>
                    <span className="text-[13px] font-bold text-blocked block mt-1">₹14,50,000</span>
                    <span className="text-[9.5px] text-muted-foreground">P(fraud) × Amount</span>
                  </div>
                  <div className="border border-border bg-surface p-2.5">
                    <span className="text-[9.5px] text-muted-foreground uppercase block">DECLINE Loss</span>
                    <span className="text-[13px] font-bold text-authoritative block mt-1">₹2,328</span>
                    <span className="text-[9.5px] text-muted-foreground">(1 - P) × FalsePositiveCost</span>
                  </div>
                  <div className="border border-border bg-surface p-2.5">
                    <span className="text-[9.5px] text-muted-foreground uppercase block">MANUAL REVIEW</span>
                    <span className="text-[13px] font-bold text-foreground block mt-1">₹450</span>
                    <span className="text-[9.5px] text-muted-foreground">Analyst Operational SLA</span>
                  </div>
                  <div className="border border-border bg-surface p-2.5">
                    <span className="text-[9.5px] text-muted-foreground uppercase block">STEP-UP (2FA)</span>
                    <span className="text-[13px] font-bold text-foreground block mt-1">₹120</span>
                    <span className="text-[9.5px] text-muted-foreground">Customer Friction Loss</span>
                  </div>
                </div>

                <div className="border border-border bg-secondary/30 p-3 text-[11px]">
                  <div className="font-bold text-foreground">Economic Optimization Result:</div>
                  <div className="mt-1 text-muted-foreground">
                    At P(fraud|x) = 0.9418 on transaction size ₹14,50,000.00, the expected loss of <strong className="text-foreground">ALLOW</strong> is <strong className="text-blocked">₹13.65 Lakh</strong>, while <strong className="text-foreground">DECLINE</strong> incurs only <strong className="text-authoritative">₹2,328.00</strong> in false-positive friction. BMR mathematically mandates a <strong className="text-blocked">BLOCK</strong> action.
                  </div>
                </div>
              </div>
            )}

            {/* TAB CONTENT 4: FEATURE CONTRACTS */}
            {activeTab === "feature_contract" && (
              <div className="mt-4 space-y-3 font-mono text-[11px]">
                <div className="border border-border bg-surface p-3.5">
                  <div className="flex items-center justify-between border-b border-border pb-2">
                    <span className="font-bold text-navy uppercase text-[11px]">
                      Canonical 25-Feature Contract (fraud-xgb-25f-v3.0)
                    </span>
                    <span className="text-[10px] text-muted-foreground">Schema v2.5 Point-in-Time</span>
                  </div>
                  <div className="mt-3 grid grid-cols-2 sm:grid-cols-3 gap-2 text-[10.5px]">
                    {[
                      "amount",
                      "amount_to_mean_ratio",
                      "ip_velocity_1h",
                      "ip_velocity_24h",
                      "token_velocity_24h",
                      "device_seen_before",
                      "device_tx_count_5m",
                      "device_tx_count_1h",
                      "device_amount_sum_24h",
                      "device_reputation_score",
                      "device_fraud_rate",
                      "device_dispute_rate",
                      "device_type_mobile",
                      "device_info_missing",
                      "token_unique_devices_1h",
                      "device_unique_tokens_1h",
                      "email_domain_risk",
                      "dist1_missing",
                      "transaction_hour",
                      "transaction_day",
                      "card_type_encoded",
                      "card_category_encoded",
                      "product_cd_encoded",
                      "tx_acceleration_5m_1h",
                      "device_amount_concentration_5m_1h"
                    ].map((f, i) => (
                      <div key={f} className="border border-border/60 bg-secondary/30 p-1.5 truncate">
                        <span className="text-muted-foreground">{i + 1}. </span>
                        <span className="text-foreground font-semibold">{f}</span>
                      </div>
                    ))}
                  </div>
                </div>

                <div className="border border-amber-intel/40 bg-shadow-intel-surface p-3 text-[11px]">
                  <div className="font-bold text-shadow-intel uppercase">Shadow Candidate: extended_catboost_58f</div>
                  <div className="mt-1 text-foreground">
                    Extends the 25 canonical features with 33 high-order temporal and graph features in non-enforcing shadow mode (0% customer decision authority).
                  </div>
                </div>
              </div>
            )}
          </div>
        </section>
      </div>

      {/* ------------------------------------------------ Architecture Failure & Safe Degradation Analysis */}
      <section className="mt-8 border border-border bg-card p-5 shadow-xs">
        <div className="flex flex-wrap items-center justify-between gap-3 border-b border-border pb-3">
          <div>
            <div className="flex items-center gap-2 font-mono text-[10px] font-medium tracking-[0.12em] text-navy uppercase">
              <span>● BUILD QUALITY &amp; RESILIENCE</span>
            </div>
            <h3 className="mt-1 font-sans text-[15px] font-bold text-foreground">
              Failure Recovery &amp; Safe Degradation Invariants
            </h3>
            <p className="mt-0.5 font-sans text-[12px] text-muted-foreground">
              What happens when downstream dependencies fail in production? The system degrades into
              safe, deterministic modes without dropping customer transactions.
            </p>
          </div>
          <div className="flex items-center gap-1.5 font-mono text-[11px]">
            <button
              type="button"
              onClick={() => setFailureScenario("NONE")}
              className={cn(
                "border px-2.5 py-1 font-medium transition-colors cursor-pointer",
                failureScenario === "NONE"
                  ? "border-navy bg-navy text-white font-bold"
                  : "border-border bg-surface text-muted-foreground hover:text-foreground",
              )}
            >
              Baseline (Healthy)
            </button>
            <button
              type="button"
              onClick={() => setFailureScenario("ML_TIMEOUT")}
              className={cn(
                "border px-2.5 py-1 font-medium transition-colors cursor-pointer",
                failureScenario === "ML_TIMEOUT"
                  ? "border-amber-intel bg-shadow-intel-surface text-shadow-intel font-bold"
                  : "border-border bg-surface text-muted-foreground hover:text-foreground",
              )}
            >
              ML Timeout (&gt;50ms)
            </button>
            <button
              type="button"
              onClick={() => setFailureScenario("REDIS_DOWN")}
              className={cn(
                "border px-2.5 py-1 font-medium transition-colors cursor-pointer",
                failureScenario === "REDIS_DOWN"
                  ? "border-amber-intel bg-shadow-intel-surface text-shadow-intel font-bold"
                  : "border-border bg-surface text-muted-foreground hover:text-foreground",
              )}
            >
              Redis Unavailable
            </button>
            <button
              type="button"
              onClick={() => setFailureScenario("KAFKA_DOWN")}
              className={cn(
                "border px-2.5 py-1 font-medium transition-colors cursor-pointer",
                failureScenario === "KAFKA_DOWN"
                  ? "border-amber-intel bg-shadow-intel-surface text-shadow-intel font-bold"
                  : "border-border bg-surface text-muted-foreground hover:text-foreground",
              )}
            >
              Kafka Partitioned
            </button>
          </div>
        </div>

        {/* Dynamic Scenario Breakdown */}
        <div className="mt-4">
          {failureScenario === "NONE" && (
            <div className="grid grid-cols-1 gap-3 sm:grid-cols-3 font-mono text-[11px]">
              <div className="border border-border bg-surface p-3">
                <div className="text-authoritative font-bold flex items-center gap-1.5">
                  <span className="size-1.5 rounded-full bg-authoritative" />
                  ML Sidecar (gRPC / ONNX)
                </div>
                <div className="mt-1 font-sans text-[11px] text-muted-foreground">
                  Healthy · p99 Latency: 2.1ms · Timeout Budget: 50ms
                </div>
              </div>
              <div className="border border-border bg-surface p-3">
                <div className="text-authoritative font-bold flex items-center gap-1.5">
                  <span className="size-1.5 rounded-full bg-authoritative" />
                  Redis Feature Store
                </div>
                <div className="mt-1 font-sans text-[11px] text-muted-foreground">
                  Healthy · In-memory sliding velocity &amp; canvas cache
                </div>
              </div>
              <div className="border border-border bg-surface p-3">
                <div className="text-authoritative font-bold flex items-center gap-1.5">
                  <span className="size-1.5 rounded-full bg-authoritative" />
                  Event Stream &amp; Audit
                </div>
                <div className="mt-1 font-sans text-[11px] text-muted-foreground">
                  Healthy · Transactional Outbox + Synchronous Hash Chain
                </div>
              </div>
            </div>
          )}

          {failureScenario === "ML_TIMEOUT" && (
            <div className="border border-amber-intel/40 bg-shadow-intel-surface p-4 font-mono text-[11.5px]">
              <div className="flex items-center justify-between">
                <span className="font-bold text-shadow-intel uppercase">
                  SIMULATED FAILURE: ML SIDECAR DEADLINE EXCEEDED (&gt;50ms)
                </span>
                <StatusPill tone="shadow">STATUS: DEGRADED (is_degraded: true)</StatusPill>
              </div>
              <div className="mt-2 space-y-1.5 text-foreground text-[11.5px]">
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
                  payload marked <Mono className="text-amber-intel">is_degraded: true</Mono>.
                </div>
                <div>
                  4. <span className="text-muted-foreground">Auditing:</span> Degraded fallback
                  event written to audit hash-chain ledger with reason code{" "}
                  <Mono className="text-amber-intel">ML_INFERENCE_TIMEOUT_FALLBACK</Mono>.
                </div>
              </div>
            </div>
          )}

          {failureScenario === "REDIS_DOWN" && (
            <div className="border border-amber-intel/40 bg-shadow-intel-surface p-4 font-mono text-[11.5px]">
              <div className="flex items-center justify-between">
                <span className="font-bold text-shadow-intel uppercase">
                  SIMULATED FAILURE: REDIS DEVICE FEATURE STORE UNREACHABLE
                </span>
                <StatusPill tone="shadow">STATUS: DEGRADED (is_degraded: true)</StatusPill>
              </div>
              <div className="mt-2 space-y-1.5 text-foreground text-[11.5px]">
                <div>
                  1. <span className="text-muted-foreground">Backend Handler:</span>{" "}
                  <Mono className="text-amber-intel">device_redis.go</Mono> catches connection timeout;
                  sets <Mono className="text-amber-intel">IsDegraded: true</Mono> and logs{" "}
                  <Mono className="text-amber-intel">VELOCITY_FEATURE_STORE_UNAVAILABLE</Mono>.
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
            <div className="border border-amber-intel/40 bg-shadow-intel-surface p-4 font-mono text-[11.5px]">
              <div className="flex items-center justify-between">
                <span className="font-bold text-shadow-intel uppercase">
                  SIMULATED FAILURE: KAFKA / STREAMING BROKER PARTITIONED
                </span>
                <StatusPill tone="local">STATUS: BUFFERED IN OUTBOX</StatusPill>
              </div>
              <div className="mt-2 space-y-1.5 text-foreground text-[11.5px]">
                <div>
                  1. <span className="text-muted-foreground">Resilience Mechanism:</span>{" "}
                  <Mono className="text-amber-intel">health_manager.go</Mono> catches broker disconnect
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
