"""
Bot activity: deterministic cadence, burst activity, rotating accounts and
payment tokens, headless/browser-like session fingerprints (modeled via
device_type + very regular inter-arrival time rather than a fake
"is_bot" flag).
"""
import random
from datetime import timedelta

from .common import EntityPools, IdCounters, GenResult, random_timestamp, make_transaction, make_label, tag


def generate(rng: random.Random, pools: EntityPools, ids: IdCounters, cfg: dict,
             start, end, n_transactions: int) -> GenResult:
    res = GenResult()
    bot_accounts = pools.reserved_accounts.get("bot") or []
    bot_devices = pools.reserved_devices.get("bot") or pools.devices[:5]
    bot_ips = pools.reserved_ips.get("bot") or pools.ips[:5]
    if not bot_accounts:
        return res

    n_waves = max(2, n_transactions // 15)
    for _ in range(n_waves):
        device = rng.choice(bot_devices)
        ip = rng.choice(bot_ips)
        start_ts = random_timestamp(rng, start, end - timedelta(hours=2))
        cadence_seconds = rng.choice([2, 3, 5])  # very regular -> low cadence CV
        n_events = rng.randint(10, 25)
        for i in range(n_events):
            ts = start_ts + timedelta(seconds=i * cadence_seconds)
            if ts > end:
                break
            acct = rng.choice(bot_accounts)  # rotating accounts
            token = rng.choice(pools.payment_tokens)  # rotating tokens
            merchant = rng.choice(pools.merchants)
            amount = round(rng.uniform(1.0, 15.0), 2)
            txn = make_transaction(ids, rng, acct, device["id"], ip["id"], token, merchant["id"], ts, amount)
            res.transactions.append(txn)
            res.labels.append(make_label(ids, rng, "TRANSACTION", txn["transaction_id"], ts, "BOT_ATTACK", cfg))
            res.scenario_tags.append(tag(ids, "TRANSACTION", txn["transaction_id"], "bot_attack"))
            session_id = ids.next("session")
            res.nodes_sessions.append({
                "session_id": session_id, "account_id": acct, "device_id": device["id"], "ip_id": ip["id"],
                "started_at": ts.isoformat(), "ended_at": (ts + timedelta(seconds=1)).isoformat(),
            })

    return res
