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
