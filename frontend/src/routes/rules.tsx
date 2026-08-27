import { createFileRoute, Link } from "@tanstack/react-router";
import { useState, useMemo } from "react";
import {
  DataGrid,
  Page,
  PageHead,
  SectionHead,
  TelemetryStrip,
} from "@/components/ropus/page";
import { Mono, StatusPill } from "@/components/ropus/core";
import { rules, ruleDetails, type RuleRecord } from "@/lib/ropus/platform-fixtures";
import { cn } from "@/lib/utils";
import {
  FileCode2,
  Play,
  ShieldCheck,
  AlertTriangle,
  Plus,
  Trash2,
  CheckCircle,
  Clock,
  Lock,
  Sparkles,
  Layers,
  Code,
  Save,
  CheckCircle2,
} from "lucide-react";
import { toast } from "sonner";

export const Route = createFileRoute("/rules")({
  head: () => ({
    meta: [
      { title: "Rules Engine & AST Policy Composer — ROPUS" },
      {
        name: "description",
        content:
          "Deterministic JSON-AST policy builder, Maker-Checker dual control workflow, historical dry-run backtesting, and sub-millisecond rule evaluation.",
      },
      { property: "og:title", content: "Rules Engine & AST Composer — ROPUS" },
      {
        property: "og:description",
        content: "Deterministic JSON-AST decisioning rules with Maker-Checker dual control and backtester.",
      },
      { property: "og:type", content: "website" },
      { name: "twitter:card", content: "summary" },
    ],
  }),
  component: RulesPage,
});

const actionTone: Record<RuleRecord["action"], string> = {
  BLOCK: "text-blocked font-semibold",
  REVIEW: "text-amber-intel font-semibold",
  CHALLENGE: "text-local font-semibold",
  SCORE: "text-muted-foreground",
};

const filters = ["ALL", "ENABLED", "SHADOW", "DISABLED"] as const;

const CANONICAL_FEATURES = [
  "velocity.ip.1hr",
  "velocity.device.1hr",
  "velocity.account.24hr",
  "geo_distance_km",
  "device_novelty_score",
  "amount_to_avg_ratio",
  "beneficiary_age_hours",
  "ip_datacenter_proxy",
  "payout_mule_cluster_size",
  "failed_pin_attempts_24h",
];

const OPERATORS = [
  { value: "GT", label: "> (Greater Than)" },
  { value: "LT", label: "< (Less Than)" },
  { value: "EQ", label: "== (Equals)" },
  { value: "GTE", label: ">= (Greater or Equal)" },
  { value: "IN_LIST", label: "IN (List Membership)" },
  { value: "GEO_DISTANCE_KM", label: "GEO_DISTANCE > KM" },
  { value: "MATCHES_REGEX", label: "REGEX MATCH" },
];

function StateText({ state }: { state: RuleRecord["state"] }) {
  if (state === "ENABLED") {
    return <StatusPill tone="authoritative">ENABLED</StatusPill>;
  }
  if (state === "SHADOW") {
    return <StatusPill tone="shadow">SHADOW</StatusPill>;
  }
  return <StatusPill tone="neutral">DISABLED</StatusPill>;
}

