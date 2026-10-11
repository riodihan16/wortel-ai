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
// and the error names the file, never its directory. Count and total size
// are checked against what staging still has room for before any file is
// read, so a large selection is refused without loading it into memory.
// Empty paths (non-file items in a mixed drag) are skipped.
func (s *Service) StagePaths(paths []string) ([]attach.Meta, error) {
	if s.db() == nil {
		return nil, errNoAttachmentStorage
	}
	files := make([]string, 0, len(paths))
	for _, p := range paths {
		if p != "" {
			files = append(files, p)
		}
	}
	items, budget := s.staging.Remaining()
	if len(files) > items {
		return nil, attach.ErrTooMany
	}
	if err := checkSizes(files, budget); err != nil {
		return nil, err
	}
	type pending struct {
		meta attach.Meta
		data []byte
	}
	batch := make([]pending, 0, len(files))
	for _, p := range files {
		name := attach.CleanName(p)
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

// checkSizes stats every path without reading it: each must be a regular
// file within MaxFileBytes, and together they must fit in budget.
func checkSizes(files []string, budget int64) error {
	var total int64
	for _, p := range files {
		name := attach.CleanName(p)
		info, err := os.Stat(p)
		if err != nil {
			return fmt.Errorf("%s: tidak bisa dibaca", name)
		}
		if info.IsDir() {
			return fmt.Errorf("%s: folder tidak bisa dilampirkan", name)
		}
		if info.Size() > attach.MaxFileBytes {
			return fmt.Errorf("%s: melebihi %d MB", name, attach.MaxFileBytes>>20)
		}
		total += info.Size()
	}
	if total > budget {
		return attach.ErrTooBig
	}
	return nil
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
