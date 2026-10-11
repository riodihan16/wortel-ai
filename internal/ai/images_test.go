package ai

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// smallPNG is a real 1x1 PNG so http.DetectContentType sniffs image/png.
var smallPNG = []byte{
	0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a,
	0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52,
	0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
	0x08, 0x02, 0x00, 0x00, 0x00, 0x90, 0x77, 0x53,
	0xde, 0x00, 0x00, 0x00, 0x0c, 0x49, 0x44, 0x41,
	0x54, 0x08, 0xd7, 0x63, 0xf8, 0xcf, 0xc0, 0x00,
	0x00, 0x03, 0x01, 0x01, 0x00, 0x18, 0xdd, 0x8d,
	0xb0, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4e,
	0x44, 0xae, 0x42, 0x60, 0x82,
}

func b64PNG() string { return base64.StdEncoding.EncodeToString(smallPNG) }

func TestGenerateImageDecodesB64(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/images/generations" {
			t.Fatalf("path %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[{"b64_json":"` + b64PNG() + `"}]}`))
	}))
	defer srv.Close()
	mime, data, err := (&OpenAICompatible{BaseURL: srv.URL}).GenerateImage(context.Background(), ImageRequest{Model: "img", Prompt: "a red panda"})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if mime != "image/png" || len(data) != len(smallPNG) {
		t.Fatalf("mime %q len %d", mime, len(data))
	}
}

func TestGenerateImageSendsModelPromptSize(t *testing.T) {
	var got map[string]any
	var auth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth = r.Header.Get("Authorization")
		_ = json.NewDecoder(r.Body).Decode(&got)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[{"b64_json":"` + b64PNG() + `"}]}`))
	}))
	defer srv.Close()
	_, _, err := (&OpenAICompatible{BaseURL: srv.URL, APIKey: "kn-secret"}).GenerateImage(context.Background(), ImageRequest{Model: "img", Prompt: "panda", Size: "1024x1024"})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got["model"] != "img" || got["prompt"] != "panda" || got["size"] != "1024x1024" {
		t.Fatalf("body %#v", got)
	}
	if auth != "Bearer kn-secret" {
		t.Fatalf("auth %q", auth)
	}
}

func TestGenerateImageContextCancel(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer srv.Close()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, _, err := (&OpenAICompatible{BaseURL: srv.URL}).GenerateImage(ctx, ImageRequest{Model: "img", Prompt: "x"})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func TestNormalizeImagesURLNoDoubleV1(t *testing.T) {
	for _, tc := range []struct{ in, want string }{
		{"https://kenari.id/v1", "https://kenari.id/v1/images/generations"},
		{"https://kenari.id/v1/", "https://kenari.id/v1/images/generations"},
		{"https://kenari.id", "https://kenari.id/images/generations"},
		{"https://kenari.id/v1/images/generations", "https://kenari.id/v1/images/generations"},
	} {
		got, err := normalizeImagesURL(tc.in)
		if err != nil || got != tc.want {
			t.Fatalf("%s: got %q err %v", tc.in, got, err)
		}
	}
}

func TestGenerateImageRejectsCredsInURL(t *testing.T) {
	if _, err := normalizeImagesURL("https://user:pass@kenari.id/v1"); err == nil {
		t.Fatal("expected credentials rejection")
	}
}

// keep net/http and time imports used in later tasks without churn.
var _ = time.Second
var _ = strings.TrimSpace
