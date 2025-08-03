package monitor

import (
	"sort"
	"sync"
	"time"
)

// StatsAggregator aggregates and maintains historical statistics
type StatsAggregator struct {
	mu          sync.RWMutex
	hourlyStats []HourlyStats
	dailyStats  []DailyStats
	jobStats    map[string]*JobStatsHistory
	maxHours    int
	maxDays     int
}

// NewStatsAggregator creates a new statistics aggregator
func NewStatsAggregator() *StatsAggregator {
	return &StatsAggregator{
		hourlyStats: make([]HourlyStats, 0),
		dailyStats:  make([]DailyStats, 0),
		jobStats:    make(map[string]*JobStatsHistory),
		maxHours:    24, // Keep 24 hours of hourly stats
		maxDays:     30, // Keep 30 days of daily stats
	}
}

// HourlyStats represents statistics for a single hour
type HourlyStats struct {
	Timestamp        time.Time `json:"timestamp"`
	JobsExecuted     int64     `json:"jobs_executed"`
	JobsSucceeded    int64     `json:"jobs_succeeded"`
	JobsFailed       int64     `json:"jobs_failed"`
	AverageExecTime  float64   `json:"average_execution_time_ms"`
	MaxExecTime      float64   `json:"max_execution_time_ms"`
	MinExecTime      float64   `json:"min_execution_time_ms"`
	ThroughputPerMin float64   `json:"throughput_per_minute"`
}

// DailyStats represents statistics for a single day
type DailyStats struct {
	Date            time.Time `json:"date"`
	JobsExecuted    int64     `json:"jobs_executed"`
	JobsSucceeded   int64     `json:"jobs_succeeded"`
	JobsFailed      int64     `json:"jobs_failed"`
	AverageExecTime float64   `json:"average_execution_time_ms"`
	MaxExecTime     float64   `json:"max_execution_time_ms"`
	MinExecTime     float64   `json:"min_execution_time_ms"`
	UptimePercent   float64   `json:"uptime_percent"`
	PeakThroughput  float64   `json:"peak_throughput_per_minute"`
}

// JobStatsHistory maintains historical data for a specific job
type JobStatsHistory struct {
	JobName          string                `json:"job_name"`
	TotalExecutions  int64                 `json:"total_executions"`
	SuccessfulRuns   int64                 `json:"successful_runs"`
	FailedRuns       int64                 `json:"failed_runs"`
	AverageExecTime  time.Duration         `json:"average_execution_time"`
	LastExecution    time.Time             `json:"last_execution"`
	SuccessRate      float64               `json:"success_rate"`
	RecentExecutions []JobExecutionHistory `json:"recent_executions"`
}

// JobExecutionHistory represents a single job execution
type JobExecutionHistory struct {
	Timestamp    time.Time     `json:"timestamp"`
	Duration     time.Duration `json:"duration"`
	Success      bool          `json:"success"`
	ErrorMessage string        `json:"error_message,omitempty"`
}

// RecordJobExecution records a job execution in the history
func (sa *StatsAggregator) RecordJobExecution(jobName string, duration time.Duration, success bool, errorMsg string) {
	sa.mu.Lock()
	defer sa.mu.Unlock()

	// Initialize job stats if not exists
	if sa.jobStats[jobName] == nil {
		sa.jobStats[jobName] = &JobStatsHistory{
			JobName:          jobName,
			RecentExecutions: make([]JobExecutionHistory, 0),
		}
	}

	jobStats := sa.jobStats[jobName]

	// Update counters
	jobStats.TotalExecutions++
	if success {
		jobStats.SuccessfulRuns++
	} else {
		jobStats.FailedRuns++
	}

	// Update execution time average
	totalTime := jobStats.AverageExecTime * time.Duration(jobStats.TotalExecutions-1)
	jobStats.AverageExecTime = (totalTime + duration) / time.Duration(jobStats.TotalExecutions)

	// Update last execution
	jobStats.LastExecution = time.Now()

	// Calculate success rate
	if jobStats.TotalExecutions > 0 {
		jobStats.SuccessRate = float64(jobStats.SuccessfulRuns) / float64(jobStats.TotalExecutions) * 100
	}

	// Add to recent executions (keep only last 50)
	execution := JobExecutionHistory{
		Timestamp:    time.Now(),
		Duration:     duration,
		Success:      success,
		ErrorMessage: errorMsg,
	}

	jobStats.RecentExecutions = append(jobStats.RecentExecutions, execution)
	if len(jobStats.RecentExecutions) > 50 {
		jobStats.RecentExecutions = jobStats.RecentExecutions[1:]
	}
}

