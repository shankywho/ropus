import { createFileRoute } from "@tanstack/react-router";
import { Page, PageHead, SectionHead } from "@/components/ropus/page";
import { Mono } from "@/components/ropus/core";
import { tenantSettings } from "@/lib/ropus/platform-fixtures";
import { cn } from "@/lib/utils";

export const Route = createFileRoute("/settings")({
  head: () => ({
    meta: [
      { title: "Settings — ROPUS" },
      { name: "description", content: "Tenant configuration: decision thresholds, timeout behaviour, case automation and retention." },
      { property: "og:title", content: "Settings — ROPUS" },
      { property: "og:description", content: "Thresholds, fallback behaviour and retention for this tenant." },
      { property: "og:type", content: "website" },
      { name: "twitter:card", content: "summary" },
    ],
  }),
  component: SettingsPage,
});

const thresholds = [
  {
    verdict: "APPROVE",
    from: 0,
    to: 0.34,
    range: "0.00 – 0.34",
    effect: "Returned to the gateway, no case.",
    tone: "text-approve",
    bar: "bg-approve",
  },
  {
    verdict: "CHALLENGE",
    from: 0.35,
    to: 0.54,
    range: "0.35 – 0.54",
    effect: "Step-up authentication requested.",
    tone: "text-primary",
    bar: "bg-primary",
  },
  {
    verdict: "REVIEW",
    from: 0.55,
    to: 0.89,
    range: "0.55 – 0.89",
    effect: "Settles, case opened for retrospective review.",
    tone: "text-warning",
    bar: "bg-warning",
  },
  {
    verdict: "BLOCK",
    from: 0.9,
    to: 1,
    range: "0.90 – 1.00",
    effect: "Payment refused, P1 case opened.",
    tone: "text-block",
    bar: "bg-block",
  },
];

const settingGroups: Array<{ title: string; keys: string[] }> = [
  { title: "Identity", keys: ["Tenant", "Legal entity", "Default policy"] },
  { title: "Decision thresholds", keys: ["Block threshold", "Review threshold", "Challenge threshold"] },
  { title: "Runtime behaviour", keys: ["Decision timeout", "Fallback on timeout", "Case auto-open"] },
  { title: "Data", keys: ["Retention"] },
];

function SettingsPage() {
  const byKey = Object.fromEntries(tenantSettings.map((s) => [s.key, s.value]));

  return (
    <Page>
      <PageHead
        title="Settings"
        subtitle="Tenant configuration. Threshold changes take effect on the next evaluation and are recorded in the audit trail."
      />

      <div className="mt-5 grid gap-8 xl:grid-cols-[minmax(0,1fr)_minmax(0,1fr)] xl:gap-14">
        <section className="min-w-0">
          <SectionHead title="Tenant configuration" meta="read-only in this console" />
          <div className="mt-3 space-y-5">
            {settingGroups.map((g) => (
              <div key={g.title}>
                <div className="text-[10px] font-bold tracking-[0.1em] text-muted-foreground uppercase">{g.title}</div>
                <dl className="mt-1">
                  {g.keys.map((k) => (
                    <div
                      key={k}
                      className="flex items-baseline justify-between gap-6 border-b border-border py-2 last:border-b-0 hover:bg-accent"
                    >
                      <dt className="text-[12.5px] text-muted-foreground">{k}</dt>
                      <dd className="min-w-0 truncate">
                        <Mono className="font-semibold">{byKey[k]}</Mono>
                      </dd>
                    </div>
                  ))}
                </dl>
              </div>
            ))}
          </div>
        </section>

        <section className="min-w-0">
          <SectionHead title="Verdict thresholds" meta="pol_wire_outbound_v7" />

          <div className="mt-4">
            <div className="flex h-[10px] w-full overflow-hidden border border-border">
              {thresholds.map((t) => (
                <div
                  key={t.verdict}
                  className={cn("h-full", t.bar)}
                  style={{ width: `${(t.to - t.from + 0.01) * 100}%` }}
                  aria-hidden
                />
              ))}
            </div>
            <div className="relative mt-1 h-4">
              {[0.35, 0.55, 0.9].map((p) => (
                <span
                  key={p}
                  className="absolute -translate-x-1/2 font-mono text-[10.5px] font-bold tabular"
                  style={{ left: `${p * 100}%` }}
                >
                  {p.toFixed(2)}
                </span>
              ))}
              <span className="absolute left-0 font-mono text-[10.5px] text-muted-foreground">0.00</span>
              <span className="absolute right-0 font-mono text-[10.5px] text-muted-foreground">1.00</span>
            </div>
          </div>

          <dl className="mt-4">
            {thresholds.map((t) => (
              <div key={t.verdict} className="border-b border-border py-2.5 last:border-b-0 hover:bg-accent">
                <div className="flex items-baseline justify-between gap-4">
                  <dt className={cn("flex items-baseline gap-2 text-[11px] font-bold tracking-[0.08em] uppercase", t.tone)}>
                    <span aria-hidden className={cn("size-[6px] translate-y-[-1px]", t.bar)} />
                    {t.verdict}
                  </dt>
                  <dd>
                    <Mono className="font-semibold">{t.range}</Mono>
                  </dd>
                </div>
                <dd className="mt-0.5 pl-[14px] text-[11.5px] text-muted-foreground">{t.effect}</dd>
              </div>
            ))}
          </dl>

          <div className="mt-4 border border-border border-l-2 border-l-primary bg-neutral-surface px-3.5 py-2.5">
            <div className="text-[10.5px] font-bold tracking-[0.08em] text-muted-foreground uppercase">
              Human in the loop
            </div>
            <p className="mt-1 text-[12px] text-muted-foreground">
              Recommended actions are never executed automatically at any threshold. A human confirms every
              irreversible action.
            </p>
          </div>
        </section>
      </div>
    </Page>
  );
}
