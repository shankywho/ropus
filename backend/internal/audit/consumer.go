package audit

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
)

// DecisionAuditPayload matches the JSON shape published to risk.events.
type DecisionAuditPayload struct {
	DecisionID        string          `json:"decision_id"`
	TenantID          string          `json:"tenant_id"`
	TransactionID     string          `json:"transaction_id"`
	Amount            int64           `json:"amount"`
	Currency          string          `json:"currency"`
	RecommendedAction string          `json:"recommended_action"`
	RiskScore         int             `json:"risk_score"`
	ReasonCodes       []string        `json:"reason_codes"`
	FeatureSnapshot   json.RawMessage `json:"feature_snapshot"`
	EvaluatedAt       string          `json:"evaluated_at"`
}

// KafkaAuditConsumer consumes decision events and streams them to ClickHouse for analytics.
type KafkaAuditConsumer struct {
	reader   *kafka.Reader
	chClient *ClickHouseClient
}

// NewKafkaAuditConsumer creates a new Kafka consumer for ClickHouse auditing.
func NewKafkaAuditConsumer(brokers []string, topic, groupID string, chClient *ClickHouseClient) *KafkaAuditConsumer {
	if len(brokers) == 0 {
		brokers = []string{"localhost:9092"}
	}
	if topic == "" {
		topic = "risk.events"
	}
	if groupID == "" {
		groupID = "risk-audit-group"
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        brokers,
		Topic:          topic,
		GroupID:        groupID,
		MinBytes:       1e3,  // 1KB
		MaxBytes:       10e6, // 10MB
		CommitInterval: time.Second,
		StartOffset:    kafka.LastOffset,
	})

	return &KafkaAuditConsumer{
		reader:   reader,
		chClient: chClient,
	}
}

// Start runs the audit consumption loop in a blocking manner.
func (c *KafkaAuditConsumer) Start(ctx context.Context) {
	log.Printf("Starting ClickHouse Audit Kafka Consumer on topic '%s' (group: risk-audit-group)...", c.reader.Config().Topic)

	for {
		select {
		case <-ctx.Done():
			log.Println("ClickHouse Audit Consumer shutting down...")
			return
		default:
			msg, err := c.reader.ReadMessage(ctx)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					return
				}
				select {
				case <-ctx.Done():
					return
				case <-time.After(2 * time.Second):
					continue
				}
			}

			c.processEvent(ctx, msg.Value)
		}
	}
}

// processEvent decodes the message and writes it to ClickHouse.
func (c *KafkaAuditConsumer) processEvent(ctx context.Context, data []byte) {
	if len(data) == 0 || c.chClient == nil {
		return
	}

	var payload DecisionAuditPayload

	// 1. Direct JSON Unmarshal
	if err := json.Unmarshal(data, &payload); err != nil {
		// 2. Debezium Envelope Fallback
		var envelope map[string]interface{}
		if errEnv := json.Unmarshal(data, &envelope); errEnv == nil {
			if payloadStr, ok := envelope["payload"].(string); ok {
				_ = json.Unmarshal([]byte(payloadStr), &payload)
			} else if afterMap, ok := envelope["after"].(map[string]interface{}); ok {
				afterBytes, _ := json.Marshal(afterMap)
				_ = json.Unmarshal(afterBytes, &payload)
			}
		}
	}

	if payload.TransactionID == "" {
		return
	}

	if payload.TenantID == "" {
		payload.TenantID = "00000000-0000-0000-0000-000000000001"
	}

	// Parse EvaluatedAt or fallback to current time
	createdAt := time.Now().UTC()
	if payload.EvaluatedAt != "" {
		if t, err := time.Parse(time.RFC3339, payload.EvaluatedAt); err == nil {
			createdAt = t
		}
	}

	ruleTriggered := strings.Join(payload.ReasonCodes, ", ")
	if ruleTriggered == "" {
		ruleTriggered = payload.RecommendedAction
	}

	featureSnapshotStr := string(payload.FeatureSnapshot)
	if featureSnapshotStr == "" || featureSnapshotStr == "null" {
		featureSnapshotStr = "{}"
	}

	record := AuditRecord{
		TransactionID:   payload.TransactionID,
		RiskScore:       int32(payload.RiskScore),
		RuleTriggered:   ruleTriggered,
		FeatureSnapshot: featureSnapshotStr,
		TenantID:        payload.TenantID,
		CreatedAt:       createdAt,
	}

	if err := c.chClient.InsertAuditRecord(ctx, record); err != nil {
		log.Printf("Warning: Failed to sink audit record to ClickHouse: %v", err)
	} else {
		log.Printf("Sinked decision event to ClickHouse OLAP [Txn: %s, Score: %d]", record.TransactionID, record.RiskScore)
	}
}

// Close gracefully closes the Kafka reader.
func (c *KafkaAuditConsumer) Close() error {
	if c.reader != nil {
		return c.reader.Close()
	}
	return nil
}
