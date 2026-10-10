package main

import (
	"context"
	"errors"
	"os"
	"sync"

	"companion/internal/ai"
	"companion/internal/app"
	"companion/internal/autostart"
	"companion/internal/config"
	"companion/internal/credentials"
	"companion/internal/hotkey"
	"companion/internal/storage"
	"companion/internal/tray"
	"companion/internal/window"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx       context.Context
	svc       *app.Service
	provs     *app.Providers
	hotkeyErr string
	version   string

	mu     sync.Mutex
	hidden bool
	mode   string // last SetWindowMode value; "" until the frontend sets one
}

type toggleAction int

const (
	toggleShowChat toggleAction = iota // show the window if needed, then chat
	toggleToPet                        // collapse the open panel back to the pet
)

// toggleFor decides what the global shortcut does. It is labelled "buka
// chat": a hidden window or the pet opens chat; an open panel (chat or
// Settings) goes back to the pet. Hiding is left to the explicit Sembunyikan
// actions, so the shortcut never makes the companion disappear (ADR-050).
func toggleFor(hidden bool, mode string) toggleAction {
	if hidden || mode != string(window.Panel) {
		return toggleShowChat
	}
	return toggleToPet
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.svc = app.NewRuntimeService(ctx, &ai.MockProvider{})
	// M3 storage + credential wiring. Failures fall back to mock-only.
	if path, err := storage.DefaultPath(); err == nil {
		if db, err := storage.Open(path); err == nil {
			a.provs = app.NewProviders(db, credentials.OSStore{})
			a.svc.SetProviders(a.provs)
		}
	}
	// M5 tray. Failure is non-fatal: app keeps running without tray.
	// Error stored and exposed via TrayError so Probe shows the cause.
	if err := tray.Start(tray.Callbacks{
		OnShow:     func() { a.setHidden(false); runtime.WindowShow(ctx); runtime.EventsEmit(ctx, "app:show") },
		OnSettings: func() { a.setHidden(false); runtime.WindowShow(ctx); runtime.EventsEmit(ctx, "app:settings") },
		OnHide:     func() { a.setHidden(true); runtime.WindowHide(ctx) },
		OnExit:     func() { runtime.Quit(ctx) },
	}); err != nil {
		tray.SetError(err)
	}
	// Edge guard: user drags stop at the monitor work area. Non-fatal; the
	// window still drags, only unconstrained.
	if err := window.KeepOnScreen(); err != nil {
		runtime.LogWarningf(ctx, "edge guard off: %v", err)
	}
	// M5 shortcut global (FR-010). Uses the stored combo when present.
	// Failure is non-fatal and reported to the user via HotkeyError so
	// they can pick another combo in Settings.
	a.applyShortcut()
}

// applyShortcut registers the configured global shortcut, replacing any
// existing registration. Never fatal: failures surface through HotkeyError.
func (a *App) applyShortcut() {
	hotkey.Stop()
	a.mu.Lock()
	a.hotkeyErr = ""
	a.mu.Unlock()
	if a.provs == nil {
		return
	}
	h, enabled := a.provs.ResolvedShortcut()
	if !enabled {
		return
	}
	if err := hotkey.Start(h, func() { a.ToggleChat() }); err != nil {
		a.mu.Lock()
		a.hotkeyErr = err.Error()
		a.mu.Unlock()
	}
}

// ToggleChat runs the global shortcut (see toggleFor). Visibility and mode
// are tracked in the App because Wails exposes no IsVisible probe.
func (a *App) ToggleChat() {
	a.mu.Lock()
	hidden, mode := a.hidden, a.mode
	a.hidden = false
	a.mu.Unlock()
	switch toggleFor(hidden, mode) {
	case toggleToPet:
		runtime.EventsEmit(a.ctx, "app:pet")
	default:
		if hidden {
			runtime.WindowShow(a.ctx)
		}
		runtime.EventsEmit(a.ctx, "app:toggle")
	}
}

