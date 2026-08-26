# synthetic_ropus

**SYNTHETIC — DEVELOPMENT / ENGINEERING USE ONLY.**

Every node, edge, transaction, and label in this dataset is generated. None
of it describes a real employee, consumer, account, device, IP, payment
instrument, or investigation. It must never be cited as production
accuracy, production recall, real-world fraud prevalence, real employee
collusion prevalence, or production economic savings. **The 52 confirmed
fraud cases in the real ROPUS frozen holdout are separate, real evidence and
are not touched, referenced, duplicated, or extended by this generator.**

This dataset exists to develop and stress-test the ROPUS pipeline itself
(GraphSAGE architecture, graph construction, temporal-leakage guards,
label-maturation handling, explainability) — not to measure how well ROPUS
will perform in production.

---

## 1. Dataset generation architecture

```
synthetic_ropus/
    README.md
    generate.py            # CLI entrypoint
    config.yaml             # proportions, date range, role mix, splits
    scenarios/
        common.py            # ID allocation, timestamps, entity pools
        legitimate.py         # bulk normal activity + hard negatives
        employee_collusion.py
        fraud_ring.py
        account_takeover.py
        card_testing.py
        bot_attack.py
        sybil.py
        graph_poisoning.py
    validation/
        validate_temporal.py  # no future-info leakage
        validate_labels.py    # label well-formedness, independence, diversity
        validate_graph.py     # referential integrity, degree sanity
        validate_privacy.py   # synthetic-id pattern, no PAN/secrets
    data/                    # generator output (created by generate.py)
        nodes/*.csv
        edges/*.csv
        transactions/transactions.csv
        labels/*.csv
```

`generate.py` builds one shared universe of entities (`build_entity_pools`
in `scenarios/common.py`), then hands each scenario module the *same*
pools, with small non-overlapping reserved sub-pools for fraud mechanisms.
This is deliberate: fraud entities are a subset of the same population as
legitimate entities, not a separately-distinguishable synthetic species.
Each scenario module returns plain rows via a shared `GenResult` container;
`generate.py` merges everything, sorts transactions chronologically, and
writes CSVs.

Run it:

```bash
pip install pyyaml
python generate.py --events 100000 --seed 42
python generate.py --events 500000 --seed 7 --outdir data_500k
python generate.py --events 1000000 --seed 7 --outdir data_1m
```

100,000 events currently generates in ~10 seconds on a single core, so
500k/1M runs are practical without a scalable-generator rewrite; the
per-scenario-function structure means any module can be swapped for a
batched/vectorized version later without touching the others.

## 2. Event/node/edge schema

**Node files** (`data/nodes/*.csv`): `employee` (employee_id, role),
`consumer` (consumer_id), `account` (account_id, consumer_id),
`device` (device_id, device_type), `ip` (ip_id, ip_type),
`payment_token` (payment_token_id), `merchant` (merchant_id, category),
`case` (case_id, opened_at, status), `session` (session_id, account_id,
device_id, ip_id, started_at, ended_at).

**Edge files** (`data/edges/<edge_type>.csv`), one file per relationship
type from the spec (`employee_accesses_account`, `employee_reviews_case`,
`employee_modifies_transaction`, `employee_approves_transaction`,
`employee_uses_device`, `employee_uses_ip`, `consumer_owns_account`,
`account_uses_device`, `account_uses_ip`, `account_uses_payment_token`,
`account_transacts_with_merchant`, `case_references_account`,
`case_references_consumer`, `case_references_employee`). Every edge row
carries `edge_timestamp`.

**Transactions** (`data/transactions/transactions.csv`): transaction_id,
account_id, device_id, ip_id, payment_token_id, merchant_id, amount,
currency, txn_type, `event_timestamp`, `ingestion_timestamp`
(ingestion is always >= event).

**Labels** (`data/labels/labels.csv`): label_id, entity_type, entity_id,
ground_truth, `label_timestamp`, label_source, label_confidence. Ground
truth is generated independently of any model — there is no
GraphSAGE-derived value anywhere in this repository, so there's no
mechanism for a prediction to become a label.

**Scenario tags** (`data/labels/scenario_tags.csv`): tag_id, entity_type,
entity_id, scenario_type, is_adversarial, is_hard_negative. This file is
kept **separate from every feature/transaction/edge table on purpose** —
it's for building evaluation subsets and for auditing the generator, and
must never be joined into a model's feature set (that would trivially leak
the scenario as a feature).

## 3. Scenario distribution

Controlled by `scenario_weights` in `config.yaml` (fractions of total
transaction volume): legitimate 94%, employee_collusion 1.5%, fraud_ring
1.5%, account_takeover 0.8%, card_testing 0.8%, bot_attack 0.8%, sybil
0.4%, graph_poisoning 0.2%. Observed output at `--events 100000 --seed 42`:
92.93% LEGITIMATE, 7.07% spread across FRAUD_RING / SUSPECTED_FRAUD /
INTERNAL_COLLUSION / BOT_ATTACK / ACCOUNT_TAKEOVER / CARD_TESTING —
comfortably inside the requested 90–97% / 3–10% band, and not a
single monolithic fraud pattern.

## 4. Legitimate vs. suspicious distribution

See §3. Adjust `scenario_weights` in `config.yaml` to move the ratio; the
generator re-normalizes weights automatically.

## 5. Employee-consumer scenario distribution

