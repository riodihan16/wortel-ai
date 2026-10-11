# Lampiran Gambar + Berkas — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Pengguna dapat melampirkan gambar (png/jpeg/webp), PDF, dan berkas teks ke pesan chat; lampiran dikirim ke model multimodal provider, tersimpan di riwayat, dan tampil kembali.

**Architecture:** Bytes lampiran di-stage di memori Go (`internal/attach`), dipindahkan ke SQLite (BLOB, tabel `attachments`) dalam satu transaksi bersama pesan user saat kirim, dan disajikan ke WebView lewat route asset-server `/attachments/{id}`. Request ke provider memakai `content` array standar OpenAI (`text` / `image_url` / `file`); hanya 3 turn user terakhir yang membawa lampiran penuh, sisanya placeholder teks.

**Tech Stack:** Go 1.25 + Wails v2.16.0 (`runtime.OpenMultipleFilesDialog`, `OnFileDrop`, `assetserver.Options.Handler`), `modernc.org/sqlite`, React 19 + TypeScript 5.6 + Vitest 3 (jsdom).

**Spec:** `docs/superpowers/specs/2026-10-11-attachments-design.md`

## Global Constraints

- Toolchain terkunci: Go 1.25, Wails v2.16.0, React 19.1.0, TypeScript 5.6.3, Vite 7.3.7. Tidak ada dependency baru (`github.com/google/uuid` sudah ada).
- Tipe berkas: gambar `image/png` `image/jpeg` `image/webp`; PDF `application/pdf`; teks `.txt .md .csv .json`. Tipe dari magic bytes (`http.DetectContentType`), bukan ekstensi.
- Batas: ≤ 4 lampiran/pesan; gambar ≤ 8 MB; PDF ≤ 15 MB; teks ≤ 256 KB + UTF-8 valid tanpa NUL; total ≤ 20 MB; nama ≤ 120 rune, basename saja.
- `AttachmentTurns = 3`: hanya 3 pesan user terakhir membawa part lampiran; lebih lama → `[gambar: nama]` / `[berkas: nama]`.
- Tanpa gating model. `413` → `payload_too_large`; `400` yang menyebut image/file/vision/modal/document → `unsupported_input`.
- Tidak ada path penuh, bytes, atau base64 di log, event, maupun pesan error. Berkas dibaca hanya dari aksi eksplisit (dialog, drop, paste).
- Route asset: hanya `GET /attachments/{uuid-v4-lowercase}`; selain itu 404/405; header `nosniff`.
- Frontend: tidak ada `any`, listener dibersihkan saat unmount, semua warna/ukuran lewat `var(--token)`, teks UI bahasa Indonesia.
- Tes: tulis dulu (RED) → implementasi (GREEN). Perintah verifikasi: `go test ./...`, `gofmt -l app.go main.go internal/`, `go vet ./...`, `cd frontend && npx tsc --noEmit && npm test && npm run build`.
- Jangan jalankan `wails build`/`wails dev` sebelum Task 16; jangan menyentuh DB nyata di `%LOCALAPPDATA%\WortelAI` (tes memakai `t.TempDir()`).

## Review Focus

1. **Ekstensi gambar yang bukan gambar** (`foto.heic`, `.jpg` berisi HEIC): pengguna mengharapkan pesan "tipe tidak didukung" yang menyebut nama berkas, bukan request 400 dari provider. → tes `TestClassifyRejectsUnknownType` (Task 3) memuat kasus `foto.heic`.
2. **Folder di-drop ke panel**: harus ditolak dengan pesan jelas, tidak panic saat `os.ReadFile`. → tes `TestStagePathsRejectsDirectory` (Task 7).
3. **Retry setelah provider gagal pada turn berlampiran**: lampiran harus ikut dikirim ulang dari DB, pesan user tidak dobel. → tes `TestRetryResendsStoredAttachments` (Task 7).
4. **Thumbnail percakapan yang sudah dihapus**: route harus 404, bukan 500 atau bytes lama. → tes `TestAttachmentHandlerAfterConversationDelete` (Task 8).
5. **Paste gambar saat pending sudah 4**: ditolak lokal dengan pesan, backend tidak dipanggil, daftar tidak berubah. → tes `refuses a fifth attachment locally` (Task 11).

---

## File Structure

| File | Tanggung jawab |
|---|---|
| `internal/ai/parts.go` (baru) | Tipe `Part`, `ImageURL`, `FilePart`; konstruktor part; `DataURL` |
| `internal/ai/provider.go` | `Message` + `MarshalJSON` (string vs array); baca body ≤ 4 KB saat non-2xx |
| `internal/ai/errors.go` | `MapHTTPErrorWithBody`: 413, 400 modality |
| `internal/attach/validate.go` (baru) | Konstanta batas, `Meta`, `Classify`, `CleanName` |
| `internal/attach/staging.go` (baru) | `Staging` memori: `Add/Get/Select/Remove/Clear/Len` |
| `internal/storage/db.go` | migrasi v5 (`attachments`) |
| `internal/storage/attachments.go` (baru) | `AttachmentRow`, `AppendUserTurn`, `AttachmentsFor`, `AttachmentData` |
| `internal/storage/chat.go` | `insertMessage` dipakai bersama; `RecentContext` menggantikan `RecentMessages` |
| `internal/app/context.go` (baru) | `AttachmentTurns`, `buildContext`, `partFor`, placeholder |
| `internal/app/attachments.go` (baru) | `Service.StagePaths/StageData/Unstage/Attachment` |
| `internal/app/attachment_handler.go` (baru) | `Service.AttachmentHandler()` route `/attachments/{id}` |
| `internal/app/service.go` | `SendMessage(…, attachmentIDs)`, `generate` memakai staging + `AppendUserTurn` + `buildContext` |
| `internal/app/history.go` | `ChatMessage.Attachments` |
| `app.go` | binding `PickAttachments`, `StageAttachmentPaths`, `StageAttachmentData`, `UnstageAttachment`, `SendMessage` baru, `assetHandler()` |
| `main.go` | `AssetServer.Handler`, `DragAndDrop.EnableFileDrop` |
| `frontend/wailsjs/**` | regenerasi binding |
| `frontend/src/types/chat.ts` | `AttachmentMeta`, `ChatMessage.attachments` |
| `frontend/src/test/wails.ts` | stub `OnFileDrop/OnFileDropOff` |
| `frontend/src/hooks/useChat.ts` | `send(text, attachments)`; history membawa `attachments` |
| `frontend/src/hooks/useAttachments.ts` (baru) | state pending + binding stage/unstage |
| `frontend/src/lib/formatSize.ts` (baru) | "512 B" / "2.0 KB" / "1.5 MB" |
| `frontend/src/components/ui/Icon.tsx` | `IconPaperclip`, `IconFile` |
| `frontend/src/components/chat/ChatPanel.tsx` | tombol klip, baris pending, paste, drop zone, kirim tanpa teks |
| `frontend/src/components/chat/AttachmentList.tsx` (baru) | thumbnail/chip di transkrip |
| `frontend/src/components/chat/MessageItem.tsx` | render `AttachmentList` pada bubble user |
| `frontend/src/styles/chat.css` | drop overlay, pending row, thumbnail, chip |
| `frontend/src/App.tsx`, `frontend/src/main.tsx` | wiring `useAttachments`; guard drop global |
| `docs/02,05,07,09,10`, `README.md` | dokumentasi + ADR-054 |

---

### Task 1: `internal/ai` — Part types dan `Message.MarshalJSON`

**Files:**
- Create: `internal/ai/parts.go`
- Modify: `internal/ai/provider.go:15-19` (struct `Message`)
- Test: `internal/ai/parts_test.go`

**Interfaces:**
- Produces: `ai.Part{Type, Text, ImageURL *ImageURL, File *FilePart}`, `ai.TextPart(string) Part`, `ai.ImagePart(mime string, data []byte) Part`, `ai.FilePartFor(name, mime string, data []byte) Part`, `ai.InlineTextPart(name string, data []byte) Part`, `ai.DataURL(mime string, data []byte) string`, `ai.Message{Role, Content string; Parts []Part}` yang marshal ke `{"role","content":"…"}` bila `Parts` kosong, atau `content` array (teks dulu) bila terisi.

- [ ] **Step 1: Tulis tes yang gagal**

`internal/ai/parts_test.go`:

```go
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
```

- [ ] **Step 2: Jalankan, pastikan gagal**

Run: `go test ./internal/ai/ -run 'TestMessage|TestFilePart|TestInlineText|TestStreamChatSendsParts' -v`
Expected: compile error `undefined: Part` / `ImagePart`.

- [ ] **Step 3: Implementasi minimal**

`internal/ai/parts.go`:

```go
package ai

import "encoding/base64"

// Part is one element of a multimodal `content` array in an OpenAI-compatible
// Chat Completions message. Only the field matching Type is set.
type Part struct {
	Type     string    `json:"type"` // text | image_url | file
	Text     string    `json:"text,omitempty"`
	ImageURL *ImageURL `json:"image_url,omitempty"`
	File     *FilePart `json:"file,omitempty"`
}

// ImageURL carries a data URL for an image_url part.
type ImageURL struct {
	URL string `json:"url"`
}

// FilePart carries a document (PDF) as a data URL for a file part.
type FilePart struct {
	Filename string `json:"filename"`
	FileData string `json:"file_data"`
}

// DataURL builds data:<mime>;base64,<data>. Providers reject data URLs
// without a media type, so mime is always required.
func DataURL(mime string, data []byte) string {
	return "data:" + mime + ";base64," + base64.StdEncoding.EncodeToString(data)
}

func TextPart(s string) Part { return Part{Type: "text", Text: s} }

func ImagePart(mime string, data []byte) Part {
	return Part{Type: "image_url", ImageURL: &ImageURL{URL: DataURL(mime, data)}}
}

func FilePartFor(name, mime string, data []byte) Part {
	return Part{Type: "file", File: &FilePart{Filename: name, FileData: DataURL(mime, data)}}
}

// InlineTextPart embeds a small text file as a fenced block, so any
// provider can read it without a provider-specific document part.
func InlineTextPart(name string, data []byte) Part {
	return TextPart("Isi berkas `" + name + "`:\n```\n" + string(data) + "\n```")
}
```

Di `internal/ai/provider.go` ganti struct `Message` (baris 15-19) dengan:

```go
// Message is one chat message sent to a provider. Parts, when present, make
// the message multimodal: Content becomes the first text part. Without Parts
// the message marshals as the plain {"role","content":"…"} shape every
// OpenAI-compatible provider understands.
type Message struct {
	Role    string
	Content string
	Parts   []Part
}

// MarshalJSON picks the string or array content shape (see Message).
func (m Message) MarshalJSON() ([]byte, error) {
	if len(m.Parts) == 0 {
		return json.Marshal(struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		}{m.Role, m.Content})
	}
	parts := make([]Part, 0, len(m.Parts)+1)
	if m.Content != "" {
		parts = append(parts, TextPart(m.Content))
	}
	parts = append(parts, m.Parts...)
	return json.Marshal(struct {
		Role    string `json:"role"`
		Content []Part `json:"content"`
	}{m.Role, parts})
}
```

- [ ] **Step 4: Jalankan, pastikan lulus**

Run: `go test ./internal/ai/ -v`
Expected: semua PASS (termasuk tes lama `TestStreamDeltasFromServer`).

- [ ] **Step 5: Commit**

```bash
git add internal/ai/parts.go internal/ai/parts_test.go internal/ai/provider.go
git commit -m "feat(ai): multimodal message parts with string fallback"
```

---

### Task 2: `internal/ai` — error mapping 413 dan 400 modality

**Files:**
- Modify: `internal/ai/errors.go`
- Modify: `internal/ai/provider.go` (blok non-2xx di `StreamChat`)
- Test: `internal/ai/provider_test.go` (tambah 3 fungsi)

**Interfaces:**
- Produces: `ai.MapHTTPErrorWithBody(status int, h http.Header, body []byte) *PublicError`; kode baru `payload_too_large`, `unsupported_input`.

- [ ] **Step 1: Tulis tes yang gagal**

Tambahkan ke akhir `internal/ai/provider_test.go`:

```go
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

func TestError400WithoutModalityWordsStaysBadRequest(t *testing.T) {
	srv := sseServer(t, 400, `{"error":{"message":"temperature must be <= 2; check your profile"}}`, nil)
	defer srv.Close()
	var pub *PublicError
	if err := streamErr(t, &OpenAICompatible{BaseURL: srv.URL}); !errors.As(err, &pub) || pub.Code != "bad_request" {
		t.Fatalf("got %v", err)
	}
}
```

- [ ] **Step 2: Jalankan, pastikan gagal**

Run: `go test ./internal/ai/ -run 'TestErrorMapping413|TestError400' -v`
Expected: 413 → kode `provider_error`; 400 image → `bad_request`. FAIL.

- [ ] **Step 3: Implementasi**

Di `internal/ai/errors.go` tambahkan import `"regexp"` dan:

```go
// attachmentWords marks a 400 that is about the input modality rather than
// a generic parameter problem, so the UI can say "this model does not take
// images" instead of "request rejected". Word-bounded: "profile" must not
// match "file".
var attachmentWords = regexp.MustCompile(`(?i)\b(image|images|image_url|file|files|vision|modal|multimodal|document|documents)\b`)

// MapHTTPErrorWithBody is MapHTTPErrorWithHeaders plus the two attachment
// cases: 413 (body too large) and a 400 whose message names images/files.
// body is at most a few KB and is never logged or forwarded.
func MapHTTPErrorWithBody(status int, h http.Header, body []byte) *PublicError {
	switch {
	case status == http.StatusRequestEntityTooLarge:
		return &PublicError{Code: "payload_too_large", Message: "Lampiran terlalu besar untuk provider. Kurangi ukuran atau jumlah berkas.", StatusCode: status}
	case status == http.StatusBadRequest && attachmentWords.Match(body):
		return &PublicError{Code: "unsupported_input", Message: "Model ini tidak menerima gambar/berkas. Pilih model lain atau kirim teks saja.", StatusCode: status}
	}
	return MapHTTPErrorWithHeaders(status, h)
}
```

Di `internal/ai/provider.go` tambahkan import `"io"` dan ganti blok:

```go
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return MapHTTPErrorWithHeaders(resp.StatusCode, resp.Header)
	}
```

dengan:

```go
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// A few KB is enough to classify the error; never logged.
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return MapHTTPErrorWithBody(resp.StatusCode, resp.Header, body)
	}
```

- [ ] **Step 4: Jalankan, pastikan lulus**

Run: `go test ./internal/ai/ -v`
Expected: PASS semua.

- [ ] **Step 5: Commit**

```bash
git add internal/ai/errors.go internal/ai/provider.go internal/ai/provider_test.go
git commit -m "feat(ai): map 413 and modality 400 to attachment errors"
```

---

### Task 3: `internal/attach` — validasi `Classify` + `CleanName`

**Files:**
- Create: `internal/attach/validate.go`
- Test: `internal/attach/validate_test.go`

**Interfaces:**
- Produces: konstanta `MaxPerMessage=4`, `MaxImageBytes=8<<20`, `MaxPDFBytes=15<<20`, `MaxTextBytes=256<<10`, `MaxMessageBytes=20<<20`, `MaxFileBytes=MaxPDFBytes`, `MaxNameRunes=120`; `KindImage/KindPDF/KindText`; `type Meta{ID, Name, Mime string; Size int64; Kind string}` (JSON `id,name,mime,size,kind`); `Classify(name string, data []byte) (Meta, error)` (ID kosong); `CleanName(string) string`; `ErrUnsupported`.

