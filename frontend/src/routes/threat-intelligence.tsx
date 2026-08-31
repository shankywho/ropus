import { createFileRoute } from "@tanstack/react-router";
import { useState, useMemo } from "react";
import { DataGrid, MetricStrip, Page, PageHead, SectionHead } from "@/components/ropus/page";
import { Mono, StatusPill } from "@/components/ropus/core";
import { indicators, type IndicatorRecord } from "@/lib/ropus/platform-fixtures";
import { cn } from "@/lib/utils";
import { Globe2, Search, Compass, ShieldAlert, CheckCircle2, AlertTriangle } from "lucide-react";
import { toast } from "sonner";

export const Route = createFileRoute("/threat-intelligence")({
  head: () => ({
    meta: [
      { title: "Threat Intelligence & Compromised Indicators — ROPUS" },
      {
        name: "description",
        content:
          "Global indicator registry: IPs, datacenter proxy ASNs, headless emulator fingerprints, mule payout accounts and compromised BINs with feed provenance and confidence scores.",
      },
      { property: "og:title", content: "Threat Intelligence — ROPUS" },
      {
        property: "og:description",
        content: "Indicators of compromise feeding the risk score with provenance and confidence.",
      },
      { property: "og:type", content: "website" },
      { name: "twitter:card", content: "summary" },
    ],
  }),
  component: ThreatIntelPage,
});

const classTone: Record<IndicatorRecord["classification"], string> = {
  MALICIOUS: "text-blocked font-bold",
  SUSPICIOUS: "text-amber-intel font-bold",
  BENIGN: "text-muted-foreground",
};

const feeds = [
  {
    name: "internal-graph",
    records: "1,204,884",
    refreshed: "2026-08-28 02:40Z",
    state: "HEALTHY",
  },
  { name: "consortium-fraud", records: "88,412", refreshed: "2026-08-28 01:00Z", state: "HEALTHY" },
  { name: "proxy-registry", records: "412,009", refreshed: "2026-08-28 00:15Z", state: "HEALTHY" },
  {
    name: "internal-device",
    records: "9,884,201",
    refreshed: "2026-08-28 02:41Z",
    state: "HEALTHY",
  },
  { name: "issuer-registry", records: "301,884", refreshed: "2026-08-27 18:00Z", state: "HEALTHY" },
];

