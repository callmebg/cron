package parser

import (
	"strings"
	"testing"
	"time"
)

func TestParseValidExpressions(t *testing.T) {
	tests := []struct {
		name       string
		expression string
		expected   bool
	}{
		// 5-field format (legacy)
		{"every minute", "* * * * *", true},
		{"every hour", "0 * * * *", true},
		{"every day at midnight", "0 0 * * *", true},
		{"every Monday at 9 AM", "0 9 * * 1", true},
		{"every 5 minutes", "*/5 * * * *", true},
		{"range of hours", "0 9-17 * * *", true},
		{"specific minutes", "0,15,30,45 * * * *", true},
		{"complex expression", "15 2,14 1 * 1-5", true},

		// 6-field format (with seconds)
		{"every second", "* * * * * *", true},
		{"every 30 seconds", "*/30 * * * * *", true},
		{"every minute at 0 seconds", "0 * * * * *", true},
		{"every hour at 0:0", "0 0 * * * *", true},
		{"every day at midnight", "0 0 0 * * *", true},
		{"every Monday at 9:00:00 AM", "0 0 9 * * 1", true},
		{"every 5 seconds", "*/5 * * * * *", true},
		{"specific seconds", "0,15,30,45 * * * * *", true},
		{"complex 6-field expression", "30 15 2,14 1 * 1-5", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schedule, err := Parse(tt.expression)
			if tt.expected {
				if err != nil {
					t.Errorf("Parse(%q) returned error: %v", tt.expression, err)
				}
				if schedule == nil {
					t.Errorf("Parse(%q) returned nil schedule", tt.expression)
				}
			} else if err == nil {
				t.Errorf("Parse(%q) should have returned error", tt.expression)
			}
		})
	}
}

func TestParseInvalidExpressions(t *testing.T) {
	tests := []struct {
		name       string
		expression string
	}{
		{"too few fields", "* * *"},
		{"too many fields", "* * * * * * *"},
		{"invalid minute in 5-field", "60 * * * *"},
		{"invalid hour in 5-field", "0 24 * * *"},
		{"invalid day in 5-field", "0 0 32 * *"},
		{"invalid month in 5-field", "0 0 1 13 *"},
		{"invalid weekday in 5-field", "0 0 * * 7"},
		{"invalid second in 6-field", "60 * * * * *"},
		{"invalid minute in 6-field", "0 60 * * * *"},
		{"invalid hour in 6-field", "0 0 24 * * *"},
		{"invalid day in 6-field", "0 0 0 32 * *"},
		{"invalid month in 6-field", "0 0 0 1 13 *"},
		{"invalid weekday in 6-field", "0 0 0 * * 7"},
		{"invalid range", "0 25-30 * * *"},
		{"invalid step", "*/0 * * * *"},
		{"empty expression", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schedule, err := Parse(tt.expression)
			if err == nil {
				t.Errorf("Parse(%q) should have returned error", tt.expression)
			}
			if schedule != nil {
				t.Errorf("Parse(%q) should have returned nil schedule", tt.expression)
			}
		})
	}
}

func TestScheduleNext(t *testing.T) {
	// Test "every minute" expression (5-field)
	schedule, err := Parse("* * * * *")
	if err != nil {
		t.Fatalf("Failed to parse expression: %v", err)
	}

	now := time.Date(2024, 1, 1, 12, 30, 0, 0, time.UTC)
	next := schedule.Next(now)
	expected := time.Date(2024, 1, 1, 12, 31, 0, 0, time.UTC)

	if !next.Equal(expected) {
		t.Errorf("Next(%v) = %v; want %v", now, next, expected)
	}
}

func TestScheduleNextWithSeconds(t *testing.T) {
	// Test "every second" expression (6-field)
	schedule, err := Parse("* * * * * *")
	if err != nil {
		t.Fatalf("Failed to parse expression: %v", err)
	}

	now := time.Date(2024, 1, 1, 12, 30, 30, 0, time.UTC)
	next := schedule.Next(now)
	expected := time.Date(2024, 1, 1, 12, 30, 31, 0, time.UTC)

	if !next.Equal(expected) {
		t.Errorf("Next(%v) = %v; want %v", now, next, expected)
	}
}

func TestScheduleNextEvery30Seconds(t *testing.T) {
	// Test "every 30 seconds" expression
	schedule, err := Parse("*/30 * * * * *")
	if err != nil {
		t.Fatalf("Failed to parse expression: %v", err)
	}

	now := time.Date(2024, 1, 1, 12, 30, 15, 0, time.UTC)
	next := schedule.Next(now)
	expected := time.Date(2024, 1, 1, 12, 30, 30, 0, time.UTC)

	if !next.Equal(expected) {
		t.Errorf("Next(%v) = %v; want %v", now, next, expected)
	}

	// Test from 30 seconds - should go to next minute
	now = time.Date(2024, 1, 1, 12, 30, 30, 0, time.UTC)
	next = schedule.Next(now)
	expected = time.Date(2024, 1, 1, 12, 31, 0, 0, time.UTC)

	if !next.Equal(expected) {
		t.Errorf("Next(%v) = %v; want %v", now, next, expected)
	}
}

