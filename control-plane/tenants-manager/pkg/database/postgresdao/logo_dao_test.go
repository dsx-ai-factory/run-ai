/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package postgresdao_test

import (
	"context"
	"testing"

	"github.com/run-ai/runai/control-plane/tenants-manager/pkg/database/dao"
	"github.com/run-ai/runai/control-plane/tenants-manager/pkg/database/postgresdao"
	"github.com/run-ai/runai/control-plane/tenants-manager/pkg/domain/models"
	"github.com/stretchr/testify/assert"
)

func TestSetByTenantId_Creates(t *testing.T) {
	t.Skip(" skipping TestSetByTenantId_Creates")

	logoDaoSubject := &postgresdao.Logo{}
	logoData := []byte("test")
	actionErr := logoDaoSubject.SetByTenantId(context.Background(), 1, &logoData)
	assert.Nil(t, actionErr)
	result := models.Logo{
		TenantId: 1,
	}
	postgresdao.DB().First(&result)
	assert.Equal(t, result.LogoPng, &logoData)
}

func TestSetByTenantId_Updates(t *testing.T) {
	t.Skip(" skipping TestSetByTenantId_Updates")

	logoDaoSubject := &postgresdao.Logo{}
	logoData := []byte("test")
	logoDataUpdate := []byte("test2")
	actionErr := logoDaoSubject.SetByTenantId(context.Background(), 1, &logoData)
	assert.Nil(t, actionErr)
	actionErr = logoDaoSubject.SetByTenantId(context.Background(), 1, &logoDataUpdate)
	assert.Nil(t, actionErr)

	result := models.Logo{
		TenantId: 1,
	}
	postgresdao.DB().First(&result)
	assert.Equal(t, result.LogoPng, &logoDataUpdate)
}

func TestGetByTenantId_Gets(t *testing.T) {
	t.Skip(" skipping TestGetByTenantId_Gets")

	logoDaoSubject := &postgresdao.Logo{}
	logoData := []byte("test")
	actionErr := logoDaoSubject.SetByTenantId(context.Background(), 1, &logoData)
	assert.Nil(t, actionErr)
	result, actionErr := logoDaoSubject.GetByTenantId(context.Background(), 1)
	assert.Nil(t, actionErr)
	assert.Equal(t, result, &logoData)
}

func TestGetByTenantId_GetShouldFail(t *testing.T) {
	t.Skip(" skipping TestGetByTenantId_GetShouldFail")

	logoDaoSubject := &postgresdao.Logo{}
	logoData := []byte("test")
	actionErr := logoDaoSubject.SetByTenantId(context.Background(), 1, &logoData)
	assert.Nil(t, actionErr)
	result, actionErr := logoDaoSubject.GetByTenantId(context.Background(), 2)
	assert.Equal(t, dao.NoContent, actionErr.Reason)
	assert.Equal(t, actionErr.Err.Error(), "record not found")
	assert.Nil(t, result)
}
