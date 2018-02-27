package keyring

import "fmt"

// provider set in the init function by the relevant os file e.g.:
// keyring_linux.go
var provider Keyring

var (
	// ErrNotFound is the expected error if the secret isn't found in the
	// keyring.
	ErrNotFound = fmt.Errorf("secret not found in keyring")
)

// Keyring provides a simple set/get interface for a keyring service.
type Keyring interface {
	// Set password in keyring for user.
	Set(service, user, password string) error
	// Set Internet password in keyring for user.
	IntSet(service, user, password string) error
	// Get Internet password from keyring given service and user name.
	IntGet(service, user string) (string, error)
	Get(service, user string) (string, error)
	// Delete secret from keyring.
	Delete(service, user string) error
}

// Set password in keyring for user.
func Set(service, user, password string) error {
	return provider.Set(service, user, password)
}

// IntSet set Internet password in keyring for user.
func IntSet(service, user, password string) error {
	return provider.IntSet(service, user, password)
}

// IntGet get internet password from keyring given service and user name.
func IntGet(service, user string) (string, error) {
	return provider.IntGet(service, user)
}

// Get password from keyring given service and user name.
func Get(service, user string) (string, error) {
	return provider.Get(service, user)
}

// Delete secret from keyring.
func Delete(service, user string) error {
	return provider.Delete(service, user)
}
