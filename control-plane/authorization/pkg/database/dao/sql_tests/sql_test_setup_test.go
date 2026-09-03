/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package sql_tests

import (
	"os"
	"testing"

	authorization_config "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/config"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/orm"
	dbtesting "github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/testing"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	sqlTestTenantID = uint(1)
	sqlTestRoleID   = uint(10)
)

type sqlCountResult struct {
	Count int64
}

func TestMain(m *testing.M) {
	_ = authorization_config.Init()
	exitCode := m.Run()
	runaidb.SetMock(nil)
	os.Exit(exitCode)
}

func TestSQLTests(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Authorization DAO SQL Tests")
}

func setupSQLMock() dbtesting.Mocker {
	return dbtesting.SetSQLMock(GinkgoT())
}

func setupSQLMockWithDB() (dbtesting.Mocker, orm.DB) {
	ormDB, mocker := dbtesting.NewMocker(GinkgoT())
	runaidb.SetMock(ormDB)
	return mocker, ormDB
}
