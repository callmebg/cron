package main

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/callmebg/cron/pkg/cron"
)

// This example demonstrates testing patterns for cron jobs
func main() {
	fmt.Println("=== Cron Testing Patterns Example ===")
	fmt.Println("This example shows how to test cron jobs and schedulers")
	fmt.Println()

	// Run different testing scenarios
	fmt.Println("🧪 Running test scenarios...")

	testBasicJobExecution()
	testJobWithMockTime()
	testErrorHandling()
	testJobRetries()
	testSchedulerLifecycle()
	testConcurrentJobs()

	fmt.Println("\n✅ All test scenarios completed!")
	fmt.Println("💡 See the source code for testing patterns you can use in your own tests")
}

// testBasicJobExecution demonstrates testing basic job execution
func testBasicJobExecution() {
	fmt.Println("\n1. Testing Basic Job Execution")
	fmt.Println("   Creating scheduler and testing job execution...")

	scheduler := cron.New()

	// Use a channel to verify job execution
	executed := make(chan bool, 1)

	job := func() {
		fmt.Println("   ✓ Test job executed successfully")
		executed <- true
	}

	// Add job that runs every minute
	err := scheduler.AddNamedJob("test-job", "* * * * *", job)
	if err != nil {
		log.Printf("   ❌ Failed to add job: %v", err)
		return
	}

	// Start scheduler
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = scheduler.Start(ctx)
	if err != nil {
		log.Printf("   ❌ Failed to start scheduler: %v", err)
		return
	}

	// Wait for job execution or timeout
	select {
	case <-executed:
		fmt.Println("   ✅ Job execution test passed")
	case <-time.After(70 * time.Second): // Wait slightly more than a minute
		fmt.Println("   ❌ Job execution test failed - timeout")
	}

	scheduler.Stop()
}

// testJobWithMockTime demonstrates testing with controlled time
func testJobWithMockTime() {
	fmt.Println("\n2. Testing with Mock Time Patterns")
	fmt.Println("   Demonstrating how to test time-dependent behavior...")

	scheduler := cron.New()

	executionTimes := make([]time.Time, 0)
	var executionCount int

	job := func() {
		executionCount++
		executionTimes = append(executionTimes, time.Now())
		fmt.Printf("   ✓ Job execution #%d at %s\n", executionCount, time.Now().Format("15:04:05"))
	}

	// Add job
	err := scheduler.AddNamedJob("time-test-job", "* * * * *", job)
	if err != nil {
		log.Printf("   ❌ Failed to add job: %v", err)
		return
	}

	// Test scheduler stats
	stats := scheduler.GetStats()
	if stats.TotalJobs != 1 {
		fmt.Printf("   ❌ Expected 1 job, got %d\n", stats.TotalJobs)
		return
	}

	fmt.Println("   ✅ Mock time patterns test setup completed")
	fmt.Println("   💡 In real tests, you would use time mocking libraries")
}

// testErrorHandling demonstrates testing error scenarios
func testErrorHandling() {
	fmt.Println("\n3. Testing Error Handling")
	fmt.Println("   Testing job error scenarios...")

	scheduler := cron.New()

	var errorCount int
	expectedError := fmt.Errorf("test error")

	job := func() error {
		fmt.Println("   🎯 Job intentionally failing for test")
		return expectedError
	}

	errorHandler := func(err error) {
		errorCount++
		fmt.Printf("   ✓ Error handled: %v (count: %d)\n", err, errorCount)
	}

	err := scheduler.AddJobWithErrorHandler("error-test-job", "* * * * *", job, errorHandler)
	if err != nil {
		log.Printf("   ❌ Failed to add error job: %v", err)
		return
	}

	// Start scheduler briefly
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	scheduler.Start(ctx)
	time.Sleep(70 * time.Second) // Wait for potential execution
	scheduler.Stop()

	fmt.Println("   ✅ Error handling test completed")
	fmt.Printf("   📊 Errors handled: %d\n", errorCount)
}

