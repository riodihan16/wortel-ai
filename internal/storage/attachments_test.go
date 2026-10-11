package storage

import (
	"path/filepath"
	"testing"
)

func openTemp(t *testing.T) *DB {
	t.Helper()
	db, err := Open(filepath.Join(t.TempDir(), "a.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func countAttachments(t *testing.T, db *DB) int {
	t.Helper()
	var n int
	if err := db.inner.QueryRow(`SELECT COUNT(*) FROM attachments`).Scan(&n); err != nil {
		t.Fatalf("count: %v", err)
	}
	return n
}

func userRow(id, text string) MessageRow {
	return MessageRow{ID: id, ConversationID: "c1", Role: "user", Content: text, Status: "complete"}
}

func asstRow(id string) MessageRow {
	return MessageRow{ID: id, ConversationID: "c1", Role: "assistant", Content: "", Status: "streaming"}
}

func TestMigrationV5CreatesAttachments(t *testing.T) {
	db := openTemp(t)
	var v int
	if err := db.inner.QueryRow(`SELECT MAX(version) FROM schema_version`).Scan(&v); err != nil || v != 5 {
		t.Fatalf("version %d %v", v, err)
	}
	if countAttachments(t, db) != 0 {
		t.Fatal("fresh table not empty")
	}
}

func TestMigrationV4toV5KeepsExistingData(t *testing.T) {
	path := filepath.Join(t.TempDir(), "old.db")
	db, err := Open(path)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if err := db.EnsureConversation("c1", "lama"); err != nil {
		t.Fatal(err)
	}
	if err := db.AppendMessage(userRow("u1", "halo")); err != nil {
		t.Fatal(err)
	}
	// Roll the schema back to v4 so the next Open must run migration 5 on a
	// database that already holds conversations.
	if _, err := db.inner.Exec(`DROP TABLE attachments`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.inner.Exec(`DELETE FROM schema_version WHERE version=5`); err != nil {
		t.Fatal(err)
	}
	_ = db.Close()
	db2, err := Open(path)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	defer db2.Close()
	if countAttachments(t, db2) != 0 {
		t.Fatal("attachments missing after upgrade")
	}
	convs, err := db2.ListConversations()
	if err != nil || len(convs) != 1 || convs[0].Title != "lama" {
		t.Fatalf("data lost: %+v %v", convs, err)
	}
}

func TestAppendUserTurnIsAtomic(t *testing.T) {
	db := openTemp(t)
	_ = db.EnsureConversation("c1", "x")
	// Two attachments with the same id: the second insert fails, so nothing
	// from this turn may remain.
	err := db.AppendUserTurn(userRow("u1", "lihat"), []AttachmentRow{
		{ID: "a1", Kind: "image", Mime: "image/png", Name: "x.png", Data: []byte{1}},
		{ID: "a1", Kind: "image", Mime: "image/png", Name: "y.png", Data: []byte{2}},
	}, asstRow("m1"))
	if err == nil {
		t.Fatal("expected duplicate id error")
	}
	rows, _ := db.RecentRows("c1", 10)
	if len(rows) != 0 || countAttachments(t, db) != 0 {
		t.Fatalf("partial turn stored: %+v", rows)
	}
}

func TestAppendUserTurnStoresAndCascades(t *testing.T) {
	db := openTemp(t)
	_ = db.EnsureConversation("c1", "x")
	err := db.AppendUserTurn(userRow("u1", "lihat"), []AttachmentRow{
		{ID: "a1", Kind: "image", Mime: "image/png", Name: "x.png", Data: []byte{1}},
		{ID: "a2", Kind: "pdf", Mime: "application/pdf", Name: "nota.pdf", Data: []byte("%PDF")},
	}, asstRow("m1"))
	if err != nil {
		t.Fatalf("append: %v", err)
	}
	atts, err := db.AttachmentsFor([]string{"u1", "m1"})
	if err != nil || len(atts["u1"]) != 2 || len(atts["m1"]) != 0 {
		t.Fatalf("for: %+v %v", atts, err)
	}
	if a := atts["u1"][1]; a.Name != "nota.pdf" || a.Position != 1 || a.Size != 4 || a.Data != nil {
		t.Fatalf("metadata: %+v", a)
	}
	mime, data, err := db.AttachmentData("a1")
	if err != nil || mime != "image/png" || len(data) != 1 || data[0] != 1 {
		t.Fatalf("data: %q %v %v", mime, data, err)
	}
	ctx, err := db.RecentContext("c1", 20)
	if err != nil || len(ctx) != 1 || ctx[0].ID != "u1" || len(ctx[0].Attachments) != 2 || ctx[0].Attachments[0].Data != nil {
		t.Fatalf("context: %+v %v", ctx, err)
	}
	if err := db.DeleteConversation("c1"); err != nil {
		t.Fatal(err)
	}
	if countAttachments(t, db) != 0 {
		t.Fatal("cascade did not remove attachments")
	}
	if _, _, err := db.AttachmentData("a1"); err == nil {
		t.Fatal("deleted attachment still readable")
	}
}

func TestRecentContextSkipsEmptyAssistantRowAndKeepsOrder(t *testing.T) {
	db := openTemp(t)
	_ = db.EnsureConversation("c1", "x")
	if err := db.AppendUserTurn(userRow("u1", "satu"), nil, asstRow("m1")); err != nil {
		t.Fatal(err)
	}
	_ = db.UpdateMessage("m1", "jawab", "complete")
	if err := db.AppendUserTurn(userRow("u2", "dua"), nil, asstRow("m2")); err != nil {
		t.Fatal(err)
	}
	ctx, err := db.RecentContext("c1", 20)
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, r := range ctx {
		got = append(got, r.Role+":"+r.Content)
	}
	want := []string{"user:satu", "assistant:jawab", "user:dua"}
	if len(got) != len(want) {
		t.Fatalf("got %v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v want %v", got, want)
		}
	}
}