// UpdateHourlyStats updates the hourly statistics
func (sa *StatsAggregator) UpdateHourlyStats(metrics *Metrics) {
	sa.mu.Lock()
	defer sa.mu.Unlock()

	now := time.Now().Truncate(time.Hour)
	snapshot := metrics.GetSnapshot()

	// Find or create current hour stats
	var currentHour *HourlyStats
	for i := len(sa.hourlyStats) - 1; i >= 0; i-- {
		if sa.hourlyStats[i].Timestamp.Equal(now) {
			currentHour = &sa.hourlyStats[i]
			break
		}
	}

	if currentHour == nil {
		// Create new hour stats
		newStats := HourlyStats{
			Timestamp: now,
		}
		sa.hourlyStats = append(sa.hourlyStats, newStats)
		currentHour = &sa.hourlyStats[len(sa.hourlyStats)-1]
	}

	// Update stats
	currentHour.JobsExecuted = snapshot.TotalJobs
	currentHour.JobsSucceeded = snapshot.CompletedJobs
	currentHour.JobsFailed = snapshot.FailedJobs
	currentHour.AverageExecTime = float64(snapshot.AverageExecutionTime.Nanoseconds()) / 1e6 // Convert to milliseconds
	currentHour.MaxExecTime = float64(snapshot.MaxExecutionTime.Nanoseconds()) / 1e6
	currentHour.MinExecTime = float64(snapshot.MinExecutionTime.Nanoseconds()) / 1e6
	currentHour.ThroughputPerMin = snapshot.JobsPerMinute

	// Keep only recent hours
	if len(sa.hourlyStats) > sa.maxHours {
		sa.hourlyStats = sa.hourlyStats[len(sa.hourlyStats)-sa.maxHours:]
	}
}

// UpdateDailyStats updates the daily statistics
func (sa *StatsAggregator) UpdateDailyStats(metrics *Metrics, uptimePercent float64) {
	sa.mu.Lock()
	defer sa.mu.Unlock()

	today := time.Now().Truncate(24 * time.Hour)
	snapshot := metrics.GetSnapshot()

	// Find or create today's stats
	var currentDay *DailyStats
	for i := len(sa.dailyStats) - 1; i >= 0; i-- {
		if sa.dailyStats[i].Date.Equal(today) {
			currentDay = &sa.dailyStats[i]
			break
		}
	}

	if currentDay == nil {
		// Create new day stats
		newStats := DailyStats{
			Date: today,
		}
		sa.dailyStats = append(sa.dailyStats, newStats)
		currentDay = &sa.dailyStats[len(sa.dailyStats)-1]
	}

	// Update stats
	currentDay.JobsExecuted = snapshot.TotalJobs
	currentDay.JobsSucceeded = snapshot.CompletedJobs
	currentDay.JobsFailed = snapshot.FailedJobs
	currentDay.AverageExecTime = float64(snapshot.AverageExecutionTime.Nanoseconds()) / 1e6
	currentDay.MaxExecTime = float64(snapshot.MaxExecutionTime.Nanoseconds()) / 1e6
	currentDay.MinExecTime = float64(snapshot.MinExecutionTime.Nanoseconds()) / 1e6
	currentDay.UptimePercent = uptimePercent

	// Calculate peak throughput from hourly stats for today
	currentDay.PeakThroughput = sa.calculatePeakThroughputForDay(today)

	// Keep only recent days
	if len(sa.dailyStats) > sa.maxDays {
		sa.dailyStats = sa.dailyStats[len(sa.dailyStats)-sa.maxDays:]
	}
}

