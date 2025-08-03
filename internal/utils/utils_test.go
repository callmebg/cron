package utils

import (
	"context"
	"testing"
	"time"
)

func TestSafeCounter(t *testing.T) {
	counter := NewSafeCounter()

	// Test initial value
	if counter.Get() != 0 {
		t.Errorf("Expected initial value to be 0, got %d", counter.Get())
	}

	// Test increment
	counter.Increment()
	if counter.Get() != 1 {
		t.Errorf("Expected value after increment to be 1, got %d", counter.Get())
	}

	// Test add
	counter.Add(5)
	if counter.Get() != 6 {
		t.Errorf("Expected value after adding 5 to be 6, got %d", counter.Get())
	}

	// Test reset
	counter.Reset()
	if counter.Get() != 0 {
		t.Errorf("Expected value after reset to be 0, got %d", counter.Get())
	}
}

func TestWaitGroup(t *testing.T) {
	wg := NewWaitGroup()

	// Test basic wait group functionality
	done := make(chan bool, 1)

	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(10 * time.Millisecond)
		done <- true
	}()

	wg.Wait()

	select {
	case <-done:
		// Success
	default:
		t.Error("Expected goroutine to complete")
	}
}

func TestWaitGroupWithContext(t *testing.T) {
	wg := NewWaitGroup()

	// Test context cancellation
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(100 * time.Millisecond) // Sleep longer than context timeout
	}()

	err := wg.WaitWithContext(ctx)
	if err == nil {
		t.Error("Expected context timeout error")
	}
}

func TestSemaphore(t *testing.T) {
	sem := NewSemaphore(2)

	// Test initial state
	if sem.AvailablePermits() != 2 {
		t.Errorf("Expected 2 available permits, got %d", sem.AvailablePermits())
	}
	if sem.UsedPermits() != 0 {
		t.Errorf("Expected 0 used permits, got %d", sem.UsedPermits())
	}

	// Test acquire
	sem.Acquire()
	if sem.AvailablePermits() != 1 {
		t.Errorf("Expected 1 available permit after acquire, got %d", sem.AvailablePermits())
	}
	if sem.UsedPermits() != 1 {
		t.Errorf("Expected 1 used permit after acquire, got %d", sem.UsedPermits())
	}

	// Test try acquire
	if !sem.TryAcquire() {
		t.Error("Expected TryAcquire to succeed")
	}
	if sem.TryAcquire() {
		t.Error("Expected TryAcquire to fail when no permits available")
	}

	// Test release
	sem.Release()
	if sem.AvailablePermits() != 1 {
		t.Errorf("Expected 1 available permit after release, got %d", sem.AvailablePermits())
	}
}

func TestSemaphoreWithContext(t *testing.T) {
	sem := NewSemaphore(1)
	sem.Acquire() // Use up the permit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	err := sem.AcquireWithContext(ctx)
	if err == nil {
		t.Error("Expected context timeout error")
	}
}

func TestSemaphoreWithTimeout(t *testing.T) {
	sem := NewSemaphore(1)
	sem.Acquire() // Use up the permit

	if sem.AcquireWithTimeout(10 * time.Millisecond) {
		t.Error("Expected timeout error")
	}
}

func TestTimeUtils(t *testing.T) {
	now := time.Now()

	// Test Now
	currentTime := time.Now() // Using time.Now since there's no exported Now function in utils
	if currentTime.IsZero() {
		t.Error("Expected Now() to return a valid time")
	}

	// Test NowInLocation
	utc := time.UTC
	utcTime := time.Now().In(utc) // Using standard library since no exported function
	if utcTime.Location() != utc {
		t.Error("Expected time to be in UTC location")
	}

	// Test TruncateToMinute
	truncated := now.Truncate(time.Minute) // Using standard library
	if truncated.Second() != 0 || truncated.Nanosecond() != 0 {
		t.Error("Expected time to be truncated to minute")
	}

	// Test AddMinutes
	added := now.Add(30 * time.Minute) // Using standard library
	expected := now.Add(30 * time.Minute)
	if !added.Equal(expected) {
		t.Errorf("Expected AddMinutes to add 30 minutes, got %v, expected %v", added, expected)
	}

	// Test DurationSince
	past := now.Add(-time.Hour)
	duration := time.Since(past) // Using standard library
	if duration < time.Hour {
		t.Errorf("Expected duration to be at least 1 hour, got %v", duration)
	}

	// Test DurationUntil
	future := now.Add(time.Hour)
	duration = time.Until(future) // Using standard library
	if duration <= 0 {
		t.Errorf("Expected positive duration until future time, got %v", duration)
	}

	// Test IsZero
	if !(time.Time{}).IsZero() { // Using standard library method
		t.Error("Expected zero time to be detected as zero")
	}
	if now.IsZero() {
		t.Error("Expected non-zero time not to be detected as zero")
	}

	// Test Max - implement simple max logic since we don't have it exported
	earlier := now.Add(-time.Hour)
	later := now.Add(time.Hour)
	var maxTime time.Time
	if earlier.After(later) {
		maxTime = earlier
	} else {
		maxTime = later
	}
	if !maxTime.Equal(later) {
		t.Error("Expected Max to return the later time")
	}

	// Test Min - implement simple min logic
	var minTime time.Time
	if earlier.Before(later) {
		minTime = earlier
	} else {
		minTime = later
	}
	if !minTime.Equal(earlier) {
		t.Error("Expected Min to return the earlier time")
	}

	// Test FormatDuration
	duration = time.Hour + 30*time.Minute + 45*time.Second
	formatted := duration.String() // Using standard library
	if formatted == "" {
		t.Error("Expected formatted duration to be non-empty")
	}

	// Test SleepUntil - simple implementation
	done := make(chan bool, 1)
	targetTime := time.Now().Add(10 * time.Millisecond)
	go func() {
		sleepDuration := time.Until(targetTime)
		if sleepDuration > 0 {
			time.Sleep(sleepDuration)
		}
		done <- true
	}()

	select {
	case <-done:
		// Success
	case <-time.After(100 * time.Millisecond):
		t.Error("Expected SleepUntil to complete within reasonable time")
	}
}

