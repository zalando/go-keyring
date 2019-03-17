package keyring

import (
	"errors"

	ss "github.com/zalando/go-keyring/secret_service"
)

// All of the following method errors out on BSD
var errBSD = errors.New("Unsupported platform: FreeBSD")

type secretServiceProvider struct{}

// Set stores user and pass in the keyring under the defined service
// name.
func (s secretServiceProvider) Set(service, user, pass string) error {
	return errBSD
}

// findItem looksup an item by service and user.
func (s secretServiceProvider) findItem(svc *ss.SecretService, service, user string) (interface{}, error) {
	return nil, errBSD
}

// Get gets a secret from the keyring given a service name and a user.
func (s secretServiceProvider) Get(service, user string) (string, error) {
	return "", errBSD
}

// Delete deletes a secret, identified by service & user, from the keyring.
func (s secretServiceProvider) Delete(service, user string) error {
	return errBSD
}

func init() {
	provider = secretServiceProvider{}
}
