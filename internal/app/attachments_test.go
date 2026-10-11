package app

import (
	"context"
	"encoding/base64"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"companion/internal/ai"
	"companion/internal/attach"
)

var pngBytes = []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n', 0, 0, 0, 13, 'I', 'H', 'D', 'R'}

func pngB64() string { return base64.StdEncoding.EncodeToString(pngBytes) }

// serviceWith wires a Service to a temp DB and a provider pointing at handler.
func serviceWith(t *testing.T, handler http.HandlerFunc) (*Service, *eventSink) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	p := testProviders(t)
	if _, err := p.Save(context.Background(), SaveInput{Name: "t", BaseURL: srv.URL, Model: "m", APIKey: "k"}); err != nil {
		t.Fatalf("save: %v", err)
	}
	var sink eventSink
	svc := NewService(&ai.MockProvider{}, sink.emit)
	svc.SetProviders(p)
	return svc, &sink
}

func sseOK(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/event-stream")
	_, _ = w.Write([]byte("data: {\"choices\":[{\"delta\":{\"content\":\"ok\"}}]}\n\ndata: [DONE]\n\n"))
}

// waitDone blocks until n terminal events (completed/cancelled/failed) exist
// AND no generation is still registered: the event is emitted before the
// goroutine's deferred cleanup, so a retry or delete issued right after the
// event could otherwise see "generation already active".
func waitDone(t *testing.T, svc *Service, sink *eventSink, n int) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		c := 0
		for _, e := range sink.snapshot() {
			if e.Type == "completed" || e.Type == "cancelled" || e.Type == "failed" {
				c++
			}
		}
		svc.mu.Lock()
		active := len(svc.active)
		svc.mu.Unlock()
		if c >= n && active == 0 {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("no terminal event or generation still active: %+v", sink.snapshot())
}

func TestSendWithAttachmentsSendsPartsAndStoresRows(t *testing.T) {
	var mu sync.Mutex
	var body string
	svc, sink := serviceWith(t, func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		mu.Lock()
		body = string(b)
		mu.Unlock()
		sseOK(w)
	})
	m, err := svc.StageData("foto.png", "image/png", pngB64())
	if err != nil {
		t.Fatalf("stage: %v", err)
	}
	if _, err := svc.SendMessage(context.Background(), "c1", "apa ini?", []string{m.ID}); err != nil {
		t.Fatalf("send: %v", err)
	}
	waitDone(t, svc, sink, 1)
	mu.Lock()
	got := body
	mu.Unlock()
	if !strings.Contains(got, `"type":"image_url"`) || !strings.Contains(got, `"text":"apa ini?"`) || strings.Contains(got, `"content":""`) {
		t.Fatalf("body: %s", got)
	}
	if svc.staging.Len() != 0 {
		t.Fatal("staging not cleared after send")
	}
	hist := svc.History("c1", 10)
	if len(hist) != 2 || len(hist[0].Attachments) != 1 || hist[0].Attachments[0].Name != "foto.png" || hist[0].Attachments[0].Kind != attach.KindImage {
		t.Fatalf("history: %+v", hist)
	}
	if hist[1].Attachments == nil || len(hist[1].Attachments) != 0 {
		t.Fatalf("assistant must carry an empty (non-nil) list: %+v", hist[1])
	}
}

func TestSendAttachmentOnlyWithoutText(t *testing.T) {
	svc, sink := serviceWith(t, func(w http.ResponseWriter, _ *http.Request) { sseOK(w) })
	m, _ := svc.StageData("foto.png", "image/png", pngB64())
	if _, err := svc.SendMessage(context.Background(), "c1", "   ", []string{m.ID}); err != nil {
		t.Fatalf("send: %v", err)
	}
	waitDone(t, svc, sink, 1)
	hist := svc.History("c1", 10)
	if len(hist) != 2 || hist[0].Content != "" || len(hist[0].Attachments) != 1 {
		t.Fatalf("history: %+v", hist)
	}
	convs := svc.ListConversations()
	if len(convs) != 1 || convs[0].Title != "foto.png" {
		t.Fatalf("title should fall back to the file name: %+v", convs)
	}
}

func TestSendRejectsUnknownAttachmentIDWithoutLocking(t *testing.T) {
	svc, sink := serviceWith(t, func(w http.ResponseWriter, _ *http.Request) { sseOK(w) })
	if _, err := svc.SendMessage(context.Background(), "c1", "x", []string{"nope"}); !errors.Is(err, attach.ErrUnknown) {
		t.Fatalf("got %v", err)
	}
	if _, err := svc.SendMessage(context.Background(), "c1", "x", nil); err != nil {
		t.Fatalf("conversation stayed locked: %v", err)
	}
	waitDone(t, svc, sink, 1)
}

