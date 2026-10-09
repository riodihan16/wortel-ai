package app

// ChatMessage is the frontend-safe history item.
type ChatMessage struct {
	ID      string `json:"id"`
	Role    string `json:"role"`
	Content string `json:"content"`
	Status  string `json:"status"`
}

// History returns recent messages oldest-first for a conversation.
// Empty slice when unwired or none. Never includes secrets.
func (s *Service) History(conversationID string, n int) []ChatMessage {
	db := s.db()
	if db == nil {
		return []ChatMessage{}
	}
	if conversationID == "" {
		conversationID = "default"
	}
	rows, err := db.RecentRows(conversationID, n)
	if err != nil {
		return []ChatMessage{}
	}
	out := make([]ChatMessage, 0, len(rows))
	for _, r := range rows {
		if r.Role != "user" && r.Role != "assistant" {
			continue
		}
		out = append(out, ChatMessage{
			ID: r.ID, Role: r.Role, Content: r.Content, Status: r.Status,
		})
	}
	return out
}
