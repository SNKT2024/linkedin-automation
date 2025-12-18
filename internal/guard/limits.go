package guard

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// CheckWorkingHours returns an error if the current time is outside working hours
// or if it's a weekend (Saturday/Sunday).
// Working hours: 09:00 - 21:00 (9 AM - 9 PM) on weekdays only.
func CheckWorkingHours() error {
    now := time.Now()

    // Check if it's a weekend
    weekday := now.Weekday()
    if weekday == time.Saturday || weekday == time.Sunday {
        return errors.New("bot cannot run on weekends (Saturday/Sunday)")
    }

    // Get current hour
    currentHour := now.Hour()

    // Check if time is before 09:00
    if currentHour < 9 {
        return fmt.Errorf("bot cannot run before 09:00 (current time: %s)", now.Format("15:04"))
    }

    // Check if time is after 21:00 (9 PM)
    if currentHour >= 21 {
        return fmt.Errorf("bot cannot run after 21:00 (current time: %s)", now.Format("15:04"))
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