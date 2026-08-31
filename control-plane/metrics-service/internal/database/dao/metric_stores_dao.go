/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"context"
	"errors"

	"github.com/run-ai/runai/control-plane/metrics-service/internal/database/models"
	metricerrors "github.com/run-ai/runai/control-plane/metrics-service/internal/errors"
	"github.com/run-ai/runai/runai-common-packages/runaidb"
	"gorm.io/gorm"
)

type MetricStoresDao struct{}

func NewMetricStoresDao() *MetricStoresDao {
	return &MetricStoresDao{}
}

func (msd *MetricStoresDao) Create(ctx context.Context, metricStore models.MetricStore) (*models.MetricStore, error) {
	_, err := runaidb.DB().WithContext(ctx).
		Table(models.MetricStoreTable).
		Create(&metricStore)

	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, metricerrors.ErrMetricStoreAlreadyExists
		}
		return nil, err
	}

	return &metricStore, nil
}

func (msd *MetricStoresDao) GetByTenantId(ctx context.Context, tenantId int32) (*models.MetricStore, error) {
	tms := []models.MetricStore{}
	err := runaidb.DB().WithContext(ctx).
		Table(models.MetricStoreTable).
		Where(models.MetricStore{TenantId: tenantId}).
		Find(&tms)
	if err != nil {
		return nil, err
	}

	if len(tms) == 0 {
		return nil, nil
	}
	return &tms[0], nil
}

func (msd *MetricStoresDao) Delete(ctx context.Context, tenantId int32) error {
	_, err := runaidb.DB().WithContext(ctx).
		Table(models.MetricStoreTable).
		Where(models.MetricStore{TenantId: tenantId}).
		Unscoped().
		Delete(&models.MetricStore{})
	return err
}
