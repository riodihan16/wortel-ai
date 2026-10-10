//go:build windows

package autostart

import (
	"os"
	"testing"
)

// TestRoundTrip touches the real HKCU Run key. It restores prior state,
// so it is safe to run on a dev machine.
func TestRoundTrip(t *testing.T) {
	priorEnabled := Enabled()
	priorValue := readRaw()

	defer func() {
		if priorEnabled {
			exe, _ := os.Executable()
			_ = Enable(exe)
		} else {
			_ = Disable()
		}
		_ = priorValue
	}()

	_ = Disable()
	if Enabled() {
		t.Fatal("Disabled state should report Enabled()==false")
	}

	exe, err := os.Executable()
	if err != nil {
		t.Fatalf("executable: %v", err)
	}
	if err := Enable(exe); err != nil {
		t.Fatalf("Enable: %v", err)
	}
	if !Enabled() {
		t.Fatal("Enabled() should be true after Enable")
	}
	if got := readRaw(); got != `"`+exe+`"` {
		t.Fatalf("registry value = %q, want quoted exe", got)
	}

	if err := Disable(); err != nil {
		t.Fatalf("Disable: %v", err)
	}
	if Enabled() {
		t.Fatal("Enabled() should be false after Disable")
	}
	// Disable twice must not error.
	if err := Disable(); err != nil {
		t.Fatalf("second Disable: %v", err)
	}
}

func TestEnableEmptyPath(t *testing.T) {
	if err := Enable(""); err == nil {
		t.Fatal("Enable(\"\") should error")
	}
}
