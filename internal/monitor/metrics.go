package monitor

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/callmebg/cron/internal/types"
)

// Metrics holds all monitoring metrics for the scheduler
type Metrics struct {
	// Basic counters
	totalJobs     int64
	runningJobs   int64
	completedJobs int64
	failedJobs    int64

	// Timing metrics
	totalExecutionTime int64 // nanoseconds
	minExecutionTime   int64 // nanoseconds
	maxExecutionTime   int64 // nanoseconds

	// System metrics
	startTime        time.Time
	lastJobExecution time.Time

	// Thread safety
	mu sync.RWMutex
}

// NewMetrics creates a new metrics collector
func NewMetrics() *Metrics {
	return &Metrics{
		startTime:        time.Now(),
		minExecutionTime: ^int64(0), // max int64 as initial min
	}
}

// RecordJobStart records when a job starts execution
func (m *Metrics) RecordJobStart() {
	atomic.AddInt64(&m.totalJobs, 1)
	atomic.AddInt64(&m.runningJobs, 1)
}

// RecordJobCompletion records when a job completes successfully
func (m *Metrics) RecordJobCompletion(duration time.Duration) {
	atomic.AddInt64(&m.runningJobs, -1)
	atomic.AddInt64(&m.completedJobs, 1)

	durationNanos := duration.Nanoseconds()
	atomic.AddInt64(&m.totalExecutionTime, durationNanos)

	// Update min/max execution times
	m.updateExecutionTimeBounds(durationNanos)

	m.mu.Lock()
	m.lastJobExecution = time.Now()
	m.mu.Unlock()
}

// RecordJobFailure records when a job fails
func (m *Metrics) RecordJobFailure(duration time.Duration) {
	atomic.AddInt64(&m.runningJobs, -1)
	atomic.AddInt64(&m.failedJobs, 1)

	durationNanos := duration.Nanoseconds()
	atomic.AddInt64(&m.totalExecutionTime, durationNanos)

	// Update min/max execution times
	m.updateExecutionTimeBounds(durationNanos)

	m.mu.Lock()
	m.lastJobExecution = time.Now()
	m.mu.Unlock()
}

// updateExecutionTimeBounds atomically updates min/max execution times
func (m *Metrics) updateExecutionTimeBounds(duration int64) {
	// Update minimum
	for {
		current := atomic.LoadInt64(&m.minExecutionTime)
		if duration >= current {
			break
		}
		if atomic.CompareAndSwapInt64(&m.minExecutionTime, current, duration) {
			break
		}
	}

	// Update maximum
	for {
		current := atomic.LoadInt64(&m.maxExecutionTime)
		if duration <= current {
			break
		}
		if atomic.CompareAndSwapInt64(&m.maxExecutionTime, current, duration) {
			break
		}
	}
}

// GetSnapshot returns a snapshot of current metrics
func (m *Metrics) GetSnapshot() MetricsSnapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()

	totalJobs := atomic.LoadInt64(&m.totalJobs)
	completedJobs := atomic.LoadInt64(&m.completedJobs)
	failedJobs := atomic.LoadInt64(&m.failedJobs)
	totalExecutionTime := atomic.LoadInt64(&m.totalExecutionTime)
	minExecutionTime := atomic.LoadInt64(&m.minExecutionTime)
	maxExecutionTime := atomic.LoadInt64(&m.maxExecutionTime)

	uptime := time.Since(m.startTime)

	// Calculate rates
	var jobsPerMinute float64
	if uptime.Minutes() > 0 {
		jobsPerMinute = float64(totalJobs) / uptime.Minutes()
	}

	var errorRate float64
	if totalJobs > 0 {
		errorRate = float64(failedJobs) / float64(totalJobs) * 100
	}

	var avgExecutionTime time.Duration
	if totalJobs > 0 {
		avgExecutionTime = time.Duration(totalExecutionTime / totalJobs)
	}

	// Handle case where no jobs have run yet
	minDuration := time.Duration(0)
	if minExecutionTime != ^int64(0) {
		minDuration = time.Duration(minExecutionTime)
	}

	return MetricsSnapshot{
		TotalJobs:            totalJobs,
		RunningJobs:          atomic.LoadInt64(&m.runningJobs),
		CompletedJobs:        completedJobs,
		FailedJobs:           failedJobs,
		JobsPerMinute:        jobsPerMinute,
		ErrorRate:            errorRate,
		Uptime:               uptime,
		AverageExecutionTime: avgExecutionTime,
		MinExecutionTime:     minDuration,
		MaxExecutionTime:     time.Duration(maxExecutionTime),
		LastJobExecution:     m.lastJobExecution,
	}
}

// GetStats converts metrics to types.Stats format
func (m *Metrics) GetStats() types.Stats {
	snapshot := m.GetSnapshot()
	return types.Stats{
		TotalJobs:            int(snapshot.TotalJobs),
		RunningJobs:          int(snapshot.RunningJobs),
		SuccessfulExecutions: snapshot.CompletedJobs,
		FailedExecutions:     snapshot.FailedJobs,
		AverageExecutionTime: snapshot.AverageExecutionTime,
		Uptime:               snapshot.Uptime,
	}
}

// Reset resets all metrics (useful for testing)
func (m *Metrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	atomic.StoreInt64(&m.totalJobs, 0)
	atomic.StoreInt64(&m.runningJobs, 0)
	atomic.StoreInt64(&m.completedJobs, 0)
	atomic.StoreInt64(&m.failedJobs, 0)
	atomic.StoreInt64(&m.totalExecutionTime, 0)
	atomic.StoreInt64(&m.minExecutionTime, ^int64(0))
	atomic.StoreInt64(&m.maxExecutionTime, 0)

	m.startTime = time.Now()
	m.lastJobExecution = time.Time{}
}

// MetricsSnapshot represents a point-in-time snapshot of metrics
type MetricsSnapshot struct {
	TotalJobs            int64         `json:"total_jobs"`
	RunningJobs          int64         `json:"running_jobs"`
	CompletedJobs        int64         `json:"completed_jobs"`
	FailedJobs           int64         `json:"failed_jobs"`
	JobsPerMinute        float64       `json:"jobs_per_minute"`
	ErrorRate            float64       `json:"error_rate_percent"`
	Uptime               time.Duration `json:"uptime"`
	AverageExecutionTime time.Duration `json:"average_execution_time"`
	MinExecutionTime     time.Duration `json:"min_execution_time"`
	MaxExecutionTime     time.Duration `json:"max_execution_time"`
	LastJobExecution     time.Time     `json:"last_job_execution"`
}
