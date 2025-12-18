package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/SNKT2024/linkedin-automation/internal/browser"
	"github.com/SNKT2024/linkedin-automation/internal/config"
	"github.com/SNKT2024/linkedin-automation/internal/guard"
	"github.com/SNKT2024/linkedin-automation/internal/linkedin"
	"github.com/SNKT2024/linkedin-automation/internal/stealth"
	"github.com/SNKT2024/linkedin-automation/internal/storage"
	"github.com/go-rod/rod"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	// ==========================================
	// CONFIGURATION LOADING
	// ==========================================
	log.Println("Loading configuration from .env...")
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("❌ Failed to load configuration: %v", err)
	}
	log.Println("✅ Configuration loaded successfully")
	log.Printf("   Email: %s", cfg.Email)
	log.Printf("   Search Keyword: %s", cfg.SearchKeyword)
	log.Printf("   Max Pages: %d", cfg.MaxPages)
	log.Printf("   Daily Invite Limit: %d", cfg.InviteLimit)
	log.Printf("   Daily Search Limit: %d", cfg.SearchLimit)
	log.Printf("   Working Hours: %s - %s", cfg.WorkStart, cfg.WorkEnd)
	log.Printf("   Default Mode: %s", cfg.DefaultMode)

	// ==========================================
	// COMMAND-LINE FLAGS
	// ==========================================
	mode := flag.String("mode", cfg.DefaultMode, "Execution mode: search, connect, demo, login, message")
	flag.Parse()

	log.Printf("\n🎯 Execution Mode: %s\n", *mode)

	// ==========================================
	// SAFETY CHECKS
	// ==========================================
	log.Println("==========================================")
	log.Println("Performing Safety Checks...")
	log.Println("==========================================")

	// Check working hours (Mon-Fri, configured hours)
	log.Println("Checking working hours...")
	if err := guard.CheckWorkingHours(cfg); err != nil {
		log.Printf("⚠️ SAFETY STOP: %v", err)
		log.Println("The bot will not run outside of configured working hours.")
		os.Exit(1)
	}
	log.Println("✅ Working hours check passed")

	// ==========================================
	// DATABASE INITIALIZATION
	// ==========================================
	log.Println("\nInitializing database...")
	db, err := storage.InitDB()
	if err != nil {
		log.Fatalf("❌ Failed to initialize database: %v", err)
	}
	defer storage.CloseDB(db)
	log.Println("✅ Database initialized successfully")

	// Show current database stats
	stats, _ := storage.GetStats(db)
	log.Printf("📊 Current Database Status:")
	log.Printf("   Total Profiles: %d", stats.Total)
	log.Printf("   Found (ready): %d", stats.Found)
	log.Printf("   Invited: %d", stats.Invited)
	log.Printf("   Connected: %d", stats.Connected)

	// Check daily limits
	todayCount, _ := guard.GetTodayCount(db)
	remaining, _ := guard.GetRemainingLimit(db, cfg.SearchLimit)
	log.Printf("\n📅 Today's Activity:")
	log.Printf("   Collected Today: %d/%d", todayCount, cfg.SearchLimit)
	log.Printf("   Remaining: %d", remaining)

	if todayCount >= cfg.SearchLimit {
		log.Printf("⚠️ Daily search limit reached (%d/%d)", todayCount, cfg.SearchLimit)
		log.Println("Continuing with existing profiles only...")
	}

	log.Println("\n✅ All safety checks passed!")

	// ==========================================
	// BROWSER INITIALIZATION
	// ==========================================
	log.Println("\n==========================================")
	log.Println("Initializing Browser...")
	log.Println("==========================================")

	log.Println("Creating browser instance...")
	b, err := browser.NewBrowser()
	if err != nil {
		log.Fatalf("❌ Failed to initialize browser: %v", err)
	}
	defer b.MustClose()
	log.Println("✅ Browser created successfully")

	log.Println("Creating stealth page...")
	page, err := browser.NewStealthPage(b)
	if err != nil {
		log.Fatalf("❌ Failed to create stealth page: %v", err)
	}
	log.Println("✅ Stealth page created")

	// ==========================================
	// LINKEDIN AUTHENTICATION
	// ==========================================
	log.Println("\n==========================================")
	log.Println("Authenticating with LinkedIn...")
	log.Println("==========================================")

	if err := linkedin.Login(b, page, cfg); err != nil {
		log.Fatalf("❌ LinkedIn login failed: %v", err)
	}
	log.Println("✅ Successfully logged into LinkedIn")

	// ==========================================
	// MODE EXECUTION
	// ==========================================
	log.Println("\n==========================================")
	log.Printf("Executing Mode: %s", strings.ToUpper(*mode))
	log.Println("==========================================\n")

	switch strings.ToLower(*mode) {
	case "search":
		runSearchMode(page, db, cfg)

	case "connect":
		runConnectMode(page, db, cfg)

	case "demo":
		runDemoMode(page, db, cfg)

	case "login":
		log.Println("🔵 Execution Mode: LOGIN ONLY")
		log.Println("✅ Login successful. Browser will remain open for 5 minutes for manual inspection.")
		log.Println("💡 You can manually browse LinkedIn to build cookies/history.")
		log.Println("📍 This mode is useful for:")
		log.Println("   • Testing authentication")
		log.Println("   • Building cookie cache")
		log.Println("   • Manual profile exploration")
		log.Println("   • Debugging browser behavior")

		log.Println("\n⏳ Keeping browser open for 2 minutes...")
		for i := 2; i > 0; i-- {
			log.Printf("   Time remaining: %d minute(s)...", i)
			time.Sleep(1 * time.Minute)
		}

		log.Println("✅ Login mode complete. Closing browser...")

	case "message":
		log.Println("🟠 Execution Mode: MESSAGE")
		log.Println("⚠️ Messaging logic is not yet implemented. Please wait for the next update.")
		log.Println("📋 Planned features:")
		log.Println("   • Fetch profiles with status 'connected'")
		log.Println("   • Send personalized messages to connections")
		log.Println("   • Track message status in database")
		log.Println("   • Respect daily messaging limits")
		log.Println("\n💡 For now, you can use 'search' and 'connect' modes to build your network.")

	default:
		log.Fatalf("❌ Invalid mode: %s. Valid modes: search, connect, demo, login, message", *mode)
	}

	// ==========================================
	// FINAL STATISTICS
	// ==========================================
	showFinalStatistics(db, cfg)

	// Keep browser open
	fmt.Println("\n✅ Execution complete. Press Enter to exit...")
	fmt.Scanln()
}

