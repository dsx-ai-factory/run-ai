/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models_crud

import (
	"fmt"
	"github.com/run-ai/runai/control-plane/workloads/internal/generated/pkg/api"
	"time"
)

const (
	TableNamePolicyCompliance        = "policy_compliance"
	PolicyComplianceColumnTenantId   = "tenant_id"
	PolicyComplianceColumnProjectId  = "project_id"
	PolicyComplianceColumnTemplateId = "template_id"
	PolicyComplianceColumnPolicyTime = "policy_time"
)

// PolicyComplianceRecord stores compliance information of a given template to the effective policy of a given project.
// we keep it here in order not to re-validate the policy each time compliance info is requested
type PolicyComplianceRecord struct {
	TenantId     int32  `gorm:"uniqueIndex:by_template_workload_and_project"`
	ProjectId    int32  `gorm:"uniqueIndex:by_template_workload_and_project"`
	TemplateId   string `gorm:"uniqueIndex:by_template_workload_and_project"`
	WorkloadType string `gorm:"uniqueIndex:by_template_workload_and_project"`
	PolicyTime   time.Time
	Aligned      bool                       // is the asset aligned with the policy
	Reason       []api.ComplianceInfoReason `json:"reason,omitempty" gorm:"serializer:json; type:jsonb;"`
	ImposedAsset bool                       // asset is imposed on all workloads by the administrator
}

func (PolicyComplianceRecord) TableName() string {
	return TableNamePolicyCompliance
}

func (PolicyComplianceRecord) TemplateIdColumn() string {
	return PolicyComplianceColumnTemplateId
}

func (record PolicyComplianceRecord) LogInfo() string {
	return fmt.Sprintf("[%d] project %d template %s workload type %s",
		record.TenantId, record.ProjectId, record.TemplateId, record.WorkloadType)
}
