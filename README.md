# Go Cron - Native Cron Job Scheduler

A lightweight, efficient cron job scheduler for Go applications built exclusively with the Go standard library. This library provides robust task scheduling with comprehensive monitoring capabilities while maintaining zero external dependencies.

English | [中文](README_zh.md)

## Features

- **Pure Go Standard Library**: No external dependencies, leveraging Go's native capabilities
- **Standard Cron Syntax**: Full support for traditional cron expressions
- **Concurrent Execution**: Efficient task execution using goroutines and channels
- **Built-in Monitoring**: Comprehensive metrics and execution tracking
- **Graceful Shutdown**: Clean termination with context-based cancellation
- **Thread-Safe**: Safe for concurrent use across multiple goroutines
- **Flexible Logging**: Configurable logging with standard `log` package integration
- **Memory Efficient**: Minimal memory footprint with efficient scheduling algorithms

## Project Structure

The project follows Go conventions and is organized for clarity, maintainability, and ease of contribution:

```
cron/
├── cmd/
│   └── examples/
│       ├── basic/
│       │   └── main.go              # Basic usage example
│       ├── advanced/
│       │   └── main.go              # Advanced configuration example
│       ├── monitoring/
│       │   └── main.go              # Monitoring and metrics example
│       └── testing/
│           └── main.go              # Testing patterns example
├── internal/
│   ├── parser/
│   │   ├── parser.go                # Cron expression parser
│   │   ├── parser_test.go           # Parser unit tests
│   │   └── validator.go             # Expression validation
│   ├── scheduler/
│   │   ├── scheduler.go             # Core scheduling logic
│   │   ├── scheduler_test.go        # Scheduler tests
│   │   ├── job.go                   # Job definition and management
│   │   ├── job_test.go              # Job tests
│   │   └── queue.go                 # Priority queue implementation
│   ├── monitor/
│   │   ├── metrics.go               # Metrics collection
│   │   ├── metrics_test.go          # Metrics tests
│   │   ├── http.go                  # HTTP monitoring endpoint
│   │   └── stats.go                 # Statistics aggregation
│   └── utils/
│       ├── time.go                  # Time utilities
│       ├── time_test.go             # Time utility tests
│       └── sync.go                  # Synchronization helpers
├── pkg/
│   └── cron/
│       ├── cron.go                  # Public API and main entry point
│       ├── cron_test.go             # Integration tests
│       ├── config.go                # Configuration types and defaults
│       ├── config_test.go           # Configuration tests
│       ├── errors.go                # Error definitions
│       └── types.go                 # Public type definitions
├── test/
│   ├── integration/
│   │   ├── scheduler_test.go        # Integration tests
│   │   └── monitoring_test.go       # Monitoring integration tests
│   ├── benchmark/
│   │   ├── scheduler_bench_test.go  # Performance benchmarks
│   │   └── memory_bench_test.go     # Memory usage benchmarks
│   └── testdata/
│       ├── cron_expressions.txt     # Test cron expressions
│       └── expected_schedules.json  # Expected scheduling results
├── docs/
│   ├── architecture.md              # Architecture overview
│   ├── performance.md               # Performance characteristics
│   └── contributing.md              # Detailed contribution guide
├── scripts/
│   ├── test.sh                      # Comprehensive testing script
│   ├── benchmark.sh                 # Benchmarking script
│   └── lint.sh                      # Code quality checks
├── .github/
│   └── workflows/
│       ├── ci.yml                   # Continuous integration
│       ├── release.yml              # Release automation
│       └── security.yml             # Security scanning
├── go.mod                           # Go module definition
├── go.sum                           # Dependency checksums
├── LICENSE                          # MIT license
├── README.md                        # This file
├── CHANGELOG.md                     # Version history
└── Makefile                         # Build automation

```

### Directory Explanations

