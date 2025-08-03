package cron

// Re-export configuration functions from internal/types for public API
import "github.com/callmebg/cron/internal/types"

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() Config {
	return types.DefaultConfig()
}

// DefaultJobConfig returns a job configuration with sensible defaults
func DefaultJobConfig() JobConfig {
	return types.DefaultJobConfig()
}