Both directions are generated so a model cannot learn "employee touched
consumer = fraud": `legitimate.py` produces high-volume customer-support
access, fraud-analyst repeated review of known-suspicious accounts, shared
corporate IP/device usage, and access-immediately-before-a-legitimate-
transaction — all tagged `is_hard_negative=True`. `employee_collusion.py`
produces the fraudulent counterpart: a small number of employees
repeatedly accessing a small cluster of accounts with tight access-to-
transaction timing, sometimes on a shared device, sometimes also
self-approving the resulting transaction. Roughly half of collusion rings
are "obvious" (many correlated signals) and half are "subtle" (2-3 weaker
signals), per the hard-detection requirement.

## 6. Temporal split strategy

Transactions are sorted by `event_timestamp` and split by **index after
sorting**, never by random shuffle: earliest 70% train, next 15%
validation, latest 15% test (`config.yaml: splits`). `validate_temporal.py`
confirms `max(event_timestamp)` in an earlier split never exceeds
`min(event_timestamp)` in a later split.

## 7. Label-generation methodology

Ground truth is assigned by the scenario module that generated the event,
completely independent of any feature computation or model — it comes from
the generator's scenario logic, not from thresholding a score. Every label
carries `label_source` (internal_investigation / fraud_analyst_review /
automated_rule_confirmation / consumer_dispute / automated_baseline) and
`label_confidence`, and delayed maturation is modeled explicitly:
legitimate/automated labels mature in minutes-to-a-day; fraud/collusion
labels mature 1-30 days after the event (`add_label_delay` in
`scenarios/common.py`), mirroring the spec's discovery -> investigation ->
confirmation timeline.

## 8. Hard-negative methodology

`legitimate.py` deliberately generates behavior that *shares surface
features with fraud signals* but is legitimately benign: high employee
account-access volume (customer support), high employee-to-suspicious-
account concentration (fraud analysts doing their job), shared
corporate/residential infrastructure, and access-immediately-preceding-a-
transaction. All are tagged `is_hard_negative=True` in `scenario_tags.csv`
and labeled `LEGITIMATE`, so any classifier trained on this data has to
distinguish frequency/timing/context, not just presence of the raw
feature.

## 9. Leakage-validation methodology

`validation/validate_temporal.py` checks (a) ingestion_timestamp >=
event_timestamp for every transaction, (b) label_timestamp >=
event_timestamp for every label, (c) the dedicated
`temporal_leakage_test.csv` subset genuinely demonstrates labels that must
be withheld from a T-scoped evaluation view, and (d) split boundaries are
strictly chronological. `validation/validate_labels.py` separately checks
label provenance and that ground truth doesn't collapse to one mechanism.

## 10. Graph poisoning scenarios

`scenarios/graph_poisoning.py` generates: (a) a high-degree flood node — one
device attached to ~200 accounts with no matching transactions, to distort
degree-based features; (b) injected employee-accesses-account edges with
plausible timestamps but no genuine business context; (c) low-and-slow
transactions from a poisoned account pool using rotating device/IP/token
combinations, specifically to evade short-window velocity rules.
`is_adversarial=True` is recorded **only** in `scenario_tags.csv`, never as
a column on the transaction/edge tables themselves.

## 11. Reproducibility instructions

```bash
python generate.py --events 100000 --seed 42
```

Re-running with the same `--events` and `--seed` reproduces byte-identical
output (verified: two runs at `--events 5000 --seed 99` produced identical
directory trees). All randomness flows through a single `random.Random(seed)`
instance threaded through every scenario call — no unseeded global state.

## 12. Expected dataset size

At `--events 100000` (the default): ~83k transactions actually materialize
(some scenario transactions land after the configured end_date and are
skipped, which is intentional realism, not a bug) with ~18k consumers, 22k
accounts, 600 employees, 1200 merchants, and roughly 130k+ edges/labels/tags
combined across all files. Passing `--events 500000` or `--events 1000000`
scales entity pools and per-scenario volumes proportionally via
`config.yaml: entity_scaling`.

## 13. Limitations

- This is a **stress-test and architecture-validation dataset**, not a
  calibrated fraud-rate estimate. Real-world class balance, ring topology,
  and behavioral distributions will differ from any real deployment.
- Statistical properties (log-normal amounts, business-hours bias, etc.)
  are simplified approximations, not fit to real transaction data.
- Graph poisoning scenarios here are illustrative examples, not a
  red-team-complete adversarial suite.
- The generator is single-threaded pure Python; at very large scales
  (multi-million events) a batched/vectorized rewrite of the hottest
  scenario (`legitimate.py`) would be worth doing before relying on wall-
  clock time.

## 14. Synthetic data statement

**All data produced by this generator is synthetic.** No real people, real
accounts, real devices, real IP addresses, real payment instruments, or
real investigations are represented. Do not use this dataset to report
production metrics, real fraud/collusion prevalence, or real economic
impact. The real ROPUS benchmark (52 confirmed fraud cases, frozen holdout)
is the only real evidence and remains entirely separate from everything in
this repository.

---

## Validation

```bash
python validation/validate_temporal.py --data data
python validation/validate_labels.py --data data
python validation/validate_graph.py --data data
python validation/validate_privacy.py --data data
```

All four currently pass against a 100k-event / seed-42 run.

## Special evaluation subsets

Written to `data/labels/`: `employee_collusion_test.csv`,
`fraud_ring_test.csv` (includes sybil clusters), `bot_attack_test.csv`,
`graph_poisoning_test.csv`, `hard_negative_test.csv`,
`temporal_leakage_test.csv`.
