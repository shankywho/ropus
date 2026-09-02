import { createFileRoute } from "@tanstack/react-router";
import { useState } from "react";
import { DataGrid, MetricStrip, Page, PageHead, SectionHead } from "@/components/ropus/page";
import { Mono } from "@/components/ropus/core";
import { apiKeys, type ApiKeyRecord } from "@/lib/ropus/platform-fixtures";
import { cn } from "@/lib/utils";
import { Key, Plus, Copy, Check, ShieldAlert, RotateCw, Trash2 } from "lucide-react";
import { toast } from "sonner";

export const Route = createFileRoute("/api-keys")({
  head: () => ({
    meta: [
      { title: "API Keys & Credentials — ROPUS" },
      {
        name: "description",
        content:
          "Issued API keys, scopes, environments, 24-hour overlapping key rotation, and cryptographic audit proofs.",
      },
      { property: "og:title", content: "API Keys — ROPUS" },
      {
        property: "og:description",
        content: "Key inventory with scopes, rotation state and last use.",
      },
      { property: "og:type", content: "website" },
      { name: "twitter:card", content: "summary" },
    ],
  }),
  component: ApiKeysPage,
});

const stateTone: Record<ApiKeyRecord["state"], string> = {
  ACTIVE: "text-approve font-bold",
  ROTATING: "text-amber-intel font-bold",
  REVOKED: "text-muted-foreground",
};

