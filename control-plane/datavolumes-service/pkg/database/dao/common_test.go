/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/run-ai/runai/control-plane/datavolumes-service/internal/generated/pkg/api"
	"github.com/stretchr/testify/require"
)

func mustUUID() uuid.UUID { return uuid.New() }

const (
	scopePluck              = "SELECT id FROM datavolumes WHERE tenant_id = ? AND project_id = ? ORDER BY id ASC LIMIT ?"
	scopeDeleteSharedScopes = "DELETE FROM shared_scopes WHERE datavolume_id IN (?)"
	scopeDeleteDatavolumes  = "DELETE FROM datavolumes WHERE id IN (?)"
)

// ---------- DeleteDataVolumesByScope ----------

func TestDeleteDataVolumesByScope_ScopeIdColumnError_System(t *testing.T) {
	setupDaoMock(t)
	err := DeleteDataVolumesByScope(testCtx("1", nil), "system", "scope-1")
	require.Error(t, err)
}

func TestDeleteDataVolumesByScope_ScopeIdColumnError_Unknown(t *testing.T) {
	setupDaoMock(t)
	err := DeleteDataVolumesByScope(testCtx("1", nil), api.ScopeType("bogus"), "scope-1")
	require.Error(t, err)
}

func TestDeleteDataVolumesByScope_NothingToDelete(t *testing.T) {
	m := setupDaoMock(t)
	tx := m.ExpectTransaction()
	tx.Begin()
	m.ExpectQuery(scopePluck).Returns([]idRow{})
	tx.Commit()
	err := DeleteDataVolumesByScope(testCtx("1", nil), api.SCOPETYPE_PROJECT, "proj-1")
	require.NoError(t, err)
}

func TestDeleteDataVolumesByScope_Success(t *testing.T) {
	m := setupDaoMock(t)
	tx := m.ExpectTransaction()
	tx.Begin()
	m.ExpectQuery(scopePluck).Returns([]idRow{{ID: mustUUID()}})
	m.ExpectExec(scopeDeleteSharedScopes).Returns(1)
	m.ExpectExec(scopeDeleteDatavolumes).Returns(1)
	// second loop iteration - nothing left to delete
	m.ExpectQuery(scopePluck).Returns([]idRow{})
	tx.Commit()
	err := DeleteDataVolumesByScope(testCtx("1", nil), api.SCOPETYPE_PROJECT, "proj-1")
	require.NoError(t, err)
}

func TestDeleteDataVolumesByScope_PluckError(t *testing.T) {
	m := setupDaoMock(t)
	tx := m.ExpectTransaction()
	tx.Begin()
	m.ExpectQuery(scopePluck).ReturnsError(errors.New("boom"))
	tx.Rollback()
	err := DeleteDataVolumesByScope(testCtx("1", nil), api.SCOPETYPE_PROJECT, "proj-1")
	require.Error(t, err)
}

func TestDeleteDataVolumesByScope_DeleteSharedScopesError(t *testing.T) {
	m := setupDaoMock(t)
	tx := m.ExpectTransaction()
	tx.Begin()
	m.ExpectQuery(scopePluck).Returns([]idRow{{ID: mustUUID()}})
	m.ExpectExec(scopeDeleteSharedScopes).ReturnsError(errors.New("boom"))
	tx.Rollback()
	err := DeleteDataVolumesByScope(testCtx("1", nil), api.SCOPETYPE_PROJECT, "proj-1")
	require.Error(t, err)
}

func TestDeleteDataVolumesByScope_DeleteDatavolumesError(t *testing.T) {
	m := setupDaoMock(t)
	tx := m.ExpectTransaction()
	tx.Begin()
	m.ExpectQuery(scopePluck).Returns([]idRow{{ID: mustUUID()}})
	m.ExpectExec(scopeDeleteSharedScopes).Returns(1)
	m.ExpectExec(scopeDeleteDatavolumes).ReturnsError(errors.New("boom"))
	tx.Rollback()
	err := DeleteDataVolumesByScope(testCtx("1", nil), api.SCOPETYPE_PROJECT, "proj-1")
	require.Error(t, err)
}
