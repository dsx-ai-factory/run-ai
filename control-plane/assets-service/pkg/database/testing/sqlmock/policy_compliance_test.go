/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package sqlmock

import (
	"testing"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dbmodel"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/auth"
	dbtesting "github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/testing"
	"github.com/stretchr/testify/require"
)

const (
	pcFirst     = `SELECT * FROM policy_compliance WHERE policy_compliance.asset_id = ? AND policy_compliance.workload_type = ? ORDER BY policy_compliance.tenant_id LIMIT ?`
	pcInsert    = `INSERT INTO policy_compliance (tenant_id,project_id,asset_id,workload_type,policy_time,aligned,reason,imposed_asset) VALUES (?,?,?,?,?,?,?,?)`
	pcDelNonRel = `DELETE FROM policy_compliance WHERE policy_compliance.workload_type = ? AND policy_time != (?)`
	pcList      = `SELECT * FROM policy_compliance WHERE policy_compliance.workload_type = ? AND asset_id IN (?)`
)

func pcPolicy() dao.PolicyInfo {
	return dao.PolicyInfo{
		WorkloadType: "training",
		PolicyTime:   time.Now(),
	}
}

func TestUpsertPolicyCompliance(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(pcFirst, "asset-1", "training", dbtesting.AnyArg()).Returns([]dbmodel.PolicyComplianceRecord{})
	m.ExpectTransaction().Begin()
	m.ExpectExec(pcInsert).Returns(1)
	m.ExpectTransaction().Commit()
	apiErr := dao.UpsertPolicyCompliance(auth.MockAuthInfo(), "asset-1", pcPolicy(), api.ComplianceInfo{Compliance: true})
	require.Nil(t, apiErr)
}

func TestListPolicyComplianceOfAssets(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectTransaction().Begin()
	m.ExpectExec(pcDelNonRel).Returns(0)
	m.ExpectTransaction().Commit()
	m.ExpectQuery(pcList, "training", dbtesting.AnyArg()).Returns([]dbmodel.PolicyComplianceRecord{{AssetId: "asset-1", WorkloadType: "training", Aligned: true}})
	got, apiErr := dao.ListPolicyComplianceOfAssets(auth.MockAuthInfo(), []string{"asset-1"}, pcPolicy())
	require.Nil(t, apiErr)
	require.NotEmpty(t, got)
}
