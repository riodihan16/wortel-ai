//go:build windows

package autostart

import "golang.org/x/sys/windows/registry"

// readRaw returns the raw Run value, or "" when absent.
func readRaw() string {
	k, err := registry.OpenKey(registry.CURRENT_USER, RunKey, registry.QUERY_VALUE)
	if err != nil {
		return ""
	}
	defer k.Close()
	v, _, err := k.GetStringValue(ValueName)
	if err != nil {
		return ""
	}
	return v
}