// shutdown cancels active streams, removes the tray icon, and releases
// the global shortcut. Runs synchronously: no goroutine lingers (FR-011).
func (a *App) shutdown(ctx context.Context) {
	if a.svc != nil {
		a.svc.CancelAll()
	}
	hotkey.Stop()
	tray.Stop()
}

// Show brings the window up from tray-hidden state.
func (a *App) Show() {
	a.setHidden(false)
	runtime.WindowShow(a.ctx)
}

// HideWindow sends the window to the tray (M5 hide to tray).
func (a *App) HideWindow() {
	a.setHidden(true)
	runtime.WindowHide(a.ctx)
}

// SetWindowMode sizes the window for "pet" or "panel" (chat/settings) and
// keeps it inside its monitor's work area, so a pet parked near an edge
// opens a panel that is not cut off.
func (a *App) SetWindowMode(mode string) error {
	m := window.Mode(mode)
	size, err := window.SizeOf(m)
	if err != nil {
		return err
	}
	if err := window.Apply(m); err != nil {
		// Placement failed (no handle/monitor info): degrade to the old
		// plain resize so the mode switch itself still works.
		runtime.LogWarningf(a.ctx, "window place: %v", err)
		runtime.WindowSetSize(a.ctx, size.W, size.H)
	}
	a.mu.Lock()
	a.mode = mode
	a.mu.Unlock()
	return nil
}

func (a *App) setHidden(v bool) {
	a.mu.Lock()
	a.hidden = v
	a.mu.Unlock()
}

// Version returns the build version stamped at link time ("dev" otherwise).
// Shown in Settings so users and support can identify a build (release checklist).
func (a *App) Version() string {
	if a.version == "" {
		return "dev"
	}
	return a.version
}

// HotkeyError returns the global-shortcut registration error, empty when OK.
// Non-empty means the combo was taken or invalid; the user must be told (FR-010).
func (a *App) HotkeyError() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.hotkeyErr
}

// GetShortcut returns the saved shortcut setting (enabled flag + combo).
func (a *App) GetShortcut() app.Shortcut {
	if a.provs == nil {
		return app.DefaultShortcut()
	}
	return a.provs.GetShortcut()
}

// SetShortcut validates, persists, and re-registers the global shortcut.
// A combo the OS refuses returns an error and leaves the previous binding
// itself; the caller sees the new HotkeyError value too.
func (a *App) SetShortcut(in app.Shortcut) error {
	if a.provs == nil {
		return errors.New("storage belum tersedia, restart aplikasi")
	}
	if err := a.provs.SetShortcut(in); err != nil {
		return err
	}
	a.applyShortcut()
	return nil
}

// GetAppearance returns stored theme + font scale, or defaults when unwired.
func (a *App) GetAppearance() app.Appearance {
	if a.provs == nil {
		return app.DefaultAppearance()
	}
	return a.provs.GetAppearance()
}

// SetAppearance validates and stores appearance settings.
// Group "Appearance" docs 06 §1; contrast + scrollability docs 06 §6.
func (a *App) SetAppearance(in app.Appearance) error {
	if a.provs == nil {
		return errors.New("storage belum tersedia, restart aplikasi")
	}
	return a.provs.SetAppearance(in)
}

// TrayStatus reports whether the tray icon loop owns an icon.
// Probe answers hide-vs-close: true = app alive in tray.
func (a *App) TrayStatus() bool {
	return tray.Running()
}

// TrayError returns the last tray init failure, empty when none.
func (a *App) TrayError() string {
	return tray.LastError()
}

// History returns recent messages oldest-first. Empty when none.
func (a *App) History(conversationID string, n int) []app.ChatMessage {
	if a.svc == nil {
		return []app.ChatMessage{}
	}
	return a.svc.History(conversationID, n)
}

// ListConversations returns id+title ordered by recent activity.
func (a *App) ListConversations() []app.Conversation {
	if a.svc == nil {
		return []app.Conversation{}
	}
	return a.svc.ListConversations()
}

