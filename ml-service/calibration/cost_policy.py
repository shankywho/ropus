"""
Cost-Sensitive Decision Policy Engine for AI Risk Manager
Evaluates expected monetary loss for ALLOW, MANUAL_REVIEW, and DECLINE actions.
Performs threshold sweeps, review capacity allocations, and multi-scenario sensitivity analyses.
"""

import os
import json
import numpy as np
import pandas as pd
from typing import Dict, Any, Tuple, List
try:
    import matplotlib
    matplotlib.use('Agg')
    import matplotlib.pyplot as plt
except ImportError:
    matplotlib = None
    plt = None

class CostSensitivePolicyEngine:
    """
    Cost-sensitive policy engine that selects the optimal risk action minimizing expected business loss:
      E[Cost(ALLOW)]   = P(fraud) * Amount * fraud_multiplier
      E[Cost(DECLINE)] = (1 - P(fraud)) * false_positive_cost
      E[Cost(REVIEW)]  = review_cost + (residual_fraud_rate * P(fraud) * Amount * fraud_multiplier)
    """
    def __init__(
        self,
        false_positive_cost: float = 500.0,
        manual_review_cost: float = 100.0,
        fraud_multiplier: float = 1.0,
        residual_review_rate: float = 0.05,
        review_capacity_pct: float = 0.10
    ):
        self.fp_cost = float(false_positive_cost)
        self.review_cost = float(manual_review_cost)
        self.fraud_multiplier = float(fraud_multiplier)
        self.residual_rate = float(residual_review_rate)
        self.review_capacity_pct = float(review_capacity_pct)

    def calculate_expected_costs(self, p_calibrated: float, amount: float) -> Dict[str, float]:
        """
        Calculates expected cost for all 3 actions for a single transaction.
        """
        p = float(np.clip(p_calibrated, 0.0, 1.0))
        amt = float(max(amount, 1.0))
        
        cost_allow = p * amt * self.fraud_multiplier
        cost_decline = (1.0 - p) * self.fp_cost
        cost_review = self.review_cost + (self.residual_rate * p * amt * self.fraud_multiplier)
        
        return {
            "ALLOW": round(cost_allow, 2),
            "MANUAL_REVIEW": round(cost_review, 2),
            "DECLINE": round(cost_decline, 2)
        }

    def select_action(self, p_calibrated: float, amount: float, allow_review: bool = True) -> Tuple[str, Dict[str, float]]:
        """
        Selects the action with the lowest expected monetary cost.
        """
        costs = self.calculate_expected_costs(p_calibrated, amount)
        if not allow_review:
            # If review capacity is exhausted, choose between ALLOW and DECLINE
            action = "ALLOW" if costs["ALLOW"] <= costs["DECLINE"] else "DECLINE"
        else:
            action = min(costs, key=costs.get)
        return action, costs

def run_cost_threshold_analysis(
    y_true: np.ndarray,
    p_calibrated: np.ndarray,
    amounts: np.ndarray,
    policy: CostSensitivePolicyEngine,
    output_csv: str,
    output_png: str
) -> pd.DataFrame:
    """
    Simulates threshold decisioning across thresholds 0.01 to 0.99 with expected monetary costs.
    """
    thresholds = np.linspace(0.01, 0.99, 99)
    rows = []
    n = len(y_true)
    
    for t in thresholds:
        # Binary threshold simulation: decline if p >= t, allow if p < t
        preds = (p_calibrated >= t).astype(int)
        
        # Fraud outcomes
        tp = int(np.sum((preds == 1) & (y_true == 1)))
        fp = int(np.sum((preds == 1) & (y_true == 0)))
        fn = int(np.sum((preds == 0) & (y_true == 1)))
        tn = int(np.sum((preds == 0) & (y_true == 0)))
        
        prec = float(tp / (tp + fp)) if (tp + fp) > 0 else 0.0
        rec = float(tp / (tp + fn)) if (tp + fn) > 0 else 0.0
        f1 = float(2 * prec * rec / (prec + rec)) if (prec + rec) > 0 else 0.0
        
        # Total Realized Cost
        # False Positives cost fp_cost; False Negatives cost actual amount * multiplier
        fn_loss = float(np.sum(amounts[(preds == 0) & (y_true == 1)] * policy.fraud_multiplier))
        fp_loss = float(fp * policy.fp_cost)
        total_cost = fn_loss + fp_loss
        avg_cost_per_txn = total_cost / n
        
        rows.append({
            "threshold": round(float(t), 2),
            "precision": round(prec, 4),
            "recall": round(rec, 4),
            "f1_score": round(f1, 4),
            "fraud_caught": tp,
            "fraud_missed": fn,
            "false_positives": fp,
            "true_negatives": tn,
            "false_negative_loss": round(fn_loss, 2),
            "false_positive_loss": round(fp_loss, 2),
            "total_realized_cost": round(total_cost, 2),
            "expected_cost_per_txn": round(avg_cost_per_txn, 2)
        })
        
    df_cost = pd.DataFrame(rows)
    os.makedirs(os.path.dirname(os.path.abspath(output_csv)), exist_ok=True)
    df_cost.to_csv(output_csv, index=False)
    print(f"Saved cost threshold analysis CSV: {output_csv}")
    
    # Plot Cost vs Threshold Curve
    if plt is not None:
        plt.figure(figsize=(9, 5))
        plt.plot(df_cost["threshold"], df_cost["expected_cost_per_txn"], label="Expected Cost per Transaction (INR)", color="crimson", lw=2)
        min_row = df_cost.loc[df_cost["expected_cost_per_txn"].idxmin()]
        opt_t = min_row["threshold"]
        min_c = min_row["expected_cost_per_txn"]
        plt.scatter([opt_t], [min_c], color="darkgreen", s=80, zorder=5, label=f"Optimal Cost Threshold ({opt_t:.2f} @ ₹{min_c:.2f})")
        plt.axvline(x=opt_t, color="darkgreen", linestyle=":")
        plt.title("Expected Business Loss vs Operating Decision Threshold")
        plt.xlabel("Calibrated Fraud Probability Threshold")
        plt.ylabel("Cost per Transaction (INR)")
        plt.grid(alpha=0.3)
        plt.legend(loc="upper right")
        plt.tight_layout()
        plt.savefig(output_png, dpi=150)
        plt.close()
        print(f"Saved cost threshold plot: {output_png}")
    
    return df_cost

