package linkedin

import (
	"database/sql"
	"log"
	"math/rand"
	"net/url"
	"strings"
	"time"

	"github.com/SNKT2024/linkedin-automation/internal/stealth"
	"github.com/SNKT2024/linkedin-automation/internal/storage"
	"github.com/go-rod/rod"
)

// RandomScroll simulates human-like scrolling behavior with reading pauses.
func RandomScroll(page *rod.Page) {
	// Number of scroll actions (5-10 times for more natural behavior)
	scrollCount := 5 + rand.Intn(6) // 5-10 scrolls
	log.Printf("Starting random scroll sequence (%d scrolls)...", scrollCount)

	for i := 0; i < scrollCount; i++ {
		// 30% chance to scroll up slightly (simulating re-reading)
		if rand.Float64() < 0.3 {
			scrollUpAmount := 100 + rand.Intn(200) // 100-300 pixels up
			log.Printf("Scroll %d/%d: Scrolling UP %d pixels (re-reading)", i+1, scrollCount, scrollUpAmount)
			page.MustEval(`(amount) => window.scrollBy(0, -amount)`, scrollUpAmount)
			stealth.RandomSleep(1000, 2000) // Longer pause while re-reading
		}

		// Scroll down by random amount (200-600 pixels)
		scrollAmount := 200 + rand.Intn(400) // 200-600 pixels down
		log.Printf("Scroll %d/%d: Scrolling DOWN %d pixels", i+1, scrollCount, scrollAmount)
		page.MustEval(`(amount) => window.scrollBy(0, amount)`, scrollAmount)

		// Simulate reading time after scrolling (1500-3500ms)
		readingTime := 1500 + rand.Intn(2000) // 1500-3500ms
		log.Printf("Reading for %dms...", readingTime)
		stealth.RandomSleep(readingTime, readingTime+500)
	}
	log.Println("Random scroll sequence completed.")
}

