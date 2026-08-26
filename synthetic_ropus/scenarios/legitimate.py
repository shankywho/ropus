"""
Legitimate activity, including HARD NEGATIVES: legitimate behavior that
superficially resembles fraud signals (shared IP/device, high employee
access volume, pre-transaction access, etc). These exist specifically so a
model cannot learn shortcuts like "employee+consumer relationship = fraud"
or "shared device = fraud".
"""
import random
from datetime import timedelta

from .common import (EntityPools, IdCounters, GenResult, business_hours_timestamp,
                      random_timestamp, make_transaction, make_label, tag)


def generate(rng: random.Random, pools: EntityPools, ids: IdCounters, cfg: dict,
             start, end, n_transactions: int) -> GenResult:
    res = GenResult()
    role_employees = {}
    for e in pools.employees:
        role_employees.setdefault(e["role"], []).append(e)

    accounts = list(pools.accounts.keys())

    # --- 1. bulk ordinary consumer transactions --------------------------
    n_bulk = int(n_transactions * 0.78)
    for _ in range(n_bulk):
        acct = rng.choice(accounts)
        device = rng.choice(pools.devices)
        ip = rng.choice(pools.ips)
        token = rng.choice(pools.payment_tokens)
        merchant = rng.choice(pools.merchants)
        ts = random_timestamp(rng, start, end)
        amount = round(rng.lognormvariate(3.2, 1.1), 2)
        txn = make_transaction(ids, rng, acct, device["id"], ip["id"], token, merchant["id"], ts, amount)
        res.transactions.append(txn)
        res.labels.append(make_label(ids, rng, "TRANSACTION", txn["transaction_id"], ts, "LEGITIMATE", cfg))
        res.add_edge("account_uses_device", {"account_id": acct, "device_id": device["id"], "edge_timestamp": ts.isoformat()})
        res.add_edge("account_uses_ip", {"account_id": acct, "ip_id": ip["id"], "edge_timestamp": ts.isoformat()})
        res.add_edge("account_uses_payment_token", {"account_id": acct, "payment_token_id": token, "edge_timestamp": ts.isoformat()})
        res.add_edge("account_transacts_with_merchant", {"account_id": acct, "merchant_id": merchant["id"], "edge_timestamp": ts.isoformat()})
        # occasional legitimate refund shortly after
        if rng.random() < 0.04:
            refund_ts = ts + timedelta(hours=rng.randint(1, 72))
            if refund_ts <= end:
                rtxn = make_transaction(ids, rng, acct, device["id"], ip["id"], token, merchant["id"],
                                         refund_ts, -abs(amount), txn_type="refund")
                res.transactions.append(rtxn)
                res.labels.append(make_label(ids, rng, "TRANSACTION", rtxn["transaction_id"], refund_ts, "LEGITIMATE", cfg))

    # --- 2. customer support: HIGH account-access volume (hard negative) -
    # customer_support employees legitimately touch hundreds of accounts.
    # This must NOT be treated as suspicious on its own.
    n_support_events = int(n_transactions * 0.06)
    support_staff = role_employees.get("customer_support", [])
    if support_staff:
        for _ in range(n_support_events):
            emp = rng.choice(support_staff)
            acct = rng.choice(accounts)
            ts = business_hours_timestamp(rng, start, end)
            res.add_edge("employee_accesses_account", {"employee_id": emp["id"], "account_id": acct, "edge_timestamp": ts.isoformat()})
            res.scenario_tags.append(tag(ids, "EMPLOYEE", emp["id"], "legitimate_high_volume_support", is_hard_negative=True))
            # legitimate access sometimes directly precedes a legitimate txn
            # (customer called in, support helped them complete a purchase)
            if rng.random() < 0.25:
                device = rng.choice(pools.devices)
                ip = rng.choice(pools.ips)
                token = rng.choice(pools.payment_tokens)
                merchant = rng.choice(pools.merchants)
                txn_ts = ts + timedelta(minutes=rng.randint(1, 20))
                amount = round(rng.lognormvariate(3.0, 1.0), 2)
                txn = make_transaction(ids, rng, acct, device["id"], ip["id"], token, merchant["id"], txn_ts, amount)
                res.transactions.append(txn)
                res.labels.append(make_label(ids, rng, "TRANSACTION", txn["transaction_id"], txn_ts, "LEGITIMATE", cfg))
                res.scenario_tags.append(tag(ids, "TRANSACTION", txn["transaction_id"], "legitimate_access_before_txn", is_hard_negative=True))

    # --- 3. fraud analysts repeatedly reviewing KNOWN fraud accounts -----
    # (hard negative: high concentration on "suspicious-looking" accounts,
    # but the employee behavior itself is the job, not collusion)
    analysts = role_employees.get("fraud_analyst", [])
    review_targets = rng.sample(accounts, k=min(40, len(accounts)))
    if analysts:
        for _ in range(int(n_transactions * 0.015)):
            emp = rng.choice(analysts)
            acct = rng.choice(review_targets)
            ts = business_hours_timestamp(rng, start, end)
            case_id = ids.next("case")
            res.nodes_cases.append({"case_id": case_id, "opened_at": ts.isoformat(), "status": "closed_no_action"})
            res.add_edge("employee_reviews_case", {"employee_id": emp["id"], "case_id": case_id, "edge_timestamp": ts.isoformat()})
            res.add_edge("case_references_account", {"case_id": case_id, "account_id": acct, "edge_timestamp": ts.isoformat()})
            res.labels.append(make_label(ids, rng, "CASE", case_id, ts, "LEGITIMATE", cfg))
            res.scenario_tags.append(tag(ids, "EMPLOYEE", emp["id"], "legitimate_analyst_repeated_review", is_hard_negative=True))

    # --- 4. shared corporate IP / device across many employees (hard negative)
    corp_ips = [i for i in pools.ips if i["ip_type"] == "corporate_nat"] or [rng.choice(pools.ips)]
    corp_devices = [d for d in pools.devices if d["device_type"] in
                    ("corporate_workstation", "customer_service_terminal")] or [rng.choice(pools.devices)]
    for _ in range(int(n_transactions * 0.01)):
        emp = rng.choice(pools.employees)
        ip = rng.choice(corp_ips)
        device = rng.choice(corp_devices)
        ts = business_hours_timestamp(rng, start, end, night_shift=(emp["role"] in ("risk_operations", "fraud_analyst") and rng.random() < 0.3))
        res.add_edge("employee_uses_ip", {"employee_id": emp["id"], "ip_id": ip["id"], "edge_timestamp": ts.isoformat()})
        res.add_edge("employee_uses_device", {"employee_id": emp["id"], "device_id": device["id"], "edge_timestamp": ts.isoformat()})
        res.scenario_tags.append(tag(ids, "EMPLOYEE", emp["id"], "legitimate_shared_corporate_infra", is_hard_negative=True))

    # --- 5. household device / residential NAT sharing among consumers ---
    household_accounts = rng.sample(accounts, k=min(len(accounts), max(20, len(accounts) // 15)))
    for i in range(0, len(household_accounts) - 1, 2):
        shared_device = rng.choice(pools.devices)
        shared_ip = rng.choice([i for i in pools.ips if i["ip_type"] == "residential"] or pools.ips)
        for acct in household_accounts[i:i + 2]:
            ts = random_timestamp(rng, start, end)
            res.add_edge("account_uses_device", {"account_id": acct, "device_id": shared_device["id"], "edge_timestamp": ts.isoformat()})
            res.add_edge("account_uses_ip", {"account_id": acct, "ip_id": shared_ip["id"], "edge_timestamp": ts.isoformat()})
            res.scenario_tags.append(tag(ids, "ACCOUNT", acct, "legitimate_household_sharing", is_hard_negative=True))

    # --- 6. legitimate transaction modifications by payments_operations --
    ops_staff = role_employees.get("payments_operations", [])
    if ops_staff and res.transactions:
        for _ in range(int(n_transactions * 0.01)):
            emp = rng.choice(ops_staff)
            txn = rng.choice(res.transactions)
            ts = business_hours_timestamp(rng, start, end)
            res.add_edge("employee_modifies_transaction", {"employee_id": emp["id"], "transaction_id": txn["transaction_id"], "edge_timestamp": ts.isoformat()})
            res.scenario_tags.append(tag(ids, "EMPLOYEE", emp["id"], "legitimate_transaction_modification", is_hard_negative=True))

    # --- 7. risk_operations approving large but legitimate transactions --
    risk_staff = role_employees.get("risk_operations", [])
    if risk_staff and res.transactions:
        for _ in range(int(n_transactions * 0.01)):
            emp = rng.choice(risk_staff)
            txn = rng.choice(res.transactions)
            ts = business_hours_timestamp(rng, start, end)
            res.add_edge("employee_approves_transaction", {"employee_id": emp["id"], "transaction_id": txn["transaction_id"], "edge_timestamp": ts.isoformat()})

    # --- 8. legitimate account recovery (consumer regains access) --------
    for _ in range(int(n_transactions * 0.005)):
        acct = rng.choice(accounts)
        new_device = rng.choice(pools.devices)
        new_ip = rng.choice(pools.ips)
        ts = random_timestamp(rng, start, end)
        case_id = ids.next("case")
        res.nodes_cases.append({"case_id": case_id, "opened_at": ts.isoformat(), "status": "resolved_recovery"})
        support = role_employees.get("customer_support", pools.employees)
        emp = rng.choice(support)
        res.add_edge("employee_reviews_case", {"employee_id": emp["id"], "case_id": case_id, "edge_timestamp": ts.isoformat()})
        res.add_edge("case_references_account", {"case_id": case_id, "account_id": acct, "edge_timestamp": ts.isoformat()})
        res.add_edge("account_uses_device", {"account_id": acct, "device_id": new_device["id"], "edge_timestamp": ts.isoformat()})
        res.add_edge("account_uses_ip", {"account_id": acct, "ip_id": new_ip["id"], "edge_timestamp": ts.isoformat()})
        res.labels.append(make_label(ids, rng, "CASE", case_id, ts, "LEGITIMATE", cfg))
        res.scenario_tags.append(tag(ids, "ACCOUNT", acct, "legitimate_account_recovery", is_hard_negative=True))

    # base consumer->account ownership + engineering low-volume access
    for acct, cons in pools.accounts.items():
        ts = random_timestamp(rng, start, end)
        res.add_edge("consumer_owns_account", {"consumer_id": cons, "account_id": acct, "edge_timestamp": ts.isoformat()})

    eng_staff = role_employees.get("engineering", [])
    for emp in eng_staff:
        for _ in range(rng.randint(0, 3)):  # engineering: low account-access volume
            acct = rng.choice(accounts)
            ts = business_hours_timestamp(rng, start, end)
            res.add_edge("employee_accesses_account", {"employee_id": emp["id"], "account_id": acct, "edge_timestamp": ts.isoformat()})

    return res
