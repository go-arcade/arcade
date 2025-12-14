// Copyright 2025 Arcade Team
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package metrics

import (
	"fmt"
	"sync"

	"github.com/go-arcade/arcade/pkg/log"
	"github.com/hibiken/asynq"
	"github.com/prometheus/client_golang/prometheus"
)

// AsynqMetricsCollector collects Asynq queue metrics for Prometheus
type AsynqMetricsCollector struct {
	inspector *asynq.Inspector
	mu        sync.RWMutex

	// Queue metrics
	queueSize           *prometheus.Desc
	queuePending        *prometheus.Desc
	queueActive         *prometheus.Desc
	queueScheduled      *prometheus.Desc
	queueRetry          *prometheus.Desc
	queueArchived       *prometheus.Desc
	queueCompleted      *prometheus.Desc
	queueAggregating    *prometheus.Desc
	queueProcessedTotal *prometheus.Desc
	queueFailedTotal    *prometheus.Desc
}

// NewAsynqMetricsCollector creates a new Task Queue metrics collector
func NewAsynqMetricsCollector(inspector *asynq.Inspector) *AsynqMetricsCollector {
	return &AsynqMetricsCollector{
		inspector: inspector,
		queueSize: prometheus.NewDesc(
			"asynq_queue_size",
			"Current size of the queue",
			[]string{"queue"},
			nil,
		),
		queuePending: prometheus.NewDesc(
			"asynq_queue_pending",
			"Number of pending tasks in the queue",
			[]string{"queue"},
			nil,
		),
		queueActive: prometheus.NewDesc(
			"asynq_queue_active",
			"Number of active tasks in the queue",
			[]string{"queue"},
			nil,
		),
		queueScheduled: prometheus.NewDesc(
			"asynq_queue_scheduled",
			"Number of scheduled tasks in the queue",
			[]string{"queue"},
			nil,
		),
		queueRetry: prometheus.NewDesc(
			"asynq_queue_retry",
			"Number of tasks in retry state",
			[]string{"queue"},
			nil,
		),
		queueArchived: prometheus.NewDesc(
			"asynq_queue_archived",
			"Number of archived tasks",
			[]string{"queue"},
			nil,
		),
		queueCompleted: prometheus.NewDesc(
			"asynq_queue_completed",
			"Number of completed tasks",
			[]string{"queue"},
			nil,
		),
		queueAggregating: prometheus.NewDesc(
			"asynq_queue_aggregating",
			"Number of aggregating tasks",
			[]string{"queue"},
			nil,
		),
		queueProcessedTotal: prometheus.NewDesc(
			"asynq_queue_processed_total",
			"Total number of processed tasks",
			[]string{"queue"},
			nil,
		),
		queueFailedTotal: prometheus.NewDesc(
			"asynq_queue_failed_total",
			"Total number of failed tasks",
			[]string{"queue"},
			nil,
		),
	}
}

// Describe implements prometheus.Collector interface
func (c *AsynqMetricsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.queueSize
	ch <- c.queuePending
	ch <- c.queueActive
	ch <- c.queueScheduled
	ch <- c.queueRetry
	ch <- c.queueArchived
	ch <- c.queueCompleted
	ch <- c.queueAggregating
	ch <- c.queueProcessedTotal
	ch <- c.queueFailedTotal
}

// Collect implements prometheus.Collector interface
func (c *AsynqMetricsCollector) Collect(ch chan<- prometheus.Metric) {
	c.mu.RLock()
	inspector := c.inspector
	c.mu.RUnlock()

	if inspector == nil {
		log.Debug("Asynq inspector is nil, skipping metrics collection")
		return
	}

	// Get all queues
	queues, err := inspector.Queues()
	if err != nil {
		log.Warnw("Failed to get queues for metrics", "error", err)
		// Even if we can't get queues, we should still report zero metrics for known queues
		// This ensures the metrics are always present
		return
	}

	// If no queues exist, still report zero metrics for default queues
	if len(queues) == 0 {
		// log.Debug("No queues found, reporting zero metrics for default queues")
		defaultQueues := []string{"critical", "default", "low"}
		for _, queueName := range defaultQueues {
			c.emitZeroMetrics(ch, queueName)
		}
		return
	}

	// Collect metrics for each queue
	for _, queueName := range queues {
		info, err := inspector.GetQueueInfo(queueName)
		if err != nil {
			log.Warnw("Failed to get queue info", "queue", queueName, "error", err)
			// Report zero metrics if we can't get info
			c.emitZeroMetrics(ch, queueName)
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			c.queueSize,
			prometheus.GaugeValue,
			float64(info.Size),
			queueName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.queuePending,
			prometheus.GaugeValue,
			float64(info.Pending),
			queueName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.queueActive,
			prometheus.GaugeValue,
			float64(info.Active),
			queueName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.queueScheduled,
			prometheus.GaugeValue,
			float64(info.Scheduled),
			queueName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.queueRetry,
			prometheus.GaugeValue,
			float64(info.Retry),
			queueName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.queueArchived,
			prometheus.GaugeValue,
			float64(info.Archived),
			queueName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.queueCompleted,
			prometheus.GaugeValue,
			float64(info.Completed),
			queueName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.queueAggregating,
			prometheus.GaugeValue,
			float64(info.Aggregating),
			queueName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.queueProcessedTotal,
			prometheus.CounterValue,
			float64(info.Processed),
			queueName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.queueFailedTotal,
			prometheus.CounterValue,
			float64(info.Failed),
			queueName,
		)
	}
}

