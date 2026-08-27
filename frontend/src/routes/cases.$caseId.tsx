import { createFileRoute, Link } from "@tanstack/react-router";
import { useState } from "react";
import {
  CaseStatusTag,
  CaseTimeline,
  DemoTag,
  EvidenceList,
  Mono,
  PriorityTag,
  RiskFactorList,
  RiskScore,
  VerdictBadge,
} from "@/components/ropus/core";
import { Disclosure, CodeBlock, KeyValue } from "@/components/ropus/primitives";
import { useSuspenseQuery } from "@tanstack/react-query";
import { caseQuery, decisionQuery } from "@/lib/ropus/api";
import type { CaseEvent, CaseRecord } from "@/lib/ropus/contracts";
import { cn } from "@/lib/utils";
import {
  ShieldAlert,
  User,
  Smartphone,
  Globe2,
  Building2,
  FileCheck2,
  Clock,
  Sparkles,
  Lock,
  ArrowRight,
  AlertOctagon,
  CheckCircle2,
  UserX,
} from "lucide-react";
import { toast } from "sonner";

export const Route = createFileRoute("/cases/$caseId")({
  head: () => ({
    meta: [
      { title: "Tripartite Forensic Evidence Dossier — ROPUS" },
      {
        name: "description",
        content:
          "Tripartite forensic evidence dossier, signal convergence breakdown, identity telemetry, and immutable audit disposition history.",
      },
      { property: "og:title", content: "Case Dossier — ROPUS" },
      {
        property: "og:description",
        content: "Forensic evidence dossier with identity, signal convergence, and immutable ledger disposition.",
      },
      { property: "og:type", content: "website" },
      { name: "twitter:card", content: "summary" },
    ],
  }),
  loader: ({ context, params }) => context.queryClient.ensureQueryData(caseQuery(params.caseId)),
  component: CaseDetail,
});

const money = (amount: number, currency: string) =>
  new Intl.NumberFormat("en-IN", { style: "currency", currency: currency || "INR" }).format(amount);

function CaseDetail() {
  const { caseId } = Route.useParams();
  const { data: record } = useSuspenseQuery(caseQuery(caseId));

  if (!record) {
    return (
      <div className="mx-auto w-full max-w-[1560px] px-5 py-12 lg:px-8">
        <h1 className="text-[19px] font-semibold tracking-tight">Case not found</h1>
        <p className="mt-2 text-[13px] text-muted-foreground">
          No case matches <Mono>{caseId}</Mono> in this tenant.
        </p>
        <Link to="/cases" className="mt-4 inline-block text-[13px] text-primary hover:underline">
          Back to queue
        </Link>
      </div>
    );
  }

  return <CaseBody record={record} />;
}

