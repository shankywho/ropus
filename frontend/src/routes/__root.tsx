import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import {
  Outlet,
  Link,
  createRootRouteWithContext,
  useRouter,
  HeadContent,
  Scripts,
} from "@tanstack/react-router";
import { useEffect, type ReactNode } from "react";

import appCss from "../styles.css?url";
import { reportLovableError } from "../lib/lovable-error-reporting";
import { AppShell } from "@/components/ropus/app-shell";
import { SessionProvider } from "@/lib/rbac";

function NotFoundComponent() {
  return (
    <div className="flex min-h-[60vh] items-center justify-center px-6">
      <div className="max-w-md border border-border bg-card p-6 shadow-xs">
        <span className="font-mono text-[9.5px] font-semibold tracking-[0.12em] text-muted-foreground uppercase">
          HTTP 404
        </span>
        <h1 className="mt-2 font-sans text-lg font-bold tracking-tight">Route not found</h1>
        <p className="mt-1 font-sans text-[12.5px] text-muted-foreground">
          No control-plane surface is registered at this path.
        </p>
        <Link
          to="/"
          className="mt-4 inline-flex items-center border border-border-strong bg-surface px-3 py-1.5 font-sans text-[12px] font-medium hover:bg-accent"
        >
          Return to Overview
        </Link>
      </div>
    </div>
  );
}

function ErrorComponent({ error, reset }: { error: Error; reset: () => void }) {
  console.error(error);
  const router = useRouter();
  useEffect(() => {
    reportLovableError(error, { boundary: "tanstack_root_error_component" });
  }, [error]);

  return (
    <div className="flex min-h-[60vh] items-center justify-center px-6">
      <div className="max-w-md border border-block/40 border-l-2 border-l-block bg-card p-6 shadow-xs">
        <span className="font-mono text-[9.5px] font-semibold tracking-[0.12em] text-block uppercase">
          Render failure
        </span>
        <h1 className="mt-2 font-sans text-lg font-bold tracking-tight">
          This surface did not load
        </h1>
        <p className="mt-1 font-sans text-[12.5px] text-muted-foreground">
          The view failed before data could be rendered. Retry, or return to the Overview.
        </p>
        <div className="mt-4 flex gap-2">
          <button
            onClick={() => {
              router.invalidate();
              reset();
            }}
            className="border border-primary bg-primary px-3 py-1.5 font-sans text-[12px] font-medium text-primary-foreground cursor-pointer"
          >
            Retry
          </button>
          <a
            href="/"
            className="border border-border-strong bg-surface px-3 py-1.5 font-sans text-[12px] font-medium hover:bg-accent"
          >
            Overview
          </a>
        </div>
      </div>
    </div>
  );
}

export const Route = createRootRouteWithContext<{ queryClient: QueryClient }>()({
  head: () => ({
    meta: [
      { charSet: "utf-8" },
      { name: "viewport", content: "width=device-width, initial-scale=1" },
      { title: "ROPUS — Risk Intelligence Control Plane" },
      {
        name: "description",
        content:
          "ROPUS control plane for transaction risk decisioning, fraud investigation, rules, models and platform operations.",
      },
      { property: "og:title", content: "ROPUS — Risk Intelligence Control Plane" },
      {
        property: "og:description",
        content:
          "Transaction risk decisioning, fraud investigation and platform operations in one control surface.",
      },
      { property: "og:type", content: "website" },
      { name: "twitter:card", content: "summary_large_image" },
    ],
    links: [
      { rel: "stylesheet", href: appCss },
      { rel: "preconnect", href: "https://fonts.googleapis.com" },
      { rel: "preconnect", href: "https://fonts.gstatic.com", crossOrigin: "anonymous" },
      {
        rel: "stylesheet",
        href: "https://fonts.googleapis.com/css2?family=DM+Mono:ital,wght@0,300;0,400;0,500;1,400;1,500&family=Manrope:wght@300;400;500;600;700;800&family=JetBrains+Mono:wght@400;500;600;700&display=swap",
      },
      { rel: "icon", href: "/favicon.ico", type: "image/x-icon" },
    ],
  }),
  shellComponent: RootShell,
  component: RootComponent,
  notFoundComponent: NotFoundComponent,
  errorComponent: ErrorComponent,
});

function RootShell({ children }: { children: ReactNode }) {
  return (
    <html lang="en">
      <head>
        <HeadContent />
      </head>
      <body>
        {children}
        <Scripts />
      </body>
    </html>
  );
}

function RootComponent() {
  const { queryClient } = Route.useRouteContext();

  return (
    <QueryClientProvider client={queryClient}>
      <SessionProvider>
        <AppShell>
          {/* Required: nested routes render here. Removing <Outlet /> breaks all child routes. */}
          <Outlet />
        </AppShell>
      </SessionProvider>
    </QueryClientProvider>
  );
}
