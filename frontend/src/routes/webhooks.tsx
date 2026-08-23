import { createFileRoute } from "@tanstack/react-router";
import { DataGrid, MetricStrip, Page, PageHead, SectionHead } from "@/components/ropus/page";
import { Mono } from "@/components/ropus/core";
import { webhookDeliveries, webhookEndpoints } from "@/lib/ropus/platform-fixtures";
import { cn } from "@/lib/utils";

export const Route = createFileRoute("/webhooks")({
  head: () => ({
    meta: [
      { title: "Webhooks — ROPUS" },
      {
        name: "description",
        content:
          "Webhook endpoints, subscribed events, delivery success rates and recent delivery attempts.",
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
  return (
    <Page>
      <PageHead
        title="Webhooks"
        subtitle="Asynchronous delivery of decision and case events. Deliveries retry with exponential backoff for six hours, then the endpoint is paused."
      />

      <MetricStrip
        items={[
          { label: "Endpoints", value: String(webhookEndpoints.length), sub: "1 paused" },
          { label: "Delivered", value: "48,112", sub: "last 24 hours" },
          { label: "Success rate", value: "99.62%", sub: "all endpoints" },
          { label: "p95 delivery", value: "148.2 ms", sub: "whk_ledger_01" },
          { label: "Failed", value: "14", sub: "awaiting retry" },
        ]}
      />

      <div className="mt-6">
        <SectionHead title="Endpoints" />
        <div className="mt-1">
          <DataGrid
            columns={[
              { key: "id", label: "Endpoint" },
              { key: "url", label: "URL" },
              { key: "events", label: "Events" },
              { key: "rate", label: "Success", align: "right" },
              { key: "p95", label: "p95", align: "right" },
              { key: "last", label: "Last delivery" },
              { key: "state", label: "State" },
            ]}
            rows={webhookEndpoints.map((e) => ({
              id: e.id,
              cells: [
                <Mono className="text-muted-foreground">{e.id}</Mono>,
                <Mono className="font-medium">{e.url}</Mono>,
                <span className="text-muted-foreground">{e.events.join(", ")}</span>,
                <Mono className={e.successRate < 0.99 ? "text-warning" : undefined}>
                  {(e.successRate * 100).toFixed(2)}%
                </Mono>,
                <Mono className="text-muted-foreground">{e.p95Ms.toFixed(1)}ms</Mono>,
                <Mono className="text-muted-foreground">{e.lastDelivery}</Mono>,
                <span
                  className={cn(
                    "text-[11px] font-semibold tracking-[0.06em]",
                    e.state === "ACTIVE" ? "text-approve" : "text-warning",
                  )}
                >
                  {e.state}
                </span>,
              ],
            }))}
          />
        </div>
      </div>

      <div className="mt-8">
        <SectionHead title="Recent deliveries" meta="most recent first" />
        <div className="mt-1">
          <DataGrid
            columns={[
              { key: "at", label: "Timestamp" },
              { key: "id", label: "Delivery" },
              { key: "endpoint", label: "Endpoint" },
              { key: "event", label: "Event" },
              { key: "attempt", label: "Attempt", align: "right" },
              { key: "status", label: "Status", align: "right" },
              { key: "latency", label: "Latency", align: "right" },
            ]}
            rows={[...webhookDeliveries]
              .sort((a, b) => b.at.localeCompare(a.at))
              .map((d) => ({
                id: d.id,
                cells: [
                  <Mono className="text-muted-foreground">{d.at}</Mono>,
                  <Mono>{d.id}</Mono>,
                  <Mono className="text-muted-foreground">{d.endpoint}</Mono>,
                  <span className="text-muted-foreground">{d.event}</span>,
                  <Mono className={d.attempt > 1 ? "text-warning" : "text-muted-foreground"}>
                    {d.attempt}
                  </Mono>,
                  <Mono className={d.status >= 400 ? "text-block" : "text-approve"}>
                    {d.status}
                  </Mono>,
                  <Mono className="text-muted-foreground">{(d.latencyMs / 1000).toFixed(2)}s</Mono>,
                ],
              }))}
          />
        </div>
      </div>
    </Page>
  );
}