function CaseBody({ record }: { record: CaseRecord }) {
  const { data: decision } = useSuspenseQuery(decisionQuery(record.decisionId));
  const [extra, setExtra] = useState<CaseEvent[]>([]);
  const [resolution, setResolution] = useState<string | null>(null);

  const resolve = (label: string, text: string) => {
    setResolution(label);
    setExtra((prev) => [
      ...prev,
      {
        id: `ev_local_${prev.length}`,
        at: new Date().toISOString(),
        actor: "m.okafor",
        actorKind: "ANALYST",
        text,
      },
    ]);
    toast.success(`Case disposition recorded: ${label}`, {
      description: "Cryptographic SHA-256 block hash committed to immutable audit ledger.",
    });
  };

  return (
    <div className="mx-auto w-full max-w-[1560px] px-5 py-5 lg:px-8 space-y-5">
      {/* Breadcrumbs & Navigation */}
      <div className="flex items-center gap-2 font-mono text-[11px] text-muted-foreground">
        <Link to="/cases" className="hover:text-foreground">
          ← Back to Investigation Queue
        </Link>
        <span>/</span>
        <span className="text-foreground font-bold">{record.caseId}</span>
      </div>

      {/* Header */}
      <header className="flex flex-wrap items-start justify-between gap-6 border-b border-border pb-5">
        <div className="min-w-0 space-y-2">
          <div className="flex flex-wrap items-center gap-3">
            <h1 className="font-mono text-[24px] font-extrabold tracking-tight text-foreground">
              {record.caseId}
            </h1>
            <PriorityTag priority={record.priority} />
            <CaseStatusTag status={resolution ? ("CLOSED" as any) : record.status} />
            <VerdictBadge verdict={record.verdict} />
            <DemoTag />
          </div>
          <p className="font-sans text-[13px] text-muted-foreground">
            Forensic incident opened following automated BLOCK disposition on transaction <Mono className="text-navy font-bold">{decision.transactionId}</Mono>.
          </p>
        </div>

        <div className="flex items-center gap-6">
          <div className="text-right font-mono">
            <div className="text-[10px] text-muted-foreground uppercase">Transaction Amount</div>
            <div className="text-[20px] font-extrabold text-foreground">{money(record.amount, record.currency)}</div>
          </div>
          <RiskScore value={record.riskScore} size="hero" showBand />
        </div>
      </header>

      {/* ------------------------------------------------ TRIPARTITE EVIDENCE DOSSIER GRID */}
      <div className="grid gap-6 lg:grid-cols-3">
        {/* COLUMN 1: TRANSACTION & IDENTITY TELEMETRY */}
        <section className="border border-border bg-card p-4 space-y-4 shadow-xs font-mono text-[11.5px]">
          <div className="flex items-center justify-between border-b border-border pb-2">
            <span className="font-bold text-navy text-[11px] uppercase tracking-wider flex items-center gap-1.5">
              <User className="size-3.5" /> 1. Identity &amp; Session Telemetry
            </span>
          </div>

          {/* Customer Profile Box */}
          <div className="border border-border bg-surface p-3 space-y-2">
            <div className="flex justify-between">
              <span className="text-muted-foreground uppercase text-[9.5px]">Customer ID:</span>
              <span className="font-bold text-foreground">{record.customerId}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-muted-foreground uppercase text-[9.5px]">Tenure:</span>
              <span className="text-foreground">445 Days (14 Months)</span>
            </div>
            <div className="flex justify-between">
              <span className="text-muted-foreground uppercase text-[9.5px]">Historical Disputes:</span>
              <span className="text-authoritative font-bold">0 Disputes</span>
            </div>
          </div>

          {/* Device & Hardware Canvas */}
          <div className="border border-border bg-surface p-3 space-y-2">
            <div className="flex items-center gap-1.5 font-bold text-foreground text-[12px]">
              <Smartphone className="size-3.5 text-navy" />
              <span>Hardware Device Canvas</span>
            </div>
            <div className="space-y-1 text-[11px]">
              <div className="flex justify-between">
                <span className="text-muted-foreground">Fingerprint:</span>
                <Mono className="text-foreground">{String(decision.rawRequest?.["device_fingerprint"] ?? "dev_emulator_linux_9f8a")}</Mono>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">Platform:</span>
                <span className="text-foreground">Linux x86_64 / Headless Chrome</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">Canvas Entropy:</span>
                <span className="text-blocked font-bold">0.96 (Headless Emulator)</span>
              </div>
            </div>
          </div>

          {/* Network & GeoIP */}
          <div className="border border-border bg-surface p-3 space-y-2">
            <div className="flex items-center gap-1.5 font-bold text-foreground text-[12px]">
              <Globe2 className="size-3.5 text-navy" />
              <span>Network &amp; Geo Haversine</span>
            </div>
            <div className="space-y-1 text-[11px]">
              <div className="flex justify-between">
                <span className="text-muted-foreground">Observed IP:</span>
                <Mono className="text-foreground">{decision.threatIntel?.ip ?? "198.51.100.44"}</Mono>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">Origin:</span>
                <span className="text-foreground">Limassol, Cyprus (Datacenter Proxy)</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">Network/ASN:</span>
                <span className="text-foreground">{decision.threatIntel?.asn ?? "ASN 13335 (Cloudflare/Proxy)"}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">Velocity Hop:</span>
                <span className="text-blocked font-bold">36,250 km/h (Impossible Travel)</span>
              </div>
            </div>
          </div>
        </section>

        {/* COLUMN 2: SIGNAL CONVERGENCE (RULES + ML + GRAPH) */}
        <section className="border border-border bg-card p-4 space-y-4 shadow-xs font-mono text-[11.5px]">
          <div className="flex items-center justify-between border-b border-border pb-2">
            <span className="font-bold text-navy text-[11px] uppercase tracking-wider flex items-center gap-1.5">
              <Sparkles className="size-3.5" /> 2. Signal Convergence Dossier
            </span>
          </div>

          {/* Level 2: Triggered AST Rules */}
          <div className="border border-border bg-surface p-3 space-y-2">
            <div className="text-[10px] text-muted-foreground uppercase font-bold flex justify-between">
              <span>Triggered AST Policies</span>
              <span className="text-blocked font-bold">{decision.rules.length} Rules Fired</span>
            </div>
            <div className="space-y-1.5">
              {decision.rules.map((r) => (
                <div key={r.id} className="border border-border/60 bg-card p-2">
                  <div className="flex justify-between">
                    <span className="font-bold text-navy">{r.id}</span>
                    <span className="text-blocked font-bold text-[10px]">BLOCK</span>
                  </div>
                  <div className="text-muted-foreground text-[10.5px] font-sans mt-0.5">{r.name}</div>
                </div>
              ))}
            </div>
          </div>

          {/* Level 4: Calibrated ML Inference */}
          <div className="border border-border bg-surface p-3 space-y-2">
            <div className="text-[10px] text-muted-foreground uppercase font-bold flex justify-between">
              <span>ONNX ML Model Inference</span>
              <span className="text-authoritative font-bold">Beta Calibrated</span>
            </div>
            <div className="flex items-baseline justify-between">
              <span className="text-muted-foreground">Model:</span>
              <span className="font-bold text-foreground">{decision.inference.model}</span>
            </div>
            <div className="flex items-baseline justify-between">
              <span className="text-muted-foreground">Calibrated P(fraud):</span>
              <span className="text-[16px] font-extrabold text-blocked">{decision.inference.probability.toFixed(4)}</span>
            </div>
          </div>

          {/* Level 3: GraphSAGE Shadow Syndicate Cluster */}
          <div className="border border-border bg-surface p-3 space-y-2">
            <div className="text-[10px] text-muted-foreground uppercase font-bold flex justify-between">
              <span>GraphSAGE Shadow BFS</span>
              <span className="text-shadow-intel font-bold">0% Authority</span>
            </div>
            <div className="text-[11px] font-sans text-muted-foreground">
              3-hop BFS identified device canvas shared across <strong>14 synthetic accounts</strong> routing funds to cashout mule <strong>PA-77120</strong>.
            </div>
            <Link
              to="/graph"
              className="text-navy hover:underline text-[10.5px] font-bold block pt-1"
            >
              Explore 3-Hop Graph Topology →
            </Link>
          </div>
        </section>

        {/* COLUMN 3: IMMUTABLE AUDIT TIMELINE & ACTION CONSOLE */}
        <section className="border border-border bg-card p-4 space-y-4 shadow-xs font-mono text-[11.5px] flex flex-col justify-between">
          <div className="space-y-4">
            <div className="flex items-center justify-between border-b border-border pb-2">
              <span className="font-bold text-navy text-[11px] uppercase tracking-wider flex items-center gap-1.5">
                <Lock className="size-3.5" /> 3. Audit Ledger &amp; Console
              </span>
              <span className="text-[10px] text-authoritative font-bold">ACID Append-Only</span>
            </div>

            {/* Case Event Timeline */}
            <div className="space-y-2">
              <div className="text-[10px] text-muted-foreground uppercase font-bold">Investigation History</div>
              <CaseTimeline events={[...(record.timeline ?? []), ...extra]} />
            </div>
          </div>

          {/* Analyst Action Console */}
          <div className="border-t border-border pt-4 space-y-2">
            <div className="text-[10px] text-muted-foreground uppercase font-bold">Operator Dispositions</div>
            {!resolution ? (
              <div className="grid grid-cols-2 gap-2">
                <button
                  type="button"
                  onClick={() => resolve("CONFIRM_FRAUD_BLOCK", "Confirmed syndicate mule attack; permanent account lock & beneficiary freeze applied.")}
                  className="border border-blocked bg-blocked py-2 text-white font-bold hover:bg-blocked/90 cursor-pointer text-[11px] flex items-center justify-center gap-1"
                >
                  <AlertOctagon className="size-3" />
                  <span>Confirm Block</span>
                </button>
                <button
                  type="button"
                  onClick={() => resolve("ESCALATE_L2", "Escalated to Financial Intelligence Unit (FIU) / L2 Investigation team.")}
                  className="border border-amber-intel bg-shadow-intel-surface py-2 text-shadow-intel font-bold hover:bg-secondary cursor-pointer text-[11px] flex items-center justify-center gap-1"
                >
                  <ShieldAlert className="size-3" />
                  <span>Escalate L2</span>
                </button>
                <button
                  type="button"
                  onClick={() => resolve("APPROVE_OVERRIDE", "Manual analyst override; customer verified via out-of-band video KYC.")}
                  className="border border-border bg-surface py-2 text-foreground font-medium hover:bg-secondary cursor-pointer text-[11px] col-span-2 flex items-center justify-center gap-1"
                >
                  <CheckCircle2 className="size-3" />
                  <span>Approve with Out-of-Band KYC</span>
                </button>
              </div>
            ) : (
              <div className="border border-authoritative bg-approve-surface p-3 text-center space-y-1">
                <div className="text-authoritative font-bold text-[12px] flex items-center justify-center gap-1">
                  <CheckCircle2 className="size-3.5" />
                  <span>Case Resolved: {resolution}</span>
                </div>
                <div className="text-[10px] text-muted-foreground">
                  SHA-256 Ledger Block committed to immutable storage.
                </div>
              </div>
            )}
          </div>
        </section>
      </div>
    </div>
  );
}
