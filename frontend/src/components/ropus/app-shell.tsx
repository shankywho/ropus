import { Link, useRouterState, useNavigate } from "@tanstack/react-router";
import { useEffect, useState, type ReactNode } from "react";
import { cn } from "@/lib/utils";
import { useSession } from "@/lib/rbac";
import { isLiveBackend } from "@/lib/ropus/api";
import { CommandPalette } from "./command-palette";
import {
  Search,
  Command as CommandIcon,
  Keyboard,
  LayoutDashboard,
  PlayCircle,
  Network,
  History,
  FolderLock,
  FileSearch,
  FileCode2,
  Cpu,
  Globe2,
  Activity,
  ShieldCheck,
  Terminal,
  Key,
  Webhook,
  Settings2,
  ShieldAlert,
} from "lucide-react";

interface NavItem {
  label: string;
  to: string;
  icon: React.ElementType;
  badge?: string;
  badgeTone?: "authoritative" | "shadow" | "blocked" | "neutral";
}

const coreEngineNav: NavItem[] = [
  { label: "Overview", to: "/", icon: LayoutDashboard },
  { label: "7-Stage Replay", to: "/demo", icon: PlayCircle, badge: "DEMO", badgeTone: "shadow" },
  { label: "Fraud Graph", to: "/graph", icon: Network, badge: "GNN", badgeTone: "shadow" },
  { label: "Decisions Log", to: "/decisions", icon: History },
  { label: "Case Queue", to: "/cases", icon: FolderLock, badge: "24h SLA", badgeTone: "blocked" },
  { label: "Investigations", to: "/investigations", icon: FileSearch },
];

const governanceNav: NavItem[] = [
  {
    label: "Rules Engine",
    to: "/rules",
    icon: FileCode2,
    badge: "AST",
    badgeTone: "authoritative",
  },
  { label: "Model Registry", to: "/models", icon: Cpu, badge: "BMR", badgeTone: "authoritative" },
  { label: "Threat Intel", to: "/threat-intelligence", icon: Globe2 },
  {
    label: "SRE Operations",
    to: "/operations",
    icon: Activity,
    badge: "99.99%",
    badgeTone: "authoritative",
  },
  { label: "Security & KMS", to: "/security", icon: ShieldCheck },
];

const developerNav: NavItem[] = [
  { label: "API Reference", to: "/api", icon: Terminal },
  { label: "API Keys", to: "/api-keys", icon: Key },
  { label: "Webhooks", to: "/webhooks", icon: Webhook },
  { label: "Settings", to: "/settings", icon: Settings2 },
];

function NavGroup({
  heading,
  items,
  pathname,
  onNavigate,
}: {
  heading: string;
  items: NavItem[];
  pathname: string;
  onNavigate?: () => void;
}) {
  return (
    <div className="mb-4">
      <div className="px-4 py-1">
        <span className="font-mono text-[9px] font-bold tracking-[0.14em] text-sidebar-muted uppercase">
          {heading}
        </span>
      </div>
      <ul className="mt-1 space-y-0.5">
        {items.map((item) => {
          const active = item.to === "/" ? pathname === "/" : pathname.startsWith(item.to);
          const Icon = item.icon;
          return (
            <li key={item.to}>
              <Link
                to={item.to}
                onClick={onNavigate}
                aria-current={active ? "page" : undefined}
                className={cn(
                  "flex items-center justify-between border-l-2 px-3.5 py-[5.5px] text-[12px] font-sans transition-all",
                  active
                    ? "border-l-sidebar-primary bg-sidebar-primary/12 font-semibold text-white"
                    : "border-l-transparent text-sidebar-muted hover:bg-sidebar-accent/50 hover:text-white",
                )}
              >
                <div className="flex items-center gap-2 min-w-0">
                  <Icon
                    className={cn(
                      "size-3.5 shrink-0 transition-colors",
                      active ? "text-sidebar-primary" : "text-sidebar-muted group-hover:text-white",
                    )}
                  />
                  <span className="truncate">{item.label}</span>
                </div>
                {item.badge && (
                  <span
                    className={cn(
                      "shrink-0 font-mono text-[8.5px] font-bold px-1.5 py-0.2 tracking-wider uppercase border",
                      item.badgeTone === "shadow"
                        ? "border-amber-intel/50 bg-amber-intel/15 text-sidebar-primary"
                        : item.badgeTone === "blocked"
                          ? "border-destructive/60 bg-destructive/20 text-rose-300"
                          : item.badgeTone === "authoritative"
                            ? "border-emerald-500/40 bg-emerald-500/15 text-emerald-300"
                            : "border-sidebar-border bg-sidebar-accent text-sidebar-muted",
                    )}
                  >
                    {item.badge}
                  </span>
                )}
              </Link>
            </li>
          );
        })}
      </ul>
    </div>
  );
}

