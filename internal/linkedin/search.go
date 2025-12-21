package linkedin

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/SNKT2024/linkedin-automation/internal/stealth"
	"github.com/SNKT2024/linkedin-automation/internal/storage"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/input"
)

// SearchPeople orchestrates the search workflow
func SearchPeople(page *rod.Page, db *sql.DB, keyword string, maxPages int) ([]string, error) {
	log.Printf("🔍 Searching for people with keyword: '%s'", keyword)

	// === CRITICAL FIX: Wait for Feed to Settle ===
	// This prevents the bot from checking for the search bar 
	// while the page is still white/loading after login.
	log.Println("   ⏳ Waiting for feed to render...")
	page.MustWaitLoad()
	stealth.RandomSleep(3000, 5000)
	// =============================================

	// 1. Navigation (Safety check)
	if !strings.Contains(page.MustInfo().URL, "/feed/") {
		log.Println("   🔄 Navigating to Feed...")
		page.MustNavigate("https://www.linkedin.com/feed/")
		page.MustWaitLoad()
		stealth.RandomSleep(3000, 5000)
	}

	// 2. Search Bar (Safe Find Pattern)
	log.Println("🔍 Looking for search bar...")
	
	// We check for multiple possible selectors to be robust
	searchSelectors := []string{"input.search-global-typeahead__input", "input[placeholder*='Search']"}
	var searchInput *rod.Element
	var found bool

	// Try finding it for up to 10 seconds
	for i := 0; i < 5; i++ {
		for _, sel := range searchSelectors {
			if has, _, _ := page.Has(sel); has {
				searchInput = page.MustElement(sel)
				found = true
				break
			}
		}
		if found { break }
		time.Sleep(2 * time.Second)
	}

	if !found {
		return nil, fmt.Errorf("could not find search bar within 10s")
	}

	// Safe Typing Logic
	searchInput.MustClick()
	stealth.RandomSleep(500, 1000)
	humanTypeWithMistakes(searchInput, keyword)
	
	log.Println("⌨️ Pressing Enter...")
	searchInput.MustType(input.Enter)
	page.MustWaitLoad()
	stealth.RandomSleep(4000, 6000)

	// 3. People Filter
	// Only click if we aren't already on the people tab
	if !strings.Contains(page.MustInfo().URL, "/people/") {
		log.Println("👥 Checking 'People' filter...")
		
		// Try finding the button by text "People"
		if found, _, _ := page.Timeout(5 * time.Second).HasR("button", "People"); found {
			btn := page.MustElementR("button", "People")
			// Only click if not already active (pressed)
			if pressed, _ := btn.Attribute("aria-pressed"); pressed == nil || *pressed != "true" {
				btn.MustClick()
				page.MustWaitLoad()
				stealth.RandomSleep(3000, 5000)
			}
		}
	}

	var newProfiles []string

	for pageNum := 1; pageNum <= maxPages; pageNum++ {
		log.Printf("\n========== Page %d/%d ==========", pageNum, maxPages)

		// 4. Check for Blocking Modals (Safe Check)
		if found, _, _ := page.Timeout(2 * time.Second).HasR("button", "Got it|Close"); found {
			log.Println("⚠️ Dismissing blocking modal...")
			page.MustElementR("button", "Got it|Close").MustClick()
			stealth.RandomSleep(1000, 2000)
		}

		// 5. Smart Scroll
		log.Println("📜 Scrolling to load results...")
		SmartScroll(page)

		// 6. Extraction
		log.Println("📥 Scanning page for profile links...")
		elements, err := page.Elements("a")
		if err != nil {
			log.Printf("❌ Error scanning page: %v", err)
			continue
		}

		count := 0
		uniqueOnPage := make(map[string]bool)

		for _, el := range elements {
			link, err := el.Property("href")
			if err != nil { continue }
			urlStr := link.String()

			if strings.Contains(urlStr, "linkedin.com/in/") && 
			   !strings.Contains(urlStr, "/minis/") &&
			   !strings.Contains(urlStr, "google.com") {
				
				if idx := strings.Index(urlStr, "?"); idx != -1 { urlStr = urlStr[:idx] }
				if uniqueOnPage[urlStr] { continue }
				uniqueOnPage[urlStr] = true
				
				// Skip yourself if needed (optional)
				// if strings.Contains(urlStr, "sanket-kumbhar") { continue }

				added, _ := storage.AddProfile(db, urlStr)
				if added {
					newProfiles = append(newProfiles, urlStr)
					count++
				}
			}
		}
		log.Printf("💾 Saved %d NEW profiles from this page", count)

		// 7. Pagination (Next Button)
		if pageNum < maxPages {
			log.Println("➡️ Looking for 'Next' button...")
			
			// Try Primary Selector (Desktop)
			if found, _, _ := page.Timeout(3 * time.Second).Has(`button[aria-label="Next"]`); found {
				nextBtn := page.MustElement(`button[aria-label="Next"]`)
				clickNext(page, nextBtn)
			} else {
				// Fallback Text Selector
				if foundFallback, _, _ := page.Timeout(2 * time.Second).HasR("button, span", "^Next$"); foundFallback {
					nextBtn := page.MustElementR("button, span", "^Next$")
					clickNext(page, nextBtn)
				} else {
					log.Println("🛑 No 'Next' button found. End of search.")
					break
				}
			}
		}
	}
	return newProfiles, nil
}

// Helper to safely click next
func clickNext(page *rod.Page, btn *rod.Element) {
	// Check visibility before scrolling
	if visible, _ := btn.Visible(); !visible {
		log.Println("⚠️ Next button found but hidden.")
		return
	}
	
	btn.MustScrollIntoView()
	stealth.RandomSleep(500, 1000)
	
	log.Println("👆 Clicking Next...")
	stealth.HumanClick(page, btn)
	page.MustWaitLoad()
	stealth.RandomSleep(4000, 6000)
}

func humanTypeWithMistakes(element *rod.Element, text string) {
	log.Printf("⌨️ Typing: '%s'", text)
	for _, char := range text {
		element.MustInput(string(char))
		stealth.RandomSleep(80, 200)
	}
	stealth.RandomSleep(500, 1000)
}

func SmartScroll(page *rod.Page) {
	// Scroll using NaturalScroll (Center mouse first)
	for i := 0; i < 5; i++ {
		stealth.NaturalScroll(page, 400)
		stealth.RandomSleep(800, 1200)
	}
	// Final JS nudge to ensure we hit the footer
	page.MustEval(`() => window.scrollTo({ top: document.body.scrollHeight, behavior: 'smooth' })`)
	stealth.RandomSleep(2000, 3000)
}