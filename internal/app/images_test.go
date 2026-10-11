package app

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
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