function ThreatIntelPage() {
  const [filterClass, setFilterClass] = useState<string>("ALL");
  const [search, setSearch] = useState<string>("");

  const filtered = useMemo(() => {
    return indicators.filter((i) => {
      const matchClass = filterClass === "ALL" || i.classification === filterClass;
      const matchSearch =
        search === "" ||
        i.value.toLowerCase().includes(search.toLowerCase()) ||
        i.type.toLowerCase().includes(search.toLowerCase()) ||
        i.feed.toLowerCase().includes(search.toLowerCase());
      return matchClass && matchSearch;
    });
  }, [filterClass, search]);

  const handleSimulateHaversine = () => {
    toast.error("Simulated Anomaly: 36,250 km/h Implied Velocity", {
      description: "Triggered RULE_IMPOSSIBLE_TRAVEL_SPEED (+0.21 risk score attribution).",
    });
  };

  return (
    <Page>
      <PageHead
        title="Threat Intelligence &amp; Global Indicators"
        subtitle="Live indicators contributed by internal graph analysis, syndicate consortiums, and proxy registries. Indicators add calibrated score; they never issue a verdict on their own."
      />

      <MetricStrip
        items={[
          { label: "Active Indicators", value: "1,986,506", sub: "across 5 feeds" },
          {
            label: "Malicious Matches",
            value: String(indicators.filter((i) => i.classification === "MALICIOUS").length),
            sub: "matched in last 24h",
            tone: "text-blocked",
          },
          {
            label: "Total Hits (24h)",
            value: indicators.reduce((s, i) => s + i.hits24h, 0).toLocaleString(),
            sub: "active evaluation matches",
            tone: "text-authoritative",
          },
          { label: "Feeds Healthy", value: "5 / 5", sub: "all registries current", tone: "text-authoritative" },
          { label: "Score Ceiling", value: "+0.14", sub: "max threat-intel weight", tone: "text-navy" },
        ]}
      />

      {/* Spherical Haversine Anomaly Simulator Box */}
      <div className="mt-6 border border-border bg-card p-4 shadow-xs flex flex-wrap items-center justify-between gap-4 font-mono text-[11.5px]">
        <div className="flex items-center gap-3">
          <Compass className="size-5 text-navy shrink-0" />
          <div>
            <div className="font-bold text-foreground text-[12px] font-sans">
              Spherical Haversine Speed Anomaly Simulator
            </div>
            <div className="text-muted-foreground text-[11px] font-sans">
              Test origin: Bengaluru, KA (12.97°N) → Limassol Proxy (34.70°N) · 7,250 km in 12 min (36,250 km/h)
            </div>
          </div>
        </div>

        <button
          type="button"
          onClick={handleSimulateHaversine}
          className="border border-navy bg-navy px-3 py-1 text-white font-bold text-[11px] hover:bg-navy/90 cursor-pointer shadow-2xs"
        >
          Test Travel Anomaly
        </button>
      </div>

      <div className="mt-6 grid gap-8 xl:grid-cols-[minmax(0,1fr)_340px] xl:gap-10">
        {/* Indicators Table */}
        <section className="min-w-0 space-y-4">
          <div className="flex flex-wrap items-center justify-between gap-3 border-b border-border pb-3">
            <SectionHead
              title="MATCHED COMPROMISED INDICATORS"
              meta={`${filtered.length} of ${indicators.length} records`}
            />

            <div className="flex items-center gap-2 font-mono text-[10.5px]">
              {["ALL", "MALICIOUS", "SUSPICIOUS"].map((c) => (
                <button
                  key={c}
                  type="button"
                  onClick={() => setFilterClass(c)}
                  className={cn(
                    "px-2.5 py-0.5 border cursor-pointer uppercase font-semibold",
                    filterClass === c
                      ? "border-navy bg-navy text-white font-bold"
                      : "border-border bg-surface text-muted-foreground hover:text-foreground",
                  )}
                >
                  {c}
                </button>
              ))}
              <input
                type="text"
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                placeholder="Search indicator, type, feed..."
                className="border border-border bg-surface px-2 py-0.5 text-[11px] font-mono text-foreground outline-none"
              />
            </div>
          </div>

          <div className="border border-border bg-card shadow-xs overflow-x-auto">
            <DataGrid
              columns={[
                { key: "value", label: "Indicator" },
                { key: "type", label: "Type" },
                { key: "class", label: "Classification" },
                { key: "feed", label: "Feed Source" },
                { key: "conf", label: "Confidence", align: "right" },
                { key: "hits", label: "Hits (24h)", align: "right" },
                { key: "first", label: "First Seen" },
                { key: "last", label: "Last Seen" },
              ]}
              rows={filtered.map((i) => ({
                id: i.value,
                cells: [
                  <Mono key="val" className="font-bold text-navy text-[11.5px]">{i.value}</Mono>,
                  <span key="type" className="text-muted-foreground font-sans text-[11.5px] font-medium">{i.type}</span>,
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
                  <Mono key="conf" className="font-bold text-foreground">{i.confidence.toFixed(2)}</Mono>,
                  <Mono key="hits" className="text-muted-foreground font-semibold">{i.hits24h.toLocaleString()}</Mono>,
                  <Mono key="first" className="text-muted-foreground text-[10.5px]">{i.firstSeen}</Mono>,
                  <Mono key="last" className="text-muted-foreground text-[10.5px]">{i.lastSeen}</Mono>,
                ],
              }))}
            />
          </div>
        </section>

        {/* Right Column: Active Feed Registries */}
        <aside className="min-w-0 space-y-4">
          <SectionHead title="FEED PROVENANCE REGISTRIES" meta="5 feeds" />
          <div className="border border-border bg-card p-4 shadow-xs space-y-2">
            <table className="w-full text-[11.5px] font-mono border-collapse">
              <thead>
                <tr className="border-b border-border text-[9px] uppercase font-bold text-muted-foreground text-left">
                  <th className="pb-2">Feed</th>
                  <th className="pb-2 text-right">Records</th>
                  <th className="pb-2 text-right">Status</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border/50">
                {feeds.map((f) => (
                  <tr key={f.name}>
                    <td className="py-2.5 font-bold text-foreground">
                      <Mono className="text-[11.5px]">{f.name}</Mono>
                      <div className="text-[9.5px] text-muted-foreground font-sans">{f.refreshed}</div>
                    </td>
                    <td className="py-2.5 text-right">
                      <Mono className="text-muted-foreground font-semibold">{f.records}</Mono>
                    </td>
                    <td className="py-2.5 text-right text-[10px] font-bold tracking-[0.06em] uppercase text-authoritative">
                      {f.state}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          <div className="border border-border border-l-2 border-l-navy bg-surface p-3 text-[11.5px] text-muted-foreground font-sans leading-relaxed">
            <strong className="text-foreground">Oracle Defense Boundary:</strong> Threat intelligence feeds contribute continuous additive risk weights during synchronous scoring. Raw indicators and reason codes are strictly restricted to authenticated analyst sessions.
          </div>
        </aside>
      </div>
    </Page>
  );
}
