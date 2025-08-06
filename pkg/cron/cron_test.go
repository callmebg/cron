package cron

import (
	"context"
	"sync"
	"testing"
	"time"
)

const (
	testJobNamePkg = "test-job"
)

func TestNew(t *testing.T) {
	scheduler := New()
	if scheduler == nil {
		t.Fatal("New() returned nil")
	}

	if scheduler.IsRunning() {
		t.Error("New scheduler should not be running")
	}

	if len(scheduler.ListJobs()) != 0 {
		t.Error("New scheduler should have no jobs")
	}
}

func TestNewWithConfig(t *testing.T) {
	config := DefaultConfig()
	config.MaxConcurrentJobs = 50

	scheduler := NewWithConfig(config)
	if scheduler == nil {
		t.Fatal("NewWithConfig() returned nil")
	}

	if scheduler.config.MaxConcurrentJobs != 50 {
		t.Errorf("Scheduler MaxConcurrentJobs = %d; want 50", scheduler.config.MaxConcurrentJobs)
	}
}

func TestNewWithInvalidConfig(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("NewWithConfig should panic with invalid config")
		}
	}()

	config := Config{
		Logger:            nil, // Invalid
		Timezone:          time.Local,
		MaxConcurrentJobs: 10,
	}

	NewWithConfig(config)
}

func TestSchedulerStartStop(t *testing.T) {
	scheduler := New()
	ctx := context.Background()

	// Test start
	err := scheduler.Start(ctx)
	if err != nil {
		t.Errorf("Start() returned error: %v", err)
	}

	if !scheduler.IsRunning() {
		t.Error("Scheduler should be running after Start()")
	}

	// Test start when already running
	err = scheduler.Start(ctx)
	if err != ErrSchedulerAlreadyStarted {
		t.Errorf("Start() when running should return ErrSchedulerAlreadyStarted; got %v", err)
	}

	// Test stop
	err = scheduler.Stop()
	if err != nil {
		t.Errorf("Stop() returned error: %v", err)
	}

	// Give some time for the goroutine to stop
	time.Sleep(time.Millisecond * 100)

	if scheduler.IsRunning() {
		t.Error("Scheduler should not be running after Stop()")
	}

	// Test stop when not running
	err = scheduler.Stop()
	if err != ErrSchedulerNotStarted {
		t.Errorf("Stop() when not running should return ErrSchedulerNotStarted; got %v", err)
	}
}

func TestAddJob(t *testing.T) {
	scheduler := New()
	executed := false

	job := func() {
		executed = true
	}

	err := scheduler.AddJob("* * * * *", job)
	if err != nil {
		t.Errorf("AddJob() returned error: %v", err)
	}

	jobs := scheduler.ListJobs()
	if len(jobs) != 1 {
		t.Errorf("Job count = %d; want 1", len(jobs))
	}

	stats := scheduler.GetStats()
	if stats.TotalJobs != 1 {
		t.Errorf("Stats.TotalJobs = %d; want 1", stats.TotalJobs)
	}

	// Use the executed variable to avoid "declared and not used" error
	_ = executed
}

func TestAddNamedJob(t *testing.T) {
	scheduler := New()
	executed := false

	job := func() {
		executed = true
	}

	err := scheduler.AddNamedJob("test-job", "* * * * *", job)
	if err != nil {
		t.Errorf("AddNamedJob() returned error: %v", err)
	}

	jobs := scheduler.ListJobs()
	if len(jobs) != 1 {
		t.Errorf("Job count = %d; want 1", len(jobs))
	}

	if jobs[0] != testJobNamePkg {
		t.Errorf("Job name = %q; want test-job", jobs[0])
	}

	// Use the executed variable to avoid "declared and not used" error
	_ = executed
}

func TestAddJobWithErrorHandler(t *testing.T) {
	scheduler := New()
	executed := false
	errorHandled := false

	job := func() error {
		executed = true
		return ErrJobTimeout
	}

	errorHandler := func(err error) {
		errorHandled = true
	}

	err := scheduler.AddJobWithErrorHandler("test-job", "* * * * *", job, errorHandler)
	if err != nil {
		t.Errorf("AddJobWithErrorHandler() returned error: %v", err)
	}

	jobs := scheduler.ListJobs()
	if len(jobs) != 1 {
		t.Errorf("Job count = %d; want 1", len(jobs))
	}

	// Use the variables to avoid "declared and not used" error
	_ = executed
	_ = errorHandled
}

