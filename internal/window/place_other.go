//go:build !windows

package window

import "errors"

// Apply needs Win32 monitor info; off Windows the caller falls back to a
// plain resize.
func Apply(m Mode) error {
	if _, err := SizeOf(m); err != nil {
		return err
	}
	return errors.ErrUnsupported
}

// KeepOnScreen is Windows-only; elsewhere drags are left to the OS.
func KeepOnScreen() error { return errors.ErrUnsupported }

// Activate is Windows-only; elsewhere WindowShow is all the app does.
func Activate() error { return errors.ErrUnsupported }
