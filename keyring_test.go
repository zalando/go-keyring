package keyring

import "testing"

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
