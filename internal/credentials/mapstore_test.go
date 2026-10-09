package credentials

import (
	"context"
	"testing"
)

func TestMapStoreRoundTrip(t *testing.T) {
	ctx := context.Background()
	m := NewMapStore()
	if err := m.Set(ctx, "ref1", []byte("secret")); err != nil {
		t.Fatalf("set: %v", err)
	}
	got, err := m.Get(ctx, "ref1")
	if err != nil || string(got) != "secret" {
		t.Fatalf("get: %q %v", got, err)
	}
	// Mutating returned copy must not affect stored value.
	got[0] = 'X'
	again, _ := m.Get(ctx, "ref1")
	if string(again) != "secret" {
		t.Fatal("store leaks mutable reference")
	}
	has, _ := m.Exists(ctx, "ref1")
	if !has {
		t.Fatal("exists false after set")
	}
	if err := m.Delete(ctx, "ref1"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := m.Get(ctx, "ref1"); err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestMapStoreRejectsEmpty(t *testing.T) {
	ctx := context.Background()
	m := NewMapStore()
	if err := m.Set(ctx, "", []byte("x")); err == nil {
		t.Fatal("expected ref error")
	}
	if err := m.Set(ctx, "r", nil); err == nil {
		t.Fatal("expected secret error")
	}
}
