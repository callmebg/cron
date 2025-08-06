package monitor

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"time"

	"github.com/callmebg/cron/internal/types"
)

// HTTP server configuration constants
const (
	defaultReadTimeout   = 10 * time.Second
	defaultWriteTimeout  = 10 * time.Second
	defaultIdleTimeout   = 60 * time.Second
	defaultScheduleCount = 10
	nextExecutionsCount  = 5
	maxScheduleCount     = 100
	decimalBase          = 10
)

// HTTPMonitor provides HTTP endpoints for monitoring the scheduler
type HTTPMonitor struct {
	metrics   *Metrics
	scheduler SchedulerInterface
	server    *http.Server
}

// SchedulerInterface defines the interface needed for monitoring
type SchedulerInterface interface {
	GetStats() types.Stats
	GetJobStats(name string) (types.JobStats, error)
	ListJobs() []string
	IsRunning() bool
	GetNextExecutions(count int) map[string][]time.Time
}

// NewHTTPMonitor creates a new HTTP monitor
func NewHTTPMonitor(metrics *Metrics, scheduler SchedulerInterface, port int) *HTTPMonitor {
	monitor := &HTTPMonitor{
		metrics:   metrics,
		scheduler: scheduler,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/metrics", monitor.handleMetrics)
	mux.HandleFunc("/health", monitor.handleHealth)
	mux.HandleFunc("/jobs", monitor.handleJobs)
	mux.HandleFunc("/jobs/", monitor.handleJobDetail)
	mux.HandleFunc("/schedule", monitor.handleSchedule)
	mux.HandleFunc("/", monitor.handleIndex)

	monitor.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      mux,
		ReadTimeout:  defaultReadTimeout,
		WriteTimeout: defaultWriteTimeout,
		IdleTimeout:  defaultIdleTimeout,
	}

	return monitor
}

// Start starts the HTTP monitoring server
func (h *HTTPMonitor) Start() error {
	go func() {
		if err := h.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			// Log error but don't panic - monitoring is optional
			fmt.Printf("HTTP monitor server error: %v\n", err)
		}
	}()
	return nil
}

// Stop stops the HTTP monitoring server
func (h *HTTPMonitor) Stop() error {
	if h.server != nil {
		return h.server.Close()
	}
	return nil
}

