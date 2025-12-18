package storage

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	_ "modernc.org/sqlite" // Pure Go SQLite driver (no CGO required)
)

const (
	dbFile      = "linkedin.db"
	createTable = `
        CREATE TABLE IF NOT EXISTS profiles (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            url TEXT UNIQUE NOT NULL,
            status TEXT NOT NULL DEFAULT 'found',
            created_at DATETIME NOT NULL,
            updated_at DATETIME NOT NULL
        );
        CREATE INDEX IF NOT EXISTS idx_url ON profiles(url);
        CREATE INDEX IF NOT EXISTS idx_status ON profiles(status);
        CREATE INDEX IF NOT EXISTS idx_created_at ON profiles(created_at);
        CREATE INDEX IF NOT EXISTS idx_updated_at ON profiles(updated_at);
    `
)

// ProfileStats holds statistics about profiles in different lifecycle stages
type ProfileStats struct {
	Total     int
	Found     int
	Invited   int
	Connected int
	Messaged  int
	Pending   int
	Premium   int
	Failed    int
}

// InitDB initializes the SQLite database and creates necessary tables.
func InitDB() (*sql.DB, error) {
	log.Printf("Initializing database: %s", dbFile)

	// Open database connection (note: driver name is "sqlite" not "sqlite3")
	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		return nil, err
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, err
	}

	// Enable WAL mode for better concurrency
	if _, err := db.Exec("PRAGMA journal_mode=WAL;"); err != nil {
		log.Printf("Warning: Failed to enable WAL mode: %v", err)
	} else {
		log.Println("WAL mode enabled for better concurrency")
	}

	// Create tables and indexes
	if _, err := db.Exec(createTable); err != nil {
		return nil, err
	}

	log.Println("Database initialized successfully")
	return db, nil
}

// AddProfile inserts a new profile URL into the database with status "found".
// Uses INSERT OR IGNORE to handle duplicate URLs gracefully.
// This is the entry point for all newly discovered profiles.
func AddProfile(db *sql.DB, url string) error {
	query := `
        INSERT OR IGNORE INTO profiles (url, status, created_at, updated_at)
        VALUES (?, 'found', ?, ?)
    `
	now := time.Now()
	result, err := db.Exec(query, url, now, now)
	if err != nil {
		log.Printf("Error adding profile %s: %v", url, err)
		return err
	}

	// Check if row was actually inserted
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		log.Printf("✅ Added new profile: %s", url)
	}

	return nil
}

// IsProfileVisited checks if a profile URL exists in the database.
// Used for deduplication during search phase.
func IsProfileVisited(db *sql.DB, url string) bool {
	var count int
	query := "SELECT COUNT(*) FROM profiles WHERE url = ?"
	err := db.QueryRow(query, url).Scan(&count)
	if err != nil {
		log.Printf("Error checking if profile exists: %v", err)
		return false
	}
	return count > 0
}

