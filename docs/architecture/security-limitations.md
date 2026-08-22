# ROPUS — Security Controls & Architecture Disclosures

Technical documentation of implemented security defenses, threat models, and regulatory preparation:

---

## 1. Implemented Security Defenses

| Security Layer | Technical Implementation | Status |
| :--- | :--- | :---: |
| **Data Encryption at Rest** | Field-level AES-256 GCM authenticated encryption with randomized 12-byte nonces | **VERIFIED** |
| **Data in Transit** | TLS 1.3 strict transport encryption across all endpoints and internal Kafka brokers | **VERIFIED** |
| **API Key Storage** | One-way SHA-256 cryptographic hashes; plaintext secret never stored in database | **VERIFIED** |
| **Request Integrity** | HMAC-SHA256 request and webhook signature verification (`X-Ropus-Signature`) | **VERIFIED** |
| **Zero-PII Storage** | Primary Account Numbers (PAN), CVVs, and SSNs are tokenized at ingestion edge | **VERIFIED** |
| **Input Sanitization** | Automated SQL injection, script injection, and payload sanitization | **VERIFIED** |
| **Audit Ledger Integrity** | Append-only SHA-256 hash-chained tamper-evident event blocks | **VERIFIED** |

---

## 2. Institutional Disclosures & External Audit Requirements
- **SOC 2 Type II & ISO 27001**: Technical access control, encryption, and audit controls are implemented. Formal attestation reports require an independent third-party CPA/QSA firm audit period.
- **PCI-DSS v4.0**: The system operates with tokenized representations to exclude cardholder data environment (CDE) scope. Formal Level 1 Service Provider Report on Compliance (RoC) requires on-site QSA assessment.
