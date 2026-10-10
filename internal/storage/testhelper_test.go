package storage

import "os"

func writeEmptyFile(path string) error {
	return os.WriteFile(path, []byte(""), 0o644)
}
