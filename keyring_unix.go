//go:build (dragonfly && cgo) || (freebsd && cgo) || netbsd || openbsd

package keyring

func init() {
	p := secretServiceProvider{}
	provider = p
	restoreProvider = func() { provider = p }
}