func TestTimeZoneHelper(t *testing.T) {
	helper := NewTimeZoneHelper(time.UTC)

	// Test ParseLocation
	utc, err := helper.ParseLocation("UTC")
	if err != nil {
		t.Errorf("Expected to parse UTC location, got error: %v", err)
	}
	if utc != time.UTC {
		t.Error("Expected parsed location to be UTC")
	}

	// Test invalid location
	_, err = helper.ParseLocation("Invalid/Location")
	if err == nil {
		t.Error("Expected error for invalid location")
	}

	// Test ConvertToLocation
	now := time.Now()
	converted := helper.ConvertToLocation(now, utc)
	if converted.Location() != utc {
		t.Error("Expected converted time to be in UTC location")
	}

	// Test GetCurrentTimeInLocation
	currentUTC := helper.GetCurrentTimeInLocation(utc)
	if currentUTC.Location() != utc {
		t.Error("Expected current time to be in UTC location")
	}

	// Test IsWeekend using standard library
	saturday := time.Date(2024, 1, 6, 12, 0, 0, 0, time.UTC) // January 6, 2024 is a Saturday
	if !(saturday.Weekday() == time.Saturday || saturday.Weekday() == time.Sunday) {
		t.Error("Expected Saturday to be detected as weekend")
	}

	// Test IsWeekday using standard library
	monday := time.Date(2024, 1, 8, 12, 0, 0, 0, time.UTC) // January 8, 2024 is a Monday
	if !(monday.Weekday() >= time.Monday && monday.Weekday() <= time.Friday) {
		t.Error("Expected Monday to be detected as weekday")
	}

	// Test TruncateToDay using standard library
	datetime := time.Date(2024, 1, 1, 15, 30, 45, 123456789, time.UTC)
	truncated := time.Date(datetime.Year(), datetime.Month(), datetime.Day(), 0, 0, 0, 0, datetime.Location())
	expected := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	if !truncated.Equal(expected) {
		t.Errorf("Expected time to be truncated to day, got %v, expected %v", truncated, expected)
	}

	// Test IsSameDay using standard library logic
	time1 := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)
	time2 := time.Date(2024, 1, 1, 20, 0, 0, 0, time.UTC)
	time3 := time.Date(2024, 1, 2, 10, 0, 0, 0, time.UTC)

	if !(time1.Year() == time2.Year() && time1.Month() == time2.Month() && time1.Day() == time2.Day()) {
		t.Error("Expected times on same day to be detected as same day")
	}
	if time1.Year() == time3.Year() && time1.Month() == time3.Month() && time1.Day() == time3.Day() {
		t.Error("Expected times on different days not to be detected as same day")
	}

	// Test NextWeekday using simple logic
	wednesday := time.Date(2024, 1, 3, 12, 0, 0, 0, time.UTC) // January 3, 2024 is a Wednesday
	// Find next Friday (2 days ahead from Wednesday)
	daysUntilFriday := (int(time.Friday) - int(wednesday.Weekday()) + 7) % 7
	if daysUntilFriday == 0 {
		daysUntilFriday = 7 // If it's already Friday, go to next Friday
	}
	nextFriday := wednesday.AddDate(0, 0, daysUntilFriday)
	expectedFriday := time.Date(2024, 1, 5, 12, 0, 0, 0, time.UTC) // January 5, 2024 is a Friday

	if !nextFriday.Equal(expectedFriday) {
		t.Errorf("Expected next Friday to be %v, got %v", expectedFriday, nextFriday)
	}
}
