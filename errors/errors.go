package errors

const (
	ErrNotFound = KeyringError("secret not found in keyring")
)

type KeyringError string

func (e KeyringError) Error() string {
	return string(e)
}
