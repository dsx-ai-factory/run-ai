/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package test_v223_upgrade_migration

import (
	"os"
	"strconv"
	"strings"
	ginkgoTesting "testing"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/config"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/integdb"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	dbconfig "github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/config"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/utils/ptr"
)

const (
	migrationsDir = "../../migrations"

	// Newest migration v2.23 and main have in common.
	lastSharedV223Version = "2230018"
	// v2.23-only backport of RUN-39587 (forward-ported to main as 02240011). A database
	// upgraded from v2.23 reports this version, and golang-migrate refuses to run Up
	// when the current version has no file.
	v223BackportVersion = "2230019"
)

func TestV223UpgradeMigration(t *ginkgoTesting.T) {
	config.Init()
	RegisterFailHandler(Fail)
	conf, cleanup, setupErr := integdb.SetupPostgresDB(&integdb.IntegdbConfig{Schema: "workloads", MigrationDirPath: "file://" + migrationsDir})
	if setupErr != nil {
		t.Fatalf("failed to set up test postgres: %v", setupErr)
	}
	t.Cleanup(cleanup)
	// Each spec drives the migrator through its own version path.
	conf.Migration.Auto = ptr.To(false)
	conf.UseServiceSession = new(bool)
	if initErr := runaidb.Init(conf); initErr != nil {
		t.Fatalf("failed to init database: %v", initErr)
	}
	t.Cleanup(func() { _ = runaidb.Close() })

	RunSpecs(t, "v2.23 upgrade migration test")
}

var _ = Describe("Upgrade from a v2.23 database", func() {
	It("migrates up to the latest version from the v2.23 backport version", func() {
		migrate(dbconfig.MigrationDirectionUp, lastSharedV223Version)
		migrate(dbconfig.MigrationDirectionForce, v223BackportVersion)

		Expect(runaidb.Migrate(&runaidb.MigrationParams{Direction: dbconfig.MigrationDirectionUp})).To(Succeed())

		Expect(currentVersion()).To(Equal(latestUpVersion()))
	})
})

func migrate(direction dbconfig.MigrationDirection, version string) {
	Expect(runaidb.Migrate(&runaidb.MigrationParams{Direction: direction, TargetVersion: ptr.To(version)})).To(Succeed())
}

func currentVersion() uint64 {
	var version uint64
	Expect(runaidb.DB().Actual().Raw("SELECT version FROM workloads.schema_migrations").Scan(&version).Error).To(Succeed())
	return version
}

func latestUpVersion() uint64 {
	entries, err := os.ReadDir(migrationsDir + "/up")
	Expect(err).ToNot(HaveOccurred())
	var latest uint64
	for _, entry := range entries {
		prefix, _, _ := strings.Cut(entry.Name(), "_")
		version, err := strconv.ParseUint(prefix, 10, 64)
		Expect(err).ToNot(HaveOccurred(), "migration file %q has no numeric version prefix", entry.Name())
		latest = max(latest, version)
	}
	return latest
}
