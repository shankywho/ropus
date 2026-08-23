import { createFileRoute } from "@tanstack/react-router";
import { Page, PageHead, SectionHead, TelemetryStrip } from "@/components/ropus/page";
import { Mono } from "@/components/ropus/core";
import { accessEvents, securityPosture, teamMembers } from "@/lib/ropus/platform-fixtures";
import { cn } from "@/lib/utils";

export const Route = createFileRoute("/security")({
  head: () => ({
    meta: [
      { title: "Security — ROPUS" },
      {
        name: "description",
        content:
          "Access controls, tenant security posture and the immutable audit trail for analyst and service actions.",
      },
      { property: "og:title", content: "Security — ROPUS" },
      {
        property: "og:description",
        content: "Posture, operators and the audit trail for this tenant.",
      },
      { property: "og:type", content: "website" },
      { name: "twitter:card", content: "summary" },
    ],
  }),
  component: SecurityPage,
});

const postureTone: Record<string, string> = {
  Enforced: "text-approve",
  Enabled: "text-approve",
  Streaming: "text-warning",
  Paused: "text-warning",
};

function ResultTag({ result }: { result: "ALLOWED" | "DENIED" }) {
  const denied = result === "DENIED";
  return (
    <span
      className={cn(
        "inline-flex items-center gap-1.5 rounded-[3px] border px-1.5 py-[1px] text-[10px] font-bold tracking-[0.07em] uppercase",
        denied
          ? "border-block/45 bg-block/12 text-block"
          : "border-approve/35 bg-approve/10 text-approve",
      )}
    >
      <span aria-hidden className="size-[5px] bg-current" />
      {result}
    </span>
  );
}

