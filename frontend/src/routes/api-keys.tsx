import { createFileRoute } from "@tanstack/react-router";
import { DataGrid, MetricStrip, Page, PageHead, SectionHead } from "@/components/ropus/page";
import { Mono } from "@/components/ropus/core";
import { apiKeys, type ApiKeyRecord } from "@/lib/ropus/platform-fixtures";
import { cn } from "@/lib/utils";

export const Route = createFileRoute("/api-keys")({
  head: () => ({
    meta: [
      { title: "API Keys — ROPUS" },
      { name: "description", content: "Issued API keys, scopes, environments, rotation state and last use." },
      { property: "og:title", content: "API Keys — ROPUS" },
      { property: "og:description", content: "Key inventory with scopes, rotation state and last use." },
      { property: "og:type", content: "website" },
      { name: "twitter:card", content: "summary" },
    ],
  }),
  component: ApiKeysPage,
});

const stateTone: Record<ApiKeyRecord["state"], string> = {
  ACTIVE: "text-approve",
  ROTATING: "text-warning",
  REVOKED: "text-muted-foreground",
};

function ApiKeysPage() {
  return (
    <Page>
      <PageHead
        title="API keys"
        subtitle="Secrets are shown once at creation and never again. Rotation overlaps for 24 hours so a gateway can cut over without dropping traffic."
      />

      <MetricStrip
        items={[
          { label: "Active keys", value: String(apiKeys.filter((k) => k.state === "ACTIVE").length), sub: "production and sandbox" },
          { label: "Rotating", value: String(apiKeys.filter((k) => k.state === "ROTATING").length), sub: "overlap window open" },
          { label: "Revoked", value: String(apiKeys.filter((k) => k.state === "REVOKED").length), sub: "retained for audit" },
          { label: "Rotation policy", value: "90 days", sub: "enforced" },
          { label: "Oldest active", value: "343 days", sub: "key_9f21ab" },
        ]}
      />

      <div className="mt-6">
        <SectionHead title="Key inventory" meta={`${apiKeys.length} keys`} />
        <div className="mt-1">
          <DataGrid
            columns={[
              { key: "id", label: "Key" },
              { key: "label", label: "Label" },
              { key: "prefix", label: "Prefix" },
              { key: "env", label: "Environment" },
              { key: "scopes", label: "Scopes" },
              { key: "created", label: "Created" },
              { key: "used", label: "Last used" },
              { key: "state", label: "State" },
            ]}
            rows={apiKeys.map((k) => ({
              id: k.id,
              cells: [
                <Mono className="text-muted-foreground">{k.id}</Mono>,
                <span className="font-medium">{k.label}</span>,
                <Mono>{k.prefix}</Mono>,
                <span className={cn("text-[11px] font-semibold tracking-[0.06em]", k.environment === "PRODUCTION" ? "text-foreground" : "text-muted-foreground")}>
                  {k.environment}
                </span>,
                <Mono className="text-muted-foreground">{k.scopes.join(", ")}</Mono>,
                <Mono className="text-muted-foreground">{k.created}</Mono>,
                <Mono className="text-muted-foreground">{k.lastUsed}</Mono>,
                <span className={cn("text-[11px] font-semibold tracking-[0.06em]", stateTone[k.state])}>{k.state}</span>,
              ],
            }))}
          />
        </div>
        <p className="mt-3 text-[11.5px] text-muted-foreground">
          key_7c8de2 is mid-rotation; its replacement has served traffic since 2026-08-21 and the old key expires
          2026-08-23 16:00Z.
        </p>
      </div>
    </Page>
  );
}
