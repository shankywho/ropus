"use client";

import React from "react";
import { Sliders, Cpu } from "lucide-react";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";

const mockModels = [
  {
    name: "LightGBM Supervised Ensemble",
    version: "v3.34.1-prod",
    auc: 0.988,
    ks: 0.742,
    p95Latency: "4.8ms",
    status: "PRODUCTION_ACTIVE",
    driftPSI: 0.024,
    fairnessRatio: "94.2%",
    lastAudited: "Today, 04:00 UTC",
  },
  {
    name: "Graph Neural Network Ring Classifier",
    version: "v3.27.0-prod",
    auc: 0.975,
    ks: 0.710,
    p95Latency: "8.2ms",
    status: "PRODUCTION_ACTIVE",
    driftPSI: 0.031,
    fairnessRatio: "91.8%",
    lastAudited: "Yesterday",
  },
  {
    name: "Autoencoder Unsupervised Anomaly Engine",
    version: "v3.16.4-canary",
    auc: 0.945,
    ks: 0.680,
    p95Latency: "3.2ms",
    status: "CANARY_EVALUATION",
    driftPSI: 0.015,
    fairnessRatio: "96.5%",
    lastAudited: "3 days ago",
  },
];

export default function ModelGovernancePage() {
  return (
    <div className="flex-1 p-6 md:p-8 space-y-6 max-w-7xl mx-auto">
      {/* Header */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 border-b border-slate-800 pb-5">
        <div>
          <h1 className="text-2xl font-bold text-white tracking-tight flex items-center gap-2.5">
            <Sliders className="size-6 text-indigo-400" />
            <span>AI Model Governance & MRM Platform</span>
          </h1>
          <p className="text-sm text-slate-400 mt-1">
            Continuous Model Risk Management (SR 11-7), feature drift telemetry, fairness audits, and automated retraining gating.
          </p>
        </div>
        <div className="flex items-center gap-3">
          <Badge className="bg-emerald-500/20 text-emerald-300 border-emerald-500/30 text-xs px-3 py-1 font-mono">
            Audit Ready • Tier 1 High Risk
          </Badge>
        </div>
      </div>

      {/* Model Roster */}
      <div className="grid grid-cols-1 gap-4">
        {mockModels.map((model, idx) => (
          <Card key={idx} className="bg-slate-900 border-slate-800 text-slate-100 shadow-xl">
            <CardHeader className="border-b border-slate-800/80 p-5 flex flex-row items-center justify-between">
              <div className="flex items-center gap-3">
                <div className="size-10 rounded-xl bg-indigo-500/10 border border-indigo-500/20 flex items-center justify-center">
                  <Cpu className="size-5 text-indigo-400" />
                </div>
                <div>
                  <CardTitle className="text-base font-semibold text-white flex items-center gap-2">
                    <span>{model.name}</span>
                    <span className="text-xs font-mono text-indigo-300 bg-indigo-950/60 px-2 py-0.5 rounded border border-indigo-500/30">
                      {model.version}
                    </span>
                  </CardTitle>
                  <p className="text-xs text-slate-400 font-mono mt-0.5">Last Audited: {model.lastAudited}</p>
                </div>
              </div>
              <Badge
                className={`text-xs px-3 py-1 font-semibold ${
                  model.status === "PRODUCTION_ACTIVE"
                    ? "bg-emerald-500/20 text-emerald-300 border-emerald-500/30"
                    : "bg-purple-500/20 text-purple-300 border-purple-500/30"
                }`}
              >
                {model.status}
              </Badge>
            </CardHeader>

            <CardContent className="p-5">
              <div className="grid grid-cols-2 md:grid-cols-5 gap-4">
                <div className="p-3 rounded-lg bg-slate-950/60 border border-slate-800">
                  <div className="text-[11px] text-slate-400">Discriminative AUC</div>
                  <div className="text-sm font-bold font-mono text-emerald-400 mt-1">{model.auc}</div>
                </div>
                <div className="p-3 rounded-lg bg-slate-950/60 border border-slate-800">
                  <div className="text-[11px] text-slate-400">Kolmogorov-Smirnov</div>
                  <div className="text-sm font-bold font-mono text-indigo-300 mt-1">{model.ks}</div>
                </div>
                <div className="p-3 rounded-lg bg-slate-950/60 border border-slate-800">
                  <div className="text-[11px] text-slate-400">Inference Latency (P95)</div>
                  <div className="text-sm font-bold font-mono text-slate-200 mt-1">{model.p95Latency}</div>
                </div>
                <div className="p-3 rounded-lg bg-slate-950/60 border border-slate-800">
                  <div className="text-[11px] text-slate-400">Feature Drift (PSI)</div>
                  <div className="text-sm font-bold font-mono text-emerald-400 mt-1">{model.driftPSI} (Stable)</div>
                </div>
                <div className="p-3 rounded-lg bg-slate-950/60 border border-slate-800">
                  <div className="text-[11px] text-slate-400">Fairness / Parity</div>
                  <div className="text-sm font-bold font-mono text-indigo-300 mt-1">{model.fairnessRatio}</div>
                </div>
              </div>
            </CardContent>
          </Card>
        ))}
      </div>
    </div>
  );
}
