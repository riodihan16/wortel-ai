package app

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"companion/internal/ai"
	"companion/internal/credentials"
	"companion/internal/storage"
)

// imageSink collects emitted ImageEvents.
type imageSink struct {
	mu     sync.Mutex
	events []ImageEvent
}

func (s *imageSink) emit(_ string, p any) {
	ev, ok := p.(ImageEvent)
	if !ok {
		return
	}
	s.mu.Lock()
	s.events = append(s.events, ev)
	s.mu.Unlock()
}

func (s *imageSink) snapshot() []ImageEvent {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]ImageEvent(nil), s.events...)
}

func (s *imageSink) has(t string) bool {
	for _, e := range s.snapshot() {
		if e.Type == t {
			return true
		}
	}
	return false
}

// waitImageDone waits until the terminal image event appears and the
// generation is deregistered.
func waitImageDone(t *testing.T, svc *Service, sink *imageSink) ImageEvent {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		for _, e := range sink.snapshot() {
			if e.Type == "completed" || e.Type == "cancelled" || e.Type == "failed" {
				svc.mu.Lock()
				n := len(svc.imageActive)
				svc.mu.Unlock()
				if n == 0 {
					return e
				}
			}
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("no terminal image event: %+v", sink.snapshot())
	return ImageEvent{}
}

func b64ImageServer(t *testing.T) http.HandlerFunc {
	t.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/images/generations" {
			t.Errorf("path %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[{"b64_json":"` + pngB64() + `"}]}`))
	}
}

func TestGenerateEmitsImageEvents(t *testing.T) {
	svc, _ := serviceWith(t, b64ImageServer(t))
	var sink imageSink
	svc.SetImageEmitter(sink.emit)
	id, err := svc.GenerateImage(context.Background(), "c1", "a red panda", "img-model")
	if err != nil || id == "" {
		t.Fatalf("generate: %v id %q", err, id)
	}
	done := waitImageDone(t, svc, &sink)
	if done.Type != "completed" || done.Mime != "image/png" {
		t.Fatalf("terminal %+v", done)
	}
	if !sink.has("started") {
		t.Fatal("missing started event")
	}
	// Stored image is served by the same asset route.
	mime, data, err := svc.db().MessageImage(done.MessageID)
	if err != nil || mime != "image/png" || len(data) != len(pngBytes) {
		t.Fatalf("stored image mime %q len %d err %v", mime, len(data), err)
	}
}

func TestGenerateImageRejectsEmptyPrompt(t *testing.T) {
	svc, _ := serviceWith(t, b64ImageServer(t))
	if _, err := svc.GenerateImage(context.Background(), "c1", "   ", "img-model"); err == nil {
		t.Fatal("expected empty prompt rejection")
	}
}

func TestGenerateImageOneActivePerConversation(t *testing.T) {
	block := make(chan struct{})
	svc, _ := serviceWith(t, func(w http.ResponseWriter, r *http.Request) {
		<-block
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[{"b64_json":"` + pngB64() + `"}]}`))
	})
	if _, err := svc.GenerateImage(context.Background(), "c1", "x", "img"); err != nil {
		t.Fatalf("first: %v", err)
	}
	if _, err := svc.GenerateImage(context.Background(), "c1", "y", "img"); err == nil {
		t.Fatal("expected second generation rejection")
	}
	close(block)
}

func TestGenerateImageLongTimeout(t *testing.T) {
	// The image generation budget must be far longer than the chat client's
	// short timeout (Kenari docs §22.4); assert the named constant.
	if imageTimeout < 240*time.Second {
		t.Fatalf("imageTimeout %v too short for image generation", imageTimeout)
	}
}

func TestGenerateImageURLResponseStoredLocally(t *testing.T) {
	imgSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write(pngBytes)
	}))
	t.Cleanup(imgSrv.Close)
	svc, _ := serviceWith(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[{"url":"` + imgSrv.URL + `/i.png"}]}`))
	})
	var sink imageSink
	svc.SetImageEmitter(sink.emit)
	if _, err := svc.GenerateImage(context.Background(), "c1", "panda", "img"); err != nil {
		t.Fatalf("generate: %v", err)
	}
	done := waitImageDone(t, svc, &sink)
	if done.Type != "completed" || done.Mime != "image/png" {
		t.Fatalf("terminal %+v", done)
	}
}