function SecurityPage() {
  const denied = accessEvents.filter((e) => e.result === "DENIED");

  return (
    <Page>
      <PageHead
        title="Security"
        subtitle="Authorization is enforced by the backend; this console reflects it. Every read of a decision payload and every rule change is recorded in the audit trail."
      />

      <TelemetryStrip
        className="mt-3"
        items={[
          {
            label: "Operators",
            value: String(teamMembers.length),
            sub: "including 1 service account",
          },
          { label: "Audit events", value: "44,120", sub: "trailing 30 days" },
          {
            label: "Denied actions",
            value: String(denied.length),
            sub: "last 24 hours",
            tone: denied.length ? "text-block" : "",
          },
          { label: "MFA coverage", value: "100%", sub: "WebAuthn enforced" },
          { label: "Audit export", value: "Paused", sub: "whk_siem_03", tone: "text-warning" },
        ]}
      />

      <div className="mt-4 border border-border border-l-2 border-l-warning bg-warning/[0.06] px-4 py-3">
        <div className="text-[10.5px] font-bold tracking-[0.08em] text-warning uppercase">
          Security attention
        </div>
        <p className="mt-1 text-[12.5px]">
          <span className="font-semibold">1 denied action in the last 24 hours.</span>{" "}
          <span className="text-muted-foreground">
            <Mono className="text-[12px]">t.novak</Mono> attempted to reveal{" "}
            <Mono className="text-[12px]">key_9f21ab</Mono> without the{" "}
            <Mono className="text-[12px]">api:admin</Mono> scope. Audit export to{" "}
            <Mono className="text-[12px]">whk_siem_03</Mono> is paused, so events are buffered
            locally.
          </span>
        </p>
      </div>

      <div className="mt-6 grid gap-8 xl:grid-cols-[minmax(0,1fr)_360px] xl:gap-10">
        <section className="min-w-0">
          <SectionHead title="Audit trail" meta="most recent first" />
          <div className="mt-1 overflow-x-auto">
            <table className="w-full min-w-[820px] border-collapse text-[12.5px]">
              <thead>
                <tr className="border-b border-border">
                  {["Timestamp", "Actor", "Action", "Target", "Source IP", "Result"].map((h) => (
                    <th
                      key={h}
                      scope="col"
                      className="py-1.5 pr-4 text-left text-[10.5px] font-semibold tracking-[0.08em] whitespace-nowrap text-muted-foreground uppercase last:pr-0"
                    >
                      {h}
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {accessEvents.map((e) => {
                  const bad = e.result === "DENIED";
                  return (
                    <tr
                      key={e.id}
                      className={cn(
                        "border-b border-border last:border-b-0 hover:bg-accent",
                        bad && "bg-block/[0.05]",
                      )}
                    >
                      <td className="relative py-2.5 pr-4 align-middle">
                        {bad && (
                          <span
                            aria-hidden
                            className="absolute top-0 bottom-0 -left-2 w-[2px] bg-block"
                          />
                        )}
                        <Mono className="text-[11.5px] text-muted-foreground">{e.at}</Mono>
                      </td>
                      <td className="py-2.5 pr-4 align-middle">
                        <span className="text-[12.5px] font-medium">{e.actor}</span>
                      </td>
                      <td className="py-2.5 pr-4 align-middle">
                        <Mono className="text-[12.5px] font-semibold">{e.action}</Mono>
                      </td>
                      <td className="py-2.5 pr-4 align-middle">
                        <Mono className="text-[11.5px] text-muted-foreground">{e.target}</Mono>
                      </td>
                      <td className="py-2.5 pr-4 align-middle">
                        <Mono className="text-[11.5px] text-muted-foreground">{e.ip}</Mono>
                      </td>
                      <td className="py-2.5 align-middle">
                        <ResultTag result={e.result} />
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>

          <div className="mt-8">
            <SectionHead title="Operators" meta={`${teamMembers.length} with tenant access`} />
            <div className="mt-1 overflow-x-auto">
              <table className="w-full min-w-[720px] border-collapse text-[12.5px]">
                <thead>
                  <tr className="border-b border-border">
                    {["User", "Role", "Scopes", "Last active"].map((h, i) => (
                      <th
                        key={h}
                        scope="col"
                        className={cn(
                          "py-1.5 pr-4 text-[10.5px] font-semibold tracking-[0.08em] whitespace-nowrap text-muted-foreground uppercase last:pr-0",
                          i === 3 ? "text-right" : "text-left",
                        )}
                      >
                        {h}
                      </th>
                    ))}
                  </tr>
                </thead>
                <tbody>
                  {teamMembers.map((m) => {
                    const service = m.role === "Service account";
                    return (
                      <tr
                        key={m.user}
                        className="border-b border-border last:border-b-0 hover:bg-accent"
                      >
                        <td className="py-2.5 pr-4 align-middle">
                          <div className="flex items-baseline gap-2">
                            <Mono className="text-[12.5px] font-semibold">{m.user}</Mono>
                            {service && (
                              <span className="rounded-[3px] border border-border bg-neutral-surface px-1 py-[1px] text-[9.5px] font-bold tracking-[0.07em] text-muted-foreground uppercase">
                                svc
                              </span>
                            )}
                          </div>
                          <div className="mt-0.5 text-[11px] text-muted-foreground">{m.name}</div>
                        </td>
                        <td className="py-2.5 pr-4 align-middle">
                          <span
                            className={cn(
                              "text-[11.5px] font-semibold",
                              service ? "text-muted-foreground" : "text-foreground",
                            )}
                          >
                            {m.role}
                          </span>
                        </td>
                        <td className="py-2.5 pr-4 align-middle">
                          <div className="flex flex-wrap gap-1">
                            {m.scopes.split(",").map((s) => (
                              <span
                                key={s}
                                className="rounded-[3px] border border-border bg-neutral-surface px-1.5 py-[1px] font-mono text-[10.5px] text-muted-foreground"
                              >
                                {s.trim()}
                              </span>
                            ))}
                          </div>
                        </td>
                        <td className="py-2.5 text-right align-middle">
                          <Mono className="text-[11.5px] text-muted-foreground">
                            {m.lastActive}
                          </Mono>
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>
          </div>
        </section>

        <aside className="min-w-0">
          <SectionHead title="Posture" meta="tenant controls" />
          <dl className="mt-1">
            {securityPosture.map((p) => {
              const state = p.control === "Audit export" ? "Paused" : p.state;
              const tone = postureTone[state] ?? "text-foreground";
              return (
                <div key={p.control} className="border-b border-border py-2.5 last:border-b-0">
                  <div className="flex items-baseline justify-between gap-4">
                    <dt className="flex items-baseline gap-2 text-[12.5px] font-medium">
                      <span
                        aria-hidden
                        className={cn(
                          "size-[6px] translate-y-[-1px]",
                          tone === "text-approve"
                            ? "bg-approve"
                            : tone === "text-warning"
                              ? "bg-warning"
                              : "bg-border-strong",
                        )}
                      />
                      {p.control}
                    </dt>
                    <dd className={cn("font-mono text-[12px] font-semibold tabular", tone)}>
                      {state}
                    </dd>
                  </div>
                  <dd className="mt-0.5 pl-[14px] text-[11.5px] text-muted-foreground">
                    {p.detail}
                  </dd>
                </div>
              );
            })}
          </dl>
        </aside>
      </div>
    </Page>
  );
}