function SidebarBody({ onNavigate }: { onNavigate?: () => void }) {
  const session = useSession();
  const pathname = useRouterState({ select: (s) => s.location.pathname });

  return (
    <div className="flex h-full flex-col bg-sidebar text-sidebar-foreground select-none">
      {/* Application Wordmark & YC Flagship Emblem */}
      <div className="flex items-center justify-between border-b border-sidebar-border px-4 py-3.5">
        <Link to="/" className="flex items-center gap-2.5 group">
          <span
            aria-hidden
            className="grid size-6 place-items-center bg-sidebar-primary font-mono text-[11px] font-extrabold text-sidebar-primary-foreground shadow-xs"
          >
            R
          </span>
          <div className="leading-tight">
            <div className="flex items-center gap-1.5">
              <span className="font-sans text-[15px] font-extrabold tracking-[0.14em] text-white group-hover:text-sidebar-primary transition-colors">
                ROPUS
              </span>
              <span className="border border-emerald-500/40 bg-emerald-500/10 px-1 py-0.2 font-mono text-[8px] font-bold text-emerald-400">
                v8.2
              </span>
            </div>
            <div className="font-mono text-[8.5px] font-semibold tracking-[0.13em] text-sidebar-muted uppercase">
              AI Risk Intelligence
            </div>
          </div>
        </Link>
      </div>

      {/* Tenant / Workspace Selector */}
      <div className="border-b border-sidebar-border px-4 py-2.5 bg-sidebar-accent/30">
        <div className="flex items-center justify-between">
          <span className="font-mono text-[8.5px] font-bold tracking-[0.14em] text-sidebar-muted uppercase">
            TENANT ENVIRONMENT
          </span>
          <span className="size-1.5 rounded-full bg-emerald-400 animate-pulse" />
        </div>
        <div className="mt-0.5 truncate text-[11.5px] font-bold text-white flex items-center gap-1">
          <span>{session.organization}</span>
        </div>
        <div className="mt-0.5 flex items-center justify-between font-mono text-[9.5px] text-sidebar-muted">
          <span>{session.tenantId}</span>
          <span className="text-emerald-400 font-semibold">&lt;100ms SLA</span>
        </div>
      </div>

      {/* Navigation Groups */}
      <nav aria-label="Primary" className="flex-1 overflow-y-auto py-2.5 scrollbar-thin">
        <NavGroup
          heading="CORE RISK ENGINE"
          items={coreEngineNav}
          pathname={pathname}
          onNavigate={onNavigate}
        />
        <NavGroup
          heading="INTELLIGENCE & GOVERNANCE"
          items={governanceNav}
          pathname={pathname}
          onNavigate={onNavigate}
        />
        <NavGroup
          heading="DEVELOPER PLATFORM"
          items={developerNav}
          pathname={pathname}
          onNavigate={onNavigate}
        />
      </nav>

      {/* User Session & Status */}
      <div className="border-t border-sidebar-border px-4 py-2.5 bg-sidebar-accent/20">
        <div className="flex items-center justify-between gap-2">
          <div className="min-w-0">
            <div className="truncate text-[11.5px] font-semibold text-white">{session.user}</div>
            <div className="truncate font-mono text-[9.5px] text-sidebar-muted">{session.role}</div>
          </div>
          <span
            className={cn(
              "shrink-0 border px-1.5 py-0.5 font-mono text-[8.5px] font-bold tracking-[0.06em] uppercase",
              session.environment === "PRODUCTION"
                ? "border-emerald-500/60 bg-emerald-500/20 text-emerald-300"
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
    <span
      className="font-mono text-[11px] text-muted-foreground tabular font-medium"
      suppressHydrationWarning
    >
      UTC {now ?? "--:--:--"}
    </span>
  );
}

export function AppShell({ children }: { children: ReactNode }) {
  const [open, setOpen] = useState(false);
  const [paletteOpen, setPaletteOpen] = useState(false);
  const [shortcutsModalOpen, setShortcutsModalOpen] = useState(false);
  const session = useSession();
  const navigate = useNavigate();

  // Vim-style key chord listeners: g then h/d/c/m/r/g/s/t/o/a
  useEffect(() => {
    let lastKey = "";
    let lastKeyTime = 0;

    const handleKeyDown = (e: KeyboardEvent) => {
      if (
        e.target instanceof HTMLInputElement ||
        e.target instanceof HTMLTextAreaElement ||
        (e.target as HTMLElement).isContentEditable
      ) {
        return;
      }

      const now = Date.now();
      if (e.key === "?" && !e.metaKey && !e.ctrlKey) {
        e.preventDefault();
        setShortcutsModalOpen((prev) => !prev);
        return;
      }

      if (e.key === "g") {
        lastKey = "g";
        lastKeyTime = now;
        return;
      }

      if (lastKey === "g" && now - lastKeyTime < 1000) {
        if (e.key === "h") {
          e.preventDefault();
          navigate({ to: "/" });
        } else if (e.key === "d") {
          e.preventDefault();
          navigate({ to: "/decisions" });
        } else if (e.key === "c") {
          e.preventDefault();
          navigate({ to: "/cases" });
        } else if (e.key === "m") {
          e.preventDefault();
          navigate({ to: "/models" });
        } else if (e.key === "r") {
          e.preventDefault();
          navigate({ to: "/rules" });
        } else if (e.key === "g") {
          e.preventDefault();
          navigate({ to: "/graph" });
        } else if (e.key === "s") {
          e.preventDefault();
          navigate({ to: "/security" });
        } else if (e.key === "t") {
          e.preventDefault();
          navigate({ to: "/threat-intelligence" });
        } else if (e.key === "o") {
          e.preventDefault();
          navigate({ to: "/operations" });
        } else if (e.key === "a") {
          e.preventDefault();
          navigate({ to: "/api" });
        }
        lastKey = "";
      }
    };

    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [navigate]);

  return (
    <div className="min-h-screen w-full bg-background lg:grid lg:grid-cols-[220px_1fr]">
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
          <div className="absolute inset-y-0 left-0 w-[220px] shadow-xl">
            <SidebarBody onNavigate={() => setOpen(false)} />
          </div>
        </div>
      )}

      <div className="flex min-w-0 flex-col">
        <header className="sticky top-0 z-30 flex h-11 items-center justify-between gap-4 border-b border-border bg-surface px-4 lg:px-8">
          <div className="flex min-w-0 items-center gap-3">
            <button
              type="button"
              onClick={() => setOpen(true)}
              aria-label="Open navigation"
              className="grid size-8 place-items-center border border-border bg-surface text-muted-foreground lg:hidden cursor-pointer"
            >
              <span aria-hidden>≡</span>
            </button>

            {/* Global Omnibar Search Trigger */}
            <button
              type="button"
              onClick={() => setPaletteOpen(true)}
              className="flex items-center gap-2 border border-border bg-secondary/40 hover:bg-secondary px-3 py-1 text-left text-muted-foreground hover:text-foreground transition-colors cursor-pointer"
            >
              <Search className="size-3.5 text-muted-foreground" />
              <span className="font-mono text-[11px] hidden sm:inline">Search control deck...</span>
              <kbd className="hidden sm:inline-flex items-center gap-0.5 border border-border bg-surface px-1 py-0.2 font-mono text-[9px] text-muted-foreground">
                <CommandIcon className="size-2.5" /> K
              </kbd>
            </button>

            {isLiveBackend ? (
              <span
                title="Connected to authoritative Go API at localhost:8080"
                className="inline-flex items-center gap-1.5 border border-approve/40 bg-approve-surface px-2 py-0.5 font-mono text-[9px] font-semibold tracking-[0.06em] text-approve uppercase"
              >
                <span aria-hidden className="size-1.5 rounded-full bg-approve animate-pulse" />
                LIVE GO BACKEND (:8080)
              </span>
            ) : (
              <span
                title="Local high-fidelity offline telemetry fixtures"
                className="inline-flex items-center gap-1.5 border border-amber-intel/40 bg-shadow-intel-surface px-2 py-0.5 font-mono text-[9.5px] font-semibold tracking-[0.06em] text-shadow-intel uppercase"
              >
                <span aria-hidden className="size-1.5 rounded-full bg-amber-intel" />
                OFFLINE FIXTURES
              </span>
            )}

            <div className="hidden xl:flex items-center gap-2 border-l border-border pl-3 font-mono text-[10px] text-muted-foreground">
              <span className="text-foreground font-semibold">&lt;100ms p95 SLA</span>
              <span>·</span>
              <span>AES-256-GCM KMS</span>
              <span>·</span>
              <span>SHA-256 Hash Chain</span>
            </div>
          </div>

          <div className="flex items-center gap-3 sm:gap-4">
            <button
              type="button"
              onClick={() => setShortcutsModalOpen(true)}
              title="Keyboard Shortcuts (?)"
              className="text-muted-foreground hover:text-foreground text-[11px] font-mono flex items-center gap-1 cursor-pointer hidden md:flex"
            >
              <Keyboard className="size-3.5" />
              <span>Shortcuts (?)</span>
            </button>
            <UtcClock />
            <span className="hidden font-mono text-[10.5px] text-muted-foreground sm:inline border-l border-border pl-3">
              {session.user}
            </span>
          </div>
        </header>

        <main className="min-w-0 flex-1 paper-deck-grid">{children}</main>
      </div>

      {/* Global Command Palette Component */}
      <CommandPalette open={paletteOpen} onOpenChange={setPaletteOpen} />

      {/* Shortcuts Cheat Sheet Modal */}
      {shortcutsModalOpen && (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-xs p-4 animate-in fade-in"
          onClick={() => setShortcutsModalOpen(false)}
        >
          <div
            className="w-full max-w-lg border border-border bg-card p-5 shadow-2xl space-y-4"
            onClick={(e) => e.stopPropagation()}
          >
            <div className="flex items-center justify-between border-b border-border pb-2">
              <span className="font-bold text-navy font-mono text-[11.5px] uppercase tracking-wider flex items-center gap-2">
                <Keyboard className="size-4" /> Operator Keyboard Shortcuts
              </span>
              <kbd className="border border-border bg-surface px-1.5 py-0.5 font-mono text-[9.5px] text-muted-foreground">
                ESC
              </kbd>
            </div>
            <div className="grid grid-cols-2 gap-3 font-mono text-[11.5px]">
              <div className="border border-border/60 bg-secondary/30 p-2.5 space-y-1.5">
                <div className="text-[9.5px] text-muted-foreground uppercase font-bold tracking-wider">
                  Navigation (G then Key)
                </div>
                <div className="flex justify-between">
                  <span>Overview</span>
                  <kbd className="text-navy font-bold">G H</kbd>
                </div>
                <div className="flex justify-between">
                  <span>Decisions</span>
                  <kbd className="text-navy font-bold">G D</kbd>
                </div>
                <div className="flex justify-between">
                  <span>Cases Queue</span>
                  <kbd className="text-navy font-bold">G C</kbd>
                </div>
                <div className="flex justify-between">
                  <span>Models Registry</span>
                  <kbd className="text-navy font-bold">G M</kbd>
                </div>
                <div className="flex justify-between">
                  <span>Rules AST</span>
                  <kbd className="text-navy font-bold">G R</kbd>
                </div>
                <div className="flex justify-between">
                  <span>Fraud Graph</span>
                  <kbd className="text-navy font-bold">G G</kbd>
                </div>
                <div className="flex justify-between">
                  <span>Threat Intel</span>
                  <kbd className="text-navy font-bold">G T</kbd>
                </div>
                <div className="flex justify-between">
                  <span>Operations</span>
                  <kbd className="text-navy font-bold">G O</kbd>
                </div>
                <div className="flex justify-between">
                  <span>Security KMS</span>
                  <kbd className="text-navy font-bold">G S</kbd>
                </div>
                <div className="flex justify-between">
                  <span>API Reference</span>
                  <kbd className="text-navy font-bold">G A</kbd>
                </div>
              </div>
              <div className="border border-border/60 bg-secondary/30 p-2.5 space-y-1.5">
                <div className="text-[9.5px] text-muted-foreground uppercase font-bold tracking-wider">
                  Global Omnibar &amp; Actions
                </div>
                <div className="flex justify-between">
                  <span>Open Omnibar</span>
                  <kbd className="text-navy font-bold">⌘ K / /</kbd>
                </div>
                <div className="flex justify-between">
                  <span>Shortcuts Help</span>
                  <kbd className="text-navy font-bold">?</kbd>
                </div>
                <div className="flex justify-between">
                  <span>Close Modal</span>
                  <kbd className="text-navy font-bold">ESC</kbd>
                </div>
                <div className="flex justify-between">
                  <span>Run Action</span>
                  <kbd className="text-navy font-bold">↵</kbd>
                </div>
              </div>
            </div>
            <div className="flex justify-end">
              <button
                type="button"
                onClick={() => setShortcutsModalOpen(false)}
                className="border border-navy bg-navy px-3 py-1 font-bold text-white text-[11px] hover:bg-navy/90 cursor-pointer font-mono"
              >
                Close
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
