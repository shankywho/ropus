import { createFileRoute } from "@tanstack/react-router";
import { DataGrid, MetricStrip, Page, PageHead, SectionHead } from "@/components/ropus/page";
import { Mono } from "@/components/ropus/core";
import { indicators, type IndicatorRecord } from "@/lib/ropus/platform-fixtures";
import { cn } from "@/lib/utils";

export const Route = createFileRoute("/threat-intelligence")({
  head: () => ({
    meta: [
      { title: "Threat Intelligence — ROPUS" },
      { name: "description", content: "Indicator register: IPs, networks, devices, payout accounts and BINs with feed provenance and confidence." },
      { property: "og:title", content: "Threat Intelligence — ROPUS" },
      { property: "og:description", content: "Indicators of compromise feeding the risk score, with provenance and confidence." },
      { property: "og:type", content: "website" },
      { name: "twitter:card", content: "summary" },
    ],
  }),
  component: ThreatIntelPage,
});

const classTone: Record<IndicatorRecord["classification"], string> = {
  MALICIOUS: "text-block",
  SUSPICIOUS: "text-warning",
  BENIGN: "text-muted-foreground",
};

const feeds = [
  { name: "internal-graph", records: "1,204,884", refreshed: "2026-08-22 17:40Z", state: "HEALTHY" },
  { name: "consortium-fraud", records: "88,412", refreshed: "2026-08-22 06:00Z", state: "HEALTHY" },
  { name: "proxy-registry", records: "412,009", refreshed: "2026-08-22 00:15Z", state: "HEALTHY" },
  { name: "internal-device", records: "9,884,201", refreshed: "2026-08-22 17:41Z", state: "HEALTHY" },
  { name: "issuer-registry", records: "301,884", refreshed: "2026-08-18 04:00Z", state: "STALE" },
];

function ThreatIntelPage() {
  return (
    <Page>
      <PageHead
        title="Threat intelligence"
        subtitle="Indicators contributed by internal graph analysis, consortium reporting and external registries. Indicators add score; they never issue a verdict on their own."
      />

      <MetricStrip
        items={[
          { label: "Indicators", value: "1,986,506", sub: "across five feeds" },
          { label: "Malicious", value: String(indicators.filter((i) => i.classification === "MALICIOUS").length), sub: "matched in last 24h" },
          { label: "Matches", value: indicators.reduce((s, i) => s + i.hits24h, 0).toLocaleString(), sub: "last 24 hours" },
          { label: "Feeds healthy", value: "4 / 5", sub: "issuer-registry is stale" },
          { label: "Score ceiling", value: "+0.14", sub: "max threat-intel contribution" },
        ]}
      />

      <div className="mt-6 grid gap-8 xl:grid-cols-[minmax(0,1fr)_340px] xl:gap-10">
        <section className="min-w-0">
          <SectionHead title="Indicators matched in this window" meta={`${indicators.length} records`} />
          <div className="mt-1">
            <DataGrid
              columns={[
                { key: "value", label: "Indicator" },
                { key: "type", label: "Type" },
                { key: "class", label: "Classification" },
                { key: "feed", label: "Feed" },
                { key: "conf", label: "Confidence", align: "right" },
                { key: "hits", label: "Hits 24h", align: "right" },
                { key: "first", label: "First seen" },
                { key: "last", label: "Last seen" },
              ]}
              rows={indicators.map((i) => ({
                id: i.value,
                cells: [
                  <Mono className="font-semibold">{i.value}</Mono>,
                  <span className="text-muted-foreground">{i.type}</span>,
                  <span className={cn("text-[11px] font-semibold tracking-[0.06em]", classTone[i.classification])}>
                    {i.classification}
                  </span>,
                  <Mono className="text-muted-foreground">{i.feed}</Mono>,
                  <Mono>{i.confidence.toFixed(2)}</Mono>,
                  <Mono className="text-muted-foreground">{i.hits24h.toLocaleString()}</Mono>,
                  <Mono className="text-muted-foreground">{i.firstSeen}</Mono>,
                  <Mono className="text-muted-foreground">{i.lastSeen}</Mono>,
                ],
              }))}
            />
          </div>
        </section>

        <aside className="min-w-0">
          <SectionHead title="Feeds" />
          <table className="mt-1 w-full text-[12.5px]">
            <tbody>
              {feeds.map((f) => (
                <tr key={f.name} className="border-b border-border last:border-b-0">
                  <td className="py-2 pr-3">
                    <Mono>{f.name}</Mono>
                  </td>
                  <td className="py-2 pr-3 text-right">
                    <Mono className="text-muted-foreground">{f.records}</Mono>
                  </td>
                  <td
                    className={cn(
                      "py-2 text-right text-[11px] font-semibold tracking-[0.06em]",
                      f.state === "STALE" ? "text-warning" : "text-muted-foreground",
                    )}
                  >
                    {f.state}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
          <p className="mt-3 text-[11.5px] text-muted-foreground">
            issuer-registry last refreshed 2026-08-18; BIN classifications older than 72 hours are ignored by the
            scoring path.
          </p>
        </aside>
      </div>
    </Page>
  );
}
