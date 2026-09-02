import { createFileRoute } from "@tanstack/react-router";
import { useState } from "react";
import { DataGrid, MetricStrip, Page, PageHead, SectionHead } from "@/components/ropus/page";
import { Mono } from "@/components/ropus/core";
import { webhookDeliveries, webhookEndpoints } from "@/lib/ropus/platform-fixtures";
import { cn } from "@/lib/utils";
import { Webhook, Plus, Send, CheckCircle2, AlertTriangle, RefreshCw } from "lucide-react";
import { toast } from "sonner";

export const Route = createFileRoute("/webhooks")({
  head: () => ({
    meta: [
      { title: "Webhooks & CDC Streaming — ROPUS" },
      {
        name: "description",
        content:
          "Asynchronous webhook delivery, Debezium CDC event routing, retry exponential backoff, and signature verification.",
      },
      { property: "og:title", content: "Webhooks — ROPUS" },
      {
        property: "og:description",
        content: "Endpoint health and recent delivery attempts for decision and case events.",
      },
      { property: "og:type", content: "website" },
      { name: "twitter:card", content: "summary" },
    ],
  }),
  component: WebhooksPage,
});

function WebhooksPage() {
  const [testingPing, setTestingPing] = useState<string | null>(null);

  const handleTestPing = (endpointId: string) => {
    setTestingPing(endpointId);
    setTimeout(() => {
      setTestingPing(null);
      toast.success(`Webhook test payload sent to ${endpointId}`, {
        description: "HTTP 200 OK received in 142ms. HMAC-SHA256 signature verified.",
      });
    }, 450);
  };

  return (
    <Page>
      <PageHead
        title="Webhooks &amp; CDC Event Streaming"
        subtitle="Asynchronous streaming of risk decisions and analyst case dispositions via Debezium CDC and Redpanda. Deliveries retry with exponential backoff for 6 hours."
      />

      <MetricStrip
        items={[
          {
            label: "Active Endpoints",
            value: String(webhookEndpoints.length),
            sub: "1 paused for maintenance",
          },
          {
            label: "Delivered (24h)",
            value: "48,112",
            sub: "events dispatched",
            tone: "text-authoritative",
          },
          {
            label: "Delivery Rate",
            value: "99.62%",
            sub: "across all endpoints",
            tone: "text-authoritative",
          },
          { label: "p95 Latency", value: "148.2 ms", sub: "whk_ledger_01" },
          {
            label: "Retry Queue",
            value: "14",
            sub: "exponential backoff active",
            tone: "text-amber-intel",
          },
        ]}
      />

      <div className="mt-6 space-y-6">
        <div>
          <SectionHead
            title="Configured Endpoints"
            meta={`${webhookEndpoints.length} registered`}
          />
          <div className="mt-2 border border-border bg-card shadow-xs overflow-x-auto">
            <DataGrid
              columns={[
                { key: "id", label: "Endpoint ID" },
                { key: "url", label: "Target URL" },
                { key: "events", label: "Subscribed Events" },
                { key: "rate", label: "Success Rate", align: "right" },
                { key: "p95", label: "p95 Latency", align: "right" },
                { key: "last", label: "Last Delivery" },
                { key: "state", label: "State" },
                { key: "action", label: "Test" },
              ]}
              rows={webhookEndpoints.map((e) => ({
                id: e.id,
                cells: [
                  <Mono key="id" className="font-bold text-navy text-[11.5px]">
                    {e.id}
                  </Mono>,
                  <Mono key="url" className="font-semibold text-foreground text-[12px]">
                    {e.url}
                  </Mono>,
                  <span key="ev" className="text-muted-foreground font-mono text-[10.5px]">
                    {e.events.join(", ")}
                  </span>,
                  <Mono
                    key="rt"
                    className={
                      e.successRate < 0.99
                        ? "text-amber-intel font-bold"
                        : "text-authoritative font-bold"
                    }
                  >
                    {(e.successRate * 100).toFixed(2)}%
                  </Mono>,
                  <Mono key="lat" className="text-muted-foreground font-semibold">
                    {e.p99Ms ? `${e.p99Ms.toFixed(1)}ms` : "148.2ms"}
                  </Mono>,
                  <Mono key="ls" className="text-muted-foreground">
                    {e.lastDelivery}
                  </Mono>,
                  <span
                    key="st"
                    className={cn(
                      "font-mono text-[10.5px] font-bold uppercase",
                      e.state === "ACTIVE" ? "text-authoritative" : "text-amber-intel",
                    )}
                  >
                    {e.state}
                  </span>,
                  <button
                    key="act"
                    type="button"
                    onClick={() => handleTestPing(e.id)}
                    disabled={testingPing === e.id}
                    className="border border-border bg-surface px-2.5 py-0.5 text-[10.5px] font-mono text-foreground hover:bg-secondary cursor-pointer flex items-center gap-1 shadow-2xs disabled:opacity-50"
                  >
                    <Send className={cn("size-2.5", testingPing === e.id && "animate-pulse")} />
                    <span>{testingPing === e.id ? "Pinging..." : "Test Ping"}</span>
                  </button>,
                ],
              }))}
            />
          </div>
        </div>

        <div>
          <SectionHead title="Recent Delivery Attempts" meta="most recent deliveries (FIFO)" />
          <div className="mt-2 border border-border bg-card shadow-xs overflow-x-auto">
            <DataGrid
              columns={[
                { key: "at", label: "Timestamp" },
                { key: "id", label: "Delivery ID" },
                { key: "endpoint", label: "Endpoint" },
                { key: "event", label: "Event Type" },
                { key: "attempt", label: "Attempt", align: "right" },
                { key: "status", label: "HTTP Status", align: "right" },
                { key: "latency", label: "Latency", align: "right" },
              ]}
              rows={[...webhookDeliveries]
                .sort((a, b) => b.at.localeCompare(a.at))
                .map((d) => ({
                  id: d.id,
                  cells: [
                    <Mono key="at" className="text-muted-foreground">
                      {d.at}
                    </Mono>,
                    <Mono key="id" className="font-bold text-navy">
                      {d.id}
                    </Mono>,
                    <Mono key="ep" className="text-muted-foreground">
                      {d.endpoint}
                    </Mono>,
                    <span key="ev" className="font-mono text-[11px] font-semibold text-foreground">
                      {d.event}
                    </span>,
                    <Mono
                      key="att"
                      className={
                        d.attempt > 1 ? "text-amber-intel font-bold" : "text-muted-foreground"
                      }
                    >
                      #{d.attempt}
                    </Mono>,
                    <Mono
                      key="st"
                      className={
                        d.status >= 400 ? "text-blocked font-bold" : "text-authoritative font-bold"
                      }
                    >
                      HTTP {d.status}
                    </Mono>,
                    <Mono key="lat" className="text-muted-foreground font-semibold">
                      {(d.latencyMs / 1000).toFixed(2)}s
                    </Mono>,
                  ],
                }))}
            />
          </div>
        </div>
      </div>
    </Page>
  );
}