#### `/cmd/examples/`
Contains practical examples demonstrating different use cases:
- **basic/**: Simple cron job setup and execution
- **advanced/**: Complex configurations with error handling
- **monitoring/**: HTTP monitoring and metrics collection
- **testing/**: Testing patterns and best practices

#### `/internal/`
Private implementation packages not intended for external use:
- **parser/**: Cron expression parsing and validation logic
- **scheduler/**: Core scheduling engine and job queue management
- **monitor/**: Metrics collection and HTTP monitoring endpoints
- **utils/**: Shared utilities for time handling and synchronization

#### `/pkg/cron/`
Public API package that users import:
- **cron.go**: Main scheduler interface and public methods
- **config.go**: Configuration structures and defaults
- **types.go**: Public type definitions and interfaces
- **errors.go**: Public error types and error handling

#### `/test/`
Comprehensive testing suite:
- **integration/**: End-to-end integration tests
- **benchmark/**: Performance and memory benchmarks
- **testdata/**: Test fixtures and expected results

#### `/docs/`
Additional documentation beyond the README:
- **architecture.md**: Internal architecture and design decisions
- **performance.md**: Performance characteristics and optimization notes
- **contributing.md**: Detailed development and contribution guidelines

#### `/scripts/`
Development and CI automation scripts:
- **test.sh**: Runs all tests with coverage reporting
- **benchmark.sh**: Executes performance benchmarks
- **lint.sh**: Code quality and formatting checks

#### `/.github/workflows/`
GitHub Actions for CI/CD:
- **ci.yml**: Automated testing across Go versions and platforms
- **release.yml**: Automated release and tagging
- **security.yml**: Security vulnerability scanning

### Package Organization Principles

1. **Clear Separation**: Public API in `/pkg/cron/`, implementation in `/internal/`
2. **Single Responsibility**: Each package has a focused, well-defined purpose
3. **Testability**: All packages designed for easy unit and integration testing
4. **Documentation**: Comprehensive examples and documentation for all public APIs
5. **Standard Library Only**: Zero external dependencies, leveraging Go's built-in capabilities

### Import Path Structure

```go
// Public API - what users import
import "github.com/callmebg/cron/pkg/cron"

// Internal packages - not accessible to users
// internal/parser
// internal/scheduler  
// internal/monitor
// internal/utils
```

This structure ensures clean separation between public API and implementation details while providing comprehensive testing, documentation, and examples.

## Installation

```bash
go get github.com/callmebg/cron
```

## Quick Start

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
    // Create a new cron scheduler
    c := cron.New()

    // Add a job that runs every minute
    err := c.AddJob("* * * * *", func() {
        fmt.Println("Job executed at:", time.Now())
    })
    if err != nil {
        log.Fatal(err)
    }

    // Start the scheduler
    ctx := context.Background()
    c.Start(ctx)

    // Keep the program running
    select {}
}
```

## Cron Syntax

The library supports the standard 5-field cron syntax:

```
┌─────────────── minute (0 - 59)
│ ┌───────────── hour (0 - 23)
│ │ ┌─────────── day of month (1 - 31)
│ │ │ ┌───────── month (1 - 12)
│ │ │ │ ┌─────── day of week (0 - 6) (Sunday to Saturday)
│ │ │ │ │
│ │ │ │ │
* * * * *
```

### Special Characters

- `*` - Any value
- `,` - Value list separator
- `-` - Range of values
- `/` - Step values

### Examples

| Expression | Description |
|------------|-------------|
| `0 */2 * * *` | Every 2 hours |
| `30 9 * * 1-5` | 9:30 AM on weekdays |
| `0 0 1 * *` | First day of every month |
| `*/15 * * * *` | Every 15 minutes |
| `0 22 * * 0` | 10 PM every Sunday |

## Usage Examples

### Basic Job Scheduling

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/callmebg/cron/pkg/cron"
)

func main() {
    c := cron.New()

    // Simple periodic task
    c.AddJob("*/5 * * * *", func() {
        fmt.Println("This runs every 5 minutes")
    })

    // Named job for better monitoring
    c.AddNamedJob("backup", "0 2 * * *", func() {
        fmt.Println("Running nightly backup")
        // Your backup logic here
    })

    // Job with error handling
    c.AddJobWithErrorHandler("0 */6 * * *", 
        func() error {
            // Your job logic here
            return doSomeWork()
        },
        func(err error) {
            log.Printf("Job failed: %v", err)
        },
    )

    c.Start(context.Background())
    select {}
}

func doSomeWork() error {
    // Simulate work that might fail
    return nil
}
```

