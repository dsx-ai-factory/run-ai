package testing

import "github.com/run-ai/runai/runai-common-packages/cache/v2/internal/lock"

type locksMockManager struct {
	t          TestingT
	mockLocker *MockLocker
}

func newLocksMockManager(t TestingT, mockLocker *MockLocker) *locksMockManager {
	return &locksMockManager{
		t:          t,
		mockLocker: mockLocker,
	}
}

func (m *locksMockManager) CreateLock(name string, shared bool) lock.Locker {
	return m.mockLocker
}
