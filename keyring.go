package keyring

import (
	"fmt"
)

// provider set in the init function by the relevant os file e.g.:
// keyring_linux.go
var provider Keyring = fallbackServiceProvider{}

var (
	// ErrNotFound is the expected error if the secret isn't found in the
	// keyring.
	ErrNotFound = fmt.Errorf("secret not found in keyring")
)

// Keyring provides a simple set/get interface for a keyring service.
type Keyring interface {
	// Set password in keyring for user.
	Set(service, user, password string) error
	// Get password from keyring given service and user name.
	Get(service, user string) (string, error)
	// Delete secret from keyring.
	Delete(service, user string) error
}

// Set password in keyring for user.
func Set(service, user, password string) error {
	return provider.Set(service, user, password)
}

// Get password from keyring given service and user name.
func Get(service, user string) (string, error) {
	return provider.Get(service, user)
}

// Delete secret from keyring.
func Delete(service, user string) error {
	return provider.Delete(service, user)
}

// CustomKeyring allows to use custom keyring names
type CustomKeyring struct {
	keyring     Keyring
	keyringName string
}

// NewCustomKeyring create a new custom keyring.
// A custom keyring allows you to use a custom
// keyring name on supported platforms
func NewCustomKeyring(name string) *CustomKeyring {
	ck := CustomKeyring{
		keyringName: name,
	}

	// Set customs keyringprovider to current provider
	ck.keyring = provider

	// If provider is 'supported' set its keyring name
	if pv, ok := provider.(secretServiceProvider); ok {
		pv.keyringName = name
		ck.keyring = pv
	}

	return &ck
}

// Set password in keyring for user.
func (ck *CustomKeyring) Set(service, user, password string) error {
	return ck.keyring.Set(service, user, password)
}

// Get password from keyring given service and user name.
func (ck *CustomKeyring) Get(service, user string) (string, error) {
	return ck.keyring.Get(service, user)
}

// Delete secret from keyring.
func (ck *CustomKeyring) Delete(service, user string) error {
	return ck.keyring.Delete(service, user)
}
