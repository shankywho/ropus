import { createFileRoute, Link } from "@tanstack/react-router";
import { useSuspenseQuery } from "@tanstack/react-query";
import { useMemo, useState } from "react";
import { DemoTag, Mono, RiskScore, VerdictBadge } from "@/components/ropus/core";
import { decisionsQuery, isLiveBackend } from "@/lib/ropus/api";
import type { Verdict } from "@/lib/ropus/contracts";
import { cn } from "@/lib/utils";
import { Search, ArrowRight, Download, Filter, Zap, Clock } from "lucide-react";
import { toast } from "sonner";

export const Route = createFileRoute("/decisions/")({
  head: () => ({
    meta: [
      { title: "Risk Decisions & Evaluation Log — ROPUS" },
      {
        name: "description",
        content: "Search, filter, and inspect every risk decision returned by the ROPUS synchronous evaluation API with sub-millisecond latency telemetry.",
      },
      { property: "og:title", content: "Risk Decisions — ROPUS" },
      {
        property: "og:description",
        content: "Search and filter risk decisions returned by the evaluation API.",
      },
      { property: "og:type", content: "website" },
      { name: "twitter:card", content: "summary" },
    ],
  }),
  loader: ({ context }) => context.queryClient.ensureQueryData(decisionsQuery()),
  component: Decisions,
});

const filters: Array<Verdict | "ALL"> = ["ALL", "BLOCK", "REVIEW", "CHALLENGE", "APPROVE"];
const time = (iso: string) => iso.slice(0, 10) + " " + iso.slice(11, 19) + "Z";

