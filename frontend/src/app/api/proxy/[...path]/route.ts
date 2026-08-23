import { NextRequest, NextResponse } from "next/server";

const BACKEND_URL = process.env.BACKEND_INTERNAL_URL || process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";
const ADMIN_API_KEY = process.env.ADMIN_API_KEY || "adm_risk_super_secret_key_98765";
const TENANT_ID = process.env.DEFAULT_TENANT_ID || "00000000-0000-0000-0000-000000000001";

export async function GET(req: NextRequest, { params }: { params: Promise<{ path: string[] }> }) {
  return handleProxy(req, await params);
}

export async function POST(req: NextRequest, { params }: { params: Promise<{ path: string[] }> }) {
  return handleProxy(req, await params);
}

export async function PUT(req: NextRequest, { params }: { params: Promise<{ path: string[] }> }) {
  return handleProxy(req, await params);
}

export async function DELETE(req: NextRequest, { params }: { params: Promise<{ path: string[] }> }) {
  return handleProxy(req, await params);
}

async function handleProxy(req: NextRequest, { path }: { path: string[] }) {
  const targetPath = path.join("/");
  const url = new URL(req.url);
  const targetUrl = `${BACKEND_URL}/${targetPath}${url.search}`;

  const headers = new Headers();
  req.headers.forEach((value, key) => {
    // Exclude host and connection headers
    if (key.toLowerCase() !== "host" && key.toLowerCase() !== "connection") {
      headers.set(key, value);
    }
  });

  // Inject server-side secure admin key and default tenant ID if not provided by browser
  if (!headers.has("X-Admin-API-Key")) {
    headers.set("X-Admin-API-Key", ADMIN_API_KEY);
  }
  if (!headers.has("X-Tenant-ID")) {
    headers.set("X-Tenant-ID", TENANT_ID);
  }

  let body: BodyInit | undefined = undefined;
  if (req.method !== "GET" && req.method !== "HEAD") {
    try {
      body = await req.text();
    } catch {
      // no body
    }
  }

  try {
    const response = await fetch(targetUrl, {
      method: req.method,
      headers,
      body,
      cache: "no-store",
    });

    const responseHeaders = new Headers();
    response.headers.forEach((value, key) => {
      responseHeaders.set(key, value);
    });

    const responseBody = await response.arrayBuffer();
    return new NextResponse(responseBody, {
      status: response.status,
      statusText: response.statusText,
      headers: responseHeaders,
    });
  } catch (err: any) {
    return NextResponse.json(
      {
        error: "backend_unavailable",
        message: `Failed to connect to local Go backend at ${BACKEND_URL}: ${err.message}`,
      },
      { status: 503 }
    );
  }
}
