"""
Canonical Schema Definitions for Raw IEEE-CIS Dataset and Serving Features
"""

RAW_SCHEMA = {
    "TransactionID": "int64",
    "isFraud": "int64",
    "TransactionDT": "int64",
    "TransactionAmt": "float64",
    "ProductCD": "object",
    "card1": "int64",
    "card2": "float64",
    "card3": "float64",
    "card4": "object",
    "card5": "float64",
    "card6": "object",
    "addr1": "float64",
    "addr2": "float64",
    "dist1": "float64",
    "P_emaildomain": "object",
    "R_emaildomain": "object",
    "C1": "float64",
    "C2": "float64",
    "C13": "float64",
    "C14": "float64",
    "D1": "float64",
    "DeviceType": "object",
    "DeviceInfo": "object"
}

# The 15 canonical point-in-time safe features (Contract V1.5)
CANONICAL_15_FEATURE_COLS = [
    "amount",
    "ip_velocity_1h",
    "ip_velocity_24h",
    "token_velocity_24h",
    "device_seen_before",
    "transaction_hour",
    "transaction_day",
    "product_cd_encoded",
    "card_type_encoded",
    "card_category_encoded",
    "email_domain_risk",
    "dist1_missing",
    "device_type_mobile",
    "device_info_missing",
    "amount_to_mean_ratio"
]

CANONICAL_FEATURE_COLS = CANONICAL_15_FEATURE_COLS

# The expanded 25 canonical point-in-time safe features (Contract V2.5)
CANONICAL_25_FEATURE_COLS = [
    # Core 15 Features
    "amount",
    "ip_velocity_1h",
    "ip_velocity_24h",
    "token_velocity_24h",
    "device_seen_before",
    "transaction_hour",
    "transaction_day",
    "product_cd_encoded",
    "card_type_encoded",
    "card_category_encoded",
    "email_domain_risk",
    "dist1_missing",
    "device_type_mobile",
    "device_info_missing",
    "amount_to_mean_ratio",
    # 10 New Advanced Behavioral, Velocity, Token & Reputation Features
    "device_tx_count_5m",
    "device_tx_count_1h",
    "device_amount_sum_24h",
    "tx_acceleration_5m_1h",
    "device_amount_concentration_5m_1h",
    "device_unique_tokens_1h",
    "token_unique_devices_1h",
    "device_reputation_score",
    "device_fraud_rate",
    "device_dispute_rate"
]

# Legacy 5 features for baseline model comparison
LEGACY_FEATURE_COLS = [
    "amount",
    "ip_velocity_1h",
    "token_velocity_24h",
    "is_new_device",
    "hour_of_day"
]
