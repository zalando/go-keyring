//go:build darwin

package darwin

import (
	"unsafe"

	"github.com/ebitengine/purego"
)

const (
	kCFStringEncodingUTF8 = 0x08000100
	kCFAllocatorDefault   = 0
)

type osStatus int32

const (
	errSecSuccess       osStatus = 0      // No error.
	errSecDuplicateItem osStatus = -25299 // The specified item already exists in the keychain.
	errSecItemNotFound  osStatus = -25300 // The specified item could not be found in the keychain.
)

type _CFRange struct {
	location int64
	length   int64
}

var (
	kCFTypeDictionaryKeyCallBacks   uintptr
	kCFTypeDictionaryValueCallBacks uintptr
	kCFBooleanTrue                  uintptr
)

var (
	kSecClass                uintptr
	kSecClassGenericPassword uintptr
	kSecAttrService          uintptr
	kSecAttrAccount          uintptr
	kSecValueData            uintptr
	kSecReturnData           uintptr
	kSecMatchLimit           uintptr
	kSecMatchLimitAll        uintptr
)

var (
	CFDictionaryCreate        func(allocator uintptr, keys, values *uintptr, numValues int64, keyCallBacks, valueCallBacks uintptr) uintptr
	CFStringCreateWithCString func(allocator uintptr, cStr string, encoding uint32) uintptr
	CFDataCreate              func(alloc uintptr, bytes []byte, length int64) uintptr
	CFDataGetLength           func(theData uintptr) int64
	CFDataGetBytes            func(theData uintptr, theRange _CFRange, buffer []byte)
	CFRelease                 func(cf uintptr)
)

var (
	SecItemCopyMatching func(query uintptr, result *uintptr) osStatus
	SecItemAdd          func(query uintptr, result uintptr) osStatus
	SecItemUpdate       func(query uintptr, attributesToUpdate uintptr) osStatus
	SecItemDelete       func(query uintptr) osStatus
)

func init() {
	cfLib := must(purego.Dlopen("/System/Library/Frameworks/CoreFoundation.framework/CoreFoundation", purego.RTLD_NOW|purego.RTLD_GLOBAL))

	kCFTypeDictionaryKeyCallBacks = must(purego.Dlsym(cfLib, "kCFTypeDictionaryKeyCallBacks"))
	kCFTypeDictionaryValueCallBacks = must(purego.Dlsym(cfLib, "kCFTypeDictionaryValueCallBacks"))
	kCFBooleanTrue = deref(must(purego.Dlsym(cfLib, "kCFBooleanTrue")))

	purego.RegisterLibFunc(&CFDictionaryCreate, cfLib, "CFDictionaryCreate")
	purego.RegisterLibFunc(&CFStringCreateWithCString, cfLib, "CFStringCreateWithCString")
	purego.RegisterLibFunc(&CFDataCreate, cfLib, "CFDataCreate")
	purego.RegisterLibFunc(&CFDataGetLength, cfLib, "CFDataGetLength")
	purego.RegisterLibFunc(&CFDataGetBytes, cfLib, "CFDataGetBytes")
	purego.RegisterLibFunc(&CFRelease, cfLib, "CFRelease")

	secLib := must(purego.Dlopen("/System/Library/Frameworks/Security.framework/Security", purego.RTLD_NOW|purego.RTLD_GLOBAL))

	kSecClass = deref(must(purego.Dlsym(secLib, "kSecClass")))
	kSecClassGenericPassword = deref(must(purego.Dlsym(secLib, "kSecClassGenericPassword")))
	kSecAttrService = deref(must(purego.Dlsym(secLib, "kSecAttrService")))
	kSecAttrAccount = deref(must(purego.Dlsym(secLib, "kSecAttrAccount")))
	kSecValueData = deref(must(purego.Dlsym(secLib, "kSecValueData")))
	kSecReturnData = deref(must(purego.Dlsym(secLib, "kSecReturnData")))
	kSecMatchLimit = deref(must(purego.Dlsym(secLib, "kSecMatchLimit")))
	kSecMatchLimitAll = deref(must(purego.Dlsym(secLib, "kSecMatchLimitAll")))

	purego.RegisterLibFunc(&SecItemCopyMatching, secLib, "SecItemCopyMatching")
	purego.RegisterLibFunc(&SecItemAdd, secLib, "SecItemAdd")
	purego.RegisterLibFunc(&SecItemUpdate, secLib, "SecItemUpdate")
	purego.RegisterLibFunc(&SecItemDelete, secLib, "SecItemDelete")
}

func deref(ptr uintptr) uintptr {
	// We take the address and then dereference it to trick
	// go vet from creating a possible miss-use of unsafe.Pointer
	// See https://github.com/golang/go/issues/41205
	return **(**uintptr)(unsafe.Pointer(&ptr))
}

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}
