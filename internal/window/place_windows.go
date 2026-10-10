//go:build windows

package window

import (
	"errors"
	"fmt"
	"os"
	"sync"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	user32             = windows.NewLazySystemDLL("user32.dll")
	pGetWindowRect     = user32.NewProc("GetWindowRect")
	pMonitorFromWindow = user32.NewProc("MonitorFromWindow")
	pGetMonitorInfo    = user32.NewProc("GetMonitorInfoW")
	pGetDpiForWindow   = user32.NewProc("GetDpiForWindow")
	pSetWindowPos      = user32.NewProc("SetWindowPos")
)

const (
	monitorDefaultToNearest = 2
	swpNoZOrder             = 0x0004
	swpNoActivate           = 0x0010
	baseDPI                 = 96
)

type monitorInfo struct {
	cbSize  uint32
	monitor windows.Rect
	work    windows.Rect
	flags   uint32
}

var errNoWindow = errors.New("window aplikasi tidak ditemukan")

// All state below is guarded by mu. Bindings and tray callbacks run on
// different goroutines and may switch modes concurrently.
var (
	mu       sync.Mutex
	placer   Placer
	mainHWND windows.HWND
	enumHit  windows.HWND
	selfPID  = uint32(os.Getpid())
	// NewCallback slots are never freed, so create the enum callback once.
	enumCB = windows.NewCallback(matchMainWindow)
)

// Apply resizes the main window for mode and moves it the least needed to
// stay inside the work area (taskbar excluded) of the monitor it is on.
func Apply(m Mode) error {
	size, err := SizeOf(m)
	if err != nil {
		return err
	}
	mu.Lock()
	defer mu.Unlock()

	h, err := mainWindow()
	if err != nil {
		return err
	}
	var r windows.Rect
	if ok, _, e := pGetWindowRect.Call(uintptr(h), uintptr(unsafe.Pointer(&r))); ok == 0 {
		return fmt.Errorf("GetWindowRect: %w", e)
	}
	work, err := workArea(h)
	if err != nil {
		return err
	}
	dpi, _, _ := pGetDpiForWindow.Call(uintptr(h))
	if dpi == 0 {
		dpi = baseDPI
	}
	w, ht := size.W*int(dpi)/baseDPI, size.H*int(dpi)/baseDPI

	// Plan on a copy: the remembered shift must only change if the window
	// really moved, or the next pet return would undo a shift never applied.
	next := placer
	x, y := next.Place(m, int(r.Left), int(r.Top), w, ht, work)
	ok, _, e := pSetWindowPos.Call(uintptr(h), 0,
		uintptr(int32(x)), uintptr(int32(y)), uintptr(w), uintptr(ht),
		swpNoZOrder|swpNoActivate)
	if ok == 0 {
		return fmt.Errorf("SetWindowPos: %w", e)
	}
	placer = next
	return nil
}

func workArea(h windows.HWND) (Rect, error) {
	mon, _, _ := pMonitorFromWindow.Call(uintptr(h), monitorDefaultToNearest)
	return monitorWorkArea(mon)
}

func monitorWorkArea(mon uintptr) (Rect, error) {
	mi := monitorInfo{cbSize: uint32(unsafe.Sizeof(monitorInfo{}))}
	if ok, _, e := pGetMonitorInfo.Call(mon, uintptr(unsafe.Pointer(&mi))); ok == 0 {
		return Rect{}, fmt.Errorf("GetMonitorInfo: %w", e)
	}
	return Rect{Left: int(mi.work.Left), Top: int(mi.work.Top), Right: int(mi.work.Right), Bottom: int(mi.work.Bottom)}, nil
}

// mainWindow finds this process's top-level window of ClassName once and
// caches it; the Wails window lives as long as the app.
func mainWindow() (windows.HWND, error) {
	if mainHWND != 0 {
		return mainHWND, nil
	}
	enumHit = 0
	// EnumWindows reports an error whenever the callback stops early, which
	// is exactly the success path here, so judge by enumHit instead.
	_ = windows.EnumWindows(enumCB, nil)
	if enumHit == 0 {
		return 0, errNoWindow
	}
	mainHWND = enumHit
	return mainHWND, nil
}

func matchMainWindow(h windows.HWND, _ uintptr) uintptr {
	var pid uint32
	if _, err := windows.GetWindowThreadProcessId(h, &pid); err != nil || pid != selfPID {
		return 1
	}
	var buf [64]uint16
	n, err := windows.GetClassName(h, &buf[0], int32(len(buf)))
	if err != nil || windows.UTF16ToString(buf[:n]) != ClassName {
		return 1
	}
	enumHit = h
	return 0
}
