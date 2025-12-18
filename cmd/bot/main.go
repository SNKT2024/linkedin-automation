package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/SNKT2024/linkedin-automation/internal/browser"
	"github.com/SNKT2024/linkedin-automation/internal/linkedin"
	"github.com/joho/godotenv"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

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
		log.Fatalf("LinkedIn login failed: %v", err)  // ✅ Only exits on ERROR
	}
	log.Println("Logged into LinkedIn successfully.")  // ✅ This confirms success

	// Start searching for profiles
	log.Println("Starting Search for 'Software Engineer'...")  // ✅ Guaranteed to run
	profiles, err := linkedin.SearchPeople(page, "Software Engineer")
	if err != nil {
		log.Fatalf("Search failed: %v", err)
	}

	// Display results
	fmt.Printf("\n==========================================\n")
	fmt.Printf("Search Results: Found %d profiles\n", len(profiles))
	fmt.Printf("==========================================\n\n")

	if len(profiles) == 0 {
		fmt.Println("No profiles found.")
	} else {
		for i, url := range profiles {
			fmt.Printf("%d. Found Profile: %s\n", i+1, url)
		}
	}

	// Wait to keep the browser open
	fmt.Println("\n==========================================")
	fmt.Println("Press Enter to exit...")
	fmt.Println("==========================================")
	fmt.Scanln()
}