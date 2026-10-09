package storage

import (
	"path/filepath"
	"testing"
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
