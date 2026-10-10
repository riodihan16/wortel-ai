//go:build windows

package hotkey

import (
	"errors"
	"fmt"
	"runtime"
	"sync"
	"unsafe"

	"golang.org/x/sys/windows"
)

// Win32 message ids and helpers used by the hotkey loop.
const (
	wmHotkey = 0x0312
	wmQuit   = 0x0012
	hotkeyID = 1
)

var (
	user32   = windows.NewLazySystemDLL("user32.dll")
	kernel32 = windows.NewLazySystemDLL("kernel32.dll")

	pRegisterHotKey     = user32.NewProc("RegisterHotKey")
	pUnregisterHotKey   = user32.NewProc("UnregisterHotKey")
	pGetMessage         = user32.NewProc("GetMessageW")
	pPostThreadMessage  = user32.NewProc("PostThreadMessageW")
	pGetCurrentThreadID = kernel32.NewProc("GetCurrentThreadId")
)

var (
	mu       sync.Mutex
	running  bool
	done     chan struct{}
	threadID uint32
)

// msgT mirrors Win32 MSG layout for GetMessageW.
type msgT struct {
	hwnd    windows.Handle
	message uint32
	_       uint32
	wParam  uintptr
	lParam  uintptr
	time    uint32
	_       uint32
	ptX     int32
	ptY     int32
}

func start(h Hotkey, onPress func()) error {
	mu.Lock()
	if running {
		mu.Unlock()
		return nil
	}
	running = true
	done = make(chan struct{})
	mu.Unlock()

	ready := make(chan error, 1)
	go func() {
		// RegisterHotKey posts WM_HOTKEY to THIS thread's queue, so the
		// goroutine must stay pinned to one OS thread for GetMessage.
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()

		tid, _, _ := pGetCurrentThreadID.Call()
		mu.Lock()
		threadID = uint32(tid)
		mu.Unlock()

		r, _, e := pRegisterHotKey.Call(0, hotkeyID, uintptr(h.Mods), uintptr(h.VK))
		if r == 0 {
			if e == nil || errors.Is(e, windows.ERROR_SUCCESS) {
				e = errors.New("kombinasi sudah dipakai aplikasi lain")
			}
			ready <- fmt.Errorf("hotkey: registrasi gagal: %w", e)
			mu.Lock()
			running = false
			d := done
			mu.Unlock()
			close(d)
			return
		}
		ready <- nil

		var m msgT
		for {
			ret, _, _ := pGetMessage.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
			if ret == 0 || int32(ret) == -1 {
				break
			}
			if m.message == wmHotkey && m.wParam == hotkeyID && onPress != nil {
				go onPress()
			}
		}

		_, _, _ = pUnregisterHotKey.Call(0, hotkeyID)
		mu.Lock()
		running = false
		d := done
		mu.Unlock()
		close(d)
	}()

	return <-ready
}

func stop() {
	mu.Lock()
	r := running
	d := done
	tid := threadID
	mu.Unlock()
	if !r {
		return
	}
	// Post WM_QUIT to the loop thread to end GetMessage.
	_, _, _ = pPostThreadMessage.Call(uintptr(tid), uintptr(wmQuit), 0, 0)
	<-d
}
