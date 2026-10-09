package app

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"companion/internal/ai"
	"companion/internal/config"
	"companion/internal/credentials"
	"companion/internal/storage"

	"github.com/google/uuid"
)

// ActiveProviderKey is the app_settings key for the active profile id.
const ActiveProviderKey = "active_provider_id"

// ErrNoActiveProvider signals chat attempted with no provider selected.
var ErrNoActiveProvider = errors.New("tidak ada provider aktif: tambahkan provider di Settings")

// ErrActiveDelete blocks deleting the active profile without switching first.
var ErrActiveDelete = errors.New("provider sedang aktif: pilih provider lain dulu sebelum menghapus")

// TestResult is safe for frontend. Never carries secrets.
type TestResult struct {
	OK            bool   `json:"ok"`
	Code          string `json:"code,omitempty"`
	Message       string `json:"message"`
	Retryable     bool   `json:"retryable,omitempty"`
	StatusCode    int    `json:"statusCode,omitempty"`
	RetryAfterSec int    `json:"retryAfterSec,omitempty"`
	Insecure      bool   `json:"insecure,omitempty"`
}

// Providers manages profiles + OS credentials. No secret ever returned.
type Providers struct {
	db    *storage.DB
	creds credentials.Store
	http  *http.Client
}

func NewProviders(db *storage.DB, creds credentials.Store) *Providers {
	return &Providers{db: db, creds: creds, http: &http.Client{Timeout: 20 * time.Second}}
}

// List returns public profiles with credential presence.
func (p *Providers) List(ctx context.Context) ([]config.ProviderProfile, error) {
	rows, err := p.db.ListProviders()
	if err != nil {
		return nil, err
	}
	out := make([]config.ProviderProfile, 0, len(rows))
	for _, r := range rows {
		has, err := p.creds.Exists(ctx, r.CredentialRef)
		if err != nil {
			has = false
		}
		out = append(out, storage.ToPublic(r, has))
	}
	return out, nil
}

// SaveInput carries the settings form. ID empty = create.
type SaveInput struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	BaseURL string `json:"baseURL"`
	Model   string `json:"model"`
	APIKey  string `json:"apiKey"`
	Enabled bool   `json:"enabled"`
}

// Save validates, writes credential to OS store, persists profile.
func (p *Providers) Save(ctx context.Context, in SaveInput) (config.ProviderProfile, error) {
	endpoint, insecure, err := config.Validate(config.ProviderInput{
		Name: in.Name, BaseURL: in.BaseURL, Model: in.Model,
	})
	if err != nil {
		return config.ProviderProfile{}, err
	}
	_ = endpoint
	id := strings.TrimSpace(in.ID)
	if id == "" {
		id = uuid.NewString()
	}
	ref := "wortel-provider-" + id
	if strings.TrimSpace(in.APIKey) != "" {
		if err := p.creds.Set(ctx, ref, []byte(strings.TrimSpace(in.APIKey))); err != nil {
			return config.ProviderProfile{}, errors.New("gagal menyimpan API key ke credential store")
		}
	} else {
		has, err := p.creds.Exists(ctx, ref)
		if err != nil || !has {
			return config.ProviderProfile{}, errors.New("API key wajib diisi untuk provider baru")
		}
	}
	row := storage.ProviderRow{
		ID: id, Name: strings.TrimSpace(in.Name),
		BaseURL: strings.TrimSpace(in.BaseURL), Model: strings.TrimSpace(in.Model),
		CredentialRef: ref, Enabled: true,
	}
	if err := p.db.SaveProvider(row); err != nil {
		return config.ProviderProfile{}, errors.New("gagal menyimpan provider")
	}
	// First provider becomes active automatically.
	if active, _ := p.db.GetSetting(ActiveProviderKey); active == "" {
		_ = p.db.SetSetting(ActiveProviderKey, id)
	}
	has, _ := p.creds.Exists(ctx, ref)
	prof := storage.ToPublic(row, has)
	prof.InsecureTransport = insecure
	return prof, nil
}

