package benchmark

import (
	"context"
	"testing"
	"time"

	"github.com/callmebg/cron/internal/parser"
	"github.com/callmebg/cron/pkg/cron"
)

func BenchmarkCronParsing(b *testing.B) {
	expressions := []string{
		"* * * * * *",
		"0 0 12 * * *",
		"*/5 0 * * * *",
		"0 30 9-17 * * 1-5",
		"0 0,12 1 */2 * *",
		"0 0 0 1 1 * 2020",
		"@yearly",
		"@monthly",
		"@weekly",
		"@daily",
		"@hourly",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		expr := expressions[i%len(expressions)]
		_, err := parser.Parse(expr)
		if err != nil {
			b.Fatalf("Failed to parse expression %s: %v", expr, err)
		}
	}
}

func BenchmarkCronParsingComplex(b *testing.B) {
	complexExpressions := []string{
		"0 0/5 14,18,3-39,52 * JAN,MAR,SEP MON-FRI 2002-2010",
		"0 23-7/2,8 * * * *",
		"0 11 4 * 5-6 *",
		"0 */5 * * * * *",
		"0 0 */2 1 * * *",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		expr := complexExpressions[i%len(complexExpressions)]
		_, err := parser.Parse(expr)
		if err != nil {
			b.Fatalf("Failed to parse complex expression %s: %v", expr, err)
		}
	}
}

func BenchmarkJobScheduling(b *testing.B) {
	scheduler := cron.New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := scheduler.AddJob("*/1 * * * *", func() {
			// Empty job
		})
		if err != nil {
			b.Fatalf("Failed to add job: %v", err)
		}
	}
}

func BenchmarkJobExecution(b *testing.B) {
	scheduler := cron.New()
	ctx := context.Background()

	var counter int
	err := scheduler.AddJob("*/1 * * * *", func() {
		counter++
	})
	if err != nil {
		b.Fatalf("Failed to add job: %v", err)
	}

	scheduler.Start(ctx)
	defer scheduler.Stop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		time.Sleep(10 * time.Microsecond)
	}
}

func BenchmarkConcurrentJobs(b *testing.B) {
	scheduler := cron.New()
	ctx := context.Background()
	scheduler.Start(ctx)
	defer scheduler.Stop()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			err := scheduler.AddJob("*/1 * * * *", func() {
				// Empty job
			})
			if err != nil {
				b.Fatalf("Failed to add job: %v", err)
			}
		}
	})
}

func BenchmarkMultipleJobsExecution(b *testing.B) {
	scheduler := cron.New()
	ctx := context.Background()

	// Add multiple jobs
	for i := 0; i < 100; i++ {
		err := scheduler.AddJob("*/1 * * * *", func() {
			// Empty job
		})
		if err != nil {
			b.Fatalf("Failed to add job %d: %v", i, err)
		}
	}

	scheduler.Start(ctx)
	defer scheduler.Stop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		time.Sleep(1 * time.Millisecond)
	}
}

func BenchmarkNextExecutionTime(b *testing.B) {
	schedule, err := parser.Parse("0 0 12 * * *")
	if err != nil {
		b.Fatalf("Failed to parse schedule: %v", err)
	}

	now := time.Now()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = schedule.Next(now)
	}
}

func BenchmarkMemoryAllocation(b *testing.B) {
	b.ReportAllocs()

	scheduler := cron.New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := scheduler.AddJob("*/1 * * * *", func() {
			// Memory allocation test
			data := make([]byte, 1024)
			_ = data
		})
		if err != nil {
			b.Fatalf("Failed to add job: %v", err)
		}
	}
}

func BenchmarkSchedulerStartStop(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scheduler := cron.New()
		ctx := context.Background()
		scheduler.Start(ctx)
		scheduler.Stop()
	}
}

// Comparative benchmarks
func BenchmarkCompareCronExpressions(b *testing.B) {
	expressions := map[string]string{
		"Simple":  "* * * * * *",
		"Hourly":  "0 0 * * * *",
		"Daily":   "0 0 12 * * *",
		"Weekly":  "0 0 12 * * 0",
		"Monthly": "0 0 12 1 * *",
		"Complex": "0 0/5 14,18 * JAN-MAR MON-FRI",
	}

	for name, expr := range expressions {
		b.Run(name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, err := parser.Parse(expr)
				if err != nil {
					b.Fatalf("Failed to parse %s expression: %v", name, err)
				}
			}
		})
	}
}
