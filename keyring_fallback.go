package keyring

import (
	"runtime"

	"github.com/zalando/go-keyring/errors"
)

// All of the following methods error out on unsupported platforms
const (
	ErrUnsupportedPlatform = errors.KeyringError("Unsupported platform: " + runtime.GOOS)
)

type fallbackServiceProvider struct{}

func (fallbackServiceProvider) Set(service, user, pass string) error {
	return ErrUnsupportedPlatform
}

func (fallbackServiceProvider) Get(service, user string) (string, error) {
	return "", ErrUnsupportedPlatform
}

func (fallbackServiceProvider) Delete(service, user string) error {
	return ErrUnsupportedPlatform
}
