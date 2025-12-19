package stealth

import (
	"math"
	"math/rand"
	"time"

	"github.com/go-rod/rod"
)

// Point represents a 2D point for mouse movement.
type Point struct {
	X float64
	Y float64
}

// GenerateBezierPath calculates a quadratic Bezier curve with random control points.
func GenerateBezierPath(fromX, fromY, toX, toY float64) []Point {
	controlX := (fromX + toX) / 2
	controlY := (fromY + toY) / 2
	controlX += rand.Float64()*100 - 50
	controlY += rand.Float64()*100 - 50

	steps := 50
	path := make([]Point, steps)

	for i := 0; i < steps; i++ {
		t := float64(i) / float64(steps-1)
		x := math.Pow(1-t, 2)*fromX + 2*(1-t)*t*controlX + math.Pow(t, 2)*toX
		y := math.Pow(1-t, 2)*fromY + 2*(1-t)*t*controlY + math.Pow(t, 2)*toY
		
		// Add noise
		x += rand.Float64()*2 - 1
		y += rand.Float64()*2 - 1
		
		path[i] = Point{X: x, Y: y}
	}
	return path
}

// MoveMouseSmoothly moves the mouse along a Bezier curve to the target position.
// Implements human overshoot behavior where the cursor occasionally overshoots
// the target and then corrects back to the actual position.
func MoveMouseSmoothly(page *rod.Page, toX, toY float64) {
	// Get the real current mouse position
	pos := page.Mouse.Position()
	currentX := pos.X
	currentY := pos.Y

	// 30% chance of triggering overshoot behavior
	shouldOvershoot := rand.Float64() < 0.3

	if shouldOvershoot {
		// Calculate direction vector from current to target
		dirX := toX - currentX
		dirY := toY - currentY

		// Normalize the direction vector
		length := math.Sqrt(dirX*dirX + dirY*dirY)
		if length > 0 {
			dirX /= length
			dirY /= length
		}

		// Calculate overshoot distance (10-60 pixels past the target)
		overshootDistance := 10.0 + rand.Float64()*50.0

		// Calculate overshoot position
		overshootX := toX + dirX*overshootDistance
		overshootY := toY + dirY*overshootDistance

		// First, move to overshoot position
		overshootPath := GenerateBezierPath(currentX, currentY, overshootX, overshootY)
		for _, p := range overshootPath {
			page.Mouse.MustMoveTo(p.X, p.Y)
			time.Sleep(time.Duration(rand.Intn(10)+5) * time.Millisecond)
		}

		// Small pause at overshoot point (simulating "oops" moment)
		RandomSleep(50, 150)

		// Then, correct back to the actual target
		correctionPath := GenerateBezierPath(overshootX, overshootY, toX, toY)
		for _, p := range correctionPath {
			page.Mouse.MustMoveTo(p.X, p.Y)
			time.Sleep(time.Duration(rand.Intn(10)+5) * time.Millisecond)
		}
	} else {
		// Normal movement without overshoot
		path := GenerateBezierPath(currentX, currentY, toX, toY)

		// Move the mouse along the path
		for _, p := range path {
			page.Mouse.MustMoveTo(p.X, p.Y)
			time.Sleep(time.Duration(rand.Intn(10)+5) * time.Millisecond)
		}
	}
}

// NaturalScroll simulates natural mouse wheel scrolling with inertia and acceleration/deceleration.
// Uses page.Mouse.Scroll instead of window.scrollBy to mimic physical mouse wheel rotation.
func NaturalScroll(page *rod.Page, deltaY int) {
	// 1. CRITICAL FIX: Move Mouse to Center
	// This ensures we are scrolling the 'body' and not stuck hovering on the fixed header/nav bar.
	page.Mouse.MustMoveTo(500, 500)

	// Determine scroll direction
	direction := 1.0
	if deltaY < 0 {
		direction = -1.0
		deltaY = -deltaY 
	}

	// Break scrolling into small steps
	stepSize := 40 + rand.Intn(20) 
	numSteps := deltaY / stepSize
	if numSteps < 1 {
		numSteps = 1
	}

	// Generate delays with inertia
	delays := generateInertiaDelays(numSteps)

	// Scroll in steps
	for i := 0; i < numSteps; i++ {
		scrollAmount := direction * float64(stepSize)
		page.Mouse.MustScroll(0, scrollAmount)
		time.Sleep(time.Duration(delays[i]) * time.Millisecond)
	}
}

