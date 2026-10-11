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