// SearchPeople searches for people on LinkedIn by keyword and returns their profile URLs.
// It handles pagination and checks the database to avoid duplicates.
func SearchPeople(page *rod.Page, db *sql.DB, keyword string) ([]string, error) {
	log.Printf("Searching for people with keyword: '%s'", keyword)

	// Build search URL for people results
	searchURL := "https://www.linkedin.com/search/results/people/?keywords=" + url.QueryEscape(keyword)
	log.Printf("Navigating to: %s", searchURL)

	// Navigate to search results
	page.MustNavigate(searchURL)
	page.MustWaitLoad()
	stealth.RandomSleep(2000, 3500)

	// Initialize results slice
	newProfiles := []string{}
	maxPages := 5 // Limit to 5 pages for safety

	// Pagination loop
	for pageNum := 1; pageNum <= maxPages; pageNum++ {
		log.Printf("\n========== Page %d/%d ==========", pageNum, maxPages)

		// Wait for results container with timeout
		log.Println("Waiting for search results to load...")
		errResults := rod.Try(func() {
			page.Timeout(10 * time.Second).MustElement(
				".reusable-search__result-container, .search-results-container, ul.reusable-search__entity-result-list",
			)
		})

		if errResults != nil {
			log.Printf("⚠️ No search results found or timeout on page %d", pageNum)
			break
		}

		// Check if results exist
		hasResults := page.MustHas(".reusable-search__result-container") ||
			page.MustHas(".search-results-container") ||
			page.MustHas("ul.reusable-search__entity-result-list")

		if !hasResults {
			log.Printf("⚠️ No search results container found on page %d", pageNum)
			break
		}

		// Perform human-like scrolling to load lazy content
		log.Println("Scrolling to load lazy-loaded content...")
		RandomScroll(page)

		// Additional wait to ensure all content is loaded
		log.Println("Waiting for all content to render...")
		stealth.RandomSleep(2000, 3500)

		// Extract profile URLs
		log.Println("Extracting profile links...")
		links := page.MustElements("a[href*='/in/']")
		log.Printf("Found %d potential profile links", len(links))

		// Track unique URLs on this page
		pageURLs := make(map[string]bool)
		pageNewCount := 0
		pageSkipCount := 0

		for _, link := range links {
			href := link.MustProperty("href").Str()

			// Validate that it's a profile URL (not company, jobs, etc.)
			if !strings.Contains(href, "/in/") ||
				strings.Contains(href, "/search/") ||
				strings.Contains(href, "/company/") ||
				strings.Contains(href, "/jobs/") ||
				strings.Contains(href, "/posts/") {
				continue
			}

			// Clean up the URL (remove query parameters and fragments)
			if idx := strings.Index(href, "?"); idx != -1 {
				href = href[:idx]
			}
			if idx := strings.Index(href, "#"); idx != -1 {
				href = href[:idx]
			}

			// Ensure it ends with / for consistency
			if !strings.HasSuffix(href, "/") {
				href += "/"
			}

			// Skip if we've already seen this URL on this page
			if pageURLs[href] {
				continue
			}
			pageURLs[href] = true

			// Check database for duplicates
			if storage.IsProfileVisited(db, href) {
				log.Printf("⏭️  Skipping duplicate: %s", href)
				pageSkipCount++
				continue
			}

			// New profile found!
			log.Printf("✅ New profile: %s", href)
			newProfiles = append(newProfiles, href)
			pageNewCount++

			// Add to database
			if err := storage.AddProfile(db, href); err != nil {
				log.Printf("⚠️ Failed to save profile to database: %v", err)
			}
		}

		log.Printf("Page %d Summary: %d new, %d duplicates, %d total unique on page",
			pageNum, pageNewCount, pageSkipCount, len(pageURLs))

		// Look for Next button
		log.Println("Looking for 'Next' button...")
		hasNext := false

		errNext := rod.Try(func() {
			// Try multiple selectors for the Next button
			nextButton := page.Timeout(5 * time.Second).MustElement(
				`button[aria-label*="Next"], button.artdeco-pagination__button--next`,
			)

			// Check if button is disabled
			isDisabled := nextButton.MustProperty("disabled").Bool()
			if isDisabled {
				log.Println("Next button is disabled (last page reached)")
				return
			}

			// Check if button has 'disabled' class
			classList := nextButton.MustProperty("className").Str()
			if strings.Contains(classList, "artdeco-button--disabled") {
				log.Println("Next button has disabled class (last page reached)")
				return
			}

			hasNext = true

			// Scroll to the Next button
			log.Println("Scrolling to Next button...")
			nextButton.MustScrollIntoView()
			stealth.RandomSleep(800, 1500)

			// Use stealth click
			log.Println("Clicking Next button...")
			stealth.HumanClick(page, nextButton)

			// Wait for new page to load
			log.Println("Waiting for next page to load...")
			page.MustWaitLoad()
			stealth.RandomSleep(3000, 5000)
		})

		if errNext != nil || !hasNext {
			log.Println("No more pages available (Next button not found or disabled)")
			break
		}
	}

	// Final summary
	log.Printf("\n========================================")
	log.Printf("✅ Search completed!")
	log.Printf("📊 Total NEW profiles found: %d", len(newProfiles))
	log.Printf("========================================\n")

	return newProfiles, nil
}

// ScrollToBottom scrolls to the bottom of the page gradually with human-like behavior.
func ScrollToBottom(page *rod.Page) {
	log.Println("Scrolling to bottom of page...")

	for {
		// Get current scroll position
		scrollHeight := page.MustEval(`() => document.documentElement.scrollHeight`).Int()
		currentScroll := page.MustEval(`() => window.pageYOffset`).Int()

		// Scroll down by a chunk
		scrollAmount := 400 + rand.Intn(300) // 400-700 pixels
		page.MustEval(`(amount) => window.scrollBy(0, amount)`, scrollAmount)

		// Wait for content to load
		stealth.RandomSleep(800, 1500)

		// Check if we've reached the bottom
		newScroll := page.MustEval(`() => window.pageYOffset`).Int()
		if newScroll+page.MustEval(`() => window.innerHeight`).Int() >= scrollHeight-100 {
			log.Println("Reached bottom of page")
			break
		}

		// Safety check to avoid infinite loop
		if newScroll == currentScroll {
			log.Println("No more scrolling possible")
			break
		}
	}
}