"use client";

import React, { useState } from "react";
import { Key, Plus, Copy, Check, RotateCw, Trash2 } from "lucide-react";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";

interface KeyRecord {
  id: string;
  name: string;
  prefix: string;
  environment: "live" | "test";
  created: string;
  lastUsed: string;
  status: "ACTIVE" | "REVOKED";
}

export default function APIKeysPage() {
  const [keys, setKeys] = useState<KeyRecord[]>([
    {
      id: "key_88a91c2b",
      name: "Production Payment Gateway",
      prefix: "rop_live_8a19bc...",
      environment: "live",
      created: "2026-08-01",
      lastUsed: "Just now",
      status: "ACTIVE",
    },
    {
      id: "key_33f481e0",
      name: "Staging Test Sandbox",
      prefix: "rop_test_99f412...",
      environment: "test",
      created: "2026-08-10",
      lastUsed: "2 hours ago",
      status: "ACTIVE",
    },
  ]);
  const [copiedId, setCopiedId] = useState<string | null>(null);

  const handleCopy = (id: string, text: string) => {
    navigator.clipboard.writeText(text);
    setCopiedId(id);
    setTimeout(() => setCopiedId(null), 2000);
  };

  const handleRotate = (id: string) => {
    setKeys(
      keys.map((k) =>
        k.id === id ? { ...k, prefix: "rop_live_rotated_" + Math.random().toString(36).substring(2, 8) + "..." } : k
      )
    );
  };

  const handleRevoke = (id: string) => {
    setKeys(keys.map((k) => (k.id === id ? { ...k, status: "REVOKED" } : k)));
  };

  return (
    <div className="flex-1 p-6 md:p-8 space-y-6 max-w-6xl mx-auto">
      {/* Header */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 border-b border-slate-800 pb-5">
        <div>
          <h1 className="text-2xl font-bold text-white tracking-tight flex items-center gap-2.5">
            <Key className="size-6 text-indigo-400" />
            <span>API Key Management</span>
          </h1>
          <p className="text-sm text-slate-400 mt-1">
            Provision cryptographically secure, SHA-256 hashed API keys for production transaction evaluation.
          </p>
        </div>
        <Button className="bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-semibold">
          <Plus className="size-3.5 mr-1.5" /> Create Secret Key
        </Button>
      </div>

      {/* Keys Table Card */}
      <Card className="bg-slate-900/80 border-slate-800">
        <CardHeader>
          <CardTitle className="text-base font-semibold text-white">Active Organization Keys</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="overflow-x-auto">
            <table className="w-full text-left text-xs">
              <thead className="border-b border-slate-800 text-slate-400 uppercase font-mono">
                <tr>
                  <th className="pb-3">Name / ID</th>
                  <th className="pb-3">Key Token Prefix</th>
                  <th className="pb-3">Environment</th>
                  <th className="pb-3">Created</th>
                  <th className="pb-3">Last Active</th>
                  <th className="pb-3">Status</th>
                  <th className="pb-3 text-right">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-800/60 font-mono">
                {keys.map((k) => (
                  <tr key={k.id} className="hover:bg-slate-800/30">
                    <td className="py-3.5 font-sans font-medium text-slate-200">
                      <div>{k.name}</div>
                      <span className="text-[11px] text-slate-500 font-mono">{k.id}</span>
                    </td>
                    <td className="py-3.5 text-indigo-300 flex items-center gap-2">
                      <span>{k.prefix}</span>
                      <button
                        onClick={() => handleCopy(k.id, k.prefix)}
                        className="text-slate-500 hover:text-slate-200"
                      >
                        {copiedId === k.id ? <Check className="size-3.5 text-emerald-400" /> : <Copy className="size-3.5" />}
                      </button>
                    </td>
                    <td className="py-3.5">
                      <Badge variant="outline" className={k.environment === "live" ? "border-emerald-500 text-emerald-300" : "border-amber-500 text-amber-300"}>
                        {k.environment.toUpperCase()}
                      </Badge>
                    </td>
                    <td className="py-3.5 text-slate-400 font-sans">{k.created}</td>
                    <td className="py-3.5 text-slate-400 font-sans">{k.lastUsed}</td>
                    <td className="py-3.5">
                      <Badge className={k.status === "ACTIVE" ? "bg-emerald-500/20 text-emerald-300" : "bg-rose-500/20 text-rose-300"}>
                        {k.status}
                      </Badge>
                    </td>
                    <td className="py-3.5 text-right space-x-2">
                      {k.status === "ACTIVE" && (
                        <>
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => handleRotate(k.id)}
                            className="h-7 text-xs text-slate-300 hover:text-white"
                          >
                            <RotateCw className="size-3 mr-1" /> Rotate
                          </Button>
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => handleRevoke(k.id)}
                            className="h-7 text-xs text-rose-400 hover:text-rose-300 hover:bg-rose-500/10"
                          >
                            <Trash2 className="size-3 mr-1" /> Revoke
                          </Button>
                        </>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