### Advanced Configuration

```go
package main

import (
    "context"
    "log"
    "os"
    "time"

    "github.com/callmebg/cron/pkg/cron"
)

func main() {
    // Custom configuration
    config := cron.Config{
        Logger:           log.New(os.Stdout, "CRON: ", log.LstdFlags),
        Timezone:         time.UTC,
        MaxConcurrentJobs: 10,
        EnableMonitoring: true,
        MonitoringPort:   8080,
    }

    c := cron.NewWithConfig(config)

    // Add jobs with different configurations
    jobConfig := cron.JobConfig{
        MaxRetries:    3,
        RetryInterval: time.Minute * 5,
        Timeout:       time.Hour,
    }

    c.AddJobWithConfig("0 */1 * * *", jobConfig, func() {
        // This job will retry up to 3 times if it fails
        // and will timeout after 1 hour
    })

    c.Start(context.Background())
    select {}
}
```

## Monitoring Capabilities

### Built-in Metrics

The library provides comprehensive monitoring out of the box:

```go
// Get scheduler statistics
stats := c.GetStats()
fmt.Printf("Total jobs: %d\n", stats.TotalJobs)
fmt.Printf("Running jobs: %d\n", stats.RunningJobs)
fmt.Printf("Successful executions: %d\n", stats.SuccessfulExecutions)
fmt.Printf("Failed executions: %d\n", stats.FailedExecutions)
fmt.Printf("Average execution time: %v\n", stats.AverageExecutionTime)
```

### Job-Level Monitoring

```go
// Get statistics for a specific job
jobStats := c.GetJobStats("backup")
fmt.Printf("Last execution: %v\n", jobStats.LastExecution)
fmt.Printf("Next execution: %v\n", jobStats.NextExecution)
fmt.Printf("Success rate: %.2f%%\n", jobStats.SuccessRate)
fmt.Printf("Total runs: %d\n", jobStats.TotalRuns)
```

### HTTP Monitoring Endpoint

When monitoring is enabled, the scheduler exposes an HTTP endpoint:

```go
c := cron.NewWithConfig(cron.Config{
    EnableMonitoring: true,
    MonitoringPort:   8080,
})
```

Access monitoring data at `http://localhost:8080/metrics`:

```json
{
  "scheduler": {
    "uptime": "2h45m30s",
    "total_jobs": 5,
    "running_jobs": 2,
    "successful_executions": 127,
    "failed_executions": 3
  },
  "jobs": [
    {
      "name": "backup",
      "schedule": "0 2 * * *",
      "last_execution": "2024-01-15T02:00:00Z",
      "next_execution": "2024-01-16T02:00:00Z",
      "status": "completed",
      "execution_time": "45s",
      "success_rate": 98.5
    }
  ]
}
```

## API Reference

### Core Types

#### Scheduler

```go
type Scheduler struct {
    // Internal fields
}

// Create a new scheduler with default configuration
func New() *Scheduler

// Create a new scheduler with custom configuration
func NewWithConfig(config Config) *Scheduler

// Start the scheduler
func (s *Scheduler) Start(ctx context.Context)

// Stop the scheduler gracefully
func (s *Scheduler) Stop() error

// Add a job with cron expression
func (s *Scheduler) AddJob(schedule string, job func()) error

// Add a named job for better monitoring
func (s *Scheduler) AddNamedJob(name, schedule string, job func()) error

// Add a job with error handling
func (s *Scheduler) AddJobWithErrorHandler(schedule string, job func() error, errorHandler func(error)) error

// Add a job with custom configuration
func (s *Scheduler) AddJobWithConfig(schedule string, config JobConfig, job func()) error

// Remove a job by name
func (s *Scheduler) RemoveJob(name string) error

// Get scheduler statistics
func (s *Scheduler) GetStats() Stats

// Get job-specific statistics
func (s *Scheduler) GetJobStats(name string) JobStats
```

