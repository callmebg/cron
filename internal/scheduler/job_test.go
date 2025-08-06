package scheduler

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/callmebg/cron/internal/types"
)

const (
	testJobID   = "test-id"
	testJobName = "test-job"
)

func TestNewJob(t *testing.T) {
	jobFunc := func() {
		// Test job function
	}

	job, err := NewJob("test-id", "test-job", "* * * * *", jobFunc, types.DefaultJobConfig(), time.UTC)
	if err != nil {
		t.Fatalf("NewJob failed: %v", err)
	}

	if job.ID != testJobID {
		t.Errorf("Job ID = %q; want %q", job.ID, testJobID)
	}

	if job.Name != testJobName {
		t.Errorf("Job Name = %q; want %q", job.Name, testJobName)
	}

	if job.Status != types.JobStatusPending {
		t.Errorf("Job Status = %q; want %q", job.Status, types.JobStatusPending)
	}

	if job.Function == nil {
		t.Error("Job Function is nil")
	}
}

func TestNewJobWithError(t *testing.T) {
	jobFunc := func() error {
		return nil
	}

	errorHandler := func(_ error) {
		// Handle error
	}

	job, err := NewJobWithError(testJobID, testJobName, "0 9 * * *", jobFunc, errorHandler, types.DefaultJobConfig(), time.UTC)
	if err != nil {
		t.Fatalf("NewJobWithError failed: %v", err)
	}

	if job.FunctionWithError == nil {
		t.Error("Job FunctionWithError is nil")
	}

	if job.ErrorHandler == nil {
		t.Error("Job ErrorHandler is nil")
	}
}

func TestJobShouldRun(t *testing.T) {
	jobFunc := func() {
		// Test job function
	}

	// Create job that runs every minute
	job, err := NewJob("test-id", "test-job", "* * * * *", jobFunc, types.DefaultJobConfig(), time.UTC)
	if err != nil {
		t.Fatalf("NewJob failed: %v", err)
	}

	// Set next execution to past time
	now := time.Now()
	job.NextExecution = now.Add(-time.Minute)

	if !job.ShouldRun(now) {
		t.Error("Job should run when next execution time has passed")
	}

	// Set next execution to future time
	job.NextExecution = now.Add(time.Minute)

	if job.ShouldRun(now) {
		t.Error("Job should not run when next execution time is in the future")
	}

	// Test running job
	job.IsRunning = true
	job.NextExecution = now.Add(-time.Minute)

	if job.ShouldRun(now) {
		t.Error("Job should not run when already running")
	}
}

func TestJobUpdateNextExecution(t *testing.T) {
	jobFunc := func() {
		// Test job function
	}

	job, err := NewJob(testJobID, testJobName, "0 9 * * *", jobFunc, types.DefaultJobConfig(), time.UTC)
	if err != nil {
		t.Fatalf("NewJob failed: %v", err)
	}

	now := time.Date(2024, 1, 1, 8, 0, 0, 0, time.UTC)
	job.UpdateNextExecution(now)

	expected := time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC)
	if !job.NextExecution.Equal(expected) {
		t.Errorf("NextExecution = %v; want %v", job.NextExecution, expected)
	}
}

func TestJobMarkRunning(t *testing.T) {
	jobFunc := func() {
		// Test job function
	}

	job, err := NewJob("test-id", "test-job", "* * * * *", jobFunc, types.DefaultJobConfig(), time.UTC)
	if err != nil {
		t.Fatalf("NewJob failed: %v", err)
	}

	job.MarkRunning()

	if !job.IsRunning {
		t.Error("Job should be marked as running")
	}

	if job.Status != types.JobStatusRunning {
		t.Errorf("Job Status = %q; want %q", job.Status, types.JobStatusRunning)
	}

	if job.LastExecution.IsZero() {
		t.Error("LastExecution should be set")
	}
}

