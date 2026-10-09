package storage

import (
	"path/filepath"
	"testing"
	"time"
)

func TestProviderCRUD(t *testing.T) {
	dir := t.TempDir()
	db, err := Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()
	row := ProviderRow{
		ID: "p1", Name: "Kenari", BaseURL: "https://kenari.id/v1",
		Model: "m", CredentialRef: "wortel-provider-p1",
		Enabled: true, CreatedAt: time.Now().UTC(),
	}
	if err := db.SaveProvider(row); err != nil {
		t.Fatalf("save: %v", err)
	}
	list, err := db.ListProviders()
	if err != nil || len(list) != 1 || list[0].Name != "Kenari" {
		t.Fatalf("list: %+v %v", list, err)
	}
	if err := db.SetSetting(ActiveProviderKeyCompat, "p1"); err != nil {
		t.Fatalf("setting: %v", err)
	}
	if v, _ := db.GetSetting(ActiveProviderKeyCompat); v != "p1" {
		t.Fatalf("active: %q", v)
	}
	if err := db.DeleteProvider("p1"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	list, _ = db.ListProviders()
	if len(list) != 0 {
		t.Fatalf("expected empty, got %d", len(list))
	}
}

// ActiveProviderKeyCompat mirrors app.ActiveProviderKey without import cycle.
// Real wiring uses the app constant; value must match.
const ActiveProviderKeyCompat = "active_provider_id"
