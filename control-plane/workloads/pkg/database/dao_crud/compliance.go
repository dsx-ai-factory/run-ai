/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao_crud

import (
	"context"
	"fmt"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/auth"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/models_crud"
	dao_utils "github.com/dsx-ai-factory/run-ai/runai-common-packages/dao-utils"
	"gorm.io/gorm"
	"net/http"
	"time"
)

// PolicyInfo holds information about the policy that the template is expected to comply with
type PolicyInfo struct {
	TenantId     int32
	WorkloadType string
	ReplicaType  string
	ProjectId    int32
	PolicyTime   time.Time // the update time of the policy that the asset complies with. if policy changes, need to re-check compliance
}

// WorkloadAndReplicaType returns the complete type identifier which should include the workload
// type itself, but also the replica type, because the compliance of template to master and worker
// may be different.
func (pi *PolicyInfo) WorkloadAndReplicaType() string {
	if pi.ReplicaType != "" {
		return fmt.Sprintf("%s:%s", pi.WorkloadType, pi.ReplicaType)
	} else {
		return pi.WorkloadType
	}
}

func UpsertPolicyComplianceOfTemplate(ctx context.Context, templateId string, whichPolicy PolicyInfo, reason []api.ComplianceInfoReason) *api.Error {
	tx := dao_utils.DB().WithContext(ctx)

	record := models_crud.PolicyComplianceRecord{
		TenantId:     whichPolicy.TenantId,
		ProjectId:    whichPolicy.ProjectId,
		WorkloadType: whichPolicy.WorkloadAndReplicaType(),
		TemplateId:   templateId,

		PolicyTime: whichPolicy.PolicyTime,
		Aligned:    len(reason) == 0,
		Reason:     reason,
	}

	result := tx.Where(models_crud.PolicyComplianceRecord{
		TenantId:     record.TenantId,
		ProjectId:    record.ProjectId,
		TemplateId:   record.TemplateId,
		WorkloadType: whichPolicy.WorkloadAndReplicaType(),
	}).Assign(record).FirstOrCreate(&record)

	if result.Error != nil {
		return api.NewError(http.StatusInternalServerError,
			fmt.Sprintf("failed to upsert policy <-> template compliance for %s", record.LogInfo()), result.Error).WithDetailsFromDb(result)
	}

	return nil
}

func ListPolicyComplianceOfTemplates(ctx context.Context, assetIds []string, whichPolicy PolicyInfo) ([]models_crud.PolicyComplianceRecord, *api.Error) {
	tx := dao_utils.DB().WithContext(ctx)

	//
	//    remove from policy compliance all the entries of this policy that their policy time is different
	//    from the update time of the policy, as they are no longer up to date (policy changed after record created
	//
	err := deleteNonRelevantCompliance(ctx, tx, whichPolicy)
	if err != nil {
		return nil, err
	}

	projectAndWorkloadKey := models_crud.PolicyComplianceRecord{
		TenantId:     whichPolicy.TenantId,
		ProjectId:    whichPolicy.ProjectId,
		WorkloadType: whichPolicy.WorkloadAndReplicaType(),
	}

	complianceRecords := make([]models_crud.PolicyComplianceRecord, 0)

	result := tx.Model(models_crud.PolicyComplianceRecord{}).
		Where(&projectAndWorkloadKey).
		Where(fmt.Sprintf("%s IN (?)", models_crud.PolicyComplianceColumnTemplateId), assetIds).
		Scan(&complianceRecords)
	if result.Error != nil {
		return nil, api.NewError(http.StatusInternalServerError, fmt.Sprintf("failed to list policy <-> template compliance for %s", projectAndWorkloadKey.LogInfo()), result.Error).WithDetailsFromDb(result)
	}

	return complianceRecords, nil
}

// delete ALL the policy compliance information of a given template.
func deletePolicyComplianceOfTemplate(ctx context.Context, tx *gorm.DB, templateId string) *api.Error {
	searchCriteria := models_crud.PolicyComplianceRecord{
		TenantId:   auth.TenantIdOf(ctx),
		TemplateId: templateId,
	}

	result := tx.Model(models_crud.PolicyComplianceRecord{}).Where(&searchCriteria).Delete(&searchCriteria)
	if result.Error != nil {
		return api.NewError(http.StatusInternalServerError, fmt.Sprintf("failed to delete policy <-> template compliance for %s", searchCriteria.LogInfo()), result.Error).WithDetailsFromDb(result)
	}

	return nil
}

// Given a policy that updated in date X, we have to remove from the DB all records related to this policy
// that have policy update time Y, because they are no longer relevant
func deleteNonRelevantCompliance(_ context.Context, tx *gorm.DB, whichPolicy PolicyInfo) *api.Error {
	projectAndTypeKey := models_crud.PolicyComplianceRecord{
		TenantId:     whichPolicy.TenantId,
		ProjectId:    whichPolicy.ProjectId,
		WorkloadType: whichPolicy.WorkloadAndReplicaType(),
	}

	result := tx.Model(models_crud.PolicyComplianceRecord{}).
		Where(&projectAndTypeKey).
		Where(fmt.Sprintf("%s != (?)", models_crud.PolicyComplianceColumnPolicyTime), whichPolicy.PolicyTime).
		Delete(&projectAndTypeKey)
	if result.Error != nil {
		return api.NewError(http.StatusInternalServerError, fmt.Sprintf("failed to delete policy <-> template compliance for %s", projectAndTypeKey.LogInfo()), result.Error).WithDetailsFromDb(result)
	}

	return nil
}
