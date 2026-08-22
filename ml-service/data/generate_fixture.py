import os
import numpy as np
import pandas as pd

def generate_sample_ieee_fixture(output_path="sample_ieee_fixture.csv", n_rows=8000, random_seed=42):
    np.random.seed(random_seed)
    
    # Chronological timestamps spanning 6 months (15,552,000 seconds)
    # Generated in strictly ascending order with random inter-arrival times
    dt_intervals = np.random.exponential(scale=15552000 / n_rows, size=n_rows).astype(int)
    dt_intervals[dt_intervals < 1] = 1
    transaction_dts = np.cumsum(dt_intervals)
    
    transaction_ids = np.arange(3000000, 3000000 + n_rows)
    
    # Real-world skewed amount distribution (log-normal with long tail)
    amounts = np.round(np.random.lognormal(mean=4.2, sigma=1.1, size=n_rows), 2)
    amounts = np.clip(amounts, 1.0, 15000.0)
    
    product_cds = np.random.choice(['W', 'C', 'H', 'R', 'S'], size=n_rows, p=[0.70, 0.12, 0.08, 0.06, 0.04])
    
    # Cards and addresses
    card1_pool = np.random.randint(1000, 19000, size=500)
    card1 = np.random.choice(card1_pool, size=n_rows)
    card2 = np.random.choice([111, 170, 224, 321, 490, 555, np.nan], size=n_rows, p=[0.20, 0.20, 0.20, 0.18, 0.12, 0.08, 0.02])
    card3 = np.random.choice([150, 185, np.nan], size=n_rows, p=[0.90, 0.08, 0.02])
    card4 = np.random.choice(['visa', 'mastercard', 'discover', 'american express', np.nan], size=n_rows, p=[0.65, 0.28, 0.04, 0.02, 0.01])
    card5 = np.random.choice([102, 117, 126, 166, 226, np.nan], size=n_rows, p=[0.10, 0.10, 0.15, 0.30, 0.30, 0.05])
    card6 = np.random.choice(['debit', 'credit', np.nan], size=n_rows, p=[0.75, 0.24, 0.01])
    
    addr1 = np.random.choice([126, 204, 299, 315, 325, 441, np.nan], size=n_rows, p=[0.18, 0.18, 0.18, 0.16, 0.14, 0.10, 0.06])
    addr2 = np.random.choice([87, 60, np.nan], size=n_rows, p=[0.92, 0.02, 0.06])
    dist1 = np.random.choice([0, 5, 12, 45, 120, np.nan], size=n_rows, p=[0.25, 0.15, 0.10, 0.05, 0.05, 0.40])
    
    # Email domains
    email_pool = ['gmail.com', 'yahoo.com', 'hotmail.com', 'anonymous.com', 'aol.com', 'comcast.net', 'icloud.com', np.nan]
    p_email = np.random.choice(email_pool, size=n_rows, p=[0.40, 0.20, 0.12, 0.08, 0.05, 0.05, 0.03, 0.07])
    r_email = np.random.choice(email_pool, size=n_rows, p=[0.25, 0.15, 0.08, 0.04, 0.03, 0.03, 0.02, 0.40])
    
    # Counting metrics (C1..C14)
    c1 = np.random.poisson(lam=1.5, size=n_rows)
    c2 = np.random.poisson(lam=1.8, size=n_rows)
    c5 = np.random.poisson(lam=0.8, size=n_rows)
    c6 = np.random.poisson(lam=1.2, size=n_rows)
    c11 = np.random.poisson(lam=1.0, size=n_rows)
    c13 = np.random.poisson(lam=2.5, size=n_rows)
    c14 = np.random.poisson(lam=1.4, size=n_rows)
    
    # Timedeltas (D1..D15)
    d1 = np.random.choice([0, 1, 5, 14, 45, 180, np.nan], size=n_rows, p=[0.20, 0.15, 0.15, 0.15, 0.15, 0.10, 0.10])
    d2 = np.random.choice([0, 1, 10, 60, 200, np.nan], size=n_rows, p=[0.10, 0.10, 0.15, 0.15, 0.10, 0.40])
    d3 = np.random.choice([0, 2, 7, 30, np.nan], size=n_rows, p=[0.15, 0.15, 0.15, 0.15, 0.40])
    d4 = np.random.choice([0, 1, 14, 90, np.nan], size=n_rows, p=[0.15, 0.15, 0.15, 0.15, 0.40])
    d10 = np.random.choice([0, 3, 30, 120, np.nan], size=n_rows, p=[0.20, 0.15, 0.15, 0.15, 0.35])
    d15 = np.random.choice([0, 5, 45, 300, np.nan], size=n_rows, p=[0.20, 0.15, 0.15, 0.15, 0.35])
    
    # Identity features
    device_types = np.random.choice(['desktop', 'mobile', np.nan], size=n_rows, p=[0.35, 0.25, 0.40])
    device_infos = np.random.choice(['Windows', 'iOS Device', 'MacOS', 'SM-G960N', 'Trident/7.0', np.nan], size=n_rows, p=[0.25, 0.15, 0.10, 0.05, 0.05, 0.40])
    id_01 = np.random.choice([-5.0, -10.0, -20.0, 0.0, np.nan], size=n_rows, p=[0.10, 0.08, 0.05, 0.02, 0.75])
    id_12 = np.random.choice(['NotFound', 'Found', np.nan], size=n_rows, p=[0.20, 0.15, 0.65])
    id_30 = np.random.choice(['Android 7.0', 'iOS 11.1.2', 'Windows 10', 'Mac OS X 10_12_6', np.nan], size=n_rows, p=[0.08, 0.08, 0.12, 0.05, 0.67])
    id_31 = np.random.choice(['chrome 63.0', 'mobile safari 11.0', 'ie 11.0', 'safari 11.0', np.nan], size=n_rows, p=[0.12, 0.08, 0.05, 0.05, 0.70])
    
    # V-features (V1..V30)
    v1 = np.random.choice([1.0, np.nan], size=n_rows, p=[0.40, 0.60])
    v12 = np.random.choice([0.0, 1.0, 2.0, np.nan], size=n_rows, p=[0.60, 0.20, 0.05, 0.15])
    v29 = np.random.choice([0.0, 1.0, 2.0, np.nan], size=n_rows, p=[0.65, 0.15, 0.05, 0.15])
    v44 = np.random.choice([0.0, 1.0, np.nan], size=n_rows, p=[0.55, 0.15, 0.30])
    v75 = np.random.choice([0.0, 1.0, 2.0, np.nan], size=n_rows, p=[0.50, 0.20, 0.05, 0.25])
    v281 = np.random.choice([0.0, 1.0, 2.0, np.nan], size=n_rows, p=[0.85, 0.08, 0.02, 0.05])
    v283 = np.random.choice([0.0, 1.0, 3.0, np.nan], size=n_rows, p=[0.80, 0.12, 0.03, 0.05])
    v307 = np.random.choice([0.0, 50.0, 250.0, 1000.0, np.nan], size=n_rows, p=[0.40, 0.25, 0.15, 0.15, 0.05])
    
    # Non-linear fraud generation matching IEEE-CIS characteristics
    fraud_probs = np.zeros(n_rows)
    # Base probability
    fraud_probs += 0.015
    # Product C risk (international / digital)
    fraud_probs += np.where(product_cds == 'C', 0.08, 0.0)
    # Anonymous / rapid email domain risk
    fraud_probs += np.where(p_email == 'anonymous.com', 0.12, 0.0)
    fraud_probs += np.where((p_email == 'mail.com') | (p_email == 'protonmail.com'), 0.10, 0.0)
    # High amount + high card count
    fraud_probs += np.where((amounts > 500) & (c1 > 3), 0.15, 0.0)
    # Velocity burst (high c13/c14 and low d1)
    fraud_probs += np.where((c13 > 6) & (d1 == 0), 0.20, 0.0)
    # Mobile with unusual id_01 signal
    fraud_probs += np.where((device_types == 'mobile') & (id_01 < -10), 0.18, 0.0)
    # Off-hour nocturnal risk
    hours = ((transaction_dts % 86400) / 3600).astype(int)
    fraud_probs += np.where((hours >= 1) & (hours <= 4), 0.04, 0.0)
    
    fraud_probs = np.clip(fraud_probs, 0.005, 0.92)
    is_fraud = (np.random.uniform(0, 1, size=n_rows) < fraud_probs).astype(int)
    
    df = pd.DataFrame({
        'TransactionID': transaction_ids,
        'isFraud': is_fraud,
        'TransactionDT': transaction_dts,
        'TransactionAmt': amounts,
        'ProductCD': product_cds,
        'card1': card1,
        'card2': card2,
        'card3': card3,
        'card4': card4,
        'card5': card5,
        'card6': card6,
        'addr1': addr1,
        'addr2': addr2,
        'dist1': dist1,
        'P_emaildomain': p_email,
        'R_emaildomain': r_email,
        'C1': c1,
        'C2': c2,
        'C5': c5,
        'C6': c6,
        'C11': c11,
        'C13': c13,
        'C14': c14,
        'D1': d1,
        'D2': d2,
        'D3': d3,
        'D4': d4,
        'D10': d10,
        'D15': d15,
        'V1': v1,
        'V12': v12,
        'V29': v29,
        'V44': v44,
        'V75': v75,
        'V281': v281,
        'V283': v283,
        'V307': v307,
        'DeviceType': device_types,
        'DeviceInfo': device_infos,
        'id_01': id_01,
        'id_12': id_12,
        'id_30': id_30,
        'id_31': id_31
    })
    
    os.makedirs(os.path.dirname(os.path.abspath(output_path)), exist_ok=True)
    df.to_csv(output_path, index=False)
    print(f"Generated sample IEEE-CIS fixture: {output_path} ({len(df)} rows, {df['isFraud'].sum()} fraud / {len(df)-df['isFraud'].sum()} legit, fraud rate: {df['isFraud'].mean()*100:.2f}%)")
    return df

if __name__ == "__main__":
    current_dir = os.path.dirname(os.path.abspath(__file__))
    out_csv = os.path.join(current_dir, "sample_ieee_fixture.csv")
    generate_sample_ieee_fixture(output_path=out_csv, n_rows=8000, random_seed=42)
