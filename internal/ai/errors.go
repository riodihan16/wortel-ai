package ai

import "net/http"

// MapHTTPError converts provider status to a safe PublicError.
func MapHTTPError(status int) *PublicError {
	switch status {
	case http.StatusUnauthorized:
		return &PublicError{Code: "unauthorized", Message: "API key ditolak. Cek key dan endpoint.", Retryable: false, StatusCode: status}
	case http.StatusForbidden:
		return &PublicError{Code: "forbidden", Message: "Akses ditolak. Mungkin permission/quota.", Retryable: false, StatusCode: status}
	case http.StatusNotFound:
		return &PublicError{Code: "not_found", Message: "Endpoint atau model tidak ditemukan.", Retryable: false, StatusCode: status}
	case http.StatusTooManyRequests:
		return &PublicError{Code: "rate_limited", Message: "Rate limit. Tunggu lalu coba lagi.", Retryable: true, StatusCode: status}
	case http.StatusRequestTimeout:
		return &PublicError{Code: "timeout", Message: "Request timeout. Coba lagi.", Retryable: true, StatusCode: status}
	default:
		if status >= 500 {
			return &PublicError{Code: "provider_server", Message: "Gangguan provider. Coba lagi.", Retryable: true, StatusCode: status}
		}
		return &PublicError{Code: "provider_error", Message: "Request ditolak provider.", Retryable: false, StatusCode: status}
	}
}
