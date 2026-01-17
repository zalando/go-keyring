package keyring

import (
	"runtime"
	"strings"
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

func TestSetTooLong(t *testing.T) {
	extraLongPassword := "ba" + strings.Repeat("na", 5000)
	err := Set(service, user, extraLongPassword)

	if runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
		// should fail on those platforms
		if err != ErrSetDataTooBig {
			t.Errorf("Should have failed, got: %s", err)
		}
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

// TestGetMultiline tests getting a multi-line password from the keyring
func TestGetUmlaut(t *testing.T) {
	umlautPassword := "at least on OSX üöäÜÖÄß will be encoded"
	err := Set(service, user, umlautPassword)
	if err != nil {
		t.Errorf("Should not fail, got: %s", err)
	}

	pw, err := Get(service, user)
	if err != nil {
		t.Errorf("Should not fail, got: %s", err)
	}

	if umlautPassword != pw {
		t.Errorf("Expected password %s, got %s", umlautPassword, pw)
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

// TestDeleteAll tests deleting all secrets for a given service.
func TestDeleteAll(t *testing.T) {
	// Set up multiple secrets for the same service
	err := Set(service, user, password)
	if err != nil {
		t.Errorf("Should not fail, got: %s", err)
	}

	err = Set(service, user+"2", password+"2")
	if err != nil {
		t.Errorf("Should not fail, got: %s", err)
	}

	// Delete all secrets for the service
	err = DeleteAll(service)
	if err != nil {
		t.Errorf("Should not fail, got: %s", err)
	}

	// Verify that all secrets for the service are deleted
	_, err = Get(service, user)
	if err != ErrNotFound {
		t.Errorf("Expected error ErrNotFound, got %s", err)
	}

	_, err = Get(service, user+"2")
	if err != ErrNotFound {
		t.Errorf("Expected error ErrNotFound, got %s", err)
	}

	// Verify that DeleteAll on an empty service doesn't cause an error
	err = DeleteAll(service)
	if err != nil {
		t.Errorf("Should not fail on empty service, got: %s", err)
	}
}

// TestDeleteAll with empty service name
func TestDeleteAllEmptyService(t *testing.T) {
	err := Set(service, user, password)

	if err != nil {
		t.Errorf("Should not fail, got: %s", err)
	}
	_ = DeleteAll("")
	_, err = Get(service, user)
	if err == ErrNotFound {
		t.Errorf("Should not have deleted secret from another service")
	}
}

// TestListUsers tests listing all users for a service.
func TestListUsers(t *testing.T) {
	// Set up multiple secrets for the same service
	const service2 = "test-service-list"
	err := Set(service2, "user1", "password1")
	if err != nil {
		t.Errorf("Should not fail, got: %s", err)
	}

	err = Set(service2, "user2", "password2")
	if err != nil {
		t.Errorf("Should not fail, got: %s", err)
	}

	err = Set(service2, "user3", "password3")
	if err != nil {
		t.Errorf("Should not fail, got: %s", err)
	}

	// List all users for the service
	users, err := ListUsers(service2)
	if err != nil {
		t.Errorf("Should not fail, got: %s", err)
	}

	// Verify we got all three users
	if len(users) != 3 {
		t.Errorf("Expected 3 users, got %d", len(users))
	}

	// Verify the users are correct (order doesn't matter)
	expectedUsers := map[string]bool{"user1": true, "user2": true, "user3": true}
	for _, user := range users {
		if !expectedUsers[user] {
			t.Errorf("Unexpected user: %s", user)
		}
		delete(expectedUsers, user)
	}

	if len(expectedUsers) > 0 {
		t.Errorf("Missing users: %v", expectedUsers)
	}

	// Clean up
	_ = DeleteAll(service2)
}

// TestListUsersEmpty tests listing users for a service with no secrets.
func TestListUsersEmpty(t *testing.T) {
	const nonExistentService = "non-existent-service-12345"
	users, err := ListUsers(nonExistentService)
	if err != nil {
		t.Errorf("Should not fail on empty service, got: %s", err)
	}

	if len(users) != 0 {
		t.Errorf("Expected 0 users for non-existent service, got %d", len(users))
	}
}

// TestListUsersSingleUser tests listing users for a service with a single user.
func TestListUsersSingleUser(t *testing.T) {
	const service3 = "test-service-single"
	err := Set(service3, "single-user", "password")
	if err != nil {
		t.Errorf("Should not fail, got: %s", err)
	}

	users, err := ListUsers(service3)
	if err != nil {
		t.Errorf("Should not fail, got: %s", err)
	}

	if len(users) != 1 {
		t.Errorf("Expected 1 user, got %d", len(users))
	}

	if users[0] != "single-user" {
		t.Errorf("Expected user 'single-user', got '%s'", users[0])
	}

	// Clean up
	_ = Delete(service3, "single-user")
}

// TestListUsersMultipleServices tests that ListUsers only returns users for the specified service.
func TestListUsersMultipleServices(t *testing.T) {
	const serviceA = "service-a"
	const serviceB = "service-b"

	// Set up users for service A
	_ = Set(serviceA, "userA1", "passwordA1")
	_ = Set(serviceA, "userA2", "passwordA2")

	// Set up users for service B
	_ = Set(serviceB, "userB1", "passwordB1")

	// List users for service A
	usersA, err := ListUsers(serviceA)
	if err != nil {
		t.Errorf("Should not fail, got: %s", err)
	}

	if len(usersA) != 2 {
		t.Errorf("Expected 2 users for service A, got %d", len(usersA))
	}

	// Verify service A users don't include service B users
	for _, user := range usersA {
		if user == "userB1" {
			t.Errorf("Service A should not include users from service B")
		}
	}

	// List users for service B
	usersB, err := ListUsers(serviceB)
	if err != nil {
		t.Errorf("Should not fail, got: %s", err)
	}

	if len(usersB) != 1 {
		t.Errorf("Expected 1 user for service B, got %d", len(usersB))
	}

	if usersB[0] != "userB1" {
		t.Errorf("Expected user 'userB1', got '%s'", usersB[0])
	}

	// Clean up
	_ = DeleteAll(serviceA)
	_ = DeleteAll(serviceB)
}