// GetJobStatsHistory returns historical statistics for a specific job
func (sa *StatsAggregator) GetJobStatsHistory(jobName string) (*JobStatsHistory, bool) {
	sa.mu.RLock()
	defer sa.mu.RUnlock()

	stats, exists := sa.jobStats[jobName]
	if !exists {
		return nil, false
	}

	// Return a copy to avoid race conditions
	statsCopy := *stats
	statsCopy.RecentExecutions = make([]JobExecutionHistory, len(stats.RecentExecutions))
	copy(statsCopy.RecentExecutions, stats.RecentExecutions)

	return &statsCopy, true
}

// GetHourlyStats returns the hourly statistics
func (sa *StatsAggregator) GetHourlyStats() []HourlyStats {
	sa.mu.RLock()
	defer sa.mu.RUnlock()

	// Return a copy to avoid race conditions
	stats := make([]HourlyStats, len(sa.hourlyStats))
	copy(stats, sa.hourlyStats)
	return stats
}

// GetDailyStats returns the daily statistics
func (sa *StatsAggregator) GetDailyStats() []DailyStats {
	sa.mu.RLock()
	defer sa.mu.RUnlock()

	// Return a copy to avoid race conditions
	stats := make([]DailyStats, len(sa.dailyStats))
	copy(stats, sa.dailyStats)
	return stats
}

// GetTopJobsByExecutions returns jobs sorted by execution count
func (sa *StatsAggregator) GetTopJobsByExecutions(limit int) []JobStatsHistory {
	sa.mu.RLock()
	defer sa.mu.RUnlock()

	var jobs []JobStatsHistory
	for _, stats := range sa.jobStats {
		jobs = append(jobs, *stats)
	}

	// Sort by total executions (descending)
	sort.Slice(jobs, func(i, j int) bool {
		return jobs[i].TotalExecutions > jobs[j].TotalExecutions
	})

	if limit > 0 && len(jobs) > limit {
		jobs = jobs[:limit]
	}

	return jobs
}

// GetTopJobsByErrors returns jobs sorted by error rate
func (sa *StatsAggregator) GetTopJobsByErrors(limit int) []JobStatsHistory {
	sa.mu.RLock()
	defer sa.mu.RUnlock()

	var jobs []JobStatsHistory
	for _, stats := range sa.jobStats {
		if stats.TotalExecutions > 0 { // Only include jobs that have run
			jobs = append(jobs, *stats)
		}
	}

	// Sort by error rate (descending)
	sort.Slice(jobs, func(i, j int) bool {
		errorRateI := float64(jobs[i].FailedRuns) / float64(jobs[i].TotalExecutions)
		errorRateJ := float64(jobs[j].FailedRuns) / float64(jobs[j].TotalExecutions)
		return errorRateI > errorRateJ
	})

	if limit > 0 && len(jobs) > limit {
		jobs = jobs[:limit]
	}

	return jobs
}

// calculatePeakThroughputForDay calculates the peak throughput for a given day from hourly stats
func (sa *StatsAggregator) calculatePeakThroughputForDay(day time.Time) float64 {
	var peak float64

	for _, hourStats := range sa.hourlyStats {
		if hourStats.Timestamp.Truncate(24 * time.Hour).Equal(day) {
			if hourStats.ThroughputPerMin > peak {
				peak = hourStats.ThroughputPerMin
			}
		}
	}

	return peak
}

// CleanupOldData removes data older than the retention period
func (sa *StatsAggregator) CleanupOldData() {
	sa.mu.Lock()
	defer sa.mu.Unlock()

	// This is automatically handled by the maxHours and maxDays limits
	// in UpdateHourlyStats and UpdateDailyStats methods

	// Additional cleanup for job executions older than 7 days
	cutoff := time.Now().AddDate(0, 0, -7)
	for _, jobStats := range sa.jobStats {
		var recentExecutions []JobExecutionHistory
		for _, exec := range jobStats.RecentExecutions {
			if exec.Timestamp.After(cutoff) {
				recentExecutions = append(recentExecutions, exec)
			}
		}
		jobStats.RecentExecutions = recentExecutions
	}
}
