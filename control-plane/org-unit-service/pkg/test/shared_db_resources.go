package testing

import (
	"sync"
	"testing"

	dao_utils "github.com/run-ai/runai/runai-common-packages/dao-utils"
)

var (
	activeTests   int
	testMutex     sync.Mutex
	dbMutex       sync.Mutex
	dbCond        *sync.Cond
	isInitialized bool
	isCleaning    bool
)

func init() {
	dbCond = sync.NewCond(&dbMutex)
}

// RegisterTest marks that a test has started
// This should be called at the beginning of each test
func RegisterTest() {
	testMutex.Lock()
	defer testMutex.Unlock()
	activeTests++
}

// UnregisterTest marks that a test has finished
// This should be called at the end of each test
func UnregisterTest() {
	if activeTests == 0 {
		return
	}
	testMutex.Lock()
	defer testMutex.Unlock()
	activeTests--

	// If this was the last test, trigger cleanup
	if activeTests == 0 {
		SharedCleanup()
	}
}

// SharedCleanup ensures that table cleanup only happens once
// This prevents conflicts when multiple tests run together
func SharedCleanup() {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	// Only cleanup if database is initialized and not already cleaning
	if isInitialized && !isCleaning {
		isCleaning = true
		CleanTables()
		isInitialized = false
		isCleaning = false
		// Signal any waiting initialization operations
		dbCond.Broadcast()
	}
}

// SharedOpenPostgresForTesting ensures that database initialization only happens once
// This prevents conflicts when multiple tests try to initialize the database simultaneously
func SharedOpenPostgresForTesting(t *testing.T, conf *dao_utils.DataBaseConfig) {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	// Wait for any ongoing cleanup to complete
	for isCleaning {
		dbCond.Wait()
	}

	// Only initialize if not already initialized
	if !isInitialized {
		dao_utils.OpenPostgresForTesting(t, conf)
		isInitialized = true
	}
}
