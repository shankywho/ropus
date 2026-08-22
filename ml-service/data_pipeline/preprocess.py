"""
Canonical Preprocessor Module
Handles Missing Values, Strict Train-Fitted Encodings, and Feature Transformations
"""

import numpy as np
import pandas as pd
from typing import Dict, Any, List
from .schema import CANONICAL_FEATURE_COLS, CANONICAL_15_FEATURE_COLS, CANONICAL_25_FEATURE_COLS

class CanonicalPreprocessor:
    """
    Stateful preprocessor that fits encoding maps and imputation statistics
    STRICTLY on the training split, preventing any validation/test data leakage.
    Supports both 15-feature legacy and 25-feature canonical contracts.
    """
    def __init__(self, feature_contract: str = "v2.5"):
        self.is_fitted = False
        self.feature_contract = feature_contract
        self.medians: Dict[str, float] = {}
        self.product_map: Dict[str, int] = {}
        self.card_type_map: Dict[str, int] = {}
        self.card_cat_map: Dict[str, int] = {}
        self.email_domain_risk: Dict[str, float] = {}
        self.global_fraud_prior: float = 0.035
        
    def fit(self, df_train: pd.DataFrame) -> "CanonicalPreprocessor":
        """
        Fits encodings, domain risk weights, and imputation values strictly on df_train.
        """
        # 1. Product CD Ordinal Map
        unique_prods = sorted(df_train["raw_product_cd"].dropna().unique().tolist())
        self.product_map = {prod: idx for idx, prod in enumerate(unique_prods)}
        
        # 2. Card Type Ordinal Map
        unique_cards = sorted(df_train["raw_card_type"].dropna().unique().tolist())
        self.card_type_map = {card: idx for idx, card in enumerate(unique_cards)}
        
        # 3. Card Category Ordinal Map
        unique_cats = sorted(df_train["raw_card_category"].dropna().unique().tolist())
        self.card_cat_map = {cat: idx for idx, cat in enumerate(unique_cats)}
        
        # 4. Strict Train-Fitted Email Domain Risk (Smoothed Target Rate)
        if "isFraud" in df_train.columns:
            self.global_fraud_prior = float(df_train["isFraud"].mean())
            domain_stats = df_train.groupby("raw_email_domain")["isFraud"].agg(["count", "sum"])
            # Empirical Bayes / Laplace smoothing
            smoothing_weight = 20.0
            for domain, row in domain_stats.iterrows():
                cnt = row["count"]
                fraud_sum = row["sum"]
                smoothed_risk = (fraud_sum + (smoothing_weight * self.global_fraud_prior)) / (cnt + smoothing_weight)
                self.email_domain_risk[str(domain)] = float(round(smoothed_risk, 4))
        else:
            self.global_fraud_prior = 0.035
            
        # 5. Numerical Medians for Imputation (Core + Advanced)
        num_cols = [
            "amount", "ip_velocity_1h", "ip_velocity_24h", "token_velocity_24h", "amount_to_mean_ratio",
            "device_tx_count_5m", "device_tx_count_1h", "device_amount_sum_24h", "tx_acceleration_5m_1h",
            "device_amount_concentration_5m_1h", "device_unique_tokens_1h", "token_unique_devices_1h",
            "device_reputation_score", "device_fraud_rate", "device_dispute_rate"
        ]
        for col in num_cols:
            if col in df_train.columns:
                self.medians[col] = float(df_train[col].median())
            else:
                self.medians[col] = 0.0
                
        self.is_fitted = True
        return self
        
    def transform(self, df: pd.DataFrame, feature_contract: str = None) -> pd.DataFrame:
        """
        Transforms raw feature dataframe into canonical matrix with zero NaNs.
        """
        if not self.is_fitted:
            raise RuntimeError("CanonicalPreprocessor must be fitted before calling transform()")
            
        contract = feature_contract or self.feature_contract
        out = pd.DataFrame(index=df.index)
        
        # 1. Numerical with median imputation
        out["amount"] = df["amount"].fillna(self.medians.get("amount", 100.0)).astype(np.float32)
        out["ip_velocity_1h"] = df["ip_velocity_1h"].fillna(self.medians.get("ip_velocity_1h", 0.0)).astype(np.float32)
        out["ip_velocity_24h"] = df["ip_velocity_24h"].fillna(self.medians.get("ip_velocity_24h", 0.0)).astype(np.float32)
        out["token_velocity_24h"] = df["token_velocity_24h"].fillna(self.medians.get("token_velocity_24h", 0.0)).astype(np.float32)
        out["amount_to_mean_ratio"] = df["amount_to_mean_ratio"].fillna(self.medians.get("amount_to_mean_ratio", 1.0)).astype(np.float32)
        
        # 2. Binary / Discrete
        out["device_seen_before"] = df["device_seen_before"].fillna(0).astype(np.int32)
        out["transaction_hour"] = df["transaction_hour"].fillna(12).astype(np.int32)
        out["transaction_day"] = df["transaction_day"].fillna(0).astype(np.int32)
        
        # 3. Categoricals encoded with fallback for unseen values
        out["product_cd_encoded"] = df["raw_product_cd"].map(lambda x: self.product_map.get(str(x), -1)).astype(np.int32)
        out["card_type_encoded"] = df["raw_card_type"].map(lambda x: self.card_type_map.get(str(x), -1)).astype(np.int32)
        out["card_category_encoded"] = df["raw_card_category"].map(lambda x: self.card_cat_map.get(str(x), -1)).astype(np.int32)
        
        # 4. Target Risk Mapped with Prior Fallback
        out["email_domain_risk"] = df["raw_email_domain"].map(
            lambda x: self.email_domain_risk.get(str(x), self.global_fraud_prior)
        ).astype(np.float32)
        
        # 5. Missing Indicators
        out["dist1_missing"] = df["dist1_missing"].fillna(1).astype(np.int32)
        out["device_type_mobile"] = df["device_type_mobile"].fillna(0).astype(np.int32)
        out["device_info_missing"] = df["device_info_missing"].fillna(0).astype(np.int32)

        if contract in ["v1.5", "15", 15]:
            return out[CANONICAL_15_FEATURE_COLS]
            
        # 6. Advanced 10 Features (for V2.5 Contract)
        out["device_tx_count_5m"] = df["device_tx_count_5m"].fillna(self.medians.get("device_tx_count_5m", 0.0)).astype(np.float32)
        out["device_tx_count_1h"] = df["device_tx_count_1h"].fillna(self.medians.get("device_tx_count_1h", 0.0)).astype(np.float32)
        out["device_amount_sum_24h"] = df["device_amount_sum_24h"].fillna(self.medians.get("device_amount_sum_24h", 0.0)).astype(np.float32)
        out["tx_acceleration_5m_1h"] = df["tx_acceleration_5m_1h"].fillna(self.medians.get("tx_acceleration_5m_1h", 0.0)).astype(np.float32)
        out["device_amount_concentration_5m_1h"] = df["device_amount_concentration_5m_1h"].fillna(self.medians.get("device_amount_concentration_5m_1h", 0.0)).astype(np.float32)
        out["device_unique_tokens_1h"] = df["device_unique_tokens_1h"].fillna(self.medians.get("device_unique_tokens_1h", 0.0)).astype(np.float32)
        out["token_unique_devices_1h"] = df["token_unique_devices_1h"].fillna(self.medians.get("token_unique_devices_1h", 0.0)).astype(np.float32)
        out["device_reputation_score"] = df["device_reputation_score"].fillna(self.medians.get("device_reputation_score", 0.50)).astype(np.float32)
        out["device_fraud_rate"] = df["device_fraud_rate"].fillna(self.medians.get("device_fraud_rate", 0.0)).astype(np.float32)
        out["device_dispute_rate"] = df["device_dispute_rate"].fillna(self.medians.get("device_dispute_rate", 0.0)).astype(np.float32)
        
        # Ensure exact column ordering
        return out[CANONICAL_25_FEATURE_COLS]
        
    def to_dict(self) -> Dict[str, Any]:
        """Serializes fitted statistics for inference serving."""
        return {
            "is_fitted": self.is_fitted,
            "feature_contract": self.feature_contract,
            "medians": self.medians,
            "product_map": self.product_map,
            "card_type_map": self.card_type_map,
            "card_cat_map": self.card_cat_map,
            "email_domain_risk": self.email_domain_risk,
            "global_fraud_prior": self.global_fraud_prior,
            "canonical_feature_cols": CANONICAL_25_FEATURE_COLS if self.feature_contract == "v2.5" else CANONICAL_15_FEATURE_COLS
        }
        
    def from_dict(self, state: Dict[str, Any]) -> "CanonicalPreprocessor":
        """Loads fitted statistics for inference serving."""
        self.is_fitted = state.get("is_fitted", True)
        self.feature_contract = state.get("feature_contract", "v2.5")
        self.medians = state.get("medians", {})
        self.product_map = state.get("product_map", {})
        self.card_type_map = state.get("card_type_map", {})
        self.card_cat_map = state.get("card_cat_map", {})
        self.email_domain_risk = state.get("email_domain_risk", {})
        self.global_fraud_prior = state.get("global_fraud_prior", 0.035)
        return self
