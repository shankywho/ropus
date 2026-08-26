"""
Real-Data Connectivity & Environment Integration Checker (Phase 62)
Audits live Kafka / Redpanda brokers, ClickHouse cluster, and Prometheus infrastructure honestly without fabricating connectivity.
"""

import os
import socket
import urllib.request
import urllib.error
from typing import Dict, List, Any, Optional
from pydantic import BaseModel, Field
from datetime import datetime, timezone


class RealDataConnectivityReport(BaseModel):
    audited_at: datetime
    is_live_connected: bool
    integration_status: str  # "READY_SHADOW" or "INTEGRATION_BLOCKED"
    missing_configurations: List[str] = Field(default_factory=list)
    kafka_broker_reachable: bool
    kafka_broker_address: Optional[str] = None
    shadow_topics_found: List[str] = Field(default_factory=list)
    clickhouse_reachable: bool
    clickhouse_address: Optional[str] = None
    prometheus_reachable: bool
    prometheus_address: Optional[str] = None
    summary_message: str


class RealDataConnectivityChecker:
    """
    Evaluates real environmental connectivity against staging / production infrastructure.
    """

    def __init__(self, timeout_sec: float = 1.0):
        self.timeout_sec = timeout_sec

    def _check_tcp_port(self, host: str, port: int) -> bool:
        s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        s.settimeout(self.timeout_sec)
        try:
            s.connect((host, port))
            s.close()
            return True
        except Exception:
            return False

    def audit_environment(self) -> RealDataConnectivityReport:
        now = datetime.now(timezone.utc)
        missing_configs: List[str] = []

        # 1. Kafka / Redpanda Inspection
        kafka_brokers_env = os.environ.get("KAFKA_BROKERS") or os.environ.get("KAFKA_BROKER")
        kafka_reachable = False
        kafka_addr = None

        if not kafka_brokers_env:
            missing_configs.append("KAFKA_BROKERS (e.g. kafka-staging-01.internal:9092,kafka-staging-02.internal:9092)")
            # Check local fallback port 9092
            if self._check_tcp_port("127.0.0.1", 9092):
                kafka_reachable = True
                kafka_addr = "127.0.0.1:9092 (Local Dev Broker)"
        else:
            kafka_addr = kafka_brokers_env
            # Attempt to parse and probe first host:port
            first_broker = kafka_brokers_env.split(",")[0].strip()
            if ":" in first_broker:
                host, port_str = first_broker.split(":", 1)
                try:
                    port = int(port_str)
                    kafka_reachable = self._check_tcp_port(host, port)
                except ValueError:
                    kafka_reachable = False

        # Check required credentials
        if not os.environ.get("KAFKA_SASL_USERNAME") and not os.environ.get("KAFKA_SSL_CA_LOCATION"):
            missing_configs.append("KAFKA_AUTH_CREDENTIALS (SASL_SSL username/password or mTLS certificates)")

        # 2. ClickHouse Inspection
        ch_host = os.environ.get("CLICKHOUSE_HOST") or os.environ.get("CLICKHOUSE_ADDR")
        ch_port_str = os.environ.get("CLICKHOUSE_HTTP_PORT") or "8123"
        ch_reachable = False
        ch_addr = None

        if not ch_host:
            missing_configs.append("CLICKHOUSE_HOST (e.g. clickhouse-cluster.internal)")
            # Check local fallback
            if self._check_tcp_port("127.0.0.1", 8123) or self._check_tcp_port("127.0.0.1", 8124):
                ch_reachable = True
                ch_addr = "127.0.0.1:8123 (Local Dev Instance)"
        else:
            try:
                ch_port = int(ch_port_str)
                ch_reachable = self._check_tcp_port(ch_host, ch_port)
                ch_addr = f"{ch_host}:{ch_port}"
            except ValueError:
                ch_reachable = False

        if not os.environ.get("CLICKHOUSE_PASSWORD") and not os.environ.get("CLICKHOUSE_USER"):
            missing_configs.append("CLICKHOUSE_AUTH (CLICKHOUSE_USER / CLICKHOUSE_PASSWORD)")

        # 3. Prometheus Inspection
        prom_host = os.environ.get("PROMETHEUS_HOST") or "127.0.0.1"
        prom_port_str = os.environ.get("PROMETHEUS_PORT") or "9090"
        prom_reachable = False
        prom_addr = f"{prom_host}:{prom_port_str}"

        try:
            prom_port = int(prom_port_str)
            prom_reachable = self._check_tcp_port(prom_host, prom_port)
        except ValueError:
            prom_reachable = False

        if not os.environ.get("PROMETHEUS_PUSHGATEWAY_URL") and not os.environ.get("PROMETHEUS_HOST"):
            missing_configs.append("PROMETHEUS_CONFIG (PROMETHEUS_PUSHGATEWAY_URL / PROMETHEUS_HOST)")

        # Determine integration status
        is_live_ready = (len(missing_configs) == 0 and kafka_reachable and ch_reachable)
        status_str = "READY_SHADOW" if is_live_ready else "INTEGRATION_BLOCKED"

        if is_live_ready:
            summary = "Live staging shadow broker and ClickHouse evidence store are reachable and configured."
        else:
            summary = (
                f"Operational shadow integration is INTEGRATION_BLOCKED due to {len(missing_configs)} missing configuration parameters. "
                "Passive shadow telemetry runs safely in offline simulation mode."
            )

        return RealDataConnectivityReport(
            audited_at=now,
            is_live_connected=is_live_ready,
            integration_status=status_str,
            missing_configurations=missing_configs,
            kafka_broker_reachable=kafka_reachable,
            kafka_broker_address=kafka_addr,
            shadow_topics_found=[],
            clickhouse_reachable=ch_reachable,
            clickhouse_address=ch_addr,
            prometheus_reachable=prom_reachable,
            prometheus_address=prom_addr,
            summary_message=summary
        )
