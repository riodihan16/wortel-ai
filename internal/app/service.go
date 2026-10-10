package app

import (
	"context"
	"errors"
	"strings"
	"sync"

	"companion/internal/ai"
	"companion/internal/storage"

	"github.com/google/uuid"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// Service orchestrates one generation per conversation with SQLite history (M4).
type Service struct {
	mu       sync.Mutex
	active   map[string]context.CancelFunc // conversationID -> cancel
	requests map[string]string             // requestID -> conversationID
	provider ai.Provider                   // fallback (mock in dev)
	provs    *Providers                    // nil = fallback only
	emit     func(event string, payload ai.ChatEvent)
	wg       sync.WaitGroup // tracks running generation goroutines
}

func NewService(provider ai.Provider, emit func(string, ai.ChatEvent)) *Service {
	return &Service{
		active:   map[string]context.CancelFunc{},
		requests: map[string]string{},
		provider: provider,
		emit:     emit,
	}
}

// SetProviders wires the M3 provider manager. Live provider used when resolvable.
func (s *Service) SetProviders(p *Providers) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.provs = p
}

// db returns the SQLite handle via providers, nil when unwired (mock/test).
func (s *Service) db() *storage.DB {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.provs == nil {
		return nil
	}
	return s.provs.db
}

// resolve picks the live provider when M3 is wired, else the fallback.
func (s *Service) resolve(ctx context.Context) (ai.Provider, string, error) {
	s.mu.Lock()
	p := s.provs
	s.mu.Unlock()
	if p == nil {
		return s.provider, "", nil
	}
	live, model, err := p.Resolve(ctx)
	if err != nil {
		return nil, "", err
	}
	return live, model, nil
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
		return "", errors.New("empty conversation")
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

	s.wg.Add(1)
	go func() {
		defer func() {
			s.mu.Lock()
			delete(s.active, conversationID)
			delete(s.requests, requestID)
			s.mu.Unlock()
			s.wg.Done()
		}()
		live, model, resolveErr := s.resolve(genCtx)
		if resolveErr != nil {
			s.emit("chat:event", ai.ChatEvent{Type: "failed", RequestID: requestID, ConversationID: conversationID, MessageID: messageID, Error: &ai.PublicError{Code: "no_provider", Message: resolveErr.Error()}})
			return
		}
		db := s.db()
		userMsgID := uuid.NewString()
		if db != nil {
			_ = db.EnsureConversation(conversationID, storage.TitleFor(text))
			_ = db.AppendMessage(storage.MessageRow{ID: userMsgID, ConversationID: conversationID, Role: "user", Content: text, Status: "complete", Model: model})
			_ = db.AppendMessage(storage.MessageRow{ID: messageID, ConversationID: conversationID, Role: "assistant", Content: "", Status: "streaming", Model: model})
		}
		msgs := []ai.Message{{Role: "user", Content: text}}
		if db != nil {
			if hist, err := db.RecentMessages(conversationID, 20); err == nil && len(hist) > 0 {
				// hist already includes the just-saved user msg; drop dup tail then cap.
				msgs = hist
				if len(msgs) > 20 {
					msgs = msgs[len(msgs)-20:]
				}
			}
		}
		var sb strings.Builder
		err := live.StreamChat(genCtx, ai.ChatRequest{Model: model, Messages: msgs}, func(chunk string) error {
			sb.WriteString(chunk)
			s.emit("chat:event", ai.ChatEvent{Type: "delta", RequestID: requestID, ConversationID: conversationID, MessageID: messageID, Delta: chunk})
			return nil
		})
		final := sb.String()
		switch {
		case err == nil:
			if db != nil {
				_ = db.UpdateMessage(messageID, final, "complete")
			}
			s.emit("chat:event", ai.ChatEvent{Type: "completed", RequestID: requestID, ConversationID: conversationID, MessageID: messageID})
		case errors.Is(err, context.Canceled):
			if db != nil {
				_ = db.UpdateMessage(messageID, final, "cancelled")
			}
			s.emit("chat:event", ai.ChatEvent{Type: "cancelled", RequestID: requestID, ConversationID: conversationID, MessageID: messageID})
		default:
			if db != nil {
				_ = db.UpdateMessage(messageID, final, "failed")
			}
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
		return
	}
	s.mu.Lock()
	cancel, ok := s.active[conversationID]
	s.mu.Unlock()
	if ok {
		cancel()
	}
}

// CancelAll stops every active stream and waits for their goroutines to
// finish writing history. Called on shutdown so no goroutine/request lingers
// after exit (FR-011) and no DB write races process teardown.
func (s *Service) CancelAll() {
	s.mu.Lock()
	cancels := make([]context.CancelFunc, 0, len(s.active))
	for _, cancel := range s.active {
		cancels = append(cancels, cancel)
	}
	s.mu.Unlock()
	for _, cancel := range cancels {
		cancel()
	}
	s.wg.Wait()
}
