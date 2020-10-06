package keyring

import (
	"testing"
)

const (
	service  = "test-service"
	user     = "test-user"
	password = "test-password"
)

// TestSet tests setting a user and password in the keyring.
func TestSet(t *testing.T) {
	err := Set(service, user, password)
	if err != nil {
		t.Errorf("Should not fail, got: %s", err)
	}
}

// TestGetMultiline tests getting a multi-line password from the keyring
func TestGetMultiLine(t *testing.T) {
	multilinePassword := `this password
has multiple
lines and will be
encoded by some keyring implementiations
like osx`
	err := Set(service, user, multilinePassword)
	if err != nil {
		t.Errorf("Should not fail, got: %s", err)
	}

	pw, err := Get(service, user)
	if err != nil {
		t.Errorf("Should not fail, got: %s", err)
	}

	if multilinePassword != pw {
		t.Errorf("Expected password %s, got %s", multilinePassword, pw)
	}
}

// TestGetSingleLineHex tests getting a single line hex string password from the keyring.
func TestGetSingleLineHex(t *testing.T) {
	hexPassword := "abcdef123abcdef123"
	err := Set(service, user, hexPassword)
	if err != nil {
		t.Errorf("Should not fail, got: %s", err)
	}

	pw, err := Get(service, user)
	if err != nil {
		t.Errorf("Should not fail, got: %s", err)
	}

	if hexPassword != pw {
		t.Errorf("Expected password %s, got %s", hexPassword, pw)
	}
}

// TestGet tests getting a password from the keyring.
func TestGet(t *testing.T) {
	err := Set(service, user, password)
	if err != nil {
		t.Errorf("Should not fail, got: %s", err)
	}

	pw, err := Get(service, user)
	if err != nil {
		t.Errorf("Should not fail, got: %s", err)
	}

	if password != pw {
		t.Errorf("Expected password %s, got %s", password, pw)
	}
}

// TestGetNonExisting tests getting a secret not in the keyring.
func TestGetNonExisting(t *testing.T) {
	_, err := Get(service, user+"fake")
	if err != ErrNotFound {
		t.Errorf("Expected error ErrNotFound, got %s", err)
	}
}

// TestDelete tests deleting a secret from the keyring.
func TestDelete(t *testing.T) {
	err := Delete(service, user)
	if err != nil {
		t.Errorf("Should not fail, got: %s", err)
	}
}

// TestDeleteNonExisting tests deleting a secret not in the keyring.
func TestDeleteNonExisting(t *testing.T) {
	err := Delete(service, user+"fake")
	if err != ErrNotFound {
		t.Errorf("Expected error ErrNotFound, got %s", err)
	}
}

// TestCustomKeyringSet tests deleting a secret from a custom keyring.
func TestCustomKeyringSet(t *testing.T) {
	keyRing := NewCustomKeyring("_keyring_test")
	err := keyRing.Set(service, user, password)
	if err != nil {
		t.Errorf("Should not fail: %s", err)
	}
}

// TestCustomKeyringGet tests getting a password from a custom keyring.
func TestCustomKeyringGet(t *testing.T) {
	keyRing := NewCustomKeyring("_keyring_test")
	pass, err := keyRing.Get(service, user)
	if err != nil {
		t.Errorf("Should not fail: %s", err)
	}

	if pass != password {
		t.Errorf("Expected password %s, got %s", password, pass)
	}
}

// TestCustomKeyringDelete tests deleting a secret from a custom keyring.
func TestCustomKeyringDelete(t *testing.T) {
	keyRing := NewCustomKeyring("_keyring_test")
	err := keyRing.Delete(service, user)
	if err != nil {
		t.Errorf("Should not fail, got: %s", err)
	}
}