- [ ] **Step 1: Tulis tes yang gagal**

`internal/attach/validate_test.go`:

```go
package attach

import (
	"errors"
	"strings"
	"testing"
)

var (
	pngHead  = []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n', 0, 0, 0, 13, 'I', 'H', 'D', 'R'}
	jpegHead = []byte{0xFF, 0xD8, 0xFF, 0xE0, 0, 0x10, 'J', 'F', 'I', 'F', 0}
	webpHead = []byte("RIFF\x00\x00\x00\x00WEBPVP8 ")
	pdfHead  = []byte("%PDF-1.4\n%\xE2\xE3\xCF\xD3\n")
)

func TestClassifyImagesByMagicBytes(t *testing.T) {
	for _, tc := range []struct {
		name string
		data []byte
		mime string
	}{
		{"foto.png", pngHead, "image/png"},
		{"foto.jpg", jpegHead, "image/jpeg"},
		{"foto.webp", webpHead, "image/webp"},
		{"salah-ekstensi.txt", pngHead, "image/png"}, // bytes win over the name
	} {
		m, err := Classify(tc.name, tc.data)
		if err != nil || m.Kind != KindImage || m.Mime != tc.mime || m.Size != int64(len(tc.data)) || m.ID != "" {
			t.Fatalf("%s: %+v %v", tc.name, m, err)
		}
	}
}

func TestClassifyPDF(t *testing.T) {
	m, err := Classify("nota.pdf", pdfHead)
	if err != nil || m.Kind != KindPDF || m.Mime != "application/pdf" {
		t.Fatalf("%+v %v", m, err)
	}
}

func TestClassifyTextByExtension(t *testing.T) {
	for ext, mime := range map[string]string{".txt": "text/plain", ".md": "text/markdown", ".csv": "text/csv", ".json": "application/json", ".TXT": "text/plain"} {
		m, err := Classify("berkas"+ext, []byte("halo\n"))
		if err != nil || m.Kind != KindText || m.Mime != mime {
			t.Fatalf("%s: %+v %v", ext, m, err)
		}
	}
}

func TestClassifyRejectsUnknownType(t *testing.T) {
	for _, tc := range []struct {
		name string
		data []byte
	}{
		{"app.exe", []byte("MZ\x90\x00\x03\x00\x00\x00\x04\x00\x00\x00\xff\xff")},
		{"halaman.html", []byte("<html><body>x</body></html>")},
		{"foto.heic", []byte("\x00\x00\x00\x18ftypheic\x00\x00\x00\x00")},
	} {
		_, err := Classify(tc.name, tc.data)
		if !errors.Is(err, ErrUnsupported) || !strings.Contains(err.Error(), tc.name) {
			t.Fatalf("%s: %v", tc.name, err)
		}
	}
}

func TestClassifyRejectsOversize(t *testing.T) {
	img := make([]byte, MaxImageBytes+1)
	copy(img, pngHead)
	if _, err := Classify("besar.png", img); err == nil || !strings.Contains(err.Error(), "besar.png") {
		t.Fatalf("image: %v", err)
	}
	pdf := make([]byte, MaxPDFBytes+1)
	copy(pdf, pdfHead)
	if _, err := Classify("besar.pdf", pdf); err == nil {
		t.Fatal("pdf oversize accepted")
	}
	txt := make([]byte, MaxTextBytes+1)
	for i := range txt {
		txt[i] = 'a'
	}
	if _, err := Classify("besar.txt", txt); err == nil {
		t.Fatal("text oversize accepted")
	}
}

func TestClassifyRejectsNonUTF8OrNULText(t *testing.T) {
	if _, err := Classify("a.txt", []byte{0xff, 0xfe, 'a', 'b'}); err == nil {
		t.Fatal("non-utf8 accepted")
	}
	if _, err := Classify("a.txt", []byte("a\x00b")); err == nil {
		t.Fatal("NUL accepted")
	}
}

func TestClassifyRejectsEmpty(t *testing.T) {
	if _, err := Classify("kosong.png", nil); err == nil {
		t.Fatal("empty accepted")
	}
}

func TestCleanName(t *testing.T) {
	long := strings.Repeat("a", 130) + ".png"
	for in, want := range map[string]string{
		`C:\Users\x\foto.png`: "foto.png",
		"/tmp/a/b.txt":        "b.txt",
		"na\x00me\t.png":      "name.png",
		"":                    "berkas",
		"   ":                 "berkas",
	} {
		if got := CleanName(in); got != want {
			t.Fatalf("%q: got %q want %q", in, got, want)
		}
	}
	if got := []rune(CleanName(long)); len(got) != MaxNameRunes {
		t.Fatalf("long name: %d runes", len(got))
	}
}
```

- [ ] **Step 2: Jalankan, pastikan gagal**

Run: `go test ./internal/attach/ -v`
Expected: compile error (package belum ada).

- [ ] **Step 3: Implementasi**

`internal/attach/validate.go`:

```go
// Package attach validates and stages user-chosen files before they are
// sent to a provider. Nothing here touches the network or the database.
package attach

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"path"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	MaxPerMessage   = 4
	MaxImageBytes   = 8 << 20
	MaxPDFBytes     = 15 << 20
	MaxTextBytes    = 256 << 10
	MaxMessageBytes = 20 << 20
	MaxNameRunes    = 120
	// MaxFileBytes is the largest single file any kind allows; callers refuse
	// to read bigger files before loading them into memory.
	MaxFileBytes = MaxPDFBytes
)

const (
	KindImage = "image"
	KindPDF   = "pdf"
	KindText  = "text"
)

// Meta is the frontend-safe description of one attachment: no bytes, no path.
type Meta struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Mime string `json:"mime"`
	Size int64  `json:"size"`
	Kind string `json:"kind"`
}

// ErrUnsupported names the accepted types so the message is actionable.
var ErrUnsupported = errors.New("tipe berkas tidak didukung (png, jpeg, webp, pdf, txt, md, csv, json)")

var textByExt = map[string]string{
	".txt":  "text/plain",
	".md":   "text/markdown",
	".csv":  "text/csv",
	".json": "application/json",
}

var imageMimes = map[string]bool{"image/png": true, "image/jpeg": true, "image/webp": true}

// Classify decides kind and mime from the bytes (magic numbers via
// http.DetectContentType), never from the extension alone, and applies the
// per-kind size rule. Text files are the one case where the extension
// matters, since bytes alone cannot tell CSV from Markdown. The returned
// Meta has no ID yet.
func Classify(name string, data []byte) (Meta, error) {
	name = CleanName(name)
	if len(data) == 0 {
		return Meta{}, fmt.Errorf("%s: berkas kosong", name)
	}
	sniff := http.DetectContentType(data)
	if i := strings.IndexByte(sniff, ';'); i >= 0 {
		sniff = sniff[:i]
	}
	m := Meta{Name: name, Size: int64(len(data))}
	switch {
	case imageMimes[sniff]:
		if len(data) > MaxImageBytes {
			return Meta{}, fmt.Errorf("%s: gambar melebihi %d MB", name, MaxImageBytes>>20)
		}
		m.Kind, m.Mime = KindImage, sniff
	case sniff == "application/pdf":
		if len(data) > MaxPDFBytes {
			return Meta{}, fmt.Errorf("%s: PDF melebihi %d MB", name, MaxPDFBytes>>20)
		}
		m.Kind, m.Mime = KindPDF, sniff
	default:
		mime, ok := textByExt[strings.ToLower(path.Ext(name))]
		if !ok {
			return Meta{}, fmt.Errorf("%s: %w", name, ErrUnsupported)
		}
		if len(data) > MaxTextBytes {
			return Meta{}, fmt.Errorf("%s: berkas teks melebihi %d KB", name, MaxTextBytes>>10)
		}
		if !utf8.Valid(data) || bytes.IndexByte(data, 0) >= 0 {
			return Meta{}, fmt.Errorf("%s: bukan teks UTF-8", name)
		}
		m.Kind, m.Mime = KindText, mime
	}
	return m, nil
}

// CleanName keeps only the base name (either separator), drops control
// characters, trims, and caps the length. Empty input becomes "berkas".
func CleanName(name string) string {
	name = path.Base(strings.ReplaceAll(strings.TrimSpace(name), `\`, "/"))
	var b strings.Builder
	for _, r := range name {
		if !unicode.IsControl(r) {
			b.WriteRune(r)
		}
	}
	name = strings.TrimSpace(b.String())
	if name == "" || name == "." || name == "/" {
		return "berkas"
	}
	if r := []rune(name); len(r) > MaxNameRunes {
		name = string(r[:MaxNameRunes])
	}
	return name
}
```

- [ ] **Step 4: Jalankan, pastikan lulus**

Run: `go test ./internal/attach/ -v`
Expected: PASS semua.

- [ ] **Step 5: Commit**

```bash
git add internal/attach/validate.go internal/attach/validate_test.go
git commit -m "feat(attach): classify attachments by magic bytes with size rules"
```

---

### Task 4: `internal/attach` — `Staging`

**Files:**
- Create: `internal/attach/staging.go`
- Test: `internal/attach/staging_test.go`

**Interfaces:**
- Consumes: `Meta`, batas dari Task 3.
- Produces: `type Item{Meta Meta; Data []byte}`; `NewStaging() *Staging`; `(*Staging).Add(Meta, []byte) (Meta, error)` (isi `ID`, `Size`; `ErrTooMany`, `ErrTooBig`); `Get(id) (Item, bool)`; `Select(ids []string) ([]Item, error)` (`ErrUnknown`, `ErrDuplicate`; tidak menghapus); `Remove(ids ...string)`; `Clear()`; `Len() int`. Field `newID func() string` untuk tes.

- [ ] **Step 1: Tulis tes yang gagal**

`internal/attach/staging_test.go`:

```go
package attach

import (
	"errors"
	"fmt"
	"testing"
)

func stagingWithIDs(ids ...string) *Staging {
	s := NewStaging()
	i := 0
	s.newID = func() string {
		id := ids[i]
		i++
		return id
	}
	return s
}

func TestStagingAddAndGet(t *testing.T) {
	s := stagingWithIDs("a")
	m, err := s.Add(Meta{Name: "x.png", Mime: "image/png", Kind: KindImage}, []byte{1, 2, 3})
	if err != nil || m.ID != "a" || m.Size != 3 {
		t.Fatalf("%+v %v", m, err)
	}
	it, ok := s.Get("a")
	if !ok || it.Meta.Name != "x.png" || len(it.Data) != 3 || s.Len() != 1 {
		t.Fatalf("get: %+v %v", it, ok)
	}
}

func TestStagingRejectsFifthItem(t *testing.T) {
	s := stagingWithIDs("a", "b", "c", "d", "e")
	for i := 0; i < MaxPerMessage; i++ {
		if _, err := s.Add(Meta{Name: fmt.Sprintf("%d.png", i)}, []byte{1}); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := s.Add(Meta{Name: "5.png"}, []byte{1}); !errors.Is(err, ErrTooMany) {
		t.Fatalf("got %v", err)
	}
}

func TestStagingRejectsOverTotal(t *testing.T) {
	s := stagingWithIDs("a", "b", "c")
	seven := make([]byte, 7<<20)
	for i := 0; i < 2; i++ {
		if _, err := s.Add(Meta{Name: "x.pdf"}, seven); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := s.Add(Meta{Name: "y.pdf"}, seven); !errors.Is(err, ErrTooBig) {
		t.Fatalf("got %v", err)
	}
}

func TestStagingSelectKeepsOrderAndRejectsUnknownOrDuplicate(t *testing.T) {
	s := stagingWithIDs("a", "b")
	_, _ = s.Add(Meta{Name: "a.png"}, []byte{1})
	_, _ = s.Add(Meta{Name: "b.png"}, []byte{2})
	items, err := s.Select([]string{"b", "a"})
	if err != nil || len(items) != 2 || items[0].Meta.Name != "b.png" || items[1].Meta.Name != "a.png" {
		t.Fatalf("%+v %v", items, err)
	}
	if _, err := s.Select([]string{"a", "zz"}); !errors.Is(err, ErrUnknown) {
		t.Fatalf("unknown: %v", err)
	}
	if _, err := s.Select([]string{"a", "a"}); !errors.Is(err, ErrDuplicate) {
		t.Fatalf("duplicate: %v", err)
	}
	if s.Len() != 2 {
		t.Fatal("select must not remove")
	}
}

func TestStagingRemoveIdempotentAndFreesBudget(t *testing.T) {
	s := stagingWithIDs("a", "b")
	big := make([]byte, MaxMessageBytes)
	if _, err := s.Add(Meta{Name: "a.pdf"}, big); err != nil {
		t.Fatal(err)
	}
	s.Remove("a")
	s.Remove("a", "nope")
	if s.Len() != 0 {
		t.Fatal("not removed")
	}
	if _, err := s.Add(Meta{Name: "b.pdf"}, big); err != nil {
		t.Fatalf("budget not freed: %v", err)
	}
}

func TestStagingClear(t *testing.T) {
	s := stagingWithIDs("a")
	_, _ = s.Add(Meta{Name: "a.png"}, []byte{1})
	s.Clear()
	if _, ok := s.Get("a"); ok || s.Len() != 0 {
		t.Fatal("clear failed")
	}
}
```

- [ ] **Step 2: Jalankan, pastikan gagal**

Run: `go test ./internal/attach/ -run TestStaging -v`
Expected: compile error `undefined: NewStaging`.

- [ ] **Step 3: Implementasi**

`internal/attach/staging.go`:

```go
package attach

import (
	"errors"
	"fmt"
	"sync"

	"github.com/google/uuid"
)

var (
	ErrTooMany   = fmt.Errorf("maksimal %d lampiran per pesan", MaxPerMessage)
	ErrTooBig    = fmt.Errorf("total lampiran melebihi %d MB per pesan", MaxMessageBytes>>20)
	ErrUnknown   = errors.New("lampiran tidak ditemukan: lampirkan ulang")
	ErrDuplicate = errors.New("lampiran yang sama disebut dua kali")
)

// Item is a staged attachment: metadata plus the bytes to send.
type Item struct {
	Meta Meta
	Data []byte
}

// Staging holds the attachments of the message being composed, in memory,
// until they are sent (moved to the database) or removed. The per-message
// limits are enforced here, so a send is never refused for size or count.
// Safe for concurrent use.
type Staging struct {
	mu    sync.Mutex
	items map[string]Item
	total int64
	newID func() string
}

func NewStaging() *Staging {
	return &Staging{items: map[string]Item{}, newID: uuid.NewString}
}

// Add stores data under a fresh id and returns the completed Meta.
func (s *Staging) Add(m Meta, data []byte) (Meta, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.items) >= MaxPerMessage {
		return Meta{}, ErrTooMany
	}
	if s.total+int64(len(data)) > MaxMessageBytes {
		return Meta{}, ErrTooBig
	}
	m.ID = s.newID()
	m.Size = int64(len(data))
	s.items[m.ID] = Item{Meta: m, Data: data}
	s.total += m.Size
	return m, nil
}

// Get returns one staged item. Data is shared, not copied; callers only read.
func (s *Staging) Get(id string) (Item, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	it, ok := s.items[id]
	return it, ok
}

// Select returns the items for ids, in that order. Any unknown or repeated
// id fails the whole call, so a send never silently drops an attachment or
// stores one twice. Nothing is removed.
func (s *Staging) Select(ids []string) ([]Item, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Item, 0, len(ids))
	seen := make(map[string]bool, len(ids))
	for _, id := range ids {
		if seen[id] {
			return nil, ErrDuplicate
		}
		seen[id] = true
		it, ok := s.items[id]
		if !ok {
			return nil, ErrUnknown
		}
		out = append(out, it)
	}
	return out, nil
}

// Remove drops ids; unknown ids are ignored.
func (s *Staging) Remove(ids ...string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, id := range ids {
		if it, ok := s.items[id]; ok {
			s.total -= it.Meta.Size
			delete(s.items, id)
		}
	}
}

// Clear drops everything (shutdown).
func (s *Staging) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items = map[string]Item{}
	s.total = 0
}

// Len reports how many items are staged.
func (s *Staging) Len() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.items)
}
```

