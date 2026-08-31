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
	"github.com/run-ai/runai/control-plane/catalog-service/pkg/database/models"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/auth"
	"github.com/stretchr/testify/require"
)

const (
	hrSQLGet         = `SELECT * FROM helm_repos WHERE helm_repos.id = ? AND ((tenant_id IS NULL or tenant_id = '' or tenant_id = ?)) AND helm_repos.id = ?`
	hrSQLList        = `SELECT * FROM helm_repos WHERE (tenant_id IS NULL or tenant_id = '' or tenant_id = ?)`
	hrSQLPreset      = `SELECT * FROM helm_repos WHERE helm_repos.scope_type = ?`
	hrSQLCount       = `SELECT count(*) FROM helm_repos WHERE name = ? AND ((scope_type = ? AND scope_id = ?) OR scope_type = ?)`
	hrSQLCountExcl   = `SELECT count(*) FROM helm_repos WHERE name = ? AND ((scope_type = ? AND scope_id = ?) OR scope_type = ?) AND id != ?`
	hrSQLCreate      = `INSERT INTO helm_repos (id,scope_type,scope_id,tenant_id,cluster_id,department_id,project_id,created_at,created_by,updated_at,updated_by,url,name,is_private,description,version) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`
	hrSQLUpdate      = `UPDATE helm_repos SET id=?,scope_type=?,scope_id=?,tenant_id=?,cluster_id=?,department_id=?,project_id=?,created_at=?,created_by=?,updated_at=?,updated_by=?,url=?,name=?,is_private=?,description=?,version=? WHERE helm_repos.id = ? AND id = ?`
	hrSQLDeleteFirst = `SELECT * FROM helm_repos WHERE helm_repos.id = ? AND ((tenant_id IS NULL or tenant_id = '' or tenant_id = ?)) AND helm_repos.id = ? ORDER BY helm_repos.id LIMIT ?`
	hrSQLDeleteExec  = `DELETE FROM helm_repos WHERE helm_repos.id = ?`
)

// GetHelmRepoById -----------------------------------------------------------

func TestGetHelmRepoById_InvalidUuid(t *testing.T) {
	setupDaoMock(t)
	got, apiErr := GetHelmRepoById(context.Background(), "not-a-uuid")
	require.Nil(t, got)
	require.NotNil(t, apiErr)
}

func TestGetHelmRepoById_Found(t *testing.T) {
	m := setupDaoMock(t)
	id := uuid.New()
	m.ExpectQuery(hrSQLGet).Returns([]models.HelmRepo{{ID: id, Name: "h1"}})
	got, apiErr := GetHelmRepoById(auth.MockAuthInfo(), id.String())
	require.Nil(t, apiErr)
	require.NotNil(t, got)
	require.Equal(t, id, got.ID)
}

func TestGetHelmRepoById_NotFound(t *testing.T) {
	m := setupDaoMock(t)
	id := uuid.New()
	m.ExpectQuery(hrSQLGet).Returns([]models.HelmRepo{})
	got, apiErr := GetHelmRepoById(auth.MockAuthInfo(), id.String())
	require.Nil(t, got)
	require.NotNil(t, apiErr)
}

func TestGetHelmRepoById_DbError(t *testing.T) {
	m := setupDaoMock(t)
	id := uuid.New()
	m.ExpectQuery(hrSQLGet).ReturnsError(errors.New("boom"))
	got, apiErr := GetHelmRepoById(auth.MockAuthInfo(), id.String())
	require.Nil(t, got)
	require.NotNil(t, apiErr)
}

// ListHelmRepos -------------------------------------------------------------

func TestListHelmRepos_Found(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(hrSQLList).Returns([]models.HelmRepo{{ID: uuid.New(), Name: "h1"}, {ID: uuid.New(), Name: "h2"}})
	got, apiErr := ListHelmRepos(auth.MockAuthInfo())
	require.Nil(t, apiErr)
	require.Len(t, got, 2)
}

func TestListHelmRepos_Empty(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(hrSQLList).Returns([]models.HelmRepo{})
	got, apiErr := ListHelmRepos(auth.MockAuthInfo())
	require.Nil(t, apiErr)
	require.Len(t, got, 0)
}

func TestListHelmRepos_DbError(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(hrSQLList).ReturnsError(errors.New("boom"))
	got, apiErr := ListHelmRepos(auth.MockAuthInfo())
	require.Nil(t, got)
	require.NotNil(t, apiErr)
}

// ListPresetHelmRepos -------------------------------------------------------

func TestListPresetHelmRepos_Found(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(hrSQLPreset).Returns([]models.HelmRepo{{ID: uuid.New(), Name: "sys1"}, {ID: uuid.New(), Name: "sys2"}})
	got, apiErr := ListPresetHelmRepos(auth.MockAuthInfo())
	require.Nil(t, apiErr)
	require.Len(t, got, 2)
	require.Contains(t, got, "sys1")
	require.Contains(t, got, "sys2")
}

func TestListPresetHelmRepos_Empty(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(hrSQLPreset).Returns([]models.HelmRepo{})
	got, apiErr := ListPresetHelmRepos(auth.MockAuthInfo())
	require.Nil(t, apiErr)
	require.Len(t, got, 0)
}

func TestListPresetHelmRepos_DbError(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(hrSQLPreset).ReturnsError(errors.New("boom"))
	got, apiErr := ListPresetHelmRepos(auth.MockAuthInfo())
	require.Nil(t, got)
	require.NotNil(t, apiErr)
}

