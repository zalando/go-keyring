package keyring

import (
	"fmt"
	"github.com/godbus/dbus"
	"github.com/ainmosni/go-keyring/secret_service"
)

type secretServiceProvider struct{}

// Set stores user and pass in the keyring under the defined service
// name.
func (s secretServiceProvider) Set(service, user, pass string) error {
	svc, err := ss.NewSecretService()
	if err != nil {
		return err
	}

	// open a session
	session, err := svc.OpenSession()
	if err != nil {
		return err
	}
	defer svc.Close(session)

	attributes := map[string]string{
		"username": user,
		"service":  service,
	}

	secret := ss.NewSecret(session.Path(), pass)

	collection := svc.GetLoginCollection()

	err = svc.Unlock(collection.Path())
	if err != nil {
		return err
	}

	err = svc.CreateItem(collection,
		fmt.Sprintf("Password for '%s' on '%s'", user, service),
		attributes, secret)
	if err != nil {
		return err
	}

	return nil
}

// findItem looksup an item by service and user.
func (s secretServiceProvider) findItem(svc *ss.SecretService, search map[string]string) (dbus.ObjectPath, error) {
	collection := svc.GetLoginCollection()

	search := map[string]string{
		"username": user,
		"service":  service,
	}

	err := svc.Unlock(collection.Path())
	if err != nil {
		return "", err
	}

	results, err := svc.SearchItems(collection, search)
	if err != nil {
		return "", err
	}

	if len(results) == 0 {
		return "", ErrNotFound
	}

	return results[0], nil
}

// Get gets a secret from the keyring given a service name and a user.
func (s secretServiceProvider) Get(service, user string) (string, error) {
	return s.find(map[string][string]{
		"username": user,
		"service": service,
	})
}

// Query does a custom keychain query, for entries that don't follow service/user.
func (s secretServiceProvider) Query(query map[string]string) (string, error) {
	return s.find(query)
}

func (s secretServiceProvider) find(search map[string]string) (string, error) {
	svc, err := ss.NewSecretService()
	if err != nil {
		return "", err
	}

	item, err := s.findItem(svc, search)
	if err != nil {
		return "", err
	}

	// open a session
	session, err := svc.OpenSession()
	if err != nil {
		return "", err
	}
	defer svc.Close(session)

	secret, err := svc.GetSecret(item, session.Path())
	if err != nil {
		return "", err
	}

	return string(secret.Value), nil
}


// Delete deletes a secret, identified by service & user, from the keyring.
func (s secretServiceProvider) Delete(service, user string) error {
	svc, err := ss.NewSecretService()
	if err != nil {
		return err
	}

	item, err := s.findItem(svc, service, user)
	if err != nil {
		return err
	}

	return svc.Delete(item)
}

func init() {
	provider = secretServiceProvider{}
}
