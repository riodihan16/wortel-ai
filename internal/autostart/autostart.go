package autostart

// Autostart toggles "run at Windows login" via the per-user Run registry key.
// Per-user (HKCU) only: no admin rights, easy to remove. Default is OFF;
// the user must opt in from Settings (FR-011).

// RunKey is the per-user autostart registry path.
const RunKey = `Software\Microsoft\Windows\CurrentVersion\Run`

// ValueName identifies our entry under RunKey.
const ValueName = "WortelAI"

// ErrUnsupported is returned on non-Windows builds.
var ErrUnsupported = errUnsupported

// Enable registers the app to start at login using exePath.
// Replaces any existing entry. exePath must be absolute.
func Enable(exePath string) error { return enable(exePath) }

// Disable removes the autostart entry. Missing entry is not an error.
func Disable() error { return disable() }

// Enabled reports whether the autostart entry exists.
func Enabled() bool { return enabled() }
