"""
Sybil clusters: many synthetic consumer/account identities funneling through
a small shared device/IP/token footprint. Distinguished from legitimate
household sharing (legitimate.py) by degree: sybil clusters are much larger
and combine with rapid succession + low transaction diversity.
"""
import random
from datetime import timedelta

from .common import EntityPools, IdCounters, GenResult, random_timestamp, make_transaction, make_label, tag


def generate(rng: random.Random, pools: EntityPools, ids: IdCounters, cfg: dict,
             start, end, n_transactions: int) -> GenResult:
    res = GenResult()
    sybil_accounts = pools.reserved_accounts.get("sybil") or []
    sybil_devices = pools.reserved_devices.get("sybil") or pools.devices[:5]
    sybil_ips = pools.reserved_ips.get("sybil") or pools.ips[:5]
    if not sybil_accounts:
        return res

    n_clusters = max(2, len(sybil_accounts) // 8)
    txns_per_cluster = max(1, n_transactions // n_clusters)
    for _ in range(n_clusters):
        cluster_size = rng.randint(6, 12)
        members = rng.sample(sybil_accounts, k=min(cluster_size, len(sybil_accounts)))
        device = rng.choice(sybil_devices)
        ip = rng.choice(sybil_ips)
        token = rng.choice(pools.payment_tokens)

        case_id = ids.next("case")
        case_ts = random_timestamp(rng, start, end)
        res.nodes_cases.append({"case_id": case_id, "opened_at": case_ts.isoformat(), "status": "under_review"})
        for acct in members:
            res.add_edge("case_references_account", {"case_id": case_id, "account_id": acct, "edge_timestamp": case_ts.isoformat()})
            res.add_edge("account_uses_device", {"account_id": acct, "device_id": device["id"], "edge_timestamp": case_ts.isoformat()})
            res.add_edge("account_uses_ip", {"account_id": acct, "ip_id": ip["id"], "edge_timestamp": case_ts.isoformat()})

        for _ in range(txns_per_cluster):
            acct = rng.choice(members)
            merchant = rng.choice(pools.merchants)
            ts = random_timestamp(rng, start, end)
            amount = round(rng.uniform(2.0, 40.0), 2)
            txn = make_transaction(ids, rng, acct, device["id"], ip["id"], token, merchant["id"], ts, amount)
            res.transactions.append(txn)
            res.labels.append(make_label(ids, rng, "TRANSACTION", txn["transaction_id"], ts, "FRAUD_RING", cfg))
            res.scenario_tags.append(tag(ids, "TRANSACTION", txn["transaction_id"], "sybil"))

        res.labels.append(make_label(ids, rng, "CASE", case_id, case_ts, "FRAUD_RING", cfg))

    return res