- [ ] **Step 4: Jalankan, pastikan lulus**

Run: `go test ./internal/attach/ -v`
Expected: PASS semua.

- [ ] **Step 5: Commit**

```bash
git add internal/attach/staging.go internal/attach/staging_test.go
git commit -m "feat(attach): in-memory staging with per-message limits"
```

---

### Task 5: storage — migrasi v5, `AppendUserTurn`, `RecentContext`

**Files:**
- Modify: `internal/storage/db.go` (`currentVersion`, `migrations[5]`)
- Create: `internal/storage/attachments.go`
- Modify: `internal/storage/chat.go` (`insertMessage`, hapus `RecentMessages`, tambah `ContextRow` + `RecentContext`, hapus import `companion/internal/ai`)
- Modify: `internal/storage/chat_test.go:30-35`, `internal/storage/db_test.go:39`
- Test: `internal/storage/attachments_test.go`

**Interfaces:**
- Produces: `storage.AttachmentRow{ID, MessageID, Kind, Mime, Name string; Size int64; Position int; Data []byte}`; `(*DB).AppendUserTurn(user MessageRow, atts []AttachmentRow, assistant MessageRow) error`; `(*DB).AttachmentsFor(messageIDs []string) (map[string][]AttachmentRow, error)` (tanpa `Data`); `(*DB).AttachmentData(id string) (mime string, data []byte, err error)`; `storage.ContextRow{ID, Role, Content string; Attachments []AttachmentRow}`; `(*DB).RecentContext(conversationID string, n int) ([]ContextRow, error)` oldest-first, tanpa baris assistant kosong.
- Removes: `(*DB).RecentMessages`.

- [ ] **Step 1: Tulis tes yang gagal**

`internal/storage/attachments_test.go`:

```go
package storage

import (
	"path/filepath"
	"testing"
)

func openTemp(t *testing.T) *DB {
	t.Helper()
	db, err := Open(filepath.Join(t.TempDir(), "a.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func countAttachments(t *testing.T, db *DB) int {
	t.Helper()
	var n int
	if err := db.inner.QueryRow(`SELECT COUNT(*) FROM attachments`).Scan(&n); err != nil {
		t.Fatalf("count: %v", err)
	}
	return n
}

func userRow(id, text string) MessageRow {
	return MessageRow{ID: id, ConversationID: "c1", Role: "user", Content: text, Status: "complete"}
}

func asstRow(id string) MessageRow {
	return MessageRow{ID: id, ConversationID: "c1", Role: "assistant", Content: "", Status: "streaming"}
}

func TestMigrationV5CreatesAttachments(t *testing.T) {
	db := openTemp(t)
	var v int
	if err := db.inner.QueryRow(`SELECT MAX(version) FROM schema_version`).Scan(&v); err != nil || v != 5 {
		t.Fatalf("version %d %v", v, err)
	}
	if countAttachments(t, db) != 0 {
		t.Fatal("fresh table not empty")
	}
}

func TestMigrationV4toV5KeepsExistingData(t *testing.T) {
	path := filepath.Join(t.TempDir(), "old.db")
	db, err := Open(path)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if err := db.EnsureConversation("c1", "lama"); err != nil {
		t.Fatal(err)
	}
	if err := db.AppendMessage(userRow("u1", "halo")); err != nil {
		t.Fatal(err)
	}
	// Roll the schema back to v4 so the next Open must run migration 5 on a
	// database that already holds conversations.
	if _, err := db.inner.Exec(`DROP TABLE attachments`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.inner.Exec(`DELETE FROM schema_version WHERE version=5`); err != nil {
		t.Fatal(err)
	}
	_ = db.Close()
	db2, err := Open(path)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	defer db2.Close()
	if countAttachments(t, db2) != 0 {
		t.Fatal("attachments missing after upgrade")
	}
	convs, err := db2.ListConversations()
	if err != nil || len(convs) != 1 || convs[0].Title != "lama" {
		t.Fatalf("data lost: %+v %v", convs, err)
	}
}

func TestAppendUserTurnIsAtomic(t *testing.T) {
	db := openTemp(t)
	_ = db.EnsureConversation("c1", "x")
	// Two attachments with the same id: the second insert fails, so nothing
	// from this turn may remain.
	err := db.AppendUserTurn(userRow("u1", "lihat"), []AttachmentRow{
		{ID: "a1", Kind: "image", Mime: "image/png", Name: "x.png", Data: []byte{1}},
		{ID: "a1", Kind: "image", Mime: "image/png", Name: "y.png", Data: []byte{2}},
	}, asstRow("m1"))
	if err == nil {
		t.Fatal("expected duplicate id error")
	}
	rows, _ := db.RecentRows("c1", 10)
	if len(rows) != 0 || countAttachments(t, db) != 0 {
		t.Fatalf("partial turn stored: %+v", rows)
	}
}

func TestAppendUserTurnStoresAndCascades(t *testing.T) {
	db := openTemp(t)
	_ = db.EnsureConversation("c1", "x")
	err := db.AppendUserTurn(userRow("u1", "lihat"), []AttachmentRow{
		{ID: "a1", Kind: "image", Mime: "image/png", Name: "x.png", Data: []byte{1}},
		{ID: "a2", Kind: "pdf", Mime: "application/pdf", Name: "nota.pdf", Data: []byte("%PDF")},
	}, asstRow("m1"))
	if err != nil {
		t.Fatalf("append: %v", err)
	}
	atts, err := db.AttachmentsFor([]string{"u1", "m1"})
	if err != nil || len(atts["u1"]) != 2 || len(atts["m1"]) != 0 {
		t.Fatalf("for: %+v %v", atts, err)
	}
	if a := atts["u1"][1]; a.Name != "nota.pdf" || a.Position != 1 || a.Size != 4 || a.Data != nil {
		t.Fatalf("metadata: %+v", a)
	}
	mime, data, err := db.AttachmentData("a1")
	if err != nil || mime != "image/png" || len(data) != 1 || data[0] != 1 {
		t.Fatalf("data: %q %v %v", mime, data, err)
	}
	ctx, err := db.RecentContext("c1", 20)
	if err != nil || len(ctx) != 1 || ctx[0].ID != "u1" || len(ctx[0].Attachments) != 2 || ctx[0].Attachments[0].Data != nil {
		t.Fatalf("context: %+v %v", ctx, err)
	}
	if err := db.DeleteConversation("c1"); err != nil {
		t.Fatal(err)
	}
	if countAttachments(t, db) != 0 {
		t.Fatal("cascade did not remove attachments")
	}
	if _, _, err := db.AttachmentData("a1"); err == nil {
		t.Fatal("deleted attachment still readable")
	}
}

func TestRecentContextSkipsEmptyAssistantRowAndKeepsOrder(t *testing.T) {
	db := openTemp(t)
	_ = db.EnsureConversation("c1", "x")
	if err := db.AppendUserTurn(userRow("u1", "satu"), nil, asstRow("m1")); err != nil {
		t.Fatal(err)
	}
	_ = db.UpdateMessage("m1", "jawab", "complete")
	if err := db.AppendUserTurn(userRow("u2", "dua"), nil, asstRow("m2")); err != nil {
		t.Fatal(err)
	}
	ctx, err := db.RecentContext("c1", 20)
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, r := range ctx {
		got = append(got, r.Role+":"+r.Content)
	}
	want := []string{"user:satu", "assistant:jawab", "user:dua"}
	if len(got) != len(want) {
		t.Fatalf("got %v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v want %v", got, want)
		}
	}
}
```

Ubah `internal/storage/chat_test.go` baris 30-35: ganti `db.RecentMessages("c1", 20)` → `db.RecentContext("c1", 20)` (field `Role`/`Content` sama). Ubah `internal/storage/db_test.go` baris 39: `db.RecentMessages("c1", 5)` → `db.RecentContext("c1", 5)`.

- [ ] **Step 2: Jalankan, pastikan gagal**

Run: `go test ./internal/storage/ -v`
Expected: compile error `undefined: AttachmentRow` / `RecentContext`.

- [ ] **Step 3: Implementasi**

`internal/storage/db.go`: `const currentVersion = 5` dan tambahkan ke map `migrations`:

```go
	// v5: attachments (images, PDFs, text files) stored as BLOBs next to the
	// user message they belong to. Cascade from messages, so deleting a
	// conversation frees them; metadata reads never select `data`.
	5: {
		`CREATE TABLE attachments (
			id TEXT PRIMARY KEY,
			message_id TEXT NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
			kind TEXT NOT NULL,
			mime TEXT NOT NULL,
			name TEXT NOT NULL,
			size INTEGER NOT NULL,
			data BLOB NOT NULL,
			position INTEGER NOT NULL,
			created_at TEXT NOT NULL
		)`,
		`CREATE INDEX idx_attachments_msg ON attachments(message_id, position)`,
	},
```

`internal/storage/chat.go`: ganti `AppendMessage` dengan:

```go
// execer is satisfied by *sql.DB and *sql.Tx, so one insert serves both the
// single-row path and the transactional user-turn path.
type execer interface {
	Exec(query string, args ...any) (sql.Result, error)
}

func insertMessage(ex execer, m MessageRow) error {
	now := time.Now().UTC().Format(time.RFC3339)
	if m.CreatedAt.IsZero() {
		m.CreatedAt = time.Now().UTC()
	}
	_, err := ex.Exec(`INSERT INTO messages
		(id, conversation_id, role, content, status, provider_id, model, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		m.ID, m.ConversationID, m.Role, m.Content, m.Status,
		nullStr(m.ProviderID), nullStr(m.Model),
		m.CreatedAt.Format(time.RFC3339), now)
	return err
}

// AppendMessage inserts one message with complete metadata.
func (d *DB) AppendMessage(m MessageRow) error { return insertMessage(d.inner, m) }
```

Hapus seluruh fungsi `RecentMessages` (dan komentarnya) serta import `"companion/internal/ai"`; tambahkan:

```go
// ContextRow is one message for building the provider request: text plus
// attachment metadata. Bytes are loaded separately, only when needed.
type ContextRow struct {
	ID          string
	Role        string
	Content     string
	Attachments []AttachmentRow
}