// The composer clears its pending list as soon as it calls SendMessage, so
// ids a synchronous rejection leaves in staging would be invisible yet
// still count toward the per-message cap.
func TestSendRejectionDropsStagedIDs(t *testing.T) {
	release := make(chan struct{})
	svc, sink := serviceWith(t, func(_ http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-release:
		}
	})
	t.Cleanup(func() { close(release) }) // before srv.Close (cleanups run LIFO)
	if _, err := svc.SendMessage(context.Background(), "c1", "lama", nil); err != nil {
		t.Fatal(err)
	}
	m, err := svc.StageData("foto.png", "image/png", pngB64())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.SendMessage(context.Background(), "c1", "x", []string{m.ID}); err == nil || err.Error() != "generation already active" {
		t.Fatalf("got %v", err)
	}
	if svc.staging.Len() != 0 {
		t.Fatalf("busy rejection left %d staged", svc.staging.Len())
	}
	m2, err := svc.StageData("foto.png", "image/png", pngB64())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.SendMessage(context.Background(), "c2", "x", []string{m2.ID, m2.ID}); !errors.Is(err, attach.ErrDuplicate) {
		t.Fatalf("got %v", err)
	}
	if svc.staging.Len() != 0 {
		t.Fatalf("select rejection left %d staged", svc.staging.Len())
	}
	svc.CancelGeneration("c1")
	waitDone(t, svc, sink, 1)
}

func TestSendAttachmentsRequireStorage(t *testing.T) {
	svc := NewService(&ai.MockProvider{}, func(string, ai.ChatEvent) {})
	if _, err := svc.SendMessage(context.Background(), "c1", "x", []string{"a"}); !errors.Is(err, errNoAttachmentStorage) {
		t.Fatalf("got %v", err)
	}
	if _, err := svc.StageData("a.png", "image/png", pngB64()); !errors.Is(err, errNoAttachmentStorage) {
		t.Fatalf("stage: %v", err)
	}
}

func TestRetryResendsStoredAttachments(t *testing.T) {
	var mu sync.Mutex
	var bodies []string
	fail := true
	svc, sink := serviceWith(t, func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		mu.Lock()
		bodies = append(bodies, string(b))
		f := fail
		fail = false
		mu.Unlock()
		if f {
			w.WriteHeader(500)
			return
		}
		sseOK(w)
	})
	m, _ := svc.StageData("foto.png", "image/png", pngB64())
	if _, err := svc.SendMessage(context.Background(), "c1", "apa ini?", []string{m.ID}); err != nil {
		t.Fatal(err)
	}
	waitDone(t, svc, sink, 1)
	if _, err := svc.RetryMessage(context.Background(), "c1", "apa ini?"); err != nil {
		t.Fatal(err)
	}
	waitDone(t, svc, sink, 2)
	mu.Lock()
	defer mu.Unlock()
	if len(bodies) != 2 || !strings.Contains(bodies[1], `"type":"image_url"`) {
		t.Fatalf("retry body lost the image: %v", bodies)
	}
	users := 0
	for _, h := range svc.History("c1", 10) {
		if h.Role == "user" {
			users++
		}
	}
	if users != 1 {
		t.Fatalf("user turn duplicated: %d", users)
	}
}

func TestStagePathsReadsFilesAndRejectsWholeBatch(t *testing.T) {
	svc, _ := serviceWith(t, func(w http.ResponseWriter, _ *http.Request) { sseOK(w) })
	dir := t.TempDir()
	png := filepath.Join(dir, "foto.png")
	exe := filepath.Join(dir, "app.exe")
	_ = os.WriteFile(png, pngBytes, 0o644)
	_ = os.WriteFile(exe, []byte("MZ\x90\x00\x03\x00\x00\x00"), 0o644)
	if _, err := svc.StagePaths([]string{png, exe}); err == nil || strings.Contains(err.Error(), dir) {
		t.Fatalf("want error naming only the file, got %v", err)
	}
	if svc.staging.Len() != 0 {
		t.Fatal("failed batch left items staged")
	}
	metas, err := svc.StagePaths([]string{png})
	if err != nil || len(metas) != 1 || metas[0].Name != "foto.png" || metas[0].ID == "" {
		t.Fatalf("%+v %v", metas, err)
	}
}

