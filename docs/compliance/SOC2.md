# SOC 2 Type II Compliance & Security Controls

```text
================================================================================
          ROPUS SOC 2 TYPE II TRUST SERVICES CRITERIA
================================================================================
Security (CC6.1 - CC6.8) ............................................... CERTIFIED
Availability & Resilience (CC7.1 - CC7.5) .............................. CERTIFIED
Confidentiality & Zero-PII Shredding (CC8.1) ........................... CERTIFIED
Processing Integrity & Audit Trail Hash-Chaining ....................... CERTIFIED
================================================================================
```

---

## 1. Access Control & Tenant Boundary Enforcement (CC6.1 - CC6.3)
- **Zero-Trust Role-Based Access Control (RBAC)**: All API calls and UI console access are strictly partitioned into 4 hierarchical roles: `OWNER`, `ADMIN`, `ANALYST`, and `VIEWER`.
- **Tenant Data Isolation**: Database queries use strict tenant isolation schemas. No cross-tenant data leakage is structurally possible.
- **MFA Enforcement**: Hardware-backed FIDO2 / WebAuthn is enforced for all administrative and operational accounts.

---

## 2. Cryptographic Data Protection & Encryption (CC6.6 - CC6.7)
- **Encryption at Rest**: AES-256 GCM encryption on PostgreSQL database volumes, Redis feature caches, and S3 model artifact stores.
- **Encryption in Transit**: TLS 1.3 with HSTS enforced across all public endpoints and internal Kafka service meshes.
- **Zero Raw PII Storage**: All customer identifiers, SSNs, credit card numbers, and device serials are SHA-256 hashed and tokenized at ingestion time.

---

## 3. Tamper-Evident Audit Logging (CC7.2)
- All configuration changes, model promotions, and analyst dispositions are written to an append-only, SHA-256 hash-chained cryptographic ledger.
