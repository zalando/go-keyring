package kw

import (
	"errors"
	"fmt"

	"github.com/godbus/dbus/v5"
	errs "github.com/zalando/go-keyring/errors"
)

const (
	serviceName     = "org.kde.kwalletd5"
	servicePath     = "/modules/kwalletd5"
	methodInterface = "org.kde.KWallet"
)

// KWallet is an interface for the KWallet dbus API.
type KWallet struct {
	*dbus.Conn
	object dbus.BusObject
	handle int
}

// NewKWallet inializes a new NewKwallet object.
func NewKWallet() (*KWallet, error) {
	conn, err := dbus.SessionBus()
	if err != nil {
		return nil, err
	}

	kw := &KWallet{
		Conn:   conn,
		object: conn.Object(serviceName, servicePath),
	}

	var wallet string
	if err := kw.object.Call(methodInterface+".networkWallet", 0).Store(&wallet); err != nil {
		return nil, fmt.Errorf("Kwallet is not available: %w", err)
	}

	return kw, nil
}

// Open the wallet
func (k *KWallet) Open(service string) error {
	var wallet string
	if err := k.object.Call(methodInterface+".networkWallet", 0).Store(&wallet); err != nil {
		return err
	}

	if err := k.object.Call(methodInterface+".open", 0, wallet, int64(0), service).Store(&k.handle); err != nil {
		return err
	}
	return nil
}

// Set stores user and pass in the keyring under the defined service
// name.
func (k *KWallet) Set(service, user, pass string) error {
	if err := k.Open(service); err != nil {
		return err
	}

	var i int
	if err := k.object.Call(methodInterface+".writePassword", 0, k.handle, service, user, pass, service).Store(&i); err != nil {
		return err
	}
	if i < 0 {
		return errors.New("Could not write password")
	}
	return nil
}

// Get gets a secret from the keyring given a service name and a user.
func (k *KWallet) Get(service, user string) (string, error) {
	if err := k.Open(service); err != nil {
		return "", err
	}
	if b, err := k.Has(service, user); err != nil {
		return "", err
	} else if !b {
		return "", errs.ErrNotFound
	}
	var password string
	err := k.object.Call(methodInterface+".readPassword", 0, k.handle, service, user, service).Store(&password)
	return password, err
}

// Delete deletes a secret, identified by service & user, from the keyring.
func (k *KWallet) Delete(service, user string) error {
	if err := k.Open(service); err != nil {
		return err
	}
	if b, err := k.Has(service, user); err != nil {
		return err
	} else if !b {
		return errs.ErrNotFound
	}

	var i int
	if err := k.object.Call(methodInterface+".removeEntry", 0, k.handle, service, user, service).Store(&i); err != nil {
		return err
	}

	if i < 0 {
		return errors.New("Could not delete password")
	}
	return nil
}

// Has a key
func (k *KWallet) Has(service, key string) (bool, error) {
	var b bool
	if err := k.object.Call(methodInterface+".hasEntry", 0, k.handle, service, key, service).Store(&b); err != nil {
		return b, err
	}
	return b, nil
}
