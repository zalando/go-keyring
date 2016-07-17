package keyring

type windowsKeychain struct{}

// Get gets a secret from the keyring given a service name and a user. This
// method is currently a NO-OP on windows.
func (k windowsKeychain) Get(service, username string) (string, error) {
	return nil, ErrNotFound
}

// Set stores stores user and pass in the keyring under the defined service
// name. This method is currently a NO-OP on windows.
func (k windowsKeychain) Set(service, username, password string) error {
	return nil
}

// Delete deletes a secret, identified by service & user, from the keyring.
// This method is currently a NO-OP on windows.
func (k MacOSXKeychain) Delete(service, username string) error {
	return nil
}

func init() {
	provider = windowsKeychain{}
}
