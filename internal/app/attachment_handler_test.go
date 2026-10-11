package app

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
)

func TestAttachmentHandlerServesStagedBytes(t *testing.T) {
	svc, _ := serviceWith(t, func(w http.ResponseWriter, _ *http.Request) { sseOK(w) })
	m, _ := svc.StageData("a.png", "image/png", pngB64())
	rec := httptest.NewRecorder()
	svc.AttachmentHandler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/attachments/"+m.ID, nil))
	if rec.Code != 200 || rec.Header().Get("Content-Type") != "image/png" || rec.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Fatalf("code %d headers %v", rec.Code, rec.Header())
	}
	if rec.Body.Len() != len(pngBytes) {
		t.Fatalf("body %d bytes", rec.Body.Len())
	}
}

func TestAttachmentHandlerRejectsBadRequests(t *testing.T) {
	svc, _ := serviceWith(t, func(w http.ResponseWriter, _ *http.Request) { sseOK(w) })
	for _, p := range []string{"/attachments/../x", "/attachments/abc", "/attachments/" + uuid.NewString(), "/other", "/attachments/"} {
		rec := httptest.NewRecorder()
		svc.AttachmentHandler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, p, nil))
		if rec.Code != 404 {
			t.Fatalf("%s: code %d", p, rec.Code)
		}
	}
	rec := httptest.NewRecorder()
	svc.AttachmentHandler().ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/attachments/"+uuid.NewString(), nil))
	if rec.Code != 405 {
		t.Fatalf("POST: code %d", rec.Code)
	}
}

func TestAttachmentHandlerAfterConversationDelete(t *testing.T) {
	svc, sink := serviceWith(t, func(w http.ResponseWriter, _ *http.Request) { sseOK(w) })
	m, _ := svc.StageData("a.png", "image/png", pngB64())
	_, _ = svc.SendMessage(context.Background(), "c1", "x", []string{m.ID})
	waitDone(t, svc, sink, 1)
	rec := httptest.NewRecorder()
	svc.AttachmentHandler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/attachments/"+m.ID, nil))
	if rec.Code != 200 {
		t.Fatalf("stored: code %d", rec.Code)
	}
	if err := svc.DeleteConversation("c1"); err != nil {
		t.Fatal(err)
	}
	rec = httptest.NewRecorder()
	svc.AttachmentHandler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/attachments/"+m.ID, nil))
	if rec.Code != 404 {
		t.Fatalf("after delete: code %d", rec.Code)
	}
}