func TestAddJobWithInvalidSchedule(t *testing.T) {
	scheduler := New()

	job := func() {
		// Test job
	}

	err := scheduler.AddJob("invalid schedule", job)
	if err == nil {
		t.Error("AddJob() with invalid schedule should return error")
	}
}

func TestAddJobWithConfig(t *testing.T) {
	scheduler := New()

	job := func() {
		// Test job
	}

	config := JobConfig{
		MaxRetries:    3,
		RetryInterval: time.Second * 5,
		Timeout:       time.Minute,
	}

	err := scheduler.AddJobWithConfig("test-job", "* * * * *", config, job)
	if err != nil {
		t.Errorf("AddJobWithConfig() returned error: %v", err)
	}

	jobs := scheduler.ListJobs()
	if len(jobs) != 1 {
		t.Errorf("Job count = %d; want 1", len(jobs))
	}
}

func TestAddDuplicateJob(t *testing.T) {
	scheduler := New()

	job := func() {
		// Test job
	}

	// Add first job
	err := scheduler.AddNamedJob("test-job", "* * * * *", job)
	if err != nil {
		t.Errorf("First AddNamedJob() returned error: %v", err)
	}

	// Try to add job with same name
	err = scheduler.AddNamedJob("test-job", "0 0 * * *", job)
	if err != ErrJobAlreadyExists {
		t.Errorf("Duplicate AddNamedJob() should return ErrJobAlreadyExists; got %v", err)
	}
}

func TestRemoveJob(t *testing.T) {
	scheduler := New()

	job := func() {
		// Test job
	}

	// Add job
	err := scheduler.AddNamedJob("test-job", "* * * * *", job)
	if err != nil {
		t.Errorf("AddNamedJob() returned error: %v", err)
	}

	// Remove job
	err = scheduler.RemoveJob("test-job")
	if err != nil {
		t.Errorf("RemoveJob() returned error: %v", err)
	}

	jobs := scheduler.ListJobs()
	if len(jobs) != 0 {
		t.Errorf("Job count after remove = %d; want 0", len(jobs))
	}

	// Try to remove non-existing job
	err = scheduler.RemoveJob("non-existing")
	if err != ErrJobNotFound {
		t.Errorf("RemoveJob() for non-existing job should return ErrJobNotFound; got %v", err)
	}
}

func TestGetJobStats(t *testing.T) {
	scheduler := New()

	job := func() {
		// Test job
	}

	err := scheduler.AddNamedJob("test-job", "* * * * *", job)
	if err != nil {
		t.Errorf("AddNamedJob() returned error: %v", err)
	}

	stats, err := scheduler.GetJobStats("test-job")
	if err != nil {
		t.Errorf("GetJobStats() returned error: %v", err)
	}

	if stats.Name != "test-job" {
		t.Errorf("JobStats.Name = %q; want test-job", stats.Name)
	}

	// Test non-existing job
	_, err = scheduler.GetJobStats("non-existing")
	if err != ErrJobNotFound {
		t.Errorf("GetJobStats() for non-existing job should return ErrJobNotFound; got %v", err)
	}
}

func TestGetStats(t *testing.T) {
	scheduler := New()

	// Initial stats
	stats := scheduler.GetStats()
	if stats.TotalJobs != 0 {
		t.Errorf("Initial TotalJobs = %d; want 0", stats.TotalJobs)
	}

	// Add a job
	job := func() {
		// Test job
	}

	err := scheduler.AddNamedJob("test-job", "* * * * *", job)
	if err != nil {
		t.Errorf("AddNamedJob() returned error: %v", err)
	}

	stats = scheduler.GetStats()
	if stats.TotalJobs != 1 {
		t.Errorf("TotalJobs after adding job = %d; want 1", stats.TotalJobs)
	}
}

