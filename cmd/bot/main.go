package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/SNKT2024/linkedin-automation/internal/browser"
	"github.com/SNKT2024/linkedin-automation/internal/linkedin"
	"github.com/SNKT2024/linkedin-automation/internal/storage"
	"github.com/joho/godotenv"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Initialize Database
	log.Println("Initializing Database...")
	db, err := storage.InitDB()
	if err != nil {
		log.Fatalf("Failed to init DB: %v", err)
	}
	defer db.Close()
	log.Println("Database initialized and ready for duplicate detection.")

	// Initialize the browser
	log.Println("Calling NewBrowser...")
	b, err := browser.NewBrowser()
	if err != nil {
		log.Fatalf("Failed to initialize browser: %v", err)
	}
	defer b.MustClose()
	log.Println("Browser initialized.")

	// Create a new stealth page
	log.Println("Creating a new stealth page...")
	page, err := browser.NewStealthPage(b)
	if err != nil {
		log.Fatalf("Failed to create stealth page: %v", err)
	}
	log.Println("Stealth page created.")

	// Enable cursor visualization for debugging BEFORE any navigation
	log.Println("Enabling cursor visualization...")
	browser.ShowCursor(page)
	log.Println("Cursor visualization enabled.")

	// Log into LinkedIn
	log.Println("Logging into LinkedIn...")
	if err := linkedin.Login(b, page); err != nil {
		log.Fatalf("LinkedIn login failed: %v", err)
	}
	log.Println("✅ Logged into LinkedIn successfully.")

	// Start searching for profiles with database integration
	log.Println("\nStarting Search for 'Software Engineer'...")
	profiles, err := linkedin.SearchPeople(page, db, "Software Engineer")
	if err != nil {
		log.Fatalf("Search failed: %v", err)
	}

	// Log completion with count
	log.Printf("Search complete. Found %d NEW profiles.", len(profiles))

	// Display results
	fmt.Printf("\n==========================================\n")
	fmt.Printf("Search Results: Found %d NEW profiles\n", len(profiles))
	fmt.Printf("==========================================\n\n")

	if len(profiles) == 0 {
		fmt.Println("No new profiles found (all were already in database).")
	} else {
		for i, url := range profiles {
			fmt.Printf("%d. New Profile: %s\n", i+1, url)
		}
	}

	// Show database statistics
	totalCount, _ := storage.GetTotalProfileCount(db)
	foundCount, _ := storage.GetProfileCountByStatus(db, "found")

	fmt.Printf("\n==========================================\n")
	fmt.Printf("Database Statistics:\n")
	fmt.Printf("  Total Profiles: %d\n", totalCount)
	fmt.Printf("  Status 'found': %d\n", foundCount)
	fmt.Printf("==========================================\n")

	// Wait to keep the browser open
	fmt.Println("\nPress Enter to exit...")
	fmt.Scanln()
}