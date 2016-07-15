package keyring

import "fmt"

// provider set in the init function by the relevant os file e.g.:
// keyring_linux.go
var provider keyring

var (
	ErrNotFound = fmt.Errorf("secret not found in keyring")
)

// keyring provides a simple set/get interface for a keyring service.
type keyring interface {
	// Get password from keyring given service and user name.
	Get(service, user string) (string, error)
	// Set password in keyring for user
	Set(service, user, password string) error
}

// Get password from keyring given service and user name.
func Get(service, user string) (string, error) {
	return provider.Get(service, user)
}

// Set password in keyring for user
func Set(service, user, password string) error {
	return provider.Set(service, user, password)
}
