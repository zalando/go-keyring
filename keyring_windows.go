package keyring

import (
	"bytes"
	"encoding/binary"
	"slices"
	"sort"
	"strings"
	"syscall"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/danieljoos/wincred"
)

type windowsKeychain struct{}

// Get gets a secret from the keyring given a service name and a user.
func (k windowsKeychain) Get(service, username string) (string, error) {
	cred, err := wincred.GetGenericCredential(k.credName(service, username))
	if err != nil {
		if err == syscall.ERROR_NOT_FOUND {
			return "", ErrNotFound
		}
		return "", err
	}

	return decodeCredentialBlob(cred.CredentialBlob), nil
}

// decodeCredentialBlob returns the secret held in a credential blob. go-keyring
// stores UTF-8, but cmdkey, the Credential Manager UI and other tools (such as
// Python's keyring) store UTF-16LE. A blob is read as UTF-16LE when it has an
// even length, contains a NUL byte or is not valid UTF-8, and decodes to valid
// UTF-16; anything else is returned as is.
func decodeCredentialBlob(blob []byte) string {
	if len(blob) == 0 || len(blob)%2 != 0 || (bytes.IndexByte(blob, 0) < 0 && utf8.Valid(blob)) {
		return string(blob)
	}
	units := make([]uint16, len(blob)/2)
	for i := range units {
		units[i] = binary.LittleEndian.Uint16(blob[2*i:])
	}
	for i := 0; i < len(units); i++ {
		switch u := units[i]; {
		case utf16.IsSurrogate(rune(u)) && u < 0xdc00 && i+1 < len(units) && units[i+1] >= 0xdc00 && units[i+1] <= 0xdfff:
			i++ // valid surrogate pair
		case utf16.IsSurrogate(rune(u)):
			return string(blob)
		}
	}
	return string(utf16.Decode(units))
}

// Set stores stores user and pass in the keyring under the defined service
// name.
func (k windowsKeychain) Set(service, username, password string) error {
	// password may not exceed 2560 bytes (https://github.com/jaraco/keyring/issues/540#issuecomment-968329967)
	if len(password) > 2560 {
		return ErrSetDataTooBig
	}

	// service may not exceed 512 bytes (might need more testing)
	if len(service) >= 512 {
		return ErrSetDataTooBig
	}

	// service may not exceed 32k but problems occur before that
	// so we limit it to 30k
	if len(service) > 1024*30 {
		return ErrSetDataTooBig
	}

	cred := wincred.NewGenericCredential(k.credName(service, username))
	cred.UserName = username
	cred.CredentialBlob = []byte(password)
	return cred.Write()
}

// Delete deletes a secret, identified by service & user, from the keyring.
func (k windowsKeychain) Delete(service, username string) error {
	cred, err := wincred.GetGenericCredential(k.credName(service, username))
	if err != nil {
		if err == syscall.ERROR_NOT_FOUND {
			return ErrNotFound
		}
		return err
	}

	return cred.Delete()
}

func (k windowsKeychain) DeleteAll(service string) error {
	// if service is empty, do nothing otherwise it might accidentally delete all secrets
	if service == "" {
		return ErrNotFound
	}

	creds, err := wincred.List()
	if err != nil {
		return err
	}

	prefix := k.credName(service, "")
	deletedCount := 0

	for _, cred := range creds {
		if strings.HasPrefix(cred.TargetName, prefix) {
			genericCred, err := wincred.GetGenericCredential(cred.TargetName)
			if err != nil {
				if err != syscall.ERROR_NOT_FOUND {
					return err
				}
			} else {
				err := genericCred.Delete()
				if err != nil {
					return err
				}
				deletedCount++
			}
		}
	}
	return nil
}

// ListUsers returns a list of all users for a given service
func (k windowsKeychain) ListUsers(service string) ([]string, error) {
	if service == "" {
		return []string{}, nil
	}

	creds, err := wincred.List()
	if err != nil {
		return nil, err
	}

	prefix := k.credName(service, "")
	var users []string

	for _, cred := range creds {
		if strings.HasPrefix(cred.TargetName, prefix) {
			username := strings.TrimPrefix(cred.TargetName, prefix)
			if username != "" {
				users = append(users, username)
			}
		}
	}
	sort.Strings(users)
	return slices.Compact(users), nil
}

// credName combines service and username to a single string.
func (k windowsKeychain) credName(service, username string) string {
	return service + ":" + username
}

func init() {
	p := windowsKeychain{}
	provider = p
	restoreProvider = func() { provider = p }
}
