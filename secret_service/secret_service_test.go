package ss

import (
	"bufio"
	"errors"
	"os/exec"
	"strings"
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

// fakePrompt is a Secret Service prompt that completes with a fixed result.
type fakePrompt struct {
	conn      *dbus.Conn
	path      dbus.ObjectPath
	dismissed bool
	result    []dbus.ObjectPath
}

func (p *fakePrompt) Prompt(windowID string) *dbus.Error {
	go func() {
		_ = p.conn.Emit(p.path, promptInterface+".Completed", p.dismissed, dbus.MakeVariant(p.result))
	}()
	return nil
}

// fakeService is a Secret Service that always asks for a prompt to unlock.
type fakeService struct {
	prompt dbus.ObjectPath
}

func (s *fakeService) Unlock(objects []dbus.ObjectPath) ([]dbus.ObjectPath, dbus.ObjectPath, *dbus.Error) {
	return []dbus.ObjectPath{}, s.prompt, nil
}

// newTestSecretService starts a private D-Bus daemon serving a fake Secret
// Service whose unlock prompt completes with the given result.
func newTestSecretService(t *testing.T, dismissed bool, result []dbus.ObjectPath) *SecretService {
	t.Helper()
	if _, err := exec.LookPath("dbus-daemon"); err != nil {
		t.Skip("dbus-daemon not found")
	}

	cmd := exec.Command("dbus-daemon", "--session", "--nofork", "--print-address")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
	})
	address, err := bufio.NewReader(stdout).ReadString('\n')
	if err != nil {
		t.Fatal(err)
	}
	address = strings.TrimSpace(address)

	connect := func() *dbus.Conn {
		conn, err := dbus.Connect(address)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = conn.Close() })
		return conn
	}

	serviceConn := connect()
	promptPath := dbus.ObjectPath("/org/freedesktop/secrets/prompt/p1")
	if err := serviceConn.Export(&fakeService{prompt: promptPath}, servicePath, serviceInterface); err != nil {
		t.Fatal(err)
	}
	prompt := &fakePrompt{conn: serviceConn, path: promptPath, dismissed: dismissed, result: result}
	if err := serviceConn.Export(prompt, promptPath, promptInterface); err != nil {
		t.Fatal(err)
	}
	if _, err := serviceConn.RequestName(serviceName, dbus.NameFlagDoNotQueue); err != nil {
		t.Fatal(err)
	}

	conn := connect()
	return &SecretService{conn, conn.Object(serviceName, servicePath)}
}

// TestUnlockPromptDismissed tests that Unlock fails when the user dismisses
// the prompt, even if the prompt result names an object (as KWallet 6 does).
func TestUnlockPromptDismissed(t *testing.T) {
	svc := newTestSecretService(t, true, []dbus.ObjectPath{"/"})

	err := svc.Unlock(loginCollectionAlias)
	if !errors.Is(err, ErrPromptDismissed) {
		t.Errorf("Expected error %v, got %v", ErrPromptDismissed, err)
	}
}

// TestUnlockPromptAccepted tests that Unlock succeeds when the user accepts
// the prompt.
func TestUnlockPromptAccepted(t *testing.T) {
	svc := newTestSecretService(t, false, []dbus.ObjectPath{loginCollectionAlias})

	if err := svc.Unlock(loginCollectionAlias); err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}
