package app

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"companion/internal/config"
	"companion/internal/credentials"
	"companion/internal/storage"
)

func testProviders(t *testing.T) *Providers {
	t.Helper()
	db, err := storage.Open(filepath.Join(t.TempDir(), "p.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return NewProviders(db, credentials.NewMapStore())
}

func TestSaveAndResolve(t *testing.T) {
	ctx := context.Background()
	p := testProviders(t)
	saved, err := p.Save(ctx, SaveInput{
		Name: "Kenari", BaseURL: "https://kenari.id/v1", Model: "m", APIKey: "k1",
	})
	if err != nil {
		t.Fatalf("save: %v", err)
	}
	if saved.HasCredential != true || saved.ID == "" {
		t.Fatalf("bad profile: %+v", saved)
	}
	// Secret must not leak into the public profile JSON.
	raw, _ := json.Marshal(saved)
	if strings.Contains(string(raw), "k1") {
		t.Fatal("secret leaked in profile JSON")
	}
	list, err := p.List(ctx)
	if err != nil || len(list) != 1 || !list[0].HasCredential {
		t.Fatalf("list: %+v %v", list, err)
	}
	if p.ActiveID() != saved.ID {
		t.Fatalf("first provider should auto-activate, got %q", p.ActiveID())
	}
	live, model, err := p.Resolve(ctx)
	if err != nil || live == nil || model != "m" {
		t.Fatalf("resolve: %v %v %q", live, err, model)
	}
	if live.APIKey != "k1" {
		t.Fatal("resolve must inject stored key")
	}
}

func TestSaveMultiModelAndSwitch(t *testing.T) {
	ctx := context.Background()
	p := testProviders(t)
	saved, err := p.Save(ctx, SaveInput{
		Name: "Kenari", BaseURL: "https://kenari.id/v1",
		Models: []config.ModelEntry{{ID: "m1"}, {ID: "m2", Alias: "Two"}, {ID: "m3"}}, Model: "m2", APIKey: "k",
	})
	if err != nil {
		t.Fatalf("save: %v", err)
	}
	if len(saved.Models) != 3 || saved.ActiveModel != "m2" {
		t.Fatalf("models not persisted: %+v", saved)
	}
	if saved.Models[1].Alias != "Two" {
		t.Fatalf("alias not persisted: %+v", saved.Models)
	}
	// Resolve must use the active model.
	_, model, err := p.Resolve(ctx)
	if err != nil || model != "m2" {
		t.Fatalf("resolve active: %v %q", err, model)
	}
	// Switch to another model in the list.
	if err := p.SetActiveModel(saved.ID, "m3"); err != nil {
		t.Fatalf("switch: %v", err)
	}
	_, model, err = p.Resolve(ctx)
	if err != nil || model != "m3" {
		t.Fatalf("resolve after switch: %v %q", err, model)
	}
	// Unknown model is rejected (no silent typo).
	if err := p.SetActiveModel(saved.ID, "nope"); err == nil {
		t.Fatal("unknown model must be rejected")
	}
	// ActiveProfile carries the list for the picker.
	prof, ok := p.ActiveProfile(ctx)
	if !ok || len(prof.Models) != 3 || prof.ActiveModel != "m3" {
		t.Fatalf("active profile: %+v ok=%v", prof, ok)
	}
}

func TestSaveRequiresKeyForNew(t *testing.T) {
	p := testProviders(t)
	if _, err := p.Save(context.Background(), SaveInput{
		Name: "n", BaseURL: "https://x.id/v1", Model: "m",
	}); err == nil {
		t.Fatal("expected key error for new provider")
	}
}

func TestDeleteRefusesActive(t *testing.T) {
	ctx := context.Background()
	p := testProviders(t)
	saved, err := p.Save(ctx, SaveInput{
		Name: "n", BaseURL: "https://x.id/v1", Model: "m", APIKey: "k",
	})
	if err != nil {
		t.Fatalf("save: %v", err)
	}
	if err := p.Delete(ctx, saved.ID); err != ErrActiveDelete {
		t.Fatalf("expected ErrActiveDelete, got %v", err)
	}
	if _, err := p.Save(ctx, SaveInput{
		Name: "second", BaseURL: "https://y.id/v1", Model: "m2", APIKey: "k2",
	}); err != nil {
		t.Fatalf("setup second: %v", err)
	}
	rows, _ := p.db.ListProviders()
	for _, r := range rows {
		if r.ID != saved.ID {
			_ = p.SetActive(r.ID)
		}
	}
	if err := p.Delete(ctx, saved.ID); err != nil {
		t.Fatalf("delete after switch: %v", err)
	}
	if has, _ := p.creds.Exists(ctx, saved.CredentialRef); has {
		t.Fatal("credential must be cleaned on delete")
	}
}

func TestTestConnectionMapping(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"bad key"}`))
	}))
	defer srv.Close()
	p := testProviders(t)
	got := p.Test(context.Background(), SaveInput{
		Name: "n", BaseURL: srv.URL, Model: "m", APIKey: "bad",
	})
	if got.OK || got.Code != "unauthorized" {
		t.Fatalf("expected 401 mapping, got %+v", got)
	}
	got = p.Test(context.Background(), SaveInput{
		Name: "n", BaseURL: srv.URL, Model: "m",
	})
	if got.OK || !strings.Contains(got.Message, "API key") {
		t.Fatalf("expected missing key message, got %+v", got)
	}
}

func TestSwitchActiveKeepsHistory(t *testing.T) {
	ctx := context.Background()
	p := testProviders(t)
	first, err := p.Save(ctx, SaveInput{
		Name: "one", BaseURL: "https://a.id/v1", Model: "m1", APIKey: "k1",
	})
	if err != nil {
		t.Fatalf("save one: %v", err)
	}
	second, err := p.Save(ctx, SaveInput{
		Name: "two", BaseURL: "https://b.id/v1", Model: "m2", APIKey: "k2",
	})
	if err != nil {
		t.Fatalf("save two: %v", err)
	}
	// Chat history under "default" with first provider active.
	if err := p.db.EnsureConversation("default", "t"); err != nil {
		t.Fatalf("ensure: %v", err)
	}
	if err := p.db.AppendMessage(storage.MessageRow{
		ID: "u1", ConversationID: "default", Role: "user", Content: "halo", Status: "complete",
	}); err != nil {
		t.Fatalf("append: %v", err)
	}
	if err := p.SetActive(second.ID); err != nil {
		t.Fatalf("switch: %v", err)
	}
	if p.ActiveID() != second.ID {
		t.Fatalf("active not switched: %q", p.ActiveID())
	}
	rows, err := p.db.RecentRows("default", 10)
	if err != nil || len(rows) != 1 || rows[0].Content != "halo" {
		t.Fatalf("history must survive switch: %+v %v", rows, err)
	}
	_ = first
}