func TestScheduleNextSpecificTime(t *testing.T) {
	// Test "every day at 9 AM" expression
	schedule, err := ParseInLocation("0 9 * * *", time.UTC)
	if err != nil {
		t.Fatalf("Failed to parse expression: %v", err)
	}

	// Test from 8 AM - should give 9 AM same day
	now := time.Date(2024, 1, 1, 8, 0, 0, 0, time.UTC)
	next := schedule.Next(now)
	expected := time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC)

	if !next.Equal(expected) {
		t.Errorf("Next(%v) = %v; want %v", now, next, expected)
	}

	// Test from 10 AM - should give 9 AM next day
	now = time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)
	next = schedule.Next(now)
	expected = time.Date(2024, 1, 2, 9, 0, 0, 0, time.UTC)

	if !next.Equal(expected) {
		t.Errorf("Next(%v) = %v; want %v", now, next, expected)
	}
}

func TestScheduleNextWeekly(t *testing.T) {
	// Test "every Monday at 9 AM" expression
	schedule, err := ParseInLocation("0 9 * * 1", time.UTC)
	if err != nil {
		t.Fatalf("Failed to parse expression: %v", err)
	}

	// Start on a Friday
	now := time.Date(2024, 1, 5, 12, 0, 0, 0, time.UTC) // Friday
	next := schedule.Next(now)
	expected := time.Date(2024, 1, 8, 9, 0, 0, 0, time.UTC) // Next Monday

	if !next.Equal(expected) {
		t.Errorf("Next(%v) = %v; want %v", now, next, expected)
	}
}

func TestParseFieldParts(t *testing.T) {
	tests := []struct {
		name     string
		field    string
		min      int
		max      int
		expected []int
	}{
		{"wildcard", "*", 0, 59, []int{0, 1, 2, 3, 4, 5}}, // First few values
		{"single value", "5", 0, 59, []int{5}},
		{"range", "1-5", 0, 59, []int{1, 2, 3, 4, 5}},
		{"step from wildcard", "*/10", 0, 59, []int{0, 10, 20, 30, 40, 50}},
		{"step from range", "10-20/5", 0, 59, []int{10, 15, 20}},
		{"comma separated", "1,3,5", 0, 59, []int{1, 3, 5}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseField(tt.field, tt.min, tt.max)
			if err != nil {
				t.Errorf("parseField(%q, %d, %d) returned error: %v", tt.field, tt.min, tt.max, err)
				return
			}

			// For wildcard test, only check first few values
			if tt.field == "*" {
				if len(result) != 60 { // 0-59
					t.Errorf("parseField(%q, %d, %d) returned %d values; want 60", tt.field, tt.min, tt.max, len(result))
					return
				}
				// Check first few values
				for i, expected := range tt.expected {
					if result[i] != expected {
						t.Errorf("parseField(%q, %d, %d)[%d] = %d; want %d", tt.field, tt.min, tt.max, i, result[i], expected)
					}
				}
				return
			}

			if len(result) != len(tt.expected) {
				t.Errorf("parseField(%q, %d, %d) returned %d values; want %d", tt.field, tt.min, tt.max, len(result), len(tt.expected))
				return
			}

			for i, expected := range tt.expected {
				if result[i] != expected {
					t.Errorf("parseField(%q, %d, %d)[%d] = %d; want %d", tt.field, tt.min, tt.max, i, result[i], expected)
				}
			}
		})
	}
}

func TestValidateExpression(t *testing.T) {
	tests := []struct {
		name       string
		expression string
		valid      bool
	}{
		{"valid simple", "0 0 * * *", true},
		{"valid complex", "*/15 9-17 * * 1-5", true},
		{"invalid fields", "* * *", false},
		{"invalid minute", "60 * * * *", false},
		{"invalid hour", "* 24 * * *", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.expression)
			if tt.valid && err != nil {
				t.Errorf("Validate(%q) returned error: %v", tt.expression, err)
			}
			if !tt.valid && err == nil {
				t.Errorf("Validate(%q) should have returned error", tt.expression)
			}
		})
	}
}

func TestIsValid(t *testing.T) {
	tests := []struct {
		expression string
		expected   bool
	}{
		{"* * * * *", true},
		{"0 0 * * *", true},
		{"invalid", false},
		{"* * *", false},
	}

	for _, tt := range tests {
		result := IsValid(tt.expression)
		if result != tt.expected {
			t.Errorf("IsValid(%q) = %v; want %v", tt.expression, result, tt.expected)
		}
	}
}