func TestJobMarkCompleted(t *testing.T) {
	jobFunc := func() {
		// Test job function
	}

	job, err := NewJob("test-id", "test-job", "* * * * *", jobFunc, types.DefaultJobConfig(), time.UTC)
	if err != nil {
		t.Fatalf("NewJob failed: %v", err)
	}

	job.MarkRunning()
	duration := time.Second * 5
	job.MarkCompleted(duration, true)

	if job.IsRunning {
		t.Error("Job should not be running after completion")
	}

	if job.Status != types.JobStatusCompleted {
		t.Errorf("Job Status = %q; want %q", job.Status, types.JobStatusCompleted)
	}

	if job.LastDuration != duration {
		t.Errorf("LastDuration = %v; want %v", job.LastDuration, duration)
	}

	if job.RunCount != 1 {
		t.Errorf("RunCount = %d; want 1", job.RunCount)
	}

	if job.SuccessCount != 1 {
		t.Errorf("SuccessCount = %d; want 1", job.SuccessCount)
	}

	if job.FailureCount != 0 {
		t.Errorf("FailureCount = %d; want 0", job.FailureCount)
	}
}

func TestJobMarkCompletedFailed(t *testing.T) {
	jobFunc := func() {
		// Test job function
	}

	job, err := NewJob("test-id", "test-job", "* * * * *", jobFunc, types.DefaultJobConfig(), time.UTC)
	if err != nil {
		t.Fatalf("NewJob failed: %v", err)
	}

	job.MarkRunning()
	duration := time.Second * 3
	job.MarkCompleted(duration, false)

	if job.Status != types.JobStatusFailed {
		t.Errorf("Job Status = %q; want %q", job.Status, types.JobStatusFailed)
	}

	if job.SuccessCount != 0 {
		t.Errorf("SuccessCount = %d; want 0", job.SuccessCount)
	}

	if job.FailureCount != 1 {
		t.Errorf("FailureCount = %d; want 1", job.FailureCount)
	}
}

func TestJobExecute(t *testing.T) {
	executed := false
	jobFunc := func() {
		executed = true
	}

	job, err := NewJob("test-id", "test-job", "* * * * *", jobFunc, types.DefaultJobConfig(), time.UTC)
	if err != nil {
		t.Fatalf("NewJob failed: %v", err)
	}

	ctx := context.Background()
	err = job.Execute(ctx)
	if err != nil {
		t.Errorf("Execute failed: %v", err)
	}

	if !executed {
		t.Error("Job function was not executed")
	}

	if job.IsRunning {
		t.Error("Job should not be running after execution")
	}

	if job.RunCount != 1 {
		t.Errorf("RunCount = %d; want 1", job.RunCount)
	}
}

func TestJobExecuteWithError(t *testing.T) {
	executed := false
	errorHandled := false
	testError := types.ErrJobTimeout

	jobFunc := func() error {
		executed = true
		return testError
	}

	errorHandler := func(_ error) {
		errorHandled = true
		if err != testError {
			t.Errorf("Error handler received %v; want %v", err, testError)
		}
	}

	job, err := NewJobWithError("test-id", "test-job", "* * * * *", jobFunc, errorHandler, types.DefaultJobConfig(), time.UTC)
	if err != nil {
		t.Fatalf("NewJobWithError failed: %v", err)
	}

	ctx := context.Background()
	err = job.Execute(ctx)
	if err != testError {
		t.Errorf("Execute returned %v; want %v", err, testError)
	}

	if !executed {
		t.Error("Job function was not executed")
	}

	if !errorHandled {
		t.Error("Error handler was not called")
	}
}

func TestJobExecuteWithTimeout(t *testing.T) {
	jobFunc := func() {
		time.Sleep(time.Second * 2) // Sleep longer than timeout
	}

	config := types.JobConfig{
		Timeout: time.Millisecond * 100, // Short timeout
	}

	job, err := NewJob("test-id", "test-job", "* * * * *", jobFunc, config, time.UTC)
	if err != nil {
		t.Fatalf("NewJob failed: %v", err)
	}

	ctx := context.Background()
	start := time.Now()
	err = job.Execute(ctx)
	duration := time.Since(start)

	if err == nil {
		t.Error("Execute should have timed out")
	}

	// Should complete quickly due to timeout
	if duration > time.Second {
		t.Errorf("Execute took %v; should have timed out quickly", duration)
	}
}

