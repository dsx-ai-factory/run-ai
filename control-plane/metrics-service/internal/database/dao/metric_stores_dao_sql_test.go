/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"context"
	"testing"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/internal/database/models"
	metricerrors "github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/internal/errors"
	dbtesting "github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/testing"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

const (
	sqlInsertMetricStore = `INSERT INTO metric_stores (id,deleted_at,tenant_id,datasource_id,metric_store_details) VALUES (?,?,?,?,?) RETURNING created_at,updated_at`
	sqlSelectMetricStore = `SELECT * FROM metric_stores WHERE metric_stores.tenant_id = ? AND metric_stores.deleted_at IS NULL`
	sqlDeleteMetricStore = `DELETE FROM metric_stores WHERE metric_stores.tenant_id = ?`
)

// insertReturning is the pair of columns the database fills in on insert.
type insertReturning struct {
	CreatedAt time.Time
	UpdatedAt time.Time
}

// metricStoreRow carries the jsonb column as raw bytes, which is what its Scan method expects.
type metricStoreRow struct {
	ID                 uuid.UUID `gorm:"type:uuid;primaryKey"`
	CreatedAt          time.Time
	UpdatedAt          time.Time
	DeletedAt          gorm.DeletedAt
	TenantId           int32
	DatasourceId       string
	MetricStoreDetails []byte
}

func TestMetricStoresDaoSQLCreate(t *testing.T) {
	store := models.MetricStore{TenantId: testTenantID, DatasourceId: "mimir_1"}

	t.Run("returns the created store", func(t *testing.T) {
		mocker := dbtesting.SetSQLMock(t)
		tx := mocker.ExpectTransaction()
		tx.Begin()
		mocker.ExpectQuery(sqlInsertMetricStore).Returns([]insertReturning{{CreatedAt: time.Now(), UpdatedAt: time.Now()}})
		tx.Commit()

		got, err := NewMetricStoresDao().Create(context.Background(), store)

		require.NoError(t, err)
		require.Equal(t, testTenantID, got.TenantId)
		require.NotEqual(t, uuid.Nil, got.ID)
	})

	t.Run("a duplicate tenant maps to ErrMetricStoreAlreadyExists", func(t *testing.T) {
		mocker := dbtesting.SetSQLMock(t)
		tx := mocker.ExpectTransaction()
		tx.Begin()
		mocker.ExpectQuery(sqlInsertMetricStore).ReturnsError(gorm.ErrDuplicatedKey)
		tx.Rollback()

		got, err := NewMetricStoresDao().Create(context.Background(), store)

		require.ErrorIs(t, err, metricerrors.ErrMetricStoreAlreadyExists)
		require.Nil(t, got)
	})

	t.Run("other db failures are returned unchanged", func(t *testing.T) {
		mocker := dbtesting.SetSQLMock(t)
		tx := mocker.ExpectTransaction()
		tx.Begin()
		mocker.ExpectQuery(sqlInsertMetricStore).ReturnsError(errDB)
		tx.Rollback()

		got, err := NewMetricStoresDao().Create(context.Background(), store)

		require.ErrorIs(t, err, errDB)
		require.NotErrorIs(t, err, metricerrors.ErrMetricStoreAlreadyExists)
		require.Nil(t, got)
	})
}

func TestMetricStoresDaoSQLGetByTenantId(t *testing.T) {
	t.Run("returns the tenant's store", func(t *testing.T) {
		mocker := dbtesting.SetSQLMock(t)
		mocker.ExpectQuery(sqlSelectMetricStore, testTenantID).Returns([]metricStoreRow{{
			ID:                 uuid.New(),
			TenantId:           testTenantID,
			DatasourceId:       "mimir_1",
			MetricStoreDetails: []byte(`{"type":"mimir"}`),
		}})

		got, err := NewMetricStoresDao().GetByTenantId(context.Background(), testTenantID)

		require.NoError(t, err)
		require.NotNil(t, got)
		require.Equal(t, testTenantID, got.TenantId)
		require.Equal(t, "mimir_1", got.DatasourceId)
	})

	t.Run("no rows returns nil without an error", func(t *testing.T) {
		mocker := dbtesting.SetSQLMock(t)
		mocker.ExpectQuery(sqlSelectMetricStore, testTenantID).Returns([]metricStoreRow{})

		got, err := NewMetricStoresDao().GetByTenantId(context.Background(), testTenantID)

		require.NoError(t, err)
		require.Nil(t, got)
	})

	t.Run("db failure is returned", func(t *testing.T) {
		mocker := dbtesting.SetSQLMock(t)
		mocker.ExpectQuery(sqlSelectMetricStore, testTenantID).ReturnsError(errDB)

		got, err := NewMetricStoresDao().GetByTenantId(context.Background(), testTenantID)

		require.ErrorIs(t, err, errDB)
		require.Nil(t, got)
	})
}

func TestMetricStoresDaoSQLDelete(t *testing.T) {
	t.Run("hard deletes the tenant's store", func(t *testing.T) {
		mocker := dbtesting.SetSQLMock(t)
		expectExecInTx(mocker, 1, sqlDeleteMetricStore, testTenantID)

		require.NoError(t, NewMetricStoresDao().Delete(context.Background(), testTenantID))
	})

	t.Run("deleting a store that does not exist is not an error", func(t *testing.T) {
		mocker := dbtesting.SetSQLMock(t)
		expectExecInTx(mocker, 0, sqlDeleteMetricStore, testTenantID)

		require.NoError(t, NewMetricStoresDao().Delete(context.Background(), testTenantID))
	})

	t.Run("db failure is returned", func(t *testing.T) {
		mocker := dbtesting.SetSQLMock(t)
		expectExecErrorInTx(mocker, sqlDeleteMetricStore, testTenantID)

		require.ErrorIs(t, NewMetricStoresDao().Delete(context.Background(), testTenantID), errDB)
	})
}
