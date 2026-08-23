import { Link, useRouterState } from "@tanstack/react-router";
import { useEffect, useState, type ReactNode } from "react";
import { cn } from "@/lib/utils";
import { useSession } from "@/lib/rbac";
import { DATA_SOURCE } from "@/lib/ropus-data";

const nav: Array<{ group: string; items: Array<{ label: string; to: string }> }> = [
  {
    group: "Control plane",
    items: [
      { label: "Overview", to: "/" },
      { label: "Risk Decisions", to: "/decisions" },
      { label: "Fraud Graph", to: "/graph" },
      { label: "Cases", to: "/cases" },
    ],
  },
  {
    group: "Intelligence",
    items: [
      { label: "Investigations", to: "/investigations" },
      { label: "Threat Intelligence", to: "/threat-intelligence" },
    ],
  },
  {
    group: "Decisioning",
    items: [
      { label: "Rules", to: "/rules" },
      { label: "Models", to: "/models" },
    ],
  },
  {
    group: "Developer",
    items: [
      { label: "API", to: "/api" },
      { label: "API Keys", to: "/api-keys" },
      { label: "Webhooks", to: "/webhooks" },
    ],
  },
  {
    group: "Operations",
    items: [
      { label: "Operations", to: "/operations" },
      { label: "Security", to: "/security" },
      { label: "Settings", to: "/settings" },
    ],
  },
  {
    group: "Scratch pad",
    items: [{ label: "Demo", to: "/demo" }],
  },

];


function SidebarBody({ onNavigate }: { onNavigate?: () => void }) {
  const session = useSession();
  const pathname = useRouterState({ select: (s) => s.location.pathname });

  return (
    <div className="flex h-full flex-col bg-sidebar text-sidebar-foreground">
      <div className="flex items-center gap-2.5 border-b border-sidebar-border px-4 py-3.5">
        <span aria-hidden className="grid size-6 place-items-center bg-sidebar-primary font-mono text-[11px] font-bold text-sidebar-primary-foreground">
          R
        </span>
        <div className="leading-tight">
          <div className="text-[13px] font-bold tracking-tight">ROPUS</div>
          <div className="font-mono text-[10px] tracking-[0.08em] text-sidebar-muted uppercase">Risk control plane</div>
        </div>
      </div>

      <div className="border-b border-sidebar-border px-4 py-2.5">
        <div className="text-[10px] font-semibold tracking-[0.08em] text-sidebar-muted uppercase">Workspace</div>
        <div className="mt-0.5 truncate text-[13px] font-medium">{session.organization}</div>
        <div className="mt-0.5 font-mono text-[11px] text-sidebar-muted">{session.tenantId}</div>
      </div>

      <nav aria-label="Primary" className="flex-1 overflow-y-auto py-2">
        {nav.map((section) => (
          <div key={section.group} className="mb-0.5">
            <div className="bg-black px-4 py-1.5">
              <span className="text-[9.5px] font-semibold tracking-[0.12em] text-sidebar-muted/80 uppercase">
                {section.group}
              </span>
            </div>
            <ul>

              {section.items.map((item) => {
                const active = item.to === "/" ? pathname === "/" : pathname.startsWith(item.to);
                return (
                  <li key={item.to}>
                    <Link
                      to={item.to}
                      onClick={onNavigate}
                      aria-current={active ? "page" : undefined}
                      className={cn(
                        "flex items-center border-l-2 px-4 py-[4.5px] text-[12.5px] transition-colors",
                        active
                          ? "border-l-sidebar-primary bg-sidebar-accent font-semibold text-white"
                          : "border-l-transparent text-sidebar-foreground/55 hover:text-white",
                      )}
                    >
                      {item.label}
                    </Link>
                  </li>
                );
              })}
            </ul>
          </div>
        ))}
      </nav>

      <div className="border-t border-sidebar-border px-4 py-3">
        <div className="flex items-center justify-between gap-2">
          <div className="min-w-0">
            <div className="truncate text-[13px] font-medium">{session.user}</div>
            <div className="truncate text-[11px] text-sidebar-muted">{session.role}</div>
          </div>
          <span
            className={cn(
              "shrink-0 border px-1.5 py-0.5 font-mono text-[10px] font-bold tracking-[0.08em]",
              session.environment === "PRODUCTION"
                ? "border-block/60 bg-block/20 text-white"
                : "border-sidebar-border bg-sidebar-accent text-sidebar-muted",
            )}
          >
            {session.environment}
          </span>
        </div>
      </div>
    </div>
  );
}

function UtcClock() {
  const [now, setNow] = useState<string | null>(null);
  useEffect(() => {
    const tick = () => setNow(new Date().toISOString().slice(11, 19));
    tick();
    const id = setInterval(tick, 1000);
    return () => clearInterval(id);
  }, []);
  return (
    <span className="font-mono text-[11px] text-muted-foreground" suppressHydrationWarning>
      UTC {now ?? "--:--:--"}
    </span>
  );
}

export function AppShell({ children }: { children: ReactNode }) {
  const [open, setOpen] = useState(false);
  const session = useSession();

  return (
    <div className="min-h-screen w-full bg-background lg:grid lg:grid-cols-[212px_1fr]">
      <aside className="hidden border-r border-sidebar-border lg:sticky lg:top-0 lg:block lg:h-screen">
        <SidebarBody />
      </aside>

      {open && (
        <div className="fixed inset-0 z-50 lg:hidden">
          <button
            aria-label="Close navigation"
            className="absolute inset-0 bg-navy/60"
            onClick={() => setOpen(false)}
          />
          <div className="absolute inset-y-0 left-0 w-[212px] shadow-lg">
            <SidebarBody onNavigate={() => setOpen(false)} />
          </div>
        </div>
      )}

      <div className="flex min-w-0 flex-col">
        <header className="sticky top-0 z-30 flex h-11 items-center justify-between gap-4 border-b border-border bg-surface px-4 lg:px-10">
          <div className="flex min-w-0 items-center gap-3">
            <button
              type="button"
              onClick={() => setOpen(true)}
              aria-label="Open navigation"
              className="grid size-8 place-items-center border border-border text-muted-foreground lg:hidden"
            >
              <span aria-hidden>≡</span>
            </button>
            <span className="hidden font-mono text-[11px] tracking-[0.06em] text-muted-foreground sm:inline">
              {session.organization} / {session.tenantId}
            </span>
            <span
              title={DATA_SOURCE.note}
              className="border border-warning/40 bg-warning-surface px-1.5 py-0.5 text-[10px] font-semibold tracking-[0.08em] text-warning uppercase"
            >
              {DATA_SOURCE.label}
            </span>
          </div>
          <div className="flex items-center gap-4">
            <span className="hidden items-center gap-1.5 text-[11px] font-semibold tracking-[0.06em] text-review md:inline-flex">
              <span aria-hidden className="size-1.5 bg-review" /> 2 SERVICES DEGRADED
            </span>
            <UtcClock />
            <span className="hidden font-mono text-[11px] text-muted-foreground sm:inline">{session.user}</span>
          </div>
        </header>
        <main className="min-w-0 flex-1">{children}</main>
      </div>
    </div>
  );
}