// RecentContext returns the last N messages oldest-first with attachment
// metadata. The empty assistant row created for the reply in flight is left
// out: sending {"role":"assistant","content":""} confuses stricter providers.
func (d *DB) RecentContext(conversationID string, n int) ([]ContextRow, error) {
	if n <= 0 {
		n = 20
	}
	rows, err := d.inner.Query(`SELECT id, role, content FROM messages
		WHERE conversation_id=? AND status IN ('complete','streaming')
		AND NOT (role='assistant' AND content='')
		ORDER BY created_at DESC, rowid DESC LIMIT ?`, conversationID, n)
	if err != nil {
		return nil, err
	}
	var rev []ContextRow
	for rows.Next() {
		var r ContextRow
		if err := rows.Scan(&r.ID, &r.Role, &r.Content); err != nil {
			rows.Close()
			return nil, err
		}
		if r.Role != "user" && r.Role != "assistant" && r.Role != "system" {
			continue
		}
		rev = append(rev, r)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	rows.Close()
	out := make([]ContextRow, 0, len(rev))
	ids := make([]string, 0, len(rev))
	for i := len(rev) - 1; i >= 0; i-- {
		out = append(out, rev[i])
		ids = append(ids, rev[i].ID)
	}
	atts, err := d.AttachmentsFor(ids)
	if err != nil {
		return nil, err
	}
	for i := range out {
		out[i].Attachments = atts[out[i].ID]
	}
	return out, nil
}
```

`internal/storage/attachments.go`:

```go
package storage

import (
	"strings"
	"time"
)

// AttachmentRow maps the attachments table. Data is set only on insert and
// by AttachmentData; metadata reads leave it nil.
type AttachmentRow struct {
	ID        string
	MessageID string
	Kind      string
	Mime      string
	Name      string
	Size      int64
	Position  int
	Data      []byte
}

// AppendUserTurn stores a user message, its attachments, and the empty
// streaming assistant row in one transaction, so a crash can never leave a
// question without its images or an orphan reply.
func (d *DB) AppendUserTurn(user MessageRow, atts []AttachmentRow, assistant MessageRow) error {
	tx, err := d.inner.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck // no-op after Commit
	if err := insertMessage(tx, user); err != nil {
		return err
	}
	now := time.Now().UTC().Format(time.RFC3339)
	for i, a := range atts {
		if _, err := tx.Exec(`INSERT INTO attachments
			(id, message_id, kind, mime, name, size, data, position, created_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			a.ID, user.ID, a.Kind, a.Mime, a.Name, int64(len(a.Data)), a.Data, i, now); err != nil {
			return err
		}
	}
	if err := insertMessage(tx, assistant); err != nil {
		return err
	}
	return tx.Commit()
}

// AttachmentsFor returns metadata (no bytes) grouped by message id, each
// group in position order. Missing ids simply have no entry.
func (d *DB) AttachmentsFor(messageIDs []string) (map[string][]AttachmentRow, error) {
	out := map[string][]AttachmentRow{}
	if len(messageIDs) == 0 {
		return out, nil
	}
	q := `SELECT id, message_id, kind, mime, name, size, position FROM attachments
		WHERE message_id IN (?` + strings.Repeat(",?", len(messageIDs)-1) + `)
		ORDER BY message_id, position`
	args := make([]any, len(messageIDs))
	for i, id := range messageIDs {
		args[i] = id
	}
	rows, err := d.inner.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var a AttachmentRow
		if err := rows.Scan(&a.ID, &a.MessageID, &a.Kind, &a.Mime, &a.Name, &a.Size, &a.Position); err != nil {
			return nil, err
		}
		out[a.MessageID] = append(out[a.MessageID], a)
	}
	return out, rows.Err()
}

// AttachmentData returns one attachment's mime and bytes. sql.ErrNoRows
// when the id is unknown (or its conversation was deleted).
func (d *DB) AttachmentData(id string) (string, []byte, error) {
	var mime string
	var data []byte
	err := d.inner.QueryRow(`SELECT mime, data FROM attachments WHERE id=?`, id).Scan(&mime, &data)
	return mime, data, err
}
```

- [ ] **Step 4: Jalankan, pastikan lulus**

Run: `go test ./internal/storage/ -v && go build ./...`
Expected: PASS semua; `go build` gagal hanya di `internal/app` (`RecentMessages` hilang) — diperbaiki Task 7.

- [ ] **Step 5: Commit**

```bash
git add internal/storage/
git commit -m "feat(storage): attachments table, atomic user turn, context with metadata"
```

---

### Task 6: `internal/app` — `buildContext` (kebijakan 3 turn)

**Files:**
- Create: `internal/app/context.go`
- Test: `internal/app/context_test.go`

**Interfaces:**
- Consumes: `storage.ContextRow`, `storage.AttachmentRow`, `ai.Part` konstruktor, `attach.Kind*`.
- Produces: `const AttachmentTurns = 3`; `buildContext(rows []storage.ContextRow, turns int, load func(id string) (string, []byte, error)) []ai.Message`; `partFor(kind, name, mime string, data []byte) ai.Part`; `placeholder(a storage.AttachmentRow) string`.

- [ ] **Step 1: Tulis tes yang gagal**

`internal/app/context_test.go`:

```go
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
```

- [ ] **Step 2: Jalankan, pastikan gagal**

Run: `go test ./internal/app/ -run TestBuildContext -v`
Expected: compile error (`buildContext`, dan `service.go` masih memanggil `RecentMessages`). Lanjut ke Step 3; `service.go` dibereskan sementara dengan mengganti blok `RecentMessages` di `generate` (akan ditulis ulang penuh di Task 7):

```go
		if db != nil {
			if rows, err := db.RecentContext(conversationID, 20); err == nil && len(rows) > 0 {
				msgs = buildContext(rows, AttachmentTurns, db.AttachmentData)
			}
		}
```

- [ ] **Step 3: Implementasi**

`internal/app/context.go`:

```go
package app

import (
	"strings"

	"companion/internal/ai"
	"companion/internal/attach"
	"companion/internal/storage"
)

// AttachmentTurns is how many of the newest user turns carry their
// attachments in full; older turns name them in text instead. Providers
// re-bill images on every request (docs 05 §4), so the window stays small
// while follow-ups about a recent image still work.
const AttachmentTurns = 3

// buildContext turns stored rows into provider messages, applying the
// AttachmentTurns window. load fetches bytes by attachment id; a load error
// degrades that attachment to its placeholder instead of failing the turn.
func buildContext(rows []storage.ContextRow, turns int, load func(id string) (string, []byte, error)) []ai.Message {
	full := map[string]bool{}
	seen := 0
	for i := len(rows) - 1; i >= 0 && seen < turns; i-- {
		if rows[i].Role == "user" {
			full[rows[i].ID] = true
			seen++
		}
	}
	out := make([]ai.Message, 0, len(rows))
	for _, r := range rows {
		m := ai.Message{Role: r.Role, Content: r.Content}
		if r.Role == "user" && len(r.Attachments) > 0 {
			if full[r.ID] {
				m.Parts = partsFromRows(r.Attachments, load)
			} else {
				m.Content = withPlaceholders(r.Content, r.Attachments)
			}
		}
		out = append(out, m)
	}
	return out
}

func partsFromRows(atts []storage.AttachmentRow, load func(string) (string, []byte, error)) []ai.Part {
	parts := make([]ai.Part, 0, len(atts))
	for _, a := range atts {
		mime, data, err := load(a.ID)
		if err != nil {
			parts = append(parts, ai.TextPart(placeholder(a)))
			continue
		}
		parts = append(parts, partFor(a.Kind, a.Name, mime, data))
	}
	return parts
}

// partFor maps one attachment to its wire part: images as image_url, PDFs
// as a file part, text inlined as a fenced block (works on any provider).
func partFor(kind, name, mime string, data []byte) ai.Part {
	switch kind {
	case attach.KindImage:
		return ai.ImagePart(mime, data)
	case attach.KindPDF:
		return ai.FilePartFor(name, mime, data)
	default:
		return ai.InlineTextPart(name, data)
	}
}

func placeholder(a storage.AttachmentRow) string {
	if a.Kind == attach.KindImage {
		return "[gambar: " + a.Name + "]"
	}
	return "[berkas: " + a.Name + "]"
}

func withPlaceholders(text string, atts []storage.AttachmentRow) string {
	lines := make([]string, 0, len(atts)+1)
	if text != "" {
		lines = append(lines, text)
	}
	for _, a := range atts {
		lines = append(lines, placeholder(a))
	}
	return strings.Join(lines, "\n")
}
```

- [ ] **Step 4: Jalankan, pastikan lulus**

Run: `go test ./internal/app/ -v`
Expected: PASS semua (termasuk tes service lama; `generate` sementara memakai `RecentContext`).

- [ ] **Step 5: Commit**

```bash
git add internal/app/context.go internal/app/context_test.go internal/app/service.go
git commit -m "feat(app): build provider context with a 3-turn attachment window"
```

---

### Task 7: `internal/app` — staging di `Service`, `SendMessage` dengan lampiran, `History`

**Files:**
- Create: `internal/app/attachments.go`
- Modify: `internal/app/service.go` (`Service.staging`, `SendMessage`, `RetryMessage`, `generate`, `CancelAll`)
- Modify: `internal/app/history.go` (`ChatMessage.Attachments`, `History`)
- Modify: `internal/app/service_test.go` (pemanggil `SendMessage` lama → tambah `nil`)
- Test: `internal/app/attachments_test.go`

**Interfaces:**
- Consumes: `attach.Staging`, `attach.Classify`, `storage.AppendUserTurn`, `storage.RecentContext`, `buildContext`.
- Produces: `(*Service).SendMessage(ctx, conversationID, text string, attachmentIDs []string) (string, error)`; `(*Service).StagePaths([]string) ([]attach.Meta, error)`; `(*Service).StageData(name, mime, b64 string) (attach.Meta, error)`; `(*Service).Unstage(id string)`; `(*Service).Attachment(id string) (mime string, data []byte, ok bool)`; `ChatMessage.Attachments []attach.Meta` (`json:"attachments"`, selalu non-nil); `errNoAttachmentStorage`.

- [ ] **Step 1: Tulis tes yang gagal**

`internal/app/attachments_test.go`:

```go
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
```

Di `internal/app/service_test.go` ganti setiap `svc.SendMessage(ctx, "c1", "halo")`-style call menjadi `svc.SendMessage(ctx, "c1", "halo", nil)` (empat tempat: `TestSendAndCancel` ×2, `TestEmptyRejected`, `TestCancelAllStopsStreams`/`TestCancelAllWaitsForGoroutines` bila memanggil `SendMessage`; cek dengan `grep -n "SendMessage(" internal/app/service_test.go`).

- [ ] **Step 2: Jalankan, pastikan gagal**

Run: `go test ./internal/app/ -run 'TestSend|TestRetryResends|TestStage|TestUnstage' -v`
Expected: compile error (`StageData`, `errNoAttachmentStorage`, `Attachments`).

- [ ] **Step 3: Implementasi**

`internal/app/attachments.go`:

```go
package app

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path"
	"strings"

	"companion/internal/ai"
	"companion/internal/attach"
	"companion/internal/storage"
)

var errNoAttachmentStorage = errors.New("lampiran butuh storage, restart aplikasi")

var extByMime = map[string]string{"image/png": ".png", "image/jpeg": ".jpg", "image/webp": ".webp"}

// StagePaths reads user-chosen files (dialog or drop) into staging. The
// batch is all-or-nothing: one bad file rejects the call and stages nothing,
// and the error names the file, never its directory.
func (s *Service) StagePaths(paths []string) ([]attach.Meta, error) {
	if s.db() == nil {
		return nil, errNoAttachmentStorage
	}
	type pending struct {
		meta attach.Meta
		data []byte
	}
	batch := make([]pending, 0, len(paths))
	for _, p := range paths {
		name := attach.CleanName(p)
		info, err := os.Stat(p)
		if err != nil {
			return nil, fmt.Errorf("%s: tidak bisa dibaca", name)
		}
		if info.IsDir() {
			return nil, fmt.Errorf("%s: folder tidak bisa dilampirkan", name)
		}
		if info.Size() > attach.MaxFileBytes {
			return nil, fmt.Errorf("%s: melebihi %d MB", name, attach.MaxFileBytes>>20)
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return nil, fmt.Errorf("%s: tidak bisa dibaca", name)
		}
		m, err := attach.Classify(p, data)
		if err != nil {
			return nil, err
		}
		batch = append(batch, pending{meta: m, data: data})
	}
	out := make([]attach.Meta, 0, len(batch))
	for _, b := range batch {
		m, err := s.staging.Add(b.meta, b.data)
		if err != nil {
			for _, done := range out {
				s.staging.Remove(done.ID)
			}
			return nil, err
		}
		out = append(out, m)
	}
	return out, nil
}

// StageData stages bytes the WebView already holds (a pasted image). b64 is
// standard base64 without a data: prefix. A missing name or extension is
// derived from mime so the stored name still says what it is.
func (s *Service) StageData(name, mime, b64 string) (attach.Meta, error) {
	if s.db() == nil {
		return attach.Meta{}, errNoAttachmentStorage
	}
	data, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return attach.Meta{}, errors.New("data lampiran tidak valid")
	}
	if strings.TrimSpace(name) == "" || path.Ext(attach.CleanName(name)) == "" {
		name = "tempel" + extByMime[strings.ToLower(mime)]
	}
	m, err := attach.Classify(name, data)
	if err != nil {
		return attach.Meta{}, err
	}
	return s.staging.Add(m, data)
}

// Unstage drops a pending attachment; unknown ids are ignored.
func (s *Service) Unstage(id string) { s.staging.Remove(id) }

// Attachment finds bytes for the asset route: staging first (not yet sent),
// then the database (sent earlier).
func (s *Service) Attachment(id string) (string, []byte, bool) {
	if it, ok := s.staging.Get(id); ok {
		return it.Meta.Mime, it.Data, true
	}
	db := s.db()
	if db == nil {
		return "", nil, false
	}
	mime, data, err := db.AttachmentData(id)
	if err != nil {
		return "", nil, false
	}
	return mime, data, true
}

func partsFromItems(items []attach.Item) []ai.Part {
	if len(items) == 0 {
		return nil
	}
	parts := make([]ai.Part, 0, len(items))
	for _, it := range items {
		parts = append(parts, partFor(it.Meta.Kind, it.Meta.Name, it.Meta.Mime, it.Data))
	}
	return parts
}

func attachmentRows(messageID string, items []attach.Item) []storage.AttachmentRow {
	rows := make([]storage.AttachmentRow, 0, len(items))
	for i, it := range items {
		rows = append(rows, storage.AttachmentRow{
			ID: it.Meta.ID, MessageID: messageID, Kind: it.Meta.Kind, Mime: it.Meta.Mime,
			Name: it.Meta.Name, Size: it.Meta.Size, Position: i, Data: it.Data,
		})
	}
	return rows
}

func metasOf(rows []storage.AttachmentRow) []attach.Meta {
	out := make([]attach.Meta, 0, len(rows))
	for _, a := range rows {
		out = append(out, attach.Meta{ID: a.ID, Name: a.Name, Mime: a.Mime, Size: a.Size, Kind: a.Kind})
	}
	return out
}

// titleFor names a conversation from its first text, or the first file when
// the user sent only attachments.
func titleFor(text string, items []attach.Item) string {
	if text != "" || len(items) == 0 {
		return storage.TitleFor(text)
	}
	return storage.TitleFor(items[0].Meta.Name)
}

func hasRow(rows []storage.ContextRow, id string) bool {
	for _, r := range rows {
		if r.ID == id {
			return true
		}
	}
	return false
}
```

`internal/app/service.go`:

1. Tambah import `"companion/internal/attach"`.
2. Field `staging *attach.Staging` di `Service`; di `NewService`: `staging: attach.NewStaging(),`.
3. Ganti `SendMessage`, `RetryMessage`, dan seluruh `generate` dengan:

```go
// SendMessage starts streaming. Text may be empty when attachments are
// given; both empty is rejected. One active generation per conversation.
func (s *Service) SendMessage(ctx context.Context, conversationID, text string, attachmentIDs []string) (string, error) {
	return s.generate(ctx, conversationID, text, false, attachmentIDs)
}

// RetryMessage regenerates the newest turn after a failure. Unlike
// SendMessage it does not store the user's text again when that turn is
// already in history; the failed reply is replaced (storage.PrepareRetry)
// and the turn's stored attachments are sent again from the database.
func (s *Service) RetryMessage(ctx context.Context, conversationID, text string) (string, error) {
	return s.generate(ctx, conversationID, text, true, nil)
}

func (s *Service) generate(ctx context.Context, conversationID, text string, retry bool, attachmentIDs []string) (string, error) {
	text = strings.TrimSpace(text)
	// A retry may carry empty text: the stored turn decides (PrepareRetry).
	if text == "" && len(attachmentIDs) == 0 && !retry {
		return "", errors.New("empty message")
	}
	if conversationID == "" {
		return "", errors.New("empty conversation")
	}
	if len(attachmentIDs) > 0 && s.db() == nil {
		return "", errNoAttachmentStorage
	}
	s.mu.Lock()
	if _, ok := s.active[conversationID]; ok {
		s.mu.Unlock()
		return "", errors.New("generation already active")
	}
	// Attachments are resolved before the goroutine so a bad id is a
	// synchronous rejection, never a half-sent turn. The items hold the bytes
	// from here on; staging is cleared now so the composer can start fresh.
	var items []attach.Item
	if len(attachmentIDs) > 0 {
		var err error
		if items, err = s.staging.Select(attachmentIDs); err != nil {
			s.mu.Unlock()
			return "", err
		}
		s.staging.Remove(attachmentIDs...)
	}
	genCtx, cancel := context.WithCancel(ctx)
	s.active[conversationID] = cancel
	s.mu.Unlock()

	requestID := uuid.NewString()
	messageID := uuid.NewString()
	s.mu.Lock()
	s.requests[requestID] = conversationID
	s.mu.Unlock()

	s.emit("chat:event", ai.ChatEvent{Type: "started", RequestID: requestID, ConversationID: conversationID, MessageID: messageID})

	s.wg.Add(1)
	go func() {
		defer func() {
			s.mu.Lock()
			delete(s.active, conversationID)
			delete(s.requests, requestID)
			s.mu.Unlock()
			s.wg.Done()
		}()
		fail := func(pub *ai.PublicError) {
			s.emit("chat:event", ai.ChatEvent{Type: "failed", RequestID: requestID, ConversationID: conversationID, MessageID: messageID, Error: pub})
		}
		db := s.db()
		// Reconcile a retry before anything can fail, so history and the UI
		// (which drops the failed bubble once the retry is accepted) agree.
		// On a DB error, assume the user row is there: appending it again
		// would duplicate the question, which is the bug retry exists to avoid.
		reused := false
		if retry && db != nil {
			ok, err := db.PrepareRetry(conversationID, text)
			reused = ok || err != nil
		}
		if !reused && text == "" && len(items) == 0 {
			fail(&ai.PublicError{Code: "empty", Message: "Tidak ada yang bisa diulang."})
			return
		}
		live, model, resolveErr := s.resolve(genCtx)
		if resolveErr != nil {
			fail(&ai.PublicError{Code: "no_provider", Message: resolveErr.Error()})
			return
		}
		userMsgID := uuid.NewString()
		current := ai.Message{Role: "user", Content: text, Parts: partsFromItems(items)}
		if db != nil {
			// Persistence is best-effort by design: a DB failure must not abort
			// an in-flight stream. The error surfaces later when the history
			// list is empty after restart; see ADR-037.
			_ = db.EnsureConversation(conversationID, titleFor(text, items))
			assistant := storage.MessageRow{ID: messageID, ConversationID: conversationID, Role: "assistant", Content: "", Status: "streaming", Model: model}
			if reused {
				_ = db.AppendMessage(assistant)
			} else {
				user := storage.MessageRow{ID: userMsgID, ConversationID: conversationID, Role: "user", Content: text, Status: "complete", Model: model}
				_ = db.AppendUserTurn(user, attachmentRows(userMsgID, items), assistant)
			}
		}
		msgs := []ai.Message{current}
		if db != nil {
			if rows, err := db.RecentContext(conversationID, 20); err == nil && len(rows) > 0 {
				msgs = buildContext(rows, AttachmentTurns, db.AttachmentData)
				// A failed insert must not drop the question itself.
				if !reused && !hasRow(rows, userMsgID) {
					msgs = append(msgs, current)
				}
			}
		}
		var sb strings.Builder
		err := live.StreamChat(genCtx, ai.ChatRequest{Model: model, Messages: msgs}, func(chunk string) error {
			sb.WriteString(chunk)
			s.emit("chat:event", ai.ChatEvent{Type: "delta", RequestID: requestID, ConversationID: conversationID, MessageID: messageID, Delta: chunk})
			return nil
		})
		final := sb.String()
		switch {
		case err == nil:
			if db != nil {
				_ = db.UpdateMessage(messageID, final, "complete")
			}
			s.emit("chat:event", ai.ChatEvent{Type: "completed", RequestID: requestID, ConversationID: conversationID, MessageID: messageID})
		case errors.Is(err, context.Canceled):
			if db != nil {
				_ = db.UpdateMessage(messageID, final, "cancelled")
			}
			s.emit("chat:event", ai.ChatEvent{Type: "cancelled", RequestID: requestID, ConversationID: conversationID, MessageID: messageID})
		default:
			if db != nil {
				_ = db.UpdateMessage(messageID, final, "failed")
			}
			var pub *ai.PublicError
			if errors.As(err, &pub) {
				fail(pub)
				return
			}
			fail(&ai.PublicError{Code: "network", Message: "Koneksi gagal. Coba lagi.", Retryable: true})
		}
	}()
	return requestID, nil
}
```

4. Di `CancelAll`, setelah `s.wg.Wait()` tambahkan `s.staging.Clear()` dengan komentar `// Shutdown: pending attachments are not persisted.`

`internal/app/history.go`: import `"companion/internal/attach"` dan `"companion/internal/storage"`; `ChatMessage` tambah field `Attachments []attach.Meta \`json:"attachments"\``; ganti badan `History` setelah `rows, err := db.RecentRows(...)` dengan:

```go
	ids := make([]string, 0, len(rows))
	for _, r := range rows {
		ids = append(ids, r.ID)
	}
	atts, err := db.AttachmentsFor(ids)
	if err != nil {
		atts = map[string][]storage.AttachmentRow{}
	}
	out := make([]ChatMessage, 0, len(rows))
	for _, r := range rows {
		if r.Role != "user" && r.Role != "assistant" {
			continue
		}
		out = append(out, ChatMessage{
			ID: r.ID, Role: r.Role, Content: r.Content, Status: r.Status,
			Attachments: metasOf(atts[r.ID]),
		})
	}
	return out
```

- [ ] **Step 4: Jalankan, pastikan lulus**

Run: `go test ./internal/... -v 2>&1 | tail -40 && go vet ./internal/...`
Expected: PASS semua (`app.go` masih gagal build untuk `./...` sampai Task 9; `go vet ./internal/...` bersih).

- [ ] **Step 5: Commit**

```bash
git add internal/app/
git commit -m "feat(app): stage, persist, and send attachments with the user turn"
```

---

### Task 8: `internal/app` — route `/attachments/{id}`

**Files:**
- Create: `internal/app/attachment_handler.go`
- Test: `internal/app/attachment_handler_test.go`

**Interfaces:**
- Consumes: `(*Service).Attachment(id)`.
- Produces: `(*Service).AttachmentHandler() http.Handler`.

- [ ] **Step 1: Tulis tes yang gagal**

`internal/app/attachment_handler_test.go`:

```go
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
```

- [ ] **Step 2: Jalankan, pastikan gagal**

Run: `go test ./internal/app/ -run TestAttachmentHandler -v`
Expected: compile error `AttachmentHandler undefined`.

- [ ] **Step 3: Implementasi**

`internal/app/attachment_handler.go`:

```go
package app

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

var attachmentIDRe = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

// AttachmentHandler serves staged or stored attachment bytes to the WebView
// at GET /attachments/{id}. Only a well-formed uuid is looked up; everything
// else is 404, so the route can never be used to probe paths or files.
func (s *Service) AttachmentHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		id, ok := strings.CutPrefix(r.URL.Path, "/attachments/")
		if !ok || !attachmentIDRe.MatchString(id) {
			http.NotFound(w, r)
			return
		}
		mime, data, found := s.Attachment(id)
		if !found {
			http.NotFound(w, r)
			return
		}
		h := w.Header()
		h.Set("Content-Type", mime)
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("Cache-Control", "private, max-age=3600")
		h.Set("Content-Disposition", "inline")
		h.Set("Content-Length", strconv.Itoa(len(data)))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
	})
}
```

- [ ] **Step 4: Jalankan, pastikan lulus**

Run: `go test ./internal/app/ -v 2>&1 | tail -30`
Expected: PASS semua.

- [ ] **Step 5: Commit**

```bash
git add internal/app/attachment_handler.go internal/app/attachment_handler_test.go
git commit -m "feat(app): serve attachments to the webview by uuid"
```

---

### Task 9: `app.go`, `main.go`, regenerasi binding `wailsjs`

**Files:**
- Modify: `app.go` (`SendMessage`; tambah `PickAttachments`, `StageAttachmentPaths`, `StageAttachmentData`, `UnstageAttachment`, `assetHandler`)
- Modify: `main.go` (`AssetServer.Handler`, `DragAndDrop`)
- Modify (generated): `frontend/wailsjs/go/main/App.d.ts`, `App.js`, `frontend/wailsjs/go/models.ts`

**Interfaces:**
- Produces (JS): `PickAttachments(): Promise<attach.Meta[]>`, `StageAttachmentPaths(paths: string[]): Promise<attach.Meta[]>`, `StageAttachmentData(name, mime, data: string): Promise<attach.Meta>`, `UnstageAttachment(id: string): Promise<void>`, `SendMessage(conv, text: string, attachmentIds: string[]): Promise<string>`; `app.ChatMessage.attachments: attach.Meta[]`.

- [ ] **Step 1: Ubah `app.go`**

Tambah import `"net/http"` dan `"companion/internal/attach"`. Ganti `SendMessage` dan tambahkan di bawahnya:

```go
// SendMessage starts a streaming generation. attachmentIDs name staged
// attachments to send with the text; text may be empty when there are any.
func (a *App) SendMessage(conversationID string, text string, attachmentIDs []string) (string, error) {
	return a.svc.SendMessage(a.ctx, conversationID, text, attachmentIDs)
}

// PickAttachments opens the native multi-file dialog and stages the chosen
// files. Cancel returns an empty list, not an error.
func (a *App) PickAttachments() ([]attach.Meta, error) {
	if a.svc == nil {
		return nil, errStorageUnavailable()
	}
	paths, err := runtime.OpenMultipleFilesDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Lampirkan berkas",
		Filters: []runtime.FileFilter{
			{DisplayName: "Gambar, PDF, teks (*.png;*.jpg;*.jpeg;*.webp;*.pdf;*.txt;*.md;*.csv;*.json)", Pattern: "*.png;*.jpg;*.jpeg;*.webp;*.pdf;*.txt;*.md;*.csv;*.json"},
			{DisplayName: "Semua berkas (*.*)", Pattern: "*.*"},
		},
	})
	if err != nil {
		return nil, errors.New("dialog berkas gagal dibuka")
	}
	if len(paths) == 0 {
		return []attach.Meta{}, nil
	}
	return a.svc.StagePaths(paths)
}

// StageAttachmentPaths stages files dropped onto the chat panel.
func (a *App) StageAttachmentPaths(paths []string) ([]attach.Meta, error) {
	if a.svc == nil {
		return nil, errStorageUnavailable()
	}
	return a.svc.StagePaths(paths)
}

// StageAttachmentData stages an image pasted from the clipboard (base64,
// no data: prefix). The name may be empty; the mime then picks the extension.
func (a *App) StageAttachmentData(name string, mime string, data string) (attach.Meta, error) {
	if a.svc == nil {
		return attach.Meta{}, errStorageUnavailable()
	}
	return a.svc.StageData(name, mime, data)
}

// UnstageAttachment drops a pending attachment. Unknown ids are ignored.
func (a *App) UnstageAttachment(id string) {
	if a.svc != nil {
		a.svc.Unstage(id)
	}
}

// assetHandler serves /attachments/{id} to the WebView. main.go needs it
// before startup creates the service, so it looks the service up per call.
func (a *App) assetHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if a.svc == nil {
			http.NotFound(w, r)
			return
		}
		a.svc.AttachmentHandler().ServeHTTP(w, r)
	})
}
```

- [ ] **Step 2: Ubah `main.go`**

```go
		AssetServer: &assetserver.Options{
			Assets:  assets,
			Handler: app.assetHandler(),
		},
		// Files dropped on an element carrying `--wails-drop-target: drop`
		// reach runtime.OnFileDrop as absolute paths; elsewhere they are
		// ignored (main.tsx also cancels the WebView's own file navigation).
		DragAndDrop: &options.DragAndDrop{EnableFileDrop: true},
```

- [ ] **Step 3: Build + vet + format**

Run: `go build ./... && go vet ./... && gofmt -l app.go main.go internal/ && go test ./...`
Expected: build OK, vet bersih, `gofmt -l` tidak mencetak apa pun, semua tes PASS.

- [ ] **Step 4: Regenerasi binding**

Coba (PowerShell): `wails generate module` dari root repo. Bila `wails` tidak ada di PATH: `go install github.com/wailsapp/wails/v2/cmd/wails@v2.16.0` lalu `$env:USERPROFILE\go\bin\wails.exe generate module` (atau path dari `go env GOPATH`).

Bila generator tidak dapat dijalankan, edit tangan mengikuti format generator:

`frontend/wailsjs/go/main/App.d.ts` — tambah import `import {attach} from '../models';` dan baris (urut abjad):

```ts
export function PickAttachments():Promise<Array<attach.Meta>>;
export function SendMessage(arg1:string,arg2:string,arg3:Array<string>):Promise<string>;
export function StageAttachmentData(arg1:string,arg2:string,arg3:string):Promise<attach.Meta>;
export function StageAttachmentPaths(arg1:Array<string>):Promise<Array<attach.Meta>>;
export function UnstageAttachment(arg1:string):Promise<void>;
```

`frontend/wailsjs/go/main/App.js`:

```js
export function PickAttachments() {
  return window['go']['main']['App']['PickAttachments']();
}

export function SendMessage(arg1, arg2, arg3) {
  return window['go']['main']['App']['SendMessage'](arg1, arg2, arg3);
}

export function StageAttachmentData(arg1, arg2, arg3) {
  return window['go']['main']['App']['StageAttachmentData'](arg1, arg2, arg3);
}

export function StageAttachmentPaths(arg1) {
  return window['go']['main']['App']['StageAttachmentPaths'](arg1);
}

export function UnstageAttachment(arg1) {
  return window['go']['main']['App']['UnstageAttachment'](arg1);
}
```

`frontend/wailsjs/go/models.ts` — tambah di atas `export namespace app`:

```ts
export namespace attach {
	
	export class Meta {
	    id: string;
	    name: string;
	    mime: string;
	    size: number;
	    kind: string;
	
	    static createFrom(source: any = {}) {
	        return new Meta(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.mime = source["mime"];
	        this.size = source["size"];
	        this.kind = source["kind"];
	    }
	}

}
```

dan di `app.ChatMessage` tambah field `attachments: attach.Meta[];` serta di constructor `this.attachments = source["attachments"];`.

Verifikasi: `cd frontend && npx tsc --noEmit` (akan gagal di `useChat.ts` karena `SendMessage` butuh 3 argumen — dibereskan Task 10; pastikan tidak ada error lain).

- [ ] **Step 5: Commit**

```bash
git add app.go main.go frontend/wailsjs
git commit -m "feat(app): attachment bindings, asset route, and file drop option"
```

---

### Task 10: frontend — tipe, `useChat.send(text, attachments)`, stub drop

**Files:**
- Modify: `frontend/src/types/chat.ts`
- Modify: `frontend/src/test/wails.ts`
- Modify: `frontend/src/hooks/useChat.ts`
- Test: `frontend/src/hooks/useChat.test.ts` (tambah describe)

**Interfaces:**
- Produces: `AttachmentKind`, `AttachmentMeta{id,name,mime,size,kind}`, `ChatMessage.attachments?: AttachmentMeta[]`; `useChat().send(text: string, attachments?: AttachmentMeta[]): Promise<void>`; `runtimeMock.fileDrop: ((x, y, paths) => void) | null`.

- [ ] **Step 1: Tulis tes yang gagal**

Tambahkan ke akhir `frontend/src/hooks/useChat.test.ts`:

```ts
describe("useChat attachments", () => {
  const att = { id: "a1", name: "foto.png", mime: "image/png", size: 10, kind: "image" as const };

  it("passes attachment ids to SendMessage and shows them on the user bubble", async () => {
    const { result } = await mountChat();
    const conv = result.current.conversationId;
    await act(() => result.current.send("lihat", [att]));
    expect(backend.send).toHaveBeenCalledWith(conv, "lihat", ["a1"]);
    expect(result.current.messages.at(-1)).toMatchObject({ role: "user", content: "lihat", attachments: [att] });
  });

  it("accepts a message with attachments and no text", async () => {
    const { result } = await mountChat();
    const conv = result.current.conversationId;
    await act(() => result.current.send("   ", [att]));
    expect(backend.send).toHaveBeenCalledWith(conv, "", ["a1"]);
  });

  it("still rejects a message with neither text nor attachments", async () => {
    const { result } = await mountChat();
    await act(() => result.current.send("   ", []));
    expect(backend.send).not.toHaveBeenCalled();
  });

  it("loads attachment metadata from history", async () => {
    backend.list.mockResolvedValue([{ id: "c9", title: "t" }]);
    backend.history.mockResolvedValue([{ id: "u1", role: "user", content: "x", status: "complete", attachments: [att] }]);
    const { result } = renderHook(() => useChat());
    await waitFor(() => expect(result.current.messages).toHaveLength(1));
    expect(result.current.messages[0].attachments).toEqual([att]);
  });
});
```

- [ ] **Step 2: Jalankan, pastikan gagal**

Run: `cd frontend && npx vitest run src/hooks/useChat.test.ts`
Expected: 4 tes baru FAIL (`send` mengabaikan argumen kedua; `attachments` undefined).

- [ ] **Step 3: Implementasi**

`frontend/src/types/chat.ts` — tambah:

```ts
export type AttachmentKind = "image" | "pdf" | "text";

/** Frontend-safe attachment description. Bytes stay in the Go backend. */
export interface AttachmentMeta {
  id: string;
  name: string;
  mime: string;
  size: number;
  kind: AttachmentKind;
}
```

dan di `ChatMessage`: `attachments?: AttachmentMeta[];`.

`frontend/src/test/wails.ts` — ganti isi dengan:

```ts
import { vi } from "vitest";

// Shared test double for the Wails runtime. The real module calls
// window.runtime.*, which jsdom does not provide; this emulates the same
// subscribe/emit contract so hooks receive synthetic backend events, and
// records the file-drop listener so a test can fire a drop.
type Handler = (...args: unknown[]) => void;
type DropHandler = (x: number, y: number, paths: string[]) => void;

const listeners = new Map<string, Set<Handler>>();

export const runtimeMock = {
  fileDrop: null as DropHandler | null,
  on(event: string, cb: Handler) {
    let set = listeners.get(event);
    if (!set) {
      set = new Set();
      listeners.set(event, set);
    }
    set.add(cb);
    return () => set!.delete(cb);
  },
  emit(event: string, ...args: unknown[]) {
    listeners.get(event)?.forEach((cb) => cb(...args));
  },
  reset() {
    listeners.clear();
    runtimeMock.fileDrop = null;
  },
};

vi.mock("../../wailsjs/runtime/runtime", () => ({
  EventsOn: (event: string, cb: Handler) => runtimeMock.on(event, cb),
  EventsOff: (event: string) => {
    listeners.delete(event);
  },
  EventsEmit: () => {},
  OnFileDrop: (cb: DropHandler) => {
    runtimeMock.fileDrop = cb;
  },
  OnFileDropOff: () => {
    runtimeMock.fileDrop = null;
  },
}));
```

`frontend/src/hooks/useChat.ts`:

1. Import tipe: `import type { AttachmentMeta, ChatEvent, ChatMessage, ChatStatus } from "../types/chat";`
2. Di `loadConversation`, tipe `rows` tambah `attachments?: AttachmentMeta[]` dan map tambah `attachments: r.attachments ?? [],`.
3. Ganti `send`:

```ts
  const send = useCallback(
    async (text: string, attachments: AttachmentMeta[] = []) => {
      const trimmed = text.trim();
      if ((!trimmed && attachments.length === 0) || sendingRef.current || activeRequest.current) return;
      let conv = convRef.current;
      if (!conv) {
        conv = freshId();
        convRef.current = conv;
        setConversationId(conv);
      }
      sendingRef.current = true;
      setError(null);
      setStatus("sending");
      setMessages((prev) => [
        ...prev,
        { id: `user-${Date.now()}`, role: "user", content: trimmed, status: "complete", attachments },
      ]);
      const ids = attachments.map((a) => a.id);
      await begin(conv, trimmed, (c, t) => SendMessage(c, t, ids));
    },
    [begin]
  );
```

- [ ] **Step 4: Jalankan, pastikan lulus**

Run: `cd frontend && npx tsc --noEmit && npx vitest run`
Expected: tsc bersih; semua tes PASS.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/types/chat.ts frontend/src/test/wails.ts frontend/src/hooks/useChat.ts frontend/src/hooks/useChat.test.ts
git commit -m "feat(chat): send attachment ids and carry metadata in history"
```

---

### Task 11: frontend — `useAttachments` + `formatSize`

**Files:**
- Create: `frontend/src/hooks/useAttachments.ts`
- Create: `frontend/src/lib/formatSize.ts`
- Test: `frontend/src/hooks/useAttachments.test.ts`, `frontend/src/lib/formatSize.test.ts`

**Interfaces:**
- Consumes: binding `PickAttachments`, `StageAttachmentPaths`, `StageAttachmentData`, `UnstageAttachment`.
- Produces: `MAX_ATTACHMENTS = 4`; `interface AttachmentsApi { pending; error; pick(); stagePaths(paths); stageFiles(files); remove(id); reset(); clearError() }`; `useAttachments(): AttachmentsApi`; `formatSize(bytes: number): string`.

- [ ] **Step 1: Tulis tes yang gagal**

`frontend/src/lib/formatSize.test.ts`:

```ts
import { describe, expect, it } from "vitest";
import { formatSize } from "./formatSize";

describe("formatSize", () => {
  it("picks the unit by size", () => {
    expect(formatSize(512)).toBe("512 B");
    expect(formatSize(2048)).toBe("2.0 KB");
    expect(formatSize(1572864)).toBe("1.5 MB");
  });
});
```

`frontend/src/hooks/useAttachments.test.ts`:

```ts
import { describe, expect, it, vi, beforeEach } from "vitest";
import { act, renderHook } from "@testing-library/react";
import type { AttachmentMeta } from "../types/chat";

function meta(id: string): AttachmentMeta {
  return { id, name: id + ".png", mime: "image/png", size: 3, kind: "image" };
}

const backend = {
  pick: vi.fn(async (): Promise<AttachmentMeta[]> => []),
  paths: vi.fn(async (_p: string[]): Promise<AttachmentMeta[]> => []),
  data: vi.fn(async (_n: string, _m: string, _d: string): Promise<AttachmentMeta> => meta("x")),
  unstage: vi.fn(async (_id: string) => {}),
};

vi.mock("../../wailsjs/go/main/App", () => ({
  PickAttachments: () => backend.pick(),
  StageAttachmentPaths: (p: string[]) => backend.paths(p),
  StageAttachmentData: (n: string, m: string, d: string) => backend.data(n, m, d),
  UnstageAttachment: (id: string) => backend.unstage(id),
}));

import { useAttachments, MAX_ATTACHMENTS } from "./useAttachments";

beforeEach(() => {
  vi.clearAllMocks();
  backend.pick.mockResolvedValue([]);
  backend.paths.mockResolvedValue([]);
  backend.unstage.mockResolvedValue(undefined);
});

describe("useAttachments", () => {
  it("adds what the dialog returns", async () => {
    backend.pick.mockResolvedValueOnce([meta("a"), meta("b")]);
    const { result } = renderHook(() => useAttachments());
    await act(() => result.current.pick());
    expect(result.current.pending.map((a) => a.id)).toEqual(["a", "b"]);
    expect(result.current.error).toBeNull();
  });

  it("keeps the list and reports the reason when the backend refuses", async () => {
    backend.paths.mockRejectedValueOnce("foto.exe: tipe berkas tidak didukung");
    const { result } = renderHook(() => useAttachments());
    await act(() => result.current.stagePaths(["C:\\x\\foto.exe"]));
    expect(result.current.pending).toEqual([]);
    expect(result.current.error).toBe("foto.exe: tipe berkas tidak didukung");
  });

  it("refuses a fifth attachment locally", async () => {
    backend.pick.mockResolvedValueOnce([meta("a"), meta("b"), meta("c"), meta("d")]);
    const { result } = renderHook(() => useAttachments());
    await act(() => result.current.pick());
    expect(result.current.pending).toHaveLength(MAX_ATTACHMENTS);
    await act(() => result.current.stagePaths(["x.png"]));
    expect(backend.paths).not.toHaveBeenCalled();
    expect(result.current.error).toContain("Maksimal 4");
    expect(result.current.pending).toHaveLength(MAX_ATTACHMENTS);
  });

  it("reads pasted files as base64 and stages each", async () => {
    backend.data.mockResolvedValueOnce(meta("p1"));
    const file = new File([new Uint8Array([1, 2, 3])], "tempel.png", { type: "image/png" });
    const { result } = renderHook(() => useAttachments());
    await act(() => result.current.stageFiles([file]));
    expect(backend.data).toHaveBeenCalledWith("tempel.png", "image/png", "AQID");
    expect(result.current.pending[0].id).toBe("p1");
  });

  it("unstages already-staged files when a later paste fails", async () => {
    backend.data.mockResolvedValueOnce(meta("p1")).mockRejectedValueOnce("rusak");
    const f = (n: string) => new File([new Uint8Array([1])], n, { type: "image/png" });
    const { result } = renderHook(() => useAttachments());
    await act(() => result.current.stageFiles([f("a.png"), f("b.png")]));
    expect(result.current.pending).toEqual([]);
    expect(backend.unstage).toHaveBeenCalledWith("p1");
    expect(result.current.error).toBe("rusak");
  });

  it("remove drops the item and tells the backend", async () => {
    backend.pick.mockResolvedValueOnce([meta("a"), meta("b")]);
    const { result } = renderHook(() => useAttachments());
    await act(() => result.current.pick());
    await act(() => result.current.remove("a"));
    expect(result.current.pending.map((a) => a.id)).toEqual(["b"]);
    expect(backend.unstage).toHaveBeenCalledWith("a");
  });

  it("reset clears the list without touching the backend", async () => {
    backend.pick.mockResolvedValueOnce([meta("a")]);
    const { result } = renderHook(() => useAttachments());
    await act(() => result.current.pick());
    act(() => result.current.reset());
    expect(result.current.pending).toEqual([]);
    expect(backend.unstage).not.toHaveBeenCalled();
  });
});
```

- [ ] **Step 2: Jalankan, pastikan gagal**

Run: `cd frontend && npx vitest run src/hooks/useAttachments.test.ts src/lib/formatSize.test.ts`
Expected: FAIL, modul tidak ditemukan.

- [ ] **Step 3: Implementasi**

`frontend/src/lib/formatSize.ts`:

```ts
// formatSize renders a byte count the way the pending row and transcript
// chips show it: whole bytes under 1 KB, else one decimal.
export function formatSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}
```

`frontend/src/hooks/useAttachments.ts`:

```ts
import { useCallback, useRef, useState } from "react";
import {
  PickAttachments,
  StageAttachmentData,
  StageAttachmentPaths,
  UnstageAttachment,
} from "../../wailsjs/go/main/App";
import type { AttachmentMeta } from "../types/chat";

export const MAX_ATTACHMENTS = 4;

export interface AttachmentsApi {
  pending: AttachmentMeta[];
  error: string | null;
  pick: () => Promise<void>;
  stagePaths: (paths: string[]) => Promise<void>;
  stageFiles: (files: File[]) => Promise<void>;
  remove: (id: string) => Promise<void>;
  /** Forget the list locally; the backend moved the items to history on send. */
  reset: () => void;
  clearError: () => void;
}

// Wails rejects with the Go error string; tests reject with Error.
function messageOf(e: unknown, fallback: string): string {
  if (typeof e === "string" && e) return e;
  if (e instanceof Error && e.message) return e.message;
  return fallback;
}

function toBase64(file: File): Promise<string> {
  return new Promise((resolve, reject) => {
    const r = new FileReader();
    r.onerror = () => reject(new Error("Berkas gagal dibaca."));
    r.onload = () => {
      const s = String(r.result);
      resolve(s.slice(s.indexOf(",") + 1));
    };
    r.readAsDataURL(file);
  });
}

// useAttachments owns the composer's pending attachments. Bytes live in the
// Go staging area; this hook holds metadata only. Dialog, drop, and paste all
// end in the same list. A failed binding leaves the list unchanged and
// reports why; the per-message cap is checked here first so a fifth file
// never round-trips to the backend.
export function useAttachments(): AttachmentsApi {
  const [pending, setPending] = useState<AttachmentMeta[]>([]);
  const [error, setError] = useState<string | null>(null);
  const pendingRef = useRef(pending);
  pendingRef.current = pending;

  const stage = useCallback(
    async (count: number, run: () => Promise<AttachmentMeta[] | null>, fallback: string) => {
      if (count === 0) return;
      if (pendingRef.current.length + count > MAX_ATTACHMENTS) {
        setError(`Maksimal ${MAX_ATTACHMENTS} lampiran per pesan.`);
        return;
      }
      setError(null);
      try {
        const added = (await run()) ?? [];
        if (added.length > 0) setPending((prev) => [...prev, ...added]);
      } catch (e) {
        setError(messageOf(e, fallback));
      }
    },
    []
  );

  // The generated bindings type `kind` as string; the backend only ever
  // sends image|pdf|text, so narrow at the boundary.
  const asMetas = (p: Promise<unknown>) => p as Promise<AttachmentMeta[] | null>;

  // The dialog's count is unknown until it closes; 1 checks there is room at
  // all, the backend enforces the cap on the real batch.
  const pick = useCallback(() => stage(1, () => asMetas(PickAttachments()), "Dialog berkas gagal dibuka."), [stage]);

  const stagePaths = useCallback(
    (paths: string[]) => stage(paths.length, () => asMetas(StageAttachmentPaths(paths)), "Berkas gagal dilampirkan."),
    [stage]
  );

  const stageFiles = useCallback(
    (files: File[]) =>
      stage(
        files.length,
        async () => {
          const out: AttachmentMeta[] = [];
          try {
            for (const f of files) {
              out.push((await StageAttachmentData(f.name, f.type, await toBase64(f))) as AttachmentMeta);
            }
          } catch (e) {
            // Keep backend staging in step with the list the user sees.
            for (const m of out) void UnstageAttachment(m.id).catch(() => {});
            throw e;
          }
          return out;
        },
        "Gambar gagal ditempel."
      ),
    [stage]
  );

  const remove = useCallback(async (id: string) => {
    setPending((prev) => prev.filter((a) => a.id !== id));
    try {
      await UnstageAttachment(id);
    } catch {
      // The list is what the user sees; the backend drops leftovers at exit.
    }
  }, []);

  const reset = useCallback(() => setPending([]), []);
  const clearError = useCallback(() => setError(null), []);

  return { pending, error, pick, stagePaths, stageFiles, remove, reset, clearError };
}
```

- [ ] **Step 4: Jalankan, pastikan lulus**

Run: `cd frontend && npx tsc --noEmit && npx vitest run`
Expected: PASS semua.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/hooks/useAttachments.ts frontend/src/hooks/useAttachments.test.ts frontend/src/lib/formatSize.ts frontend/src/lib/formatSize.test.ts
git commit -m "feat(chat): useAttachments hook for pending files"
```

---

### Task 12: frontend — composer: tombol klip, baris pending, paste, drop zone

**Files:**
- Modify: `frontend/src/components/ui/Icon.tsx` (tambah `IconPaperclip`, `IconFile`)
- Modify: `frontend/src/components/chat/ChatPanel.tsx`
- Modify: `frontend/src/styles/chat.css`
- Test: `frontend/src/components/chat/ChatPanel.attach.test.tsx`

**Interfaces:**
- Consumes: `AttachmentsApi` (Task 11), `OnFileDrop/OnFileDropOff` (wailsjs runtime), `formatSize`.
- Produces: prop `attach: AttachmentsApi`; `onSend: (text: string, attachments: AttachmentMeta[]) => void`; elemen `#btn-attach`, `#chat-pending`, `#chat-drop`, `#attach-error`.

- [ ] **Step 1: Tulis tes yang gagal**

`frontend/src/components/chat/ChatPanel.attach.test.tsx`:

```tsx
import { describe, expect, it, vi, beforeEach, afterEach } from "vitest";
import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { runtimeMock } from "../../test/wails";
import type { AttachmentMeta } from "../../types/chat";
import type { AttachmentsApi } from "../../hooks/useAttachments";
import { ChatPanel } from "./ChatPanel";

function meta(id: string, kind: AttachmentMeta["kind"] = "image"): AttachmentMeta {
  return { id, name: `${id}.${kind === "image" ? "png" : "pdf"}`, mime: kind === "image" ? "image/png" : "application/pdf", size: 2048, kind };
}

function api(over: Partial<AttachmentsApi> = {}): AttachmentsApi {
  return {
    pending: [],
    error: null,
    pick: vi.fn(async () => {}),
    stagePaths: vi.fn(async () => {}),
    stageFiles: vi.fn(async () => {}),
    remove: vi.fn(async () => {}),
    reset: vi.fn(),
    clearError: vi.fn(),
    ...over,
  };
}

function renderPanel(over: { attach?: AttachmentsApi; draft?: string; busy?: boolean } = {}) {
  const onSend = vi.fn();
  const attach = over.attach ?? api();
  const utils = render(
    <ChatPanel
      messages={[]}
      status="idle"
      error={null}
      busy={over.busy ?? false}
      conversations={[]}
      conversationId="c1"
      onSend={onSend}
      onCancel={() => {}}
      onRetry={() => {}}
      onDismissError={() => {}}
      onNew={() => {}}
      onSwitch={() => {}}
      onRename={() => {}}
      onDelete={() => {}}
      draft={over.draft ?? ""}
      onDraftChange={() => {}}
      modelPicker={null}
      onOpenSettings={() => {}}
      hasProvider
      attach={attach}
    />
  );
  return { ...utils, onSend, attach };
}

beforeEach(() => runtimeMock.reset());
afterEach(() => cleanup());

describe("ChatPanel attachments", () => {
  it("has a clip button that opens the picker", () => {
    const { attach } = renderPanel();
    fireEvent.click(screen.getByRole("button", { name: "Lampirkan berkas" }));
    expect(attach.pick).toHaveBeenCalled();
  });

  it("disables the clip button while busy", () => {
    renderPanel({ busy: true });
    expect((screen.getByRole("button", { name: "Lampirkan berkas" }) as HTMLButtonElement).disabled).toBe(true);
  });

  it("shows pending items as thumbnail or chip with a remove button", () => {
    const attach = api({ pending: [meta("a"), meta("b", "pdf")] });
    renderPanel({ attach });
    const img = screen.getByRole("img", { name: "a.png" }) as HTMLImageElement;
    expect(img.getAttribute("src")).toBe("/attachments/a");
    expect(screen.getByText("b.pdf")).toBeTruthy();
    expect(screen.getByText("2.0 KB")).toBeTruthy();
    fireEvent.click(screen.getByRole("button", { name: "Hapus lampiran a.png" }));
    expect(attach.remove).toHaveBeenCalledWith("a");
  });

  it("sends with attachments only, then resets draft and pending", () => {
    const attach = api({ pending: [meta("a")] });
    const { onSend } = renderPanel({ attach });
    const send = screen.getByRole("button", { name: "Kirim pesan" }) as HTMLButtonElement;
    expect(send.disabled).toBe(false);
    fireEvent.click(send);
    expect(onSend).toHaveBeenCalledWith("", attach.pending);
    expect(attach.reset).toHaveBeenCalled();
  });

  it("keeps send disabled with neither text nor attachments", () => {
    renderPanel();
    expect((screen.getByRole("button", { name: "Kirim pesan" }) as HTMLButtonElement).disabled).toBe(true);
  });

  it("stages pasted images instead of inserting them", () => {
    const { attach } = renderPanel();
    const file = new File([new Uint8Array([1])], "shot.png", { type: "image/png" });
    fireEvent.paste(screen.getByRole("textbox", { name: "Tulis pesan" }), {
      clipboardData: { files: [file], getData: () => "" },
    });
    expect(attach.stageFiles).toHaveBeenCalledWith([file]);
  });

  it("leaves a plain text paste alone", () => {
    const { attach } = renderPanel();
    fireEvent.paste(screen.getByRole("textbox", { name: "Tulis pesan" }), {
      clipboardData: { files: [], getData: () => "halo" },
    });
    expect(attach.stageFiles).not.toHaveBeenCalled();
  });

  it("listens for dropped files only while mounted", () => {
    const { attach, unmount } = renderPanel();
    expect(runtimeMock.fileDrop).not.toBeNull();
    runtimeMock.fileDrop?.(0, 0, ["C:\\x\\a.png"]);
    expect(attach.stagePaths).toHaveBeenCalledWith(["C:\\x\\a.png"]);
    unmount();
    expect(runtimeMock.fileDrop).toBeNull();
  });

  it("shows the attachment error with a close button", () => {
    const attach = api({ error: "foto.exe: tipe berkas tidak didukung" });
    renderPanel({ attach });
    expect(screen.getByRole("alert").textContent).toContain("foto.exe");
    fireEvent.click(screen.getByRole("button", { name: "Tutup pesan lampiran" }));
    expect(attach.clearError).toHaveBeenCalled();
  });
});
```

- [ ] **Step 2: Jalankan, pastikan gagal**

Run: `cd frontend && npx vitest run src/components/chat/ChatPanel.attach.test.tsx`
Expected: FAIL (prop `attach` tidak dikenal; tombol tidak ada).

- [ ] **Step 3: Implementasi**

`frontend/src/components/ui/Icon.tsx` — tambahkan di akhir:

```tsx
export const IconPaperclip = (p: IconProps) => (
  <Svg {...p}>
    <path d="m21.44 11.05-9.19 9.19a6 6 0 0 1-8.49-8.49l9.19-9.19a4 4 0 0 1 5.66 5.66l-9.2 9.19a2 2 0 0 1-2.83-2.83l8.49-8.48" />
  </Svg>
);

export const IconFile = (p: IconProps) => (
  <Svg {...p}>
    <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" />
    <path d="M14 2v6h6" />
    <path d="M16 13H8" />
    <path d="M16 17H8" />
  </Svg>
);
```

`frontend/src/components/chat/ChatPanel.tsx`:

1. Import:

```tsx
import type { ClipboardEvent, DragEvent, JSX, KeyboardEvent, ReactNode } from "react";
import { OnFileDrop, OnFileDropOff } from "../../../wailsjs/runtime/runtime";
import type { AttachmentMeta, ChatMessage, ChatStatus } from "../../types/chat";
import type { AttachmentsApi } from "../../hooks/useAttachments";
import { formatSize } from "../../lib/formatSize";
```

dan tambahkan `IconFile`, `IconPaperclip` ke import ikon.

2. Props: ganti `onSend: (text: string) => void;` → `onSend: (text: string, attachments: AttachmentMeta[]) => void;` dan tambah `/** Pending attachments; owned by App so they survive pet mode. */ attach: AttachmentsApi;`. Destructure `attach` di parameter.

3. Di badan komponen, setelah `composerRef`:

```tsx
  // Drop zone depth: dragenter/dragleave fire for every child, so a counter
  // keeps the overlay steady until the pointer really leaves the panel.
  const [dragDepth, setDragDepth] = useState(0);
  const stageRef = useRef(attach.stagePaths);
  stageRef.current = attach.stagePaths;

  // Dropped files arrive as paths from the Wails runtime, only while the
  // chat panel is mounted: the pet and Settings never take a drop.
  useEffect(() => {
    OnFileDrop((_x, _y, paths) => {
      setDragDepth(0);
      void stageRef.current(paths);
    }, true);
    return () => OnFileDropOff();
  }, []);

  const hasFiles = (e: DragEvent<HTMLDivElement>) => Array.from(e.dataTransfer?.types ?? []).includes("Files");
  const onDragEnter = (e: DragEvent<HTMLDivElement>) => {
    if (hasFiles(e)) setDragDepth((d) => d + 1);
  };
  const onDragLeave = (e: DragEvent<HTMLDivElement>) => {
    if (hasFiles(e)) setDragDepth((d) => Math.max(0, d - 1));
  };
  const onDrop = () => setDragDepth(0);

  // Pasted images go to staging; pasted text keeps the browser default.
  const onPaste = (e: ClipboardEvent<HTMLTextAreaElement>) => {
    const files = Array.from(e.clipboardData?.files ?? []).filter((f) => f.type.startsWith("image/"));
    if (files.length === 0) return;
    e.preventDefault();
    void attach.stageFiles(files);
  };
```

4. Ganti `submit`:

```tsx
  const hasContent = draft.trim() !== "" || attach.pending.length > 0;
  const submit = () => {
    if (!hasContent || busy || !hasProvider) return;
    onSend(draft, attach.pending);
    onDraftChange("");
    attach.reset();
  };
```

5. Root: `<div id="chat-main" onDragEnter={onDragEnter} onDragLeave={onDragLeave} onDrop={onDrop}>` dan tepat di dalamnya:

```tsx
      {dragDepth > 0 && (
        <div id="chat-drop" role="status" aria-live="polite">
          <IconPaperclip size={20} />
          <span>Lepas untuk melampirkan</span>
        </div>
      )}
```

6. Di empty state `hasProvider`, setelah tombol contoh: `<p className="chat-empty__tip">Tempel atau seret gambar/PDF ke sini.</p>`

7. Sebelum `<div id="chat-composer">`:

```tsx
      {attach.error && (
        <div id="attach-error" className="banner banner--error" role="alert">
          <IconAlert size={15} />
          <span className="banner__text">{attach.error}</span>
          <button type="button" className="icon-btn" onClick={attach.clearError} aria-label="Tutup pesan lampiran" title="Tutup">
            <IconX size={14} />
          </button>
        </div>
      )}
      {attach.pending.length > 0 && (
        <div id="chat-pending" role="list" aria-label="Lampiran menunggu">
          {attach.pending.map((a) => (
            <div key={a.id} role="listitem" className="pending-item">
              {a.kind === "image" ? (
                <img className="pending-item__thumb" src={`/attachments/${a.id}`} alt={a.name} />
              ) : (
                <span className="att-chip">
                  <IconFile size={14} />
                  <span className="att-chip__name">{a.name}</span>
                  <span className="att-chip__size">{formatSize(a.size)}</span>
                </span>
              )}
              <button
                type="button"
                className="icon-btn pending-item__remove"
                onClick={() => void attach.remove(a.id)}
                aria-label={`Hapus lampiran ${a.name}`}
                title="Hapus lampiran"
                disabled={busy}
              >
                <IconX size={12} />
              </button>
            </div>
          ))}
        </div>
      )}
```

8. Di `#chat-composer`, sebelum `<textarea>`:

```tsx
        <button
          type="button"
          id="btn-attach"
          className="btn btn--icon"
          onClick={() => void attach.pick()}
          disabled={busy || !hasProvider}
          aria-label="Lampirkan berkas"
          title="Lampirkan gambar, PDF, atau teks"
        >
          <IconPaperclip size={15} />
        </button>
```

`<textarea>` tambah `onPaste={onPaste}`. Tombol kirim: `disabled={busy ? false : !hasProvider || !hasContent}`.

`frontend/src/styles/chat.css` — tambahkan di akhir:

```css
/* ---- attachments: drop zone, pending row, composer clip, chips ----
   The custom property marks the whole panel (children inherit it) as the
   Wails file-drop target; the overlay never takes pointer events so the
   drop still lands on the panel. */
#chat-main {
    --wails-drop-target: drop;
    position: relative;
}

#chat-drop {
    position: absolute;
    inset: var(--space-2);
    z-index: 2;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: var(--space-2);
    border: 2px dashed var(--color-ink-3);
    border-radius: var(--radius-lg);
    background: var(--color-paper);
    color: var(--color-ink-2);
    font-size: var(--text-sm);
    pointer-events: none;
}

#chat-pending {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-2);
    flex-shrink: 0;
    padding-top: var(--space-1);
}

.pending-item {
    position: relative;
    display: inline-flex;
    align-items: center;
    max-width: 100%;
}

.pending-item__thumb {
    display: block;
    width: 56px;
    height: 56px;
    object-fit: cover;
    border: 1px solid var(--color-rule-strong);
    border-radius: var(--radius-sm);
}

.pending-item__remove {
    position: absolute;
    top: -6px;
    right: -6px;
    width: 20px;
    height: 20px;
    background: var(--color-paper-3);
    border: 1px solid var(--color-rule-strong);
}

.att-chip {
    display: inline-flex;
    align-items: center;
    gap: var(--space-1);
    height: 28px;
    max-width: 100%;
    padding: 0 var(--space-3) 0 var(--space-2);
    border: 1px solid var(--color-rule-strong);
    border-radius: var(--radius-pill);
    background: var(--color-paper-2);
    color: var(--color-ink-2);
    font-size: var(--text-sm);
}

.pending-item .att-chip {
    padding-right: var(--space-5);
}

.att-chip__name {
    max-width: 18ch;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.att-chip__size {
    color: var(--color-ink-3);
}

#btn-attach {
    width: 32px;
    height: 32px;
    flex-shrink: 0;
}

#btn-attach:focus-visible,
.pending-item__remove:focus-visible {
    outline: var(--focus-ring);
    outline-offset: var(--focus-offset);
}
```

- [ ] **Step 4: Jalankan, pastikan lulus**

Run: `cd frontend && npx tsc --noEmit && npx vitest run`
Expected: tsc gagal hanya di `App.tsx` (prop `attach` wajib) — itu Task 14; vitest PASS semua.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/components/ui/Icon.tsx frontend/src/components/chat/ChatPanel.tsx frontend/src/components/chat/ChatPanel.attach.test.tsx frontend/src/styles/chat.css
git commit -m "feat(chat): attach button, pending row, paste and drop in the composer"
```

---

### Task 13: frontend — `AttachmentList` di transkrip

**Files:**
- Create: `frontend/src/components/chat/AttachmentList.tsx`
- Modify: `frontend/src/components/chat/MessageItem.tsx`
- Modify: `frontend/src/styles/chat.css`
- Test: `frontend/src/components/chat/MessageItem.attach.test.tsx`

**Interfaces:**
- Consumes: `ChatMessage.attachments`, `formatSize`, `IconFile`, `.att-chip`.
- Produces: `<AttachmentList items={AttachmentMeta[]} />`.

- [ ] **Step 1: Tulis tes yang gagal**

`frontend/src/components/chat/MessageItem.attach.test.tsx`:

```tsx
import { describe, expect, it, afterEach } from "vitest";
import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import type { ChatMessage } from "../../types/chat";
import { MessageItem } from "./MessageItem";

const img = { id: "a", name: "foto.png", mime: "image/png", size: 100, kind: "image" as const };
const pdf = { id: "b", name: "nota.pdf", mime: "application/pdf", size: 3072, kind: "pdf" as const };

function user(over: Partial<ChatMessage> = {}): ChatMessage {
  return { id: "u1", role: "user", content: "lihat", status: "complete", ...over };
}

afterEach(() => cleanup());

describe("MessageItem attachments", () => {
  it("renders image thumbnails and file chips for a user turn", () => {
    render(<MessageItem message={user({ attachments: [img, pdf] })} isLast busy={false} onRetry={() => {}} />);
    const el = screen.getByRole("img", { name: "foto.png" });
    expect(el.getAttribute("src")).toBe("/attachments/a");
    expect(screen.getByText("nota.pdf")).toBeTruthy();
    expect(screen.getByText("3.0 KB")).toBeTruthy();
    expect(screen.getByRole("group", { name: "Lampiran: 2 berkas" })).toBeTruthy();
  });

  it("toggles a thumbnail to full size", () => {
    render(<MessageItem message={user({ attachments: [img] })} isLast busy={false} onRetry={() => {}} />);
    const btn = screen.getByRole("button", { name: "foto.png" });
    expect(btn.getAttribute("aria-pressed")).toBe("false");
    fireEvent.click(btn);
    expect(btn.getAttribute("aria-pressed")).toBe("true");
  });

  it("renders nothing extra without attachments", () => {
    const { container } = render(<MessageItem message={user()} isLast busy={false} onRetry={() => {}} />);
    expect(container.querySelector(".att-list")).toBeNull();
  });

  it("hides the empty bubble when only attachments were sent", () => {
    const { container } = render(<MessageItem message={user({ content: "", attachments: [img] })} isLast busy={false} onRetry={() => {}} />);
    expect(container.querySelector(".msg-body")).toBeNull();
    expect(container.querySelector(".att-list")).not.toBeNull();
  });
});
```

- [ ] **Step 2: Jalankan, pastikan gagal**

Run: `cd frontend && npx vitest run src/components/chat/MessageItem.attach.test.tsx`
Expected: FAIL (tidak ada `img`/`group`).

- [ ] **Step 3: Implementasi**

`frontend/src/components/chat/AttachmentList.tsx`:

```tsx
// AttachmentList shows a sent message's attachments: image thumbnails that
// toggle to full width on click, and name+size chips for files. Images come
// from the local asset route, never from provider output.
import { useState } from "react";
import type { JSX } from "react";
import type { AttachmentMeta } from "../../types/chat";
import { IconFile } from "../ui/Icon";
import { formatSize } from "../../lib/formatSize";

export function AttachmentList({ items }: { items: AttachmentMeta[] }): JSX.Element {
  const [open, setOpen] = useState<string | null>(null);
  return (
    <div className="att-list" role="group" aria-label={`Lampiran: ${items.length} berkas`}>
      {items.map((a) =>
        a.kind === "image" ? (
          <button
            key={a.id}
            type="button"
            className={open === a.id ? "att-thumb att-thumb--open" : "att-thumb"}
            onClick={() => setOpen(open === a.id ? null : a.id)}
            aria-pressed={open === a.id}
            aria-label={a.name}
            title={open === a.id ? "Perkecil" : "Perbesar"}
          >
            <img src={`/attachments/${a.id}`} alt={a.name} />
          </button>
        ) : (
          <span key={a.id} className="att-chip">
            <IconFile size={14} />
            <span className="att-chip__name">{a.name}</span>
            <span className="att-chip__size">{formatSize(a.size)}</span>
          </span>
        )
      )}
    </div>
  );
}
```

`frontend/src/components/chat/MessageItem.tsx`:

1. `import { AttachmentList } from "./AttachmentList";`
2. Setelah `const waiting = …`:

```tsx
  const atts = m.role === "user" ? m.attachments ?? [] : [];
  // No bubble for a turn that was only attachments; the list stands alone.
  const showBody = !(failed && !m.content) && !(m.role === "user" && !m.content);
```

(hapus baris `const showBody = !(failed && !m.content);` yang lama.)

3. Di dalam `<div className="msg-main">`, sebelum `{showBody && (`: `{atts.length > 0 && <AttachmentList items={atts} />}`

`frontend/src/styles/chat.css` — tambahkan:

```css
/* ---- attachments in the transcript ---- */
.att-list {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-2);
    max-width: 100%;
    margin-bottom: var(--space-1);
}

.msg-user .att-list {
    justify-content: flex-end;
}

.att-thumb {
    padding: 0;
    max-width: 160px;
    overflow: hidden;
    border: 1px solid var(--color-rule-strong);
    border-radius: var(--radius-sm);
    background: none;
    cursor: zoom-in;
}

.att-thumb img {
    display: block;
    max-width: 160px;
    max-height: 160px;
    object-fit: cover;
}

.att-thumb--open {
    max-width: 100%;
    cursor: zoom-out;
}

.att-thumb--open img {
    max-width: 100%;
    max-height: none;
    object-fit: contain;
}

.att-thumb:focus-visible {
    outline: var(--focus-ring);
    outline-offset: var(--focus-offset);
}
```

- [ ] **Step 4: Jalankan, pastikan lulus**

Run: `cd frontend && npx vitest run`
Expected: PASS semua.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/components/chat/AttachmentList.tsx frontend/src/components/chat/MessageItem.tsx frontend/src/components/chat/MessageItem.attach.test.tsx frontend/src/styles/chat.css
git commit -m "feat(chat): show attachments on sent messages"
```

---

### Task 14: frontend — wiring `App.tsx`, guard drop global, build

**Files:**
- Modify: `frontend/src/App.tsx`
- Modify: `frontend/src/main.tsx`

- [ ] **Step 1: Wiring**

`frontend/src/App.tsx`: `import {useAttachments} from './hooks/useAttachments';`; setelah `const [draft, setDraft] = useState<string>("");` tambah:

```tsx
    // Pending attachments live next to the draft for the same reason: pet
    // mode unmounts the panel and the list must survive it.
    const attach = useAttachments();
```

Pada `<ChatPanel … />` tambah prop `attach={attach}` (prop `onSend={chat.send}` tetap; tipenya kini `(text, attachments)`).

`frontend/src/main.tsx`: sebelum pemanggilan `createRoot(...)`:

```ts
// A file dropped anywhere outside the chat drop zone (pet, Settings, panel
// header) must not make WebView2 navigate to it. The Wails runtime has its
// own drop listener and decides by the target's CSS whether to report the
// drop; cancelling the default here does not interfere with that.
window.addEventListener("dragover", (e) => e.preventDefault());
window.addEventListener("drop", (e) => e.preventDefault());
```

- [ ] **Step 2: Verifikasi frontend penuh**

Run: `cd frontend && npx tsc --noEmit && npm test && npm run build`
Expected: tsc bersih; semua tes PASS; `vite build` sukses ke `frontend/dist`.

- [ ] **Step 3: Verifikasi backend penuh**

Run: `go build ./... && go vet ./... && gofmt -l app.go main.go internal/ && go test ./...`
Expected: semua bersih/PASS.

- [ ] **Step 4: Commit**

```bash
git add frontend/src/App.tsx frontend/src/main.tsx
git commit -m "feat(app): wire attachments into the chat surface"
```

---

### Task 15: Dokumentasi + ADR-054

**Files:**
- Modify: `docs/02-mvp-requirements.md`, `docs/05-ai-provider-integration.md`, `docs/07-storage-security-privacy.md`, `docs/09-decisions.md`, `docs/10-frontend-ui-inventory.md`, `README.md`

- [ ] **Step 1: `docs/02-mvp-requirements.md`** — setelah FR-011 tambahkan:

```markdown
### FR-012 — Lampiran gambar dan berkas (P1)

Pengguna dapat melampirkan gambar (png/jpeg/webp), PDF, dan berkas teks (txt/md/csv/json) pada pesan, lewat tombol klip, Ctrl+V gambar, atau drag-drop ke panel chat (ditambahkan 2026-10-11; lihat ADR-054).

**Kriteria penerimaan**
- Maksimal 4 lampiran per pesan; gambar ≤ 8 MB, PDF ≤ 15 MB, teks ≤ 256 KB, total ≤ 20 MB; tipe ditentukan dari isi berkas, bukan ekstensi.
- Pesan boleh tanpa teks bila ada lampiran; tanpa keduanya tidak dikirim.
- Lampiran tersimpan di riwayat lokal dan tampil kembali saat percakapan dibuka; hapus percakapan menghapus lampirannya.
- Hanya 3 pesan user terakhir yang mengirim lampiran penuh ke model; yang lebih lama diganti placeholder teks.
- Provider yang menolak gambar/berkas menghasilkan pesan "Model ini tidak menerima gambar/berkas", bukan error generik.
- Berkas hanya dibaca dari aksi eksplisit pengguna; path penuh dan isi berkas tidak masuk log.
```

- [ ] **Step 2: `docs/05-ai-provider-integration.md`**

Bagian 4, setelah contoh payload: 

```markdown
Pesan user dengan lampiran memakai `content` array (bentuk standar OpenAI): part `text` untuk teks pengguna, `image_url` dengan data URL base64 untuk gambar, `file` (`filename` + `file_data` data URL) untuk PDF, dan berkas teks disisipkan sebagai part `text` berpagar. Hanya 3 pesan user terakhir membawa part; yang lebih lama memakai `[gambar: nama]` / `[berkas: nama]` sehingga gambar tidak ditagih ulang tanpa batas (Kenari menagih gambar di history pada tiap request). Tanpa lampiran, `content` tetap string.
```

Bagian 8, tambah dua baris tabel:

```markdown
| 413 | Lampiran terlalu besar untuk provider; minta pengguna mengurangi ukuran/jumlah. |
| 400 yang menyebut image/file/vision/modal/document | Model tidak menerima gambar/berkas; sarankan model lain atau kirim teks saja. |
```

Bagian 9: pindahkan "Multimodal/image input" dan "Upload file" dari "Tidak wajib" ke "Dukung" dengan catatan "(gambar, PDF, teks; ADR-054)".

- [ ] **Step 3: `docs/07-storage-security-privacy.md`** — setelah tabel `messages`:

```markdown
### `attachments`

- `id` TEXT PRIMARY KEY (UUID).
- `message_id` TEXT NOT NULL FK → `messages(id)` ON DELETE CASCADE.
- `kind` TEXT (`image` | `pdf` | `text`), `mime` TEXT, `name` TEXT (basename saja, tanpa path), `size` INTEGER.
- `data` BLOB NOT NULL — isi berkas pengguna. Bukan secret, tetapi data pribadi: tidak pernah di-log.
- `position` INTEGER (urutan dalam pesan), `created_at`.

Indeks `attachments(message_id, position)`. Pembacaan metadata tidak memilih kolom `data`. Hapus percakapan membebaskan halaman untuk dipakai ulang; ukuran file DB tidak mengecil otomatis.
```

- [ ] **Step 4: `docs/09-decisions.md`** — baris tabel status:

```markdown
| ADR-054 | Lampiran gambar/PDF/teks: staging Go, BLOB SQLite, 3 turn, tanpa gating | Done 2026-10-11 | `content` array standar OpenAI; bytes di-stage di Go lalu disimpan BLOB bersama pesan; thumbnail via route `/attachments/{id}`; 413/400 modality dipetakan ke pesan khusus. |
```

dan entri lengkap sebelum "## Template entri keputusan baru":

```markdown
## ADR-054 — Lampiran gambar dan berkas (2026-10-11)

- **Status:** Done, perlu smoke manual Windows (dialog, drop, paste, model vision Kenari).
- **Konteks:** Provider yang dipakai (Kenari) meneruskan model multimodal; pengguna ingin bertanya tentang screenshot dan dokumen tanpa pindah ke browser. Spec: `docs/superpowers/specs/2026-10-11-attachments-design.md`.
- **Pilihan:** (A) staging bytes di Go + BLOB SQLite + route asset-server; (B) staging base64 di state React; (C) seperti A tetapi berkas di disk.
- **Keputusan:** (A). Dialog Wails hanya memberi path, jadi Go harus membaca; satu tabel `attachments` dengan cascade membuat kirim + simpan atomik tanpa sweep file; `<img src="/attachments/{id}">` menghindari base64 lewat IPC. Format request memakai `content` array standar OpenAI (`image_url` data URL, `file` untuk PDF, teks disisipkan sebagai part `text`) sehingga tidak ada parameter khusus provider. Hanya 3 turn user terakhir membawa lampiran (`AttachmentTurns`); lebih lama jadi placeholder teks karena gambar ditagih ulang tiap request. Tanpa gating kemampuan model: 400 yang menyebut image/file/vision/modal/document → `unsupported_input`, 413 → `payload_too_large`. Batas: 4 lampiran, gambar 8 MB, PDF 15 MB, teks 256 KB, total 20 MB; tipe dari magic bytes. Staging adalah isi pesan yang sedang disusun, batas ditegakkan saat menambah.
- **Perbaikan ikutan:** `RecentContext` tidak lagi mengirim baris assistant kosong (`streaming`) yang dulu ikut sebagai `{"role":"assistant","content":""}`.
- **Validasi:** test-first di `internal/ai`, `internal/attach`, `internal/storage`, `internal/app` (marshal parts, klasifikasi, staging, migrasi v5 + cascade + atomik, jendela 3 turn, retry membawa lampiran, route 404/405, error map) dan vitest (`useChat`, `useAttachments`, `ChatPanel`, `MessageItem`, `formatSize`).
- **Risiko:** screenshot PNG besar tanpa downscale → token mahal (downscale jadi kandidat lanjutan). `OnFileDrop` butuh `--wails-drop-target` di elemen dan belum diuji di build nyata sebelum smoke. Provider OpenAI-compatible lain mungkin tidak mengenal part `file`; dipetakan ke `unsupported_input`. Retry setelah penolakan sinkron `SendMessage` (bukan kegagalan provider) berjalan tanpa lampiran. Ukuran DB bertambah sebesar lampiran.
```

- [ ] **Step 5: `docs/10-frontend-ui-inventory.md`** — di tabel S2 tambah baris:

```markdown
| Baris lampiran | `#chat-pending`, `.pending-item` | ✔ | Thumbnail 56 px / chip nama+ukuran, tombol × per item (ADR-054). |
| Tombol lampir | `#btn-attach` | ✔ | Kiri textarea; nonaktif saat busy/tanpa provider. |
| Drop overlay | `#chat-drop` | ✔ | Hanya saat drag berkas di atas panel chat. |
| Lampiran di transkrip | `.att-list`, `.att-thumb`, `.att-chip` | ✔ | Klik thumbnail = perbesar inline, tanpa lightbox. |
```

- [ ] **Step 6: `README.md`** — di daftar fitur tambahkan: `- Lampiran gambar (png/jpeg/webp), PDF, dan berkas teks ke pesan: tombol klip, Ctrl+V, atau drag-drop. Maks 4 berkas / 20 MB per pesan; hanya 3 pesan terakhir yang mengirim ulang lampiran ke model.`

- [ ] **Step 7: Commit**

```bash
git add docs README.md
git commit -m "docs: attachments (FR-012, ADR-054, storage and provider notes)"
```

---

### Task 16: Build rilis + smoke manual

**Files:** tidak ada perubahan kode; hasil dicatat di `docs/09-decisions.md` ADR-054 (Status) bila ada temuan.

- [ ] **Step 1: Build**

Run (PowerShell): `wails build -ldflags "-X main.Version=0.1.0"`
Expected: `build/bin/companion.exe` terbentuk; tidak ada warning binding.

- [ ] **Step 2: Smoke manual (Kenari, model vision)** — catat hasil tiap poin:

1. Tombol klip → dialog filter "Gambar, PDF, teks" → pilih PNG → thumbnail muncul di baris pending; × menghapus.
2. Ctrl+V screenshot di composer → thumbnail; Ctrl+V teks → masuk draft seperti biasa.
3. Seret PNG dari Explorer ke panel chat → overlay "Lepas untuk melampirkan" → terlampir. Seret ke pet (mode pet) → tidak ada reaksi, WebView tidak berpindah ke file.
4. Kirim gambar + "apa ini?" → balasan menyebut isi gambar. Follow-up tanpa gambar ("warnanya apa?") masih nyambung.
5. Kirim PDF ber-text-layer → ringkasan benar. Coba PDF > 15 MB → ditolak lokal dengan nama berkas, tidak ada request.
6. Ganti ke model non-vision → kirim gambar → pesan "Model ini tidak menerima gambar/berkas…" di balasan; "Coba lagi" setelah ganti model vision berhasil dengan gambar yang sama.
7. Pilih 5 berkas di dialog → error "maksimal 4 lampiran per pesan", tidak ada yang terlampir.
8. Restart aplikasi → buka percakapan → thumbnail tampil dari DB; klik → perbesar; hapus percakapan → `SELECT COUNT(*) FROM attachments` (hanya SELECT) turun.
9. Mode pet lalu kembali ke chat → lampiran pending masih ada.
10. RAM idle dan ukuran exe dibandingkan dengan rilis sebelumnya; selisih dicatat.

- [ ] **Step 3: Catat hasil**

Ubah baris `**Status:**` ADR-054 menjadi `Done, smoke manual Windows lulus <tanggal>` atau catat temuan yang gagal sebagai risiko/bug terbuka.

```bash
git add docs/09-decisions.md
git commit -m "docs: ADR-054 smoke results"
```
