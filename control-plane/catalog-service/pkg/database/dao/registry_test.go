/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/dsx-ai-factory/run-ai/control-plane/catalog-service/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/auth"
	"github.com/stretchr/testify/require"
)

type countRow struct {
	Count int64
}

const (
	sqlGet         = `SELECT * FROM registries WHERE registries.id = ? AND ((tenant_id IS NULL or tenant_id = '' or tenant_id = ?)) AND registries.id = ?`
	sqlList        = `SELECT * FROM registries WHERE (tenant_id IS NULL or tenant_id = '' or tenant_id = ?)`
	sqlPreset      = `SELECT * FROM registries WHERE registries.scope_type = ?`
	sqlCount       = `SELECT count(*) FROM registries WHERE name = ? AND ((scope_type = ? AND scope_id = ?) OR scope_type = ?)`
	sqlCountExcl   = `SELECT count(*) FROM registries WHERE name = ? AND ((scope_type = ? AND scope_id = ?) OR scope_type = ?) AND id != ?`
	sqlCreate      = `INSERT INTO registries (id,scope_type,scope_id,tenant_id,cluster_id,department_id,project_id,created_at,created_by,updated_at,updated_by,url,name,description,version) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`
	sqlUpdate      = `UPDATE registries SET id=?,scope_type=?,scope_id=?,tenant_id=?,cluster_id=?,department_id=?,project_id=?,created_at=?,created_by=?,updated_at=?,updated_by=?,url=?,name=?,description=?,version=? WHERE registries.id = ? AND id = ?`
	sqlDeleteFirst = `SELECT * FROM registries WHERE registries.id = ? AND ((tenant_id IS NULL or tenant_id = '' or tenant_id = ?)) AND registries.id = ? ORDER BY registries.id LIMIT ?`
	sqlDeleteExec  = `DELETE FROM registries WHERE registries.id = ?`
)

// GetRegistryById -----------------------------------------------------------

func TestGetRegistryById_InvalidUuid(t *testing.T) {
	setupDaoMock(t)
	got, apiErr := GetRegistryById(context.Background(), "not-a-uuid")
	require.Nil(t, got)
	require.NotNil(t, apiErr)
}

func TestGetRegistryById_Found(t *testing.T) {
	m := setupDaoMock(t)
	id := uuid.New()
	m.ExpectQuery(sqlGet).Returns([]models.Registry{{ID: id, Name: "r1"}})
	got, apiErr := GetRegistryById(auth.MockAuthInfo(), id.String())
	require.Nil(t, apiErr)
	require.NotNil(t, got)
	require.Equal(t, id, got.ID)
}

func TestGetRegistryById_NotFound(t *testing.T) {
	m := setupDaoMock(t)
	id := uuid.New()
	m.ExpectQuery(sqlGet).Returns([]models.Registry{})
	got, apiErr := GetRegistryById(auth.MockAuthInfo(), id.String())
	require.Nil(t, got)
	require.NotNil(t, apiErr)
}

func TestGetRegistryById_DbError(t *testing.T) {
	m := setupDaoMock(t)
	id := uuid.New()
	m.ExpectQuery(sqlGet).ReturnsError(errors.New("boom"))
	got, apiErr := GetRegistryById(auth.MockAuthInfo(), id.String())
	require.Nil(t, got)
	require.NotNil(t, apiErr)
}

// ListRegistries ------------------------------------------------------------

func TestListRegistries_Found(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(sqlList).Returns([]models.Registry{{ID: uuid.New(), Name: "r1"}, {ID: uuid.New(), Name: "r2"}})
	got, apiErr := ListRegistries(auth.MockAuthInfo())
	require.Nil(t, apiErr)
	require.Len(t, got, 2)
}

func TestListRegistries_Empty(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(sqlList).Returns([]models.Registry{})
	got, apiErr := ListRegistries(auth.MockAuthInfo())
	require.Nil(t, apiErr)
	require.Len(t, got, 0)
}

func TestListRegistries_DbError(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(sqlList).ReturnsError(errors.New("boom"))
	got, apiErr := ListRegistries(auth.MockAuthInfo())
	require.Nil(t, got)
	require.NotNil(t, apiErr)
}

// ListPresetRegistries ------------------------------------------------------

func TestListPresetRegistries_Found(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(sqlPreset).Returns([]models.Registry{{ID: uuid.New(), Name: "sys1"}, {ID: uuid.New(), Name: "sys2"}})
	got, apiErr := ListPresetRegistries(auth.MockAuthInfo())
	require.Nil(t, apiErr)
	require.Len(t, got, 2)
	require.Contains(t, got, "sys1")
	require.Contains(t, got, "sys2")
}

func TestListPresetRegistries_Empty(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(sqlPreset).Returns([]models.Registry{})
	got, apiErr := ListPresetRegistries(auth.MockAuthInfo())
	require.Nil(t, apiErr)
	require.Len(t, got, 0)
}

func TestListPresetRegistries_DbError(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(sqlPreset).ReturnsError(errors.New("boom"))
	got, apiErr := ListPresetRegistries(auth.MockAuthInfo())
	require.Nil(t, got)
	require.NotNil(t, apiErr)
}

