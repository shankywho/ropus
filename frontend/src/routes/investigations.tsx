import { createFileRoute, Link } from "@tanstack/react-router";
import { useState } from "react";
import { DataGrid, MetricStrip, Page, PageHead, SectionHead } from "@/components/ropus/page";
import { Mono } from "@/components/ropus/core";
import { investigations, type InvestigationRecord } from "@/lib/ropus/platform-fixtures";
import { cn } from "@/lib/utils";

export const Route = createFileRoute("/investigations")({
  head: () => ({
    meta: [
      { title: "Investigations — ROPUS" },
      {
        name: "description",
        content:
          "Multi-case investigations grouping related entities, exposure and analyst ownership.",
      },
      { property: "og:title", content: "Investigations — ROPUS" },
      {
        property: "og:description",
        content: "Entity-level investigations spanning multiple cases and decisions.",
      },
      { property: "og:type", content: "website" },
      { name: "twitter:card", content: "summary" },
    ],
  }),
  component: InvestigationsPage,
});

const stateTone: Record<InvestigationRecord["state"], string> = {
  ACTIVE: "text-block",
  MONITORING: "text-warning",
  CLOSED: "text-muted-foreground",
};

const usd = (n: number) => n.toLocaleString("en-US", { minimumFractionDigits: 2 });