func TestListJobs(t *testing.T) {
	scheduler := New()

	// Empty list
	jobs := scheduler.ListJobs()
	if len(jobs) != 0 {
		t.Errorf("Initial job count = %d; want 0", len(jobs))
	}

	// Add jobs
	job := func() {
		// Test job
	}

	scheduler.AddNamedJob("job1", "* * * * *", job)
	scheduler.AddNamedJob("job2", "0 0 * * *", job)

	jobs = scheduler.ListJobs()
	if len(jobs) != 2 {
		t.Errorf("Job count = %d; want 2", len(jobs))
	}

	// Check job names are present
	jobMap := make(map[string]bool)
	for _, name := range jobs {
		jobMap[name] = true
	}

	if !jobMap["job1"] {
		t.Error("job1 not found in job list")
	}

	if !jobMap["job2"] {
		t.Error("job2 not found in job list")
	}
}

func TestGetNextExecutions(t *testing.T) {
	scheduler := New()

	job := func() {
		// Test job
	}

	err := scheduler.AddNamedJob("daily-job", "0 9 * * *", job) // Daily at 9 AM
	if err != nil {
		t.Errorf("AddNamedJob() returned error: %v", err)
	}

	executions := scheduler.GetNextExecutions(3)
	if len(executions) != 1 {
		t.Errorf("GetNextExecutions returned %d jobs; want 1", len(executions))
	}

	jobExecutions := executions["daily-job"]
	if len(jobExecutions) != 3 {
		t.Errorf("GetNextExecutions returned %d executions for daily-job; want 3", len(jobExecutions))
	}

	// Check that executions are in chronological order
	for i := 1; i < len(jobExecutions); i++ {
		if !jobExecutions[i].After(jobExecutions[i-1]) {
			t.Errorf("Execution %d is not after execution %d", i, i-1)
		}
	}
}

func TestSchedulerConcurrency(t *testing.T) {
	scheduler := New()
	ctx := context.Background()

	var executed int64
	var mu sync.Mutex

	job := func() {
		mu.Lock()
		executed++
		mu.Unlock()
	}

	// Add multiple jobs
	for i := 0; i < 10; i++ {
		err := scheduler.AddJob("* * * * *", job)
		if err != nil {
			t.Errorf("AddJob() returned error: %v", err)
		}
	}

	// Start scheduler
	err := scheduler.Start(ctx)
	if err != nil {
		t.Errorf("Start() returned error: %v", err)
	}

	// Run concurrent operations
	var wg sync.WaitGroup

	// Concurrent job additions
	wg.Add(5)
	for i := 0; i < 5; i++ {
		go func(index int) {
			defer wg.Done()
			jobName := "concurrent-job-" + string(rune('a'+index))
			scheduler.AddNamedJob(jobName, "* * * * *", job)
		}(i)
	}

	// Concurrent stats queries
	wg.Add(5)
	for i := 0; i < 5; i++ {
		go func() {
			defer wg.Done()
			scheduler.GetStats()
		}()
	}

	wg.Wait()

	// Stop scheduler
	err = scheduler.Stop()
	if err != nil {
		t.Errorf("Stop() returned error: %v", err)
	}

	// Check final state
	jobs := scheduler.ListJobs()
	if len(jobs) != 15 { // 10 initial + 5 concurrent
		t.Errorf("Final job count = %d; want 15", len(jobs))
	}
}

func TestJobExecution(t *testing.T) {
	scheduler := New()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	executed := make(chan bool, 1)
	job := func() {
		select {
		case executed <- true:
		default:
		}
	}

	// Add a job that runs every minute
	err := scheduler.AddNamedJob("test-job", "* * * * *", job)
	if err != nil {
		t.Errorf("AddNamedJob() returned error: %v", err)
	}

	// Manually trigger job execution by manipulating the job's next execution time
	// This is a bit of a hack for testing, but it works
	scheduler.mu.Lock()
	if testJob, exists := scheduler.jobs["test-job"]; exists {
		testJob.NextExecution = time.Now().Add(-time.Minute) // Set to past time
	}
	scheduler.mu.Unlock()

	// Start scheduler
	err = scheduler.Start(ctx)
	if err != nil {
		t.Errorf("Start() returned error: %v", err)
	}

	// Wait for job execution or timeout
	select {
	case <-executed:
		// Job executed successfully
		t.Log("Job executed successfully")
	case <-time.After(time.Second * 3):
		t.Error("Job was not executed within timeout")
	}

	// Stop scheduler
	err = scheduler.Stop()
	if err != nil {
		t.Errorf("Stop() returned error: %v", err)
	}
}
