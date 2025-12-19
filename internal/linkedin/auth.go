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

// Login handles LinkedIn authentication with "Fail Fast" logic
func Login(browser *rod.Browser, page *rod.Page, cfg *config.Config) error {
	email := cfg.Email
	password := cfg.Password

	// 1. Try Cookie Login
	if err := loadCookies(browser); err == nil {
		log.Println("🍪 Cookies loaded. Checking validity...")

		page.MustNavigate("https://www.linkedin.com/feed/")
		
		// Wait a moment for redirect to happen
		// (LinkedIn takes 1-2 seconds to decide if cookies are good or bad)
		time.Sleep(3 * time.Second)

		// 2. FAIL FAST CHECK
		// Instead of waiting 15s, we check URL immediately.
		currentURL := page.MustInfo().URL
		
		if strings.Contains(currentURL, "/feed") || strings.Contains(currentURL, "/mini-profile") {
			log.Println("✅ Cookies are valid! (Feed detected)")
			return nil
		}

		// If we are redirected to /login or /uas/login, cookies are dead.
		if strings.Contains(currentURL, "/login") || strings.Contains(currentURL, "uas/authenticate") {
			log.Println("🚫 Cookies expired (Redirected to Login). Switching to manual login immediately...")
			// Fall through to Manual Login below
		} else {
			// Edge case: Maybe internet is slow? Give it one last verification check.
			if verifyLogin(page) {
				return nil
			}
			log.Println("⚠️ Cookie login inconclusive. Switching to manual.")
		}
	}

	// 3. Manual Login (The Fallback)
	log.Println("🔓 Starting Manual Login...")
	
	// Critical: Clear invalid cookies first so LinkedIn doesn't loop
	browser.MustSetCookies() // Clears all cookies
	
	page.MustNavigate("https://www.linkedin.com/login")
	page.MustWaitLoad()
	stealth.RandomSleep(2000, 3000)

	// Fill Email
	log.Println("   ✍️ Filling Email...")
	emailInput, err := page.Element("#username")
	if err != nil { return err }
	stealth.HumanType(emailInput, email)
	stealth.RandomSleep(1000, 2000)

	// Fill Password
	log.Println("   ✍️ Filling Password...")
	passInput, err := page.Element("#password")
	if err != nil { return err }
	stealth.HumanType(passInput, password)
	stealth.RandomSleep(1000, 2000)

	// Click Sign In
	log.Println("   🚀 Clicking Sign In...")
	// Try multiple selectors for the button
	btn, err := page.Element("button[type='submit'], .login__form_action_container button")
	if err != nil { return errors.New("could not find login button") }
	
	stealth.HumanClick(page, btn)
	page.MustWaitLoad()
	
	// Wait for feed to confirm success
	log.Println("   ⏳ Waiting for Feed...")
	
	// Robust verification loop (Wait up to 30s for manual login to process)
	if verifyLogin(page) {
		log.Println("✅ Manual Login Successful!")
		saveCookies(browser) // Save fresh cookies for next time
		return nil
	}

	return errors.New("manual login failed (timeout waiting for feed)")
}

// verifyLogin waits up to 15 seconds for signs of a successful login
func verifyLogin(page *rod.Page) bool {
	// Poll every 1 second for 15 seconds
	for i := 0; i < 15; i++ {
		if strings.Contains(page.MustInfo().URL, "/feed") {
			return true
		}
		// Check for global nav bar (strong indicator of logged-in state)
		if _, err := page.Element("#global-nav"); err == nil {
			return true
		}
		time.Sleep(1 * time.Second)
	}
	return false
}

// loadCookies loads cookies from file
func loadCookies(browser *rod.Browser) error {
	file, err := os.Open(cookiesFile)
	if err != nil { return err }
	defer file.Close()

	var cookies []*proto.NetworkCookie
	if err := json.NewDecoder(file).Decode(&cookies); err != nil { return err }

	// Convert NetworkCookie to NetworkCookieParam
	cookieParams := make([]*proto.NetworkCookieParam, len(cookies))
	for i, cookie := range cookies {
		cookieParams[i] = &proto.NetworkCookieParam{
			Name:     cookie.Name,
			Value:    cookie.Value,
			Domain:   cookie.Domain,
			Path:     cookie.Path,
			Secure:   cookie.Secure,
			HTTPOnly: cookie.HTTPOnly,
			SameSite: cookie.SameSite,
			Expires:  cookie.Expires,
		}
	}

	return browser.SetCookies(cookieParams)
}

// saveCookies saves active cookies to file
func saveCookies(browser *rod.Browser) error {
	cookies, err := browser.GetCookies()
	if err != nil { return err }

	data, err := json.MarshalIndent(cookies, "", "  ")
	if err != nil { return err }

	return os.WriteFile(cookiesFile, data, 0644)
}