func TestJobGetStats(t *testing.T) {
	jobFunc := func() {
		// Test job function
	}

	job, err := NewJob(testJobID, testJobName, "0 9 * * *", jobFunc, types.DefaultJobConfig(), time.UTC)
	if err != nil {
		t.Fatalf("NewJob failed: %v", err)
	}

	// Simulate some runs
	job.RunCount = 10
	job.SuccessCount = 8
	job.FailureCount = 2
	job.LastDuration = time.Second * 5

	stats := job.GetStats()

	if stats.Name != testJobName {
		t.Errorf("Stats Name = %q; want %q", stats.Name, testJobName)
	}

	if stats.TotalRuns != 10 {
		t.Errorf("Stats TotalRuns = %d; want 10", stats.TotalRuns)
	}

	expectedSuccessRate := 80.0 // 8/10 * 100
	if stats.SuccessRate != expectedSuccessRate {
		t.Errorf("Stats SuccessRate = %f; want %f", stats.SuccessRate, expectedSuccessRate)
	}

	if stats.ExecutionTime != time.Second*5 {
		t.Errorf("Stats ExecutionTime = %v; want %v", stats.ExecutionTime, time.Second*5)
	}
}

func TestJobExecuteWithRetry(t *testing.T) {
	attempts := 0
	jobFunc := func() error {
		attempts++
		if attempts < 3 {
			return types.ErrJobTimeout // Fail first 2 attempts
		}
		return nil // Succeed on 3rd attempt
	}

	errorHandler := func(_ error) {
		// Handle error
	}

	config := types.JobConfig{
		MaxRetries:    3,
		RetryInterval: time.Millisecond * 10, // Short interval for testing
	}

	job, err := NewJobWithError("test-id", "test-job", "* * * * *", jobFunc, errorHandler, config, time.UTC)
	if err != nil {
		t.Fatalf("NewJobWithError failed: %v", err)
	}

	ctx := context.Background()
	err = job.Execute(ctx)
	if err != nil {
		t.Errorf("Execute failed after retries: %v", err)
	}

	if attempts != 3 {
		t.Errorf("Job was attempted %d times; want 3", attempts)
	}
}

func FuzzJobScheduling(f *testing.F) {
	// Seed with valid cron expressions
	f.Add("* * * * *", "test-job")
	f.Add("0 0 * * *", "daily-job")
	f.Add("*/5 * * * *", "every-5-min")
	f.Add("0 9-17 * * 1-5", "work-hours")
	f.Add("@hourly", "hourly-job")
	f.Add("@daily", "daily-job")

	f.Fuzz(func(t *testing.T, cronExpr, jobName string) {
		// Sanitize job name to avoid control characters
		jobName = strings.ReplaceAll(jobName, "\n", "")
		jobName = strings.ReplaceAll(jobName, "\r", "")
		jobName = strings.ReplaceAll(jobName, "\t", "")

		if len(jobName) > 100 {
			jobName = jobName[:100] // Limit length
		}

		if jobName == "" {
			jobName = "fuzz-job"
		}

		jobFunc := func() {
			// Simple job that doesn't crash
		}

		// Try to create a job - should not crash
		job, err := NewJob("fuzz-id", jobName, cronExpr, jobFunc, types.DefaultJobConfig(), time.UTC)

		if err != nil {
			// If job creation fails due to invalid cron expression, that's fine
			return
		}

		if job == nil {
			t.Error("NewJob returned nil job without error")
			return
		}

		// Test basic job operations don't crash
		now := time.Now()

		// Test ShouldRun
		_ = job.ShouldRun(now)

		// Test UpdateNextExecution
		job.UpdateNextExecution(now)

		// Test MarkRunning
		job.MarkRunning()

		// Test MarkCompleted
		job.MarkCompleted(time.Millisecond*100, true)

		// Test GetStats
		_ = job.GetStats()

		// Test Execute (with timeout to prevent infinite execution)
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		_ = job.Execute(ctx)
	})
}

