package parser

import (
	"errors"
	"strings"
	"time"
)

// Error definitions for the parser package
var (
	ErrInvalidExpression = errors.New("invalid cron expression")
	ErrValueOutOfRange   = errors.New("value out of range")
)

// Validate validates a cron expression string
func Validate(expr string) error {
	_, err := Parse(expr)
	return err
}

// ValidateInLocation validates a cron expression string in a specific timezone
func ValidateInLocation(expr string, loc *time.Location) error {
	_, err := ParseInLocation(expr, loc)
	return err
}

// IsValid checks if a cron expression is valid
func IsValid(expr string) bool {
	return Validate(expr) == nil
}

// NormalizeExpression normalizes a cron expression by removing extra whitespace
func NormalizeExpression(expr string) string {
	fields := strings.Fields(expr)
	return strings.Join(fields, " ")
}

// GetNextExecutions returns the next n execution times for a cron expression
func GetNextExecutions(expr string, after time.Time, count int) ([]time.Time, error) {
	schedule, err := ParseInLocation(expr, after.Location())
	if err != nil {
		return nil, err
	}

	var executions []time.Time
	current := after
	for i := 0; i < count; i++ {
		next := schedule.Next(current)
		if next.IsZero() {
			break
		}
		executions = append(executions, next)
		current = next
	}

	return executions, nil
}

// GetExecutionsBetween returns all execution times between start and end for a cron expression
func GetExecutionsBetween(expr string, start, end time.Time) ([]time.Time, error) {
	schedule, err := Parse(expr)
	if err != nil {
		return nil, err
	}

	var executions []time.Time
	current := start.Add(-time.Minute) // Start one minute before to include start time if it matches

	for {
		next := schedule.Next(current)
		if next.IsZero() || next.After(end) {
			break
		}
		if !next.Before(start) {
			executions = append(executions, next)
		}
		current = next

		// Safety check to prevent infinite loops
		if len(executions) > 10000 {
			break
		}
	}

	return executions, nil
}
