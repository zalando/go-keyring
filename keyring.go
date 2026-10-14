package keyring

import "errors"

// provider set in the init function by the relevant os file e.g.:
// keyring_unix.go
var provider Keyring = fallbackServiceProvider{}

// restoreProvider is set by platform init to restore the real provider after MockInit
var restoreProvider func()

var (
	// ErrNotFound is the expected error if the secret isn't found in the
	// keyring.
	ErrNotFound = errors.New("secret not found in keyring")
	// ErrUnsupportedBackend is returned by New if the requested backend is
	// unknown or not available on this platform.
	ErrUnsupportedBackend = errors.New("keyring backend not supported on this platform")
	// ErrSetDataTooBig is returned if `Set` was called with too much data.
	// On MacOS: The combination of service, username & password should not exceed ~3000 bytes
	// On Windows: The service is limited to 32KiB while the password is limited to 2560 bytes
	// On Linux/Unix: There is no theoretical limit but performance suffers with big values (>100KiB)
	ErrSetDataTooBig = errors.New("data passed to Set was too big")
)

// Keyring provides a simple set/get interface for a keyring service.
type Keyring interface {
	// Set password in keyring for user.
	Set(service, user, password string) error
	// Get password from keyring given service and user name.
	Get(service, user string) (string, error)
	// Delete secret from keyring.
	Delete(service, user string) error
	// DeleteAll deletes all secrets for a given service
	DeleteAll(service string) error
	// ListUsers returns a list of all users for a given service
	ListUsers(service string) ([]string, error)
}

// Backend selects the implementation used by a Keyring created with New.
type Backend int

const (
	// BackendDefault is the default implementation for the platform. It is
	// what the package level functions (Set, Get, ...) use. On macOS it
	// drives the /usr/bin/security command.
	BackendDefault Backend = iota
	// BackendSecurityFramework calls the macOS Security framework directly
	// from the current process, without cgo.
	//
	// Keychain items are then owned by the calling application, so other
	// processes (including /usr/bin/security) get a confirmation prompt
	// before they can read them, unlike items created through
	// BackendDefault, which are readable by anything that can run
	// /usr/bin/security. Conversely, macOS asks for confirmation when the
	// application's code signature changes, and when this backend reads an
	// item created through BackendDefault. The two backends use the same
	// item format, but do not silently share items.
	//
	// Only available on macOS; New returns ErrUnsupportedBackend elsewhere.
	BackendSecurityFramework
)

// Options configure a Keyring created with New. The zero value selects the
// default behavior.
type Options struct {
	// Backend selects the implementation. Defaults to BackendDefault.
	Backend Backend
}

// New returns a Keyring using the backend selected by opts.
//
// The package level functions (Set, Get, Delete, ...) are unaffected and keep
// using the default backend.
func New(opts Options) (Keyring, error) {
	switch opts.Backend {
	case BackendDefault:
		return provider, nil
	case BackendSecurityFramework:
		return newSecurityFramework()
	default:
		return nil, ErrUnsupportedBackend
	}
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

// DeleteAll deletes all secrets for a given service
func DeleteAll(service string) error {
	return provider.DeleteAll(service)
}

// ListUsers returns a list of all users for a given service
func ListUsers(service string) ([]string, error) {
	return provider.ListUsers(service)
}
