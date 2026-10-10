//go:build windows

package tray

import (
	"fmt"
	"sync"
	"unsafe"

	"golang.org/x/sys/windows"
)

func errNIDSize(n uintptr) error {
	return fmt.Errorf("tray: NOTIFYICONDATA size %d, want %d", n, wantNIDSize)
}

var (
	user32   = windows.NewLazySystemDLL("user32.dll")
	shell32  = windows.NewLazySystemDLL("shell32.dll")
	kernel32 = windows.NewLazySystemDLL("kernel32.dll")

	pRegisterClassEx  = user32.NewProc("RegisterClassExW")
	pCreateWindowEx   = user32.NewProc("CreateWindowExW")
	pDefWindowProc    = user32.NewProc("DefWindowProcW")
	pDestroyWindow    = user32.NewProc("DestroyWindow")
	pGetMessage       = user32.NewProc("GetMessageW")
	pTranslateMessage = user32.NewProc("TranslateMessage")
	pDispatchMessage  = user32.NewProc("DispatchMessageW")
	pPostMessage      = user32.NewProc("PostMessageW")
	pLoadIcon         = user32.NewProc("LoadIconW")
	pLoadCursor       = user32.NewProc("LoadCursorW")
	pCreatePopupMenu  = user32.NewProc("CreatePopupMenu")
	pAppendMenu       = user32.NewProc("AppendMenuW")
	pDestroyMenu      = user32.NewProc("DestroyMenu")
	pTrackPopupMenu   = user32.NewProc("TrackPopupMenu")
	pGetCursorPos     = user32.NewProc("GetCursorPos")
	pSetForeground    = user32.NewProc("SetForegroundWindow")
	pUnregisterClass  = user32.NewProc("UnregisterClassW")
	pRegWindowMsg     = user32.NewProc("RegisterWindowMessageW")
	pPostQuitMessage  = user32.NewProc("PostQuitMessage")

	pShellNotifyIcon = shell32.NewProc("Shell_NotifyIconW")
	pGetModuleHandle = kernel32.NewProc("GetModuleHandleW")
)

const (
	wmApp      = 0x8000
	wmTray     = wmApp + 1
	wmClose    = 0x0010
	wmDestroy  = 0x0002
	wmCommand  = 0x0111
	wmNull     = 0x0000
	wmLBtnUp   = 0x0202
	wmLBtnDbl  = 0x0203
	wmRBtnUp   = 0x0205
	idiApp     = 32512
	idcArrow   = 32512
	mfString   = 0x00000000
	mfSep      = 0x00000800
	tpmButtons = 0x0002 | 0x0100 // RIGHTBUTTON | RETURNCMD
	nifMessage = 0x00000001
	nifIcon    = 0x00000002
	nifTip     = 0x00000004
	nimAdd     = 0x00000000
	nimModify  = 0x00000001
	nimDelete  = 0x00000002
)

type wndClassEx struct {
	size       uint32
	style      uint32
	wndProc    uintptr
	clsExtra   int32
	wndExtra   int32
	instance   windows.Handle
	icon       windows.Handle
	cursor     windows.Handle
	background windows.Handle
	menuName   *uint16
	className  *uint16
	iconSm     windows.Handle
}

// wantNIDSize is the measured NOTIFYICONDATAW size on 64-bit (976).
// Field order follows MSDN; union uTimeout/uVersion is ONE uint32.
// Pinned by TestNIDSize so layout drift fails fast.
const wantNIDSize = 976

type notifyIconData struct {
	size            uint32
	wnd             windows.Handle
	id              uint32
	flags           uint32
	callbackMessage uint32
	icon            windows.Handle
	tip             [128]uint16
	state           uint32
	stateMask       uint32
	info            [256]uint16
	verOrTimeout    uint32 // union uTimeout/uVersion: SATU field; padding follows
	infoTitle       [64]uint16
	infoFlags       uint32
	guid            windows.GUID
	balloonIcon     windows.Handle
}

type point struct{ x, y int32 }

type msg struct {
	hwnd    windows.Handle
	message uint32
	_       uint32 // padding
	wParam  uintptr
	lParam  uintptr
	time    uint32
	pt      point
	private uint32
}

