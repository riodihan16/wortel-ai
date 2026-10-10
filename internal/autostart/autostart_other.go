//go:build !windows

package autostart

import "errors"

var errUnsupported = errors.New("autostart: hanya Windows yang didukung")

func enable(exePath string) error { return errUnsupported }
func disable() error              { return errUnsupported }
func enabled() bool               { return false }
