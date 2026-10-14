//go:build !darwin

package keyring

func newSecurityFramework() (Keyring, error) {
	return nil, ErrUnsupportedBackend
}
