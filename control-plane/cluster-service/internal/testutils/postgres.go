/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package testutils

import (
	"encoding/json"
	"fmt"

	. "github.com/onsi/ginkgo/v2" //nolint:staticcheck // dot-import is the idiomatic ginkgo usage in test helpers
	. "github.com/onsi/gomega"    //nolint:staticcheck // dot-import is the idiomatic gomega usage in test helpers
	"github.com/run-ai/runai/runai-common-packages/integdb"
	"github.com/run-ai/runai/runai-common-packages/runaidb"
	dbconfig "github.com/run-ai/runai/runai-common-packages/runaidb/config"
)

// terminateContainer is only set on ginkgo process 1, which owns the container.
var terminateContainer func()

// RegisterPostgresSuite starts one postgres container for the whole suite and
// initializes runaidb per ginkgo process against its own schema, so parallel
// processes never share tables.
func RegisterPostgresSuite(schema, migrationDirPath string) {
	SynchronizedBeforeSuite(func() []byte {
		conf, cleanup, err := integdb.SetupPostgresDB(&integdb.IntegdbConfig{
			Schema:           schema,
			MigrationDirPath: migrationDirPath,
		})
		Expect(err).NotTo(HaveOccurred())
		terminateContainer = cleanup

		payload, err := json.Marshal(conf)
		Expect(err).NotTo(HaveOccurred())
		return payload
	}, func(payload []byte) {
		conf := &dbconfig.Config{}
		Expect(json.Unmarshal(payload, conf)).To(Succeed())

		// Schema and service-session role must be unique per process, otherwise
		// parallel processes race on migrations and on CREATE ROLE.
		conf.Connection.Schema = fmt.Sprintf("%s_%d", schema, GinkgoParallelProcess())
		conf.ServiceName = fmt.Sprintf("%s_%d", schema, GinkgoParallelProcess())

		Expect(runaidb.Init(conf)).To(Succeed())
	})

	SynchronizedAfterSuite(func() {
		_ = runaidb.Close()
	}, func() {
		if terminateContainer != nil {
			terminateContainer()
		}
	})
}
