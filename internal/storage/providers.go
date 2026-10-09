package storage

import (
	"time"

	"companion/internal/config"
)

// ProviderRow maps provider_profiles. No secret column, ever.
type ProviderRow struct {
	ID            string
	Name          string
	BaseURL       string
	Model         string
	CredentialRef string
	Enabled       bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// SaveProvider inserts or replaces a profile (same id). Last write wins.
func (d *DB) SaveProvider(p ProviderRow) error {
	now := time.Now().UTC().Format(time.RFC3339)
	if p.CreatedAt.IsZero() {
		p.CreatedAt = time.Now().UTC()
	}
	_, err := d.inner.Exec(`INSERT INTO provider_profiles
		(id, name, base_url, model, credential_ref, is_enabled, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			name=excluded.name, base_url=excluded.base_url, model=excluded.model,
			credential_ref=excluded.credential_ref, is_enabled=excluded.is_enabled,
			updated_at=excluded.updated_at`,
		p.ID, p.Name, p.BaseURL, p.Model, p.CredentialRef, boolToInt(p.Enabled),
		p.CreatedAt.Format(time.RFC3339), now,
	)
	return err
}

// ListProviders returns profiles ordered by creation.
func (d *DB) ListProviders() ([]ProviderRow, error) {
	rows, err := d.inner.Query(`SELECT id, name, base_url, model, credential_ref,
		is_enabled, created_at, updated_at FROM provider_profiles ORDER BY created_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ProviderRow
	for rows.Next() {
		var p ProviderRow
		var enabled int
		var created, updated string
		if err := rows.Scan(&p.ID, &p.Name, &p.BaseURL, &p.Model,
			&p.CredentialRef, &enabled, &created, &updated); err != nil {
			return nil, err
		}
		p.Enabled = enabled != 0
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
	return config.ProviderProfile{
		ID:            p.ID,
		Name:          p.Name,
		BaseURL:       p.BaseURL,
		Model:         p.Model,
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
