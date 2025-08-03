package benchmark

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/callmebg/cron/pkg/cron"
)

func BenchmarkMonitoringOverhead(b *testing.B) {
	b.Run("WithMonitoring", func(b *testing.B) {
		config := cron.Config{
			EnableMonitoring: true,
		}
		scheduler := cron.NewWithConfig(config)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			err := scheduler.AddJob("*/1 * * * *", func() {})
			if err != nil {
				b.Fatalf("Failed to add job: %v", err)
			}
		}
	})

	b.Run("WithoutMonitoring", func(b *testing.B) {
		scheduler := cron.New()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			err := scheduler.AddJob("*/1 * * * *", func() {})
			if err != nil {
				b.Fatalf("Failed to add job: %v", err)
			}
		}
	})
}

func BenchmarkConcurrentJobAccess(b *testing.B) {
	scheduler := cron.New()
	ctx := context.Background()
	scheduler.Start(ctx)
	defer scheduler.Stop()

	// Pre-populate with some jobs
	for i := 0; i < 50; i++ {
		err := scheduler.AddJob("*/1 * * * *", func() {})
		if err != nil {
			b.Fatalf("Failed to add initial job: %v", err)
		}
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// Mix of operations
			switch b.N % 2 {
			case 0:
				// Add job
				_ = scheduler.AddJob("*/1 * * * *", func() {})
			case 1:
				// Get stats
				_ = scheduler.GetStats()
			}
		}
	})
}

func BenchmarkJobQueuePerformance(b *testing.B) {
	scheduler := cron.New()
	ctx := context.Background()

	b.ResetTimer()
	scheduler.Start(ctx)

	for i := 0; i < b.N; i++ {
		err := scheduler.AddJob("*/1 * * * *", func() {
			// Simulate varying job execution times
			time.Sleep(time.Duration(i%10) * time.Microsecond)
		})
		if err != nil {
			b.Fatalf("Failed to add job: %v", err)
		}
	}

	// Let jobs run briefly
	time.Sleep(100 * time.Millisecond)

	scheduler.Stop()
}

func BenchmarkRetryMechanism(b *testing.B) {
	scheduler := cron.New()
	ctx := context.Background()
	scheduler.Start(ctx)
	defer scheduler.Stop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		attempts := 0
		config := cron.JobConfig{
			MaxRetries:    3,
			RetryInterval: 1 * time.Millisecond,
		}

		err := scheduler.AddJobWithConfig("test", "*/1 * * * *", config, func() {
			attempts++
			if attempts < 2 {
				panic("retry test")
			}
		})
		if err != nil {
			b.Fatalf("Failed to add job: %v", err)
		}

		time.Sleep(5 * time.Millisecond)
	}
}

func BenchmarkTimeZoneHandling(b *testing.B) {
	locations := []string{
		"UTC",
		"America/New_York",
		"Europe/London",
		"Asia/Tokyo",
		"Australia/Sydney",
	}

	for _, locName := range locations {
		b.Run(locName, func(b *testing.B) {
			loc, err := time.LoadLocation(locName)
			if err != nil {
				b.Skip("Timezone not available:", locName)
			}

			config := cron.Config{
				Timezone: loc,
			}
			scheduler := cron.NewWithConfig(config)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				err := scheduler.AddJob("0 0 12 * * *", func() {})
				if err != nil {
					b.Fatalf("Failed to add job: %v", err)
				}
			}
		})
	}
}

func BenchmarkErrorHandling(b *testing.B) {
	scheduler := cron.New()
	ctx := context.Background()
	scheduler.Start(ctx)
	defer scheduler.Stop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := scheduler.AddJobWithErrorHandler("test", "*/1 * * * *", func() error {
			return cron.ErrJobTimeout
		}, func(err error) {
			// Handle error
		})
		if err != nil {
			b.Fatalf("Failed to add job: %v", err)
		}

		time.Sleep(2 * time.Millisecond)
	}
}

func BenchmarkContextCancellation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)

		scheduler := cron.New()
		scheduler.Start(ctx)

		err := scheduler.AddJob("*/1 * * * *", func() {
			select {
			case <-ctx.Done():
				return
			default:
				// Job logic
			}
		})
		if err != nil {
			b.Fatalf("Failed to add job: %v", err)
		}

		<-ctx.Done()
		scheduler.Stop()
		cancel()
	}
}

func BenchmarkHighFrequencyJobs(b *testing.B) {
	scheduler := cron.New()
	ctx := context.Background()
	scheduler.Start(ctx)
	defer scheduler.Stop()

	var counter int64
	var mu sync.Mutex

	err := scheduler.AddJob("* * * * * *", func() {
		mu.Lock()
		counter++
		mu.Unlock()
	})
	if err != nil {
		b.Fatalf("Failed to add job: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		time.Sleep(1 * time.Microsecond)
	}
}
