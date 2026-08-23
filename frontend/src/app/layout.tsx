"use client";

import React, { useState, useEffect } from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import "./globals.css";
import {
  LayoutDashboard,
  ShieldAlert,
  GitGraph,
  FolderKanban,
  PlayCircle,
  FileCode2,
  Cpu,
  Terminal,
  Key,
  Webhook,
  Activity,
  ShieldCheck,
  Sliders,
  Radio,
  FileSearch,
  ChevronRight,
  Server,
  AlertTriangle,
} from "lucide-react";
import { operationsApi } from "@/api/operations";

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const pathname = usePathname();
  const [isPlatformDrawerOpen, setIsPlatformDrawerOpen] = useState(false);
  const [backendHealth, setBackendHealth] = useState<string>("CHECKING");
  const [isDegradedBannerOpen, setIsDegradedBannerOpen] = useState(false);

  useEffect(() => {
    let isMounted = true;
    const checkHealth = async () => {
      try {
        const rep = await operationsApi.getHealth();
        if (isMounted) {
          setBackendHealth(rep.overall_status || "HEALTHY");
          setIsDegradedBannerOpen(rep.overall_status === "DEGRADED" || rep.overall_status === "UNHEALTHY");
        }
      } catch {
        if (isMounted) {
          setBackendHealth("OFFLINE");
          setIsDegradedBannerOpen(true);
        }
      }
    };
    checkHealth();
    const interval = setInterval(checkHealth, 30000);
    return () => {
      isMounted = false;
      clearInterval(interval);
    };
  }, []);

  const navGroups = [
    {
      title: "Control Plane",
      items: [
        { name: "Overview", href: "/", icon: LayoutDashboard },
        { name: "Risk Decisions", href: "/transactions", icon: ShieldAlert },
        { name: "Fraud Graph", href: "/graph", icon: GitGraph },
        { name: "Cases", href: "/cases", icon: FolderKanban },
      ],
    },
    {
      title: "Intelligence",
      items: [
        { name: "Investigations", href: "/investigations", icon: FileSearch },
        { name: "Threat Intelligence", href: "/threat-intel", icon: Radio },
      ],
    },
    {
      title: "Decisioning",
      items: [
        { name: "Rules Engine", href: "/rules", icon: FileCode2 },
        { name: "Model Registry", href: "/models", icon: Cpu },
      ],
    },
    {
      title: "Developer",
      items: [
        { name: "API Reference", href: "/playground", icon: Terminal },
        { name: "API Keys", href: "/api-keys", icon: Key },
        { name: "Webhooks", href: "/webhooks", icon: Webhook },
      ],
    },
    {
      title: "Operations",
      items: [
        { name: "Cluster Health", href: "/operations", icon: Activity },
        { name: "Security & Audit", href: "/security", icon: ShieldCheck },
        { name: "Tenant Settings", href: "/settings", icon: Sliders },
      ],
    },
  ];

  return (
    <html lang="en">
      <head>
        <title>ROPUS — Risk Control Plane</title>
        <meta
          name="description"
          content="Production-grade AI fraud decisioning, graph intelligence & autonomous safety control plane"
        />
      </head>
      <body className="flex h-screen overflow-hidden bg-[#070e1c] text-[#f4f5f7]">
        {/* Sidebar */}
        <aside className="w-64 bg-[#011638] border-r border-[#0a2552] flex flex-col justify-between shrink-0 select-none shadow-md">
          <div className="overflow-y-auto max-h-[calc(100vh-140px)]">
            {/* Header Brand */}
            <div className="px-5 py-4 border-b border-[#0a2552] flex items-center justify-between sticky top-0 bg-[#011638] z-10">
              <Link href="/" className="flex items-center gap-2.5">
                <div className="w-6 h-6 bg-[#0d94fb] rounded-[4px] flex items-center justify-center font-mono text-xs font-black text-white shadow-xs">
                  R
                </div>
                <div>
                  <div className="flex items-center gap-1.5">
                    <span className="font-bold text-sm tracking-tight text-[#f4f5f7]">
                      ROPUS
                    </span>
                    <span className="text-[9px] font-mono text-[#0d94fb] bg-[#0d94fb15] px-1 py-0.2 rounded-[2px] border border-[#0d94fb33]">
                      CONTROL
                    </span>
                  </div>
                  <p className="text-[10px] text-[#97a0af] font-mono leading-none mt-0.5">
                    AI Risk Manager Mesh
                  </p>
                </div>
              </Link>
            </div>

            {/* Navigation Groups */}
            <nav className="p-3 space-y-4">
              {navGroups.map((grp) => (
                <div key={grp.title} className="space-y-1">
                  <div className="px-3 py-1 text-[10px] font-mono font-bold text-[#5e6c84] uppercase tracking-wider">
                    {grp.title}
                  </div>
                  {grp.items.map((item) => {
                    const isActive =
                      item.href === "/" ? pathname === "/" : pathname.startsWith(item.href);
                    const Icon = item.icon;

                    return (
                      <Link
                        key={item.href}
                        href={item.href}
                        className={`flex items-center gap-2.5 px-3 py-1.5 rounded-[4px] text-xs font-semibold transition-all ${
                          isActive
                            ? "bg-[#0d94fb] text-white shadow-xs"
                            : "text-[#97a0af] hover:text-[#f4f5f7] hover:bg-[#06204c]"
                        }`}
                      >
                        <Icon className="w-3.5 h-3.5 shrink-0" />
                        <span>{item.name}</span>
                      </Link>
                    );
                  })}
                </div>
              ))}

              {/* Visual Divider Before Demo */}
              <div className="pt-2 pb-1">
                <div className="h-px bg-[#0a2552] w-full" />
              </div>

              {/* Demo Walkthrough Item */}
              <div className="space-y-1">
                <div className="px-3 py-1 text-[10px] font-mono font-bold text-[#5e6c84] uppercase tracking-wider">
                  Simulation & Demo
                </div>
                <Link
                  href="/demo"
                  className={`flex items-center justify-between px-3 py-2 rounded-[4px] text-xs font-semibold transition-all ${
                    pathname.startsWith("/demo")
                      ? "bg-[#f59e0b] text-[#011638] font-bold shadow-xs"
                      : "text-[#f59e0b] hover:bg-[#f59e0b12]"
                  }`}
                >
                  <div className="flex items-center gap-2.5">
                    <PlayCircle className="w-4 h-4 shrink-0" />
                    <span>Demo Walkthrough</span>
                  </div>
                  <span
                    className={`text-[9px] font-mono px-1.5 py-0.2 rounded-[2px] uppercase ${
                      pathname.startsWith("/demo")
                        ? "bg-[#011638] text-[#f59e0b]"
                        : "bg-[#f59e0b22] text-[#f59e0b]"
                    }`}
                  >
                    7-Stage
                  </span>
                </Link>
              </div>
            </nav>
          </div>

          {/* Sidebar Footer: Multi-Tenancy / Auth Context & Platform Link */}
          <div className="p-3 border-t border-[#0a2552] bg-[#01112b] space-y-2 shrink-0">
            <button
              onClick={() => setIsPlatformDrawerOpen(!isPlatformDrawerOpen)}
              className="w-full px-2.5 py-1 text-[11px] font-mono text-[#97a0af] hover:text-[#f4f5f7] hover:bg-[#06204c] rounded-[4px] flex items-center justify-between transition-colors cursor-pointer"
            >
              <div className="flex items-center gap-1.5">
                <Server className="w-3.5 h-3.5 text-[#0d94fb]" />
                <span>Backend Port: 8080</span>
              </div>
              <ChevronRight
                className={`w-3 h-3 transition-transform ${
                  isPlatformDrawerOpen ? "rotate-90" : ""
                }`}
              />
            </button>

            {isPlatformDrawerOpen && (
              <div className="p-2.5 bg-[#011638] border border-[#0a2552] rounded-[4px] space-y-1.5 text-[10px] font-mono">
                <div className="flex justify-between text-[#97a0af]">
                  <span>API Service:</span>
                  <span className="text-[#0d94fb] font-semibold">http://localhost:8080</span>
                </div>
                <div className="flex justify-between text-[#97a0af]">
                  <span>ML Sidecar:</span>
                  <span className="text-[#04db7c]">http://localhost:8000</span>
                </div>
                <div className="flex justify-between text-[#97a0af]">
                  <span>RBAC Scope:</span>
                  <span className="text-[#f4f5f7]">ADMIN_OWNER</span>
                </div>
                <div className="flex justify-between text-[#97a0af]">
                  <span>Health:</span>
                  <span className={backendHealth === "HEALTHY" ? "text-[#04db7c]" : "text-[#f59e0b]"}>
                    {backendHealth}
                  </span>
                </div>
              </div>
            )}

            <div className="px-2.5 py-1.5 bg-[#011638] border border-[#0a2552] rounded-[4px] font-mono text-[10px] space-y-0.5">
              <div className="flex items-center justify-between text-[#5e6c84]">
                <span>TENANT:</span>
                <span className="text-[#f4f5f7] font-semibold">org_prod_us_east</span>
              </div>
              <div className="flex items-center justify-between text-[#5e6c84]">
                <span>STATUS:</span>
                <span className="text-[#04db7c] font-bold flex items-center gap-1">
                  <span
                    className={`w-1.5 h-1.5 rounded-full inline-block ${
                      backendHealth === "HEALTHY"
                        ? "bg-[#04db7c]"
                        : backendHealth === "OFFLINE"
                        ? "bg-[#f05252]"
                        : "bg-[#f59e0b]"
                    }`}
                  />
                  {backendHealth === "HEALTHY" ? "LIVE (LOCAL API)" : backendHealth}
                </span>
              </div>
            </div>
          </div>
        </aside>

        {/* Main Content Area */}
        <div className="flex-1 flex flex-col overflow-hidden">
          {/* Degraded Service Banner */}
          {isDegradedBannerOpen && (
            <div className="px-6 py-2 bg-[#f59e0b18] border-b border-[#f59e0b44] text-[#f59e0b] text-xs font-mono flex items-center justify-between shrink-0">
              <div className="flex items-center gap-2">
                <AlertTriangle className="w-4 h-4 text-[#f59e0b] shrink-0" />
                <span>
                  {backendHealth === "OFFLINE"
                    ? "Backend server is currently offline at http://localhost:8080. Operating with resilient mock adapters."
                    : `Backend service is currently ${backendHealth}. Real-time decisioning operational with fallback protection.`}
                </span>
              </div>
              <button
                onClick={() => setIsDegradedBannerOpen(false)}
                className="px-2 py-0.5 bg-[#f59e0b25] hover:bg-[#f59e0b40] rounded text-[10px] cursor-pointer"
              >
                Dismiss
              </button>
            </div>
          )}

          <main className="flex-1 overflow-y-auto p-6 lg:p-8 bg-[#070e1c]">
            {children}
          </main>
        </div>
      </body>
    </html>
  );
}
