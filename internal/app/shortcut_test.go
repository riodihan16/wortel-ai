package app

import (
	"path/filepath"
	"testing"

	"companion/internal/credentials"
	"companion/internal/hotkey"
	"companion/internal/storage"
)

func newShortcutProviders(t *testing.T) *Providers {
	t.Helper()
	db, err := storage.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return NewProviders(db, &credentials.MapStore{})
}

func TestShortcutDefault(t *testing.T) {
	p := newShortcutProviders(t)
	got := p.GetShortcut()
	if !got.Enabled || got.Combo != "Ctrl+Shift+W" {
		t.Fatalf("default shortcut = %+v", got)
	}
}

func TestShortcutRoundTrip(t *testing.T) {
	p := newShortcutProviders(t)
	in := Shortcut{Enabled: false, Combo: "Ctrl+Alt+Q"}
	if err := p.SetShortcut(in); err != nil {
		t.Fatalf("set: %v", err)
	}
	if got := p.GetShortcut(); got != in {
		t.Fatalf("round trip = %+v, want %+v", got, in)
	}
}

func TestShortcutRejectsUnparsable(t *testing.T) {
	p := newShortcutProviders(t)
	if err := p.SetShortcut(Shortcut{Enabled: true, Combo: "Ctrl+Shift+É"}); err == nil {
		t.Fatal("unsupported key must be rejected")
	}
	if err := p.SetShortcut(Shortcut{Enabled: true, Combo: "W"}); err == nil {
		t.Fatal("modifier-less combo must be rejected")
	}
	if got := p.GetShortcut(); got != DefaultShortcut() {
		t.Fatalf("state changed after rejected write: %+v", got)
	}
}

func TestResolvedShortcutFallback(t *testing.T) {
	p := newShortcutProviders(t)
	// Force a corrupt combo directly through storage.
	if err := p.db.SetSetting(keyShortcutCombo, "not-a-combo"); err != nil {
		t.Fatalf("seed: %v", err)
	}
	h, enabled := p.ResolvedShortcut()
	if !enabled || h != hotkey.Default {
		t.Fatalf("fallback = %+v enabled=%v", h, enabled)
	}
}
