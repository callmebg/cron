package integration

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/callmebg/cron/pkg/cron"
)

func TestCronIntegration(t *testing.T) {
	// Test basic scheduling
	t.Run("BasicScheduling", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		var executed sync.Map
		scheduler := cron.New()

		err := scheduler.AddJob("*/2 * * * * *", func() {
			executed.Store("basic", time.Now())
		})
		if err != nil {
			t.Fatalf("Failed to add job: %v", err)
		}

		scheduler.Start(ctx)
		defer scheduler.Stop()

		// Wait for at least one execution
		time.Sleep(3 * time.Second)

		if _, ok := executed.Load("basic"); !ok {
			t.Error("Basic job was not executed")
		}
	})

	// Test multiple jobs
	t.Run("MultipleJobs", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		var executed sync.Map
		scheduler := cron.New()

		err := scheduler.AddJob("*/1 * * * * *", func() {
			executed.Store("job1", time.Now())
		})
		if err != nil {
			t.Fatalf("Failed to add job1: %v", err)
		}

		err = scheduler.AddJob("*/2 * * * * *", func() {
			executed.Store("job2", time.Now())
		})
		if err != nil {
			t.Fatalf("Failed to add job2: %v", err)
		}

		scheduler.Start(ctx)
		defer scheduler.Stop()

		time.Sleep(3 * time.Second)

		if _, ok := executed.Load("job1"); !ok {
			t.Error("Job1 was not executed")
		}
		if _, ok := executed.Load("job2"); !ok {
			t.Error("Job2 was not executed")
		}
	})

	// Test job removal during execution
	t.Run("JobRemovalDuringExecution", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		var executed sync.Map
		scheduler := cron.New()

		err := scheduler.AddJob("*/1 * * * * *", func() {
			executed.Store("removable", time.Now())
		})
		if err != nil {
			t.Fatalf("Failed to add job: %v", err)
		}

		scheduler.Start(ctx)
		time.Sleep(1500 * time.Millisecond)
		scheduler.Stop()

		// Clear the executed map
		executed.Delete("removable")

		// Wait and ensure no more executions after stop
		time.Sleep(2 * time.Second)

		if _, ok := executed.Load("removable"); ok {
			t.Error("Job continued executing after scheduler stop")
		}
	})

	// Test scheduler start/stop
	t.Run("StartStopScheduler", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		var executed sync.Map
		scheduler := cron.New()

		err := scheduler.AddJob("*/1 * * * * *", func() {
			executed.Store("startstop", time.Now())
		})
		if err != nil {
			t.Fatalf("Failed to add job: %v", err)
		}

		// Start and immediately stop
		scheduler.Start(ctx)
		scheduler.Stop()

		// Clear any potential executions
		executed.Delete("startstop")

		// Wait and ensure no executions
		time.Sleep(2 * time.Second)

		if _, ok := executed.Load("startstop"); ok {
			t.Error("Job executed after scheduler was stopped")
		}

		// Restart and verify it works
		scheduler.Start(ctx)
		time.Sleep(1500 * time.Millisecond)
		scheduler.Stop()

		if _, ok := executed.Load("startstop"); !ok {
			t.Error("Job did not execute after scheduler restart")
		}
	})
}

func TestCronWithContext(t *testing.T) {
	t.Run("ContextCancellation", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		scheduler := cron.New()

		var executed int32
		err := scheduler.AddJob("*/1 * * * * *", func() {
			atomic.StoreInt32(&executed, 1)
		})
		if err != nil {
			t.Fatalf("Failed to add job: %v", err)
		}

		scheduler.Start(ctx)

		// Wait for context to cancel
		<-ctx.Done()
		scheduler.Stop()

		if atomic.LoadInt32(&executed) == 0 {
			t.Error("Job should have executed at least once before context cancellation")
		}
	})
}

func TestCronErrorHandling(t *testing.T) {
	t.Run("InvalidCronExpression", func(t *testing.T) {
		scheduler := cron.New()

		err := scheduler.AddJob("invalid cron", func() {})
		if err == nil {
			t.Error("Expected error for invalid cron expression")
		}
	})

	t.Run("PanicRecovery", func(t *testing.T) {
		scheduler := cron.New()
		ctx := context.Background()

		var recovered bool
		err := scheduler.AddJobWithErrorHandler("panic-job", "*/1 * * * * *", func() error {
			panic("test panic")
		}, func(err error) {
			recovered = true
		})
		if err != nil {
			t.Fatalf("Failed to add job: %v", err)
		}

		scheduler.Start(ctx)
		time.Sleep(1500 * time.Millisecond)
		scheduler.Stop()

		if !recovered {
			t.Error("Panic was not recovered")
		}
	})
}

func TestCronConcurrency(t *testing.T) {
	t.Run("ConcurrentJobExecution", func(t *testing.T) {
		scheduler := cron.New()
		ctx := context.Background()

		var mu sync.Mutex
		executions := make(map[string]int)

		// Add multiple jobs that could run concurrently
		for i := 0; i < 5; i++ {
			jobName := string(rune('A' + i))
			err := scheduler.AddJob("*/1 * * * * *", func() {
				mu.Lock()
				executions[jobName]++
				mu.Unlock()
			})
			if err != nil {
				t.Fatalf("Failed to add job %s: %v", jobName, err)
			}
		}

		scheduler.Start(ctx)
		time.Sleep(2500 * time.Millisecond)
		scheduler.Stop()

		mu.Lock()
		totalExecutions := 0
		for _, count := range executions {
			totalExecutions += count
		}
		mu.Unlock()

		if totalExecutions == 0 {
			t.Error("No jobs were executed")
		}

		t.Logf("Total executions: %d", totalExecutions)
	})
}

func TestCronLongRunningJobs(t *testing.T) {
	t.Run("LongRunningJob", func(t *testing.T) {
		scheduler := cron.New()
		ctx := context.Background()

		var startTime, endTime time.Time
		var jobCompleted bool
		var mu sync.Mutex

		err := scheduler.AddJob("*/2 * * * * *", func() {
			mu.Lock()
			startTime = time.Now()
			time.Sleep(1 * time.Second) // Simulate long-running job
			endTime = time.Now()
			jobCompleted = true
			mu.Unlock()
		})
		if err != nil {
			t.Fatalf("Failed to add job: %v", err)
		}

		scheduler.Start(ctx)
		time.Sleep(4 * time.Second)
		scheduler.Stop()

		mu.Lock()
		completed := jobCompleted
		start := startTime
		end := endTime
		mu.Unlock()

		if !completed {
			t.Error("Long-running job was not completed")
		}

		duration := end.Sub(start)
		if duration < 1*time.Second {
			t.Errorf("Job duration %v is less than expected 1 second", duration)
		}
	})
}