function InvestigationsPage() {
  const open = investigations.filter((i) => i.state !== "CLOSED");
  const [selectedId, setSelectedId] = useState("inv_2026_0184");
  const activeInv = investigations.find((i) => i.id === selectedId) || investigations[0]!;

  return (
    <Page>
      <PageHead
        title="Investigations"
        subtitle="Evidentiary dossiers grouping multi-case fraud clusters, shared hardware infrastructure, and cross-border money mule networks."
      />

      <MetricStrip
        items={[
          { label: "Open investigations", value: String(open.length), sub: "active or monitoring" },
          {
            label: "Entities under review",
            value: String(open.reduce((s, i) => s + i.entities, 0)),
            sub: "across all clusters",
          },
          {
            label: "Linked cases",
            value: String(investigations.flatMap((i) => i.linkedCases).length),
            sub: "attached to investigations",
          },
          {
            label: "Exposure",
            value: `₹${usd(open.reduce((s, i) => s + i.exposure * 100, 0))}`,
            sub: "INR at risk, open only",
          },
          { label: "Oldest open", value: "18 days", sub: "inv_2026_0166" },
        ]}
      />

      {/* ------------------------------------------------ Active Evidentiary Dossier View */}
      <section className="mt-6 rounded border border-border bg-surface p-5">
        <div className="flex flex-wrap items-start justify-between gap-4 border-b border-border pb-4">
          <div>
            <div className="flex items-center gap-2">
              <span className="font-mono text-[11px] font-bold text-primary">ACTIVE DOSSIER</span>
              <Mono className="text-[12px] font-bold">{activeInv.id}</Mono>
              <span
                className={cn(
                  "text-[11px] font-semibold tracking-wider uppercase",
                  stateTone[activeInv.state],
                )}
              >
                ({activeInv.state})
              </span>
            </div>
            <h2 className="mt-1 text-[16px] font-bold text-foreground">{activeInv.title}</h2>
            <p className="mt-0.5 text-[12.5px] text-muted-foreground">
              Lead Analyst: <Mono className="text-foreground font-bold">{activeInv.owner}</Mono> ·
              Total Syndicate Exposure:{" "}
              <Mono className="text-foreground font-bold">
                ₹{usd(activeInv.exposure * 100)} INR
              </Mono>
            </p>
          </div>

          <div className="flex items-center gap-3">
            <Link
              to="/graph"
              className="rounded border border-border bg-surface px-3 py-1.5 font-mono text-[12px] font-bold text-primary hover:bg-accent"
            >
              Open Cluster in Fraud Graph →
            </Link>
          </div>
        </div>

        <div className="mt-4 grid grid-cols-1 gap-4 lg:grid-cols-3 text-[12.5px]">
          <div className="rounded border border-border bg-accent/20 p-3.5">
            <div className="text-[11px] font-semibold text-muted-foreground uppercase">
              1. Observed Facts
            </div>
            <ul className="mt-2 space-y-1.5 text-muted-foreground list-disc list-inside">
              <li>
                • Session from datacenter proxy <Mono className="text-[11px]">198.51.100.44</Mono>.
              </li>
              <li>
                • Headless Linux emulator canvas <Mono className="text-[11px]">9f8a84b12c</Mono>.
              </li>
              <li>
                • Beneficiary node <Mono className="text-[11px]">PA-77120</Mono> added 9m before
                transfer.
              </li>
            </ul>
          </div>

          <div className="rounded border border-border bg-accent/20 p-3.5">
            <div className="text-[11px] font-semibold text-muted-foreground uppercase">
              2. Inferred Graph Structure
            </div>
            <ul className="mt-2 space-y-1.5 text-muted-foreground list-disc list-inside">
              <li>• 14 synthetic accounts share same hardware canvas hash.</li>
              <li>• Degree centrality = 16 (high density syndicate ring).</li>
              <li>• 3 linked accounts have confirmed prior chargebacks.</li>
            </ul>
          </div>

          <div className="rounded border border-border bg-accent/20 p-3.5">
            <div className="text-[11px] font-semibold text-muted-foreground uppercase">
              3. Recommended Actions
            </div>
            <ul className="mt-2 space-y-1.5 text-muted-foreground">
              <li>• Execute immediate permanent block on payout request.</li>
              <li>• Freeze 14 connected mule accounts in graph ring.</li>
              <li>
                • Add Beneficiary <Mono className="text-[11px]">PA-77120</Mono> to global blacklist.
              </li>
              <li>• File mandatory FIU-IND STR/SAR narrative batch.</li>
            </ul>
          </div>
        </div>
      </section>

      {/* ------------------------------------------------ All Investigations Table */}
      <div className="mt-8">
        <SectionHead title="Investigation register" meta={`${investigations.length} records`} />
        <div className="mt-1">
          <DataGrid
            columns={[
              { key: "id", label: "Investigation" },
              { key: "title", label: "Subject" },
              { key: "entities", label: "Entities", align: "right" },
              { key: "cases", label: "Linked cases" },
              { key: "exposure", label: "Exposure INR", align: "right" },
              { key: "owner", label: "Owner" },
              { key: "state", label: "State" },
              { key: "opened", label: "Opened" },
              { key: "activity", label: "Last activity" },
            ]}
            rows={investigations.map((i) => ({
              id: i.id,
              cells: [
                <button
                  key="id"
                  type="button"
                  onClick={() => setSelectedId(i.id)}
                  className={cn(
                    "font-mono text-[12px] hover:underline",
                    selectedId === i.id ? "font-bold text-primary" : "text-muted-foreground",
                  )}
                >
                  {i.id}
                </button>,
                <span className="font-medium">{i.title}</span>,
                <Mono className="text-muted-foreground">{i.entities}</Mono>,
                i.linkedCases.length ? (
                  <span className="flex flex-wrap gap-x-2">
                    {i.linkedCases.map((c) => (
                      <Link
                        key={c}
                        to="/cases/$caseId"
                        params={{ caseId: c }}
                        className="font-mono text-[12px] text-primary hover:underline"
                      >
                        {c}
                      </Link>
                    ))}
                  </span>
                ) : (
                  <span className="text-muted-foreground">—</span>
                ),
                <Mono>{usd(i.exposure)}</Mono>,
                <span className="text-muted-foreground">{i.owner}</span>,
                <span
                  className={cn("text-[11px] font-semibold tracking-[0.06em]", stateTone[i.state])}
                >
                  {i.state}
                </span>,
                <Mono className="text-muted-foreground">{i.opened}</Mono>,
                <Mono className="text-muted-foreground">{i.lastActivity}</Mono>,
              ],
            }))}
          />
        </div>
      </div>
    </Page>
  );
}
