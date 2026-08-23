import { createFileRoute, Link } from "@tanstack/react-router";
import { useState } from "react";
import {
  CaseStatusTag,
  CaseTimeline,
  DemoTag,
  EvidenceList,
  Mono,
  PriorityTag,
  RiskFactorList,
  RiskScore,
  SlaTag,
  VerdictBadge,
} from "@/components/ropus/core";
import { Disclosure } from "@/components/ropus/primitives";
import { useSuspenseQuery } from "@tanstack/react-query";
import { caseQuery, decisionQuery } from "@/lib/ropus/api";
import type { CaseEvent, CaseRecord } from "@/lib/ropus/contracts";
import { cn } from "@/lib/utils";

export const Route = createFileRoute("/cases/$caseId")({
  head: () => ({
    meta: [
      { title: "Case detail — ROPUS" },
      {
        name: "description",
        content: "Case outcome, evidence, score attribution and append-only investigation history.",
      },
      { property: "og:title", content: "Case detail — ROPUS" },
      {
        property: "og:description",
        content: "Case outcome, evidence, score attribution and append-only investigation history.",
      },
      { property: "og:type", content: "website" },
      { name: "twitter:card", content: "summary" },
    ],
  }),
  loader: ({ context, params }) => context.queryClient.ensureQueryData(caseQuery(params.caseId)),
  component: CaseDetail,
});

const money = (amount: number, currency: string) =>
  `${amount.toLocaleString("en-US", { minimumFractionDigits: 2 })} ${currency}`;

function CaseDetail() {
  const { caseId } = Route.useParams();
  const { data: record } = useSuspenseQuery(caseQuery(caseId));

  if (!record) {
    return (
      <div className="mx-auto w-full max-w-[1560px] px-5 py-12 lg:px-7">
        <h1 className="text-[19px] font-semibold tracking-tight">Case not found</h1>
        <p className="mt-2 text-[13px] text-muted-foreground">
          No case matches <Mono>{caseId}</Mono> in this tenant.
        </p>
        <Link to="/cases" className="mt-4 inline-block text-[13px] text-primary hover:underline">
          Back to queue
        </Link>
      </div>
    );
  }

  return <CaseBody record={record} />;
}

