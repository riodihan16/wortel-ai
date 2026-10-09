package storage

import (
	"time"

	"companion/internal/ai"
)

// MessageRow maps the messages table.
type MessageRow struct {
	ID             string
	ConversationID string
	Role           string
	Content        string
	Status         string
	ProviderID     string
	Model          string
	CreatedAt      time.Time
}

// EnsureConversation creates the row when absent. Title from first text.
func (d *DB) EnsureConversation(id, title string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	if title == "" {
		title = "Percakapan"
	}
	_, err := d.inner.Exec(`INSERT INTO conversations (id, title, created_at, updated_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET updated_at=excluded.updated_at`,
		id, title, now, now)
	return err
}

// AppendMessage inserts one message with complete metadata.
func (d *DB) AppendMessage(m MessageRow) error {
	now := time.Now().UTC().Format(time.RFC3339)
	if m.CreatedAt.IsZero() {
		m.CreatedAt = time.Now().UTC()
	}
	_, err := d.inner.Exec(`INSERT INTO messages
		(id, conversation_id, role, content, status, provider_id, model, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		m.ID, m.ConversationID, m.Role, m.Content, m.Status,
		nullStr(m.ProviderID), nullStr(m.Model),
		m.CreatedAt.Format(time.RFC3339), now)
	return err
}

// UpdateMessage sets content + status after stream ends.
func (d *DB) UpdateMessage(id, content, status string) error {
	_, err := d.inner.Exec(`UPDATE messages SET content=?, status=?,
		updated_at=? WHERE id=?`,
		content, status, time.Now().UTC().Format(time.RFC3339), id)
	return err
}

// RecentMessages returns the last N messages oldest-first for context.
func (d *DB) RecentMessages(conversationID string, n int) ([]ai.Message, error) {
	if n <= 0 {
		n = 20
	}
	rows, err := d.inner.Query(`SELECT role, content FROM messages
		WHERE conversation_id=? AND status IN ('complete','streaming')
		ORDER BY created_at DESC LIMIT ?`, conversationID, n)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var rev []ai.Message
	for rows.Next() {
		var role, content string
		if err := rows.Scan(&role, &content); err != nil {
			return nil, err
		}
		if role != "user" && role != "assistant" && role != "system" {
			continue
		}
		rev = append(rev, ai.Message{Role: role, Content: content})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	// Reverse to oldest-first.
	out := make([]ai.Message, 0, len(rev))
	for i := len(rev) - 1; i >= 0; i-- {
		out = append(out, rev[i])
	}
	return out, nil
}

// RecentRows returns full rows oldest-first for history display.
func (d *DB) RecentRows(conversationID string, n int) ([]MessageRow, error) {
	if n <= 0 {
		n = 100
	}
	rows, err := d.inner.Query(`SELECT id, conversation_id, role, content, status,
		COALESCE(provider_id,''), COALESCE(model,''), created_at
		FROM messages WHERE conversation_id=?
		ORDER BY created_at DESC LIMIT ?`, conversationID, n)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var rev []MessageRow
	for rows.Next() {
		var m MessageRow
		var created string
		if err := rows.Scan(&m.ID, &m.ConversationID, &m.Role, &m.Content,
			&m.Status, &m.ProviderID, &m.Model, &created); err != nil {
			return nil, err
		}
		rev = append(rev, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	out := make([]MessageRow, 0, len(rev))
	for i := len(rev) - 1; i >= 0; i-- {
		out = append(out, rev[i])
	}
	return out, nil
}

// TitleFor makes a local title from the first user text (no extra API cost).
func TitleFor(text string) string {
	r := []rune(text)
	if len(r) > 48 {
		r = r[:48]
	}
	return string(r)
}

func nullStr(s string) any {
	if s == "" {
		return nil
	}
	return s
}
