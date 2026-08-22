"use client";

import React, { useState } from "react";
import { Users, UserPlus, Shield, Trash2 } from "lucide-react";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";

interface TeamMember {
  id: string;
  name: string;
  email: string;
  role: "OWNER" | "ADMIN" | "ANALYST" | "VIEWER";
  joined: string;
}

export default function TeamPage() {
  const [members, setMembers] = useState<TeamMember[]>([
    { id: "usr_01", name: "Sarah Jenkins", email: "sarah.j@acmebank.com", role: "OWNER", joined: "2026-01-15" },
    { id: "usr_02", name: "David Chen", email: "david.c@acmebank.com", role: "ADMIN", joined: "2026-02-01" },
    { id: "usr_03", name: "Elena Rostova", email: "elena.r@acmebank.com", role: "ANALYST", joined: "2026-03-10" },
    { id: "usr_04", name: "Marcus Vance", email: "marcus.v@acmebank.com", role: "VIEWER", joined: "2026-04-22" },
  ]);

  return (
    <div className="flex-1 p-6 md:p-8 space-y-6 max-w-6xl mx-auto">
      {/* Header */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 border-b border-slate-800 pb-5">
        <div>
          <h1 className="text-2xl font-bold text-white tracking-tight flex items-center gap-2.5">
            <Users className="size-6 text-indigo-400" />
            <span>Team & Access Control</span>
          </h1>
          <p className="text-sm text-slate-400 mt-1">
            Manage organization members, assign role-based access permissions, and audit user logins.
          </p>
        </div>
        <Button className="bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-semibold">
          <UserPlus className="size-3.5 mr-1.5" /> Invite Team Member
        </Button>
      </div>

      {/* Members Table */}
      <Card className="bg-slate-900/80 border-slate-800">
        <CardHeader>
          <CardTitle className="text-base font-semibold text-white">Active Members ({members.length})</CardTitle>
        </CardHeader>
        <CardContent>
          <table className="w-full text-left text-xs">
            <thead className="border-b border-slate-800 text-slate-400 uppercase font-mono">
              <tr>
                <th className="pb-3">User</th>
                <th className="pb-3">Role</th>
                <th className="pb-3">Joined Date</th>
                <th className="pb-3 text-right">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-800/60 font-mono">
              {members.map((m) => (
                <tr key={m.id} className="hover:bg-slate-800/30">
                  <td className="py-3 font-sans">
                    <div className="font-semibold text-slate-200">{m.name}</div>
                    <div className="text-slate-400 text-[11px] font-mono">{m.email}</div>
                  </td>
                  <td className="py-3">
                    <Badge
                      className={
                        m.role === "OWNER"
                          ? "bg-purple-500/20 text-purple-300"
                          : m.role === "ADMIN"
                          ? "bg-indigo-500/20 text-indigo-300"
                          : m.role === "ANALYST"
                          ? "bg-blue-500/20 text-blue-300"
                          : "bg-slate-500/20 text-slate-300"
                      }
                    >
                      <Shield className="size-3 mr-1" />
                      {m.role}
                    </Badge>
                  </td>
                  <td className="py-3 text-slate-400 font-sans">{m.joined}</td>
                  <td className="py-3 text-right">
                    {m.role !== "OWNER" && (
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => setMembers(members.filter((u) => u.id !== m.id))}
                        className="h-7 text-xs text-rose-400 hover:text-rose-300 hover:bg-rose-500/10"
                      >
                        <Trash2 className="size-3 mr-1" /> Remove
                      </Button>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </CardContent>
      </Card>
    </div>
  );
}
