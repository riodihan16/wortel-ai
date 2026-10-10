package config

import (
	"strings"

	"companion/internal/ai"
)

// MaxModels caps the per-provider model list. Keeps the settings UI and the
// stored JSON bounded; there is no product reason to allow hundreds.
const MaxModels = 50

// ProviderInput is user-supplied provider data from settings UI.
// APIKey travels only on save/test, never persisted or returned.
type ProviderInput struct {
	Name    string
	BaseURL string
	Models  []string
	Model   string // active model; when empty the first of Models is used
	APIKey  string
}

// ProviderProfile is safe for frontend: no secret, only configured flag.
type ProviderProfile struct {
	ID                string   `json:"id"`
	Name              string   `json:"name"`
	BaseURL           string   `json:"baseURL"`
	Model             string   `json:"model"`
	Models            []string `json:"models"`
	ActiveModel       string   `json:"activeModel"`
	CredentialRef     string   `json:"credentialRef"`
	HasCredential     bool     `json:"hasCredential"`
	Enabled           bool     `json:"enabled"`
	InsecureTransport bool     `json:"insecureTransport,omitempty"`
}

// NormalizeModels trims entries, drops blanks and duplicates, and caps the
// list at MaxModels. Order is preserved so the first entry stays the default.
// When the list is empty but the legacy single-model field is set, that value
// is used as a one-item list. Returns the cleaned list and the active model
// (falls back to the first entry).
func NormalizeModels(models []string, active string) (clean []string, activeModel string) {
	if len(models) == 0 {
		models = []string{active}
	}
	seen := map[string]bool{}
	for _, m := range models {
		t := strings.TrimSpace(m)
		if t == "" || len(t) > 256 || seen[t] {
			continue
		}
		seen[t] = true
		clean = append(clean, t)
		if len(clean) == MaxModels {
			break
		}
	}
	if len(clean) == 0 {
		return clean, ""
	}
	activeModel = strings.TrimSpace(active)
	if activeModel == "" || !seen[activeModel] {
		activeModel = clean[0]
	}
	return clean, activeModel
}

// Validate checks display name, URL, and at least one model. Returns the
// normalized endpoint, the cleaned models + active model, and warnings.
func Validate(in ProviderInput) (endpoint string, models []string, active string, insecure bool, err error) {
	if strings.TrimSpace(in.Name) == "" {
		return "", nil, "", false, errName()
	}
	models, active = NormalizeModels(in.Models, in.Model)
	if len(models) == 0 {
		return "", nil, "", false, errModel()
	}
	endpoint, err = ai.NormalizeBaseURL(in.BaseURL)
	if err != nil {
		return "", nil, "", false, errURL()
	}
	if strings.HasPrefix(endpoint, "http://") && !isLoopback(endpoint) {
		insecure = true
	}
	return endpoint, models, active, insecure, nil
}

func isLoopback(endpoint string) bool {
	return strings.Contains(endpoint, "://localhost") ||
		strings.Contains(endpoint, "://127.") ||
		strings.Contains(endpoint, "://[::1]")
}
