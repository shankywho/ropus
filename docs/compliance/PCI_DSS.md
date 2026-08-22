# PCI-DSS v4.0 Compliance & Cardholder Data Protection

ROPUS operates as a Level 1 Service Provider under PCI-DSS v4.0.

---

## 1. Cardholder Data Environment (CDE) Isolation
- **Scope Exclusion through Tokenization**: ROPUS does not ingest or store Primary Account Numbers (PAN), CVVs, or PIN blocks.
- **Card Hash Ingestion**: Only SHA-256 tokenized card fingerprint representations (`fp_card_8a19...`) and BIN/IIN routing prefixes (first 6 / 8 digits) are ingested for card testing velocity checks.

---

## 2. Secure Transmission & Key Management (Requirement 3 & 4)
- **HMAC-SHA256 Request Signing**: All developer API requests require cryptographic signature validation (`X-Ropus-Signature`).
- **Key Rotation**: Automated 90-day rotation and instant revocation for `rop_live_` API credentials.
