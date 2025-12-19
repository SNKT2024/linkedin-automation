package config

import (
	"errors"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all configuration values loaded from environment variables
type Config struct {
    // LinkedIn Credentials
    Email    string
    Password string

    // Search Settings
    SearchKeyword  string
    SearchLocation string
    MaxPages       int

    // Stealth & Human Behavior
    DelayFactor float64
    ScrollMin   int
    ScrollMax   int

    // Safety Limits (Daily)
    InviteLimit int
    SearchLimit int

    // Working Hours (24h format)
    WorkStart string
    WorkEnd   string

    // Message Templates
    ConnectMessageTemplate  string
	FollowupMessageTemplate string

    // Execution Defaults
    DefaultMode string
}

// Load reads configuration from environment variables and returns a Config struct.
// Automatically loads .env file and returns an error if required credentials are missing.
func Load() (*Config, error) {
    // Load .env file (ignore error if file doesn't exist - allow system env vars)
    _ = godotenv.Load()

    // Required credentials
    email := os.Getenv("LINKEDIN_EMAIL")
    password := os.Getenv("LINKEDIN_PASSWORD")

    if email == "" || password == "" {
        return nil, errors.New("LINKEDIN_EMAIL and LINKEDIN_PASSWORD must be set in .env file")
    }

    cfg := &Config{
        Email:    email,
        Password: password,

        // Search Settings with defaults
        SearchKeyword:  getEnvOrDefault("SEARCH_KEYWORD", "Software Engineer"),
        MaxPages:       getEnvAsInt("MAX_PAGES_TO_SCRAPE", 3),

        // Stealth & Human Behavior with defaults
        DelayFactor: getEnvAsFloat("DELAY_FACTOR", 1.0),
        ScrollMin:   getEnvAsInt("SCROLL_COUNT_MIN", 3),
        ScrollMax:   getEnvAsInt("SCROLL_COUNT_MAX", 7),

        // Safety Limits with defaults
        InviteLimit: getEnvAsInt("DAILY_INVITE_LIMIT", 10),
        SearchLimit: getEnvAsInt("DAILY_SEARCH_LIMIT", 50),

        // Working Hours with defaults
        WorkStart: getEnvOrDefault("WORKING_HOURS_START", "09:00"),
        WorkEnd:   getEnvOrDefault("WORKING_HOURS_END", "21:00"),

        // Message and Note Template
		ConnectMessageTemplate:  getEnvOrDefault("CONNECT_MESSAGE_TEMPLATE", "Hi {firstName}, I noticed your profile and would love to connect!"),
		FollowupMessageTemplate: getEnvOrDefault("FOLLOW_UP_MESSAGE_TEMPLATE", "Hi {firstName}, thanks for connecting! Great to meet you."),

        // Execution Defaults
        DefaultMode: getEnvOrDefault("DEFAULT_MODE", "demo"),
    }

    // Validate working hours format (basic check)
    if !isValidTimeFormat(cfg.WorkStart) || !isValidTimeFormat(cfg.WorkEnd) {
        return nil, errors.New("WORKING_HOURS_START and WORKING_HOURS_END must be in HH:MM format")
    }

    return cfg, nil
}

// getEnvOrDefault returns the environment variable value or a default if not set
func getEnvOrDefault(key, defaultValue string) string {
    value := os.Getenv(key)
    if value == "" {
        return defaultValue
    }
    return value
}

// getEnvAsInt returns the environment variable as an integer or a default if not set/invalid
func getEnvAsInt(key string, defaultValue int) int {
    valueStr := os.Getenv(key)
    if valueStr == "" {
        return defaultValue
    }

    value, err := strconv.Atoi(valueStr)
    if err != nil {
        return defaultValue
    }

    return value
}

// getEnvAsFloat returns the environment variable as a float64 or a default if not set/invalid
func getEnvAsFloat(key string, defaultValue float64) float64 {
    valueStr := os.Getenv(key)
    if valueStr == "" {
        return defaultValue
    }

    value, err := strconv.ParseFloat(valueStr, 64)
    if err != nil {
        return defaultValue
    }

    return value
}

// isValidTimeFormat checks if a time string is in HH:MM format
func isValidTimeFormat(timeStr string) bool {
    if len(timeStr) != 5 {
        return false
    }
    if timeStr[2] != ':' {
        return false
    }
    // Check if HH and MM are numbers
    hour, err1 := strconv.Atoi(timeStr[0:2])
    minute, err2 := strconv.Atoi(timeStr[3:5])
    if err1 != nil || err2 != nil {
        return false
    }
    // Validate ranges
    return hour >= 0 && hour <= 23 && minute >= 0 && minute <= 59
}