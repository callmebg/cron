package cron

import (
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.Logger == nil {
		t.Error("DefaultConfig Logger should not be nil")
	}

	if config.Timezone == nil {
		t.Error("DefaultConfig Timezone should not be nil")
	}

	if config.MaxConcurrentJobs <= 0 {
		t.Errorf("DefaultConfig MaxConcurrentJobs = %d; want > 0", config.MaxConcurrentJobs)
	}

	if config.MonitoringPort <= 0 || config.MonitoringPort > 65535 {
		t.Errorf("DefaultConfig MonitoringPort = %d; want valid port", config.MonitoringPort)
	}
}

func TestDefaultJobConfig(t *testing.T) {
	config := DefaultJobConfig()

	if config.MaxRetries < 0 {
		t.Errorf("DefaultJobConfig MaxRetries = %d; want >= 0", config.MaxRetries)
	}

	if config.RetryInterval < 0 {
		t.Errorf("DefaultJobConfig RetryInterval = %v; want >= 0", config.RetryInterval)
	}

	if config.Timeout < 0 {
		t.Errorf("DefaultJobConfig Timeout = %v; want >= 0", config.Timeout)
	}
}

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name:    "valid default config",
			config:  DefaultConfig(),
			wantErr: false,
		},
		{
			name: "nil logger",
			config: Config{
				Logger:            nil,
				Timezone:          time.Local,
				MaxConcurrentJobs: 10,
				EnableMonitoring:  false,
				MonitoringPort:    8080,
			},
			wantErr: true,
		},
		{
			name: "nil timezone",
			config: Config{
				Logger:            DefaultConfig().Logger,
				Timezone:          nil,
				MaxConcurrentJobs: 10,
				EnableMonitoring:  false,
				MonitoringPort:    8080,
			},
			wantErr: true,
		},
		{
			name: "zero max concurrent jobs",
			config: Config{
				Logger:            DefaultConfig().Logger,
				Timezone:          time.Local,
				MaxConcurrentJobs: 0,
				EnableMonitoring:  false,
				MonitoringPort:    8080,
			},
			wantErr: true,
		},
		{
			name: "negative max concurrent jobs",
			config: Config{
				Logger:            DefaultConfig().Logger,
				Timezone:          time.Local,
				MaxConcurrentJobs: -1,
				EnableMonitoring:  false,
				MonitoringPort:    8080,
			},
			wantErr: true,
		},
		{
			name: "invalid monitoring port - too low",
			config: Config{
				Logger:            DefaultConfig().Logger,
				Timezone:          time.Local,
				MaxConcurrentJobs: 10,
				EnableMonitoring:  true,
				MonitoringPort:    0,
			},
			wantErr: true,
		},
		{
			name: "invalid monitoring port - too high",
			config: Config{
				Logger:            DefaultConfig().Logger,
				Timezone:          time.Local,
				MaxConcurrentJobs: 10,
				EnableMonitoring:  true,
				MonitoringPort:    65536,
			},
			wantErr: true,
		},
		{
			name: "monitoring disabled - port doesn't matter",
			config: Config{
				Logger:            DefaultConfig().Logger,
				Timezone:          time.Local,
				MaxConcurrentJobs: 10,
				EnableMonitoring:  false,
				MonitoringPort:    0, // Invalid port but monitoring is disabled
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestJobConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  JobConfig
		wantErr bool
	}{
		{
			name:    "valid default config",
			config:  DefaultJobConfig(),
			wantErr: false,
		},
		{
			name: "negative max retries",
			config: JobConfig{
				MaxRetries:    -1,
				RetryInterval: time.Minute,
				Timeout:       time.Hour,
			},
			wantErr: true,
		},
		{
			name: "negative retry interval",
			config: JobConfig{
				MaxRetries:    3,
				RetryInterval: -time.Minute,
				Timeout:       time.Hour,
			},
			wantErr: true,
		},
		{
			name: "negative timeout",
			config: JobConfig{
				MaxRetries:    3,
				RetryInterval: time.Minute,
				Timeout:       -time.Hour,
			},
			wantErr: true,
		},
		{
			name: "zero values are valid",
			config: JobConfig{
				MaxRetries:    0,
				RetryInterval: 0,
				Timeout:       0,
			},
			wantErr: false,
		},
		{
			name: "large valid values",
			config: JobConfig{
				MaxRetries:    100,
				RetryInterval: time.Hour * 24,
				Timeout:       time.Hour * 24 * 7,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("JobConfig.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestJobStatusConstants(t *testing.T) {
	expectedStatuses := []string{
		JobStatusPending,
		JobStatusRunning,
		JobStatusCompleted,
		JobStatusFailed,
		JobStatusCanceled,
	}

	for _, status := range expectedStatuses {
		if status == "" {
			t.Errorf("Job status constant should not be empty")
		}
	}

	// Test that constants are unique
	statusMap := make(map[string]bool)
	for _, status := range expectedStatuses {
		if statusMap[status] {
			t.Errorf("Duplicate job status: %s", status)
		}
		statusMap[status] = true
	}
}
