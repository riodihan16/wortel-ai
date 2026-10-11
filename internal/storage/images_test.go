package storage

import (
	"database/sql"
	"errors"
	"path/filepath"
	"testing"
)

func openImgDB(t *testing.T) *DB {
	t.Helper()
	db, err := Open(filepath.Join(t.TempDir(), "img.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func imgBytes() []byte { return []byte{0x89, 'P', 'N', 'G', 0x0d, 0x0a, 0x1a, 0x0a} }

func TestSetMessageImageRoundTrip(t *testing.T) {
	db := openImgDB(t)
	if err := db.EnsureConversation("c1", "t"); err != nil {
		t.Fatal(err)
	}
	id, err := db.AppendImageMessage("c1", "img-model")
	if err != nil {
		t.Fatalf("append: %v", err)
	}
	if id == "" {
		t.Fatal("empty id")
	}
	if err := db.SetMessageImage(id, "image/png", imgBytes()); err != nil {
		t.Fatalf("set: %v", err)
	}
	mime, data, err := db.MessageImage(id)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if mime != "image/png" || len(data) != len(imgBytes()) {
		t.Fatalf("mime %q len %d", mime, len(data))
	}
	if err := db.DeleteConversation("c1"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, _, err := db.MessageImage(id); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected ErrNoRows after cascade, got %v", err)
	}
}

func TestMessageImageAbsentReturnsEmpty(t *testing.T) {
	db := openImgDB(t)
	if err := db.EnsureConversation("c1", "t"); err != nil {
		t.Fatal(err)
	}
	if err := db.AppendMessage(MessageRow{ID: "m1", ConversationID: "c1", Role: "user", Content: "hi", Status: "complete"}); err != nil {
		t.Fatal(err)
	}
	mime, data, err := db.MessageImage("m1")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if mime != "" || data != nil {
		t.Fatalf("expected empty image, got %q %v", mime, data)
	}
}

func TestMigrationV6AddsImageColumns(t *testing.T) {
	// A fresh DB runs all migrations; the columns must exist and be writable.
	db := openImgDB(t)
	var version int
	if err := db.inner.QueryRow(`SELECT COALESCE(MAX(version),0) FROM schema_version`).Scan(&version); err != nil {
		t.Fatal(err)
	}
	if version < 6 {
		t.Fatalf("schema version %d, want >=6", version)
	}
	if err := db.EnsureConversation("c1", "t"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.AppendImageMessage("c1", "m"); err != nil {
		t.Fatalf("append image row: %v", err)
	}
}

func TestMigrationV5toV6KeepsExistingData(t *testing.T) {
	path := filepath.Join(t.TempDir(), "mig6.db")
	db, err := Open(path)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if err := db.EnsureConversation("c1", "lama"); err != nil {
		t.Fatal(err)
	}
	if err := db.AppendMessage(MessageRow{ID: "u1", ConversationID: "c1", Role: "user", Content: "halo", Status: "complete"}); err != nil {
		t.Fatal(err)
	}
	// Roll back to v5: drop the image columns and stamp v5 so reopen runs v6
	// on a populated messages table.
	if _, err := db.inner.Exec(`ALTER TABLE messages DROP COLUMN image_mime`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.inner.Exec(`ALTER TABLE messages DROP COLUMN image_data`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.inner.Exec(`DELETE FROM schema_version WHERE version >= 6`); err != nil {
		t.Fatal(err)
	}
	_ = db.Close()
	db2, err := Open(path)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	defer db2.Close()
	rows, err := db2.RecentRows("c1", 10)
	if err != nil || len(rows) != 1 || rows[0].Content != "halo" {
		t.Fatalf("data lost: %+v %v", rows, err)
	}
	// The new columns must be usable after the migration.
	id, err := db2.AppendImageMessage("c1", "img")
	if err != nil {
		t.Fatalf("append image: %v", err)
	}
	if err := db2.SetMessageImage(id, "image/png", imgBytes()); err != nil {
		t.Fatalf("set image: %v", err)
	}
}

func TestImageRowExcludedFromContext(t *testing.T) {
	db := openImgDB(t)
	if err := db.EnsureConversation("c1", "t"); err != nil {
		t.Fatal(err)
	}
	if err := db.AppendMessage(MessageRow{ID: "u1", ConversationID: "c1", Role: "user", Content: "gambar apa ini", Status: "complete"}); err != nil {
		t.Fatal(err)
	}
	id, err := db.AppendImageMessage("c1", "img")
	if err != nil {
		t.Fatal(err)
	}
	if err := db.SetMessageImage(id, "image/png", imgBytes()); err != nil {
		t.Fatal(err)
	}
	rows, err := db.RecentContext("c1", 20)
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range rows {
		if r.ID == id {
			t.Fatalf("image assistant row must not enter chat context")
		}
	}
}
