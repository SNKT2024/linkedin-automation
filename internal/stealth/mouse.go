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

// GenerateBezierPath calculates a quadratic Bezier curve with random control points and noise.
func GenerateBezierPath(fromX, fromY, toX, toY float64) []Point {
	// Random control point to create an arc
	controlX := (fromX + toX) / 2
	controlY := (fromY + toY) / 2
	controlX += rand.Float64()*100 - 50 // Add random noise to control point
	controlY += rand.Float64()*100 - 50

	// Number of steps in the curve
	steps := 50
	path := make([]Point, steps)

	for i := 0; i < steps; i++ {
		t := float64(i) / float64(steps-1) // Parameter t ranges from 0 to 1

		// Quadratic Bezier formula: B(t) = (1-t)^2 * P0 + 2(1-t)t * P1 + t^2 * P2
		x := math.Pow(1-t, 2)*fromX + 2*(1-t)*t*controlX + math.Pow(t, 2)*toX
		y := math.Pow(1-t, 2)*fromY + 2*(1-t)*t*controlY + math.Pow(t, 2)*toY

		// Add slight random noise to each point
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
			// Update visual cursor position
			page.MustEval(`(x, y) => {
				if (window.updateGhostCursor) {
					window.updateGhostCursor(x, y, 'red');
				}
			}`, p.X, p.Y)
			time.Sleep(time.Duration(rand.Intn(10)+5) * time.Millisecond)
		}

		// Small pause at overshoot point (simulating "oops" moment)
		RandomSleep(50, 150)

		// Then, correct back to the actual target
		correctionPath := GenerateBezierPath(overshootX, overshootY, toX, toY)
		for _, p := range correctionPath {
			page.Mouse.MustMoveTo(p.X, p.Y)
			// Update visual cursor position
			page.MustEval(`(x, y) => {
				if (window.updateGhostCursor) {
					window.updateGhostCursor(x, y, 'orange');
				}
			}`, p.X, p.Y)
			time.Sleep(time.Duration(rand.Intn(10)+5) * time.Millisecond)
		}

		// Change cursor back to red after correction
		page.MustEval(`(x, y) => {
			if (window.updateGhostCursor) {
				window.updateGhostCursor(x, y, 'red');
			}
		}`, toX, toY)

	} else {
		// Normal movement without overshoot
		path := GenerateBezierPath(currentX, currentY, toX, toY)

		// Move the mouse along the path
		for _, p := range path {
			page.Mouse.MustMoveTo(p.X, p.Y)
			// Update visual cursor position
			page.MustEval(`(x, y) => {
				if (window.updateGhostCursor) {
					window.updateGhostCursor(x, y, 'red');
				}
			}`, p.X, p.Y)
			time.Sleep(time.Duration(rand.Intn(10)+5) * time.Millisecond)
		}
	}
}