package hotkey

import (
	"errors"
	"fmt"
	"strings"
)

// keyNames maps supported virtual keys to display names. Letters, digits,
// and function keys cover the realistic set for a chat toggle; unknown keys
// from a capture are rejected instead of guessed.
var keyNames = map[uint32]string{
	// A-Z
	0x41: "A", 0x42: "B", 0x43: "C", 0x44: "D", 0x45: "E", 0x46: "F",
	0x47: "G", 0x48: "H", 0x49: "I", 0x4A: "J", 0x4B: "K", 0x4C: "L",
	0x4D: "M", 0x4E: "N", 0x4F: "O", 0x50: "P", 0x51: "Q", 0x52: "R",
	0x53: "S", 0x54: "T", 0x55: "U", 0x56: "V", 0x57: "W", 0x58: "X",
	0x59: "Y", 0x5A: "Z",
	// 0-9
	0x30: "0", 0x31: "1", 0x32: "2", 0x33: "3", 0x34: "4",
	0x35: "5", 0x36: "6", 0x37: "7", 0x38: "8", 0x39: "9",
	// F1-F12
	0x70: "F1", 0x71: "F2", 0x72: "F3", 0x73: "F4", 0x74: "F5",
	0x75: "F6", 0x76: "F7", 0x77: "F8", 0x78: "F9", 0x79: "F10",
	0x7A: "F11", 0x7B: "F12",
}

// Reverse lookup built once at init.
var nameKeys = func() map[string]uint32 {
	m := make(map[string]uint32, len(keyNames))
	for vk, name := range keyNames {
		m[name] = vk
	}
	return m
}()

// ErrUnsupportedKey is returned when a key has no display mapping.
var ErrUnsupportedKey = errors.New("tombol tidak didukung: pakai huruf, angka, atau F1-F12")

// String renders the combo in display order: Ctrl+Alt+Shift+Win+Key.
func (h Hotkey) String() string {
	parts := make([]string, 0, 5)
	if h.Mods&ModControl != 0 {
		parts = append(parts, "Ctrl")
	}
	if h.Mods&ModAlt != 0 {
		parts = append(parts, "Alt")
	}
	if h.Mods&ModShift != 0 {
		parts = append(parts, "Shift")
	}
	if h.Mods&ModWin != 0 {
		parts = append(parts, "Win")
	}
	name, ok := keyNames[h.VK]
	if !ok {
		name = fmt.Sprintf("0x%02X", h.VK)
	}
	return strings.Join(append(parts, name), "+")
}

// Parse reads a combo created by String. Case-insensitive on modifier names
// and key letters. MOD_NOREPEAT is always applied. Requires >=1 modifier
// so a bare letter cannot hijack normal typing system-wide.
func Parse(s string) (Hotkey, error) {
	var h Hotkey
	seenKey := false
	for _, part := range strings.Split(s, "+") {
		p := strings.TrimSpace(part)
		if p == "" {
			continue
		}
		switch strings.ToLower(p) {
		case "ctrl", "control":
			h.Mods |= ModControl
		case "alt":
			h.Mods |= ModAlt
		case "shift":
			h.Mods |= ModShift
		case "win", "super", "meta":
			h.Mods |= ModWin
		default:
			vk, ok := nameKeys[strings.ToUpper(p)]
			if !ok {
				return Hotkey{}, ErrUnsupportedKey
			}
			h.VK = vk
			seenKey = true
		}
	}
	if !seenKey {
		return Hotkey{}, errors.New("tidak ada tombol utama pada shortcut")
	}
	if h.Mods&(ModControl|ModAlt|ModShift|ModWin) == 0 {
		return Hotkey{}, errors.New("shortcut butuh minimal satu modifier (Ctrl/Alt/Shift/Win)")
	}
	h.Mods |= ModNoRepeat
	return h, nil
}

// ParseOrDefault returns the default combo when s is empty or invalid.
// Used at startup so a corrupt setting never disables the app silently.
func ParseOrDefault(s string) (Hotkey, bool) {
	if strings.TrimSpace(s) == "" {
		return Default, true
	}
	h, err := Parse(s)
	if err != nil {
		return Default, false
	}
	return h, true
}
