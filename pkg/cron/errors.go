package cron

// Re-export errors from internal/types for public API
import "github.com/callmebg/cron/internal/types"

// Error definitions for the cron package
var (
	// ErrInvalidSchedule is returned when a cron schedule expression is invalid
	ErrInvalidSchedule = types.ErrInvalidSchedule

	// ErrJobNotFound is returned when trying to access a non-existent job
	ErrJobNotFound = types.ErrJobNotFound

	// ErrJobAlreadyExists is returned when trying to add a job with a name that already exists
	ErrJobAlreadyExists = types.ErrJobAlreadyExists

	// ErrSchedulerNotStarted is returned when trying to perform operations on a stopped scheduler
	ErrSchedulerNotStarted = types.ErrSchedulerNotStarted

	// ErrSchedulerAlreadyStarted is returned when trying to start an already running scheduler
	ErrSchedulerAlreadyStarted = types.ErrSchedulerAlreadyStarted

	// ErrSchedulerStopped is returned when the scheduler has been stopped
	ErrSchedulerStopped = types.ErrSchedulerStopped

	// ErrJobTimeout is returned when a job exceeds its configured timeout
	ErrJobTimeout = types.ErrJobTimeout

	// ErrMaxRetriesExceeded is returned when a job fails after all retry attempts
	ErrMaxRetriesExceeded = types.ErrMaxRetriesExceeded

	// ErrInvalidConfiguration is returned when the scheduler configuration is invalid
	ErrInvalidConfiguration = types.ErrInvalidConfiguration
)
