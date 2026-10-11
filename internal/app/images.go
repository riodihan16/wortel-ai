package app

import (
	"context"
	"errors"
	"strings"
	"time"

	"companion/internal/ai"

	"github.com/google/uuid"
)

// imageTimeout is the generation budget for one image. Kenari docs §22.4
// call out image generation as needing a longer timeout than chat (the
// curated example allows up to 5 minutes), so it is set explicitly here and
// never inherited from the chat client's short timeout.
const imageTimeout = 5 * time.Minute

// ImageEvent is the structured event forwarded to the frontend for one image
// generation. Distinct from ai.ChatEvent so the WebView can route it. Never
// carries the API key.
type ImageEvent struct {
	Type           string          `json:"type"` // started|completed|cancelled|failed
	RequestID      string          `json:"requestId"`
	ConversationID string          `json:"conversationId"`
	MessageID      string          `json:"messageId"`
	Mime           string          `json:"mime,omitempty"`
	Prompt         string          `json:"prompt,omitempty"`
	Error          *ai.PublicError `json:"error,omitempty"`
}

// GenerateImage starts a non-streaming image generation. One active image
// generation per conversation, tracked in a cancel map separate from chat's
// so the two never collide. Returns the requestID.
func (s *Service) GenerateImage(ctx context.Context, conversationID, prompt, model string) (string, error) {
	prompt = strings.TrimSpace(prompt)
	if prompt == "" {
		return "", errors.New("prompt kosong")
	}
	if conversationID == "" {
		return "", errors.New("empty conversation")
	}
	if s.db() == nil {
		return "", errNoDB
	}
	s.mu.Lock()
	if _, ok := s.imageActive[conversationID]; ok {
		s.mu.Unlock()
		return "", errors.New("pembuatan gambar sudah berjalan")
	}
	genCtx, cancel := context.WithTimeout(ctx, imageTimeout)
	s.imageActive[conversationID] = cancel
	s.mu.Unlock()

	requestID := uuid.NewString()
	s.wg.Add(1)
	go func() {
		defer func() {
			cancel()
			s.mu.Lock()
			delete(s.imageActive, conversationID)
			s.mu.Unlock()
			s.wg.Done()
		}()
		fail := func(pub *ai.PublicError) {
			s.emitImage(ImageEvent{Type: "failed", RequestID: requestID, ConversationID: conversationID, Prompt: prompt, Error: pub})
		}
		db := s.db()
		live, model2, resolveErr := s.resolveImage(genCtx)
		if resolveErr != nil {
			fail(&ai.PublicError{Code: "no_provider", Message: resolveErr.Error()})
			return
		}
		if model2 != "" {
			model = model2
		}
		// The conversation row must exist for the message FK; a fresh chat
		// that never sent text would otherwise fail the insert.
		_ = db.EnsureConversation(conversationID, titleFor(prompt, nil))
		messageID, err := db.AppendImageMessage(conversationID, model)
		if err != nil {
			fail(&ai.PublicError{Code: "storage", Message: "Gagal menyimpan gambar."})
			return
		}
		s.emitImage(ImageEvent{Type: "started", RequestID: requestID, ConversationID: conversationID, MessageID: messageID, Prompt: prompt})

		img, ok := live.(*ai.OpenAICompatible)
		if !ok {
			_ = db.MarkImageFailed(messageID)
			fail(&ai.PublicError{Code: "no_provider", Message: "Provider gambar tidak tersedia."})
			return
		}
		mime, data, err := img.GenerateImage(genCtx, ai.ImageRequest{Model: model, Prompt: prompt, Size: "1024x1024", N: 1})
		if err != nil {
			_ = db.MarkImageFailed(messageID)
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				s.emitImage(ImageEvent{Type: "cancelled", RequestID: requestID, ConversationID: conversationID, MessageID: messageID})
				return
			}
			var pub *ai.PublicError
			if errors.As(err, &pub) {
				fail(pub)
				return
			}
			fail(&ai.PublicError{Code: "network", Message: "Koneksi gagal. Coba lagi.", Retryable: true})
			return
		}
		if err := db.SetMessageImage(messageID, mime, data); err != nil {
			_ = db.MarkImageFailed(messageID)
			fail(&ai.PublicError{Code: "storage", Message: "Gagal menyimpan gambar."})
			return
		}
		s.emitImage(ImageEvent{Type: "completed", RequestID: requestID, ConversationID: conversationID, MessageID: messageID, Mime: mime, Prompt: prompt})
	}()
	return requestID, nil
}

// CancelImage stops the active image generation for a conversation.
func (s *Service) CancelImage(conversationID string) {
	if conversationID == "" {
		return
	}
	s.mu.Lock()
	cancel, ok := s.imageActive[conversationID]
	s.mu.Unlock()
	if ok {
		cancel()
	}
}

func (s *Service) emitImage(ev ImageEvent) {
	// Reuse the same emitter surface; the App wires it to Wails. The event
	// name is distinct from chat:event.
	s.mu.Lock()
	emit := s.emitRaw
	s.mu.Unlock()
	if emit != nil {
		emit("image:event", ev)
	}
}
