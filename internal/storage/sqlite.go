package storage

import (
	"database/sql"
	"log"
	"time"

	_ "modernc.org/sqlite" // Pure Go SQLite driver (no CGO required)
)

const (
	dbFile      = "linkedin.db"
	createTable = `
        CREATE TABLE IF NOT EXISTS profiles (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            url TEXT UNIQUE NOT NULL,
            status TEXT NOT NULL,
            created_at DATETIME NOT NULL,
            updated_at DATETIME NOT NULL
        );
        CREATE INDEX IF NOT EXISTS idx_url ON profiles(url);
        CREATE INDEX IF NOT EXISTS idx_status ON profiles(status);
    `
)

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

	// Create tables
	if _, err := db.Exec(createTable); err != nil {
		return nil, err
	}

	log.Println("Database initialized successfully")
	return db, nil
}

// IsProfileVisited checks if a profile URL exists in the database.
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

// AddProfile inserts a new profile URL into the database with status "found".
// Uses INSERT OR IGNORE to handle duplicate URLs gracefully.
func AddProfile(db *sql.DB, url string) error {
	query := `
        INSERT OR IGNORE INTO profiles (url, status, created_at, updated_at)
        VALUES (?, ?, ?, ?)
    `
	now := time.Now()
	_, err := db.Exec(query, url, "found", now, now)
	if err != nil {
		log.Printf("Error adding profile %s: %v", url, err)
		return err
	}
	return nil
}

// UpdateProfileStatus updates the status of a profile.
func UpdateProfileStatus(db *sql.DB, url string, status string) error {
	query := `
        UPDATE profiles
        SET status = ?, updated_at = ?
        WHERE url = ?
    `
	_, err := db.Exec(query, status, time.Now(), url)
	if err != nil {
		log.Printf("Error updating profile %s status to %s: %v", url, status, err)
		return err
	}
	return nil
}

// GetProfilesByStatus retrieves all profiles with a specific status.
func GetProfilesByStatus(db *sql.DB, status string) ([]string, error) {
	query := "SELECT url FROM profiles WHERE status = ? ORDER BY created_at ASC"
	rows, err := db.Query(query, status)
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

// GetTotalProfileCount returns the total number of profiles in the database.
func GetTotalProfileCount(db *sql.DB) (int, error) {
	var count int
	query := "SELECT COUNT(*) FROM profiles"
	err := db.QueryRow(query).Scan(&count)
	return count, err
}

// GetProfileCountByStatus returns the count of profiles with a specific status.
func GetProfileCountByStatus(db *sql.DB, status string) (int, error) {
	var count int
	query := "SELECT COUNT(*) FROM profiles WHERE status = ?"
	err := db.QueryRow(query, status).Scan(&count)
	return count, err
}

// GetAllProfiles retrieves all profile URLs from the database.
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

// CloseDB closes the database connection.
func CloseDB(db *sql.DB) error {
	log.Println("Closing database connection...")
	return db.Close()
}