// generateInertiaDelays creates a delay pattern with acceleration and deceleration.
// generateInertiaDelays creates a delay pattern with acceleration and deceleration.
func generateInertiaDelays(numSteps int) []int {
	delays := make([]int, numSteps)

	// Split into three phases: acceleration (25%), constant (50%), deceleration (25%)
	accelPhase := numSteps / 4
	decelPhase := numSteps / 4
	constPhase := numSteps - accelPhase - decelPhase

	idx := 0

	// Acceleration phase
	for i := 0; i < accelPhase; i++ {
		delays[idx] = 80 - i*15 
		if delays[idx] < 20 { delays[idx] = 20 }
		idx++
	}

	// Constant phase
	for i := 0; i < constPhase; i++ {
		delays[idx] = 20 + rand.Intn(10)
		idx++
	}

	// Deceleration phase
	for i := 0; i < decelPhase; i++ {
		delays[idx] = 20 + i*15
		if delays[idx] > 80 { delays[idx] = 80 }
		idx++
	}

	return delays
}

// RandomWander simulates idle mouse movement by moving the mouse to a random location.
func RandomWander(page *rod.Page) {
	// Get viewport size
	viewport := page.MustEval(`() => {
		return { width: window.innerWidth, height: window.innerHeight };
	}`).Val().(map[string]interface{})

	viewportWidth := viewport["width"].(float64)
	viewportHeight := viewport["height"].(float64)

	// Pick a random point within the viewport (avoiding edges)
	margin := 100.0 // Keep 100px away from edges
	targetX := margin + rand.Float64()*(viewportWidth-2*margin)
	targetY := margin + rand.Float64()*(viewportHeight-2*margin)

	// Move to that point smoothly using Bezier curve
	MoveMouseSmoothly(page, targetX, targetY)

	// Hover at that location (simulating reading/thinking)
	hoverTime := 500 + rand.Intn(1000) // 0.5-1.5 seconds
	time.Sleep(time.Duration(hoverTime) * time.Millisecond)
}

// ScrollWithReading simulates natural scrolling behavior while reading content.
// It scrolls down in chunks, pauses to "read", and occasionally scrolls back up slightly.
func ScrollWithReading(page *rod.Page, totalDistance int) {
	scrolled := 0

	for scrolled < totalDistance {
		// Decide how much to scroll this iteration (200-500 pixels)
		scrollAmount := 200 + rand.Intn(300)
		if scrolled+scrollAmount > totalDistance {
			scrollAmount = totalDistance - scrolled
		}

		// 20% chance to scroll up a bit (re-reading behavior)
		if rand.Float64() < 0.2 && scrolled > 0 {
			scrollUpAmount := 50 + rand.Intn(100) // 50-150 pixels
			NaturalScroll(page, -scrollUpAmount)
			RandomSleep(800, 1500) // Pause while "re-reading"
		}

		// Scroll down naturally
		NaturalScroll(page, scrollAmount)
		scrolled += scrollAmount

		// Pause to "read" the content (longer pause = more realistic)
		readingTime := 1500 + rand.Intn(2500) // 1.5-4 seconds
		time.Sleep(time.Duration(readingTime) * time.Millisecond)

		// 30% chance to wander mouse while reading
		if rand.Float64() < 0.3 {
			RandomWander(page)
		}
	}
}