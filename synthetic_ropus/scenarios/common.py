"""
Shared utilities for all scenario generators.

Design notes
------------
- All IDs are synthetic, sequential, zero-padded (emp_000001, consumer_000001,
  acct_000001, ...). No real identities, no real payment data ever.
- Every scenario module receives pre-built entity pools (see `EntityPools`)
  and a shared `IdCounters` object so IDs never collide across scenarios.
- Scenario modules return plain dict rows for nodes/edges/transactions/labels
  plus a separate `scenario_tags` list. scenario_tags is written to its own
  file and is NEVER merged into the transaction/feature tables — it exists
  purely for evaluation-subset construction and dataset auditing.
"""
from __future__ import annotations

import itertools
import random
from dataclasses import dataclass, field
from datetime import datetime, timedelta
from typing import Dict, List, Optional


# ---------------------------------------------------------------------------
# ID allocation
# ---------------------------------------------------------------------------
class IdCounters:
    def __init__(self):
        self._counters: Dict[str, itertools.count] = {}

    def next(self, prefix: str) -> str:
        if prefix not in self._counters:
            self._counters[prefix] = itertools.count(1)
        n = next(self._counters[prefix])
        return f"{prefix}_{n:06d}"


# ---------------------------------------------------------------------------
# Time helpers
# ---------------------------------------------------------------------------
def parse_date(s: str) -> datetime:
    return datetime.strptime(s, "%Y-%m-%d")


def random_timestamp(rng: random.Random, start: datetime, end: datetime) -> datetime:
    delta = end - start
    total_seconds = int(delta.total_seconds())
    offset = rng.randint(0, max(total_seconds, 1))
    return start + timedelta(seconds=offset)


def business_hours_timestamp(rng: random.Random, start: datetime, end: datetime,
                              night_shift: bool = False) -> datetime:
    """Timestamp biased toward 09:00-18:00 local (or 22:00-06:00 for night shift),
    on a random weekday-biased day. Used for employee activity so 'normal'
    behavior has realistic temporal structure that anomaly detection can
    actually learn against."""
    day = start + timedelta(days=rng.randint(0, max((end - start).days, 1)))
    if night_shift:
        hour = rng.choice(list(range(22, 24)) + list(range(0, 6)))
    else:
        # slight weekday bias
        if day.weekday() >= 5 and rng.random() < 0.85:
            day += timedelta(days=(7 - day.weekday()))
        hour = rng.randint(9, 18)
    minute = rng.randint(0, 59)
    second = rng.randint(0, 59)
    return day.replace(hour=hour % 24, minute=minute, second=second, microsecond=0)


def add_ingestion_delay(rng: random.Random, event_ts: datetime) -> datetime:
    """Ingestion always strictly >= event time (pipeline latency)."""
    return event_ts + timedelta(seconds=rng.randint(1, 240))


def add_label_delay(rng: random.Random, event_ts: datetime, fraud: bool) -> datetime:
    """Models delayed label maturation. Fraud/suspicious labels take longer
    to confirm than routine/legitimate baseline labels."""
    if fraud:
        days = rng.choice([1, 2, 3, 5, 7, 10, 14, 20, 30])
        return event_ts + timedelta(days=days, hours=rng.randint(0, 23))
    else:
        # legitimate baseline labels can be near-immediate (automated) or,
        # for hard negatives that got reviewed and cleared, also delayed.
        if rng.random() < 0.9:
            return event_ts + timedelta(minutes=rng.randint(1, 90))
        return event_ts + timedelta(days=rng.randint(1, 10))


# ---------------------------------------------------------------------------
# Entity pools
# ---------------------------------------------------------------------------
@dataclass
class EntityPools:
    employees: List[dict] = field(default_factory=list)          # {id, role}
    consumers: List[str] = field(default_factory=list)
    accounts: Dict[str, str] = field(default_factory=dict)        # account_id -> consumer_id
    devices: List[dict] = field(default_factory=list)             # {id, device_type}
    ips: List[dict] = field(default_factory=list)                 # {id, ip_type}
    payment_tokens: List[str] = field(default_factory=list)
    merchants: List[dict] = field(default_factory=list)           # {id, category}

    # reserved sub-pools carved out for specific fraud mechanisms so that
    # scenario generators don't collide with each other or with the bulk
    # legitimate population.
    reserved_consumers: Dict[str, List[str]] = field(default_factory=dict)
    reserved_accounts: Dict[str, List[str]] = field(default_factory=dict)
    reserved_employees: Dict[str, List[dict]] = field(default_factory=dict)
    reserved_devices: Dict[str, List[dict]] = field(default_factory=dict)
    reserved_ips: Dict[str, List[dict]] = field(default_factory=dict)


