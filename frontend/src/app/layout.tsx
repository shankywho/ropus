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
          content="Razorpay Blade developer-first AI risk & fraud control plane"
        />
      </head>
      <body className="flex h-screen overflow-hidden bg-[#070e1c] text-[#f4f5f7]">
        {/* Razorpay Blade Prussian Blue Sidebar */}
        <aside className="w-64 bg-[#011638] border-r border-[#0a2552] flex flex-col justify-between shrink-0 select-none shadow-md">
          <div>
            {/* Header Brand */}
            <div className="px-5 py-4 border-b border-[#0a2552] flex items-center justify-between">
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
                    Risk & Fraud Mesh
                  </p>
                </div>
              </Link>
            </div>

            {/* Main Navigation */}
            <nav className="p-3 space-y-1">
              <div className="px-3 py-1 text-[10px] font-mono font-bold text-[#5e6c84] uppercase tracking-wider">
                CORE PLATFORM
              </div>

              {mainNavItems.map((item) => {
                const isActive =
                  item.href === "/" ? pathname === "/" : pathname.startsWith(item.href);
                const Icon = item.icon;

                return (
                  <Link
                    key={item.href}
                    href={item.href}
                    className={`flex items-center gap-2.5 px-3 py-2 rounded-[4px] text-xs font-semibold transition-all ${
                      isActive
                        ? "bg-[#0d94fb] text-white shadow-xs"
                        : "text-[#97a0af] hover:text-[#f4f5f7] hover:bg-[#06204c]"
                    }`}
                  >
                    <Icon className="w-4 h-4 shrink-0" />
                    <span>{item.name}</span>
                  </Link>
                );
              })}

              {/* Visual Divider Before Demo */}
              <div className="pt-3 pb-1">
                <div className="h-px bg-[#0a2552] w-full" />
              </div>

              <div className="px-3 py-1 text-[10px] font-mono font-bold text-[#5e6c84] uppercase tracking-wider">
                SIMULATION & DEMO
              </div>

              {/* Demo Walkthrough Item */}
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
                  <span>Demo Experience</span>
                </div>
                <span
                  className={`text-[9px] font-mono px-1.5 py-0.2 rounded-[2px] uppercase ${
                    pathname.startsWith("/demo")
                      ? "bg-[#011638] text-[#f59e0b]"
                      : "bg-[#f59e0b22] text-[#f59e0b]"
                  }`}
                >
                  5-min
                </span>
              </Link>
            </nav>
          </div>

          {/* Sidebar Footer: Multi-Tenancy / Auth Context & Platform Link */}
          <div className="p-3 border-t border-[#0a2552] bg-[#01112b] space-y-2">
            {/* Secondary Platform Expander Trigger */}
            <button
              onClick={() => setIsPlatformDrawerOpen(!isPlatformDrawerOpen)}
              className="w-full px-2.5 py-1.5 text-[11px] font-mono text-[#97a0af] hover:text-[#f4f5f7] hover:bg-[#06204c] rounded-[4px] flex items-center justify-between transition-colors cursor-pointer"
            >
              <div className="flex items-center gap-1.5">
                <Server className="w-3.5 h-3.5 text-[#0d94fb]" />
                <span>Blade Infrastructure</span>
              </div>
              <ChevronRight
                className={`w-3 h-3 transition-transform ${
                  isPlatformDrawerOpen ? "rotate-90" : ""
                }`}
              />
            </button>

            {/* Collapsed Platform Drawer */}
            {isPlatformDrawerOpen && (
              <div className="p-2.5 bg-[#011638] border border-[#0a2552] rounded-[4px] space-y-1.5 text-[10px] font-mono">
                <div className="flex justify-between text-[#97a0af]">
                  <span>Design System:</span>
                  <span className="text-[#0d94fb] font-semibold">Razorpay Blade</span>
                </div>
                <div className="flex justify-between text-[#97a0af]">
                  <span>RBAC Role:</span>
                  <span className="text-[#f4f5f7] font-semibold">OWNER</span>
                </div>
                <div className="flex justify-between text-[#97a0af]">
                  <span>Key Hash:</span>
                  <span className="text-[#04db7c]">SHA-256 Valid</span>
                </div>
                <div className="flex justify-between text-[#97a0af]">
                  <span>Envelope Crypto:</span>
                  <span className="text-[#04db7c]">AES-256 GCM</span>
                </div>
                <div className="flex justify-between text-[#97a0af]">
                  <span>Rate Quota:</span>
                  <span className="text-[#f4f5f7]">5000 RPS</span>
                </div>
              </div>
            )}

            {/* Persistent Tenant & Auth Scoping */}
            <div className="px-2.5 py-2 bg-[#011638] border border-[#0a2552] rounded-[4px] font-mono text-[10px] space-y-1">
              <div className="flex items-center justify-between text-[#5e6c84]">
                <span>TENANT:</span>
                <span className="text-[#f4f5f7] font-semibold">org_prod_us_east</span>
              </div>
              <div className="flex items-center justify-between text-[#5e6c84]">
                <span>ENVIRONMENT:</span>
                <span className="text-[#04db7c] font-bold flex items-center gap-1">
                  <span className="w-1.5 h-1.5 rounded-full bg-[#04db7c] inline-block" />
                  LIVE
                </span>
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
