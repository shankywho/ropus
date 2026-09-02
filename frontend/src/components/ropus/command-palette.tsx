import React, { useEffect, useState } from "react";
import { Command } from "cmdk";
import { useNavigate } from "@tanstack/react-router";
import {
  Search,
  LayoutDashboard,
  PlayCircle,
  Network,
  Cpu,
  FileCode2,
  FolderLock,
  History,
  Activity,
  ShieldCheck,
  Globe2,
  Key,
  Zap,
  Flame,
  Compass,
  CornerDownLeft,
  Terminal,
  Webhook,
  Settings2,
  FileSearch,
} from "lucide-react";
import { cn } from "@/lib/utils";

interface CommandPaletteProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function CommandPalette({ open, onOpenChange }: CommandPaletteProps) {
  const navigate = useNavigate();
  const [search, setSearch] = useState("");

  useEffect(() => {
    const down = (e: KeyboardEvent) => {
      if ((e.key === "k" && (e.metaKey || e.ctrlKey)) || e.key === "/") {
        if (
          e.target instanceof HTMLInputElement ||
          e.target instanceof HTMLTextAreaElement ||
          (e.target as HTMLElement).isContentEditable
        ) {
          return;
        }
        e.preventDefault();
        onOpenChange(!open);
      }
    };

    document.addEventListener("keydown", down);
    return () => document.removeEventListener("keydown", down);
  }, [open, onOpenChange]);

  const handleSelect = (callback: () => void) => {
    onOpenChange(false);
    callback();
  };

  if (!open) return null;

