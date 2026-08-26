"""
Card testing: rapid small-value authorizations across many merchants using
rotating payment tokens from a small device/IP footprint. Includes
legitimate micro-transaction lookalikes (e.g. a consumer legitimately buying
several cheap digital goods in a row) mixed in as noise.
"""
import random
from datetime import timedelta

from .common import EntityPools, IdCounters, GenResult, random_timestamp, make_transaction, make_label, tag


def generate(rng: random.Random, pools: EntityPools, ids: IdCounters, cfg: dict,
             start, end, n_transactions: int) -> GenResult:
    res = GenResult()
    ct_accounts = pools.reserved_accounts.get("card_testing") or []
    if not ct_accounts:
        return res

    n_sessions = max(2, n_transactions // 8)
    for _ in range(n_sessions):
        device = rng.choice(pools.devices)
        ip = rng.choice(pools.ips)
        acct = rng.choice(ct_accounts)
        start_ts = random_timestamp(rng, start, end - timedelta(hours=1))
        n_probes = rng.randint(5, 12)
        for i in range(n_probes):
            ts = start_ts + timedelta(seconds=i * rng.randint(3, 20))  # deterministic-ish rapid cadence
            if ts > end:
                break
            token = rng.choice(pools.payment_tokens)  # rapid token rotation
            merchant = rng.choice(pools.merchants)
            amount = round(rng.uniform(0.5, 3.0), 2)  # micro-transaction
            txn = make_transaction(ids, rng, acct, device["id"], ip["id"], token, merchant["id"], ts, amount)
            res.transactions.append(txn)
            gt = "CARD_TESTING" if i >= 2 else "SUSPECTED_FRAUD"  # first couple look ambiguous alone
            res.labels.append(make_label(ids, rng, "TRANSACTION", txn["transaction_id"], ts, gt, cfg))
            res.scenario_tags.append(tag(ids, "TRANSACTION", txn["transaction_id"], "card_testing"))

    return res
