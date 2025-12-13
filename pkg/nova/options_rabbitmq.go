package nova

import (
	"crypto/tls"
	"time"
)

// RabbitMQOption is the interface for RabbitMQ-specific configuration options
type RabbitMQOption interface {
	apply(*RabbitMQConfig)
}

type rabbitmqOptionFunc func(*RabbitMQConfig)

func (f rabbitmqOptionFunc) apply(c *RabbitMQConfig) {
	f(c)
}

// WithRabbitMQExchange sets the RabbitMQ exchange name
func WithRabbitMQExchange(exchange string) RabbitMQOption {
	return rabbitmqOptionFunc(func(c *RabbitMQConfig) {
		c.Exchange = exchange
	})
}

// WithRabbitMQTopicPrefix sets the RabbitMQ topic prefix
func WithRabbitMQTopicPrefix(prefix string) RabbitMQOption {
	return rabbitmqOptionFunc(func(c *RabbitMQConfig) {
		c.TopicPrefix = prefix
	})
}

// WithRabbitMQPrefetch sets the RabbitMQ prefetch configuration
func WithRabbitMQPrefetch(count, size int) RabbitMQOption {
	return rabbitmqOptionFunc(func(c *RabbitMQConfig) {
		c.PrefetchCount = count
		c.PrefetchSize = size
	})
}

// WithRabbitMQAuth sets the RabbitMQ authentication configuration
func WithRabbitMQAuth(username, password string) RabbitMQOption {
	return rabbitmqOptionFunc(func(c *RabbitMQConfig) {
		c.Username = username
		c.Password = password
	})
}

// WithRabbitMQTLS sets the RabbitMQ TLS configuration
func WithRabbitMQTLS(tlsConfig *tls.Config) RabbitMQOption {
	return rabbitmqOptionFunc(func(c *RabbitMQConfig) {
		c.TLSConfig = tlsConfig
	})
}

// WithRabbitMQDelaySlots sets the RabbitMQ delay slot configuration
func WithRabbitMQDelaySlots(count int, duration time.Duration) RabbitMQOption {
	return rabbitmqOptionFunc(func(c *RabbitMQConfig) {
		c.DelaySlotCount = count
		c.DelaySlotDuration = duration
	})
}
