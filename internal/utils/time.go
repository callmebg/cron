package utils

import (
	"time"
)

// TimeUtils provides utility functions for time operations
type TimeUtils struct{}

// Now returns the current time
func (tu *TimeUtils) Now() time.Time {
	return time.Now()
}

// NowInLocation returns the current time in the specified location
func (tu *TimeUtils) NowInLocation(loc *time.Location) time.Time {
	return time.Now().In(loc)
}

// TruncateToMinute truncates a time to the minute precision
func (tu *TimeUtils) TruncateToMinute(t time.Time) time.Time {
	return t.Truncate(time.Minute)
}

// AddMinutes adds the specified number of minutes to a time
func (tu *TimeUtils) AddMinutes(t time.Time, minutes int) time.Time {
	return t.Add(time.Duration(minutes) * time.Minute)
}

// DurationSince returns the duration since the given time
func (tu *TimeUtils) DurationSince(t time.Time) time.Duration {
	return time.Since(t)
}

// DurationUntil returns the duration until the given time
func (tu *TimeUtils) DurationUntil(t time.Time) time.Duration {
	return time.Until(t)
}

// IsZero checks if a time is the zero time
func (tu *TimeUtils) IsZero(t time.Time) bool {
	return t.IsZero()
}

// Max returns the later of two times
func (tu *TimeUtils) Max(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}

// Min returns the earlier of two times
func (tu *TimeUtils) Min(a, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}
	return b
}

// FormatDuration formats a duration in a human-readable way
func (tu *TimeUtils) FormatDuration(d time.Duration) string {
	if d < time.Second {
		return d.String()
	}
	if d < time.Minute {
		return d.Truncate(time.Millisecond).String()
	}
	if d < time.Hour {
		return d.Truncate(time.Second).String()
	}
	return d.Truncate(time.Minute).String()
}

// SleepUntil sleeps until the specified time
func (tu *TimeUtils) SleepUntil(t time.Time) {
	duration := time.Until(t)
	if duration > 0 {
		time.Sleep(duration)
	}
}

// NewTimeUtils creates a new TimeUtils instance
func NewTimeUtils() *TimeUtils {
	return &TimeUtils{}
}

// TimeZoneHelper provides utilities for working with timezones in cron schedules
type TimeZoneHelper struct {
	defaultLocation *time.Location
}

// NewTimeZoneHelper creates a new timezone helper
func NewTimeZoneHelper(defaultLocation *time.Location) *TimeZoneHelper {
	if defaultLocation == nil {
		defaultLocation = time.Local
	}
	return &TimeZoneHelper{
		defaultLocation: defaultLocation,
	}
}

// ParseLocation safely parses a timezone location string
func (tz *TimeZoneHelper) ParseLocation(locationName string) (*time.Location, error) {
	if locationName == "" {
		return tz.defaultLocation, nil
	}

	location, err := time.LoadLocation(locationName)
	if err != nil {
		return tz.defaultLocation, err
	}

	return location, nil
}

// ConvertToLocation converts a time from one timezone to another
func (tz *TimeZoneHelper) ConvertToLocation(t time.Time, location *time.Location) time.Time {
	if location == nil {
		location = tz.defaultLocation
	}
	return t.In(location)
}

// GetCurrentTimeInLocation returns the current time in the specified location
func (tz *TimeZoneHelper) GetCurrentTimeInLocation(location *time.Location) time.Time {
	if location == nil {
		location = tz.defaultLocation
	}
	return time.Now().In(location)
}

// IsWeekend checks if the given time falls on a weekend
func IsWeekend(t time.Time) bool {
	weekday := t.Weekday()
	return weekday == time.Saturday || weekday == time.Sunday
}

// IsWeekday checks if the given time falls on a weekday
func IsWeekday(t time.Time) bool {
	return !IsWeekend(t)
}

// TruncateToDay truncates a time to the beginning of the day
func TruncateToDay(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, t.Location())
}

// IsSameDay checks if two times are on the same day
func IsSameDay(t1, t2 time.Time) bool {
	y1, m1, d1 := t1.Date()
	y2, m2, d2 := t2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

// NextWeekday returns the next occurrence of the specified weekday
func NextWeekday(t time.Time, weekday time.Weekday) time.Time {
	daysUntilWeekday := int(weekday - t.Weekday())
	if daysUntilWeekday <= 0 {
		daysUntilWeekday += 7 // Next week
	}

	nextWeekday := t.AddDate(0, 0, daysUntilWeekday)
	return time.Date(nextWeekday.Year(), nextWeekday.Month(), nextWeekday.Day(), 0, 0, 0, 0, t.Location())
}
