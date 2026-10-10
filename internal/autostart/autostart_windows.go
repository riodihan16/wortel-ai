//go:build windows

package autostart

import (
	"errors"
	"fmt"

	"golang.org/x/sys/windows/registry"
)

var errUnsupported = errors.New("autostart: tidak didukung di platform ini")

func enable(exePath string) error {
	if exePath == "" {
		return errors.New("autostart: path executable kosong")
	}
	k, _, err := registry.CreateKey(registry.CURRENT_USER, RunKey, registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("autostart: buka registry: %w", err)
	}
	defer k.Close()
	// Quote path so spaces are handled by Windows.
	if err := k.SetStringValue(ValueName, `"`+exePath+`"`); err != nil {
		return fmt.Errorf("autostart: tulis registry: %w", err)
	}
	return nil
}

func disable() error {
	k, err := registry.OpenKey(registry.CURRENT_USER, RunKey, registry.SET_VALUE)
	if err != nil {
		if errors.Is(err, registry.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("autostart: buka registry: %w", err)
	}
	defer k.Close()
	if err := k.DeleteValue(ValueName); err != nil && !errors.Is(err, registry.ErrNotExist) {
		return fmt.Errorf("autostart: hapus registry: %w", err)
	}
	return nil
}

func enabled() bool {
	k, err := registry.OpenKey(registry.CURRENT_USER, RunKey, registry.QUERY_VALUE)
	if err != nil {
		return false
	}
	defer k.Close()
	_, _, err = k.GetStringValue(ValueName)
	return err == nil
}