// Delete removes profile + credential. Refuses when active.
func (p *Providers) Delete(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return errors.New("id provider kosong")
	}
	if active, _ := p.db.GetSetting(ActiveProviderKey); active == id {
		return ErrActiveDelete
	}
	rows, err := p.db.ListProviders()
	if err != nil {
		return err
	}
	ref := ""
	for _, r := range rows {
		if r.ID == id {
			ref = r.CredentialRef
		}
	}
	if err := p.db.DeleteProvider(id); err != nil {
		return errors.New("gagal menghapus provider")
	}
	if ref != "" {
		_ = p.creds.Delete(ctx, ref)
	}
	return nil
}

// SetActive selects the provider used for chat.
func (p *Providers) SetActive(id string) error {
	rows, err := p.db.ListProviders()
	if err != nil {
		return err
	}
	for _, r := range rows {
		if r.ID == id {
			return p.db.SetSetting(ActiveProviderKey, id)
		}
	}
	return errors.New("provider tidak ditemukan")
}

// ActiveID returns the selected profile id, empty when none.
func (p *Providers) ActiveID() string {
	id, _ := p.db.GetSetting(ActiveProviderKey)
	return id
}

// Resolve builds a live provider client from the active profile.
// Returns provider, model, error. Secret stays in backend memory only.
func (p *Providers) Resolve(ctx context.Context) (*ai.OpenAICompatible, string, error) {
	id := p.ActiveID()
	if id == "" {
		return nil, "", ErrNoActiveProvider
	}
	rows, err := p.db.ListProviders()
	if err != nil {
		return nil, "", err
	}
	for _, r := range rows {
		if r.ID != id {
			continue
		}
		secret, err := p.creds.Get(ctx, r.CredentialRef)
		if err != nil {
			return nil, "", errors.New("credential provider tidak tersedia: isi ulang API key")
		}
		key := string(secret)
		// Best-effort wipe of the copy; original slice from store is a clone.
		defer func() {
			for i := range secret {
				secret[i] = 0
			}
		}()
		return &ai.OpenAICompatible{BaseURL: r.BaseURL, APIKey: key, Client: p.http}, r.Model, nil
	}
	return nil, "", ErrNoActiveProvider
}

// Test checks URL/model/credential with a minimal non-stream request.
// Uses the form API key when given, else the stored credential for id.
func (p *Providers) Test(ctx context.Context, in SaveInput) TestResult {
	_, insecure, err := config.Validate(config.ProviderInput{
		Name: firstOr(in.Name, "test"), BaseURL: in.BaseURL, Model: in.Model,
	})
	if err != nil {
		return TestResult{Message: err.Error()}
	}
	endpoint, err := ai.NormalizeBaseURL(in.BaseURL)
	if err != nil {
		return TestResult{Message: "base URL tidak valid"}
	}
	key := strings.TrimSpace(in.APIKey)
	if key == "" && strings.TrimSpace(in.ID) != "" {
		rows, err := p.db.ListProviders()
		if err == nil {
			for _, r := range rows {
				if r.ID == strings.TrimSpace(in.ID) {
					if b, err := p.creds.Get(ctx, r.CredentialRef); err == nil {
						key = string(b)
						for i := range b {
							b[i] = 0
						}
					}
				}
			}
		}
	}
	if key == "" {
		return TestResult{Message: "API key belum diisi"}
	}
	tctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	payload, _ := json.Marshal(map[string]any{
		"model":      strings.TrimSpace(in.Model),
		"messages":   []ai.Message{{Role: "user", Content: "ping"}},
		"stream":     false,
		"max_tokens": 1,
	})
	req, err := http.NewRequestWithContext(tctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return TestResult{Message: "gagal membuat request"}
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+key)
	resp, err := p.http.Do(req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return TestResult{Code: "timeout", Message: "Timeout. Coba lagi.", Retryable: true}
		}
		return TestResult{Code: "network", Message: "Koneksi gagal. Cek URL dan internet.", Retryable: true}
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		pub := ai.MapHTTPErrorWithHeaders(resp.StatusCode, resp.Header)
		return TestResult{Code: pub.Code, Message: pub.Message,
			Retryable: pub.Retryable, StatusCode: pub.StatusCode,
			RetryAfterSec: pub.RetryAfterSec, Insecure: insecure}
	}
	var probe struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&probe)
	_ = probe
	return TestResult{OK: true, Message: "Koneksi OK.", Insecure: insecure}
}

func firstOr(s, fb string) string {
	if strings.TrimSpace(s) == "" {
		return fb
	}
	return s
}
