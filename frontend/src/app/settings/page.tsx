"use client";

import React, { useState } from "react";
import { Settings, Save, Check } from "lucide-react";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";

export default function SettingsPage() {
  const [blockThreshold, setBlockThreshold] = useState(80);
  const [reviewThreshold, setReviewThreshold] = useState(30);
  const [activeModel, setActiveModel] = useState("fraud-xgb-v5-prod");
  const [webhookUrl, setWebhookUrl] = useState("https://api.acmebank.com/v1/ropus-webhooks");
  const [saved, setSaved] = useState(false);

  const handleSave = () => {
    setSaved(true);
    setTimeout(() => setSaved(false), 2000);
  };

  return (
    <div className="flex-1 p-6 md:p-8 space-y-6 max-w-4xl mx-auto">
      {/* Header */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 border-b border-slate-800 pb-5">
        <div>
          <h1 className="text-2xl font-bold text-white tracking-tight flex items-center gap-2.5">
            <Settings className="size-6 text-indigo-400" />
            <span>Tenant Configuration & Policies</span>
          </h1>
          <p className="text-sm text-slate-400 mt-1">
            Customize risk thresholds, configure production webhook endpoints, and choose active inference models.
          </p>
        </div>
        <Button onClick={handleSave} className="bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-semibold">
          {saved ? <Check className="size-3.5 mr-1.5 text-emerald-300" /> : <Save className="size-3.5 mr-1.5" />}
          {saved ? "Settings Saved" : "Save Changes"}
        </Button>
      </div>

      {/* Settings Sections */}
      <div className="space-y-6">
        {/* Risk Thresholds */}
        <Card className="bg-slate-900/80 border-slate-800">
          <CardHeader>
            <CardTitle className="text-base font-semibold text-white">Risk Evaluation Thresholds</CardTitle>
          </CardHeader>
          <CardContent className="space-y-6">
            <div>
              <div className="flex justify-between text-xs font-mono mb-2">
                <span className="text-slate-400">Hard Block Threshold: Score &ge; {blockThreshold}%</span>
                <span className="text-rose-400 font-bold">BLOCK</span>
              </div>
              <input
                type="range"
                min="50"
                max="95"
                value={blockThreshold}
                onChange={(e) => setBlockThreshold(Number(e.target.value))}
                className="w-full h-2 bg-slate-800 rounded-lg appearance-none cursor-pointer accent-indigo-500"
              />
            </div>

            <div>
              <div className="flex justify-between text-xs font-mono mb-2">
                <span className="text-slate-400">Manual Review Threshold: Score &ge; {reviewThreshold}%</span>
                <span className="text-amber-400 font-bold">REVIEW / CHALLENGE</span>
              </div>
              <input
                type="range"
                min="10"
                max="50"
                value={reviewThreshold}
                onChange={(e) => setReviewThreshold(Number(e.target.value))}
                className="w-full h-2 bg-slate-800 rounded-lg appearance-none cursor-pointer accent-indigo-500"
              />
            </div>
          </CardContent>
        </Card>

        {/* Model Selection */}
        <Card className="bg-slate-900/80 border-slate-800">
          <CardHeader>
            <CardTitle className="text-base font-semibold text-white">Active Machine Learning Model</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <label className="text-xs font-mono text-slate-400">Production Inference Model Version</label>
              <select
                value={activeModel}
                onChange={(e) => setActiveModel(e.target.value)}
                className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-xs text-white focus:outline-none focus:border-indigo-500"
              >
                <option value="fraud-xgb-v5-prod">fraud-xgb-v5-prod (Gradient Boosted Tree • AUC 0.982)</option>
                <option value="fraud-lgbm-v4-prod">fraud-lgbm-v4-prod (LightGBM High-Throughput • AUC 0.978)</option>
                <option value="fraud-gnn-v2-canary">fraud-gnn-v2-canary (Graph Neural Network • 10% Traffic Split)</option>
              </select>
            </div>
          </CardContent>
        </Card>

        {/* Webhooks */}
        <Card className="bg-slate-900/80 border-slate-800">
          <CardHeader>
            <CardTitle className="text-base font-semibold text-white">Customer Webhook Dispatch URL</CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            <label className="text-xs font-mono text-slate-400">Endpoint Destination (Receives HMAC-SHA256 Signed JSON)</label>
            <input
              type="text"
              value={webhookUrl}
              onChange={(e) => setWebhookUrl(e.target.value)}
              className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-xs text-white font-mono focus:outline-none focus:border-indigo-500"
            />
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
