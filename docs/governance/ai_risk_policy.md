# Enterprise AI Risk Management & Governance Policy

## 1. Objective & Scope
This policy establishes mandatory operational governance standards for all artificial intelligence and machine learning models deployed within the AI Risk Manager / Ropus risk decision platform.

## 2. Model Risk Classification Tiers
- **Tier 1 (High Risk)**: Automated transactional blocking models capable of preventing payments or freezing customer accounts. Requires quarterly model re-validation and multi-stakeholder sign-off.
- **Tier 2 (Medium Risk)**: Step-up authentication and manual review dispatch models. Requires semi-annual audit.
- **Tier 3 (Low Risk)**: Analytical and offline risk clustering algorithms. Requires annual review.

## 3. Mandatory Governance Controls
1. **Cryptographic Provenance**: Every production model artifact must have an immutable audit trail linking back to its exact training dataset checksum and configuration hash.
2. **Deterministic Explainability**: Every high-risk score must yield human-interpretable feature contribution factors without leaking confidential customer PII.
3. **Disparate Impact & Fairness**: Models must be evaluated against the 80% Four-Fifths rule across all operational payment channels and geographies.
4. **Audit Trail Immutability**: All risk decisions must be appended to the sequential cryptographic hash chain.
