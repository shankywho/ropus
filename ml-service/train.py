import os
import numpy as np
import pandas as pd
import joblib
from datetime import datetime
from sklearn.model_selection import train_test_split
from sklearn.metrics import roc_auc_score, classification_report
import xgboost as xgb

def generate_synthetic_data(n_samples=25000, random_seed=42):
    """
    Generates a realistic synthetic transaction dataset for fraud detection.
    Features:
      - amount: Transaction amount in base units (e.g. INR / cents)
      - ip_velocity_1h: Number of transactions from same IP in last 1 hour
      - token_velocity_24h: Number of transactions for payment token in last 24 hours
      - is_new_device: Binary flag (1 if new device fingerprint, 0 if recognized)
      - hour_of_day: Transaction hour (0 to 23)
    """
    np.random.seed(random_seed)

    # Base features for legitimate transactions
    amounts = np.random.exponential(scale=2500, size=n_samples) + 100
    ip_velocity_1h = np.random.poisson(lam=0.8, size=n_samples)
    token_velocity_24h = np.random.poisson(lam=1.5, size=n_samples)
    is_new_device = np.random.binomial(n=1, p=0.15, size=n_samples)
    hour_of_day = np.random.randint(0, 24, size=n_samples)

    df = pd.DataFrame({
        'amount': amounts,
        'ip_velocity_1h': ip_velocity_1h,
        'token_velocity_24h': token_velocity_24h,
        'is_new_device': is_new_device,
        'hour_of_day': hour_of_day
    })

    # Fraud probability calculation (latent risk score based on risk vectors)
    # 1. High velocity spikes
    velocity_risk = (df['ip_velocity_1h'] > 4) * 0.35 + (df['token_velocity_24h'] > 6) * 0.40
    # 2. Large amount anomalies with new device
    amount_device_risk = ((df['amount'] > 15000) & (df['is_new_device'] == 1)) * 0.45
    # 3. Off-peak hour burst attacks (e.g. 1 AM to 4 AM)
    night_risk = ((df['hour_of_day'] >= 1) & (df['hour_of_day'] <= 4) & (df['ip_velocity_1h'] >= 3)) * 0.50
    # 4. Small baseline noise fraud
    base_prob = 0.015

    fraud_probs = np.clip(base_prob + velocity_risk + amount_device_risk + night_risk, 0, 0.95)
    df['is_fraud'] = (np.random.rand(n_samples) < fraud_probs).astype(int)

    return df

def train_and_export_model(output_dir="model", model_filename="fraud_model.joblib"):
    print(f"[{datetime.utcnow().isoformat()}] Generating synthetic fraud dataset...")
    df = generate_synthetic_data(n_samples=30000)
    
    feature_cols = ['amount', 'ip_velocity_1h', 'token_velocity_24h', 'is_new_device', 'hour_of_day']
    X = df[feature_cols]
    y = df['is_fraud']

    fraud_count = y.sum()
    print(f"Dataset generated: {len(df)} samples, {fraud_count} fraud cases ({fraud_count/len(df)*100:.2f}% fraud rate)")

    X_train, X_test, y_train, y_test = train_test_split(
        X, y, test_size=0.20, random_state=42, stratify=y
    )

    scale_pos_weight = (len(y_train) - sum(y_train)) / sum(y_train)

    print("Training XGBoost Classifier...")
    model = xgb.XGBClassifier(
        n_estimators=120,
        max_depth=4,
        learning_rate=0.08,
        scale_pos_weight=scale_pos_weight,
        eval_metric="logloss",
        random_state=42,
        tree_method="hist"
    )

    model.fit(X_train, y_train)

    # Evaluate Model
    y_pred_proba = model.predict_proba(X_test)[:, 1]
    roc_auc = roc_auc_score(y_test, y_pred_proba)
    print(f"=========================================")
    print(f" Model Evaluation ROC-AUC: {roc_auc:.4f}")
    print(f"=========================================")

    y_pred = (y_pred_proba > 0.5).astype(int)
    print(classification_report(y_test, y_pred, target_names=["Legitimate", "Fraud"]))

    # Prepare model bundle
    os.makedirs(output_dir, exist_ok=True)
    export_path = os.path.join(output_dir, model_filename)

    model_bundle = {
        "model": model,
        "features": feature_cols,
        "feature_medians": X.median().to_dict(),
        "feature_stds": X.std().to_dict(),
        "feature_importances": dict(zip(feature_cols, model.feature_importances_.tolist())),
        "roc_auc": float(roc_auc),
        "trained_at": datetime.utcnow().isoformat()
    }

    joblib.dump(model_bundle, export_path)
    print(f"Successfully saved model bundle to: {export_path}")
    return export_path

if __name__ == "__main__":
    current_dir = os.path.dirname(os.path.abspath(__file__))
    model_dir = os.path.join(current_dir, "model")
    train_and_export_model(output_dir=model_dir)
