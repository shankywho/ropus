import { createFileRoute } from "@tanstack/react-router";
import { DataGrid, MetricStrip, Page, PageHead, SectionHead } from "@/components/ropus/page";
import { Mono } from "@/components/ropus/core";
import { indicators, type IndicatorRecord } from "@/lib/ropus/platform-fixtures";
import { cn } from "@/lib/utils";

export const Route = createFileRoute("/threat-intelligence")({
  head: () => ({
    meta: [
      { title: "Threat Intelligence — ROPUS" },
      {
        name: "description",
        content:
          "Indicator register: IPs, networks, devices, payout accounts and BINs with feed provenance and confidence.",
      },
      { property: "og:title", content: "Threat Intelligence — ROPUS" },
      {
        property: "og:description",
        content: "Indicators of compromise feeding the risk score, with provenance and confidence.",
      },
      { property: "og:type", content: "website" },
      { name: "twitter:card", content: "summary" },
    ],
  }),
  component: ThreatIntelPage,
});

const classTone: Record<IndicatorRecord["classification"], string> = {
  MALICIOUS: "text-blocked font-bold",
  SUSPICIOUS: "text-shadow-intel font-bold",
  BENIGN: "text-muted-foreground",
};

const feeds = [
  {
    name: "internal-graph",
    records: "1,204,884",
    refreshed: "2026-08-22 17:40Z",
    state: "HEALTHY",
  },
  { name: "consortium-fraud", records: "88,412", refreshed: "2026-08-22 06:00Z", state: "HEALTHY" },
  { name: "proxy-registry", records: "412,009", refreshed: "2026-08-22 00:15Z", state: "HEALTHY" },
  {
    name: "internal-device",
    records: "9,884,201",
    refreshed: "2026-08-22 17:41Z",
    state: "HEALTHY",
  },
  { name: "issuer-registry", records: "301,884", refreshed: "2026-08-18 04:00Z", state: "STALE" },
];

function ThreatIntelPage() {
  return (
    <Page>
      <PageHead
        title="Threat Intelligence &amp; Feeds"
        subtitle="Indicators contributed by internal graph analysis, consortium reporting and external registries. Indicators add score; they never issue a verdict on their own."
      />

      <MetricStrip
        items={[
          { label: "Indicators", value: "1,986,506", sub: "across five feeds" },
          {
            label: "Malicious",
            value: String(indicators.filter((i) => i.classification === "MALICIOUS").length),
            sub: "matched in last 24h",
            tone: "text-blocked",
          },
          {
            label: "Matches",
            value: indicators.reduce((s, i) => s + i.hits24h, 0).toLocaleString(),
            sub: "last 24 hours",
          },
          { label: "Feeds Healthy", value: "4 / 5", sub: "issuer-registry is stale", tone: "text-shadow-intel" },
          { label: "Score Ceiling", value: "+0.14", sub: "max threat-intel contribution", tone: "text-authoritative" },
        ]}
      />

      <div className="mt-6 grid gap-8 xl:grid-cols-[minmax(0,1fr)_340px] xl:gap-10">
        <section className="min-w-0 space-y-4">
          <SectionHead
            title="INDICATORS MATCHED IN THIS WINDOW"
            meta={`${indicators.length} records`}
          />
          <div className="border border-border bg-card shadow-xs">
            <DataGrid
              columns={[
                { key: "value", label: "Indicator" },
                { key: "type", label: "Type" },
                { key: "class", label: "Classification" },
                { key: "feed", label: "Feed" },
                { key: "conf", label: "Confidence", align: "right" },
                { key: "hits", label: "Hits 24h", align: "right" },
                { key: "first", label: "First Seen" },
                { key: "last", label: "Last Seen" },
              ]}
              rows={indicators.map((i) => ({
                id: i.value,
                cells: [
                  <Mono key="val" className="font-bold text-navy text-[11.5px]">{i.value}</Mono>,
                  <span key="type" className="text-muted-foreground font-sans text-[11.5px]">{i.type}</span>,
                  <span
                    key="class"
                    className={cn(
                      "font-mono text-[10px] uppercase font-bold tracking-wider",
                      classTone[i.classification],
                    )}
                  >
                    {i.classification}
                  </span>,
                  <Mono key="feed" className="text-muted-foreground text-[11px]">{i.feed}</Mono>,
                  <Mono key="conf" className="font-semibold">{i.confidence.toFixed(2)}</Mono>,
                  <Mono key="hits" className="text-muted-foreground font-medium">{i.hits24h.toLocaleString()}</Mono>,
                  <Mono key="first" className="text-muted-foreground text-[10.5px]">{i.firstSeen}</Mono>,
                  <Mono key="last" className="text-muted-foreground text-[10.5px]">{i.lastSeen}</Mono>,
                ],
              }))}
            />
          </div>
        </section>

        <aside className="min-w-0 space-y-3">
          <SectionHead title="ACTIVE FEEDS" />
          <div className="border border-border bg-card p-3 shadow-xs">
            <table className="w-full text-[11.5px] font-mono">
              <tbody>
                {feeds.map((f) => (
                  <tr key={f.name} className="border-b border-border/50 last:border-b-0">
                    <td className="py-2 pr-3 font-semibold text-foreground">
                      <Mono className="text-[11.5px]">{f.name}</Mono>
                    </td>
                    <td className="py-2 pr-3 text-right">
                      <Mono className="text-muted-foreground">{f.records}</Mono>
                    </td>
                    <td
                      className={cn(
                        "py-2 text-right text-[10px] font-bold tracking-[0.06em] uppercase",
                        f.state === "STALE" ? "text-amber-intel" : "text-authoritative",
                      )}
                    >
                      {f.state}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          <p className="text-[11px] font-sans text-muted-foreground leading-relaxed">
            issuer-registry last refreshed 2026-08-18; BIN classifications older than 72 hours are
            ignored by the scoring path.
          </p>
        </aside>
      </div>
    </Page>
  );
}
