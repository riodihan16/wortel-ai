package credentials

import "errors"

func errEmpty() error   { return errors.New("empty secret") }
func errStoreOp() error { return errors.New("credential store unavailable") }
