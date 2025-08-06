package types

import (
	"testing"
	"time"
)

func TestDefaultJobConfig(t *testing.T) {
	config := DefaultJobConfig()

	if config.MaxRetries != 0 {
		t.Errorf("Expected MaxRetries to be 0, got %d", config.MaxRetries)
	}

	if config.RetryInterval != time.Minute {
		t.Errorf("Expected RetryInterval to be %v, got %v", time.Minute, config.RetryInterval)
	}

	if config.Timeout != 0 {
		t.Errorf("Expected Timeout to be 0, got %v", config.Timeout)
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.MaxConcurrentJobs != 100 {
		t.Errorf("Expected MaxConcurrentJobs to be 100, got %d", config.MaxConcurrentJobs)
	}

	if config.EnableMonitoring != false {
		t.Errorf("Expected EnableMonitoring to be false, got %v", config.EnableMonitoring)
	}

	if config.MonitoringPort != 8080 {
		t.Errorf("Expected MonitoringPort to be 8080, got %d", config.MonitoringPort)
	}

	if config.Logger == nil {
		t.Error("Expected Logger to be set")
	}

	if config.Timezone == nil {
		t.Error("Expected Timezone to be set")
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name   string
		config Config
		valid  bool
	}{
		{
			name:   "valid config",
			config: DefaultConfig(),
			valid:  true,
		},
		{
			name: "invalid max concurrent jobs",
			config: Config{
				MaxConcurrentJobs: 0,
			},
			valid: false,
		},
		{
			name: "invalid monitoring port when monitoring enabled",
			config: Config{
				MaxConcurrentJobs: 10,
				EnableMonitoring:  true,
				MonitoringPort:    -1,
			},
			valid: false,
		},
		{
			name: "valid monitoring port when monitoring disabled",
			config: Config{
				MaxConcurrentJobs: 10,
				EnableMonitoring:  false,
				MonitoringPort:    -1,
			},
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set required fields if not set
			if tt.config.Logger == nil {
				tt.config.Logger = DefaultConfig().Logger
			}
			if tt.config.Timezone == nil {
				tt.config.Timezone = DefaultConfig().Timezone
			}

			err := tt.config.Validate()
			if tt.valid && err != nil {
				t.Errorf("Expected config to be valid, got error: %v", err)
			}
			if !tt.valid && err == nil {
				t.Error("Expected config to be invalid, got no error")
			}
		})
	}
}

func TestJobConfigValidation(t *testing.T) {
	tests := []struct {
		name   string
		config JobConfig
		valid  bool
	}{
		{
			name:   "valid config",
			config: DefaultJobConfig(),
			valid:  true,
		},
		{
			name: "negative max retries",
			config: JobConfig{
				MaxRetries: -1,
			},
			valid: false,
		},
		{
			name: "negative retry interval",
			config: JobConfig{
				RetryInterval: -time.Second,
			},
			valid: false,
		},
		{
			name: "negative timeout",
			config: JobConfig{
				Timeout: -time.Second,
			},
			valid: false,
		},
		{
			name: "all valid values",
			config: JobConfig{
				MaxRetries:    3,
				RetryInterval: time.Minute,
				Timeout:       time.Minute * 5,
			},
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.valid && err != nil {
				t.Errorf("Expected config to be valid, got error: %v", err)
			}
			if !tt.valid && err == nil {
				t.Error("Expected config to be invalid, got no error")
			}
		})
	}
}

func TestJobStatusConstants(t *testing.T) {
	statuses := []string{
		JobStatusPending,
		JobStatusRunning,
		JobStatusCompleted,
		JobStatusFailed,
		JobStatusCanceled,
	}

	expected := []string{
		"pending",
		"running",
		"completed",
		"failed",
		"canceled",
	}

	for i, status := range statuses {
		if status != expected[i] {
			t.Errorf("Expected status %d to be %q, got %q", i, expected[i], status)
		}
	}
}

func TestErrorConstants(t *testing.T) {
	errors := []error{
		ErrInvalidSchedule,
		ErrJobNotFound,
		ErrJobAlreadyExists,
		ErrSchedulerNotStarted,
		ErrSchedulerAlreadyStarted,
		ErrSchedulerStopped,
		ErrJobTimeout,
		ErrMaxRetriesExceeded,
		ErrInvalidConfiguration,
	}

	// Just test that they are not nil and have sensible messages
	for i, err := range errors {
		if err == nil {
			t.Errorf("Error %d should not be nil", i)
		}
		if err.Error() == "" {
			t.Errorf("Error %d should have a non-empty message", i)
		}
	}
}
