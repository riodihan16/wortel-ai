package storage

import (
	"path/filepath"
	"testing"
	"time"
)

func TestChatHistoryRoundTrip(t *testing.T) {
	db, err := Open(filepath.Join(t.TempDir(), "chat.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()
	if err := db.EnsureConversation("c1", "Halo dunia"); err != nil {
		t.Fatalf("ensure: %v", err)
	}
	if err := db.AppendMessage(MessageRow{
		ID: "u1", ConversationID: "c1", Role: "user", Content: "halo", Status: "complete",
	}); err != nil {
		t.Fatalf("append user: %v", err)
	}
	if err := db.AppendMessage(MessageRow{
		ID: "a1", ConversationID: "c1", Role: "assistant", Content: "hai juga", Status: "complete",
	}); err != nil {
		t.Fatalf("append asst: %v", err)
	}
	msgs, err := db.RecentMessages("c1", 20)
	if err != nil || len(msgs) != 2 {
		t.Fatalf("recent: %+v %v", msgs, err)
	}
	if msgs[0].Role != "user" || msgs[1].Content != "hai juga" {
		t.Fatalf("order/content: %+v", msgs)
	}
	if err := db.UpdateMessage("a1", "hai juga!", "cancelled"); err != nil {
		t.Fatalf("update: %v", err)
	}
	rows, err := db.RecentRows("c1", 10)
	if err != nil || len(rows) != 2 || rows[1].Status != "cancelled" {
		t.Fatalf("rows: %+v %v", rows, err)
	}
}

func TestConversationCRUD(t *testing.T) {
	db, err := Open(filepath.Join(t.TempDir(), "conv.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()
	if err := db.EnsureConversation("c1", "Satu"); err != nil {
		t.Fatalf("ensure c1: %v", err)
	}
	if err := db.EnsureConversation("c2", "Dua"); err != nil {
		t.Fatalf("ensure c2: %v", err)
	}
	if err := db.AppendMessage(MessageRow{
		ID: "u1", ConversationID: "c1", Role: "user", Content: "halo", Status: "complete",
	}); err != nil {
		t.Fatalf("append: %v", err)
	}
	list, err := db.ListConversations()
	if err != nil || len(list) != 2 {
		t.Fatalf("list: %+v %v", list, err)
	}
	if err := db.RenameConversation("c1", "  Judul Baru  "); err != nil {
		t.Fatalf("rename: %v", err)
	}
	if err := db.RenameConversation("c1", "   "); err == nil {
		t.Fatalf("rename empty should fail")
	}
	if err := db.DeleteConversation("c1"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	rows, err := db.RecentRows("c1", 10)
	if err != nil || len(rows) != 0 {
		t.Fatalf("cascade: %+v %v", rows, err)
	}
	if err := db.DeleteConversation("missing"); err == nil {
		t.Fatalf("delete missing should fail")
	}
}

func TestMigrationV1toV2(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "mig.db")
	db, err := Open(path)
	if err != nil {
		t.Fatalf("open v2: %v", err)
	}
	_ = db.Close()
	// Reopen must keep data (no silent wipe).
	db2, err := Open(path)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	defer db2.Close()
	if err := db2.EnsureConversation("c", "t"); err != nil {
		t.Fatalf("write after reopen: %v", err)
	}
}

// TestMigrationV2toV3 builds a v2 provider row by hand (no models/active_model
// columns), then reopens: migration must add the columns and seed models from
// the legacy single-model column so existing providers stay usable.
func TestMigrationV2toV3(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "mig3.db")
	db, err := Open(path)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	// Simulate a pre-v3 row: drop the new columns and stamp version 2.
	if _, err := db.inner.Exec(`
		CREATE TABLE pp_v2 AS SELECT id, name, base_url, model, credential_ref,
			is_enabled, created_at, updated_at FROM provider_profiles;
		DROP TABLE provider_profiles;
		ALTER TABLE pp_v2 RENAME TO provider_profiles;
		DELETE FROM schema_version WHERE version = 3;
	`); err != nil {
		t.Fatalf("downgrade: %v", err)
	}
	now := time.Now().UTC().Format(time.RFC3339)
	if _, err := db.inner.Exec(`INSERT INTO provider_profiles
		(id, name, base_url, model, credential_ref, is_enabled, created_at, updated_at)
		VALUES ('old','Old','https://a.id/v1','legacy-model','ref-old',1,?,?)`, now, now); err != nil {
		t.Fatalf("seed v2 row: %v", err)
	}
	_ = db.Close()

	db2, err := Open(path)
	if err != nil {
		t.Fatalf("reopen (migrate): %v", err)
	}
	defer db2.Close()
	list, err := db2.ListProviders()
	if err != nil || len(list) != 1 {
		t.Fatalf("list after migration: %+v %v", list, err)
	}
	p := list[0]
	if len(p.Models) != 1 || p.Models[0] != "legacy-model" {
		t.Fatalf("models not seeded: %+v", p.Models)
	}
	if p.ActiveModel != "legacy-model" {
		t.Fatalf("active_model not seeded: %q", p.ActiveModel)
	}
}

// TestProviderModelsRoundTrip covers multi-model persistence + legacy column sync.
func TestProviderModelsRoundTrip(t *testing.T) {
	db, err := Open(filepath.Join(t.TempDir(), "models.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()
	if err := db.SaveProvider(ProviderRow{
		ID: "p1", Name: "n", BaseURL: "https://a.id/v1",
		Models: []string{"m1", "m2", "m3"}, ActiveModel: "m2",
		CredentialRef: "r", Enabled: true,
	}); err != nil {
		t.Fatalf("save: %v", err)
	}
	list, err := db.ListProviders()
	if err != nil || len(list) != 1 {
		t.Fatalf("list: %+v %v", list, err)
	}
	got := list[0]
	if len(got.Models) != 3 || got.Models[2] != "m3" || got.ActiveModel != "m2" {
		t.Fatalf("models round-trip: %+v", got)
	}
	// Legacy column must mirror the active model for old readers.
	if got.Model != "m2" {
		t.Fatalf("legacy model column not synced: %q", got.Model)
	}
}
