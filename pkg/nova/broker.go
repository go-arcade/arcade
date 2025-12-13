package nova

import (
	"context"
)

// QueueType represents the message queue type
type QueueType string

const (
	QueueTypeKafka    QueueType = "kafka"
	QueueTypeRocketMQ QueueType = "rocketmq"
	QueueTypeRabbitMQ QueueType = "rabbitmq"
)

// MessageQueueBroker is the interface for message queue brokers
// All message queue implementations must implement this interface
type MessageQueueBroker interface {
	// SendMessage sends a single message
	SendMessage(ctx context.Context, topic string, key string, value []byte, headers map[string]string) error

	// SendBatchMessages sends multiple messages in batch
	SendBatchMessages(ctx context.Context, topic string, messages []Message) error

	// Subscribe subscribes to topics and consumes messages
	Subscribe(ctx context.Context, topics []string, handler MessageHandler) error

	// Close closes the connection
	Close() error
}

// Message represents a message structure
type Message struct {
	Key     string
	Value   []byte
	Headers map[string]string
}

// MessageHandler is the function type for message handlers
type MessageHandler func(ctx context.Context, msg *Message) error

// BrokerConfig is the interface for broker configuration
type BrokerConfig interface {
	GetType() QueueType
	GetBootstrapServers() string
	GetGroupID() string
	GetTopicPrefix() string
}
