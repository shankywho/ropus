import { createFileRoute, Link } from "@tanstack/react-router";
import { useMemo, useState } from "react";
import { CaseStatusTag, DemoTag, PriorityTag, RiskScore } from "@/components/ropus/core";
import { useSuspenseQuery } from "@tanstack/react-query";
import { casesQuery } from "@/lib/ropus/api";
import type { CaseStatus } from "@/lib/ropus/contracts";
import { cn } from "@/lib/utils";
import { Clock, AlertTriangle, CheckSquare, Square, ArrowRight } from "lucide-react";
import { toast } from "sonner";

export const Route = createFileRoute("/cases/")({
  head: () => ({
    meta: [
      { title: "Case Management Queue & Tripartite Dossier — ROPUS" },
      {
        name: "description",
        content:
          "High-density analyst investigation queue with 24-hour SLA countdown, batch triage actions, and forensic evidence dossiers.",
      },
      { property: "og:title", content: "Case Management Queue — ROPUS" },
      {
        property: "og:description",
        content:
          "Investigation queue for blocked, challenged and review-flagged transactions with 24h SLA urgency.",
      },
      { property: "og:type", content: "website" },
      { name: "twitter:card", content: "summary" },
    ],
  }),
  loader: ({ context }) => context.queryClient.ensureQueryData(casesQuery()),
  component: CaseQueue,
});

const filters: Array<{ id: "ALL" | CaseStatus; label: string }> = [
  { id: "ALL", label: "All Cases" },
  { id: "OPEN", label: "Open (Unassigned)" },
  { id: "IN_REVIEW", label: "In Review" },
  { id: "ESCALATED", label: "Escalated (L2)" },
  { id: "CLOSED", label: "Closed / Resolved" },
];

const money = (amount: number, currency: string) =>
  new Intl.NumberFormat("en-IN", { style: "currency", currency: currency || "INR" }).format(amount);

function SlaUrgencyBadge({ minutes }: { minutes: number }) {
  if (minutes < 0) {
    return (
      <span className="inline-flex items-center gap-1 border border-blocked bg-blocked-surface px-2 py-0.5 font-mono text-[9.5px] font-bold text-blocked uppercase animate-pulse">
        <AlertTriangle className="size-3" />
        BREACHED ({Math.abs(minutes)}m ago)
      </span>
    );
  }
  if (minutes <= 240) {
    return (
      <span className="inline-flex items-center gap-1 border border-blocked/60 bg-blocked-surface px-2 py-0.5 font-mono text-[9.5px] font-bold text-blocked uppercase">
        <Clock className="size-3" />
        CRITICAL ({Math.floor(minutes / 60)}h {minutes % 60}m)
      </span>
    );
  }
  if (minutes <= 720) {
    return (
      <span className="inline-flex items-center gap-1 border border-amber-intel/60 bg-shadow-intel-surface px-2 py-0.5 font-mono text-[9.5px] font-bold text-shadow-intel uppercase">
        <Clock className="size-3" />
        {Math.floor(minutes / 60)}h {minutes % 60}m REMAINING
      </span>
    );
  }
  return (
    <span className="inline-flex items-center gap-1 border border-authoritative/40 bg-approve-surface px-2 py-0.5 font-mono text-[9.5px] font-semibold text-authoritative uppercase">
      <Clock className="size-3" />
      {Math.floor(minutes / 60)}h {minutes % 60}m
    </span>
  );
}

