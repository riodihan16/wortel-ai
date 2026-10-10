//go:build !windows

package hotkey

import "errors"

// Non-Windows stub: global hotkeys are a Windows-only feature for the MVP.

func start(h Hotkey, onPress func()) error {
	return errors.New("hotkey: hanya Windows yang didukung")
}

func stop() {}
