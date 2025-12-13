package metrics

import (
	"sync"
	"time"

	"github.com/go-arcade/arcade/pkg/cron"
	"github.com/prometheus/client_golang/prometheus"
)

// CronMetricsRecorder implements cron.MetricsRecorder interface
type CronMetricsRecorder struct{}

var (
	// CronJobRunsTotal counts the total number of cron job runs
	CronJobRunsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cron_job_runs_total",
			Help: "Total number of cron job runs",
		},
		[]string{"job_name"},
	)

	// CronJobRunDurationSeconds measures the duration of cron job runs
	CronJobRunDurationSeconds = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "cron_job_run_duration_seconds",
			Help:    "Duration of cron job runs in seconds",
			Buckets: prometheus.ExponentialBuckets(0.001, 2, 15), // 1ms to ~32s
		},
		[]string{"job_name"},
	)

	// CronJobErrorsTotal counts the total number of cron job errors
	CronJobErrorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cron_job_errors_total",
			Help: "Total number of cron job errors",
		},
		[]string{"job_name"},
	)

	// CronJobLastRunTime records the last run time of each cron job
	CronJobLastRunTime = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cron_job_last_run_time_seconds",
			Help: "Last run time of cron job in seconds since epoch",
		},
		[]string{"job_name"},
	)

	// CronJobNextRunTime records the next scheduled run time of each cron job
	CronJobNextRunTime = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cron_job_next_run_time_seconds",
			Help: "Next scheduled run time of cron job in seconds since epoch",
		},
		[]string{"job_name"},
	)

	// CronJobsTotal counts the total number of registered cron jobs
	CronJobsTotal = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "cron_jobs_total",
			Help: "Total number of registered cron jobs",
		},
	)

	// cronMetricsOnce ensures metrics are registered only once
	cronMetricsOnce sync.Once
)

// NewCronMetricsRecorder creates a new cron metrics recorder
func NewCronMetricsRecorder() *CronMetricsRecorder {
	return &CronMetricsRecorder{}
}

// RecordJobRun records a cron job run
func (r *CronMetricsRecorder) RecordJobRun(jobName string, duration time.Duration, err error) {
	RecordCronJobRun(jobName, duration, err)
}

// UpdateNextRun updates the next run time for a cron job
func (r *CronMetricsRecorder) UpdateNextRun(jobName string, nextRun time.Time) {
	UpdateCronJobNextRun(jobName, nextRun)
}

// UpdateJobsCount updates the total number of registered cron jobs
func (r *CronMetricsRecorder) UpdateJobsCount(count int) {
	UpdateCronJobsCount(count)
}

// SetupCronMetrics sets up cron metrics recording
func SetupCronMetrics(registry *prometheus.Registry) {
	RegisterCronMetrics(registry)
	cron.SetMetricsRecorder(NewCronMetricsRecorder())
}

// RegisterCronMetrics registers all cron-related metrics
func RegisterCronMetrics(registry *prometheus.Registry) {
	cronMetricsOnce.Do(func() {
		registry.MustRegister(
			CronJobRunsTotal,
			CronJobRunDurationSeconds,
			CronJobErrorsTotal,
			CronJobLastRunTime,
			CronJobNextRunTime,
			CronJobsTotal,
		)
	})
}

// RecordCronJobRun records a cron job run
func RecordCronJobRun(jobName string, duration time.Duration, err error) {
	if err != nil {
		CronJobErrorsTotal.WithLabelValues(jobName).Inc()
	}
	CronJobRunsTotal.WithLabelValues(jobName).Inc()
	CronJobRunDurationSeconds.WithLabelValues(jobName).Observe(duration.Seconds())
	CronJobLastRunTime.WithLabelValues(jobName).Set(float64(time.Now().Unix()))
}

// UpdateCronJobNextRun updates the next run time for a cron job
func UpdateCronJobNextRun(jobName string, nextRun time.Time) {
	if !nextRun.IsZero() {
		CronJobNextRunTime.WithLabelValues(jobName).Set(float64(nextRun.Unix()))
	}
}

// UpdateCronJobsCount updates the total number of registered cron jobs
func UpdateCronJobsCount(count int) {
	CronJobsTotal.Set(float64(count))
}
