"""
FastAPI ONNX Runtime ML Serving Sidecar with Probability Calibration & Cost Policy
Exposes Raw and Calibrated Probabilities, Expected Decision Loss, and Structured Signals
"""

import os
import time
import json
from typing import Optional, List, Dict, Any
from datetime import datetime
from contextlib import asynccontextmanager

import numpy as np
import onnxruntime as ort
from fastapi import FastAPI, HTTPException
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel, Field

# Ensure ml-service root in path
import sys
current_dir = os.path.dirname(os.path.abspath(__file__))
if current_dir not in sys.path:
    sys.path.insert(0, current_dir)

from calibration.calibrator import ModelCalibrator
from calibration.cost_policy import CostSensitivePolicyEngine

# Global ONNX session, calibrator, and policy state
ONNX_SESSION: Optional[ort.InferenceSession] = None
INPUT_NAME: str = "float_input"
NUM_INPUT_FEATURES: int = 15
METADATA: Dict[str, Any] = {}
PREPROCESSOR_STATE: Dict[str, Any] = {}
CALIBRATOR: Optional[ModelCalibrator] = None
COST_POLICY: Optional[CostSensitivePolicyEngine] = None

# Candidate Shadow Model & Calibrator State
CANDIDATE_ONNX_SESSION: Optional[ort.InferenceSession] = None
CANDIDATE_INPUT_NAME: str = "float_input"
CANDIDATE_NUM_FEATURES: int = 25
CANDIDATE_CALIBRATOR: Optional[ModelCalibrator] = None
CANDIDATE_METADATA: Dict[str, Any] = {}
CANDIDATE_LOADED: bool = False
CANDIDATE_ERROR: Optional[str] = None

# Backup / Recovery 15F Model Session & Calibrator State
FALLBACK_ONNX_SESSION: Optional[ort.InferenceSession] = None
FALLBACK_INPUT_NAME: str = "float_input"
FALLBACK_NUM_FEATURES: int = 15
FALLBACK_CALIBRATOR: Optional[ModelCalibrator] = None
FALLBACK_LOADED: bool = False
FALLBACK_ERROR: Optional[str] = None

from data_pipeline.schema import CANONICAL_25_FEATURE_COLS, CANONICAL_15_FEATURE_COLS
CANONICAL_FEATURE_COLS = CANONICAL_15_FEATURE_COLS

def get_model_paths():
    onnx_path = os.getenv("ONNX_MODEL_PATH", os.path.join(current_dir, "model", "fraud_model.onnx"))
    metadata_path = os.getenv("METADATA_PATH", os.path.join(current_dir, "model", "model_metadata.json"))
    calibration_path = os.getenv("CALIBRATION_PATH", os.path.join(current_dir, "model", "calibration.json"))
    return onnx_path, metadata_path, calibration_path

def get_candidate_paths():
    cand_dir = os.path.join(current_dir, "model", "candidates")
    cand_onnx = os.path.join(cand_dir, "fraud_model_25f_candidate.onnx")
    cand_cal = os.path.join(cand_dir, "calibration_25f_candidate.json")
    cand_meta = os.path.join(cand_dir, "metadata.json")
    return cand_onnx, cand_cal, cand_meta

