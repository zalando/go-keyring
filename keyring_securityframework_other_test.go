//go:build !darwin

package keyring

import (
	"errors"
	"testing"
)

func TestNewSecurityFrameworkUnsupported(t *testing.T) {
	if _, err := New(Options{Backend: BackendSecurityFramework}); !errors.Is(err, ErrUnsupportedBackend) {
		t.Errorf("Expected ErrUnsupportedBackend, got %v", err)
	}
}
