package cases

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

// DecisionEventPayload represents the structure of decision messages arriving over Kafka/Redpanda.
type DecisionEventPayload struct {
	DecisionID        string `json:"decision_id"`
	TenantID          string `json:"tenant_id"`
	TransactionID     string `json:"transaction_id"`
	RecommendedAction string `json:"recommended_action"`
	RiskScore         int    `json:"risk_score"`
}

// KafkaConsumer listens to risk decision events and generates analyst review cases asynchronously.
type KafkaConsumer struct {
	reader      *kafka.Reader
	caseService *Service
}

// NewKafkaConsumer initializes a KafkaConsumer configured with the Redpanda/Kafka broker.
func NewKafkaConsumer(brokers []string, topic, groupID string, caseService *Service) *KafkaConsumer {
	if len(brokers) == 0 {
		brokers = []string{"localhost:9092"}
	}
	if topic == "" {
		topic = "risk.events"
	}
	if groupID == "" {
		groupID = "risk-case-manager-group"
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

	return &KafkaConsumer{
		reader:      reader,
		caseService: caseService,
	}
}

// Start begins the event consumption loop in a blocking manner (should be invoked in a goroutine).
func (c *KafkaConsumer) Start(ctx context.Context) {
	log.Printf("Starting asynchronous Kafka Case Consumer on topic '%s'...", c.reader.Config().Topic)

	for {
		select {
		case <-ctx.Done():
			log.Println("Kafka Case Consumer shutting down...")
			return
		default:
			msg, err := c.reader.ReadMessage(ctx)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					return
				}
				// If broker is offline in local dev mode, backoff gracefully
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

// processEvent parses the event and provisions a case if recommended_action == MANUAL_REVIEW.
func (c *KafkaConsumer) processEvent(ctx context.Context, data []byte) {
	if len(data) == 0 {
		return
	}

	var payload DecisionEventPayload

	// 1. Direct JSON unmarshal
	if err := json.Unmarshal(data, &payload); err != nil {
		// 2. Fallback: Check if wrapped in a Debezium CDC envelope or string payload
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

	if payload.RecommendedAction != "MANUAL_REVIEW" {
		return
	}

	if payload.TenantID == "" {
		payload.TenantID = "00000000-0000-0000-0000-000000000001"
	}

	priority := "MEDIUM"
	if payload.RiskScore >= 80 {
		priority = "HIGH"
	}

	createdCase, err := c.caseService.CreateCaseFromDecision(
		ctx,
		payload.TenantID,
		payload.DecisionID,
		payload.TransactionID,
		priority,
	)
	if err != nil {
		log.Printf("Warning: Failed to asynchronously create case for decision %s: %v", payload.DecisionID, err)
		return
	}

	if createdCase != nil {
		log.Printf("Async Case Created [CaseID: %s, TxnID: %s, SLA: %s]",
			createdCase.ID, createdCase.TransactionID, createdCase.SLAExpiresAt.Format(time.RFC3339))
	}
}

// Close gracefully closes the Kafka reader.
func (c *KafkaConsumer) Close() error {
	if c.reader != nil {
		return c.reader.Close()
	}
	return nil
}
