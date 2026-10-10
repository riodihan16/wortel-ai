//go:build windows

package hotkey

import "testing"

// TestRebindSwapsCombo registers a combo, stops it, then registers a
// different one — the path App.applyShortcut takes on every change.
func TestRebindSwapsCombo(t *testing.T) {
	h1, err := Parse("Ctrl+Shift+W")
	if err != nil {
		t.Fatalf("parse h1: %v", err)
	}
	if err := Start(h1, func() {}); err != nil {
		t.Fatalf("start h1: %v", err)
	}
	Stop()

	h2, err := Parse("Ctrl+Alt+Q")
	if err != nil {
		t.Fatalf("parse h2: %v", err)
	}
	if err := Start(h2, func() {}); err != nil {
		t.Fatalf("start h2 after stop: %v", err)
	}
	Stop()
}

// TestStopWhenNeverStarted must not panic or block.
func TestStopWhenNeverStarted(t *testing.T) {
	Stop()
	Stop()
}
