package keyring

import (
	"github.com/zalando/go-keyring/internal/darwin"
)

// macOSSecurityFramework is a Keyring that calls the Security framework
// directly instead of running /usr/bin/security. See BackendSecurityFramework.
type macOSSecurityFramework struct{}

func newSecurityFramework() (Keyring, error) {
	return macOSSecurityFramework{}, nil
}

// Get password from macos keyring given service and user name.
func (macOSSecurityFramework) Get(service, username string) (string, error) {
	out, err := darwin.FindGenericPassword(service, username)
	if err == darwin.ErrNotFound {
		return "", ErrNotFound
	} else if err != nil {
		return "", err
	}
	return decodeSecret(out)
}

// Set stores a secret in the macos keyring given a service name and a user.
func (macOSSecurityFramework) Set(service, username, password string) error {
	return darwin.AddGenericPassword(service, username, encodeSecret(password))
}

// Delete deletes a secret, identified by service & user, from the keyring.
func (macOSSecurityFramework) Delete(service, username string) error {
	err := darwin.DeleteGenericPassword(service, username)
	if err == darwin.ErrNotFound {
		return ErrNotFound
	}
	return err
}

// DeleteAll deletes all secrets for a given service
func (macOSSecurityFramework) DeleteAll(service string) error {
	// if service is empty, do nothing otherwise it might accidentally delete all secrets
	if service == "" {
		return ErrNotFound
	}
	return darwin.DeleteGenericPasswords(service)
}

// ListUsers returns a list of all users for a given service
func (macOSSecurityFramework) ListUsers(service string) ([]string, error) {
	if service == "" {
		return []string{}, nil
	}
	return darwin.ListGenericPasswordAccounts(service)
}
