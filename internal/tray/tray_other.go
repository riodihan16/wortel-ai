//go:build !windows

package tray

// Non-Windows stub: Tray only runs on Windows (MVP target).
// Start/Stop are no-ops so tests and builds stay green elsewhere.
func Start(c Callbacks) error { return nil }

func Stop() {}

// Running always false off Windows.
func Running() bool { return false }

// LastError always empty off Windows.
func LastError() string { return "" }
