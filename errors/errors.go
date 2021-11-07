package errs

import "runtime"

const (
	ErrNotFound            = KeyringError("secret not found in keyring")
	ErrUnsupportedPlatform = KeyringError("Unsupported platform: " + runtime.GOOS)
)

type KeyringError string

func (e KeyringError) Error() string {
	return string(e)
}