func TestAttachmentHandlerServesMessageImage(t *testing.T) {
	svc, _ := serviceWith(t, b64ImageServer(t))
	var sink imageSink
	svc.SetImageEmitter(sink.emit)
	if _, err := svc.GenerateImage(context.Background(), "c1", "panda", "img-model"); err != nil {
		t.Fatalf("generate: %v", err)
	}
	done := waitImageDone(t, svc, &sink)
	rec := httptest.NewRecorder()
	svc.AttachmentHandler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/attachments/"+done.MessageID, nil))
	if rec.Code != 200 || rec.Header().Get("Content-Type") != "image/png" || rec.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Fatalf("code %d headers %v", rec.Code, rec.Header())
	}
	if rec.Body.Len() != len(pngBytes) {
		t.Fatalf("body %d bytes", rec.Body.Len())
	}
}

// serviceNoProvider wires storage but saves no provider, so image generation
// resolves to "no active provider".
func serviceNoProvider(t *testing.T) (*Service, *imageSink) {
	t.Helper()
	db, err := storage.Open(filepath.Join(t.TempDir(), "np.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	p := NewProviders(db, credentials.NewMapStore())
	svc := NewService(&ai.MockProvider{}, func(string, ai.ChatEvent) {})
	svc.SetProviders(p)
	sink := &imageSink{}
	svc.SetImageEmitter(sink.emit)
	return svc, sink
}

func TestGenerateRejectsWhenNoProvider(t *testing.T) {
	svc, sink := serviceNoProvider(t)
	if _, err := svc.GenerateImage(context.Background(), "c1", "panda", "img"); err != nil {
		t.Fatalf("generate returned sync err: %v", err)
	}
	var term ImageEvent
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		for _, e := range sink.snapshot() {
			if e.Type == "failed" {
				term = e
			}
		}
		if term.Type == "failed" {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if term.Type != "failed" || term.Error == nil || term.Error.Code != "no_provider" {
		t.Fatalf("expected no_provider failure, got %+v", sink.snapshot())
	}
	// No half-written image row.
	if _, _, err := svc.db().MessageImage(term.MessageID); err == nil && term.MessageID != "" {
		t.Fatal("unexpected image row after failure")
	}
}

func TestCancelAllStopsImageGenerations(t *testing.T) {
	block := make(chan struct{})
	svc, _ := serviceWith(t, func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-block:
		case <-r.Context().Done():
		}
	})
	var sink imageSink
	svc.SetImageEmitter(sink.emit)
	for _, c := range []string{"c1", "c2"} {
		if _, err := svc.GenerateImage(context.Background(), c, "x", "img"); err != nil {
			t.Fatalf("generate %s: %v", c, err)
		}
	}
	svc.CancelAll() // must not return until image goroutines finish
	svc.mu.Lock()
	n := len(svc.imageActive)
	svc.mu.Unlock()
	if n != 0 {
		t.Fatalf("imageActive leaked after CancelAll: %d", n)
	}
	if sink.has("cancelled") == false {
		t.Fatalf("expected cancelled events, got %+v", sink.snapshot())
	}
	close(block)
}

func TestDeleteConversationCascadesImages(t *testing.T) {
	svc, _ := serviceWith(t, b64ImageServer(t))
	var sink imageSink
	svc.SetImageEmitter(sink.emit)
	if _, err := svc.GenerateImage(context.Background(), "c1", "panda", "img"); err != nil {
		t.Fatalf("generate: %v", err)
	}
	done := waitImageDone(t, svc, &sink)
	if _, _, err := svc.db().MessageImage(done.MessageID); err != nil {
		t.Fatalf("image should exist: %v", err)
	}
	if err := svc.DeleteConversation("c1"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, _, err := svc.db().MessageImage(done.MessageID); err == nil {
		t.Fatal("image row survived conversation delete")
	}
}

func TestGenerateImageTooLargePayload(t *testing.T) {
	svc, _ := serviceWith(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusRequestEntityTooLarge)
		_, _ = w.Write([]byte(`{"error":"too large"}`))
	})
	var sink imageSink
	svc.SetImageEmitter(sink.emit)
	if _, err := svc.GenerateImage(context.Background(), "c1", "panda", "img"); err != nil {
		t.Fatalf("generate: %v", err)
	}
	done := waitImageDone(t, svc, &sink)
	if done.Type != "failed" || done.Error == nil || done.Error.Code != "payload_too_large" {
		t.Fatalf("got %+v", done)
	}
	_ = errors.Is
}