// runSearchMode executes the search workflow
func runSearchMode(page *rod.Page, db *sql.DB, cfg *config.Config) {
	log.Println("🔍 Testing Search Mode...")
	log.Printf("   Keyword: %s", cfg.SearchKeyword)
	log.Printf("   Max Pages: %d", cfg.MaxPages)

	newProfiles, err := linkedin.SearchPeople(page, db, cfg.SearchKeyword, cfg.MaxPages)
	if err != nil {
		log.Printf("❌ Search failed: %v", err)
		return
	}

	log.Printf("\n✅ Search Test Complete!")
	log.Printf("📊 Found %d NEW profiles", len(newProfiles))
	log.Println("💾 Check database for profiles with status 'found'")

	if len(newProfiles) > 0 {
		log.Println("\nSample of new profiles:")
		for i, url := range newProfiles {
			if i >= 5 {
				log.Printf("   ... and %d more", len(newProfiles)-5)
				break
			}
			log.Printf("   %d. %s", i+1, url)
		}
	}
}

// runConnectMode executes the connection workflow
func runConnectMode(page *rod.Page, db *sql.DB, cfg *config.Config) {
	log.Println("🤝 Starting Connect Mode...")

	// Fetch profiles to invite
	log.Printf("Fetching up to %d profiles to invite...", cfg.InviteLimit)
	profiles, err := storage.GetProfilesToInvite(db, cfg.InviteLimit)
	if err != nil {
		log.Printf("❌ Failed to fetch profiles: %v", err)
		return
	}

	if len(profiles) == 0 {
		log.Println("⚠️ No profiles available for connection")
		log.Println("💡 Run in 'search' mode first to collect profiles")
		return
	}

	log.Printf("Found %d profiles ready for connection\n", len(profiles))

	// Connection statistics
	var (
		successCount     = 0
		pendingCount     = 0
		alreadyConnected = 0
		premiumSkipped   = 0
		failedCount      = 0
	)

	// Process each profile
	for i, profileURL := range profiles {
		log.Printf("\n========== Profile %d/%d ==========", i+1, len(profiles))
		log.Printf("Processing: %s", profileURL)

		// Navigate to profile
		log.Println("Navigating to profile...")
		page.MustNavigate(profileURL)
		page.MustWaitLoad()
		stealth.RandomSleep(2000, 4000)

		// Extract first name from profile
		firstName := "there" // Default fallback
		err := rod.Try(func() {
			// Find the h1 element containing the name
			nameElement := page.Timeout(5 * time.Second).MustElement("h1")
			fullName := strings.TrimSpace(nameElement.MustText())

			// Split by space and take first name
			nameParts := strings.Fields(fullName)
			if len(nameParts) > 0 {
				firstName = nameParts[0]
				log.Printf("Extracted name: %s (full: %s)", firstName, fullName)
			}
		})

		if err != nil {
			log.Printf("⚠️ Could not extract name, using default: %s", firstName)
		}

		// Compose personalized message (for future use)
		message := fmt.Sprintf("Hi %s, I came across your profile and would love to connect!", firstName)
		log.Printf("Composed message: %s", message)

		// Attempt to connect
		status, connErr := linkedin.ConnectWithProfile(page, profileURL)

		// Handle the result
		switch status {
		case "clicked":
			log.Println("✅ Connection request sent successfully")
			successCount++
			storage.UpdateStatus(db, profileURL, "invited")

		case "skipped_pending":
			log.Println("⏭️  Connection already pending")
			pendingCount++
			storage.UpdateStatus(db, profileURL, "pending")

		case "skipped_connected":
			log.Println("⏭️  Already connected")
			alreadyConnected++
			storage.UpdateStatus(db, profileURL, "already_connected")

		case "skipped_premium":
			log.Println("⏭️  Premium profile - InMail required")
			premiumSkipped++
			storage.UpdateStatus(db, profileURL, "premium_only")

		case "failed":
			log.Printf("❌ Failed to connect: %v", connErr)
			failedCount++
			// Keep status as 'found' so it can be retried

		default:
			log.Printf("⚠️ Unknown status: %s", status)
			failedCount++
		}

		// Critical safety delay between connection attempts
		if i < len(profiles)-1 {
			waitTime := 15000 + rand.Intn(15000) // 15-30 seconds
			log.Printf("⏳ Safety delay: waiting %d ms before next connection...", waitTime)
			stealth.RandomSleep(waitTime, waitTime+1000)
		}
	}

	// Connection summary
	log.Println("\n==========================================")
	log.Println("Connect Mode Complete")
	log.Println("==========================================")
	log.Printf("✅ Connections Sent:     %d\n", successCount)
	log.Printf("⏭️  Already Pending:      %d\n", pendingCount)
	log.Printf("⏭️  Already Connected:    %d\n", alreadyConnected)
	log.Printf("💎 Premium/InMail Only:  %d\n", premiumSkipped)
	log.Printf("❌ Failed (will retry):  %d\n", failedCount)
	log.Printf("📊 Total Processed:      %d\n", len(profiles))
	log.Println("==========================================")
}

