/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package sqlmock

import (
	"testing"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/query_types"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/auth"
	dbtesting "github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/testing"
	"github.com/stretchr/testify/require"
)

const listWhere = ` WHERE (assets_metadata.auto_delete is null or assets_metadata.auto_delete = ?)` +
	` AND COALESCE(assets_metadata.private, FALSE) = false AND assets_metadata.kind = ? AND ` +
	pvcScope + ` AND assets_metadata.deleted_at IS NULL`

const (
	pvcList = pvcCols + ` ` + pvcJoins + listWhere
	cmList  = cmCols + ` ` + cmJoins + listWhere
	svList  = svCols + ` ` + svJoins + listWhere
)

func listQueryParams(t *testing.T) *query_types.QueryParams {
	qp, err := query_types.NewQueryParams("", "", "", false, 0, "", "", nil)
	require.Nil(t, err)
	return qp
}

func TestListPvcAssets(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(pvcList, dbtesting.AnyArg(), "pvc", dbtesting.AnyArg(), dbtesting.AnyArg()).Returns(pvcRow())
	got, apiErr := dao.ListPvcAssets(auth.MockAuthInfo(), listQueryParams(t), dao.DatasourcesOptions{})
	require.Nil(t, apiErr)
	require.NotNil(t, got)
	require.GreaterOrEqual(t, len(got.Entries), 1)
}

func TestListConfigMapAssets(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(cmList, dbtesting.AnyArg(), "config-map", dbtesting.AnyArg(), dbtesting.AnyArg()).Returns(cmRow())
	got, apiErr := dao.ConfigMapsDBAccessor.ListConfigMapAssets(auth.MockAuthInfo(), listQueryParams(t))
	require.Nil(t, apiErr)
	require.NotNil(t, got)
	require.GreaterOrEqual(t, len(got.Entries), 1)
}

func TestListSecretVolumeAssets(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(svList, dbtesting.AnyArg(), "secret-volume", dbtesting.AnyArg(), dbtesting.AnyArg()).Returns(svRow())
	got, apiErr := dao.NewSecretVolumeDBAccessor().ListSecretVolumeAssets(auth.MockAuthInfo(), listQueryParams(t))
	require.Nil(t, apiErr)
	require.NotNil(t, got)
	require.GreaterOrEqual(t, len(got.Entries), 1)
}