func TestStagePathsRejectsDirectory(t *testing.T) {
	svc, _ := serviceWith(t, func(w http.ResponseWriter, _ *http.Request) { sseOK(w) })
	if _, err := svc.StagePaths([]string{t.TempDir()}); err == nil || !strings.Contains(err.Error(), "folder") {
		t.Fatalf("got %v", err)
	}
}

func TestStageDataNamesPastedImageByMime(t *testing.T) {
	svc, _ := serviceWith(t, func(w http.ResponseWriter, _ *http.Request) { sseOK(w) })
	m, err := svc.StageData("", "image/png", pngB64())
	if err != nil || m.Name != "tempel.png" || m.Kind != attach.KindImage {
		t.Fatalf("%+v %v", m, err)
	}
	if _, err := svc.StageData("a.png", "image/png", "!!not base64!!"); err == nil {
		t.Fatal("bad base64 accepted")
	}
}

func TestStageDataRejectsOversizedBeforeDecoding(t *testing.T) {
	svc, _ := serviceWith(t, func(w http.ResponseWriter, _ *http.Request) { sseOK(w) })
	// DecodedLen only needs the length: this decodes to 16 MiB - 1 of zeros.
	b64 := strings.Repeat("A", (16<<20)/3*4)
	if _, err := svc.StageData("", "image/png", b64); err == nil || err.Error() != "tempel.png: melebihi 15 MB" {
		t.Fatalf("got %v", err)
	}
	if svc.staging.Len() != 0 {
		t.Fatal("oversized paste staged")
	}
}

func TestUnstageAndLookup(t *testing.T) {
	svc, sink := serviceWith(t, func(w http.ResponseWriter, _ *http.Request) { sseOK(w) })
	m, _ := svc.StageData("a.png", "image/png", pngB64())
	if mime, data, ok := svc.Attachment(m.ID); !ok || mime != "image/png" || len(data) != len(pngBytes) {
		t.Fatal("staged lookup failed")
	}
	svc.Unstage(m.ID)
	svc.Unstage(m.ID)
	if _, _, ok := svc.Attachment(m.ID); ok {
		t.Fatal("unstaged item still served")
	}
	m2, _ := svc.StageData("b.png", "image/png", pngB64())
	_, _ = svc.SendMessage(context.Background(), "c1", "x", []string{m2.ID})
	waitDone(t, svc, sink, 1)
	if _, _, ok := svc.Attachment(m2.ID); !ok {
		t.Fatal("stored attachment not found via DB")
	}
}

func TestResolveFailureKeepsAttachmentsForRetry(t *testing.T) {
	var sink eventSink
	svc := NewService(&ai.MockProvider{}, sink.emit)
	svc.SetProviders(testProviders(t))
	m, err := svc.StageData("foto.png", "image/png", pngB64())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.SendMessage(context.Background(), "c1", "x", []string{m.ID}); err != nil {
		t.Fatal(err)
	}
	waitDone(t, svc, &sink, 1)
	evs := sink.snapshot()
	last := evs[len(evs)-1]
	if last.Type != "failed" || last.Error == nil || last.Error.Code != "no_provider" {
		t.Fatalf("last event: %+v", last)
	}
	hist := svc.History("c1", 10)
	if len(hist) != 2 || hist[0].Role != "user" || len(hist[0].Attachments) != 1 || hist[0].Attachments[0].Name != "foto.png" || hist[1].Role != "assistant" || hist[1].Status != "failed" {
		t.Fatalf("history: %+v", hist)
	}
	if _, _, ok := svc.Attachment(m.ID); !ok {
		t.Fatal("attachment lost after resolve failure")
	}
}

func TestStagePathsRollsBackOnCapError(t *testing.T) {
	svc, _ := serviceWith(t, func(w http.ResponseWriter, _ *http.Request) { sseOK(w) })
	for i := 0; i < 3; i++ {
		if _, err := svc.StageData("s.png", "image/png", pngB64()); err != nil {
			t.Fatal(err)
		}
	}
	dir := t.TempDir()
	var paths []string
	for _, n := range []string{"a.png", "b.png"} {
		p := filepath.Join(dir, n)
		_ = os.WriteFile(p, pngBytes, 0o644)
		paths = append(paths, p)
	}
	if _, err := svc.StagePaths(paths); !errors.Is(err, attach.ErrTooMany) {
		t.Fatalf("got %v", err)
	}
	if svc.staging.Len() != 3 {
		t.Fatalf("staging len %d, want 3", svc.staging.Len())
	}
}

