/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package testing

import (
	"sync"
	"testing"

	org_unit_config "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/config"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/integdb"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
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
func SharedOpenPostgresForTesting(t *testing.T) {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	// Wait for any ongoing cleanup to complete
	for isCleaning {
		dbCond.Wait()
	}

	// Only initialize if not already initialized
	if !isInitialized {
		conf, cleanup, setupErr := integdb.SetupPostgresDB(&integdb.IntegdbConfig{
			Schema:           org_unit_config.Schema,
			MigrationDirPath: "file://../database/migrations",
		})
		if setupErr != nil {
			t.Fatalf("failed to set up test postgres: %v", setupErr)
		}
		t.Cleanup(cleanup)
		conf.FullSaveAssociations = true
		conf.SkipDefaultTransaction = true
		if err := runaidb.Init(conf); err != nil {
			t.Fatalf("failed to init database: %v", err)
		}
		t.Cleanup(func() { _ = runaidb.Close() })
		isInitialized = true
	}
}