function CaseBody({ record }: { record: CaseRecord }) {
  const { data: decision } = useSuspenseQuery(decisionQuery(record.decisionId));
  const [extra, setExtra] = useState<CaseEvent[]>([]);
  const [resolution, setResolution] = useState<string | null>(null);

  const closed = record.status === "CLOSED" || resolution !== null;


  const resolve = (label: string, text: string) => {
    setResolution(label);
    setExtra((prev) => [
      ...prev,
      {
        id: `ev_local_${prev.length}`,
        at: new Date().toISOString(),
        actor: "m.okafor",
        actorKind: "ANALYST",
        text,
      },
    ]);
  };

  return (
    <div className="mx-auto w-full max-w-[1560px] px-5 py-5 lg:px-7">
      <Link to="/cases" className="text-[12px] text-muted-foreground hover:text-foreground">
        ← Queue
      </Link>

      <header className="mt-3 flex flex-wrap items-start justify-between gap-6 border-b border-border pb-6">
        <div className="min-w-0">
          <div className="flex flex-wrap items-center gap-3">
            <h1 className="font-mono text-[19px] font-semibold tracking-tight">{record.caseId}</h1>
            <PriorityTag priority={record.priority} />
            <CaseStatusTag status={record.status} />
            <VerdictBadge verdict={record.verdict} />
          </div>
          <p className="mt-2 max-w-[70ch] text-[13.5px] leading-relaxed text-muted-foreground">
            {record.summary}
          </p>
        </div>
        <div className="flex items-start gap-8">
          <div className="text-right">
            <RiskScore value={record.riskScore} size="hero" showBand />
          </div>
          <DemoTag />
        </div>
      </header>

      <div className="mt-8 grid gap-10 lg:grid-cols-[minmax(0,1fr)_320px]">
        <div className="min-w-0 space-y-10">
          <section>
            <h2 className="label-xs">Evidence</h2>
            <div className="mt-4 space-y-6">
              <EvidenceList items={decision.evidence} kind="OBSERVED" />
              <EvidenceList items={decision.evidence} kind="INFERRED" />
              <EvidenceList items={decision.evidence} kind="RECOMMENDED" />
            </div>
          </section>

          <section>
            <h2 className="label-xs">Investigation history</h2>
            <p className="mt-1 text-[12px] text-muted-foreground">
              Append-only. Each entry is hash-chained into the tenant audit log.
            </p>
            <div className="mt-3">
              <CaseTimeline events={[...record.timeline, ...extra]} />
            </div>
          </section>

          <Disclosure summary="Score attribution from the originating decision">
            <div className="pt-3">
              <RiskFactorList factors={decision.factors} />
            </div>
          </Disclosure>
        </div>

        <aside className="space-y-8">
          <section>
            <h2 className="label-xs">Case record</h2>
            <dl className="mt-3 divide-y divide-border border-y border-border">
              {(
                [
                  ["Customer", <Mono key="c">{record.customerId}</Mono>],
                  ["Transaction", <Mono key="t">{record.transactionId}</Mono>],
                  [
                    "Decision",
                    <Link
                      key="d"
                      to="/decisions/$decisionId"
                      params={{ decisionId: record.decisionId }}
                      className="font-mono text-[12.5px] text-primary hover:underline"
                    >
                      {record.decisionId}
                    </Link>,
                  ],
                  ["Amount", <Mono key="a">{money(record.amount, record.currency)}</Mono>],
                  ["Opened by", <Mono key="o">{record.openedBy}</Mono>],
                  ["Opened", <Mono key="oa">{record.openedAt.replace("T", " ")}</Mono>],
                  ["Assignee", record.assignee ?? "Unassigned"],
                  ["SLA", <SlaTag key="s" minutes={record.slaMinutesRemaining} />],
                  [
                    "Outcome",
                    resolution ?? (record.outcome ? record.outcome.replace("_", " ").toLowerCase() : "pending"),
                  ],
                ] as Array<[string, React.ReactNode]>
              ).map(([k, v]) => (
                <div key={k} className="flex items-baseline justify-between gap-6 py-1.5">
                  <dt className="label-xs">{k}</dt>
                  <dd className="text-right text-[12.5px]">{v}</dd>
                </div>
              ))}
            </dl>
          </section>

          <section>
            <h2 className="label-xs">Resolution</h2>
            <p className="mt-1 text-[12px] text-muted-foreground">
              Requires a case outcome. The backend remains the authority on whether the action is permitted.
            </p>
            {closed ? (
              <p className="mt-3 text-[12.5px]">
                Closed —{" "}
                <span className="font-medium">
                  {resolution ?? record.outcome?.replace("_", " ").toLowerCase()}
                </span>
              </p>
            ) : (
              <div className="mt-3 flex flex-col gap-2">
                {([
                  ["Confirm fraud", "Confirmed fraud. Transaction held and payee added to the deny list."],
                  ["Release transaction", "Released the transaction to settlement after verification."],
                  ["Issue challenge", "Issued a step-up challenge to the customer."],
                ] as Array<[string, string]>).map(([label, text], i) => (
                  <button
                    key={label}
                    type="button"
                    onClick={() => resolve(label.toLowerCase(), text)}
                    className={cn(
                      "border px-3 py-1.5 text-left text-[12.5px] font-medium",
                      i === 0
                        ? "border-block text-block hover:bg-block/5"
                        : "border-border-strong hover:bg-accent",
                    )}
                  >
                    {label}
                  </button>
                ))}
              </div>
            )}
          </section>

          <section>
            <h2 className="label-xs">Related</h2>
            <div className="mt-3 flex flex-col gap-1.5 text-[12.5px]">
              <Link to="/graph" className="text-primary hover:underline">
                Fraud graph for {record.customerId}
              </Link>
              <Link
                to="/decisions/$decisionId"
                params={{ decisionId: record.decisionId }}
                className="text-primary hover:underline"
              >
                Full decision record
              </Link>
            </div>
          </section>
        </aside>
      </div>
    </div>
  );
}
