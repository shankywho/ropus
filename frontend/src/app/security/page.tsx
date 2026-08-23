"use client";

import React from "react";
import { DataProvenanceBadge } from "@/components/ropus/DataProvenanceBadge";
import { ShieldCheck, Lock, AlertTriangle, Key, Terminal } from "lucide-react";

export default function SecurityAuditPage() {
  const securityEvents = [
    {
      id: "sec_99182",
      type: "IP_BLOCKLIST_HIT",
      source: "198.51.100.44 (Bulletproof Proxy)",
      action: "REJECTED_403",
      time: "2 mins ago",
      severity: "HIGH",
    },
    {
      id: "sec_99181",
      type: "KEY_ROTATION_SUCCESS",
      source: "admin@acmebank.com",
      action: "ROP_LIVE_ROTATED",
      time: "15 mins ago",
      severity: "LOW",
    },
    {
      id: "sec_99180",
      type: "SQL_INJECTION_ATTEMPT",
      source: "POST /v1/risk/evaluate (Parameter: customer_id)",
      action: "SANITIZED_AND_BLOCKED",
      time: "1 hour ago",
      severity: "CRITICAL",
    },
  ];

  const auditLedger = [
    {
      id: "aud_1724330101",
      actor: "system_scheduler",
      action: "SNAPSHOT_BACKUP_COMPLETED",
      resource: "s3://ropus-backups-dr-us-west-2/bkp_full_01.tar.gz",
      hash: "8f4b1e9c...3a71",
      timestamp: "2026-08-22 17:00:00 UTC",
    },
    {
      id: "aud_1724330100",
      actor: "elena.r@acmebank.com",
      action: "RULE_THRESHOLD_MODIFIED",
      resource: "policy_carding_velocity_v2",
      hash: "2d9c0a4e...8f12",
      timestamp: "2026-08-22 16:45:12 UTC",
    },
    {
      id: "aud_1724330099",
      actor: "mlops_pipeline_worker",
      action: "MODEL_PROMOTION_APPROVED",
      resource: "fraud-xgb-v5-prod",
      hash: "e5a88c3f...99bc",
      timestamp: "2026-08-22 15:30:00 UTC",
    },
  ];

  return (
    <div className="space-y-5">
      {/* 1. Header */}
      <div className="flex flex-wrap items-center justify-between gap-4 pb-3.5 border-b border-[#1c2536]">
        <div>
          <div className="flex items-center gap-2.5">
            <h1 className="text-lg font-bold text-[#f4f5f7] tracking-tight font-mono uppercase">
              Security Operations &amp; Tamper-Evident Ledger
            </h1>
            <DataProvenanceBadge type="LIVE_BACKEND" sublabel="SHA-256 Chained Hash" />
          </div>
          <p className="text-xs text-[#5e6c84] font-mono mt-1">
            Zero-PII tokenization, AES-256 GCM encryption at rest &amp; cryptographic audit trail verification
          </p>
        </div>
      </div>

      {/* 2. Security Posture Strip */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-3 font-mono">
        <div className="p-3.5 bg-[#0f172a] border border-[#1c2536] rounded-[4px] flex items-center justify-between">
          <div>
            <span className="text-[10px] text-[#5e6c84] uppercase tracking-wider block">ENCRYPTION AT REST</span>
            <span className="text-base font-bold text-[#04db7c] mt-1 block">AES-256 GCM Active</span>
            <span className="text-[10px] text-[#5e6c84]">Zero-PII Tokenized Storage</span>
          </div>
          <div className="p-2 bg-[#04db7c15] text-[#04db7c] rounded-[3px]">
            <Lock className="w-4 h-4" />
          </div>
        </div>

        <div className="p-3.5 bg-[#0f172a] border border-[#1c2536] rounded-[4px] flex items-center justify-between">
          <div>
            <span className="text-[10px] text-[#5e6c84] uppercase tracking-wider block">AUDIT HASH INTEGRITY</span>
            <span className="text-base font-bold text-[#0d94fb] mt-1 block">100% Verified</span>
            <span className="text-[10px] text-[#5e6c84]">Zero Broken Hash Chains</span>
          </div>
          <div className="p-2 bg-[#0d94fb15] text-[#0d94fb] rounded-[3px]">
            <Terminal className="w-4 h-4" />
          </div>
        </div>

        <div className="p-3.5 bg-[#0f172a] border border-[#1c2536] rounded-[4px] flex items-center justify-between">
          <div>
            <span className="text-[10px] text-[#5e6c84] uppercase tracking-wider block">ZERO-TRUST WAF / GATEWAY</span>
            <span className="text-base font-bold text-[#f4f5f7] mt-1 block">Enforced (TLS 1.3)</span>
            <span className="text-[10px] text-[#5e6c84]">HMAC-SHA256 Request Signing</span>
          </div>
          <div className="p-2 bg-[#0a1324] text-[#97a0af] rounded-[3px]">
            <Key className="w-4 h-4" />
          </div>
        </div>
      </div>

      {/* 3. Cryptographic Audit Ledger Table */}
      <div className="bg-[#0f172a] border border-[#1c2536] rounded-[4px] p-4 font-mono text-xs">
        <div className="flex items-center justify-between pb-2.5 mb-3 border-b border-[#1c2536]">
          <div className="flex items-center gap-2">
            <Terminal className="w-4 h-4 text-[#0d94fb]" />
            <h2 className="font-bold uppercase tracking-wider text-[#f4f5f7] text-xs">
              Immutable Hash-Chained Audit Trail ({auditLedger.length})
            </h2>
          </div>
          <span className="text-[#5e6c84] text-[10px]">Merkle Tree Root: 0x8f4b1e...</span>
        </div>

        <div className="overflow-x-auto">
          <table className="w-full text-left">
            <thead>
              <tr className="border-b border-[#1c2536] text-[#5e6c84] text-[11px]">
                <th className="pb-2 font-medium">LOG ID</th>
                <th className="pb-2 font-medium">ACTOR</th>
                <th className="pb-2 font-medium">ACTION</th>
                <th className="pb-2 font-medium">RESOURCE TARGET</th>
                <th className="pb-2 font-medium">SHA-256 HASH</th>
                <th className="pb-2 font-medium text-right">TIMESTAMP</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-[#1c2536]">
              {auditLedger.map((l) => (
                <tr key={l.id} className="hover:bg-[#142036] transition-colors">
                  <td className="py-2.5 text-[#f4f5f7] font-semibold">{l.id}</td>
                  <td className="py-2.5 text-[#0d94fb]">{l.actor}</td>
                  <td className="py-2.5 font-bold text-[#04db7c]">{l.action}</td>
                  <td className="py-2.5 text-[#97a0af]">{l.resource}</td>
                  <td className="py-2.5 text-[#5e6c84]">{l.hash}</td>
                  <td className="py-2.5 text-right text-[#5e6c84]">{l.timestamp}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}
