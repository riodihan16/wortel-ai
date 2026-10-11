package ai

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func streamOK(t *testing.T, c *OpenAICompatible) string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var got string
	err := c.StreamChat(ctx, ChatRequest{Model: "m", Messages: []Message{{Role: "user", Content: "hi"}}}, func(s string) error {
		got += s
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	return got
}

func streamErr(t *testing.T, c *OpenAICompatible) error {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return c.StreamChat(ctx, ChatRequest{Model: "m", Messages: []Message{{Role: "user", Content: "hi"}}}, func(string) error { return nil })
}

func sseServer(t *testing.T, status int, body string, headers map[string]string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for k, v := range headers {
			w.Header().Set(k, v)
		}
		if ct := w.Header().Get("Content-Type"); ct == "" && status == 200 {
			w.Header().Set("Content-Type", "text/event-stream")
		}
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
}

func TestStreamDeltasFromServer(t *testing.T) {
	srv := sseServer(t, 200, "data: {\"choices\":[{\"delta\":{\"content\":\"halo\"}}]}\n\ndata: {\"choices\":[{\"delta\":{\"content\":\" dunia\"}}]}\n\ndata: [DONE]\n\n", nil)
	defer srv.Close()
	if got := streamOK(t, &OpenAICompatible{BaseURL: srv.URL}); got != "halo dunia" {
		t.Fatalf("got %q", got)
	}
}

func TestErrorMapping429500408400(t *testing.T) {
	for _, tc := range []struct {
		status int
		code   string
		retry  bool
	}{
		{429, "rate_limited", true},
		{500, "provider_server", true},
		{408, "timeout", true},
		{400, "bad_request", false},
	} {
		srv := sseServer(t, tc.status, `{"error":"x"}`, nil)
		pub, ok := streamErr(t, &OpenAICompatible{BaseURL: srv.URL}).(*PublicError)
		srv.Close()
		if !ok || pub.Code != tc.code || pub.Retryable != tc.retry || pub.StatusCode != tc.status {
			t.Fatalf("status %d: got %#v", tc.status, pub)
		}
	}
}

func TestRetryAfterRespected(t *testing.T) {
	srv := sseServer(t, 429, `{"error":"slow"}`, map[string]string{"Retry-After": "7"})
	defer srv.Close()
	pub, ok := streamErr(t, &OpenAICompatible{BaseURL: srv.URL}).(*PublicError)
	if !ok || pub.RetryAfterSec != 7 {
		t.Fatalf("expected retry-after 7, got %#v", pub)
	}
}

func TestMidStreamDisconnectKeepsPartial(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(200)
		_, _ = w.Write([]byte("data: {\"choices\":[{\"delta\":{\"content\":\"parsial\"}}]}\n\n"))
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
		// Hijack lalu tutup kasar: simulasi putus di tengah stream.
		if hj, ok := w.(http.Hijacker); ok {
			if conn, _, err := hj.Hijack(); err == nil {
				conn.Close()
				return
			}
		}
	}))
	defer srv.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var got string
	err := (&OpenAICompatible{BaseURL: srv.URL}).StreamChat(ctx, ChatRequest{Model: "m", Messages: []Message{{Role: "user", Content: "hi"}}}, func(s string) error {
		got += s
		return nil
	})
	if err != nil {
		t.Fatalf("disconnect must end clean with partial, got err: %v", err)
	}
	if got != "parsial" {
		t.Fatalf("partial lost: %q", got)
	}
}

func TestTimeoutSurfaced(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(300 * time.Millisecond)
	}))
	defer srv.Close()
	c := &OpenAICompatible{BaseURL: srv.URL, Client: &http.Client{Timeout: 100 * time.Millisecond}}
	if err := streamErr(t, c); !errors.Is(err, context.DeadlineExceeded) && !isTimeout(err) {
		t.Fatalf("expected timeout, got %#v", err)
	}
}

func isTimeout(err error) bool {
	var te interface{ Timeout() bool }
	return errors.As(err, &te) && te.Timeout()
}

func TestErrorMapping413IsPayloadTooLarge(t *testing.T) {
	srv := sseServer(t, 413, `{"error":{"message":"request entity too large"}}`, nil)
	defer srv.Close()
	var pub *PublicError
	if err := streamErr(t, &OpenAICompatible{BaseURL: srv.URL}); !errors.As(err, &pub) || pub.Code != "payload_too_large" {
		t.Fatalf("got %v", err)
	}
}

func TestError400NamingImagesIsUnsupportedInput(t *testing.T) {
	srv := sseServer(t, 400, `{"error":{"message":"This model does not support image input"}}`, nil)
	defer srv.Close()
	var pub *PublicError
	if err := streamErr(t, &OpenAICompatible{BaseURL: srv.URL}); !errors.As(err, &pub) || pub.Code != "unsupported_input" {
		t.Fatalf("got %v", err)
	}
}

func TestError400NamingModalityIsUnsupportedInput(t *testing.T) {
	for _, body := range []string{
		`{"error":{"message":"unsupported modality: image"}}`,
		`{"error":{"message":"Unsupported modality for this model"}}`,
		`{"error":{"message":"model only supports text modalities"}}`,
		`{"error":{"message":"content type input_image is not allowed"}}`,
		`{"error":{"message":"image_input is disabled for this deployment"}}`,
	} {
		srv := sseServer(t, 400, body, nil)
		var pub *PublicError
		err := streamErr(t, &OpenAICompatible{BaseURL: srv.URL})
		srv.Close()
		if !errors.As(err, &pub) || pub.Code != "unsupported_input" {
			t.Fatalf("%s: got %v", body, err)
		}
	}
}

func TestError400WithoutModalityWordsStaysBadRequest(t *testing.T) {
	srv := sseServer(t, 400, `{"error":{"message":"temperature must be <= 2; check your profile"}}`, nil)
	defer srv.Close()
	var pub *PublicError
	if err := streamErr(t, &OpenAICompatible{BaseURL: srv.URL}); !errors.As(err, &pub) || pub.Code != "bad_request" {
		t.Fatalf("got %v", err)
	}
}