// RenameConversation sets the conversation title.
func (a *App) RenameConversation(conversationID string, title string) error {
	if a.svc == nil {
		return errStorageUnavailable()
	}
	return a.svc.RenameConversation(conversationID, title)
}

// DeleteConversation removes a conversation and its messages.
// Refuses when a generation is active on it.
func (a *App) DeleteConversation(conversationID string) error {
	if a.svc == nil {
		return errStorageUnavailable()
	}
	return a.svc.DeleteConversation(conversationID)
}

// SendMessage starts a streaming generation. Returns requestId.
func (a *App) SendMessage(conversationID string, text string) (string, error) {
	return a.svc.SendMessage(a.ctx, conversationID, text)
}

// RetryMessage regenerates the newest turn after a failure without storing
// the user's text a second time.
func (a *App) RetryMessage(conversationID string, text string) (string, error) {
	return a.svc.RetryMessage(a.ctx, conversationID, text)
}

// CancelGeneration stops the active stream for a conversation.
func (a *App) CancelGeneration(conversationID string) {
	a.svc.CancelGeneration(conversationID)
}

// ListProviders returns frontend-safe profiles. Never includes secrets.
func (a *App) ListProviders() ([]config.ProviderProfile, error) {
	if a.provs == nil {
		return []config.ProviderProfile{}, nil
	}
	return a.provs.List(a.ctx)
}

// SaveProvider validates, stores the key in the OS credential store,
// and persists the profile. API key never returned.
func (a *App) SaveProvider(in app.SaveInput) (config.ProviderProfile, error) {
	if a.provs == nil {
		return config.ProviderProfile{}, errStorageUnavailable()
	}
	return a.provs.Save(a.ctx, in)
}

// DeleteProvider removes a profile and its credential. Refuses when active.
func (a *App) DeleteProvider(id string) error {
	if a.provs == nil {
		return errStorageUnavailable()
	}
	return a.provs.Delete(a.ctx, id)
}

// SetActiveProvider selects the provider used for chat.
func (a *App) SetActiveProvider(id string) error {
	if a.provs == nil {
		return errStorageUnavailable()
	}
	return a.provs.SetActive(id)
}

// SetActiveModel switches the chat model on the active provider. The model
// must already be in that provider's list.
func (a *App) SetActiveModel(id, model string) error {
	if a.provs == nil {
		return errStorageUnavailable()
	}
	return a.provs.SetActiveModel(id, model)
}

// ActiveProvider returns the active provider's safe profile (models list,
// active model), or null when no provider is active.
//
// Wails drops a second non-error return value (its binder treats output 2 as
// the error slot), so a (config.ProviderProfile, bool) tuple reaches the
// frontend as a bare struct and a zero profile cannot be told from a real
// one. Use a pointer instead: nil marshals to JSON null.
func (a *App) ActiveProvider() *config.ProviderProfile {
	if a.provs == nil {
		return nil
	}
	profile, ok := a.provs.ActiveProfile(a.ctx)
	if !ok {
		return nil
	}
	return &profile
}

// TestProvider checks URL/model/credential with a minimal request.
func (a *App) TestProvider(in app.SaveInput) (app.TestResult, error) {
	if a.provs == nil {
		return app.TestResult{}, errStorageUnavailable()
	}
	return a.provs.Test(a.ctx, in), nil
}

func errStorageUnavailable() error { return errors.New("storage belum tersedia, restart aplikasi") }

// AutostartEnabled reports whether the app starts with Windows.
// Default is off; the user opts in from Settings (FR-011).
func (a *App) AutostartEnabled() bool {
	return autostart.Enabled()
}

// SetAutostart turns "run at login" on or off. Uses the per-user Run key;
// no admin rights and no secret involved.
func (a *App) SetAutostart(on bool) error {
	if !on {
		return autostart.Disable()
	}
	exe, err := os.Executable()
	if err != nil {
		return errors.New("tidak dapat menemukan path aplikasi")
	}
	return autostart.Enable(exe)
}
