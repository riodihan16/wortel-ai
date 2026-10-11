package storage

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
)

// TestOpenRejectsEmptyPath must fail fast with a clear error, never panic,
// so the app can fall back and tell the user (FR-009).
func TestOpenRejectsEmptyPath(t *testing.T) {
	if _, err := Open(""); err == nil {
		t.Fatal("expected error for empty path")
	}
}

// TestOpenFailsWhenDirUncreatable: a path under a file (not a dir) cannot have
// its parent created. Open must return an error, not panic.
func TestOpenFailsWhenDirUncreatable(t *testing.T) {
	base := t.TempDir()
	file := filepath.Join(base, "notadir")
	if err := writeEmptyFile(file); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if _, err := Open(filepath.Join(file, "sub", "x.db")); err == nil {
		t.Fatal("expected error when parent dir cannot be created")
	}
}

// TestUseAfterCloseDoesNotPanic: queries on a closed handle must return an
// error, so a DB failure surfaces as a message rather than a crash.
func TestUseAfterCloseDoesNotPanic(t *testing.T) {
	db, err := Open(filepath.Join(t.TempDir(), "x.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
	if _, err := db.RecentContext("c1", 5); err == nil {
		t.Fatal("expected error on closed handle")
	}
	if err := db.AppendMessage(MessageRow{ID: "m1", ConversationID: "c1", Role: "user", Content: "x", Status: "complete"}); err == nil {
		t.Fatal("expected write error on closed handle")
	}
}

// TestForeignKeysOnEveryConnection: the FK pragma is per connection, so it
// must come from the DSN, not from a one-off Exec on the pool.
func TestForeignKeysOnEveryConnection(t *testing.T) {
	db, err := Open(filepath.Join(t.TempDir(), "with space", "fk.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()
	ctx := context.Background()
	c1, err := db.inner.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer c1.Close()
	c2, err := db.inner.Conn(ctx) // forces a second physical connection while c1 is held
	if err != nil {
		t.Fatal(err)
	}
	defer c2.Close()
	for i, c := range []*sql.Conn{c1, c2} {
		var on int
		if err := c.QueryRowContext(ctx, `PRAGMA foreign_keys`).Scan(&on); err != nil || on != 1 {
			t.Fatalf("conn %d: foreign_keys=%d err=%v", i, on, err)
		}
	}
}
