# Go Cron - Native Cron Job Scheduler

[![Version](https://img.shields.io/badge/version-v0.2.0--beta-blue.svg)](https://github.com/callmebg/cron)
[![Go Version](https://img.shields.io/badge/go-%3E%3D1.19-blue.svg)](https://golang.org/)
[![Test Coverage](https://img.shields.io/badge/coverage-75.4%25-green.svg)](#test-coverage)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

A lightweight, efficient cron job scheduler for Go applications built exclusively with the Go standard library. This library provides robust task scheduling with comprehensive monitoring capabilities while maintaining zero external dependencies.

English | [中文](README_zh.md)

## ✨ Features

- **Pure Go Standard Library**: No external dependencies, leveraging Go's native capabilities
- **Standard Cron Syntax**: Full support for 5-field and 6-field cron expressions
- **Concurrent Execution**: Efficient task execution using goroutines and worker pools
- **Built-in Monitoring**: Comprehensive metrics, execution tracking, and HTTP endpoints
- **Graceful Shutdown**: Clean termination with context-based cancellation
- **Thread-Safe**: Safe for concurrent use across multiple goroutines with robust synchronization
- **Flexible Configuration**: Configurable job timeouts, retries, and error handling
- **Memory Efficient**: Minimal memory footprint with priority queue-based scheduling
- **Production Ready**: Comprehensive test coverage (75.4%) with integration and benchmark tests

## 🚀 Quick Start

### Installation

```bash
go get github.com/callmebg/cron
```

### Basic Usage

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/callmebg/cron/pkg/cron"
)

func main() {
    // Create a new scheduler
    scheduler := cron.New()

    // Add a job that runs every minute
    err := scheduler.AddJob("0 * * * * *", func() {
        fmt.Println("Job executed at:", time.Now().Format("2006-01-02 15:04:05"))
    })
    if err != nil {
        log.Fatal("Failed to add job:", err)
    }

    // Start the scheduler
    ctx := context.Background()
    scheduler.Start(ctx)

    // Run for 5 minutes
    time.Sleep(5 * time.Minute)

    // Gracefully stop the scheduler
    scheduler.Stop()
    fmt.Println("Scheduler stopped")
}
```

### Advanced Usage with Configuration

```go
package main

import (
    "context"
    "log"
    "time"

    "github.com/callmebg/cron/pkg/cron"
)

func main() {
    // Create custom configuration
    config := cron.DefaultConfig()
    config.MaxConcurrentJobs = 50
    config.EnableMonitoring = true
    config.MonitoringPort = 8080
    
    // Create scheduler with custom config
    scheduler := cron.NewWithConfig(config)

    // Add job with error handling and retry configuration
    jobConfig := cron.JobConfig{
        MaxRetries:    3,
        RetryInterval: 30 * time.Second,
        Timeout:       5 * time.Minute,
    }

    err := scheduler.AddJobWithConfig("backup-job", "0 0 2 * * *", jobConfig, func() {
        // Simulate backup operation
        log.Println("Running backup job...")
        time.Sleep(2 * time.Second)
        log.Println("Backup completed successfully")
    })
    if err != nil {
        log.Fatal("Failed to add backup job:", err)
    }

    // Add job with error handler
    err = scheduler.AddJobWithErrorHandler("monitor-job", "*/30 * * * * *", 
        func() error {
            // Job that might fail
            return performHealthCheck()
        },
        func(err error) {
            log.Printf("Health check failed: %v", err)
        })
    if err != nil {
        log.Fatal("Failed to add monitor job:", err)
    }

    // Start with context
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    scheduler.Start(ctx)

    // Monitor metrics via HTTP (if monitoring enabled)
    // curl http://localhost:8080/metrics
    
    // Keep running
    select {}
}

func performHealthCheck() error {
    // Simulate health check logic
    return nil
}
```

## 📁 Project Structure

The project follows Go best practices and is organized for clarity, maintainability, and ease of contribution:

```
cron/
├── cmd/examples/              # 🎯 Example applications
│   ├── basic/                 # Basic usage patterns
│   ├── advanced/              # Advanced configuration
│   ├── monitoring/            # Monitoring integration
│   └── testing/               # Testing patterns
├── internal/                  # 🔒 Internal packages
│   ├── parser/                # Cron expression parsing
│   ├── scheduler/             # Core scheduling logic
│   ├── monitor/               # Metrics and monitoring
│   ├── types/                 # Internal type definitions
│   └── utils/                 # Utility functions
├── pkg/cron/                  # 📦 Public API
│   ├── cron.go                # Main scheduler implementation
│   ├── config.go              # Configuration management
│   ├── errors.go              # Error definitions
│   └── types.go               # Public types
├── test/                      # 🧪 Test suites
│   ├── integration/           # Integration tests
│   ├── benchmark/             # Performance benchmarks
│   └── testdata/              # Test data and fixtures
├── docs/                      # 📚 Documentation
├── scripts/                   # 🛠️ Build and test scripts
└── .github/workflows/         # 🔄 CI/CD pipelines
```

## 🎯 Comprehensive Examples

The `cmd/examples/` directory contains comprehensive, runnable examples demonstrating various aspects of the cron library:

### 📚 Available Examples

| Example | Description | Key Features |
|---------|-------------|--------------|
| **[basic/](cmd/examples/basic/)** | Essential usage patterns | Simple scheduling, multiple cron formats, graceful shutdown |
| **[advanced/](cmd/examples/advanced/)** | Advanced configurations | Error handling, retries, timeouts, custom configuration |
| **[monitoring/](cmd/examples/monitoring/)** | Monitoring & metrics | HTTP endpoints, real-time metrics, performance tracking |
| **[testing/](cmd/examples/testing/)** | Testing patterns | Unit tests, mocking, benchmarks, test helpers |

### 🏃 Running Examples

```bash
# Basic usage example
go run cmd/examples/basic/main.go

# Advanced features with error handling
go run cmd/examples/advanced/main.go

# Monitoring with HTTP endpoints
go run cmd/examples/monitoring/main.go
# Visit http://localhost:8080 for metrics dashboard

# Testing patterns (includes test functions and helpers)
go run cmd/examples/testing/main.go
```

### 💡 Example Highlights

**Basic Example**: Start here to understand fundamental concepts
- Multiple job scheduling patterns
- 5-field and 6-field cron expressions
- Proper startup and shutdown procedures

**Advanced Example**: Production-ready patterns
- Custom scheduler configuration
- Job retry mechanisms and timeouts
- Comprehensive error handling
- Business hours scheduling
- Real-time status monitoring

**Monitoring Example**: Observability and metrics
- HTTP monitoring endpoints (`/health`, `/metrics`, `/jobs`)
- Real-time performance tracking
- Job execution statistics
- Web-based dashboard access

**Testing Example**: Quality assurance patterns
- Unit testing strategies
- Mock time implementations
- Test helper utilities
- Benchmark writing guide

## 📋 Cron Expression Format

This library supports both 5-field and 6-field cron expressions:

### 6-field format (recommended):
```
┌─────────────── second (0-59)
│ ┌─────────────── minute (0-59)
│ │ ┌─────────────── hour (0-23)
│ │ │ ┌─────────────── day of month (1-31)
│ │ │ │ ┌─────────────── month (1-12)
│ │ │ │ │ ┌─────────────── day of week (0-6, Sunday=0)
│ │ │ │ │ │
* * * * * *
```

### 5-field format (traditional):
```
┌─────────────── minute (0-59)
│ ┌─────────────── hour (0-23)
│ │ ┌─────────────── day of month (1-31)
│ │ │ ┌─────────────── month (1-12)
│ │ │ │ ┌─────────────── day of week (0-6, Sunday=0)
│ │ │ │ │
* * * * *
```

### Supported Operators

| Operator | Description | Example |
|----------|-------------|---------|
| `*` | Any value | `* * * * * *` (every second) |
| `,` | Value list | `0,15,30,45 * * * * *` (every 15 seconds) |
| `-` | Range | `0-5 * * * * *` (seconds 0 through 5) |
| `/` | Step values | `*/10 * * * * *` (every 10 seconds) |

### Example Expressions

| Expression | Description |
|------------|-------------|
| `0 * * * * *` | Every minute |
| `0 0 * * * *` | Every hour |
| `0 0 0 * * *` | Every day at midnight |
| `0 0 9 * * 1-5` | Every weekday at 9 AM |
| `0 */15 * * * *` | Every 15 minutes |
| `0 0 0 1 * *` | First day of every month |
| `*/30 * * * * *` | Every 30 seconds |

## ⚙️ Configuration Options

### Scheduler Configuration

```go
type Config struct {
    Logger            *log.Logger    // Custom logger instance
    Timezone          *time.Location // Timezone for schedule evaluation
    MaxConcurrentJobs int            // Maximum concurrent job executions
    EnableMonitoring  bool           // Enable HTTP monitoring endpoint
    MonitoringPort    int            // Port for monitoring HTTP server
}
```

### Job Configuration

```go
type JobConfig struct {
    MaxRetries    int           // Maximum retry attempts on failure
    RetryInterval time.Duration // Delay between retry attempts
    Timeout       time.Duration // Maximum execution time per job
}
```

### Default Values

```go
// Default scheduler configuration
config := cron.DefaultConfig()
// Logger: os.Stdout with "CRON: " prefix
// Timezone: time.Local
// MaxConcurrentJobs: 100
// EnableMonitoring: false
// MonitoringPort: 8080

// Default job configuration
jobConfig := cron.DefaultJobConfig()
// MaxRetries: 0 (no retries)
// RetryInterval: 1 minute
// Timeout: 0 (no timeout)
```

## 📊 Monitoring and Metrics

The scheduler provides comprehensive monitoring capabilities:

### HTTP Monitoring Endpoint

When monitoring is enabled, the scheduler exposes metrics via HTTP:

```bash
# Get scheduler statistics
curl http://localhost:8080/stats

# Get detailed job information
curl http://localhost:8080/jobs

# Get health check
curl http://localhost:8080/health
```

### Metrics Available

```go
type Stats struct {
    TotalJobs            int           // Total scheduled jobs
    RunningJobs          int           // Currently executing jobs
    SuccessfulExecutions int64         // Total successful executions
    FailedExecutions     int64         // Total failed executions
    AverageExecutionTime time.Duration // Average execution time
    Uptime               time.Duration // Scheduler uptime
}
```

### Programmatic Access

```go
// Get scheduler statistics
stats := scheduler.GetStats()
fmt.Printf("Total jobs: %d\n", stats.TotalJobs)
fmt.Printf("Success rate: %.2f%%\n", 
    float64(stats.SuccessfulExecutions) / 
    float64(stats.SuccessfulExecutions + stats.FailedExecutions) * 100)

// Get individual job statistics
jobStats := scheduler.GetJobStats("backup-job")
fmt.Printf("Job %s: %d runs, %.2f%% success rate\n", 
    jobStats.Name, jobStats.TotalRuns, jobStats.SuccessRate)
```

## 🚦 Error Handling

The library provides comprehensive error handling mechanisms:

### Built-in Errors

```go
var (
    ErrInvalidSchedule       = errors.New("invalid cron schedule expression")
    ErrJobNotFound          = errors.New("job not found")
    ErrJobAlreadyExists     = errors.New("job with this name already exists")
    ErrSchedulerNotStarted  = errors.New("scheduler is not started")
    ErrSchedulerAlreadyStarted = errors.New("scheduler is already started")
    ErrJobTimeout           = errors.New("job execution timeout")
    ErrMaxRetriesExceeded   = errors.New("maximum retry attempts exceeded")
)
```

### Error Handling Patterns

```go
// Method 1: Job with error return
scheduler.AddJobWithErrorHandler("risky-job", "*/5 * * * * *",
    func() error {
        if err := riskyOperation(); err != nil {
            return fmt.Errorf("operation failed: %w", err)
        }
        return nil
    },
    func(err error) {
        log.Printf("Job failed: %v", err)
        // Send alert, log to external system, etc.
    })

// Method 2: Panic recovery (automatic)
scheduler.AddJob("panic-prone-job", "*/1 * * * * *", func() {
    // Even if this panics, the scheduler continues running
    panic("something went wrong")
})

// Method 3: Timeout handling
jobConfig := cron.JobConfig{
    Timeout: 30 * time.Second, // Job will be cancelled after 30 seconds
}
scheduler.AddJobWithConfig("long-job", "0 * * * * *", jobConfig, func() {
    time.Sleep(60 * time.Second) // This will be cancelled
})
```

## 🔒 Thread Safety

The scheduler is fully thread-safe and supports:

- **Concurrent job execution**: Multiple jobs can run simultaneously
- **Safe job management**: Add/remove jobs while scheduler is running
- **Context cancellation**: Graceful shutdown with proper cleanup
- **Resource synchronization**: Internal locks prevent race conditions

```go
// Safe to call from multiple goroutines
go func() {
    scheduler.AddJob("job1", "*/1 * * * * *", func() { /* work */ })
}()

go func() {
    scheduler.RemoveJob("old-job")
}()

go func() {
    stats := scheduler.GetStats() // Thread-safe read access
}()
```

## 🧪 Test Coverage

The project maintains high test coverage across all components:

| Component | Coverage | Type |
|-----------|----------|------|
| Parser | 80.8% | Unit + Integration |
| Scheduler | 97.3% | Unit + Integration + Benchmark |
| Queue | 97.3% | Unit + Integration |
| Core API | 87.3% | Unit + Integration + Benchmark |
| Types | 88.9% | Unit |
| Utils | 52.2% | Unit |
| **Overall** | **75.4%** | **All Types** |

### Running Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run integration tests
go test ./test/integration/...

# Run benchmarks
go test -bench=. ./test/benchmark/...

# Use provided scripts
./scripts/test.sh        # Comprehensive test suite
./scripts/benchmark.sh   # Performance benchmarks
```

## 📈 Performance Characteristics

### Benchmarks (Go 1.19+, Intel i7-12700)

| Operation | Performance | Memory |
|-----------|-------------|---------|
| Cron Parsing | ~3,400 ns/op | Minimal allocation |
| Job Scheduling | ~1,500 ns/op | < 1KB per job |
| Queue Operations | ~100 ns/op | O(log n) complexity |
| Concurrent Access | Linear scaling | Thread-safe |

### Scalability

- **Job Capacity**: Tested with 10,000+ concurrent jobs
- **Memory Usage**: ~10MB for 1,000 active jobs
- **CPU Overhead**: < 1% for typical workloads
- **Scheduling Latency**: < 100ms under normal load

## 🔄 Recent Updates (v0.2.0-beta)

### 🐛 Bug Fixes
- Fixed race condition in integration tests (`TestCronLongRunningJobs`)
- Corrected invalid configuration in benchmark tests
- Standardized cron expression formats across all tests

### ✨ Enhancements
- **Complete Integration Test Suite**: End-to-end testing coverage
- **Comprehensive Benchmark Suite**: Performance testing for all components
- **Improved Thread Safety**: Enhanced synchronization mechanisms
- **Better Configuration Management**: Unified configuration patterns

### 📊 Quality Improvements
- **Test Coverage**: Increased from 60% to 75.4%
- **Test Stability**: Eliminated flaky tests and race conditions
- **Code Quality**: Enhanced error handling and configuration validation

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

### Development Setup

```bash
# Clone the repository
git clone https://github.com/callmebg/cron.git
cd cron

# Install dependencies (none required - pure Go!)
go mod download

# Run tests
./scripts/test.sh

# Run benchmarks
./scripts/benchmark.sh

# Check code quality
./scripts/lint.sh
```

### Project Guidelines

- **Zero Dependencies**: Maintain pure Go standard library usage
- **High Test Coverage**: Aim for >90% test coverage
- **Comprehensive Documentation**: Update docs for all changes
- **Performance Focus**: Benchmark critical paths

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🌟 Acknowledgments

- Inspired by the Unix cron system
- Built with Go's excellent standard library
- Thanks to all contributors and users

---

**Project Status**: 🟢 Beta (Stable) - Ready for Production Use  
**Version**: v0.2.0-beta  
**Go Version**: 1.19+  
**Last Updated**: August 6, 2025  

For questions, bug reports, or feature requests, please [open an issue](https://github.com/callmebg/cron/issues).