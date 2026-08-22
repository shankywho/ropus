"""
Canonical Data and Feature Engineering Pipeline for AI Risk Manager
IEEE-CIS Real Fraud Dataset & Production Serving Contract
"""

from .schema import RAW_SCHEMA, CANONICAL_FEATURE_COLS
from .data_loader import load_raw_dataset
from .split import temporal_train_val_test_split
from .features import extract_canonical_features
from .preprocess import CanonicalPreprocessor
from .validate import validate_pipeline_integrity

__all__ = [
    "RAW_SCHEMA",
    "CANONICAL_FEATURE_COLS",
    "load_raw_dataset",
    "temporal_train_val_test_split",
    "extract_canonical_features",
    "CanonicalPreprocessor",
    "validate_pipeline_integrity"
]