// handleMetrics returns comprehensive metrics in JSON format
func (h *HTTPMonitor) handleMetrics(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	snapshot := h.metrics.GetSnapshot()
	schedulerStats := h.scheduler.GetStats()

	response := map[string]interface{}{
		"timestamp": time.Now().Format(time.RFC3339),
		"scheduler": map[string]interface{}{
			"running":               h.scheduler.IsRunning(),
			"uptime":                snapshot.Uptime.String(),
			"total_jobs":            snapshot.TotalJobs,
			"running_jobs":          snapshot.RunningJobs,
			"completed_jobs":        snapshot.CompletedJobs,
			"failed_jobs":           snapshot.FailedJobs,
			"successful_executions": schedulerStats.SuccessfulExecutions,
			"failed_executions":     schedulerStats.FailedExecutions,
		},
		"performance": map[string]interface{}{
			"jobs_per_minute":        snapshot.JobsPerMinute,
			"error_rate_percent":     snapshot.ErrorRate,
			"average_execution_time": snapshot.AverageExecutionTime.String(),
			"min_execution_time":     snapshot.MinExecutionTime.String(),
			"max_execution_time":     snapshot.MaxExecutionTime.String(),
			"last_job_execution":     snapshot.LastJobExecution.Format(time.RFC3339),
		},
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// handleHealth returns a simple health check
func (h *HTTPMonitor) handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	status := "ok"
	if !h.scheduler.IsRunning() {
		status = "stopped"
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	response := map[string]interface{}{
		"status":    status,
		"timestamp": time.Now().Format(time.RFC3339),
		"uptime":    h.metrics.GetSnapshot().Uptime.String(),
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// handleJobs returns information about all jobs
func (h *HTTPMonitor) handleJobs(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	jobNames := h.scheduler.ListJobs()
	sort.Strings(jobNames) // Sort for consistent output

	var jobs []map[string]interface{}
	for _, name := range jobNames {
		jobStats, err := h.scheduler.GetJobStats(name)
		if err != nil {
			continue // Skip jobs that can't be found
		}

		jobInfo := map[string]interface{}{
			"name":           jobStats.Name,
			"schedule":       jobStats.Schedule,
			"status":         jobStats.Status,
			"last_execution": formatTime(jobStats.LastExecution),
			"next_execution": formatTime(jobStats.NextExecution),
			"execution_time": jobStats.ExecutionTime.String(),
			"success_rate":   jobStats.SuccessRate,
			"total_runs":     jobStats.TotalRuns,
		}
		jobs = append(jobs, jobInfo)
	}

	response := map[string]interface{}{
		"timestamp":  time.Now().Format(time.RFC3339),
		"total_jobs": len(jobs),
		"jobs":       jobs,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// handleJobDetail returns detailed information about a specific job
func (h *HTTPMonitor) handleJobDetail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extract job name from URL path
	jobName := r.URL.Path[len("/jobs/"):]
	if jobName == "" {
		http.Error(w, "Job name required", http.StatusBadRequest)
		return
	}

	jobStats, err := h.scheduler.GetJobStats(jobName)
	if err != nil {
		http.Error(w, "Job not found", http.StatusNotFound)
		return
	}

	// Get next executions
	nextExecutions := h.scheduler.GetNextExecutions(nextExecutionsCount)
	jobNextExecutions := nextExecutions[jobName]

	var nextExecStrings []string
	for _, t := range jobNextExecutions {
		nextExecStrings = append(nextExecStrings, t.Format(time.RFC3339))
	}

	response := map[string]interface{}{
		"timestamp":         time.Now().Format(time.RFC3339),
		"name":              jobStats.Name,
		"schedule":          jobStats.Schedule,
		"status":            jobStats.Status,
		"last_execution":    formatTime(jobStats.LastExecution),
		"next_execution":    formatTime(jobStats.NextExecution),
		"execution_time":    jobStats.ExecutionTime.String(),
		"success_rate":      jobStats.SuccessRate,
		"total_runs":        jobStats.TotalRuns,
		"next_5_executions": nextExecStrings,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// handleSchedule returns upcoming executions for all jobs
func (h *HTTPMonitor) handleSchedule(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	count := defaultScheduleCount // Default number of upcoming executions
	if countStr := r.URL.Query().Get("count"); countStr != "" {
		if parsedCount := parseInt(countStr, decimalBase); parsedCount > 0 && parsedCount <= maxScheduleCount {
			count = parsedCount
		}
	}

	nextExecutions := h.scheduler.GetNextExecutions(count)

	// Convert to a more readable format
	schedule := make(map[string][]string)
	for jobName, executions := range nextExecutions {
		var execStrings []string
		for _, t := range executions {
			execStrings = append(execStrings, t.Format(time.RFC3339))
		}
		schedule[jobName] = execStrings
	}

	response := map[string]interface{}{
		"timestamp": time.Now().Format(time.RFC3339),
		"count":     count,
		"schedule":  schedule,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// handleIndex returns a simple HTML interface
func (h *HTTPMonitor) handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	html := `
<!DOCTYPE html>
<html>
<head>
    <title>Cron Scheduler Monitor</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .endpoint { margin: 10px 0; padding: 10px; background: #f5f5f5; border-radius: 4px; }
        .method { color: #007acc; font-weight: bold; }
    </style>
</head>
<body>
    <h1>Cron Scheduler Monitor</h1>
    <p>Available endpoints:</p>
    
    <div class="endpoint">
        <span class="method">GET</span> <a href="/health">/health</a> - Health check
    </div>
    
    <div class="endpoint">
        <span class="method">GET</span> <a href="/metrics">/metrics</a> - Complete metrics
    </div>
    
    <div class="endpoint">
        <span class="method">GET</span> <a href="/jobs">/jobs</a> - All jobs information
    </div>
    
    <div class="endpoint">
        <span class="method">GET</span> /jobs/{name} - Specific job details
    </div>
    
    <div class="endpoint">
        <span class="method">GET</span> <a href="/schedule?count=5">/schedule</a> - Upcoming executions
    </div>
</body>
</html>`

	if _, err := w.Write([]byte(html)); err != nil {
		// Log error but don't return error response since headers already sent
		fmt.Printf("Failed to write HTML response: %v\n", err)
	}
}

// Helper functions

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339)
}

func parseInt(s string, defaultValue int) int {
	if s == "" {
		return defaultValue
	}
	// Simple integer parsing - in production you'd use strconv.Atoi
	var result int
	for _, c := range s {
		if c >= '0' && c <= '9' {
			result = result*decimalBase + int(c-'0')
		} else {
			return defaultValue
		}
	}
	return result
}
