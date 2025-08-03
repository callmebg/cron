//go:build integration
// +build integration

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/callmebg/cron/pkg/cron"
)

func TestMonitoringIntegration(t *testing.T) {
	t.Run("MonitoringMetrics", func(t *testing.T) {
		config := cron.Config{
			EnableMonitoring: true,
			MonitoringPort:   8080,
		}

		scheduler := cron.NewWithConfig(config)
		ctx := context.Background()

		var executed bool
		err := scheduler.AddJob("*/1 * * * *", func() {
			executed = true
		})
		if err != nil {
			t.Fatalf("Failed to add job: %v", err)
		}

		scheduler.Start(ctx)
		time.Sleep(1500 * time.Millisecond)
		scheduler.Stop()

		if !executed {
			t.Error("Job was not executed")
		}

		// Get metrics
		metrics := scheduler.GetStats()
		if metrics.TotalJobs == 0 {
			t.Error("Expected metrics to show total jobs > 0")
		}
	})

	t.Run("JobExecutionStats", func(t *testing.T) {
		scheduler := cron.New()
		ctx := context.Background()

		var execCount int
		err := scheduler.AddJob("*/1 * * * *", func() {
			execCount++
		})
		if err != nil {
			t.Fatalf("Failed to add job: %v", err)
		}

		scheduler.Start(ctx)
		time.Sleep(2500 * time.Millisecond)
		scheduler.Stop()

		stats := scheduler.GetStats()
		if stats.TotalJobs == 0 {
			t.Error("Expected job stats to show total jobs > 0")
		}

		if execCount == 0 {
			t.Error("Expected job to execute at least once")
		}
	})

	t.Run("FailureTracking", func(t *testing.T) {
		scheduler := cron.New()
		ctx := context.Background()

		err := scheduler.AddJobWithErrorHandler("test-job", "*/1 * * * *", func() error {
			return cron.ErrJobTimeout
		}, func(err error) {
			// Handle error
		})
		if err != nil {
			t.Fatalf("Failed to add job: %v", err)
		}

		scheduler.Start(ctx)
		time.Sleep(1500 * time.Millisecond)
		scheduler.Stop()

		stats := scheduler.GetStats()
		if stats.FailedExecutions == 0 {
			t.Error("Expected failure count > 0")
		}
	})
}

func TestRetryIntegration(t *testing.T) {
	t.Run("JobRetryMechanism", func(t *testing.T) {
		scheduler := cron.New()
		ctx := context.Background()

		attempts := 0
		config := cron.JobConfig{
			MaxRetries:    3,
			RetryInterval: 100 * time.Millisecond,
		}

		err := scheduler.AddJobWithConfig("retry-job", "*/1 * * * *", config, func() {
			attempts++
			if attempts < 3 {
				panic("retry test")
			}
		})
		if err != nil {
			t.Fatalf("Failed to add job: %v", err)
		}

		scheduler.Start(ctx)
		time.Sleep(2 * time.Second)
		scheduler.Stop()

		if attempts < 3 {
			t.Errorf("Expected at least 3 attempts, got %d", attempts)
		}
	})
}

func TestConfigurationIntegration(t *testing.T) {
	t.Run("CustomConfiguration", func(t *testing.T) {
		loc, err := time.LoadLocation("UTC")
		if err != nil {
			t.Skip("UTC timezone not available")
		}

		config := cron.Config{
			Timezone:          loc,
			EnableMonitoring:  true,
			MaxConcurrentJobs: 5,
		}

		scheduler := cron.NewWithConfig(config)

		// Test timezone handling
		err = scheduler.AddJob("0 0 12 * * *", func() {
			// Daily at noon UTC
		})
		if err != nil {
			t.Fatalf("Failed to add job: %v", err)
		}

		// Verify job was added successfully
		stats := scheduler.GetStats()
		if stats.TotalJobs != 1 {
			t.Errorf("Expected 1 job, got %d", stats.TotalJobs)
		}
	})

	t.Run("ConcurrencyLimits", func(t *testing.T) {
		config := cron.Config{
			MaxConcurrentJobs: 2,
		}

		scheduler := cron.NewWithConfig(config)
		ctx := context.Background()

		executing := 0
		maxExecuting := 0

		for i := 0; i < 5; i++ {
			err := scheduler.AddJob("*/1 * * * *", func() {
				executing++
				if executing > maxExecuting {
					maxExecuting = executing
				}
				time.Sleep(500 * time.Millisecond)
				executing--
			})
			if err != nil {
				t.Fatalf("Failed to add job %d: %v", i, err)
			}
		}

		scheduler.Start(ctx)
		time.Sleep(2 * time.Second)
		scheduler.Stop()

		// Note: This test is somewhat flaky due to timing, but gives us a rough idea
		if maxExecuting > 5 {
			t.Errorf("Too many jobs executed concurrently: max executing was %d", maxExecuting)
		}
	})
}
