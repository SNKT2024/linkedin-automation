package linkedin

import (
	"database/sql"
	"log"
	"strings"
	"time"

	"github.com/SNKT2024/linkedin-automation/internal/stealth"
	"github.com/SNKT2024/linkedin-automation/internal/storage"
	"github.com/go-rod/rod"
)

// SendMessages checks profiles and sends a welcome message if connected
func SendMessages(page *rod.Page, db *sql.DB, messageTemplate string, limit int) error {
	log.Println("📨 Starting Messaging Service...")

	// 1. Get profiles
	profiles, err := storage.GetProfilesByStatus(db, "invited", limit)
	if err != nil { return err }

	if len(profiles) == 0 {
		log.Println("⚠️ No 'invited' profiles found to check. Run 'connect' mode first.")
		return nil
	}

	log.Printf("Found %d invited profiles to check for acceptance", len(profiles))

	sentCount := 0

	for _, profileURL := range profiles {
		if sentCount >= limit {
			log.Println("🛑 Message session limit reached.")
			break
		}

		log.Printf("👉 Checking status for: %s", profileURL)

		// Navigate
		page.MustNavigate(profileURL)
		page.MustWaitLoad()
		stealth.RandomSleep(3000, 5000)

		// 2. DETECT CONNECTION STATUS
		// Use Timeout for detection only
		msgBtnSelector := "button, a"
		foundMsgBtn, _, _ := page.Timeout(3 * time.Second).HasR(msgBtnSelector, "^Message$")
		
		if !foundMsgBtn {
			if foundPending, _, _ := page.Timeout(2 * time.Second).HasR("button", "Pending|Withdraw"); foundPending {
				log.Println("   ⏳ Still Pending. Skipping.")
				storage.UpdateStatus(db, profileURL, "pending")
			} else {
				log.Println("   ❌ Not connected (No 'Message' button). Skipping.")
			}
			continue
		}

		// Grab the button safely (without timeout) to check attributes/click
		// We use Elements() and filter manually to be safe, or just FindR if confident
		msgBtn, err := page.ElementR(msgBtnSelector, "^Message$")
		if err != nil { continue }

		// Check for locked Premium InMail icon
		if lockIcon, _ := msgBtn.Element("svg[data-test-icon='lock-small']"); lockIcon != nil {
			log.Println("   🔒 Message button is locked (Premium only). Skipping.")
			storage.UpdateStatus(db, profileURL, "premium_only")
			continue
		}

		log.Println("   ✅ Message button found. Clicking...")
		stealth.HumanClick(page, msgBtn)
		stealth.RandomSleep(2000, 3000)

		// 3. PRIORITY CHECK: DID THE CHAT BOX OPEN?
		// Selector for the chat box
		chatSelector := "div[role='textbox'][aria-label*='Write a message']"
		
		// Wait up to 5 seconds for it to appear
		if found, _, _ := page.Timeout(5 * time.Second).Has(chatSelector); found {
			// === SUCCESS PATH: CHAT IS OPEN ===
			log.Println("   ✅ Chat input found! Connection active.")
			
			// CRITICAL FIX: Grab the element using the original 'page' (no timeout)
			// This prevents the "Context Deadline Exceeded" panic while typing
			chatBox := page.MustElement(chatSelector)

			// Personalize
			firstName := "there"
			if nameEl, err := page.Timeout(2 * time.Second).Element("h1"); err == nil {
				text := nameEl.MustText()
				parts := strings.Split(text, " ")
				if len(parts) > 0 { firstName = parts[0] }
			}
			finalMsg := strings.ReplaceAll(messageTemplate, "{firstName}", firstName)

			// Type & Send (Now safe from timeouts)
			log.Printf("   ✍️ Typing: '%s...'", finalMsg)
			stealth.HumanType(chatBox, finalMsg)
			stealth.RandomSleep(2000, 3000)

			// Find Send Button
			if sendBtn, err := page.Timeout(3 * time.Second).Element("button[type='submit']"); err == nil {
				log.Println("   🚀 Clicking Send...")
				stealth.HumanClick(page, sendBtn)
				stealth.RandomSleep(2000, 3000)
				
				storage.UpdateStatus(db, profileURL, "messaged")
				log.Println("   ✅ Message sent & DB updated.")
				sentCount++

				// === ☕ NEW: COFFEE BREAK LOGIC ===
            // After every 3 messages, take a break
            if sentCount > 0 && sentCount%3 == 0 {
                log.Println("   ☕ Taking a short break to mimic human behavior...")
                stealth.RandomSleep(45000, 90000) // 45s - 90s
                continue
            }
            // ==================================
			} else {
				log.Println("   ⚠️ Could not find Send button.")
			}

			closeChat(page)

		} else {
			// === FAILURE PATH: CHAT DID NOT OPEN ===
			log.Println("   ⚠️ Chat box did not appear. Checking for Premium Popup...")
			
			// Check for popup (Wait 2s)
			popupSelector := "div[role='dialog'], div.artdeco-modal"
			if foundPopup, _, _ := page.Timeout(2 * time.Second).HasR(popupSelector, "Message with Premium|Try Premium|Unlock InMail"); foundPopup {
				log.Println("   🛑 Blocked by Premium/InMail Popup. (Not fully connected).")
				
				// Close popup
				if closeBtn, err := page.Timeout(2 * time.Second).Element(`button[aria-label="Dismiss"], button[aria-label="Close"]`); err == nil {
					closeBtn.MustClick()
				} else {
					page.Keyboard.Press(27) // Escape
				}
				
				storage.UpdateStatus(db, profileURL, "pending")
			} else {
				log.Println("   ❌ Unknown state: Clicked message but no chat and no popup.")
			}
		}

		log.Println("   ❄️ Cooling down...")
		stealth.RandomSleep(5000, 10000)
	}

	return nil
}

// Helper to close chat windows
func closeChat(page *rod.Page) {
	if closeBtn, err := page.Timeout(2 * time.Second).Element(`button[aria-label*="Close"]`); err == nil {
		if visible, _ := closeBtn.Visible(); visible {
			closeBtn.MustClick()
		}
	}
}