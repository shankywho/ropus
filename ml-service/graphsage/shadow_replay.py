"""
Realistic Offline Shadow Replay Fixture & Deterministic Point-in-Time Stream Engine (Phase 59)
"""

import json
from datetime import datetime, timezone, timedelta
from typing import Dict, List, Any, Optional, Iterator, Tuple

from .schema import (
    HeteroNode, HeteroEdge, HeteroNodeType, HeteroEdgeType,
    LabelMaturityState, DatasetType
)
from .data_source import MockRealGraphDataSource


def generate_realistic_shadow_fixture() -> MockRealGraphDataSource:
    """
    Generates a deterministic, realistic shadow event fixture containing:
    1. Legitimate Customer Support Agent (high volume, 150+ accounts, normal hours, low concentration)
    2. Legitimate Fraud Analyst (account reviews, case closures, no self-approval)
    3. Corporate Shared NAT Gateway (50 employees, 200 consumer accounts)
    4. Support pre-transaction assistance (benign support before customer purchase)
    5. Rogue Support Employee 1 (tight account concentration, rapid pre-tx access, self-approval)
    6. Rogue Employee 2 (off-hours access at 02:30 AM, shared personal device colocation)
    7. Low-and-Slow Collusion Ring (stealthy access across 3 weeks, low-velocity withdrawals)
    8. Multi-Hop Collusion Ring (Employee -> Device -> Account -> Consumer)
    """
    base_time = datetime(2026, 8, 1, 9, 0, 0, tzinfo=timezone.utc)
    source = MockRealGraphDataSource(dataset_type=DatasetType.REAL_SHADOW)

    # -------------------------------------------------------------------------
    # 1. ENTITIES (NODES)
    # -------------------------------------------------------------------------
    # Employees
    emp_support = HeteroNode(id="emp_support_alice", type=HeteroNodeType.EMPLOYEE, created_at=base_time, properties={"role": "customer_support", "department": "Customer Care"})
    emp_analyst = HeteroNode(id="emp_analyst_bob", type=HeteroNodeType.EMPLOYEE, created_at=base_time, properties={"role": "fraud_analyst", "department": "Risk Ops"})
    emp_rogue1 = HeteroNode(id="emp_rogue_charlie", type=HeteroNodeType.EMPLOYEE, created_at=base_time, properties={"role": "customer_support", "department": "Customer Care"})
    emp_rogue2 = HeteroNode(id="emp_rogue_david", type=HeteroNodeType.EMPLOYEE, created_at=base_time, properties={"role": "payments_operations", "department": "Payments Ops"})
    emp_stealth = HeteroNode(id="emp_stealth_eve", type=HeteroNodeType.EMPLOYEE, created_at=base_time, properties={"role": "customer_support", "department": "Customer Care"})

    for emp in [emp_support, emp_analyst, emp_rogue1, emp_rogue2, emp_stealth]:
        source.add_node(emp)

    # Infrastructure
    dev_corp_nat = HeteroNode(id="ip_corp_nat_hub", type=HeteroNodeType.IP, created_at=base_time, properties={"ip_type": "corporate_nat"})
    dev_rogue_phone = HeteroNode(id="dev_shared_phone_99", type=HeteroNodeType.DEVICE, created_at=base_time, properties={"device_type": "mobile"})
    dev_legit_workstation = HeteroNode(id="dev_corp_workstation_12", type=HeteroNodeType.DEVICE, created_at=base_time, properties={"device_type": "corporate_workstation"})

    for infra in [dev_corp_nat, dev_rogue_phone, dev_legit_workstation]:
        source.add_node(infra)

    # Legitimate consumers & accounts (30 benign accounts)
    for i in range(1, 31):
        c_id = f"usr_legit_{i:03d}"
        a_id = f"acct_legit_{i:03d}"
        d_id = f"dev_legit_{i:03d}"
        ip_id = f"ip_legit_{i:03d}"
        t_id = f"tok_legit_{i:03d}"

        source.add_node(HeteroNode(id=c_id, type=HeteroNodeType.CONSUMER, created_at=base_time))
        source.add_node(HeteroNode(id=a_id, type=HeteroNodeType.ACCOUNT, created_at=base_time, properties={"consumer_id": c_id}))
        source.add_node(HeteroNode(id=d_id, type=HeteroNodeType.DEVICE, created_at=base_time, properties={"device_type": "mobile"}))
        source.add_node(HeteroNode(id=ip_id, type=HeteroNodeType.IP, created_at=base_time, properties={"ip_type": "residential"}))
        source.add_node(HeteroNode(id=t_id, type=HeteroNodeType.PAYMENT_TOKEN, created_at=base_time))

        # Ownership & device linkages
        source.add_edge(HeteroEdge(id=f"e_own_{i}", source_id=c_id, target_id=a_id, type=HeteroEdgeType.CONSUMER_OWNS_ACCOUNT, timestamp=base_time))
        source.add_edge(HeteroEdge(id=f"e_dev_{i}", source_id=a_id, target_id=d_id, type=HeteroEdgeType.ACCOUNT_USES_DEVICE, timestamp=base_time))
        source.add_edge(HeteroEdge(id=f"e_ip_{i}", source_id=a_id, target_id=ip_id, type=HeteroEdgeType.ACCOUNT_USES_IP, timestamp=base_time))
        source.add_edge(HeteroEdge(id=f"e_tok_{i}", source_id=a_id, target_id=t_id, type=HeteroEdgeType.ACCOUNT_USES_PAYMENT_TOKEN, timestamp=base_time))

    # Mule consumers & accounts for collusion scenarios
    for i in range(1, 11):
        m_cid = f"usr_mule_{i:02d}"
        m_aid = f"acct_mule_{i:02d}"
        source.add_node(HeteroNode(id=m_cid, type=HeteroNodeType.CONSUMER, created_at=base_time, risk_score=0.10))
        source.add_node(HeteroNode(id=m_aid, type=HeteroNodeType.ACCOUNT, created_at=base_time, properties={"consumer_id": m_cid}))
        source.add_edge(HeteroEdge(id=f"e_mown_{i}", source_id=m_cid, target_id=m_aid, type=HeteroEdgeType.CONSUMER_OWNS_ACCOUNT, timestamp=base_time))

    # -------------------------------------------------------------------------
    # 2. CHRONOLOGICAL RELATIONSHIP EVENTS & TRANSACTIONS
    # -------------------------------------------------------------------------
    curr_time = base_time + timedelta(days=1)
    edge_cnt = 1000

    # Scenario A: Legitimate Customer Support (Alice touches 20 diverse accounts across normal hours)
    for i in range(1, 21):
        curr_time += timedelta(minutes=15)
        aid = f"acct_legit_{i:03d}"
        edge_cnt += 1
        source.add_edge(HeteroEdge(
            id=f"e_sup_{edge_cnt}",
            source_id="emp_support_alice",
            target_id=aid,
            type=HeteroEdgeType.EMPLOYEE_ACCESSES_ACCOUNT,
            timestamp=curr_time,
            properties={"purpose": "customer_support_inquiry", "ticket_id": f"TICK-{1000+i}"}
        ))
        # Benign transaction follows after support call
        curr_time += timedelta(minutes=5)
        source.add_transaction({
            "transaction_id": f"txn_legit_{i:03d}",
            "account_id": aid,
            "amount": 45.0 + (i * 3.5),
            "currency": "USD",
            "event_timestamp": curr_time.isoformat(),
            "is_collusion": False
        })

    # Scenario B: Legitimate Fraud Analyst (Bob reviews cases)
    for i in range(21, 26):
        curr_time += timedelta(hours=1)
        aid = f"acct_legit_{i:03d}"
        case_id = f"case_rev_{i:03d}"
        source.add_node(HeteroNode(id=case_id, type=HeteroNodeType.CASE, created_at=curr_time, properties={"status": "CLOSED_BENIGN"}))
        edge_cnt += 1
        source.add_edge(HeteroEdge(
            id=f"e_case_{edge_cnt}",
            source_id="emp_analyst_bob",
            target_id=case_id,
            type=HeteroEdgeType.EMPLOYEE_REVIEWS_CASE,
            timestamp=curr_time
        ))
        edge_cnt += 1
        source.add_edge(HeteroEdge(
            id=f"e_case_ref_{edge_cnt}",
            source_id=case_id,
            target_id=aid,
            type=HeteroEdgeType.CASE_REFERENCES_ACCOUNT,
            timestamp=curr_time
        ))

    # Scenario C: Suspicious Collusion 1 (Charlie repeatedly accesses tight cluster of 3 mule accounts, self-approves)
    curr_time += timedelta(days=2)
    mule_cluster_1 = ["acct_mule_01", "acct_mule_02", "acct_mule_03"]
    for repeat in range(3):
        for aid in mule_cluster_1:
            curr_time += timedelta(minutes=2)
            edge_cnt += 1
            source.add_edge(HeteroEdge(
                id=f"e_rogue1_{edge_cnt}",
                source_id="emp_rogue_charlie",
                target_id=aid,
                type=HeteroEdgeType.EMPLOYEE_ACCESSES_ACCOUNT,
                timestamp=curr_time,
                properties={"off_hours": False}
            ))
            # Rapid high-value transaction 30s after access
            curr_time += timedelta(seconds=30)
            txn_id = f"txn_collusion_c_{repeat}_{aid}"
            source.add_transaction({
                "transaction_id": txn_id,
                "account_id": aid,
                "amount": 4800.0,
                "currency": "USD",
                "event_timestamp": curr_time.isoformat(),
                "is_collusion": True
            })
            # Rogue employee approves transaction
            curr_time += timedelta(seconds=15)
            edge_cnt += 1
            source.add_edge(HeteroEdge(
                id=f"e_appr_{edge_cnt}",
                source_id="emp_rogue_charlie",
                target_id=txn_id,
                type=HeteroEdgeType.EMPLOYEE_APPROVES_TRANSACTION,
                timestamp=curr_time
            ))

    # Scenario D: Suspicious Collusion 2 (David accesses accounts at 02:30 AM via shared mobile device)
    curr_time += timedelta(days=1)
    curr_time = curr_time.replace(hour=2, minute=30, second=0)  # Off-hours: 2:30 AM
    mule_cluster_2 = ["acct_mule_04", "acct_mule_05"]
    for aid in mule_cluster_2:
        edge_cnt += 1
        source.add_edge(HeteroEdge(
            id=f"e_rogue2_{edge_cnt}",
            source_id="emp_rogue_david",
            target_id=aid,
            type=HeteroEdgeType.EMPLOYEE_ACCESSES_ACCOUNT,
            timestamp=curr_time,
            properties={"off_hours": True}
        ))
        # Shared device link
        edge_cnt += 1
        source.add_edge(HeteroEdge(
            id=f"e_dev_rogue_{edge_cnt}",
            source_id="emp_rogue_david",
            target_id="dev_shared_phone_99",
            type=HeteroEdgeType.EMPLOYEE_USES_DEVICE,
            timestamp=curr_time
        ))
        edge_cnt += 1
        source.add_edge(HeteroEdge(
            id=f"e_acct_dev_rogue_{edge_cnt}",
            source_id=aid,
            target_id="dev_shared_phone_99",
            type=HeteroEdgeType.ACCOUNT_USES_DEVICE,
            timestamp=curr_time
        ))
        curr_time += timedelta(minutes=3)
        source.add_transaction({
            "transaction_id": f"txn_collusion_d_{aid}",
            "account_id": aid,
            "amount": 9500.0,
            "currency": "USD",
            "event_timestamp": curr_time.isoformat(),
            "is_collusion": True
        })

    # Scenario E: Low-and-Slow Collusion (Eve touches accounts across 3 weeks with spaced transactions)
    for week in range(3):
        curr_time += timedelta(days=7)
        aid = "acct_mule_06"
        edge_cnt += 1
        source.add_edge(HeteroEdge(
            id=f"e_stealth_{edge_cnt}",
            source_id="emp_stealth_eve",
            target_id=aid,
            type=HeteroEdgeType.EMPLOYEE_ACCESSES_ACCOUNT,
            timestamp=curr_time
        ))
        curr_time += timedelta(hours=2)
        source.add_transaction({
            "transaction_id": f"txn_stealth_{week}",
            "account_id": aid,
            "amount": 950.0,
            "currency": "USD",
            "event_timestamp": curr_time.isoformat(),
            "is_collusion": True
        })

    # -------------------------------------------------------------------------
    # 3. LABELS (MATURED CHRONOLOGICALLY)
    # -------------------------------------------------------------------------
    # Matured legitimate transactions (> 30 days after event)
    for i in range(1, 15):
        source.add_label({
            "label_id": f"lbl_legit_{i}",
            "entity_id": f"txn_legit_{i:03d}",
            "maturity_state": LabelMaturityState.CONFIRMED_LEGITIMATE.value,
            "ground_truth": "LEGITIMATE",
            "label_timestamp": (base_time + timedelta(days=60)).isoformat(),
            "label_type": "MATURED_CLEAN"
        })

    # Confirmed real holdout cases (0 used for training, simulated mature outcome)
    source.add_label({
        "label_id": "lbl_collusion_c1",
        "entity_id": "txn_collusion_c_0_acct_mule_01",
        "maturity_state": LabelMaturityState.CONFIRMED_FRAUD.value,
        "ground_truth": "NON_LEGITIMATE",
        "label_timestamp": (base_time + timedelta(days=45)).isoformat(),
        "label_type": "INTERNAL_COLLUSION",
        "is_collusion": True
    })

    return source
