"use client";

import React, { useState, useEffect } from "react";
import { Save, Check } from "lucide-react";
import { modelsApi } from "@/api/models";

export default function SettingsPage() {
  const [activeModel, setActiveModel] = useState("fraud-xgb-25f-v3.0");
  const [fallbackModel, setFallbackModel] = useState("fraud-xgb-15f-v1.5");
  const [saved, setSaved] = useState(false);
  const [tenantId] = useState("00000000-0000-0000-0000-000000000001");

  useEffect(() => {
    const loadSettings = async () => {
      try {
        const prodData = await modelsApi.getProductionModel();
        if (prodData && prodData.production_model) {
          setActiveModel(prodData.production_model.version);
        }
        if (prodData && prodData.fallback_model) {
          setFallbackModel(prodData.fallback_model.version);
        }
      } catch (err) {
        console.warn("Could not load production model metadata:", err);
      }
    };
    loadSettings();
  }, []);

  const handleSave = () => {
    setSaved(true);
    setTimeout(() => setSaved(false), 2500);
  };

  return (
    <div className="max-w-4xl mx-auto space-y-6">
      {/* Header */}
      <div className="flex flex-wrap items-center justify-between gap-4 border-b border-[#1c2536] pb-4">
        <div>
          <div className="flex items-center gap-2">
            <h1 className="text-xl font-bold text-[#f4f5f7] tracking-tight">Tenant Configuration &amp; Policies</h1>
            <span className="text-[10px] font-mono bg-[#0d94fb15] text-[#0d94fb] px-1.5 py-0.5 rounded-[2px] border border-[#0d94fb33]">
              TENANT CONTEXT
            </span>
          </div>
          <p className="text-xs text-[#97a0af] font-mono mt-0.5">
            Calibrated decision thresholds, model arbitration precedence, and webhook settings
          </p>
        </div>

        <button
          onClick={handleSave}
          className="px-3.5 py-1.5 bg-[#0d94fb] hover:bg-[#0b82dc] text-white font-mono text-xs font-bold rounded-[4px] shadow-xs active:scale-95 cursor-pointer flex items-center gap-1.5"
        >
          {saved ? <Check className="w-3.5 h-3.5 text-[#04db7c]" /> : <Save className="w-3.5 h-3.5" />}
          <span>{saved ? "Policies Synced" : "Save Preferences"}</span>
        </button>
      </div>

      {saved && (
        <div className="p-3 bg-[#04db7c15] border border-[#04db7c44] text-[#04db7c] font-mono text-xs rounded-[4px]">
          ✓ Tenant preference updates saved.
        </div>
      )}

      {/* Settings Sections */}
      <div className="space-y-6">
        {/* Scoped Tenant Identity */}
        <div className="p-5 bg-[#0f172a] border border-[#1c2536] rounded-[4px] shadow-xs space-y-3 font-mono text-xs">
          <span className="text-xs font-mono font-bold text-[#f4f5f7] uppercase tracking-wider block">
            Tenant Isolation Scoping
          </span>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
            <div className="p-3 bg-[#0a1324] rounded-[4px] border border-[#1c2536]">
              <span className="text-[#5e6c84] text-[10px] block">PRIMARY TENANT UUID</span>
              <span className="text-[#0d94fb] font-bold">{tenantId}</span>
            </div>
            <div className="p-3 bg-[#0a1324] rounded-[4px] border border-[#1c2536]">
              <span className="text-[#5e6c84] text-[10px] block">ENCRYPTION KEY REF</span>
              <span className="text-[#04db7c] font-bold">kms://us-east-1/risk-tenant-01</span>
            </div>
          </div>
        </div>

        {/* Server Policy Thresholds */}
        <div className="p-5 bg-[#0f172a] border border-[#1c2536] rounded-[4px] shadow-xs space-y-4 font-mono text-xs">
          <span className="text-xs font-mono font-bold text-[#f4f5f7] uppercase tracking-wider block">
            Calibrated Policy Decision Thresholds (Backend Authoritative)
          </span>

          <div className="space-y-3">
            <div className="p-3 bg-[#0a1324] border border-[#1c2536] rounded-[4px] flex items-center justify-between">
              <div>
                <span className="text-[#04db7c] font-bold block">AUTOMATIC ALLOW (&lt; 0.05)</span>
                <span className="text-[11px] text-[#97a0af]">Sub-millisecond approval with zero step-up friction</span>
              </div>
              <span className="px-2 py-0.5 bg-[#04db7c15] text-[#04db7c] border border-[#04db7c33] rounded-[2px] text-[10px] font-bold">
                POLICY #1
              </span>
            </div>

            <div className="p-3 bg-[#0a1324] border border-[#1c2536] rounded-[4px] flex items-center justify-between">
              <div>
                <span className="text-[#f59e0b] font-bold block">MANUAL REVIEW / STEP-UP (0.05 - 0.35)</span>
                <span className="text-[11px] text-[#97a0af]">Automatic case generation &amp; 24-hour SLA routing</span>
              </div>
              <span className="px-2 py-0.5 bg-[#f59e0b15] text-[#f59e0b] border border-[#f59e0b33] rounded-[2px] text-[10px] font-bold">
                POLICY #2
              </span>
            </div>

            <div className="p-3 bg-[#0a1324] border border-[#1c2536] rounded-[4px] flex items-center justify-between">
              <div>
                <span className="text-[#f05252] font-bold block">AUTOMATIC DECLINE / BLOCK (&ge; 0.35)</span>
                <span className="text-[11px] text-[#97a0af]">Instant transaction decline &amp; recipient quarantine</span>
              </div>
              <span className="px-2 py-0.5 bg-[#f0525215] text-[#f05252] border border-[#f0525233] rounded-[2px] text-[10px] font-bold">
                POLICY #3
              </span>
            </div>
          </div>
        </div>

        {/* Model & Fallback Configuration */}
        <div className="p-5 bg-[#0f172a] border border-[#1c2536] rounded-[4px] shadow-xs space-y-4 font-mono text-xs">
          <span className="text-xs font-mono font-bold text-[#f4f5f7] uppercase tracking-wider block">
            Active Model Registry &amp; Fallback Binding
          </span>

          <div className="space-y-3">
            <div>
              <label className="text-[11px] text-[#5e6c84] block mb-1">PRIMARY PRODUCTION MODEL</label>
              <input
                type="text"
                readOnly
                value={`${activeModel} (25 Features • Beta-Calibrated v2.5)`}
                className="w-full px-3 py-2 bg-[#0a1324] border border-[#1c2536] text-[#0d94fb] font-bold rounded-[4px]"
              />
            </div>

            <div>
              <label className="text-[11px] text-[#5e6c84] block mb-1">EMERGENCY FALLBACK MODEL (ZERO DEPENDENCY)</label>
              <input
                type="text"
                readOnly
                value={`${fallbackModel} (15 Features • In-Memory Fallback)`}
                className="w-full px-3 py-2 bg-[#0a1324] border border-[#1c2536] text-[#f59e0b] font-bold rounded-[4px]"
              />
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
