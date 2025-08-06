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
	fmt.Println("=== Go Cron - Basic Example ===")
	fmt.Println("This example demonstrates basic usage of the Go Cron library.")
	fmt.Println()

	// Create a new cron scheduler with default configuration
	scheduler := cron.New()

	// Example 1: Simple job that runs every minute
	fmt.Println("Adding job that runs every minute...")
	err := scheduler.AddJob("0 * * * * *", func() {
		fmt.Printf("[%s] ⏰ Every minute job executed\n", time.Now().Format("15:04:05"))
	})
	if err != nil {
		log.Fatal("Failed to add minute job:", err)
	}

	// Example 2: Job that runs every 30 seconds (6-field cron format)
	fmt.Println("Adding job that runs every 30 seconds...")
	err = scheduler.AddJob("*/30 * * * * *", func() {
		fmt.Printf("[%s] ⚡ Thirty-second job executed\n", time.Now().Format("15:04:05"))
	})
	if err != nil {
		log.Fatal("Failed to add thirty-second job:", err)
	}

	// Example 3: Job with specific time (hourly at minute 0)
	fmt.Println("Adding job that runs at the start of every hour...")
	err = scheduler.AddJob("0 0 * * * *", func() {
		fmt.Printf("[%s] 📊 Hourly report generated\n", time.Now().Format("15:04:05"))
		simulateWork("report generation", 1*time.Second)
	})
	if err != nil {
		log.Fatal("Failed to add hourly job:", err)
	}

	// Example 4: Daily job (5-field format)
	fmt.Println("Adding job that runs daily at 9:00 AM...")
	err = scheduler.AddJob("0 9 * * *", func() {
		fmt.Printf("[%s] 🔄 Daily backup started\n", time.Now().Format("15:04:05"))
		simulateWork("backup", 2*time.Second)
		fmt.Printf("[%s] ✅ Daily backup completed\n", time.Now().Format("15:04:05"))
	})
	if err != nil {
		log.Fatal("Failed to add daily job:", err)
	}

	// Example 5: Weekly job (Mondays at 9:00 AM)
	fmt.Println("Adding job that runs weekly on Mondays at 9:00 AM...")
	err = scheduler.AddJob("0 9 * * 1", func() {
		fmt.Printf("[%s] 🔧 Weekly maintenance started\n", time.Now().Format("15:04:05"))
		simulateWork("maintenance", 1500*time.Millisecond)
		fmt.Printf("[%s] ✅ Weekly maintenance completed\n", time.Now().Format("15:04:05"))
	})
	if err != nil {
		log.Fatal("Failed to add weekly job:", err)
	}

	// Show scheduler statistics before starting
	fmt.Println("\n=== Scheduler Information ===")
	printSchedulerStats(scheduler)

	// Start the scheduler
	fmt.Println("\n=== Starting Scheduler ===")
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		if err := scheduler.Start(ctx); err != nil {
			log.Printf("Scheduler error: %v", err)
		}
	}()

	fmt.Printf("✅ Scheduler started at %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Println("🚀 Running... Press Ctrl+C to stop")

	// Set up signal handling for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Wait for either signal or timeout
	select {
	case <-sigCh:
		fmt.Println("\n\n=== Received shutdown signal ===")
	case <-time.After(2 * time.Minute):
		fmt.Println("\n\n=== Demo completed (2 minutes) ===")
	}

	// Show final statistics
	fmt.Println("=== Final Statistics ===")
	printSchedulerStats(scheduler)

	// Stop the scheduler gracefully
	fmt.Println("\n=== Stopping Scheduler ===")
	cancel()
	scheduler.Stop()
	fmt.Println("✅ Scheduler stopped gracefully")
}

// simulateWork simulates some work being done
func simulateWork(workType string, duration time.Duration) {
	fmt.Printf("  → Performing %s work for %v...\n", workType, duration)
	time.Sleep(duration)
}

// printSchedulerStats displays current scheduler statistics
func printSchedulerStats(scheduler *cron.Scheduler) {
	stats := scheduler.GetStats()

	fmt.Printf("📈 Total Jobs: %d\n", stats.TotalJobs)
	fmt.Printf("🏃 Running Jobs: %d\n", stats.RunningJobs)
	fmt.Printf("✅ Successful Executions: %d\n", stats.SuccessfulExecutions)
	fmt.Printf("❌ Failed Executions: %d\n", stats.FailedExecutions)
	fmt.Printf("⏱️  Average Execution Time: %v\n", stats.AverageExecutionTime)
	fmt.Printf("⏰ Uptime: %v\n", stats.Uptime)

	if stats.SuccessfulExecutions+stats.FailedExecutions > 0 {
		successRate := float64(stats.SuccessfulExecutions) / float64(stats.SuccessfulExecutions+stats.FailedExecutions) * 100
		fmt.Printf("📊 Success Rate: %.1f%%\n", successRate)
	}
}
