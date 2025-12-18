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
	"github.com/go-rod/rod/lib/input"
)

// RandomScroll simulates human-like scrolling behavior with reading pauses.
// Now uses NaturalScroll with physics-based wheel movement and RandomWander for mouse jitter.
func RandomScroll(page *rod.Page) {
	// Number of scroll actions (4-7 times for natural behavior)
	scrollCount := 4 + rand.Intn(4)
	log.Printf("Starting natural scroll sequence (%d scrolls)...", scrollCount)

	for i := 0; i < scrollCount; i++ {
		// 30% chance to scroll up slightly (simulating re-reading)
		if rand.Float64() < 0.3 {
			scrollUpAmount := 100 + rand.Intn(200) // 100-300 pixels up
			log.Printf("Scroll %d/%d: Scrolling UP %d pixels (re-reading)", i+1, scrollCount, scrollUpAmount)
			stealth.NaturalScroll(page, -scrollUpAmount) // Use physics-based scroll
			stealth.RandomSleep(1000, 2000)              // Longer pause while re-reading
		}

		// Scroll down by random amount (200-600 pixels)
		scrollAmount := 200 + rand.Intn(400) // 200-600 pixels down
		log.Printf("Scroll %d/%d: Scrolling DOWN %d pixels", i+1, scrollCount, scrollAmount)
		stealth.NaturalScroll(page, scrollAmount) // Use physics-based scroll

		// 40% chance to wander mouse while scrolling (checking content)
		if rand.Float64() < 0.4 {
			log.Println("Mouse wandering while reading...")
			stealth.RandomWander(page)
		}

		// Simulate reading time after scrolling (1500-3500ms)
		readingTime := 1500 + rand.Intn(2000) // 1500-3500ms
		log.Printf("Reading for %dms...", readingTime)
		stealth.RandomSleep(readingTime, readingTime+500)

		// Additional 20% chance for a second wander (user getting distracted)
		if rand.Float64() < 0.2 {
			log.Println("Additional mouse movement (distracted reading)...")
			stealth.RandomWander(page)
		}
	}
	log.Println("Natural scroll sequence completed.")
}

// humanTypeWithMistakes types text with realistic typos and corrections
// Updated with slower, more natural typing speeds (approx 60 WPM)
func humanTypeWithMistakes(element *rod.Element, text string) {
	log.Printf("Typing with human-like behavior: '%s'", text)

	// 30% chance to make a typo somewhere in the text
	if rand.Float64() < 0.3 {
		// Pick a random position to make the typo (not at the very end)
		typoPos := rand.Intn(len(text) - 1)

		// Type up to the typo position with human-like delays
		for i := 0; i < typoPos; i++ {
			element.MustInput(string(text[i]))
			// Slower typing: 120-300ms per character (approx 60 WPM)
			stealth.RandomSleep(120, 300)
		}

		// Make a typo - type a random wrong character
		wrongChars := "qwertyuiopasdfghjklzxcvbnm"
		typoChar := string(wrongChars[rand.Intn(len(wrongChars))])
		log.Printf("Making typo at position %d: typing '%s' instead of '%s'", typoPos, typoChar, string(text[typoPos]))
		element.MustInput(typoChar)
		stealth.RandomSleep(120, 300)

		// Pause - user notices the mistake (longer thinking pause)
		stealth.RandomSleep(1000, 2000)

		// Backspace to delete the typo
		log.Println("Correcting typo with backspace...")
		element.MustType(input.Backspace)
		stealth.RandomSleep(200, 400)

		// Continue typing correctly from the typo position
		for i := typoPos; i < len(text); i++ {
			element.MustInput(string(text[i]))
			stealth.RandomSleep(120, 300)
		}
	} else {
		// No typo - just type slowly with human delays
		for _, char := range text {
			element.MustInput(string(char))
			// Slower typing: 120-300ms per character
			stealth.RandomSleep(120, 300)
		}
	}

	// Longer thinking pause after typing (1-2 seconds)
	log.Println("Pausing after typing (thinking)...")
	stealth.RandomSleep(1000, 2000)
}