def run_review_capacity_analysis(
    y_true: np.ndarray,
    p_calibrated: np.ndarray,
    amounts: np.ndarray,
    policy: CostSensitivePolicyEngine,
    output_csv: str,
    capacities: List[float] = [0.01, 0.05, 0.10, 0.20]
) -> pd.DataFrame:
    """
    Simulates operational manual review queue performance at 1%, 5%, 10%, and 20% review capacities.
    """
    n = len(y_true)
    total_fraud = int(np.sum(y_true))
    total_fraud_loss = float(np.sum(amounts[y_true == 1] * policy.fraud_multiplier))
    
    rows = []
    # Rank descending by calibrated probability
    rank_idx = np.argsort(-p_calibrated)
    
    for cap in capacities:
        k = max(1, int(n * cap))
        review_indices = rank_idx[:k]
        
        # Fraud caught in review
        reviewed_true = y_true[review_indices]
        reviewed_amt = amounts[review_indices]
        
        fraud_caught = int(np.sum(reviewed_true == 1))
        fraud_missed = int(total_fraud - fraud_caught)
        
        prec_at_k = float(fraud_caught / k) if k > 0 else 0.0
        rec_at_k = float(fraud_caught / total_fraud) if total_fraud > 0 else 0.0
        
        # Review operations cost
        review_ops_cost = float(k * policy.review_cost)
        # Residual fraud in reviewed items (5% analyst slip)
        residual_fraud_loss = float(np.sum(reviewed_amt[reviewed_true == 1]) * policy.residual_rate)
        # Unreviewed fraud loss
        unreviewed_indices = rank_idx[k:]
        unreviewed_fraud_loss = float(np.sum(amounts[unreviewed_indices][y_true[unreviewed_indices] == 1] * policy.fraud_multiplier))
        
        total_expected_cost = review_ops_cost + residual_fraud_loss + unreviewed_fraud_loss
        
        rows.append({
            "review_capacity_pct": f"{int(cap*100)}%",
            "review_queue_volume": k,
            "precision": round(prec_at_k, 4),
            "recall": round(rec_at_k, 4),
            "fraud_caught": fraud_caught,
            "fraud_missed": fraud_missed,
            "review_operational_cost": round(review_ops_cost, 2),
            "total_expected_loss": round(total_expected_cost, 2),
            "loss_reduction_pct": round((1.0 - (total_expected_cost / total_fraud_loss)) * 100, 2) if total_fraud_loss > 0 else 0.0
        })
        
    df_cap = pd.DataFrame(rows)
    os.makedirs(os.path.dirname(os.path.abspath(output_csv)), exist_ok=True)
    df_cap.to_csv(output_csv, index=False)
    print(f"Saved review capacity analysis CSV: {output_csv}")
    return df_cap

def run_cost_sensitivity_scenarios(
    y_true: np.ndarray,
    p_calibrated: np.ndarray,
    amounts: np.ndarray,
    config_path: str
) -> Dict[str, Any]:
    """
    Evaluates optimal threshold and expected loss across Scenarios A, B, C, and D.
    """
    with open(config_path, "r") as f:
        policy_cfg = json.load(f)
        
    scenarios = policy_cfg.get("scenarios", {})
    scenario_results = {}
    
    thresholds = np.linspace(0.01, 0.99, 99)
    
    for s_name, s_params in scenarios.items():
        fp_cost = float(s_params.get("false_positive_cost", 500.0))
        mult = float(s_params.get("fraud_loss_multiplier", 1.0))
        
        best_t = 0.50
        min_cost = float("inf")
        
        for t in thresholds:
            preds = (p_calibrated >= t).astype(int)
            fp = int(np.sum((preds == 1) & (y_true == 0)))
            fn_loss = float(np.sum(amounts[(preds == 0) & (y_true == 1)] * mult))
            fp_loss = float(fp * fp_cost)
            tot_cost = fn_loss + fp_loss
            
            if tot_cost < min_cost:
                min_cost = tot_cost
                best_t = float(t)
                
        scenario_results[s_name] = {
            "description": s_params.get("description"),
            "false_positive_cost": fp_cost,
            "fraud_loss_multiplier": mult,
            "optimal_threshold": round(best_t, 2),
            "min_expected_loss": round(min_cost, 2),
            "loss_per_transaction": round(min_cost / len(y_true), 2)
        }
        
    return scenario_results
