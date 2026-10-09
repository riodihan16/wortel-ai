package credentials

import (
	"context"
	"errors"
)

// ErrNotFound is returned when no credential exists for ref.
var ErrNotFound = errors.New("credential not found")

// Store persists secrets outside SQLite/settings/logs.
// Frontend must never receive secret bytes.
type Store interface {
	Set(ctx context.Context, ref string, secret []byte) error
	Get(ctx context.Context, ref string) ([]byte, error)
	Exists(ctx context.Context, ref string) (bool, error)
	Delete(ctx context.Context, ref string) error
}

func validateRef(ref string) error {
	if ref == "" {
		return errors.New("empty credential ref")
	}
	return nil
}