// emitZeroMetrics emits zero values for all metrics for a given queue
func (c *AsynqMetricsCollector) emitZeroMetrics(ch chan<- prometheus.Metric, queueName string) {
	ch <- prometheus.MustNewConstMetric(c.queueSize, prometheus.GaugeValue, 0, queueName)
	ch <- prometheus.MustNewConstMetric(c.queuePending, prometheus.GaugeValue, 0, queueName)
	ch <- prometheus.MustNewConstMetric(c.queueActive, prometheus.GaugeValue, 0, queueName)
	ch <- prometheus.MustNewConstMetric(c.queueScheduled, prometheus.GaugeValue, 0, queueName)
	ch <- prometheus.MustNewConstMetric(c.queueRetry, prometheus.GaugeValue, 0, queueName)
	ch <- prometheus.MustNewConstMetric(c.queueArchived, prometheus.GaugeValue, 0, queueName)
	ch <- prometheus.MustNewConstMetric(c.queueCompleted, prometheus.GaugeValue, 0, queueName)
	ch <- prometheus.MustNewConstMetric(c.queueAggregating, prometheus.GaugeValue, 0, queueName)
	ch <- prometheus.MustNewConstMetric(c.queueProcessedTotal, prometheus.CounterValue, 0, queueName)
	ch <- prometheus.MustNewConstMetric(c.queueFailedTotal, prometheus.CounterValue, 0, queueName)
}

// asynqMetricsOnce ensures metrics are registered only once
var asynqMetricsOnce sync.Once

// SetupAsynqMetrics sets up Task Queue metrics collection
func SetupAsynqMetrics(registry *prometheus.Registry, inspector *asynq.Inspector) error {
	if inspector == nil {
		return nil // Skip if inspector is not available
	}

	collector := NewAsynqMetricsCollector(inspector)
	return registry.Register(collector)
}

// RegisterAsynqMetrics registers Task Queue metrics collector
func RegisterAsynqMetrics(registry *prometheus.Registry, inspector *asynq.Inspector) error {
	if registry == nil {
		return fmt.Errorf("registry is nil")
	}
	if inspector == nil {
		return fmt.Errorf("inspector is nil")
	}

	var err error
	asynqMetricsOnce.Do(func() {
		err = SetupAsynqMetrics(registry, inspector)
		if err != nil {
			log.Errorw("Failed to setup Task Queue metrics", "error", err)
		}
	})
	return err
}

// RegisterAsynqMetricsFromQueueServer registers Task Queue metrics from queue server
func RegisterAsynqMetricsFromQueueServer(registry *prometheus.Registry, queueServer interface{}) {
	if registry == nil {
		log.Warn("Metrics registry is nil, cannot register Task Queue metrics")
		return
	}
	if queueServer == nil {
		log.Warn("Queue server is nil, cannot register Task Queue metrics")
		return
	}

	// Use type assertion to get RedisConnOpt
	// This avoids importing queue package directly
	type QueueServerWithRedis interface {
		GetRedisConnOpt() asynq.RedisConnOpt
	}

	server, ok := queueServer.(QueueServerWithRedis)
	if !ok {
		log.Warn("Queue server does not implement GetRedisConnOpt method")
		return
	}

	redisOpt := server.GetRedisConnOpt()
	if redisOpt == nil {
		log.Warn("Redis connection option is nil")
		return
	}

	log.Infow("Creating Asynq inspector for metrics", "redisOpt", redisOpt)
	inspector := asynq.NewInspector(redisOpt)
	if inspector == nil {
		log.Error("Failed to create Asynq inspector")
		return
	}

	if err := RegisterAsynqMetrics(registry, inspector); err != nil {
		log.Errorw("Failed to register Task Queue metrics", "error", err)
	}
}

// RegisterAsynqMetricsFromQueueClient registers Task Queue metrics from queue client
func RegisterAsynqMetricsFromQueueClient(registry *prometheus.Registry, queueClient interface{}) {
	if registry == nil {
		log.Warn("Metrics registry is nil, cannot register Task Queue metrics")
		return
	}
	if queueClient == nil {
		log.Warn("Queue client is nil, cannot register Task Queue metrics")
		return
	}

	// Use type assertion to get RedisConnOpt
	// This avoids importing queue package directly
	type QueueClientWithRedis interface {
		GetRedisConnOpt() asynq.RedisConnOpt
	}

	client, ok := queueClient.(QueueClientWithRedis)
	if !ok {
		log.Warn("Queue client does not implement GetRedisConnOpt method")
		return
	}

	redisOpt := client.GetRedisConnOpt()
	if redisOpt == nil {
		log.Warn("Redis connection option is nil")
		return
	}

	log.Infow("Creating Asynq inspector for metrics", "redisOpt", redisOpt)
	inspector := asynq.NewInspector(redisOpt)
	if inspector == nil {
		log.Error("Failed to create Asynq inspector")
		return
	}

	if err := RegisterAsynqMetrics(registry, inspector); err != nil {
		log.Errorw("Failed to register Task Queue metrics", "error", err)
	}
}
