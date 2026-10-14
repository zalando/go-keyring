//go:build darwin

package darwin

import (
	"errors"
	"fmt"
)

var ErrNotFound = errors.New("secret not found in keyring")

func FindGenericPassword(service, username string) (string, error) {
	cfService := CFStringCreateWithCString(kCFAllocatorDefault, service, kCFStringEncodingUTF8)
	defer CFRelease(cfService)

	cfAccount := CFStringCreateWithCString(kCFAllocatorDefault, username, kCFStringEncodingUTF8)
	defer CFRelease(cfAccount)

	keys := []uintptr{
		kSecClass,
		kSecAttrService,
		kSecAttrAccount,
		kSecReturnData,
	}
	values := []uintptr{
		kSecClassGenericPassword,
		cfService,
		cfAccount,
		kCFBooleanTrue,
	}

	query := CFDictionaryCreate(
		kCFAllocatorDefault,
		&keys[0], &values[0], int64(len(keys)),
		kCFTypeDictionaryKeyCallBacks,
		kCFTypeDictionaryValueCallBacks,
	)
	defer CFRelease(query)

	var data uintptr
	st := SecItemCopyMatching(query, &data)
	if st == errSecItemNotFound {
		return "", ErrNotFound
	} else if st != errSecSuccess {
		return "", fmt.Errorf("error SecItemCopyMatching: %d", st)
	}
	defer CFRelease(data)

	length := CFDataGetLength(data)
	if length < 0 {
		return "", fmt.Errorf("error CFDataGetLength: %d", length)
	}

	buffer := make([]byte, length)
	CFDataGetBytes(data, _CFRange{0, length}, buffer)

	return string(buffer), nil
}

func AddGenericPassword(service, username, password string) error {
	cfService := CFStringCreateWithCString(kCFAllocatorDefault, service, kCFStringEncodingUTF8)
	defer CFRelease(cfService)

	cfAccount := CFStringCreateWithCString(kCFAllocatorDefault, username, kCFStringEncodingUTF8)
	defer CFRelease(cfAccount)

	cfPasswordData := CFDataCreate(kCFAllocatorDefault, []byte(password), int64(len(password)))
	defer CFRelease(cfPasswordData)

	keys := []uintptr{
		kSecClass,
		kSecAttrService,
		kSecAttrAccount,
		kSecValueData,
	}
	values := []uintptr{
		kSecClassGenericPassword,
		cfService,
		cfAccount,
		cfPasswordData,
	}

	query := CFDictionaryCreate(
		kCFAllocatorDefault,
		&keys[0], &values[0], int64(len(keys)),
		kCFTypeDictionaryKeyCallBacks,
		kCFTypeDictionaryValueCallBacks,
	)
	defer CFRelease(query)

	sa := SecItemAdd(query, 0)
	if sa == errSecDuplicateItem {
		su := SecItemUpdate(query, query)
		if su != errSecSuccess {
			return fmt.Errorf("error SecItemUpdate: %d", su)
		}
	} else if sa != errSecSuccess {
		return fmt.Errorf("error SecItemAdd: %d", sa)
	}
	return nil
}

func DeleteGenericPassword(service, username string) error {
	cfService := CFStringCreateWithCString(kCFAllocatorDefault, service, kCFStringEncodingUTF8)
	defer CFRelease(cfService)

	cfAccount := CFStringCreateWithCString(kCFAllocatorDefault, username, kCFStringEncodingUTF8)
	defer CFRelease(cfAccount)

	keys := []uintptr{
		kSecClass,
		kSecAttrService,
		kSecAttrAccount,
	}
	values := []uintptr{
		kSecClassGenericPassword,
		cfService,
		cfAccount,
	}

	query := CFDictionaryCreate(
		kCFAllocatorDefault,
		&keys[0], &values[0], int64(len(keys)),
		kCFTypeDictionaryKeyCallBacks,
		kCFTypeDictionaryValueCallBacks,
	)
	defer CFRelease(query)

	st := SecItemDelete(query)
	if st == errSecItemNotFound {
		return ErrNotFound
	} else if st != errSecSuccess {
		return fmt.Errorf("error SecItemDelete: %d", st)
	}
	return nil
}

func DeleteGenericPasswords(service string) error {
	cfService := CFStringCreateWithCString(kCFAllocatorDefault, service, kCFStringEncodingUTF8)
	defer CFRelease(cfService)

	keys := []uintptr{
		kSecClass,
		kSecAttrService,
		kSecMatchLimit,
	}
	values := []uintptr{
		kSecClassGenericPassword,
		cfService,
		kSecMatchLimitAll,
	}

	query := CFDictionaryCreate(
		kCFAllocatorDefault,
		&keys[0], &values[0], int64(len(keys)),
		kCFTypeDictionaryKeyCallBacks,
		kCFTypeDictionaryValueCallBacks,
	)
	defer CFRelease(query)

	st := SecItemDelete(query)
	if st == errSecItemNotFound {
		return nil
	} else if st != errSecSuccess {
		return fmt.Errorf("error SecItemDelete: %d", st)
	}
	return nil
}

// ListGenericPasswordAccounts returns the distinct account names of all
// generic passwords stored for service.
func ListGenericPasswordAccounts(service string) ([]string, error) {
	cfService := CFStringCreateWithCString(kCFAllocatorDefault, service, kCFStringEncodingUTF8)
	defer CFRelease(cfService)

	keys := []uintptr{
		kSecClass,
		kSecAttrService,
		kSecMatchLimit,
		kSecReturnAttributes,
	}
	values := []uintptr{
		kSecClassGenericPassword,
		cfService,
		kSecMatchLimitAll,
		kCFBooleanTrue,
	}

	query := CFDictionaryCreate(
		kCFAllocatorDefault,
		&keys[0], &values[0], int64(len(keys)),
		kCFTypeDictionaryKeyCallBacks,
		kCFTypeDictionaryValueCallBacks,
	)
	defer CFRelease(query)

	var result uintptr
	st := SecItemCopyMatching(query, &result)
	if st == errSecItemNotFound {
		return []string{}, nil
	} else if st != errSecSuccess {
		return nil, fmt.Errorf("error SecItemCopyMatching: %d", st)
	}
	defer CFRelease(result)

	users := []string{}
	seen := map[string]bool{}
	for i, n := int64(0), CFArrayGetCount(result); i < n; i++ {
		attrs := CFArrayGetValueAtIndex(result, i)
		account := CFDictionaryGetValue(attrs, kSecAttrAccount)
		if account == 0 {
			continue
		}
		user, err := goString(account)
		if err != nil {
			return nil, err
		}
		if !seen[user] {
			seen[user] = true
			users = append(users, user)
		}
	}
	return users, nil
}

// goString copies a CFString into a Go string.
func goString(cfString uintptr) (string, error) {
	size := CFStringGetMaximumSizeForEncoding(CFStringGetLength(cfString), kCFStringEncodingUTF8) + 1
	buffer := make([]byte, size)
	if !CFStringGetCString(cfString, buffer, size, kCFStringEncodingUTF8) {
		return "", errors.New("error CFStringGetCString")
	}
	for i, b := range buffer {
		if b == 0 {
			return string(buffer[:i]), nil
		}
	}
	return string(buffer), nil
}