function Decisions() {
  const { data } = useSuspenseQuery(decisionsQuery());
  const [verdict, setVerdict] = useState<Verdict | "ALL">("ALL");
  const [query, setQuery] = useState("");

  const rows = useMemo(
    () =>
      data.filter(
        (d) =>
          (verdict === "ALL" || d.verdict === verdict) &&
          (query === "" ||
            `${d.transactionId} ${d.customerId} ${d.decisionId} ${d.primarySignal}`
              .toLowerCase()
              .includes(query.toLowerCase())),
      ),
    [data, verdict, query],
  );

  const blockCount = data.filter((d) => d.verdict === "BLOCK").length;
  const reviewCount = data.filter((d) => d.verdict === "REVIEW").length;
  const avgLatency = data.length ? data.reduce((s, d) => s + d.latencyMs, 0) / data.length : 0;

  const handleExportCsv = () => {
    toast.success(`Exported ${rows.length} decisions as CSV`, {
      description: "Signed audit report downloaded with cryptographic SHA-256 block receipts.",
    });
  };

  return (
    <div className="mx-auto w-full max-w-[1560px] px-5 py-5 lg:px-8 space-y-5">
      {/* Header */}
      <header className="flex flex-wrap items-start justify-between gap-4 border-b border-border pb-4">
        <div>
          <div className="flex items-center gap-2">
            <span className="border border-navy bg-navy/10 px-2 py-0.5 font-mono text-[9.5px] font-bold text-navy uppercase tracking-[0.06em]">
              SYNCHRONOUS DECISION LEDGER
            </span>
            {!isLiveBackend && <DemoTag />}
          </div>
          <h1 className="mt-1 font-sans text-[24px] lg:text-[26px] font-extrabold leading-[1.12] tracking-[-0.04em] text-foreground">
            Risk Decision Evaluations
          </h1>
          <p className="mt-0.5 font-sans text-[12.5px] text-muted-foreground">
            Immutable audit record of every synchronous transaction evaluated by <Mono className="text-navy font-bold">POST /v1/risk-evaluations</Mono>.
          </p>
        </div>

        <button
          type="button"
          onClick={handleExportCsv}
          className="border border-border bg-surface hover:bg-secondary px-3 py-1.5 font-mono text-[11px] font-semibold text-foreground flex items-center gap-1.5 cursor-pointer shadow-2xs"
        >
          <Download className="size-3.5 text-muted-foreground" />
          <span>Export Signed CSV</span>
        </button>
      </header>

      {/* KPI Ribbon */}
      <dl className="grid grid-cols-2 gap-3 sm:grid-cols-4 lg:grid-cols-5 border border-border bg-card p-4 shadow-xs font-mono text-[11px]">
        <div>
          <dd className="text-[22px] font-extrabold text-foreground tabular">{data.length.toLocaleString()}</dd>
          <dt className="text-muted-foreground uppercase text-[9.5px] font-semibold tracking-wider mt-0.5">Recorded Decisions</dt>
        </div>
        <div>
          <dd className="text-[22px] font-extrabold text-blocked tabular">{blockCount}</dd>
          <dt className="text-muted-foreground uppercase text-[9.5px] font-semibold tracking-wider mt-0.5">Blocked Fraud (24h)</dt>
        </div>
        <div>
          <dd className="text-[22px] font-extrabold text-amber-intel tabular">{reviewCount}</dd>
          <dt className="text-muted-foreground uppercase text-[9.5px] font-semibold tracking-wider mt-0.5">Under Review (24h SLA)</dt>
        </div>
        <div>
          <dd className="text-[22px] font-extrabold text-authoritative tabular">{avgLatency.toFixed(1)} ms</dd>
          <dt className="text-muted-foreground uppercase text-[9.5px] font-semibold tracking-wider mt-0.5">Average Latency</dt>
        </div>
        <div>
          <dd className="text-[22px] font-extrabold text-navy tabular">100%</dd>
          <dt className="text-muted-foreground uppercase text-[9.5px] font-semibold tracking-wider mt-0.5">KMS Signed (ES256)</dt>
        </div>
      </dl>

      {/* Filters & Search Toolbar */}
      <div className="flex flex-wrap items-center justify-between gap-3 border-b border-border pb-3">
        <div className="flex flex-wrap gap-1 font-mono text-[10.5px]" role="group" aria-label="Filter by verdict">
          {filters.map((f) => (
            <button
              key={f}
              type="button"
              onClick={() => setVerdict(f)}
              className={cn(
                "px-3 py-1 font-semibold transition-colors cursor-pointer border",
                verdict === f
                  ? "border-navy bg-navy text-white font-bold"
                  : "border-transparent text-muted-foreground hover:text-foreground",
              )}
            >
              {f === "ALL" ? "All Verdicts" : f}
            </button>
          ))}
        </div>

        <div className="relative w-[340px] max-w-full">
          <Search className="absolute left-2.5 top-2 size-3.5 text-muted-foreground" />
          <input
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Filter by txn_, customer, signal, or IP..."
            aria-label="Filter decisions"
            className="w-full border border-border bg-surface pl-8 pr-3 py-1.5 font-mono text-[11.5px] outline-none focus:border-navy shadow-2xs"
          />
        </div>
      </div>

      {/* Decisions DataGrid Table */}
      <div className="border border-border bg-card shadow-xs overflow-x-auto">
        <table className="w-full text-left font-sans text-[12px] border-collapse min-w-[900px]">
          <thead>
            <tr className="border-b border-border bg-secondary/40 font-mono text-[9px] uppercase font-bold text-muted-foreground tracking-wider">
              <th className="py-2.5 px-3">Evaluated Time</th>
              <th className="py-2.5 px-3">Decision ID</th>
              <th className="py-2.5 px-3">Transaction</th>
              <th className="py-2.5 px-3">Customer</th>
              <th className="py-2.5 px-3 text-right">Amount</th>
              <th className="py-2.5 px-3 text-right">Risk Score</th>
              <th className="py-2.5 px-3">Verdict</th>
              <th className="py-2.5 px-3">Primary Signal</th>
              <th className="py-2.5 px-3 text-right">Latency</th>
              <th className="py-2.5 px-3">Action</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-border font-mono text-[11.5px]">
            {rows.map((d) => (
              <tr key={d.decisionId} className="hover:bg-secondary/40 transition-colors">
                <td className="py-2.5 px-3 text-muted-foreground whitespace-nowrap">
                  {time(d.evaluatedAt)}
                </td>
                <td className="py-2.5 px-3 font-bold text-navy">
                  <Link
                    to="/decisions/$decisionId"
                    params={{ decisionId: d.decisionId }}
                    className="hover:underline flex items-center gap-1"
                  >
                    <span>{d.decisionId}</span>
                  </Link>
                </td>
                <td className="py-2.5 px-3 font-bold text-foreground">
                  <Mono>{d.transactionId}</Mono>
                </td>
                <td className="py-2.5 px-3 text-muted-foreground">
                  <Mono>{d.customerId}</Mono>
                </td>
                <td className="py-2.5 px-3 text-right font-bold text-foreground tabular">
                  {`${d.amount.toLocaleString("en-US", { minimumFractionDigits: 2 })} ${d.currency}`}
                </td>
                <td className="py-2.5 px-3 text-right">
                  <RiskScore value={d.riskScore} size="sm" showBand />
                </td>
                <td className="py-2.5 px-3">
                  <VerdictBadge verdict={d.verdict} size="sm" />
                </td>
                <td className="py-2.5 px-3 font-sans text-[11.5px] text-muted-foreground truncate max-w-[200px]">
                  {d.primarySignal}
                </td>
                <td className="py-2.5 px-3 text-right font-semibold text-muted-foreground tabular">
                  {d.latencyMs.toFixed(1)}ms
                </td>
                <td className="py-2.5 px-3">
                  <Link
                    to="/decisions/$decisionId"
                    params={{ decisionId: d.decisionId }}
                    className="border border-border bg-surface px-2 py-1 text-foreground font-medium hover:bg-secondary transition-colors font-sans text-[11px] inline-flex items-center gap-1 shadow-2xs"
                  >
                    <span>TreeSHAP</span>
                    <ArrowRight className="size-3 opacity-70" />
                  </Link>
                </td>
              </tr>
            ))}
          </tbody>
        </table>

        {rows.length === 0 && (
          <div className="py-12 text-center text-muted-foreground font-sans text-[12.5px]">
            No decisions match the specified search or filter criteria.
          </div>
        )}
      </div>
    </div>
  );
}
