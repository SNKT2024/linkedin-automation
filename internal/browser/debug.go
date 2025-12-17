package browser

import (
	"log"

	"github.com/go-rod/rod"
)

// ShowCursor injects JavaScript to visualize the mouse cursor for debugging purposes.
// The script persists across page navigations using EvalOnNewDocument.
func ShowCursor(page *rod.Page) {
	page.MustEvalOnNewDocument(`() => {
		// Wait for body to be available
		const initCursor = () => {
			// Check if cursor already exists to prevent duplicates
			if (document.getElementById('ghost-cursor')) return;
			
			const box = document.createElement('div');
			box.id = 'ghost-cursor';
			box.setAttribute('style', 'position:fixed; z-index:999999; pointer-events:none; width:20px; height:20px; background:red; border-radius:50%; box-shadow: 0 0 10px rgba(255, 0, 0, 0.5); transform: translate(-50%, -50%); left:100px; top:100px;');
			
			// Append to body when ready
			if (document.body) {
				document.body.appendChild(box);
				console.log('Ghost cursor added to page');
			} else {
				document.addEventListener('DOMContentLoaded', () => {
					document.body.appendChild(box);
					console.log('Ghost cursor added to page after DOMContentLoaded');
				});
			}
			
			// Create global function to update cursor position
			window.updateGhostCursor = (x, y, color) => {
				const cursor = document.getElementById('ghost-cursor');
				if (cursor) {
					cursor.style.left = x + 'px';
					cursor.style.top = y + 'px';
					if (color) {
						cursor.style.background = color;
						cursor.style.boxShadow = color === 'blue' ? '0 0 10px rgba(0, 0, 255, 0.5)' : '0 0 10px rgba(255, 0, 0, 0.5)';
					}
				}
			};
		};
		
		// Run immediately if possible
		if (document.readyState === 'loading') {
			document.addEventListener('DOMContentLoaded', initCursor);
		} else {
			initCursor();
		}
	}`)
}

// TestCursor shows the cursor in the center of the page for testing
func TestCursor(page *rod.Page) {
	log.Println("Testing cursor visibility...")
	result := page.MustEval(`() => {
		const cursor = document.getElementById('ghost-cursor');
		if (cursor) {
			cursor.style.left = '500px';
			cursor.style.top = '300px';
			cursor.style.background = 'lime';
			return 'Cursor found and moved to center';
		}
		return 'Cursor NOT found!';
	}`)
	log.Println("Test result:", result.Str())
}

// UpdateCursor updates the visual cursor position on the page.
func UpdateCursor(page *rod.Page, x, y float64, color string) {
	if color == "" {
		color = "red"
	}
	page.MustEval(`(x, y, color) => {
		if (window.updateGhostCursor) {
			window.updateGhostCursor(x, y, color);
		}
	}`, x, y, color)
}