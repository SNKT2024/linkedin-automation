package guard

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/SNKT2024/linkedin-automation/internal/config"
)

// CheckWorkingHours returns an error if the current time is outside working hours
// or if it's a weekend (Saturday/Sunday).
// Now uses configurable working hours from Config instead of hardcoded values.
func CheckWorkingHours(cfg *config.Config) error {
	now := time.Now()

	// Check if it's a weekend
	weekday := now.Weekday()
	if weekday == time.Saturday || weekday == time.Sunday {
		return errors.New("bot cannot run on weekends (Saturday/Sunday)")
	}

	// Parse WorkStart time (format: "HH:MM")
	startParts := strings.Split(cfg.WorkStart, ":")
	if len(startParts) != 2 {
		return fmt.Errorf("invalid WorkStart format: %s (expected HH:MM)", cfg.WorkStart)
	}
	startHour, err := strconv.Atoi(startParts[0])
	if err != nil {
		return fmt.Errorf("invalid WorkStart hour: %s", cfg.WorkStart)
	}
	startMinute, err := strconv.Atoi(startParts[1])
	if err != nil {
		return fmt.Errorf("invalid WorkStart minute: %s", cfg.WorkStart)
	}

	// Parse WorkEnd time (format: "HH:MM")
	endParts := strings.Split(cfg.WorkEnd, ":")
	if len(endParts) != 2 {
		return fmt.Errorf("invalid WorkEnd format: %s (expected HH:MM)", cfg.WorkEnd)
	}
	endHour, err := strconv.Atoi(endParts[0])
	if err != nil {
		return fmt.Errorf("invalid WorkEnd hour: %s", cfg.WorkEnd)
	}
	endMinute, err := strconv.Atoi(endParts[1])
	if err != nil {
		return fmt.Errorf("invalid WorkEnd minute: %s", cfg.WorkEnd)
	}

	// Get current time components
	currentHour := now.Hour()
	currentMinute := now.Minute()

	// Convert times to minutes since midnight for easier comparison
	currentMinutes := currentHour*60 + currentMinute
	startMinutes := startHour*60 + startMinute
	endMinutes := endHour*60 + endMinute

	// Check if current time is before work start
	if currentMinutes < startMinutes {
		return fmt.Errorf("bot cannot run before %s (current time: %s)",
			cfg.WorkStart, now.Format("15:04"))
	}

	// Check if current time is after work end
	if currentMinutes >= endMinutes {
		return fmt.Errorf("bot cannot run after %s (current time: %s)",
			cfg.WorkEnd, now.Format("15:04"))
	}

	// Within working hours
	return nil
}

// CheckDailyLimit checks if the daily profile collection limit has been reached.
// It counts how many profiles were added today and compares against the limit.
func CheckDailyLimit(db *sql.DB, limit int) error {
	// Get today's date at midnight (start of day)
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	// Query to count profiles created today
	query := `
        SELECT COUNT(*) 
        FROM profiles 
        WHERE created_at >= ?
    `

	var count int
	err := db.QueryRow(query, startOfDay).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check daily limit: %w", err)
	}

	// Check if limit is reached
	if count >= limit {
		return fmt.Errorf("daily limit reached: %d/%d profiles collected today", count, limit)
	}

	// Return remaining count for logging
	return nil
}

// GetTodayCount returns the number of profiles collected today.
// Useful for displaying progress without enforcing limits.
func GetTodayCount(db *sql.DB) (int, error) {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	query := `
        SELECT COUNT(*) 
        FROM profiles 
        WHERE created_at >= ?
    `

	var count int
	err := db.QueryRow(query, startOfDay).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get today's count: %w", err)
	}

	return count, nil
}

// GetRemainingLimit returns how many more profiles can be collected today.
func GetRemainingLimit(db *sql.DB, dailyLimit int) (int, error) {
	todayCount, err := GetTodayCount(db)
	if err != nil {
		return 0, err
	}

	remaining := dailyLimit - todayCount
	if remaining < 0 {
		remaining = 0
	}

	return remaining, nil
}