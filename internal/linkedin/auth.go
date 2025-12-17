package linkedin

import (
	"encoding/json"
	"errors"
	"log"
	"os"

	"github.com/SNKT2024/linkedin-automation/internal/stealth"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

const cookiesFile = "cookies.json"

// Login handles LinkedIn authentication, using cookies if available or logging in manually.
func Login(browser *rod.Browser, page *rod.Page) error {
	// Attempt to load cookies from cookies.json
	if err := loadCookies(browser); err == nil {
		log.Println("Cookies loaded successfully. Refreshing the page...")
		page.MustReload()
		if page.MustHas("#global-nav") {
			log.Println("Logged in using cookies.")
			return nil
		}
		log.Println("Cookies invalid or expired. Proceeding with manual login.")
	}

	// Navigate to LinkedIn login page
	log.Println("Navigating to LinkedIn login page...")
	page.MustNavigate("https://www.linkedin.com/login")

	// Simulate user looking at the page before typing
	stealth.RandomSleep(1000, 3000)

	// Find and fill the username field
	log.Println("Filling in email...")
	email := os.Getenv("LINKEDIN_EMAIL")
	password := os.Getenv("LINKEDIN_PASSWORD")

	usernameField := page.MustElement("#username")
	stealth.HumanType(usernameField, email)

	// Simulate checking if email is correct
	log.Println("Pausing after email entry...")
	stealth.RandomSleep(800, 1500)

	// Move mouse smoothly to password field
	log.Println("Moving to password field...")
	passwordField := page.MustElement("#password")

	// Get password field position
	box := passwordField.MustEval(`() => {
        const rect = this.getBoundingClientRect();
        return { x: rect.x, y: rect.y, width: rect.width, height: rect.height };
    }`).Val().(map[string]interface{})

	passwordX := box["x"].(float64) + box["width"].(float64)/2
	passwordY := box["y"].(float64) + box["height"].(float64)/2

	// Move mouse smoothly to password field
	stealth.MoveMouseSmoothly(page, passwordX, passwordY)

	// Click the password field
	page.Mouse.MustClick("left")

	// Wait before starting to type password
	stealth.RandomSleep(200, 500)

	// Fill in the password
	log.Println("Filling in password...")
	stealth.HumanType(passwordField, password)

	// Simulate thinking/checking before clicking sign in
	log.Println("Pausing before clicking Sign in...")
	stealth.RandomSleep(1000, 2000)

	// Move to the "Sign in" button smoothly
	log.Println("Moving to Sign in button...")
	signInButton := page.MustElement("button[type=submit]")

	// Get button position
	buttonBox := signInButton.MustEval(`() => {
        const rect = this.getBoundingClientRect();
        return { x: rect.x, y: rect.y, width: rect.width, height: rect.height };
    }`).Val().(map[string]interface{})

	buttonX := buttonBox["x"].(float64) + buttonBox["width"].(float64)/2
	buttonY := buttonBox["y"].(float64) + buttonBox["height"].(float64)/2

	// Move mouse smoothly to button
	stealth.MoveMouseSmoothly(page, buttonX, buttonY)

	// Hover for a moment
	stealth.RandomSleep(200, 400)

	// Click the "Sign in" button
	log.Println("Clicking the 'Sign in' button...")
	page.Mouse.MustClick("left")

	// Wait for the navigation to complete
	log.Println("Waiting for navigation to complete...")
	page.MustWaitLoad()
	if !page.MustHas("#global-nav") {
		return errors.New("login failed: could not find the global navigation bar")
	}

	// Save cookies after successful login
	log.Println("Login successful. Saving cookies...")
	if err := saveCookies(browser); err != nil {
		log.Printf("Failed to save cookies: %v", err)
	}

	return nil
}

// loadCookies loads cookies from cookies.json and sets them in the browser.
func loadCookies(browser *rod.Browser) error {
	file, err := os.Open(cookiesFile)
	if err != nil {
		return err
	}
	defer file.Close()

	var cookies []*proto.NetworkCookie
	if err := json.NewDecoder(file).Decode(&cookies); err != nil {
		return err
	}

	browser.MustSetCookies(cookies...)
	return nil
}

// saveCookies saves the current browser cookies to cookies.json.
func saveCookies(browser *rod.Browser) error {
	cookies := browser.MustGetCookies()
	file, err := os.Create(cookiesFile)
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewEncoder(file).Encode(cookies)
}