// SearchPeople searches for people on LinkedIn by keyword and returns their profile URLs.
// It handles pagination and checks the database to avoid duplicates.
func SearchPeople(page *rod.Page, db *sql.DB, keyword string) ([]string, error) {
	log.Printf("Searching for people with keyword: '%s'", keyword)

	// Check if already on feed, if not navigate there
	currentURL := page.MustInfo().URL
	if !strings.Contains(currentURL, "/feed/") {
		log.Println("Not on feed page, navigating to LinkedIn feed...")
		page.MustNavigate("https://www.linkedin.com/feed/")
		page.MustWaitLoad()
		stealth.RandomSleep(2000, 4000)
	} else {
		log.Println("Already on feed page, proceeding with search...")
		stealth.RandomSleep(1000, 2000) // Brief pause to simulate looking at feed
	}

	// Wander mouse before looking for search bar (simulate scanning the page)
	log.Println("Scanning the page before searching...")
	stealth.RandomWander(page)

	// Find the search bar (using robust selector)
	log.Println("Looking for search bar...")
	stealth.RandomSleep(800, 1500) // Pause before searching

	// Use primary robust selector
	searchInput := page.MustElement("input.search-global-typeahead__input, input[placeholder*='Search'], input[aria-label*='Search']")

	// Get search bar position
	box := searchInput.MustEval(`() => {
        const rect = this.getBoundingClientRect();
        return { x: rect.x, y: rect.y, width: rect.width, height: rect.height };
    }`).Val().(map[string]interface{})

	searchX := box["x"].(float64) + box["width"].(float64)/2
	searchY := box["y"].(float64) + box["height"].(float64)/2

	// Move mouse to search bar naturally
	log.Println("Moving mouse to search bar...")
	stealth.MoveMouseSmoothly(page, searchX, searchY)
	stealth.RandomSleep(300, 700)

	// Click the search bar to focus it
	log.Println("Clicking search bar...")
	page.Mouse.MustClick("left")
	stealth.RandomSleep(500, 1000)

	// Type the keyword with realistic mistakes and slower typing
	log.Printf("Typing search keyword with human-like behavior...")
	humanTypeWithMistakes(searchInput, keyword)

	// Pause after typing (already included in humanTypeWithMistakes)
	// Additional thinking pause before pressing enter
	stealth.RandomSleep(500, 1000)

	// Press Enter
	log.Println("Pressing Enter to search...")
	searchInput.MustType(input.Enter)
	page.MustWaitLoad()
	stealth.RandomSleep(2000, 3500)

	// Self-healing navigation: Check if we're on people results page
	currentURL = page.MustInfo().URL
	if !strings.Contains(currentURL, "/search/results/people/") {
		log.Println("Not on people results page. Attempting to click 'People' filter...")

		// Wander before clicking the People filter (simulate looking for it)
		log.Println("Looking for People filter...")
		stealth.RandomWander(page)

		// Try to find and click the People button using text content (more robust)
		errPeople := rod.Try(func() {
			// Use ElementR (Regex) to find button by text content
			peopleButton := page.Timeout(3 * time.Second).MustElementR("button", "People")

			// Get button position
			buttonBox := peopleButton.MustEval(`() => {
                const rect = this.getBoundingClientRect();
                return { x: rect.x, y: rect.y, width: rect.width, height: rect.height };
            }`).Val().(map[string]interface{})

			peopleX := buttonBox["x"].(float64) + buttonBox["width"].(float64)/2
			peopleY := buttonBox["y"].(float64) + buttonBox["height"].(float64)/2

			// Move mouse to People button
			log.Println("Moving to 'People' filter...")
			stealth.MoveMouseSmoothly(page, peopleX, peopleY)
			stealth.RandomSleep(300, 700)

			// Click People filter
			log.Println("Clicking 'People' filter...")
			page.Mouse.MustClick("left")
			page.MustWaitLoad()
			stealth.RandomSleep(2000, 3000)
		})

		// Fallback: Force navigation if clicking failed
		if errPeople != nil {
			log.Println("⚠️ Could not click People filter. Using fallback navigation...")
			fallbackURL := "https://www.linkedin.com/search/results/people/?keywords=" + url.QueryEscape(keyword)
			log.Printf("Navigating directly to: %s", fallbackURL)
			page.MustNavigate(fallbackURL)
			page.MustWaitLoad()
			stealth.RandomSleep(2000, 3500)
		}
	} else {
		log.Println("Already on people search results page")
	}

	// Verify we're on the correct page
	finalURL := page.MustInfo().URL
	if !strings.Contains(finalURL, "/search/results/people/") {
		log.Println("❌ Failed to navigate to people results. Aborting search.")
		return []string{}, nil
	}

	// Initialize results slice
	newProfiles := []string{}
	maxPages := 3 // Limit to 3 pages for safety

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

		// Perform human-like scrolling with natural physics and mouse wandering
		log.Println("Scrolling to load lazy-loaded content with natural behavior...")
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

			// Check database for duplicates - IsProfileVisited returns bool only
			if storage.IsProfileVisited(db, href) {
				log.Printf("⏭️  Skipping duplicate: %s", href)
				pageSkipCount++
				continue
			}

			// New profile found!
			log.Printf("✅ New profile: %s", href)
			newProfiles = append(newProfiles, href)
			pageNewCount++

			// Add to database - AddProfile takes (db, url) only
			if err := storage.AddProfile(db, href); err != nil {
				log.Printf("⚠️ Failed to save profile to database: %v", err)
			}
		}

		log.Printf("Page %d Summary: %d new, %d duplicates, %d total unique on page",
			pageNum, pageNewCount, pageSkipCount, len(pageURLs))

		// Look for Next button
		if pageNum < maxPages {
			log.Println("Looking for 'Next' button...")

			// Wander mouse before clicking Next (simulate scanning for pagination)
			if rand.Float64() < 0.5 {
				log.Println("Scanning page for Next button...")
				stealth.RandomWander(page)
			}

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

				// Scroll to the Next button naturally
				log.Println("Scrolling to Next button...")
				nextButton.MustScrollIntoView()
				stealth.RandomSleep(800, 1500)

				// Wander before clicking Next button (final check)
				log.Println("Preparing to click Next...")
				stealth.RandomWander(page)

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
	log.Println("Scrolling to bottom of page with natural physics...")

	for {
		// Get current scroll position
		scrollHeight := page.MustEval(`() => document.documentElement.scrollHeight`).Int()
		currentScroll := page.MustEval(`() => window.pageYOffset`).Int()

		// Scroll down by a chunk (400-700 pixels)
		scrollAmount := 400 + rand.Intn(300)
		stealth.NaturalScroll(page, scrollAmount) // Use physics-based scroll

		// 30% chance to wander while scrolling to bottom
		if rand.Float64() < 0.3 {
			log.Println("Wandering while scrolling...")
			stealth.RandomWander(page)
		}
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