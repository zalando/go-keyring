package keyring

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"
	"time"
)

// newFrameworkKeyring returns a Security framework Keyring and a service name
// that is unique to this test run. Keychain items are owned by the application
// that created them, so a name shared with an earlier run (and thus an earlier
// test binary) could make macOS ask for confirmation.
func newFrameworkKeyring(t *testing.T) (Keyring, string) {
	t.Helper()
	k, err := New(Options{Backend: BackendSecurityFramework})
	if err != nil {
		t.Fatalf("Should not fail, got: %s", err)
	}
	svc := fmt.Sprintf("test-service-framework-%d-%d", os.Getpid(), time.Now().UnixNano())
	t.Cleanup(func() { _ = k.DeleteAll(svc) })
	return k, svc
}

func TestFrameworkSetGet(t *testing.T) {
	k, svc := newFrameworkKeyring(t)

	if err := k.Set(svc, user, password); err != nil {
		t.Fatalf("Should not fail, got: %s", err)
	}
	pw, err := k.Get(svc, user)
	if err != nil || pw != password {
		t.Errorf("Expected password %q, got %q, %v", password, pw, err)
	}

	// Setting again overwrites.
	if err := k.Set(svc, user, "other"); err != nil {
		t.Fatalf("Should not fail, got: %s", err)
	}
	pw, err = k.Get(svc, user)
	if err != nil || pw != "other" {
		t.Errorf("Expected password %q, got %q, %v", "other", pw, err)
	}
}

func TestFrameworkSpecialValues(t *testing.T) {
	k, svc := newFrameworkKeyring(t)

	for name, value := range map[string]string{
		"multiline": "this password\ncontains\nmultiple lines",
		"umlaut":    "at least on OSX üöäÜÖÄß will be encoded",
		"hex":       "abcdef123abcdef123",
		"large":     strings.Repeat("na", 5000), // beyond the limit of the security command
	} {
		if err := k.Set(svc, name, value); err != nil {
			t.Errorf("%s: should not fail, got: %s", name, err)
			continue
		}
		pw, err := k.Get(svc, name)
		if err != nil || pw != value {
			t.Errorf("%s: expected %q, got %q, %v", name, value, pw, err)
		}
	}
}

func TestFrameworkGetNonExisting(t *testing.T) {
	k, svc := newFrameworkKeyring(t)

	if _, err := k.Get(svc, user+"fake"); err != ErrNotFound {
		t.Errorf("Expected error ErrNotFound, got %v", err)
	}
}

func TestFrameworkDelete(t *testing.T) {
	k, svc := newFrameworkKeyring(t)

	if err := k.Set(svc, user, password); err != nil {
		t.Fatalf("Should not fail, got: %s", err)
	}
	if err := k.Delete(svc, user); err != nil {
		t.Errorf("Should not fail, got: %s", err)
	}
	if _, err := k.Get(svc, user); err != ErrNotFound {
		t.Errorf("Expected error ErrNotFound, got %v", err)
	}
	if err := k.Delete(svc, user); err != ErrNotFound {
		t.Errorf("Expected error ErrNotFound, got %v", err)
	}
}

func TestFrameworkDeleteAll(t *testing.T) {
	k, svc := newFrameworkKeyring(t)

	for _, u := range []string{"user1", "user2", "user3"} {
		if err := k.Set(svc, u, password); err != nil {
			t.Fatalf("Should not fail, got: %s", err)
		}
	}
	if err := k.DeleteAll(svc); err != nil {
		t.Errorf("Should not fail, got: %s", err)
	}
	for _, u := range []string{"user1", "user2", "user3"} {
		if _, err := k.Get(svc, u); err != ErrNotFound {
			t.Errorf("%s: expected error ErrNotFound, got %v", u, err)
		}
	}
	// Nothing left to delete is not an error.
	if err := k.DeleteAll(svc); err != nil {
		t.Errorf("Should not fail, got: %s", err)
	}
}

func TestFrameworkDeleteAllEmptyService(t *testing.T) {
	k, _ := newFrameworkKeyring(t)

	if err := k.DeleteAll(""); err != ErrNotFound {
		t.Errorf("Expected error ErrNotFound, got %v", err)
	}
}

func TestFrameworkListUsers(t *testing.T) {
	k, svc := newFrameworkKeyring(t)
	other := svc + "-other"
	t.Cleanup(func() { _ = k.DeleteAll(other) })

	users, err := k.ListUsers(svc)
	if err != nil || len(users) != 0 {
		t.Errorf("Expected no users, got %v, %v", users, err)
	}

	for _, u := range []string{"bob", "alice", "ünïcode"} {
		if err := k.Set(svc, u, password); err != nil {
			t.Fatalf("Should not fail, got: %s", err)
		}
	}
	if err := k.Set(other, "carol", password); err != nil {
		t.Fatalf("Should not fail, got: %s", err)
	}

	users, err = k.ListUsers(svc)
	if err != nil {
		t.Fatalf("Should not fail, got: %s", err)
	}
	sort.Strings(users)
	want := []string{"alice", "bob", "ünïcode"}
	if strings.Join(users, ",") != strings.Join(want, ",") {
		t.Errorf("Expected users %v, got %v", want, users)
	}

	if users, err := k.ListUsers(""); err != nil || len(users) != 0 {
		t.Errorf("Expected no users for empty service, got %v, %v", users, err)
	}
}
