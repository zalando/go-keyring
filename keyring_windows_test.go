package keyring

import (
	"testing"
	"unicode/utf16"

	"github.com/danieljoos/wincred"
)

func utf16le(s string) []byte {
	var b []byte
	for _, u := range utf16.Encode([]rune(s)) {
		b = append(b, byte(u), byte(u>>8))
	}
	return b
}

func TestDecodeCredentialBlob(t *testing.T) {
	tests := []struct {
		name string
		blob []byte
		want string
	}{
		{"utf8", []byte("s3cret"), "s3cret"},
		{"utf8 non-ascii", []byte("pässwörd"), "pässwörd"},
		{"empty", nil, ""},
		{"utf16 ascii", utf16le("s3cret"), "s3cret"},
		{"utf16 non-ascii", utf16le("pässwörd"), "pässwörd"},
		{"utf16 cjk", utf16le("密码"), "密码"},
		{"utf16 surrogate pair", utf16le("key🔑"), "key🔑"},
		{"odd length", []byte("abc"), "abc"},
		{"unpaired surrogate", []byte{0x00, 0xd8, 0x41, 0x00}, "\x00\xd8A\x00"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := decodeCredentialBlob(tt.blob); got != tt.want {
				t.Errorf("decodeCredentialBlob(%q) = %q, want %q", tt.blob, got, tt.want)
			}
		})
	}
}

// A secret stored as UTF-16LE, the way cmdkey and the Credential Manager UI
// store it, is returned as text.
func TestGetUTF16Secret(t *testing.T) {
	cred := wincred.NewGenericCredential(provider.(windowsKeychain).credName(service, user))
	cred.UserName = user
	cred.CredentialBlob = utf16le(password)
	if err := cred.Write(); err != nil {
		t.Fatalf("Should not fail, got: %s", err)
	}
	defer func() { _ = Delete(service, user) }()

	pw, err := Get(service, user)
	if err != nil {
		t.Errorf("Should not fail, got: %s", err)
	}
	if pw != password {
		t.Errorf("Expected password %s, got %s", password, pw)
	}
}
