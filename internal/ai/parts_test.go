package ai

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMessageMarshalsPlainStringWithoutParts(t *testing.T) {
	b, err := json.Marshal(Message{Role: "user", Content: "halo"})
	if err != nil {
		t.Fatal(err)
	}
	if got := string(b); got != `{"role":"user","content":"halo"}` {
		t.Fatalf("got %s", got)
	}
}

type partsWire struct {
	Role    string           `json:"role"`
	Content []map[string]any `json:"content"`
}

func unmarshalParts(t *testing.T, m Message) partsWire {
	t.Helper()
	b, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	var w partsWire
	if err := json.Unmarshal(b, &w); err != nil {
		t.Fatalf("content is not an array: %s", b)
	}
	return w
}

func TestMessageMarshalsContentArrayTextFirst(t *testing.T) {
	w := unmarshalParts(t, Message{Role: "user", Content: "lihat ini", Parts: []Part{ImagePart("image/png", []byte{1, 2})}})
	if len(w.Content) != 2 {
		t.Fatalf("want 2 parts, got %+v", w.Content)
	}
	if w.Content[0]["type"] != "text" || w.Content[0]["text"] != "lihat ini" {
		t.Fatalf("text part first: %+v", w.Content[0])
	}
	img, _ := w.Content[1]["image_url"].(map[string]any)
	if w.Content[1]["type"] != "image_url" || img["url"] != "data:image/png;base64,AQI=" {
		t.Fatalf("image part: %+v", w.Content[1])
	}
}

func TestMessageWithPartsAndEmptyTextHasNoTextPart(t *testing.T) {
	w := unmarshalParts(t, Message{Role: "user", Parts: []Part{ImagePart("image/jpeg", []byte{9})}})
	if len(w.Content) != 1 || w.Content[0]["type"] != "image_url" {
		t.Fatalf("got %+v", w.Content)
	}
}

func TestFilePartCarriesFilenameAndDataURL(t *testing.T) {
	w := unmarshalParts(t, Message{Role: "user", Parts: []Part{FilePartFor("nota.pdf", "application/pdf", []byte("%PDF"))}})
	f, _ := w.Content[0]["file"].(map[string]any)
	if w.Content[0]["type"] != "file" || f["filename"] != "nota.pdf" || f["file_data"] != "data:application/pdf;base64,JVBERg==" {
		t.Fatalf("got %+v", w.Content[0])
	}
}

func TestInlineTextPartFencesContent(t *testing.T) {
	p := InlineTextPart("catatan.txt", []byte("a,b\n1,2"))
	if p.Type != "text" || !strings.HasPrefix(p.Text, "Isi berkas `catatan.txt`:\n```\n") || !strings.HasSuffix(p.Text, "a,b\n1,2\n```") {
		t.Fatalf("got %q", p.Text)
	}
}

func TestStreamChatSendsPartsAsContentArray(t *testing.T) {
	var body []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("data: [DONE]\n\n"))
	}))
	defer srv.Close()
	c := &OpenAICompatible{BaseURL: srv.URL}
	req := ChatRequest{Model: "m", Messages: []Message{{Role: "user", Content: "apa ini", Parts: []Part{ImagePart("image/png", []byte{1})}}}}
	if err := c.StreamChat(context.Background(), req, func(string) error { return nil }); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), `"type":"image_url"`) || !strings.Contains(string(body), `"text":"apa ini"`) {
		t.Fatalf("body: %s", body)
	}
}
