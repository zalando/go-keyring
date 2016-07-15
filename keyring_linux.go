package keyring

import (
	"fmt"

	"github.com/mikkeloscar/go-keyring/secret_service"
)

type secretServiceProvider struct{}

// Set stores stores user and pass in the keyring under the defined service
// name.
func (s secretServiceProvider) Set(service, user, pass string) error {
	svc, err := ss.NewSecretService()
	if err != nil {
		return err
	}

	// create or get collection
	// TODO: make work
	// c, err := svc.CreateCollection(service)
	// if err != nil {
	// 	return err
	// }

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

	collection := svc.GetCollection("login")

	err = svc.CreateItem(collection,
		fmt.Sprintf("Password for '%s' on '%s'", user, service),
		attributes, secret)
	if err != nil {
		return err
	}

	return nil
}

// Get gets a secret from the keyring given a service name and a user.
func (s secretServiceProvider) Get(service, user string) (string, error) {
	svc, err := ss.NewSecretService()
	if err != nil {
		return "", err
	}

	// open a session
	session, err := svc.OpenSession()
	if err != nil {
		return "", err
	}
	defer svc.Close(session)

	search := map[string]string{
		"username": user,
		"service":  service,
	}

	collection := svc.GetCollection("login")

	results, err := svc.SearchItems(collection, search)
	if err != nil {
		return "", err
	}

	if len(results) == 0 {
		return "", ErrNotFound
	}

	secret, err := svc.GetSecret(results[0], session.Path())
	if err != nil {
		return "", err
	}

	return string(secret.Value), nil
}

func init() {
	provider = secretServiceProvider{}
}
