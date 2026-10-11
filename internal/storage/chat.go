package storage

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// ConversationRow maps the conversations table.
type ConversationRow struct {
	ID        string
	Title     string
	UpdatedAt string
}

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

// execer is satisfied by *sql.DB and *sql.Tx, so one insert serves both the
// single-row path and the transactional user-turn path.
type execer interface {
	Exec(query string, args ...any) (sql.Result, error)
}

func insertMessage(ex execer, m MessageRow) error {
	now := time.Now().UTC().Format(time.RFC3339)
	if m.CreatedAt.IsZero() {
		m.CreatedAt = time.Now().UTC()
	}
	_, err := ex.Exec(`INSERT INTO messages
		(id, conversation_id, role, content, status, provider_id, model, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		m.ID, m.ConversationID, m.Role, m.Content, m.Status,
		nullStr(m.ProviderID), nullStr(m.Model),
		m.CreatedAt.Format(time.RFC3339), now)
	return err
}

// AppendMessage inserts one message with complete metadata.
func (d *DB) AppendMessage(m MessageRow) error { return insertMessage(d.inner, m) }

// UpdateMessage sets content + status after stream ends.
func (d *DB) UpdateMessage(id, content, status string) error {
	_, err := d.inner.Exec(`UPDATE messages SET content=?, status=?,
		updated_at=? WHERE id=?`,
		content, status, time.Now().UTC().Format(time.RFC3339), id)
	return err
}

// PrepareRetry readies a conversation to regenerate its newest turn without
// storing the user's text twice. When the newest rows are that user message
// followed by an unfinished reply (failed, or still "streaming" after a crash;
// the service never retries while a generation is live), the reply is deleted
// and the user row is kept for reuse (true). A user row with no reply yet is
// reused as is. Any other shape (a completed reply, different text, nothing
// stored because the first attempt never reached the DB) returns false and
// changes nothing, so the caller stores the user message as a normal send
// would. Check and delete run in one transaction.
func (d *DB) PrepareRetry(conversationID, text string) (bool, error) {
	tx, err := d.inner.Begin()
	if err != nil {
		return false, err
	}
	defer tx.Rollback() //nolint:errcheck // no-op after Commit

	// rowid breaks ties: a user row and its reply are written in the same
	// second, so created_at alone does not order them.
	rows, err := tx.Query(`SELECT id, role, content, status FROM messages
		WHERE conversation_id=? ORDER BY created_at DESC, rowid DESC LIMIT 2`, conversationID)
	if err != nil {
		return false, err
	}
	type row struct{ id, role, content, status string }
	var last []row
	for rows.Next() {
		var r row
		if err := rows.Scan(&r.id, &r.role, &r.content, &r.status); err != nil {
			rows.Close()
			return false, err
		}
		last = append(last, r)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return false, err
	}
	rows.Close()

	switch {
	case len(last) >= 1 && last[0].role == "user" && last[0].content == text:
		return true, nil
	case len(last) == 2 && last[0].role == "assistant" &&
		(last[0].status == "failed" || last[0].status == "streaming") &&
		last[1].role == "user" && last[1].content == text:
		if _, err := tx.Exec(`DELETE FROM messages WHERE id=? AND conversation_id=?
			AND role='assistant' AND status IN ('failed','streaming')`, last[0].id, conversationID); err != nil {
			return false, err
		}
		if err := tx.Commit(); err != nil {
			return false, err
		}
		return true, nil
	default:
		return false, nil
	}
}

// ContextRow is one message for building the provider request: text plus
// attachment metadata. Bytes are loaded separately, only when needed.
type ContextRow struct {
	ID          string
	Role        string
	Content     string
	Attachments []AttachmentRow
}

// RecentContext returns the last N messages oldest-first with attachment
// metadata. The empty assistant row created for the reply in flight is left
// out: sending {"role":"assistant","content":""} confuses stricter providers.
func (d *DB) RecentContext(conversationID string, n int) ([]ContextRow, error) {
	if n <= 0 {
		n = 20
	}
	rows, err := d.inner.Query(`SELECT id, role, content FROM messages
		WHERE conversation_id=? AND status IN ('complete','streaming')
		AND NOT (role='assistant' AND content='')
		ORDER BY created_at DESC, rowid DESC LIMIT ?`, conversationID, n)
	if err != nil {
		return nil, err
	}
	var rev []ContextRow
	for rows.Next() {
		var r ContextRow
		if err := rows.Scan(&r.ID, &r.Role, &r.Content); err != nil {
			rows.Close()
			return nil, err
		}
		if r.Role != "user" && r.Role != "assistant" && r.Role != "system" {
			continue
		}
		rev = append(rev, r)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	rows.Close()
	out := make([]ContextRow, 0, len(rev))
	ids := make([]string, 0, len(rev))
	for i := len(rev) - 1; i >= 0; i-- {
		out = append(out, rev[i])
		ids = append(ids, rev[i].ID)
	}
	atts, err := d.AttachmentsFor(ids)
	if err != nil {
		return nil, err
	}
	for i := range out {
		out[i].Attachments = atts[out[i].ID]
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

// ListConversations returns id+title ordered by recent activity.
func (d *DB) ListConversations() ([]ConversationRow, error) {
	rows, err := d.inner.Query(`SELECT id, title, updated_at FROM conversations
		ORDER BY updated_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []ConversationRow{}
	for rows.Next() {
		var c ConversationRow
		if err := rows.Scan(&c.ID, &c.Title, &c.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// RenameConversation sets the title. Empty title rejected, max 120 runes.
func (d *DB) RenameConversation(id, title string) error {
	title = strings.TrimSpace(title)
	if r := []rune(title); len(r) > 120 {
		title = string(r[:120])
	}
	if title == "" {
		return fmt.Errorf("empty title")
	}
	res, err := d.inner.Exec(`UPDATE conversations SET title=?, updated_at=? WHERE id=?`,
		title, time.Now().UTC().Format(time.RFC3339), id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// DeleteConversation removes a conversation; messages cascade via FK.
func (d *DB) DeleteConversation(id string) error {
	res, err := d.inner.Exec(`DELETE FROM conversations WHERE id=?`, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func nullStr(s string) any {
	if s == "" {
		return nil
	}
	return s
}
