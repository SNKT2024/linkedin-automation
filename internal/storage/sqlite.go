package storage

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	_ "modernc.org/sqlite" // Pure Go SQLite driver
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

	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	if _, err := db.Exec("PRAGMA journal_mode=WAL;"); err != nil {
		log.Printf("Warning: Failed to enable WAL mode: %v", err)
	} else {
		log.Println("WAL mode enabled for better concurrency")
	}

	if _, err := db.Exec(createTable); err != nil {
		return nil, err
	}

	log.Println("Database initialized successfully")
	return db, nil
}

// AddProfile inserts a new profile URL into the database.
// RETURNS: (bool, error) -> true if added, false if duplicate/ignored
func AddProfile(db *sql.DB, url string) (bool, error) {
	query := `
        INSERT OR IGNORE INTO profiles (url, status, created_at, updated_at)
        VALUES (?, 'found', ?, ?)
    `
	now := time.Now()
	result, err := db.Exec(query, url, now, now)
	if err != nil {
		log.Printf("Error adding profile %s: %v", url, err)
		return false, err
	}

	// Check if row was actually inserted
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		// log.Printf("✅ Added new profile: %s", url) // Optional: uncomment for verbose logs
		return true, nil
	}

	return false, nil
}

// IsProfileVisited checks if a profile URL exists in the database.
func IsProfileVisited(db *sql.DB, url string) bool {
	var count int
	query := "SELECT COUNT(*) FROM profiles WHERE url = ?"
	err := db.QueryRow(query, url).Scan(&count)
	if err != nil {
		return false
	}
	return count > 0
}

// GetProfilesToInvite retrieves profiles with status 'found' that need connection invites.
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
		return nil, err
	}
	defer rows.Close()

	var urls []string
	for rows.Next() {
		var url string
		if err := rows.Scan(&url); err != nil {
			continue
		}
		urls = append(urls, url)
	}

	log.Printf("Found %d profiles ready for invitation", len(urls))
	return urls, nil
}

// UpdateStatus updates the status of a profile.
func UpdateStatus(db *sql.DB, url string, newStatus string) error {
	query := `
        UPDATE profiles
        SET status = ?, updated_at = CURRENT_TIMESTAMP
        WHERE url = ?
    `
	result, err := db.Exec(query, newStatus, url)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		log.Printf("✅ Updated %s status to '%s'", url, newStatus)
	}

	return nil
}

// GetProfilesByStatus retrieves profiles with a specific status.
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
		return nil, err
	}
	defer rows.Close()

	var urls []string
	for rows.Next() {
		var url string
		if err := rows.Scan(&url); err != nil {
			continue
		}
		urls = append(urls, url)
	}

	return urls, nil
}

// GetStats returns comprehensive statistics about profiles in the database.
func GetStats(db *sql.DB) (*ProfileStats, error) {
	stats := &ProfileStats{}

	// Get total count
	if err := db.QueryRow("SELECT COUNT(*) FROM profiles").Scan(&stats.Total); err != nil {
		return nil, err
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
		_ = db.QueryRow("SELECT COUNT(*) FROM profiles WHERE status = ?", status).Scan(countPtr)
	}

	return stats, nil
}

// Count returns the number of profiles with a specific status.
func Count(db *sql.DB, status string) (int, error) {
	var count int
	query := "SELECT COUNT(*) FROM profiles WHERE status = ?"
	err := db.QueryRow(query, status).Scan(&count)
	return count, err
}

// GetProfilesCreatedToday returns count of profiles created today (for daily limits).
func GetProfilesCreatedToday(db *sql.DB) (int, error) {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM profiles WHERE created_at >= ?", startOfDay).Scan(&count)
	return count, err
}

// GetProfilesUpdatedToday returns count of profiles updated today (for activity tracking).
func GetProfilesUpdatedToday(db *sql.DB) (int, error) {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM profiles WHERE updated_at >= ?", startOfDay).Scan(&count)
	return count, err
}

// PrintStats displays a formatted summary of database statistics.
func PrintStats(db *sql.DB) error {
	stats, err := GetStats(db)
	if err != nil {
		return err
	}

	fmt.Println("\n" + strings.Repeat("=", 50))
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

// CloseDB closes the database connection gracefully.
func CloseDB(db *sql.DB) error {
	return db.Close()
}