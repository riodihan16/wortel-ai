package app

import (
	"context"
	"errors"
	"strings"
	"sync"

	"companion/internal/ai"
	"companion/internal/attach"
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
	staging  *attach.Staging
}

func NewService(provider ai.Provider, emit func(string, ai.ChatEvent)) *Service {
	return &Service{
		active:   map[string]context.CancelFunc{},
		requests: map[string]string{},
		provider: provider,
		emit:     emit,
		staging:  attach.NewStaging(),
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

// SendMessage starts streaming. Text may be empty when attachments are
// given; both empty is rejected. One active generation per conversation.
func (s *Service) SendMessage(ctx context.Context, conversationID, text string, attachmentIDs []string) (string, error) {
	return s.generate(ctx, conversationID, text, false, attachmentIDs)
}

// RetryMessage regenerates the newest turn after a failure. Unlike
// SendMessage it does not store the user's text again when that turn is
// already in history; the failed reply is replaced (storage.PrepareRetry)
// and the turn's stored attachments are sent again from the database.
func (s *Service) RetryMessage(ctx context.Context, conversationID, text string) (string, error) {
	return s.generate(ctx, conversationID, text, true, nil)
}

func (s *Service) generate(ctx context.Context, conversationID, text string, retry bool, attachmentIDs []string) (string, error) {
	text = strings.TrimSpace(text)
	// A retry may carry empty text (an attachment-only turn): with a DB,
	// PrepareRetry decides from the stored turn; without one there is
	// nothing to reuse, so reject as a plain send would.
	if text == "" && len(attachmentIDs) == 0 && (!retry || s.db() == nil) {
		return "", errors.New("empty message")
	}
	if conversationID == "" {
		return "", errors.New("empty conversation")
	}
	if len(attachmentIDs) > 0 && s.db() == nil {
		return "", errNoAttachmentStorage
	}
	s.mu.Lock()
	if _, ok := s.active[conversationID]; ok {
		s.mu.Unlock()
		return "", errors.New("generation already active")
	}
	// Attachments are resolved before the goroutine so a bad id is a
	// synchronous rejection, never a half-sent turn. The items hold the bytes
	// from here on; staging is cleared now so the composer can start fresh.
	var items []attach.Item
	if len(attachmentIDs) > 0 {
		var err error
		if items, err = s.staging.Select(attachmentIDs); err != nil {
			s.mu.Unlock()
			return "", err
		}
		s.staging.Remove(attachmentIDs...)
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
		fail := func(pub *ai.PublicError) {
			s.emit("chat:event", ai.ChatEvent{Type: "failed", RequestID: requestID, ConversationID: conversationID, MessageID: messageID, Error: pub})
		}
		db := s.db()
		// Reconcile a retry before anything can fail, so history and the UI
		// (which drops the failed bubble once the retry is accepted) agree.
		// On a DB error, assume the user row is there: appending it again
		// would duplicate the question, which is the bug retry exists to avoid.
		reused := false
		if retry && db != nil {
			ok, err := db.PrepareRetry(conversationID, text)
			reused = ok || err != nil
		}
		if !reused && text == "" && len(items) == 0 {
			fail(&ai.PublicError{Code: "empty", Message: "Tidak ada yang bisa diulang."})
			return
		}
		live, model, resolveErr := s.resolve(genCtx)
		userMsgID := uuid.NewString()
		current := ai.Message{Role: "user", Content: text, Parts: partsFromItems(items)}
		if db != nil {
			// Persistence is best-effort by design: a DB failure must not abort
			// an in-flight stream. The error surfaces later when the history
			// list is empty after restart; see ADR-037.
			_ = db.EnsureConversation(conversationID, titleFor(text, items))
			assistant := storage.MessageRow{ID: messageID, ConversationID: conversationID, Role: "assistant", Content: "", Status: "streaming", Model: model}
			if reused {
				_ = db.AppendMessage(assistant)
			} else {
				user := storage.MessageRow{ID: userMsgID, ConversationID: conversationID, Role: "user", Content: text, Status: "complete", Model: model}
				_ = db.AppendUserTurn(user, attachmentRows(userMsgID, items), assistant)
			}
		}
		if resolveErr != nil {
			// The turn (and its attachments) is already stored, so a retry
			// can reuse it once a provider is configured.
			if db != nil {
				_ = db.UpdateMessage(messageID, "", "failed")
			}
			fail(&ai.PublicError{Code: "no_provider", Message: resolveErr.Error()})
			return
		}
		msgs := []ai.Message{current}
		if db != nil {
			if rows, err := db.RecentContext(conversationID, 20); err == nil && len(rows) > 0 {
				msgs = buildContext(rows, AttachmentTurns, db.AttachmentData)
				// A failed insert must not drop the question itself.
				if !reused && !hasRow(rows, userMsgID) {
					msgs = append(msgs, current)
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
				fail(pub)
				return
			}
			fail(&ai.PublicError{Code: "network", Message: "Koneksi gagal. Coba lagi.", Retryable: true})
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
	// Shutdown: pending attachments are not persisted.
	s.staging.Clear()
}
