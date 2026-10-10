//go:build windows

package tray

import (
	"os"
	"path/filepath"
	"time"
)

// trace appends a timestamped line to %LOCALAPPDATA%\WortelAI\tray.log.
// Diagnostic only; errors ignored. Remove once tray is stable.
func trace(msg string) {
	base := os.Getenv("LOCALAPPDATA")
	if base == "" {
		return
	}
	dir := filepath.Join(base, "WortelAI")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return
	}
	f, err := os.OpenFile(filepath.Join(dir, "tray.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer f.Close()
	_, _ = f.WriteString(time.Now().Format("15:04:05.000") + " " + msg + "\n")
}