#### Configuration

```go
type Config struct {
    Logger           *log.Logger
    Timezone         *time.Location
    MaxConcurrentJobs int
    EnableMonitoring bool
    MonitoringPort   int
}

type JobConfig struct {
    MaxRetries    int
    RetryInterval time.Duration
    Timeout       time.Duration
}
```

#### Statistics

```go
type Stats struct {
    TotalJobs              int
    RunningJobs            int
    SuccessfulExecutions   int64
    FailedExecutions       int64
    AverageExecutionTime   time.Duration
    Uptime                 time.Duration
}

type JobStats struct {
    Name            string
    Schedule        string
    LastExecution   time.Time
    NextExecution   time.Time
    Status          string
    ExecutionTime   time.Duration
    SuccessRate     float64
    TotalRuns       int64
}
```

## Configuration Options

### Scheduler Configuration

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `Logger` | `*log.Logger` | `log.Default()` | Logger instance for scheduler events |
| `Timezone` | `*time.Location` | `time.Local` | Timezone for schedule evaluation |
| `MaxConcurrentJobs` | `int` | `100` | Maximum number of concurrent job executions |
| `EnableMonitoring` | `bool` | `false` | Enable HTTP monitoring endpoint |
| `MonitoringPort` | `int` | `8080` | Port for monitoring HTTP server |

### Job Configuration

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `MaxRetries` | `int` | `0` | Maximum number of retry attempts |
| `RetryInterval` | `time.Duration` | `time.Minute` | Delay between retry attempts |
| `Timeout` | `time.Duration` | `0` (no timeout) | Maximum execution time per job |

## Best Practices

### 1. Use Context for Cancellation

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

c.Start(ctx)

// Graceful shutdown
go func() {
    <-sigterm
    cancel()
    c.Stop()
}()
```

### 2. Handle Errors Properly

```go
c.AddJobWithErrorHandler("*/5 * * * *",
    func() error {
        return performTask()
    },
    func(err error) {
        log.Printf("Task failed: %v", err)
        // Send alert, log to external system, etc.
    },
)
```

### 3. Use Named Jobs for Monitoring

```go
// Good: Named job for easy identification
c.AddNamedJob("database-cleanup", "0 3 * * *", cleanupDatabase)

// Better: With configuration
c.AddJobWithConfig("user-sync", "*/30 * * * *", 
    cron.JobConfig{
        MaxRetries: 3,
        Timeout: time.Minute * 10,
    },
    syncUsers,
)
```

### 4. Resource Management

```go
func resourceIntensiveJob() {
    // Limit resource usage
    semaphore := make(chan struct{}, 5) // Max 5 concurrent operations
    
    var wg sync.WaitGroup
    for _, item := range items {
        wg.Add(1)
        go func(item string) {
            defer wg.Done()
            semaphore <- struct{}{}        // Acquire
            defer func() { <-semaphore }() // Release
            
            processItem(item)
        }(item)
    }
    wg.Wait()
}
```

### 5. Testing Cron Jobs

```go
func TestJobExecution(t *testing.T) {
    c := cron.New()
    
    executed := false
    err := c.AddJob("* * * * *", func() {
        executed = true
    })
    require.NoError(t, err)
    
    ctx, cancel := context.WithTimeout(context.Background(), time.Minute*2)
    defer cancel()
    
    go c.Start(ctx)
    
    // Wait for execution
    time.Sleep(time.Minute + time.Second*5)
    assert.True(t, executed)
}
```

## Performance Considerations

### Memory Usage

- The scheduler uses a heap-based priority queue for efficient job scheduling
- Memory usage scales linearly with the number of scheduled jobs
- Each job maintains minimal metadata for monitoring

### CPU Usage

- Scheduling overhead is minimal, using time-based triggers
- Job execution happens in separate goroutines to prevent blocking
- The scheduler sleeps between evaluations to minimize CPU usage

### Concurrency

- Safe for concurrent access from multiple goroutines
- Uses sync.RWMutex for thread-safe operations
- Configurable maximum concurrent job execution limit

## Error Handling

The library provides multiple levels of error handling:

1. **Schedule Parsing Errors**: Returned immediately when adding jobs
2. **Job Execution Errors**: Handled via error handlers or logged
3. **System Errors**: Logged and reported through monitoring

Example comprehensive error handling:

```go
c := cron.NewWithConfig(cron.Config{
    Logger: log.New(os.Stderr, "CRON-ERROR: ", log.LstdFlags),
})