DEVICE_TYPES = ["mobile", "desktop", "corporate_workstation", "tablet", "customer_service_terminal"]
IP_TYPES = ["residential", "corporate_nat", "mobile_carrier", "datacenter_proxy", "unknown"]
MERCHANT_CATEGORIES = ["grocery", "electronics", "travel", "subscription", "gaming",
                       "marketplace", "utilities", "restaurant", "fashion", "digital_goods"]


def build_entity_pools(rng: random.Random, cfg: dict, n_events: int, ids: IdCounters) -> EntityPools:
    scaling = cfg["entity_scaling"]
    n_consumers = max(200, int(n_events / 1000 * scaling["consumers_per_1000_events"]))
    n_employees = max(scaling["min_employees"], int(n_events / 1000 * scaling["employees_per_1000_events"]))
    n_merchants = max(scaling["min_merchants"], int(n_events / 1000 * scaling["merchants_per_1000_events"]))

    pools = EntityPools()

    # employees, assigned roles by configured distribution
    role_dist = cfg["employee_roles"]
    roles = list(role_dist.keys())
    weights = list(role_dist.values())
    for _ in range(n_employees):
        role = rng.choices(roles, weights=weights, k=1)[0]
        pools.employees.append({"id": ids.next("emp"), "role": role})

    # consumers + one-or-more accounts each
    for _ in range(n_consumers):
        cid = ids.next("consumer")
        pools.consumers.append(cid)
        n_accts = 1 if rng.random() < 0.85 else rng.randint(2, 3)
        for _ in range(n_accts):
            aid = ids.next("acct")
            pools.accounts[aid] = cid

    n_devices = int(n_consumers * 0.65)
    for _ in range(max(n_devices, 20)):
        pools.devices.append({"id": ids.next("device"), "device_type": rng.choice(DEVICE_TYPES)})

    n_ips = int(n_consumers * 0.55)
    for _ in range(max(n_ips, 20)):
        pools.ips.append({"id": ids.next("ip"), "ip_type": rng.choice(IP_TYPES)})

    n_tokens = int(len(pools.accounts) * 1.1)
    for _ in range(max(n_tokens, 20)):
        pools.payment_tokens.append(ids.next("token"))

    for _ in range(n_merchants):
        pools.merchants.append({"id": ids.next("merchant"), "category": rng.choice(MERCHANT_CATEGORIES)})

    # carve out reserved sub-pools for fraud mechanisms (non-overlapping
    # slices of the already-built pools, so "fraud entities" are a subset of
    # the same universe as legitimate entities rather than a separate
    # trivially-distinguishable population)
    def slice_pool(pool: list, frac: float, cursor: dict, key: str) -> list:
        start = cursor.get(key, 0)
        n = max(3, int(len(pool) * frac))
        end = min(start + n, len(pool))
        cursor[key] = end
        return pool[start:end]

    cursor = {}
    pools.reserved_consumers["collusion"] = slice_pool(pools.consumers, 0.01, cursor, "c")
    pools.reserved_consumers["fraud_ring"] = slice_pool(pools.consumers, 0.02, cursor, "c")
    pools.reserved_consumers["ato"] = slice_pool(pools.consumers, 0.01, cursor, "c")
    pools.reserved_consumers["card_testing"] = slice_pool(pools.consumers, 0.01, cursor, "c")
    pools.reserved_consumers["bot"] = slice_pool(pools.consumers, 0.015, cursor, "c")
    pools.reserved_consumers["sybil"] = slice_pool(pools.consumers, 0.03, cursor, "c")
    pools.reserved_consumers["poisoning"] = slice_pool(pools.consumers, 0.01, cursor, "c")

    acct_ids = list(pools.accounts.keys())
    cursor2 = {}
    pools.reserved_accounts["collusion"] = slice_pool(acct_ids, 0.01, cursor2, "a")
    pools.reserved_accounts["fraud_ring"] = slice_pool(acct_ids, 0.02, cursor2, "a")
    pools.reserved_accounts["ato"] = slice_pool(acct_ids, 0.01, cursor2, "a")
    pools.reserved_accounts["card_testing"] = slice_pool(acct_ids, 0.01, cursor2, "a")
    pools.reserved_accounts["bot"] = slice_pool(acct_ids, 0.015, cursor2, "a")
    pools.reserved_accounts["sybil"] = slice_pool(acct_ids, 0.03, cursor2, "a")
    pools.reserved_accounts["poisoning"] = slice_pool(acct_ids, 0.01, cursor2, "a")

    non_engineering = [e for e in pools.employees if e["role"] != "engineering"]
    cursor3 = {}
    pools.reserved_employees["collusion"] = slice_pool(non_engineering, 0.06, cursor3, "e")
    pools.reserved_employees["fraud_ring"] = slice_pool(non_engineering, 0.04, cursor3, "e")

    cursor4 = {}
    pools.reserved_devices["collusion"] = slice_pool(pools.devices, 0.02, cursor4, "d")
    pools.reserved_devices["sybil"] = slice_pool(pools.devices, 0.02, cursor4, "d")
    pools.reserved_devices["bot"] = slice_pool(pools.devices, 0.02, cursor4, "d")
    pools.reserved_devices["poisoning"] = slice_pool(pools.devices, 0.02, cursor4, "d")

    cursor5 = {}
    pools.reserved_ips["sybil"] = slice_pool(pools.ips, 0.02, cursor5, "i")
    pools.reserved_ips["bot"] = slice_pool(pools.ips, 0.02, cursor5, "i")
    pools.reserved_ips["poisoning"] = slice_pool(pools.ips, 0.02, cursor5, "i")

    return pools


