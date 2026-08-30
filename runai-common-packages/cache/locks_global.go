package cache

import (
	"fmt"
	"time"

	"github.com/lithammer/shortuuid/v4"
	cache_config "github.com/run-ai/runai/runai-common-packages/cache/v2/config"
	"github.com/run-ai/runai/runai-common-packages/cache/v2/internal/lock"
	"github.com/run-ai/runai/runai-common-packages/cache/v2/options"
	"github.com/run-ai/runai/runai-common-packages/config"
)

var (
	lockManagerInstance lockManagerer
)

func initLocks(cacheClient cacheClienter, observabilityHandler observabilityHandlerer) error {
	storeOpts := &options.StoreOptions{
		Ttl:             24 * time.Hour,
		MaxStoreSizeMB:  10,
		MaxRecordSizeKB: 1,
		ReadOnly:        false,
	}
	serviceLockStore := options.Store{
		Name:    fmt.Sprintf("locks-%s", config.Get[*cache_config.Cache]().ServiceName),
		Options: storeOpts,
	}
	if err := cacheClient.Store(serviceLockStore); err != nil {
		return fmt.Errorf("failed to create service lock store: %w", err)
	}
	sharedLockStore := options.Store{
		Name:    "locks-shared",
		Options: storeOpts,
	}
	if err := cacheClient.Store(sharedLockStore); err != nil {
		return fmt.Errorf("failed to create shared lock store: %w", err)
	}
	lockUniqueId := shortuuid.New()
	lockManagerInstance = lock.NewManager(cacheClient, observabilityHandler, serviceLockStore, sharedLockStore, lockUniqueId)
	return nil
}

func Lock(name string) lock.Locker {
	return lockManagerInstance.CreateLock(name, false)
}

func SharedLock(name string) lock.Locker {
	return lockManagerInstance.CreateLock(name, true)
}

type lockManagerer interface {
	CreateLock(name string, shared bool) lock.Locker
}

func SetLocksMock(lockManager lockManagerer) {
	lockManagerInstance = lockManager
}
