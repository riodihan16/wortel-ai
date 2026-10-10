package hotkey

// Hotkey is a parsed global shortcut: modifiers + virtual key.
// Registration is best-effort: Windows rejects combos already owned
// by another app, so the caller must surface failures (FR-010).

// Modifier bitmask matching Win32 MOD_* flags.
const (
	ModAlt      = 0x0001
	ModControl  = 0x0002
	ModShift    = 0x0004
	ModWin      = 0x0008
	ModNoRepeat = 0x4000
)

// Hotkey is a normalized global shortcut.
type Hotkey struct {
	Mods uint32
	VK   uint32
}

// Default toggles the chat: Ctrl+Shift+W (W for Wortel).
// Chosen to avoid common system shortcuts; user can change it later.
var Default = Hotkey{Mods: ModControl | ModShift | ModNoRepeat, VK: 0x57}

// Start registers the hotkey and invokes onPress each time it fires.
// Returns an error when Windows refuses the registration.
func Start(h Hotkey, onPress func()) error { return start(h, onPress) }

// Stop unregisters the hotkey. Safe to call when not registered.
func Stop() { stop() }