def load_or_train_onnx_model():
    global ONNX_SESSION, INPUT_NAME, NUM_INPUT_FEATURES, METADATA, PREPROCESSOR_STATE, CALIBRATOR, COST_POLICY
    global CANDIDATE_ONNX_SESSION, CANDIDATE_INPUT_NAME, CANDIDATE_NUM_FEATURES, CANDIDATE_CALIBRATOR, CANDIDATE_METADATA, CANDIDATE_LOADED, CANDIDATE_ERROR
    onnx_path, metadata_path, cal_path = get_model_paths()

    # 1. If ONNX model does not exist, trigger training pipeline
    if not os.path.exists(onnx_path):
        print(f"ONNX model not found at {onnx_path}. Triggering training & calibration pipeline...")
        try:
            from train import train_and_evaluate_pipeline
            train_and_evaluate_pipeline()
        except Exception as e:
            print(f"Failed to auto-train ONNX model: {e}")

    # 2. Load Production ONNX Runtime InferenceSession (15 Features)
    if os.path.exists(onnx_path):
        try:
            opts = ort.SessionOptions()
            opts.intra_op_num_threads = 2
            opts.graph_optimization_level = ort.GraphOptimizationLevel.ORT_ENABLE_ALL

            ONNX_SESSION = ort.InferenceSession(onnx_path, sess_options=opts, providers=['CPUExecutionProvider'])
            input_shape = ONNX_SESSION.get_inputs()[0].shape
            INPUT_NAME = ONNX_SESSION.get_inputs()[0].name
            NUM_INPUT_FEATURES = input_shape[1] if len(input_shape) > 1 and isinstance(input_shape[1], int) else 15
            print(f"Successfully loaded production ONNX session from {onnx_path} (Input: '{INPUT_NAME}', Features: {NUM_INPUT_FEATURES})")
        except Exception as e:
            print(f"Error loading production ONNX session: {e}")

    # 3. Load Production Metadata & Preprocessor State
    if os.path.exists(metadata_path):
        try:
            with open(metadata_path, "r") as f:
                METADATA = json.load(f)
            PREPROCESSOR_STATE = METADATA.get("preprocessor_state", {})
            print(f"Loaded production metadata: version={METADATA.get('model_version')}")
        except Exception as e:
            print(f"Error loading metadata: {e}")

    # 4. Load Production Calibration Artifact
    if os.path.exists(cal_path):
        try:
            with open(cal_path, "r") as f:
                cal_state = json.load(f)
            CALIBRATOR = ModelCalibrator.from_dict(cal_state)
            print(f"Loaded production calibrator: method={CALIBRATOR.method}, version={CALIBRATOR.version}")
        except Exception as e:
            print(f"Error loading calibrator: {e}")
            CALIBRATOR = ModelCalibrator(method="beta")
    else:
        CALIBRATOR = ModelCalibrator(method="beta")

    # 5. Initialize Cost Policy Engine
    COST_POLICY = CostSensitivePolicyEngine(
        false_positive_cost=500.0,
        manual_review_cost=100.0,
        fraud_multiplier=1.0,
        residual_review_rate=0.05,
        review_capacity_pct=0.10
    )

    # 6. Load Candidate Shadow Model (25 Features) & Candidate Beta Calibrator
    cand_onnx, cand_cal, cand_meta = get_candidate_paths()
    if os.path.exists(cand_onnx):
        try:
            opts_cand = ort.SessionOptions()
            opts_cand.intra_op_num_threads = 2
            opts_cand.graph_optimization_level = ort.GraphOptimizationLevel.ORT_ENABLE_ALL

            CANDIDATE_ONNX_SESSION = ort.InferenceSession(cand_onnx, sess_options=opts_cand, providers=['CPUExecutionProvider'])
            cand_shape = CANDIDATE_ONNX_SESSION.get_inputs()[0].shape
            CANDIDATE_INPUT_NAME = CANDIDATE_ONNX_SESSION.get_inputs()[0].name
            CANDIDATE_NUM_FEATURES = cand_shape[1] if len(cand_shape) > 1 and isinstance(cand_shape[1], int) else 25

            if os.path.exists(cand_cal):
                with open(cand_cal, "r") as f:
                    cand_cal_state = json.load(f)
                CANDIDATE_CALIBRATOR = ModelCalibrator.from_dict(cand_cal_state)
            else:
                CANDIDATE_CALIBRATOR = ModelCalibrator(method="beta")

            if os.path.exists(cand_meta):
                with open(cand_meta, "r") as f:
                    CANDIDATE_METADATA = json.load(f)

            # Synthetic test vector validation
            test_vec = np.ones((1, 25), dtype=np.float32)
            _ = CANDIDATE_ONNX_SESSION.run(None, {CANDIDATE_INPUT_NAME: test_vec})

            CANDIDATE_LOADED = True
            CANDIDATE_ERROR = None
            print(f"Successfully loaded candidate shadow model from {cand_onnx} (Features: {CANDIDATE_NUM_FEATURES}, Calibrator: {CANDIDATE_CALIBRATOR.version})")
        except Exception as e:
            CANDIDATE_LOADED = False
            CANDIDATE_ERROR = str(e)
            print(f"Warning: Failed to load candidate shadow model: {e}")
    else:
        CANDIDATE_LOADED = False
        CANDIDATE_ERROR = "Candidate ONNX artifact not found"

    # 7. Load Backup / Recovery 15F Model
    global FALLBACK_ONNX_SESSION, FALLBACK_INPUT_NAME, FALLBACK_NUM_FEATURES, FALLBACK_CALIBRATOR, FALLBACK_LOADED, FALLBACK_ERROR
    fallback_onnx = os.path.join(current_dir, "model", "backup", "fraud_model_15f_v1.onnx")
    fallback_cal = os.path.join(current_dir, "model", "backup", "calibration_15f_v1.json")
    if os.path.exists(fallback_onnx):
        try:
            opts_fb = ort.SessionOptions()
            opts_fb.intra_op_num_threads = 1
            opts_fb.graph_optimization_level = ort.GraphOptimizationLevel.ORT_ENABLE_ALL

            FALLBACK_ONNX_SESSION = ort.InferenceSession(fallback_onnx, sess_options=opts_fb, providers=['CPUExecutionProvider'])
            FALLBACK_INPUT_NAME = FALLBACK_ONNX_SESSION.get_inputs()[0].name
            FALLBACK_NUM_FEATURES = 15

            if os.path.exists(fallback_cal):
                with open(fallback_cal, "r") as f:
                    fb_cal_state = json.load(f)
                FALLBACK_CALIBRATOR = ModelCalibrator.from_dict(fb_cal_state)
            else:
                FALLBACK_CALIBRATOR = ModelCalibrator(method="beta")

            FALLBACK_LOADED = True
            FALLBACK_ERROR = None
            print(f"Successfully loaded emergency backup 15F model from {fallback_onnx}")
        except Exception as e:
            FALLBACK_LOADED = False
            FALLBACK_ERROR = str(e)
            print(f"Warning: Failed to load backup 15F model: {e}")
    else:
        FALLBACK_LOADED = False
        FALLBACK_ERROR = "Backup 15F model artifact not found"

