package config

import "testing"

func TestValidateOK(t *testing.T) {
	ep, models, active, insecure, err := Validate(ProviderInput{
		Name: "Kenari", BaseURL: "https://kenari.id/v1",
		Models: []string{"gpt-4o", "gpt-4o-mini"},
	})
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if ep != "https://kenari.id/v1/chat/completions" {
		t.Fatalf("bad endpoint: %q", ep)
	}
	if len(models) != 2 || models[0] != "gpt-4o" {
		t.Fatalf("bad models: %v", models)
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
	if len(models) != 1 || models[0] != "m" || active != "m" {
		t.Fatalf("single model not seeded: %v active=%q", models, active)
	}
}

func TestNormalizeModelsDedupesAndPreservesActive(t *testing.T) {
	models, active := NormalizeModels(
		[]string{" a ", "b", "a", "", "  ", "b"},
		"b",
	)
	if len(models) != 2 || models[0] != "a" || models[1] != "b" {
		t.Fatalf("dedupe/trim failed: %v", models)
	}
	if active != "b" {
		t.Fatalf("active kept: %q", active)
	}
	// Unknown active falls back to the first entry.
	_, active = NormalizeModels([]string{"x", "y"}, "zzz")
	if active != "x" {
		t.Fatalf("unknown active must fall back to first: %q", active)
	}
}

func TestNormalizeModelsCaps(t *testing.T) {
	in := make([]string, 0, MaxModels+10)
	for i := range MaxModels + 10 {
		in = append(in, string(rune('a'+i%26))+string(rune('a'+i/26)))
	}
	models, _ := NormalizeModels(in, "")
	if len(models) != MaxModels {
		t.Fatalf("cap: want %d got %d", MaxModels, len(models))
	}
}

func TestValidateRejects(t *testing.T) {
	for _, in := range []ProviderInput{
		{BaseURL: "https://x.id/v1", Models: []string{"m"}},
		{Name: "n", BaseURL: "https://x.id/v1"},
		{Name: "n", BaseURL: "https://x.id/v1", Models: []string{"", "  "}},
		{Name: "n", BaseURL: "ftp://x.id/v1", Models: []string{"m"}},
		{Name: "n", BaseURL: "https://user:pw@x.id/v1", Models: []string{"m"}},
	} {
		if _, _, _, _, err := Validate(in); err == nil {
			t.Fatalf("expected error for %+v", in)
		}
	}
}

func TestValidateHTTPFlagsInsecure(t *testing.T) {
	_, _, _, insecure, err := Validate(ProviderInput{
		Name: "n", BaseURL: "http://192.168.1.5:8080/v1", Models: []string{"m"},
	})
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if !insecure {
		t.Fatal("non-loopback http must flag insecure")
	}
	_, _, _, insecure, err = Validate(ProviderInput{
		Name: "n", BaseURL: "http://localhost:8080/v1", Models: []string{"m"},
	})
	if err != nil || insecure {
		t.Fatalf("loopback http ok, no flag: %v %v", err, insecure)
	}
}
