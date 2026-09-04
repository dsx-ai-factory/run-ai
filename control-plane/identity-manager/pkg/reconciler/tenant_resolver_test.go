/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package reconciler

import (
	"context"
	"errors"
	"testing"

	businessMocks "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/business/mocks"
	pkgModels "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/models"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	testRealm    = "test-realm"
	testTenantId = "tenant-1001"
)

func TestGetRealmTenantId_FromContext(t *testing.T) {
	realmDal := businessMocks.NewRealmDal(t)

	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, testTenantId)
	resolver := NewRealmTenantResolver(ctx, realmDal)

	result := resolver.GetRealmTenantId(ctx, testRealm)

	assert.Equal(t, testTenantId, result)
	realmDal.AssertNotCalled(t, "GetRealmByName", mock.Anything, mock.Anything)
}

func TestGetRealmTenantId_FromDb(t *testing.T) {
	realmDal := businessMocks.NewRealmDal(t)
	realmDal.On("GetRealmByName", mock.Anything, testRealm).
		Return(&pkgModels.Realm{TenantId: testTenantId}, nil)

	resolver := NewRealmTenantResolver(context.Background(), realmDal)

	result := resolver.GetRealmTenantId(context.Background(), testRealm)

	assert.Equal(t, testTenantId, result)
}

func TestGetRealmTenantId_DbError_ReturnsEmpty(t *testing.T) {
	realmDal := businessMocks.NewRealmDal(t)
	realmDal.On("GetRealmByName", mock.Anything, testRealm).
		Return(nil, errors.New("db connection failed"))

	resolver := NewRealmTenantResolver(context.Background(), realmDal)

	result := resolver.GetRealmTenantId(context.Background(), testRealm)

	assert.Empty(t, result)
}

func TestGetRealmTenantId_EmptyContextValue_FallsBackToDb(t *testing.T) {
	realmDal := businessMocks.NewRealmDal(t)
	realmDal.On("GetRealmByName", mock.Anything, testRealm).
		Return(&pkgModels.Realm{TenantId: testTenantId}, nil)

	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, "")
	resolver := NewRealmTenantResolver(ctx, realmDal)

	result := resolver.GetRealmTenantId(ctx, testRealm)

	assert.Equal(t, testTenantId, result)
}
