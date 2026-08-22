"use client";

import React, { useState } from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import "./globals.css";
import {
  LayoutDashboard,
  ShieldAlert,
  GitGraph,
  FolderKanban,
  PlayCircle,
  Server,
  ChevronRight,
} from "lucide-react";

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const pathname = usePathname();
  const [isPlatformDrawerOpen, setIsPlatformDrawerOpen] = useState(false);

  const mainNavItems = [
    { name: "Overview", href: "/", icon: LayoutDashboard },
    { name: "Risk Decisions", href: "/transactions", icon: ShieldAlert },
    { name: "Fraud Graph", href: "/graph", icon: GitGraph },
    { name: "Cases", href: "/cases", icon: FolderKanban },
  ];

  return (
    <html lang="en">
      <head>
        <title>ROPUS — Risk Control Plane</title>
        <meta
          name="description"
          content="Sub-10ms explainable AI fraud decisioning, 3-hop graph intelligence & autonomous investigation platform"
        />
      </head>
      <body className="flex h-screen overflow-hidden bg-[#070e1c] text-[#f4f5f7]">
        {/* Sidebar Navigation */}
        <aside className="w-60 bg-[#0b1528] border-r border-[#1c2b48] flex flex-col justify-between shrink-0 select-none">
          <div>
            {/* Logo / Header */}
            <div className="px-5 py-4 border-b border-[#1c2b48] flex items-center justify-between">
              <Link href="/" className="flex items-center gap-2">
                <div className="w-5 h-5 bg-[#0d94fb] rounded flex items-center justify-center font-mono text-[11px] font-black text-white">
                  R
                </div>
                <span className="font-mono font-bold text-sm tracking-wider text-[#f4f5f7]">
                  ROPUS
                </span>
                <span className="text-[10px] font-mono text-[#6b778c] bg-[#14223d] px-1 py-0.2 rounded">
                  v3.39
                </span>
              </Link>
            </div>

            {/* 4 Primary Navigation Items */}
            <nav className="p-3 space-y-1">
              {mainNavItems.map((item) => {
                const isActive =
                  item.href === "/" ? pathname === "/" : pathname.startsWith(item.href);
                const Icon = item.icon;

                return (
                  <Link
                    key={item.href}
                    href={item.href}
                    className={`flex items-center gap-2.5 px-3 py-2 rounded text-xs font-mono transition-colors ${
                      isActive
                        ? "bg-[#0d94fb18] text-[#0d94fb] font-semibold border border-[#0d94fb44]"
                        : "text-[#97a0af] hover:text-[#f4f5f7] hover:bg-[#0f1c34]"
                    }`}
                  >
                    <Icon className="w-4 h-4" />
                    <span>{item.name}</span>
                  </Link>
                );
              })}

              {/* Visual Divider Before Demo */}
              <div className="pt-3 pb-1">
                <div className="h-px bg-[#1c2b48] w-full" />
              </div>

              {/* 5. Demo Walkthrough */}
              <Link
                href="/demo"
                className={`flex items-center justify-between px-3 py-2 rounded text-xs font-mono transition-colors ${
                  pathname.startsWith("/demo")
                    ? "bg-[#f59e0b18] text-[#f59e0b] font-bold border border-[#f59e0b44]"
                    : "text-[#f59e0b] hover:bg-[#f59e0b10]"
                }`}
              >
                <div className="flex items-center gap-2.5">
                  <PlayCircle className="w-4 h-4 text-[#f59e0b]" />
                  <span>Demo Experience</span>
                </div>
                <span className="text-[9px] bg-[#f59e0b22] px-1 py-0.2 rounded uppercase">
                  5-min
                </span>
              </Link>
            </nav>
          </div>

          {/* Sidebar Footer: Multi-Tenancy / Auth Context & Minimal Platform Link */}
          <div className="p-3 border-t border-[#1c2b48] bg-[#070e1c] space-y-2">
            {/* Secondary Platform Expander Trigger */}
            <button
              onClick={() => setIsPlatformDrawerOpen(!isPlatformDrawerOpen)}
              className="w-full px-2.5 py-1.5 text-[11px] font-mono text-[#6b778c] hover:text-[#f4f5f7] hover:bg-[#0b1528] rounded flex items-center justify-between transition-colors"
            >
              <div className="flex items-center gap-1.5">
                <Server className="w-3.5 h-3.5" />
                <span>Platform Controls</span>
              </div>
              <ChevronRight
                className={`w-3 h-3 transition-transform ${
                  isPlatformDrawerOpen ? "rotate-90" : ""
                }`}
              />
            </button>

            {/* Collapsed Minimal Platform Drawer */}
            {isPlatformDrawerOpen && (
              <div className="p-2 bg-[#0b1528] border border-[#1c2b48] rounded space-y-1 text-[10px] font-mono">
                <div className="flex justify-between text-[#97a0af]">
                  <span>RBAC Role:</span>
                  <span className="text-[#f4f5f7]">OWNER</span>
                </div>
                <div className="flex justify-between text-[#97a0af]">
                  <span>Crypto Key:</span>
                  <span className="text-[#04db7c]">SHA-256 Valid</span>
                </div>
                <div className="flex justify-between text-[#97a0af]">
                  <span>AES-256 GCM:</span>
                  <span className="text-[#04db7c]">Enabled</span>
                </div>
                <div className="flex justify-between text-[#97a0af]">
                  <span>Rate Quota:</span>
                  <span className="text-[#f4f5f7]">5000 RPS</span>
                </div>
              </div>
            )}

            {/* Persistent Tenant & Auth Scoping */}
            <div className="px-2 py-1.5 bg-[#0b1528] border border-[#14223d] rounded font-mono text-[10px] space-y-0.5">
              <div className="flex items-center justify-between text-[#6b778c]">
                <span>TENANT:</span>
                <span className="text-[#f4f5f7] font-semibold">org_prod_us_east</span>
              </div>
              <div className="flex items-center justify-between text-[#6b778c]">
                <span>ENV:</span>
                <span className="text-[#04db7c] font-bold">LIVE PRODUCTION</span>
              </div>
            </div>
          </div>
        </aside>

        {/* Main Content Area */}
        <main className="flex-1 overflow-y-auto p-6 lg:p-8 bg-[#070e1c]">
          {children}
        </main>
      </body>
    </html>
  );
}
