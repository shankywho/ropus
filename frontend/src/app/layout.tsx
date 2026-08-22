import type { Metadata } from "next";
import { Inter } from "next/font/google";
import Link from "next/link";
import "./globals.css";
import {
  ShieldAlert,
  PlayCircle,
  FolderKanban,
  Sliders,
  Activity,
  CheckCircle2,
  Lock,
  Key,
  CreditCard,
  Users,
  Settings,
  Server,
  ShieldCheck,
} from "lucide-react";

const inter = Inter({ subsets: ["latin"] });

export const metadata: Metadata = {
  title: "AI Risk Manager — Fintech Enterprise Production Platform",
  description:
    "Real-time fraud decisioning, rules engine, and analyst case management platform",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" className="dark">
      <body
        className={`${inter.className} min-h-screen bg-slate-950 text-slate-100 antialiased flex flex-col md:flex-row`}
      >
        {/* Sidebar Navigation */}
        <aside className="w-full md:w-64 bg-slate-900 border-r border-slate-800 flex flex-col shrink-0">
          {/* Logo & Title */}
          <div className="p-5 border-b border-slate-800 flex items-center gap-3">
            <div className="size-9 rounded-lg bg-gradient-to-br from-indigo-500 to-purple-600 flex items-center justify-center shadow-lg shadow-indigo-500/20">
              <ShieldAlert className="size-5 text-white" />
            </div>
            <div>
              <h1 className="font-bold text-sm tracking-tight text-white flex items-center gap-1.5">
                AI Risk Manager
              </h1>
              <p className="text-xs text-slate-400 font-mono">v3.38 Hardened Production</p>
            </div>
          </div>

          {/* Navigation Links */}
          <nav className="flex-1 p-3 space-y-4 overflow-y-auto">
            <div className="space-y-1">
              <p className="px-3 text-[10px] font-mono uppercase text-slate-500 font-bold tracking-wider">
                Risk Operations
              </p>
              <Link
                href="/"
                className="flex items-center gap-3 px-3 py-2 rounded-lg text-xs font-medium text-slate-300 hover:text-white hover:bg-slate-800/60 transition-colors"
              >
                <Activity className="size-4 text-indigo-400" />
                <span>Overview</span>
              </Link>

              <Link
                href="/transactions"
                className="flex items-center gap-3 px-3 py-2 rounded-lg text-xs font-medium text-slate-300 hover:text-white hover:bg-slate-800/60 transition-colors"
              >
                <Activity className="size-4 text-cyan-400" />
                <span>Transactions</span>
                <span className="ml-auto text-[10px] bg-cyan-500/20 text-cyan-300 px-1.5 py-0.5 rounded font-mono">
                  Live
                </span>
              </Link>

              <Link
                href="/graph"
                className="flex items-center gap-3 px-3 py-2 rounded-lg text-xs font-medium text-slate-300 hover:text-white hover:bg-slate-800/60 transition-colors"
              >
                <ShieldAlert className="size-4 text-rose-400" />
                <span>Fraud Graph</span>
              </Link>

              <Link
                href="/cases"
                className="flex items-center gap-3 px-3 py-2 rounded-lg text-xs font-medium text-slate-300 hover:text-white hover:bg-slate-800/60 transition-colors"
              >
                <FolderKanban className="size-4 text-amber-400" />
                <span>Review Queue</span>
              </Link>

              <Link
                href="/rules"
                className="flex items-center gap-3 px-3 py-2 rounded-lg text-xs font-medium text-slate-300 hover:text-white hover:bg-slate-800/60 transition-colors"
              >
                <Sliders className="size-4 text-purple-400" />
                <span>Rules Engine</span>
              </Link>

              <Link
                href="/governance"
                className="flex items-center gap-3 px-3 py-2 rounded-lg text-xs font-medium text-slate-300 hover:text-white hover:bg-slate-800/60 transition-colors"
              >
                <CheckCircle2 className="size-4 text-teal-400" />
                <span>Model Governance</span>
              </Link>

              <Link
                href="/demo"
                className="flex items-center gap-3 px-3 py-2 rounded-lg text-xs font-medium text-slate-300 hover:text-white hover:bg-slate-800/60 transition-colors"
              >
                <PlayCircle className="size-4 text-emerald-400" />
                <span>Command Center</span>
                <span className="ml-auto text-[10px] bg-emerald-500/20 text-emerald-300 px-1.5 py-0.5 rounded font-mono font-bold">
                  Demo
                </span>
              </Link>
            </div>

            <div className="space-y-1 pt-2 border-t border-slate-800">
              <p className="px-3 text-[10px] font-mono uppercase text-slate-500 font-bold tracking-wider">
                Fintech Operations
              </p>
              <Link
                href="/operations"
                className="flex items-center gap-3 px-3 py-2 rounded-lg text-xs font-medium text-slate-300 hover:text-white hover:bg-slate-800/60 transition-colors"
              >
                <Server className="size-4 text-indigo-400" />
                <span>Infra Control Plane</span>
              </Link>

              <Link
                href="/security"
                className="flex items-center gap-3 px-3 py-2 rounded-lg text-xs font-medium text-slate-300 hover:text-white hover:bg-slate-800/60 transition-colors"
              >
                <ShieldCheck className="size-4 text-emerald-400" />
                <span>Security & Audit</span>
              </Link>
            </div>

            <div className="space-y-1 pt-2 border-t border-slate-800">
              <p className="px-3 text-[10px] font-mono uppercase text-slate-500 font-bold tracking-wider">
                Enterprise SaaS
              </p>
              <Link
                href="/api-keys"
                className="flex items-center gap-3 px-3 py-2 rounded-lg text-xs font-medium text-slate-300 hover:text-white hover:bg-slate-800/60 transition-colors"
              >
                <Key className="size-4 text-indigo-400" />
                <span>API Keys</span>
              </Link>

              <Link
                href="/billing"
                className="flex items-center gap-3 px-3 py-2 rounded-lg text-xs font-medium text-slate-300 hover:text-white hover:bg-slate-800/60 transition-colors"
              >
                <CreditCard className="size-4 text-amber-400" />
                <span>Subscription & Billing</span>
              </Link>

              <Link
                href="/team"
                className="flex items-center gap-3 px-3 py-2 rounded-lg text-xs font-medium text-slate-300 hover:text-white hover:bg-slate-800/60 transition-colors"
              >
                <Users className="size-4 text-blue-400" />
                <span>Team & Access</span>
              </Link>

              <Link
                href="/settings"
                className="flex items-center gap-3 px-3 py-2 rounded-lg text-xs font-medium text-slate-300 hover:text-white hover:bg-slate-800/60 transition-colors"
              >
                <Settings className="size-4 text-slate-400" />
                <span>Settings & Policies</span>
              </Link>
            </div>
          </nav>

          {/* Architecture Status Panel */}
          <div className="p-3 m-3 rounded-lg bg-slate-950/60 border border-slate-800 text-xs space-y-2">
            <div className="flex items-center justify-between text-slate-400 font-medium">
              <span>Cluster State</span>
              <span className="flex items-center gap-1 text-emerald-400 text-[11px]">
                <span className="size-1.5 rounded-full bg-emerald-500 animate-pulse" />
                99.995% SLA
              </span>
            </div>
            <div className="space-y-1 text-[11px] text-slate-400 font-mono">
              <div className="flex justify-between">
                <span>P99 SLA:</span>
                <span className="text-slate-200">6.8ms</span>
              </div>
              <div className="flex justify-between">
                <span>Throughput:</span>
                <span className="text-slate-200">104k/sec</span>
              </div>
            </div>
          </div>
        </aside>

        {/* Main Content Area */}
        <main className="flex-1 flex flex-col min-w-0 overflow-y-auto">
          {children}
        </main>
      </body>
    </html>
  );
}
