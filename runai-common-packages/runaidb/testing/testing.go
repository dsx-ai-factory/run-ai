/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package testing

import (
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/orm/mocks"
	"github.com/stretchr/testify/mock"
)

type TestingT interface {
	mock.TestingT
	Cleanup(func())
}

// SetOrmMock sets a mocked DB to assert the DB interface methods that are called.
func SetOrmMock(t TestingT) *mocks.DB {
	m := mocks.NewDB(t)
	runaidb.SetMock(m)
	t.Cleanup(func() { runaidb.SetMock(nil) })
	return m
}

// SetSQLMock sets a mocked DB to assert the SQL queries that are executed.
func SetSQLMock(t TestingT) Mocker {
	registerEncryptionSerializer(t)
	db, mocker := NewMocker(t)
	runaidb.SetMock(db)
	t.Cleanup(func() {
		runaidb.SetMock(nil)
	})
	return mocker
}
