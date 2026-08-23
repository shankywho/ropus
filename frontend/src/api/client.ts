/**
 * Base HTTP Client for ROPUS Control Plane.
 * Communicates with the local Go backend directly or via secure Next.js server proxy.
 */

export interface ApiClientConfig {
  baseUrl?: string;
  tenantId?: string;
  timeoutMs?: number;
}

export class ApiError extends Error {
  public statusCode: number;
  public details?: any;

  constructor(message: string, statusCode: number, details?: any) {
    super(message);
    this.name = "ApiError";
    this.statusCode = statusCode;
    this.details = details;
  }
}

export class HttpClient {
  private baseUrl: string;
  private tenantId: string;
  private timeoutMs: number;

  constructor(config?: ApiClientConfig) {
    if (config?.baseUrl) {
      this.baseUrl = config.baseUrl;
    } else if (typeof window !== "undefined") {
      // In browser, route via secure server-side Next.js proxy to protect secrets
      this.baseUrl = "/api/proxy";
    } else {
      // In server runtime, call backend directly
      this.baseUrl =
        process.env.BACKEND_INTERNAL_URL ||
        process.env.NEXT_PUBLIC_API_URL ||
        "http://localhost:8080";
    }

    this.tenantId =
      config?.tenantId ||
      process.env.NEXT_PUBLIC_TENANT_ID ||
      "00000000-0000-0000-0000-000000000001";
    this.timeoutMs = config?.timeoutMs || 10000;
  }

  public getBaseUrl(): string {
    return this.baseUrl;
  }

  public async request<T>(
    endpoint: string,
    options: RequestInit = {},
    customHeaders?: Record<string, string>
  ): Promise<T> {
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), this.timeoutMs);

    const cleanEndpoint = endpoint.startsWith("/") ? endpoint : "/" + endpoint;
    const url = `${this.baseUrl}${cleanEndpoint}`;

    const headers: Record<string, string> = {
      "Content-Type": "application/json",
      "X-Tenant-ID": this.tenantId,
      ...(options.headers as Record<string, string>),
      ...customHeaders,
    };

    try {
      const response = await fetch(url, {
        ...options,
        headers,
        signal: controller.signal,
        cache: "no-store",
      });

      if (!response.ok) {
        let errorMsg = `HTTP Error ${response.status} (${response.statusText})`;
        let errorDetails: any = null;

        try {
          const body = await response.json();
          errorMsg = body.message || body.error || errorMsg;
          errorDetails = body;
        } catch {
          // non-JSON response
        }

        throw new ApiError(errorMsg, response.status, errorDetails);
      }

      // Handle 204 No Content
      if (response.status === 204) {
        return {} as T;
      }

      return await response.json();
    } catch (err: any) {
      if (err.name === "AbortError") {
        throw new ApiError(`Request timeout after ${this.timeoutMs}ms`, 408);
      }
      if (err instanceof ApiError) {
        throw err;
      }
      throw new ApiError(err.message || "Network error occurred", 503);
    } finally {
      clearTimeout(timeoutId);
    }
  }
}

export const defaultHttpClient = new HttpClient();
