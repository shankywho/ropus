import { createFileRoute } from "@tanstack/react-router";
import { useState } from "react";
import { DataGrid, MetricStrip, Page, PageHead, SectionHead } from "@/components/ropus/page";
import { Mono } from "@/components/ropus/core";
import { CodeBlock } from "@/components/ropus/primitives";
import { apiEndpoints, evaluateResponse, evaluateSample } from "@/lib/ropus/platform-fixtures";
import { cn } from "@/lib/utils";
import { Terminal, Copy, Check, Sparkles, Send } from "lucide-react";
import { toast } from "sonner";

export const Route = createFileRoute("/api")({
  head: () => ({
    meta: [
      { title: "Developer API & SDK Reference — ROPUS" },
      {
        name: "description",
        content:
          "Decision API reference, synchronous evaluation endpoint, Go/Python/Node SDK snippets, and sub-100ms SLA integration guides.",
      },
      { property: "og:title", content: "Developer API — ROPUS" },
      {
        property: "og:description",
        content: "Synchronous risk evaluation endpoints with live latency and SDK code generation.",
      },
      { property: "og:type", content: "website" },
      { name: "twitter:card", content: "summary" },
    ],
  }),
  component: ApiPage,
});

const PYTHON_SAMPLE = `import requests

url = "https://api.ropus.internal/v1/risk-evaluations"
headers = {
    "Authorization": "Bearer rop_live_9f8a84b12c77120e88419",
    "Content-Type": "application/json"
}
payload = {
    "transaction_id": "txn_order_88419",
    "customer_id": "cus_4471029",
    "amount": 145000000,
    "currency": "INR",
    "ip_address": "198.51.100.44",
    "device_fingerprint": "dev_emulator_linux_9f8a",
    "payment_method": {
        "type": "imps",
        "token": "payout_pa_77120"
    }
}

response = requests.post(url, json=payload, headers=headers, timeout=0.10)
decision = response.json()
print(f"Verdict: {decision['verdict']}, Score: {decision['risk_score']}")`;

const GO_SAMPLE = `package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type EvaluationRequest struct {
	TransactionID     string        \`json:"transaction_id"\`
	CustomerID        string        \`json:"customer_id"\`
	Amount            int64         \`json:"amount"\`
	Currency          string        \`json:"currency"\`
	IPAddress         string        \`json:"ip_address"\`
	DeviceFingerprint string        \`json:"device_fingerprint"\`
	PaymentMethod     PaymentMethod \`json:"payment_method"\`
}

type PaymentMethod struct {
	Type  string \`json:"type"\`
	Token string \`json:"token"\`
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	payload, _ := json.Marshal(EvaluationRequest{
		TransactionID:     "txn_order_88419",
		CustomerID:        "cus_4471029",
		Amount:            145000000,
		Currency:          "INR",
		IPAddress:         "198.51.100.44",
		DeviceFingerprint: "dev_emulator_linux_9f8a",
		PaymentMethod:     PaymentMethod{Type: "imps", Token: "payout_pa_77120"},
	})

	req, _ := http.NewRequestWithContext(ctx, "POST", "https://api.ropus.internal/v1/risk-evaluations", bytes.NewBuffer(payload))
	req.Header.Set("Authorization", "Bearer rop_live_9f8a84b12c77120e88419")
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("Evaluation failed (Degrading to conservative rule fallback): %v\n", err)
		return
	}
	defer resp.Body.Close()
	fmt.Printf("HTTP Status: %d\n", resp.StatusCode)
}`;

const TS_SAMPLE = `import { RopusClient } from "@ropus/sdk";

const ropus = new RopusClient({
  apiKey: process.env.ROPUS_API_KEY!,
  timeoutMs: 100, // Enforces <100ms SLA
});

const evaluation = await ropus.evaluations.create({
  transactionId: "txn_order_88419",
  customerId: "cus_4471029",
  amount: 145000000,
  currency: "INR",
  ipAddress: "198.51.100.44",
  deviceFingerprint: "dev_emulator_linux_9f8a",
  paymentMethod: {
    type: "imps",
    token: "payout_pa_77120",
  },
});

if (evaluation.verdict === "BLOCK") {
  // Reject outbound payment immediately
  console.log(\`Transaction blocked: \${evaluation.reasonCodes.join(", ")}\`);
}`;

