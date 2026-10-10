//go:build windows

package main

import (
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	user32Fatal  = windows.NewLazySystemDLL("user32.dll")
	pMessageBoxW = user32Fatal.NewProc("MessageBoxW")
)

const (
	mbOK            = 0x00000000
	mbIconError     = 0x00000010
	mbSetForeground = 0x00010000
)

// fatalDialog shows a blocking message box so a start failure is visible in a
// GUI build: stdout is not attached to a console, so println() would vanish.
// The WebView2-missing case gets an actionable message (FR-001).
func fatalDialog(err error) {
	msg := err.Error()
	if isWebView2Missing(msg) {
		msg = "Wortel AI membutuhkan Microsoft Edge WebView2 Runtime, " +
			"yang tidak ditemukan di sistem ini.\n\n" +
			"Pasang WebView2 Runtime dari Microsoft lalu jalankan ulang aplikasi:\n" +
			"https://developer.microsoft.com/microsoft-edge/webview2/\n\n" +
			"Detail: " + msg
	}
	text, _ := windows.UTF16PtrFromString(msg)
	caption, _ := windows.UTF16PtrFromString("Wortel AI")
	// Owner handle 0 = desktop; a module handle is not a valid HWND and makes
	// MessageBoxW return without showing anything.
	pMessageBoxW.Call(0, uintptr(unsafe.Pointer(text)), uintptr(unsafe.Pointer(caption)), mbOK|mbIconError|mbSetForeground)
}

func isWebView2Missing(msg string) bool {
	m := strings.ToLower(msg)
	return strings.Contains(m, "webview2") || strings.Contains(m, "runtime version not available")
}
