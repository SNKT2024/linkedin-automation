package browser

import (
	"log"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/stealth"
)

// NewBrowser initializes and returns a Rod browser instance in headful mode.
func NewBrowser() (*rod.Browser, error) {
	log.Println("Initializing browser...")
    url := launcher.New().Headless(false).Leakless(false).MustLaunch()
	browser := rod.New().ControlURL(url).MustConnect()
	log.Println("Browser initialized successfully.")
	return browser, nil
}

func NewStealthPage(browser *rod.Browser) (*rod.Page, error) {
    // Create a new page and apply stealth scripts
    page := stealth.MustPage(browser)
    return page, nil
}