"""
Adversarial / graph-poisoning scenarios: high-degree attack nodes,
relationship flooding, injected misleading-but-superficially-legitimate
edges, rotating infra. `is_adversarial=True` is recorded ONLY in the
scenario_tags side-table -- never as a column in the transaction/edge
tables themselves, so it can't leak into model features directly.
"""
import random

from .common import EntityPools, IdCounters, GenResult, random_timestamp, make_transaction, make_label, tag


def generate(rng: random.Random, pools: EntityPools, ids: IdCounters, cfg: dict,
             start, end, n_transactions: int) -> GenResult:
    res = GenResult()
    p_accounts = pools.reserved_accounts.get("poisoning") or []
    p_consumers = pools.reserved_consumers.get("poisoning") or []
    p_devices = pools.reserved_devices.get("poisoning") or pools.devices[:3]
    if not p_accounts:
        return res

    # 1. high-degree attack node: one device edge-flooded onto many accounts
    #    with NO corresponding transaction activity (pure relationship
    #    flooding, designed to distort embeddings / graph degree features)
    flood_device = rng.choice(p_devices)
    flood_targets = rng.sample(list(pools.accounts.keys()), k=min(200, len(pools.accounts)))
    for acct in flood_targets:
        ts = random_timestamp(rng, start, end)
        res.add_edge("account_uses_device", {"account_id": acct, "device_id": flood_device["id"], "edge_timestamp": ts.isoformat()})
    res.scenario_tags.append(tag(ids, "DEVICE", flood_device["id"], "graph_poisoning_flood", is_adversarial=True))

    # 2. injected misleading "legitimate-looking" relationship: attacker
    #    manufactures an employee_accesses_account edge with a plausible
    #    timestamp but no real business justification.
    if pools.employees:
        for _ in range(max(3, n_transactions // 20)):
            emp = rng.choice(pools.employees)
            acct = rng.choice(p_accounts)
            ts = random_timestamp(rng, start, end)
            res.add_edge("employee_accesses_account", {"employee_id": emp["id"], "account_id": acct, "edge_timestamp": ts.isoformat()})
            res.scenario_tags.append(tag(ids, "EMPLOYEE", emp["id"], "graph_poisoning_injected_edge", is_adversarial=True))

    # 3. rotating device/ip/token low-and-slow transactions from poisoned accounts
    for _ in range(max(3, n_transactions // 5)):
        acct = rng.choice(p_accounts)
        device = rng.choice(pools.devices)
        ip = rng.choice(pools.ips)
        token = rng.choice(pools.payment_tokens)
        merchant = rng.choice(pools.merchants)
        ts = random_timestamp(rng, start, end)
        amount = round(rng.uniform(5.0, 60.0), 2)
        txn = make_transaction(ids, rng, acct, device["id"], ip["id"], token, merchant["id"], ts, amount)
        res.transactions.append(txn)
        res.labels.append(make_label(ids, rng, "TRANSACTION", txn["transaction_id"], ts, "SUSPECTED_FRAUD", cfg))
        res.scenario_tags.append(tag(ids, "TRANSACTION", txn["transaction_id"], "graph_poisoning_low_and_slow", is_adversarial=True))

    return res
