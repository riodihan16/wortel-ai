package app

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"companion/internal/attach"
	"companion/internal/storage"
)

func imgRow(id string) storage.AttachmentRow {
	return storage.AttachmentRow{ID: id, Kind: attach.KindImage, Mime: "image/png", Name: id + ".png"}
}

func loadOK(string) (string, []byte, error) { return "image/png", []byte{1}, nil }

func TestBuildContextKeepsPartsForNewestTurnsOnly(t *testing.T) {
	var rows []storage.ContextRow
	for i := 1; i <= 5; i++ {
		u := fmt.Sprintf("u%d", i)
		rows = append(rows,
			storage.ContextRow{ID: u, Role: "user", Content: "turn " + u, Attachments: []storage.AttachmentRow{imgRow("a" + u)}},
			storage.ContextRow{ID: "m" + u, Role: "assistant", Content: "ok"},
		)
	}
	msgs := buildContext(rows, AttachmentTurns, loadOK)
	if len(msgs) != 10 {
		t.Fatalf("len %d", len(msgs))
	}
	for i, m := range msgs {
		if m.Role == "assistant" {
			if len(m.Parts) != 0 || m.Content != "ok" {
				t.Fatalf("assistant %d changed: %+v", i, m)
			}
			continue
		}
		turn := i/2 + 1
		if turn <= 2 {
			if len(m.Parts) != 0 || !strings.HasSuffix(m.Content, fmt.Sprintf("\n[gambar: au%d.png]", turn)) {
				t.Fatalf("old turn %d: %+v", turn, m)
			}
		} else if len(m.Parts) != 1 || m.Parts[0].Type != "image_url" || strings.Contains(m.Content, "[gambar") {
			t.Fatalf("new turn %d: %+v", turn, m)
		}
	}
}

func TestBuildContextLoadErrorFallsBackToPlaceholderPart(t *testing.T) {
	rows := []storage.ContextRow{{ID: "u1", Role: "user", Content: "x", Attachments: []storage.AttachmentRow{imgRow("a")}}}
	msgs := buildContext(rows, AttachmentTurns, func(string) (string, []byte, error) { return "", nil, errors.New("gone") })
	if len(msgs[0].Parts) != 1 || msgs[0].Parts[0].Type != "text" || msgs[0].Parts[0].Text != "[gambar: a.png]" {
		t.Fatalf("%+v", msgs[0].Parts)
	}
}

func TestBuildContextPDFAndTextParts(t *testing.T) {
	rows := []storage.ContextRow{{ID: "u1", Role: "user", Content: "baca", Attachments: []storage.AttachmentRow{
		{ID: "p", Kind: attach.KindPDF, Mime: "application/pdf", Name: "nota.pdf"},
		{ID: "t", Kind: attach.KindText, Mime: "text/plain", Name: "catatan.txt"},
	}}}
	load := func(id string) (string, []byte, error) {
		if id == "p" {
			return "application/pdf", []byte("%PDF"), nil
		}
		return "text/plain", []byte("isi"), nil
	}
	msgs := buildContext(rows, AttachmentTurns, load)
	p := msgs[0].Parts
	if len(p) != 2 || p[0].Type != "file" || p[0].File == nil || p[0].File.Filename != "nota.pdf" {
		t.Fatalf("pdf: %+v", p)
	}
	if p[1].Type != "text" || !strings.Contains(p[1].Text, "Isi berkas `catatan.txt`") || !strings.Contains(p[1].Text, "isi") {
		t.Fatalf("text: %+v", p[1])
	}
}

func TestBuildContextPlaceholderWithEmptyText(t *testing.T) {
	rows := []storage.ContextRow{
		{ID: "u1", Role: "user", Content: "", Attachments: []storage.AttachmentRow{{ID: "p", Kind: attach.KindPDF, Mime: "application/pdf", Name: "nota.pdf"}}},
		{ID: "u2", Role: "user", Content: "a"}, {ID: "u3", Role: "user", Content: "b"}, {ID: "u4", Role: "user", Content: "c"},
	}
	msgs := buildContext(rows, AttachmentTurns, loadOK)
	if msgs[0].Content != "[berkas: nota.pdf]" || len(msgs[0].Parts) != 0 {
		t.Fatalf("%+v", msgs[0])
	}
}

func TestBuildContextLoadsBytesOnlyInsideWindow(t *testing.T) {
	var rows []storage.ContextRow
	for i := 1; i <= 5; i++ {
		u := fmt.Sprintf("u%d", i)
		rows = append(rows,
			storage.ContextRow{ID: u, Role: "user", Content: "turn " + u, Attachments: []storage.AttachmentRow{imgRow("a" + u)}},
			storage.ContextRow{ID: "m" + u, Role: "assistant", Content: "ok"},
		)
	}
	var loaded []string
	load := func(id string) (string, []byte, error) {
		loaded = append(loaded, id)
		return "image/png", []byte{1}, nil
	}
	buildContext(rows, AttachmentTurns, load)
	want := []string{"au3", "au4", "au5"}
	if len(loaded) != len(want) {
		t.Fatalf("loaded %v, want %v", loaded, want)
	}
	for i := range want {
		if loaded[i] != want[i] {
			t.Fatalf("loaded %v, want %v", loaded, want)
		}
	}
}
