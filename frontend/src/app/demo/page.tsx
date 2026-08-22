"use client";

import React, { useState, useEffect } from "react";
import {
  PlayCircle,
  RotateCcw,
  Pause,
  Play,
  ShieldAlert,
  ShieldCheck,
  Activity,
  Terminal,
  Share2,
  Cpu,
  UserCheck,
  Clock,
  ArrowRight,
  Sparkles,
  FileText,
  AlertTriangle,
  Lock,
  CheckCircle2,
} from "lucide-react";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";

export default function DemoPage() {
  const [activeStep, setActiveStep] = useState(1);
  const [isPlaying, setIsPlaying] = useState(false);
  const [analystDecision, setAnalystDecision] = useState<string | null>(null);

  // Auto-play timer for presentation mode
  useEffect(() => {
    let timer: NodeJS.Timeout;
    if (isPlaying && activeStep < 7) {
      timer = setTimeout(() => {
        setActiveStep((prev) => prev + 1);
      }, 3500);
    } else if (activeStep === 7) {
      setIsPlaying(false);
    }
    return () => clearTimeout(timer);
  }, [isPlaying, activeStep]);

  const handleStart = () => {
    setActiveStep(1);
    setIsPlaying(true);
    setAnalystDecision(null);
  };

  const handleReset = () => {
    setActiveStep(1);
    setIsPlaying(false);
    setAnalystDecision(null);
  };

  const liveEvents = [
    { time: "17:42:01", event: "DEVICE_CHANGE", actor: "usr_sarah_connor", desc: "Unseen device canvas fingerprint dev_mule_cluster_99", severity: "WARN" },
    { time: "17:42:02", event: "IMPOSSIBLE_TRAVEL", actor: "usr_sarah_connor", desc: "Geo distance 7,850 km (Limassol, CY -> NYC) within 12m", severity: "HIGH" },
    { time: "17:42:03", event: "IP_REPUTATION_ALERT", actor: "198.51.100.44", desc: "Bulletproof anonymous proxy match", severity: "CRITICAL" },
    { time: "17:42:04", event: "GRAPH_RELATION_DISCOVERED", actor: "GraphEngine", desc: "3-hop cyclic connection to 14 synthetic accounts", severity: "CRITICAL" },
    { time: "17:42:05", event: "RISK_SCORE_UPDATED", actor: "RiskEngine", desc: "Score escalated from 0.04 (APPROVE) -> 0.96 (BLOCK)", severity: "CRITICAL" },
    { time: "17:42:06", event: "ML_DECISION_ISSUED", actor: "XGBoost_v5", desc: "Hard BLOCK verdict issued with 0.94 confidence", severity: "CRITICAL" },
    { time: "17:42:07", event: "AI_INVESTIGATION_STARTED", actor: "Claude 3.7 Agent", desc: "Automated evidentiary dossier compiled", severity: "INFO" },
    { time: "17:42:09", event: "CASE_CREATED", actor: "CaseManager", desc: "Priority P0 Case #CASE-88419 opened in analyst review queue", severity: "HIGH" },
    { time: "17:42:10", event: "WEBHOOK_DELIVERED", actor: "WebhookGateway", desc: "HMAC-SHA256 event dispatched to bank destination", severity: "INFO" },
  ];

  return (
    <div className="flex-1 p-6 md:p-8 space-y-8 max-w-7xl mx-auto">
      {/* Header */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 border-b border-slate-800 pb-5">
        <div>
          <h1 className="text-2xl font-bold text-white tracking-tight flex items-center gap-2.5">
            <PlayCircle className="size-6 text-indigo-400" />
            <span>Interactive 5-Minute Investor Demo</span>
          </h1>
          <p className="text-sm text-slate-400 mt-1">
            End-to-end demonstration of autonomous fraud prevention: from behavioral baseline to AI investigation and human governance.
          </p>
        </div>

        {/* Demo Controls */}
        <div className="flex items-center gap-2">
          {!isPlaying ? (
            <Button onClick={handleStart} className="bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-semibold">
              <Play className="size-3.5 mr-1.5 fill-current" /> Start Attack Demo
            </Button>
          ) : (
            <Button onClick={() => setIsPlaying(false)} variant="outline" className="border-slate-800 bg-slate-900 text-xs">
              <Pause className="size-3.5 mr-1.5" /> Pause
            </Button>
          )}
          <Button onClick={handleReset} variant="outline" className="border-slate-800 bg-slate-900 text-xs">
            <RotateCcw className="size-3.5 mr-1.5" /> Reset
          </Button>
        </div>
      </div>

      {/* Stepper Navigation */}
      <div className="grid grid-cols-7 gap-2">
        {[
          { num: 1, title: "1. Baseline" },
          { num: 2, title: "2. Attack Starts" },
          { num: 3, title: "3. Graph Link" },
          { num: 4, title: "4. ML Fusion" },
          { num: 5, title: "5. AI Dossier" },
          { num: 6, title: "6. Case Opened" },
          { num: 7, title: "7. Governance" },
        ].map((s) => (
          <button
            key={s.num}
            onClick={() => {
              setActiveStep(s.num);
              setIsPlaying(false);
            }}
            className={`p-3 rounded-lg border text-left transition-all ${
              activeStep === s.num
                ? "bg-indigo-950/60 border-indigo-500/80 shadow-lg shadow-indigo-500/10 text-white"
                : activeStep > s.num
                ? "bg-slate-900/60 border-slate-800 text-emerald-400"
                : "bg-slate-950/40 border-slate-900 text-slate-500"
            }`}
          >
            <div className="text-[10px] font-mono uppercase tracking-wider">Step {s.num}</div>
            <div className="text-xs font-bold truncate mt-0.5">{s.title.replace(/^\d+\.\s*/, "")}</div>
          </button>
        ))}
      </div>

      {/* Main Interactive Stage Display */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Left 2 Cols: Main Narrative Stage View */}
        <div className="lg:col-span-2 space-y-6">
          {/* STEP 1: NORMAL CUSTOMER BASELINE */}
          {activeStep === 1 && (
            <Card className="bg-slate-900/80 border-slate-800">
              <CardHeader className="border-b border-slate-800/80 pb-4">
                <div className="flex items-center justify-between">
                  <CardTitle className="text-base font-semibold text-white flex items-center gap-2">
                    <UserCheck className="size-4 text-emerald-400" />
                    <span>Step 1 — Legitimate Customer Behavioral Profile</span>
                  </CardTitle>
                  <Badge className="bg-emerald-500/20 text-emerald-300">Baseline Established</Badge>
                </div>
              </CardHeader>
              <CardContent className="p-6 space-y-5">
                <div className="p-4 bg-slate-950/60 border border-slate-800 rounded-lg space-y-2 text-xs">
                  <div className="grid grid-cols-2 gap-4 font-mono">
                    <div><span className="text-slate-500">Customer ID:</span> <span className="text-slate-200">usr_sarah_connor</span></div>
                    <div><span className="text-slate-500">Account Age:</span> <span className="text-slate-200">3 Years (Spotless)</span></div>
                    <div><span className="text-slate-500">Avg Transaction:</span> <span className="text-slate-200">$64.50 USD</span></div>
                    <div><span className="text-slate-500">Primary Geo:</span> <span className="text-slate-200">New York, NY (US)</span></div>
                    <div><span className="text-slate-500">Trusted Device:</span> <span className="text-slate-200">dev_macbook_pro_16</span></div>
                    <div><span className="text-slate-500">Baseline Risk Score:</span> <span className="text-emerald-400 font-bold">0.04 (APPROVE)</span></div>
                  </div>
                </div>

                <div className="p-4 bg-indigo-950/20 border border-indigo-500/30 rounded-lg text-xs text-indigo-200">
                  <strong>Narration:</strong> &ldquo;This is a legitimate user. ROPUS continuously tracks behavioral baselines, typical velocities, and trusted hardware signatures before making real-time risk decisions.&rdquo;
                </div>
              </CardContent>
            </Card>
          )}

          {/* STEP 2: ATTACK BEGINS */}
          {activeStep === 2 && (
            <Card className="bg-slate-900/80 border-slate-800">
              <CardHeader className="border-b border-slate-800/80 pb-4">
                <div className="flex items-center justify-between">
                  <CardTitle className="text-base font-semibold text-white flex items-center gap-2">
                    <ShieldAlert className="size-4 text-amber-400" />
                    <span>Step 2 — Adversary Compromise & Impossible Travel</span>
                  </CardTitle>
                  <Badge className="bg-amber-500/20 text-amber-300">Anomaly Ingested</Badge>
                </div>
              </CardHeader>
              <CardContent className="p-6 space-y-5">
                <div className="p-4 bg-slate-950/60 border border-slate-800 rounded-lg space-y-3 text-xs font-mono">
                  <div className="flex justify-between border-b border-slate-800 pb-2">
                    <span className="text-slate-400">Inbound Transaction ID:</span>
                    <span className="text-white font-bold">tx_order_88419</span>
                  </div>
                  <div className="grid grid-cols-2 gap-3 text-slate-300">
                    <div><span className="text-slate-500">Attempted Amount:</span> <span className="text-amber-300 font-bold">$14,500.00 USD</span></div>
                    <div><span className="text-slate-500">Merchant Target:</span> <span>CryptoLiquidityExpress</span></div>
                    <div><span className="text-slate-500">Reported IP:</span> <span className="text-rose-300">198.51.100.44 (Bulletproof VPN)</span></div>
                    <div><span className="text-slate-500">Origin Geo:</span> <span className="text-rose-300">Limassol, Cyprus (7,850 km jump)</span></div>
                  </div>
                </div>

                <div className="p-4 bg-amber-950/20 border border-amber-500/30 rounded-lg text-xs text-amber-200">
                  <strong>Narration:</strong> &ldquo;Within 12 minutes of domestic New York activity, a session initiates from Cyprus requesting a $14,500 outbound wire transfer over an anonymous VPN.&rdquo;
                </div>
              </CardContent>
            </Card>
          )}

          {/* STEP 3: FRAUD GRAPH LINKAGE */}
          {activeStep === 3 && (
            <Card className="bg-slate-900/80 border-slate-800">
              <CardHeader className="border-b border-slate-800/80 pb-4">
                <div className="flex items-center justify-between">
                  <CardTitle className="text-base font-semibold text-white flex items-center gap-2">
                    <Share2 className="size-4 text-rose-400" />
                    <span>Step 3 — Graph Intelligence: Uncovering the Syndicate Ring</span>
                  </CardTitle>
                  <Badge className="bg-rose-500/20 text-rose-300">14 Linked Nodes Found</Badge>
                </div>
              </CardHeader>
              <CardContent className="p-6 space-y-5">
                <div className="p-4 bg-slate-950/60 border border-slate-800 rounded-lg space-y-3">
                  <p className="text-xs text-slate-300">
                    ROPUS traverses the 3-hop entity neighborhood, discovering that the hardware canvas hash (<code className="text-indigo-400">dev_mule_cluster_99</code>) has been reused across 14 synthetic identities.
                  </p>
                  <div className="flex items-center justify-center p-6 bg-slate-900/90 border border-slate-800 rounded-lg">
                    <div className="text-center space-y-2">
                      <div className="flex items-center justify-center gap-3 text-xs font-mono">
                        <span className="p-2 bg-slate-800 rounded text-slate-300">Victim Account</span>
                        <ArrowRight className="size-4 text-slate-600" />
                        <span className="p-2 bg-rose-500/20 text-rose-300 rounded border border-rose-500/40">Shared Emulator Canvas</span>
                        <ArrowRight className="size-4 text-slate-600" />
                        <span className="p-2 bg-purple-500/20 text-purple-300 rounded border border-purple-500/40">14 Mule Accounts</span>
                      </div>
                      <p className="text-[11px] text-slate-500 font-mono">3-Hop Cyclic Syndicate Neighborhood (Degree: 14)</p>
                    </div>
                  </div>
                </div>

                <div className="p-4 bg-rose-950/20 border border-rose-500/30 rounded-lg text-xs text-rose-200">
                  <strong>Narration:</strong> &ldquo;Ropus does not evaluate this transaction in isolation. It immediately connects the hardware fingerprint to a wider transnational fraud cluster.&rdquo;
                </div>
              </CardContent>
            </Card>
          )}

          {/* STEP 4: ML + FACTOR FUSION */}
          {activeStep === 4 && (
            <Card className="bg-slate-900/80 border-slate-800">
              <CardHeader className="border-b border-slate-800/80 pb-4">
                <div className="flex items-center justify-between">
                  <CardTitle className="text-base font-semibold text-white flex items-center gap-2">
                    <Cpu className="size-4 text-indigo-400" />
                    <span>Step 4 — Real ML & Mathematical Factor Attribution</span>
                  </CardTitle>
                  <Badge className="bg-rose-500/20 text-rose-300 font-bold">Hard BLOCK Verdict</Badge>
                </div>
              </CardHeader>
              <CardContent className="p-6 space-y-5">
                <div className="space-y-2">
                  <div className="flex justify-between items-center text-xs font-mono">
                    <span className="text-slate-400">Composite Risk Score:</span>
                    <span className="text-xl font-black text-rose-400">0.96 (BLOCK)</span>
                  </div>
                  <div className="p-3 bg-slate-950/60 border border-slate-800 rounded-lg space-y-2 text-xs font-mono">
                    <div className="flex justify-between text-slate-300"><span>Impossible Travel / Geolocation:</span> <span className="text-rose-300 font-bold">+0.21</span></div>
                    <div className="flex justify-between text-slate-300"><span>Amount Velocity Surge ($14,500):</span> <span className="text-rose-300 font-bold">+0.22</span></div>
                    <div className="flex justify-between text-slate-300"><span>Device Novelty / Emulator Canvas:</span> <span className="text-rose-300 font-bold">+0.18</span></div>
                    <div className="flex justify-between text-slate-300"><span>IP Reputation / Bulletproof Proxy:</span> <span className="text-rose-300 font-bold">+0.18</span></div>
                    <div className="flex justify-between text-slate-300"><span>Graph Syndicate Exposure:</span> <span className="text-rose-300 font-bold">+0.17</span></div>
                    <div className="flex justify-between text-slate-300"><span>XGBoost ML Probability (0.982):</span> <span className="text-rose-300 font-bold">+0.20</span></div>
                  </div>
                </div>

                <div className="p-4 bg-indigo-950/20 border border-indigo-500/30 rounded-lg text-xs text-indigo-200">
                  <strong>Narration:</strong> &ldquo;Every risk factor has an exact mathematical contribution derived from the backend evaluation pipeline. No black-box guesses.&rdquo;
                </div>
              </CardContent>
            </Card>
          )}

          {/* STEP 5: AI INVESTIGATION DOSSIER */}
          {activeStep === 5 && (
            <Card className="bg-slate-900/80 border-slate-800">
              <CardHeader className="border-b border-slate-800/80 pb-4">
                <div className="flex items-center justify-between">
                  <CardTitle className="text-base font-semibold text-white flex items-center gap-2">
                    <Sparkles className="size-4 text-purple-400" />
                    <span>Step 5 — Autonomous AI Investigator Dossier</span>
                  </CardTitle>
                  <Badge className="bg-purple-500/20 text-purple-300">Facts vs Inferences</Badge>
                </div>
              </CardHeader>
              <CardContent className="p-6 space-y-4 text-xs font-mono">
                <div className="p-3 bg-slate-950/60 border border-slate-800 rounded-lg space-y-2">
                  <div className="text-emerald-400 font-bold flex items-center gap-1.5">
                    <CheckCircle2 className="size-3.5" /> OBSERVED FACTS:
                  </div>
                  <p className="text-slate-300">
                    &bull; Session IP: 198.51.100.44 confirmed bulletproof proxy subnet.<br />
                    &bull; Geo distance jump of 7,850 km in 12 minutes constitutes physical impossibility.
                  </p>
                </div>

                <div className="p-3 bg-slate-950/60 border border-slate-800 rounded-lg space-y-2">
                  <div className="text-amber-400 font-bold flex items-center gap-1.5">
                    <AlertTriangle className="size-3.5" /> INFERRED ATTACK PATTERN:
                  </div>
                  <p className="text-slate-300">
                    &bull; Credential stuffing account takeover followed by immediate automated liquidity drain via synthetic mule network.
                  </p>
                </div>

                <div className="p-3 bg-slate-950/60 border border-slate-800 rounded-lg space-y-2">
                  <div className="text-indigo-400 font-bold flex items-center gap-1.5">
                    <ShieldCheck className="size-3.5" /> RECOMMENDED ACTIONS:
                  </div>
                  <p className="text-slate-300">
                    &bull; Freeze outgoing settlement on tx_order_88419 immediately.<br />
                    &bull; Invalidate all active JWT tokens for usr_sarah_connor and require in-person / WebAuthn identity verification.
                  </p>
                </div>
              </CardContent>
            </Card>
          )}

          {/* STEP 6: CASE CREATION & TIMELINE */}
          {activeStep === 6 && (
            <Card className="bg-slate-900/80 border-slate-800">
              <CardHeader className="border-b border-slate-800/80 pb-4">
                <div className="flex items-center justify-between">
                  <CardTitle className="text-base font-semibold text-white flex items-center gap-2">
                    <FileText className="size-4 text-cyan-400" />
                    <span>Step 6 — Persistent Review Case #CASE-88419</span>
                  </CardTitle>
                  <Badge className="bg-rose-500/20 text-rose-300 font-bold">P0 Critical Priority</Badge>
                </div>
              </CardHeader>
              <CardContent className="p-6 space-y-4">
                <div className="p-4 bg-slate-950/60 border border-slate-800 rounded-lg space-y-3 text-xs font-mono">
                  <div className="flex justify-between border-b border-slate-800 pb-2">
                    <span className="text-slate-400">Case ID:</span>
                    <span className="text-cyan-300 font-bold">CASE-88419</span>
                  </div>
                  <div className="flex justify-between border-b border-slate-800 pb-2">
                    <span className="text-slate-400">Attached Evidence:</span>
                    <span className="text-slate-200">5 Artifacts (IP, Geo, Graph, ML, Device)</span>
                  </div>
                  <div className="flex justify-between border-b border-slate-800 pb-2">
                    <span className="text-slate-400">Audit Hash Chain:</span>
                    <span className="text-slate-500">aud_88419_blk_01 (4f8a...e91c)</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-slate-400">Webhook Status:</span>
                    <span className="text-emerald-400 font-bold">Dispatched (200 OK)</span>
                  </div>
                </div>
              </CardContent>
            </Card>
          )}

          {/* STEP 7: HUMAN-IN-THE-LOOP GOVERNANCE */}
          {activeStep === 7 && (
            <Card className="bg-slate-900/80 border-slate-800">
              <CardHeader className="border-b border-slate-800/80 pb-4">
                <div className="flex items-center justify-between">
                  <CardTitle className="text-base font-semibold text-white flex items-center gap-2">
                    <Lock className="size-4 text-emerald-400" />
                    <span>Step 7 — Human-in-the-Loop Analyst Governance</span>
                  </CardTitle>
                  <Badge className="bg-emerald-500/20 text-emerald-300">Analyst Authority</Badge>
                </div>
              </CardHeader>
              <CardContent className="p-6 space-y-5">
                <p className="text-xs text-slate-300">
                  Analysts retain complete authority to confirm the AI block, request step-up biometric challenge, or override with reason logging.
                </p>

                <div className="flex items-center gap-3">
                  <Button
                    onClick={() => setAnalystDecision("CONFIRMED_BLOCK")}
                    className={`text-xs font-semibold ${
                      analystDecision === "CONFIRMED_BLOCK"
                        ? "bg-rose-600 text-white"
                        : "bg-slate-800 hover:bg-slate-700 text-slate-200"
                    }`}
                  >
                    Confirm Block & Freeze
                  </Button>
                  <Button
                    onClick={() => setAnalystDecision("ESCALATED_AML")}
                    className={`text-xs font-semibold ${
                      analystDecision === "ESCALATED_AML"
                        ? "bg-amber-600 text-white"
                        : "bg-slate-800 hover:bg-slate-700 text-slate-200"
                    }`}
                  >
                    Escalate to AML Compliance
                  </Button>
                  <Button
                    onClick={() => setAnalystDecision("OVERRIDDEN")}
                    variant="outline"
                    className="border-slate-700 text-xs text-slate-400"
                  >
                    Override (False Positive)
                  </Button>
                </div>

                {analystDecision && (
                  <div className="p-3 bg-slate-950 border border-slate-800 rounded-lg text-xs font-mono text-emerald-300">
                    &bull; Analyst action <strong>{analystDecision}</strong> recorded by <code className="text-white">elena.r@acmebank.com</code> with immutable SHA-256 audit ledger entry.
                  </div>
                )}
              </CardContent>
            </Card>
          )}
        </div>

        {/* Right Col: Live Event Stream Ticker */}
        <div className="space-y-4">
          <Card className="bg-slate-900/80 border-slate-800">
            <CardHeader className="pb-3 border-b border-slate-800">
              <CardTitle className="text-sm font-semibold text-white flex items-center justify-between">
                <span className="flex items-center gap-2">
                  <Terminal className="size-4 text-indigo-400" />
                  <span>Live Event Stream</span>
                </span>
                <span className="size-2 rounded-full bg-emerald-500 animate-pulse" />
              </CardTitle>
            </CardHeader>
            <CardContent className="p-3 max-h-[420px] overflow-y-auto space-y-2 font-mono text-[11px]">
              {liveEvents.slice(0, activeStep + 2).map((e, idx) => (
                <div key={idx} className="p-2 bg-slate-950/60 border border-slate-800/80 rounded space-y-1">
                  <div className="flex justify-between text-slate-500">
                    <span>{e.time}</span>
                    <span
                      className={
                        e.severity === "CRITICAL"
                          ? "text-rose-400 font-bold"
                          : e.severity === "HIGH"
                          ? "text-amber-400 font-bold"
                          : "text-indigo-400"
                      }
                    >
                      {e.event}
                    </span>
                  </div>
                  <p className="text-slate-300 text-[10px]">{e.desc}</p>
                </div>
              ))}
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
}
