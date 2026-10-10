//go:build windows

package window

import "fmt"

var (
	pSetForegroundWindow = user32.NewProc("SetForegroundWindow")
	pGetForegroundWindow = user32.NewProc("GetForegroundWindow")
)

// IsForeground reports whether the companion is the window the user is
// typing in. An unknown window counts as not foreground, so the shortcut
// brings it forward rather than collapsing it.
func IsForeground() bool {
	mu.Lock() // mainWindow's cache is shared with Apply and the drag guard
	h, err := mainWindow()
	mu.Unlock()
	if err != nil {
		return false
	}
	fg, _, _ := pGetForegroundWindow.Call()
	return fg == uintptr(h)
}

// Activate makes the companion the foreground window so keys typed right
// after the global shortcut reach the composer instead of the app the user
// was in. The process that receives WM_HOTKEY may take the foreground, so
// this succeeds from the shortcut path (ADR-051).
func Activate() error {
	mu.Lock()
	h, err := mainWindow()
	mu.Unlock()
	if err != nil {
		return err
	}
	if r, _, err := pSetForegroundWindow.Call(uintptr(h)); r == 0 {
		return fmt.Errorf("SetForegroundWindow: %w", err)
	}
	return nil
}
