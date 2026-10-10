//go:build windows

package window

import "fmt"

var pSetForegroundWindow = user32.NewProc("SetForegroundWindow")

// Activate makes the companion the foreground window so keys typed right
// after the global shortcut reach the composer instead of the app the user
// was in. The process that receives WM_HOTKEY may take the foreground, so
// this succeeds from the shortcut path (ADR-051).
func Activate() error {
	h, err := mainWindow()
	if err != nil {
		return err
	}
	if r, _, _ := pSetForegroundWindow.Call(uintptr(h)); r == 0 {
		return fmt.Errorf("SetForegroundWindow refused")
	}
	return nil
}