@asynccontextmanager
async def lifespan(app: FastAPI):
    load_or_train_onnx_model()
    yield

app = FastAPI(
    title="AI Risk Manager — Calibrated ONNX ML Inference & Policy Sidecar",
    version="2.2.0",
    description="High-performance ONNX Runtime fraud scoring with probability calibration and cost-sensitive decisioning",
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
    amount: float = Field(..., description="Transaction amount in base units", example=1250.0)
    ip_velocity_1h: float = Field(0.0, description="Transactions from IP in last 1 hour", example=3.0)
    ip_velocity_24h: Optional[float] = Field(None, description="Transactions from IP in last 24 hours", example=5.0)
    token_velocity_24h: float = Field(0.0, description="Transactions for Token in last 24 hours", example=4.0)
    is_new_device: int = Field(0, description="1 if unrecognised device, 0 if known", example=1)
    device_seen_before: Optional[int] = Field(None, description="1 if device seen before with account, 0 if novel", example=0)
    hour_of_day: Optional[int] = Field(None, description="Hour of day (0-23). Defaults to current UTC hour.", example=14)
    day_of_week: Optional[int] = Field(None, description="Day of week (0-6). Defaults to current UTC weekday.", example=2)
    product_cd: Optional[str] = Field("W", description="Product category code", example="W")
    card_type: Optional[str] = Field("visa", description="Card network (visa, mastercard)", example="visa")
    card_category: Optional[str] = Field("debit", description="Card category (debit, credit)", example="debit")
    email_domain: Optional[str] = Field("gmail.com", description="Purchaser email domain", example="gmail.com")
    dist1_missing: Optional[int] = Field(1, description="1 if address distance missing", example=1)
    device_type_mobile: Optional[int] = Field(0, description="1 if mobile device", example=0)
    device_info_missing: Optional[int] = Field(0, description="1 if device info missing", example=0)
    amount_to_mean_ratio: Optional[float] = Field(1.0, description="Ratio of amount to user historical mean", example=1.2)
    feature_contract_version: Optional[str] = Field("v1.5", description="Feature contract version", example="v1.5")

class PredictResponse(BaseModel):
    risk_score: int = Field(..., description="Presentation risk score between 0 and 100 derived from calibrated probability", example=78)
    probability: float = Field(..., description="Calibrated fraud probability (for backward compatibility)", example=0.7842)
    raw_probability: float = Field(..., description="Raw direct output from XGBoost model", example=0.8250)
    calibrated_probability: float = Field(..., description="Calibrated empirical posterior fraud probability", example=0.7842)
    expected_costs: Dict[str, float] = Field(default_factory=dict, description="Expected monetary loss for ALLOW, MANUAL_REVIEW, DECLINE")
    policy_recommended_action: str = Field("ALLOW", description="Lowest-cost optimal policy action")
    reason_codes: List[str] = Field(..., description="Structured explainability signals (MODEL_SIGNAL, RULE_SIGNAL, POLICY_SIGNAL)")
    feature_attributions: Dict[str, float] = Field(..., description="Local feature importance scores")
    latency_ms: float = Field(..., description="Inference latency in milliseconds")
    runtime: str = Field("onnxruntime+calibrated", description="Underlying inference and calibration engine")

class ShadowPredictRequest(BaseModel):
    features: Optional[List[float]] = Field(None, description="25-feature canonical float array")
    features_dict: Optional[Dict[str, float]] = Field(None, description="25-feature dictionary by canonical column name")
    evaluation_id: Optional[str] = Field(None, description="Risk evaluation ID")
    tenant_id: Optional[str] = Field(None, description="Tenant ID")
    transaction_id: Optional[str] = Field(None, description="Transaction ID")
    feature_contract_version: Optional[str] = Field("v2.5", description="Feature contract version (v2.5)")

class ShadowPredictResponse(BaseModel):
    model_version: str = Field(..., description="Candidate model version")
    feature_contract_version: str = Field("v2.5", description="Candidate feature contract version")
    feature_count: int = Field(25, description="Input feature count")
    raw_probability: float = Field(..., description="Raw output probability from candidate 25F XGBoost")
    calibrated_probability: float = Field(..., description="Beta calibrated probability from candidate calibrator")
    risk_score: int = Field(..., description="Derived integer risk score in [0, 100]")
    shadow_decision: str = Field(..., description="Shadow risk action (ALLOW_RECOMMENDATION, MANUAL_REVIEW, DECLINE_RECOMMENDATION)")
    latency_ms: float = Field(..., description="Shadow scoring inference latency in milliseconds")
    runtime: str = Field("onnxruntime+candidate_beta_calibrated", description="Underlying runtime engine")

@app.get("/health")
def health():
    is_loaded = ONNX_SESSION is not None
    return {
        "status": "ok",
        "service": "calibrated-onnx-ml-sidecar",
        "engine": "ONNX Runtime",
        "model_loaded": is_loaded,
        "model_version": METADATA.get("model_version", "fraud-xgb-25f-v3.0"),
        "feature_contract": "fraud-risk-25f-v2.5" if NUM_INPUT_FEATURES == 25 else "fraud-risk-15f-v1.5",
        "features_count": NUM_INPUT_FEATURES,
        "calibration_method": CALIBRATOR.method if CALIBRATOR else "none",
        "calibration_version": CALIBRATOR.version if CALIBRATOR else "none",
        "fallback_recovery_model": {
            "loaded": FALLBACK_LOADED,
            "error": FALLBACK_ERROR,
            "version": "fraud-xgb-15f-v1.5",
            "features_count": FALLBACK_NUM_FEATURES
        },
        "shadow_scoring": {
            "candidate_loaded": CANDIDATE_LOADED,
            "candidate_error": CANDIDATE_ERROR,
            "model_version": CANDIDATE_METADATA.get("model_version", "unknown"),
            "calibrator_version": CANDIDATE_CALIBRATOR.version if CANDIDATE_CALIBRATOR else "none",
            "features_count": CANDIDATE_NUM_FEATURES
        }
    }

@app.get("/model/candidate")
def get_candidate_info():
    candidate_meta_path = os.path.join(current_dir, "model", "candidates", "metadata.json")
    if not os.path.exists(candidate_meta_path):
        return {
            "status": "not_found",
            "message": "No candidate model trained yet."
        }
    with open(candidate_meta_path, "r") as f:
        meta = json.load(f)
    return {
        "status": "offline_candidate_ready",
        "model_version": meta.get("model_version"),
        "feature_contract": meta.get("feature_contract"),
        "feature_count": meta.get("feature_count"),
        "is_production_active": False,
        "test_metrics_25f": meta.get("test_metrics_25f"),
        "ablation_delta": meta.get("ablation_delta"),
        "onnx_checksum": meta.get("onnx_artifact", {}).get("checksum_sha256"),
        "notes": meta.get("notes")
    }

@app.post("/predict/shadow", response_model=ShadowPredictResponse)
@app.post("/predict/candidate", response_model=ShadowPredictResponse)
def predict_shadow(req: ShadowPredictRequest):
    start_time = time.perf_counter()

    if not CANDIDATE_LOADED or CANDIDATE_ONNX_SESSION is None:
        raise HTTPException(status_code=503, detail=f"Candidate shadow model not available: {CANDIDATE_ERROR}")

    # Build 25-feature vector
    if req.features is not None:
        if len(req.features) != 25:
            raise HTTPException(status_code=400, detail=f"Expected exactly 25 features, received {len(req.features)}")
        vec = np.array([req.features], dtype=np.float32)
    elif req.features_dict is not None:
        vec_list = []
        for col in CANONICAL_25_FEATURE_COLS:
            if col not in req.features_dict:
                raise HTTPException(status_code=400, detail=f"Missing required feature in 25-feature contract: '{col}'")
            val = float(req.features_dict[col])
            vec_list.append(0.0 if np.isnan(val) else val)
        vec = np.array([vec_list], dtype=np.float32)
    else:
        raise HTTPException(status_code=400, detail="Must provide either 'features' (list) or 'features_dict'")

    # Run Candidate ONNX inference
    try:
        raw_preds = CANDIDATE_ONNX_SESSION.run(None, {CANDIDATE_INPUT_NAME: vec})
        prob_output = raw_preds[1]
        if isinstance(prob_output, list) and len(prob_output) > 0 and isinstance(prob_output[0], dict):
            p_raw = float(prob_output[0].get(1, 0.05))
        elif isinstance(prob_output, np.ndarray) and prob_output.ndim >= 2:
            p_raw = float(prob_output[0, 1])
        else:
            p_raw = float(raw_preds[0][0])
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Candidate ONNX inference failure: {e}")

    # Bounds sanitation
    if np.isnan(p_raw) or np.isinf(p_raw):
        p_raw = 0.05
    p_raw = float(np.clip(p_raw, 0.0001, 0.9999))

    # Apply Candidate Beta Calibrator
    if CANDIDATE_CALIBRATOR is not None and CANDIDATE_CALIBRATOR.is_fitted:
        p_cal = float(CANDIDATE_CALIBRATOR.predict_proba(np.array([p_raw]), method="beta")[0])
    else:
        p_cal = p_raw
    p_cal = float(np.clip(p_cal, 0.0001, 0.9999))

    risk_score = int(round(p_cal * 100))

    # Apply existing policy decision thresholds (<0.05 / 0.05-0.35 / >=0.35)
    if p_cal >= 0.35:
        shadow_decision = "DECLINE_RECOMMENDATION"
    elif p_cal >= 0.05:
        shadow_decision = "MANUAL_REVIEW"
    else:
        shadow_decision = "ALLOW_RECOMMENDATION"

    latency_ms = (time.perf_counter() - start_time) * 1000.0

    return ShadowPredictResponse(
        model_version=CANDIDATE_METADATA.get("model_version", "fraud-xgb-25f-candidate-v1"),
        feature_contract_version="v2.5",
        feature_count=25,
        raw_probability=round(p_raw, 4),
        calibrated_probability=round(p_cal, 4),
        risk_score=risk_score,
        shadow_decision=shadow_decision,
        latency_ms=round(latency_ms, 3),
        runtime="onnxruntime+candidate_beta_calibrated"
    )

@app.post("/predict", response_model=PredictResponse)
def predict(req: PredictRequest):
    start_time = time.perf_counter()

    # Probability & Input Sanitation
    amt = float(req.amount) if (not np.isnan(req.amount) and req.amount > 0) else 100.0
    hour = req.hour_of_day if (req.hour_of_day is not None and 0 <= req.hour_of_day <= 23) else datetime.utcnow().hour
    day = req.day_of_week if (req.day_of_week is not None and 0 <= req.day_of_week <= 6) else datetime.utcnow().weekday()
    
    device_seen = req.device_seen_before if req.device_seen_before is not None else (0 if req.is_new_device == 1 else 1)
    ip_24h = req.ip_velocity_24h if req.ip_velocity_24h is not None else max(req.ip_velocity_1h, req.token_velocity_24h)

    # Encode categoricals using preprocessor state maps
    prod_map = PREPROCESSOR_STATE.get("product_map", {"W": 0, "C": 1, "H": 2, "R": 3, "S": 4})
    card_map = PREPROCESSOR_STATE.get("card_type_map", {"visa": 0, "mastercard": 1, "discover": 2, "american express": 3})
    cat_map = PREPROCESSOR_STATE.get("card_cat_map", {"debit": 0, "credit": 1, "charge": 2})
    domain_risk_map = PREPROCESSOR_STATE.get("email_domain_risk", {})
    prior_risk = PREPROCESSOR_STATE.get("global_fraud_prior", 0.035)

    prod_encoded = prod_map.get(str(req.product_cd), 0)
    card_encoded = card_map.get(str(req.card_type), 0)
    cat_encoded = cat_map.get(str(req.card_category), 0)
    email_risk = domain_risk_map.get(str(req.email_domain), prior_risk)

    # Build 15 canonical feature dictionary
    canonical_dict = {
        "amount": amt,
        "ip_velocity_1h": float(req.ip_velocity_1h),
        "ip_velocity_24h": float(ip_24h),
        "token_velocity_24h": float(req.token_velocity_24h),
        "device_seen_before": int(device_seen),
        "transaction_hour": int(hour),
        "transaction_day": int(day),
        "product_cd_encoded": int(prod_encoded),
        "card_type_encoded": int(card_encoded),
        "card_category_encoded": int(cat_encoded),
        "email_domain_risk": float(email_risk),
        "dist1_missing": int(req.dist1_missing or 1),
        "device_type_mobile": int(req.device_type_mobile or 0),
        "device_info_missing": int(req.device_info_missing or 0),
        "amount_to_mean_ratio": float(req.amount_to_mean_ratio or 1.0)
    }

    # 1. ONNX Inference Execution
    if ONNX_SESSION is not None:
        try:
            if NUM_INPUT_FEATURES == 15:
                input_vector = np.array([[canonical_dict[col] for col in CANONICAL_FEATURE_COLS]], dtype=np.float32)
            else:
                input_vector = np.array([[
                    amt,
                    float(req.ip_velocity_1h),
                    float(req.token_velocity_24h),
                    int(req.is_new_device),
                    int(hour)
                ]], dtype=np.float32)

            raw_preds = ONNX_SESSION.run(None, {INPUT_NAME: input_vector})
            prob_output = raw_preds[1]
            if isinstance(prob_output, list) and len(prob_output) > 0 and isinstance(prob_output[0], dict):
                p_raw = float(prob_output[0].get(1, 0.05))
            elif isinstance(prob_output, np.ndarray) and prob_output.ndim >= 2:
                p_raw = float(prob_output[0, 1])
            else:
                p_raw = float(raw_preds[0][0])
        except Exception as e:
            print(f"ONNX inference error: {e}. Falling back to heuristic.")
            p_raw = fallback_score(req)
    else:
        p_raw = fallback_score(req)

    # 2. Probability Validation & Sanitization (Strict bounds [0.0001, 0.9999], zero NaNs)
    if np.isnan(p_raw) or np.isinf(p_raw):
        p_raw = 0.05
    p_raw = float(np.clip(p_raw, 0.0001, 0.9999))

    # 3. Probability Calibration
    if CALIBRATOR is not None and CALIBRATOR.is_fitted:
        p_calibrated = float(CALIBRATOR.predict_proba(np.array([p_raw]))[0])
    else:
        p_calibrated = p_raw
    p_calibrated = float(np.clip(p_calibrated, 0.0001, 0.9999))

    # Risk Score derived from calibrated probability
    risk_score = int(round(p_calibrated * 100))

    # 4. Cost-Sensitive Action Optimization
    if COST_POLICY is not None:
        recommended_action, expected_costs = COST_POLICY.select_action(p_calibrated, amt)
    else:
        expected_costs = {"ALLOW": 0.0, "MANUAL_REVIEW": 0.0, "DECLINE": 0.0}
        recommended_action = "ALLOW" if risk_score < 45 else ("MANUAL_REVIEW" if risk_score < 85 else "DECLINE")

    # 5. Structured Reason Codes (MODEL_SIGNAL, RULE_SIGNAL, POLICY_SIGNAL)
    reason_codes = []
    attributions = {}
    importances = METADATA.get("feature_importances", {})

    for col in CANONICAL_FEATURE_COLS:
        imp = importances.get(col, 0.05)
        val = canonical_dict[col]
        attributions[col] = round(float(imp * min(val, 10.0)), 4)

    # Model & Feature Signals
    if p_calibrated >= 0.50:
        reason_codes.append("MODEL_SIGNAL:HIGH_FRAUD_PROBABILITY")
    if req.ip_velocity_1h >= 4:
        reason_codes.append("RULE_SIGNAL:HIGH_IP_VELOCITY_1H")
    if req.token_velocity_24h >= 6:
        reason_codes.append("RULE_SIGNAL:HIGH_TOKEN_VELOCITY_24H")
    if amt >= 15000:
        reason_codes.append("RULE_SIGNAL:HIGH_TRANSACTION_AMOUNT")
    if device_seen == 0 or req.is_new_device == 1:
        reason_codes.append("MODEL_SIGNAL:NEW_DEVICE_FINGERPRINT")
    if 1 <= hour <= 4:
        reason_codes.append("MODEL_SIGNAL:OFF_HOURS_ACTIVITY")
    if email_risk > 0.08:
        reason_codes.append("MODEL_SIGNAL:SUSPICIOUS_EMAIL_DOMAIN")
    if canonical_dict["amount_to_mean_ratio"] >= 3.0:
        reason_codes.append("MODEL_SIGNAL:AMOUNT_ANOMALY_VS_HISTORY")

    # Policy Signals
    if recommended_action == "MANUAL_REVIEW":
        reason_codes.append("POLICY_SIGNAL:EXPECTED_LOSS_ABOVE_REVIEW_THRESHOLD")
    elif recommended_action == "DECLINE":
        reason_codes.append("POLICY_SIGNAL:EXPECTED_LOSS_EXCEEDS_DECLINE_THRESHOLD")

    if not reason_codes and risk_score < 25:
        reason_codes.append("MODEL_SIGNAL:LOW_ANOMALY_BASELINE")

    latency_ms = (time.perf_counter() - start_time) * 1000.0

    return PredictResponse(
        risk_score=risk_score,
        probability=round(p_calibrated, 4),
        raw_probability=round(p_raw, 4),
        calibrated_probability=round(p_calibrated, 4),
        expected_costs=expected_costs,
        policy_recommended_action=recommended_action,
        reason_codes=reason_codes,
        feature_attributions=attributions,
        latency_ms=round(latency_ms, 2),
        runtime="onnxruntime+calibrated"
    )

def fallback_score(req: PredictRequest) -> float:
    base = 0.05
    if req.amount > 10000: base += 0.25
    if req.ip_velocity_1h > 3: base += 0.30
    if req.token_velocity_24h > 5: base += 0.25
    if req.is_new_device == 1: base += 0.15
    return min(base, 0.95)

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8000)