func TestStagePathsRejectsTooManyBeforeReading(t *testing.T) {
	svc, _ := serviceWith(t, func(w http.ResponseWriter, _ *http.Request) { sseOK(w) })
	for i := 0; i < 3; i++ {
		if _, err := svc.StageData("s.png", "image/png", pngB64()); err != nil {
			t.Fatal(err)
		}
	}
	// Neither file exists: only a check made before any Stat or read can
	// answer ErrTooMany here.
	dir := t.TempDir()
	paths := []string{filepath.Join(dir, "a.png"), filepath.Join(dir, "b.png")}
	if _, err := svc.StagePaths(paths); !errors.Is(err, attach.ErrTooMany) {
		t.Fatalf("got %v", err)
	}
	if svc.staging.Len() != 3 {
		t.Fatalf("staging len %d, want 3", svc.staging.Len())
	}
}

func TestStagePathsRejectsOverBudgetBeforeReading(t *testing.T) {
	svc, _ := serviceWith(t, func(w http.ResponseWriter, _ *http.Request) { sseOK(w) })
	dir := t.TempDir()
	var paths []string
	// Two 11 MiB files: each passes MaxFileBytes, together they exceed the
	// 20 MiB message budget. Sparse via Truncate, so nothing is written.
	for _, n := range []string{"a.pdf", "b.pdf"} {
		p := filepath.Join(dir, n)
		f, err := os.Create(p)
		if err != nil {
			t.Fatal(err)
		}
		if err := f.Truncate(11 << 20); err != nil {
			t.Fatal(err)
		}
		_ = f.Close()
		paths = append(paths, p)
	}
	if _, err := svc.StagePaths(paths); !errors.Is(err, attach.ErrTooBig) {
		t.Fatalf("got %v", err)
	}
	if svc.staging.Len() != 0 {
		t.Fatalf("staging len %d, want 0", svc.staging.Len())
	}
}

func TestStagePathsSkipsEmptyPaths(t *testing.T) {
	svc, _ := serviceWith(t, func(w http.ResponseWriter, _ *http.Request) { sseOK(w) })
	png := filepath.Join(t.TempDir(), "foto.png")
	_ = os.WriteFile(png, pngBytes, 0o644)
	metas, err := svc.StagePaths([]string{"", png})
	if err != nil || len(metas) != 1 || metas[0].Name != "foto.png" {
		t.Fatalf("%+v %v", metas, err)
	}
}

func TestRetryEmptyTextResendsAttachmentOnlyTurn(t *testing.T) {
	var mu sync.Mutex
	var bodies []string
	fail := true
	svc, sink := serviceWith(t, func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		mu.Lock()
		bodies = append(bodies, string(b))
		f := fail
		fail = false
		mu.Unlock()
		if f {
			w.WriteHeader(500)
			return
		}
		sseOK(w)
	})
	m, _ := svc.StageData("foto.png", "image/png", pngB64())
	if _, err := svc.SendMessage(context.Background(), "c1", "", []string{m.ID}); err != nil {
		t.Fatal(err)
	}
	waitDone(t, svc, sink, 1)
	if _, err := svc.RetryMessage(context.Background(), "c1", ""); err != nil {
		t.Fatal(err)
	}
	waitDone(t, svc, sink, 2)
	mu.Lock()
	defer mu.Unlock()
	if len(bodies) != 2 || !strings.Contains(bodies[1], `"type":"image_url"`) {
		t.Fatalf("retry body lost the image: %v", bodies)
	}
	users := 0
	for _, h := range svc.History("c1", 10) {
		if h.Role == "user" {
			users++
		}
	}
	if users != 1 {
		t.Fatalf("user rows: %d", users)
	}
}

func TestRetryEmptyTextWithNothingToReuseFails(t *testing.T) {
	svc, sink := serviceWith(t, func(w http.ResponseWriter, _ *http.Request) { sseOK(w) })
	if _, err := svc.SendMessage(context.Background(), "c1", "halo", nil); err != nil {
		t.Fatal(err)
	}
	waitDone(t, svc, sink, 1)
	if _, err := svc.RetryMessage(context.Background(), "c1", ""); err != nil {
		t.Fatalf("retry should be accepted synchronously: %v", err)
	}
	waitDone(t, svc, sink, 2)
	evs := sink.snapshot()
	last := evs[len(evs)-1]
	if last.Type != "failed" || last.Error == nil || last.Error.Code != "empty" {
		t.Fatalf("last event: %+v", last)
	}
	users := 0
	for _, h := range svc.History("c1", 10) {
		if h.Role == "user" {
			users++
		}
	}
	if users != 1 {
		t.Fatalf("user rows: %d", users)
	}
}
