package config

import "errors"

func errName() error  { return errors.New("nama provider wajib diisi") }
func errModel() error { return errors.New("model ID wajib diisi (maks 256 karakter)") }
func errURL() error {
	return errors.New("base URL tidak valid: gunakan http(s)://host tanpa kredensial")
}