export function CaseQueue() {
  const { data: allCases } = useSuspenseQuery(casesQuery());
  const [status, setStatus] = useState<"ALL" | CaseStatus>("ALL");
  const [selectedCaseIds, setSelectedCaseIds] = useState<Set<string>>(new Set());

  const rows = useMemo(
    () => (status === "ALL" ? allCases : allCases.filter((c) => c.status === status)),
    [allCases, status],
  );

  const open = allCases.filter((c) => c.status !== "CLOSED");
  const breached = open.filter((c) => c.slaMinutesRemaining < 0);
  const critical = open.filter((c) => c.slaMinutesRemaining > 0 && c.slaMinutesRemaining <= 240);

  const toggleSelect = (id: string) => {
    setSelectedCaseIds((prev) => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });
  };

  const selectAll = () => {
    if (selectedCaseIds.size === rows.length) {
      setSelectedCaseIds(new Set());
    } else {
      setSelectedCaseIds(new Set(rows.map((r) => r.caseId)));
    }
  };

  const handleBatchClaim = () => {
    toast.success(`Claimed ${selectedCaseIds.size} cases for review`, {
      description: "Cases assigned to analyst m.okafor with 24-hour SLA timer active.",
    });
    setSelectedCaseIds(new Set());
  };

  const handleBatchConfirmBlock = () => {
    toast.success(`Batch Block confirmed for ${selectedCaseIds.size} cases`, {
      description: "Cryptographic hash-chain ledger entries committed and accounts frozen.",
    });
    setSelectedCaseIds(new Set());
  };

  return (
    <div className="mx-auto w-full max-w-[1560px] px-5 py-5 lg:px-8 space-y-5">
      {/* Header */}
      <header className="flex flex-wrap items-baseline justify-between gap-4 border-b border-border pb-4">
        <div>
          <div className="flex items-center gap-2">
            <span className="border border-navy bg-navy/10 px-2 py-0.5 font-mono text-[9.5px] font-bold text-navy uppercase tracking-[0.06em]">
              ANALYST OPERATIONS
            </span>
            <DemoTag />
          </div>
          <h1 className="mt-1 font-sans text-[24px] lg:text-[26px] font-extrabold leading-[1.12] tracking-[-0.04em] text-foreground">
            Case Management Queue &amp; SLA Monitor
          </h1>
          <p className="mt-0.5 font-sans text-[12.5px] text-muted-foreground">
            Deterministic 24-hour investigation SLA. All analyst dispositions are committed to an
            immutable append-only ledger.
          </p>
        </div>
      </header>

      {/* Telemetry KPI Strip */}
      <dl className="grid grid-cols-2 gap-3 sm:grid-cols-4 lg:grid-cols-5 border border-border bg-card p-4 shadow-xs font-mono text-[11px]">
        <div>
          <dd className="text-[22px] font-extrabold text-foreground tabular">{open.length}</dd>
          <dt className="text-muted-foreground uppercase text-[9.5px] font-semibold tracking-wider mt-0.5">
            Active Open Cases
          </dt>
        </div>
        <div>
          <dd className="text-[22px] font-extrabold text-blocked tabular">{breached.length}</dd>
          <dt className="text-muted-foreground uppercase text-[9.5px] font-semibold tracking-wider mt-0.5">
            SLA Breached (&lt;0m)
          </dt>
        </div>
        <div>
          <dd className="text-[22px] font-extrabold text-amber-intel tabular">{critical.length}</dd>
          <dt className="text-muted-foreground uppercase text-[9.5px] font-semibold tracking-wider mt-0.5">
            Critical Urgency (&lt;4h)
          </dt>
        </div>
        <div>
          <dd className="text-[22px] font-extrabold text-navy tabular">
            {open.filter((c) => c.priority === "P1").length}
          </dd>
          <dt className="text-muted-foreground uppercase text-[9.5px] font-semibold tracking-wider mt-0.5">
            P1 Priority
          </dt>
        </div>
        <div>
          <dd className="text-[22px] font-extrabold text-authoritative tabular">
            {open.filter((c) => !c.assignee).length}
          </dd>
          <dt className="text-muted-foreground uppercase text-[9.5px] font-semibold tracking-wider mt-0.5">
            Unassigned
          </dt>
        </div>
      </dl>

      {/* Filter Tabs & Batch Actions Bar */}
      <div className="flex flex-wrap items-center justify-between gap-3 border-b border-border pb-3">
        <nav aria-label="Filter by status" className="flex flex-wrap gap-1 font-mono text-[10.5px]">
          {filters.map((f) => (
            <button
              key={f.id}
              type="button"
              onClick={() => setStatus(f.id)}
              className={cn(
                "px-3 py-1 transition-colors cursor-pointer border font-semibold",
                status === f.id
                  ? "border-navy bg-navy text-white font-bold"
                  : "border-transparent text-muted-foreground hover:text-foreground",
              )}
            >
              {f.label}
            </button>
          ))}
        </nav>

        {/* Batch Action Toolbar */}
        {selectedCaseIds.size > 0 && (
          <div className="flex items-center gap-2 font-mono text-[10.5px] bg-secondary/70 border border-border px-3 py-1 shadow-2xs">
            <span className="font-bold text-foreground">{selectedCaseIds.size} Selected</span>
            <button
              type="button"
              onClick={handleBatchClaim}
              className="border border-navy bg-navy px-2.5 py-0.5 text-white font-bold hover:bg-navy/90 cursor-pointer"
            >
              Claim Selected
            </button>
            <button
              type="button"
              onClick={handleBatchConfirmBlock}
              className="border border-blocked bg-blocked px-2.5 py-0.5 text-white font-bold hover:bg-blocked/90 cursor-pointer"
            >
              Batch Block &amp; Freeze
            </button>
          </div>
        )}
      </div>

      {/* Case DataGrid Table */}
      <div className="border border-border bg-card shadow-xs overflow-x-auto">
        <table className="w-full text-left font-sans text-[12px] border-collapse">
          <thead>
            <tr className="border-b border-border bg-secondary/40 font-mono text-[9px] uppercase font-bold text-muted-foreground tracking-wider">
              <th className="py-2.5 px-3 w-[40px]">
                <button type="button" onClick={selectAll} className="cursor-pointer">
                  {selectedCaseIds.size === rows.length && rows.length > 0 ? (
                    <CheckSquare className="size-3.5 text-navy" />
                  ) : (
                    <Square className="size-3.5 text-muted-foreground" />
                  )}
                </button>
              </th>
              <th className="py-2.5 px-3">Case ID</th>
              <th className="py-2.5 px-3">Priority</th>
              <th className="py-2.5 px-3">Status</th>
              <th className="py-2.5 px-3">SLA Countdown</th>
              <th className="py-2.5 px-3">Amount</th>
              <th className="py-2.5 px-3">Risk Score</th>
              <th className="py-2.5 px-3">Assignee</th>
              <th className="py-2.5 px-3">Action</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-border font-mono text-[11.5px]">
            {rows.map((c) => {
              const isChecked = selectedCaseIds.has(c.caseId);
              return (
                <tr
                  key={c.caseId}
                  className={cn(
                    "hover:bg-secondary/40 transition-colors",
                    isChecked && "bg-secondary/60",
                  )}
                >
                  <td className="py-2.5 px-3">
                    <button
                      type="button"
                      onClick={() => toggleSelect(c.caseId)}
                      className="cursor-pointer"
                    >
                      {isChecked ? (
                        <CheckSquare className="size-3.5 text-navy" />
                      ) : (
                        <Square className="size-3.5 text-muted-foreground" />
                      )}
                    </button>
                  </td>
                  <td className="py-2.5 px-3 font-bold text-navy">
                    <Link
                      to={`/cases/${c.caseId}`}
                      className="hover:underline flex items-center gap-1"
                    >
                      <span>{c.caseId}</span>
                      <ArrowRight className="size-3 opacity-60" />
                    </Link>
                  </td>
                  <td className="py-2.5 px-3">
                    <PriorityTag priority={c.priority} />
                  </td>
                  <td className="py-2.5 px-3">
                    <CaseStatusTag status={c.status} />
                  </td>
                  <td className="py-2.5 px-3">
                    <SlaUrgencyBadge minutes={c.slaMinutesRemaining} />
                  </td>
                  <td className="py-2.5 px-3 font-bold text-foreground tabular">
                    {money(c.amount, c.currency)}
                  </td>
                  <td className="py-2.5 px-3">
                    <RiskScore value={c.riskScore} size="sm" showBand />
                  </td>
                  <td className="py-2.5 px-3 text-muted-foreground font-sans text-[11.5px]">
                    {c.assignee ?? (
                      <span className="text-amber-intel italic font-semibold font-mono text-[11px]">
                        Unassigned
                      </span>
                    )}
                  </td>
                  <td className="py-2.5 px-3">
                    <Link
                      to={`/cases/${c.caseId}`}
                      className="border border-border bg-surface px-2 py-1 text-foreground font-medium hover:bg-secondary transition-colors font-sans text-[11.5px]"
                    >
                      Dossier →
                    </Link>
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>
    </div>
  );
}