func FuzzJobWithError(f *testing.F) {
	// Seed with valid patterns
	f.Add("* * * * *", "error-job", true)
	f.Add("0 0 * * *", "panic-job", false)
	f.Add("*/5 * * * *", "timeout-job", true)

	f.Fuzz(func(t *testing.T, cronExpr, jobName string, shouldPanic bool) {
		// Sanitize inputs
		jobName = strings.ReplaceAll(jobName, "\n", "")
		jobName = strings.ReplaceAll(jobName, "\r", "")
		if len(jobName) > 50 {
			jobName = jobName[:50]
		}
		if jobName == "" {
			jobName = "fuzz-error-job"
		}

		var jobFunc func() error
		var errorHandler func(error)

		// Add panic recovery to prevent fuzz test from crashing
		defer func() {
			if r := recover(); r != nil {
				// Expected panic in fuzz test, just continue
				t.Logf("Recovered from panic: %v", r)
			}
		}()

		if shouldPanic {
			jobFunc = func() error {
				panic("fuzz panic test")
			}
		} else {
			jobFunc = func() error {
				return fmt.Errorf("fuzz error test")
			}
		}

		errorHandler = func(_ error) {
			// Handle error - should not crash
		}

		// Try to create a job with error handling
		job, err := NewJobWithError("fuzz-error-id", jobName, cronExpr, jobFunc, errorHandler, types.DefaultJobConfig(), time.UTC)

		if err != nil {
			// If job creation fails due to invalid cron expression, that's fine
			return
		}

		if job == nil {
			t.Error("NewJobWithError returned nil job without error")
			return
		}

		// Test execution with panic/error recovery
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		// Should not crash even with panic/error
		_ = job.Execute(ctx)

		// Test that job state is still valid after error
		_ = job.GetStats()
		_ = job.ShouldRun(time.Now())
	})
}

func FuzzJobConfiguration(f *testing.F) {
	// Seed with various configurations
	f.Add(int64(1000), int64(100), 3, true) // timeout, retry interval, max retries, allow overlap
	f.Add(int64(5000), int64(500), 1, false)
	f.Add(int64(100), int64(10), 5, true)

	f.Fuzz(func(t *testing.T, timeoutMs, retryMs int64, maxRetries int, _ bool) {
		// Ensure reasonable bounds
		if timeoutMs < 0 {
			timeoutMs = -timeoutMs
		}
		if retryMs < 0 {
			retryMs = -retryMs
		}
		if maxRetries < 0 {
			maxRetries = -maxRetries
		}

		// Cap values to reasonable limits
		if timeoutMs > 60000 {
			timeoutMs = 60000 // Max 1 minute
		}
		if retryMs > 10000 {
			retryMs = 10000 // Max 10 seconds
		}
		if maxRetries > 10 {
			maxRetries = 10
		}

		config := types.JobConfig{
			Timeout:       time.Duration(timeoutMs) * time.Millisecond,
			RetryInterval: time.Duration(retryMs) * time.Millisecond,
			MaxRetries:    maxRetries,
		}

		jobFunc := func() {
			time.Sleep(time.Duration(timeoutMs/10) * time.Millisecond) // Sleep for fraction of timeout
		}

		// Try to create job with fuzzed configuration
		job, err := NewJob("fuzz-config-id", "fuzz-config-job", "* * * * *", jobFunc, config, time.UTC)

		if err != nil {
			return // Invalid configuration is fine
		}

		if job == nil {
			t.Error("NewJob returned nil job without error")
			return
		}

		// Test that job operations work with various configurations
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutMs*2)*time.Millisecond)
		defer cancel()

		// Execute job and verify it doesn't crash
		_ = job.Execute(ctx)

		// Verify job stats are still accessible
		stats := job.GetStats()
		if stats.Name != "fuzz-config-job" {
			t.Errorf("Expected job name 'fuzz-config-job', got %q", stats.Name)
		}
	})
}
