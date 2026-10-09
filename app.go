package main

import (
	"context"
	"errors"

	"companion/internal/ai"
	"companion/internal/app"
	"companion/internal/config"
	"companion/internal/credentials"
	"companion/internal/storage"
)

// App struct
type App struct {
	ctx   context.Context
	svc   *app.Service
	provs *app.Providers
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
