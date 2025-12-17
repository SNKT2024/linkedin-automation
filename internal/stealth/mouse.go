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
func MoveMouseSmoothly(page *rod.Page, toX, toY float64) {
	// Start from current position (0, 0) as Rod doesn't track position
	// We'll generate the path from a reasonable starting point
	currentX := 0.0
	currentY := 0.0

	// Generate the Bezier path
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
				time.Sleep(time.Duration(rand.Intn(10)+5) * time.Millisecond) // Sleep 5-15ms between steps
	}
}