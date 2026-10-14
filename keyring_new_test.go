package keyring

import (
	"errors"
	"testing"
)

func TestNewDefaultBackend(t *testing.T) {
	MockInit()
	defer MockRestore()

	k, err := New(Options{})
	if err != nil {
		t.Fatalf("Should not fail, got: %s", err)
	}

	if err := k.Set(service, user, password); err != nil {
		t.Fatalf("Should not fail, got: %s", err)
	}
	// The default backend is the one behind the package level functions.
	pw, err := Get(service, user)
	if err != nil || pw != password {
		t.Errorf("Expected password %q, got %q, %v", password, pw, err)
	}
}

func TestNewUnknownBackend(t *testing.T) {
	if _, err := New(Options{Backend: Backend(-1)}); !errors.Is(err, ErrUnsupportedBackend) {
		t.Errorf("Expected ErrUnsupportedBackend, got %v", err)
	}
}