function ApiPage() {
  const [activeLang, setActiveLang] = useState<"curl" | "python" | "go" | "typescript">("curl");
  const [copied, setCopied] = useState(false);

  const getCode = () => {
    switch (activeLang) {
      case "python":
        return PYTHON_SAMPLE;
      case "go":
        return GO_SAMPLE;
      case "typescript":
        return TS_SAMPLE;
      default:
        return evaluateSample;
    }
  };

  const handleCopy = () => {
    navigator.clipboard.writeText(getCode());
    setCopied(true);
    toast.success("Code snippet copied to clipboard");
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <Page>
      <PageHead
        title="Developer API & SDK Reference"
        subtitle="One synchronous endpoint carries the decision path (<100ms SLA); everything else is read or workflow. Latency and error rates below are measured on this tenant's traffic."
      />

      <MetricStrip
        items={[
          { label: "Calls (24h)", value: "1,345,142", sub: "production requests" },
          { label: "Decision p99", value: "62.4 ms", sub: "POST /v1/risk-evaluations", tone: "text-authoritative" },
          { label: "Error rate", value: "0.03%", sub: "5xx and timeouts", tone: "text-authoritative" },
          { label: "Timeout Budget", value: "100 ms", sub: "falls back to AST rules" },
          { label: "API version", value: "2026-08-01", sub: "pinned per key" },
        ]}
      />

      <div className="mt-6 grid gap-8 xl:grid-cols-[minmax(0,1fr)_480px] xl:gap-10">
        <section className="min-w-0 space-y-6">
          <div>
            <SectionHead title="Endpoints" meta={`${apiEndpoints.length} public APIs`} />
            <div className="mt-2 border border-border bg-card shadow-xs">
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
                      key="m"
                      className={cn(
                        "font-mono text-[11px] font-bold",
                        e.method === "POST" ? "text-primary" : "text-muted-foreground",
                      )}
                    >
                      {e.method}
                    </span>,
                    <Mono key="p" className="font-bold text-navy">{e.path}</Mono>,
                    <span key="s" className="text-muted-foreground font-sans text-[12px]">{e.summary}</span>,
                    <Mono key="lat" className={e.p99Ms > 250 ? "text-amber-intel font-bold" : "text-muted-foreground"}>
                      {e.p99Ms.toFixed(1)}ms
                    </Mono>,
                    <Mono key="c" className="text-muted-foreground font-medium">{e.calls24h.toLocaleString()}</Mono>,
                    <Mono key="err" className={e.errorRate > 0.01 ? "text-blocked font-bold" : "text-muted-foreground"}>
                      {(e.errorRate * 100).toFixed(2)}%
                    </Mono>,
                  ],
                }))}
              />
            </div>
          </div>

          <div>
            <SectionHead title="Integration &amp; SLA Guarantees" />
            <dl className="mt-2 divide-y divide-border border border-border bg-card p-4 shadow-xs text-[12px] font-mono">
              {[
                [
                  "Authentication",
                  "Bearer key, scoped per environment. Keys are pinned to one API version and rotatable with 24h overlap.",
                ],
                [
                  "Idempotency",
                  "Repeat a transaction_id within 24h to receive the original decision, not a duplicate evaluation.",
                ],
                ["Timeouts & Fallback", "If downstream ML times out (>50ms), engine returns conservative deterministic AST rule verdict (is_degraded: true)."],
                [
                  "Webhooks & CDC",
                  "decision.returned is emitted asynchronously via Debezium CDC and Redpanda; never gate fund movement on webhooks.",
                ],
                ["Rate limit", "2,500 requests per second per production key with automatic token-bucket burst capacity."],
              ].map(([k, v]) => (
                <div key={k} className="flex flex-col sm:flex-row gap-2 sm:gap-6 py-2.5 first:pt-0 last:pb-0">
                  <dt className="w-[160px] shrink-0 text-[10.5px] font-bold tracking-[0.06em] text-navy uppercase">
                    {k}
                  </dt>
                  <dd className="text-muted-foreground font-sans leading-relaxed">{v}</dd>
                </div>
              ))}
            </dl>
          </div>
        </section>

        {/* Right Column: Code Snippets & Response Explorer */}
        <aside className="min-w-0 space-y-4">
          <div className="border border-border bg-card p-4 shadow-xs space-y-3">
            <div className="flex flex-wrap items-center justify-between border-b border-border pb-2.5 gap-2">
              <div className="flex items-center gap-1.5 font-mono text-[11px] font-bold text-navy uppercase tracking-wider">
                <Terminal className="size-4" />
                <span>POST /v1/risk-evaluations</span>
              </div>
              <div className="flex items-center gap-1 font-mono text-[10px]">
                {(["curl", "python", "go", "typescript"] as const).map((lang) => (
                  <button
                    key={lang}
                    type="button"
                    onClick={() => setActiveLang(lang)}
                    className={cn(
                      "px-2 py-0.5 border cursor-pointer uppercase font-semibold",
                      activeLang === lang
                        ? "border-navy bg-navy text-white font-bold"
                        : "border-border bg-surface text-muted-foreground hover:text-foreground",
                    )}
                  >
                    {lang}
                  </button>
                ))}
              </div>
            </div>

            <div className="relative">
              <button
                type="button"
                onClick={handleCopy}
                className="absolute top-2 right-2 z-10 border border-border/80 bg-surface/90 hover:bg-secondary p-1.5 text-muted-foreground hover:text-foreground cursor-pointer shadow-2xs font-mono text-[10px] flex items-center gap-1"
              >
                {copied ? <Check className="size-3 text-authoritative" /> : <Copy className="size-3" />}
                <span>{copied ? "Copied" : "Copy"}</span>
              </button>
              <CodeBlock code={getCode()} language={activeLang === "curl" ? "bash" : activeLang} />
            </div>
          </div>

          <div className="border border-border bg-card p-4 shadow-xs space-y-2 font-mono text-[11px]">
            <div className="flex items-center justify-between border-b border-border pb-2">
              <span className="font-bold text-navy uppercase text-[10px] tracking-wider">
                HTTP 200 Synchronous Response (&lt;18.4ms)
              </span>
              <span className="border border-authoritative/40 bg-approve-surface px-1.5 py-0.2 text-[9px] text-authoritative font-bold">
                APPLICATION/JSON
              </span>
            </div>
            <CodeBlock code={evaluateResponse} language="json" />
            <p className="text-[11px] text-muted-foreground font-sans leading-relaxed pt-1">
              The synchronous payload carries the decision verdict, calibrated risk score, and reason codes. Detailed TreeSHAP attributions and 3-hop graph topologies are indexed asynchronously for analyst review.
            </p>
          </div>
        </aside>
      </div>
    </Page>
  );
}
