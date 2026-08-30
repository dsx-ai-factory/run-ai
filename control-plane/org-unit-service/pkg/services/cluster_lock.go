package services

import (
	"sync"
)

// ClusterLockManager manages read-write locks per cluster ID.
// This allows multiple CreateProject/CreateDepartment operations to run concurrently (read lock),
// while NodePoolUpdated takes an exclusive write lock that blocks all CreateProject/CreateDepartment
// operations for that cluster.
type ClusterLockManager struct {
	locks sync.Map // map[string]*sync.RWMutex
}

var clusterLockManager = &ClusterLockManager{}

// GetClusterLockManager returns the singleton cluster lock manager
func GetClusterLockManager() *ClusterLockManager {
	return clusterLockManager
}

// getOrCreateLock gets or creates a RWMutex for the given cluster ID
func (m *ClusterLockManager) getOrCreateLock(clusterId string) *sync.RWMutex {
	if lock, ok := m.locks.Load(clusterId); ok {
		return lock.(*sync.RWMutex)
	}

	newLock := &sync.RWMutex{}
	actual, _ := m.locks.LoadOrStore(clusterId, newLock)
	return actual.(*sync.RWMutex)
}

// RLock acquires a read lock for the given cluster ID.
// Multiple goroutines can hold the read lock simultaneously.
func (m *ClusterLockManager) RLock(clusterId string) {
	m.getOrCreateLock(clusterId).RLock()
}

// RUnlock releases the read lock for the given cluster ID.
func (m *ClusterLockManager) RUnlock(clusterId string) {
	m.getOrCreateLock(clusterId).RUnlock()
}

// Lock acquires an exclusive write lock for the given cluster ID.
// This blocks until all read locks are released and prevents new read locks.
// Use this in NodePoolUpdated to ensure exclusive access.
func (m *ClusterLockManager) Lock(clusterId string) {
	m.getOrCreateLock(clusterId).Lock()
}

// Unlock releases the exclusive write lock for the given cluster ID.
func (m *ClusterLockManager) Unlock(clusterId string) {
	m.getOrCreateLock(clusterId).Unlock()
}
