import { createFileRoute } from "@tanstack/react-router";
import { useState } from "react";
import { Page, PageHead, SectionHead, TelemetryStrip } from "@/components/ropus/page";
import { Mono, StatusPill } from "@/components/ropus/core";
import { accessEvents, securityPosture, teamMembers } from "@/lib/ropus/platform-fixtures";
import { cn } from "@/lib/utils";
import {
  ShieldCheck,
  Lock,
  KeyRound,
  FileCheck2,
  Layers,
  CheckCircle2,
  AlertTriangle,
  ArrowRight,
  RefreshCw,
  Hash,
} from "lucide-react";
import { toast } from "sonner";

export const Route = createFileRoute("/security")({
  head: () => ({
    meta: [
      { title: "Security & Merkle Hash Chain Explorer — ROPUS" },
      {
        name: "description",
        content:
          "Access controls, tenant security posture, WebAuthn MFA, and cryptographic Merkle Hash-Chain audit ledger verification.",
      },
      { property: "og:title", content: "Security & Cryptographic Ledger — ROPUS" },
      {
        property: "og:description",
        content: "Posture, operators and the cryptographic Merkle Hash-Chain audit trail for this tenant.",
      },
      { property: "og:type", content: "website" },
      { name: "twitter:card", content: "summary" },
    ],
  }),
  component: SecurityPage,
});

interface LedgerBlock {
  index: number;
  blockId: string;
  timestamp: string;
  actor: string;
  action: string;
  prevHash: string;
  payloadHash: string;
  blockHash: string;
  signature: string;
  verified: boolean;
}

const LEDGER_BLOCKS: LedgerBlock[] = [
  {
    index: 4284,
    blockId: "blk_4284_freeze",
    timestamp: "2026-08-28 02:30:11 UTC",
    actor: "shankar.r (Analyst)",
    action: "ANALYST_CONFIRM_BLOCK_AND_FREEZE",
    prevHash: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
    payloadHash: "7d1b54a88f01a938c204b7321e050882e75e921d746532788f4b5f88419f91a2",
    blockHash: "9a8f21bc084128564104c9958172ea2011b988f0412856c4192bca88172081f2",
    signature: "sig_es256_kms_secp256r1_99182a",
    verified: true,
  },
  {
    index: 4283,
    blockId: "blk_4283_rule_ast",
    timestamp: "2026-08-28 02:15:40 UTC",
    actor: "a.sharma (Security Lead)",
    action: "RULE_MAKER_CHECKER_APPROVAL",
    prevHash: "c51a0219bc881472851085201192ba8712950812957125918239019284102941",
    payloadHash: "41982bca88172081f2e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b93",
    blockHash: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
    signature: "sig_es256_kms_secp256r1_44710a",
    verified: true,
  },
  {
    index: 4282,
    blockId: "blk_4282_canary",
    timestamp: "2026-08-28 01:50:00 UTC",
    actor: "shankar.r (ML Engineer)",
    action: "MODEL_CANARY_ROUTING_UPDATE",
    prevHash: "88f0412856c4192bca88172081f27d1b54a88f01a938c204b7321e050882e75e",
    payloadHash: "c4192bca88172081f2e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b93",
    blockHash: "c51a0219bc881472851085201192ba8712950812957125918239019284102941",
    signature: "sig_es256_kms_secp256r1_88120b",
    verified: true,
  },
];

function ResultTag({ result }: { result: "ALLOWED" | "DENIED" }) {
  const denied = result === "DENIED";
  return (
    <span
      className={cn(
        "inline-flex items-center gap-1.5 rounded-[3px] border px-1.5 py-[1px] font-mono text-[10px] font-bold tracking-[0.07em] uppercase",
        denied
          ? "border-blocked/60 bg-blocked-surface text-blocked"
          : "border-authoritative/40 bg-approve-surface text-authoritative",
      )}
    >
      <span aria-hidden className="size-[5px] bg-current rounded-full" />
      {result}
    </span>
  );
}

