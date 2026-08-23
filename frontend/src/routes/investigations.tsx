import { createFileRoute, Link } from "@tanstack/react-router";
import { DataGrid, MetricStrip, Page, PageHead, SectionHead } from "@/components/ropus/page";
import { Mono } from "@/components/ropus/core";
import { investigations, type InvestigationRecord } from "@/lib/ropus/platform-fixtures";
import { cn } from "@/lib/utils";

export const Route = createFileRoute("/investigations")({
  head: () => ({
    meta: [
      { title: "Investigations — ROPUS" },
      { name: "description", content: "Multi-case investigations grouping related entities, exposure and analyst ownership." },
      { property: "og:title", content: "Investigations — ROPUS" },
      { property: "og:description", content: "Entity-level investigations spanning multiple cases and decisions." },
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

  return (
    <Page>
      <PageHead
        title="Investigations"
        subtitle="Long-running work that spans more than one case: a payout cluster, a network origin, a device family. Cases attach to an investigation; decisions attach to cases."
      />

      <MetricStrip
        items={[
          { label: "Open investigations", value: String(open.length), sub: "active or monitoring" },
          { label: "Entities under review", value: String(open.reduce((s, i) => s + i.entities, 0)), sub: "across all graphs" },
          { label: "Linked cases", value: String(investigations.flatMap((i) => i.linkedCases).length), sub: "attached to investigations" },
          { label: "Exposure", value: usd(open.reduce((s, i) => s + i.exposure, 0)), sub: "USD at risk, open only" },
          { label: "Oldest open", value: "18 days", sub: "inv_2026_0166" },
        ]}
      />

      <div className="mt-6">
        <SectionHead title="Investigation register" meta={`${investigations.length} records`} />
        <div className="mt-1">
          <DataGrid
            columns={[
              { key: "id", label: "Investigation" },
              { key: "title", label: "Subject" },
              { key: "entities", label: "Entities", align: "right" },
              { key: "cases", label: "Linked cases" },
              { key: "exposure", label: "Exposure USD", align: "right" },
              { key: "owner", label: "Owner" },
              { key: "state", label: "State" },
              { key: "opened", label: "Opened" },
              { key: "activity", label: "Last activity" },
            ]}
            rows={investigations.map((i) => ({
              id: i.id,
              cells: [
                <Mono className="text-muted-foreground">{i.id}</Mono>,
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
                <span className={cn("text-[11px] font-semibold tracking-[0.06em]", stateTone[i.state])}>{i.state}</span>,
                <Mono className="text-muted-foreground">{i.opened}</Mono>,
                <Mono className="text-muted-foreground">{i.lastActivity}</Mono>,
              ],
            }))}
          />
        </div>
        <p className="mt-3 text-[11.5px] text-muted-foreground">
          Investigation <Mono className="text-foreground">inv_2026_0184</Mono> is the cluster behind the currently
          blocked wire —{" "}
          <Link to="/graph" className="text-primary hover:underline">
            open it in the fraud graph
          </Link>
          .
        </p>
      </div>
    </Page>
  );
}
