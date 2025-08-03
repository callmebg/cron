package parser

import (
	"strconv"
	"strings"
	"time"
)

// Schedule represents a parsed cron schedule
type Schedule struct {
	Seconds  []int // 0-59
	Minutes  []int // 0-59
	Hours    []int // 0-23
	Days     []int // 1-31
	Months   []int // 1-12
	Weekdays []int // 0-6 (Sunday=0)
	Timezone *time.Location
}

// Next calculates the next execution time after the given time
func (s *Schedule) Next(after time.Time) time.Time {
	t := after.In(s.Timezone).Add(time.Second).Truncate(time.Second)

	// Find next valid time
	for i := 0; i < 4*365*24*60*60; i++ { // Limit iterations to prevent infinite loops
		if s.matches(t) {
			return t.In(after.Location())
		}
		t = t.Add(time.Second)
	}

	// Return zero time if no match found (should not happen with valid cron expressions)
	return time.Time{}
}

// matches checks if the given time matches the schedule
func (s *Schedule) matches(t time.Time) bool {
	return s.matchesField(t.Second(), s.Seconds) &&
		s.matchesField(t.Minute(), s.Minutes) &&
		s.matchesField(t.Hour(), s.Hours) &&
		s.matchesField(t.Day(), s.Days) &&
		s.matchesField(int(t.Month()), s.Months) &&
		s.matchesField(int(t.Weekday()), s.Weekdays)
}

// matchesField checks if a value matches any value in the field slice
func (s *Schedule) matchesField(value int, field []int) bool {
	for _, v := range field {
		if v == value {
			return true
		}
	}
	return false
}

// Parse parses a cron expression string into a Schedule
func Parse(expr string) (*Schedule, error) {
	return ParseInLocation(expr, time.Local)
}

// ParseInLocation parses a cron expression string in a specific timezone
func ParseInLocation(expr string, loc *time.Location) (*Schedule, error) {
	fields := strings.Fields(expr)
	// Support both 5-field (minute-based) and 6-field (second-based) formats
	if len(fields) != 5 && len(fields) != 6 {
		return nil, ErrInvalidExpression
	}

	schedule := &Schedule{
		Timezone: loc,
	}

	var err error
	// Handle 6-field format: seconds minutes hours days months weekdays
	if len(fields) == 6 {
		if schedule.Seconds, err = parseField(fields[0], 0, 59); err != nil {
			return nil, err
		}
		if schedule.Minutes, err = parseField(fields[1], 0, 59); err != nil {
			return nil, err
		}
		if schedule.Hours, err = parseField(fields[2], 0, 23); err != nil {
			return nil, err
		}
		if schedule.Days, err = parseField(fields[3], 1, 31); err != nil {
			return nil, err
		}
		if schedule.Months, err = parseField(fields[4], 1, 12); err != nil {
			return nil, err
		}
		if schedule.Weekdays, err = parseField(fields[5], 0, 6); err != nil {
			return nil, err
		}
	} else {
		// Handle 5-field format: minutes hours days months weekdays (legacy support)
		// Default seconds to 0 for backward compatibility
		schedule.Seconds = []int{0}
		if schedule.Minutes, err = parseField(fields[0], 0, 59); err != nil {
			return nil, err
		}
		if schedule.Hours, err = parseField(fields[1], 0, 23); err != nil {
			return nil, err
		}
		if schedule.Days, err = parseField(fields[2], 1, 31); err != nil {
			return nil, err
		}
		if schedule.Months, err = parseField(fields[3], 1, 12); err != nil {
			return nil, err
		}
		if schedule.Weekdays, err = parseField(fields[4], 0, 6); err != nil {
			return nil, err
		}
	}

	return schedule, nil
}

// parseField parses a single field of a cron expression
func parseField(field string, min, max int) ([]int, error) {
	var values []int

	// Handle comma-separated values
	parts := strings.Split(field, ",")
	for _, part := range parts {
		vals, err := parseFieldPart(part, min, max)
		if err != nil {
			return nil, err
		}
		values = append(values, vals...)
	}

	// Remove duplicates and sort
	return removeDuplicates(values), nil
}

// parseFieldPart parses a part of a field (handling *, ranges, and steps)
func parseFieldPart(part string, min, max int) ([]int, error) {
	// Handle step values (e.g., */5, 1-10/2)
	if strings.Contains(part, "/") {
		return parseStepValue(part, min, max)
	}

	// Handle ranges (e.g., 1-5)
	if strings.Contains(part, "-") {
		return parseRange(part, min, max)
	}

	// Handle wildcard
	if part == "*" {
		return generateRange(min, max), nil
	}

	// Handle single value
	value, err := strconv.Atoi(part)
	if err != nil {
		return nil, ErrInvalidExpression
	}
	if value < min || value > max {
		return nil, ErrValueOutOfRange
	}
	return []int{value}, nil
}

// parseStepValue parses step values like */5 or 1-10/2
func parseStepValue(part string, min, max int) ([]int, error) {
	stepParts := strings.Split(part, "/")
	if len(stepParts) != 2 {
		return nil, ErrInvalidExpression
	}

	step, err := strconv.Atoi(stepParts[1])
	if err != nil || step <= 0 {
		return nil, ErrInvalidExpression
	}

	var baseValues []int
	if stepParts[0] == "*" {
		baseValues = generateRange(min, max)
	} else {
		baseValues, err = parseFieldPart(stepParts[0], min, max)
		if err != nil {
			return nil, err
		}
	}

	var result []int
	for i, value := range baseValues {
		if i%step == 0 {
			result = append(result, value)
		}
	}

	return result, nil
}

// parseRange parses range values like 1-5
func parseRange(part string, min, max int) ([]int, error) {
	rangeParts := strings.Split(part, "-")
	if len(rangeParts) != 2 {
		return nil, ErrInvalidExpression
	}

	start, err := strconv.Atoi(rangeParts[0])
	if err != nil {
		return nil, ErrInvalidExpression
	}
	end, err := strconv.Atoi(rangeParts[1])
	if err != nil {
		return nil, ErrInvalidExpression
	}

	if start < min || start > max || end < min || end > max || start > end {
		return nil, ErrValueOutOfRange
	}

	var result []int
	for i := start; i <= end; i++ {
		result = append(result, i)
	}
	return result, nil
}

// generateRange generates a slice of integers from min to max (inclusive)
func generateRange(min, max int) []int {
	result := make([]int, max-min+1)
	for i := range result {
		result[i] = min + i
	}
	return result
}

// removeDuplicates removes duplicate values from a slice and returns a sorted slice
func removeDuplicates(values []int) []int {
	seen := make(map[int]bool)
	var result []int

	for _, value := range values {
		if !seen[value] {
			seen[value] = true
			result = append(result, value)
		}
	}

	// Simple insertion sort for small slices
	for i := 1; i < len(result); i++ {
		key := result[i]
		j := i - 1
		for j >= 0 && result[j] > key {
			result[j+1] = result[j]
			j--
		}
		result[j+1] = key
	}

	return result
}
