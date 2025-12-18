package linkedin

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"strings"
	"time"

	"github.com/SNKT2024/linkedin-automation/internal/config"
	"github.com/SNKT2024/linkedin-automation/internal/stealth"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

const cookiesFile = "cookies.json"

// Login handles LinkedIn authentication, using cookies if available or logging in manually.
// Now accepts Config struct instead of reading from environment directly.
func Login(browser *rod.Browser, page *rod.Page, cfg *config.Config) error {
	// Get credentials from config
	email := cfg.Email
	password := cfg.Password

	// Attempt to load cookies from cookies.json
	if err := loadCookies(browser); err == nil {
		log.Println("Cookies loaded successfully. Navigating to feed...")

		// Navigate directly to feed instead of reloading
		page.MustNavigate("https://www.linkedin.com/feed/")
		page.MustWaitLoad()

		// Wait up to 15 seconds and check for success conditions
		log.Println("Checking if cookies are valid...")
		loginSuccess := false

		// Try for up to 15 seconds
		startTime := time.Now()
		for time.Since(startTime) < 15*time.Second {
			currentURL := page.MustInfo().URL
			log.Printf("Current URL: %s", currentURL)

			// Check condition 1: URL contains "/feed" (not redirected to login)
			if strings.Contains(currentURL, "/feed") {
				log.Println("Login successful via cookies (detected feed URL).")
				loginSuccess = true
				break
			}

			// Check condition 2: URL contains "/check/challenge" or "/challenge" (security check)
			if strings.Contains(currentURL, "/challenge") || strings.Contains(currentURL, "/checkpoint") {
				log.Println("Security challenge detected. Cookies are valid but verification needed.")
				log.Println("Please complete the security check in the browser...")
				// Wait for user to complete security check
				time.Sleep(30 * time.Second)
				continue
			}

			// Check condition 3: #global-nav exists
			err := rod.Try(func() {
				page.Timeout(1 * time.Second).MustElement("#global-nav")
			})
			if err == nil {
				log.Println("Login successful via cookies (detected global-nav element).")
				loginSuccess = true
				break
			}

			// Wait a bit before checking again
			time.Sleep(500 * time.Millisecond)
		}

		if loginSuccess {
			log.Println("✅ Cookie-based authentication successful!")
			return nil
		}

		log.Printf("Cookies appear invalid. Final URL: %s", page.MustInfo().URL)
		log.Println("Proceeding with manual login.")
	}

	// Ensure we are on the login page before attempting manual login
	currentURL := page.MustInfo().URL
	if !strings.Contains(currentURL, "/login") {
		log.Println("Not on login page. Navigating to login page...")
		page.MustNavigate("https://www.linkedin.com/login")
		page.MustWaitLoad()
	}

	// Navigate to LinkedIn login page (if not already there)
	log.Println("Navigating to LinkedIn login page...")
	page.MustNavigate("https://www.linkedin.com/login")
	page.MustWaitLoad()

	// Simulate user looking at the page before typing
	stealth.RandomSleep(1000, 3000)

	// Find and fill the username field
	log.Println("Filling in email...")

	// Verify we have the login form
	err := rod.Try(func() {
		page.Timeout(5 * time.Second).MustElement("#username")
	})
	if err != nil {
		return errors.New("login failed: could not find username field on login page")
	}

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

	// Use Race to handle three possible outcomes
	log.Println("Verifying login outcome...")

	// Race for multiple outcomes with proper error handling
	outcome := ""
	err = rod.Try(func() {
		page.Race().Element("#global-nav").MustHandle(func(e *rod.Element) {
			outcome = "success"
		}).Element("input[name='pin']").MustHandle(func(e *rod.Element) {
			outcome = "2fa"
		}).Element(".secondary-action").MustHandle(func(e *rod.Element) {
			outcome = "2fa"
		}).Element(".error-for-username").MustHandle(func(e *rod.Element) {
			outcome = "error"
		}).Element(".error-for-password").MustHandle(func(e *rod.Element) {
			outcome = "error"
		}).Element(".form__input--error").MustHandle(func(e *rod.Element) {
			outcome = "error"
		}).MustDo()
	})

	// Handle timeout (no outcome matched within timeout)
	if err != nil {
		log.Printf("Timeout or error during login verification: %v", err)
		// Fallback: check URL and global-nav manually
		currentURL := page.MustInfo().URL
		if strings.Contains(currentURL, "/feed") || page.MustHas("#global-nav") {
			outcome = "success"
		} else if strings.Contains(currentURL, "/challenge") || strings.Contains(currentURL, "/checkpoint") {
			outcome = "2fa"
		} else {
			outcome = "unknown"
		}
	}

	// Handle outcomes
	switch outcome {
	case "success":
		log.Println("✅ Login successful (detected global-nav element).")
		// Save cookies after successful login
		log.Println("Saving cookies...")
		if err := saveCookies(browser); err != nil {
			log.Printf("Failed to save cookies: %v", err)
		}
		return nil

	case "2fa":
		log.Println("⚠️ 2FA/VERIFICATION DETECTED")
		log.Println("=" + strings.Repeat("=", 50))
		log.Println("Please solve the captcha/enter PIN in the browser manually.")
		log.Println("The bot will wait until you complete the verification...")
		log.Println("=" + strings.Repeat("=", 50))

		// Wait indefinitely for user to complete 2FA
		log.Println("Waiting for #global-nav to appear after 2FA...")
		page.MustElement("#global-nav") // Wait indefinitely

		log.Println("✅ 2FA completed successfully!")
		// Save cookies after successful login
		log.Println("Saving cookies...")
		if err := saveCookies(browser); err != nil {
			log.Printf("Failed to save cookies: %v", err)
		}
		return nil

	case "error":
		log.Println("❌ LOGIN FAILED: Invalid Credentials")
		log.Println("Please check your LINKEDIN_EMAIL and LINKEDIN_PASSWORD in the .env file.")
		return errors.New("login failed: invalid username or password")

	default:
		log.Println("⚠️ Unknown login outcome. Attempting fallback verification...")
		// Try one more time to check for success
		startTime := time.Now()
		for time.Since(startTime) < 10*time.Second {
			currentURL := page.MustInfo().URL
			if strings.Contains(currentURL, "/feed") {
				log.Println("✅ Login successful (detected feed URL in fallback).")
				if err := saveCookies(browser); err != nil {
					log.Printf("Failed to save cookies: %v", err)
				}
				return nil
			}

			err := rod.Try(func() {
				page.Timeout(1 * time.Second).MustElement("#global-nav")
			})
			if err == nil {
				log.Println("✅ Login successful (detected global-nav in fallback).")
				if err := saveCookies(browser); err != nil {
					log.Printf("Failed to save cookies: %v", err)
				}
				return nil
			}

			time.Sleep(500 * time.Millisecond)
		}

		return errors.New("login failed: could not verify successful login")
	}
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