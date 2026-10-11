package attach

import (
	"errors"
	"fmt"
	"sync"

	"github.com/google/uuid"
)

var (
	ErrTooMany   = fmt.Errorf("maksimal %d lampiran per pesan", MaxPerMessage)
	ErrTooBig    = fmt.Errorf("total lampiran melebihi %d MB per pesan", MaxMessageBytes>>20)
	ErrUnknown   = errors.New("lampiran tidak ditemukan: lampirkan ulang")
	ErrDuplicate = errors.New("lampiran yang sama disebut dua kali")
)

// Item is a staged attachment: metadata plus the bytes to send.
type Item struct {
	Meta Meta
	Data []byte
}

// Staging holds the attachments of the message being composed, in memory,
// until they are sent (moved to the database) or removed. The per-message
// limits are enforced here, so a send is never refused for size or count.
// Safe for concurrent use.
type Staging struct {
	mu    sync.Mutex
	items map[string]Item
	total int64
	newID func() string
}

func NewStaging() *Staging {
	return &Staging{items: map[string]Item{}, newID: uuid.NewString}
}

// Add stores data under a fresh id and returns the completed Meta.
func (s *Staging) Add(m Meta, data []byte) (Meta, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.items) >= MaxPerMessage {
		return Meta{}, ErrTooMany
	}
	if s.total+int64(len(data)) > MaxMessageBytes {
		return Meta{}, ErrTooBig
	}
	m.ID = s.newID()
	m.Size = int64(len(data))
	s.items[m.ID] = Item{Meta: m, Data: data}
	s.total += m.Size
	return m, nil
}

// Get returns one staged item. Data is shared, not copied; callers only read.
func (s *Staging) Get(id string) (Item, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	it, ok := s.items[id]
	return it, ok
}

// Select returns the items for ids, in that order. Any unknown or repeated
// id fails the whole call, so a send never silently drops an attachment or
// stores one twice. Nothing is removed.
func (s *Staging) Select(ids []string) ([]Item, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Item, 0, len(ids))
	seen := make(map[string]bool, len(ids))
	for _, id := range ids {
		if seen[id] {
			return nil, ErrDuplicate
		}
		seen[id] = true
		it, ok := s.items[id]
		if !ok {
			return nil, ErrUnknown
		}
		out = append(out, it)
	}
	return out, nil
}

// Remove drops ids; unknown ids are ignored.
func (s *Staging) Remove(ids ...string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, id := range ids {
		if it, ok := s.items[id]; ok {
			s.total -= it.Meta.Size
			delete(s.items, id)
		}
	}
}

// Clear drops everything (shutdown).
func (s *Staging) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items = map[string]Item{}
	s.total = 0
}

// Len reports how many items are staged.
func (s *Staging) Len() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.items)
}
