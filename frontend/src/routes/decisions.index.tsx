import { createFileRoute, Link } from "@tanstack/react-router";
import { useSuspenseQuery } from "@tanstack/react-query";
import { useMemo, useState } from "react";
import { DemoTag, Mono, RiskScore, VerdictBadge } from "@/components/ropus/core";
import { decisionsQuery, isLiveBackend } from "@/lib/ropus/api";
import type { Verdict } from "@/lib/ropus/contracts";
import { cn } from "@/lib/utils";

export const Route = createFileRoute("/decisions/")({
  head: () => ({
    meta: [
      { title: "Risk Decisions — ROPUS" },
      {
        name: "description",
        content: "Search and filter every risk decision returned by the ROPUS evaluation API.",
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

const filters: Array<Verdict | "ALL"> = ["ALL", "APPROVE", "REVIEW", "CHALLENGE", "BLOCK"];
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

  return (
    <div className="mx-auto w-full max-w-[1560px] px-5 py-5 lg:px-7">
      <header className="flex flex-wrap items-start justify-between gap-4">
        <div>
          <h1 className="text-[19px] font-semibold tracking-tight">Risk decisions</h1>
          <p className="mt-1 max-w-[62ch] text-[13px] text-muted-foreground">
            Every evaluation returned by <Mono>POST /v1/risk/evaluate</Mono>, with the verdict,
            score and served latency. Open a decision for its full attribution.
          </p>
        </div>
        {!isLiveBackend && <DemoTag />}
      </header>

      <div className="mt-7 flex flex-wrap items-center justify-between gap-4 border-y border-border py-3">
        <div className="flex flex-wrap gap-4" role="group" aria-label="Filter by verdict">
          {filters.map((f) => (
            <button
              key={f}
              type="button"
              onClick={() => setVerdict(f)}
              aria-pressed={verdict === f}
              className={cn(
                "text-[12.5px]",
                verdict === f
                  ? "font-semibold text-foreground"
                  : "text-muted-foreground hover:text-foreground",
              )}
            >
              {f === "ALL" ? "All" : f}
            </button>
          ))}
        </div>
        <input
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          placeholder="Filter by transaction, customer or signal"
          aria-label="Filter decisions"
          className="w-[300px] max-w-full border border-border bg-transparent px-2.5 py-1.5 font-mono text-[12px] outline-none focus:border-primary"
        />
      </div>

      <table className="mt-2 w-full border-collapse text-[13px]">
        <caption className="sr-only">Risk decisions</caption>
        <thead>
          <tr className="border-b border-border">
            {[
              "Evaluated",
              "Decision",
              "Transaction",
              "Customer",
              "Amount",
              "Risk",
              "Verdict",
              "Latency",
              "Case",
            ].map((h, i) => (
              <th
                key={h}
                scope="col"
                className={cn(
                  "label-xs py-2 pr-4 font-semibold whitespace-nowrap",
                  i === 4 || i === 5 || i === 7 ? "text-right" : "text-left",
                )}
              >
                {h}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {rows.map((d) => (
            <tr key={d.decisionId} className="border-b border-border hover:bg-accent">
              <td className="py-2.5 pr-4">
                <Mono className="text-muted-foreground">{time(d.evaluatedAt)}</Mono>
              </td>
              <td className="py-2.5 pr-4">
                <Link
                  to="/decisions/$decisionId"
                  params={{ decisionId: d.decisionId }}
                  className="font-mono text-[12.5px] text-primary hover:underline"
                >
                  {d.decisionId}
                </Link>
              </td>
              <td className="py-2.5 pr-4">
                <Mono>{d.transactionId}</Mono>
              </td>
              <td className="py-2.5 pr-4">
                <Mono className="text-muted-foreground">{d.customerId}</Mono>
              </td>
              <td className="py-2.5 pr-4 text-right">
                <Mono>{`${d.amount.toLocaleString("en-US", { minimumFractionDigits: 2 })} ${d.currency}`}</Mono>
              </td>
              <td className="py-2.5 pr-4 text-right">
                <RiskScore value={d.riskScore} />
              </td>
              <td className="py-2.5 pr-4">
                <VerdictBadge verdict={d.verdict} />
              </td>
              <td className="py-2.5 pr-4 text-right">
                <Mono className="text-muted-foreground">{d.latencyMs.toFixed(1)}ms</Mono>
              </td>
              <td className="py-2.5">
                {d.caseId ? (
                  <Link
                    to="/cases/$caseId"
                    params={{ caseId: d.caseId }}
                    className="font-mono text-[12.5px] text-primary hover:underline"
                  >
                    {d.caseId}
                  </Link>
                ) : (
                  <span className="text-muted-foreground">—</span>
                )}
              </td>
            </tr>
          ))}
        </tbody>
      </table>

      {rows.length === 0 && (
        <p className="py-8 text-[13px] text-muted-foreground">No decisions match this filter.</p>
      )}
    </div>
  );
}
