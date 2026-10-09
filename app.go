package main

import (
	"context"

	"companion/internal/ai"
	"companion/internal/app"
)

// App struct
type App struct {
	ctx context.Context
	svc *app.Service
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
}

// SendMessage starts a mock streaming generation. Returns requestId.
func (a *App) SendMessage(conversationID string, text string) (string, error) {
	return a.svc.SendMessage(a.ctx, conversationID, text)
}

// CancelGeneration stops the active stream for a conversation.
func (a *App) CancelGeneration(conversationID string) {
	a.svc.CancelGeneration(conversationID)
}
