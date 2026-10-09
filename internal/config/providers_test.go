package config

import "testing"

func TestValidateOK(t *testing.T) {
	ep, insecure, err := Validate(ProviderInput{
		Name: "Kenari", BaseURL: "https://kenari.id/v1", Model: "m",
	})
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if ep != "https://kenari.id/v1/chat/completions" {
		t.Fatalf("bad endpoint: %q", ep)
	}
	if insecure {
		t.Fatal("https must not flag insecure")
	}
}

func TestValidateRejects(t *testing.T) {
	for _, in := range []ProviderInput{
		{BaseURL: "https://x.id/v1", Model: "m"},
		{Name: "n", BaseURL: "https://x.id/v1"},
		{Name: "n", BaseURL: "ftp://x.id/v1", Model: "m"},
		{Name: "n", BaseURL: "https://user:pw@x.id/v1", Model: "m"},
	} {
		if _, _, err := Validate(in); err == nil {
			t.Fatalf("expected error for %+v", in)
		}
	}
}

func TestValidateHTTPFlagsInsecure(t *testing.T) {
	_, insecure, err := Validate(ProviderInput{
		Name: "n", BaseURL: "http://192.168.1.5:8080/v1", Model: "m",
	})
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if !insecure {
		t.Fatal("non-loopback http must flag insecure")
	}
	_, insecure, err = Validate(ProviderInput{
		Name: "n", BaseURL: "http://localhost:8080/v1", Model: "m",
	})
	if err != nil || insecure {
		t.Fatalf("loopback http ok, no flag: %v %v", err, insecure)
	}
}
