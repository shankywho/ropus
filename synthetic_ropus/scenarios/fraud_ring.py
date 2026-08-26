"""
Multi-hop fraud rings: EMPLOYEE -> CONSUMER -> ACCOUNT -> DEVICE -> IP ->
PAYMENT_TOKEN -> TRANSACTION chains. Some rings mix in legitimate entities
(a genuine consumer's account briefly touched by a ring) to avoid rings
being trivially separable by "every node in the ring is flagged".
"""
import random
from datetime import timedelta

from .common import (EntityPools, IdCounters, GenResult, business_hours_timestamp,
                      random_timestamp, make_transaction, make_label, tag)


def generate(rng: random.Random, pools: EntityPools, ids: IdCounters, cfg: dict,
             start, end, n_transactions: int) -> GenResult:
    res = GenResult()
    ring_accounts = pools.reserved_accounts.get("fraud_ring") or []
    ring_consumers = pools.reserved_consumers.get("fraud_ring") or []
    ring_employees = pools.reserved_employees.get("fraud_ring") or []
    if not ring_accounts:
        return res

    n_rings = max(2, len(ring_accounts) // 4)
    txns_per_ring = max(1, n_transactions // n_rings)

    for r in range(n_rings):
        subtle = rng.random() < 0.5
        size = rng.randint(3, 6)
        members = rng.sample(ring_accounts, k=min(size, len(ring_accounts)))
        # mix in one legitimate, uninvolved account to break trivial separability
        if subtle and len(pools.accounts) > len(ring_accounts):
            legit_pool = [a for a in pools.accounts if a not in ring_accounts]
            if legit_pool:
                members.append(rng.choice(legit_pool))

        ring_device = rng.choice(pools.devices)
        ring_ip = rng.choice(pools.ips)
        ring_token = rng.choice(pools.payment_tokens)
        emp = rng.choice(ring_employees) if ring_employees and rng.random() < 0.6 else None
        consumer = rng.choice(ring_consumers) if ring_consumers else None

        case_id = ids.next("case")
        case_ts = business_hours_timestamp(rng, start, end)
        res.nodes_cases.append({"case_id": case_id, "opened_at": case_ts.isoformat(),
                                 "status": "confirmed_ring" if not subtle else "under_review"})
        if consumer:
            res.add_edge("case_references_consumer", {"case_id": case_id, "consumer_id": consumer, "edge_timestamp": case_ts.isoformat()})
        if emp:
            res.add_edge("case_references_employee", {"case_id": case_id, "employee_id": emp["id"], "edge_timestamp": case_ts.isoformat()})
            access_ts = business_hours_timestamp(rng, start, end)
            res.add_edge("employee_accesses_account", {"employee_id": emp["id"], "account_id": members[0], "edge_timestamp": access_ts.isoformat()})

        for acct in members:
            res.add_edge("case_references_account", {"case_id": case_id, "account_id": acct, "edge_timestamp": case_ts.isoformat()})

        for _ in range(txns_per_ring):
            acct = rng.choice(members)
            use_shared_infra = rng.random() < (0.7 if not subtle else 0.35)
            device = ring_device if use_shared_infra else rng.choice(pools.devices)
            ip = ring_ip if use_shared_infra else rng.choice(pools.ips)
            token = ring_token if rng.random() < (0.6 if not subtle else 0.3) else rng.choice(pools.payment_tokens)
            merchant = rng.choice(pools.merchants)
            ts = random_timestamp(rng, start, end)
            amount = round(rng.lognormvariate(4.0 if not subtle else 3.3, 1.0), 2)
            txn = make_transaction(ids, rng, acct, device["id"], ip["id"], token, merchant["id"], ts, amount)
            res.transactions.append(txn)
            gt = "FRAUD_RING" if not subtle else ("SUSPECTED_FRAUD" if rng.random() < 0.5 else "FRAUD_RING")
            res.labels.append(make_label(ids, rng, "TRANSACTION", txn["transaction_id"], ts, gt, cfg))
            res.scenario_tags.append(tag(ids, "TRANSACTION", txn["transaction_id"], "fraud_ring"))

        res.labels.append(make_label(ids, rng, "CASE", case_id, case_ts,
                                      "FRAUD_RING" if not subtle else "SUSPECTED_FRAUD", cfg))

    return res
