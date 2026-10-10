package config

import (
	"strings"
	"testing"
)

func TestValidateOK(t *testing.T) {
	ep, models, active, insecure, err := Validate(ProviderInput{
		Name: "Kenari", BaseURL: "https://kenari.id/v1",
		Models: []ModelEntry{{ID: "gpt-4o"}, {ID: "gpt-4o-mini", Alias: "Mini"}},
	})
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if ep != "https://kenari.id/v1/chat/completions" {
		t.Fatalf("bad endpoint: %q", ep)
	}
	if len(models) != 2 || models[0].ID != "gpt-4o" {
		t.Fatalf("bad models: %v", models)
	}
	if models[1].Alias != "Mini" {
		t.Fatalf("alias not kept: %v", models)
	}
	if active != "gpt-4o" {
		t.Fatalf("active must default to first model, got %q", active)
	}
	if insecure {
		t.Fatal("https must not flag insecure")
	}
}

func TestValidateSingleModelFallback(t *testing.T) {
	// Legacy single Model field still works and seeds the list.
	_, models, active, _, err := Validate(ProviderInput{
		Name: "n", BaseURL: "https://x.id/v1", Model: "m",
	})
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if len(models) != 1 || models[0].ID != "m" || active != "m" {
		t.Fatalf("single model not seeded: %v active=%q", models, active)
	}
}

func TestNormalizeModelsDedupesAndPreservesActive(t *testing.T) {
	models, active := NormalizeModels(
		[]ModelEntry{{ID: " a "}, {ID: "b"}, {ID: "a"}, {ID: ""}, {ID: "  "}, {ID: "b"}},
		"b",
	)
	if len(models) != 2 || models[0].ID != "a" || models[1].ID != "b" {
		t.Fatalf("dedupe/trim failed: %v", models)
	}
	if active != "b" {
		t.Fatalf("active kept: %q", active)
	}
	// Unknown active falls back to the first entry.
	_, active = NormalizeModels([]ModelEntry{{ID: "x"}, {ID: "y"}}, "zzz")
	if active != "x" {
		t.Fatalf("unknown active must fall back to first: %q", active)
	}
}

func TestNormalizeModelsCaps(t *testing.T) {
	in := make([]ModelEntry, 0, MaxModels+10)
	for i := range MaxModels + 10 {
		in = append(in, ModelEntry{ID: string(rune('a'+i%26)) + string(rune('a'+i/26))})
	}
	models, _ := NormalizeModels(in, "")
	if len(models) != MaxModels {
		t.Fatalf("cap: want %d got %d", MaxModels, len(models))
	}
}

func TestNormalizeModelsAliases(t *testing.T) {
	long := strings.Repeat("a", 100)
	models, active := NormalizeModels([]ModelEntry{
		{ID: "  m1  ", Alias: "  Fast  "},
		{ID: "m1", Alias: "dup"},
		{ID: "   ", Alias: "blank"},
		{ID: "m2", Alias: long},
	}, "missing")
	if len(models) != 2 {
		t.Fatalf("want 2 entries, got %v", models)
	}
	if models[0].ID != "m1" || models[0].Alias != "Fast" {
		t.Fatalf("trim failed: %+v", models[0])
	}
	if models[1].ID != "m2" || len(models[1].Alias) != 64 {
		t.Fatalf("alias must truncate to 64: %q", models[1].Alias)
	}
	if active != "m1" {
		t.Fatalf("active must fall back to first: %q", active)
	}
	// Empty list + empty active -> empty.
	clean, act := NormalizeModels(nil, "")
	if len(clean) != 0 || act != "" {
		t.Fatalf("empty must stay empty: %v %q", clean, act)
	}
}

func TestValidateRejects(t *testing.T) {
	for _, in := range []ProviderInput{
		{BaseURL: "https://x.id/v1", Models: []ModelEntry{{ID: "m"}}},
		{Name: "n", BaseURL: "https://x.id/v1"},
		{Name: "n", BaseURL: "https://x.id/v1", Models: []ModelEntry{{ID: ""}, {ID: "  "}}},
		{Name: "n", BaseURL: "ftp://x.id/v1", Models: []ModelEntry{{ID: "m"}}},
		{Name: "n", BaseURL: "https://user:pw@x.id/v1", Models: []ModelEntry{{ID: "m"}}},
	} {
		if _, _, _, _, err := Validate(in); err == nil {
			t.Fatalf("expected error for %+v", in)
		}
	}
}

func TestValidateHTTPFlagsInsecure(t *testing.T) {
	_, _, _, insecure, err := Validate(ProviderInput{
		Name: "n", BaseURL: "http://192.168.1.5:8080/v1", Models: []ModelEntry{{ID: "m"}},
	})
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if !insecure {
		t.Fatal("non-loopback http must flag insecure")
	}
	_, _, _, insecure, err = Validate(ProviderInput{
		Name: "n", BaseURL: "http://localhost:8080/v1", Models: []ModelEntry{{ID: "m"}},
	})
	if err != nil || insecure {
		t.Fatalf("loopback http ok, no flag: %v %v", err, insecure)
	}
}
