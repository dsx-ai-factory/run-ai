/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"path/filepath"
	"runtime"
	"testing"

	authClient "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/auth"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/pkg/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/pkg/settings"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/integdb"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	setting_pkg "github.com/dsx-ai-factory/run-ai/runai-common-packages/setting"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const DB_SCHEMA_FOR_TEST = "workloads_manager"

var authMock = auth.NewSimpleMockAuthService()

func TestCrud(t *testing.T) {
	RegisterFailHandler(Fail)

	config.Init()
	RegisterSuspendDefCacheStore()

	conf, cleanup, err := integdb.SetupPostgresDB(&integdb.IntegdbConfig{
		Schema:           DB_SCHEMA_FOR_TEST,
		MigrationDirPath: migrationDirURL(),
	})
	if err != nil {
		t.Fatalf("failed to set up test postgres: %v", err)
	}
	t.Cleanup(cleanup)

	// tests run against a single-user DB; use the owner role (not runaidb's limited
	// service session) so the test role can read/write the migrated tables.
	useServiceSession := false
	conf.UseServiceSession = &useServiceSession
	conf.SkipDefaultTransaction = true

	// Close resets runaidb's sync.Once so a failed/empty Init from another
	// package init cannot make this Init a silent no-op with a nil DB().
	_ = runaidb.Close()
	runaidb.SetLogger(logging.Actual())
	if err := runaidb.Init(conf); err != nil {
		t.Fatalf("failed to init database: %v", err)
	}
	t.Cleanup(func() { _ = runaidb.Close() })

	if runaidb.DB() == nil {
		t.Fatal("runaidb.DB() is nil after Init")
	}
	if _, err := runaidb.DB().Exec("SELECT 1"); err != nil {
		t.Fatalf("runaidb ping failed: %v", err)
	}

	setting_pkg.Init(nil, nil)
	if err := setting_pkg.Register[*settings.YamlSubmission](); err != nil {
		t.Fatalf("failed to register yaml submission setting: %v", err)
	}

	authClient.SetMockAuthorizationClient(authMock)

	RunSpecs(t, "workloads-manager")

	t.Cleanup(func() {
		cleanTables()
	})
}

func migrationDirURL() string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return "file://../database/migrations"
	}
	return "file://" + filepath.ToSlash(filepath.Join(filepath.Dir(filename), "..", "database", "migrations"))
}
