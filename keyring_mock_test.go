package keyring

import "testing"

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

	pw, err := mp.Get(service, user)
	if err != nil {
		t.Errorf("Should not fail, got: %s", err)
	}

	if password != pw {
		t.Errorf("Expected password %s, got %s", password, pw)
	}
}

// TestGetNonExisting tests getting a secret not in the keyring.
func TestMockGetNonExisting(t *testing.T) {
	mp := mockProvider{}

	_, err := mp.Get(service, user+"fake")
	if err != ErrNotFound {
		t.Errorf("Expected error ErrNotFound, got %s", err)
	}
}

// TestDelete tests deleting a secret from the keyring.
func TestMockDelete(t *testing.T) {
	mp := mockProvider{}

	err := mp.Set(service, user, password)
	if err != nil {
		t.Errorf("Should not fail, got: %s", err)
	}

	err = mp.Delete(service, user)
	if err != nil {
		t.Errorf("Should not fail, got: %s", err)
	}
}

// TestDeleteNonExisting tests deleting a secret not in the keyring.
func TestMockDeleteNonExisting(t *testing.T) {
	mp := mockProvider{}

	err := mp.Delete(service, user+"fake")
	if err != ErrNotFound {
		t.Errorf("Expected error ErrNotFound, got %s", err)
	}
}
