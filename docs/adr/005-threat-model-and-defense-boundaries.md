Context: Track requirement enforcing strictly defense-only architecture.

Threat Vectors Identified: Model extraction, fraud oracle probing, DoS-induced fail-open, and insider collusion.

Decision & Mitigations: Closed AST engine, fail-safe degradation, merchant-only reason code boundaries, and immutable audit logs.

Residual Risks: Two-party collusion (mitigated via immutable audit logging) and image extraction (mitigated via private registry IAM).