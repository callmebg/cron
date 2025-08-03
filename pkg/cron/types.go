package cron

// Re-export types from internal/types for public API
import "github.com/callmebg/cron/internal/types"

// Job function types
type JobFunc = types.JobFunc
type JobFuncWithError = types.JobFuncWithError
type ErrorHandler = types.ErrorHandler

// Configuration types
type JobConfig = types.JobConfig
type Config = types.Config

// Statistics types
type Stats = types.Stats
type JobStats = types.JobStats

// Job status constants
const (
	JobStatusPending   = types.JobStatusPending
	JobStatusRunning   = types.JobStatusRunning
	JobStatusCompleted = types.JobStatusCompleted
	JobStatusFailed    = types.JobStatusFailed
	JobStatusCancelled = types.JobStatusCancelled
)
