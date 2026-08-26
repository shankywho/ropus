import { Link, useRouterState } from "@tanstack/react-router";
import { useEffect, useState, type ReactNode } from "react";
import { cn } from "@/lib/utils";
import { useSession } from "@/lib/rbac";
import { isLiveBackend } from "@/lib/ropus/api";

const controlPlaneNav = [
  { label: "Overview", to: "/" },
  { label: "Fraud Graph", to: "/graph" },
  { label: "Demo", to: "/demo" },
];

const platformServicesNav = [
  { label: "Rules", to: "/rules" },
  { label: "Models", to: "/models" },
  { label: "Cases", to: "/cases" },
  { label: "Operations", to: "/operations" },
  { label: "Security", to: "/security" },
];

function SidebarBody({ onNavigate }: { onNavigate?: () => void }) {
  const session = useSession();
  const pathname = useRouterState({ select: (s) => s.location.pathname });

  return (
    <div className="flex h-full flex-col bg-sidebar text-sidebar-foreground">
      {/* Application Wordmark: 18px, Weight 800, Letter spacing .18em */}
      <div className="flex items-center gap-3 border-b border-sidebar-border px-4 py-3.5">
        <span
          aria-hidden
          className="grid size-6 place-items-center bg-sidebar-primary font-mono text-[11px] font-bold text-sidebar-primary-foreground"
        >
          R
        </span>
        <div className="leading-tight">
          <div className="text-[18px] font-extrabold tracking-[0.18em] text-white">ROPUS</div>
          <div className="font-mono text-[9.5px] tracking-[0.13em] text-sidebar-muted uppercase">
            Risk Control Plane
          </div>
        </div>
      </div>

      <div className="border-b border-sidebar-border px-4 py-2.5">
        <div className="font-mono text-[10px] tracking-[0.13em] text-sidebar-muted uppercase">
          WORKSPACE
        </div>
        <div className="mt-0.5 truncate text-[12.5px] font-semibold text-white">{session.organization}</div>
        <div className="mt-0.5 font-mono text-[10.5px] text-sidebar-muted">{session.tenantId}</div>
      </div>

      <nav aria-label="Primary" className="flex-1 overflow-y-auto py-2">
        {/* Control Plane Group */}
        <div className="mb-4">
          <div className="px-4 py-1.5">
            <span className="font-mono text-[9.5px] tracking-[0.13em] text-sidebar-muted uppercase">
              CONTROL PLANE
            </span>
          </div>
          <ul className="mt-1 space-y-0.5">
            {controlPlaneNav.map((item) => {
              const active = item.to === "/" ? pathname === "/" : pathname.startsWith(item.to);
              return (
                <li key={item.to}>
                  <Link
                    to={item.to}
                    onClick={onNavigate}
                    aria-current={active ? "page" : undefined}
                    className={cn(
                      "flex items-center justify-between border-l-2 px-4 py-[6px] text-[12px] font-sans transition-colors",
                      active
                        ? "border-l-sidebar-primary bg-sidebar-primary/12 font-semibold text-white"
                        : "border-l-transparent text-sidebar-muted hover:bg-sidebar-accent/50 hover:text-white",
                      item.to === "/demo" && !active && "text-sidebar-foreground font-medium",
                    )}
                  >
                    <span>{item.label}</span>
                    {item.to === "/demo" && (
                      <span className="font-mono text-[10px] text-sidebar-primary">●</span>
                    )}
                  </Link>
                </li>
              );
            })}
          </ul>
        </div>

        {/* Platform Services Group */}
        <div>
          <div className="px-4 py-1.5">
            <span className="font-mono text-[9.5px] tracking-[0.13em] text-sidebar-muted uppercase">
              PLATFORM SERVICES
            </span>
          </div>
          <ul className="mt-1 space-y-0.5">
            {platformServicesNav.map((item) => {
              const active = item.to === "/" ? pathname === "/" : pathname.startsWith(item.to);
              return (
                <li key={item.to}>
                  <Link
                    to={item.to}
                    onClick={onNavigate}
                    aria-current={active ? "page" : undefined}
                    className={cn(
                      "flex items-center border-l-2 px-4 py-[5.5px] text-[12px] font-sans transition-colors",
                      active
                        ? "border-l-sidebar-primary bg-sidebar-primary/12 font-semibold text-white"
                        : "border-l-transparent text-sidebar-muted hover:bg-sidebar-accent/50 hover:text-white",
                    )}
                  >
                    {item.label}
                  </Link>
                </li>
              );
            })}
          </ul>
        </div>
      </nav>

      <div className="border-t border-sidebar-border px-4 py-3">
        <div className="flex items-center justify-between gap-2">
          <div className="min-w-0">
            <div className="truncate text-[12.5px] font-semibold text-white">{session.user}</div>
            <div className="truncate font-mono text-[10.5px] text-sidebar-muted">{session.role}</div>
          </div>
          <span
            className={cn(
              "shrink-0 border px-1.5 py-0.5 font-mono text-[9.5px] font-medium tracking-[0.06em]",
              session.environment === "PRODUCTION"
                ? "border-destructive/60 bg-destructive/20 text-white"
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
    <span className="font-mono text-[11px] text-muted-foreground tabular" suppressHydrationWarning>
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
            className="absolute inset-0 bg-navy/60 backdrop-blur-xs"
            onClick={() => setOpen(false)}
          />
          <div className="absolute inset-y-0 left-0 w-[212px] shadow-xl">
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
              className="grid size-8 place-items-center border border-border bg-surface text-muted-foreground lg:hidden"
            >
              <span aria-hidden>≡</span>
            </button>
            <span className="hidden font-mono text-[11px] tracking-[0.06em] text-muted-foreground sm:inline">
              {session.organization} / {session.tenantId}
            </span>
            {isLiveBackend ? (
              <span
                title="Connected to authoritative Go API at localhost:8080"
                className="inline-flex items-center gap-1 border border-approve/40 bg-approve-surface px-1.5 py-0.5 font-mono text-[9.5px] font-medium tracking-[0.06em] text-approve uppercase"
              >
                <span aria-hidden className="size-1 rounded-full bg-approve" />
                LIVE BACKEND
              </span>
            ) : (
              <span
                title="Offline fixtures rendered locally"
                className="inline-flex items-center gap-1 border border-amber-intel/40 bg-shadow-intel-surface px-1.5 py-0.5 font-mono text-[9.5px] font-medium tracking-[0.06em] text-shadow-intel uppercase"
              >
                <span aria-hidden className="size-1 rounded-full bg-amber-intel" />
                DEMO DATA
              </span>
            )}
          </div>
          <div className="flex items-center gap-4">
            <span className="hidden items-center gap-1.5 font-mono text-[10px] tracking-[0.06em] text-shadow-intel md:inline-flex">
              <span aria-hidden className="size-1.5 rounded-full bg-amber-intel animate-pulse" /> 2 SERVICES DEGRADED
            </span>
            <UtcClock />
            <span className="hidden font-mono text-[11px] text-muted-foreground sm:inline">
              {session.user}
            </span>
          </div>
        </header>
        <main className="min-w-0 flex-1 paper-deck-grid">{children}</main>
      </div>
    </div>
  );
}
