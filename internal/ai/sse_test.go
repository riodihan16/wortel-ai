package ai

import (
	"bufio"
	"context"
	"strings"
	"testing"
)

func collect(t *testing.T, input string) []StreamChunk {
	t.Helper()
	var out []StreamChunk
	err := ScanSSE(context.Background(), bufio.NewReader(strings.NewReader(input)), func(c StreamChunk) error {
		out = append(out, c)
		return nil
	})
	if err != nil {
		t.Fatalf("ScanSSE err: %v", err)
	}
	return out
}

func TestSSEBasicDeltas(t *testing.T) {
	in := "data: {\"choices\":[{\"delta\":{\"content\":\"halo\"}}]}\n\n" +
		"data: {\"choices\":[{\"delta\":{\"content\":\" dunia\"}}]}\n\n" +
		"data: [DONE]\n\n"
	got := collect(t, in)
	if len(got) != 3 || got[0].Text != "halo" || got[1].Text != " dunia" || !got[2].Done {
		t.Fatalf("unexpected chunks: %+v", got)
	}
}

func TestSSEIgnoresCommentsAndMalformed(t *testing.T) {
	in := ": keepalive\n\n" +
		"data: not-json\n\n" +
		"data: {\"choices\":[{\"delta\":{\"content\":\"ok\"}}]}\n\n"
	got := collect(t, in)
	if len(got) != 1 || got[0].Text != "ok" {
		t.Fatalf("unexpected chunks: %+v", got)
	}
}

func TestSSEMultiDataLinesJoined(t *testing.T) {
	in := "data: {\"choices\":\n" +
		"data: [{\"delta\":{\"content\":\"x\"}}]}\n\n"
	got := collect(t, in)
	if len(got) != 1 || got[0].Text != "x" {
		t.Fatalf("unexpected chunks: %+v", got)
	}
}

func TestSSECancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := ScanSSE(ctx, bufio.NewReader(strings.NewReader("data: {}\n\n")), func(StreamChunk) error { return nil })
	if err == nil {
		t.Fatal("expected cancel error")
	}
}

func TestNormalizeBaseURL(t *testing.T) {
	u, err := NormalizeBaseURL("https://api.example.com/v1")
	if err != nil || u != "https://api.example.com/v1/chat/completions" {
		t.Fatalf("got %q %v", u, err)
	}
	u, err = NormalizeBaseURL("https://gateway.example.com/openai/v1/")
	if err != nil || u != "https://gateway.example.com/openai/v1/chat/completions" {
		t.Fatalf("got %q %v", u, err)
	}
	if _, err := NormalizeBaseURL("ftp://x"); err == nil {
		t.Fatal("expected scheme error")
	}
	if _, err := NormalizeBaseURL("https://user:pass@x.com/v1"); err == nil {
		t.Fatal("expected credential error")
	}
}

func TestMapHTTPError(t *testing.T) {
	for _, tc := range []struct {
		status int
		code   string
	}{
		{401, "unauthorized"}, {403, "forbidden"}, {404, "not_found"},
		{429, "rate_limited"}, {500, "provider_server"},
	} {
		if got := MapHTTPError(tc.status); got.Code != tc.code {
			t.Fatalf("status %d: got %q", tc.status, got.Code)
		}
	}
}