// ValidateHelmRepoNameAvailable ---------------------------------------------

func TestValidateHelmRepoNameAvailable_Available(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(hrSQLCount).Returns([]countRow{{Count: 0}})
	apiErr := ValidateHelmRepoNameAvailable(auth.MockAuthInfo(), "myname", nil)
	require.Nil(t, apiErr)
}

func TestValidateHelmRepoNameAvailable_Conflict(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(hrSQLCount).Returns([]countRow{{Count: 1}})
	apiErr := ValidateHelmRepoNameAvailable(auth.MockAuthInfo(), "myname", nil)
	require.NotNil(t, apiErr)
	require.Equal(t, http.StatusConflict, int(apiErr.Code))
}

func TestValidateHelmRepoNameAvailable_WithExclude(t *testing.T) {
	m := setupDaoMock(t)
	exclude := uuid.New()
	m.ExpectQuery(hrSQLCountExcl).Returns([]countRow{{Count: 0}})
	apiErr := ValidateHelmRepoNameAvailable(auth.MockAuthInfo(), "myname", &exclude)
	require.Nil(t, apiErr)
}

func TestValidateHelmRepoNameAvailable_DbError(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(hrSQLCount).ReturnsError(errors.New("boom"))
	apiErr := ValidateHelmRepoNameAvailable(auth.MockAuthInfo(), "myname", nil)
	require.NotNil(t, apiErr)
}

// CreateHelmRepo ------------------------------------------------------------

func TestCreateHelmRepo_Success(t *testing.T) {
	m := setupDaoMock(t)
	tx := m.ExpectTransaction()
	tx.Begin()
	m.ExpectExec(hrSQLCreate).Returns(1)
	tx.Commit()
	rec := &models.HelmRepo{ID: uuid.New(), Name: "h1"}
	apiErr := CreateHelmRepo(auth.MockAuthInfo(), rec)
	require.Nil(t, apiErr)
}

func TestCreateHelmRepo_DbError(t *testing.T) {
	m := setupDaoMock(t)
	tx := m.ExpectTransaction()
	tx.Begin()
	m.ExpectExec(hrSQLCreate).ReturnsError(errors.New("boom"))
	tx.Rollback()
	rec := &models.HelmRepo{ID: uuid.New(), Name: "h1"}
	apiErr := CreateHelmRepo(auth.MockAuthInfo(), rec)
	require.NotNil(t, apiErr)
}

// UpdateHelmRepo ------------------------------------------------------------

func TestUpdateHelmRepo_Success(t *testing.T) {
	m := setupDaoMock(t)
	tx := m.ExpectTransaction()
	tx.Begin()
	m.ExpectExec(hrSQLUpdate).Returns(1)
	tx.Commit()
	rec := &models.HelmRepo{ID: uuid.New(), Name: "h1"}
	apiErr := UpdateHelmRepo(auth.MockAuthInfo(), rec)
	require.Nil(t, apiErr)
}

func TestUpdateHelmRepo_DbError(t *testing.T) {
	m := setupDaoMock(t)
	tx := m.ExpectTransaction()
	tx.Begin()
	m.ExpectExec(hrSQLUpdate).ReturnsError(errors.New("boom"))
	tx.Rollback()
	rec := &models.HelmRepo{ID: uuid.New(), Name: "h1"}
	apiErr := UpdateHelmRepo(auth.MockAuthInfo(), rec)
	require.NotNil(t, apiErr)
}

// DeleteHelmRepo ------------------------------------------------------------

func TestDeleteHelmRepo_InvalidUuid(t *testing.T) {
	setupDaoMock(t)
	apiErr := DeleteHelmRepo(context.Background(), "not-a-uuid")
	require.NotNil(t, apiErr)
}

func TestDeleteHelmRepo_NotFound(t *testing.T) {
	m := setupDaoMock(t)
	id := uuid.New()
	m.ExpectQuery(hrSQLDeleteFirst).Returns([]models.HelmRepo{})
	apiErr := DeleteHelmRepo(auth.MockAuthInfo(), id.String())
	require.NotNil(t, apiErr)
	require.Equal(t, http.StatusNotFound, int(apiErr.Code))
}

func TestDeleteHelmRepo_Success(t *testing.T) {
	m := setupDaoMock(t)
	id := uuid.New()
	m.ExpectQuery(hrSQLDeleteFirst).Returns([]models.HelmRepo{{ID: id, Name: "h1"}})
	tx := m.ExpectTransaction()
	tx.Begin()
	m.ExpectExec(hrSQLDeleteExec).Returns(1)
	tx.Commit()
	apiErr := DeleteHelmRepo(auth.MockAuthInfo(), id.String())
	require.Nil(t, apiErr)
}

func TestDeleteHelmRepo_DeleteError(t *testing.T) {
	m := setupDaoMock(t)
	id := uuid.New()
	m.ExpectQuery(hrSQLDeleteFirst).Returns([]models.HelmRepo{{ID: id, Name: "h1"}})
	tx := m.ExpectTransaction()
	tx.Begin()
	m.ExpectExec(hrSQLDeleteExec).ReturnsError(errors.New("boom"))
	tx.Rollback()
	apiErr := DeleteHelmRepo(auth.MockAuthInfo(), id.String())
	require.NotNil(t, apiErr)
}
