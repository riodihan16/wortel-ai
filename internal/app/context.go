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
// while follow-ups about a recent image still work. The whole window also
// shares one byte budget (attach.MaxMessageBytes), so a request stays under
// provider body limits: a turn that would overflow it, and every older
// turn, falls back to placeholders.
const AttachmentTurns = 3

// buildContext turns stored rows into provider messages, applying the
// AttachmentTurns window and its byte budget. The budget is decided from
// stored sizes, so no bytes are loaded for turns outside the window. load
// fetches bytes by attachment id; a load error degrades that attachment to
// its placeholder instead of failing the turn.
func buildContext(rows []storage.ContextRow, turns int, load func(id string) (string, []byte, error)) []ai.Message {
	full := map[string]bool{}
	seen := 0
	var used int64
	for i := len(rows) - 1; i >= 0; i-- {
		r := rows[i]
		if r.Role != "user" {
			continue
		}
		size := attachmentBytes(r.Attachments)
		if seen >= turns || used+size > attach.MaxMessageBytes {
			break
		}
		full[r.ID] = true
		seen++
		used += size
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

func attachmentBytes(atts []storage.AttachmentRow) int64 {
	var n int64
	for _, a := range atts {
		n += a.Size
	}
	return n
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
