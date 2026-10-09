package app

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

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
