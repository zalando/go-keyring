package keyring

type mockProviderItem struct {
	Value  string
	Locked bool
}

type mockProvider struct {
	mockStore map[string]map[string]*mockProviderItem
	mockError error
}

// Set stores user and pass in the keyring under the defined service
// name.
func (m *mockProvider) Set(service, user, pass string) error {
	if m.mockError != nil {
		return m.mockError
	}
	if m.mockStore == nil {
		m.mockStore = make(map[string]map[string]*mockProviderItem)
	}
	if m.mockStore[service] == nil {
		m.mockStore[service] = make(map[string]*mockProviderItem)
	}
	m.mockStore[service][user] = &mockProviderItem{Value: pass, Locked: true}
	return nil
}

// Get gets a secret from the keyring given a service name and a user.
func (m *mockProvider) Get(service, user string) (string, error) {
	if m.mockError != nil {
		return "", m.mockError
	}
	if b, ok := m.mockStore[service]; ok {
		if item, ok := b[user]; ok {
			if item.Locked {
				return "", ErrNotFound
			}
			_ = m.Lock(service, user)
			return item.Value, nil
		}
	}
	return "", ErrNotFound
}

// Delete deletes a secret, identified by service & user, from the keyring.
func (m *mockProvider) Delete(service, user string) error {
	if m.mockError != nil {
		return m.mockError
	}
	if m.mockStore != nil {
		if _, ok := m.mockStore[service]; ok {
			if item, ok := m.mockStore[service][user]; ok {
				if item.Locked {
					return ErrNotFound
				}
				delete(m.mockStore[service], user)
				return nil
			}
		}
	}
	return ErrNotFound
}

// Unlock unlocks item from the keyring given a service name and a user
func (m *mockProvider) Unlock(service, user string) error {
	if m.mockError != nil {
		return m.mockError
	}
	if m.mockStore != nil {
		if _, ok := m.mockStore[service]; ok {
			if item, ok := m.mockStore[service][user]; ok {
				item.Locked = false
				return nil
			}
		}
	}
	return ErrNotFound
}

// Lock locks item from the keyring given a service name and a user
func (m *mockProvider) Lock(service, user string) error {
	if m.mockError != nil {
		return m.mockError
	}
	if m.mockStore != nil {
		if _, ok := m.mockStore[service]; ok {
			if item, ok := m.mockStore[service][user]; ok {
				item.Locked = true
				return nil
			}
		}
	}
	return ErrNotFound
}

// MockInit sets the provider to a mocked memory store
func MockInit() {
	provider = &mockProvider{}
}

// MockInitWithError sets the provider to a mocked memory store
// that returns the given error on all operations
func MockInitWithError(err error) {
	provider = &mockProvider{mockError: err}
}
