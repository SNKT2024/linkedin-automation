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
	log.Printf("   Daily Invite Limit: %d", cfg.InviteLimit)
	log.Printf("   Daily Search Limit: %d", cfg.SearchLimit)
	log.Printf("   Working Hours: %s - %s", cfg.WorkStart, cfg.WorkEnd)

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

	// 1. Check working hours
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

	// ==========================================
	// BROWSER INITIALIZATION
	// ==========================================
	log.Println("\n==========================================")
	log.Println("Initializing Browser...")
	log.Println("==========================================")

	b, err := browser.NewBrowser()
	if err != nil {
		log.Fatalf("❌ Failed to initialize browser: %v", err)
	}
	defer b.MustClose()

	page, err := browser.NewStealthPage(b)
	if err != nil {
		log.Fatalf("❌ Failed to create stealth page: %v", err)
	}
	log.Println("✅ Browser & Stealth Page Ready")

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
		log.Println("🔵 Login Mode: Keeping browser open for manual inspection.")
		for i := 2; i > 0; i-- {
			log.Printf("   Time remaining: %d minute(s)...", i)
			time.Sleep(1 * time.Minute)
		}

	case "message":
		runMessageMode(page,db,cfg)

	default:
		log.Fatalf("❌ Invalid mode: %s", *mode)
	}

	// ==========================================
	// FINAL STATISTICS
	// ==========================================
	showFinalStatistics(db, cfg)

	fmt.Println("\n✅ Execution complete. Press Enter to exit...")
	fmt.Scanln()
}

// runSearchMode executes the search workflow with rate limiting
func runSearchMode(page *rod.Page, db *sql.DB, cfg *config.Config) {
	log.Println("🔍 Starting Search Mode...")

	// 1. RATE LIMIT CHECK
	todayCount, err := guard.GetTodayCount(db)
	if err != nil {
		log.Printf("⚠️ Error checking search limits: %v", err)
		return
	}

	log.Printf("📊 Search Limit Status: %d/%d profiles collected today", todayCount, cfg.SearchLimit)

	if todayCount >= cfg.SearchLimit {
		log.Println("🛑 Daily search limit reached. Skipping search execution.")
		return
	}

	// Calculate allowable pages (optional optimization)
	// We run the search anyway, relying on the loop to stop or just run max pages 
	// since we want to fill the buffer.
	
	newProfiles, err := linkedin.SearchPeople(page, db, cfg.SearchKeyword, cfg.MaxPages)
	if err != nil {
		log.Printf("❌ Search failed: %v", err)
		return
	}

	log.Printf("\n✅ Search Complete. Found %d NEW profiles.", len(newProfiles))
}

