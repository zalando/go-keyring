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
	object     dbus.BusObject
	walletName string
	handle     int
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

	kw.walletName, err = kw.defaultWallet()
	return kw, err
}

// Set stores user and pass in the keyring under the defined service
// name.
func (k *KWallet) Set(service, user, pass string) error {
	if err := k.open(service); err != nil {
		return err
	}

	var i int
	// org.kde.KWallet.writePassword(handle int, folder string, key string, value string, appId string) int
	if err := k.object.Call(methodInterface+".writePassword", 0, k.handle, service, user, pass, service).Store(&i); err != nil {
		return fmt.Errorf("failed to write password: %w", err)
	}
	if i < 0 {
		return errors.New("Could not write password")
	}
	return nil
}

// Get gets a secret from the keyring given a service name and a user.
func (k *KWallet) Get(service, user string) (string, error) {
	if err := k.open(service); err != nil {
		return "", err
	}
	if b, err := k.hasEntry(service, user); err != nil {
		return "", err
	} else if !b {
		return "", errs.ErrNotFound
	}

	var password string
	// org.kde.KWallet.readPassword(handle int, folder string, key string, appId string) string
	if err := k.object.Call(methodInterface+".readPassword", 0, k.handle, service, user, service).Store(&password); err != nil {
		return "", fmt.Errorf("failed to read password: %w", err)
	}
	return password, nil
}

// Delete deletes a secret, identified by service & user, from the keyring.
func (k *KWallet) Delete(service, user string) error {
	if err := k.open(service); err != nil {
		return err
	}

	if b, err := k.hasEntry(service, user); err != nil {
		return err
	} else if !b {
		return errs.ErrNotFound
	}

	return k.removeEntry(service, user)
}

func (k *KWallet) open(service string) error {
	var alreadyOpen bool
	// org.kde.KWallet.isOpen(wallet string) bool
	if err := k.object.Call(methodInterface+".isOpen", 0, k.handle).Store(&alreadyOpen); err != nil {
		return fmt.Errorf("failed to check if wallet is open: %w", err)
	}
	if alreadyOpen {
		return nil
	}

	// org.kde.KWallet.open(wallet string, wId string, appId string) int
	if err := k.object.Call(methodInterface+".open", 0, k.walletName, int64(0), service).Store(&k.handle); err != nil {
		return fmt.Errorf("failed to open wallet: %w", err)
	}
	return nil
}

func (k *KWallet) defaultWallet() (string, error) {
	var wallet string
	// org.kde.KWallet.networkWallet() string
	if err := k.object.Call(methodInterface+".networkWallet", 0).Store(&wallet); err != nil {
		return "", fmt.Errorf("KWallet is not available: %w", err)
	}

	return wallet, nil
}

func (k *KWallet) removeEntry(service, key string) error {
	var i int
	// org.kde.KWallet.removeEntry(handle int, folder string, key string, appId string) int
	if err := k.object.Call(methodInterface+".removeEntry", 0, k.handle, service, key, service).Store(&i); err != nil {
		return fmt.Errorf("failed to delete entry: %w", err)
	}
	if i < 0 {
		return errors.New("Could not delete password")
	}

	return nil
}

func (k *KWallet) hasEntry(service, key string) (bool, error) {
	var b bool
	// org.kde.KWallet.hasEntry(handle int, folder string, key string, appId string) bool
	if err := k.object.Call(methodInterface+".hasEntry", 0, k.handle, service, key, service).Store(&b); err != nil {
		return b, fmt.Errorf("failed to check if entry exists: %w", err)
	}
	return b, nil
}
