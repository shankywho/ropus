import { createFileRoute, Link } from "@tanstack/react-router";
import { useState } from "react";
import { DataGrid, MetricStrip, Page, PageHead, SectionHead } from "@/components/ropus/page";
import { Mono, StatusPill } from "@/components/ropus/core";
import { investigations, type InvestigationRecord } from "@/lib/ropus/platform-fixtures";
import { cn } from "@/lib/utils";
import { FileSearch, Network, ShieldAlert, ArrowRight, CheckCircle2, Lock } from "lucide-react";
import { toast } from "sonner";

export const Route = createFileRoute("/investigations")({
  head: () => ({
    meta: [
      { title: "Evidentiary Dossiers & Investigations — ROPUS" },
      {
        name: "description",
        content:
          "Multi-case forensic investigations grouping related entities, mule clusters, and shared hardware canvas infrastructure.",
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
  ACTIVE: "text-blocked font-bold",
  MONITORING: "text-amber-intel font-bold",
  CLOSED: "text-muted-foreground",
};

const money = (n: number) =>
  new Intl.NumberFormat("en-IN", { style: "currency", currency: "INR" }).format(n);

function InvestigationsPage() {
  const open = investigations.filter((i) => i.state !== "CLOSED");
  const [selectedId, setSelectedId] = useState("inv_2026_0184");
  const activeInv = investigations.find((i) => i.id === selectedId) || investigations[0]!;

  const handleFreezeCluster = () => {
    toast.success(`Cluster frozen for ${activeInv.id}`, {
      description: "14 connected synthetic accounts and payout destination PA-77120 locked.",
    });
  };

  return (
    <Page>
      <PageHead
        title="Forensic Investigations &amp; Mule Dossiers"
        subtitle="Evidentiary dossiers grouping multi-case fraud clusters, shared headless hardware infrastructure, and cross-border money mule networks."
      />

      <MetricStrip
        items={[
          {
            label: "Active Investigations",
            value: String(open.length),
            sub: "active syndicate clusters",
            tone: "text-navy",
          },
          {
            label: "Entities Under Review",
            value: String(open.reduce((s, i) => s + i.entities, 0)),
            sub: "accounts, IPs & devices",
            tone: "text-blocked",
          },
          {
            label: "Linked Cases",
            value: String(investigations.flatMap((i) => i.linkedCases).length),
            sub: "attached to open dossiers",
          },
          {
            label: "Syndicate Exposure",
            value: money(open.reduce((s, i) => s + i.exposure * 100, 0)),
            sub: "total capital at risk",
            tone: "text-blocked",
          },
          { label: "Oldest Open", value: "18 days", sub: "inv_2026_0166" },
        ]}
      />

      {/* ------------------------------------------------ Active Evidentiary Dossier View */}
      <section className="mt-6 border border-border bg-card p-5 shadow-xs">
        <div className="flex flex-wrap items-start justify-between gap-4 border-b border-border pb-4">
          <div>
            <div className="flex items-center gap-2 font-mono text-[10.5px]">
              <span className="border border-navy bg-navy/10 px-2 py-0.5 font-bold text-navy uppercase tracking-wider">
                ACTIVE DOSSIER
              </span>
              <Mono className="text-[12px] font-bold text-foreground">{activeInv.id}</Mono>
              <span
                className={cn(
                  "font-bold uppercase tracking-wider text-[10px]",
                  stateTone[activeInv.state],
                )}
              >
                ({activeInv.state})
              </span>
            </div>
            <h2 className="mt-1 font-sans text-[18px] font-extrabold text-foreground tracking-tight">
              {activeInv.title}
            </h2>
            <p className="mt-0.5 font-sans text-[12.5px] text-muted-foreground">
              Lead Fraud Investigator:{" "}
              <Mono className="text-foreground font-bold">{activeInv.owner}</Mono> · Total Syndicate
              Exposure:{" "}
              <Mono className="text-blocked font-bold text-[13px]">
                {money(activeInv.exposure * 100)}
              </Mono>
            </p>
          </div>

          <div className="flex items-center gap-2 font-mono text-[11px]">
            <button
              type="button"
              onClick={handleFreezeCluster}
              className="border border-blocked bg-blocked px-3 py-1.5 font-bold text-white hover:bg-blocked/90 cursor-pointer shadow-xs flex items-center gap-1.5"
            >
              <Lock className="size-3" />
              <span>Freeze Syndicate Cluster</span>
            </button>
            <Link
              to="/graph"
              className="border border-border bg-surface hover:bg-secondary px-3 py-1.5 font-bold text-navy flex items-center gap-1.5 transition-colors shadow-2xs"
            >
              <Network className="size-3.5 text-navy" />
              <span>Explore in Fraud Graph →</span>
            </Link>
          </div>
        </div>

        {/* 3-Column Dossier Decomposition */}
        <div className="mt-4 grid grid-cols-1 gap-4 lg:grid-cols-3 font-mono text-[11.5px]">
          <div className="border border-border bg-surface p-3.5 space-y-2 shadow-2xs">
            <div className="text-[9.5px] font-bold text-navy uppercase tracking-wider border-b border-border pb-1.5">
              1. Observed Forensic Facts
            </div>
            <ul className="space-y-1.5 text-muted-foreground font-sans text-[11.5px] list-disc list-inside">
              <li>
                Inbound session from Limassol datacenter proxy{" "}
                <Mono className="text-[11px] font-bold text-foreground">198.51.100.44</Mono> (ASN
                13335).
              </li>
              <li>
                Headless Linux emulator canvas{" "}
                <Mono className="text-[11px] font-bold text-foreground">9f8a84b12c</Mono> with 0.96
                entropy score.
              </li>
              <li>
                Mule beneficiary node{" "}
                <Mono className="text-[11px] font-bold text-foreground">PA-77120</Mono> created 9
                minutes before payout attempt.
              </li>
            </ul>
          </div>

          <div className="border border-border bg-surface p-3.5 space-y-2 shadow-2xs">
            <div className="text-[9.5px] font-bold text-shadow-intel uppercase tracking-wider border-b border-border pb-1.5">
              2. GraphSAGE GNN Ring Topology
            </div>
            <ul className="space-y-1.5 text-muted-foreground font-sans text-[11.5px] list-disc list-inside">
              <li>
                <strong>14 synthetic accounts</strong> share the identical hardware canvas
                fingerprint.
              </li>
              <li>
                Degree centrality = <strong>16</strong> (High-density syndicate ring structure).
              </li>
              <li>3 connected nodes have confirmed historical chargebacks on record.</li>
            </ul>
          </div>

          <div className="border border-border bg-surface p-3.5 space-y-2 shadow-2xs">
            <div className="text-[9.5px] font-bold text-blocked uppercase tracking-wider border-b border-border pb-1.5">
              3. Recommended Legal Actions
            </div>
            <ul className="space-y-1.5 text-muted-foreground font-sans text-[11.5px] list-disc list-inside">
              <li>
                Permanent payment refusal on transaction{" "}
                <Mono className="text-[11px]">txn_order_88419</Mono>.
              </li>
              <li>Freeze 14 connected accounts in graph syndicate ring.</li>
              <li>
                Add Beneficiary <Mono className="text-[11px]">PA-77120</Mono> to global blacklist.
              </li>
              <li>Export cryptographic audit evidence bundle for FIU-IND STR submission.</li>
            </ul>
          </div>
        </div>
      </section>

      {/* ------------------------------------------------ All Investigations Table */}
      <div className="mt-8">
        <SectionHead title="Investigation Register" meta={`${investigations.length} dossiers`} />
        <div className="mt-2 border border-border bg-card shadow-xs overflow-x-auto">
          <DataGrid
            columns={[
              { key: "id", label: "Investigation ID" },
              { key: "title", label: "Subject" },
              { key: "entities", label: "Entities", align: "right" },
              { key: "cases", label: "Linked Cases" },
              { key: "exposure", label: "Exposure", align: "right" },
              { key: "owner", label: "Lead Owner" },
              { key: "state", label: "State" },
              { key: "opened", label: "Opened" },
              { key: "activity", label: "Last Activity" },
            ]}
            rows={investigations.map((i) => ({
              id: i.id,
              cells: [
                <button
                  key="id"
                  type="button"
                  onClick={() => setSelectedId(i.id)}
                  className={cn(
                    "font-mono text-[11.5px] hover:underline cursor-pointer font-bold",
                    selectedId === i.id ? "text-navy" : "text-muted-foreground",
                  )}
                >
                  {i.id}
                </button>,
                <span key="title" className="font-sans font-semibold text-foreground text-[12px]">
                  {i.title}
                </span>,
                <Mono key="ent" className="text-muted-foreground font-semibold">
                  {i.entities}
                </Mono>,
                i.linkedCases.length ? (
                  <span key="cs" className="flex flex-wrap gap-x-2 font-mono text-[11px]">
                    {i.linkedCases.map((c) => (
                      <Link
                        key={c}
                        to="/cases/$caseId"
                        params={{ caseId: c }}
                        className="text-navy font-bold hover:underline"
                      >
                        {c}
                      </Link>
                    ))}
                  </span>
                ) : (
                  <span key="cs" className="text-muted-foreground font-mono text-[11px]">
                    —
                  </span>
                ),
                <Mono key="exp" className="font-bold text-foreground tabular">
                  {money(i.exposure * 100)}
                </Mono>,
                <span
                  key="own"
                  className="text-muted-foreground font-medium font-sans text-[11.5px]"
                >
                  {i.owner}
                </span>,
                <span
                  key="st"
                  className={cn("font-mono text-[10.5px] font-bold uppercase", stateTone[i.state])}
                >
                  {i.state}
                </span>,
                <Mono key="op" className="text-muted-foreground text-[10.5px]">
                  {i.opened}
                </Mono>,
                <Mono key="act" className="text-muted-foreground text-[10.5px]">
                  {i.lastActivity}
                </Mono>,
              ],
            }))}
          />
        </div>
      </div>
    </Page>
  );
}
