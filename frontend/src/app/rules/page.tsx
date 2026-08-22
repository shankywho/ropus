"use client";

import React, { useEffect, useState } from "react";
import Link from "next/link";
import {
  Sliders,
  PlusCircle,
  ShieldCheck,
  ShieldAlert,
  AlertTriangle,
  CheckCircle2,
  XCircle,
  Users,
  RefreshCw,
  ArrowRight,
  Code2,
  Lock,
  Info,
} from "lucide-react";
import {
  Card,
  CardHeader,
  CardTitle,
  CardDescription,
  CardContent,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
  Table,
  TableHeader,
  TableBody,
  TableHead,
  TableRow,
  TableCell,
} from "@/components/ui/table";
import { api, Rule, RuleStatus } from "@/lib/api";

export default function RulesManagementPage() {
  const [rules, setRules] = useState<Rule[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [currentAnalyst, setCurrentAnalyst] = useState<string>("analyst_a");
  const [actionLoadingId, setActionLoadingId] = useState<string | null>(null);
  const [statusMessage, setStatusMessage] = useState<{ text: string; type: "success" | "error" } | null>(null);

  // Fallback demo mock rules
  const mockFallbackRules: Rule[] = [
    {
      rule_id: "rule_01_high_velocity_block",
      tenant_id: "00000000-0000-0000-0000-000000000001",
      name: "High IP Velocity & Amount Block",
      description: "Hard decline if amount > 50,000 and 1h IP transactions >= 5",
      dsl_ast: {
        action: "DECLINE_RECOMMENDATION",
        reason_code: "HIGH_IP_VELOCITY_BLOCK",
        condition: {
          AND: [
            { field: "amount", operator: ">", value: 50000 },
            { field: "velocity.ip.1hr", operator: ">=", value: 5 },
          ],
        },
      },
      status: "ACTIVE",
      version: 1,
      created_by: "analyst_a",
      approved_by: "analyst_b",
      created_at: new Date(Date.now() - 72 * 3600 * 1000).toISOString(),
      updated_at: new Date(Date.now() - 48 * 3600 * 1000).toISOString(),
    },
    {
      rule_id: "rule_02_new_device_review",
      tenant_id: "00000000-0000-0000-0000-000000000001",
      name: "Untrusted Device High Ticket Review",
      description: "Send to analyst queue if device is new and amount > 30,000",
      dsl_ast: {
        action: "MANUAL_REVIEW",
        reason_code: "NEW_DEVICE_HIGH_TICKET",
        condition: {
          AND: [
            { field: "device_fingerprint", operator: "==", value: "new_device" },
            { field: "amount", operator: ">", value: 30000 },
          ],
        },
      },
      status: "PENDING_APPROVAL",
      version: 1,
      created_by: "analyst_a",
      created_at: new Date(Date.now() - 4 * 3600 * 1000).toISOString(),
      updated_at: new Date(Date.now() - 4 * 3600 * 1000).toISOString(),
    },
    {
      rule_id: "rule_03_token_burst_draft",
      tenant_id: "00000000-0000-0000-0000-000000000001",
      name: "Token 24h Velocity Spike Guard",
      description: "Flag card tokens with over 10 transactions in 24 hours",
      dsl_ast: {
        action: "MANUAL_REVIEW",
        reason_code: "TOKEN_VELOCITY_SPIKE",
        condition: {
          field: "velocity.token.24hr",
          operator: ">=",
          value: 10,
        },
      },
      status: "DRAFT",
      version: 1,
      created_by: "analyst_b",
      created_at: new Date(Date.now() - 1 * 3600 * 1000).toISOString(),
      updated_at: new Date(Date.now() - 1 * 3600 * 1000).toISOString(),
    },
  ];

  const fetchRules = async () => {
    setLoading(true);
    try {
      const data = await api.getRules();
      if (data && data.rules && data.rules.length > 0) {
        setRules(data.rules);
      } else {
        setRules(mockFallbackRules);
      }
    } catch (err: any) {
      console.warn("Could not fetch rules from backend, using fallback list:", err.message);
      setRules(mockFallbackRules);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchRules();
  }, []);

  const handleTransition = async (ruleId: string, newStatus: RuleStatus) => {
    setActionLoadingId(ruleId);
    setStatusMessage(null);

    const rule = rules.find((r) => r.rule_id === ruleId);

    // Enforce client-side check for maker-checker
    if (newStatus === "ACTIVE" && rule && rule.created_by === currentAnalyst) {
      setStatusMessage({
        text: `Maker-Checker Violation: ${currentAnalyst} created this rule and cannot approve it. Switch analyst above to approve.`,
        type: "error",
      });
      setActionLoadingId(null);
      return;
    }

    try {
      await api.transitionRule(ruleId, newStatus, currentAnalyst);
      setStatusMessage({
        text: `Rule successfully transitioned to ${newStatus} by ${currentAnalyst}.`,
        type: "success",
      });
      // Update local state
      setRules((prev) =>
        prev.map((r) =>
          r.rule_id === ruleId
            ? {
                ...r,
                status: newStatus,
                approved_by: newStatus === "ACTIVE" ? currentAnalyst : r.approved_by,
              }
            : r
        )
      );
    } catch (err: any) {
      console.warn("Transition API failed:", err.message);
      setStatusMessage({
        text: err.message || `Failed to transition rule.`,
        type: "error",
      });
    } finally {
      setActionLoadingId(null);
    }
  };

  const renderStatusBadge = (status: RuleStatus) => {
    switch (status) {
      case "ACTIVE":
        return (
          <Badge variant="success" className="gap-1 bg-emerald-500/15 text-emerald-400 border-emerald-500/30">
            <CheckCircle2 className="size-3" />
            ACTIVE
          </Badge>
        );
      case "PENDING_APPROVAL":
        return (
          <Badge variant="warning" className="gap-1 bg-amber-500/15 text-amber-400 border-amber-500/30">
            <AlertTriangle className="size-3" />
            PENDING APPROVAL
          </Badge>
        );
      case "DRAFT":
        return (
          <Badge variant="outline" className="gap-1 border-slate-700 text-slate-400">
            DRAFT
          </Badge>
        );
      case "SHADOW":
        return (
          <Badge variant="info" className="gap-1 bg-blue-500/15 text-blue-400 border-blue-500/30">
            SHADOW
          </Badge>
        );
      default:
        return <Badge variant="outline">{status}</Badge>;
    }
  };

  const formatConditionPreview = (dsl: Record<string, any>) => {
    if (!dsl) return "Always True";
    const cond = dsl.condition;
    if (!cond) return "Always True";

    if (cond.AND && Array.isArray(cond.AND)) {
      return cond.AND.map((c: any) => `${c.field} ${c.operator} ${c.value}`).join(" AND ");
    }
    if (cond.OR && Array.isArray(cond.OR)) {
      return cond.OR.map((c: any) => `${c.field} ${c.operator} ${c.value}`).join(" OR ");
    }
    if (cond.field) {
      return `${cond.field} ${cond.operator} ${cond.value}`;
    }
    return JSON.stringify(cond);
  };

  return (
    <div className="p-6 md:p-10 max-w-7xl mx-auto w-full space-y-8">
      {/* Header */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 border-b border-slate-800 pb-6">
        <div>
          <div className="flex items-center gap-2 mb-1">
            <h1 className="text-2xl font-bold tracking-tight text-white flex items-center gap-2">
              <Sliders className="size-6 text-purple-400" />
              Rules Governance &amp; Maker-Checker
            </h1>
            <Badge variant="info">AST Engine</Badge>
          </div>
          <p className="text-sm text-slate-400">
            Declarative JSON-AST business rules with strict dual-control Maker-Checker authorization enforcement.
          </p>
        </div>

        {/* Top Controls: Analyst Switcher & Create Rule Button */}
        <div className="flex flex-wrap items-center gap-3">
          {/* Analyst Role Switcher */}
          <div className="flex items-center gap-2 p-1.5 bg-slate-900 border border-slate-800 rounded-lg text-xs">
            <Users className="size-4 text-slate-400 ml-1" />
            <span className="text-slate-400 text-[11px]">Active Persona:</span>
            <button
              type="button"
              onClick={() => setCurrentAnalyst("analyst_a")}
              className={`px-2.5 py-1 rounded text-xs font-mono font-medium transition-all ${
                currentAnalyst === "analyst_a"
                  ? "bg-indigo-600 text-white shadow-sm"
                  : "text-slate-400 hover:text-white"
              }`}
            >
              Analyst A (Creator)
            </button>
            <button
              type="button"
              onClick={() => setCurrentAnalyst("analyst_b")}
              className={`px-2.5 py-1 rounded text-xs font-mono font-medium transition-all ${
                currentAnalyst === "analyst_b"
                  ? "bg-purple-600 text-white shadow-sm"
                  : "text-slate-400 hover:text-white"
              }`}
            >
              Analyst B (Checker)
            </button>
          </div>

          <Link href="/rules/new">
            <Button className="bg-gradient-to-r from-purple-600 to-indigo-600 hover:from-purple-500 hover:to-indigo-500 text-white gap-2 shadow-lg shadow-purple-500/20">
              <PlusCircle className="size-4" />
              Create New Rule
            </Button>
          </Link>
        </div>
      </div>

      {/* Status Alert Banner */}
      {statusMessage && (
        <div
          className={`p-3 rounded-lg border text-xs flex items-center justify-between ${
            statusMessage.type === "error"
              ? "bg-rose-500/10 border-rose-500/30 text-rose-300"
              : "bg-emerald-500/10 border-emerald-500/30 text-emerald-300"
          }`}
        >
          <div className="flex items-center gap-2">
            {statusMessage.type === "error" ? (
              <ShieldAlert className="size-4 text-rose-400 shrink-0" />
            ) : (
              <CheckCircle2 className="size-4 text-emerald-400 shrink-0" />
            )}
            <span>{statusMessage.text}</span>
          </div>
          <button
            type="button"
            onClick={() => setStatusMessage(null)}
            className="text-slate-400 hover:text-white text-xs ml-4"
          >
            &times;
          </button>
        </div>
      )}

      {/* Rules Table Card */}
      <Card className="bg-slate-900/70 border-slate-800 shadow-xl overflow-hidden backdrop-blur">
        <CardHeader className="pb-3">
          <div className="flex items-center justify-between">
            <CardTitle className="text-base text-white">Configured Risk Rules</CardTitle>
            <Button
              variant="ghost"
              size="sm"
              onClick={fetchRules}
              disabled={loading}
              className="text-xs text-slate-400 hover:text-white gap-1.5 h-8"
            >
              <RefreshCw className={`size-3.5 ${loading ? "animate-spin" : ""}`} />
              Refresh
            </Button>
          </div>
          <CardDescription className="text-xs text-slate-400">
            Rules execute in real-time as Pre-Rules (Hard blocks/allows) or Post-Rules (Risk score thresholding).
          </CardDescription>
        </CardHeader>

        <CardContent className="p-0">
          <Table>
            <TableHeader className="bg-slate-950/70 border-b border-slate-800">
              <TableRow className="border-slate-800 hover:bg-transparent">
                <TableHead className="text-slate-400 font-semibold text-xs py-3.5 pl-6">Rule Name &amp; Description</TableHead>
                <TableHead className="text-slate-400 font-semibold text-xs py-3.5">Condition AST Preview</TableHead>
                <TableHead className="text-slate-400 font-semibold text-xs py-3.5">Action</TableHead>
                <TableHead className="text-slate-400 font-semibold text-xs py-3.5">Status</TableHead>
                <TableHead className="text-slate-400 font-semibold text-xs py-3.5">Creator / Approver</TableHead>
                <TableHead className="text-slate-400 font-semibold text-xs py-3.5 text-right pr-6">Maker-Checker Action</TableHead>
              </TableRow>
            </TableHeader>

            <TableBody>
              {loading ? (
                <TableRow className="border-slate-800">
                  <TableCell colSpan={6} className="text-center py-12 text-slate-400">
                    <RefreshCw className="size-5 animate-spin mx-auto mb-2 text-purple-400" />
                    Loading rules...
                  </TableCell>
                </TableRow>
              ) : rules.length > 0 ? (
                rules.map((rule) => {
                  const isCreator = rule.created_by === currentAnalyst;

                  return (
                    <TableRow
                      key={rule.rule_id}
                      className="border-slate-800/80 hover:bg-slate-800/40 transition-colors"
                    >
                      {/* Name & Description */}
                      <TableCell className="pl-6 py-4 max-w-xs">
                        <p className="text-sm font-semibold text-white truncate">{rule.name}</p>
                        <p className="text-xs text-slate-400 truncate mt-0.5">{rule.description}</p>
                      </TableCell>

                      {/* Condition AST Preview */}
                      <TableCell className="max-w-xs font-mono text-xs text-purple-300">
                        <span className="bg-slate-950 px-2 py-1 rounded border border-slate-800 truncate block">
                          {formatConditionPreview(rule.dsl_ast)}
                        </span>
                      </TableCell>

                      {/* Action */}
                      <TableCell>
                        <Badge
                          variant={
                            rule.dsl_ast?.action === "DECLINE_RECOMMENDATION"
                              ? "danger"
                              : rule.dsl_ast?.action === "MANUAL_REVIEW"
                              ? "warning"
                              : "success"
                          }
                          className="text-[10px] font-mono"
                        >
                          {rule.dsl_ast?.action || "ALLOW"}
                        </Badge>
                      </TableCell>

                      {/* Status */}
                      <TableCell>{renderStatusBadge(rule.status)}</TableCell>

                      {/* Creator / Approver */}
                      <TableCell className="text-xs font-mono">
                        <p className="text-slate-300">
                          By: <span className="text-indigo-400">{rule.created_by}</span>
                        </p>
                        {rule.approved_by && (
                          <p className="text-slate-400 text-[11px]">
                            Appr: <span className="text-emerald-400">{rule.approved_by}</span>
                          </p>
                        )}
                      </TableCell>

                      {/* Maker-Checker Actions */}
                      <TableCell className="text-right pr-6">
                        {rule.status === "DRAFT" && (
                          <Button
                            size="sm"
                            disabled={actionLoadingId === rule.rule_id}
                            onClick={() => handleTransition(rule.rule_id, "PENDING_APPROVAL")}
                            className="h-7 px-3 text-xs bg-amber-600 hover:bg-amber-500 text-white"
                          >
                            Submit for Approval
                          </Button>
                        )}

                        {rule.status === "PENDING_APPROVAL" && (
                          <div className="flex items-center justify-end gap-2">
                            {isCreator ? (
                              <span
                                title="Dual-Control Violation: You created this rule and cannot approve it."
                                className="text-[11px] font-mono text-rose-400 bg-rose-500/10 border border-rose-500/20 px-2 py-1 rounded flex items-center gap-1 cursor-not-allowed"
                              >
                                <Lock className="size-3" /> Cannot self-approve
                              </span>
                            ) : (
                              <Button
                                size="sm"
                                disabled={actionLoadingId === rule.rule_id}
                                onClick={() => handleTransition(rule.rule_id, "ACTIVE")}
                                className="h-7 px-3 text-xs bg-emerald-600 hover:bg-emerald-500 text-white gap-1 shadow-sm"
                              >
                                <CheckCircle2 className="size-3" />
                                Approve &amp; Activate
                              </Button>
                            )}
                          </div>
                        )}

                        {rule.status === "ACTIVE" && (
                          <span className="text-[11px] font-mono text-emerald-400 flex items-center justify-end gap-1">
                            <CheckCircle2 className="size-3" /> Live in Pipeline
                          </span>
                        )}
                      </TableCell>
                    </TableRow>
                  );
                })
              ) : (
                <TableRow className="border-slate-800">
                  <TableCell colSpan={6} className="text-center py-12 text-slate-500">
                    No rules found. Click &quot;Create New Rule&quot; to build your first AST rule.
                  </TableCell>
                </TableRow>
              )}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </div>
  );
}
