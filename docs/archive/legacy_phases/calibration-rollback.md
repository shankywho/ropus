# Calibration Model Rollback Procedure

**Document Version:** 1.0  
**Scope:** Immediate emergency rollback from Beta Calibration to previous Isotonic Calibration  

---

## 1. Overview & Rollback Guarantee

The AI Risk Manager preserves complete backward-compatible rollback capability. The previous production Isotonic calibration artifact is stored at:

`ml-service/model/calibration.isotonic.backup.json`

The serving engine (`ml-service/serve.py`) and deserialization routine (`ModelCalibrator.from_dict()`) dynamically recognize both `"type": "beta"` and `"type": "isotonic"` payloads without requiring code modifications or backend restarts.

---

## 2. Step-by-Step Rollback Instructions

### Step 1: Restore the Backup Artifact
Copy the backup Isotonic JSON artifact over the active calibration artifact:

```bash
# On local host
cp ml-service/model/calibration.isotonic.backup.json ml-service/model/calibration.json

# Copy to the running ML container
docker cp ml-service/model/calibration.json risk-ml-service:/app/model/calibration.json
```

### Step 2: Restart the ML Service Sidecar
Restart only the ML sidecar container to reload the calibration parameters into memory:

```bash
docker compose restart ml-service
```

### Step 3: Verify Successful Rollback
Check the health endpoint to verify `calibration_method` is restored to `isotonic`:

```bash
curl -s http://localhost:8000/health
```

Expected Response:
```json
{
  "status": "ok",
  "service": "calibrated-onnx-ml-sidecar",
  "engine": "ONNX Runtime",
  "model_loaded": true,
  "model_version": "xgb-ieee-canonical-v2-calibrated",
  "calibration_method": "isotonic",
  "features_count": 15
}
```

---

## 3. Rollback Safety & Verification
- **Go API Impact:** ZERO impact or downtime on the Go decision pipeline.
- **Data Integrity:** No database migrations or Redis state modifications are required.