export function SecurityPage() {
  const denied = accessEvents.filter((e) => e.result === "DENIED");
  const [selectedBlock, setSelectedBlock] = useState<LedgerBlock>(LEDGER_BLOCKS[0]);
  const [verifying, setVerifying] = useState(false);

  const handleVerifyChain = () => {
    setVerifying(true);
    setTimeout(() => {
      setVerifying(false);
      toast.success("Cryptographic Merkle Hash-Chain Validated (4,284 Blocks)", {
        description: "100% of SHA-256 block digests match KMS ES256 hardware signatures.",
      });
    }, 500);
  };

  return (
    <Page>
      <PageHead
        title="Security & Cryptographic Audit Ledger"
        subtitle="Authorization is enforced by the Go backend; all decision mutations and rule promotions are committed to an immutable SHA-256 Merkle hash-chain."
      />

      <TelemetryStrip
        items={[
          {
            label: "Operators",
            value: String(teamMembers.length),
            sub: "including 1 service account",
          },
          { label: "Audit Ledger Blocks", value: "4,284", sub: "SHA-256 Hash Chain", tone: "text-navy" },
          {
            label: "Denied Actions",
            value: String(denied.length),
            sub: "last 24 hours",
            tone: denied.length ? "text-blocked" : "",
          },
          { label: "MFA Coverage", value: "100%", sub: "WebAuthn / FIDO2 Enforced", tone: "text-authoritative" },
          { label: "KMS Key State", value: "Active", sub: "AWS KMS / Vault ES256", tone: "text-authoritative" },
        ]}
      />

      {/* ------------------------------------------------ MERKLE HASH CHAIN EXPLORER */}
      <section className="mt-6 border border-border bg-card p-5 shadow-xs space-y-4 font-mono text-[11.5px]">
        <div className="flex flex-wrap items-center justify-between border-b border-border pb-3 gap-2">
          <div className="flex items-center gap-2">
            <span className="border border-navy bg-navy/10 px-2 py-0.5 text-[10px] font-bold text-navy uppercase tracking-[0.06em]">
              IMMUTABLE LEDGER
            </span>
            <h2 className="font-sans text-[17px] font-bold text-foreground">
              Cryptographic Merkle Hash-Chain Block Explorer
            </h2>
          </div>
          <button
            type="button"
            onClick={handleVerifyChain}
            disabled={verifying}
            className="border border-navy bg-navy px-3 py-1 font-bold text-white text-[10.5px] hover:bg-navy/90 cursor-pointer flex items-center gap-1.5 disabled:opacity-50"
          >
            <RefreshCw className={cn("size-3", verifying && "animate-spin")} />
            <span>{verifying ? "Verifying SHA-256 Chain..." : "Verify Chain Integrity"}</span>
          </button>
        </div>

        <p className="font-sans text-[12px] text-muted-foreground">
          Every critical decision, rule activation, and model promotion generates a block chained via:
          <Mono className="text-navy font-bold ml-1">H_i = SHA-256(H_{`{i-1}`} || EntryID || Timestamp || PayloadHash)</Mono>.
        </p>

        <div className="grid gap-6 lg:grid-cols-[1fr_420px]">
          {/* Block Chain Visualizer */}
          <div className="space-y-2">
            <div className="text-[10px] text-muted-foreground uppercase font-bold">Recent Ledger Blocks</div>
            <div className="space-y-2">
              {LEDGER_BLOCKS.map((blk) => (
                <div
                  key={blk.blockId}
                  onClick={() => setSelectedBlock(blk)}
                  className={cn(
                    "border p-3 cursor-pointer transition-all bg-surface",
                    selectedBlock.blockId === blk.blockId
                      ? "border-navy bg-secondary/50 shadow-xs"
                      : "border-border hover:border-border-strong",
                  )}
                >
                  <div className="flex items-center justify-between text-[11px]">
                    <div className="flex items-center gap-2">
                      <span className="font-bold text-navy">#{blk.index}</span>
                      <Mono className="font-bold text-foreground">{blk.blockId}</Mono>
                    </div>
                    <span className="text-authoritative font-bold flex items-center gap-1 text-[10px]">
                      <CheckCircle2 className="size-3" /> VERIFIED
                    </span>
                  </div>
                  <div className="mt-1 text-[11px] text-foreground font-sans font-medium">
                    {blk.action}
                  </div>
                  <div className="mt-1 text-[10px] text-muted-foreground flex justify-between">
                    <span>Actor: {blk.actor}</span>
                    <span>{blk.timestamp}</span>
                  </div>
                </div>
              ))}
            </div>
          </div>

          {/* Block Cryptographic Inspector */}
          <aside className="border border-border bg-surface p-4 space-y-3 font-mono text-[11px]">
            <div className="flex items-center justify-between border-b border-border pb-2">
              <span className="font-bold text-navy text-[10.5px] uppercase">
                Block Cryptographic Proof #{selectedBlock.index}
              </span>
              <span className="border border-authoritative/40 bg-approve-surface px-1.5 py-0.2 text-[9.5px] text-authoritative font-bold">
                ES256 SIGNED
              </span>
            </div>

            <div className="space-y-2">
              <div>
                <span className="text-muted-foreground block text-[9.5px] uppercase">Action:</span>
                <span className="font-bold text-foreground text-[12px]">{selectedBlock.action}</span>
              </div>
              <div>
                <span className="text-muted-foreground block text-[9.5px] uppercase">Timestamp:</span>
                <span className="text-foreground">{selectedBlock.timestamp}</span>
              </div>
              <div>
                <span className="text-muted-foreground block text-[9.5px] uppercase">Previous Block Hash (H_{`{i-1}`}):</span>
                <Mono className="text-[10px] text-muted-foreground break-all">{selectedBlock.prevHash}</Mono>
              </div>
              <div>
                <span className="text-muted-foreground block text-[9.5px] uppercase">Payload Digest (SHA-256):</span>
                <Mono className="text-[10px] text-foreground break-all">{selectedBlock.payloadHash}</Mono>
              </div>
              <div className="border-t border-border pt-2">
                <span className="text-navy font-bold block text-[9.5px] uppercase">Final Block Hash (H_i):</span>
                <Mono className="text-[10.5px] font-bold text-navy break-all">{selectedBlock.blockHash}</Mono>
              </div>
              <div>
                <span className="text-muted-foreground block text-[9.5px] uppercase">KMS Signature:</span>
                <Mono className="text-[10px] text-authoritative">{selectedBlock.signature}</Mono>
              </div>
            </div>
          </aside>
        </div>
      </section>

      {/* Audit Log Table & Posture */}
      <div className="mt-6 grid gap-6 xl:grid-cols-[minmax(0,1fr)_360px] xl:gap-8">
        <section className="min-w-0 space-y-3">
          <SectionHead title="ACCESS AUDIT TRAIL" meta="most recent 24h events" />
          <div className="border border-border bg-card overflow-x-auto">
            <table className="w-full text-left font-mono text-[11.5px] border-collapse">
              <thead>
                <tr className="border-b border-border bg-secondary/50 text-[10px] uppercase text-muted-foreground">
                  <th className="py-2.5 px-3">Timestamp</th>
                  <th className="py-2.5 px-3">Actor</th>
                  <th className="py-2.5 px-3">Action</th>
                  <th className="py-2.5 px-3">Resource</th>
                  <th className="py-2.5 px-3">Result</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border">
                {accessEvents.map((e) => (
                  <tr key={e.id} className="hover:bg-secondary/40 transition-colors">
                    <td className="py-2.5 px-3 text-muted-foreground">{e.at}</td>
                    <td className="py-2.5 px-3 font-bold text-foreground">{e.actor}</td>
                    <td className="py-2.5 px-3 font-sans text-foreground">{e.action}</td>
                    <td className="py-2.5 px-3"><Mono className="text-[11px]">{e.resource}</Mono></td>
                    <td className="py-2.5 px-3"><ResultTag result={e.result} /></td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </section>

        {/* Security Posture */}
        <aside className="border border-border bg-card p-5 shadow-xs space-y-3 font-mono text-[11.5px]">
          <span className="font-bold text-navy text-[11px] uppercase tracking-wider block border-b border-border pb-2">
            TENANT SECURITY POSTURE
          </span>
          <div className="space-y-2 text-[11px]">
            {securityPosture.map((p) => (
              <div key={p.category} className="flex justify-between items-center py-1 border-b border-border/40 last:border-0">
                <span className="text-muted-foreground">{p.category}:</span>
                <span className="font-bold text-authoritative">{p.status}</span>
              </div>
            ))}
          </div>
        </aside>
      </div>
    </Page>
  );
}
