package stealth

import (
	"math"
	"math/rand"
	"time"

	"github.com/go-rod/rod"
)

// RandomSleep sleeps for a random duration between min and max milliseconds.
func RandomSleep(min, max int) {
	duration := time.Duration(rand.Intn(max-min)+min) * time.Millisecond
	time.Sleep(duration)
}

// HumanType simulates realistic human-like typing into an input element.
// It uses WPM-based delays with Gaussian distribution and occasional micro-pauses.
func HumanType(element *rod.Element, text string) {
	// Average typing speed in WPM (Words Per Minute)
	// Typical human typing ranges from 40-60 WPM
	wpm := 45.0 + rand.Float64()*15.0 // Random WPM between 45-60

	// Calculate base delay per character in milliseconds
	// Average word length is ~5 characters, so chars per minute = WPM * 5
	charsPerMinute := wpm * 5
	baseDelayMs := 60000.0 / charsPerMinute // Convert to milliseconds per character

	for _, char := range text {
		element.MustInput(string(char))

		// Generate delay using Gaussian distribution
		// Mean is baseDelayMs, standard deviation is 30% of the mean
		stdDev := baseDelayMs * 0.3
		delay := gaussianDelay(baseDelayMs, stdDev)

		// Ensure delay is positive and reasonable
		if delay < 20 {
			delay = 20
		}
		if delay > 300 {
			delay = 300
		}

		// 20% chance of adding a micro-pause (hesitation)
		if rand.Float64() < 0.2 {
			delay += float64(300 + rand.Intn(200)) // Add 300-500ms pause
		}

		time.Sleep(time.Duration(delay) * time.Millisecond)
	}
}

// gaussianDelay generates a delay using Gaussian (normal) distribution.
// This creates more natural variation than uniform random distribution.
func gaussianDelay(mean, stdDev float64) float64 {
	// Box-Muller transform to generate normal distribution
	u1 := rand.Float64()
	u2 := rand.Float64()

	// Avoid log(0)
	if u1 < 1e-10 {
		u1 = 1e-10
	}

	z0 := math.Sqrt(-2.0*math.Log(u1)) * math.Cos(2.0*math.Pi*u2)
	return mean + z0*stdDev
}

// HumanClick simulates a human-like click on an element with smooth Bezier curve movement.
func HumanClick(page *rod.Page, element *rod.Element) {
	// Get the element's dimensions and position using JavaScript
	box := element.MustEval(`() => {
		const rect = this.getBoundingClientRect();
		return { x: rect.x, y: rect.y, width: rect.width, height: rect.height };
	}`).Val().(map[string]interface{})

	x := box["x"].(float64) + box["width"].(float64)/2
	y := box["y"].(float64) + box["height"].(float64)/2

	// Move the mouse smoothly to the element's center using Bezier curve
	MoveMouseSmoothly(page, x, y)

	// Simulate "aiming" before clicking
	RandomSleep(300, 700)

	// Update cursor to blue before clicking
	page.MustEval(`(x, y) => {
		if (window.updateGhostCursor) {
			window.updateGhostCursor(x, y, 'blue');
		}
	}`, x, y)

	// Perform the click
	page.Mouse.MustClick("left")

	// Update cursor back to red after clicking
	page.MustEval(`(x, y) => {
		if (window.updateGhostCursor) {
			window.updateGhostCursor(x, y, 'red');
		}
	}`, x, y)
}