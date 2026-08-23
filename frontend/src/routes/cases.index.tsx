import { createFileRoute, Link } from "@tanstack/react-router";
import { useMemo, useState } from "react";
import {
  CaseStatusTag,
  DemoTag,
  Mono,
  PriorityTag,
  RiskScore,
  SlaTag,
  VerdictBadge,
} from "@/components/ropus/core";
import { useSuspenseQuery } from "@tanstack/react-query";
import { casesQuery } from "@/lib/ropus/api";
import type { CaseStatus } from "@/lib/ropus/contracts";
import { cn } from "@/lib/utils";

export const Route = createFileRoute("/cases/")({
  head: () => ({
    meta: [
      { title: "Cases — ROPUS" },
      {
        name: "description",
        content: "Investigation queue for blocked, challenged and review-flagged transactions.",
      },
      { property: "og:title", content: "Cases — ROPUS" },
      {
        property: "og:description",
        content: "Investigation queue for blocked, challenged and review-flagged transactions.",
      },
      { property: "og:type", content: "website" },
      { name: "twitter:card", content: "summary" },
    ],
  }),
  loader: ({ context }) => context.queryClient.ensureQueryData(casesQuery()),
  component: CaseQueue,
});

const filters: Array<{ id: "ALL" | CaseStatus; label: string }> = [
  { id: "ALL", label: "All" },
  { id: "OPEN", label: "Open" },
  { id: "IN_REVIEW", label: "In review" },
  { id: "ESCALATED", label: "Escalated" },
  { id: "CLOSED", label: "Closed" },
];

const money = (amount: number, currency: string) =>
  `${amount.toLocaleString("en-US", { minimumFractionDigits: 2 })} ${currency}`;

function CaseQueue() {
  const { data: allCases } = useSuspenseQuery(casesQuery());
  const [status, setStatus] = useState<"ALL" | CaseStatus>("ALL");
  const rows = useMemo(
    () => (status === "ALL" ? allCases : allCases.filter((c) => c.status === status)),
    [allCases, status],
  );

  const open = allCases.filter((c) => c.status !== "CLOSED");
  const breached = open.filter((c) => c.slaMinutesRemaining < 0);

  return (
    <div className="mx-auto w-full max-w-[1560px] px-5 py-5 lg:px-7">
      <header className="flex flex-wrap items-baseline justify-between gap-4">
        <div>
          <h1 className="text-[19px] font-semibold tracking-tight">Cases</h1>
          <p className="mt-1 max-w-[62ch] text-[13px] text-muted-foreground">
            Every case originates from a decision. Nothing is closed without a recorded outcome and an
            append-only history.
          </p>
        </div>
        <DemoTag />
      </header>

      <dl className="mt-7 flex flex-wrap gap-x-12 gap-y-4 border-y border-border py-4">
        {[
          ["Open", String(open.length)],
          ["P1", String(open.filter((c) => c.priority === "P1").length)],
          ["SLA breached", String(breached.length)],
          ["Unassigned", String(open.filter((c) => !c.assignee).length)],
        ].map(([label, value]) => (
          <div key={label}>
            <dd className="font-mono text-[20px] leading-none font-semibold tabular">{value}</dd>
            <dt className="mt-1.5 text-[11.5px] text-muted-foreground">{label}</dt>
          </div>
        ))}
      </dl>

      <nav aria-label="Filter by status" className="mt-6 flex flex-wrap gap-4">
        {filters.map((f) => (
          <button
            key={f.id}
            type="button"
            onClick={() => setStatus(f.id)}
            aria-pressed={status === f.id}
            className={cn(
              "border-b-2 pb-1 text-[12.5px]",
              status === f.id
                ? "border-b-primary font-semibold text-foreground"
                : "border-b-transparent text-muted-foreground hover:text-foreground",
            )}
          >
            {f.label}
          </button>
        ))}
      </nav>

      <table className="mt-2 w-full border-collapse text-[13px]">
        <caption className="sr-only">Case queue</caption>
        <thead>
          <tr className="border-b border-border">
            {["Case", "Pri", "Customer", "Amount", "Risk", "Verdict", "Assignee", "Status", "SLA"].map(
              (h, i) => (
                <th
                  key={h}
                  scope="col"
                  className={cn(
                    "label-xs py-2 pr-4 font-semibold whitespace-nowrap",
                    i > 3 && i < 6 ? "text-right" : "text-left",
                    i === 8 && "pr-0 text-right",
                  )}
                >
                  {h}
                </th>
              ),
            )}
          </tr>
        </thead>
        <tbody>
          {rows.map((c) => (
            <tr key={c.caseId} className="border-b border-border hover:bg-accent">
              <td className="py-2.5 pr-4">
                <Link
                  to="/cases/$caseId"
                  params={{ caseId: c.caseId }}
                  className="font-mono text-[12.5px] text-primary hover:underline"
                >
                  {c.caseId}
                </Link>
              </td>
              <td className="py-2.5 pr-4">
                <PriorityTag priority={c.priority} />
              </td>
              <td className="py-2.5 pr-4">
                <Mono className="text-muted-foreground">{c.customerId}</Mono>
              </td>
              <td className="py-2.5 pr-4">
                <Mono>{money(c.amount, c.currency)}</Mono>
              </td>
              <td className="py-2.5 pr-4 text-right">
                <RiskScore value={c.riskScore} />
              </td>
              <td className="py-2.5 pr-4 text-right">
                <VerdictBadge verdict={c.verdict} />
              </td>
              <td className="py-2.5 pr-4 text-[12.5px]">
                {c.assignee ?? <span className="text-muted-foreground">Unassigned</span>}
              </td>
              <td className="py-2.5 pr-4">
                <CaseStatusTag status={c.status} />
              </td>
              <td className="py-2.5 text-right">
                <SlaTag minutes={c.slaMinutesRemaining} />
              </td>
            </tr>
          ))}
        </tbody>
      </table>

      {rows.length === 0 && (
        <p className="py-8 text-[13px] text-muted-foreground">No cases in this state.</p>
      )}
    </div>
  );
}
