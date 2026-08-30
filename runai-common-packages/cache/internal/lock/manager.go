package lock

import (
	"context"
	"time"

	"github.com/run-ai/runai/runai-common-packages/cache/v2/options"
)

type Locker interface {
	Lock(ctx context.Context, ttl time.Duration) error
	Unlock(ctx context.Context) error
}

type manager struct {
	cacheClient          cacheClienter
	observabilityHandler observabilityHandlerer
	serviceLockStore     options.Store
	sharedLockStore      options.Store
	lockUniqueId         string
}

func NewManager(cacheClient cacheClienter, observabilityHandler observabilityHandlerer, serviceLockStore options.Store, sharedLockStore options.Store, lockUniqueId string) *manager {
	return &manager{
		cacheClient:          cacheClient,
		observabilityHandler: observabilityHandler,
		serviceLockStore:     serviceLockStore,
		sharedLockStore:      sharedLockStore,
		lockUniqueId:         lockUniqueId,
	}
}

func (l *manager) CreateLock(name string, shared bool) Locker {
	store := l.serviceLockStore
	if shared {
		store = l.sharedLockStore
	}

	return newLocker(l.cacheClient, l.observabilityHandler, store, l.lockUniqueId, name)
}