export function RulesPage() {
  const [filter, setFilter] = useState<(typeof filters)[number]>("ALL");
  const [selectedId, setSelectedId] = useState<string | null>("rule_vel_geo_impossible_v1");
  const [composerOpen, setComposerOpen] = useState<boolean>(false);
  const [composerMode, setComposerMode] = useState<"visual" | "json">("visual");

  // Rule Composer state
  const [ruleName, setRuleName] = useState("RULE_HIGH_VELOCITY_PROXY_PAYOUT");
  const [ruleAction, setRuleAction] = useState<"BLOCK" | "REVIEW" | "CHALLENGE">("BLOCK");
  const [conditions, setConditions] = useState([
    { field: "velocity.ip.1hr", op: "GT", value: "8" },
    { field: "ip_datacenter_proxy", op: "EQ", value: "true" },
    { field: "amount_to_avg_ratio", op: "GT", value: "5.0" },
  ]);
  const [jsonAst, setJsonAst] = useState(
    JSON.stringify(
      {
        and: [
          { field: "velocity.ip.1hr", op: ">", value: 8 },
          { field: "ip_datacenter_proxy", op: "==", value: true },
          { field: "amount_to_avg_ratio", op: ">", value: 5.0 },
        ],
      },
      null,
      2,
    ),
  );

  // Backtest State
  const [backtestRunning, setBacktestRunning] = useState(false);
  const [backtestResult, setBacktestResult] = useState<{
    testedCount: number;
    hitCount: number;
    hitRate: string;
    precision: string;
    estimatedFpMonthly: string;
  } | null>(null);

  const rows = rules.filter((r) => filter === "ALL" || r.state === filter);
  const selected = rows.find((r) => r.id === selectedId) ?? rules[0]!;
  const detail = selected ? ruleDetails[selected.id] : undefined;

  const handleAddCondition = () => {
    setConditions((prev) => [...prev, { field: "velocity.account.24hr", op: "GT", value: "10" }]);
  };

  const handleRemoveCondition = (index: number) => {
    setConditions((prev) => prev.filter((_, i) => i !== index));
  };

  const handleRunBacktest = () => {
    setBacktestRunning(true);
    setTimeout(() => {
      setBacktestRunning(false);
      setBacktestResult({
        testedCount: 100000,
        hitCount: 184,
        hitRate: "0.184%",
        precision: "99.45%",
        estimatedFpMonthly: "₹38,500 INR",
      });
      toast.success("Dry-Run Backtest Completed on 100k Transactions", {
        description: "Zero false-positive critical breaches projected across 90-day holdout dataset.",
      });
    }, 600);
  };

  const handleSubmitForApproval = () => {
    toast.success("Rule submitted for Maker-Checker Dual Control Approval", {
      description: "State moved to PENDING_APPROVAL. Self-approval blocked by policy.",
    });
    setComposerOpen(false);
  };

  return (
    <Page>
      <PageHead
        title="Deterministic Rules Engine"
        subtitle="Deterministic JSON-AST conditions evaluated before the ML model on every request. Evaluated in <0.5ms with strict Maker-Checker dual control."
      />

      <TelemetryStrip
        items={[
          { label: "Rules", value: String(rules.length), sub: "in this tenant" },
          {
            label: "Enabled",
            value: String(rules.filter((r) => r.state === "ENABLED").length),
            sub: "scoring live traffic",
            tone: "text-authoritative",
          },
          {
            label: "Shadow",
            value: String(rules.filter((r) => r.state === "SHADOW").length),
            sub: "evaluated, non-blocking",
            tone: "text-amber-intel",
          },
          {
            label: "Hits (24h)",
            value: rules.reduce((s, r) => s + r.hits24h, 0).toLocaleString(),
            sub: "active triggers",
          },
          {
            label: "Hard Blocks",
            value: String(rules.filter((r) => r.action === "BLOCK").length),
            sub: "instant decline",
            tone: "text-blocked",
          },
        ]}
      />

      {/* Action Bar */}
      <div className="mt-4 flex flex-wrap items-center justify-between gap-3 border-b border-border pb-3">
        <div className="flex items-center gap-1 font-mono text-[11px]">
          {filters.map((f) => (
            <button
              key={f}
              type="button"
              onClick={() => setFilter(f)}
              className={cn(
                "px-2.5 py-1 transition-colors cursor-pointer border",
                filter === f
                  ? "border-navy bg-navy text-white font-bold"
                  : "border-transparent text-muted-foreground hover:text-foreground",
              )}
            >
              {f}
            </button>
          ))}
        </div>

        <button
          type="button"
          onClick={() => setComposerOpen(!composerOpen)}
          className="border border-navy bg-navy px-3 py-1.5 font-mono text-[11px] font-bold text-white hover:bg-navy/90 transition-opacity flex items-center gap-1.5 cursor-pointer"
        >
          <Plus className="size-3.5" />
          <span>{composerOpen ? "Close Composer" : "+ Compose New AST Rule"}</span>
        </button>
      </div>

      {/* ------------------------------------------------ VISUAL AST RULE COMPOSER MODAL / DRAWER */}
      {composerOpen && (
        <section className="mt-4 border border-navy bg-card p-5 shadow-lg space-y-4 animate-in fade-in duration-150">
          <div className="flex flex-wrap items-center justify-between border-b border-border pb-3 gap-2">
            <div className="flex items-center gap-2">
              <span className="border border-navy bg-navy/10 px-2 py-0.5 font-mono text-[10px] font-bold text-navy uppercase tracking-[0.06em]">
                RULE COMPOSER
              </span>
              <h3 className="font-sans text-[16px] font-bold text-foreground">
                Dual-Mode JSON-AST Policy Builder
              </h3>
            </div>
            <div className="flex items-center gap-1 border border-border bg-surface p-0.5 font-mono text-[10.5px]">
              <button
                type="button"
                onClick={() => setComposerMode("visual")}
                className={cn(
                  "px-2.5 py-0.5 cursor-pointer",
                  composerMode === "visual" ? "bg-navy text-white font-bold" : "text-muted-foreground",
                )}
              >
                Visual Nodes
              </button>
              <button
                type="button"
                onClick={() => setComposerMode("json")}
                className={cn(
                  "px-2.5 py-0.5 cursor-pointer",
                  composerMode === "json" ? "bg-navy text-white font-bold" : "text-muted-foreground",
                )}
              >
                Monaco JSON-AST
              </button>
            </div>
          </div>

          <div className="grid gap-6 lg:grid-cols-[1fr_360px]">
            {/* Left: Visual / JSON Builder */}
            <div className="space-y-4 font-mono text-[11.5px]">
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="text-[10px] text-muted-foreground uppercase block font-bold">Rule Identifier</label>
                  <input
                    type="text"
                    value={ruleName}
                    onChange={(e) => setRuleName(e.target.value)}
                    className="mt-1 w-full border border-border bg-surface px-2.5 py-1 text-foreground font-mono text-[12px]"
                  />
                </div>
                <div>
                  <label className="text-[10px] text-muted-foreground uppercase block font-bold">Enforcement Action</label>
                  <select
                    value={ruleAction}
                    onChange={(e) => setRuleAction(e.target.value as any)}
                    className="mt-1 w-full border border-border bg-surface px-2.5 py-1 text-foreground font-mono text-[12px]"
                  >
                    <option value="BLOCK">BLOCK (Hard Decline)</option>
                    <option value="REVIEW">REVIEW (Manual Case Queue)</option>
                    <option value="CHALLENGE">CHALLENGE (Step-Up OTP)</option>
                  </select>
                </div>
              </div>

              {composerMode === "visual" ? (
                <div className="space-y-2 border border-border bg-surface/50 p-3.5">
                  <div className="flex items-center justify-between text-[10px] text-muted-foreground uppercase font-bold">
                    <span>Boolean Conjunction: AND Logic Block</span>
                    <span>{conditions.length} Conditions</span>
                  </div>

                  <div className="space-y-2">
                    {conditions.map((cond, idx) => (
                      <div key={idx} className="flex items-center gap-2 border border-border bg-card p-2">
                        <select
                          value={cond.field}
                          onChange={(e) => {
                            const newConds = [...conditions];
                            newConds[idx].field = e.target.value;
                            setConditions(newConds);
                          }}
                          className="border border-border bg-surface px-2 py-1 text-[11px] font-mono flex-1"
                        >
                          {CANONICAL_FEATURES.map((f) => (
                            <option key={f} value={f}>{f}</option>
                          ))}
                        </select>

                        <select
                          value={cond.op}
                          onChange={(e) => {
                            const newConds = [...conditions];
                            newConds[idx].op = e.target.value;
                            setConditions(newConds);
                          }}
                          className="border border-border bg-surface px-2 py-1 text-[11px] font-mono w-[140px]"
                        >
                          {OPERATORS.map((op) => (
                            <option key={op.value} value={op.value}>{op.label}</option>
                          ))}
                        </select>

                        <input
                          type="text"
                          value={cond.value}
                          onChange={(e) => {
                            const newConds = [...conditions];
                            newConds[idx].value = e.target.value;
                            setConditions(newConds);
                          }}
                          className="border border-border bg-surface px-2 py-1 text-[11px] font-mono w-[90px]"
                        />

                        <button
                          type="button"
                          onClick={() => handleRemoveCondition(idx)}
                          className="p-1 text-muted-foreground hover:text-blocked cursor-pointer"
                        >
                          <Trash2 className="size-3.5" />
                        </button>
                      </div>
                    ))}
                  </div>

                  <button
                    type="button"
                    onClick={handleAddCondition}
                    className="mt-2 text-[10.5px] text-navy font-bold hover:underline cursor-pointer flex items-center gap-1"
                  >
                    <Plus className="size-3" /> Add Condition Clause
                  </button>
                </div>
              ) : (
                <textarea
                  value={jsonAst}
                  onChange={(e) => setJsonAst(e.target.value)}
                  rows={8}
                  className="w-full border border-border bg-surface p-3 font-mono text-[11.5px] text-foreground"
                />
              )}
            </div>

            {/* Right: Dry-Run Backtest & Maker-Checker Governance */}
            <div className="space-y-4 border-l border-border/70 pl-5 font-mono text-[11px]">
              {/* Backtest Section */}
              <div className="border border-border bg-surface p-3 space-y-2">
                <div className="flex items-center justify-between text-[10px] text-muted-foreground uppercase font-bold">
                  <span>Dry-Run Backtester</span>
                  <span className="text-authoritative">IEEE-CIS Fixture</span>
                </div>
                <p className="text-[10.5px] text-muted-foreground font-sans">
                  Evaluate projected block rate &amp; false-positive friction against 100,000 historical transactions.
                </p>
                <button
                  type="button"
                  onClick={handleRunBacktest}
                  disabled={backtestRunning}
                  className="w-full border border-border bg-card hover:bg-secondary py-1.5 text-foreground font-bold flex items-center justify-center gap-1.5 cursor-pointer disabled:opacity-50"
                >
                  <Play className="size-3" />
                  <span>{backtestRunning ? "Simulating on 100k Events..." : "Run Dry-Run Simulation"}</span>
                </button>

                {backtestResult && (
                  <div className="border-t border-border pt-2 space-y-1 text-[10.5px]">
                    <div className="flex justify-between"><span>Projected Hits:</span><span className="font-bold text-foreground">{backtestResult.hitCount} ({backtestResult.hitRate})</span></div>
                    <div className="flex justify-between"><span>Precision:</span><span className="font-bold text-authoritative">{backtestResult.precision}</span></div>
                    <div className="flex justify-between"><span>Est. FP Friction:</span><span className="font-bold text-navy">{backtestResult.estimatedFpMonthly}</span></div>
                  </div>
                )}
              </div>

              {/* Maker-Checker Governance Safeguard */}
              <div className="border border-amber-intel/40 bg-shadow-intel-surface p-3 space-y-1.5">
                <div className="flex items-center gap-1.5 text-shadow-intel font-bold text-[10.5px]">
                  <AlertTriangle className="size-3.5" />
                  <span>Maker-Checker Dual Control Policy</span>
                </div>
                <div className="text-[10px] text-muted-foreground font-sans leading-snug">
                  Rules created by <strong>shankar.r</strong> cannot be activated without independent second-officer cryptographic sign-off.
                </div>
                <div className="border-t border-border/40 pt-1 text-[9.5px] text-muted-foreground">
                  Workflow: DRAFT → PENDING_APPROVAL → ACTIVE
                </div>
              </div>

              <button
                type="button"
                onClick={handleSubmitForApproval}
                className="w-full border border-navy bg-navy py-2 text-white font-bold hover:bg-navy/90 cursor-pointer text-[11.5px]"
              >
                Submit Rule for Multi-Party Approval →
              </button>
            </div>
          </div>
        </section>
      )}

      {/* ------------------------------------------------ Main Grid */}
      <div
        className={cn(
          "mt-6 grid gap-6",
          selected ? "xl:grid-cols-[minmax(0,1fr)_380px]" : "grid-cols-1",
        )}
      >
        <section className="min-w-0 space-y-3">
          <SectionHead
            title="REGISTERED POLICIES"
            meta={`${rows.length} of ${rules.length} rules visible`}
          />
          <div className="overflow-x-auto border border-border bg-card">
            <DataGrid
              onSelect={(id) => setSelectedId((cur) => (cur === id ? null : id))}
              selectedId={selectedId}
              rowClassName="py-2.5"
              columns={[
                { key: "id", label: "Rule ID" },
                { key: "name", label: "Name" },
                { key: "policy", label: "Policy Group" },
                { key: "action", label: "Action" },
                { key: "state", label: "State" },
                { key: "hits24h", label: "Hits (24h)" },
                { key: "precision", label: "Precision" },
              ]}
              rows={rows.map((r) => ({
                id: r.id,
                cells: [
                  <Mono key="id" className="font-bold text-navy text-[11.5px]">{r.id}</Mono>,
                  <span key="name" className="text-[12px] font-sans">{r.name}</span>,
                  <Mono key="pol" className="text-[11.5px]">{r.policy}</Mono>,
                  <span key="act" className={actionTone[r.action]}>{r.action}</span>,
                  <StateText key="state" state={r.state} />,
                  <span key="hits" className="font-mono text-[12px] tabular">{r.hits24h.toLocaleString()}</span>,
                  <span key="prec" className="font-mono text-[12px] font-semibold tabular">
                    {(r.precision * 100).toFixed(1)}%
                  </span>,
                ],
              }))}
            />
          </div>
        </section>

        {/* Right Inspector */}
        {selected && detail && (
          <aside className="border border-border bg-card p-5 shadow-xs space-y-4 font-mono text-[11.5px]">
            <div className="flex items-center justify-between border-b border-border pb-2">
              <span className="font-bold text-navy text-[11px] uppercase tracking-wider">
                RULE GOVERNANCE SPECIFICATION
              </span>
              <StateText state={selected.state} />
            </div>

            <div className="space-y-1">
              <div className="font-bold text-foreground text-[13px]">{selected.id}</div>
              <div className="font-sans text-[11.5px] text-muted-foreground">{selected.name}</div>
            </div>

            <div className="border-t border-border pt-3 space-y-2 text-[11px]">
              <div className="flex justify-between">
                <span className="text-muted-foreground">Action:</span>
                <span className={actionTone[selected.action]}>{selected.action}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">24h Hit Volume:</span>
                <span className="font-bold text-foreground">{selected.hits24h.toLocaleString()}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">Historical Precision:</span>
                <span className="font-bold text-authoritative">{(selected.precision * 100).toFixed(1)}%</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">Author:</span>
                <span className="text-foreground">{detail.author}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">Approving Officer:</span>
                <span className="text-foreground font-bold">{detail.approver}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">SHA-256 Digest:</span>
                <Mono className="text-[10px] text-muted-foreground truncate max-w-[160px]">{detail.sha256Hash}</Mono>
              </div>
            </div>

            {/* AST Logic Preview */}
            <div className="border-t border-border pt-3 space-y-1.5">
              <span className="text-[10px] text-muted-foreground uppercase font-bold block">AST Expression</span>
              <pre className="border border-border bg-surface p-2.5 text-[10.5px] text-foreground font-mono overflow-x-auto">
                {JSON.stringify(detail.ast, null, 2)}
              </pre>
            </div>
          </aside>
        )}
      </div>
    </Page>
  );
}
