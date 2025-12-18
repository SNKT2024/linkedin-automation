package linkedin

import (
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/SNKT2024/linkedin-automation/internal/stealth"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/input"
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
func SearchPeople(page *rod.Page, keyword string) ([]string, error) {
	log.Printf("Searching for people with keyword: %s", keyword)

	// Simulate thinking before searching
	log.Println("Pausing before starting search...")
	stealth.RandomSleep(2000, 4000)

	// Navigate to LinkedIn search page (not directly to results)
	log.Println("Navigating to LinkedIn homepage...")
	page.MustNavigate("https://www.linkedin.com")
	page.MustWaitLoad()
	stealth.RandomSleep(1500, 2500)

	// Find and click the search bar
	log.Println("Looking for search bar...")
	stealth.RandomSleep(800, 1500)

	// Try to find the search input
	searchInput := page.MustElement("input[placeholder*='Search'], input.search-global-typeahead__input, input[aria-label*='Search']")

	// Move mouse to search bar and click it
	box := searchInput.MustEval(`() => {
		const rect = this.getBoundingClientRect();
		return { x: rect.x, y: rect.y, width: rect.width, height: rect.height };
	}`).Val().(map[string]interface{})

	searchX := box["x"].(float64) + box["width"].(float64)/2
	searchY := box["y"].(float64) + box["height"].(float64)/2

	log.Println("Moving mouse to search bar...")
	stealth.MoveMouseSmoothly(page, searchX, searchY)
	stealth.RandomSleep(300, 600)

	log.Println("Clicking search bar...")
	page.Mouse.MustClick("left")
	stealth.RandomSleep(500, 1000)

	// Type the search keyword with human-like typing
	log.Printf("Typing search keyword: %s", keyword)
	stealth.HumanType(searchInput, keyword)
	stealth.RandomSleep(800, 1500)

	// Press Enter to search
	log.Println("Pressing Enter to search...")
	searchInput.MustType(input.Enter) // Enter key
	page.MustWaitLoad()
	stealth.RandomSleep(2000, 3500)

	// Click on "People" filter if not already filtered
	log.Println("Looking for People filter...")
	errPeople := rod.Try(func() {
		peopleButton := page.Timeout(5 * time.Second).MustElement("button[aria-label*='People'], button:has-text('People')")

		// Check if already on people results
		currentURL := page.MustInfo().URL
		if !strings.Contains(currentURL, "/search/results/people/") {
			log.Println("Clicking People filter...")

			// Move mouse to People button
			buttonBox := peopleButton.MustEval(`() => {
				const rect = this.getBoundingClientRect();
				return { x: rect.x, y: rect.y, width: rect.width, height: rect.height };
			}`).Val().(map[string]interface{})

			peopleX := buttonBox["x"].(float64) + buttonBox["width"].(float64)/2
			peopleY := buttonBox["y"].(float64) + buttonBox["height"].(float64)/2

			stealth.MoveMouseSmoothly(page, peopleX, peopleY)
			stealth.RandomSleep(300, 700)

			peopleButton.MustClick()
			page.MustWaitLoad()
			stealth.RandomSleep(2000, 3000)
		}
	})
	if errPeople != nil {
		log.Println("Could not find People filter, assuming already on people results page")
	}

	// Wait for results to load
	log.Println("Waiting for search results to load...")
	stealth.RandomSleep(2000, 3500)

	// Wait for results container with timeout
	errResults := rod.Try(func() {
		page.Timeout(10 * time.Second).MustElement(".reusable-search__result-container, .search-results-container, ul.reusable-search__entity-result-list")
	})
	if errResults != nil {
		log.Println("No search results found or timeout waiting for results.")
		return []string{}, fmt.Errorf("search results not found: %w", errResults)
	}

	// Check if results exist
	hasResults := page.MustHas(".reusable-search__result-container") ||
		page.MustHas(".search-results-container") ||
		page.MustHas("ul.reusable-search__entity-result-list")

	if !hasResults {
		log.Println("No search results found.")
		return []string{}, nil
	}

	// Perform human-like scrolling to load lazy content
	log.Println("Performing human-like scrolling to load all results...")
	RandomScroll(page)

	// Additional wait to ensure all lazy-loaded content appears
	log.Println("Waiting for all content to load...")
	stealth.RandomSleep(2000, 3500)

	// Extract profile URLs
	log.Println("Extracting profile URLs...")
	profileURLs := make(map[string]bool) // Use map to ensure uniqueness

	// Find all links in the search results
	links := page.MustElements("a[href*='/in/']")
	log.Printf("Found %d potential profile links", len(links))

	for _, link := range links {
		href := link.MustProperty("href").Str()

		// Validate that it's a profile URL
		if strings.Contains(href, "/in/") &&
			!strings.Contains(href, "/search/") &&
			!strings.Contains(href, "/company/") &&
			!strings.Contains(href, "/jobs/") {

			// Clean up the URL (remove query parameters)
			if idx := strings.Index(href, "?"); idx != -1 {
				href = href[:idx]
			}

			// Ensure it ends with / for consistency
			if !strings.HasSuffix(href, "/") {
				href += "/"
			}

			// Add to map for deduplication
			profileURLs[href] = true
		}
	}

	// Convert map to slice
	uniqueURLs := make([]string, 0, len(profileURLs))
	for profileURL := range profileURLs {
		uniqueURLs = append(uniqueURLs, profileURL)
	}

	log.Printf("Extracted %d unique profile URLs", len(uniqueURLs))
	return uniqueURLs, nil
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