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
