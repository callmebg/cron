package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/callmebg/cron/internal/monitor"
	"github.com/callmebg/cron/pkg/cron"
)

func main() {
	fmt.Println("=== Cron Scheduler with Monitoring Example ===")
	fmt.Println("This example demonstrates:")
	fmt.Println("- HTTP monitoring endpoints")
	fmt.Println("- Real-time metrics collection")
	fmt.Println("- Performance tracking")
	fmt.Println("- Web dashboard access")
	fmt.Println()

	// Create configuration with monitoring enabled
	config := cron.Config{
		Logger:            log.New(os.Stdout, "CRON: ", log.LstdFlags),
		Timezone:          time.Local,
		MaxConcurrentJobs: 10,
		EnableMonitoring:  true,
		MonitoringPort:    8080,
	}

	// Create scheduler with monitoring
	scheduler := cron.NewWithConfig(config)

	// Create metrics collector
	metrics := monitor.NewMetrics()

	// Create HTTP monitor (this would be integrated into the scheduler in a real implementation)
	httpMonitor := monitor.NewHTTPMonitor(metrics, scheduler, config.MonitoringPort)

	// Add various jobs to generate interesting metrics
	setupJobs(scheduler, metrics)

	// Start HTTP monitoring server
	fmt.Printf("Starting HTTP monitoring server on port %d...\n", config.MonitoringPort)
	err := httpMonitor.Start()
	if err != nil {
		log.Fatal("Failed to start HTTP monitor:", err)
	}

	fmt.Printf("📊 Monitoring dashboard available at: http://localhost:%d\n", config.MonitoringPort)
	fmt.Println("📊 API endpoints:")
	fmt.Printf("   http://localhost:%d/health - Health check\n", config.MonitoringPort)
	fmt.Printf("   http://localhost:%d/metrics - Complete metrics\n", config.MonitoringPort)
	fmt.Printf("   http://localhost:%d/jobs - All jobs information\n", config.MonitoringPort)
	fmt.Printf("   http://localhost:%d/schedule - Upcoming executions\n", config.MonitoringPort)
	fmt.Println()

	// Set up graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start the scheduler
	fmt.Println("🚀 Starting scheduler...")
	err = scheduler.Start(ctx)
	if err != nil {
		log.Fatal("Failed to start scheduler:", err)
	}

	fmt.Printf("✅ Scheduler started at %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Println("📈 Collecting metrics and serving HTTP endpoints...")
	fmt.Println("🌐 Visit the monitoring URLs above to see real-time data")
	fmt.Println("⏹️  Press Ctrl+C to stop")

	// Periodic status updates
	statusTicker := time.NewTicker(15 * time.Second)
	defer statusTicker.Stop()

	metricsUpdateTicker := time.NewTicker(5 * time.Second)
	defer metricsUpdateTicker.Stop()

	// Main monitoring loop
	for {
		select {
		case <-sigChan:
			fmt.Println("\n🛑 Shutdown signal received...")
			goto shutdown

		case <-statusTicker.C:
			printMonitoringStatus(scheduler, metrics)

		case <-metricsUpdateTicker.C:
			// In a real implementation, this would be handled automatically by the scheduler
			// Here we're simulating the metrics updates
			updateMetrics(scheduler, metrics)

		case <-ctx.Done():
			goto shutdown
		}
	}

shutdown:
	fmt.Println("\n=== Graceful Shutdown ===")

	// Stop HTTP monitor
	err = httpMonitor.Stop()
	if err != nil {
		log.Printf("Error stopping HTTP monitor: %v", err)
	} else {
		fmt.Println("✅ HTTP monitor stopped")
	}

	// Stop scheduler
	err = scheduler.Stop()
	if err != nil {
		log.Printf("Error stopping scheduler: %v", err)
	} else {
		fmt.Println("✅ Scheduler stopped")
	}

	// Show final metrics
	fmt.Println("\n=== Final Metrics Report ===")
	printFinalMetrics(metrics)
}

// setupJobs adds various jobs to demonstrate monitoring capabilities
func setupJobs(scheduler *cron.Scheduler, metrics *monitor.Metrics) {
	fmt.Println("Setting up demonstration jobs...")

	// Fast job for frequent metrics
	err := scheduler.AddNamedJob("heartbeat", "* * * * *", func() {
		start := time.Now()
		fmt.Printf("[%s] 💓 Heartbeat\n", time.Now().Format("15:04:05"))

		// Simulate quick work
		time.Sleep(time.Millisecond * 100)

		duration := time.Since(start)
		metrics.RecordJobStart()
		metrics.RecordJobCompletion(duration)
	})
	if err != nil {
		log.Fatal("Failed to add heartbeat job:", err)
	}

	// Variable duration job
	err = scheduler.AddNamedJob("variable-work", "*/2 * * * *", func() {
		start := time.Now()
		workTime := time.Duration(1+time.Now().Second()%5) * time.Second

		fmt.Printf("[%s] 🔧 Variable work starting (will take %v)\n", time.Now().Format("15:04:05"), workTime)

		time.Sleep(workTime)

		duration := time.Since(start)
		fmt.Printf("[%s] 🔧 Variable work completed in %v\n", time.Now().Format("15:04:05"), duration)

		metrics.RecordJobStart()
		metrics.RecordJobCompletion(duration)
	})
	if err != nil {
		log.Fatal("Failed to add variable work job:", err)
	}

	// Occasionally failing job
	err = scheduler.AddJobWithErrorHandler(
		"flaky-job",
		"*/3 * * * *",
		func() error {
			start := time.Now()
			fmt.Printf("[%s] 🎲 Flaky job executing...\n", time.Now().Format("15:04:05"))

			// Simulate work
			time.Sleep(time.Millisecond * 500)

			duration := time.Since(start)
			metrics.RecordJobStart()

			// 30% chance of failure
			if time.Now().Second()%10 < 3 {
				metrics.RecordJobFailure(duration)
				return fmt.Errorf("random failure occurred")
			}

			metrics.RecordJobCompletion(duration)
			fmt.Printf("[%s] 🎲 Flaky job completed successfully\n", time.Now().Format("15:04:05"))
			return nil
		},
		func(err error) {
			fmt.Printf("[%s] 🚨 Flaky job failed: %v\n", time.Now().Format("15:04:05"), err)
		},
	)
	if err != nil {
		log.Fatal("Failed to add flaky job:", err)
	}

	// Heavy computation job
	err = scheduler.AddNamedJob("heavy-computation", "*/5 * * * *", func() {
		start := time.Now()
		fmt.Printf("[%s] 🏋️ Heavy computation starting...\n", time.Now().Format("15:04:05"))

		// Simulate CPU-intensive work
		count := 0
		for i := 0; i < 1000000; i++ {
			count += i % 100
		}

		duration := time.Since(start)
		fmt.Printf("[%s] 🏋️ Heavy computation completed (result: %d) in %v\n",
			time.Now().Format("15:04:05"), count, duration)

		metrics.RecordJobStart()
		metrics.RecordJobCompletion(duration)
	})
	if err != nil {
		log.Fatal("Failed to add heavy computation job:", err)
	}

	// Data processing job
	err = scheduler.AddNamedJob("data-processor", "*/4 * * * *", func() {
		start := time.Now()
		fmt.Printf("[%s] 📊 Data processing started\n", time.Now().Format("15:04:05"))

		// Simulate data processing stages
		stages := []string{"collecting", "validating", "transforming", "storing"}
		for _, stage := range stages {
			fmt.Printf("[%s] 📊 %s data...\n", time.Now().Format("15:04:05"), stage)
			time.Sleep(time.Millisecond * 300)
		}

		duration := time.Since(start)
		fmt.Printf("[%s] 📊 Data processing completed in %v\n", time.Now().Format("15:04:05"), duration)

		metrics.RecordJobStart()
		metrics.RecordJobCompletion(duration)
	})
	if err != nil {
		log.Fatal("Failed to add data processor job:", err)
	}

	fmt.Printf("✅ Added %d demonstration jobs\n", len(scheduler.ListJobs()))
}

// printMonitoringStatus displays current monitoring information
func printMonitoringStatus(scheduler *cron.Scheduler, metrics *monitor.Metrics) {
	fmt.Println("\n=== 📊 Monitoring Status ===")

	// Scheduler stats
	stats := scheduler.GetStats()
	fmt.Printf("🔧 Scheduler Status: %s\n", getStatusEmoji(scheduler.IsRunning()))
	fmt.Printf("📋 Active Jobs: %d\n", stats.TotalJobs)
	fmt.Printf("🏃 Running Jobs: %d\n", stats.RunningJobs)

	// Metrics snapshot
	snapshot := metrics.GetSnapshot()
	fmt.Printf("✅ Successful: %d\n", snapshot.CompletedJobs)
	fmt.Printf("❌ Failed: %d\n", snapshot.FailedJobs)
	fmt.Printf("📈 Jobs/min: %.1f\n", snapshot.JobsPerMinute)
	fmt.Printf("📊 Error Rate: %.1f%%\n", snapshot.ErrorRate)
	fmt.Printf("⏱️  Avg Duration: %v\n", snapshot.AverageExecutionTime)
	fmt.Printf("⏰ Uptime: %v\n", snapshot.Uptime)

	if !snapshot.LastJobExecution.IsZero() {
		timeSinceLastJob := time.Since(snapshot.LastJobExecution)
		fmt.Printf("🕐 Last Job: %v ago\n", timeSinceLastJob.Truncate(time.Second))
	}
}

// updateMetrics simulates metrics updates (in real implementation this would be automatic)
func updateMetrics(scheduler *cron.Scheduler, metrics *monitor.Metrics) {
	// This is a simplified update - in the real implementation,
	// the scheduler would automatically update metrics as jobs run
}

// printFinalMetrics displays comprehensive final metrics
func printFinalMetrics(metrics *monitor.Metrics) {
	snapshot := metrics.GetSnapshot()

	fmt.Println("📈 Final Performance Metrics:")
	fmt.Printf("   Total Jobs Executed: %d\n", snapshot.TotalJobs)
	fmt.Printf("   Successful: %d (%.1f%%)\n", snapshot.CompletedJobs,
		float64(snapshot.CompletedJobs)/float64(snapshot.TotalJobs)*100)
	fmt.Printf("   Failed: %d (%.1f%%)\n", snapshot.FailedJobs,
		float64(snapshot.FailedJobs)/float64(snapshot.TotalJobs)*100)

	fmt.Println("\n⏱️  Timing Metrics:")
	fmt.Printf("   Average Execution: %v\n", snapshot.AverageExecutionTime)
	fmt.Printf("   Fastest Execution: %v\n", snapshot.MinExecutionTime)
	fmt.Printf("   Slowest Execution: %v\n", snapshot.MaxExecutionTime)

	fmt.Println("\n📊 Throughput Metrics:")
	fmt.Printf("   Jobs per Minute: %.2f\n", snapshot.JobsPerMinute)
	fmt.Printf("   Total Uptime: %v\n", snapshot.Uptime)

	if snapshot.TotalJobs > 0 {
		fmt.Printf("   Overall Success Rate: %.1f%%\n",
			float64(snapshot.CompletedJobs)/float64(snapshot.TotalJobs)*100)
	}
}

// Helper functions

func getStatusEmoji(running bool) string {
	if running {
		return "🟢 Running"
	}
	return "🔴 Stopped"
}

// demonstrateAPIEndpoints shows how to programmatically access the monitoring API
func demonstrateAPIEndpoints(port int) {
	fmt.Println("\n=== 🌐 API Endpoint Demonstration ===")

	baseURL := fmt.Sprintf("http://localhost:%d", port)

	// Example API calls (in a real scenario, you'd parse the JSON responses)
	endpoints := []string{
		"/health",
		"/metrics",
		"/jobs",
		"/schedule?count=5",
	}

	for _, endpoint := range endpoints {
		url := baseURL + endpoint
		fmt.Printf("📡 GET %s\n", url)

		resp, err := http.Get(url)
		if err != nil {
			fmt.Printf("   ❌ Error: %v\n", err)
			continue
		}

		fmt.Printf("   ✅ Status: %s\n", resp.Status)
		resp.Body.Close()
	}

	fmt.Println("💡 Tip: Open these URLs in your browser for formatted JSON")
}
