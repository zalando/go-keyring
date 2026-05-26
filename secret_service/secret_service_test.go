package ss

import (
	"testing"

	dbus "github.com/godbus/dbus/v5"
)

const testSession = dbus.ObjectPath("/session")

// TestNewSecretTextContent tests that valid UTF-8 strings get text/plain content type.
func TestNewSecretTextContent(t *testing.T) {
	secret := NewSecret(testSession, "plain text password")

	expected := "text/plain; charset=utf-8"
	if secret.ContentType != expected {
		t.Errorf("Expected content type %s, got %s", expected, secret.ContentType)
	}
}

// TestNewSecretBinaryContent tests that non-UTF-8 data gets application/octet-stream content type.
func TestNewSecretBinaryContent(t *testing.T) {
	secret := NewSecret(testSession, string([]byte{0x1f, 0x8b, 0x08, 0x00, 0xff, 0xfe}))

	expected := "application/octet-stream"
	if secret.ContentType != expected {
		t.Errorf("Expected content type %s, got %s", expected, secret.ContentType)
	}
}

// TestNewSecretEmptyString tests that an empty string gets text/plain content type.
func TestNewSecretEmptyString(t *testing.T) {
	secret := NewSecret(testSession, "")

	expected := "text/plain; charset=utf-8"
	if secret.ContentType != expected {
		t.Errorf("Expected content type %s, got %s", expected, secret.ContentType)
	}
}

// TestNewSecretUnicodeContent tests that valid UTF-8 with non-ASCII chars gets text/plain content type.
func TestNewSecretUnicodeContent(t *testing.T) {
	secret := NewSecret(testSession, "passwort mit Ümlauten äöü")

	expected := "text/plain; charset=utf-8"
	if secret.ContentType != expected {
		t.Errorf("Expected content type %s, got %s", expected, secret.ContentType)
	}
}

// TestNewSecretValuePreserved tests that the secret value and session are stored correctly.
func TestNewSecretValuePreserved(t *testing.T) {
	input := "test secret"
	secret := NewSecret(testSession, input)

	if string(secret.Value) != input {
		t.Errorf("Expected value %s, got %s", input, string(secret.Value))
	}
	if secret.Session != testSession {
		t.Errorf("Expected session %s, got %s", testSession, secret.Session)
	}
}
