package linkedin

import (
	"errors"
	"log"
	"math/rand"
	"time"

	"github.com/SNKT2024/linkedin-automation/internal/stealth"
	"github.com/go-rod/rod"
)

// ConnectWithProfile attempts to send a connection request to a LinkedIn profile.
// Returns status and error:
// - "skipped_pending": Connection request already sent (Pending/Withdraw button found)
// - "skipped_connected": Already connected (Message button found)
// - "skipped_premium": Premium profile requiring InMail (cannot connect for free)
// - "clicked": Successfully clicked Connect button
// - "failed": Could not find Connect button or action failed
func ConnectWithProfile(page *rod.Page, profileURL string) (string, error) {
    log.Printf("Navigating to profile: %s", profileURL)

    // Navigate to profile page
    page.MustNavigate(profileURL)
    page.MustWaitLoad()

    // Simulate reading the profile (3-5 seconds)
    log.Println("Reading profile...")
    stealth.RandomSleep(3000, 5000)

    // Simulate scrolling to view profile (human behavior)
    log.Println("Scrolling profile page...")
    stealth.NaturalScroll(page, 300+rand.Intn(200)) // Scroll 300-500px
    stealth.RandomSleep(1500, 2500)

    // Occasionally wander mouse (20% chance)
    if rand.Float64() < 0.2 {
        log.Println("Mouse wandering while viewing profile...")
        stealth.RandomWander(page)
    }

    // ==========================================
    // STATUS CHECK: Pending/Withdraw
    // ==========================================
    log.Println("Checking connection status...")

    // Check for "Pending" button using ElementR (regex text match)
    errPending := rod.Try(func() {
        page.Timeout(2 * time.Second).MustElementR("button", "Pending")
        log.Println("⏭️  Connection already pending (skipping)")
    })
    if errPending == nil {
        return "skipped_pending", nil
    }

    // Check for "Withdraw" button
    errWithdraw := rod.Try(func() {
        page.Timeout(2 * time.Second).MustElementR("button", "Withdraw")
        log.Println("⏭️  Connection pending - Withdraw option available (skipping)")
    })
    if errWithdraw == nil {
        return "skipped_pending", nil
    }

    // ==========================================
    // STATUS CHECK: Already Connected
    // ==========================================

    // Check for "Message" button (already connected)
    errMessage := rod.Try(func() {
        page.Timeout(2 * time.Second).MustElementR("button", "^Message$")
        log.Println("⏭️  Already connected to this profile (skipping)")
    })
    if errMessage == nil {
        return "skipped_connected", nil
    }

    // Check for "Following" button (another already connected state)
    errFollowing := rod.Try(func() {
        page.Timeout(2 * time.Second).MustElementR("button", "Following")
        log.Println("⏭️  Already following/connected to this profile (skipping)")
    })
    if errFollowing == nil {
        return "skipped_connected", nil
    }

    // ==========================================
    // STATUS CHECK: Premium/InMail Only
    // ==========================================

    // Check for "Message" button with Premium badge (InMail required)
    errInMail := rod.Try(func() {
        // Look for InMail indicator or Premium message button
        page.Timeout(2 * time.Second).MustElement(`button[aria-label*="Send InMail"], .premium-inmail-button`)
        log.Println("⏭️  Premium profile - InMail required (skipping)")
    })
    if errInMail == nil {
        return "skipped_premium", nil
    }

    // Check if profile is "Out of network" (3rd degree) requiring Premium
    errOutOfNetwork := rod.Try(func() {
        page.Timeout(2 * time.Second).MustElementR("span, div", "Out of network|3rd")
        log.Println("⏭️  Out of network - may require Premium (attempting anyway)")
    })
    _ = errOutOfNetwork // Log but don't skip, try to connect anyway

    // ==========================================
    // PRIORITY A: Direct "Connect" Button
    // ==========================================
    log.Println("Looking for direct 'Connect' button...")

    errConnect := rod.Try(func() {
        // Use ElementR with regex to find button strictly containing "Connect"
        connectButton := page.Timeout(3 * time.Second).MustElementR("button", "^Connect$")

        log.Println("✅ Found direct 'Connect' button")

        // Scroll button into view if needed
        connectButton.MustScrollIntoView()
        stealth.RandomSleep(500, 1000)

        // Use stealth click
        log.Println("Clicking 'Connect' button...")
        stealth.HumanClick(page, connectButton)

        // Wait for modal/dialog to appear
        log.Println("Waiting for connection dialog...")
        stealth.RandomSleep(2000, 3000)

        // Handle the connection dialog
        handleConnectionDialog(page)
    })

    if errConnect == nil {
        log.Println("✅ Successfully clicked Connect button (Priority A)")
        return "clicked", nil
    }

    log.Println("Direct 'Connect' button not found. Trying 'More' dropdown...")

    // ==========================================
    // PRIORITY B: "More" Dropdown → "Connect"
    // ==========================================
    log.Println("Looking for 'More' button...")

    errMore := rod.Try(func() {
        // Find "More" button using multiple approaches
        var moreButton *rod.Element

        // Approach 1: Try ElementR with text "More"
        errMoreText := rod.Try(func() {
            moreButton = page.Timeout(3 * time.Second).MustElementR("button", "^More$|More actions")
        })

        // Approach 2: Try aria-label selector
        if errMoreText != nil {
            errAria := rod.Try(func() {
                moreButton = page.Timeout(3 * time.Second).MustElement(`button[aria-label*="More actions"]`)
            })
            if errAria != nil {
                // Approach 3: Try class selector
                moreButton = page.Timeout(3 * time.Second).MustElement(`button.artdeco-dropdown__trigger, button.pvs-overflow-actions-dropdown__trigger`)
            }
        }

        if moreButton == nil {
            panic("More button not found")
        }

        log.Println("✅ Found 'More' button")

        // Scroll into view
        moreButton.MustScrollIntoView()
        stealth.RandomSleep(500, 1000)

        // Click "More" to open dropdown
        log.Println("Clicking 'More' button to open dropdown...")
        stealth.HumanClick(page, moreButton)

        // Wait for dropdown to appear
        log.Println("Waiting for dropdown menu...")
        stealth.RandomSleep(1500, 2500)

        // Wait for dropdown menu to be visible
        page.Timeout(3 * time.Second).MustElement(`div[role="menu"], ul[role="menu"], div.artdeco-dropdown__content`)

        // Find Connect option - check multiple possible text variations
        log.Println("Looking for 'Connect' in dropdown...")
        connectOption := page.Timeout(3 * time.Second).MustElementR("button, a, div[role='menuitem'], li", "^Connect$|^Connect ")

        log.Println("✅ Found 'Connect' option in dropdown")

        // Click "Connect" in dropdown
        log.Println("Clicking 'Connect' option...")
        stealth.HumanClick(page, connectOption)

        // Wait for connection dialog
        log.Println("Waiting for connection dialog...")
        stealth.RandomSleep(2000, 3000)

        // Handle the connection dialog
        handleConnectionDialog(page)
    })

    if errMore == nil {
        log.Println("✅ Successfully clicked Connect via 'More' dropdown (Priority B)")
        return "clicked", nil
    }

    // ==========================================
    // PRIORITY C: Enterprise/Sales Navigator Profiles
    // ==========================================
    log.Println("Trying enterprise profile connection methods...")

    errEnterprise := rod.Try(func() {
        // Some enterprise profiles have "Get introduced" or "Request introduction"
        introButton := page.Timeout(3 * time.Second).MustElementR("button", "Get introduced|Request introduction")
        
        log.Println("✅ Found enterprise introduction button")
        stealth.HumanClick(page, introButton)
        stealth.RandomSleep(2000, 3000)
        
        // Try to send the introduction request
        handleConnectionDialog(page)
    })

    if errEnterprise == nil {
        log.Println("✅ Successfully requested introduction (Enterprise Profile)")
        return "clicked", nil
    }

    // ==========================================
    // FAILURE: Connect button not found
    // ==========================================
    log.Println("❌ Could not find 'Connect' button (tried direct, 'More' dropdown, and enterprise methods)")
    return "failed", errors.New("connect button not found on profile page")
}

