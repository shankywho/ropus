package graphsage

import (
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

// ConnectivityReport captures the real-world infrastructure audit in Go.
type ConnectivityReport struct {
	AuditedAt             time.Time `json:"audited_at"`
	IsLiveConnected       bool      `json:"is_live_connected"`
	IntegrationStatus     string    `json:"integration_status"` // "READY_SHADOW" or "INTEGRATION_BLOCKED"
	MissingConfigurations []string  `json:"missing_configurations"`
	KafkaBrokerReachable  bool      `json:"kafka_broker_reachable"`
	KafkaBrokerAddress    string    `json:"kafka_broker_address"`
	ClickHouseReachable   bool      `json:"clickhouse_reachable"`
	ClickHouseAddress     string    `json:"clickhouse_address"`
	PrometheusReachable   bool      `json:"prometheus_reachable"`
	PrometheusAddress     string    `json:"prometheus_address"`
	SummaryMessage        string    `json:"summary_message"`
}

// ConnectivityChecker performs genuine network reachability audits.
type ConnectivityChecker struct {
	Timeout time.Duration
}

// NewConnectivityChecker initializes the checker.
func NewConnectivityChecker(timeout time.Duration) *ConnectivityChecker {
	if timeout <= 0 {
		timeout = 1 * time.Second
	}
	return &ConnectivityChecker{Timeout: timeout}
}

func (c *ConnectivityChecker) checkPort(address string) bool {
	conn, err := net.DialTimeout("tcp", address, c.Timeout)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

// AuditEnvironment checks Kafka, ClickHouse, and Prometheus endpoints.
func (c *ConnectivityChecker) AuditEnvironment() ConnectivityReport {
	now := time.Now().UTC()
	var missing []string

	// 1. Kafka Audit
	kafkaEnv := os.Getenv("KAFKA_BROKERS")
	if kafkaEnv == "" {
		kafkaEnv = os.Getenv("KAFKA_BROKER")
	}

	kafkaReachable := false
	kafkaAddr := ""

	if kafkaEnv == "" {
		missing = append(missing, "KAFKA_BROKERS (e.g. kafka-staging-01.internal:9092)")
		if c.checkPort("127.0.0.1:9092") {
			kafkaReachable = true
			kafkaAddr = "127.0.0.1:9092 (Local Dev Broker)"
		}
	} else {
		kafkaAddr = kafkaEnv
		brokers := strings.Split(kafkaEnv, ",")
		if len(brokers) > 0 {
			kafkaReachable = c.checkPort(strings.TrimSpace(brokers[0]))
		}
	}

	if os.Getenv("KAFKA_SASL_USERNAME") == "" && os.Getenv("KAFKA_SSL_CA_LOCATION") == "" {
		missing = append(missing, "KAFKA_AUTH_CREDENTIALS (SASL_SSL or mTLS certificates)")
	}

	// 2. ClickHouse Audit
	chHost := os.Getenv("CLICKHOUSE_HOST")
	if chHost == "" {
		chHost = os.Getenv("CLICKHOUSE_ADDR")
	}
	chPort := os.Getenv("CLICKHOUSE_HTTP_PORT")
	if chPort == "" {
		chPort = "8123"
	}

	chReachable := false
	chAddr := ""

	if chHost == "" {
		missing = append(missing, "CLICKHOUSE_HOST (e.g. clickhouse-cluster.internal)")
		if c.checkPort("127.0.0.1:8123") || c.checkPort("127.0.0.1:8124") {
			chReachable = true
			chAddr = "127.0.0.1:8123 (Local Dev Instance)"
		}
	} else {
		chAddr = fmt.Sprintf("%s:%s", chHost, chPort)
		chReachable = c.checkPort(chAddr)
	}

	if os.Getenv("CLICKHOUSE_PASSWORD") == "" && os.Getenv("CLICKHOUSE_USER") == "" {
		missing = append(missing, "CLICKHOUSE_AUTH (CLICKHOUSE_USER / CLICKHOUSE_PASSWORD)")
	}

	// 3. Prometheus Audit
	promHost := os.Getenv("PROMETHEUS_HOST")
	if promHost == "" {
		promHost = "127.0.0.1"
	}
	promPort := os.Getenv("PROMETHEUS_PORT")
	if promPort == "" {
		promPort = "9090"
	}
	promAddr := fmt.Sprintf("%s:%s", promHost, promPort)
	promReachable := c.checkPort(promAddr)

	isLiveReady := len(missing) == 0 && kafkaReachable && chReachable
	statusStr := "INTEGRATION_BLOCKED"
	if isLiveReady {
		statusStr = "READY_SHADOW"
	}

	summary := fmt.Sprintf("Operational shadow integration is %s. %d missing configuration parameters.", statusStr, len(missing))
	if isLiveReady {
		summary = "Staging shadow streaming infrastructure and ClickHouse evidence store are connected."
	}

	return ConnectivityReport{
		AuditedAt:             now,
		IsLiveConnected:       isLiveReady,
		IntegrationStatus:     statusStr,
		MissingConfigurations: missing,
		KafkaBrokerReachable:  kafkaReachable,
		KafkaBrokerAddress:    kafkaAddr,
		ClickHouseReachable:   chReachable,
		ClickHouseAddress:     chAddr,
		PrometheusReachable:   promReachable,
		PrometheusAddress:     promAddr,
		SummaryMessage:        summary,
	}
}
