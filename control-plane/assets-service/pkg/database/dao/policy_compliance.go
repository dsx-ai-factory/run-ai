/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"context"
	"fmt"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dbmodel"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dbtables"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/auth"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/orm"
)

const ComplianceAsAssetKind = api.AssetKind("compliance")

// PolicyInfo holds information about the policy that an asset is expected to comply with
type PolicyInfo struct {
	TenantId     int32
	WorkloadType string
	ReplicaType  string
	ProjectId    int32
	PolicyTime   time.Time // the update time of the policy that the asset complies with. if policy changes, need to re-check compliance
}

// WorkloadAndReplicaType returns the complete type identifier which should include the workload
// type itself, but also the replica type, because the compliance of asset to master and worker may be different.
func (pi *PolicyInfo) WorkloadAndReplicaType() string {
	if pi.ReplicaType != "" {
		return fmt.Sprintf("%s:%s", pi.WorkloadType, pi.ReplicaType)
	} else {
		return pi.WorkloadType
	}
}

func UpsertPolicyCompliance(ctx context.Context, assetId string, whichPolicy PolicyInfo, complianceInfo api.ComplianceInfo) *api.Error {
	tx := runaidb.DB().WithContext(ctx)

	record := dbmodel.PolicyComplianceRecord{
		TenantId:     whichPolicy.TenantId,
		ProjectId:    whichPolicy.ProjectId,
		WorkloadType: whichPolicy.WorkloadAndReplicaType(),
		AssetId:      assetId,

		PolicyTime:   whichPolicy.PolicyTime,
		Aligned:      complianceInfo.Compliance,
		Reason:       complianceInfo.Reason,
		ImposedAsset: complianceInfo.Imposed,
	}

	_, err := tx.Where(dbmodel.PolicyComplianceRecord{
		TenantId:     record.TenantId,
		ProjectId:    record.ProjectId,
		AssetId:      record.AssetId,
		WorkloadType: whichPolicy.WorkloadAndReplicaType(),
	}).Assign(record, map[string]interface{}{
		// Struct assignment skips false values when a cached result becomes noncompliant.
		"Aligned": record.Aligned,
	}).FirstOrCreate(&record)
	if err != nil {
		return database.Error(database.UPSERT, ComplianceAsAssetKind).WithDetails(err)
	}

	return nil
}

func ListPolicyComplianceOfAssets(ctx context.Context, assetIds []string, whichPolicy PolicyInfo) ([]dbmodel.PolicyComplianceRecord, *api.Error) {
	tx := runaidb.DB().WithContext(ctx)

	//
	//    remove from policy compliance all the entries of this policy that their policy time is different
	//    from the update time of the policy, as they are no longer up to date (policy changed after record created
	//
	err := deleteNonRelevantCompliance(ctx, tx, whichPolicy)
	if err != nil {
		return nil, err
	}

	projectAndWorkloadKey := dbmodel.PolicyComplianceRecord{
		TenantId:     whichPolicy.TenantId,
		ProjectId:    whichPolicy.ProjectId,
		WorkloadType: whichPolicy.WorkloadAndReplicaType(),
	}

	complianceRecords := make([]dbmodel.PolicyComplianceRecord, 0)

	_, scanErr := tx.Model(dbmodel.PolicyComplianceRecord{}).
		Where(&projectAndWorkloadKey).
		Where(fmt.Sprintf("%s IN (?)", dbtables.PolicyComplianceColumnAssetId), assetIds).
		Scan(&complianceRecords)
	if scanErr != nil {
		return nil, database.Error(database.LIST, ComplianceAsAssetKind).WithDetails(scanErr)
	}

	return complianceRecords, nil
}

// delete ALL the policy compliance information of a given asset.
func deletePolicyComplianceOfAsset(ctx context.Context, tx orm.DB, assetId string) *api.Error {
	searchCriteria := dbmodel.PolicyComplianceRecord{
		TenantId: auth.TenantIdOf(ctx),
		AssetId:  assetId,
	}

	rows, err := tx.Model(dbmodel.PolicyComplianceRecord{}).Where(&searchCriteria).Delete(&searchCriteria)
	if err != nil {
		return database.Error(database.DELETE, ComplianceAsAssetKind).WithDetailsFromDbResult(rows, err)
	}

	return nil
}

// Given a policy that updated in date X, we have to remove from the DB all records related to this policy
// that have policy update time Y, because they are no longer relevant
func deleteNonRelevantCompliance(_ context.Context, tx orm.DB, whichPolicy PolicyInfo) *api.Error {
	projectAndTypeKey := dbmodel.PolicyComplianceRecord{
		TenantId:     whichPolicy.TenantId,
		ProjectId:    whichPolicy.ProjectId,
		WorkloadType: whichPolicy.WorkloadAndReplicaType(),
	}

	if _, delErr := tx.Model(dbmodel.PolicyComplianceRecord{}).
		Where(&projectAndTypeKey).
		Where(fmt.Sprintf("%s != (?)", dbtables.PolicyComplianceColumnPolicyTime), whichPolicy.PolicyTime).
		Delete(&projectAndTypeKey); delErr != nil {
		return database.Error(database.DELETE, ComplianceAsAssetKind).WithDetails(delErr)
	}

	return nil
}
