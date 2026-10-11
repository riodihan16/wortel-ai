package ai

import (
	"context"
	"errors"
	"net/http"
	"regexp"
	"strconv"
)

func isCtx(err error) bool {
	return errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)
}

// MapHTTPError converts provider status to a safe PublicError.
func MapHTTPError(status int) *PublicError {
	return MapHTTPErrorWithHeaders(status, nil)
}

// MapHTTPErrorWithHeaders honors Retry-After on 429/503.
func MapHTTPErrorWithHeaders(status int, h http.Header) *PublicError {
	mk := func(code, msg string, retry bool) *PublicError {
		return &PublicError{Code: code, Message: msg, Retryable: retry, StatusCode: status, RetryAfterSec: retryAfter(h)}
	}
	switch status {
	case http.StatusBadRequest:
		return mk("bad_request", "Request/parameter ditolak provider.", false)
	case http.StatusUnauthorized:
		return mk("unauthorized", "API key ditolak. Cek key dan endpoint.", false)
	case http.StatusForbidden:
		return mk("forbidden", "Akses ditolak. Mungkin permission/quota.", false)
	case http.StatusNotFound:
		return mk("not_found", "Endpoint atau model tidak ditemukan.", false)
	case http.StatusTooManyRequests:
		return mk("rate_limited", "Rate limit. Tunggu lalu coba lagi.", true)
	case http.StatusRequestTimeout:
		return mk("timeout", "Request timeout. Coba lagi.", true)
	default:
		if status >= 500 {
			return mk("provider_server", "Gangguan provider. Coba lagi.", true)
		}
		return mk("provider_error", "Request ditolak provider.", false)
	}
}

// attachmentWords marks a 400 that is about the input modality rather than
// a generic parameter problem, so the UI can say "this model does not take
// images" instead of "request rejected". Word-bounded: "profile" must not
// match "file". The underscore forms are listed whole because `_` is a word
// character, so `\bimage\b` does not match inside "input_image".
var attachmentWords = regexp.MustCompile(`(?i)\b(image|images|image_url|input_image|image_input|file|files|vision|modal|modality|modalities|multimodal|document|documents)\b`)

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

// retryAfter parses Retry-After seconds, 0 when absent/invalid.
// HTTP-date form ignored on purpose: MVP shows manual retry.
func retryAfter(h http.Header) int {
	if h == nil {
		return 0
	}
	v := h.Get("Retry-After")
	if v == "" {
		return 0
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 0 {
		return 0
	}
	if n > 600 {
		n = 600
	}
	return n
}

// toNetError wraps transport errors safe for UI. Ctx cancel stays raw
// so service can tell cancelled apart from failed.
func toNetError(err error) error {
	if err == nil {
		return nil
	}
	if isCtx(err) {
		return err
	}
	var te interface{ Timeout() bool }
	if errors.As(err, &te) && te.Timeout() {
		return &PublicError{Code: "timeout", Message: "Request timeout. Coba lagi.", Retryable: true}
	}
	return &PublicError{Code: "network", Message: "Koneksi gagal. Cek URL dan internet.", Retryable: true}
}
