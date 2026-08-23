import { createFileRoute } from "@tanstack/react-router";
import { DataGrid, MetricStrip, Page, PageHead, SectionHead } from "@/components/ropus/page";
import { Mono } from "@/components/ropus/core";
import { CodeBlock } from "@/components/ropus/primitives";
import { apiEndpoints, evaluateResponse, evaluateSample } from "@/lib/ropus/platform-fixtures";
import { cn } from "@/lib/utils";

export const Route = createFileRoute("/api")({
  head: () => ({
    meta: [
      { title: "API — ROPUS" },
      {
        name: "description",
        content:
          "Decision API reference: synchronous evaluation, decision retrieval, graph traversal and case endpoints.",
      },
      { property: "og:title", content: "API — ROPUS" },
      {
        property: "og:description",
        content: "Synchronous risk evaluation endpoints with live latency and error rates.",
      },
      { property: "og:type", content: "website" },
      { name: "twitter:card", content: "summary" },
    ],
  }),
  component: ApiPage,
});

function ApiPage() {
  return (
    <Page>
      <PageHead
        title="API"
        subtitle="One synchronous endpoint carries the decision path; everything else is read or workflow. Latency and error rates below are measured on this tenant's traffic."
      />

      <MetricStrip
        items={[
          { label: "Calls", value: "1,345,142", sub: "last 24 hours" },
          { label: "Decision p99", value: "62.4 ms", sub: "/v1/risk/evaluate" },
          { label: "Error rate", value: "0.03%", sub: "5xx and timeouts" },
          { label: "Timeout policy", value: "120 ms", sub: "falls back to REVIEW" },
          { label: "API version", value: "2026-06-01", sub: "pinned per key" },
        ]}
      />

      <div className="mt-6 grid gap-8 xl:grid-cols-[minmax(0,1fr)_440px] xl:gap-10">
        <section className="min-w-0">
          <SectionHead title="Endpoints" meta={`${apiEndpoints.length} public`} />
          <div className="mt-1">
            <DataGrid
              columns={[
                { key: "method", label: "Method", width: "70px" },
                { key: "path", label: "Path" },
                { key: "summary", label: "Purpose" },
                { key: "p99", label: "p99", align: "right" },
                { key: "calls", label: "Calls 24h", align: "right" },
                { key: "err", label: "Errors", align: "right" },
              ]}
              rows={apiEndpoints.map((e) => ({
                id: e.path,
                cells: [
                  <span
                    className={cn(
                      "font-mono text-[11px] font-bold",
                      e.method === "POST" ? "text-primary" : "text-muted-foreground",
                    )}
                  >
                    {e.method}
                  </span>,
                  <Mono className="font-medium">{e.path}</Mono>,
                  <span className="text-muted-foreground">{e.summary}</span>,
                  <Mono className={e.p99Ms > 250 ? "text-warning" : "text-muted-foreground"}>
                    {e.p99Ms.toFixed(1)}ms
                  </Mono>,
                  <Mono className="text-muted-foreground">{e.calls24h.toLocaleString()}</Mono>,
                  <Mono className={e.errorRate > 0.01 ? "text-warning" : "text-muted-foreground"}>
                    {(e.errorRate * 100).toFixed(2)}%
                  </Mono>,
                ],
              }))}
            />
          </div>

          <div className="mt-8">
            <SectionHead title="Integration notes" />
            <dl className="mt-1 divide-y divide-border text-[12.5px]">
              {[
                [
                  "Authentication",
                  "Bearer key, scoped per environment. Keys are pinned to one API version.",
                ],
                [
                  "Idempotency",
                  "Repeat a transaction_id within 24h to receive the original decision, not a new evaluation.",
                ],
                ["Timeouts", "The gateway must fail open to REVIEW if ROPUS exceeds 120 ms."],
                [
                  "Webhooks",
                  "decision.returned is delivered after the synchronous response; never gate a payment on it.",
                ],
                ["Rate limit", "1,200 requests per second per production key; burst 2,000."],
              ].map(([k, v]) => (
                <div key={k} className="flex gap-6 py-2">
                  <dt className="w-[140px] shrink-0 text-[11px] font-semibold tracking-[0.06em] text-muted-foreground uppercase">
                    {k}
                  </dt>
                  <dd className="text-muted-foreground">{v}</dd>
                </div>
              ))}
            </dl>
          </div>
        </section>

        <aside className="min-w-0">
          <SectionHead title="POST /v1/risk/evaluate" />
          <div className="mt-2">
            <CodeBlock code={evaluateSample} language="bash" />
          </div>
          <div className="mt-4">
            <div className="text-[11px] font-semibold tracking-[0.08em] text-muted-foreground uppercase">
              200 response
            </div>
            <div className="mt-1.5">
              <CodeBlock code={evaluateResponse} language="json" />
            </div>
          </div>
          <p className="mt-3 text-[11.5px] text-muted-foreground">
            The response carries the verdict only. Factors, evidence and the graph neighbourhood are
            fetched from the decision endpoint by the console, not by the payment path.
          </p>
        </aside>
      </div>
    </Page>
  );
}
