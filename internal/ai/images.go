package ai

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// maxImageJSONBytes bounds the JSON envelope read from the provider so a
// misbehaving endpoint cannot stream an unbounded body into memory.
const maxImageJSONBytes = 64 << 20

// ImageRequest is one image-generation request. Size/N are model-dependent;
// callers pass only what the configured model accepts (Kenari docs §10.5).
type ImageRequest struct {
	Model  string
	Prompt string
	Size   string
	N      int
}

// normalizeImagesURL validates scheme/host and joins the images/generations
// path. Mirrors NormalizeBaseURL's guards; never doubles a trailing /v1.
func normalizeImagesURL(base string) (string, error) {
	u, err := url.Parse(strings.TrimSpace(base))
	if err != nil || u.Host == "" {
		return "", fmt.Errorf("invalid base url")
	}
	if u.Scheme != "https" && u.Scheme != "http" {
		return "", fmt.Errorf("base url must be http or https")
	}
	if u.User != nil {
		return "", fmt.Errorf("base url must not contain credentials")
	}
	p := strings.TrimSuffix(u.Path, "/")
	if !strings.HasSuffix(p, "/images/generations") {
		p += "/images/generations"
	}
	u.Path = p
	return u.String(), nil
}

// imageResponse is the OpenAI-style shape. A provider may answer with either
// b64_json or a url/data-url; both fields are read.
type imageResponse struct {
	Data []struct {
		B64JSON *string `json:"b64_json"`
		URL     *string `json:"url"`
	} `json:"data"`
}

// GenerateImage calls POST /images/generations and returns the decoded bytes
// with their sniffed mime. Only png/jpeg/webp are accepted; anything else is a
// safe error, never stored. The caller sets the long timeout on ctx (docs
// §22.4: image generation needs a longer budget than chat).
func (o *OpenAICompatible) GenerateImage(ctx context.Context, req ImageRequest) (string, []byte, error) {
	endpoint, err := normalizeImagesURL(o.BaseURL)
	if err != nil {
		return "", nil, err
	}
	payload := map[string]any{"model": req.Model, "prompt": req.Prompt}
	if req.Size != "" {
		payload["size"] = req.Size
	}
	if req.N > 0 {
		payload["n"] = req.N
	}
	body, _ := json.Marshal(payload)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return "", nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")
	if o.APIKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+o.APIKey)
	}
	client := o.Client
	if client == nil {
		client = SafeClient()
	}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", nil, toNetError(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		errBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return "", nil, MapImageHTTPError(resp.StatusCode, resp.Header, errBody)
	}
	var out imageResponse
	if err := json.NewDecoder(io.LimitReader(resp.Body, maxImageJSONBytes)).Decode(&out); err != nil {
		return "", nil, &PublicError{Code: "bad_response", Message: "Provider mengembalikan respons gambar yang tidak dikenali."}
	}
	return decodeImageResult(out, req)
}

// decodeImageResult turns one response item into mime + bytes. b64_json is
// decoded in place; a url/data-url item is fetched by fetchImageURL.
func decodeImageResult(out imageResponse, req ImageRequest) (string, []byte, error) {
	if len(out.Data) == 0 {
		return "", nil, &PublicError{Code: "empty_image", Message: "Provider tidak mengembalikan gambar."}
	}
	item := out.Data[0]
	if item.B64JSON != nil && *item.B64JSON != "" {
		data, err := base64.StdEncoding.DecodeString(*item.B64JSON)
		if err != nil {
			return "", nil, &PublicError{Code: "bad_response", Message: "Gambar base64 dari provider tidak valid."}
		}
		return sniffImage(data)
	}
	if item.URL != nil && *item.URL != "" {
		return "", nil, &PublicError{Code: "url_unsupported", Message: "Provider mengembalikan URL gambar; belum didukung."}
	}
	return "", nil, &PublicError{Code: "empty_image", Message: "Provider tidak mengembalikan gambar."}
}

// sniffImage validates the declared content type and returns it with the
// bytes. Only png/jpeg/webp pass.
func sniffImage(data []byte) (string, []byte, error) {
	if len(data) == 0 {
		return "", nil, &PublicError{Code: "empty_image", Message: "Provider tidak mengembalikan gambar."}
	}
	mime := http.DetectContentType(data)
	switch mime {
	case "image/png", "image/jpeg", "image/webp":
		return mime, data, nil
	default:
		return "", nil, &PublicError{Code: "unsupported_image", Message: "Format gambar dari provider tidak didukung (hanya PNG, JPEG, WebP)."}
	}
}