function ApiKeysPage() {
  const [keysList, setKeysList] = useState<ApiKeyRecord[]>(apiKeys);
  const [createModalOpen, setCreateModalOpen] = useState(false);
  const [newKeyLabel, setNewKeyLabel] = useState("Production Ingress Wire API");
  const [newKeyEnv, setNewKeyEnv] = useState<"PRODUCTION" | "SANDBOX">("PRODUCTION");
  const [newGeneratedSecret, setNewGeneratedSecret] = useState<string | null>(null);
  const [copied, setCopied] = useState(false);

  const handleGenerateKey = () => {
    const rawSecret =
      `rop_${newKeyEnv === "PRODUCTION" ? "live" : "test"}_` +
      Array.from({ length: 24 }, () => Math.random().toString(36)[2]).join("");
    const newRecord: ApiKeyRecord = {
      id: `key_${Math.random().toString(36).substring(2, 8)}`,
      label: newKeyLabel,
      prefix: rawSecret.substring(0, 16) + "...",
      environment: newKeyEnv,
      scopes: ["evaluations:create", "decisions:read"],
      created: "Just now",
      lastUsed: "Never",
      state: "ACTIVE",
    };
    setKeysList([newRecord, ...keysList]);
    setNewGeneratedSecret(rawSecret);
    toast.success("API key generated successfully", {
      description: "Secret displayed once. Ensure it is stored securely in your secret manager.",
    });
  };

  const handleRotateKey = (keyId: string) => {
    toast.info(`Rotation initiated for ${keyId}`, {
      description: "24-hour overlapping grace window opened. Old key will expire in 24 hours.",
    });
  };

  return (
    <Page>
      <PageHead
        title="API Keys & Access Credentials"
        subtitle="Cryptographically scoped bearer keys. Secrets are shown once at creation and never again. Rotation overlaps for 24 hours so gateways cut over with zero downtime."
      />

      <MetricStrip
        items={[
          {
            label: "Active keys",
            value: String(keysList.filter((k) => k.state === "ACTIVE").length),
            sub: "production & sandbox",
            tone: "text-authoritative",
          },
          {
            label: "Rotating",
            value: String(keysList.filter((k) => k.state === "ROTATING").length),
            sub: "24h overlap window",
            tone: "text-amber-intel",
          },
          {
            label: "Revoked",
            value: String(keysList.filter((k) => k.state === "REVOKED").length),
            sub: "retained in audit ledger",
          },
          { label: "Rotation Policy", value: "90 days", sub: "automatic notification" },
          {
            label: "SLA Guarantee",
            value: "0ms Downtime",
            sub: "during key rotation",
            tone: "text-authoritative",
          },
        ]}
      />

      {/* Action Bar */}
      <div className="mt-6 flex flex-wrap items-center justify-between gap-4 border-b border-border pb-4">
        <div>
          <SectionHead title="Key Inventory" meta={`${keysList.length} total credentials`} />
        </div>
        <button
          type="button"
          onClick={() => {
            setNewGeneratedSecret(null);
            setCreateModalOpen(true);
          }}
          className="border border-navy bg-navy px-3.5 py-1.5 font-mono text-[11px] font-bold text-white hover:bg-navy/90 flex items-center gap-1.5 cursor-pointer shadow-xs"
        >
          <Plus className="size-3.5" />
          <span>+ Generate New API Key</span>
        </button>
      </div>

      {/* Modal / Generator Drawer */}
      {createModalOpen && (
        <div className="border border-navy bg-card p-5 shadow-lg space-y-4 font-mono text-[11.5px] animate-in fade-in">
          <div className="flex items-center justify-between border-b border-border pb-2">
            <span className="font-bold text-navy text-[12px] uppercase flex items-center gap-2">
              <Key className="size-4" /> Issue New API Key
            </span>
            <button
              type="button"
              onClick={() => setCreateModalOpen(false)}
              className="text-muted-foreground hover:text-foreground cursor-pointer text-[11px]"
            >
              ✕ Close
            </button>
          </div>

          {!newGeneratedSecret ? (
            <div className="space-y-3">
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                <div>
                  <label className="text-[9.5px] text-muted-foreground uppercase block font-bold tracking-wider">
                    Key Label
                  </label>
                  <input
                    type="text"
                    value={newKeyLabel}
                    onChange={(e) => setNewKeyLabel(e.target.value)}
                    className="mt-1 w-full border border-border bg-surface px-2.5 py-1 text-foreground font-mono text-[11.5px]"
                  />
                </div>
                <div>
                  <label className="text-[9.5px] text-muted-foreground uppercase block font-bold tracking-wider">
                    Environment
                  </label>
                  <select
                    value={newKeyEnv}
                    onChange={(e) => setNewKeyEnv(e.target.value as "PRODUCTION" | "SANDBOX")}
                    className="mt-1 w-full border border-border bg-surface px-2.5 py-1 text-foreground font-mono text-[11.5px]"
                  >
                    <option value="PRODUCTION">PRODUCTION (Authoritative)</option>
                    <option value="SANDBOX">SANDBOX (Test Traffic)</option>
                  </select>
                </div>
              </div>

              <div className="pt-2 flex justify-end">
                <button
                  type="button"
                  onClick={handleGenerateKey}
                  className="border border-navy bg-navy px-4 py-1.5 text-white font-bold hover:bg-navy/90 cursor-pointer shadow-xs"
                >
                  Generate Secret Key
                </button>
              </div>
            </div>
          ) : (
            <div className="space-y-3 border border-authoritative/40 bg-approve-surface p-4 shadow-2xs">
              <div className="text-authoritative font-bold text-[12px] flex items-center gap-1.5">
                <Check className="size-4" /> API Key Created! Store your secret now.
              </div>
              <div className="flex items-center justify-between border border-border bg-surface p-2.5">
                <Mono className="text-[12px] font-bold text-foreground select-all">
                  {newGeneratedSecret}
                </Mono>
                <button
                  type="button"
                  onClick={() => {
                    navigator.clipboard.writeText(newGeneratedSecret);
                    setCopied(true);
                    toast.success("Copied to clipboard");
                    setTimeout(() => setCopied(false), 2000);
                  }}
                  className="border border-border bg-surface px-2.5 py-1 text-[10.5px] hover:bg-secondary cursor-pointer flex items-center gap-1"
                >
                  {copied ? (
                    <Check className="size-3 text-authoritative" />
                  ) : (
                    <Copy className="size-3" />
                  )}
                  <span>{copied ? "Copied" : "Copy"}</span>
                </button>
              </div>
              <p className="text-[11px] text-muted-foreground font-sans">
                You will not be able to view this full secret key again.
              </p>
            </div>
          )}
        </div>
      )}

      {/* Keys Table */}
      <div className="mt-2 border border-border bg-card shadow-xs overflow-x-auto">
        <DataGrid
          columns={[
            { key: "id", label: "Key ID" },
            { key: "label", label: "Label" },
            { key: "prefix", label: "Key Prefix" },
            { key: "env", label: "Environment" },
            { key: "scopes", label: "Scopes" },
            { key: "created", label: "Created" },
            { key: "used", label: "Last Used" },
            { key: "state", label: "State" },
            { key: "action", label: "Rotate" },
          ]}
          rows={keysList.map((k) => ({
            id: k.id,
            cells: [
              <Mono key="id" className="font-bold text-navy text-[11.5px]">
                {k.id}
              </Mono>,
              <span key="label" className="font-sans font-semibold text-foreground text-[12px]">
                {k.label}
              </span>,
              <Mono key="pref" className="text-muted-foreground">
                {k.prefix}
              </Mono>,
              <span
                key="env"
                className={cn(
                  "font-mono text-[10px] font-bold tracking-[0.06em] uppercase",
                  k.environment === "PRODUCTION" ? "text-emerald-800" : "text-muted-foreground",
                )}
              >
                {k.environment}
              </span>,
              <Mono key="sc" className="text-muted-foreground text-[10.5px]">
                {k.scopes.join(", ")}
              </Mono>,
              <Mono key="cr" className="text-muted-foreground text-[10.5px]">
                {k.created}
              </Mono>,
              <Mono key="us" className="text-muted-foreground text-[10.5px]">
                {k.lastUsed}
              </Mono>,
              <span
                key="st"
                className={cn(
                  "font-mono text-[10.5px] font-semibold tracking-[0.06em] uppercase",
                  stateTone[k.state],
                )}
              >
                {k.state}
              </span>,
              <button
                key="act"
                type="button"
                onClick={() => handleRotateKey(k.id)}
                className="border border-border bg-surface px-2 py-0.5 text-[10.5px] font-mono text-muted-foreground hover:text-foreground hover:bg-secondary cursor-pointer flex items-center gap-1 shadow-2xs"
              >
                <RotateCw className="size-2.5" />
                <span>Rotate</span>
              </button>,
            ],
          }))}
        />
      </div>

      <p className="mt-3 text-[11.5px] font-sans text-muted-foreground leading-relaxed">
        Key <Mono className="text-foreground font-bold">key_7c8de2</Mono> is currently in rotation.
        Its replacement has served traffic since 2026-08-21 and the old key will expire
        automatically on 2026-08-29.
      </p>
    </Page>
  );
}
