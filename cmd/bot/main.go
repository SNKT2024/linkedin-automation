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

	// Navigate to a page first to test cursor
	page.MustNavigate("https://www.linkedin.com")
	page.MustWaitLoad()

	// Test cursor visibility
	browser.TestCursor(page)
	time.Sleep(2 * time.Second) // Pause so you can see the cursor

	// Log into LinkedIn
	log.Println("Logging into LinkedIn...")
	if err := linkedin.Login(b, page); err != nil {
		log.Fatalf("LinkedIn login failed: %v", err)
	}
	log.Println("Logged into LinkedIn successfully.")

	// Wait to keep the browser open
	fmt.Println("Press Enter to exit...")
	fmt.Scanln()
}