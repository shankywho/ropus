# IEEE-CIS Fraud Detection Dataset Guide

**Dataset Name:** IEEE-CIS Fraud Detection Benchmark (Kaggle / Vesta Corporation)  
**Primary Target Column:** `isFraud` (Binary: `0` = Legitimate, `1` = Fraud)  
**Time Ordering Column:** `TransactionDT` (Seconds from a given reference datetime)  

---

## 1. Dataset Source & Background

The IEEE Computational Intelligence Society (IEEE-CIS), in partnership with Vesta Corporation, released this real-world e-commerce transaction dataset containing 590,540 training transactions and 144,233 identity records with a ~3.5% overall fraud rate.

### Expected Files
When downloaded, extract the following CSV files into `ml-service/data/raw/`:
- `train_transaction.csv` (394 columns, 590,540 rows)
- `train_identity.csv` (41 columns, 144,233 rows)
- `test_transaction.csv` (393 columns, 506,691 rows)
- `test_identity.csv` (41 columns, 141,907 rows)

---

## 2. Download Instructions (Kaggle CLI)

Ensure you have accepted the competition rules on Kaggle (IEEE-CIS Fraud Detection):
```bash
# 1. Install Kaggle CLI
pip install kaggle

# 2. Configure Kaggle credentials (~/.kaggle/kaggle.json)
# 3. Download competition archive into ml-service/data/raw/
mkdir -p ml-service/data/raw
cd ml-service/data/raw
kaggle competitions download -c ieee-fraud-detection

# 4. Unzip files
unzip ieee-fraud-detection.zip
```

---

## 3. Data Schema & Core Features

- **Transaction Table (`train_transaction.csv`):**
  - `TransactionID`: Unique integer transaction identifier.
  - `isFraud`: Target label (`0` or `1`).
  - `TransactionDT`: Timedelta from a reference date (in seconds). Used for strict chronological/temporal splits.
  - `TransactionAmt`: Payment amount in USD.
  - `ProductCD`: Product code category for transaction.
  - `card1` - `card6`: Payment card properties (issuer, type, category, issuing bank).
  - `addr1`, `addr2`: Billing/shipping region and country codes.
  - `dist1`, `dist2`: Distance metrics between billing/shipping addresses.
  - `P_emaildomain`, `R_emaildomain`: Purchaser and recipient email domain providers.
  - `C1` - `C14`: Counting metrics (e.g. how many addresses/cards are associated).
  - `D1` - `D15`: Timedeltas (e.g. days between previous transactions).
  - `M1` - `M9`: Match indicators (e.g. names on card and address).
  - `V1` - `V339`: Vesta engineered behavioral and risk features.
- **Identity Table (`train_identity.csv`):**
  - `id_01` - `id_38`: Network and identity attributes (IP location, proxy indicators, browser versions).
  - `DeviceType`: Mobile / Desktop indicator.
  - `DeviceInfo`: Device model / User-Agent string.

---

## 4. Deterministic Sample Fixture

For automated continuous integration, testing, and execution in environments where the full 1.5 GB raw dataset is not downloaded, this directory includes a deterministic sample fixture:
- File: `ml-service/data/sample_ieee_fixture.csv`
- Purpose: Validates schema parsing, point-in-time feature extraction, chronological splits, and test execution deterministically without requiring external network access.
