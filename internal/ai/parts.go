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