@dataclass
class GenResult:
    """Container every scenario module returns."""
    edges: Dict[str, List[dict]] = field(default_factory=lambda: {})
    transactions: List[dict] = field(default_factory=list)
    nodes_sessions: List[dict] = field(default_factory=list)
    nodes_cases: List[dict] = field(default_factory=list)
    labels: List[dict] = field(default_factory=list)
    scenario_tags: List[dict] = field(default_factory=list)

    def add_edge(self, edge_type: str, row: dict):
        self.edges.setdefault(edge_type, []).append(row)

    def merge(self, other: "GenResult"):
        for k, v in other.edges.items():
            self.edges.setdefault(k, []).extend(v)
        self.transactions.extend(other.transactions)
        self.nodes_sessions.extend(other.nodes_sessions)
        self.nodes_cases.extend(other.nodes_cases)
        self.labels.extend(other.labels)
        self.scenario_tags.extend(other.scenario_tags)


def make_transaction(ids: IdCounters, rng: random.Random, account_id: str,
                      device_id: str, ip_id: str, token_id: str, merchant_id: str,
                      event_ts: datetime, amount: float, txn_type: str = "purchase") -> dict:
    return {
        "transaction_id": ids.next("txn"),
        "account_id": account_id,
        "device_id": device_id,
        "ip_id": ip_id,
        "payment_token_id": token_id,
        "merchant_id": merchant_id,
        "amount": round(amount, 2),
        "currency": "USD",
        "txn_type": txn_type,
        "event_timestamp": event_ts.isoformat(),
        "ingestion_timestamp": add_ingestion_delay(rng, event_ts).isoformat(),
    }


def make_label(ids: IdCounters, rng: random.Random, entity_type: str, entity_id: str,
               event_ts: datetime, ground_truth: str, cfg: dict) -> dict:
    fraud = ground_truth != "LEGITIMATE"
    label_ts = add_label_delay(rng, event_ts, fraud)
    if fraud:
        source = rng.choice(["internal_investigation", "fraud_analyst_review",
                              "automated_rule_confirmation", "consumer_dispute"])
        confidence = round(rng.uniform(0.62, 0.99), 2)
    else:
        source = rng.choice(["automated_baseline", "fraud_analyst_review"])
        confidence = round(rng.uniform(0.85, 0.99), 2)
    return {
        "label_id": ids.next("label"),
        "entity_type": entity_type,
        "entity_id": entity_id,
        "ground_truth": ground_truth,
        "label_timestamp": label_ts.isoformat(),
        "label_source": source,
        "label_confidence": confidence,
    }


def tag(ids: IdCounters, entity_type: str, entity_id: str, scenario_type: str,
        is_adversarial: bool = False, is_hard_negative: bool = False) -> dict:
    return {
        "tag_id": ids.next("tag"),
        "entity_type": entity_type,
        "entity_id": entity_id,
        "scenario_type": scenario_type,
        "is_adversarial": is_adversarial,
        "is_hard_negative": is_hard_negative,
    }
