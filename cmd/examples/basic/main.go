package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/callmebg/cron/pkg/cron"
)

func main() {
	// Create a new cron scheduler with default configuration
	c := cron.New()

	fmt.Println("=== Basic Cron Scheduler Example ===")
	fmt.Println("This example demonstrates basic job scheduling with the Go Cron library.")
	fmt.Println()

	// Example 1: Simple job that runs every minute
	fmt.Println("Adding job that runs every minute...")
	err := c.AddJob("* * * * *", func() {
		fmt.Printf("[%s] Every minute job executed\n", time.Now().Format("15:04:05"))
	})
	if err != nil {
		log.Fatal("Failed to add minute job:", err)
	}

	// Example 2: Named job for better tracking
	fmt.Println("Adding named job that runs every 30 seconds...")
	err = c.AddNamedJob("thirty-second-job", "*/30 * * * * *", func() {
		fmt.Printf("[%s] Thirty-second job executed\n", time.Now().Format("15:04:05"))
	})
	if err != nil {
		// Note: The parser might not support seconds, so this would fail with standard 5-field cron
		fmt.Printf("Note: %v (this is expected with 5-field cron syntax)\n", err)

		// Alternative: use a job that runs every 2 minutes instead
		err = c.AddNamedJob("two-minute-job", "*/2 * * * *", func() {
			fmt.Printf("[%s] Two-minute job executed\n", time.Now().Format("15:04:05"))
		})
		if err != nil {
			log.Fatal("Failed to add two-minute job:", err)
		}
	}

	// Example 3: Job with specific time
	fmt.Println("Adding job that runs at the top of every hour...")
	err = c.AddNamedJob("hourly-report", "0 * * * *", func() {
		fmt.Printf("[%s] Hourly report generated\n", time.Now().Format("15:04:05"))
	})
	if err != nil {
		log.Fatal("Failed to add hourly job:", err)
	}

	// Example 4: Daily job
	fmt.Println("Adding job that runs daily at 9:00 AM...")
	err = c.AddNamedJob("daily-backup", "0 9 * * *", func() {
		fmt.Printf("[%s] Daily backup started\n", time.Now().Format("15:04:05"))
		simulateWork("backup", 2*time.Second)
		fmt.Printf("[%s] Daily backup completed\n", time.Now().Format("15:04:05"))
	})
	if err != nil {
		log.Fatal("Failed to add daily job:", err)
	}

	// Example 5: Weekly job (Mondays at 9:00 AM)
	fmt.Println("Adding job that runs weekly on Mondays at 9:00 AM...")
	err = c.AddNamedJob("weekly-maintenance", "0 9 * * 1", func() {
		fmt.Printf("[%s] Weekly maintenance started\n", time.Now().Format("15:04:05"))
		simulateWork("maintenance", 3*time.Second)
		fmt.Printf("[%s] Weekly maintenance completed\n", time.Now().Format("15:04:05"))
	})
	if err != nil {
		log.Fatal("Failed to add weekly job:", err)
	}

	// Show scheduler statistics
	fmt.Println("\n=== Scheduler Information ===")
	printSchedulerStats(c)

	// Start the scheduler
	fmt.Println("\n=== Starting Scheduler ===")
	ctx := context.Background()
	err = c.Start(ctx)
	if err != nil {
		log.Fatal("Failed to start scheduler:", err)
	}

	fmt.Printf("Scheduler started at %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Println("Press Ctrl+C to stop, or wait 2 minutes for automatic shutdown...")

	// Run for 2 minutes to demonstrate job execution
	time.Sleep(2 * time.Minute)

	// Show final statistics
	fmt.Println("\n=== Final Statistics ===")
	printSchedulerStats(c)

	// Show job-specific statistics
	fmt.Println("\n=== Job Statistics ===")
	printJobStats(c)

	// Stop the scheduler gracefully
	fmt.Println("\n=== Stopping Scheduler ===")
	err = c.Stop()
	if err != nil {
		log.Printf("Error stopping scheduler: %v", err)
	} else {
		fmt.Println("Scheduler stopped gracefully")
	}
}

// simulateWork simulates some work being done
func simulateWork(workType string, duration time.Duration) {
	fmt.Printf("  → Performing %s work...\n", workType)
	time.Sleep(duration)
}

// printSchedulerStats displays current scheduler statistics
func printSchedulerStats(c *cron.Scheduler) {
	stats := c.GetStats()

	fmt.Printf("Total Jobs: %d\n", stats.TotalJobs)
	fmt.Printf("Running Jobs: %d\n", stats.RunningJobs)
	fmt.Printf("Successful Executions: %d\n", stats.SuccessfulExecutions)
	fmt.Printf("Failed Executions: %d\n", stats.FailedExecutions)
	fmt.Printf("Average Execution Time: %v\n", stats.AverageExecutionTime)
	fmt.Printf("Uptime: %v\n", stats.Uptime)
}

// printJobStats displays statistics for each job
func printJobStats(c *cron.Scheduler) {
	jobs := c.ListJobs()

	for _, jobName := range jobs {
		stats, err := c.GetJobStats(jobName)
		if err != nil {
			fmt.Printf("Error getting stats for job %s: %v\n", jobName, err)
			continue
		}

		fmt.Printf("\nJob: %s\n", stats.Name)
		fmt.Printf("  Schedule: %s\n", stats.Schedule)
		fmt.Printf("  Status: %s\n", stats.Status)
		fmt.Printf("  Total Runs: %d\n", stats.TotalRuns)
		fmt.Printf("  Success Rate: %.1f%%\n", stats.SuccessRate)
		fmt.Printf("  Last Execution: %s\n", formatTime(stats.LastExecution))
		fmt.Printf("  Next Execution: %s\n", formatTime(stats.NextExecution))
		fmt.Printf("  Execution Time: %v\n", stats.ExecutionTime)
	}
}

// formatTime formats a time for display
func formatTime(t time.Time) string {
	if t.IsZero() {
		return "Never"
	}
	return t.Format("2006-01-02 15:04:05")
}
