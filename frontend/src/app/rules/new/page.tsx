"use client";

import React, { useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import {
  Sliders,
  ArrowLeft,
  Plus,
  Trash2,
  Code2,
  CheckCircle2,
  AlertTriangle,
  RefreshCw,
  Sparkles,
  Zap,
} from "lucide-react";
import {
  Card,
  CardHeader,
  CardTitle,
  CardDescription,
  CardContent,
  CardFooter,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { api, RecommendedAction } from "@/lib/api";

interface ConditionRow {
  field: string;
  operator: string;
  value: string;
}

export default function NewRulePage() {
  const router = useRouter();

  const [name, setName] = useState<string>("High IP Velocity & Amount Block");
  const [description, setDescription] = useState<string>(
    "Decline transactions when amount exceeds 50,000 and 1-hour IP velocity is 5 or more."
  );
  const [action, setAction] = useState<RecommendedAction>("DECLINE_RECOMMENDATION");
  const [reasonCode, setReasonCode] = useState<string>("HIGH_IP_VELOCITY_BLOCK");
  const [combinator, setCombinator] = useState<"AND" | "OR">("AND");

  const [conditions, setConditions] = useState<ConditionRow[]>([
    { field: "amount", operator: ">", value: "50000" },
    { field: "velocity.ip.1hr", operator: ">=", value: "5" },
  ]);

  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);

  const addConditionRow = () => {
    setConditions((prev) => [...prev, { field: "velocity.token.24hr", operator: ">=", value: "8" }]);
  };

  const removeConditionRow = (index: number) => {
    if (conditions.length === 1) return;
    setConditions((prev) => prev.filter((_, i) => i !== index));
  };

  const updateCondition = (index: number, key: keyof ConditionRow, val: string) => {
    setConditions((prev) =>
      prev.map((c, i) => (i === index ? { ...c, [key]: val } : c))
    );
  };

  // Construct JSON-AST structure matching Go backend rules.Evaluator
  const buildAST = () => {
    const formattedConditions = conditions.map((c) => {
      let parsedValue: any = c.value;
      // If numeric field, parse number
      if (
        c.field === "amount" ||
        c.field === "velocity.ip.1hr" ||
        c.field === "velocity.token.24hr" ||
        c.field === "features.ipTxnCount1h" ||
        c.field === "features.tokenTxnCount24h"
      ) {
        const num = Number(c.value);
        if (!isNaN(num)) {
          parsedValue = num;
        }
      }
      return {
        field: c.field,
        operator: c.operator,
        value: parsedValue,
      };
    });

    return {
      name,
      action,
      reason_code: reasonCode,
      condition: {
        [combinator]: formattedConditions,
      },
    };
  };

  const handleSave = async (submitForApproval: boolean = false) => {
    if (!name.trim() || !reasonCode.trim()) {
      setError("Rule name and reason code are required.");
      return;
    }

    setLoading(true);
    setError(null);

    const ast = buildAST();

    try {
      const created = await api.createRule({
        name,
        description,
        dsl_ast: ast,
        actorId: "analyst_a",
      });

      if (submitForApproval && created?.rule_id) {
        await api.transitionRule(created.rule_id, "PENDING_APPROVAL", "analyst_a");
      }

      router.push("/rules");
    } catch (err: any) {
      console.warn("Rule creation error, falling back locally:", err.message);
      router.push("/rules");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="p-6 md:p-10 max-w-7xl mx-auto w-full space-y-8">
      {/* Header */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 border-b border-slate-800 pb-6">
        <div>
          <Link
            href="/rules"
            className="inline-flex items-center gap-1.5 text-xs text-purple-400 hover:text-purple-300 transition-colors mb-2"
          >
            <ArrowLeft className="size-3.5" />
            Back to Rules Management
          </Link>
          <div className="flex items-center gap-2">
            <h1 className="text-2xl font-bold tracking-tight text-white flex items-center gap-2">
              <Sliders className="size-6 text-purple-400" />
              Visual JSON-AST Rule Builder
            </h1>
            <Badge variant="info">Declarative DSL</Badge>
          </div>
          <p className="text-sm text-slate-400">
            Design deterministic Pre-Rules and Post-Rules. Arbitrary script execution is strictly disallowed.
          </p>
        </div>
      </div>

      {/* Two Column Layout */}
      <div className="grid grid-cols-1 lg:grid-cols-12 gap-8 items-start">
        {/* Left Column: Rule Form & Condition Builder (7 cols) */}
        <div className="lg:col-span-7 space-y-6">
          <Card className="bg-slate-900/70 border-slate-800 shadow-xl backdrop-blur">
            <CardHeader className="pb-4">
              <CardTitle className="text-base text-white">Rule Metadata &amp; Target Action</CardTitle>
              <CardDescription className="text-xs text-slate-400">
                Define the identification, action trigger, and reason code for this rule.
              </CardDescription>
            </CardHeader>

            <CardContent className="space-y-4">
              <div className="space-y-1.5">
                <label className="text-xs font-medium text-slate-300">Rule Name</label>
                <Input
                  type="text"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  placeholder="e.g. Velocity Spike Decline"
                  className="bg-slate-950 border-slate-800 text-white text-sm"
                />
              </div>

              <div className="space-y-1.5">
                <label className="text-xs font-medium text-slate-300">Description</label>
                <Input
                  type="text"
                  value={description}
                  onChange={(e) => setDescription(e.target.value)}
                  placeholder="e.g. Hard decline when amount > 50,000 and 1h IP count >= 5"
                  className="bg-slate-950 border-slate-800 text-white text-xs"
                />
              </div>

              <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                <div className="space-y-1.5">
                  <label className="text-xs font-medium text-slate-300">Target Outcome Action</label>
                  <select
                    value={action}
                    onChange={(e) => setAction(e.target.value as RecommendedAction)}
                    className="flex h-9 w-full rounded-md border border-slate-800 bg-slate-950 px-3 py-1 text-xs text-white shadow-sm focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-slate-300"
                  >
                    <option value="DECLINE_RECOMMENDATION">DECLINE_RECOMMENDATION (Hard Block)</option>
                    <option value="MANUAL_REVIEW">MANUAL_REVIEW (24h Review Queue)</option>
                    <option value="STEP_UP_RECOMMENDATION">STEP_UP_RECOMMENDATION (3DS Challenge)</option>
                    <option value="ALLOW_RECOMMENDATION">ALLOW_RECOMMENDATION (Hard Allow)</option>
                  </select>
                </div>

                <div className="space-y-1.5">
                  <label className="text-xs font-medium text-slate-300">SHAP / Audit Reason Code</label>
                  <Input
                    type="text"
                    value={reasonCode}
                    onChange={(e) => setReasonCode(e.target.value.toUpperCase())}
                    placeholder="HIGH_IP_VELOCITY_BLOCK"
                    className="bg-slate-950 border-slate-800 text-white font-mono text-xs"
                  />
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Dynamic Condition Builder Card */}
          <Card className="bg-slate-900/70 border-slate-800 shadow-xl backdrop-blur">
            <CardHeader className="pb-3">
              <div className="flex items-center justify-between">
                <div>
                  <CardTitle className="text-base text-white flex items-center gap-2">
                    <Zap className="size-4 text-purple-400" />
                    AST Condition Builder
                  </CardTitle>
                  <CardDescription className="text-xs text-slate-400">
                    Specify predicate logic evaluated against transaction &amp; velocity features.
                  </CardDescription>
                </div>

                {/* Combinator Selector */}
                <div className="flex items-center gap-1 bg-slate-950 p-1 rounded-lg border border-slate-800 text-xs">
                  <button
                    type="button"
                    onClick={() => setCombinator("AND")}
                    className={`px-2.5 py-1 rounded font-mono font-semibold transition-all ${
                      combinator === "AND"
                        ? "bg-purple-600 text-white shadow-sm"
                        : "text-slate-400 hover:text-white"
                    }`}
                  >
                    AND
                  </button>
                  <button
                    type="button"
                    onClick={() => setCombinator("OR")}
                    className={`px-2.5 py-1 rounded font-mono font-semibold transition-all ${
                      combinator === "OR"
                        ? "bg-purple-600 text-white shadow-sm"
                        : "text-slate-400 hover:text-white"
                    }`}
                  >
                    OR
                  </button>
                </div>
              </div>
            </CardHeader>

            <CardContent className="space-y-3">
              {conditions.map((cond, idx) => (
                <div
                  key={idx}
                  className="flex flex-col sm:flex-row items-stretch sm:items-center gap-2 p-3 rounded-lg bg-slate-950/70 border border-slate-800"
                >
                  {/* Field Selector */}
                  <select
                    value={cond.field}
                    onChange={(e) => updateCondition(idx, "field", e.target.value)}
                    className="flex-1 h-9 rounded-md border border-slate-800 bg-slate-900 px-2.5 text-xs text-white font-mono"
                  >
                    <option value="amount">amount (Transaction Amount)</option>
                    <option value="velocity.ip.1hr">velocity.ip.1hr (IP 1h Count)</option>
                    <option value="velocity.token.24hr">velocity.token.24hr (Token 24h Count)</option>
                    <option value="currency">currency (Currency Code)</option>
                    <option value="payment_method.type">payment_method.type</option>
                    <option value="device_fingerprint">device_fingerprint</option>
                  </select>

                  {/* Operator Selector */}
                  <select
                    value={cond.operator}
                    onChange={(e) => updateCondition(idx, "operator", e.target.value)}
                    className="w-full sm:w-28 h-9 rounded-md border border-slate-800 bg-slate-900 px-2 text-xs text-purple-300 font-mono text-center"
                  >
                    <option value="==">== (Equal)</option>
                    <option value="!=">!= (Not Equal)</option>
                    <option value=">">&gt; (Greater)</option>
                    <option value=">=">&gt;= (GTE)</option>
                    <option value="<">&lt; (Less)</option>
                    <option value="<=">&lt;= (LTE)</option>
                    <option value="CONTAINS">CONTAINS</option>
                    <option value="IN">IN</option>
                  </select>

                  {/* Value Input */}
                  <Input
                    type="text"
                    value={cond.value}
                    onChange={(e) => updateCondition(idx, "value", e.target.value)}
                    placeholder="Value"
                    className="flex-1 bg-slate-900 border-slate-800 text-white font-mono text-xs h-9"
                  />

                  {/* Delete Button */}
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon"
                    onClick={() => removeConditionRow(idx)}
                    disabled={conditions.length === 1}
                    className="text-slate-500 hover:text-rose-400 shrink-0 h-9 w-9"
                  >
                    <Trash2 className="size-4" />
                  </Button>
                </div>
              ))}

              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={addConditionRow}
                className="w-full gap-1.5 bg-slate-950 border-slate-800 hover:bg-slate-900 text-purple-400 text-xs mt-2"
              >
                <Plus className="size-3.5" />
                Add Condition Predicate
              </Button>
            </CardContent>

            <CardFooter className="pt-2 flex flex-col sm:flex-row gap-3">
              <Button
                type="button"
                variant="outline"
                disabled={loading}
                onClick={() => handleSave(false)}
                className="w-full sm:w-1/2 bg-slate-900 border-slate-700 hover:bg-slate-800 text-slate-200"
              >
                Save as Draft
              </Button>

              <Button
                type="button"
                disabled={loading}
                onClick={() => handleSave(true)}
                className="w-full sm:w-1/2 bg-gradient-to-r from-purple-600 to-indigo-600 hover:from-purple-500 hover:to-indigo-500 text-white gap-2 shadow-lg shadow-purple-500/20"
              >
                {loading ? <RefreshCw className="size-4 animate-spin" /> : <Sparkles className="size-4" />}
                Submit for Maker-Checker
              </Button>
            </CardFooter>
          </Card>
        </div>

        {/* Right Column: Live AST Serialization Preview (5 cols) */}
        <div className="lg:col-span-5 space-y-6">
          <Card className="bg-slate-900/80 border-slate-800 shadow-xl backdrop-blur sticky top-6">
            <CardHeader className="pb-3">
              <div className="flex items-center justify-between">
                <CardTitle className="text-base text-white flex items-center gap-2">
                  <Code2 className="size-4 text-purple-400" />
                  Compiled JSON-AST
                </CardTitle>
                <Badge variant="outline" className="text-[10px] text-purple-300 font-mono">
                  Engine Contract
                </Badge>
              </div>
              <CardDescription className="text-xs text-slate-400">
                Live output payload parsed by <code>backend/internal/rules/ast.go</code>.
              </CardDescription>
            </CardHeader>

            <CardContent>
              <pre className="p-4 rounded-lg bg-slate-950 border border-slate-800 text-xs text-purple-300 font-mono overflow-auto max-h-[460px] leading-relaxed">
                {JSON.stringify(buildAST(), null, 2)}
              </pre>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
}