// runDemoMode executes the demo workflow (search → wait → connect)
func runDemoMode(page *rod.Page, db *sql.DB, cfg *config.Config) {
	log.Println("🎯 Running Demo Sequence...")
	log.Println("This will execute: Search → Wait 10s → Connect")

	// Phase 1: Search
	log.Println("\n📍 Phase 1: Search")
	runSearchMode(page, db, cfg)

	// Phase 2: Wait
	log.Println("\n📍 Phase 2: Waiting 10 seconds...")
	for i := 10; i > 0; i-- {
		log.Printf("   %d...", i)
		time.Sleep(1 * time.Second)
	}

	// Phase 3: Connect
	log.Println("\n📍 Phase 3: Connect")
	runConnectMode(page, db, cfg)

	log.Println("\n✅ Demo sequence completed!")
}

// showFinalStatistics displays comprehensive database statistics
func showFinalStatistics(db *sql.DB, cfg *config.Config) {
	log.Println("\n==========================================")
	log.Println("FINAL DATABASE STATISTICS")
	log.Println("==========================================")

	stats, err := storage.GetStats(db)
	if err != nil {
		log.Printf("⚠️ Could not retrieve statistics: %v", err)
		return
	}

	log.Printf("Total Profiles:          %d", stats.Total)
	log.Printf("├─ Found (ready):        %d", stats.Found)
	log.Printf("├─ Invited (sent):       %d", stats.Invited)
	log.Printf("├─ Connected:            %d", stats.Connected)
	log.Printf("├─ Messaged:             %d", stats.Messaged)
	log.Printf("├─ Pending:              %d", stats.Pending)
	log.Printf("├─ Premium Only:         %d", stats.Premium)
	log.Printf("└─ Failed (retry):       %d", stats.Failed)

	// Today's activity
	todayCount, _ := guard.GetTodayCount(db)
	remaining, _ := guard.GetRemainingLimit(db, cfg.SearchLimit)

	log.Printf("\nToday's Activity:")
	log.Printf("├─ Collected Today:      %d/%d", todayCount, cfg.SearchLimit)
	log.Printf("└─ Remaining Today:      %d", remaining)

	log.Println("==========================================")
}