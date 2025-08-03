package types

import (
	"errors"
	"log"
	"os"
	"time"
)

// Error definitions for the cron package
var (
	// ErrInvalidSchedule is returned when a cron schedule expression is invalid
	ErrInvalidSchedule = errors.New("invalid cron schedule expression")

	// ErrJobNotFound is returned when trying to access a non-existent job
	ErrJobNotFound = errors.New("job not found")

	// ErrJobAlreadyExists is returned when trying to add a job with a name that already exists
	ErrJobAlreadyExists = errors.New("job with this name already exists")

	// ErrSchedulerNotStarted is returned when trying to perform operations on a stopped scheduler
	ErrSchedulerNotStarted = errors.New("scheduler is not started")

	// ErrSchedulerAlreadyStarted is returned when trying to start an already running scheduler
	ErrSchedulerAlreadyStarted = errors.New("scheduler is already started")

	// ErrSchedulerStopped is returned when the scheduler has been stopped
	ErrSchedulerStopped = errors.New("scheduler has been stopped")

	// ErrJobTimeout is returned when a job exceeds its configured timeout
	ErrJobTimeout = errors.New("job execution timeout")

	// ErrMaxRetriesExceeded is returned when a job fails after all retry attempts
	ErrMaxRetriesExceeded = errors.New("maximum retry attempts exceeded")

	// ErrInvalidConfiguration is returned when the scheduler configuration is invalid
	ErrInvalidConfiguration = errors.New("invalid scheduler configuration")
)

// JobFunc represents a job function without error handling
type JobFunc func()

// JobFuncWithError represents a job function that can return an error
type JobFuncWithError func() error

// ErrorHandler handles job execution errors
type ErrorHandler func(error)

// JobConfig contains configuration options for individual jobs
type JobConfig struct {
	MaxRetries    int           // Maximum number of retry attempts
	RetryInterval time.Duration // Delay between retry attempts
	Timeout       time.Duration // Maximum execution time per job
}

// DefaultJobConfig returns a job configuration with sensible defaults
func DefaultJobConfig() JobConfig {
	return JobConfig{
		MaxRetries:    0,
		RetryInterval: time.Minute,
		Timeout:       0, // No timeout by default
	}
}

// Config contains configuration options for the scheduler
type Config struct {
	Logger            *log.Logger    // Logger instance for scheduler events
	Timezone          *time.Location // Timezone for schedule evaluation
	MaxConcurrentJobs int            // Maximum number of concurrent job executions
	EnableMonitoring  bool           // Enable HTTP monitoring endpoint
	MonitoringPort    int            // Port for monitoring HTTP server
}

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() Config {
	return Config{
		Logger:            log.New(os.Stdout, "CRON: ", log.LstdFlags),
		Timezone:          time.Local,
		MaxConcurrentJobs: 100,
		EnableMonitoring:  false,
		MonitoringPort:    8080,
	}
}

// Validate validates the configuration
func (c Config) Validate() error {
	if c.MaxConcurrentJobs <= 0 {
		return ErrInvalidConfiguration
	}
	// Only validate monitoring port if monitoring is enabled
	if c.EnableMonitoring && (c.MonitoringPort <= 0 || c.MonitoringPort > 65535) {
		return ErrInvalidConfiguration
	}
	if c.Logger == nil {
		return ErrInvalidConfiguration
	}
	if c.Timezone == nil {
		return ErrInvalidConfiguration
	}
	return nil
}

// Validate validates the job configuration
func (jc JobConfig) Validate() error {
	if jc.MaxRetries < 0 {
		return ErrInvalidConfiguration
	}
	if jc.RetryInterval < 0 {
		return ErrInvalidConfiguration
	}
	if jc.Timeout < 0 {
		return ErrInvalidConfiguration
	}
	return nil
}

// Stats represents scheduler statistics
type Stats struct {
	TotalJobs            int           // Total number of scheduled jobs
	RunningJobs          int           // Number of currently running jobs
	SuccessfulExecutions int64         // Total successful job executions
	FailedExecutions     int64         // Total failed job executions
	AverageExecutionTime time.Duration // Average execution time across all jobs
	Uptime               time.Duration // Scheduler uptime
}

// JobStats represents statistics for a specific job
type JobStats struct {
	Name          string        // Job name
	Schedule      string        // Cron schedule expression
	LastExecution time.Time     // Time of last execution
	NextExecution time.Time     // Time of next scheduled execution
	Status        string        // Current job status
	ExecutionTime time.Duration // Duration of last execution
	SuccessRate   float64       // Success rate percentage
	TotalRuns     int64         // Total number of runs
}

// Job status constants
const (
	JobStatusPending   = "pending"
	JobStatusRunning   = "running"
	JobStatusCompleted = "completed"
	JobStatusFailed    = "failed"
	JobStatusCancelled = "cancelled"
)