// testJobRetries demonstrates testing retry behavior
func testJobRetries() {
	fmt.Println("\n4. Testing Job Retries")
	fmt.Println("   Testing retry mechanism...")

	scheduler := cron.New()

	var attemptCount int

	retryConfig := cron.JobConfig{
		MaxRetries:    3,
		RetryInterval: time.Second,
		Timeout:       5 * time.Second,
	}

	job := func() error {
		attemptCount++
		fmt.Printf("   🔄 Attempt #%d\n", attemptCount)

		if attemptCount < 3 {
			return fmt.Errorf("attempt %d failed", attemptCount)
		}

		fmt.Println("   ✅ Job succeeded on attempt 3")
		return nil
	}

	errorHandler := func(err error) {
		fmt.Printf("   ⚠️ Retry attempt failed: %v\n", err)
	}

	err := scheduler.AddJobWithErrorConfig("retry-test-job", "* * * * *", retryConfig, job, errorHandler)
	if err != nil {
		log.Printf("   ❌ Failed to add retry job: %v", err)
		return
	}

	fmt.Printf("   ✅ Retry test setup completed (max retries: %d)\n", retryConfig.MaxRetries)
	fmt.Println("   💡 In real tests, you would verify the exact retry behavior")
}

// testSchedulerLifecycle demonstrates testing scheduler start/stop
func testSchedulerLifecycle() {
	fmt.Println("\n5. Testing Scheduler Lifecycle")
	fmt.Println("   Testing scheduler start/stop behavior...")

	scheduler := cron.New()

	// Test initial state
	if scheduler.IsRunning() {
		fmt.Println("   ❌ Scheduler should not be running initially")
		return
	}
	fmt.Println("   ✓ Initial state: stopped")

	// Test start
	ctx := context.Background()
	err := scheduler.Start(ctx)
	if err != nil {
		log.Printf("   ❌ Failed to start scheduler: %v", err)
		return
	}

	if !scheduler.IsRunning() {
		fmt.Println("   ❌ Scheduler should be running after start")
		return
	}
	fmt.Println("   ✓ Successfully started")

	// Test double start (should fail)
	err = scheduler.Start(ctx)
	if err != cron.ErrSchedulerAlreadyStarted {
		fmt.Printf("   ❌ Expected ErrSchedulerAlreadyStarted, got: %v\n", err)
		return
	}
	fmt.Println("   ✓ Double start properly rejected")

	// Test stop
	err = scheduler.Stop()
	if err != nil {
		log.Printf("   ❌ Failed to stop scheduler: %v", err)
		return
	}

	// Give it a moment to stop
	time.Sleep(100 * time.Millisecond)

	if scheduler.IsRunning() {
		fmt.Println("   ❌ Scheduler should be stopped after stop")
		return
	}
	fmt.Println("   ✓ Successfully stopped")

	// Test double stop (should fail)
	err = scheduler.Stop()
	if err != cron.ErrSchedulerNotStarted {
		fmt.Printf("   ❌ Expected ErrSchedulerNotStarted, got: %v\n", err)
		return
	}
	fmt.Println("   ✓ Double stop properly rejected")

	fmt.Println("   ✅ Scheduler lifecycle test passed")
}

