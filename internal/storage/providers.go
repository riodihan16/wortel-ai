package storage

import (
	"encoding/json"
	"time"

	"companion/internal/config"
)

// ProviderRow maps provider_profiles. No secret column, ever.
type ProviderRow struct {
	ID      string
	Name    string
	BaseURL string
	// Model is the legacy single-model column. Kept in sync with the first
	// entry of Models so a v2 reader still works; Models is authoritative.
	Model         string
	Models        []config.ModelEntry
	ActiveModel   string
	CredentialRef string
	Enabled       bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// encodeModels marshals the model list for storage. Nil -> "[]".
func encodeModels(models []config.ModelEntry) (string, error) {
	if models == nil {
		return "[]", nil
	}
	b, err := json.Marshal(models)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// decodeModels parses the stored JSON array of entries. Malformed values
// degrade to nil rather than failing the whole provider read.
func decodeModels(raw string) []config.ModelEntry {
	if raw == "" {
		return nil
	}
	var out []config.ModelEntry
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil
	}
	return out
}

// SaveProvider inserts or replaces a profile (same id). Last write wins.
func (d *DB) SaveProvider(p ProviderRow) error {
	now := time.Now().UTC().Format(time.RFC3339)
	if p.CreatedAt.IsZero() {
		p.CreatedAt = time.Now().UTC()
	}
	// Keep the legacy single-model column equal to the active model (or the
	// first listed model) so a reader of the old schema still sees the model
	// actually used for chat.
	legacy := p.Model
	if p.ActiveModel != "" {
		legacy = p.ActiveModel
	} else if len(p.Models) > 0 {
		legacy = p.Models[0].ID
	}
	modelsJSON, err := encodeModels(p.Models)
	if err != nil {
		return err
	}
	_, err = d.inner.Exec(`INSERT INTO provider_profiles
		(id, name, base_url, model, models, active_model, credential_ref, is_enabled, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			name=excluded.name, base_url=excluded.base_url, model=excluded.model,
			models=excluded.models, active_model=excluded.active_model,
			credential_ref=excluded.credential_ref, is_enabled=excluded.is_enabled,
			updated_at=excluded.updated_at`,
		p.ID, p.Name, p.BaseURL, legacy, modelsJSON, p.ActiveModel,
		p.CredentialRef, boolToInt(p.Enabled),
		p.CreatedAt.Format(time.RFC3339), now,
	)
	return err
}

// ListProviders returns profiles ordered by creation.
func (d *DB) ListProviders() ([]ProviderRow, error) {
	rows, err := d.inner.Query(`SELECT id, name, base_url, model, models, active_model,
		credential_ref, is_enabled, created_at, updated_at
		FROM provider_profiles ORDER BY created_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ProviderRow
	for rows.Next() {
		var p ProviderRow
		var enabled int
		var created, updated, modelsJSON string
		if err := rows.Scan(&p.ID, &p.Name, &p.BaseURL, &p.Model, &modelsJSON,
			&p.ActiveModel, &p.CredentialRef, &enabled, &created, &updated); err != nil {
			return nil, err
		}
		p.Enabled = enabled != 0
		p.Models = decodeModels(modelsJSON)
		p.CreatedAt, _ = time.Parse(time.RFC3339, created)
		p.UpdatedAt, _ = time.Parse(time.RFC3339, updated)
		out = append(out, p)
	}
	return out, rows.Err()
}

// DeleteProvider removes a profile by id.
func (d *DB) DeleteProvider(id string) error {
	_, err := d.inner.Exec(`DELETE FROM provider_profiles WHERE id = ?`, id)
	return err
}

// GetSetting reads one app setting; empty string when absent.
func (d *DB) GetSetting(key string) (string, error) {
	var v string
	err := d.inner.QueryRow(`SELECT value FROM app_settings WHERE key = ?`, key).Scan(&v)
	if err != nil {
		return "", nil
	}
	return v, nil
}

// SetSetting upserts one app setting.
func (d *DB) SetSetting(key, value string) error {
	_, err := d.inner.Exec(`INSERT INTO app_settings (key, value) VALUES (?, ?)
		ON CONFLICT(key) DO UPDATE SET value=excluded.value`, key, value)
	return err
}

// ToPublic converts a row to the frontend-safe profile.
func ToPublic(p ProviderRow, hasCredential bool) config.ProviderProfile {
	models := p.Models
	if len(models) == 0 && p.Model != "" {
		models = []config.ModelEntry{{ID: p.Model}}
	}
	active := p.ActiveModel
	if active == "" && len(models) > 0 {
		active = models[0].ID
	}
	return config.ProviderProfile{
		ID:            p.ID,
		Name:          p.Name,
		BaseURL:       p.BaseURL,
		Model:         active,
		Models:        models,
		ActiveModel:   active,
		CredentialRef: p.CredentialRef,
		HasCredential: hasCredential,
		Enabled:       p.Enabled,
	}
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
