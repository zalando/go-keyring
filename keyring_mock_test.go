package keyring

import (
	"errors"
	"testing"
)

// TestSet tests setting a user and password in the keyring.
func TestMockSet(t *testing.T) {
	mp := mockProvider{}
	err := mp.Set(service, user, password)
	if err != nil {
		t.Errorf("Should not fail, got: %s", err)
	}
}

// TestGet tests getting a password from the keyring.
func TestMockGet(t *testing.T) {
	mp := mockProvider{}
	err := mp.Set(service, user, password)
	if err != nil {
		t.Errorf("Should not fail, got: %s", err)
	}

	_ = mp.Unlock(service, user)
	pw, err := mp.Get(service, user)
	if err != nil {
		t.Errorf("Should not fail, got: %s", err)
	}

	if password != pw {
		t.Errorf("Expected password %s, got %s", password, pw)
	}
}

// TestGetLocked tests getting a locked password from the keyring.
func TestMockGetLocked(t *testing.T) {
	mp := mockProvider{}
	err := mp.Set(service, user, password)
	if err != nil {
		t.Errorf("Should not fail, got: %s", err)
	}

	pwd, err := mp.Get(service, user)
	assertError(t, err, ErrNotFound)

	if pwd != "" {
		t.Errorf("Should not return item value, got: %s", pwd)
	}
}

// TestGetNonExisting tests getting a secret not in the keyring.
func TestMockGetNonExisting(t *testing.T) {
	mp := mockProvider{}

	_, err := mp.Get(service, user+"fake")
	assertError(t, err, ErrNotFound)
}

// TestDelete tests deleting a secret from the keyring.
func TestMockDelete(t *testing.T) {
	mp := mockProvider{}

	err := mp.Set(service, user, password)
	if err != nil {
		t.Errorf("Should not fail, got: %s", err)
	}

	_ = mp.Unlock(service, user)
	err = mp.Delete(service, user)
	if err != nil {
		t.Errorf("Should not fail, got: %s", err)
	}
}

// TestDeleteNonExisting tests deleting a secret not in the keyring.
func TestMockDeleteNonExisting(t *testing.T) {
	mp := mockProvider{}

	err := mp.Delete(service, user+"fake")
	assertError(t, err, ErrNotFound)
}

func TestMockWithError(t *testing.T) {
	mp := mockProvider{mockError: errors.New("mock error")}

	err := mp.Set(service, user, password)
	assertError(t, err, mp.mockError)

	_, err = mp.Get(service, user)
	assertError(t, err, mp.mockError)

	err = mp.Delete(service, user)
	assertError(t, err, mp.mockError)
}

func assertError(t *testing.T, err error, expected error) {
	if err != expected {
		t.Errorf("Expected error %s, got %s", expected, err)
	}
}