// runConnectMode executes the connection workflow with strict rate limiting & personalization
func runConnectMode(page *rod.Page, db *sql.DB, cfg *config.Config) {
	log.Println("🤝 Starting Connect Mode...")

	// 1. RATE LIMIT CHECK
	inviteCount, err := guard.GetDailyInviteCount(db)
	if err != nil {
		log.Printf("⚠️ Error checking invite limits: %v", err)
		return
	}

	remaining := cfg.InviteLimit - inviteCount
	log.Printf("📊 Invite Limit Status: %d/%d sent today (Remaining: %d)", inviteCount, cfg.InviteLimit, remaining)

	if remaining <= 0 {
		log.Println("🛑 Daily invite limit reached. Stopping Connect Mode.")
		return
	}

	// 2. Fetch profiles
	log.Printf("Fetching up to %d profiles to invite...", remaining)
	profiles, err := storage.GetProfilesToInvite(db, remaining)
	if err != nil {
		log.Printf("❌ Failed to fetch profiles: %v", err)
		return
	}

	if len(profiles) == 0 {
		log.Println("⚠️ No profiles available for connection (Run 'search' mode first)")
		return
	}

	log.Printf("Found %d profiles ready for connection", len(profiles))

	// 3. Process Connections
	var successCount = 0

	for i, profileURL := range profiles {
		log.Printf("\n========== Profile %d/%d ==========", i+1, len(profiles))
		
		// Navigate first to get the name
		page.MustNavigate(profileURL)
		page.MustWaitLoad()
		stealth.RandomSleep(3000, 5000)

		// Extract First Name for Personalization
		firstName := "there" // Default fallback
		if nameEl, err := page.Timeout(2 * time.Second).Element("h1"); err == nil {
			text := nameEl.MustText()
			parts := strings.Split(text, " ")
			if len(parts) > 0 {
				firstName = parts[0]
			}
		}

		// Create Personalized Message
		message := strings.ReplaceAll(cfg.ConnectMessageTemplate,"{firstName}",firstName)

		// Attempt to connect (Passing the message now!)
		status, connErr := linkedin.ConnectWithProfile(page, profileURL, message)

		// Update Database based on result
		switch status {
		case "clicked":
			log.Println("✅ Connection request sent")
			successCount++
			storage.UpdateStatus(db, profileURL, "invited")

			// === ☕ NEW: COFFEE BREAK LOGIC ===
            // After every 3 successful invites, take a long break (1-3 minutes)
            if successCount > 0 && successCount%3 == 0 {
                breakTime := 60000 + rand.Intn(120000) // 60s - 180s
                log.Printf("☕ Taking a coffee break for %d seconds (Stealth Protocol)...", breakTime/1000)
                time.Sleep(time.Duration(breakTime) * time.Millisecond)
                continue // Skip the normal safety delay since we just took a long break
            }
            // ==================================
		case "skipped_pending":
			storage.UpdateStatus(db, profileURL, "pending")
		case "skipped_connected":
			storage.UpdateStatus(db, profileURL, "already_connected")
		case "skipped_premium":
			storage.UpdateStatus(db, profileURL, "premium_only")
		case "failed":
			log.Printf("❌ Failed: %v", connErr)
		}

		// Safety Delay
		if i < len(profiles)-1 {
			waitTime := 15000 + rand.Intn(15000) // 15-30s delay
			log.Printf("⏳ Safety delay: %ds...", waitTime/1000)
			stealth.RandomSleep(waitTime, waitTime+1000)
		}
	}

	log.Printf("\n✅ Connect Mode Complete. Sent %d new invites.", successCount)
}

// runDemoMode executes search then connect
func runDemoMode(page *rod.Page, db *sql.DB, cfg *config.Config) {
	log.Println("🎯 Running Demo Sequence...")
	runSearchMode(page, db, cfg)
	
	log.Println("\n⏳ Waiting 10 seconds before connecting...")
	time.Sleep(10 * time.Second)

	runConnectMode(page, db, cfg)
	log.Println("\n✅ Demo sequence completed!")
}

// showFinalStatistics displays comprehensive database statistics
func showFinalStatistics(db *sql.DB, cfg *config.Config) {
	log.Println("\n==========================================")
	log.Println("FINAL DATABASE STATISTICS")
	log.Println("==========================================")

	stats, _ := storage.GetStats(db)
	log.Printf("Total Profiles:          %d", stats.Total)
	log.Printf("├─ Found (ready):        %d", stats.Found)
	log.Printf("├─ Invited (sent):       %d", stats.Invited)
	log.Printf("├─ Connected:            %d", stats.Connected)
	
	// Daily Stats
	todaySearch, _ := guard.GetTodayCount(db)
	todayInvites, _ := guard.GetDailyInviteCount(db)
	
	log.Println("\n📅 Today's Performance:")
	log.Printf("├─ Profiles Collected:   %d / %d", todaySearch, cfg.SearchLimit)
	log.Printf("└─ Invites Sent:         %d / %d", todayInvites, cfg.InviteLimit)
	log.Println("==========================================")
}


// runMessageMode executes the messaging workflow
func runMessageMode(page *rod.Page, db *sql.DB, cfg *config.Config) {
	log.Println("📨 Starting Message Mode...")

	//  DYNAMIC TEMPLATE: Load from Config
	template := cfg.FollowupMessageTemplate

	// Set a safe batch limit (e.g., 10 messages per run)
	// checks 'invited' profiles to see if they accepted
	err := linkedin.SendMessages(page, db, template, 10)
	if err != nil {
		log.Printf("❌ Message mode error: %v", err)
	}

	log.Println("✅ Message Mode Complete.")
}