//go:build windows

package window

import (
	"fmt"
	"sync/atomic"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	pGetWindowLongPtr = user32.NewProc("GetWindowLongPtrW")
	pSetWindowLongPtr = user32.NewProc("SetWindowLongPtrW")
	pCallWindowProc   = user32.NewProc("CallWindowProcW")
	pMonitorFromPoint = user32.NewProc("MonitorFromPoint")
	pGetCursorPos     = user32.NewProc("GetCursorPos")
)

const wmMoving = 0x0216

type point struct{ X, Y int32 }

// GWLP_WNDPROC is -4; nIndex is a C int, so pass its two's complement.
var gwlpWndProc = int32(-4)

var (
	// Read on the UI thread inside movingProc, written once before the
	// subclass goes live, hence atomic rather than mu.
	prevProc atomic.Uintptr
	movingCB = windows.NewCallback(movingProc)
)

// KeepOnScreen subclasses the main window so a user drag (pet or panel
// header) cannot push it past the work area of the monitor it is over.
// The rect is corrected on every WM_MOVING step, so the window stops at the
// edge instead of snapping back after release. Idempotent.
func KeepOnScreen() error {
	mu.Lock()
	defer mu.Unlock()
	if prevProc.Load() != 0 {
		return nil
	}
	h, err := mainWindow()
	if err != nil {
		return err
	}
	idx := uintptr(gwlpWndProc)
	old, _, e := pGetWindowLongPtr.Call(uintptr(h), idx)
	if old == 0 {
		return fmt.Errorf("GetWindowLongPtr: %w", e)
	}
	// Store first: messages may arrive on the UI thread the moment the new
	// proc is installed, and they must be able to chain.
	prevProc.Store(old)
	if r, _, e := pSetWindowLongPtr.Call(uintptr(h), idx, movingCB); r == 0 {
		prevProc.Store(0)
		return fmt.Errorf("SetWindowLongPtr: %w", e)
	}
	return nil
}

func movingProc(h, msg, wp, lp uintptr) uintptr {
	if uint32(msg) == wmMoving && lp != 0 {
		// lp points at the proposed window RECT; double indirection keeps
		// vet's unsafeptr check quiet without changing the meaning.
		wr := *(**windows.Rect)(unsafe.Pointer(&lp))
		constrainRect(wr)
	}
	r, _, _ := pCallWindowProc.Call(prevProc.Load(), h, msg, wp, lp)
	return r
}

func constrainRect(wr *windows.Rect) {
	// Clamp into the monitor under the cursor, not the one the window is
	// on: the window then follows the cursor onto another display as soon
	// as the cursor crosses, instead of walling at the shared edge until
	// most of the window is over.
	var pt point
	if ok, _, _ := pGetCursorPos.Call(uintptr(unsafe.Pointer(&pt))); ok == 0 {
		return
	}
	// POINT is passed by value; on 64-bit Windows (the only build target)
	// its two int32s pack little-endian into one register.
	packed := uintptr(uint32(pt.X)) | uintptr(uint32(pt.Y))<<32
	mon, _, _ := pMonitorFromPoint.Call(packed, monitorDefaultToNearest)
	work, err := monitorWorkArea(mon)
	if err != nil {
		return
	}
	c := Constrain(Rect{Left: int(wr.Left), Top: int(wr.Top), Right: int(wr.Right), Bottom: int(wr.Bottom)}, work)
	wr.Left, wr.Top, wr.Right, wr.Bottom = int32(c.Left), int32(c.Top), int32(c.Right), int32(c.Bottom)
}
