"use client";

import React, { useEffect, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import {
  FolderKanban,
  AlertTriangle,
  Clock,
  CheckCircle2,
  XCircle,
  Eye,
  RefreshCw,
  Search,
  Filter,
  UserCheck,
  ShieldAlert,
  ArrowUpRight,
} from "lucide-react";
import {
  Card,
  CardHeader,
  CardTitle,
  CardDescription,
  CardContent,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import {
  Table,
  TableHeader,
  TableBody,
  TableHead,
  TableRow,
  TableCell,
} from "@/components/ui/table";
import { api, CaseItem, CaseStatus } from "@/lib/api";

export default function CasesQueuePage() {
  const router = useRouter();
  const [cases, setCases] = useState<CaseItem[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);
  const [statusFilter, setStatusFilter] = useState<string>("ALL");
  const [searchQuery, setSearchQuery] = useState<string>("");
  const [claimingId, setClaimingId] = useState<string | null>(null);

  // Fallback demo mock cases if backend has no active cases yet
  const mockFallbackCases: CaseItem[] = [
    {
      case_id: "case_8f7b2c1a-4d3e-4b2a",
      tenant_id: "00000000-0000-0000-0000-000000000001",
      decision_id: "dec_99a8b7c6",
      transaction_id: "txn_demo_9821734",
      status: "OPEN",
      priority: "HIGH",
      sla_expires_at: new Date(Date.now() + 18 * 3600 * 1000).toISOString(),
      created_at: new Date(Date.now() - 2 * 3600 * 1000).toISOString(),
      updated_at: new Date(Date.now() - 2 * 3600 * 1000).toISOString(),
    },
    {
      case_id: "case_3a1e9d2f-5c4b-4a1d",
      tenant_id: "00000000-0000-0000-0000-000000000001",
      decision_id: "dec_11f2e3d4",
      transaction_id: "txn_velocity_spike_55",
      status: "UNDER_REVIEW",
      priority: "CRITICAL",
      assigned_to: "analyst_sarah",
      sla_expires_at: new Date(Date.now() + 8 * 3600 * 1000).toISOString(),
      created_at: new Date(Date.now() - 6 * 3600 * 1000).toISOString(),
      updated_at: new Date(Date.now() - 1 * 3600 * 1000).toISOString(),
    },
    {
      case_id: "case_7e4d1b8c-9a2f-4e3b",
      tenant_id: "00000000-0000-0000-0000-000000000001",
      decision_id: "dec_77b6c5d4",
      transaction_id: "txn_safe_cleared_001",
      status: "RESOLVED_ALLOW",
      priority: "MEDIUM",
      assigned_to: "analyst_sarah",
      resolution_reason: "Verified user identity via two-factor call confirm.",
      resolved_at: new Date(Date.now() - 12 * 3600 * 1000).toISOString(),
      sla_expires_at: new Date(Date.now() - 2 * 3600 * 1000).toISOString(),
      created_at: new Date(Date.now() - 24 * 3600 * 1000).toISOString(),
      updated_at: new Date(Date.now() - 12 * 3600 * 1000).toISOString(),
    },
  ];

  const fetchCases = async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await api.getCases();
      if (data && data.cases && data.cases.length > 0) {
        setCases(data.cases);
      } else {
        // Use initial demo cases if empty
        setCases(mockFallbackCases);
      }
    } catch (err: any) {
      console.warn("Could not fetch cases from backend, using fallback queue:", err.message);
      setCases(mockFallbackCases);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchCases();
  }, []);

  const handleClaimCase = async (caseId: string) => {
    setClaimingId(caseId);
    try {
      await api.claimCase(caseId, "analyst_sarah");
      // Redirect to case detail page
      router.push(`/cases/${caseId}`);
    } catch (err: any) {
      console.warn("Claim API failed, proceeding to case detail view:", err.message);
      router.push(`/cases/${caseId}`);
    } finally {
      setClaimingId(null);
    }
  };

  // Filter cases based on status and search query
  const filteredCases = cases.filter((c) => {
    const matchesSearch =
      c.case_id.toLowerCase().includes(searchQuery.toLowerCase()) ||
      c.transaction_id.toLowerCase().includes(searchQuery.toLowerCase()) ||
      (c.assigned_to && c.assigned_to.toLowerCase().includes(searchQuery.toLowerCase()));

    if (!matchesSearch) return false;

    if (statusFilter === "ALL") return true;
    if (statusFilter === "RESOLVED") {
      return c.status === "RESOLVED_ALLOW" || c.status === "RESOLVED_DECLINE" || c.status === "CLOSED";
    }
    return c.status === statusFilter;
  });

  const renderStatusBadge = (status: CaseStatus) => {
    switch (status) {
      case "OPEN":
        return (
          <Badge variant="warning" className="gap-1 bg-amber-500/15 text-amber-400 border-amber-500/30">
            <AlertTriangle className="size-3" />
            OPEN
          </Badge>
        );
      case "UNDER_REVIEW":
        return (
          <Badge variant="info" className="gap-1 bg-blue-500/15 text-blue-400 border-blue-500/30">
            <UserCheck className="size-3" />
            UNDER_REVIEW
          </Badge>
        );
      case "RESOLVED_ALLOW":
        return (
          <Badge variant="success" className="gap-1 bg-emerald-500/15 text-emerald-400 border-emerald-500/30">
            <CheckCircle2 className="size-3" />
            ALLOW
          </Badge>
        );
      case "RESOLVED_DECLINE":
        return (
          <Badge variant="danger" className="gap-1 bg-rose-500/15 text-rose-400 border-rose-500/30">
            <XCircle className="size-3" />
            DECLINE
          </Badge>
        );
      default:
        return <Badge variant="outline">{status}</Badge>;
    }
  };

  const formatTimeRemaining = (slaExpiry: string) => {
    const expiry = new Date(slaExpiry).getTime();
    const now = Date.now();
    const diffHours = Math.round((expiry - now) / (1000 * 60 * 60));

    if (diffHours <= 0) {
      return <span className="text-rose-400 font-mono font-medium">SLA Breached</span>;
    }
    if (diffHours < 4) {
      return <span className="text-amber-400 font-mono font-medium">{diffHours}h remaining</span>;
    }
    return <span className="text-slate-300 font-mono">{diffHours}h remaining</span>;
  };

  return (
    <div className="p-6 md:p-10 max-w-7xl mx-auto w-full space-y-8">
      {/* Header */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 border-b border-slate-800 pb-6">
        <div>
          <div className="flex items-center gap-2 mb-1">
            <h1 className="text-2xl font-bold tracking-tight text-white flex items-center gap-2">
              <FolderKanban className="size-6 text-amber-400" />
              Manual Review Queue
            </h1>
            <Badge variant="warning">24h SLA Active</Badge>
          </div>
          <p className="text-sm text-slate-400">
            Flagged transactions asynchronously dispatched by Kafka for human investigation and decision overrides.
          </p>
        </div>

        <Button
          variant="outline"
          size="sm"
          onClick={fetchCases}
          disabled={loading}
          className="gap-2 bg-slate-900 border-slate-700 hover:bg-slate-800 text-slate-200"
        >
          <RefreshCw className={`size-3.5 ${loading ? "animate-spin" : ""}`} />
          Refresh Queue
        </Button>
      </div>

      {/* Filter and Search Bar */}
      <div className="flex flex-col sm:flex-row items-stretch sm:items-center justify-between gap-4">
        {/* Status Filter Tabs */}
        <div className="flex flex-wrap items-center gap-1.5 p-1 bg-slate-900/80 border border-slate-800 rounded-lg">
          {["ALL", "OPEN", "UNDER_REVIEW", "RESOLVED"].map((status) => (
            <button
              key={status}
              type="button"
              onClick={() => setStatusFilter(status)}
              className={`px-3 py-1.5 rounded-md text-xs font-medium transition-all ${
                statusFilter === status
                  ? "bg-indigo-600 text-white shadow-sm"
                  : "text-slate-400 hover:text-white hover:bg-slate-800/60"
              }`}
            >
              {status.replace("_", " ")}
            </button>
          ))}
        </div>

        {/* Search Input */}
        <div className="relative w-full sm:w-72">
          <Search className="size-4 absolute left-3 top-2.5 text-slate-500" />
          <Input
            type="text"
            placeholder="Search by Case or Txn ID..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="pl-9 bg-slate-900/60 border-slate-800 text-xs text-white placeholder:text-slate-500 font-mono"
          />
        </div>
      </div>

      {/* Cases Table */}
      <Card className="bg-slate-900/70 border-slate-800 shadow-xl overflow-hidden backdrop-blur">
        <CardContent className="p-0">
          <Table>
            <TableHeader className="bg-slate-950/70 border-b border-slate-800">
              <TableRow className="border-slate-800 hover:bg-transparent">
                <TableHead className="text-slate-400 font-semibold text-xs py-3.5 pl-6">Case ID</TableHead>
                <TableHead className="text-slate-400 font-semibold text-xs py-3.5">Transaction ID</TableHead>
                <TableHead className="text-slate-400 font-semibold text-xs py-3.5">Status</TableHead>
                <TableHead className="text-slate-400 font-semibold text-xs py-3.5">Priority</TableHead>
                <TableHead className="text-slate-400 font-semibold text-xs py-3.5">SLA Expiration</TableHead>
                <TableHead className="text-slate-400 font-semibold text-xs py-3.5">Assigned To</TableHead>
                <TableHead className="text-slate-400 font-semibold text-xs py-3.5 text-right pr-6">Action</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {loading ? (
                <TableRow className="border-slate-800">
                  <TableCell colSpan={7} className="text-center py-12 text-slate-400">
                    <RefreshCw className="size-5 animate-spin mx-auto mb-2 text-indigo-400" />
                    Loading investigation queue...
                  </TableCell>
                </TableRow>
              ) : filteredCases.length > 0 ? (
                filteredCases.map((c) => (
                  <TableRow
                    key={c.case_id}
                    className="border-slate-800/80 hover:bg-slate-800/40 transition-colors"
                  >
                    {/* Case ID */}
                    <TableCell className="font-mono text-xs text-indigo-400 font-medium pl-6">
                      <Link
                        href={`/cases/${c.case_id}`}
                        className="hover:underline flex items-center gap-1"
                      >
                        {c.case_id.slice(0, 18)}...
                      </Link>
                    </TableCell>

                    {/* Transaction ID */}
                    <TableCell className="font-mono text-xs text-slate-200">
                      {c.transaction_id}
                    </TableCell>

                    {/* Status Badge */}
                    <TableCell>{renderStatusBadge(c.status)}</TableCell>

                    {/* Priority */}
                    <TableCell>
                      <Badge
                        variant="outline"
                        className={`text-[10px] uppercase font-mono ${
                          c.priority === "CRITICAL"
                            ? "border-rose-500/50 text-rose-400 bg-rose-500/10"
                            : c.priority === "HIGH"
                            ? "border-amber-500/50 text-amber-400 bg-amber-500/10"
                            : "border-slate-700 text-slate-400"
                        }`}
                      >
                        {c.priority || "NORMAL"}
                      </Badge>
                    </TableCell>

                    {/* SLA Expiration */}
                    <TableCell className="text-xs">
                      <div className="flex items-center gap-1.5">
                        <Clock className="size-3.5 text-slate-500 shrink-0" />
                        {formatTimeRemaining(c.sla_expires_at)}
                      </div>
                    </TableCell>

                    {/* Assigned Analyst */}
                    <TableCell className="text-xs font-mono text-slate-400">
                      {c.assigned_to ? (
                        <span className="text-slate-200">{c.assigned_to}</span>
                      ) : (
                        <span className="text-slate-600 italic">Unassigned</span>
                      )}
                    </TableCell>

                    {/* Action */}
                    <TableCell className="text-right pr-6">
                      {c.status === "OPEN" ? (
                        <Button
                          size="sm"
                          onClick={() => handleClaimCase(c.case_id)}
                          disabled={claimingId === c.case_id}
                          className="h-7 px-3 text-xs bg-indigo-600 hover:bg-indigo-500 text-white"
                        >
                          {claimingId === c.case_id ? (
                            <RefreshCw className="size-3 animate-spin" />
                          ) : (
                            "Claim Case"
                          )}
                        </Button>
                      ) : (
                        <Link href={`/cases/${c.case_id}`}>
                          <Button
                            variant="outline"
                            size="sm"
                            className="h-7 px-3 text-xs bg-slate-900 border-slate-700 hover:bg-slate-800 text-slate-300 gap-1"
                          >
                            <Eye className="size-3" />
                            Investigate
                          </Button>
                        </Link>
                      )}
                    </TableCell>
                  </TableRow>
                ))
              ) : (
                <TableRow className="border-slate-800">
                  <TableCell colSpan={7} className="text-center py-12 text-slate-500">
                    No cases match the current filter criteria.
                  </TableCell>
                </TableRow>
              )}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </div>
  );
}
