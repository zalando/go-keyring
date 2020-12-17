package kw

import (
	"errors"

	"github.com/godbus/dbus"
)

const (
	serviceName     = "org.kde.kwalletd5"
	servicePath     = "/modules/kwalletd5"
	methodInterface = "org.kde.KWallet"
)

// KWallet is an interface for the KWallet dbus API.
type KWallet struct {
	*dbus.Conn
	object  dbus.BusObject
	service string
	handle  int
}

// NewKWallet inializes a new NewKwallet object.
func NewKWallet(service string) (*KWallet, error) {
	conn, err := dbus.SessionBus()
	if err != nil {
		return nil, err
	}

	return &KWallet{
		conn,
		conn.Object(serviceName, servicePath),
		service,
		0,
	}, nil
}

// IsAvailable checks if the kwallet is available
func (k *KWallet) IsAvailable() bool {
	var wallet string
	if err := k.object.Call(methodInterface+".networkWallet", 0).Store(&wallet); err != nil {
		return false
	}
	return true
}

// Open the wallet
func (k *KWallet) Open() error {
	var wallet string
	if err := k.object.Call(methodInterface+".networkWallet", 0).Store(&wallet); err != nil {
		return err
	}

	if err := k.object.Call(methodInterface+".open", 0, wallet, int64(0), k.service).Store(&k.handle); err != nil {
		return err
	}
	return nil
}

// Read a value by key
func (k *KWallet) Read(key string) (string, error) {
	var password string
	if err := k.object.Call(methodInterface+".readPassword", 0, k.handle, k.service, key, k.service).Store(&password); err != nil {
		return "", err
	}
	return password, nil
}

// Write a key, value pair
func (k *KWallet) Write(key, value string) error {
	var i int
	if err := k.object.Call(methodInterface+".writePassword", 0, k.handle, k.service, key, value, k.service).Store(&i); err != nil {
		return err
	}
	if i < 0 {
		return errors.New("Could not write password")
	}
	return nil
}

// Delete a value by key
func (k *KWallet) Delete(key string) error {
	var i int
	if err := k.object.Call(methodInterface+".removeEntry", 0, k.handle, k.service, key, k.service).Store(&i); err != nil {
		return err
	}
	if i < 0 {
		return errors.New("Could not delete password")
	}
	return nil
}

// Has a  key
func (k *KWallet) Has(key string) (bool, error) {
	var b bool
	if err := k.object.Call(methodInterface+".hasEntry", 0, k.handle, k.service, key, k.service).Store(&b); err != nil {
		return b, err
	}
	return b, nil
}
