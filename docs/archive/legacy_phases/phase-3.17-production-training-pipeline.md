# Phase 3.17 — Production ML Training Pipeline Integration

## 1. Overview & Architecture

Phase 3.17 replaces the controlled training mock with a production-grade asynchronous **Training Pipeline** subsystem while strictly enforcing all multi-stage safety boundaries:

```mermaid
flowchart TD
    A[Sustained Drift / Manual Trigger] --> B[Dataset Quorum & Schema Validator]
    B -->|Check SHA-256, Quorum, Labels| C[Training Runner]
    C -->|Asynchronous Isolated Execution| D[Artifact Store & Verifier]
    D -->|Verify Checksum, Dim 25, Contracts| E[Model Registry CANDIDATE State]
    E --> F[Offline Validation Gates]
    F -->|ROC-AUC, Calibration, Latency| G[Shadow Evaluation Gates]
    G -->|Scoring Agreement & Error Rate| H[Operator Approval Gate]
    H -->|POST /v1/models/candidates/{id}/approve| I[Staged Canary Rollout 0%->100%]
    I -->|Circuit Breaker Monitored| J[Atomic Production Promotion]
    I -->|Safety Breach / Trip| K[Automatic Rollback to Fallback Model]
```

---

## 2. Implemented Subsystems & Components

### 2.1 Dataset Validator (`backend/internal/riskengine/dataset_validator.go`)
- **Schema & Feature Contract Verification**: Enforces the 25-feature canonical contract (`fraud-risk-25f-v2.5`).
- **Data Quality Gates**:
  - Sample quorum enforcement ($\ge 50$ samples, default 1,000 for production).
  - Missing value rate ceiling ($\le 15\%$).
  - Quality score threshold ($\ge 0.80$).
  - Positive label verification (guarantees training data contains positive fraud labels).
  - Numerical sanity checks: guarantees vector contains zero `NaN` or `+Inf`/`-Inf` values.
  - SHA-256 dataset checksumming and file integrity verification.

### 2.2 Artifact Store & Artifact Verifier (`backend/internal/riskengine/artifact_store.go`, `artifact_verifier.go`)
- **`ArtifactStore` Interface**: Pluggable storage abstraction supporting local filesystem and object storage.
- **Immutability Guarantee**: `LocalFilesystemArtifactStore` guarantees that artifact versions are write-once and can never be overwritten once published.
- **`ArtifactVerifier`**:
  - Verifies artifact existence and non-zero byte size.
  - Computes and verifies SHA-256 cryptographic checksum against the manifest.
  - Validates 25-feature input dimensionality.
  - Executes dry-run test vector inference sanity checks before model registration.

### 2.3 Model Registry (`backend/internal/riskengine/model_registry.go`)
- **Thread-Safe Lifecycle Management**: Tracks models across explicit lifecycle states:
  - `CANDIDATE` $\to$ `VALIDATED` $\to$ `SHADOW` $\to$ `APPROVED` $\to$ `CANARY` $\to$ `PROMOTED` / `REJECTED` / `ROLLED_BACK` / `FAILED`.
- **Atomic Promotion**: Promotes candidate to active production model while preserving the previous production model as the emergency fallback.
- **Registered Baselines**:
  - Production Primary: `fraud-xgb-25f-v3.0`
  - Emergency Fallback: `fraud-xgb-15f-v1.5`

### 2.4 Training Runner Abstraction (`backend/internal/riskengine/training_runner.go`)
- **`TrainingRunner` Interface**:
  - `StartTraining(ctx, req)`
  - `GetTrainingStatus(ctx, jobID)`
  - `CancelTraining(ctx, jobID)`
  - `ValidateDataset(ctx, metadata)`
- **`LocalProcessTrainingAdapter`**:
  - Executes isolated OS process (`exec.CommandContext`).
  - Strict timeout enforcement and bounded log buffers (64 KB ring cap).
  - Process cancellation and exit code capturing.
  - Passes dataset path, output artifact directory, and candidate version parameters.
- **`FixtureTrainingAdapter`**: Fast deterministic training runner for integration testing.

### 2.5 API Endpoints

| Endpoint | Method | Auth | Description |
|---|---|---|---|
| `/v1/models/registry` | `GET` | Public | Returns collection of all models in the registry |
| `/v1/models/production` | `GET` | Public | Returns current primary production model and fallback model |
| `/v1/models/{version}` | `GET` | Public | Returns metadata for a specific model version |
| `/v1/models/candidates` | `GET` | Public | Lists all active candidate models and validation scores |
| `/v1/models/candidates/{id}` | `GET` | Public | Gets candidate details and artifact checksum |
| `/v1/models/candidates/{id}/validation` | `GET` | Public | Returns candidate offline validation scorecard |
| `/v1/models/candidates/{id}/shadow` | `GET` | Public | Returns candidate shadow evaluation scorecard |
| `/v1/models/candidates/{id}/approve` | `POST` | Admin | Operator approval triggering staged canary rollout |
| `/v1/models/candidates/{id}/reject` | `POST` | Admin | Operator rejection of candidate |
| `/v1/retraining/jobs/{id}` | `GET` | Public | Returns retraining job state and execution duration |
| `/v1/retraining/jobs/{id}/logs` | `GET` | Public | Returns captured training stdout/stderr logs |
| `/v1/retraining/jobs/{id}/cancel` | `POST` | Admin | Cancels an in-flight retraining job |
| `/v1/retraining/trigger` | `POST` | Admin | Triggers manual retraining with audit tracking |

---

## 3. Verification & Test Evidence

- **Unit & Integration Tests**:
  - `go test -v ./internal/riskengine` $\to$ **PASS (100%)**
- **Data Race Detector**:
  - `go test -race -count=1 -timeout=120s ./...` $\to$ **PASS (0 races across all packages)**
- **Benchmarks**:
  - `BenchmarkModelRegistry_GetModel`: 15.8M ops/sec (84.95 ns/op)
  - `BenchmarkDatasetValidator`: 7.6M ops/sec (202.9 ns/op)
  - `BenchmarkRetrainingCoordinator_GetStatus`: 3.9M ops/sec (314.6 ns/op)
- **Live Docker Verification**:
  - Triggered job `job_retrain_1787319277695863634`.
  - Offline validation and shadow evaluation passed.
  - Job paused at `AWAITING_APPROVAL`.
  - Operator approved candidate `model_cand_20260821133437`.
  - Staged canary rollout ($1\% \to 5\% \to 10\% \to 25\% \to 50\% \to 100\%$) completed without circuit breaker tripping.
  - Model promoted in Model Registry to active primary production model.
