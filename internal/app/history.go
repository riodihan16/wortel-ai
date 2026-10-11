package app

import (
	"errors"

	"companion/internal/attach"
	"companion/internal/storage"
)

var (
	errNoDB             = errors.New("storage belum tersedia, restart aplikasi")
	errNoConversation   = errors.New("percakapan tidak ditemukan")
	errConversationBusy = errors.New("percakapan sedang streaming: hentikan dulu sebelum menghapus")
)

// ChatMessage is the frontend-safe history item.
type ChatMessage struct {
	ID          string        `json:"id"`
	Role        string        `json:"role"`
	Content     string        `json:"content"`
	Status      string        `json:"status"`
	Attachments []attach.Meta `json:"attachments"`
	// ImageMime is set on an assistant row that carries a generated image;
	// the bytes are served via /attachments/{id}.
	ImageMime string `json:"imageMime,omitempty"`
}

// History returns recent messages oldest-first for a conversation.
// Empty slice when unwired or none. Never includes secrets.
func (s *Service) History(conversationID string, n int) []ChatMessage {
	db := s.db()
	if db == nil {
		return []ChatMessage{}
	}
	if conversationID == "" {
		return []ChatMessage{}
	}
	rows, err := db.RecentRows(conversationID, n)
	if err != nil {
		return []ChatMessage{}
	}
	ids := make([]string, 0, len(rows))
	for _, r := range rows {
		ids = append(ids, r.ID)
	}
	atts, err := db.AttachmentsFor(ids)
	if err != nil {
		atts = map[string][]storage.AttachmentRow{}
	}
	out := make([]ChatMessage, 0, len(rows))
	for _, r := range rows {
		if r.Role != "user" && r.Role != "assistant" {
			continue
		}
		out = append(out, ChatMessage{
			ID: r.ID, Role: r.Role, Content: r.Content, Status: r.Status,
			Attachments: metasOf(atts[r.ID]),
			ImageMime:   r.ImageMime,
		})
	}
	return out
}

// Conversation is the frontend-safe conversation item.
type Conversation struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

// ListConversations returns id+title ordered by recent activity.
// Empty slice when unwired. Never includes message content.
func (s *Service) ListConversations() []Conversation {
	db := s.db()
	if db == nil {
		return []Conversation{}
	}
	rows, err := db.ListConversations()
	if err != nil {
		return []Conversation{}
	}
	out := make([]Conversation, 0, len(rows))
	for _, r := range rows {
		out = append(out, Conversation{ID: r.ID, Title: r.Title})
	}
	return out
}

// RenameConversation sets the title. Empty title rejected in storage.
func (s *Service) RenameConversation(conversationID, title string) error {
	db := s.db()
	if db == nil {
		return errNoDB
	}
	if conversationID == "" {
		return errNoConversation
	}
	return db.RenameConversation(conversationID, title)
}

// DeleteConversation removes a conversation and its messages.
// Refuses when a generation is active on it.
func (s *Service) DeleteConversation(conversationID string) error {
	db := s.db()
	if db == nil {
		return errNoDB
	}
	if conversationID == "" {
		return errNoConversation
	}
	s.mu.Lock()
	_, active := s.active[conversationID]
	_, imgActive := s.imageActive[conversationID]
	s.mu.Unlock()
	if active || imgActive {
		return errConversationBusy
	}
	return db.DeleteConversation(conversationID)
}
