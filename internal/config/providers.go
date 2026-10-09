package config

import (
	"strings"

	"companion/internal/ai"
)

// ProviderInput is user-supplied provider data from settings UI.
// APIKey travels only on save/test, never persisted or returned.
type ProviderInput struct {
	Name    string
	BaseURL string
	Model   string
	APIKey  string
}

// ProviderProfile is safe for frontend: no secret, only configured flag.
type ProviderProfile struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	BaseURL           string `json:"baseURL"`
	Model             string `json:"model"`
	CredentialRef     string `json:"credentialRef"`
	HasCredential     bool   `json:"hasCredential"`
	Enabled           bool   `json:"enabled"`
	InsecureTransport bool   `json:"insecureTransport,omitempty"`
}

// Validate checks display name, URL, model. Returns normalized endpoint + warnings.
func Validate(in ProviderInput) (endpoint string, insecure bool, err error) {
	if strings.TrimSpace(in.Name) == "" {
		return "", false, errName()
	}
	if strings.TrimSpace(in.Model) == "" {
		return "", false, errModel()
	}
	if len(in.Model) > 256 {
		return "", false, errModel()
	}
	endpoint, err = ai.NormalizeBaseURL(in.BaseURL)
	if err != nil {
		return "", false, errURL()
	}
	if strings.HasPrefix(endpoint, "http://") && !isLoopback(endpoint) {
		insecure = true
	}
	return endpoint, insecure, nil
}

func isLoopback(endpoint string) bool {
	return strings.Contains(endpoint, "://localhost") ||
		strings.Contains(endpoint, "://127.") ||
		strings.Contains(endpoint, "://[::1]")
}
