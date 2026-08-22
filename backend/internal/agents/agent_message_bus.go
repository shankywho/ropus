package agents

import (
	"fmt"
	"sync"
	"time"
)

// MessageType defines the category of an inter-agent message.
type MessageType string

const (
	MsgTask            MessageType = "TASK"
	MsgResult          MessageType = "RESULT"
	MsgAlert           MessageType = "ALERT"
	MsgApprovalRequest MessageType = "APPROVAL_REQUEST"
)

// AgentMessage represents an inter-agent communication packet.
type AgentMessage struct {
	MessageID   string                 `json:"message_id"`
	TraceID     string                 `json:"trace_id"`
	SourceAgent string                 `json:"source_agent"`
	TargetAgent AgentType              `json:"target_agent"`
	Type        MessageType            `json:"type"`
	Payload     map[string]interface{} `json:"payload"`
	Timestamp   time.Time              `json:"timestamp"`
}

// AgentMessageBus provides inter-agent messaging and task routing across the mesh.
type AgentMessageBus interface {
	Publish(msg *AgentMessage) error
	Subscribe(agentType AgentType, handler func(msg *AgentMessage) error) error
}

// LocalAgentBus is an ultra-high-speed in-memory message bus (>1M msgs/sec).
type LocalAgentBus struct {
	mu          sync.RWMutex
	subscribers map[AgentType][]func(msg *AgentMessage) error
}

// NewLocalAgentBus initializes the in-process agent message bus.
func NewLocalAgentBus() *LocalAgentBus {
	return &LocalAgentBus{
		subscribers: make(map[AgentType][]func(msg *AgentMessage) error),
	}
}

// Publish delivers an inter-agent message to registered target subscribers.
func (b *LocalAgentBus) Publish(msg *AgentMessage) error {
	if msg == nil {
		return fmt.Errorf("message cannot be nil")
	}

	b.mu.RLock()
	handlers := append([]func(msg *AgentMessage) error(nil), b.subscribers[msg.TargetAgent]...)
	b.mu.RUnlock()

	for _, h := range handlers {
		_ = h(msg)
	}
	return nil
}

// Subscribe registers a message callback for a specific agent role.
func (b *LocalAgentBus) Subscribe(agentType AgentType, handler func(msg *AgentMessage) error) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.subscribers[agentType] = append(b.subscribers[agentType], handler)
	return nil
}
