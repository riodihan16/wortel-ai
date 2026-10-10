// Package window sizes the single app window per UI mode and keeps it inside
// the work area of its monitor. Placement math is pure and platform-free;
// the Win32 glue lives in place_windows.go.
package window

import "errors"

// Mode is the UI layout the window is sized for.
type Mode string

const (
	Pet   Mode = "pet"
	Panel Mode = "panel" // chat and settings share one panel size
)

// Dim is a window size in device-independent pixels (96 DPI).
type Dim struct{ W, H int }

var (
	PetSize   = Dim{W: 180, H: 180}
	PanelSize = Dim{W: 400, H: 600}
)

// ClassName is the Win32 class of the main window, set in main.go so the
// handle can be found without relying on the Wails default.
const ClassName = "WortelAIWindow"

var errUnknownMode = errors.New("mode window tidak dikenal")

// SizeOf returns the DIP size for a mode, rejecting values the frontend
// should never send.
func SizeOf(m Mode) (Dim, error) {
	switch m {
	case Pet:
		return PetSize, nil
	case Panel:
		return PanelSize, nil
	default:
		return Dim{}, errUnknownMode
	}
}

// Rect is a screen rectangle in physical pixels; Right/Bottom are exclusive.
type Rect struct{ Left, Top, Right, Bottom int }

// Fit returns the top-left for a w×h window that wants to sit at (x, y),
// moved the least distance needed to lie inside work. A window larger than
// work is pinned to work's top-left so its title area stays reachable.
func Fit(x, y, w, h int, work Rect) (int, int) {
	return clamp(x, w, work.Left, work.Right), clamp(y, h, work.Top, work.Bottom)
}

// Constrain returns r moved, size unchanged, to lie inside work. Used on
// every step of a user drag so the window cannot be pushed off screen.
func Constrain(r, work Rect) Rect {
	w, h := r.Right-r.Left, r.Bottom-r.Top
	x, y := Fit(r.Left, r.Top, w, h, work)
	return Rect{Left: x, Top: y, Right: x + w, Bottom: y + h}
}

func clamp(pos, size, lo, hi int) int {
	pos = min(pos, hi-size)
	return max(pos, lo)
}

// Placer remembers how far the window was pushed when it grew from pet to
// panel. Shrinking back undoes that push, so the pet returns to the spot the
// user left it, or follows by the same offset if the panel was dragged.
type Placer struct {
	panel          bool
	shiftX, shiftY int
}

// Place returns the top-left for switching to mode at size w×h, given the
// current top-left (x, y) and the work area of the window's monitor.
func (p *Placer) Place(m Mode, x, y, w, h int, work Rect) (int, int) {
	switch {
	case m == Panel && !p.panel:
		nx, ny := Fit(x, y, w, h, work)
		p.panel, p.shiftX, p.shiftY = true, nx-x, ny-y
		return nx, ny
	case m == Pet && p.panel:
		sx, sy := p.shiftX, p.shiftY
		p.panel, p.shiftX, p.shiftY = false, 0, 0
		return Fit(x-sx, y-sy, w, h, work)
	default:
		// Same mode again (e.g. tray Show while already in chat): only make
		// sure it is on screen, keep the remembered shift.
		return Fit(x, y, w, h, work)
	}
}