var (
	mu       sync.Mutex
	running  bool
	hwnd     windows.Handle
	hMenu    windows.Handle
	hInst    windows.Handle
	nid      notifyIconData
	cb       Callbacks
	done     chan struct{}
	ready    chan error
	taskMsg  uint32
	wndProc  uintptr
	classBuf *uint16
	lastErr  string
)

func setErr(err error) {
	if err == nil {
		lastErr = ""
		return
	}
	lastErr = err.Error()
}

// LastError returns the last tray init failure, if any.
func LastError() string {
	mu.Lock()
	defer mu.Unlock()
	return lastErr
}

// SetError records an external tray failure (e.g. Start return).
func SetError(err error) {
	mu.Lock()
	defer mu.Unlock()
	setErr(err)
}

// Start creates the tray icon + menu on a dedicated thread.
// Safe to call once; second call is a no-op returning nil.
func Start(c Callbacks) error {
	mu.Lock()
	if running {
		mu.Unlock()
		return nil
	}
	cb = c
	done = make(chan struct{})
	ready = make(chan error, 1)
	wndProc = windows.NewCallback(wndProcFn)
	mu.Unlock()

	go runLoop()

	return <-ready
}

// Stop removes the icon and ends the message loop. No-op when not running.
// Never blocks the caller: posts WM_CLOSE and detaches.
// Called from Wails OnShutdown; blocking there hangs quit.
func Stop() {
	mu.Lock()
	h := hwnd
	r := running
	mu.Unlock()
	if !r || h == 0 {
		return
	}
	_, _, _ = pPostMessage.Call(uintptr(h), uintptr(wmClose), 0, 0)
}

// Running reports whether the tray loop owns an icon right now.
// Exposed to the frontend probe so hide-vs-close is provable.
func Running() bool {
	mu.Lock()
	defer mu.Unlock()
	return running && hwnd != 0
}

func runLoop() {
	if err := initTray(); err != nil {
		mu.Lock()
		setErr(err)
		doneCh := done
		mu.Unlock()
		ready <- err
		close(doneCh)
		return
	}
	mu.Lock()
	setErr(nil)
	running = true
	mu.Unlock()
	ready <- nil

	var m msg
	for {
		r, _, _ := pGetMessage.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
		if r == 0 {
			break
		}
		_, _, _ = pTranslateMessage.Call(uintptr(unsafe.Pointer(&m)))
		_, _, _ = pDispatchMessage.Call(uintptr(unsafe.Pointer(&m)))
	}
	cleanup()

	mu.Lock()
	running = false
	hwnd = 0
	doneCh := done
	mu.Unlock()
	close(doneCh)
}

func initTray() error {
	if n := unsafe.Sizeof(notifyIconData{}); n != wantNIDSize {
		return errNIDSize(n)
	}
	h, _, _ := pGetModuleHandle.Call(0)
	hInst = windows.Handle(h)

	var err error
	classBuf, err = windows.UTF16PtrFromString("WortelAITray")
	if err != nil {
		return err
	}
	title, _ := windows.UTF16PtrFromString("Wortel AI")
	cur, _, _ := pLoadCursor.Call(0, uintptr(idcArrow))
	icon, _, _ := pLoadIcon.Call(0, uintptr(idiApp))

	wcx := wndClassEx{
		wndProc:   wndProc,
		instance:  hInst,
		cursor:    windows.Handle(cur),
		className: classBuf,
	}
	wcx.size = uint32(unsafe.Sizeof(wcx))
	r, _, _ := pRegisterClassEx.Call(uintptr(unsafe.Pointer(&wcx)))
	if r == 0 {
		return windows.GetLastError()
	}

	tm, _, _ := pRegWindowMsg.Call(uintptr(unsafe.Pointer(mustUTF16("TaskbarCreated"))))
	taskMsg = uint32(tm)

	hw, _, _ := pCreateWindowEx.Call(0,
		uintptr(unsafe.Pointer(classBuf)),
		uintptr(unsafe.Pointer(title)),
		0, 0, 0, 0, 0, 0, 0, uintptr(hInst), 0)
	if hw == 0 {
		_, _, _ = pUnregisterClass.Call(uintptr(unsafe.Pointer(classBuf)), uintptr(hInst))
		return windows.GetLastError()
	}
	hwnd = windows.Handle(hw)

	mn, _, _ := pCreatePopupMenu.Call()
	if mn == 0 {
		_, _, _ = pDestroyWindow.Call(hw)
		_, _, _ = pUnregisterClass.Call(uintptr(unsafe.Pointer(classBuf)), uintptr(hInst))
		return windows.GetLastError()
	}
	hMenu = windows.Handle(mn)
	addItem(menuShow, "Buka")
	addItem(menuSettings, "Settings")
	addItem(menuHide, "Sembunyikan")
	addSep()
	addItem(menuExit, "Keluar")

	tipBuf, err := windows.UTF16FromString("Wortel AI")
	if err != nil {
		return err
	}
	var tipArr [128]uint16
	copy(tipArr[:], tipBuf)

	nid = notifyIconData{
		wnd:             hwnd,
		id:              1,
		flags:           nifMessage | nifIcon | nifTip,
		callbackMessage: wmTray,
		icon:            windows.Handle(icon),
		tip:             tipArr,
	}
	nid.size = uint32(unsafe.Sizeof(nid))
	r, _, _ = pShellNotifyIcon.Call(uintptr(nimAdd), uintptr(unsafe.Pointer(&nid)))
	if r == 0 {
		_, _, _ = pDestroyMenu.Call(mn)
		_, _, _ = pDestroyWindow.Call(hw)
		_, _, _ = pUnregisterClass.Call(uintptr(unsafe.Pointer(classBuf)), uintptr(hInst))
		return windows.GetLastError()
	}
	return nil
}

