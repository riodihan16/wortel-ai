package app

import (
	"path/filepath"
	"testing"

	"companion/internal/credentials"
	"companion/internal/storage"
)

func newAppearanceProviders(t *testing.T) *Providers {
	t.Helper()
	db, err := storage.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return NewProviders(db, &credentials.MapStore{})
}

func TestAppearanceDefaultsWhenUnset(t *testing.T) {
	p := newAppearanceProviders(t)
	got := p.GetAppearance()
	want := DefaultAppearance()
	if got != want {
		t.Fatalf("default appearance = %+v, want %+v", got, want)
	}
}

func TestAppearanceRoundTrip(t *testing.T) {
	p := newAppearanceProviders(t)
	in := Appearance{Theme: ThemeLight, FontScale: 1.2}
	if err := p.SetAppearance(in); err != nil {
		t.Fatalf("set: %v", err)
	}
	if got := p.GetAppearance(); got != in {
		t.Fatalf("round trip = %+v, want %+v", got, in)
	}
}

func TestAppearanceRejectsInvalid(t *testing.T) {
	p := newAppearanceProviders(t)
	if err := p.SetAppearance(Appearance{Theme: "neon", FontScale: 1}); err == nil {
		t.Fatal("unknown theme must be rejected")
	}
	if err := p.SetAppearance(Appearance{Theme: ThemeDark, FontScale: 3}); err == nil {
		t.Fatal("out-of-range scale must be rejected")
	}
	// Rejected writes must not corrupt stored state.
	if got := p.GetAppearance(); got != DefaultAppearance() {
		t.Fatalf("state changed after rejected write: %+v", got)
	}
}