// Handle schedule parsing errors
if err := c.AddJob("invalid cron", myJob); err != nil {
    log.Fatalf("Invalid cron expression: %v", err)
}

// Handle job execution errors
c.AddJobWithErrorHandler("0 */1 * * *",
    func() error {
        return riskyOperation()
    },
    func(err error) {
        // Custom error handling
        notifyAdmin(err)
        logToExternalSystem(err)
    },
)
```

## Contributing

We welcome contributions! Please follow these guidelines:

### Development Setup

1. Clone the repository:
   ```bash
   git clone https://github.com/callmebg/cron.git
   cd cron
   ```

2. Run tests:
   ```bash
   go test ./...
   ```

3. Run benchmarks:
   ```bash
   go test -bench=. ./...
   ```

### Code Standards

- Follow Go formatting conventions (`gofmt`)
- Write comprehensive tests for new features
- Update documentation for API changes
- Use only Go standard library dependencies
- Maintain backward compatibility

### Submitting Changes

1. Fork the repository
2. Create a feature branch: `git checkout -b feature-name`
3. Commit your changes: `git commit -am 'Add feature'`
4. Push to the branch: `git push origin feature-name`
5. Submit a pull request

### Testing

All contributions must include tests:

```go
func TestNewFeature(t *testing.T) {
    c := cron.New()
    
    // Test your feature
    err := c.YourNewMethod()
    assert.NoError(t, err)
}

func BenchmarkNewFeature(b *testing.B) {
    c := cron.New()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        c.YourNewMethod()
    }
}
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Testing

This project follows comprehensive testing practices to ensure reliability and performance. See our detailed [Test Plan](TEST_PLAN.md) for complete testing strategy.

### Test Coverage

- **Unit Tests**: 90%+ coverage across all packages
- **Integration Tests**: End-to-end scheduling workflows
- **Benchmark Tests**: Performance and memory usage validation
- **Compatibility Tests**: Cross-platform and Go version testing

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -coverprofile=coverage.out ./...

# Run benchmarks
go test -bench=. ./test/benchmark/

# Run integration tests
go test ./test/integration/
```

## Comparison with Other Libraries

We've conducted a detailed comparison with other popular Go cron libraries. See [COMPARISON.md](COMPARISON.md) for comprehensive analysis.

### Quick Comparison

| Feature | callmebg/cron | robfig/cron | go-co-op/gocron |
|---------|--------------|-------------|-----------------|
| Zero Dependencies | ✅ | ✅ | ❌ |
| Built-in Monitoring | ✅ | ❌ | ⚠️ |
| Error Handling | ✅ Advanced | ⚠️ Basic | ✅ Good |
| Concurrency Control | ✅ | ❌ | ⚠️ |
| Production Ready | ✅ | ✅ | ✅ |

### Why Choose callmebg/cron?

- **Zero Dependencies**: Built with Go standard library only
- **Production Ready**: Comprehensive monitoring and error handling
- **High Performance**: Optimized for low memory usage and high throughput
- **Developer Friendly**: Clean API with extensive documentation
- **Enterprise Features**: Retry mechanisms, timeouts, and graceful shutdown

## Roadmap

- [ ] Support for seconds-level precision (6-field cron syntax)
- [ ] Distributed scheduling capabilities
- [ ] Web UI for job management
- [ ] Prometheus metrics export
- [ ] Job persistence across restarts
- [ ] Dynamic job modification
- [ ] Job dependency management

## Support

- Create an issue for bug reports or feature requests
- Check existing issues before creating new ones
- Provide minimal reproduction cases for bugs
- Include Go version and OS information in bug reports

---

Built with Go's standard library - no external dependencies required.