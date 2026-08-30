package auth

import (
	"context"
	"errors"
	"sync"
)

var (
	internalAccessTokenGetter *Manager
	single                    sync.Once
)

type TokenService interface {
	GetAccessToken(ctx context.Context) (string, error)
}

// GetManager returns the initialized AccessTokenManager.
func GetManager() *Manager {
	single.Do(func() {
		internalAccessTokenGetter = &Manager{}
	})
	return internalAccessTokenGetter
}

type ConfigGetter interface {
	GetConfigValue(key string) string
	GetAuthenticationPath() string
}

type ServiceCreatorFn func(config ConfigGetter) (TokenService, error)

type Manager struct {
	config           ConfigGetter
	backendCreatorFn ServiceCreatorFn
	internalErr      error
	lazyInitOnce     sync.Once
	backend          TokenService
}

// Init initializes the TokenService only once.
// this a workaround to test the command where the backend is mocked in tests
func (m *Manager) Init(creator ServiceCreatorFn, getter ConfigGetter) {
	m.config = getter
	m.backendCreatorFn = creator
	m.internalErr = nil
	m.lazyInitOnce = sync.Once{}

}

func (m *Manager) IsInitialized() bool {
	return m.backendCreatorFn != nil
}
func (m *Manager) initBackend() (TokenService, error) {
	m.lazyInitOnce.Do(func() {
		backend, err := m.backendCreatorFn(m.config)
		if err != nil {
			m.internalErr = err
			return
		}
		m.backend = backend
	})
	return m.backend, m.internalErr
}

func (m *Manager) GetAccessToken(ctx context.Context) (string, error) {
	backend, err := m.initBackend()
	if err != nil {
		return "", err
	}
	return backend.GetAccessToken(ctx)
}

type StoredTokenError struct {
	Message string
}

func (e *StoredTokenError) Error() string {
	return e.Message
}

func NewStoredTokenError(message string) *StoredTokenError {
	return &StoredTokenError{Message: message}
}

// IsStoredTokenError checks if the error is of type StoredTokenError.
func IsStoredTokenError(err error) bool {
	var storedTokenError *StoredTokenError
	return errors.As(err, &storedTokenError)
}
