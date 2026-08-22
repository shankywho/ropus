"use client";

import React from "react";
import { ShieldAlert, AlertTriangle, Cpu, Network, UserCheck, Activity } from "lucide-react";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";

interface ExplanationBreakdown {
  graph_intelligence_weight: number;
  behavior_analysis_weight: number;
  threat_intelligence_weight: number;
  machine_learning_weight: number;
}

interface RiskExplanationProps {
  transactionId: string;
  amount: string;
  riskScore: number;
  decision: "APPROVE" | "REVIEW" | "CHALLENGE" | "BLOCK";
  reasons: string[];
  humanExplanation: string;
  breakdown: ExplanationBreakdown;
  graphSignals?: string[];
  modelVersion?: string;
}

export function RiskExplanationCard({
  transactionId,
  amount,
  riskScore,
  decision,
  reasons,
  humanExplanation,
  breakdown,
  graphSignals = [],
  modelVersion = "v3.34-ensemble-prod",
}: RiskExplanationProps) {
  const getBadgeStyle = () => {
    switch (decision) {
      case "APPROVE":
        return "bg-emerald-500/20 text-emerald-300 border-emerald-500/30";
      case "CHALLENGE":
        return "bg-blue-500/20 text-blue-300 border-blue-500/30";
      case "REVIEW":
        return "bg-amber-500/20 text-amber-300 border-amber-500/30";
      case "BLOCK":
        return "bg-rose-500/20 text-rose-300 border-rose-500/30";
    }
  };

  return (
    <Card className="bg-slate-900 border-slate-800 text-slate-100 shadow-xl overflow-hidden">
      <CardHeader className="border-b border-slate-800/80 bg-slate-950/40 p-5">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="size-10 rounded-xl bg-indigo-500/10 border border-indigo-500/20 flex items-center justify-center">
              <Activity className="size-5 text-indigo-400" />
            </div>
            <div>
              <CardTitle className="text-base font-semibold flex items-center gap-2">
                <span>Transaction Decision:</span>
                <span className="font-mono text-indigo-300">{transactionId}</span>
              </CardTitle>
              <p className="text-xs text-slate-400 font-mono mt-0.5">
                Amount: <span className="text-slate-200 font-bold">{amount}</span> • Model: {modelVersion}
              </p>
            </div>
          </div>
          <div className="flex items-center gap-3">
            <div className="text-right">
              <div className="text-xs text-slate-400 font-mono">Risk Score</div>
              <div className="text-xl font-bold font-mono text-white">{riskScore}%</div>
            </div>
            <Badge className={`px-3 py-1 text-xs font-semibold border ${getBadgeStyle()}`}>
              {decision}
            </Badge>
          </div>
        </div>
      </CardHeader>

      <CardContent className="p-5 space-y-6">
        {/* Human Readable Explanation */}
        <div className="p-4 rounded-xl bg-slate-950/70 border border-slate-800 text-sm text-slate-200 leading-relaxed">
          <span className="font-semibold text-indigo-400 mr-2">AI Summary:</span>
          {humanExplanation}
        </div>

        {/* Feature Contribution Breakdown */}
        <div>
          <h4 className="text-xs font-semibold uppercase tracking-wider text-slate-400 mb-3 flex items-center gap-2">
            <Cpu className="size-4 text-indigo-400" />
            <span>Factor Attribution & Intelligence Weights</span>
          </h4>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
            {/* Threat Intelligence */}
            <div className="p-3 rounded-lg bg-slate-950/50 border border-slate-800/80 space-y-1.5">
              <div className="flex justify-between text-xs">
                <span className="text-slate-300 flex items-center gap-1.5">
                  <ShieldAlert className="size-3.5 text-rose-400" /> Threat Intelligence
                </span>
                <span className="font-mono text-slate-200 font-bold">
                  {Math.round(breakdown.threat_intelligence_weight * 100)}%
                </span>
              </div>
              <div className="h-2 w-full bg-slate-800 rounded-full overflow-hidden">
                <div
                  className="h-full bg-gradient-to-r from-rose-500 to-amber-500 rounded-full"
                  style={{ width: `${breakdown.threat_intelligence_weight * 100}%` }}
                />
              </div>
            </div>

            {/* Graph Intelligence */}
            <div className="p-3 rounded-lg bg-slate-950/50 border border-slate-800/80 space-y-1.5">
              <div className="flex justify-between text-xs">
                <span className="text-slate-300 flex items-center gap-1.5">
                  <Network className="size-3.5 text-indigo-400" /> Graph Intelligence
                </span>
                <span className="font-mono text-slate-200 font-bold">
                  {Math.round(breakdown.graph_intelligence_weight * 100)}%
                </span>
              </div>
              <div className="h-2 w-full bg-slate-800 rounded-full overflow-hidden">
                <div
                  className="h-full bg-gradient-to-r from-indigo-500 to-purple-500 rounded-full"
                  style={{ width: `${breakdown.graph_intelligence_weight * 100}%` }}
                />
              </div>
            </div>

            {/* Behavior Analysis */}
            <div className="p-3 rounded-lg bg-slate-950/50 border border-slate-800/80 space-y-1.5">
              <div className="flex justify-between text-xs">
                <span className="text-slate-300 flex items-center gap-1.5">
                  <UserCheck className="size-3.5 text-amber-400" /> Behavioral Anomaly
                </span>
                <span className="font-mono text-slate-200 font-bold">
                  {Math.round(breakdown.behavior_analysis_weight * 100)}%
                </span>
              </div>
              <div className="h-2 w-full bg-slate-800 rounded-full overflow-hidden">
                <div
                  className="h-full bg-gradient-to-r from-amber-500 to-yellow-500 rounded-full"
                  style={{ width: `${breakdown.behavior_analysis_weight * 100}%` }}
                />
              </div>
            </div>

            {/* Machine Learning Model */}
            <div className="p-3 rounded-lg bg-slate-950/50 border border-slate-800/80 space-y-1.5">
              <div className="flex justify-between text-xs">
                <span className="text-slate-300 flex items-center gap-1.5">
                  <Activity className="size-3.5 text-emerald-400" /> ML Ensemble Inference
                </span>
                <span className="font-mono text-slate-200 font-bold">
                  {Math.round(breakdown.machine_learning_weight * 100)}%
                </span>
              </div>
              <div className="h-2 w-full bg-slate-800 rounded-full overflow-hidden">
                <div
                  className="h-full bg-gradient-to-r from-emerald-500 to-teal-500 rounded-full"
                  style={{ width: `${breakdown.machine_learning_weight * 100}%` }}
                />
              </div>
            </div>
          </div>
        </div>

        {/* Triggered Reasons & Signals */}
        {reasons.length > 0 && (
          <div>
            <h4 className="text-xs font-semibold uppercase tracking-wider text-slate-400 mb-2">
              Decision Drivers & Policy Triggers
            </h4>
            <div className="space-y-1.5">
              {reasons.map((reason, idx) => (
                <div key={idx} className="flex items-start gap-2 text-xs text-slate-300 bg-slate-950/40 p-2 rounded border border-slate-800/60">
                  <AlertTriangle className="size-3.5 text-amber-400 shrink-0 mt-0.5" />
                  <span>{reason}</span>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Graph Signals */}
        {graphSignals.length > 0 && (
          <div>
            <h4 className="text-xs font-semibold uppercase tracking-wider text-indigo-400 mb-2">
              Graph Community Signals
            </h4>
            <div className="space-y-1.5">
              {graphSignals.map((signal, idx) => (
                <div key={idx} className="flex items-start gap-2 text-xs text-indigo-200 bg-indigo-950/30 p-2 rounded border border-indigo-500/20">
                  <Network className="size-3.5 text-indigo-400 shrink-0 mt-0.5" />
                  <span>{signal}</span>
                </div>
              ))}
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  );
}
