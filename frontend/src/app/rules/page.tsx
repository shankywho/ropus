"use client";

import React, { useEffect, useState, useCallback } from "react";
import Link from "next/link";
import { DataProvenanceBadge } from "@/components/ropus/DataProvenanceBadge";
import {
  Sliders,
  PlusCircle,
  Users,
} from "lucide-react";
import { api, Rule, RuleStatus } from "@/lib/api";

export default function RulesManagementPage() {
  const [rules, setRules] = useState<Rule[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [currentAnalyst, setCurrentAnalyst] = useState<string>("analyst_a");
  const [actionLoadingId, setActionLoadingId] = useState<string | null>(null);
  const [statusMessage, setStatusMessage] = useState<{ text: string; type: "success" | "error" } | null>(null);

  const fetchRules = useCallback(async () => {
    setLoading(true);
    try {
      const data = await api.getRules();
      if (data && data.rules) {
        setRules(data.rules);
      }
    } catch (err: any) {
      console.warn("Could not fetch rules from backend:", err.message);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchRules();
  }, [fetchRules]);

  const handleTransition = async (ruleId: string, newStatus: RuleStatus) => {
    setActionLoadingId(ruleId);
    setStatusMessage(null);

    const rule = rules.find((r) => r.rule_id === ruleId);

    // Enforce client-side check for maker-checker
    if (newStatus === "ACTIVE" && rule && rule.created_by === currentAnalyst) {
      setStatusMessage({
        text: `Maker-Checker Violation: ${currentAnalyst} created this rule and cannot approve it. Switch analyst identity above to approve.`,
        type: "error",
      });
      setActionLoadingId(null);
      return;
    }

    try {
      const updatedRule = await api.transitionRule(ruleId, newStatus, currentAnalyst);
      if (updatedRule && updatedRule.rule_id) {
        setRules((prev) => prev.map((r) => (r.rule_id === ruleId ? updatedRule : r)));
        setStatusMessage({
          text: `Rule ${ruleId} transitioned to ${newStatus} by ${currentAnalyst}`,
          type: "success",
        });
      }
    } catch (err: any) {
      setStatusMessage({
        text: err.message || "Failed to update rule status in backend",
        type: "error",
      });
    } finally {
      setActionLoadingId(null);
    }
  };

  const getStatusBadge = (status: RuleStatus) => {
    switch (status) {
      case "ACTIVE":
        return "bg-[#04db7c15] text-[#04db7c] border-[#04db7c33]";
      case "SHADOW":
        return "bg-[#38bdf815] text-[#38bdf8] border-[#38bdf833]";
      case "PENDING_APPROVAL":
        return "bg-[#f59e0b15] text-[#f59e0b] border-[#f59e0b33]";
      case "DRAFT":
        return "bg-[#5e6c8415] text-[#97a0af] border-[#5e6c8433]";
      case "ARCHIVED":
        return "bg-[#f0525215] text-[#f05252] border-[#f0525233]";
      default:
        return "bg-[#5e6c8415] text-[#97a0af] border-[#5e6c8433]";
    }
  };

  return (
    <div className="space-y-5">
      {/* 1. Top Header */}
      <div className="flex flex-wrap items-center justify-between gap-4 pb-3.5 border-b border-[#1c2536]">
        <div>
          <div className="flex items-center gap-2.5">
            <h1 className="text-lg font-bold text-[#f4f5f7] tracking-tight font-mono uppercase">
              Rule Engine &amp; Policy Registry
            </h1>
            <DataProvenanceBadge type="LIVE_BACKEND" sublabel="AST Evaluator" />
          </div>
          <p className="text-xs text-[#5e6c84] font-mono mt-1">
            Deterministic rule precedence, JSON-AST conditions, and dual-analyst maker-checker governance
          </p>
        </div>

        <div className="flex items-center gap-2 font-mono text-xs">
          {/* Analyst Switcher for Dual Control */}
          <div className="flex items-center gap-1.5 bg-[#0f172a] p-1 border border-[#1c2536] rounded-[4px]">
            <Users className="w-3.5 h-3.5 text-[#0d94fb] ml-1" />
            <span className="text-[#5e6c84] text-[10px] uppercase">Actor:</span>
            <select
              value={currentAnalyst}
              onChange={(e) => setCurrentAnalyst(e.target.value)}
              className="bg-[#0a1324] text-[#f4f5f7] border border-[#1c2536] rounded-[2px] px-2 py-0.5 text-xs outline-none cursor-pointer"
            >
              <option value="analyst_a">Analyst A (Author)</option>
              <option value="analyst_b">Analyst B (Reviewer)</option>
              <option value="lead_officer">Risk Lead Officer</option>
            </select>
          </div>

          <Link
            href="/rules/new"
            className="flex items-center gap-1.5 px-3 py-1.5 bg-[#0d94fb] hover:bg-[#0b82dc] text-white font-bold rounded-[4px] shadow-xs active:scale-95 transition-all"
          >
            <PlusCircle className="w-3.5 h-3.5" />
            <span>Create Rule</span>
          </Link>
        </div>
      </div>

      {statusMessage && (
        <div
          className={`p-3 rounded-[4px] border font-mono text-xs flex items-center justify-between ${
            statusMessage.type === "success"
              ? "bg-[#04db7c15] text-[#04db7c] border-[#04db7c44]"
              : "bg-[#f0525215] text-[#f05252] border-[#f0525244]"
          }`}
        >
          <span>{statusMessage.text}</span>
          <button onClick={() => setStatusMessage(null)} className="text-[10px] underline ml-2 cursor-pointer">
            Dismiss
          </button>
        </div>
      )}

      {/* 2. Institutional Rules Table */}
      <div className="bg-[#0f172a] border border-[#1c2536] rounded-[4px] p-4 font-mono text-xs">
        <div className="flex items-center justify-between pb-2.5 mb-3 border-b border-[#1c2536]">
          <div className="flex items-center gap-2">
            <Sliders className="w-4 h-4 text-[#0d94fb]" />
            <h2 className="font-bold uppercase tracking-wider text-[#f4f5f7] text-xs">
              Configured Rule Set ({rules.length})
            </h2>
          </div>
          <span className="text-[#5e6c84] text-[10px]">Deterministic Precedence #1</span>
        </div>

        <div className="overflow-x-auto">
          <table className="w-full text-left">
            <thead>
              <tr className="border-b border-[#1c2536] text-[#5e6c84] text-[11px]">
                <th className="pb-2 font-medium">RULE ID</th>
                <th className="pb-2 font-medium">NAME &amp; DESCRIPTION</th>
                <th className="pb-2 font-medium">ACTION</th>
                <th className="pb-2 font-medium">MAKER</th>
                <th className="pb-2 font-medium">CHECKER</th>
                <th className="pb-2 font-medium">STATUS</th>
                <th className="pb-2 font-medium text-right">GOVERNANCE</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-[#1c2536]">
              {rules.length > 0 ? (
                rules.map((rule) => (
                  <tr key={rule.rule_id} className="hover:bg-[#142036] transition-colors">
                    <td className="py-3 font-semibold text-[#f4f5f7]">{rule.rule_id}</td>
                    <td className="py-3">
                      <span className="font-bold text-[#f4f5f7] block">{rule.name}</span>
                      <span className="text-[11px] text-[#97a0af]">{rule.description}</span>
                    </td>
                    <td className="py-3">
                      <span className="px-1.5 py-0.5 rounded-[2px] bg-[#0a1324] border border-[#1c2536] text-[10px] text-[#f05252] font-bold">
                        {rule.dsl_ast?.action || "MANUAL_REVIEW"}
                      </span>
                    </td>
                    <td className="py-3 text-[#97a0af]">{rule.created_by}</td>
                    <td className="py-3 text-[#97a0af]">{rule.approved_by || "—"}</td>
                    <td className="py-3">
                      <span
                        className={`text-[10px] font-bold px-2 py-0.5 rounded-[2px] border ${getStatusBadge(
                          rule.status
                        )}`}
                      >
                        {rule.status}
                      </span>
                    </td>
                    <td className="py-3 text-right">
                      {rule.status === "DRAFT" && (
                        <button
                          onClick={() => handleTransition(rule.rule_id, "PENDING_APPROVAL")}
                          disabled={actionLoadingId === rule.rule_id}
                          className="px-2 py-1 bg-[#f59e0b] hover:bg-[#d97706] text-[#011638] font-bold rounded-[2px] text-[10px] cursor-pointer"
                        >
                          Submit Review
                        </button>
                      )}

                      {rule.status === "PENDING_APPROVAL" && (
                        <button
                          onClick={() => handleTransition(rule.rule_id, "ACTIVE")}
                          disabled={actionLoadingId === rule.rule_id}
                          className="px-2 py-1 bg-[#04db7c] hover:bg-[#03b868] text-[#011638] font-bold rounded-[2px] text-[10px] cursor-pointer"
                        >
                          Approve (Checker)
                        </button>
                      )}

                      {rule.status === "ACTIVE" && (
                        <button
                          onClick={() => handleTransition(rule.rule_id, "ARCHIVED")}
                          disabled={actionLoadingId === rule.rule_id}
                          className="px-2 py-1 bg-[#0f172a] hover:bg-[#f0525225] text-[#f05252] border border-[#f0525233] font-bold rounded-[2px] text-[10px] cursor-pointer"
                        >
                          Archive
                        </button>
                      )}

                      {rule.status === "ARCHIVED" && (
                        <button
                          onClick={() => handleTransition(rule.rule_id, "PENDING_APPROVAL")}
                          disabled={actionLoadingId === rule.rule_id}
                          className="px-2 py-1 bg-[#0d94fb] hover:bg-[#0b82dc] text-white font-bold rounded-[2px] text-[10px] cursor-pointer"
                        >
                          Re-Enable
                        </button>
                      )}
                    </td>
                  </tr>
                ))
              ) : (
                <tr>
                  <td colSpan={7} className="py-8 text-center text-[#5e6c84]">
                    Loading rules from backend repository...
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}
