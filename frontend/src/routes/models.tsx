import { createFileRoute } from "@tanstack/react-router";
import { useState, useMemo } from "react";
import {
  DataGrid,
  Page,
  PageHead,
  SectionHead,
  TelemetryStrip,
} from "@/components/ropus/page";
import { Mono, StatusPill } from "@/components/ropus/core";
import { models, modelDetails, type ModelRecord } from "@/lib/ropus/platform-fixtures";
import { blockedDecision } from "@/lib/ropus/fixtures";
import { cn } from "@/lib/utils";
import {
  Cpu,
  Sliders,
  TrendingDown,
  ShieldCheck,
  AlertTriangle,
  Flame,
  CheckCircle2,
  Lock,
  RefreshCw,
  GitBranch,
} from "lucide-react";
import {
  ResponsiveContainer,
  LineChart,
  Line,
  XAxis,
  YAxis,
  Tooltip,
  CartesianGrid,
  ReferenceLine,
} from "recharts";
import { toast } from "sonner";

export const Route = createFileRoute("/models")({
  head: () => ({
    meta: [
      { title: "Models & BMR Optimizer — ROPUS Risk Control Plane" },
      {
        name: "description",
        content:
          "Deployed risk models, Bayes Minimum Risk (BMR) economic cost optimizer, canary traffic routing, feature drift, and GraphSAGE shadow evaluation.",
      },
      { property: "og:title", content: "Models & BMR Optimizer — ROPUS" },
      {
        property: "og:description",
        content: "Model registry with stage, AUC, drift, latency, and Bayes Minimum Risk loss matrix simulator.",
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

export function ModelsPage() {
  const [selectedKey, setSelectedKey] = useState<string | null>("mdl_wire_risk@v4.2.1");
  const [compareKey, setCompareKey] = useState<string | null>("mdl_wire_risk@v4.3.0-rc2");

  // BMR Simulator State
  const [costFp, setCostFp] = useState<number>(40000); // False positive friction (in INR, e.g. ₹40,000)
  const [costFn, setCostFn] = useState<number>(1450000); // False negative / fraud loss (in INR, e.g. ₹14,50,000)
  const [canaryPercent, setCanaryPercent] = useState<number>(0);
  const [canaryLocked, setCanaryLocked] = useState<boolean>(true);

  // Compute Optimal BMR Cutoff: tau* = C_FP / (C_FP + C_FN)
  const optimalTau = useMemo(() => {
    return costFp / (costFp + costFn);
  }, [costFp, costFn]);

  // Generate simulated Cost Curve data points across threshold continuum [0.01 to 0.99]
  const costCurveData = useMemo(() => {
    const points = [];
    for (let t = 0.01; t <= 0.99; t += 0.02) {
      // Estimated FPR and FNR based on calibrated XGBoost validation curve
      const fpr = Math.exp(-6 * t);
      const fnr = 1 - Math.exp(-0.8 * (1 - t));
      const expectedLoss = fpr * costFp * 500 + fnr * costFn * 25; // per 10k transactions
      points.push({
        threshold: Number(t.toFixed(2)),
        loss: Math.round(expectedLoss / 1000), // in Thousands INR
        fpr: Number((fpr * 100).toFixed(2)),
        fnr: Number((fnr * 100).toFixed(2)),
      });
    }
    return points;
  }, [costFp, costFn]);

  const production = models.filter((m) => m.stage === "PRODUCTION");
  const selected = models.find((m) => key(m) === selectedKey) ?? models[0]!;
  const detail = modelDetails[key(selected)] ?? modelDetails["mdl_wire_risk@v4.2.1"];

  const handleApplyCanary = () => {
    toast.success(`Canary traffic updated to ${canaryPercent}%`, {
      description: "Shadow candidate received updated routing weight with dual-officer maker-checker validation.",
    });
  };

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
          { label: "Serving p99", value: "19.6 ms", sub: "fraud-xgb-25f-v3.0" },
          {
            label: "Max Drift (PSI)",
            value: Math.max(...production.map((m) => m.driftPsi)).toFixed(2),
            sub: "Stable baseline (<0.10)",
            tone: "text-authoritative",
          },
          { label: "Canary Decision Share", value: `${canaryPercent}%`, sub: "0% Shadow Authority", tone: "text-shadow-intel" },
        ]}
      />

      {/* ------------------------------------------------ BAYES MINIMUM RISK (BMR) OPTIMIZER */}
      <section className="mt-6 border border-border bg-card p-5 shadow-xs space-y-5">
        <div className="flex flex-wrap items-center justify-between border-b border-border pb-3 gap-2">
          <div className="flex items-center gap-2">
            <span className="border border-navy bg-navy/10 px-2 py-0.5 font-mono text-[10px] font-bold text-navy uppercase tracking-[0.06em]">
              ECONOMIC DECISION SIMULATOR
            </span>
            <h2 className="font-sans text-[17px] font-bold text-foreground">
              Bayes Minimum Risk (BMR) Expected Monetary Loss Optimizer
            </h2>
          </div>
          <div className="font-mono text-[11px] text-muted-foreground">
            Loss Matrix Optimization: <Mono className="text-foreground font-bold">L(a, y) = E[Cost]</Mono>
          </div>
        </div>

        <div className="grid gap-6 lg:grid-cols-[380px_1fr]">
          {/* Sliders & Parameters */}
          <div className="space-y-4 font-mono text-[11.5px] border-r border-border/60 pr-4">
            <div className="border border-border bg-surface p-3 space-y-2">
              <div className="flex justify-between items-center text-[10.5px]">
                <span className="text-muted-foreground uppercase font-bold">Cost of False Positive (C_FP)</span>
                <span className="font-bold text-navy">₹{costFp.toLocaleString("en-IN")} INR</span>
              </div>
              <input
                type="range"
                min="5000"
                max="200000"
                step="5000"
                value={costFp}
                onChange={(e) => setCostFp(Number(e.target.value))}
                className="w-full accent-navy cursor-pointer"
              />
              <div className="text-[10px] text-muted-foreground font-sans">
                Customer friction, brand churn, manual support escalation ticket cost.
              </div>
            </div>

            <div className="border border-border bg-surface p-3 space-y-2">
              <div className="flex justify-between items-center text-[10.5px]">
                <span className="text-muted-foreground uppercase font-bold">Cost of False Negative (C_FN)</span>
                <span className="font-bold text-blocked">₹{costFn.toLocaleString("en-IN")} INR</span>
              </div>
              <input
                type="range"
                min="100000"
                max="5000000"
                step="50000"
                value={costFn}
                onChange={(e) => setCostFn(Number(e.target.value))}
                className="w-full accent-blocked cursor-pointer"
              />
              <div className="text-[10px] text-muted-foreground font-sans">
                Direct fraud chargeback loss, interchange scheme fine, recovery loss.
              </div>
            </div>

            {/* Computed Optimal Output */}
            <div className="border border-navy/40 bg-navy/5 p-3 space-y-2">
              <div className="text-[10px] text-navy font-bold uppercase tracking-wider">
                Analytically Optimal Cutoff Threshold (τ*)
              </div>
              <div className="flex items-baseline justify-between">
                <span className="text-[24px] font-extrabold text-navy">{optimalTau.toFixed(4)}</span>
                <span className="text-[11px] text-muted-foreground">P(fraud) ≥ {(optimalTau * 100).toFixed(2)}% → BLOCK</span>
              </div>
              <div className="text-[10.5px] font-sans text-muted-foreground leading-snug">
                Formula: <Mono className="text-navy font-bold">τ* = C_FP / (C_FP + C_FN)</Mono>. Minimizes expected rupee loss across entire transaction portfolio.
              </div>
            </div>
          </div>

          {/* Recharts Expected Rupee Loss Curve */}
          <div className="space-y-2">
            <div className="flex items-center justify-between text-[11px] font-mono text-muted-foreground">
              <span>EXPECTED PORTFOLIO LOSS vs. DECISION CUTOFF THRESHOLD</span>
              <span className="text-navy font-bold">Optimal Point marked at τ* = {optimalTau.toFixed(4)}</span>
            </div>
            <div className="h-[220px] w-full border border-border/50 bg-surface/50 p-2">
              <ResponsiveContainer width="100%" height="100%">
                <LineChart data={costCurveData} margin={{ top: 10, right: 20, left: 10, bottom: 5 }}>
                  <CartesianGrid strokeDasharray="3 3" stroke="#e2ded4" />
                  <XAxis dataKey="threshold" stroke="#71717a" fontSize={10} fontStyle="mono" />
                  <YAxis stroke="#71717a" fontSize={10} fontStyle="mono" />
                  <Tooltip
                    contentStyle={{ backgroundColor: "#1C2B45", color: "#fff", border: "none", fontSize: "11px", fontFamily: "DM Mono" }}
                    formatter={(val: any) => [`₹${Number(val).toLocaleString("en-IN")}k`, "Expected Loss"]}
                    labelFormatter={(lbl) => `Threshold τ: ${lbl}`}
                  />
                  <ReferenceLine x={Number(optimalTau.toFixed(2))} stroke="#1C2B45" strokeWidth={2} strokeDasharray="4 4" label={{ value: "τ*", fill: "#1C2B45", fontSize: 12, position: "top" }} />
                  <Line type="monotone" dataKey="loss" stroke="#A0443A" strokeWidth={2} dot={false} />
                </LineChart>
              </ResponsiveContainer>
            </div>
          </div>
        </div>
      </section>

      {/* ------------------------------------------------ CANARY ROUTING & REGISTRY */}
      <div className="mt-6 grid gap-6 xl:grid-cols-[minmax(0,1fr)_360px] xl:gap-8">
        <section className="min-w-0 space-y-4">
          <SectionHead title="MODEL REGISTRY" meta={`${models.length} registered models`} />
          <div className="overflow-x-auto border border-border bg-card">
            <DataGrid
              onSelect={(id) => {
                setSelectedKey((cur) => (cur === id ? null : id));
              }}
              selectedId={selectedKey}
              rowClassName="py-2.5"
              columns={[
                { key: "id", label: "Model" },
                { key: "name", label: "Purpose" },
                { key: "version", label: "Version" },
                { key: "stage", label: "Stage" },
                { key: "auc", label: "Val PR-AUC" },
                { key: "driftPsi", label: "PSI Drift" },
                { key: "p99Ms", label: "p99 Latency" },
              ]}
              rows={models.map((m) => ({
                id: key(m),
                cells: [
                  <span key="id" className="font-mono text-[12px] font-bold text-navy">{m.id}</span>,
                  <span key="name" className="text-[12px] font-sans">{m.name}</span>,
                  <Mono key="version" className="text-[11.5px]">{m.version}</Mono>,
                  <StageText key="stage" stage={m.stage} />,
                  <span key="auc" className="font-mono text-[12px] font-semibold tabular">
                    {m.auc.toFixed(4)}
                  </span>,
                  <span
                    key="psi"
                    className={cn(
                      "font-mono text-[12px] tabular",
                      m.driftPsi < 0.1 ? "text-authoritative font-bold" : "text-amber-intel",
                    )}
                  >
                    {m.driftPsi.toFixed(2)}
                  </span>,
                  <span key="lat" className="font-mono text-[12px] tabular">
                    {m.p99Ms.toFixed(1)} ms
                  </span>,
                ],
              }))}
            />
          </div>

          {/* Canary Routing Controls */}
          <div className="border border-border bg-card p-4 space-y-3 font-mono text-[11.5px]">
            <div className="flex items-center justify-between border-b border-border pb-2">
              <div className="flex items-center gap-2">
                <GitBranch className="size-4 text-navy" />
                <span className="font-bold text-foreground text-[12px]">
                  Canary Traffic Routing: Active Champion vs Shadow Candidate
                </span>
              </div>
              <button
                type="button"
                onClick={() => setCanaryLocked(!canaryLocked)}
                className="flex items-center gap-1 text-[10px] text-muted-foreground hover:text-foreground cursor-pointer"
              >
                <Lock className="size-3" />
                <span>{canaryLocked ? "Unlock Slider (Maker-Checker)" : "Lock Configuration"}</span>
              </button>
            </div>

            <div className="grid grid-cols-2 gap-4 text-[11px]">
              <div className="border border-authoritative/40 bg-approve-surface p-2.5">
                <div className="text-[10px] text-authoritative uppercase font-bold">Active Champion (100% - {canaryPercent}%)</div>
                <div className="font-bold text-foreground text-[12px] mt-0.5">fraud-xgb-25f-v3.0</div>
                <div className="text-muted-foreground text-[10.5px]">25 Canonical Features · Beta Calibrated</div>
              </div>
              <div className="border border-shadow-intel/40 bg-shadow-intel-surface p-2.5">
                <div className="text-[10px] text-shadow-intel uppercase font-bold">Shadow Candidate ({canaryPercent}%)</div>
                <div className="font-bold text-foreground text-[12px] mt-0.5">extended_catboost_58f</div>
                <div className="text-muted-foreground text-[10.5px]">58 Features · 0% Customer Authority (Gated)</div>
              </div>
            </div>

            <div className="space-y-1.5 pt-1">
              <div className="flex justify-between text-[10.5px]">
                <span className="text-muted-foreground">Canary Traffic Split:</span>
                <span className="font-bold text-navy">{canaryPercent}% routed to Shadow Validation</span>
              </div>
              <input
                type="range"
                min="0"
                max="50"
                step="5"
                disabled={canaryLocked}
                value={canaryPercent}
                onChange={(e) => setCanaryPercent(Number(e.target.value))}
                className="w-full accent-navy cursor-pointer disabled:opacity-40"
              />
            </div>

            <div className="flex justify-end gap-2 pt-1">
              <button
                type="button"
                disabled={canaryLocked || canaryPercent === 0}
                onClick={handleApplyCanary}
                className="border border-navy bg-navy px-3 py-1 font-bold text-white text-[10.5px] hover:bg-navy/90 cursor-pointer disabled:opacity-40 font-mono"
              >
                Apply Canary Split
              </button>
            </div>
          </div>
        </section>

        {/* Right Sidebar: Selected Model Metadata */}
        <aside className="border border-border bg-card p-5 shadow-xs space-y-4 font-mono text-[11.5px]">
          <div className="flex items-center justify-between border-b border-border pb-2">
            <span className="font-bold text-navy text-[11px] uppercase tracking-wider">
              MODEL GOVERNANCE DOSSIER
            </span>
            <StageText stage={selected.stage} />
          </div>

          <div className="space-y-2">
            <div className="text-[13px] font-bold text-foreground">{selected.id}</div>
            <div className="text-[11px] text-muted-foreground font-sans">{selected.name}</div>
          </div>

          <div className="border-t border-border pt-3 space-y-2">
            <div className="flex justify-between">
              <span className="text-muted-foreground">Version:</span>
              <span className="font-bold text-foreground">{selected.version}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-muted-foreground">Validation PR-AUC:</span>
              <span className="font-bold text-foreground">{selected.auc.toFixed(4)}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-muted-foreground">Serving Latency p99:</span>
              <span className="font-bold text-foreground">{selected.p99Ms.toFixed(1)} ms</span>
            </div>
            <div className="flex justify-between">
              <span className="text-muted-foreground">PSI Feature Drift:</span>
              <span className={cn("font-bold", selected.driftPsi < 0.1 ? "text-authoritative" : "text-amber-intel")}>
                {selected.driftPsi.toFixed(2)}
              </span>
            </div>
            <div className="flex justify-between">
              <span className="text-muted-foreground">Trained Date:</span>
              <span className="text-foreground">{selected.trainedOn}</span>
            </div>
            {detail && (
              <>
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Serving Alias:</span>
                  <span className="font-bold text-navy">{detail.servingAlias}</span>
                </div>
                {detail.promotedOn && (
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">Promoted On:</span>
                    <span className="text-foreground">{detail.promotedOn}</span>
                  </div>
                )}
              </>
            )}
          </div>
        </aside>
      </div>
    </Page>
  );
}
