package cron

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/callmebg/cron/internal/scheduler"
	"github.com/callmebg/cron/internal/types"
)

// Scheduler represents the main cron scheduler
type Scheduler struct {
	config    Config
	queue     *scheduler.JobQueue
	running   bool
	stopCh    chan struct{}
	mu        sync.RWMutex
	jobIDGen  int64
	jobs      map[string]*scheduler.Job
	stats     Stats
	startTime time.Time
}

// New creates a new scheduler with default configuration
func New() *Scheduler {
	return NewWithConfig(DefaultConfig())
}

// NewWithConfig creates a new scheduler with custom configuration
func NewWithConfig(config Config) *Scheduler {
	if err := types.Config(config).Validate(); err != nil {
		panic(fmt.Sprintf("invalid configuration: %v", err))
	}

	return &Scheduler{
		config:   config,
		queue:    scheduler.NewJobQueue(),
		stopCh:   make(chan struct{}),
		jobs:     make(map[string]*scheduler.Job),
		jobIDGen: 0,
	}
}

// Start starts the scheduler
func (s *Scheduler) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return ErrSchedulerAlreadyStarted
	}

	s.running = true
	s.startTime = time.Now()
	s.config.Logger.Println("Cron scheduler started")
	s.stopCh = make(chan struct{})

	// Start the main scheduling loop
	go s.run(ctx)

	return nil
}

// Stop stops the scheduler gracefully
func (s *Scheduler) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return ErrSchedulerNotStarted
	}

	s.running = false
	close(s.stopCh)
	s.config.Logger.Println("Cron scheduler stopped")

	return nil
}

// AddJob adds a job with cron expression
func (s *Scheduler) AddJob(schedule string, job JobFunc) error {
	return s.AddNamedJob("", schedule, job)
}

// AddNamedJob adds a named job for better monitoring
func (s *Scheduler) AddNamedJob(name, schedule string, job JobFunc) error {
	config := DefaultJobConfig()
	return s.AddJobWithConfig(name, schedule, config, job)
}

// AddJobWithErrorHandler adds a job with error handling
func (s *Scheduler) AddJobWithErrorHandler(name, schedule string, job JobFuncWithError, errorHandler ErrorHandler) error {
	config := DefaultJobConfig()
	return s.AddJobWithErrorConfig(name, schedule, config, job, errorHandler)
}

// AddJobWithConfig adds a job with custom configuration
func (s *Scheduler) AddJobWithConfig(name, schedule string, config JobConfig, job JobFunc) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := types.JobConfig(config).Validate(); err != nil {
		return err
	}

	// Generate job ID and name if not provided
	s.jobIDGen++
	jobID := fmt.Sprintf("job_%d", s.jobIDGen)
	if name == "" {
		name = jobID
	}

	// Check if job with this name already exists
	if _, exists := s.jobs[name]; exists {
		return ErrJobAlreadyExists
	}

	// Create new job
	newJob, err := scheduler.NewJob(jobID, name, schedule, types.JobFunc(job), types.JobConfig(config), s.config.Timezone)
	if err != nil {
		return err
	}

	// Add to internal tracking
	s.jobs[name] = newJob
	s.queue.Add(newJob)
	s.stats.TotalJobs++

	s.config.Logger.Printf("Added job '%s' with schedule '%s'", name, schedule)
	return nil
}

// AddJobWithErrorConfig adds a job with error handling and custom configuration
func (s *Scheduler) AddJobWithErrorConfig(name, schedule string, config JobConfig, job JobFuncWithError, errorHandler ErrorHandler) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := types.JobConfig(config).Validate(); err != nil {
		return err
	}

	// Generate job ID and name if not provided
	s.jobIDGen++
	jobID := fmt.Sprintf("job_%d", s.jobIDGen)
	if name == "" {
		name = jobID
	}

	// Check if job with this name already exists
	if _, exists := s.jobs[name]; exists {
		return ErrJobAlreadyExists
	}

	// Create new job with error handling
	newJob, err := scheduler.NewJobWithError(jobID, name, schedule, types.JobFuncWithError(job), types.ErrorHandler(errorHandler), types.JobConfig(config), s.config.Timezone)
	if err != nil {
		return err
	}

	// Add to internal tracking
	s.jobs[name] = newJob
	s.queue.Add(newJob)
	s.stats.TotalJobs++

	s.config.Logger.Printf("Added job with error handling '%s' with schedule '%s'", name, schedule)
	return nil
}