// handleConnectionDialog handles the connection request dialog that appears after clicking Connect
func handleConnectionDialog(page *rod.Page) {
    log.Println("Handling connection dialog...")

    // Try to find and click "Send without a note" or "Send" button
    errSend := rod.Try(func() {
        var sendButton *rod.Element

        // Approach 1: Try ElementR with text variations
        errSendText := rod.Try(func() {
            sendButton = page.Timeout(5 * time.Second).MustElementR("button", "Send without a note|Send invitation|^Send now$|^Send$")
        })

        // Approach 2: Try aria-label selector
        if errSendText != nil {
            errAria := rod.Try(func() {
                sendButton = page.Timeout(3 * time.Second).MustElement(`button[aria-label*="Send without a note"], button[aria-label*="Send now"]`)
            })
            if errAria != nil {
                // Approach 3: Try data-control-name (LinkedIn-specific)
                sendButton = page.Timeout(3 * time.Second).MustElement(`button[data-control-name="send_invitation"]`)
            }
        }

        if sendButton != nil {
            log.Println("✅ Found 'Send' button in dialog")
            stealth.HumanClick(page, sendButton)
            stealth.RandomSleep(1500, 2500)
            log.Println("✅ Connection request sent")
        }
    })

    if errSend != nil {
        log.Println("⚠️ Could not auto-send - dialog may require a note or email")
        
        // Check if email is required
        errEmail := rod.Try(func() {
            page.Timeout(2 * time.Second).MustElement(`input[name="email"], input[type="email"]`)
            log.Println("⚠️ Email required for this connection - manual intervention needed")
        })
        
        if errEmail == nil {
            // Email field found - this is a limitation
            return
        }
        
        // Maybe it's asking for a note - try to close and consider it partial success
        log.Println("⚠️ Dialog opened but couldn't complete automatically (may need note/email)")
    }
}