"use client";

import React from "react";
import { ShieldCheck, Lock, AlertTriangle, Key, Terminal } from "lucide-react";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";

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
    <div className="flex-1 p-6 md:p-8 space-y-8 max-w-7xl mx-auto">
      {/* Header */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 border-b border-slate-800 pb-5">
        <div>
          <h1 className="text-2xl font-bold text-white tracking-tight flex items-center gap-2.5">
            <ShieldCheck className="size-6 text-emerald-400" />
            <span>Security Operations & Tamper-Evident Audit Ledger</span>
          </h1>
          <p className="text-sm text-slate-400 mt-1">
            Real-time security threat mitigation, TLS 1.3/AES-256 GCM encryption verification, and SHA-256 hash-chained audit trails.
          </p>
        </div>
      </div>

      {/* Security Status Cards */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <Card className="bg-slate-900/80 border-slate-800">
          <CardContent className="p-4 flex items-center justify-between">
            <div>
              <p className="text-[11px] font-mono text-slate-400 uppercase">Encryption At Rest</p>
              <p className="text-xl font-bold text-emerald-400 mt-1">AES-256 GCM Active</p>
              <p className="text-[10px] text-slate-500 font-mono mt-1">Zero-PII Tokenized Storage</p>
            </div>
            <div className="p-2.5 bg-emerald-500/10 rounded-lg text-emerald-400">
              <Lock className="size-5" />
            </div>
          </CardContent>
        </Card>

        <Card className="bg-slate-900/80 border-slate-800">
          <CardContent className="p-4 flex items-center justify-between">
            <div>
              <p className="text-[11px] font-mono text-slate-400 uppercase">Audit Hash Integrity</p>
              <p className="text-xl font-bold text-indigo-400 mt-1">100% Verified</p>
              <p className="text-[10px] text-slate-500 font-mono mt-1">Zero Broken Hash Chains</p>
            </div>
            <div className="p-2.5 bg-indigo-500/10 rounded-lg text-indigo-400">
              <Terminal className="size-5" />
            </div>
          </CardContent>
        </Card>

        <Card className="bg-slate-900/80 border-slate-800">
          <CardContent className="p-4 flex items-center justify-between">
            <div>
              <p className="text-[11px] font-mono text-slate-400 uppercase">Zero-Trust WAF / Gateway</p>
              <p className="text-xl font-bold text-white mt-1">Enforced (TLS 1.3)</p>
              <p className="text-[10px] text-slate-500 font-mono mt-1">HMAC-SHA256 Request Signing</p>
            </div>
            <div className="p-2.5 bg-purple-500/10 rounded-lg text-purple-400">
              <Key className="size-5" />
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Security Threat Interceptions */}
      <Card className="bg-slate-900/80 border-slate-800">
        <CardHeader>
          <CardTitle className="text-base font-semibold text-white flex items-center gap-2">
            <AlertTriangle className="size-4 text-amber-400" />
            <span>Recent Security Threat Interceptions</span>
          </CardTitle>
        </CardHeader>
        <CardContent>
          <table className="w-full text-left text-xs">
            <thead className="border-b border-slate-800 text-slate-400 uppercase font-mono">
              <tr>
                <th className="pb-3">Event ID</th>
                <th className="pb-3">Threat Signature</th>
                <th className="pb-3">Source Target</th>
                <th className="pb-3">Gateway Action</th>
                <th className="pb-3 text-right">Time</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-800/60 font-mono">
              {securityEvents.map((e) => (
                <tr key={e.id} className="hover:bg-slate-800/30">
                  <td className="py-3 font-medium text-slate-200">{e.id}</td>
                  <td className="py-3 text-amber-300 font-bold">{e.type}</td>
                  <td className="py-3 text-slate-400 font-sans">{e.source}</td>
                  <td className="py-3">
                    <Badge className="bg-rose-500/20 text-rose-300">{e.action}</Badge>
                  </td>
                  <td className="py-3 text-right text-slate-400 font-sans">{e.time}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </CardContent>
      </Card>

      {/* Hash-Chained Audit Ledger */}
      <Card className="bg-slate-900/80 border-slate-800">
        <CardHeader>
          <CardTitle className="text-base font-semibold text-white flex items-center gap-2">
            <Terminal className="size-4 text-indigo-400" />
            <span>Immutable Hash-Chained Audit Trail</span>
          </CardTitle>
        </CardHeader>
        <CardContent>
          <table className="w-full text-left text-xs font-mono">
            <thead className="border-b border-slate-800 text-slate-400 uppercase">
              <tr>
                <th className="pb-3">Log ID</th>
                <th className="pb-3">Actor</th>
                <th className="pb-3">Action</th>
                <th className="pb-3">Resource Target</th>
                <th className="pb-3">SHA-256 Current Hash</th>
                <th className="pb-3 text-right">Timestamp</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-800/60">
              {auditLedger.map((l) => (
                <tr key={l.id} className="hover:bg-slate-800/30">
                  <td className="py-3 text-slate-300">{l.id}</td>
                  <td className="py-3 text-indigo-300">{l.actor}</td>
                  <td className="py-3 text-emerald-300 font-bold">{l.action}</td>
                  <td className="py-3 text-slate-400">{l.resource}</td>
                  <td className="py-3 text-slate-500">{l.hash}</td>
                  <td className="py-3 text-right text-slate-400 font-sans">{l.timestamp}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </CardContent>
      </Card>
    </div>
  );
}
