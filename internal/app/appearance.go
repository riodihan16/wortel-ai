package app

import (
	"errors"
	"strconv"
)

// Appearance settings keys in app_settings.
const (
	keyTheme     = "appearance_theme"
	keyFontScale = "appearance_font_scale"
)

// Theme values accepted by the frontend.
const (
	ThemeDark  = "dark"
	ThemeLight = "light"
)

// FontScale bounds (docs 06 §6: UI harus tetap terbaca, kontras memadai).
const (
	minFontScale = 0.85
	maxFontScale = 1.4
)

// Appearance is the frontend-safe appearance settings payload.
type Appearance struct {
	Theme     string  `json:"theme"`
	FontScale float64 `json:"fontScale"`
}

// DefaultAppearance is used when nothing is stored yet.
func DefaultAppearance() Appearance {
	return Appearance{Theme: ThemeDark, FontScale: 1.0}
}

// GetAppearance returns stored appearance or defaults. Never errors on
// missing keys; a DB failure falls back to defaults.
func (p *Providers) GetAppearance() Appearance {
	out := DefaultAppearance()
	if v, err := p.db.GetSetting(keyTheme); err == nil && (v == ThemeDark || v == ThemeLight) {
		out.Theme = v
	}
	if v, err := p.db.GetSetting(keyFontScale); err == nil && v != "" {
		if f, perr := strconv.ParseFloat(v, 64); perr == nil && f >= minFontScale && f <= maxFontScale {
			out.FontScale = f
		}
	}
	return out
}

// SetAppearance validates and persists appearance settings.
// Rejects unknown theme or out-of-range scale so the UI cannot be unusable.
func (p *Providers) SetAppearance(in Appearance) error {
	if in.Theme != ThemeDark && in.Theme != ThemeLight {
		return errors.New("theme harus 'dark' atau 'light'")
	}
	if in.FontScale < minFontScale || in.FontScale > maxFontScale {
		return errors.New("skala font di luar rentang yang diizinkan")
	}
	if err := p.db.SetSetting(keyTheme, in.Theme); err != nil {
		return err
	}
	return p.db.SetSetting(keyFontScale, strconv.FormatFloat(in.FontScale, 'f', 2, 64))
}
