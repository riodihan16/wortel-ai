package app

import (
	"context"
	"errors"
	"strings"
	"sync"

	"companion/internal/ai"

	"github.com/google/uuid"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// Service orchestrates one generation per conversation. No DB yet (M4).
type Service struct {
	mu       sync.Mutex
	active   map[string]context.CancelFunc // conversationID -> cancel
	requests map[string]string             // requestID -> conversationID
	provider ai.Provider
	emit     func(event string, payload ai.ChatEvent)
}

func NewService(provider ai.Provider, emit func(string, ai.ChatEvent)) *Service {
	return &Service{
		active:   map[string]context.CancelFunc{},
		requests: map[string]string{},
		provider: provider,
		emit:     emit,
	}
}

func NewRuntimeService(ctx context.Context, provider ai.Provider) *Service {
	return NewService(provider, func(ev string, p ai.ChatEvent) {
		runtime.EventsEmit(ctx, ev, p)
	})
}

// SendMessage starts streaming. Empty input rejected. One active per conversation.
func (s *Service) SendMessage(ctx context.Context, conversationID, text string) (string, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return "", errors.New("empty message")
	}
	if conversationID == "" {
		conversationID = "default"
	}
	s.mu.Lock()
	if _, ok := s.active[conversationID]; ok {
		s.mu.Unlock()
		return "", errors.New("generation already active")
	}
	genCtx, cancel := context.WithCancel(ctx)
	s.active[conversationID] = cancel
	s.mu.Unlock()

	requestID := uuid.NewString()
	messageID := uuid.NewString()
	s.mu.Lock()
	s.requests[requestID] = conversationID
	s.mu.Unlock()

	s.emit("chat:event", ai.ChatEvent{Type: "started", RequestID: requestID, ConversationID: conversationID, MessageID: messageID})

	go func() {
		defer func() {
			s.mu.Lock()
			delete(s.active, conversationID)
			s.mu.Unlock()
		}()
		err := s.provider.StreamChat(genCtx, ai.ChatRequest{Messages: []ai.Message{{Role: "user", Content: text}}}, func(chunk string) error {
			s.emit("chat:event", ai.ChatEvent{Type: "delta", RequestID: requestID, ConversationID: conversationID, MessageID: messageID, Delta: chunk})
			return nil
		})
		switch {
		case err == nil:
			s.emit("chat:event", ai.ChatEvent{Type: "completed", RequestID: requestID, ConversationID: conversationID, MessageID: messageID})
		case errors.Is(err, context.Canceled):
			s.emit("chat:event", ai.ChatEvent{Type: "cancelled", RequestID: requestID, ConversationID: conversationID, MessageID: messageID})
		default:
			var pub *ai.PublicError
			if errors.As(err, &pub) {
				s.emit("chat:event", ai.ChatEvent{Type: "failed", RequestID: requestID, ConversationID: conversationID, MessageID: messageID, Error: pub})
				return
			}
			s.emit("chat:event", ai.ChatEvent{Type: "failed", RequestID: requestID, ConversationID: conversationID, MessageID: messageID, Error: &ai.PublicError{Code: "network", Message: "Koneksi gagal. Coba lagi.", Retryable: true}})
		}
	}()
	return requestID, nil
}

// CancelGeneration stops the active stream for a conversation.
func (s *Service) CancelGeneration(conversationID string) {
	if conversationID == "" {
		conversationID = "default"
	}
	s.mu.Lock()
	cancel, ok := s.active[conversationID]
	s.mu.Unlock()
	if ok {
		cancel()
	}
}
