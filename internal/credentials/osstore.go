package credentials

import (
	"context"

	"github.com/zalando/go-keyring"
)

// serviceName groups all Wortel AI secrets in the OS credential store.
const serviceName = "WortelAI"

// OSStore persists secrets in the OS credential manager
// (Windows Credential Manager via wincred on Windows).
// ref is the credential account name, e.g. "wortel-provider-<uuid>".
// Errors never contain secret bytes.
type OSStore struct{}

func (OSStore) Set(_ context.Context, ref string, secret []byte) error {
	if err := validateRef(ref); err != nil {
		return err
	}
	if len(secret) == 0 {
		return errEmpty()
	}
	if err := keyring.Set(serviceName, ref, string(secret)); err != nil {
		return errStoreOp()
	}
	return nil
}

func (OSStore) Get(_ context.Context, ref string) ([]byte, error) {
	if err := validateRef(ref); err != nil {
		return nil, err
	}
	s, err := keyring.Get(serviceName, ref)
	if err != nil {
		if err == keyring.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, errStoreOp()
	}
	return []byte(s), nil
}

func (OSStore) Exists(_ context.Context, ref string) (bool, error) {
	if err := validateRef(ref); err != nil {
		return false, err
	}
	_, err := keyring.Get(serviceName, ref)
	if err != nil {
		if err == keyring.ErrNotFound {
			return false, nil
		}
		return false, errStoreOp()
	}
	return true, nil
}

func (OSStore) Delete(_ context.Context, ref string) error {
	if err := validateRef(ref); err != nil {
		return err
	}
	if err := keyring.Delete(serviceName, ref); err != nil {
		// Deleting a missing credential is success for cleanup paths.
		if err == keyring.ErrNotFound {
			return nil
		}
		return errStoreOp()
	}
	return nil
}
