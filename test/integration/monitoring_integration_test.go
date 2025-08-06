package integration

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/callmebg/cron/pkg/cron"
)

func TestMonitoringIntegration(t *testing.T) {
	t.Run("MonitoringMetrics", func(t *testing.T) {
		config := cron.Config{
			Logger:            log.New(os.Stdout, "CRON: ", log.LstdFlags),
			Timezone:          time.Local,
			MaxConcurrentJobs: 100,
			EnableMonitoring:  true,
			MonitoringPort:    8080,
		}

		scheduler := cron.NewWithConfig(config)
		ctx := context.Background()

		var executed bool
		err := scheduler.AddJob("*/1 * * * * *", func() {
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
		err := scheduler.AddJob("*/1 * * * * *", func() {
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

		err := scheduler.AddJobWithErrorHandler("test-job", "*/1 * * * * *", func() error {
			return cron.ErrJobTimeout
		}, func(_ error) {
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