  return (
    <div
      className="fixed inset-0 z-50 flex items-start justify-center bg-black/60 pt-[12vh] backdrop-blur-xs p-4 animate-in fade-in duration-150"
      onClick={() => onOpenChange(false)}
    >
      <div
        className="relative w-full max-w-2xl border border-border bg-card shadow-2xl overflow-hidden animate-in zoom-in-95 duration-150"
        onClick={(e) => e.stopPropagation()}
      >
        <Command className="flex flex-col bg-transparent font-sans" loop shouldFilter={true}>
          {/* Header Input */}
          <div className="flex items-center border-b border-border px-4 py-3 bg-secondary/30">
            <Search className="mr-3 size-4 text-muted-foreground shrink-0" />
            <Command.Input
              value={search}
              onValueChange={setSearch}
              placeholder="Search surfaces, entities (txn_, cus_, CASE-), or run actions..."
              className="w-full bg-transparent font-mono text-[12.5px] text-foreground placeholder:text-muted-foreground outline-hidden"
              autoFocus
            />
            <kbd className="hidden sm:inline-flex items-center gap-1 border border-border bg-surface px-1.5 py-0.5 font-mono text-[9.5px] text-muted-foreground">
              ESC
            </kbd>
          </div>

          {/* List */}
          <Command.List className="max-h-[400px] overflow-y-auto p-2 font-mono text-[12px] scrollbar-thin">
            <Command.Empty className="py-8 text-center text-muted-foreground text-[12px] font-sans">
              No matching control plane surfaces, entities, or commands found.
            </Command.Empty>

            {/* QUICK NAVIGATION */}
            <Command.Group
              heading="CONTROL PLANE SURFACES"
              className="px-2 py-1.5 font-mono text-[9.5px] font-bold tracking-[0.12em] text-muted-foreground uppercase"
            >
              <Command.Item
                value="overview dashboard home control deck kpis"
                onSelect={() => handleSelect(() => navigate({ to: "/" }))}
                className="flex items-center justify-between px-3 py-2 text-foreground cursor-pointer rounded-none aria-selected:bg-navy aria-selected:text-white transition-colors"
              >
                <div className="flex items-center gap-2.5">
                  <LayoutDashboard className="size-4 shrink-0 text-navy aria-selected:text-white" />
                  <span className="font-sans text-[12.5px] font-medium">
                    Overview & Control Deck
                  </span>
                </div>
                <kbd className="font-mono text-[9.5px] opacity-70">G H</kbd>
              </Command.Item>

              <Command.Item
                value="demo walkthrough 7-stage incident replay attack simulation"
                onSelect={() => handleSelect(() => navigate({ to: "/demo" }))}
                className="flex items-center justify-between px-3 py-2 text-foreground cursor-pointer rounded-none aria-selected:bg-navy aria-selected:text-white transition-colors"
              >
                <div className="flex items-center gap-2.5">
                  <PlayCircle className="size-4 shrink-0 text-shadow-intel aria-selected:text-white" />
                  <span className="font-sans text-[12.5px] font-medium">
                    Interactive 7-Stage Demo Replay
                  </span>
                </div>
                <span className="border border-shadow-intel/40 bg-shadow-intel-surface px-1.5 py-0.2 font-mono text-[9px] font-semibold text-shadow-intel uppercase tracking-[0.06em]">
                  REPLAY
                </span>
              </Command.Item>

              <Command.Item
                value="fraud graph syndicate explorer knowledge graph graphsage gnn"
                onSelect={() => handleSelect(() => navigate({ to: "/graph" }))}
                className="flex items-center justify-between px-3 py-2 text-foreground cursor-pointer rounded-none aria-selected:bg-navy aria-selected:text-white transition-colors"
              >
                <div className="flex items-center gap-2.5">
                  <Network className="size-4 shrink-0 text-amber-intel aria-selected:text-white" />
                  <span className="font-sans text-[12.5px] font-medium">
                    Fraud Knowledge Graph (GraphSAGE Shadow)
                  </span>
                </div>
                <kbd className="font-mono text-[9.5px] opacity-70">G G</kbd>
              </Command.Item>

              <Command.Item
                value="decisions synchronous audit logs evaluations telemetry"
                onSelect={() => handleSelect(() => navigate({ to: "/decisions" }))}
                className="flex items-center justify-between px-3 py-2 text-foreground cursor-pointer rounded-none aria-selected:bg-navy aria-selected:text-white transition-colors"
              >
                <div className="flex items-center gap-2.5">
                  <History className="size-4 shrink-0 text-navy aria-selected:text-white" />
                  <span className="font-sans text-[12.5px] font-medium">
                    Decision Evaluation Logs & Telemetry
                  </span>
                </div>
                <kbd className="font-mono text-[9.5px] opacity-70">G D</kbd>
              </Command.Item>

              <Command.Item
                value="cases analyst review queue triage investigation 24h sla"
                onSelect={() => handleSelect(() => navigate({ to: "/cases" }))}
                className="flex items-center justify-between px-3 py-2 text-foreground cursor-pointer rounded-none aria-selected:bg-navy aria-selected:text-white transition-colors"
              >
                <div className="flex items-center gap-2.5">
                  <FolderLock className="size-4 shrink-0 text-blocked aria-selected:text-white" />
                  <span className="font-sans text-[12.5px] font-medium">
                    Case Management Queue (24h SLA)
                  </span>
                </div>
                <kbd className="font-mono text-[9.5px] opacity-70">G C</kbd>
              </Command.Item>

              <Command.Item
                value="investigations dossiers multi-case clusters mule rings"
                onSelect={() => handleSelect(() => navigate({ to: "/investigations" }))}
                className="flex items-center justify-between px-3 py-2 text-foreground cursor-pointer rounded-none aria-selected:bg-navy aria-selected:text-white transition-colors"
              >
                <div className="flex items-center gap-2.5">
                  <FileSearch className="size-4 shrink-0 text-navy aria-selected:text-white" />
                  <span className="font-sans text-[12.5px] font-medium">
                    Evidentiary Dossiers & Investigations
                  </span>
                </div>
              </Command.Item>

              <Command.Item
                value="rules ast declarative maker checker policy builder"
                onSelect={() => handleSelect(() => navigate({ to: "/rules" }))}
                className="flex items-center justify-between px-3 py-2 text-foreground cursor-pointer rounded-none aria-selected:bg-navy aria-selected:text-white transition-colors"
              >
                <div className="flex items-center gap-2.5">
                  <FileCode2 className="size-4 shrink-0 text-navy aria-selected:text-white" />
                  <span className="font-sans text-[12.5px] font-medium">
                    Rules Engine & AST Policy Builder
                  </span>
                </div>
                <kbd className="font-mono text-[9.5px] opacity-70">G R</kbd>
              </Command.Item>

              <Command.Item
                value="models registry onnx bmr shadow xgboost catboost loss matrix"
                onSelect={() => handleSelect(() => navigate({ to: "/models" }))}
                className="flex items-center justify-between px-3 py-2 text-foreground cursor-pointer rounded-none aria-selected:bg-navy aria-selected:text-white transition-colors"
              >
                <div className="flex items-center gap-2.5">
                  <Cpu className="size-4 shrink-0 text-authoritative aria-selected:text-white" />
                  <span className="font-sans text-[12.5px] font-medium">
                    Model Registry & BMR Cost Optimizer
                  </span>
                </div>
                <kbd className="font-mono text-[9.5px] opacity-70">G M</kbd>
              </Command.Item>

              <Command.Item
                value="threat intelligence impossible travel asn tor proxy ip reputation"
                onSelect={() => handleSelect(() => navigate({ to: "/threat-intelligence" }))}
                className="flex items-center justify-between px-3 py-2 text-foreground cursor-pointer rounded-none aria-selected:bg-navy aria-selected:text-white transition-colors"
              >
                <div className="flex items-center gap-2.5">
                  <Globe2 className="size-4 shrink-0 text-amber-intel aria-selected:text-white" />
                  <span className="font-sans text-[12.5px] font-medium">
                    Threat Intelligence & Indicators
                  </span>
                </div>
                <kbd className="font-mono text-[9.5px] opacity-70">G T</kbd>
              </Command.Item>

              <Command.Item
                value="operations slo sre telemetry health metrics p99 latency"
                onSelect={() => handleSelect(() => navigate({ to: "/operations" }))}
                className="flex items-center justify-between px-3 py-2 text-foreground cursor-pointer rounded-none aria-selected:bg-navy aria-selected:text-white transition-colors"
              >
                <div className="flex items-center gap-2.5">
                  <Activity className="size-4 shrink-0 text-authoritative aria-selected:text-white" />
                  <span className="font-sans text-[12.5px] font-medium">
                    SRE Operations & SLO Monitor
                  </span>
                </div>
                <kbd className="font-mono text-[9.5px] opacity-70">G O</kbd>
              </Command.Item>

              <Command.Item
                value="security kms sha256 hash chain pci soc2 audit ledger"
                onSelect={() => handleSelect(() => navigate({ to: "/security" }))}
                className="flex items-center justify-between px-3 py-2 text-foreground cursor-pointer rounded-none aria-selected:bg-navy aria-selected:text-white transition-colors"
              >
                <div className="flex items-center gap-2.5">
                  <ShieldCheck className="size-4 shrink-0 text-authoritative aria-selected:text-white" />
                  <span className="font-sans text-[12.5px] font-medium">
                    Security & SHA-256 Audit Hash Chain
                  </span>
                </div>
                <kbd className="font-mono text-[9.5px] opacity-70">G S</kbd>
              </Command.Item>

              <Command.Item
                value="api reference sdk curl python go synchronous evaluate"
                onSelect={() => handleSelect(() => navigate({ to: "/api" }))}
                className="flex items-center justify-between px-3 py-2 text-foreground cursor-pointer rounded-none aria-selected:bg-navy aria-selected:text-white transition-colors"
              >
                <div className="flex items-center gap-2.5">
                  <Terminal className="size-4 shrink-0 text-navy aria-selected:text-white" />
                  <span className="font-sans text-[12.5px] font-medium">
                    Developer API Reference
                  </span>
                </div>
                <kbd className="font-mono text-[9.5px] opacity-70">G A</kbd>
              </Command.Item>

              <Command.Item
                value="api keys developer access authentication tokens rotation"
                onSelect={() => handleSelect(() => navigate({ to: "/api-keys" }))}
                className="flex items-center justify-between px-3 py-2 text-foreground cursor-pointer rounded-none aria-selected:bg-navy aria-selected:text-white transition-colors"
              >
                <div className="flex items-center gap-2.5">
                  <Key className="size-4 shrink-0 text-muted-foreground aria-selected:text-white" />
                  <span className="font-sans text-[12.5px] font-medium">
                    API Keys & Authentication
                  </span>
                </div>
              </Command.Item>

              <Command.Item
                value="webhooks endpoints retries event streaming cdc"
                onSelect={() => handleSelect(() => navigate({ to: "/webhooks" }))}
                className="flex items-center justify-between px-3 py-2 text-foreground cursor-pointer rounded-none aria-selected:bg-navy aria-selected:text-white transition-colors"
              >
                <div className="flex items-center gap-2.5">
                  <Webhook className="size-4 shrink-0 text-muted-foreground aria-selected:text-white" />
                  <span className="font-sans text-[12.5px] font-medium">
                    Webhooks & Streaming CDC
                  </span>
                </div>
              </Command.Item>

              <Command.Item
                value="settings tenant configuration thresholds timeouts rbac"
                onSelect={() => handleSelect(() => navigate({ to: "/settings" }))}
                className="flex items-center justify-between px-3 py-2 text-foreground cursor-pointer rounded-none aria-selected:bg-navy aria-selected:text-white transition-colors"
              >
                <div className="flex items-center gap-2.5">
                  <Settings2 className="size-4 shrink-0 text-muted-foreground aria-selected:text-white" />
                  <span className="font-sans text-[12.5px] font-medium">
                    Tenant Settings & Verdict Thresholds
                  </span>
                </div>
              </Command.Item>
            </Command.Group>

            {/* ENTITY SEARCH */}
            <Command.Group
              heading="SAMPLE ENTITY PREVIEWS"
              className="mt-2 px-2 py-1.5 font-mono text-[9.5px] font-bold tracking-[0.12em] text-muted-foreground uppercase border-t border-border pt-2"
            >
              <Command.Item
                value="txn_order_88419 attack transaction limassol wire inr dec_88419_attack"
                onSelect={() =>
                  handleSelect(() =>
                    navigate({
                      to: "/decisions/$decisionId",
                      params: { decisionId: "dec_88419_attack" },
                    }),
                  )
                }
                className="flex items-center justify-between px-3 py-2 text-foreground cursor-pointer rounded-none aria-selected:bg-navy aria-selected:text-white transition-colors"
              >
                <div className="flex items-center gap-2.5">
                  <Zap className="size-4 shrink-0 text-blocked" />
                  <div>
                    <div className="font-mono font-bold text-[12px]">
                      txn_order_88419 (₹14,50,000.00 IMPS Payout)
                    </div>
                    <div className="font-sans text-[11px] text-muted-foreground aria-selected:text-white/80">
                      Limassol Datacenter Proxy · Score: 96/100 · BLOCK
                    </div>
                  </div>
                </div>
                <CornerDownLeft className="size-3.5 opacity-60" />
              </Command.Item>

              <Command.Item
                value="txn_baseline_99120 organic quick commerce spend dec_99120_normal"
                onSelect={() =>
                  handleSelect(() =>
                    navigate({
                      to: "/decisions/$decisionId",
                      params: { decisionId: "dec_99120_normal" },
                    }),
                  )
                }
                className="flex items-center justify-between px-3 py-2 text-foreground cursor-pointer rounded-none aria-selected:bg-navy aria-selected:text-white transition-colors"
              >
                <div className="flex items-center gap-2.5">
                  <Zap className="size-4 shrink-0 text-authoritative" />
                  <div>
                    <div className="font-mono font-bold text-[12px]">
                      txn_baseline_99120 (₹385.00 INR)
                    </div>
                    <div className="font-sans text-[11px] text-muted-foreground aria-selected:text-white/80">
                      Bengaluru Residential · Score: 02/100 · APPROVE
                    </div>
                  </div>
                </div>
                <CornerDownLeft className="size-3.5 opacity-60" />
              </Command.Item>

              <Command.Item
                value="CASE-88419 p0 critical syndicate investigation"
                onSelect={() =>
                  handleSelect(() =>
                    navigate({ to: "/cases/$caseId", params: { caseId: "CASE-88419" } }),
                  )
                }
                className="flex items-center justify-between px-3 py-2 text-foreground cursor-pointer rounded-none aria-selected:bg-navy aria-selected:text-white transition-colors"
              >
                <div className="flex items-center gap-2.5">
                  <FolderLock className="size-4 shrink-0 text-blocked" />
                  <div>
                    <div className="font-mono font-bold text-[12px]">
                      CASE-88419 (P0 Critical: Multi-Mule Syndicate)
                    </div>
                    <div className="font-sans text-[11px] text-muted-foreground aria-selected:text-white/80">
                      Assigned: a.sharma · SLA: 14m remaining
                    </div>
                  </div>
                </div>
                <CornerDownLeft className="size-3.5 opacity-60" />
              </Command.Item>
            </Command.Group>

            {/* OPERATOR ACTIONS */}
            <Command.Group
              heading="OPERATOR ACTIONS & SIMULATIONS"
              className="mt-2 px-2 py-1.5 font-mono text-[9.5px] font-bold tracking-[0.12em] text-muted-foreground uppercase border-t border-border pt-2"
            >
              <Command.Item
                value="trigger live risk evaluation evaluate payload"
                onSelect={() => handleSelect(() => navigate({ to: "/demo" }))}
                className="flex items-center justify-between px-3 py-2 text-foreground cursor-pointer rounded-none aria-selected:bg-navy aria-selected:text-white transition-colors"
              >
                <div className="flex items-center gap-2.5">
                  <Flame className="size-4 shrink-0 text-amber-intel" />
                  <span className="font-sans text-[12.5px]">
                    Trigger Live Go Backend Risk Evaluation
                  </span>
                </div>
              </Command.Item>

              <Command.Item
                value="simulate impossible travel geo velocity jump"
                onSelect={() => handleSelect(() => navigate({ to: "/threat-intelligence" }))}
                className="flex items-center justify-between px-3 py-2 text-foreground cursor-pointer rounded-none aria-selected:bg-navy aria-selected:text-white transition-colors"
              >
                <div className="flex items-center gap-2.5">
                  <Compass className="size-4 shrink-0 text-shadow-intel" />
                  <span className="font-sans text-[12.5px]">
                    Simulate Spherical Haversine Travel Anomaly
                  </span>
                </div>
              </Command.Item>

              <Command.Item
                value="verify cryptographic sha256 hash chain audit ledger"
                onSelect={() => handleSelect(() => navigate({ to: "/security" }))}
                className="flex items-center justify-between px-3 py-2 text-foreground cursor-pointer rounded-none aria-selected:bg-navy aria-selected:text-white transition-colors"
              >
                <div className="flex items-center gap-2.5">
                  <ShieldCheck className="size-4 shrink-0 text-authoritative" />
                  <span className="font-sans text-[12.5px]">
                    Verify SHA-256 Audit Hash Chain Ledger
                  </span>
                </div>
              </Command.Item>
            </Command.Group>
          </Command.List>

          {/* Footer Bar */}
          <div className="flex items-center justify-between border-t border-border bg-secondary/50 px-4 py-2 text-[10px] font-mono text-muted-foreground">
            <div className="flex items-center gap-3">
              <span>
                <kbd className="border border-border bg-surface px-1">↑</kbd>{" "}
                <kbd className="border border-border bg-surface px-1">↓</kbd> Navigate
              </span>
              <span>
                <kbd className="border border-border bg-surface px-1">↵</kbd> Select
              </span>
            </div>
            <div className="flex items-center gap-2">
              <span className="size-1.5 rounded-full bg-authoritative" />
              <span>ROPUS Engine Online (&lt;100ms p95 SLA)</span>
            </div>
          </div>
        </Command>
      </div>
    </div>
  );
}
