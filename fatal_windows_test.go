//go:build windows

package main

import "testing"

func TestIsWebView2Missing(t *testing.T) {
	cases := map[string]bool{
		"WebView2 runtime version not available": true,
		"failed to find WebView2 Runtime":        true,
		"some other failure":                     false,
		"":                                       false,
	}
	for msg, want := range cases {
		if got := isWebView2Missing(msg); got != want {
			t.Errorf("isWebView2Missing(%q) = %v, want %v", msg, got, want)
		}
	}
}