func cleanup() {
	_, _, _ = pShellNotifyIcon.Call(uintptr(nimDelete), uintptr(unsafe.Pointer(&nid)))
	if hMenu != 0 {
		_, _, _ = pDestroyMenu.Call(uintptr(hMenu))
		hMenu = 0
	}
	if hwnd != 0 {
		_, _, _ = pDestroyWindow.Call(uintptr(hwnd))
	}
	_, _, _ = pUnregisterClass.Call(uintptr(unsafe.Pointer(classBuf)), uintptr(hInst))
}

func addItem(id uint32, label string) {
	s, _ := windows.UTF16PtrFromString(label)
	_, _, _ = pAppendMenu.Call(uintptr(hMenu), uintptr(mfString), uintptr(id), uintptr(unsafe.Pointer(s)))
}

func addSep() {
	_, _, _ = pAppendMenu.Call(uintptr(hMenu), uintptr(mfSep), 0, 0)
}

func mustUTF16(s string) *uint16 {
	p, _ := windows.UTF16PtrFromString(s)
	return p
}

func wndProcFn(h, m, w, l uintptr) uintptr {
	switch uint32(m) {
	case wmTray:
		switch uint32(l) & 0xFFFF {
		case wmRBtnUp:
			popup()
			return 0
		case wmLBtnUp, wmLBtnDbl:
			mu.Lock()
			c := cb
			mu.Unlock()
			c.dispatch(ActionShow)
			return 0
		}
		return 0
	case wmCommand:
		mu.Lock()
		c := cb
		mu.Unlock()
		c.dispatch(ActionForMenuID(uint32(w) & 0xFFFF))
		return 0
	case wmClose:
		_, _, _ = pDestroyWindow.Call(h)
		return 0
	case wmDestroy:
		_, _, _ = pPostQuitMessage.Call(0)
		return 0
	}
	if taskMsg != 0 && uint32(m) == taskMsg {
		// Explorer restarted: re-add icon.
		_, _, _ = pShellNotifyIcon.Call(uintptr(nimAdd), uintptr(unsafe.Pointer(&nid)))
		return 0
	}
	r, _, _ := pDefWindowProc.Call(h, m, w, l)
	return r
}

func popup() {
	var pt point
	_, _, _ = pGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))
	_, _, _ = pSetForeground.Call(uintptr(hwnd))
	r, _, _ := pTrackPopupMenu.Call(uintptr(hMenu), uintptr(tpmButtons),
		uintptr(pt.x), uintptr(pt.y), 0, uintptr(hwnd), 0)
	if r != 0 {
		mu.Lock()
		c := cb
		mu.Unlock()
		c.dispatch(ActionForMenuID(uint32(r)))
	}
	_, _, _ = pPostMessage.Call(uintptr(hwnd), uintptr(wmNull), 0, 0)
}
