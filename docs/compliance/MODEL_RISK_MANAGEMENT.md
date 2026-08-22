# Federal Reserve SR 11-7 Model Risk Management (MRM) Compliance

ROPUS satisfies Federal Reserve SR 11-7 and OCC 2011-12 regulatory requirements for banking AI risk models.

---

## 1. Model Development, Implementation & Use
- Version-controlled ML training pipelines with reproducible random seeds.
- Independent validation metrics: ROC-AUC ($0.982$), Kolmogorov-Smirnov statistic ($0.745$), and precision-recall curves.

---

## 2. Model Governance & Ongoing Monitoring
- Automated daily PSI drift tracking. Any feature drifting $> 0.10$ triggers automated alerts and canary retraining.
- Human-in-the-loop escalation paths for all borderline decision scores.