// ValidateRegistryNameAvailable ---------------------------------------------

func TestValidateRegistryNameAvailable_Available(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(sqlCount).Returns([]countRow{{Count: 0}})
	apiErr := ValidateRegistryNameAvailable(auth.MockAuthInfo(), "myname", nil)
	require.Nil(t, apiErr)
}

func TestValidateRegistryNameAvailable_Conflict(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(sqlCount).Returns([]countRow{{Count: 1}})
	apiErr := ValidateRegistryNameAvailable(auth.MockAuthInfo(), "myname", nil)
	require.NotNil(t, apiErr)
	require.Equal(t, http.StatusConflict, int(apiErr.Code))
}

func TestValidateRegistryNameAvailable_WithExclude(t *testing.T) {
	m := setupDaoMock(t)
	exclude := uuid.New()
	m.ExpectQuery(sqlCountExcl).Returns([]countRow{{Count: 0}})
	apiErr := ValidateRegistryNameAvailable(auth.MockAuthInfo(), "myname", &exclude)
	require.Nil(t, apiErr)
}

func TestValidateRegistryNameAvailable_DbError(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(sqlCount).ReturnsError(errors.New("boom"))
	apiErr := ValidateRegistryNameAvailable(auth.MockAuthInfo(), "myname", nil)
	require.NotNil(t, apiErr)
}

// CreateRegistry ------------------------------------------------------------

func TestCreateRegistry_Success(t *testing.T) {
	m := setupDaoMock(t)
	tx := m.ExpectTransaction()
	tx.Begin()
	m.ExpectExec(sqlCreate).Returns(1)
	tx.Commit()
	reg := &models.Registry{ID: uuid.New(), Name: "r1"}
	apiErr := CreateRegistry(auth.MockAuthInfo(), reg)
	require.Nil(t, apiErr)
}

func TestCreateRegistry_DbError(t *testing.T) {
	m := setupDaoMock(t)
	tx := m.ExpectTransaction()
	tx.Begin()
	m.ExpectExec(sqlCreate).ReturnsError(errors.New("boom"))
	tx.Rollback()
	reg := &models.Registry{ID: uuid.New(), Name: "r1"}
	apiErr := CreateRegistry(auth.MockAuthInfo(), reg)
	require.NotNil(t, apiErr)
}

// UpdateRegistry ------------------------------------------------------------

func TestUpdateRegistry_Success(t *testing.T) {
	m := setupDaoMock(t)
	tx := m.ExpectTransaction()
	tx.Begin()
	m.ExpectExec(sqlUpdate).Returns(1)
	tx.Commit()
	reg := &models.Registry{ID: uuid.New(), Name: "r1"}
	apiErr := UpdateRegistry(auth.MockAuthInfo(), reg)
	require.Nil(t, apiErr)
}

func TestUpdateRegistry_DbError(t *testing.T) {
	m := setupDaoMock(t)
	tx := m.ExpectTransaction()
	tx.Begin()
	m.ExpectExec(sqlUpdate).ReturnsError(errors.New("boom"))
	tx.Rollback()
	reg := &models.Registry{ID: uuid.New(), Name: "r1"}
	apiErr := UpdateRegistry(auth.MockAuthInfo(), reg)
	require.NotNil(t, apiErr)
}

// DeleteRegistry ------------------------------------------------------------

func TestDeleteRegistry_InvalidUuid(t *testing.T) {
	setupDaoMock(t)
	apiErr := DeleteRegistry(context.Background(), "not-a-uuid")
	require.NotNil(t, apiErr)
}

func TestDeleteRegistry_NotFound(t *testing.T) {
	m := setupDaoMock(t)
	id := uuid.New()
	m.ExpectQuery(sqlDeleteFirst).Returns([]models.Registry{})
	apiErr := DeleteRegistry(auth.MockAuthInfo(), id.String())
	require.NotNil(t, apiErr)
	require.Equal(t, http.StatusNotFound, int(apiErr.Code))
}

func TestDeleteRegistry_Success(t *testing.T) {
	m := setupDaoMock(t)
	id := uuid.New()
	m.ExpectQuery(sqlDeleteFirst).Returns([]models.Registry{{ID: id, Name: "r1"}})
	tx := m.ExpectTransaction()
	tx.Begin()
	m.ExpectExec(sqlDeleteExec).Returns(1)
	tx.Commit()
	apiErr := DeleteRegistry(auth.MockAuthInfo(), id.String())
	require.Nil(t, apiErr)
}

func TestDeleteRegistry_DeleteError(t *testing.T) {
	m := setupDaoMock(t)
	id := uuid.New()
	m.ExpectQuery(sqlDeleteFirst).Returns([]models.Registry{{ID: id, Name: "r1"}})
	tx := m.ExpectTransaction()
	tx.Begin()
	m.ExpectExec(sqlDeleteExec).ReturnsError(errors.New("boom"))
	tx.Rollback()
	apiErr := DeleteRegistry(auth.MockAuthInfo(), id.String())
	require.NotNil(t, apiErr)
}
