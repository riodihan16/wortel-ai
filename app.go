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

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx       context.Context
	svc       *app.Service
	provs     *app.Providers
	hotkeyErr string

	mu     sync.Mutex
	hidden bool
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
	// M5 shortcut global (FR-010). Failure is non-fatal and reported
	// to the user via HotkeyError so they can pick another combo.
	if err := hotkey.Start(hotkey.Default, func() { a.ToggleChat() }); err != nil {
		a.hotkeyErr = err.Error()
	}
}

// ToggleChat shows the window in chat mode, or hides it when already shown.
// Visibility is tracked in the App because Wails exposes no IsVisible probe.
func (a *App) ToggleChat() {
	a.mu.Lock()
	hidden := a.hidden
	a.hidden = !hidden
	a.mu.Unlock()
	if hidden {
		runtime.WindowShow(a.ctx)
		runtime.EventsEmit(a.ctx, "app:toggle")
		return
	}
	runtime.WindowHide(a.ctx)
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

func (a *App) setHidden(v bool) {
	a.mu.Lock()
	a.hidden = v
	a.mu.Unlock()
}

// HotkeyError returns the global-shortcut registration error, empty when OK.
// Non-empty means the default combo was taken; the user must be told (FR-010).
func (a *App) HotkeyError() string {
	return a.hotkeyErr
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
