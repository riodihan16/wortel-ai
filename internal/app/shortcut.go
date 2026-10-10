package app

import (
	"strconv"

	"companion/internal/hotkey"
)

// Shortcut settings keys in app_settings.
const (
	keyShortcutEnabled = "shortcut_enabled"
	keyShortcutCombo   = "shortcut_combo"
)

// Shortcut is the frontend-safe global-shortcut setting.
type Shortcut struct {
	Enabled bool   `json:"enabled"`
	Combo   string `json:"combo"`
}

// DefaultShortcut is used when nothing is stored yet.
func DefaultShortcut() Shortcut {
	return Shortcut{Enabled: true, Combo: hotkey.Default.String()}
}

// GetShortcut returns the stored shortcut, falling back to defaults for
// missing or unparsable values so the app never starts with a dead key.
func (p *Providers) GetShortcut() Shortcut {
	out := DefaultShortcut()
	if v, err := p.db.GetSetting(keyShortcutEnabled); err == nil && v != "" {
		if b, perr := strconv.ParseBool(v); perr == nil {
			out.Enabled = b
		}
	}
	if v, err := p.db.GetSetting(keyShortcutCombo); err == nil && v != "" {
		out.Combo = v
	}
	return out
}

// SetShortcut validates and persists shortcut settings. The combo must parse
// so an unregisterable value is rejected before it reaches the OS.
func (p *Providers) SetShortcut(in Shortcut) error {
	if _, err := hotkey.Parse(in.Combo); err != nil {
		return err
	}
	if err := p.db.SetSetting(keyShortcutEnabled, strconv.FormatBool(in.Enabled)); err != nil {
		return err
	}
	return p.db.SetSetting(keyShortcutCombo, in.Combo)
}

// ResolvedShortcut returns the parsed combo plus whether the shortcut should
// run. Unparsable stored combos fall back to the default.
func (p *Providers) ResolvedShortcut() (hotkey.Hotkey, bool) {
	s := p.GetShortcut()
	h, _ := hotkey.ParseOrDefault(s.Combo)
	return h, s.Enabled
}
