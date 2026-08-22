# Global Regulatory Compliance & Standards Mapping

The AI Risk Manager platform is architected in accordance with leading international AI governance frameworks:

## 1. EU AI Act (High-Risk AI Systems Compliance)
- **Risk Management System (Article 9)**: Implemented via [`model_risk_management.go`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/backend/internal/governance/model_risk_management.go) and real-time PSI drift triggers.
- **Data Governance & Management (Article 10)**: Point-in-time training feature snapshots with deterministic SHA-256 hashes.
- **Technical Documentation & Record-Keeping (Articles 11 & 12)**: Automated Model Cards, Data Sheets, and cryptographic decision hash chains.
- **Transparency & Provision of Information (Article 13)**: Decision explanations provided for every automated block.
- **Human Oversight (Article 14)**: Manual review queues with analyst resolution loops.

## 2. NIST AI Risk Management Framework (AI RMF 1.0)
- **GOVERN**: Enterprise AI risk policies and four-eyes approval workflow.
- **MAP & MEASURE**: Continuous monitoring of ROC-AUC, calibration error, and disparate impact.
- **MANAGE**: Emergency fallback models and automated model freeze safeguards.

## 3. Federal Reserve SR 11-7 / OCC Model Risk Management
- Independent model validation before production promotion.
- Ongoing monitoring and comprehensive model inventory.
- Documented conceptual soundness and sensitivity analysis.
