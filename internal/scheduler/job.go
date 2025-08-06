package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/callmebg/cron/internal/parser"
	"github.com/callmebg/cron/internal/types"
)

// Job represents a scheduled job
type Job struct {
	ID                string
	Name              string
	Schedule          *parser.Schedule
	Function          types.JobFunc
	FunctionWithError types.JobFuncWithError
	ErrorHandler      types.ErrorHandler
	Config            types.JobConfig
	Status            string
	LastExecution     time.Time
	NextExecution     time.Time
	RunCount          int64
	SuccessCount      int64
	FailureCount      int64
	LastDuration      time.Duration
	IsRunning         bool
	CreatedAt         time.Time
	UpdatedAt         time.Time
	mu                sync.RWMutex
}

// NewJob creates a new job
func NewJob(id, name, scheduleExpr string, fn types.JobFunc, config types.JobConfig, timezone *time.Location) (*Job, error) {
	schedule, err := parser.ParseInLocation(scheduleExpr, timezone)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	job := &Job{
		ID:            id,
		Name:          name,
		Schedule:      schedule,
		Function:      fn,
		Config:        config,
		Status:        types.JobStatusPending,
		NextExecution: schedule.Next(now),
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	return job, nil
}

// NewJobWithError creates a new job with error handling
func NewJobWithError(
	id, name, scheduleExpr string,
	fn types.JobFuncWithError,
	errorHandler types.ErrorHandler,
	config types.JobConfig,
	timezone *time.Location,
) (*Job, error) {
	schedule, err := parser.ParseInLocation(scheduleExpr, timezone)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	job := &Job{
		ID:                id,
		Name:              name,
		Schedule:          schedule,
		FunctionWithError: fn,
		ErrorHandler:      errorHandler,
		Config:            config,
		Status:            types.JobStatusPending,
		NextExecution:     schedule.Next(now),
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	return job, nil
}

// GetStats returns job statistics
func (j *Job) GetStats() types.JobStats {
	j.mu.RLock()
	defer j.mu.RUnlock()

	var successRate float64
	if j.RunCount > 0 {
		successRate = float64(j.SuccessCount) / float64(j.RunCount) * 100
	}

	return types.JobStats{
		Name:          j.Name,
		Schedule:      j.GetScheduleExpression(),
		LastExecution: j.LastExecution,
		NextExecution: j.NextExecution,
		Status:        j.Status,
		ExecutionTime: j.LastDuration,
		SuccessRate:   successRate,
		TotalRuns:     j.RunCount,
	}
}

// GetScheduleExpression returns a string representation of the schedule
func (j *Job) GetScheduleExpression() string {
	// This is a simplified version - in a real implementation,
	// you might want to store the original expression or reconstruct it
	return "cron expression"
}

// ShouldRun checks if the job should run at the given time
func (j *Job) ShouldRun(t time.Time) bool {
	j.mu.RLock()
	defer j.mu.RUnlock()
	return !j.IsRunning && !j.NextExecution.After(t)
}

// UpdateNextExecution updates the next execution time
func (j *Job) UpdateNextExecution(after time.Time) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.NextExecution = j.Schedule.Next(after)
	j.UpdatedAt = time.Now()
}

// MarkRunning marks the job as running
func (j *Job) MarkRunning() {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.IsRunning = true
	j.Status = types.JobStatusRunning
	j.LastExecution = time.Now()
	j.UpdatedAt = time.Now()
}

// MarkCompleted marks the job as completed
func (j *Job) MarkCompleted(duration time.Duration, success bool) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.IsRunning = false
	j.LastDuration = duration
	j.RunCount++
	j.UpdatedAt = time.Now()

	if success {
		j.Status = types.JobStatusCompleted
		j.SuccessCount++
	} else {
		j.Status = types.JobStatusFailed
		j.FailureCount++
	}

	// Update next execution time
	j.NextExecution = j.Schedule.Next(j.LastExecution)
}

// IsJobRunning returns whether the job is currently running (thread-safe)
func (j *Job) IsJobRunning() bool {
	j.mu.RLock()
	defer j.mu.RUnlock()
	return j.IsRunning
}

// GetJobStats returns job statistics (thread-safe)
func (j *Job) GetJobStats() (duration time.Duration, count int64) {
	j.mu.RLock()
	defer j.mu.RUnlock()
	return j.LastDuration, j.RunCount
}

// Execute runs the job function
func (j *Job) Execute(ctx context.Context) error {
	start := time.Now()
	j.MarkRunning()

	defer func() {
		duration := time.Since(start)
		success := true
		j.MarkCompleted(duration, success)
	}()

	// Handle timeout if configured
	if j.Config.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, j.Config.Timeout)
		defer cancel()
	}

	// Execute the job function
	if j.FunctionWithError != nil {
		return j.executeWithRetry(ctx)
	}

	// Execute simple function in a goroutine to respect context cancellation
	done := make(chan struct{})
	go func() {
		defer close(done)
		j.Function()
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// executeWithRetry executes a job with retry logic
func (j *Job) executeWithRetry(ctx context.Context) error {
	var lastErr error
	maxAttempts := j.Config.MaxRetries + 1

	for attempt := 0; attempt < maxAttempts; attempt++ {
		if attempt > 0 {
			// Wait before retry
			select {
			case <-time.After(j.Config.RetryInterval):
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		// Execute the function
		done := make(chan error, 1)
		go func() {
			defer func() {
				if r := recover(); r != nil {
					done <- fmt.Errorf("panic in job execution: %v", r)
				}
			}()
			done <- j.FunctionWithError()
		}()

		select {
		case err := <-done:
			if err == nil {
				return nil // Success
			}
			lastErr = err
			if j.ErrorHandler != nil {
				j.ErrorHandler(err)
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return lastErr
}
