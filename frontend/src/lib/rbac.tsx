import { createContext, useContext, type ReactNode } from "react";

/**
 * Frontend RBAC is presentational only. The ROPUS backend remains authoritative
 * for every authorization decision; hiding or disabling a control here is a
 * usability affordance, never an enforcement boundary.
 */

export type Permission =
  | "decisions:read"
  | "cases:read"
  | "cases:write"
  | "cases:approve"
  | "rules:write"
  | "models:write"
  | "keys:write"
  | "webhooks:write"
  | "billing:read"
  | "team:write";

export interface Session {
  user: string;
  email: string;
  role: string;
  organization: string;
  tenantId: string;
  environment: "PRODUCTION" | "SANDBOX";
  permissions: Permission[];
}

const demoSession: Session = {
  user: "m.okafor",
  email: "m.okafor@ropus.internal",
  role: "Risk Analyst",
  organization: "Northbank Payments",
  tenantId: "tnt_4f19ac",
  environment: "PRODUCTION",
  permissions: ["decisions:read", "cases:read", "cases:write", "cases:approve", "billing:read"],
};

const SessionContext = createContext<Session>(demoSession);

export function SessionProvider({ children }: { children: ReactNode }) {
  return <SessionContext.Provider value={demoSession}>{children}</SessionContext.Provider>;
}

export function useSession() {
  return useContext(SessionContext);
}

export function useCan(permission: Permission) {
  return useSession().permissions.includes(permission);
}

export function PermissionGate({
  permission,
  mode = "hide",
  children,
  fallback,
}: {
  permission: Permission;
  mode?: "hide" | "explain";
  children: ReactNode;
  fallback?: ReactNode;
}) {
  const allowed = useCan(permission);
  if (allowed) return <>{children}</>;
  if (fallback) return <>{fallback}</>;
  if (mode === "hide") return null;
  return (
    <p className="border border-border border-l-2 border-l-border-strong bg-neutral-surface px-3 py-2 text-[12px] text-muted-foreground">
      Requires <span className="font-mono">{permission}</span>. Your role ({useSessionRole()}) does not hold this
      grant. Backend authorization remains authoritative.
    </p>
  );
}

function useSessionRole() {
  return useSession().role;
}
