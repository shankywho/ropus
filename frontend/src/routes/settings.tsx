import { createFileRoute } from "@tanstack/react-router";
import { useState } from "react";
import { Page, PageHead, SectionHead, TelemetryStrip } from "@/components/ropus/page";
import { Mono, StatusPill } from "@/components/ropus/core";
import { tenantSettings } from "@/lib/ropus/platform-fixtures";
import { cn } from "@/lib/utils";
import { Sliders, ShieldCheck, Clock, Save, Lock, RotateCcw } from "lucide-react";
import { toast } from "sonner";

export const Route = createFileRoute("/settings")({
  head: () => ({
    meta: [
      { title: "Tenant Settings & Policy Thresholds — ROPUS" },
      {
        name: "description",
        content:
          "Tenant configuration, risk verdict cutoff thresholds, fallback timeout behavior, and Maker-Checker governance.",
      },
      { property: "og:title", content: "Settings — ROPUS" },
      {
        property: "og:description",
        content: "Thresholds, fallback behavior and retention configuration for this tenant.",
      },
      { property: "og:type", content: "website" },
      { name: "twitter:card", content: "summary" },
    ],
  }),
  component: SettingsPage,
});

export function SettingsPage() {
  const [blockThreshold, setBlockThreshold] = useState<number>(0.80);
  const [reviewThreshold, setReviewThreshold] = useState<number>(0.50);
  const [challengeThreshold, setChallengeThreshold] = useState<number>(0.30);
  const [makerCheckerEnforced, setMakerCheckerEnforced] = useState<boolean>(true);
  const [timeoutBudgetMs, setTimeoutBudgetMs] = useState<number>(100);
  const [saving, setSaving] = useState<boolean>(false);

  const handleSaveSettings = () => {
    setSaving(true);
    setTimeout(() => {
      setSaving(false);
      toast.success("Tenant configuration committed to audit ledger", {
        description: "New threshold cutoffs will take effect on next evaluation (Block: ≥" + blockThreshold.toFixed(2) + ", Review: ≥" + reviewThreshold.toFixed(2) + ").",
      });
    }, 400);
  };

  return (
    <Page>
      <PageHead
        title="Tenant Settings & Governance Policy"
        subtitle="Institutional risk configuration, decision threshold arbitration, sub-100ms timeout fallbacks, and Maker-Checker multi-party dual control."
      />

      <TelemetryStrip
        items={[
          { label: "Active Policy", value: "pol_wire_outbound_v8", sub: "Production Champion", tone: "text-navy" },
          { label: "Block Cutoff (τ*)", value: blockThreshold.toFixed(2), sub: "BMR Cost Optimal", tone: "text-blocked" },
          { label: "Review Cutoff", value: reviewThreshold.toFixed(2), sub: "24h SLA Case Queue", tone: "text-amber-intel" },
          { label: "Timeout Budget", value: `${timeoutBudgetMs} ms`, sub: "Hard deadline", tone: "text-authoritative" },
          { label: "Dual Control", value: makerCheckerEnforced ? "ENFORCED" : "OFF", sub: "Maker-Checker Policy", tone: makerCheckerEnforced ? "text-authoritative" : "text-amber-intel" },
        ]}
      />

      <div className="mt-6 grid gap-8 xl:grid-cols-2 xl:gap-10">
        {/* Left Column: Decision Thresholds Interactive Tuner */}
        <section className="border border-border bg-card p-5 shadow-xs space-y-5 font-mono text-[11.5px]">
          <div className="flex items-center justify-between border-b border-border pb-3">
            <div className="flex items-center gap-2">
              <Sliders className="size-4 text-navy" />
              <span className="font-bold text-foreground text-[13px] font-sans">
                Verdict Cutoff Thresholds (0.00 – 1.00)
              </span>
            </div>
            <StatusPill tone="authoritative">REAL-TIME APPLIED</StatusPill>
          </div>

          <div className="space-y-4">
            {/* Visual Threshold Range Bar */}
            <div className="space-y-1">
              <div className="flex h-[12px] w-full overflow-hidden border border-border">
                <div
                  className="bg-approve h-full"
                  style={{ width: `${challengeThreshold * 100}%` }}
                  title="APPROVE Range"
                />
                <div
                  className="bg-primary h-full"
                  style={{ width: `${(reviewThreshold - challengeThreshold) * 100}%` }}
                  title="CHALLENGE Range"
                />
                <div
                  className="bg-amber-intel h-full"
                  style={{ width: `${(blockThreshold - reviewThreshold) * 100}%` }}
                  title="REVIEW Range"
                />
                <div
                  className="bg-blocked h-full"
                  style={{ width: `${(1 - blockThreshold) * 100}%` }}
                  title="BLOCK Range"
                />
              </div>

              <div className="flex justify-between text-[10px] text-muted-foreground font-mono">
                <span>0.00 (APPROVE)</span>
                <span>{challengeThreshold.toFixed(2)} (2FA)</span>
                <span>{reviewThreshold.toFixed(2)} (REVIEW)</span>
                <span>{blockThreshold.toFixed(2)} (BLOCK)</span>
                <span>1.00</span>
              </div>
            </div>

            {/* Block Threshold Slider */}
            <div className="border border-border bg-surface p-3 space-y-1.5 shadow-2xs">
              <div className="flex justify-between items-center">
                <span className="font-bold text-blocked uppercase text-[10.5px]">1. Hard Block Cutoff (P(fraud) ≥ τ*)</span>
                <span className="font-bold text-blocked text-[13px]">{blockThreshold.toFixed(2)}</span>
              </div>
              <input
                type="range"
                min="0.60"
                max="0.95"
                step="0.01"
                value={blockThreshold}
                onChange={(e) => setBlockThreshold(Number(e.target.value))}
                className="w-full accent-blocked cursor-pointer"
              />
              <div className="text-[11px] text-muted-foreground font-sans leading-snug">
                Transactions scoring above {blockThreshold.toFixed(2)} are refused at payment gateway with statutory reason code.
              </div>
            </div>

            {/* Review Threshold Slider */}
            <div className="border border-border bg-surface p-3 space-y-1.5 shadow-2xs">
              <div className="flex justify-between items-center">
                <span className="font-bold text-amber-intel uppercase text-[10.5px]">2. Analyst Review Cutoff</span>
                <span className="font-bold text-amber-intel text-[13px]">{reviewThreshold.toFixed(2)}</span>
              </div>
              <input
                type="range"
                min="0.40"
                max="0.75"
                step="0.01"
                value={reviewThreshold}
                onChange={(e) => setReviewThreshold(Number(e.target.value))}
                className="w-full accent-amber-intel cursor-pointer"
              />
              <div className="text-[11px] text-muted-foreground font-sans leading-snug">
                Scores between {reviewThreshold.toFixed(2)} and {blockThreshold.toFixed(2)} provision a case in the 24-hour SLA analyst queue.
              </div>
            </div>

            {/* Step-Up Challenge Slider */}
            <div className="border border-border bg-surface p-3 space-y-1.5 shadow-2xs">
              <div className="flex justify-between items-center">
                <span className="font-bold text-navy uppercase text-[10.5px]">3. Step-Up 2FA Challenge Cutoff</span>
                <span className="font-bold text-navy text-[13px]">{challengeThreshold.toFixed(2)}</span>
              </div>
              <input
                type="range"
                min="0.20"
                max="0.45"
                step="0.01"
                value={challengeThreshold}
                onChange={(e) => setChallengeThreshold(Number(e.target.value))}
                className="w-full accent-navy cursor-pointer"
              />
              <div className="text-[11px] text-muted-foreground font-sans leading-snug">
                Scores between {challengeThreshold.toFixed(2)} and {reviewThreshold.toFixed(2)} request dynamic WebAuthn / OTP step-up authentication.
              </div>
            </div>
          </div>
        </section>

        {/* Right Column: Governance & Runtime Settings */}
        <section className="space-y-6 font-mono text-[11.5px]">
          {/* Runtime & Governance Card */}
          <div className="border border-border bg-card p-5 shadow-xs space-y-4">
            <div className="flex items-center justify-between border-b border-border pb-3">
              <div className="flex items-center gap-2">
                <ShieldCheck className="size-4 text-authoritative" />
                <span className="font-bold text-foreground text-[13px] font-sans">
                  Dual Control &amp; SRE Policies
                </span>
              </div>
              <span className="text-[10px] text-muted-foreground uppercase font-bold tracking-wider">
                COMPLIANCE MANDATE
              </span>
            </div>

            {/* Maker-Checker Toggle */}
            <div className="border border-border bg-surface p-3.5 space-y-2 shadow-2xs">
              <div className="flex items-center justify-between">
                <div>
                  <div className="font-bold text-foreground text-[12px] flex items-center gap-1.5">
                    <Lock className="size-3.5 text-navy" />
                    <span>Maker-Checker Dual Control Enforcement</span>
                  </div>
                  <div className="mt-1 font-sans text-[11.5px] text-muted-foreground leading-relaxed">
                    Rule creators and model deployers cannot self-approve production promotions. Mandates independent second officer sign-off.
                  </div>
                </div>
                <input
                  type="checkbox"
                  checked={makerCheckerEnforced}
                  onChange={(e) => setMakerCheckerEnforced(e.target.checked)}
                  className="size-4 accent-navy cursor-pointer"
                />
              </div>
            </div>

            {/* Timeout & Fallback Budget */}
            <div className="border border-border bg-surface p-3.5 space-y-2 shadow-2xs">
              <div className="flex items-center justify-between">
                <span className="font-bold text-foreground text-[12px] flex items-center gap-1.5">
                  <Clock className="size-3.5 text-authoritative" />
                  <span>Synchronous Timeout Budget (ms)</span>
                </span>
                <span className="font-bold text-navy text-[13px]">{timeoutBudgetMs} ms</span>
              </div>
              <input
                type="range"
                min="50"
                max="200"
                step="10"
                value={timeoutBudgetMs}
                onChange={(e) => setTimeoutBudgetMs(Number(e.target.value))}
                className="w-full accent-navy cursor-pointer"
              />
              <div className="font-sans text-[11px] text-muted-foreground leading-snug">
                If the synchronous pipeline exceeds {timeoutBudgetMs}ms, execution gracefully falls back to deterministic AST rules with <Mono className="text-amber-intel">is_degraded: true</Mono>.
              </div>
            </div>

            {/* Save Button */}
            <div className="pt-2 flex justify-end gap-3">
              <button
                type="button"
                onClick={() => {
                  setBlockThreshold(0.80);
                  setReviewThreshold(0.50);
                  setChallengeThreshold(0.30);
                  setTimeoutBudgetMs(100);
                  setMakerCheckerEnforced(true);
                  toast.info("Reset settings to default calibrated profile");
                }}
                className="border border-border bg-surface hover:bg-secondary px-3 py-1.5 text-muted-foreground hover:text-foreground text-[11px] cursor-pointer flex items-center gap-1.5 shadow-2xs"
              >
                <RotateCcw className="size-3" />
                <span>Reset Defaults</span>
              </button>
              <button
                type="button"
                onClick={handleSaveSettings}
                disabled={saving}
                className="border border-navy bg-navy hover:bg-navy/90 px-4 py-1.5 font-bold text-white text-[11px] cursor-pointer flex items-center gap-1.5 shadow-xs disabled:opacity-50"
              >
                <Save className="size-3.5" />
                <span>{saving ? "Committing Changes..." : "Commit Settings to Ledger"}</span>
              </button>
            </div>
          </div>

          {/* Read-Only Environment Metadata */}
          <div className="border border-border bg-card p-4 shadow-xs space-y-2">
            <span className="text-[9.5px] font-bold text-muted-foreground uppercase tracking-wider block border-b border-border pb-1.5">
              TENANT METADATA
            </span>
            <div className="space-y-1 text-[11px]">
              {tenantSettings.map((s) => (
                <div key={s.key} className="flex justify-between py-0.5">
                  <span className="text-muted-foreground font-sans">{s.key}:</span>
                  <Mono className="font-bold text-foreground">{s.value}</Mono>
                </div>
              ))}
            </div>
          </div>
        </section>
      </div>
    </Page>
  );
}
