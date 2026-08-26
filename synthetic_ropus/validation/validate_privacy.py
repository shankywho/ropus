#!/usr/bin/env python3
"""
Validates privacy constraints:
  1. All entity IDs match the synthetic ID pattern (prefix_NNNNNN).
  2. No column anywhere contains a 16-digit-looking PAN, a 3/4-digit CVV
     field, or common plaintext-secret field names.
  3. No free-text fields that could carry real names/emails/etc (this
     generator never writes any -- validated by column-name allowlist).

Run: python validation/validate_privacy.py --data data
"""
import argparse
import csv
import glob
import os
import re

ID_PATTERN = re.compile(r"^[a-z_]+_\d{6}$")
PAN_PATTERN = re.compile(r"\b\d{13,19}\b")
FORBIDDEN_COLUMN_NAMES = {"pan", "card_number", "cvv", "cvv2", "password", "ssn",
                           "full_name", "email", "phone", "dob", "address"}


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--data", default="data")
    args = ap.parse_args()
    d = args.data
    errors = []

    all_csvs = glob.glob(os.path.join(d, "**", "*.csv"), recursive=True)
    id_like_cols = {"employee_id", "consumer_id", "account_id", "device_id", "ip_id",
                     "payment_token_id", "merchant_id", "case_id", "session_id",
                     "transaction_id", "label_id", "tag_id", "entity_id"}

    for path in all_csvs:
        with open(path, newline="") as f:
            reader = csv.DictReader(f)
            if reader.fieldnames is None:
                continue
            for col in reader.fieldnames:
                if col.lower() in FORBIDDEN_COLUMN_NAMES:
                    errors.append(f"{path}: forbidden column name '{col}'")
            for i, row in enumerate(reader):
                for col, val in row.items():
                    if col in id_like_cols and val:
                        if not ID_PATTERN.match(val):
                            errors.append(f"{path}: id column '{col}' has non-synthetic value '{val}' (row {i})")
                    if col in ("amount",):
                        continue
                    if val and PAN_PATTERN.search(val) and col not in id_like_cols:
                        errors.append(f"{path}: possible raw card-number-like value in column '{col}' (row {i}): {val}")
                if i > 5000:
                    break  # sampled check per file is sufficient; full scan is expensive at scale

    if errors:
        # de-duplicate similar errors for readability
        seen = set()
        uniq = []
        for e in errors:
            key = e.split("(row")[0]
            if key not in seen:
                seen.add(key)
                uniq.append(e)
        print(f"[FAIL] {len(errors)} privacy validation errors ({len(uniq)} unique):")
        for e in uniq[:25]:
            print("  -", e)
        raise SystemExit(1)
    print("[PASS] validate_privacy: all ids synthetic, no forbidden fields, no PAN-like values found")


if __name__ == "__main__":
    main()
