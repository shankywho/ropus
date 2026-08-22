const crypto = require("crypto");

/**
 * Ropus AI Risk Manager Official Node.js / JavaScript SDK.
 * Enables enterprise risk scoring, decisioning, and case intelligence.
 */
class RopusClient {
  constructor(apiKey, options = {}) {
    this.apiKey = apiKey;
    this.baseUrl = (options.baseUrl || "http://localhost:8080").replace(/\/$/, "");
    this.timeoutMs = options.timeoutMs || 2000;
  }

  async evaluateRisk(transaction) {
    const url = `${this.baseUrl}/v1/risk/evaluate`;
    const response = await fetch(url, {
      method: "POST",
      headers: {
        "Authorization": `Bearer ${this.apiKey}`,
        "Content-Type": "application/json",
        "X-Client-SDK": "ropus-js-v3.34",
      },
      body: JSON.stringify(transaction),
    });

    if (!response.ok) {
      throw new Error(`Ropus API error: ${response.status} ${response.statusText}`);
    }
    return await response.json();
  }

  async createCase(caseData) {
    const url = `${this.baseUrl}/v1/cases/create`;
    const response = await fetch(url, {
      method: "POST",
      headers: {
        "Authorization": `Bearer ${this.apiKey}`,
        "Content-Type": "application/json",
      },
      body: JSON.stringify(caseData),
    });

    if (!response.ok) {
      throw new Error(`Ropus API error: ${response.status} ${response.statusText}`);
    }
    return await response.json();
  }

  static verifyWebhookSignature(payloadString, signatureHeader, secret) {
    const hmac = crypto.createHmac("sha256", secret);
    hmac.update(payloadString);
    const expectedSig = "sha256=" + hmac.digest("hex");
    return crypto.timingSafeEqual(Buffer.from(expectedSig), Buffer.from(signatureHeader));
  }
}

module.exports = { RopusClient };