// GetProfilesToInvite retrieves profiles with status 'found' that need connection invites.
// Returns up to 'limit' profiles, ordered by oldest first (created_at ASC).
// This ensures we process profiles in the order they were discovered.
func GetProfilesToInvite(db *sql.DB, limit int) ([]string, error) {
	query := `
        SELECT url 
        FROM profiles 
        WHERE status = 'found' 
        ORDER BY created_at ASC 
        LIMIT ?
    `
	rows, err := db.Query(query, limit)
	if err != nil {
		log.Printf("Error fetching profiles to invite: %v", err)
		return nil, err
	}
	defer rows.Close()

	var urls []string
	for rows.Next() {
		var url string
		if err := rows.Scan(&url); err != nil {
			log.Printf("Error scanning profile: %v", err)
			continue
		}
		urls = append(urls, url)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	log.Printf("Found %d profiles ready for invitation", len(urls))
	return urls, nil
}

// UpdateStatus updates the status of a profile and sets updated_at to current timestamp.
// This is the primary function for advancing profiles through the lifecycle:
// found → invited → connected → messaged
func UpdateStatus(db *sql.DB, url string, newStatus string) error {
	query := `
        UPDATE profiles
        SET status = ?, updated_at = CURRENT_TIMESTAMP
        WHERE url = ?
    `
	result, err := db.Exec(query, newStatus, url)
	if err != nil {
		log.Printf("Error updating status for %s to '%s': %v", url, newStatus, err)
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		log.Printf("⚠️ Warning: No profile found with URL %s", url)
	} else {
		log.Printf("✅ Updated %s status to '%s'", url, newStatus)
	}

	return nil
}

// GetProfilesByStatus retrieves profiles with a specific status up to the specified limit.
// Ordered by most recently updated first (updated_at ASC for fairness).
// Useful for:
// - Getting 'connected' profiles to message
// - Getting 'invited' profiles to check acceptance
// - Getting 'failed' profiles to retry
func GetProfilesByStatus(db *sql.DB, status string, limit int) ([]string, error) {
	query := `
        SELECT url 
        FROM profiles 
        WHERE status = ? 
        ORDER BY updated_at ASC 
        LIMIT ?
    `
	rows, err := db.Query(query, status, limit)
	if err != nil {
		log.Printf("Error fetching profiles with status '%s': %v", status, err)
		return nil, err
	}
	defer rows.Close()

	var urls []string
	for rows.Next() {
		var url string
		if err := rows.Scan(&url); err != nil {
			log.Printf("Error scanning profile: %v", err)
			continue
		}
		urls = append(urls, url)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	log.Printf("Found %d profiles with status '%s'", len(urls), status)
	return urls, nil
}

// GetStats returns comprehensive statistics about profiles in the database.
// Provides a complete overview of the automation pipeline state.
func GetStats(db *sql.DB) (*ProfileStats, error) {
	stats := &ProfileStats{}

	// Get total count
	err := db.QueryRow("SELECT COUNT(*) FROM profiles").Scan(&stats.Total)
	if err != nil {
		return nil, fmt.Errorf("failed to get total count: %w", err)
	}

	// Get count by status
	statusCounts := map[string]*int{
		"found":        &stats.Found,
		"invited":      &stats.Invited,
		"connected":    &stats.Connected,
		"messaged":     &stats.Messaged,
		"pending":      &stats.Pending,
		"premium_only": &stats.Premium,
		"failed":       &stats.Failed,
	}

	for status, countPtr := range statusCounts {
		err := db.QueryRow("SELECT COUNT(*) FROM profiles WHERE status = ?", status).Scan(countPtr)
		if err != nil {
			log.Printf("Warning: Failed to get count for status '%s': %v", status, err)
			*countPtr = 0
		}
	}

	log.Printf("📊 Database Stats - Total: %d, Found: %d, Invited: %d, Connected: %d, Messaged: %d",
		stats.Total, stats.Found, stats.Invited, stats.Connected, stats.Messaged)

	return stats, nil
}

// Count returns the number of profiles with a specific status.
// Convenience wrapper for quick status checks.
func Count(db *sql.DB, status string) (int, error) {
	var count int
	query := "SELECT COUNT(*) FROM profiles WHERE status = ?"
	err := db.QueryRow(query, status).Scan(&count)
	if err != nil {
		log.Printf("Error counting profiles with status '%s': %v", status, err)
		return 0, err
	}
	return count, nil
}

// GetTotalProfileCount returns the total number of profiles in the database.
func GetTotalProfileCount(db *sql.DB) (int, error) {
	var count int
	query := "SELECT COUNT(*) FROM profiles"
	err := db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get total profile count: %w", err)
	}
	return count, nil
}

// GetProfileCountByStatus returns the count of profiles with a specific status.
// Deprecated: Use Count() instead for consistency.
func GetProfileCountByStatus(db *sql.DB, status string) (int, error) {
	return Count(db, status)
}

// GetAllProfiles retrieves all profile URLs from the database.
// Use with caution on large databases - prefer GetProfilesByStatus with limit.
func GetAllProfiles(db *sql.DB) ([]string, error) {
	query := "SELECT url FROM profiles ORDER BY created_at ASC"
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var urls []string
	for rows.Next() {
		var url string
		if err := rows.Scan(&url); err != nil {
			log.Printf("Error scanning profile: %v", err)
			continue
		}
		urls = append(urls, url)
	}

	return urls, rows.Err()
}

// UpdateProfileStatus is a legacy wrapper around UpdateStatus for backward compatibility.
// Deprecated: Use UpdateStatus() directly.
func UpdateProfileStatus(db *sql.DB, url string, status string) error {
	return UpdateStatus(db, url, status)
}

// CloseDB closes the database connection gracefully.
func CloseDB(db *sql.DB) error {
	log.Println("Closing database connection...")
	return db.Close()
}

// GetProfilesCreatedToday returns count of profiles created today (for daily limits).
func GetProfilesCreatedToday(db *sql.DB) (int, error) {
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
		return 0, fmt.Errorf("failed to count today's profiles: %w", err)
	}

	return count, nil
}

// GetProfilesUpdatedToday returns count of profiles updated today (for activity tracking).
func GetProfilesUpdatedToday(db *sql.DB) (int, error) {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	query := `
        SELECT COUNT(*) 
        FROM profiles 
        WHERE updated_at >= ?
    `

	var count int
	err := db.QueryRow(query, startOfDay).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count today's activity: %w", err)
	}

	return count, nil
}

// PrintStats displays a formatted summary of database statistics.
// Useful for debugging and monitoring.
func PrintStats(db *sql.DB) error {
	stats, err := GetStats(db)
	if err != nil {
		return err
	}

	fmt.Println("\n"+strings.Repeat("=", 50))
	fmt.Println("📊 LINKEDIN AUTOMATION DATABASE STATISTICS")
	fmt.Println(strings.Repeat("=", 50))
	fmt.Printf("Total Profiles:          %d\n", stats.Total)
	fmt.Printf("├─ Found (ready):        %d\n", stats.Found)
	fmt.Printf("├─ Invited (pending):    %d\n", stats.Invited)
	fmt.Printf("├─ Connected:            %d\n", stats.Connected)
	fmt.Printf("├─ Messaged:             %d\n", stats.Messaged)
	fmt.Printf("├─ Pending Review:       %d\n", stats.Pending)
	fmt.Printf("├─ Premium Only:         %d\n", stats.Premium)
	fmt.Printf("└─ Failed (retry):       %d\n", stats.Failed)
	fmt.Println(strings.Repeat("=", 50) + "\n")

	return nil
}