func TestGetNextExecutions(t *testing.T) {
	expression := "0 9 * * *" // Every day at 9 AM
	after := time.Date(2024, 1, 1, 8, 0, 0, 0, time.UTC)
	count := 3

	executions, err := GetNextExecutions(expression, after, count)
	if err != nil {
		t.Fatalf("GetNextExecutions returned error: %v", err)
	}

	if len(executions) != count {
		t.Errorf("GetNextExecutions returned %d executions; want %d", len(executions), count)
	}

	// Check that executions are in order and correct
	expected := []time.Time{
		time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC),
		time.Date(2024, 1, 2, 9, 0, 0, 0, time.UTC),
		time.Date(2024, 1, 3, 9, 0, 0, 0, time.UTC),
	}

	for i, exec := range executions {
		if !exec.Equal(expected[i]) {
			t.Errorf("GetNextExecutions()[%d] = %v; want %v", i, exec, expected[i])
		}
	}
}

func TestNormalizeExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"  *   *   *   *   *  ", "* * * * *"},
		{"0 0 * * *", "0 0 * * *"},
		{"\t0\t0\t*\t*\t*\t", "0 0 * * *"},
		{"", ""},
	}

	for _, tt := range tests {
		result := NormalizeExpression(tt.input)
		if result != tt.expected {
			t.Errorf("NormalizeExpression(%q) = %q; want %q", tt.input, result, tt.expected)
		}
	}
}

func BenchmarkParse(b *testing.B) {
	expression := "*/5 9-17 * * 1-5"
	for i := 0; i < b.N; i++ {
		_, err := Parse(expression)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkScheduleNext(b *testing.B) {
	schedule, err := Parse("*/5 * * * *")
	if err != nil {
		b.Fatal(err)
	}

	now := time.Now()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		schedule.Next(now)
	}
}

func FuzzCronParsing(f *testing.F) {
	// Seed with valid expressions
	f.Add("* * * * *")
	f.Add("0 0 * * *")
	f.Add("*/5 * * * *")
	f.Add("0 9-17 * * 1-5")
	f.Add("15,30,45 * * * *")
	f.Add("0 0 1 * *")
	f.Add("0 0 * * 0")
	f.Add("@hourly")
	f.Add("@daily")
	f.Add("@weekly")
	f.Add("@monthly")
	f.Add("@yearly")

	f.Fuzz(func(t *testing.T, expression string) {
		// Attempt to parse the expression
		schedule, err := Parse(expression)

		if err != nil {
			// If parsing fails, that's fine for fuzz testing
			// Just ensure we don't crash
			return
		}

		if schedule == nil {
			t.Error("Parse returned nil schedule without error")
			return
		}

		// If parsing succeeds, test that Next() doesn't crash
		now := time.Now()
		next := schedule.Next(now)

		// Basic sanity check - next time should be after current time
		if next.Before(now) || next.Equal(now) {
			t.Errorf("Next time %v should be after current time %v", next, now)
		}

		// Test multiple Next() calls don't crash
		for i := 0; i < 5; i++ {
			now = next
			next = schedule.Next(now)
			if next.Before(now) || next.Equal(now) {
				t.Errorf("Next time %v should be after current time %v", next, now)
				break
			}
		}
	})
}

func FuzzParseField(f *testing.F) {
	// Seed with valid field values
	f.Add("*", 0, 59)
	f.Add("5", 0, 59)
	f.Add("1-10", 0, 59)
	f.Add("*/5", 0, 59)
	f.Add("1,3,5", 0, 59)
	f.Add("10-20/2", 0, 59)

	f.Fuzz(func(t *testing.T, field string, min, max int) {
		// Ensure min and max are reasonable
		if min < 0 || max < 0 || min >= max || max > 10000 {
			return
		}

		// Attempt to parse the field
		result, err := parseField(field, min, max)

		if err != nil {
			// If parsing fails, that's fine for fuzz testing
			return
		}

		// If parsing succeeds, validate the result
		if result == nil {
			t.Error("parseField returned nil result without error")
			return
		}

		// Check that all values are within bounds
		for _, value := range result {
			if value < min || value > max {
				t.Errorf("Value %d is outside bounds [%d, %d]", value, min, max)
			}
		}

		// Check that values are sorted and unique
		for i := 1; i < len(result); i++ {
			if result[i] <= result[i-1] {
				t.Errorf("Result is not sorted or contains duplicates: %v", result)
				break
			}
		}
	})
}

func FuzzNormalizeExpression(f *testing.F) {
	// Seed with various expressions
	f.Add("* * * * *")
	f.Add("  *   *   *   *   *  ")
	f.Add("\t0\t0\t*\t*\t*\t")
	f.Add("0 0 * * *")
	f.Add("")
	f.Add("   ")

	f.Fuzz(func(t *testing.T, input string) {
		// Normalize should never crash
		result := NormalizeExpression(input)

		// If input is empty or only whitespace, result should be empty
		if strings.TrimSpace(input) == "" {
			if result != "" {
				t.Errorf("NormalizeExpression(%q) = %q; expected empty string", input, result)
			}
			return
		}

		// Result should not contain leading/trailing whitespace
		if result != strings.TrimSpace(result) {
			t.Errorf("NormalizeExpression(%q) = %q; contains leading/trailing whitespace", input, result)
		}

		// Result should not contain multiple consecutive spaces
		if strings.Contains(result, "  ") {
			t.Errorf("NormalizeExpression(%q) = %q; contains multiple consecutive spaces", input, result)
		}
	})
}
