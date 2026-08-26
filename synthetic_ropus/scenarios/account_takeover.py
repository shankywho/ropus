"""
Credential stuffing -> account takeover -> post-takeover transaction burst.
Includes both a concentrated (one device, many accounts) and a distributed
low-rate variant so it isn't trivially detected by raw fan-out alone.
"""
import random
from datetime import timedelta

from .common import EntityPools, IdCounters, GenResult, random_timestamp, make_transaction, make_label, tag


def generate(rng: random.Random, pools: EntityPools, ids: IdCounters, cfg: dict,
             start, end, n_transactions: int) -> GenResult:
    res = GenResult()
    ato_accounts = pools.reserved_accounts.get("ato") or []
    if not ato_accounts:
        return res

    n_incidents = max(2, n_transactions // 6)
    for _ in range(n_incidents):
        distributed = rng.random() < 0.5
        attack_device = rng.choice(pools.devices)
        attack_ip = rng.choice(pools.ips)
        window_start = random_timestamp(rng, start, end - timedelta(days=1))

        n_targets = rng.randint(3, 10) if not distributed else rng.randint(1, 3)
        targets = rng.sample(ato_accounts, k=min(n_targets, len(ato_accounts)))

        stuffing_span = timedelta(minutes=rng.randint(2, 20)) if not distributed else timedelta(hours=rng.randint(6, 48))
        for acct in targets:
            attempt_ts = window_start + timedelta(seconds=rng.randint(0, int(stuffing_span.total_seconds())))
            # unsuccessful login attempts (modeled as low-value probe txns are avoided;
            # represent as an access edge if we later add SESSION nodes)
            session_id = ids.next("session")
            res.nodes_sessions.append({
                "session_id": session_id, "account_id": acct,
                "device_id": attack_device["id"], "ip_id": attack_ip["id"],
                "started_at": attempt_ts.isoformat(),
                "ended_at": (attempt_ts + timedelta(seconds=rng.randint(5, 60))).isoformat(),
            })

        # only some targets are "successfully" taken over -> post-ATO burst
        compromised = rng.sample(targets, k=max(1, len(targets) // 2))
        for acct in compromised:
            burst_start = window_start + stuffing_span + timedelta(minutes=rng.randint(1, 30))
            n_burst_txns = rng.randint(2, 5)
            for i in range(n_burst_txns):
                ts = burst_start + timedelta(minutes=i * rng.randint(1, 4))
                if ts > end:
                    break
                token = rng.choice(pools.payment_tokens)
                merchant = rng.choice(pools.merchants)
                amount = round(rng.lognormvariate(4.2, 0.8), 2)
                txn = make_transaction(ids, rng, acct, attack_device["id"], attack_ip["id"],
                                        token, merchant["id"], ts, amount)
                res.transactions.append(txn)
                res.labels.append(make_label(ids, rng, "TRANSACTION", txn["transaction_id"], ts, "ACCOUNT_TAKEOVER", cfg))
                res.scenario_tags.append(tag(ids, "TRANSACTION", txn["transaction_id"], "account_takeover"))
            res.add_edge("account_uses_device", {"account_id": acct, "device_id": attack_device["id"], "edge_timestamp": burst_start.isoformat()})
            res.add_edge("account_uses_ip", {"account_id": acct, "ip_id": attack_ip["id"], "edge_timestamp": burst_start.isoformat()})

    return res