// RemoveJob removes a job by name
func (s *Scheduler) RemoveJob(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	job, exists := s.jobs[name]
	if !exists {
		return ErrJobNotFound
	}

	// Remove from queue and tracking
	s.queue.Remove(job.ID)
	delete(s.jobs, name)
	s.stats.TotalJobs--

	s.config.Logger.Printf("Removed job '%s'", name)
	return nil
}

// GetStats returns scheduler statistics
func (s *Scheduler) GetStats() Stats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := s.stats
	stats.TotalJobs = len(s.jobs)
	stats.RunningJobs = s.countRunningJobs()

	if s.running {
		stats.Uptime = time.Since(s.startTime)
	}

	// Calculate average execution time
	if stats.SuccessfulExecutions > 0 {
		totalDuration := s.calculateTotalExecutionTime()
		stats.AverageExecutionTime = time.Duration(totalDuration.Nanoseconds() / stats.SuccessfulExecutions)
	}

	return stats
}

// GetJobStats returns job-specific statistics
func (s *Scheduler) GetJobStats(name string) (JobStats, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	job, exists := s.jobs[name]
	if !exists {
		return JobStats{}, ErrJobNotFound
	}

	return job.GetStats(), nil
}

// ListJobs returns a list of all job names
func (s *Scheduler) ListJobs() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	names := make([]string, 0, len(s.jobs))
	for name := range s.jobs {
		names = append(names, name)
	}
	return names
}

// IsRunning returns true if the scheduler is running
func (s *Scheduler) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

// run is the main scheduling loop
func (s *Scheduler) run(ctx context.Context) {
	ticker := time.NewTicker(time.Second) // Check every second
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.config.Logger.Println("Scheduler context cancelled")
			return
		case <-s.stopCh:
			s.config.Logger.Println("Scheduler stop signal received")
			return
		case now := <-ticker.C:
			s.processReadyJobs(ctx, now)
		}
	}
}

// processReadyJobs processes all jobs that are ready to run
func (s *Scheduler) processReadyJobs(ctx context.Context, now time.Time) {
	readyJobs := s.queue.PopReady(now)

	for _, job := range readyJobs {
		// Check if we've reached the concurrent job limit
		if s.countRunningJobs() >= s.config.MaxConcurrentJobs {
			// Re-add job to queue for next cycle
			s.queue.Add(job)
			s.config.Logger.Printf("Job '%s' deferred due to concurrent job limit", job.Name)
			continue
		}

		// Execute job in a separate goroutine
		go s.executeJob(ctx, job)
	}
}

// executeJob executes a single job
func (s *Scheduler) executeJob(ctx context.Context, job *scheduler.Job) {
	s.mu.Lock()
	s.stats.RunningJobs++
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		s.stats.RunningJobs--
		s.mu.Unlock()

		// Re-add job to queue for next execution
		job.UpdateNextExecution(time.Now())
		s.queue.Add(job)
	}()

	// Execute the job
	err := job.Execute(ctx)

	// Update statistics
	s.mu.Lock()
	if err != nil {
		s.stats.FailedExecutions++
		s.config.Logger.Printf("Job '%s' failed: %v", job.Name, err)
	} else {
		s.stats.SuccessfulExecutions++
		s.config.Logger.Printf("Job '%s' completed successfully", job.Name)
	}
	s.mu.Unlock()
}

// countRunningJobs counts the number of currently running jobs
func (s *Scheduler) countRunningJobs() int {
	count := 0
	for _, job := range s.jobs {
		if job.IsJobRunning() {
			count++
		}
	}
	return count
}

// calculateTotalExecutionTime calculates the total execution time of all jobs
func (s *Scheduler) calculateTotalExecutionTime() time.Duration {
	var total time.Duration
	for _, job := range s.jobs {
		lastDuration, runCount := job.GetJobStats()
		total += lastDuration * time.Duration(runCount)
	}
	return total
}

// GetNextExecutions returns the next n execution times for all jobs
func (s *Scheduler) GetNextExecutions(count int) map[string][]time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string][]time.Time)
	for name, job := range s.jobs {
		executions := make([]time.Time, 0, count)
		current := time.Now()

		for i := 0; i < count; i++ {
			next := job.Schedule.Next(current)
			if next.IsZero() {
				break
			}
			executions = append(executions, next)
			current = next
		}

		result[name] = executions
	}

	return result
}
