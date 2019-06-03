package keyring

import (
	"errors"
	"runtime"
)

// All of the following method errors out on unsupported platforms
var errUnsupportedPlatform = errors.New("Unsupported platform: " + runtime.GOOS)

type fallbackServiceProvider struct{}

func (fallbackServiceProvider) Set(service, user, pass string) error {
	return errUnsupportedPlatform
}

func (fallbackServiceProvider) Get(service, user string) (string, error) {
	return "", errUnsupportedPlatform
}

func (fallbackServiceProvider) Delete(service, user string) error {
	return errUnsupportedPlatform
}
