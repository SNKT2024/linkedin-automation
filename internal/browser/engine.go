package browser

import (
	"log"
	"math/rand"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/stealth"
)

// Common modern User Agents for fingerprint randomization
var userAgents = []string{
	// Windows 10 Chrome
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	// Windows 11 Chrome
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36",
	// Mac OS Chrome
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	// Mac OS Safari
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.1 Safari/605.1.15",
	// Linux Firefox
	"Mozilla/5.0 (X11; Linux x86_64; rv:121.0) Gecko/20100101 Firefox/121.0",
	// Windows Firefox
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:121.0) Gecko/20100101 Firefox/121.0",
}

// Common viewport resolutions
type Viewport struct {
	Width  int
	Height int
}

var viewports = []Viewport{
	{Width: 1920, Height: 1080}, // Full HD
	{Width: 1366, Height: 768},  // Common laptop
	{Width: 1440, Height: 900},  // MacBook Air
	{Width: 1536, Height: 864},  // Surface Pro
	{Width: 1280, Height: 720},  // HD
}

// NewBrowser initializes and returns a Rod browser instance in headful mode with random fingerprinting.
func NewBrowser() (*rod.Browser, error) {
	log.Println("Initializing browser with random fingerprinting...")

	// Select random User Agent
	randomUA := userAgents[rand.Intn(len(userAgents))]
	log.Printf("Selected User Agent: %s", randomUA)

	// Configure launcher with random User Agent and fixed window size
	url := launcher.New().
		Headless(false).
		Leakless(false).
		Set("user-agent", randomUA).
		Set("window-size", "1920,1080"). // Force large physical window to prevent geometry detection
		MustLaunch()

	browser := rod.New().ControlURL(url).MustConnect()
	log.Println("Browser initialized successfully.")
	return browser, nil
}

// NewStealthPage creates a new page with stealth capabilities and random viewport.
func NewStealthPage(browser *rod.Browser) (*rod.Page, error) {
	// Create a new page and apply stealth scripts
	page := stealth.MustPage(browser)

	// Select random viewport
	randomViewport := viewports[rand.Intn(len(viewports))]
	log.Printf("Selected Viewport: %dx%d", randomViewport.Width, randomViewport.Height)

	// Set the viewport
	page.MustSetViewport(randomViewport.Width, randomViewport.Height, 1.0, false)

	return page, nil
}