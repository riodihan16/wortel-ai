package storage

import (
	"strings"
	"time"
)

// AttachmentRow maps the attachments table. Data is set only on insert and
// by AttachmentData; metadata reads leave it nil.
type AttachmentRow struct {
	ID        string
	MessageID string
	Kind      string
	Mime      string
	Name      string
	Size      int64
	Position  int
	Data      []byte
}

// AppendUserTurn stores a user message, its attachments, and the empty
// streaming assistant row in one transaction, so a crash can never leave a
// question without its images or an orphan reply.
func (d *DB) AppendUserTurn(user MessageRow, atts []AttachmentRow, assistant MessageRow) error {
	tx, err := d.inner.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck // no-op after Commit
	if err := insertMessage(tx, user); err != nil {
		return err
	}
	now := time.Now().UTC().Format(time.RFC3339)
	for i, a := range atts {
		if _, err := tx.Exec(`INSERT INTO attachments
			(id, message_id, kind, mime, name, size, data, position, created_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			a.ID, user.ID, a.Kind, a.Mime, a.Name, int64(len(a.Data)), a.Data, i, now); err != nil {
			return err
		}
	}
	if err := insertMessage(tx, assistant); err != nil {
		return err
	}
	return tx.Commit()
}

// AttachmentsFor returns metadata (no bytes) grouped by message id, each
// group in position order. Missing ids simply have no entry.
func (d *DB) AttachmentsFor(messageIDs []string) (map[string][]AttachmentRow, error) {
	out := map[string][]AttachmentRow{}
	if len(messageIDs) == 0 {
		return out, nil
	}
	q := `SELECT id, message_id, kind, mime, name, size, position FROM attachments
		WHERE message_id IN (?` + strings.Repeat(",?", len(messageIDs)-1) + `)
		ORDER BY message_id, position`
	args := make([]any, len(messageIDs))
	for i, id := range messageIDs {
		args[i] = id
	}
	rows, err := d.inner.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var a AttachmentRow
		if err := rows.Scan(&a.ID, &a.MessageID, &a.Kind, &a.Mime, &a.Name, &a.Size, &a.Position); err != nil {
			return nil, err
		}
		out[a.MessageID] = append(out[a.MessageID], a)
	}
	return out, rows.Err()
}

// AttachmentData returns one attachment's mime and bytes. sql.ErrNoRows
// when the id is unknown (or its conversation was deleted).
func (d *DB) AttachmentData(id string) (string, []byte, error) {
	var mime string
	var data []byte
	err := d.inner.QueryRow(`SELECT mime, data FROM attachments WHERE id=?`, id).Scan(&mime, &data)
	return mime, data, err
}
