/**
 * Base HTTP Client for ROPUS Control Plane.
 * Communicates with the local Go backend on port 8080.
 */

export interface ApiClientConfig {
  baseUrl?: string;
  adminApiKey?: string;
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
  private adminApiKey: string;
  private tenantId: string;
  private timeoutMs: number;

  constructor(config?: ApiClientConfig) {
    this.baseUrl =
      config?.baseUrl ||
      process.env.NEXT_PUBLIC_API_URL ||
      process.env.VITE_API_BASE_URL ||
      "http://localhost:8080";
    this.adminApiKey =
      config?.adminApiKey ||
      process.env.NEXT_PUBLIC_ADMIN_API_KEY ||
      "adm_risk_super_secret_key_98765";
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

    const url = `${this.baseUrl}${endpoint.startsWith("/") ? endpoint : "/" + endpoint}`;

    const headers: Record<string, string> = {
      "Content-Type": "application/json",
      "X-Tenant-ID": this.tenantId,
      "X-Admin-API-Key": this.adminApiKey,
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
