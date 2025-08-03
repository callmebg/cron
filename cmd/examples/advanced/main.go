package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/callmebg/cron/pkg/cron"
)

func main() {
	fmt.Println("=== Advanced Cron Scheduler Example ===")
	fmt.Println("This example demonstrates advanced features including:")
	fmt.Println("- Custom configuration")
	fmt.Println("- Error handling")
	fmt.Println("- Job timeouts and retries")
	fmt.Println("- Graceful shutdown")
	fmt.Println()

	// Create custom configuration
	config := cron.Config{
		Logger:            log.New(os.Stdout, "CRON: ", log.LstdFlags|log.Lshortfile),
		Timezone:          time.UTC,
		MaxConcurrentJobs: 5,     // Limit concurrent jobs
		EnableMonitoring:  false, // We'll enable this in the monitoring example
		MonitoringPort:    8080,
	}

	// Create scheduler with custom configuration
	scheduler := cron.NewWithConfig(config)

	// Example 1: Job with error handling
	fmt.Println("Adding job with error handling...")
	err := scheduler.AddJobWithErrorHandler(
		"error-prone-job",
		"*/30 * * * *", // Every 30 seconds for demonstration
		func() error {
			fmt.Printf("[%s] Executing error-prone job...\n", time.Now().Format("15:04:05"))

			// Simulate random failures
			if time.Now().Second()%3 == 0 {
				return fmt.Errorf("simulated error at %s", time.Now().Format("15:04:05"))
			}

			fmt.Printf("[%s] Error-prone job completed successfully\n", time.Now().Format("15:04:05"))
			return nil
		},
		func(err error) {
			fmt.Printf("[%s] 🚨 Job failed: %v\n", time.Now().Format("15:04:05"), err)
			// In a real application, you might:
			// - Send notifications
			// - Log to external systems
			// - Increment error counters
		},
	)
	if err != nil {
		log.Fatal("Failed to add error-prone job:", err)
	}

	// Example 2: Job with custom configuration (retries and timeout)
	fmt.Println("Adding job with retry configuration...")
	jobConfig := cron.JobConfig{
		MaxRetries:    3,                // Retry up to 3 times
		RetryInterval: 5 * time.Second,  // Wait 5 seconds between retries
		Timeout:       30 * time.Second, // Timeout after 30 seconds
	}

	err = scheduler.AddJobWithErrorConfig(
		"resilient-job",
		"*/45 * * * *", // Every 45 seconds
		jobConfig,
		func() error {
			fmt.Printf("[%s] Starting resilient job...\n", time.Now().Format("15:04:05"))

			// Simulate work that might fail
			workDuration := time.Duration(time.Now().Second()%10) * time.Second
			fmt.Printf("[%s] Working for %v...\n", time.Now().Format("15:04:05"), workDuration)

			time.Sleep(workDuration)

			// Simulate occasional failures
			if time.Now().Second()%7 == 0 {
				return fmt.Errorf("resilient job failed after %v", workDuration)
			}

			fmt.Printf("[%s] Resilient job completed successfully\n", time.Now().Format("15:04:05"))
			return nil
		},
		func(err error) {
			fmt.Printf("[%s] 🔄 Resilient job error (will retry): %v\n", time.Now().Format("15:04:05"), err)
		},
	)
	if err != nil {
		log.Fatal("Failed to add resilient job:", err)
	}

	// Example 3: Long-running job with timeout
	fmt.Println("Adding long-running job with timeout...")
	longJobConfig := cron.JobConfig{
		MaxRetries:    1,
		RetryInterval: 10 * time.Second,
		Timeout:       15 * time.Second, // Short timeout to demonstrate timeout handling
	}

	err = scheduler.AddJobWithErrorConfig(
		"long-running-job",
		"0 */2 * * *", // Every 2 minutes
		longJobConfig,
		func() error {
			fmt.Printf("[%s] Starting long-running job...\n", time.Now().Format("15:04:05"))

			// Simulate long-running work
			for i := 0; i < 20; i++ {
				time.Sleep(2 * time.Second)
				fmt.Printf("[%s] Long job progress: %d/20\n", time.Now().Format("15:04:05"), i+1)
			}

			fmt.Printf("[%s] Long-running job completed\n", time.Now().Format("15:04:05"))
			return nil
		},
		func(err error) {
			fmt.Printf("[%s] ⏰ Long job error: %v\n", time.Now().Format("15:04:05"), err)
		},
	)
	if err != nil {
		log.Fatal("Failed to add long-running job:", err)
	}

	// Example 4: Multiple jobs with different schedules
	fmt.Println("Adding multiple jobs with different schedules...")

	// High-frequency job
	err = scheduler.AddNamedJob("heartbeat", "* * * * *", func() {
		fmt.Printf("[%s] 💓 Heartbeat\n", time.Now().Format("15:04:05"))
	})
	if err != nil {
		log.Fatal("Failed to add heartbeat job:", err)
	}

	// Business hours job (9 AM to 5 PM, Monday to Friday)
	err = scheduler.AddNamedJob("business-hours-check", "0 9-17 * * 1-5", func() {
		fmt.Printf("[%s] 🏢 Business hours check\n", time.Now().Format("15:04:05"))
	})
	if err != nil {
		log.Fatal("Failed to add business hours job:", err)
	}

	// End of day job (5 PM, Monday to Friday)
	err = scheduler.AddNamedJob("end-of-day-report", "0 17 * * 1-5", func() {
		fmt.Printf("[%s] 📊 Generating end-of-day report\n", time.Now().Format("15:04:05"))
		time.Sleep(3 * time.Second) // Simulate report generation
		fmt.Printf("[%s] 📊 End-of-day report completed\n", time.Now().Format("15:04:05"))
	})
	if err != nil {
		log.Fatal("Failed to add end-of-day job:", err)
	}

	// Show initial status
	fmt.Println("\n=== Initial Configuration ===")
	printSchedulerInfo(scheduler)

	// Set up graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupt signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start the scheduler
	fmt.Println("\n=== Starting Advanced Scheduler ===")
	err = scheduler.Start(ctx)
	if err != nil {
		log.Fatal("Failed to start scheduler:", err)
	}

	fmt.Printf("Advanced scheduler started at %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Println("Jobs will run according to their schedules...")
	fmt.Println("Press Ctrl+C for graceful shutdown")

	// Create a ticker to show periodic status updates
	statusTicker := time.NewTicker(30 * time.Second)
	defer statusTicker.Stop()

	// Main loop
	for {
		select {
		case <-sigChan:
			fmt.Println("\n🛑 Shutdown signal received...")
			cancel() // Cancel context to stop scheduler loop
			goto shutdown

		case <-statusTicker.C:
			fmt.Println("\n=== Status Update ===")
			printSchedulerInfo(scheduler)

		case <-ctx.Done():
			goto shutdown
		}
	}

shutdown:
	fmt.Println("\n=== Graceful Shutdown ===")

	// Stop the scheduler
	err = scheduler.Stop()
	if err != nil {
		log.Printf("Error during shutdown: %v", err)
	} else {
		fmt.Println("✅ Scheduler stopped gracefully")
	}

	// Show final statistics
	fmt.Println("\n=== Final Report ===")
	printFinalReport(scheduler)
}

// printSchedulerInfo displays current scheduler information
func printSchedulerInfo(scheduler *cron.Scheduler) {
	stats := scheduler.GetStats()

	fmt.Printf("Status: %s\n", getStatusString(scheduler.IsRunning()))
	fmt.Printf("Total Jobs: %d\n", stats.TotalJobs)
	fmt.Printf("Running Jobs: %d\n", stats.RunningJobs)
	fmt.Printf("Successful Executions: %d\n", stats.SuccessfulExecutions)
	fmt.Printf("Failed Executions: %d\n", stats.FailedExecutions)

	if stats.SuccessfulExecutions+stats.FailedExecutions > 0 {
		successRate := float64(stats.SuccessfulExecutions) / float64(stats.SuccessfulExecutions+stats.FailedExecutions) * 100
		fmt.Printf("Success Rate: %.1f%%\n", successRate)
	}

	fmt.Printf("Average Execution Time: %v\n", stats.AverageExecutionTime)
	fmt.Printf("Uptime: %v\n", stats.Uptime)

	// Show next upcoming executions
	fmt.Println("\nNext 3 executions:")
	nextExecutions := scheduler.GetNextExecutions(3)
	for jobName, executions := range nextExecutions {
		if len(executions) > 0 {
			fmt.Printf("  %s: %s\n", jobName, executions[0].Format("15:04:05"))
		}
	}
}

// printFinalReport displays a comprehensive final report
func printFinalReport(scheduler *cron.Scheduler) {
	jobs := scheduler.ListJobs()

	fmt.Printf("Jobs managed: %d\n", len(jobs))

	for _, jobName := range jobs {
		stats, err := scheduler.GetJobStats(jobName)
		if err != nil {
			continue
		}

		fmt.Printf("\n📋 Job: %s\n", stats.Name)
		fmt.Printf("   Schedule: %s\n", stats.Schedule)
		fmt.Printf("   Total Runs: %d\n", stats.TotalRuns)
		if stats.TotalRuns > 0 {
			fmt.Printf("   Success Rate: %.1f%%\n", stats.SuccessRate)
			fmt.Printf("   Last Execution: %s\n", formatTime(stats.LastExecution))
			fmt.Printf("   Avg Duration: %v\n", stats.ExecutionTime)
		}
	}

	stats := scheduler.GetStats()
	fmt.Printf("\n📈 Overall Statistics:\n")
	fmt.Printf("   Total Executions: %d\n", stats.SuccessfulExecutions+stats.FailedExecutions)
	fmt.Printf("   Success Rate: %.1f%%\n", getOverallSuccessRate(stats))
	fmt.Printf("   Total Uptime: %v\n", stats.Uptime)
}

// Helper functions

func getStatusString(running bool) string {
	if running {
		return "🟢 Running"
	}
	return "🔴 Stopped"
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return "Never"
	}
	return t.Format("2006-01-02 15:04:05")
}

func getOverallSuccessRate(stats cron.Stats) float64 {
	total := stats.SuccessfulExecutions + stats.FailedExecutions
	if total == 0 {
		return 0
	}
	return float64(stats.SuccessfulExecutions) / float64(total) * 100
}