// testConcurrentJobs demonstrates testing concurrent job execution
func testConcurrentJobs() {
	fmt.Println("\n6. Testing Concurrent Jobs")
	fmt.Println("   Testing multiple jobs running concurrently...")

	config := cron.Config{
		Logger:            log.New(log.Writer(), "TEST: ", log.LstdFlags),
		Timezone:          time.Local,
		MaxConcurrentJobs: 3, // Limit to test concurrency control
		EnableMonitoring:  false,
		MonitoringPort:    8080,
	}

	scheduler := cron.NewWithConfig(config)

	var job1Done, job2Done, job3Done bool

	// Job 1: Quick job
	job1 := func() {
		fmt.Println("   🏃 Quick job running")
		time.Sleep(100 * time.Millisecond)
		job1Done = true
		fmt.Println("   ✓ Quick job completed")
	}

	// Job 2: Medium job
	job2 := func() {
		fmt.Println("   🚶 Medium job running")
		time.Sleep(500 * time.Millisecond)
		job2Done = true
		fmt.Println("   ✓ Medium job completed")
	}

	// Job 3: Slow job
	job3 := func() {
		fmt.Println("   🐌 Slow job running")
		time.Sleep(1 * time.Second)
		job3Done = true
		fmt.Println("   ✓ Slow job completed")
	}

	// Add all jobs
	scheduler.AddNamedJob("quick-job", "* * * * *", job1)
	scheduler.AddNamedJob("medium-job", "* * * * *", job2)
	scheduler.AddNamedJob("slow-job", "* * * * *", job3)

	// Check job count
	jobs := scheduler.ListJobs()
	if len(jobs) != 3 {
		fmt.Printf("   ❌ Expected 3 jobs, got %d\n", len(jobs))
		return
	}
	fmt.Printf("   ✓ Added %d concurrent jobs\n", len(jobs))

	// Test job stats
	for _, jobName := range jobs {
		stats, err := scheduler.GetJobStats(jobName)
		if err != nil {
			fmt.Printf("   ❌ Failed to get stats for %s: %v\n", jobName, err)
			continue
		}
		fmt.Printf("   📊 Job %s: status=%s, runs=%d\n", stats.Name, stats.Status, stats.TotalRuns)
	}

	fmt.Println("   ✅ Concurrent jobs test setup completed")
	fmt.Println("   💡 In real tests, you would verify concurrency limits and execution order")

	// Use the completion flags to avoid "declared and not used" error
	_ = job1Done
	_ = job2Done
	_ = job3Done
}

// TestExampleFunction shows how to write actual Go tests
func TestExampleFunction(t *testing.T) {
	// This function demonstrates the structure of actual Go tests
	// that you would write in *_test.go files

	scheduler := cron.New()

	executed := false
	job := func() {
		executed = true
	}

	err := scheduler.AddJob("* * * * *", job)
	if err != nil {
		t.Errorf("AddJob failed: %v", err)
	}

	if executed {
		t.Error("Job should not execute before scheduler starts")
	}

	jobs := scheduler.ListJobs()
	if len(jobs) != 1 {
		t.Errorf("Expected 1 job, got %d", len(jobs))
	}

	stats := scheduler.GetStats()
	if stats.TotalJobs != 1 {
		t.Errorf("Expected TotalJobs=1, got %d", stats.TotalJobs)
	}
}

// BenchmarkExampleFunction shows how to write benchmarks
func BenchmarkExampleFunction(b *testing.B) {
	scheduler := cron.New()

	job := func() {
		// Minimal job for benchmarking
	}

	for i := 0; i < b.N; i++ {
		jobName := fmt.Sprintf("bench-job-%d", i)
		scheduler.AddNamedJob(jobName, "* * * * *", job)
	}
}

// Helper functions for testing

// MockTime represents a controllable time for testing
type MockTime struct {
	current time.Time
}

func NewMockTime(start time.Time) *MockTime {
	return &MockTime{current: start}
}

func (m *MockTime) Now() time.Time {
	return m.current
}

func (m *MockTime) Advance(duration time.Duration) {
	m.current = m.current.Add(duration)
}

// TestHelper provides utilities for testing cron jobs
type TestHelper struct {
	scheduler *cron.Scheduler
	executed  map[string]int
}

func NewTestHelper() *TestHelper {
	return &TestHelper{
		scheduler: cron.New(),
		executed:  make(map[string]int),
	}
}

func (th *TestHelper) AddTestJob(name, schedule string) error {
	return th.scheduler.AddNamedJob(name, schedule, func() {
		th.executed[name]++
		fmt.Printf("Test job %s executed (count: %d)\n", name, th.executed[name])
	})
}

func (th *TestHelper) GetExecutionCount(jobName string) int {
	return th.executed[jobName]
}

func (th *TestHelper) WaitForExecution(jobName string, expectedCount int, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if th.executed[jobName] >= expectedCount {
			return true
		}
		time.Sleep(10 * time.Millisecond)
	}
	return false
}
