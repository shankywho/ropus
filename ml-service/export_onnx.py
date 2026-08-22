"""
ONNX Conversion and Runtime Exporter for AI Risk Manager
Converts trained XGBoost models to high-performance ONNX format (opset 15)
"""

import os
import json
from datetime import datetime
import numpy as np
import joblib
import onnx
from onnxmltools import convert_xgboost
from onnxmltools.convert.common.data_types import FloatTensorType
import onnxruntime as rt

def export_canonical_model_to_onnx(
    model,
    preprocessor=None,
    onnx_path="model/fraud_model.onnx",
    n_features=15
):
    """
    Converts trained XGBoost model to ONNX format (opset 15) and verifies inference.
    """
    print(f"Converting trained {n_features}-feature model to ONNX format...")
    
    # Normalize booster feature names to positional format f0, f1, ... fN
    if hasattr(model, "get_booster"):
        model.get_booster().feature_names = [f"f{i}" for i in range(n_features)]
        
    initial_type = [('float_input', FloatTensorType([None, n_features]))]
    
    onnx_model = convert_xgboost(model, initial_types=initial_type, target_opset=15)
    
    os.makedirs(os.path.dirname(os.path.abspath(onnx_path)), exist_ok=True)
    onnx.save_model(onnx_model, onnx_path)
    print(f"Successfully saved canonical ONNX model to: {onnx_path} (Size: {os.path.getsize(onnx_path)} bytes)")
    
    # Verify with ONNX Runtime
    print("Verifying ONNX Runtime inference on 15-feature tensor...")
    sess = rt.InferenceSession(onnx_path, providers=['CPUExecutionProvider'])
    input_name = sess.get_inputs()[0].name
    
    sample_input = np.ones((1, n_features), dtype=np.float32)
    raw_preds = sess.run(None, {input_name: sample_input})
    
    print(f"ONNX Verification passed. Output shapes: {[type(o) for o in raw_preds]}")
    return onnx_path

def export_model_to_onnx(
    joblib_path="model/fraud_model.joblib",
    onnx_path="model/fraud_model.onnx",
    metadata_path="model/model_metadata.json"
):
    """Legacy wrapper for export from joblib bundle."""
    if not os.path.exists(joblib_path):
        from train import train_and_evaluate_pipeline
        train_and_evaluate_pipeline()
        return onnx_path

    bundle = joblib.load(joblib_path)
    model = bundle["model"]
    preprocessor = bundle.get("preprocessor")
    n_features = 15 if "canonical" in str(bundle.get("metadata", {}).get("model_version", "")) else 5
    return export_canonical_model_to_onnx(model, preprocessor, onnx_path=onnx_path, n_features=n_features)

if __name__ == "__main__":
    current_dir = os.path.dirname(os.path.abspath(__file__))
    model_dir = os.path.join(current_dir, "model")
    j_path = os.path.join(model_dir, "fraud_model.joblib")
    o_path = os.path.join(model_dir, "fraud_model.onnx")
    m_path = os.path.join(model_dir, "model_metadata.json")
    
    export_model_to_onnx(joblib_path=j_path, onnx_path=o_path, metadata_path=m_path)
