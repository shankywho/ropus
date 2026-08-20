import os
import time
from typing import Optional, List, Dict, Any
from datetime import datetime
from contextlib import asynccontextmanager

import joblib
import numpy as np
import pandas as pd
from fastapi import FastAPI, HTTPException
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel, Field

# Global model state
MODEL_BUNDLE = None
FEATURE_COLS = ['amount', 'ip_velocity_1h', 'token_velocity_24h', 'is_new_device', 'hour_of_day']

def get_model_path() -> str:
    current_dir = os.path.dirname(os.path.abspath(__file__))
    default_path = os.path.join(current_dir, "model", "fraud_model.joblib")
    return os.getenv("MODEL_PATH", default_path)

def load_or_train_model():
    global MODEL_BUNDLE
    model_path = get_model_path()

    if os.path.exists(model_path):
        try:
            MODEL_BUNDLE = joblib.load(model_path)
            print(f"Loaded ML model bundle from {model_path} (ROC-AUC: {MODEL_BUNDLE.get('roc_auc', 'N/A')})")
            return
        except Exception as e:
            print(f"Failed to load model from {model_path}: {e}")

    # If model file does not exist, trigger on-the-fly training
    print("Model bundle not found. Training a fresh model on startup...")
    try:
        from train import train_and_export_model
        output_dir = os.path.dirname(model_path)
        train_and_export_model(output_dir=output_dir, model_filename=os.path.basename(model_path))
        MODEL_BUNDLE = joblib.load(model_path)
        print("Model training complete and loaded into memory.")
    except Exception as e:
        print(f"Warning: Could not train model automatically: {e}")

@asynccontextmanager
async def lifespan(app: FastAPI):
    load_or_train_model()
    yield

app = FastAPI(
    title="AI Risk Manager — ML Inference Sidecar",
    version="1.0.0",
    description="Real-time XGBoost fraud scoring & local feature attribution service",
    lifespan=lifespan
)

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

class PredictRequest(BaseModel):
    amount: float = Field(..., description="Transaction amount in local currency/cents", example=12500.0)
    ip_velocity_1h: float = Field(0.0, description="Transactions from IP in last 1 hour", example=3.0)
    token_velocity_24h: float = Field(0.0, description="Transactions for Token in last 24 hours", example=4.0)
    is_new_device: int = Field(0, description="1 if unrecognised device, 0 if known", example=1)
    hour_of_day: Optional[int] = Field(None, description="Hour of day (0-23). Defaults to current UTC hour.", example=14)

class PredictResponse(BaseModel):
    risk_score: int = Field(..., description="Calculated risk score between 0 and 100", example=78)
    probability: float = Field(..., description="Raw fraud probability (0.0 to 1.0)", example=0.7842)
    reason_codes: List[str] = Field(..., description="Top contributing heuristic and attribution reason codes")
    feature_attributions: Dict[str, float] = Field(..., description="Local feature importance scores")
    latency_ms: float = Field(..., description="Inference latency in milliseconds")

@app.get("/health")
def health():
    is_loaded = MODEL_BUNDLE is not None and "model" in MODEL_BUNDLE
    return {
        "status": "ok",
        "service": "ml-inference-sidecar",
        "model_loaded": is_loaded,
        "roc_auc": MODEL_BUNDLE.get("roc_auc") if is_loaded else None,
        "trained_at": MODEL_BUNDLE.get("trained_at") if is_loaded else None
    }

@app.post("/predict", response_model=PredictResponse)
def predict(req: PredictRequest):
    start_time = time.perf_counter()

    hour = req.hour_of_day
    if hour is None or hour < 0 or hour > 23:
        hour = datetime.utcnow().hour

    feature_dict = {
        'amount': float(req.amount),
        'ip_velocity_1h': float(req.ip_velocity_1h),
        'token_velocity_24h': float(req.token_velocity_24h),
        'is_new_device': int(req.is_new_device),
        'hour_of_day': int(hour)
    }

    df_input = pd.DataFrame([feature_dict])[FEATURE_COLS]

    # Model inference
    if MODEL_BUNDLE is not None and "model" in MODEL_BUNDLE:
        model = MODEL_BUNDLE["model"]
        proba = float(model.predict_proba(df_input)[0, 1])
    else:
        # Graceful fallback heuristic scoring if ML model failed to load
        base = 0.05
        if req.amount > 10000: base += 0.25
        if req.ip_velocity_1h > 3: base += 0.30
        if req.token_velocity_24h > 5: base += 0.25
        if req.is_new_device == 1: base += 0.15
        proba = min(base, 0.95)

    risk_score = int(round(proba * 100))

    # Calculate SHAP-like local reason codes & feature attributions
    reason_codes = []
    attributions = {}

    medians = MODEL_BUNDLE.get("feature_medians", {}) if MODEL_BUNDLE else {}
    stds = MODEL_BUNDLE.get("feature_stds", {}) if MODEL_BUNDLE else {}
    global_importances = MODEL_BUNDLE.get("feature_importances", {}) if MODEL_BUNDLE else {}

    for col in FEATURE_COLS:
        val = feature_dict[col]
        med = medians.get(col, 0.0)
        std = stds.get(col, 1.0)
        imp = global_importances.get(col, 0.2)

        # Z-score deviation scaled by feature weight
        z_score = max(0.0, (val - med) / (std if std > 0 else 1.0))
        attribution = float(round(z_score * imp, 4))
        attributions[col] = attribution

    # Generate human-readable reason codes
    if req.ip_velocity_1h >= 4:
        reason_codes.append("HIGH_IP_VELOCITY_1H")
    if req.token_velocity_24h >= 6:
        reason_codes.append("HIGH_TOKEN_VELOCITY_24H")
    if req.amount >= 15000:
        reason_codes.append("HIGH_TRANSACTION_AMOUNT")
    if req.is_new_device == 1:
        reason_codes.append("NEW_DEVICE_FINGERPRINT")
    if 1 <= hour <= 4:
        reason_codes.append("OFF_HOURS_ACTIVITY")

    if not reason_codes and risk_score < 25:
        reason_codes.append("LOW_ANOMALY_BASELINE")

    latency_ms = (time.perf_counter() - start_time) * 1000.0

    return PredictResponse(
        risk_score=risk_score,
        probability=round(proba, 4),
        reason_codes=reason_codes,
        feature_attributions=attributions,
        latency_ms=round(latency_ms, 2)
    )

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8000)
