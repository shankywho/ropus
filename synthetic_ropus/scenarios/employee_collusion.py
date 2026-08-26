"""
Employee-consumer / employee-account collusion.

Signal is built from COMBINATIONS (repeated access + tight timing + shared
device + transaction anomaly), never from a single feature, per spec.
Some cases are obvious (many correlated signals), some are subtle (only 2-3
weaker signals) so detection isn't trivial.
"""
import random
from datetime import timedelta

from .common import (EntityPools, IdCounters, GenResult, business_hours_timestamp,
                      make_transaction, make_label, tag)


def generate(rng: random.Random, pools: EntityPools, ids: IdCounters, cfg: dict,
             start, end, n_transactions: int) -> GenResult:
    res = GenResult()
    collusive_employees = pools.reserved_employees.get("collusion") or rng.sample(pools.employees, k=min(5, len(pools.employees)))
    collusive_accounts = pools.reserved_accounts.get("collusion") or []
    if not collusive_accounts:
        return res

    n_rings = max(3, len(collusive_employees))
    txns_per_ring = max(1, n_transactions // n_rings)

    for ring_idx in range(n_rings):
        emp = rng.choice(collusive_employees)
        # a small suspicious cluster of accounts this employee repeatedly touches
        cluster_size = rng.randint(2, 5)
        cluster = rng.sample(collusive_accounts, k=min(cluster_size, len(collusive_accounts)))
        shared_device = rng.choice(pools.reserved_devices.get("collusion") or pools.devices)
        obvious = rng.random() < 0.5  # half obvious, half subtle

        for _ in range(txns_per_ring):
            acct = rng.choice(cluster)
            access_ts = business_hours_timestamp(rng, start, end)
            res.add_edge("employee_accesses_account", {"employee_id": emp["id"], "account_id": acct, "edge_timestamp": access_ts.isoformat()})

            # tight timing: transaction shortly after access
            gap_minutes = rng.randint(2, 15) if obvious else rng.randint(10, 90)
            txn_ts = access_ts + timedelta(minutes=gap_minutes)
            if txn_ts > end:
                continue
            device = shared_device if (obvious or rng.random() < 0.4) else rng.choice(pools.devices)
            ip = rng.choice(pools.ips)
            token = rng.choice(pools.payment_tokens)
            merchant = rng.choice(pools.merchants)
            amount = round(rng.lognormvariate(4.4 if obvious else 3.6, 0.9), 2)  # somewhat elevated amounts
            txn = make_transaction(ids, rng, acct, device["id"], ip["id"], token, merchant["id"], txn_ts, amount)
            res.transactions.append(txn)

            if obvious:
                # employee also approves/modifies own colluded transaction
                res.add_edge("employee_approves_transaction", {"employee_id": emp["id"], "transaction_id": txn["transaction_id"], "edge_timestamp": (txn_ts + timedelta(minutes=1)).isoformat()})
                gt = "INTERNAL_COLLUSION"
            else:
                gt = "SUSPECTED_FRAUD" if rng.random() < 0.6 else "INTERNAL_COLLUSION"

            res.labels.append(make_label(ids, rng, "TRANSACTION", txn["transaction_id"], txn_ts, gt, cfg))
            res.scenario_tags.append(tag(ids, "TRANSACTION", txn["transaction_id"], "employee_collusion",
                                          is_adversarial=False))
            res.scenario_tags.append(tag(ids, "EMPLOYEE", emp["id"], "employee_collusion"))

        case_ts = business_hours_timestamp(rng, start, end)
        case_id = ids.next("case")
        res.nodes_cases.append({"case_id": case_id, "opened_at": case_ts.isoformat(), "status": "confirmed_violation" if obvious else "under_review"})
        res.add_edge("case_references_employee", {"case_id": case_id, "employee_id": emp["id"], "edge_timestamp": case_ts.isoformat()})
        for acct in cluster:
            res.add_edge("case_references_account", {"case_id": case_id, "account_id": acct, "edge_timestamp": case_ts.isoformat()})
        res.labels.append(make_label(ids, rng, "CASE", case_id, case_ts,
                                      "INTERNAL_COLLUSION" if obvious else "SUSPECTED_FRAUD", cfg))

    return res
