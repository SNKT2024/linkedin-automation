package linkedin

import (
	"errors"
	"log"
	"math/rand"
	"time"

	"github.com/SNKT2024/linkedin-automation/internal/stealth"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

// ConnectWithProfile attempts to send a connection request with an optional note.
func ConnectWithProfile(page *rod.Page, profileURL string, message string) (string, error) {
	log.Printf("Navigating to profile: %s", profileURL)

	page.MustNavigate(profileURL)
	page.MustWaitLoad()

	log.Println("Reading profile...")
	stealth.RandomSleep(3000, 5000)
	stealth.NaturalScroll(page, 300+rand.Intn(200))
	
	// 1. CRITICAL: Only check for "Pending" first. 
	// DO NOT check for "Message" here, or we will skip Open Profiles.
	if exists(page, "button", "Pending") { return "skipped_pending", nil }
	if exists(page, "button", "Withdraw") { return "skipped_pending", nil }
	
	// 2. HUNT FOR CONNECT BUTTON (Priority A: Direct)
	log.Println("Looking for 'Connect' button...")
	var connectBtn *rod.Element
	
	// Try Direct Button
	if btn, err := page.Timeout(3 * time.Second).ElementR("button", "^Connect$"); err == nil {
		connectBtn = btn
		log.Println("✅ Found direct 'Connect' button")
	} else {
		// Try "More" Dropdown (Priority B)
		log.Println("Direct button missing. Checking 'More' dropdown...")
		// Click "More" to open the menu
		if moreBtn, err := page.Timeout(3 * time.Second).ElementR("button", "^More$|More actions"); err == nil {
			stealth.HumanClick(page, moreBtn)
			stealth.RandomSleep(1000, 2000)
			
			// Look for Connect inside the menu
			if dropBtn, err := page.Timeout(3 * time.Second).ElementR("div[role='menuitem'], button, span", "^Connect$"); err == nil {
				connectBtn = dropBtn
				log.Println("✅ Found 'Connect' in dropdown")
			} else {
				// Close dropdown if Connect wasn't found (click body)
				page.Mouse.Click(proto.InputMouseButtonLeft, 1) 
			}
		}
	}

	// 3. IF CONNECT FOUND -> CLICK IT
	if connectBtn != nil {
		// Ensure visibility
		connectBtn.MustScrollIntoView()
		stealth.RandomSleep(500, 1000)

		log.Println("🚀 Clicking 'Connect'...")
		stealth.HumanClick(page, connectBtn)
		stealth.RandomSleep(2000, 3000)

		// Handle the Note/Send Dialog
		handleConnectionDialog(page, message)
		return "clicked", nil
	}

	// 4. IF CONNECT NOT FOUND -> CHECK IF ALREADY CONNECTED
	// Now it is safe to check for "Message", because we confirmed "Connect" is missing.
	if exists(page, "button", "^Message$") {
		log.Println("⚠️ No 'Connect' button, but 'Message' exists -> Already Connected.")
		return "skipped_connected", nil
	}

	// 5. CHECK FOR LOCKED/PREMIUM
	errInMail := rod.Try(func() {
		page.Timeout(2 * time.Second).MustElement(`button[aria-label*="Send InMail"], .premium-inmail-button`)
	})
	if errInMail == nil { return "skipped_premium", nil }

	log.Println("❌ Could not find Connect button (and not connected).")
	return "failed", errors.New("connect button not found")
}

// handleConnectionDialog adds a note if message is provided
func handleConnectionDialog(page *rod.Page, message string) {
	log.Println("Handling connection dialog...")

	// IF message exists, try to click "Add a note"
	if message != "" {
		if noteBtn, err := page.Timeout(3 * time.Second).ElementR("button", "Add a note"); err == nil {
			log.Println("📝 Clicking 'Add a note'...")
			stealth.HumanClick(page, noteBtn)
			stealth.RandomSleep(1000, 2000)

			// Type Message
			if textArea, err := page.Element("textarea"); err == nil {
				// Truncate to 300 chars (LinkedIn Limit)
				if len(message) > 300 { message = message[:300] }
				
				log.Printf("✍️ Typing note: '%s...'", message[:15])
				stealth.HumanType(textArea, message)
				stealth.RandomSleep(1000, 2000)
			}
		} else {
			log.Println("⚠️ 'Add a note' button not found. Sending without note.")
		}
	}

	// Click "Send" (Works for both "Send now" and "Send" after writing note)
	if sendBtn, err := page.Timeout(3 * time.Second).ElementR("button", "Send|Send now|Send without a note"); err == nil {
		log.Println("🚀 Clicking Send...")
		stealth.HumanClick(page, sendBtn)
		stealth.RandomSleep(2000, 3000)
	} else {
		log.Println("⚠️ 'Send' button not found (Email verification might be required)")
	}
}

// Helper to quickly check for element existence by text
func exists(page *rod.Page, selector, textRegex string) bool {
	_, err := page.Timeout(1 * time.Second).ElementR(selector, textRegex)
	return err == nil
}