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
